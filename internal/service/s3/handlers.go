package s3

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	xmlHeader      = `<?xml version="1.0" encoding="UTF-8"?>`
	s3Namespace    = "http://s3.amazonaws.com/doc/2006-03-01/"
	timeFormatISO  = "2006-01-02T15:04:05.000Z"
	timeFormatHTTP = "Mon, 02 Jan 2006 15:04:05 GMT"
)

// ListBuckets handles GET / - list all buckets.
func (s *Service) ListBuckets(w http.ResponseWriter, r *http.Request) {
	buckets, err := s.storage.ListBuckets(r.Context())
	if err != nil {
		writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	bucketInfos := make([]BucketInfo, len(buckets))
	for i, b := range buckets {
		bucketInfos[i] = BucketInfo{
			Name:         b.Name,
			CreationDate: b.CreationDate.Format(timeFormatISO),
		}
	}

	result := ListAllMyBucketsResult{
		Xmlns: s3Namespace,
		Owner: Owner{
			ID:          "owner-id",
			DisplayName: "owner",
		},
		Buckets: Buckets{
			Bucket: bucketInfos,
		},
	}

	writeXMLResponse(w, result)
}

// CreateBucket handles PUT /{bucket} - create a bucket.
func (s *Service) CreateBucket(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	if bucket == "" {
		writeS3Error(w, r, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

		return
	}

	err := s.storage.CreateBucket(r.Context(), bucket)
	if err != nil {
		var bucketErr *BucketError
		if errors.As(err, &bucketErr) {
			switch bucketErr.Code {
			case "BucketAlreadyOwnedByYou":
				writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusConflict)
			default:
				writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusBadRequest)
			}

			return
		}

		writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Location", "/"+bucket)
	w.WriteHeader(http.StatusOK)
}

// DeleteBucket handles DELETE /{bucket} - delete a bucket.
func (s *Service) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	if bucket == "" {
		writeS3Error(w, r, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

		return
	}

	err := s.storage.DeleteBucket(r.Context(), bucket)
	if err != nil {
		var bucketErr *BucketError
		if errors.As(err, &bucketErr) {
			switch bucketErr.Code {
			case "NoSuchBucket":
				writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusNotFound)
			case "BucketNotEmpty":
				writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusConflict)
			default:
				writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusBadRequest)
			}

			return
		}

		writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HeadBucket handles HEAD /{bucket} - check bucket existence.
func (s *Service) HeadBucket(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	if bucket == "" {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	exists, err := s.storage.BucketExists(r.Context(), bucket)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if !exists {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListObjects handles GET /{bucket} - list objects in a bucket.
func (s *Service) ListObjects(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	if bucket == "" {
		writeS3Error(w, r, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

		return
	}

	prefix := r.URL.Query().Get("prefix")
	delimiter := r.URL.Query().Get("delimiter")
	maxKeys := 1000

	if maxKeysStr := r.URL.Query().Get("max-keys"); maxKeysStr != "" {
		if mk, err := strconv.Atoi(maxKeysStr); err == nil && mk > 0 {
			maxKeys = mk
		}
	}

	objects, commonPrefixes, err := s.storage.ListObjects(r.Context(), bucket, prefix, delimiter, maxKeys)
	if err != nil {
		var bucketErr *BucketError
		if errors.As(err, &bucketErr) {
			writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusNotFound)

			return
		}

		writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	contents := make([]ObjectInfo, len(objects))
	for i := range objects {
		contents[i] = ObjectInfo{
			Key:          objects[i].Key,
			LastModified: objects[i].LastModified.Format(timeFormatISO),
			ETag:         objects[i].ETag,
			Size:         objects[i].Size,
			StorageClass: "STANDARD",
		}
	}

	prefixes := make([]CommonPrefix, len(commonPrefixes))
	for i, p := range commonPrefixes {
		prefixes[i] = CommonPrefix{Prefix: p}
	}

	result := ListBucketResult{
		Xmlns:          s3Namespace,
		Name:           bucket,
		Prefix:         prefix,
		KeyCount:       len(objects),
		MaxKeys:        maxKeys,
		IsTruncated:    false,
		Contents:       contents,
		CommonPrefixes: prefixes,
	}

	writeXMLResponse(w, result)
}

// PutObject handles PUT /{bucket}/{key...} - upload an object.
func (s *Service) PutObject(w http.ResponseWriter, r *http.Request) {
	if !checkPresignedURL(w, r) {
		return
	}

	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	if bucket == "" {
		writeS3Error(w, r, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

		return
	}

	if key == "" {
		writeS3Error(w, r, "InvalidArgument", "Invalid key", http.StatusBadRequest)

		return
	}

	metadata := make(map[string]string)
	if ct := r.Header.Get("Content-Type"); ct != "" {
		metadata["Content-Type"] = ct
	}

	// Extract x-amz-meta-* headers
	for name, values := range r.Header {
		if metaKey, found := strings.CutPrefix(strings.ToLower(name), "x-amz-meta-"); found {
			metadata[metaKey] = values[0]
		}
	}

	obj, err := s.storage.PutObject(r.Context(), bucket, key, r.Body, metadata)
	if err != nil {
		var bucketErr *BucketError
		if errors.As(err, &bucketErr) {
			writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusNotFound)

			return
		}

		writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("ETag", obj.ETag)

	if obj.VersionID != "" {
		w.Header().Set("x-amz-version-id", obj.VersionID)
	}

	w.WriteHeader(http.StatusOK)
}

// GetObject handles GET /{bucket}/{key...} - download an object.
func (s *Service) GetObject(w http.ResponseWriter, r *http.Request) {
	if !checkPresignedURL(w, r) {
		return
	}

	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	if bucket == "" {
		writeS3Error(w, r, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

		return
	}

	if key == "" {
		writeS3Error(w, r, "InvalidArgument", "Invalid key", http.StatusBadRequest)

		return
	}

	versionID := r.URL.Query().Get("versionId")

	var obj *Object

	var err error

	if versionID != "" {
		obj, err = s.storage.GetObjectVersion(r.Context(), bucket, key, versionID)
	} else {
		obj, err = s.storage.GetObject(r.Context(), bucket, key)
	}

	if err != nil {
		handleGetObjectError(w, r, err)

		return
	}

	writeObjectResponse(w, obj)
}

// handleGetObjectError handles errors from GetObject/GetObjectVersion.
func handleGetObjectError(w http.ResponseWriter, r *http.Request, err error) {
	var bucketErr *BucketError
	if errors.As(err, &bucketErr) {
		writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusNotFound)

		return
	}

	var objErr *ObjectError
	if errors.As(err, &objErr) {
		writeS3Error(w, r, objErr.Code, objErr.Message, http.StatusNotFound)

		return
	}

	writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)
}

// writeObjectResponse writes the object response with headers and body.
func writeObjectResponse(w http.ResponseWriter, obj *Object) {
	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(obj.Size, 10))
	w.Header().Set("ETag", obj.ETag)
	w.Header().Set("Last-Modified", obj.LastModified.UTC().Format(timeFormatHTTP))

	if obj.VersionID != "" {
		w.Header().Set("x-amz-version-id", obj.VersionID)
	}

	for k, v := range obj.Metadata {
		if k != "Content-Type" {
			w.Header().Set("x-amz-meta-"+k, v)
		}
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(obj.Body)
}

// DeleteObject handles DELETE /{bucket}/{key...} - delete an object.
func (s *Service) DeleteObject(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	if bucket == "" {
		writeS3Error(w, r, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

		return
	}

	if key == "" {
		writeS3Error(w, r, "InvalidArgument", "Invalid key", http.StatusBadRequest)

		return
	}

	versionID := r.URL.Query().Get("versionId")

	var deleteMarker *Object

	var err error

	if versionID != "" {
		deleteMarker, err = s.storage.DeleteObjectVersion(r.Context(), bucket, key, versionID)
	} else {
		deleteMarker, err = s.storage.DeleteObject(r.Context(), bucket, key)
	}

	if err != nil {
		var bucketErr *BucketError
		if errors.As(err, &bucketErr) {
			writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusNotFound)

			return
		}

		writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	// Return version info in headers if applicable
	if deleteMarker != nil {
		if deleteMarker.VersionID != "" {
			w.Header().Set("x-amz-version-id", deleteMarker.VersionID)
		}

		if deleteMarker.IsDeleteMarker {
			w.Header().Set("x-amz-delete-marker", "true")
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// HeadObject handles HEAD /{bucket}/{key...} - get object metadata.
func (s *Service) HeadObject(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	if bucket == "" || key == "" {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	obj, err := s.storage.HeadObject(r.Context(), bucket, key)
	if err != nil {
		var bucketErr *BucketError
		if errors.As(err, &bucketErr) {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		var objErr *ObjectError
		if errors.As(err, &objErr) {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(obj.Size, 10))
	w.Header().Set("ETag", obj.ETag)
	w.Header().Set("Last-Modified", obj.LastModified.UTC().Format(timeFormatHTTP))

	// Set metadata headers
	for k, v := range obj.Metadata {
		if k != "Content-Type" {
			w.Header().Set("x-amz-meta-"+k, v)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// PutBucketVersioning handles PUT /{bucket}?versioning - set bucket versioning.
func (s *Service) PutBucketVersioning(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	if bucket == "" {
		writeS3Error(w, r, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

		return
	}

	var config VersioningConfiguration
	if err := xml.NewDecoder(r.Body).Decode(&config); err != nil {
		writeS3Error(w, r, "MalformedXML", "The XML you provided was not well-formed", http.StatusBadRequest)

		return
	}

	err := s.storage.PutBucketVersioning(r.Context(), bucket, config.Status)
	if err != nil {
		var bucketErr *BucketError
		if errors.As(err, &bucketErr) {
			writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusNotFound)

			return
		}

		writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetBucketVersioning handles GET /{bucket}?versioning - get bucket versioning.
func (s *Service) GetBucketVersioning(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	if bucket == "" {
		writeS3Error(w, r, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

		return
	}

	status, err := s.storage.GetBucketVersioning(r.Context(), bucket)
	if err != nil {
		var bucketErr *BucketError
		if errors.As(err, &bucketErr) {
			writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusNotFound)

			return
		}

		writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	result := VersioningConfiguration{
		Xmlns:  s3Namespace,
		Status: status,
	}

	writeXMLResponse(w, result)
}

// ListObjectVersions handles GET /{bucket}?versions - list object versions.
func (s *Service) ListObjectVersions(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	if bucket == "" {
		writeS3Error(w, r, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

		return
	}

	prefix := r.URL.Query().Get("prefix")
	delimiter := r.URL.Query().Get("delimiter")
	maxKeys := parseMaxKeys(r.URL.Query().Get("max-keys"))

	objects, commonPrefixes, err := s.storage.ListObjectVersions(r.Context(), bucket, prefix, delimiter, maxKeys)
	if err != nil {
		handleListVersionsError(w, r, err)

		return
	}

	versions, deleteMarkers := separateVersionsAndDeleteMarkers(objects)
	prefixes := toCommonPrefixes(commonPrefixes)

	result := ListVersionsResult{
		Xmlns:          s3Namespace,
		Name:           bucket,
		Prefix:         prefix,
		MaxKeys:        maxKeys,
		IsTruncated:    false,
		Versions:       versions,
		DeleteMarkers:  deleteMarkers,
		CommonPrefixes: prefixes,
	}

	writeXMLResponse(w, result)
}

// parseMaxKeys parses max-keys query parameter with default of 1000.
func parseMaxKeys(maxKeysStr string) int {
	if maxKeysStr == "" {
		return 1000
	}

	if mk, err := strconv.Atoi(maxKeysStr); err == nil && mk > 0 {
		return mk
	}

	return 1000
}

// handleListVersionsError handles errors from ListObjectVersions.
func handleListVersionsError(w http.ResponseWriter, r *http.Request, err error) {
	var bucketErr *BucketError
	if errors.As(err, &bucketErr) {
		writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusNotFound)

		return
	}

	writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)
}

// separateVersionsAndDeleteMarkers separates objects into versions and delete markers.
func separateVersionsAndDeleteMarkers(objects []Object) ([]ObjectVersionInfo, []DeleteMarkerInfo) {
	versions := make([]ObjectVersionInfo, 0, len(objects))
	deleteMarkers := make([]DeleteMarkerInfo, 0)

	for i := range objects {
		obj := &objects[i]
		isLatest := i == 0 || objects[i-1].Key != obj.Key

		if obj.IsDeleteMarker {
			deleteMarkers = append(deleteMarkers, toDeleteMarkerInfo(obj, isLatest))
		} else {
			versions = append(versions, toObjectVersionInfo(obj, isLatest))
		}
	}

	return versions, deleteMarkers
}

// toObjectVersionInfo converts an Object to ObjectVersionInfo.
func toObjectVersionInfo(obj *Object, isLatest bool) ObjectVersionInfo {
	return ObjectVersionInfo{
		Key:          obj.Key,
		VersionID:    obj.VersionID,
		IsLatest:     isLatest,
		LastModified: obj.LastModified.Format(timeFormatISO),
		ETag:         obj.ETag,
		Size:         obj.Size,
		StorageClass: "STANDARD",
		Owner:        Owner{ID: "owner-id", DisplayName: "owner"},
	}
}

// toDeleteMarkerInfo converts an Object to DeleteMarkerInfo.
func toDeleteMarkerInfo(obj *Object, isLatest bool) DeleteMarkerInfo {
	return DeleteMarkerInfo{
		Key:          obj.Key,
		VersionID:    obj.VersionID,
		IsLatest:     isLatest,
		LastModified: obj.LastModified.Format(timeFormatISO),
		Owner:        Owner{ID: "owner-id", DisplayName: "owner"},
	}
}

// toCommonPrefixes converts string slice to CommonPrefix slice.
func toCommonPrefixes(prefixes []string) []CommonPrefix {
	result := make([]CommonPrefix, len(prefixes))
	for i, p := range prefixes {
		result[i] = CommonPrefix{Prefix: p}
	}

	return result
}

// handleBucketPut routes PUT /{bucket} requests based on query parameters.
func (s *Service) handleBucketPut(w http.ResponseWriter, r *http.Request) {
	if _, ok := r.URL.Query()["versioning"]; ok {
		s.PutBucketVersioning(w, r)

		return
	}

	s.CreateBucket(w, r)
}

// handleBucketGet routes GET /{bucket} requests based on query parameters.
func (s *Service) handleBucketGet(w http.ResponseWriter, r *http.Request) {
	if _, ok := r.URL.Query()["versioning"]; ok {
		s.GetBucketVersioning(w, r)

		return
	}

	if _, ok := r.URL.Query()["versions"]; ok {
		s.ListObjectVersions(w, r)

		return
	}

	s.ListObjects(w, r)
}

// writeXMLResponse writes an XML response with HTTP 200 OK status.
func writeXMLResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)

	_, _ = io.WriteString(w, xmlHeader)
	_ = xml.NewEncoder(w).Encode(v)
}

// writeS3Error writes an S3 error response.
func writeS3Error(w http.ResponseWriter, _ *http.Request, code, message string, status int) {
	errResp := ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: uuid.New().String(),
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)

	_, _ = io.WriteString(w, xmlHeader)
	_ = xml.NewEncoder(w).Encode(errResp)
}

// isPresignedRequest checks if the request is a presigned URL request.
func isPresignedRequest(r *http.Request) bool {
	return r.URL.Query().Get("X-Amz-Signature") != ""
}

// checkPresignedURL validates presigned URL if present and writes error response if invalid.
// Returns true if the request should continue processing, false if an error was written.
func checkPresignedURL(w http.ResponseWriter, r *http.Request) bool {
	if !isPresignedRequest(r) {
		return true
	}

	if err := validatePresignedURL(r); err != nil {
		var presignErr *PresignedURLError
		if errors.As(err, &presignErr) {
			writeS3Error(w, r, presignErr.Code, presignErr.Message, http.StatusForbidden)

			return false
		}

		writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)

		return false
	}

	return true
}

// validatePresignedURL validates the presigned URL expiration.
// Returns nil if the URL is valid, or an error if expired.
func validatePresignedURL(r *http.Request) error {
	// Get the date when the URL was signed
	amzDate := r.URL.Query().Get("X-Amz-Date")
	if amzDate == "" {
		return &PresignedURLError{Code: "AuthorizationQueryParametersError", Message: "X-Amz-Date must be in the ISO8601 Long Format"}
	}

	// Get the expiration in seconds
	expiresStr := r.URL.Query().Get("X-Amz-Expires")
	if expiresStr == "" {
		return &PresignedURLError{Code: "AuthorizationQueryParametersError", Message: "X-Amz-Expires must be provided"}
	}

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return &PresignedURLError{Code: "AuthorizationQueryParametersError", Message: "X-Amz-Expires must be a number"}
	}

	// AWS allows max 7 days (604800 seconds) for presigned URLs
	const maxExpires = 604800
	if expires > maxExpires {
		return &PresignedURLError{Code: "AuthorizationQueryParametersError", Message: "X-Amz-Expires must be less than 604800 seconds"}
	}

	// Parse the signing date (format: 20060102T150405Z)
	signTime, err := time.Parse("20060102T150405Z", amzDate)
	if err != nil {
		return &PresignedURLError{Code: "AuthorizationQueryParametersError", Message: "Invalid X-Amz-Date format"}
	}

	// Check if the URL has expired
	expirationTime := signTime.Add(time.Duration(expires) * time.Second)
	if time.Now().After(expirationTime) {
		return &PresignedURLError{Code: "AccessDenied", Message: "Request has expired"}
	}

	return nil
}

// PresignedURLError represents a presigned URL validation error.
type PresignedURLError struct {
	Code    string
	Message string
}

func (e *PresignedURLError) Error() string {
	return e.Code + ": " + e.Message
}

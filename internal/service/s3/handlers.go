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

	writeXMLResponse(w, http.StatusOK, result)
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
	for i, obj := range objects {
		contents[i] = ObjectInfo{
			Key:          obj.Key,
			LastModified: obj.LastModified.Format(timeFormatISO),
			ETag:         obj.ETag,
			Size:         obj.Size,
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

	writeXMLResponse(w, http.StatusOK, result)
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

	obj, err := s.storage.GetObject(r.Context(), bucket, key)
	if err != nil {
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

	err := s.storage.DeleteObject(r.Context(), bucket, key)
	if err != nil {
		var bucketErr *BucketError
		if errors.As(err, &bucketErr) {
			writeS3Error(w, r, bucketErr.Code, bucketErr.Message, http.StatusNotFound)

			return
		}

		writeS3Error(w, r, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
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

// writeXMLResponse writes an XML response.
func writeXMLResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)

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

package s3

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

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

	bucketInfos := make([]ListBucketsResponseBucket, len(buckets))
	for i, b := range buckets {
		bucketInfos[i] = ListBucketsResponseBucket{
			Name:         b.Name,
			CreationDate: b.CreationDate.Format(timeFormatISO),
		}
	}

	result := ListBucketsResponse{
		Xmlns: s3Namespace,
		Owner: ListBucketsResponseOwner{
			ID:          "owner-id",
			DisplayName: "owner",
		},
		Buckets: ListBucketsResponseBuckets{
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

	contents := make([]ListObjectsResponseContent, len(objects))
	for i, obj := range objects {
		contents[i] = ListObjectsResponseContent{
			Key:          obj.Key,
			LastModified: obj.LastModified.Format(timeFormatISO),
			ETag:         obj.ETag,
			Size:         obj.Size,
			StorageClass: "STANDARD",
		}
	}

	prefixes := make([]ListObjectsResponsePrefix, len(commonPrefixes))
	for i, p := range commonPrefixes {
		prefixes[i] = ListObjectsResponsePrefix{Prefix: p}
	}

	result := ListObjectsResponse{
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

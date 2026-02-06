//go:build integration

package integration

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func newS3Client(t *testing.T) *s3.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
		o.UsePathStyle = true
	})
}

func TestS3_CreateAndDeleteBucket(t *testing.T) {
	client := newS3Client(t)
	ctx := t.Context()
	bucketName := "test-create-delete-bucket"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	// Head bucket to verify it exists
	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to head bucket: %v", err)
	}

	// Delete bucket
	_, err = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to delete bucket: %v", err)
	}
}

func TestS3_ListBuckets(t *testing.T) {
	client := newS3Client(t)
	ctx := t.Context()

	// Create a bucket first
	bucketName := "test-list-buckets"
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// List buckets
	result, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		t.Fatalf("failed to list buckets: %v", err)
	}

	found := false
	for _, b := range result.Buckets {
		if *b.Name == bucketName {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("bucket %s not found in list", bucketName)
	}
}

func TestS3_PutAndGetObject(t *testing.T) {
	client := newS3Client(t)
	ctx := t.Context()
	bucketName := "test-put-get-object"
	key := "test-key.txt"
	content := "Hello, awsim!"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// Put object
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte(content)),
	})
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	// Get object
	result, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("failed to get object: %v", err)
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != content {
		t.Errorf("expected content %q, got %q", content, string(body))
	}
}

func TestS3_HeadObject(t *testing.T) {
	client := newS3Client(t)
	ctx := t.Context()
	bucketName := "test-head-object"
	key := "test-key.txt"
	content := "Hello, awsim!"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// Put object
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte(content)),
	})
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	// Head object
	result, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("failed to head object: %v", err)
	}

	if *result.ContentLength != int64(len(content)) {
		t.Errorf("expected content length %d, got %d", len(content), *result.ContentLength)
	}
}

func TestS3_DeleteObject(t *testing.T) {
	client := newS3Client(t)
	ctx := t.Context()
	bucketName := "test-delete-object"
	key := "test-key.txt"
	content := "Hello, awsim!"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// Put object
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte(content)),
	})
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	// Delete object
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("failed to delete object: %v", err)
	}

	// Verify object is deleted
	_, err = client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err == nil {
		t.Error("expected error when getting deleted object, got nil")
	}
}

func TestS3_ListObjects(t *testing.T) {
	client := newS3Client(t)
	ctx := t.Context()
	bucketName := "test-list-objects"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		// Clean up objects
		for _, key := range []string{"file1.txt", "file2.txt", "dir/file3.txt"} {
			_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			})
		}
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// Put multiple objects
	keys := []string{"file1.txt", "file2.txt", "dir/file3.txt"}
	for _, key := range keys {
		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
			Body:   bytes.NewReader([]byte("content")),
		})
		if err != nil {
			t.Fatalf("failed to put object %s: %v", key, err)
		}
	}

	// List all objects
	result, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to list objects: %v", err)
	}

	if len(result.Contents) != 3 {
		t.Errorf("expected 3 objects, got %d", len(result.Contents))
	}

	// List with prefix
	result, err = client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String("dir/"),
	})
	if err != nil {
		t.Fatalf("failed to list objects with prefix: %v", err)
	}

	if len(result.Contents) != 1 {
		t.Errorf("expected 1 object with prefix 'dir/', got %d", len(result.Contents))
	}
}

func newS3PresignClient(t *testing.T) *s3.PresignClient {
	t.Helper()

	return s3.NewPresignClient(newS3Client(t))
}

func TestS3_PresignedURL_GetObject(t *testing.T) {
	client := newS3Client(t)
	presignClient := newS3PresignClient(t)
	ctx := t.Context()
	bucketName := "test-presigned-get-object"
	key := "test-presigned-key.txt"
	content := "Hello, presigned URL!"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// Put object using regular client
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte(content)),
	})
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	// Generate presigned URL for GetObject
	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	if err != nil {
		t.Fatalf("failed to presign GetObject: %v", err)
	}

	// Use presigned URL to get the object
	resp, err := http.Get(presignedReq.URL)
	if err != nil {
		t.Fatalf("failed to GET presigned URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != content {
		t.Errorf("expected content %q, got %q", content, string(body))
	}
}

func TestS3_PresignedURL_PutObject(t *testing.T) {
	client := newS3Client(t)
	presignClient := newS3PresignClient(t)
	ctx := t.Context()
	bucketName := "test-presigned-put-object"
	key := "test-presigned-put-key.txt"
	content := "Hello, presigned PUT!"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// Generate presigned URL for PutObject
	presignedReq, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	if err != nil {
		t.Fatalf("failed to presign PutObject: %v", err)
	}

	// Use presigned URL to put the object
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, presignedReq.URL, bytes.NewReader([]byte(content)))
	if err != nil {
		t.Fatalf("failed to create PUT request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to PUT presigned URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify object was uploaded using regular client
	result, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("failed to get object: %v", err)
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != content {
		t.Errorf("expected content %q, got %q", content, string(body))
	}
}

func TestS3_PresignedURL_Expired(t *testing.T) {
	client := newS3Client(t)
	ctx := t.Context()
	bucketName := "test-presigned-expired"
	key := "test-expired-key.txt"
	content := "Hello, expired!"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// Put object using regular client
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte(content)),
	})
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	// Generate presigned URL with 1 second expiration
	presignClient := newS3PresignClient(t)
	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 1 * time.Second
	})
	if err != nil {
		t.Fatalf("failed to presign GetObject: %v", err)
	}

	// Wait for URL to expire
	time.Sleep(2 * time.Second)

	// Use expired presigned URL
	resp, err := http.Get(presignedReq.URL)
	if err != nil {
		t.Fatalf("failed to GET expired presigned URL: %v", err)
	}
	defer resp.Body.Close()

	// Should return 403 Forbidden for expired URL
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403 for expired URL, got %d", resp.StatusCode)
	}
}

func TestS3_Versioning_PutAndGetBucketVersioning(t *testing.T) {
	client := newS3Client(t)
	ctx := t.Context()
	bucketName := "test-versioning-config"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// Get versioning status (should be empty initially)
	result, err := client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to get bucket versioning: %v", err)
	}

	if result.Status != "" {
		t.Errorf("expected empty versioning status, got %v", result.Status)
	}

	// Enable versioning
	_, err = client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &s3types.VersioningConfiguration{
			Status: s3types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		t.Fatalf("failed to put bucket versioning: %v", err)
	}

	// Verify versioning is enabled
	result, err = client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to get bucket versioning: %v", err)
	}

	if result.Status != s3types.BucketVersioningStatusEnabled {
		t.Errorf("expected versioning status Enabled, got %v", result.Status)
	}

	// Suspend versioning
	_, err = client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &s3types.VersioningConfiguration{
			Status: s3types.BucketVersioningStatusSuspended,
		},
	})
	if err != nil {
		t.Fatalf("failed to suspend bucket versioning: %v", err)
	}

	// Verify versioning is suspended
	result, err = client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to get bucket versioning: %v", err)
	}

	if result.Status != s3types.BucketVersioningStatusSuspended {
		t.Errorf("expected versioning status Suspended, got %v", result.Status)
	}
}

func TestS3_Versioning_PutObjectWithVersioning(t *testing.T) {
	client := newS3Client(t)
	ctx := t.Context()
	bucketName := "test-versioning-put-object"
	key := "test-versioned-key.txt"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		// List and delete all versions
		versions, _ := client.ListObjectVersions(ctx, &s3.ListObjectVersionsInput{
			Bucket: aws.String(bucketName),
		})
		if versions != nil {
			for _, v := range versions.Versions {
				_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket:    aws.String(bucketName),
					Key:       v.Key,
					VersionId: v.VersionId,
				})
			}
			for _, dm := range versions.DeleteMarkers {
				_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket:    aws.String(bucketName),
					Key:       dm.Key,
					VersionId: dm.VersionId,
				})
			}
		}
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// Enable versioning
	_, err = client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &s3types.VersioningConfiguration{
			Status: s3types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		t.Fatalf("failed to enable versioning: %v", err)
	}

	// Put first version
	putResult1, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte("version 1")),
	})
	if err != nil {
		t.Fatalf("failed to put object v1: %v", err)
	}

	if putResult1.VersionId == nil || *putResult1.VersionId == "" {
		t.Error("expected version ID for first put, got empty")
	}

	// Put second version
	putResult2, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte("version 2")),
	})
	if err != nil {
		t.Fatalf("failed to put object v2: %v", err)
	}

	if putResult2.VersionId == nil || *putResult2.VersionId == "" {
		t.Error("expected version ID for second put, got empty")
	}

	if *putResult1.VersionId == *putResult2.VersionId {
		t.Error("expected different version IDs for different puts")
	}

	// Get latest version (should be v2)
	getResult, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("failed to get object: %v", err)
	}
	defer getResult.Body.Close()

	body, _ := io.ReadAll(getResult.Body)
	if string(body) != "version 2" {
		t.Errorf("expected 'version 2', got %q", string(body))
	}

	// Get first version by version ID
	getResult1, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket:    aws.String(bucketName),
		Key:       aws.String(key),
		VersionId: putResult1.VersionId,
	})
	if err != nil {
		t.Fatalf("failed to get object v1 by version ID: %v", err)
	}
	defer getResult1.Body.Close()

	body1, _ := io.ReadAll(getResult1.Body)
	if string(body1) != "version 1" {
		t.Errorf("expected 'version 1', got %q", string(body1))
	}
}

func TestS3_Versioning_DeleteObjectCreatesDeleteMarker(t *testing.T) {
	client := newS3Client(t)
	ctx := t.Context()
	bucketName := "test-versioning-delete-marker"
	key := "test-delete-marker-key.txt"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		// List and delete all versions
		versions, _ := client.ListObjectVersions(ctx, &s3.ListObjectVersionsInput{
			Bucket: aws.String(bucketName),
		})
		if versions != nil {
			for _, v := range versions.Versions {
				_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket:    aws.String(bucketName),
					Key:       v.Key,
					VersionId: v.VersionId,
				})
			}
			for _, dm := range versions.DeleteMarkers {
				_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket:    aws.String(bucketName),
					Key:       dm.Key,
					VersionId: dm.VersionId,
				})
			}
		}
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// Enable versioning
	_, err = client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &s3types.VersioningConfiguration{
			Status: s3types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		t.Fatalf("failed to enable versioning: %v", err)
	}

	// Put object
	putResult, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte("content")),
	})
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	// Delete object (should create delete marker)
	deleteResult, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("failed to delete object: %v", err)
	}

	if deleteResult.DeleteMarker == nil || !*deleteResult.DeleteMarker {
		t.Error("expected delete marker to be true")
	}

	if deleteResult.VersionId == nil || *deleteResult.VersionId == "" {
		t.Error("expected version ID for delete marker")
	}

	// Try to get object (should fail with 404)
	_, err = client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err == nil {
		t.Error("expected error when getting deleted object, got nil")
	}

	// Get object by original version ID (should succeed)
	getResult, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket:    aws.String(bucketName),
		Key:       aws.String(key),
		VersionId: putResult.VersionId,
	})
	if err != nil {
		t.Fatalf("failed to get object by version ID: %v", err)
	}
	defer getResult.Body.Close()

	body, _ := io.ReadAll(getResult.Body)
	if string(body) != "content" {
		t.Errorf("expected 'content', got %q", string(body))
	}
}

func TestS3_Versioning_ListObjectVersions(t *testing.T) {
	client := newS3Client(t)
	ctx := t.Context()
	bucketName := "test-versioning-list-versions"
	key := "test-list-versions-key.txt"

	// Create bucket
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	t.Cleanup(func() {
		// List and delete all versions
		versions, _ := client.ListObjectVersions(ctx, &s3.ListObjectVersionsInput{
			Bucket: aws.String(bucketName),
		})
		if versions != nil {
			for _, v := range versions.Versions {
				_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket:    aws.String(bucketName),
					Key:       v.Key,
					VersionId: v.VersionId,
				})
			}
			for _, dm := range versions.DeleteMarkers {
				_, _ = client.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket:    aws.String(bucketName),
					Key:       dm.Key,
					VersionId: dm.VersionId,
				})
			}
		}
		_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	})

	// Enable versioning
	_, err = client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &s3types.VersioningConfiguration{
			Status: s3types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		t.Fatalf("failed to enable versioning: %v", err)
	}

	// Put three versions
	for i := 1; i <= 3; i++ {
		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
			Body:   bytes.NewReader([]byte("version " + string(rune('0'+i)))),
		})
		if err != nil {
			t.Fatalf("failed to put object v%d: %v", i, err)
		}
	}

	// Delete object (creates delete marker)
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("failed to delete object: %v", err)
	}

	// List object versions
	listResult, err := client.ListObjectVersions(ctx, &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("failed to list object versions: %v", err)
	}

	// Should have 3 versions
	if len(listResult.Versions) != 3 {
		t.Errorf("expected 3 versions, got %d", len(listResult.Versions))
	}

	// Should have 1 delete marker
	if len(listResult.DeleteMarkers) != 1 {
		t.Errorf("expected 1 delete marker, got %d", len(listResult.DeleteMarkers))
	}

	// Check that only one version (the latest before delete) or delete marker is marked as latest
	latestCount := 0
	for _, v := range listResult.Versions {
		if v.IsLatest != nil && *v.IsLatest {
			latestCount++
		}
	}
	for _, dm := range listResult.DeleteMarkers {
		if dm.IsLatest != nil && *dm.IsLatest {
			latestCount++
		}
	}

	if latestCount != 1 {
		t.Errorf("expected exactly 1 latest version/marker, got %d", latestCount)
	}
}

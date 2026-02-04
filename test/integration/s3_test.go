//go:build integration

package integration

import (
	"bytes"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

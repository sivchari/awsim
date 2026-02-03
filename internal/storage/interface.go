// Package storage provides storage backends for AWS service data.
package storage

import (
	"context"
	"io"
	"time"
)

// Object represents a stored object with metadata.
type Object struct {
	Key          string
	Body         []byte
	ContentType  string
	ETag         string
	Size         int64
	LastModified time.Time
	Metadata     map[string]string
}

// Bucket represents an S3 bucket.
type Bucket struct {
	Name         string
	CreationDate time.Time
}

// S3Storage is the interface for S3 storage operations.
type S3Storage interface {
	// Bucket operations
	CreateBucket(ctx context.Context, name string) error
	DeleteBucket(ctx context.Context, name string) error
	ListBuckets(ctx context.Context) ([]Bucket, error)
	BucketExists(ctx context.Context, name string) (bool, error)

	// Object operations
	PutObject(ctx context.Context, bucket, key string, body io.Reader, metadata map[string]string) (*Object, error)
	GetObject(ctx context.Context, bucket, key string) (*Object, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	ListObjects(ctx context.Context, bucket, prefix string, maxKeys int) ([]Object, error)
	HeadObject(ctx context.Context, bucket, key string) (*Object, error)
}

// Message represents an SQS message.
type Message struct {
	MessageID     string
	ReceiptHandle string
	Body          string
	Attributes    map[string]string
	SentTimestamp time.Time
}

// Queue represents an SQS queue.
type Queue struct {
	Name string
	URL  string
	Arn  string
}

// SQSStorage is the interface for SQS storage operations.
type SQSStorage interface {
	// Queue operations
	CreateQueue(ctx context.Context, name string, attributes map[string]string) (*Queue, error)
	DeleteQueue(ctx context.Context, queueURL string) error
	ListQueues(ctx context.Context, prefix string) ([]Queue, error)
	GetQueueURL(ctx context.Context, name string) (string, error)

	// Message operations
	SendMessage(ctx context.Context, queueURL, body string, attributes map[string]string) (*Message, error)
	ReceiveMessages(ctx context.Context, queueURL string, maxMessages int, waitTime time.Duration) ([]Message, error)
	DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error
}

// Item represents a DynamoDB item.
type Item map[string]any

// Table represents a DynamoDB table.
type Table struct {
	Name   string
	Status string
}

// DynamoDBStorage is the interface for DynamoDB storage operations.
type DynamoDBStorage interface {
	// Table operations
	CreateTable(ctx context.Context, name string, keySchema map[string]string) (*Table, error)
	DeleteTable(ctx context.Context, name string) error
	ListTables(ctx context.Context) ([]string, error)
	DescribeTable(ctx context.Context, name string) (*Table, error)

	// Item operations
	PutItem(ctx context.Context, table string, item Item) error
	GetItem(ctx context.Context, table string, key Item) (Item, error)
	DeleteItem(ctx context.Context, table string, key Item) error
	Query(ctx context.Context, table string, keyCondition string, values map[string]any) ([]Item, error)
	Scan(ctx context.Context, table string) ([]Item, error)
}

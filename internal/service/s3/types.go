// Package s3 provides S3 service emulation for awsim.
package s3

import (
	"encoding/xml"
	"time"
)

// Domain Models

// Bucket represents an S3 bucket.
type Bucket struct {
	Name         string
	CreationDate time.Time
}

// Object represents an S3 object.
type Object struct {
	Key          string
	Body         []byte
	ETag         string
	Size         int64
	LastModified time.Time
	ContentType  string
	Metadata     map[string]string
}

// Request Types

// ListObjectsRequest contains parameters for ListObjects operation.
type ListObjectsRequest struct {
	Bucket            string
	Prefix            string
	Delimiter         string
	MaxKeys           int
	ContinuationToken string
	StartAfter        string
}

// CreateBucketRequest contains parameters for CreateBucket operation.
type CreateBucketRequest struct {
	XMLName            xml.Name `xml:"CreateBucketConfiguration"`
	LocationConstraint string   `xml:"LocationConstraint"`
}

// Response Types

// ListBucketsResponse is the response for ListBuckets operation.
// XML element name follows AWS S3 API specification.
type ListBucketsResponse struct {
	XMLName xml.Name                   `xml:"ListAllMyBucketsResult"`
	Xmlns   string                     `xml:"xmlns,attr"`
	Owner   ListBucketsResponseOwner   `xml:"Owner"`
	Buckets ListBucketsResponseBuckets `xml:"Buckets"`
}

// ListBucketsResponseOwner represents the bucket owner in ListBuckets response.
type ListBucketsResponseOwner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

// ListBucketsResponseBuckets is a list of buckets in ListBuckets response.
type ListBucketsResponseBuckets struct {
	Bucket []ListBucketsResponseBucket `xml:"Bucket"`
}

// ListBucketsResponseBucket represents a bucket entry in ListBuckets response.
type ListBucketsResponseBucket struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

// ListObjectsResponse is the response for ListObjects operation.
// XML element name follows AWS S3 API specification (ListBucketResult).
type ListObjectsResponse struct {
	XMLName               xml.Name                     `xml:"ListBucketResult"`
	Xmlns                 string                       `xml:"xmlns,attr"`
	Name                  string                       `xml:"Name"`
	Prefix                string                       `xml:"Prefix"`
	KeyCount              int                          `xml:"KeyCount"`
	MaxKeys               int                          `xml:"MaxKeys"`
	IsTruncated           bool                         `xml:"IsTruncated"`
	Contents              []ListObjectsResponseContent `xml:"Contents"`
	ContinuationToken     string                       `xml:"ContinuationToken,omitempty"`
	NextContinuationToken string                       `xml:"NextContinuationToken,omitempty"`
	StartAfter            string                       `xml:"StartAfter,omitempty"`
	CommonPrefixes        []ListObjectsResponsePrefix  `xml:"CommonPrefixes,omitempty"`
}

// ListObjectsResponseContent represents an object entry in ListObjects response.
type ListObjectsResponseContent struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

// ListObjectsResponsePrefix represents a common prefix in ListObjects response.
type ListObjectsResponsePrefix struct {
	Prefix string `xml:"Prefix"`
}

// CopyObjectResponse is the response for CopyObject operation.
type CopyObjectResponse struct {
	XMLName      xml.Name `xml:"CopyObjectResult"`
	ETag         string   `xml:"ETag"`
	LastModified string   `xml:"LastModified"`
}

// Error Types

// ErrorResponse represents an S3 error response.
type ErrorResponse struct {
	XMLName    xml.Name `xml:"Error"`
	Code       string   `xml:"Code"`
	Message    string   `xml:"Message"`
	Resource   string   `xml:"Resource,omitempty"`
	RequestID  string   `xml:"RequestId"`
	BucketName string   `xml:"BucketName,omitempty"`
	Key        string   `xml:"Key,omitempty"`
}

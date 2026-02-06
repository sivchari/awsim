// Package s3 provides S3 service emulation for awsim.
package s3

import (
	"encoding/xml"
	"time"
)

// Bucket represents an S3 bucket.
type Bucket struct {
	Name         string
	CreationDate time.Time
}

// Object represents an S3 object.
type Object struct {
	Key            string
	Body           []byte
	ETag           string
	Size           int64
	LastModified   time.Time
	ContentType    string
	Metadata       map[string]string
	VersionID      string
	IsDeleteMarker bool
}

// XML Response Types

// ListAllMyBucketsResult is the response for ListBuckets.
type ListAllMyBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Xmlns   string   `xml:"xmlns,attr"`
	Owner   Owner    `xml:"Owner"`
	Buckets Buckets  `xml:"Buckets"`
}

// Owner represents the bucket owner.
type Owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

// Buckets is a list of buckets.
type Buckets struct {
	Bucket []BucketInfo `xml:"Bucket"`
}

// BucketInfo represents bucket information in XML response.
type BucketInfo struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

// ListBucketResult is the response for ListObjectsV2.
type ListBucketResult struct {
	XMLName               xml.Name       `xml:"ListBucketResult"`
	Xmlns                 string         `xml:"xmlns,attr"`
	Name                  string         `xml:"Name"`
	Prefix                string         `xml:"Prefix"`
	KeyCount              int            `xml:"KeyCount"`
	MaxKeys               int            `xml:"MaxKeys"`
	IsTruncated           bool           `xml:"IsTruncated"`
	Contents              []ObjectInfo   `xml:"Contents"`
	ContinuationToken     string         `xml:"ContinuationToken,omitempty"`
	NextContinuationToken string         `xml:"NextContinuationToken,omitempty"`
	StartAfter            string         `xml:"StartAfter,omitempty"`
	CommonPrefixes        []CommonPrefix `xml:"CommonPrefixes,omitempty"`
}

// ObjectInfo represents object information in XML response.
type ObjectInfo struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

// CommonPrefix represents a common prefix in ListObjects response.
type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

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

// Versioning Types

// VersioningConfiguration represents bucket versioning configuration.
type VersioningConfiguration struct {
	XMLName xml.Name `xml:"VersioningConfiguration"`
	Xmlns   string   `xml:"xmlns,attr,omitempty"`
	Status  string   `xml:"Status,omitempty"`
}

// ListVersionsResult is the response for ListObjectVersions.
type ListVersionsResult struct {
	XMLName             xml.Name            `xml:"ListVersionsResult"`
	Xmlns               string              `xml:"xmlns,attr"`
	Name                string              `xml:"Name"`
	Prefix              string              `xml:"Prefix,omitempty"`
	KeyMarker           string              `xml:"KeyMarker,omitempty"`
	VersionIDMarker     string              `xml:"VersionIdMarker,omitempty"`
	NextKeyMarker       string              `xml:"NextKeyMarker,omitempty"`
	NextVersionIDMarker string              `xml:"NextVersionIdMarker,omitempty"`
	MaxKeys             int                 `xml:"MaxKeys"`
	IsTruncated         bool                `xml:"IsTruncated"`
	Versions            []ObjectVersionInfo `xml:"Version,omitempty"`
	DeleteMarkers       []DeleteMarkerInfo  `xml:"DeleteMarker,omitempty"`
	CommonPrefixes      []CommonPrefix      `xml:"CommonPrefixes,omitempty"`
}

// ObjectVersionInfo represents an object version in ListVersionsResult.
type ObjectVersionInfo struct {
	Key          string `xml:"Key"`
	VersionID    string `xml:"VersionId"`
	IsLatest     bool   `xml:"IsLatest"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
	Owner        Owner  `xml:"Owner"`
}

// DeleteMarkerInfo represents a delete marker in ListVersionsResult.
type DeleteMarkerInfo struct {
	Key          string `xml:"Key"`
	VersionID    string `xml:"VersionId"`
	IsLatest     bool   `xml:"IsLatest"`
	LastModified string `xml:"LastModified"`
	Owner        Owner  `xml:"Owner"`
}

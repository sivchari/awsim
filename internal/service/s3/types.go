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

// Multipart Upload Types

// MultipartUpload represents an in-progress multipart upload.
type MultipartUpload struct {
	Bucket    string
	Key       string
	UploadID  string
	Initiated time.Time
	Parts     map[int]*Part // partNumber -> Part
}

// Part represents a part in a multipart upload.
type Part struct {
	PartNumber   int
	ETag         string
	Size         int64
	LastModified time.Time
	Body         []byte
}

// InitiateMultipartUploadResult is the response for CreateMultipartUpload.
type InitiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Xmlns    string   `xml:"xmlns,attr"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	UploadID string   `xml:"UploadId"`
}

// CompleteMultipartUploadRequest is the request body for CompleteMultipartUpload.
type CompleteMultipartUploadRequest struct {
	XMLName xml.Name      `xml:"CompleteMultipartUpload"`
	Parts   []PartRequest `xml:"Part"`
}

// PartRequest represents a part in the complete request.
type PartRequest struct {
	PartNumber int    `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

// CompleteMultipartUploadResult is the response for CompleteMultipartUpload.
type CompleteMultipartUploadResult struct {
	XMLName  xml.Name `xml:"CompleteMultipartUploadResult"`
	Xmlns    string   `xml:"xmlns,attr"`
	Location string   `xml:"Location"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	ETag     string   `xml:"ETag"`
}

// ListMultipartUploadsResult is the response for ListMultipartUploads.
type ListMultipartUploadsResult struct {
	XMLName            xml.Name       `xml:"ListMultipartUploadsResult"`
	Xmlns              string         `xml:"xmlns,attr"`
	Bucket             string         `xml:"Bucket"`
	KeyMarker          string         `xml:"KeyMarker,omitempty"`
	UploadIDMarker     string         `xml:"UploadIdMarker,omitempty"`
	NextKeyMarker      string         `xml:"NextKeyMarker,omitempty"`
	NextUploadIDMarker string         `xml:"NextUploadIdMarker,omitempty"`
	MaxUploads         int            `xml:"MaxUploads"`
	IsTruncated        bool           `xml:"IsTruncated"`
	Uploads            []UploadInfo   `xml:"Upload"`
	CommonPrefixes     []CommonPrefix `xml:"CommonPrefixes,omitempty"`
}

// UploadInfo represents a multipart upload in the list response.
type UploadInfo struct {
	Key       string `xml:"Key"`
	UploadID  string `xml:"UploadId"`
	Initiated string `xml:"Initiated"`
}

// ListPartsResult is the response for ListParts.
type ListPartsResult struct {
	XMLName              xml.Name   `xml:"ListPartsResult"`
	Xmlns                string     `xml:"xmlns,attr"`
	Bucket               string     `xml:"Bucket"`
	Key                  string     `xml:"Key"`
	UploadID             string     `xml:"UploadId"`
	PartNumberMarker     int        `xml:"PartNumberMarker"`
	NextPartNumberMarker int        `xml:"NextPartNumberMarker"`
	MaxParts             int        `xml:"MaxParts"`
	IsTruncated          bool       `xml:"IsTruncated"`
	Parts                []PartInfo `xml:"Part"`
}

// PartInfo represents a part in the list parts response.
type PartInfo struct {
	PartNumber   int    `xml:"PartNumber"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
}

// CopyPartResult is the response for UploadPartCopy.
type CopyPartResult struct {
	XMLName      xml.Name `xml:"CopyPartResult"`
	Xmlns        string   `xml:"xmlns,attr"`
	LastModified string   `xml:"LastModified"`
	ETag         string   `xml:"ETag"`
}

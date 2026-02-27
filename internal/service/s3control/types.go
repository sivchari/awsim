package s3control

import "encoding/xml"

// PublicAccessBlockConfiguration represents public access block settings.
type PublicAccessBlockConfiguration struct {
	XMLName               xml.Name `xml:"PublicAccessBlockConfiguration"`
	BlockPublicAcls       bool     `xml:"BlockPublicAcls"`
	IgnorePublicAcls      bool     `xml:"IgnorePublicAcls"`
	BlockPublicPolicy     bool     `xml:"BlockPublicPolicy"`
	RestrictPublicBuckets bool     `xml:"RestrictPublicBuckets"`
}

// GetPublicAccessBlockOutput is the response for GetPublicAccessBlock.
type GetPublicAccessBlockOutput struct {
	XMLName                        xml.Name                        `xml:"GetPublicAccessBlockOutput"`
	PublicAccessBlockConfiguration *PublicAccessBlockConfiguration `xml:"PublicAccessBlockConfiguration"`
}

// AccessPoint represents an S3 access point.
type AccessPoint struct {
	Name              string
	Bucket            string
	AccountID         string
	NetworkOrigin     string
	VpcConfiguration  *VpcConfiguration
	PublicAccessBlock *PublicAccessBlockConfiguration
	Alias             string
	AccessPointArn    string
	Endpoints         map[string]string
	BucketAccountID   string
}

// VpcConfiguration represents VPC settings for an access point.
type VpcConfiguration struct {
	VpcID string `xml:"VpcId"`
}

// CreateAccessPointInput is the input for CreateAccessPoint.
type CreateAccessPointInput struct {
	XMLName                        xml.Name                        `xml:"CreateAccessPointRequest"`
	Bucket                         string                          `xml:"Bucket"`
	Name                           string                          `xml:"-"`
	VpcConfiguration               *VpcConfiguration               `xml:"VpcConfiguration,omitempty"`
	PublicAccessBlockConfiguration *PublicAccessBlockConfiguration `xml:"PublicAccessBlockConfiguration,omitempty"`
	BucketAccountID                string                          `xml:"BucketAccountId,omitempty"`
}

// CreateAccessPointResult is the response for CreateAccessPoint.
type CreateAccessPointResult struct {
	XMLName        xml.Name `xml:"CreateAccessPointResult"`
	AccessPointArn string   `xml:"AccessPointArn"`
	Alias          string   `xml:"Alias"`
}

// GetAccessPointResult is the response for GetAccessPoint.
type GetAccessPointResult struct {
	XMLName                        xml.Name                        `xml:"GetAccessPointResult"`
	Name                           string                          `xml:"Name"`
	Bucket                         string                          `xml:"Bucket"`
	NetworkOrigin                  string                          `xml:"NetworkOrigin"`
	VpcConfiguration               *VpcConfiguration               `xml:"VpcConfiguration,omitempty"`
	PublicAccessBlockConfiguration *PublicAccessBlockConfiguration `xml:"PublicAccessBlockConfiguration,omitempty"`
	AccessPointArn                 string                          `xml:"AccessPointArn"`
	Alias                          string                          `xml:"Alias"`
	Endpoints                      *Endpoints                      `xml:"Endpoints,omitempty"`
	BucketAccountID                string                          `xml:"BucketAccountId,omitempty"`
}

// Endpoints represents access point endpoints.
type Endpoints struct {
	Entry []EndpointEntry `xml:"entry"`
}

// EndpointEntry represents a single endpoint entry.
type EndpointEntry struct {
	Key   string `xml:"key"`
	Value string `xml:"value"`
}

// ListAccessPointsResult is the response for ListAccessPoints.
type ListAccessPointsResult struct {
	XMLName         xml.Name              `xml:"ListAccessPointsResult"`
	AccessPointList []AccessPointListItem `xml:"AccessPointList>AccessPoint"`
	NextToken       string                `xml:"NextToken,omitempty"`
}

// AccessPointListItem represents an access point in a list.
type AccessPointListItem struct {
	Name             string            `xml:"Name"`
	NetworkOrigin    string            `xml:"NetworkOrigin"`
	VpcConfiguration *VpcConfiguration `xml:"VpcConfiguration,omitempty"`
	Bucket           string            `xml:"Bucket"`
	AccessPointArn   string            `xml:"AccessPointArn"`
	Alias            string            `xml:"Alias"`
	BucketAccountID  string            `xml:"BucketAccountId,omitempty"`
}

// Error represents an S3 Control error response.
type Error struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	Resource  string   `xml:"Resource,omitempty"`
	RequestID string   `xml:"RequestId"`
}

// Error codes for S3 Control.
const (
	ErrNoSuchAccessPoint                    = "NoSuchAccessPoint"
	ErrAccessPointAlreadyOwnedByYou         = "AccessPointAlreadyOwnedByYou"
	ErrNoSuchPublicAccessBlockConfiguration = "NoSuchPublicAccessBlockConfiguration"
	ErrInvalidRequest                       = "InvalidRequest"
	ErrInternalError                        = "InternalError"
)

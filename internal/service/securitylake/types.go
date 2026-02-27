// Package securitylake provides an in-memory implementation of AWS Security Lake.
package securitylake

// Error represents an error response.
type Error struct {
	Code    string `json:"__type"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// DataLake represents a Security Lake data lake.
type DataLake struct {
	ARN                      string                    `json:"dataLakeArn,omitempty"`
	CreateStatus             string                    `json:"createStatus,omitempty"`
	EncryptionConfiguration  *EncryptionConfiguration  `json:"encryptionConfiguration,omitempty"`
	LifecycleConfiguration   *LifecycleConfiguration   `json:"lifecycleConfiguration,omitempty"`
	Region                   string                    `json:"region,omitempty"`
	ReplicationConfiguration *ReplicationConfiguration `json:"replicationConfiguration,omitempty"`
	S3BucketARN              string                    `json:"s3BucketArn,omitempty"`
	UpdateStatus             *DataLakeUpdateStatus     `json:"updateStatus,omitempty"`
}

// EncryptionConfiguration represents the encryption configuration for a data lake.
type EncryptionConfiguration struct {
	KmsKeyID string `json:"kmsKeyId,omitempty"`
}

// LifecycleConfiguration represents the lifecycle configuration for a data lake.
type LifecycleConfiguration struct {
	Expiration  *Expiration   `json:"expiration,omitempty"`
	Transitions []*Transition `json:"transitions,omitempty"`
}

// Expiration represents the expiration settings.
type Expiration struct {
	Days int `json:"days,omitempty"`
}

// Transition represents a storage class transition.
type Transition struct {
	Days         int    `json:"days,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
}

// ReplicationConfiguration represents the replication configuration.
type ReplicationConfiguration struct {
	Regions []string `json:"regions,omitempty"`
	RoleARN string   `json:"roleArn,omitempty"`
}

// DataLakeUpdateStatus represents the update status of a data lake.
type DataLakeUpdateStatus struct {
	Exception *DataLakeException `json:"exception,omitempty"`
	RequestID string             `json:"requestId,omitempty"`
	Status    string             `json:"status,omitempty"`
}

// DataLakeException represents an exception that occurred.
type DataLakeException struct {
	Exception   string `json:"exception,omitempty"`
	Region      string `json:"region,omitempty"`
	Remediation string `json:"remediation,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

// DataLakeConfiguration represents the configuration for creating a data lake.
type DataLakeConfiguration struct {
	EncryptionConfiguration  *EncryptionConfiguration  `json:"encryptionConfiguration,omitempty"`
	LifecycleConfiguration   *LifecycleConfiguration   `json:"lifecycleConfiguration,omitempty"`
	Region                   string                    `json:"region,omitempty"`
	ReplicationConfiguration *ReplicationConfiguration `json:"replicationConfiguration,omitempty"`
}

// Subscriber represents a Security Lake subscriber.
type Subscriber struct {
	AccessTypes           []string             `json:"accessTypes,omitempty"`
	CreatedAt             string               `json:"createdAt,omitempty"`
	ResourceShareARN      string               `json:"resourceShareArn,omitempty"`
	ResourceShareName     string               `json:"resourceShareName,omitempty"`
	RoleARN               string               `json:"roleArn,omitempty"`
	S3BucketARN           string               `json:"s3BucketArn,omitempty"`
	Sources               []*LogSourceResource `json:"sources,omitempty"`
	SubscriberARN         string               `json:"subscriberArn,omitempty"`
	SubscriberDescription string               `json:"subscriberDescription,omitempty"`
	SubscriberEndpoint    string               `json:"subscriberEndpoint,omitempty"`
	SubscriberID          string               `json:"subscriberId,omitempty"`
	SubscriberIdentity    *SubscriberIdentity  `json:"subscriberIdentity,omitempty"`
	SubscriberName        string               `json:"subscriberName,omitempty"`
	SubscriberStatus      string               `json:"subscriberStatus,omitempty"`
	UpdatedAt             string               `json:"updatedAt,omitempty"`
}

// SubscriberIdentity represents the identity of a subscriber.
type SubscriberIdentity struct {
	ExternalID string `json:"externalId,omitempty"`
	Principal  string `json:"principal,omitempty"`
}

// LogSourceResource represents a log source resource.
type LogSourceResource struct {
	AwsLogSource    *AwsLogSourceResource    `json:"awsLogSource,omitempty"`
	CustomLogSource *CustomLogSourceResource `json:"customLogSource,omitempty"`
}

// AwsLogSourceResource represents an AWS log source resource.
type AwsLogSourceResource struct {
	SourceName    string `json:"sourceName,omitempty"`
	SourceVersion string `json:"sourceVersion,omitempty"`
}

// CustomLogSourceResource represents a custom log source resource.
type CustomLogSourceResource struct {
	Attributes    *CustomLogSourceAttributes `json:"attributes,omitempty"`
	Provider      *CustomLogSourceProvider   `json:"provider,omitempty"`
	SourceName    string                     `json:"sourceName,omitempty"`
	SourceVersion string                     `json:"sourceVersion,omitempty"`
}

// CustomLogSourceAttributes represents custom log source attributes.
type CustomLogSourceAttributes struct {
	CrawlerARN  string `json:"crawlerArn,omitempty"`
	DatabaseARN string `json:"databaseArn,omitempty"`
	TableARN    string `json:"tableArn,omitempty"`
}

// CustomLogSourceProvider represents a custom log source provider.
type CustomLogSourceProvider struct {
	Location string `json:"location,omitempty"`
	RoleARN  string `json:"roleArn,omitempty"`
}

// Tag represents a resource tag.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CreateDataLakeRequest represents a CreateDataLake request.
type CreateDataLakeRequest struct {
	Configurations          []*DataLakeConfiguration `json:"configurations"`
	MetaStoreManagerRoleARN string                   `json:"metaStoreManagerRoleArn,omitempty"`
	Tags                    []*Tag                   `json:"tags,omitempty"`
}

// CreateDataLakeResponse represents a CreateDataLake response.
type CreateDataLakeResponse struct {
	DataLakes []*DataLake `json:"dataLakes,omitempty"`
}

// DeleteDataLakeRequest represents a DeleteDataLake request.
type DeleteDataLakeRequest struct {
	Regions []string `json:"regions"`
}

// DeleteDataLakeResponse represents a DeleteDataLake response.
type DeleteDataLakeResponse struct{}

// ListDataLakesRequest represents a ListDataLakes request.
type ListDataLakesRequest struct {
	Regions []string `json:"regions,omitempty"`
}

// ListDataLakesResponse represents a ListDataLakes response.
type ListDataLakesResponse struct {
	DataLakes []*DataLake `json:"dataLakes,omitempty"`
}

// UpdateDataLakeRequest represents an UpdateDataLake request.
type UpdateDataLakeRequest struct {
	Configurations          []*DataLakeConfiguration `json:"configurations"`
	MetaStoreManagerRoleARN string                   `json:"metaStoreManagerRoleArn,omitempty"`
}

// UpdateDataLakeResponse represents an UpdateDataLake response.
type UpdateDataLakeResponse struct {
	DataLakes []*DataLake `json:"dataLakes,omitempty"`
}

// CreateSubscriberRequest represents a CreateSubscriber request.
type CreateSubscriberRequest struct {
	AccessTypes           []string             `json:"accessTypes,omitempty"`
	Sources               []*LogSourceResource `json:"sources"`
	SubscriberDescription string               `json:"subscriberDescription,omitempty"`
	SubscriberIdentity    *SubscriberIdentity  `json:"subscriberIdentity"`
	SubscriberName        string               `json:"subscriberName"`
	Tags                  []*Tag               `json:"tags,omitempty"`
}

// CreateSubscriberResponse represents a CreateSubscriber response.
type CreateSubscriberResponse struct {
	Subscriber *Subscriber `json:"subscriber,omitempty"`
}

// GetSubscriberRequest represents a GetSubscriber request.
type GetSubscriberRequest struct {
	SubscriberID string `json:"subscriberId"`
}

// GetSubscriberResponse represents a GetSubscriber response.
type GetSubscriberResponse struct {
	Subscriber *Subscriber `json:"subscriber,omitempty"`
}

// DeleteSubscriberRequest represents a DeleteSubscriber request.
type DeleteSubscriberRequest struct {
	SubscriberID string `json:"subscriberId"`
}

// DeleteSubscriberResponse represents a DeleteSubscriber response.
type DeleteSubscriberResponse struct{}

// UpdateSubscriberRequest represents an UpdateSubscriber request.
type UpdateSubscriberRequest struct {
	Sources               []*LogSourceResource `json:"sources,omitempty"`
	SubscriberDescription string               `json:"subscriberDescription,omitempty"`
	SubscriberID          string               `json:"subscriberId"`
	SubscriberIdentity    *SubscriberIdentity  `json:"subscriberIdentity,omitempty"`
	SubscriberName        string               `json:"subscriberName,omitempty"`
}

// UpdateSubscriberResponse represents an UpdateSubscriber response.
type UpdateSubscriberResponse struct {
	Subscriber *Subscriber `json:"subscriber,omitempty"`
}

// ListSubscribersRequest represents a ListSubscribers request.
type ListSubscribersRequest struct {
	MaxResults int    `json:"maxResults,omitempty"`
	NextToken  string `json:"nextToken,omitempty"`
}

// ListSubscribersResponse represents a ListSubscribers response.
type ListSubscribersResponse struct {
	NextToken   string        `json:"nextToken,omitempty"`
	Subscribers []*Subscriber `json:"subscribers,omitempty"`
}

// CreateAwsLogSourceRequest represents a CreateAwsLogSource request.
type CreateAwsLogSourceRequest struct {
	Sources []*AwsLogSourceConfiguration `json:"sources"`
}

// AwsLogSourceConfiguration represents an AWS log source configuration.
type AwsLogSourceConfiguration struct {
	Accounts      []string `json:"accounts,omitempty"`
	Regions       []string `json:"regions"`
	SourceName    string   `json:"sourceName"`
	SourceVersion string   `json:"sourceVersion,omitempty"`
}

// CreateAwsLogSourceResponse represents a CreateAwsLogSource response.
type CreateAwsLogSourceResponse struct {
	Failed []string `json:"failed,omitempty"`
}

// DeleteAwsLogSourceRequest represents a DeleteAwsLogSource request.
type DeleteAwsLogSourceRequest struct {
	Sources []*AwsLogSourceConfiguration `json:"sources"`
}

// DeleteAwsLogSourceResponse represents a DeleteAwsLogSource response.
type DeleteAwsLogSourceResponse struct {
	Failed []string `json:"failed,omitempty"`
}

// ListLogSourcesRequest represents a ListLogSources request.
type ListLogSourcesRequest struct {
	Accounts   []string             `json:"accounts,omitempty"`
	MaxResults int                  `json:"maxResults,omitempty"`
	NextToken  string               `json:"nextToken,omitempty"`
	Regions    []string             `json:"regions,omitempty"`
	Sources    []*LogSourceResource `json:"sources,omitempty"`
}

// LogSource represents a log source.
type LogSource struct {
	Account string               `json:"account,omitempty"`
	Region  string               `json:"region,omitempty"`
	Sources []*LogSourceResource `json:"sources,omitempty"`
}

// ListLogSourcesResponse represents a ListLogSources response.
type ListLogSourcesResponse struct {
	NextToken string       `json:"nextToken,omitempty"`
	Sources   []*LogSource `json:"sources,omitempty"`
}

// TagResourceRequest represents a TagResource request.
type TagResourceRequest struct {
	ResourceARN string `json:"resourceArn"`
	Tags        []*Tag `json:"tags"`
}

// TagResourceResponse represents a TagResource response.
type TagResourceResponse struct{}

// UntagResourceRequest represents an UntagResource request.
type UntagResourceRequest struct {
	ResourceARN string   `json:"resourceArn"`
	TagKeys     []string `json:"tagKeys"`
}

// UntagResourceResponse represents an UntagResource response.
type UntagResourceResponse struct{}

// ListTagsForResourceRequest represents a ListTagsForResource request.
type ListTagsForResourceRequest struct {
	ResourceARN string `json:"resourceArn"`
}

// ListTagsForResourceResponse represents a ListTagsForResource response.
type ListTagsForResourceResponse struct {
	Tags []*Tag `json:"tags,omitempty"`
}

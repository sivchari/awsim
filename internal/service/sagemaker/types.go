package sagemaker

import "time"

// NotebookInstance represents a SageMaker notebook instance.
type NotebookInstance struct {
	NotebookInstanceName   string
	NotebookInstanceArn    string
	NotebookInstanceStatus string
	URL                    string
	InstanceType           string
	RoleArn                string
	KmsKeyID               string
	SubnetID               string
	SecurityGroups         []string
	DirectInternetAccess   string
	VolumeSizeInGB         int32
	AcceleratorTypes       []string
	DefaultCodeRepository  string
	AdditionalCodeRepos    []string
	RootAccess             string
	PlatformIdentifier     string
	InstanceMetadataConfig *InstanceMetadataConfig
	CreationTime           time.Time
	LastModifiedTime       time.Time
}

// InstanceMetadataConfig represents instance metadata service configuration.
type InstanceMetadataConfig struct {
	MinimumInstanceMetadataServiceVersion string `json:"MinimumInstanceMetadataServiceVersion"`
}

// TrainingJob represents a SageMaker training job.
type TrainingJob struct {
	TrainingJobName   string
	TrainingJobArn    string
	TrainingJobStatus string
	SecondaryStatus   string
	AlgorithmSpec     *AlgorithmSpecification
	RoleArn           string
	InputDataConfig   []Channel
	OutputDataConfig  *OutputDataConfig
	ResourceConfig    *ResourceConfig
	StoppingCondition *StoppingCondition
	CreationTime      time.Time
	TrainingStartTime *time.Time
	TrainingEndTime   *time.Time
	FailureReason     string
}

// AlgorithmSpecification represents the algorithm specification for a training job.
type AlgorithmSpecification struct {
	TrainingImage     string `json:"TrainingImage"`
	TrainingInputMode string `json:"TrainingInputMode"`
	AlgorithmName     string `json:"AlgorithmName,omitempty"`
}

// Channel represents an input data channel for a training job.
type Channel struct {
	ChannelName string      `json:"ChannelName"`
	DataSource  *DataSource `json:"DataSource"`
	ContentType string      `json:"ContentType,omitempty"`
}

// DataSource represents the data source for a channel.
type DataSource struct {
	S3DataSource *S3DataSource `json:"S3DataSource,omitempty"`
}

// S3DataSource represents an S3 data source.
type S3DataSource struct {
	S3DataType             string `json:"S3DataType"`
	S3Uri                  string `json:"S3Uri"`
	S3DataDistributionType string `json:"S3DataDistributionType,omitempty"`
}

// OutputDataConfig represents output data configuration.
type OutputDataConfig struct {
	S3OutputPath string `json:"S3OutputPath"`
	KmsKeyID     string `json:"KmsKeyId,omitempty"`
}

// ResourceConfig represents resource configuration for training.
type ResourceConfig struct {
	InstanceType   string `json:"InstanceType"`
	InstanceCount  int32  `json:"InstanceCount"`
	VolumeSizeInGB int32  `json:"VolumeSizeInGB"`
}

// StoppingCondition represents stopping conditions for a training job.
type StoppingCondition struct {
	MaxRuntimeInSeconds int32 `json:"MaxRuntimeInSeconds"`
}

// Model represents a SageMaker model.
type Model struct {
	ModelName              string
	ModelArn               string
	PrimaryContainer       *ContainerDefinition
	ExecutionRoleArn       string
	EnableNetworkIsolation bool
	CreationTime           time.Time
}

// ContainerDefinition represents a container definition.
type ContainerDefinition struct {
	Image             string            `json:"Image"`
	Mode              string            `json:"Mode,omitempty"`
	ModelDataURL      string            `json:"ModelDataUrl,omitempty"`
	Environment       map[string]string `json:"Environment,omitempty"`
	ContainerHostname string            `json:"ContainerHostname,omitempty"`
}

// Endpoint represents a SageMaker endpoint.
type Endpoint struct {
	EndpointName       string
	EndpointArn        string
	EndpointConfigName string
	EndpointStatus     string
	CreationTime       time.Time
	LastModifiedTime   time.Time
	FailureReason      string
}

// Error represents a service error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}

// CreateNotebookInstanceRequest represents a CreateNotebookInstance request.
type CreateNotebookInstanceRequest struct {
	NotebookInstanceName   string                  `json:"NotebookInstanceName"`
	InstanceType           string                  `json:"InstanceType"`
	RoleArn                string                  `json:"RoleArn"`
	SubnetID               string                  `json:"SubnetId,omitempty"`
	SecurityGroupIDs       []string                `json:"SecurityGroupIds,omitempty"`
	KmsKeyID               string                  `json:"KmsKeyId,omitempty"`
	DirectInternetAccess   string                  `json:"DirectInternetAccess,omitempty"`
	VolumeSizeInGB         int32                   `json:"VolumeSizeInGB,omitempty"`
	AcceleratorTypes       []string                `json:"AcceleratorTypes,omitempty"`
	DefaultCodeRepository  string                  `json:"DefaultCodeRepository,omitempty"`
	AdditionalCodeRepos    []string                `json:"AdditionalCodeRepositories,omitempty"`
	RootAccess             string                  `json:"RootAccess,omitempty"`
	PlatformIdentifier     string                  `json:"PlatformIdentifier,omitempty"`
	InstanceMetadataConfig *InstanceMetadataConfig `json:"InstanceMetadataServiceConfiguration,omitempty"`
	Tags                   []Tag                   `json:"Tags,omitempty"`
}

// Tag represents a key-value tag.
type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// CreateNotebookInstanceResponse represents a CreateNotebookInstance response.
type CreateNotebookInstanceResponse struct {
	NotebookInstanceArn string `json:"NotebookInstanceArn"`
}

// DeleteNotebookInstanceRequest represents a DeleteNotebookInstance request.
type DeleteNotebookInstanceRequest struct {
	NotebookInstanceName string `json:"NotebookInstanceName"`
}

// DescribeNotebookInstanceRequest represents a DescribeNotebookInstance request.
type DescribeNotebookInstanceRequest struct {
	NotebookInstanceName string `json:"NotebookInstanceName"`
}

// DescribeNotebookInstanceResponse represents a DescribeNotebookInstance response.
type DescribeNotebookInstanceResponse struct {
	NotebookInstanceName   string   `json:"NotebookInstanceName"`
	NotebookInstanceArn    string   `json:"NotebookInstanceArn"`
	NotebookInstanceStatus string   `json:"NotebookInstanceStatus"`
	URL                    string   `json:"Url,omitempty"`
	InstanceType           string   `json:"InstanceType"`
	RoleArn                string   `json:"RoleArn"`
	KmsKeyID               string   `json:"KmsKeyId,omitempty"`
	SubnetID               string   `json:"SubnetId,omitempty"`
	SecurityGroups         []string `json:"SecurityGroups,omitempty"`
	DirectInternetAccess   string   `json:"DirectInternetAccess,omitempty"`
	VolumeSizeInGB         int32    `json:"VolumeSizeInGB,omitempty"`
	RootAccess             string   `json:"RootAccess,omitempty"`
	CreationTime           float64  `json:"CreationTime"`
	LastModifiedTime       float64  `json:"LastModifiedTime"`
}

// ListNotebookInstancesRequest represents a ListNotebookInstances request.
type ListNotebookInstancesRequest struct {
	MaxResults int32  `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
	SortBy     string `json:"SortBy,omitempty"`
	SortOrder  string `json:"SortOrder,omitempty"`
}

// NotebookInstanceSummary represents a notebook instance summary.
type NotebookInstanceSummary struct {
	NotebookInstanceName   string  `json:"NotebookInstanceName"`
	NotebookInstanceArn    string  `json:"NotebookInstanceArn"`
	NotebookInstanceStatus string  `json:"NotebookInstanceStatus"`
	URL                    string  `json:"Url,omitempty"`
	InstanceType           string  `json:"InstanceType"`
	CreationTime           float64 `json:"CreationTime"`
	LastModifiedTime       float64 `json:"LastModifiedTime"`
}

// ListNotebookInstancesResponse represents a ListNotebookInstances response.
type ListNotebookInstancesResponse struct {
	NotebookInstances []NotebookInstanceSummary `json:"NotebookInstances"`
	NextToken         string                    `json:"NextToken,omitempty"`
}

// CreateTrainingJobRequest represents a CreateTrainingJob request.
type CreateTrainingJobRequest struct {
	TrainingJobName   string                  `json:"TrainingJobName"`
	AlgorithmSpec     *AlgorithmSpecification `json:"AlgorithmSpecification"`
	RoleArn           string                  `json:"RoleArn"`
	InputDataConfig   []Channel               `json:"InputDataConfig,omitempty"`
	OutputDataConfig  *OutputDataConfig       `json:"OutputDataConfig"`
	ResourceConfig    *ResourceConfig         `json:"ResourceConfig"`
	StoppingCondition *StoppingCondition      `json:"StoppingCondition"`
	Tags              []Tag                   `json:"Tags,omitempty"`
}

// CreateTrainingJobResponse represents a CreateTrainingJob response.
type CreateTrainingJobResponse struct {
	TrainingJobArn string `json:"TrainingJobArn"`
}

// DescribeTrainingJobRequest represents a DescribeTrainingJob request.
type DescribeTrainingJobRequest struct {
	TrainingJobName string `json:"TrainingJobName"`
}

// DescribeTrainingJobResponse represents a DescribeTrainingJob response.
type DescribeTrainingJobResponse struct {
	TrainingJobName   string                  `json:"TrainingJobName"`
	TrainingJobArn    string                  `json:"TrainingJobArn"`
	TrainingJobStatus string                  `json:"TrainingJobStatus"`
	SecondaryStatus   string                  `json:"SecondaryStatus"`
	AlgorithmSpec     *AlgorithmSpecification `json:"AlgorithmSpecification,omitempty"`
	RoleArn           string                  `json:"RoleArn"`
	InputDataConfig   []Channel               `json:"InputDataConfig,omitempty"`
	OutputDataConfig  *OutputDataConfig       `json:"OutputDataConfig,omitempty"`
	ResourceConfig    *ResourceConfig         `json:"ResourceConfig,omitempty"`
	StoppingCondition *StoppingCondition      `json:"StoppingCondition,omitempty"`
	CreationTime      float64                 `json:"CreationTime"`
	TrainingStartTime *float64                `json:"TrainingStartTime,omitempty"`
	TrainingEndTime   *float64                `json:"TrainingEndTime,omitempty"`
	FailureReason     string                  `json:"FailureReason,omitempty"`
}

// CreateModelRequest represents a CreateModel request.
type CreateModelRequest struct {
	ModelName              string               `json:"ModelName"`
	PrimaryContainer       *ContainerDefinition `json:"PrimaryContainer"`
	ExecutionRoleArn       string               `json:"ExecutionRoleArn"`
	EnableNetworkIsolation bool                 `json:"EnableNetworkIsolation,omitempty"`
	Tags                   []Tag                `json:"Tags,omitempty"`
}

// CreateModelResponse represents a CreateModel response.
type CreateModelResponse struct {
	ModelArn string `json:"ModelArn"`
}

// DeleteModelRequest represents a DeleteModel request.
type DeleteModelRequest struct {
	ModelName string `json:"ModelName"`
}

// CreateEndpointRequest represents a CreateEndpoint request.
type CreateEndpointRequest struct {
	EndpointName       string `json:"EndpointName"`
	EndpointConfigName string `json:"EndpointConfigName"`
	Tags               []Tag  `json:"Tags,omitempty"`
}

// CreateEndpointResponse represents a CreateEndpoint response.
type CreateEndpointResponse struct {
	EndpointArn string `json:"EndpointArn"`
}

// DeleteEndpointRequest represents a DeleteEndpoint request.
type DeleteEndpointRequest struct {
	EndpointName string `json:"EndpointName"`
}

// DescribeEndpointRequest represents a DescribeEndpoint request.
type DescribeEndpointRequest struct {
	EndpointName string `json:"EndpointName"`
}

// DescribeEndpointResponse represents a DescribeEndpoint response.
type DescribeEndpointResponse struct {
	EndpointName       string  `json:"EndpointName"`
	EndpointArn        string  `json:"EndpointArn"`
	EndpointConfigName string  `json:"EndpointConfigName"`
	EndpointStatus     string  `json:"EndpointStatus"`
	CreationTime       float64 `json:"CreationTime"`
	LastModifiedTime   float64 `json:"LastModifiedTime"`
	FailureReason      string  `json:"FailureReason,omitempty"`
}

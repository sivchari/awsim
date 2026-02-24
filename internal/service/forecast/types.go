// Package forecast provides Amazon Forecast service emulation for awsim.
package forecast

import (
	"encoding/json"
	"time"
)

// AWSTimestamp is a time.Time that marshals to Unix timestamp (float64).
type AWSTimestamp struct {
	time.Time
}

// MarshalJSON implements json.Marshaler.
func (t AWSTimestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return json.Marshal(nil) //nolint:wrapcheck // MarshalJSON interface requirement
	}

	return json.Marshal(float64(t.Unix()) + float64(t.Nanosecond())/1e9) //nolint:wrapcheck // MarshalJSON interface requirement
}

// ToAWSTimestamp converts time.Time to AWSTimestamp.
func ToAWSTimestamp(t time.Time) AWSTimestamp {
	return AWSTimestamp{Time: t}
}

// Domain types.

// Dataset represents an Amazon Forecast dataset.
type Dataset struct {
	DatasetArn           string            `json:"DatasetArn,omitempty"`
	DatasetName          string            `json:"DatasetName,omitempty"`
	DatasetType          string            `json:"DatasetType,omitempty"`
	Domain               string            `json:"Domain,omitempty"`
	DataFrequency        string            `json:"DataFrequency,omitempty"`
	Schema               *Schema           `json:"Schema,omitempty"`
	Status               string            `json:"Status,omitempty"`
	CreationTime         AWSTimestamp      `json:"CreationTime,omitempty"`
	LastModificationTime AWSTimestamp      `json:"LastModificationTime,omitempty"`
	EncryptionConfig     *EncryptionConfig `json:"EncryptionConfig,omitempty"`
}

// Schema represents a dataset schema.
type Schema struct {
	Attributes []SchemaAttribute `json:"Attributes,omitempty"`
}

// SchemaAttribute represents a schema attribute.
type SchemaAttribute struct {
	AttributeName string `json:"AttributeName,omitempty"`
	AttributeType string `json:"AttributeType,omitempty"`
}

// EncryptionConfig represents encryption configuration.
type EncryptionConfig struct {
	KMSKeyArn string `json:"KMSKeyArn,omitempty"`
	RoleArn   string `json:"RoleArn,omitempty"`
}

// DatasetGroup represents an Amazon Forecast dataset group.
type DatasetGroup struct {
	DatasetGroupArn      string       `json:"DatasetGroupArn,omitempty"`
	DatasetGroupName     string       `json:"DatasetGroupName,omitempty"`
	Domain               string       `json:"Domain,omitempty"`
	DatasetArns          []string     `json:"DatasetArns,omitempty"`
	Status               string       `json:"Status,omitempty"`
	CreationTime         AWSTimestamp `json:"CreationTime,omitempty"`
	LastModificationTime AWSTimestamp `json:"LastModificationTime,omitempty"`
}

// Predictor represents an Amazon Forecast predictor.
type Predictor struct {
	PredictorArn         string               `json:"PredictorArn,omitempty"`
	PredictorName        string               `json:"PredictorName,omitempty"`
	AlgorithmArn         string               `json:"AlgorithmArn,omitempty"`
	ForecastHorizon      int32                `json:"ForecastHorizon,omitempty"`
	ForecastTypes        []string             `json:"ForecastTypes,omitempty"`
	InputDataConfig      *InputDataConfig     `json:"InputDataConfig,omitempty"`
	FeaturizationConfig  *FeaturizationConfig `json:"FeaturizationConfig,omitempty"`
	Status               string               `json:"Status,omitempty"`
	CreationTime         AWSTimestamp         `json:"CreationTime,omitempty"`
	LastModificationTime AWSTimestamp         `json:"LastModificationTime,omitempty"`
	Message              string               `json:"Message,omitempty"`
}

// InputDataConfig represents input data configuration.
type InputDataConfig struct {
	DatasetGroupArn       string                 `json:"DatasetGroupArn,omitempty"`
	SupplementaryFeatures []SupplementaryFeature `json:"SupplementaryFeatures,omitempty"`
}

// SupplementaryFeature represents a supplementary feature.
type SupplementaryFeature struct {
	Name  string `json:"Name,omitempty"`
	Value string `json:"Value,omitempty"`
}

// FeaturizationConfig represents featurization configuration.
type FeaturizationConfig struct {
	ForecastFrequency  string          `json:"ForecastFrequency,omitempty"`
	ForecastDimensions []string        `json:"ForecastDimensions,omitempty"`
	Featurizations     []Featurization `json:"Featurizations,omitempty"`
}

// Featurization represents a featurization.
type Featurization struct {
	AttributeName         string                `json:"AttributeName,omitempty"`
	FeaturizationPipeline []FeaturizationMethod `json:"FeaturizationPipeline,omitempty"`
}

// FeaturizationMethod represents a featurization method.
type FeaturizationMethod struct {
	FeaturizationMethodName       string            `json:"FeaturizationMethodName,omitempty"`
	FeaturizationMethodParameters map[string]string `json:"FeaturizationMethodParameters,omitempty"`
}

// Forecast represents an Amazon Forecast forecast.
type Forecast struct {
	ForecastArn                     string       `json:"ForecastArn,omitempty"`
	ForecastName                    string       `json:"ForecastName,omitempty"`
	PredictorArn                    string       `json:"PredictorArn,omitempty"`
	DatasetGroupArn                 string       `json:"DatasetGroupArn,omitempty"`
	ForecastTypes                   []string     `json:"ForecastTypes,omitempty"`
	Status                          string       `json:"Status,omitempty"`
	CreationTime                    AWSTimestamp `json:"CreationTime,omitempty"`
	LastModificationTime            AWSTimestamp `json:"LastModificationTime,omitempty"`
	EstimatedTimeRemainingInMinutes *int64       `json:"EstimatedTimeRemainingInMinutes,omitempty"`
	Message                         string       `json:"Message,omitempty"`
}

// DatasetSummary represents a dataset summary for list operations.
type DatasetSummary struct {
	DatasetArn           string       `json:"DatasetArn,omitempty"`
	DatasetName          string       `json:"DatasetName,omitempty"`
	DatasetType          string       `json:"DatasetType,omitempty"`
	Domain               string       `json:"Domain,omitempty"`
	CreationTime         AWSTimestamp `json:"CreationTime,omitempty"`
	LastModificationTime AWSTimestamp `json:"LastModificationTime,omitempty"`
}

// DatasetGroupSummary represents a dataset group summary for list operations.
type DatasetGroupSummary struct {
	DatasetGroupArn      string       `json:"DatasetGroupArn,omitempty"`
	DatasetGroupName     string       `json:"DatasetGroupName,omitempty"`
	CreationTime         AWSTimestamp `json:"CreationTime,omitempty"`
	LastModificationTime AWSTimestamp `json:"LastModificationTime,omitempty"`
}

// PredictorSummary represents a predictor summary for list operations.
type PredictorSummary struct {
	PredictorArn         string       `json:"PredictorArn,omitempty"`
	PredictorName        string       `json:"PredictorName,omitempty"`
	DatasetGroupArn      string       `json:"DatasetGroupArn,omitempty"`
	Status               string       `json:"Status,omitempty"`
	CreationTime         AWSTimestamp `json:"CreationTime,omitempty"`
	LastModificationTime AWSTimestamp `json:"LastModificationTime,omitempty"`
	Message              string       `json:"Message,omitempty"`
}

// ForecastSummary represents a forecast summary for list operations.
type ForecastSummary struct {
	ForecastArn          string       `json:"ForecastArn,omitempty"`
	ForecastName         string       `json:"ForecastName,omitempty"`
	PredictorArn         string       `json:"PredictorArn,omitempty"`
	DatasetGroupArn      string       `json:"DatasetGroupArn,omitempty"`
	Status               string       `json:"Status,omitempty"`
	CreationTime         AWSTimestamp `json:"CreationTime,omitempty"`
	LastModificationTime AWSTimestamp `json:"LastModificationTime,omitempty"`
	Message              string       `json:"Message,omitempty"`
}

// Request/Response types.

// CreateDatasetInput represents the input for CreateDataset.
type CreateDatasetInput struct {
	DatasetName      string            `json:"DatasetName"`
	DatasetType      string            `json:"DatasetType"`
	Domain           string            `json:"Domain"`
	DataFrequency    string            `json:"DataFrequency,omitempty"`
	Schema           *Schema           `json:"Schema"`
	EncryptionConfig *EncryptionConfig `json:"EncryptionConfig,omitempty"`
	Tags             []Tag             `json:"Tags,omitempty"`
}

// CreateDatasetOutput represents the output for CreateDataset.
type CreateDatasetOutput struct {
	DatasetArn string `json:"DatasetArn,omitempty"`
}

// DescribeDatasetInput represents the input for DescribeDataset.
type DescribeDatasetInput struct {
	DatasetArn string `json:"DatasetArn"`
}

// DescribeDatasetOutput represents the output for DescribeDataset.
type DescribeDatasetOutput struct {
	DatasetArn           string            `json:"DatasetArn,omitempty"`
	DatasetName          string            `json:"DatasetName,omitempty"`
	DatasetType          string            `json:"DatasetType,omitempty"`
	Domain               string            `json:"Domain,omitempty"`
	DataFrequency        string            `json:"DataFrequency,omitempty"`
	Schema               *Schema           `json:"Schema,omitempty"`
	Status               string            `json:"Status,omitempty"`
	CreationTime         AWSTimestamp      `json:"CreationTime,omitempty"`
	LastModificationTime AWSTimestamp      `json:"LastModificationTime,omitempty"`
	EncryptionConfig     *EncryptionConfig `json:"EncryptionConfig,omitempty"`
}

// ListDatasetsInput represents the input for ListDatasets.
type ListDatasetsInput struct {
	MaxResults *int32 `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// ListDatasetsOutput represents the output for ListDatasets.
type ListDatasetsOutput struct {
	Datasets  []*DatasetSummary `json:"Datasets,omitempty"`
	NextToken string            `json:"NextToken,omitempty"`
}

// DeleteDatasetInput represents the input for DeleteDataset.
type DeleteDatasetInput struct {
	DatasetArn string `json:"DatasetArn"`
}

// CreateDatasetGroupInput represents the input for CreateDatasetGroup.
type CreateDatasetGroupInput struct {
	DatasetGroupName string   `json:"DatasetGroupName"`
	Domain           string   `json:"Domain"`
	DatasetArns      []string `json:"DatasetArns,omitempty"`
	Tags             []Tag    `json:"Tags,omitempty"`
}

// CreateDatasetGroupOutput represents the output for CreateDatasetGroup.
type CreateDatasetGroupOutput struct {
	DatasetGroupArn string `json:"DatasetGroupArn,omitempty"`
}

// DescribeDatasetGroupInput represents the input for DescribeDatasetGroup.
type DescribeDatasetGroupInput struct {
	DatasetGroupArn string `json:"DatasetGroupArn"`
}

// DescribeDatasetGroupOutput represents the output for DescribeDatasetGroup.
type DescribeDatasetGroupOutput struct {
	DatasetGroupArn      string       `json:"DatasetGroupArn,omitempty"`
	DatasetGroupName     string       `json:"DatasetGroupName,omitempty"`
	Domain               string       `json:"Domain,omitempty"`
	DatasetArns          []string     `json:"DatasetArns,omitempty"`
	Status               string       `json:"Status,omitempty"`
	CreationTime         AWSTimestamp `json:"CreationTime,omitempty"`
	LastModificationTime AWSTimestamp `json:"LastModificationTime,omitempty"`
}

// ListDatasetGroupsInput represents the input for ListDatasetGroups.
type ListDatasetGroupsInput struct {
	MaxResults *int32 `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// ListDatasetGroupsOutput represents the output for ListDatasetGroups.
type ListDatasetGroupsOutput struct {
	DatasetGroups []*DatasetGroupSummary `json:"DatasetGroups,omitempty"`
	NextToken     string                 `json:"NextToken,omitempty"`
}

// DeleteDatasetGroupInput represents the input for DeleteDatasetGroup.
type DeleteDatasetGroupInput struct {
	DatasetGroupArn string `json:"DatasetGroupArn"`
}

// UpdateDatasetGroupInput represents the input for UpdateDatasetGroup.
type UpdateDatasetGroupInput struct {
	DatasetGroupArn string   `json:"DatasetGroupArn"`
	DatasetArns     []string `json:"DatasetArns"`
}

// CreatePredictorInput represents the input for CreatePredictor.
type CreatePredictorInput struct {
	PredictorName       string               `json:"PredictorName"`
	AlgorithmArn        string               `json:"AlgorithmArn,omitempty"`
	ForecastHorizon     int32                `json:"ForecastHorizon"`
	ForecastTypes       []string             `json:"ForecastTypes,omitempty"`
	InputDataConfig     *InputDataConfig     `json:"InputDataConfig"`
	FeaturizationConfig *FeaturizationConfig `json:"FeaturizationConfig"`
	EncryptionConfig    *EncryptionConfig    `json:"EncryptionConfig,omitempty"`
	Tags                []Tag                `json:"Tags,omitempty"`
}

// CreatePredictorOutput represents the output for CreatePredictor.
type CreatePredictorOutput struct {
	PredictorArn string `json:"PredictorArn,omitempty"`
}

// DescribePredictorInput represents the input for DescribePredictor.
type DescribePredictorInput struct {
	PredictorArn string `json:"PredictorArn"`
}

// DescribePredictorOutput represents the output for DescribePredictor.
type DescribePredictorOutput struct {
	PredictorArn         string               `json:"PredictorArn,omitempty"`
	PredictorName        string               `json:"PredictorName,omitempty"`
	AlgorithmArn         string               `json:"AlgorithmArn,omitempty"`
	ForecastHorizon      int32                `json:"ForecastHorizon,omitempty"`
	ForecastTypes        []string             `json:"ForecastTypes,omitempty"`
	InputDataConfig      *InputDataConfig     `json:"InputDataConfig,omitempty"`
	FeaturizationConfig  *FeaturizationConfig `json:"FeaturizationConfig,omitempty"`
	Status               string               `json:"Status,omitempty"`
	CreationTime         AWSTimestamp         `json:"CreationTime,omitempty"`
	LastModificationTime AWSTimestamp         `json:"LastModificationTime,omitempty"`
	Message              string               `json:"Message,omitempty"`
}

// ListPredictorsInput represents the input for ListPredictors.
type ListPredictorsInput struct {
	MaxResults *int32   `json:"MaxResults,omitempty"`
	NextToken  string   `json:"NextToken,omitempty"`
	Filters    []Filter `json:"Filters,omitempty"`
}

// ListPredictorsOutput represents the output for ListPredictors.
type ListPredictorsOutput struct {
	Predictors []*PredictorSummary `json:"Predictors,omitempty"`
	NextToken  string              `json:"NextToken,omitempty"`
}

// DeletePredictorInput represents the input for DeletePredictor.
type DeletePredictorInput struct {
	PredictorArn string `json:"PredictorArn"`
}

// CreateForecastInput represents the input for CreateForecast.
type CreateForecastInput struct {
	ForecastName  string   `json:"ForecastName"`
	PredictorArn  string   `json:"PredictorArn"`
	ForecastTypes []string `json:"ForecastTypes,omitempty"`
	Tags          []Tag    `json:"Tags,omitempty"`
}

// CreateForecastOutput represents the output for CreateForecast.
type CreateForecastOutput struct {
	ForecastArn string `json:"ForecastArn,omitempty"`
}

// DescribeForecastInput represents the input for DescribeForecast.
type DescribeForecastInput struct {
	ForecastArn string `json:"ForecastArn"`
}

// DescribeForecastOutput represents the output for DescribeForecast.
type DescribeForecastOutput struct {
	ForecastArn                     string       `json:"ForecastArn,omitempty"`
	ForecastName                    string       `json:"ForecastName,omitempty"`
	PredictorArn                    string       `json:"PredictorArn,omitempty"`
	DatasetGroupArn                 string       `json:"DatasetGroupArn,omitempty"`
	ForecastTypes                   []string     `json:"ForecastTypes,omitempty"`
	Status                          string       `json:"Status,omitempty"`
	CreationTime                    AWSTimestamp `json:"CreationTime,omitempty"`
	LastModificationTime            AWSTimestamp `json:"LastModificationTime,omitempty"`
	EstimatedTimeRemainingInMinutes *int64       `json:"EstimatedTimeRemainingInMinutes,omitempty"`
	Message                         string       `json:"Message,omitempty"`
}

// ListForecastsInput represents the input for ListForecasts.
type ListForecastsInput struct {
	MaxResults *int32   `json:"MaxResults,omitempty"`
	NextToken  string   `json:"NextToken,omitempty"`
	Filters    []Filter `json:"Filters,omitempty"`
}

// ListForecastsOutput represents the output for ListForecasts.
type ListForecastsOutput struct {
	Forecasts []*ForecastSummary `json:"Forecasts,omitempty"`
	NextToken string             `json:"NextToken,omitempty"`
}

// DeleteForecastInput represents the input for DeleteForecast.
type DeleteForecastInput struct {
	ForecastArn string `json:"ForecastArn"`
}

// Filter represents a filter for list operations.
type Filter struct {
	Condition string `json:"Condition,omitempty"`
	Key       string `json:"Key,omitempty"`
	Value     string `json:"Value,omitempty"`
}

// Tag represents a resource tag.
type Tag struct {
	Key   string `json:"Key,omitempty"`
	Value string `json:"Value,omitempty"`
}

// Error represents a Forecast error.
type Error struct {
	Code    string `json:"__type"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Status constants.
const (
	statusActive           = "ACTIVE"
	statusCreatePending    = "CREATE_PENDING"
	statusCreateInProgress = "CREATE_IN_PROGRESS"
	statusCreateFailed     = "CREATE_FAILED"
	statusDeletePending    = "DELETE_PENDING"
	statusDeleteInProgress = "DELETE_IN_PROGRESS"
	statusDeleteFailed     = "DELETE_FAILED"
	statusUpdatePending    = "UPDATE_PENDING"
	statusUpdateInProgress = "UPDATE_IN_PROGRESS"
	statusUpdateFailed     = "UPDATE_FAILED"
)

// Dataset type constants.
const (
	datasetTypeTargetTimeSeries  = "TARGET_TIME_SERIES"
	datasetTypeRelatedTimeSeries = "RELATED_TIME_SERIES"
	datasetTypeItemMetadata      = "ITEM_METADATA"
)

// Domain constants.
const (
	domainRetail            = "RETAIL"
	domainCustom            = "CUSTOM"
	domainInventoryPlanning = "INVENTORY_PLANNING"
	domainEC2Capacity       = "EC2_CAPACITY"
	domainWorkForce         = "WORK_FORCE"
	domainWebTraffic        = "WEB_TRAFFIC"
	domainMetrics           = "METRICS"
)

// Error code constants.
const (
	errInvalidInputException          = "InvalidInputException"
	errResourceNotFoundException      = "ResourceNotFoundException"
	errResourceAlreadyExistsException = "ResourceAlreadyExistsException"
	errResourceInUseException         = "ResourceInUseException"
)

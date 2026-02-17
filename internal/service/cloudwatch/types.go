// Package cloudwatch provides CloudWatch metrics service emulation for awsim.
package cloudwatch

// Metric represents a CloudWatch metric.
type Metric struct {
	Namespace  string      `json:"Namespace"`
	MetricName string      `json:"MetricName"`
	Dimensions []Dimension `json:"Dimensions,omitempty"`
}

// Dimension represents a CloudWatch dimension.
type Dimension struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

// MetricDatum represents a single metric data point.
type MetricDatum struct {
	MetricName        string        `json:"MetricName"`
	Dimensions        []Dimension   `json:"Dimensions,omitempty"`
	Timestamp         string        `json:"Timestamp,omitempty"`
	Value             *float64      `json:"Value,omitempty"`
	StatisticValues   *StatisticSet `json:"StatisticValues,omitempty"`
	Values            []float64     `json:"Values,omitempty"`
	Counts            []float64     `json:"Counts,omitempty"`
	Unit              string        `json:"Unit,omitempty"`
	StorageResolution *int32        `json:"StorageResolution,omitempty"`
}

// StatisticSet represents a set of statistics.
type StatisticSet struct {
	SampleCount float64 `json:"SampleCount"`
	Sum         float64 `json:"Sum"`
	Minimum     float64 `json:"Minimum"`
	Maximum     float64 `json:"Maximum"`
}

// MetricDatapoint represents a metric data point in storage.
type MetricDatapoint struct {
	Timestamp string
	Value     float64
	Unit      string
}

// Alarm represents a CloudWatch alarm.
type Alarm struct {
	AlarmName          string
	AlarmARN           string
	AlarmDescription   string
	MetricName         string
	Namespace          string
	Statistic          string
	Dimensions         []Dimension
	Period             int32
	EvaluationPeriods  int32
	Threshold          float64
	ComparisonOperator string
	ActionsEnabled     bool
	AlarmActions       []string
	OKActions          []string
	StateValue         string
	StateReason        string
	StateUpdatedAt     string
	CreatedAt          string
}

// PutMetricDataRequest is the request for PutMetricData.
type PutMetricDataRequest struct {
	Namespace  string        `json:"Namespace"`
	MetricData []MetricDatum `json:"MetricData"`
}

// GetMetricDataRequest is the request for GetMetricData.
type GetMetricDataRequest struct {
	MetricDataQueries []MetricDataQuery `json:"MetricDataQueries"`
	StartTime         string            `json:"StartTime"`
	EndTime           string            `json:"EndTime"`
	NextToken         string            `json:"NextToken,omitempty"`
	MaxDatapoints     *int32            `json:"MaxDatapoints,omitempty"`
}

// MetricDataQuery represents a metric data query.
type MetricDataQuery struct {
	ID         string      `json:"Id"`
	MetricStat *MetricStat `json:"MetricStat,omitempty"`
	Expression string      `json:"Expression,omitempty"`
	Label      string      `json:"Label,omitempty"`
	ReturnData *bool       `json:"ReturnData,omitempty"`
	Period     *int32      `json:"Period,omitempty"`
}

// MetricStat defines the metric and stat to return.
type MetricStat struct {
	Metric Metric `json:"Metric"`
	Period int32  `json:"Period"`
	Stat   string `json:"Stat"`
	Unit   string `json:"Unit,omitempty"`
}

// GetMetricStatisticsRequest is the request for GetMetricStatistics.
type GetMetricStatisticsRequest struct {
	Namespace  string      `json:"Namespace"`
	MetricName string      `json:"MetricName"`
	Dimensions []Dimension `json:"Dimensions,omitempty"`
	StartTime  string      `json:"StartTime"`
	EndTime    string      `json:"EndTime"`
	Period     int32       `json:"Period"`
	Statistics []string    `json:"Statistics,omitempty"`
	Unit       string      `json:"Unit,omitempty"`
}

// ListMetricsRequest is the request for ListMetrics.
type ListMetricsRequest struct {
	Namespace          string            `json:"Namespace,omitempty"`
	MetricName         string            `json:"MetricName,omitempty"`
	Dimensions         []DimensionFilter `json:"Dimensions,omitempty"`
	NextToken          string            `json:"NextToken,omitempty"`
	RecentlyActive     string            `json:"RecentlyActive,omitempty"`
	IncludeLinkedAccts *bool             `json:"IncludeLinkedAccounts,omitempty"`
	OwnAcctOnly        *bool             `json:"OwningAccount,omitempty"`
}

// DimensionFilter is used for filtering metrics by dimension.
type DimensionFilter struct {
	Name  string `json:"Name"`
	Value string `json:"Value,omitempty"`
}

// PutMetricAlarmRequest is the request for PutMetricAlarm.
type PutMetricAlarmRequest struct {
	AlarmName          string      `json:"AlarmName"`
	AlarmDescription   string      `json:"AlarmDescription,omitempty"`
	MetricName         string      `json:"MetricName"`
	Namespace          string      `json:"Namespace"`
	Statistic          string      `json:"Statistic,omitempty"`
	Dimensions         []Dimension `json:"Dimensions,omitempty"`
	Period             int32       `json:"Period"`
	EvaluationPeriods  int32       `json:"EvaluationPeriods"`
	Threshold          float64     `json:"Threshold"`
	ComparisonOperator string      `json:"ComparisonOperator"`
	ActionsEnabled     *bool       `json:"ActionsEnabled,omitempty"`
	AlarmActions       []string    `json:"AlarmActions,omitempty"`
	OKActions          []string    `json:"OKActions,omitempty"`
}

// DeleteAlarmsRequest is the request for DeleteAlarms.
type DeleteAlarmsRequest struct {
	AlarmNames []string `json:"AlarmNames"`
}

// DescribeAlarmsRequest is the request for DescribeAlarms.
type DescribeAlarmsRequest struct {
	AlarmNames      []string `json:"AlarmNames,omitempty"`
	AlarmNamePrefix string   `json:"AlarmNamePrefix,omitempty"`
	StateValue      string   `json:"StateValue,omitempty"`
	ActionPrefix    string   `json:"ActionPrefix,omitempty"`
	MaxRecords      *int32   `json:"MaxRecords,omitempty"`
	NextToken       string   `json:"NextToken,omitempty"`
}

// JSON Response types for CloudWatch JSON protocol.

// GetMetricDataResponse is the response for GetMetricData.
type GetMetricDataResponse struct {
	MetricDataResults []MetricDataResult `json:"MetricDataResults"`
	NextToken         string             `json:"NextToken,omitempty"`
}

// MetricDataResult represents a single metric data result.
type MetricDataResult struct {
	ID         string    `json:"Id"`
	Label      string    `json:"Label"`
	Timestamps []string  `json:"Timestamps"`
	Values     []float64 `json:"Values"`
	StatusCode string    `json:"StatusCode"`
}

// GetMetricStatisticsResponse is the response for GetMetricStatistics.
type GetMetricStatisticsResponse struct {
	Label      string      `json:"Label"`
	Datapoints []Datapoint `json:"Datapoints"`
}

// Datapoint represents a single datapoint.
type Datapoint struct {
	Timestamp   string   `json:"Timestamp"`
	SampleCount *float64 `json:"SampleCount,omitempty"`
	Average     *float64 `json:"Average,omitempty"`
	Sum         *float64 `json:"Sum,omitempty"`
	Minimum     *float64 `json:"Minimum,omitempty"`
	Maximum     *float64 `json:"Maximum,omitempty"`
	Unit        string   `json:"Unit,omitempty"`
}

// ListMetricsResponse is the response for ListMetrics.
type ListMetricsResponse struct {
	Metrics   []Metric `json:"Metrics"`
	NextToken string   `json:"NextToken,omitempty"`
}

// DescribeAlarmsResponse is the response for DescribeAlarms.
type DescribeAlarmsResponse struct {
	MetricAlarms []MetricAlarm `json:"MetricAlarms"`
	NextToken    string        `json:"NextToken,omitempty"`
}

// MetricAlarm represents a single metric alarm in JSON response.
type MetricAlarm struct {
	AlarmName                          string      `json:"AlarmName"`
	AlarmArn                           string      `json:"AlarmArn"`
	AlarmDescription                   string      `json:"AlarmDescription,omitempty"`
	MetricName                         string      `json:"MetricName"`
	Namespace                          string      `json:"Namespace"`
	Statistic                          string      `json:"Statistic,omitempty"`
	Dimensions                         []Dimension `json:"Dimensions,omitempty"`
	Period                             int32       `json:"Period"`
	EvaluationPeriods                  int32       `json:"EvaluationPeriods"`
	Threshold                          float64     `json:"Threshold"`
	ComparisonOperator                 string      `json:"ComparisonOperator"`
	ActionsEnabled                     bool        `json:"ActionsEnabled"`
	AlarmActions                       []string    `json:"AlarmActions,omitempty"`
	OKActions                          []string    `json:"OKActions,omitempty"`
	StateValue                         string      `json:"StateValue"`
	StateReason                        string      `json:"StateReason"`
	StateUpdatedTimestamp              string      `json:"StateUpdatedTimestamp"`
	AlarmConfigurationUpdatedTimestamp string      `json:"AlarmConfigurationUpdatedTimestamp"`
}

// ErrorResponse represents a CloudWatch error response in JSON format.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}

// Error represents a CloudWatch error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// GetMetricDataResult is the result for GetMetricData storage operation.
type GetMetricDataResult struct {
	MetricDataResults []MetricDataResult
	NextToken         string
}

// GetMetricStatisticsResult is the result for GetMetricStatistics storage operation.
type GetMetricStatisticsResult struct {
	Label      string
	Datapoints []Datapoint
}

// ListMetricsResult is the result for ListMetrics storage operation.
type ListMetricsResult struct {
	Metrics   []Metric
	NextToken string
}

// DescribeAlarmsResult is the result for DescribeAlarms storage operation.
type DescribeAlarmsResult struct {
	MetricAlarms []MetricAlarm
	NextToken    string
}

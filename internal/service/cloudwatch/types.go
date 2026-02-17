// Package cloudwatch provides CloudWatch metrics service emulation for awsim.
package cloudwatch

import "encoding/xml"

// Metric represents a CloudWatch metric.
type Metric struct {
	Namespace  string
	MetricName string
	Dimensions []Dimension
}

// Dimension represents a CloudWatch dimension.
type Dimension struct {
	Name  string
	Value string
}

// MetricDatum represents a single metric data point.
type MetricDatum struct {
	MetricName        string
	Dimensions        []Dimension
	Timestamp         string
	Value             *float64
	StatisticValues   *StatisticSet
	Values            []float64
	Counts            []float64
	Unit              string
	StorageResolution *int32
}

// StatisticSet represents a set of statistics.
type StatisticSet struct {
	SampleCount float64
	Sum         float64
	Minimum     float64
	Maximum     float64
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

// XML Response types for CloudWatch Query protocol.

// ResponseMetadata contains common response metadata.
type ResponseMetadata struct {
	RequestID string `xml:"RequestId"`
}

// XMLPutMetricDataResponse is the response for PutMetricData.
type XMLPutMetricDataResponse struct {
	XMLName          xml.Name         `xml:"PutMetricDataResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	ResponseMetadata ResponseMetadata `xml:"ResponseMetadata"`
}

// XMLGetMetricDataResponse is the response for GetMetricData.
type XMLGetMetricDataResponse struct {
	XMLName             xml.Name               `xml:"GetMetricDataResponse"`
	Xmlns               string                 `xml:"xmlns,attr"`
	GetMetricDataResult XMLGetMetricDataResult `xml:"GetMetricDataResult"`
	ResponseMetadata    ResponseMetadata       `xml:"ResponseMetadata"`
}

// XMLGetMetricDataResult contains the GetMetricData result.
type XMLGetMetricDataResult struct {
	MetricDataResults XMLMetricDataResults `xml:"MetricDataResults"`
	NextToken         string               `xml:"NextToken,omitempty"`
}

// XMLMetricDataResults contains metric data results.
type XMLMetricDataResults struct {
	Member []XMLMetricDataResult `xml:"member"`
}

// XMLMetricDataResult represents a single metric data result.
type XMLMetricDataResult struct {
	ID         string        `xml:"Id"`
	Label      string        `xml:"Label"`
	Timestamps XMLTimestamps `xml:"Timestamps"`
	Values     XMLValues     `xml:"Values"`
	StatusCode string        `xml:"StatusCode"`
}

// XMLTimestamps contains timestamp members.
type XMLTimestamps struct {
	Member []string `xml:"member"`
}

// XMLValues contains value members.
type XMLValues struct {
	Member []float64 `xml:"member"`
}

// XMLGetMetricStatisticsResponse is the response for GetMetricStatistics.
type XMLGetMetricStatisticsResponse struct {
	XMLName                   xml.Name                     `xml:"GetMetricStatisticsResponse"`
	Xmlns                     string                       `xml:"xmlns,attr"`
	GetMetricStatisticsResult XMLGetMetricStatisticsResult `xml:"GetMetricStatisticsResult"`
	ResponseMetadata          ResponseMetadata             `xml:"ResponseMetadata"`
}

// XMLGetMetricStatisticsResult contains the GetMetricStatistics result.
type XMLGetMetricStatisticsResult struct {
	Label      string        `xml:"Label"`
	Datapoints XMLDatapoints `xml:"Datapoints"`
}

// XMLDatapoints contains datapoint members.
type XMLDatapoints struct {
	Member []XMLDatapoint `xml:"member"`
}

// XMLDatapoint represents a single datapoint.
type XMLDatapoint struct {
	Timestamp   string  `xml:"Timestamp"`
	SampleCount float64 `xml:"SampleCount,omitempty"`
	Average     float64 `xml:"Average,omitempty"`
	Sum         float64 `xml:"Sum,omitempty"`
	Minimum     float64 `xml:"Minimum,omitempty"`
	Maximum     float64 `xml:"Maximum,omitempty"`
	Unit        string  `xml:"Unit,omitempty"`
}

// XMLListMetricsResponse is the response for ListMetrics.
type XMLListMetricsResponse struct {
	XMLName           xml.Name             `xml:"ListMetricsResponse"`
	Xmlns             string               `xml:"xmlns,attr"`
	ListMetricsResult XMLListMetricsResult `xml:"ListMetricsResult"`
	ResponseMetadata  ResponseMetadata     `xml:"ResponseMetadata"`
}

// XMLListMetricsResult contains the ListMetrics result.
type XMLListMetricsResult struct {
	Metrics   XMLMetrics `xml:"Metrics"`
	NextToken string     `xml:"NextToken,omitempty"`
}

// XMLMetrics contains metric members.
type XMLMetrics struct {
	Member []XMLMetric `xml:"member"`
}

// XMLMetric represents a single metric.
type XMLMetric struct {
	Namespace  string        `xml:"Namespace"`
	MetricName string        `xml:"MetricName"`
	Dimensions XMLDimensions `xml:"Dimensions"`
}

// XMLDimensions contains dimension members.
type XMLDimensions struct {
	Member []XMLDimension `xml:"member"`
}

// XMLDimension represents a single dimension.
type XMLDimension struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
}

// XMLPutMetricAlarmResponse is the response for PutMetricAlarm.
type XMLPutMetricAlarmResponse struct {
	XMLName          xml.Name         `xml:"PutMetricAlarmResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	ResponseMetadata ResponseMetadata `xml:"ResponseMetadata"`
}

// XMLDeleteAlarmsResponse is the response for DeleteAlarms.
type XMLDeleteAlarmsResponse struct {
	XMLName          xml.Name         `xml:"DeleteAlarmsResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	ResponseMetadata ResponseMetadata `xml:"ResponseMetadata"`
}

// XMLDescribeAlarmsResponse is the response for DescribeAlarms.
type XMLDescribeAlarmsResponse struct {
	XMLName              xml.Name                `xml:"DescribeAlarmsResponse"`
	Xmlns                string                  `xml:"xmlns,attr"`
	DescribeAlarmsResult XMLDescribeAlarmsResult `xml:"DescribeAlarmsResult"`
	ResponseMetadata     ResponseMetadata        `xml:"ResponseMetadata"`
}

// XMLDescribeAlarmsResult contains the DescribeAlarms result.
type XMLDescribeAlarmsResult struct {
	MetricAlarms XMLMetricAlarms `xml:"MetricAlarms"`
	NextToken    string          `xml:"NextToken,omitempty"`
}

// XMLMetricAlarms contains metric alarm members.
type XMLMetricAlarms struct {
	Member []XMLMetricAlarm `xml:"member"`
}

// XMLMetricAlarm represents a single metric alarm.
type XMLMetricAlarm struct {
	AlarmName                          string        `xml:"AlarmName"`
	AlarmArn                           string        `xml:"AlarmArn"`
	AlarmDescription                   string        `xml:"AlarmDescription,omitempty"`
	MetricName                         string        `xml:"MetricName"`
	Namespace                          string        `xml:"Namespace"`
	Statistic                          string        `xml:"Statistic,omitempty"`
	Dimensions                         XMLDimensions `xml:"Dimensions"`
	Period                             int32         `xml:"Period"`
	EvaluationPeriods                  int32         `xml:"EvaluationPeriods"`
	Threshold                          float64       `xml:"Threshold"`
	ComparisonOperator                 string        `xml:"ComparisonOperator"`
	ActionsEnabled                     bool          `xml:"ActionsEnabled"`
	AlarmActions                       XMLActions    `xml:"AlarmActions"`
	OKActions                          XMLActions    `xml:"OKActions"`
	StateValue                         string        `xml:"StateValue"`
	StateReason                        string        `xml:"StateReason"`
	StateUpdatedTimestamp              string        `xml:"StateUpdatedTimestamp"`
	AlarmConfigurationUpdatedTimestamp string        `xml:"AlarmConfigurationUpdatedTimestamp"`
}

// XMLActions contains action members.
type XMLActions struct {
	Member []string `xml:"member"`
}

// XMLErrorResponse represents a CloudWatch error response.
type XMLErrorResponse struct {
	XMLName   xml.Name       `xml:"ErrorResponse"`
	Xmlns     string         `xml:"xmlns,attr"`
	Error     XMLErrorDetail `xml:"Error"`
	RequestID string         `xml:"RequestId"`
}

// XMLErrorDetail contains error details.
type XMLErrorDetail struct {
	Type    string `xml:"Type"`
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
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

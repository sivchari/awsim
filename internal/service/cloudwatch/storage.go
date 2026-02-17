package cloudwatch

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

// Storage defines the CloudWatch storage interface.
type Storage interface {
	PutMetricData(ctx context.Context, namespace string, metricData []MetricDatum) error
	GetMetricData(ctx context.Context, req *GetMetricDataRequest) (*XMLGetMetricDataResult, error)
	GetMetricStatistics(ctx context.Context, req *GetMetricStatisticsRequest) (*XMLGetMetricStatisticsResult, error)
	ListMetrics(ctx context.Context, req *ListMetricsRequest) (*XMLListMetricsResult, error)
	PutMetricAlarm(ctx context.Context, req *PutMetricAlarmRequest) error
	DeleteAlarms(ctx context.Context, alarmNames []string) error
	DescribeAlarms(ctx context.Context, req *DescribeAlarmsRequest) (*XMLDescribeAlarmsResult, error)
}

// metricKey uniquely identifies a metric.
type metricKey struct {
	namespace  string
	metricName string
	dimensions string // sorted dimension string for consistency
}

// storedMetric holds metric data in memory.
type storedMetric struct {
	namespace  string
	metricName string
	dimensions []Dimension
	datapoints []MetricDatapoint
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu      sync.RWMutex
	metrics map[metricKey]*storedMetric
	alarms  map[string]*Alarm
	baseURL string
}

// NewMemoryStorage creates a new in-memory CloudWatch storage.
func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		metrics: make(map[metricKey]*storedMetric),
		alarms:  make(map[string]*Alarm),
		baseURL: baseURL,
	}
}

// PutMetricData stores metric data.
func (s *MemoryStorage) PutMetricData(_ context.Context, namespace string, metricData []MetricDatum) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range metricData {
		datum := &metricData[i]
		key := s.makeMetricKey(namespace, datum.MetricName, datum.Dimensions)

		metric, exists := s.metrics[key]
		if !exists {
			metric = &storedMetric{
				namespace:  namespace,
				metricName: datum.MetricName,
				dimensions: datum.Dimensions,
				datapoints: make([]MetricDatapoint, 0),
			}
			s.metrics[key] = metric
		}

		timestamp := datum.Timestamp
		if timestamp == "" {
			timestamp = time.Now().UTC().Format(time.RFC3339)
		}

		s.appendDatapoints(metric, datum, timestamp)
	}

	return nil
}

// appendDatapoints adds datapoints to the metric based on the datum type.
func (s *MemoryStorage) appendDatapoints(metric *storedMetric, datum *MetricDatum, timestamp string) {
	switch {
	case datum.Value != nil:
		metric.datapoints = append(metric.datapoints, MetricDatapoint{
			Timestamp: timestamp,
			Value:     *datum.Value,
			Unit:      datum.Unit,
		})
	case datum.StatisticValues != nil:
		metric.datapoints = append(metric.datapoints, MetricDatapoint{
			Timestamp: timestamp,
			Value:     datum.StatisticValues.Sum / datum.StatisticValues.SampleCount,
			Unit:      datum.Unit,
		})
	case len(datum.Values) > 0:
		for _, v := range datum.Values {
			metric.datapoints = append(metric.datapoints, MetricDatapoint{
				Timestamp: timestamp,
				Value:     v,
				Unit:      datum.Unit,
			})
		}
	}
}

// GetMetricData retrieves metric data.
func (s *MemoryStorage) GetMetricData(_ context.Context, req *GetMetricDataRequest) (*XMLGetMetricDataResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]XMLMetricDataResult, 0, len(req.MetricDataQueries))

	for _, query := range req.MetricDataQueries {
		if query.MetricStat == nil {
			continue
		}

		key := s.makeMetricKey(
			query.MetricStat.Metric.Namespace,
			query.MetricStat.Metric.MetricName,
			query.MetricStat.Metric.Dimensions,
		)

		metric, exists := s.metrics[key]
		if !exists {
			results = append(results, XMLMetricDataResult{
				ID:         query.ID,
				Label:      query.Label,
				StatusCode: "Complete",
			})

			continue
		}

		// Filter datapoints by time range
		timestamps := make([]string, 0)
		values := make([]float64, 0)

		for _, dp := range metric.datapoints {
			if s.isInTimeRange(dp.Timestamp, req.StartTime, req.EndTime) {
				timestamps = append(timestamps, dp.Timestamp)
				values = append(values, dp.Value)
			}
		}

		label := query.Label
		if label == "" {
			label = query.MetricStat.Metric.MetricName
		}

		results = append(results, XMLMetricDataResult{
			ID:    query.ID,
			Label: label,
			Timestamps: XMLTimestamps{
				Member: timestamps,
			},
			Values: XMLValues{
				Member: values,
			},
			StatusCode: "Complete",
		})
	}

	return &XMLGetMetricDataResult{
		MetricDataResults: XMLMetricDataResults{
			Member: results,
		},
	}, nil
}

// GetMetricStatistics retrieves statistics for a metric.
func (s *MemoryStorage) GetMetricStatistics(_ context.Context, req *GetMetricStatisticsRequest) (*XMLGetMetricStatisticsResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeMetricKey(req.Namespace, req.MetricName, req.Dimensions)

	metric, exists := s.metrics[key]
	if !exists {
		return &XMLGetMetricStatisticsResult{
			Label:      req.MetricName,
			Datapoints: XMLDatapoints{Member: []XMLDatapoint{}},
		}, nil
	}

	// Collect datapoints in time range
	var filteredPoints []MetricDatapoint

	for _, dp := range metric.datapoints {
		if s.isInTimeRange(dp.Timestamp, req.StartTime, req.EndTime) {
			filteredPoints = append(filteredPoints, dp)
		}
	}

	if len(filteredPoints) == 0 {
		return &XMLGetMetricStatisticsResult{
			Label:      req.MetricName,
			Datapoints: XMLDatapoints{Member: []XMLDatapoint{}},
		}, nil
	}

	// Calculate statistics
	datapoints := s.calculateStatistics(filteredPoints, req.Statistics, req.Period)

	return &XMLGetMetricStatisticsResult{
		Label: req.MetricName,
		Datapoints: XMLDatapoints{
			Member: datapoints,
		},
	}, nil
}

// ListMetrics lists available metrics.
func (s *MemoryStorage) ListMetrics(_ context.Context, req *ListMetricsRequest) (*XMLListMetricsResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := make([]XMLMetric, 0)

	for _, m := range s.metrics {
		// Filter by namespace
		if req.Namespace != "" && m.namespace != req.Namespace {
			continue
		}

		// Filter by metric name
		if req.MetricName != "" && m.metricName != req.MetricName {
			continue
		}

		// Filter by dimensions
		if !s.matchesDimensionFilters(m.dimensions, req.Dimensions) {
			continue
		}

		dims := convertDimensionsToXML(m.dimensions)
		metrics = append(metrics, XMLMetric{
			Namespace:  m.namespace,
			MetricName: m.metricName,
			Dimensions: XMLDimensions{Member: dims},
		})
	}

	// Sort metrics for consistent output
	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].Namespace != metrics[j].Namespace {
			return metrics[i].Namespace < metrics[j].Namespace
		}

		return metrics[i].MetricName < metrics[j].MetricName
	})

	return &XMLListMetricsResult{
		Metrics: XMLMetrics{Member: metrics},
	}, nil
}

// PutMetricAlarm creates or updates an alarm.
func (s *MemoryStorage) PutMetricAlarm(_ context.Context, req *PutMetricAlarmRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC().Format(time.RFC3339)
	alarmARN := fmt.Sprintf("arn:aws:cloudwatch:us-east-1:000000000000:alarm:%s", req.AlarmName)

	actionsEnabled := true
	if req.ActionsEnabled != nil {
		actionsEnabled = *req.ActionsEnabled
	}

	alarm := &Alarm{
		AlarmName:          req.AlarmName,
		AlarmARN:           alarmARN,
		AlarmDescription:   req.AlarmDescription,
		MetricName:         req.MetricName,
		Namespace:          req.Namespace,
		Statistic:          req.Statistic,
		Dimensions:         req.Dimensions,
		Period:             req.Period,
		EvaluationPeriods:  req.EvaluationPeriods,
		Threshold:          req.Threshold,
		ComparisonOperator: req.ComparisonOperator,
		ActionsEnabled:     actionsEnabled,
		AlarmActions:       req.AlarmActions,
		OKActions:          req.OKActions,
		StateValue:         "INSUFFICIENT_DATA",
		StateReason:        "Unchecked: Initial alarm creation",
		StateUpdatedAt:     now,
		CreatedAt:          now,
	}

	s.alarms[req.AlarmName] = alarm

	return nil
}

// DeleteAlarms deletes the specified alarms.
func (s *MemoryStorage) DeleteAlarms(_ context.Context, alarmNames []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, name := range alarmNames {
		if _, exists := s.alarms[name]; !exists {
			return &Error{
				Code:    "ResourceNotFound",
				Message: fmt.Sprintf("Alarm %s does not exist", name),
			}
		}
	}

	for _, name := range alarmNames {
		delete(s.alarms, name)
	}

	return nil
}

// DescribeAlarms returns information about alarms.
func (s *MemoryStorage) DescribeAlarms(_ context.Context, req *DescribeAlarmsRequest) (*XMLDescribeAlarmsResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alarms := make([]XMLMetricAlarm, 0)

	for _, alarm := range s.alarms {
		if !s.alarmMatchesFilter(alarm, req) {
			continue
		}

		alarms = append(alarms, convertAlarmToXML(alarm))
	}

	// Sort alarms by name
	sort.Slice(alarms, func(i, j int) bool {
		return alarms[i].AlarmName < alarms[j].AlarmName
	})

	// Apply MaxRecords limit
	maxRecords := 50
	if req.MaxRecords != nil && *req.MaxRecords > 0 {
		maxRecords = int(*req.MaxRecords)
	}

	if len(alarms) > maxRecords {
		alarms = alarms[:maxRecords]
	}

	return &XMLDescribeAlarmsResult{
		MetricAlarms: XMLMetricAlarms{Member: alarms},
	}, nil
}

// alarmMatchesFilter checks if an alarm matches the filter criteria.
func (s *MemoryStorage) alarmMatchesFilter(alarm *Alarm, req *DescribeAlarmsRequest) bool {
	if len(req.AlarmNames) > 0 && !slices.Contains(req.AlarmNames, alarm.AlarmName) {
		return false
	}

	if req.AlarmNamePrefix != "" && !strings.HasPrefix(alarm.AlarmName, req.AlarmNamePrefix) {
		return false
	}

	if req.StateValue != "" && alarm.StateValue != req.StateValue {
		return false
	}

	return true
}

// convertAlarmToXML converts an Alarm to XMLMetricAlarm.
func convertAlarmToXML(alarm *Alarm) XMLMetricAlarm {
	dims := convertDimensionsToXML(alarm.Dimensions)

	return XMLMetricAlarm{
		AlarmName:                          alarm.AlarmName,
		AlarmArn:                           alarm.AlarmARN,
		AlarmDescription:                   alarm.AlarmDescription,
		MetricName:                         alarm.MetricName,
		Namespace:                          alarm.Namespace,
		Statistic:                          alarm.Statistic,
		Dimensions:                         XMLDimensions{Member: dims},
		Period:                             alarm.Period,
		EvaluationPeriods:                  alarm.EvaluationPeriods,
		Threshold:                          alarm.Threshold,
		ComparisonOperator:                 alarm.ComparisonOperator,
		ActionsEnabled:                     alarm.ActionsEnabled,
		AlarmActions:                       XMLActions{Member: alarm.AlarmActions},
		OKActions:                          XMLActions{Member: alarm.OKActions},
		StateValue:                         alarm.StateValue,
		StateReason:                        alarm.StateReason,
		StateUpdatedTimestamp:              alarm.StateUpdatedAt,
		AlarmConfigurationUpdatedTimestamp: alarm.CreatedAt,
	}
}

// makeMetricKey creates a unique key for a metric.
func (s *MemoryStorage) makeMetricKey(namespace, metricName string, dimensions []Dimension) metricKey {
	// Sort dimensions for consistent key generation
	sorted := make([]Dimension, len(dimensions))
	copy(sorted, dimensions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	dimParts := make([]string, 0, len(sorted))

	for _, d := range sorted {
		dimParts = append(dimParts, fmt.Sprintf("%s=%s", d.Name, d.Value))
	}

	return metricKey{
		namespace:  namespace,
		metricName: metricName,
		dimensions: strings.Join(dimParts, ","),
	}
}

// isInTimeRange checks if a timestamp is within the given range.
func (s *MemoryStorage) isInTimeRange(timestamp, startTime, endTime string) bool {
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return false
	}

	if startTime != "" {
		start, err := time.Parse(time.RFC3339, startTime)
		if err == nil && ts.Before(start) {
			return false
		}
	}

	if endTime != "" {
		end, err := time.Parse(time.RFC3339, endTime)
		if err == nil && ts.After(end) {
			return false
		}
	}

	return true
}

// matchesDimensionFilters checks if dimensions match the filters.
func (s *MemoryStorage) matchesDimensionFilters(dimensions []Dimension, filters []DimensionFilter) bool {
	if len(filters) == 0 {
		return true
	}

	dimMap := make(map[string]string)
	for _, d := range dimensions {
		dimMap[d.Name] = d.Value
	}

	for _, f := range filters {
		val, exists := dimMap[f.Name]
		if !exists {
			return false
		}

		if f.Value != "" && val != f.Value {
			return false
		}
	}

	return true
}

// calculateStatistics calculates statistics for datapoints.
func (s *MemoryStorage) calculateStatistics(points []MetricDatapoint, statistics []string, _ int32) []XMLDatapoint {
	if len(points) == 0 {
		return nil
	}

	// Group by period (simplified: use first timestamp)
	timestamp := points[0].Timestamp
	unit := points[0].Unit
	count := float64(len(points))

	var sum float64

	minVal := points[0].Value
	maxVal := points[0].Value

	for _, p := range points {
		sum += p.Value

		if p.Value < minVal {
			minVal = p.Value
		}

		if p.Value > maxVal {
			maxVal = p.Value
		}
	}

	dp := XMLDatapoint{
		Timestamp: timestamp,
		Unit:      unit,
	}

	// Include requested statistics
	for _, stat := range statistics {
		switch stat {
		case "SampleCount":
			dp.SampleCount = count
		case "Average":
			dp.Average = sum / count
		case "Sum":
			dp.Sum = sum
		case "Minimum":
			dp.Minimum = minVal
		case "Maximum":
			dp.Maximum = maxVal
		}
	}

	// If no specific statistics requested, include all
	if len(statistics) == 0 {
		dp.SampleCount = count
		dp.Average = sum / count
		dp.Sum = sum
		dp.Minimum = minVal
		dp.Maximum = maxVal
	}

	return []XMLDatapoint{dp}
}

// convertDimensionsToXML converts Dimension slice to XMLDimension slice.
func convertDimensionsToXML(dimensions []Dimension) []XMLDimension {
	result := make([]XMLDimension, 0, len(dimensions))

	for _, d := range dimensions {
		result = append(result, XMLDimension(d))
	}

	return result
}

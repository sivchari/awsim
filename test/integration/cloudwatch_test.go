//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func newCloudWatchClient(t *testing.T) *cloudwatch.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	return cloudwatch.NewFromConfig(cfg, func(o *cloudwatch.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestCloudWatch_PutMetricData(t *testing.T) {
	client := newCloudWatchClient(t)
	ctx := t.Context()

	// Put metric data.
	_, err := client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String("TestNamespace"),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("TestMetric"),
				Value:      aws.Float64(100.0),
				Unit:       types.StandardUnitCount,
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("Environment"),
						Value: aws.String("Test"),
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to put metric data: %v", err)
	}

	t.Log("Successfully put metric data")
}

func TestCloudWatch_ListMetrics(t *testing.T) {
	client := newCloudWatchClient(t)
	ctx := t.Context()

	namespace := "TestListMetrics"
	metricName := "ListTestMetric"

	// Put metric data first.
	_, err := client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String(metricName),
				Value:      aws.Float64(42.0),
				Unit:       types.StandardUnitCount,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to put metric data: %v", err)
	}

	// List metrics.
	output, err := client.ListMetrics(ctx, &cloudwatch.ListMetricsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(metricName),
	})
	if err != nil {
		t.Fatalf("failed to list metrics: %v", err)
	}

	if len(output.Metrics) == 0 {
		t.Fatal("expected at least one metric, got none")
	}

	found := false

	for _, m := range output.Metrics {
		if *m.Namespace == namespace && *m.MetricName == metricName {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("metric %s/%s not found in list", namespace, metricName)
	}
}

func TestCloudWatch_GetMetricStatistics(t *testing.T) {
	client := newCloudWatchClient(t)
	ctx := t.Context()

	namespace := "TestGetStats"
	metricName := "StatsTestMetric"

	// Put some metric data.
	_, err := client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String(metricName),
				Value:      aws.Float64(10.0),
				Unit:       types.StandardUnitCount,
			},
			{
				MetricName: aws.String(metricName),
				Value:      aws.Float64(20.0),
				Unit:       types.StandardUnitCount,
			},
			{
				MetricName: aws.String(metricName),
				Value:      aws.Float64(30.0),
				Unit:       types.StandardUnitCount,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to put metric data: %v", err)
	}

	// Get metric statistics.
	now := time.Now()
	output, err := client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(metricName),
		StartTime:  aws.Time(now.Add(-1 * time.Hour)),
		EndTime:    aws.Time(now.Add(1 * time.Hour)),
		Period:     aws.Int32(60),
		Statistics: []types.Statistic{
			types.StatisticSum,
			types.StatisticAverage,
			types.StatisticMinimum,
			types.StatisticMaximum,
			types.StatisticSampleCount,
		},
	})
	if err != nil {
		t.Fatalf("failed to get metric statistics: %v", err)
	}

	if len(output.Datapoints) == 0 {
		t.Fatal("expected at least one datapoint, got none")
	}

	dp := output.Datapoints[0]
	t.Logf("Datapoint: Sum=%.2f, Average=%.2f, Min=%.2f, Max=%.2f, Count=%.0f",
		*dp.Sum, *dp.Average, *dp.Minimum, *dp.Maximum, *dp.SampleCount)
}

func TestCloudWatch_PutMetricAlarm(t *testing.T) {
	client := newCloudWatchClient(t)
	ctx := t.Context()

	alarmName := "test-alarm"

	// Put metric alarm.
	_, err := client.PutMetricAlarm(ctx, &cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(alarmName),
		MetricName:         aws.String("TestMetric"),
		Namespace:          aws.String("TestNamespace"),
		Statistic:          types.StatisticAverage,
		Period:             aws.Int32(60),
		EvaluationPeriods:  aws.Int32(1),
		Threshold:          aws.Float64(80.0),
		ComparisonOperator: types.ComparisonOperatorGreaterThanThreshold,
		ActionsEnabled:     aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to put metric alarm: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteAlarms(ctx, &cloudwatch.DeleteAlarmsInput{
			AlarmNames: []string{alarmName},
		})
	})

	t.Log("Successfully put metric alarm")
}

func TestCloudWatch_DescribeAlarms(t *testing.T) {
	client := newCloudWatchClient(t)
	ctx := t.Context()

	alarmName := "test-describe-alarm"

	// Put metric alarm.
	_, err := client.PutMetricAlarm(ctx, &cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(alarmName),
		MetricName:         aws.String("DescribeTestMetric"),
		Namespace:          aws.String("TestNamespace"),
		Statistic:          types.StatisticAverage,
		Period:             aws.Int32(60),
		EvaluationPeriods:  aws.Int32(1),
		Threshold:          aws.Float64(90.0),
		ComparisonOperator: types.ComparisonOperatorGreaterThanOrEqualToThreshold,
	})
	if err != nil {
		t.Fatalf("failed to put metric alarm: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteAlarms(ctx, &cloudwatch.DeleteAlarmsInput{
			AlarmNames: []string{alarmName},
		})
	})

	// Describe alarms.
	output, err := client.DescribeAlarms(ctx, &cloudwatch.DescribeAlarmsInput{
		AlarmNames: []string{alarmName},
	})
	if err != nil {
		t.Fatalf("failed to describe alarms: %v", err)
	}

	if len(output.MetricAlarms) == 0 {
		t.Fatal("expected at least one alarm, got none")
	}

	alarm := output.MetricAlarms[0]

	if *alarm.AlarmName != alarmName {
		t.Errorf("alarm name mismatch: got %s, want %s", *alarm.AlarmName, alarmName)
	}

	if *alarm.Threshold != 90.0 {
		t.Errorf("threshold mismatch: got %.2f, want 90.0", *alarm.Threshold)
	}
}

func TestCloudWatch_DeleteAlarms(t *testing.T) {
	client := newCloudWatchClient(t)
	ctx := t.Context()

	alarmName := "test-delete-alarm"

	// Put metric alarm.
	_, err := client.PutMetricAlarm(ctx, &cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(alarmName),
		MetricName:         aws.String("DeleteTestMetric"),
		Namespace:          aws.String("TestNamespace"),
		Statistic:          types.StatisticAverage,
		Period:             aws.Int32(60),
		EvaluationPeriods:  aws.Int32(1),
		Threshold:          aws.Float64(50.0),
		ComparisonOperator: types.ComparisonOperatorLessThanThreshold,
	})
	if err != nil {
		t.Fatalf("failed to put metric alarm: %v", err)
	}

	// Delete alarm.
	_, err = client.DeleteAlarms(ctx, &cloudwatch.DeleteAlarmsInput{
		AlarmNames: []string{alarmName},
	})
	if err != nil {
		t.Fatalf("failed to delete alarm: %v", err)
	}

	// Verify alarm is deleted.
	output, err := client.DescribeAlarms(ctx, &cloudwatch.DescribeAlarmsInput{
		AlarmNames: []string{alarmName},
	})
	if err != nil {
		t.Fatalf("failed to describe alarms: %v", err)
	}

	if len(output.MetricAlarms) != 0 {
		t.Errorf("expected alarm to be deleted, but found %d alarms", len(output.MetricAlarms))
	}
}

func TestCloudWatch_GetMetricData(t *testing.T) {
	client := newCloudWatchClient(t)
	ctx := t.Context()

	namespace := "TestGetMetricData"
	metricName := "DataTestMetric"

	// Put metric data.
	_, err := client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String(metricName),
				Value:      aws.Float64(100.0),
				Unit:       types.StandardUnitCount,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to put metric data: %v", err)
	}

	// Get metric data.
	now := time.Now()
	output, err := client.GetMetricData(ctx, &cloudwatch.GetMetricDataInput{
		StartTime: aws.Time(now.Add(-1 * time.Hour)),
		EndTime:   aws.Time(now.Add(1 * time.Hour)),
		MetricDataQueries: []types.MetricDataQuery{
			{
				Id: aws.String("m1"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String(namespace),
						MetricName: aws.String(metricName),
					},
					Period: aws.Int32(60),
					Stat:   aws.String("Average"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to get metric data: %v", err)
	}

	if len(output.MetricDataResults) == 0 {
		t.Fatal("expected at least one metric data result, got none")
	}

	result := output.MetricDataResults[0]

	if *result.Id != "m1" {
		t.Errorf("result ID mismatch: got %s, want m1", *result.Id)
	}

	t.Logf("Got metric data result: ID=%s, Label=%s, Values=%v",
		*result.Id, *result.Label, result.Values)
}

func TestCloudWatch_PutMetricDataWithDimensions(t *testing.T) {
	client := newCloudWatchClient(t)
	ctx := t.Context()

	namespace := "TestDimensions"
	metricName := "DimensionMetric"

	// Put metric data with dimensions.
	_, err := client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String(metricName),
				Value:      aws.Float64(50.0),
				Unit:       types.StandardUnitPercent,
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("InstanceId"),
						Value: aws.String("i-12345"),
					},
					{
						Name:  aws.String("InstanceType"),
						Value: aws.String("t2.micro"),
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to put metric data with dimensions: %v", err)
	}

	// List metrics with dimension filter.
	output, err := client.ListMetrics(ctx, &cloudwatch.ListMetricsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(metricName),
		Dimensions: []types.DimensionFilter{
			{
				Name:  aws.String("InstanceId"),
				Value: aws.String("i-12345"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to list metrics: %v", err)
	}

	if len(output.Metrics) == 0 {
		t.Fatal("expected at least one metric with matching dimensions, got none")
	}
}

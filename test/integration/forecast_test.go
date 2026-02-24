//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/forecast"
	"github.com/aws/aws-sdk-go-v2/service/forecast/types"
	"github.com/stretchr/testify/require"
)

func newForecastClient(t *testing.T) *forecast.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	return forecast.NewFromConfig(cfg, func(o *forecast.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestForecast_CreateAndDeleteDataset(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	datasetName := "test-dataset-create-delete"

	// Create dataset.
	createOutput, err := client.CreateDataset(ctx, &forecast.CreateDatasetInput{
		DatasetName: aws.String(datasetName),
		DatasetType: types.DatasetTypeTargetTimeSeries,
		Domain:      types.DomainRetail,
		Schema: &types.Schema{
			Attributes: []types.SchemaAttribute{
				{
					AttributeName: aws.String("item_id"),
					AttributeType: types.AttributeTypeString,
				},
				{
					AttributeName: aws.String("timestamp"),
					AttributeType: types.AttributeTypeTimestamp,
				},
				{
					AttributeName: aws.String("demand"),
					AttributeType: types.AttributeTypeFloat,
				},
			},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.DatasetArn)

	datasetArn := createOutput.DatasetArn

	t.Cleanup(func() {
		_, _ = client.DeleteDataset(ctx, &forecast.DeleteDatasetInput{
			DatasetArn: datasetArn,
		})
	})

	// Describe dataset.
	descOutput, err := client.DescribeDataset(ctx, &forecast.DescribeDatasetInput{
		DatasetArn: datasetArn,
	})
	require.NoError(t, err)
	require.Equal(t, datasetName, *descOutput.DatasetName)
	require.Equal(t, types.DatasetTypeTargetTimeSeries, descOutput.DatasetType)
	require.Equal(t, types.DomainRetail, descOutput.Domain)

	// Delete dataset.
	_, err = client.DeleteDataset(ctx, &forecast.DeleteDatasetInput{
		DatasetArn: datasetArn,
	})
	require.NoError(t, err)

	// Verify dataset is deleted.
	_, err = client.DescribeDataset(ctx, &forecast.DescribeDatasetInput{
		DatasetArn: datasetArn,
	})
	require.Error(t, err)
}

func TestForecast_ListDatasets(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	// Create datasets.
	var datasetArns []*string

	for i := 0; i < 2; i++ {
		createOutput, err := client.CreateDataset(ctx, &forecast.CreateDatasetInput{
			DatasetName: aws.String("test-dataset-list-" + string(rune('a'+i))),
			DatasetType: types.DatasetTypeTargetTimeSeries,
			Domain:      types.DomainRetail,
			Schema: &types.Schema{
				Attributes: []types.SchemaAttribute{
					{
						AttributeName: aws.String("item_id"),
						AttributeType: types.AttributeTypeString,
					},
					{
						AttributeName: aws.String("timestamp"),
						AttributeType: types.AttributeTypeTimestamp,
					},
					{
						AttributeName: aws.String("demand"),
						AttributeType: types.AttributeTypeFloat,
					},
				},
			},
		})
		require.NoError(t, err)
		datasetArns = append(datasetArns, createOutput.DatasetArn)
	}

	t.Cleanup(func() {
		for _, arn := range datasetArns {
			_, _ = client.DeleteDataset(ctx, &forecast.DeleteDatasetInput{
				DatasetArn: arn,
			})
		}
	})

	// List datasets.
	listOutput, err := client.ListDatasets(ctx, &forecast.ListDatasetsInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.Datasets), 2)
}

func TestForecast_CreateAndDeleteDatasetGroup(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	datasetGroupName := "test-dataset-group"

	// Create dataset group.
	createOutput, err := client.CreateDatasetGroup(ctx, &forecast.CreateDatasetGroupInput{
		DatasetGroupName: aws.String(datasetGroupName),
		Domain:           types.DomainRetail,
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.DatasetGroupArn)

	datasetGroupArn := createOutput.DatasetGroupArn

	t.Cleanup(func() {
		_, _ = client.DeleteDatasetGroup(ctx, &forecast.DeleteDatasetGroupInput{
			DatasetGroupArn: datasetGroupArn,
		})
	})

	// Describe dataset group.
	descOutput, err := client.DescribeDatasetGroup(ctx, &forecast.DescribeDatasetGroupInput{
		DatasetGroupArn: datasetGroupArn,
	})
	require.NoError(t, err)
	require.Equal(t, datasetGroupName, *descOutput.DatasetGroupName)
	require.Equal(t, types.DomainRetail, descOutput.Domain)

	// Delete dataset group.
	_, err = client.DeleteDatasetGroup(ctx, &forecast.DeleteDatasetGroupInput{
		DatasetGroupArn: datasetGroupArn,
	})
	require.NoError(t, err)

	// Verify dataset group is deleted.
	_, err = client.DescribeDatasetGroup(ctx, &forecast.DescribeDatasetGroupInput{
		DatasetGroupArn: datasetGroupArn,
	})
	require.Error(t, err)
}

func TestForecast_UpdateDatasetGroup(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	// Create a dataset.
	datasetOutput, err := client.CreateDataset(ctx, &forecast.CreateDatasetInput{
		DatasetName: aws.String("test-dataset-for-update-group"),
		DatasetType: types.DatasetTypeTargetTimeSeries,
		Domain:      types.DomainRetail,
		Schema: &types.Schema{
			Attributes: []types.SchemaAttribute{
				{
					AttributeName: aws.String("item_id"),
					AttributeType: types.AttributeTypeString,
				},
				{
					AttributeName: aws.String("timestamp"),
					AttributeType: types.AttributeTypeTimestamp,
				},
				{
					AttributeName: aws.String("demand"),
					AttributeType: types.AttributeTypeFloat,
				},
			},
		},
	})
	require.NoError(t, err)

	datasetArn := datasetOutput.DatasetArn

	t.Cleanup(func() {
		_, _ = client.DeleteDataset(ctx, &forecast.DeleteDatasetInput{
			DatasetArn: datasetArn,
		})
	})

	// Create a dataset group.
	dgOutput, err := client.CreateDatasetGroup(ctx, &forecast.CreateDatasetGroupInput{
		DatasetGroupName: aws.String("test-dataset-group-update"),
		Domain:           types.DomainRetail,
	})
	require.NoError(t, err)

	datasetGroupArn := dgOutput.DatasetGroupArn

	t.Cleanup(func() {
		_, _ = client.DeleteDatasetGroup(ctx, &forecast.DeleteDatasetGroupInput{
			DatasetGroupArn: datasetGroupArn,
		})
	})

	// Update dataset group with dataset.
	_, err = client.UpdateDatasetGroup(ctx, &forecast.UpdateDatasetGroupInput{
		DatasetGroupArn: datasetGroupArn,
		DatasetArns:     []string{*datasetArn},
	})
	require.NoError(t, err)

	// Verify the update.
	descOutput, err := client.DescribeDatasetGroup(ctx, &forecast.DescribeDatasetGroupInput{
		DatasetGroupArn: datasetGroupArn,
	})
	require.NoError(t, err)
	require.Len(t, descOutput.DatasetArns, 1)
	require.Equal(t, *datasetArn, descOutput.DatasetArns[0])
}

func TestForecast_CreatePredictor(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	// Create a dataset.
	datasetOutput, err := client.CreateDataset(ctx, &forecast.CreateDatasetInput{
		DatasetName: aws.String("test-dataset-for-predictor"),
		DatasetType: types.DatasetTypeTargetTimeSeries,
		Domain:      types.DomainRetail,
		Schema: &types.Schema{
			Attributes: []types.SchemaAttribute{
				{
					AttributeName: aws.String("item_id"),
					AttributeType: types.AttributeTypeString,
				},
				{
					AttributeName: aws.String("timestamp"),
					AttributeType: types.AttributeTypeTimestamp,
				},
				{
					AttributeName: aws.String("demand"),
					AttributeType: types.AttributeTypeFloat,
				},
			},
		},
	})
	require.NoError(t, err)

	datasetArn := datasetOutput.DatasetArn

	t.Cleanup(func() {
		_, _ = client.DeleteDataset(ctx, &forecast.DeleteDatasetInput{
			DatasetArn: datasetArn,
		})
	})

	// Create a dataset group with the dataset.
	dgOutput, err := client.CreateDatasetGroup(ctx, &forecast.CreateDatasetGroupInput{
		DatasetGroupName: aws.String("test-dataset-group-for-predictor"),
		Domain:           types.DomainRetail,
		DatasetArns:      []string{*datasetArn},
	})
	require.NoError(t, err)

	datasetGroupArn := dgOutput.DatasetGroupArn

	t.Cleanup(func() {
		_, _ = client.DeleteDatasetGroup(ctx, &forecast.DeleteDatasetGroupInput{
			DatasetGroupArn: datasetGroupArn,
		})
	})

	// Create a predictor.
	predictorOutput, err := client.CreatePredictor(ctx, &forecast.CreatePredictorInput{
		PredictorName:   aws.String("test-predictor"),
		ForecastHorizon: aws.Int32(30),
		InputDataConfig: &types.InputDataConfig{
			DatasetGroupArn: datasetGroupArn,
		},
		FeaturizationConfig: &types.FeaturizationConfig{
			ForecastFrequency: aws.String("D"),
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, predictorOutput.PredictorArn)

	predictorArn := predictorOutput.PredictorArn

	t.Cleanup(func() {
		_, _ = client.DeletePredictor(ctx, &forecast.DeletePredictorInput{
			PredictorArn: predictorArn,
		})
	})

	// Describe predictor.
	descOutput, err := client.DescribePredictor(ctx, &forecast.DescribePredictorInput{
		PredictorArn: predictorArn,
	})
	require.NoError(t, err)
	require.Equal(t, "test-predictor", *descOutput.PredictorName)
	require.Equal(t, int32(30), *descOutput.ForecastHorizon)

	// Delete predictor.
	_, err = client.DeletePredictor(ctx, &forecast.DeletePredictorInput{
		PredictorArn: predictorArn,
	})
	require.NoError(t, err)
}

func TestForecast_CreateForecast(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	// Create a dataset.
	datasetOutput, err := client.CreateDataset(ctx, &forecast.CreateDatasetInput{
		DatasetName: aws.String("test-dataset-for-forecast"),
		DatasetType: types.DatasetTypeTargetTimeSeries,
		Domain:      types.DomainRetail,
		Schema: &types.Schema{
			Attributes: []types.SchemaAttribute{
				{
					AttributeName: aws.String("item_id"),
					AttributeType: types.AttributeTypeString,
				},
				{
					AttributeName: aws.String("timestamp"),
					AttributeType: types.AttributeTypeTimestamp,
				},
				{
					AttributeName: aws.String("demand"),
					AttributeType: types.AttributeTypeFloat,
				},
			},
		},
	})
	require.NoError(t, err)

	datasetArn := datasetOutput.DatasetArn

	t.Cleanup(func() {
		_, _ = client.DeleteDataset(ctx, &forecast.DeleteDatasetInput{
			DatasetArn: datasetArn,
		})
	})

	// Create a dataset group.
	dgOutput, err := client.CreateDatasetGroup(ctx, &forecast.CreateDatasetGroupInput{
		DatasetGroupName: aws.String("test-dataset-group-for-forecast"),
		Domain:           types.DomainRetail,
		DatasetArns:      []string{*datasetArn},
	})
	require.NoError(t, err)

	datasetGroupArn := dgOutput.DatasetGroupArn

	t.Cleanup(func() {
		_, _ = client.DeleteDatasetGroup(ctx, &forecast.DeleteDatasetGroupInput{
			DatasetGroupArn: datasetGroupArn,
		})
	})

	// Create a predictor.
	predictorOutput, err := client.CreatePredictor(ctx, &forecast.CreatePredictorInput{
		PredictorName:   aws.String("test-predictor-for-forecast"),
		ForecastHorizon: aws.Int32(30),
		InputDataConfig: &types.InputDataConfig{
			DatasetGroupArn: datasetGroupArn,
		},
		FeaturizationConfig: &types.FeaturizationConfig{
			ForecastFrequency: aws.String("D"),
		},
	})
	require.NoError(t, err)

	predictorArn := predictorOutput.PredictorArn

	t.Cleanup(func() {
		_, _ = client.DeletePredictor(ctx, &forecast.DeletePredictorInput{
			PredictorArn: predictorArn,
		})
	})

	// Create a forecast.
	forecastOutput, err := client.CreateForecast(ctx, &forecast.CreateForecastInput{
		ForecastName: aws.String("test-forecast"),
		PredictorArn: predictorArn,
	})
	require.NoError(t, err)
	require.NotEmpty(t, forecastOutput.ForecastArn)

	forecastArn := forecastOutput.ForecastArn

	t.Cleanup(func() {
		_, _ = client.DeleteForecast(ctx, &forecast.DeleteForecastInput{
			ForecastArn: forecastArn,
		})
	})

	// Describe forecast.
	descOutput, err := client.DescribeForecast(ctx, &forecast.DescribeForecastInput{
		ForecastArn: forecastArn,
	})
	require.NoError(t, err)
	require.Equal(t, "test-forecast", *descOutput.ForecastName)
	require.Equal(t, *predictorArn, *descOutput.PredictorArn)

	// Delete forecast.
	_, err = client.DeleteForecast(ctx, &forecast.DeleteForecastInput{
		ForecastArn: forecastArn,
	})
	require.NoError(t, err)
}

func TestForecast_DatasetNotFound(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	// Describe non-existent dataset.
	_, err := client.DescribeDataset(ctx, &forecast.DescribeDatasetInput{
		DatasetArn: aws.String("arn:aws:forecast:us-east-1:123456789012:dataset/non-existent"),
	})
	require.Error(t, err)

	// Delete non-existent dataset.
	_, err = client.DeleteDataset(ctx, &forecast.DeleteDatasetInput{
		DatasetArn: aws.String("arn:aws:forecast:us-east-1:123456789012:dataset/non-existent"),
	})
	require.Error(t, err)
}

func TestForecast_DatasetGroupNotFound(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	// Describe non-existent dataset group.
	_, err := client.DescribeDatasetGroup(ctx, &forecast.DescribeDatasetGroupInput{
		DatasetGroupArn: aws.String("arn:aws:forecast:us-east-1:123456789012:dataset-group/non-existent"),
	})
	require.Error(t, err)

	// Delete non-existent dataset group.
	_, err = client.DeleteDatasetGroup(ctx, &forecast.DeleteDatasetGroupInput{
		DatasetGroupArn: aws.String("arn:aws:forecast:us-east-1:123456789012:dataset-group/non-existent"),
	})
	require.Error(t, err)
}

func TestForecast_PredictorNotFound(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	// Describe non-existent predictor.
	_, err := client.DescribePredictor(ctx, &forecast.DescribePredictorInput{
		PredictorArn: aws.String("arn:aws:forecast:us-east-1:123456789012:predictor/non-existent"),
	})
	require.Error(t, err)

	// Delete non-existent predictor.
	_, err = client.DeletePredictor(ctx, &forecast.DeletePredictorInput{
		PredictorArn: aws.String("arn:aws:forecast:us-east-1:123456789012:predictor/non-existent"),
	})
	require.Error(t, err)
}

func TestForecast_ForecastNotFound(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	// Describe non-existent forecast.
	_, err := client.DescribeForecast(ctx, &forecast.DescribeForecastInput{
		ForecastArn: aws.String("arn:aws:forecast:us-east-1:123456789012:forecast/non-existent"),
	})
	require.Error(t, err)

	// Delete non-existent forecast.
	_, err = client.DeleteForecast(ctx, &forecast.DeleteForecastInput{
		ForecastArn: aws.String("arn:aws:forecast:us-east-1:123456789012:forecast/non-existent"),
	})
	require.Error(t, err)
}

func TestForecast_DuplicateDatasetName(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	datasetName := "test-dataset-duplicate"

	// Create first dataset.
	createOutput, err := client.CreateDataset(ctx, &forecast.CreateDatasetInput{
		DatasetName: aws.String(datasetName),
		DatasetType: types.DatasetTypeTargetTimeSeries,
		Domain:      types.DomainRetail,
		Schema: &types.Schema{
			Attributes: []types.SchemaAttribute{
				{
					AttributeName: aws.String("item_id"),
					AttributeType: types.AttributeTypeString,
				},
				{
					AttributeName: aws.String("timestamp"),
					AttributeType: types.AttributeTypeTimestamp,
				},
				{
					AttributeName: aws.String("demand"),
					AttributeType: types.AttributeTypeFloat,
				},
			},
		},
	})
	require.NoError(t, err)

	datasetArn := createOutput.DatasetArn

	t.Cleanup(func() {
		_, _ = client.DeleteDataset(ctx, &forecast.DeleteDatasetInput{
			DatasetArn: datasetArn,
		})
	})

	// Try to create a dataset with the same name.
	_, err = client.CreateDataset(ctx, &forecast.CreateDatasetInput{
		DatasetName: aws.String(datasetName),
		DatasetType: types.DatasetTypeTargetTimeSeries,
		Domain:      types.DomainRetail,
		Schema: &types.Schema{
			Attributes: []types.SchemaAttribute{
				{
					AttributeName: aws.String("item_id"),
					AttributeType: types.AttributeTypeString,
				},
				{
					AttributeName: aws.String("timestamp"),
					AttributeType: types.AttributeTypeTimestamp,
				},
				{
					AttributeName: aws.String("demand"),
					AttributeType: types.AttributeTypeFloat,
				},
			},
		},
	})
	require.Error(t, err)
}

func TestForecast_ResourceInUse(t *testing.T) {
	client := newForecastClient(t)
	ctx := t.Context()

	// Create a dataset.
	datasetOutput, err := client.CreateDataset(ctx, &forecast.CreateDatasetInput{
		DatasetName: aws.String("test-dataset-in-use"),
		DatasetType: types.DatasetTypeTargetTimeSeries,
		Domain:      types.DomainRetail,
		Schema: &types.Schema{
			Attributes: []types.SchemaAttribute{
				{
					AttributeName: aws.String("item_id"),
					AttributeType: types.AttributeTypeString,
				},
				{
					AttributeName: aws.String("timestamp"),
					AttributeType: types.AttributeTypeTimestamp,
				},
				{
					AttributeName: aws.String("demand"),
					AttributeType: types.AttributeTypeFloat,
				},
			},
		},
	})
	require.NoError(t, err)

	datasetArn := datasetOutput.DatasetArn

	// Create a dataset group that uses the dataset.
	dgOutput, err := client.CreateDatasetGroup(ctx, &forecast.CreateDatasetGroupInput{
		DatasetGroupName: aws.String("test-dataset-group-in-use"),
		Domain:           types.DomainRetail,
		DatasetArns:      []string{*datasetArn},
	})
	require.NoError(t, err)

	datasetGroupArn := dgOutput.DatasetGroupArn

	t.Cleanup(func() {
		_, _ = client.DeleteDatasetGroup(ctx, &forecast.DeleteDatasetGroupInput{
			DatasetGroupArn: datasetGroupArn,
		})
		_, _ = client.DeleteDataset(ctx, &forecast.DeleteDatasetInput{
			DatasetArn: datasetArn,
		})
	})

	// Try to delete the dataset (should fail because it's in use).
	_, err = client.DeleteDataset(ctx, &forecast.DeleteDatasetInput{
		DatasetArn: datasetArn,
	})
	require.Error(t, err)
}

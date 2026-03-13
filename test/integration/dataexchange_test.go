//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
)

func newDataExchangeClient(t *testing.T) *dataexchange.Client {
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

	return dataexchange.NewFromConfig(cfg, func(o *dataexchange.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestDataExchange_CreateDataSet(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	result, err := client.CreateDataSet(ctx, &dataexchange.CreateDataSetInput{
		Name:        aws.String("test-dataset"),
		Description: aws.String("test description"),
		AssetType:   types.AssetTypeS3Snapshot,
	})
	if err != nil {
		t.Fatalf("failed to create data set: %v", err)
	}

	if result.Id == nil || *result.Id == "" {
		t.Error("expected Id to be set")
	}

	if *result.Name != "test-dataset" {
		t.Errorf("expected name 'test-dataset', got %s", *result.Name)
	}

	if *result.Description != "test description" {
		t.Errorf("expected description 'test description', got %s", *result.Description)
	}

	if result.Arn == nil || *result.Arn == "" {
		t.Error("expected Arn to be set")
	}
}

func TestDataExchange_GetDataSet(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	createResult, err := client.CreateDataSet(ctx, &dataexchange.CreateDataSetInput{
		Name:        aws.String("get-dataset"),
		Description: aws.String("get test"),
		AssetType:   types.AssetTypeS3Snapshot,
	})
	if err != nil {
		t.Fatalf("failed to create data set: %v", err)
	}

	result, err := client.GetDataSet(ctx, &dataexchange.GetDataSetInput{
		DataSetId: createResult.Id,
	})
	if err != nil {
		t.Fatalf("failed to get data set: %v", err)
	}

	if *result.Name != "get-dataset" {
		t.Errorf("expected name 'get-dataset', got %s", *result.Name)
	}
}

func TestDataExchange_ListDataSets(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	_, err := client.CreateDataSet(ctx, &dataexchange.CreateDataSetInput{
		Name:        aws.String("list-dataset"),
		Description: aws.String("list test"),
		AssetType:   types.AssetTypeS3Snapshot,
	})
	if err != nil {
		t.Fatalf("failed to create data set: %v", err)
	}

	result, err := client.ListDataSets(ctx, &dataexchange.ListDataSetsInput{})
	if err != nil {
		t.Fatalf("failed to list data sets: %v", err)
	}

	if len(result.DataSets) == 0 {
		t.Error("expected at least one data set")
	}
}

func TestDataExchange_UpdateDataSet(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	createResult, err := client.CreateDataSet(ctx, &dataexchange.CreateDataSetInput{
		Name:        aws.String("update-dataset"),
		Description: aws.String("original"),
		AssetType:   types.AssetTypeS3Snapshot,
	})
	if err != nil {
		t.Fatalf("failed to create data set: %v", err)
	}

	result, err := client.UpdateDataSet(ctx, &dataexchange.UpdateDataSetInput{
		DataSetId:   createResult.Id,
		Description: aws.String("updated"),
	})
	if err != nil {
		t.Fatalf("failed to update data set: %v", err)
	}

	if *result.Description != "updated" {
		t.Errorf("expected description 'updated', got %s", *result.Description)
	}
}

func TestDataExchange_DeleteDataSet(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	createResult, err := client.CreateDataSet(ctx, &dataexchange.CreateDataSetInput{
		Name:        aws.String("delete-dataset"),
		Description: aws.String("delete test"),
		AssetType:   types.AssetTypeS3Snapshot,
	})
	if err != nil {
		t.Fatalf("failed to create data set: %v", err)
	}

	_, err = client.DeleteDataSet(ctx, &dataexchange.DeleteDataSetInput{
		DataSetId: createResult.Id,
	})
	if err != nil {
		t.Fatalf("failed to delete data set: %v", err)
	}

	_, err = client.GetDataSet(ctx, &dataexchange.GetDataSetInput{
		DataSetId: createResult.Id,
	})
	if err == nil {
		t.Fatal("expected error for deleted data set")
	}
}

func TestDataExchange_CreateRevision(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	dsResult, err := client.CreateDataSet(ctx, &dataexchange.CreateDataSetInput{
		Name:        aws.String("revision-dataset"),
		Description: aws.String("revision test"),
		AssetType:   types.AssetTypeS3Snapshot,
	})
	if err != nil {
		t.Fatalf("failed to create data set: %v", err)
	}

	result, err := client.CreateRevision(ctx, &dataexchange.CreateRevisionInput{
		DataSetId: dsResult.Id,
		Comment:   aws.String("initial revision"),
	})
	if err != nil {
		t.Fatalf("failed to create revision: %v", err)
	}

	if result.Id == nil || *result.Id == "" {
		t.Error("expected Id to be set")
	}

	if *result.Comment != "initial revision" {
		t.Errorf("expected comment 'initial revision', got %s", *result.Comment)
	}
}

func TestDataExchange_GetRevision(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	dsResult, err := client.CreateDataSet(ctx, &dataexchange.CreateDataSetInput{
		Name:        aws.String("get-revision-dataset"),
		Description: aws.String("get revision test"),
		AssetType:   types.AssetTypeS3Snapshot,
	})
	if err != nil {
		t.Fatalf("failed to create data set: %v", err)
	}

	revResult, err := client.CreateRevision(ctx, &dataexchange.CreateRevisionInput{
		DataSetId: dsResult.Id,
		Comment:   aws.String("get revision"),
	})
	if err != nil {
		t.Fatalf("failed to create revision: %v", err)
	}

	result, err := client.GetRevision(ctx, &dataexchange.GetRevisionInput{
		DataSetId:  dsResult.Id,
		RevisionId: revResult.Id,
	})
	if err != nil {
		t.Fatalf("failed to get revision: %v", err)
	}

	if *result.Comment != "get revision" {
		t.Errorf("expected comment 'get revision', got %s", *result.Comment)
	}
}

func TestDataExchange_ListRevisions(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	dsResult, err := client.CreateDataSet(ctx, &dataexchange.CreateDataSetInput{
		Name:        aws.String("list-revision-dataset"),
		Description: aws.String("list revision test"),
		AssetType:   types.AssetTypeS3Snapshot,
	})
	if err != nil {
		t.Fatalf("failed to create data set: %v", err)
	}

	_, err = client.CreateRevision(ctx, &dataexchange.CreateRevisionInput{
		DataSetId: dsResult.Id,
	})
	if err != nil {
		t.Fatalf("failed to create revision: %v", err)
	}

	result, err := client.ListDataSetRevisions(ctx, &dataexchange.ListDataSetRevisionsInput{
		DataSetId: dsResult.Id,
	})
	if err != nil {
		t.Fatalf("failed to list revisions: %v", err)
	}

	if len(result.Revisions) != 1 {
		t.Errorf("expected 1 revision, got %d", len(result.Revisions))
	}
}

func TestDataExchange_DeleteRevision(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	dsResult, err := client.CreateDataSet(ctx, &dataexchange.CreateDataSetInput{
		Name:        aws.String("delete-revision-dataset"),
		Description: aws.String("delete revision test"),
		AssetType:   types.AssetTypeS3Snapshot,
	})
	if err != nil {
		t.Fatalf("failed to create data set: %v", err)
	}

	revResult, err := client.CreateRevision(ctx, &dataexchange.CreateRevisionInput{
		DataSetId: dsResult.Id,
	})
	if err != nil {
		t.Fatalf("failed to create revision: %v", err)
	}

	_, err = client.DeleteRevision(ctx, &dataexchange.DeleteRevisionInput{
		DataSetId:  dsResult.Id,
		RevisionId: revResult.Id,
	})
	if err != nil {
		t.Fatalf("failed to delete revision: %v", err)
	}

	_, err = client.GetRevision(ctx, &dataexchange.GetRevisionInput{
		DataSetId:  dsResult.Id,
		RevisionId: revResult.Id,
	})
	if err == nil {
		t.Fatal("expected error for deleted revision")
	}
}

func TestDataExchange_CreateJob(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	result, err := client.CreateJob(ctx, &dataexchange.CreateJobInput{
		Type: types.TypeImportAssetsFromS3,
		Details: &types.RequestDetails{
			ImportAssetsFromS3: &types.ImportAssetsFromS3RequestDetails{
				DataSetId:  aws.String("test-dataset-id"),
				RevisionId: aws.String("test-revision-id"),
				AssetSources: []types.AssetSourceEntry{
					{
						Bucket: aws.String("source-bucket"),
						Key:    aws.String("source-key"),
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	if result.Id == nil || *result.Id == "" {
		t.Error("expected Id to be set")
	}

	if result.State != types.StateWaiting {
		t.Errorf("expected state WAITING, got %s", result.State)
	}
}

func TestDataExchange_GetJob(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	createResult, err := client.CreateJob(ctx, &dataexchange.CreateJobInput{
		Type: types.TypeImportAssetsFromS3,
		Details: &types.RequestDetails{
			ImportAssetsFromS3: &types.ImportAssetsFromS3RequestDetails{
				DataSetId:  aws.String("test-dataset-id"),
				RevisionId: aws.String("test-revision-id"),
				AssetSources: []types.AssetSourceEntry{
					{
						Bucket: aws.String("source-bucket"),
						Key:    aws.String("source-key"),
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	result, err := client.GetJob(ctx, &dataexchange.GetJobInput{
		JobId: createResult.Id,
	})
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}

	if result.Type != types.TypeImportAssetsFromS3 {
		t.Errorf("expected type IMPORT_ASSETS_FROM_S3, got %s", result.Type)
	}
}

func TestDataExchange_ListJobs(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	_, err := client.CreateJob(ctx, &dataexchange.CreateJobInput{
		Type: types.TypeImportAssetsFromS3,
		Details: &types.RequestDetails{
			ImportAssetsFromS3: &types.ImportAssetsFromS3RequestDetails{
				DataSetId:  aws.String("test-dataset-id"),
				RevisionId: aws.String("test-revision-id"),
				AssetSources: []types.AssetSourceEntry{
					{
						Bucket: aws.String("source-bucket"),
						Key:    aws.String("source-key"),
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	result, err := client.ListJobs(ctx, &dataexchange.ListJobsInput{})
	if err != nil {
		t.Fatalf("failed to list jobs: %v", err)
	}

	if len(result.Jobs) == 0 {
		t.Error("expected at least one job")
	}
}

func TestDataExchange_DataSetNotFound(t *testing.T) {
	client := newDataExchangeClient(t)
	ctx := t.Context()

	_, err := client.GetDataSet(ctx, &dataexchange.GetDataSetInput{
		DataSetId: aws.String("nonexistent"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent data set")
	}
}

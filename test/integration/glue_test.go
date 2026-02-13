//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/aws-sdk-go-v2/service/glue/types"
)

func newGlueClient(t *testing.T) *glue.Client {
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

	return glue.NewFromConfig(cfg, func(o *glue.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestGlue_CreateAndGetDatabase(t *testing.T) {
	client := newGlueClient(t)
	ctx := t.Context()

	dbName := "test_database"

	// Create database.
	_, err := client.CreateDatabase(ctx, &glue.CreateDatabaseInput{
		DatabaseInput: &types.DatabaseInput{
			Name:        aws.String(dbName),
			Description: aws.String("Test database"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	t.Logf("Created database: %s", dbName)

	// Get database.
	getOutput, err := client.GetDatabase(ctx, &glue.GetDatabaseInput{
		Name: aws.String(dbName),
	})
	if err != nil {
		t.Fatalf("failed to get database: %v", err)
	}

	if getOutput.Database == nil {
		t.Fatal("database is nil")
	}

	if *getOutput.Database.Name != dbName {
		t.Errorf("name mismatch: got %s, want %s", *getOutput.Database.Name, dbName)
	}

	if *getOutput.Database.Description != "Test database" {
		t.Errorf("description mismatch: got %s, want Test database", *getOutput.Database.Description)
	}

	t.Logf("Retrieved database: %s", dbName)
}

func TestGlue_GetDatabases(t *testing.T) {
	client := newGlueClient(t)
	ctx := t.Context()

	// Create a database.
	dbName := "list_test_database"
	_, err := client.CreateDatabase(ctx, &glue.CreateDatabaseInput{
		DatabaseInput: &types.DatabaseInput{
			Name: aws.String(dbName),
		},
	})
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Get databases.
	listOutput, err := client.GetDatabases(ctx, &glue.GetDatabasesInput{
		MaxResults: aws.Int32(10),
	})
	if err != nil {
		t.Fatalf("failed to get databases: %v", err)
	}

	if len(listOutput.DatabaseList) == 0 {
		t.Fatal("no databases returned")
	}

	// Find our database.
	found := false

	for _, db := range listOutput.DatabaseList {
		if db.Name != nil && *db.Name == dbName {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("database %s not found in list", dbName)
	}

	t.Logf("Listed %d databases", len(listOutput.DatabaseList))
}

func TestGlue_DeleteDatabase(t *testing.T) {
	client := newGlueClient(t)
	ctx := t.Context()

	dbName := "delete_test_database"

	// Create a database.
	_, err := client.CreateDatabase(ctx, &glue.CreateDatabaseInput{
		DatabaseInput: &types.DatabaseInput{
			Name: aws.String(dbName),
		},
	})
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Delete the database.
	_, err = client.DeleteDatabase(ctx, &glue.DeleteDatabaseInput{
		Name: aws.String(dbName),
	})
	if err != nil {
		t.Fatalf("failed to delete database: %v", err)
	}

	t.Logf("Deleted database: %s", dbName)

	// Verify it's deleted.
	_, err = client.GetDatabase(ctx, &glue.GetDatabaseInput{
		Name: aws.String(dbName),
	})
	if err == nil {
		t.Fatal("expected error when getting deleted database")
	}

	t.Logf("Verified database is deleted")
}

func TestGlue_CreateAndGetTable(t *testing.T) {
	client := newGlueClient(t)
	ctx := t.Context()

	dbName := "table_test_database"
	tableName := "test_table"

	// Create database first.
	_, err := client.CreateDatabase(ctx, &glue.CreateDatabaseInput{
		DatabaseInput: &types.DatabaseInput{
			Name: aws.String(dbName),
		},
	})
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Create table.
	_, err = client.CreateTable(ctx, &glue.CreateTableInput{
		DatabaseName: aws.String(dbName),
		TableInput: &types.TableInput{
			Name:        aws.String(tableName),
			Description: aws.String("Test table"),
			TableType:   aws.String("EXTERNAL_TABLE"),
			StorageDescriptor: &types.StorageDescriptor{
				Columns: []types.Column{
					{
						Name: aws.String("id"),
						Type: aws.String("int"),
					},
					{
						Name: aws.String("name"),
						Type: aws.String("string"),
					},
				},
				Location:     aws.String("s3://test-bucket/data/"),
				InputFormat:  aws.String("org.apache.hadoop.mapred.TextInputFormat"),
				OutputFormat: aws.String("org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Logf("Created table: %s", tableName)

	// Get table.
	getOutput, err := client.GetTable(ctx, &glue.GetTableInput{
		DatabaseName: aws.String(dbName),
		Name:         aws.String(tableName),
	})
	if err != nil {
		t.Fatalf("failed to get table: %v", err)
	}

	if getOutput.Table == nil {
		t.Fatal("table is nil")
	}

	if *getOutput.Table.Name != tableName {
		t.Errorf("name mismatch: got %s, want %s", *getOutput.Table.Name, tableName)
	}

	if *getOutput.Table.TableType != "EXTERNAL_TABLE" {
		t.Errorf("table type mismatch: got %s, want EXTERNAL_TABLE", *getOutput.Table.TableType)
	}

	t.Logf("Retrieved table: %s", tableName)
}

func TestGlue_GetTables(t *testing.T) {
	client := newGlueClient(t)
	ctx := t.Context()

	dbName := "get_tables_database"
	tableName := "list_test_table"

	// Create database.
	_, err := client.CreateDatabase(ctx, &glue.CreateDatabaseInput{
		DatabaseInput: &types.DatabaseInput{
			Name: aws.String(dbName),
		},
	})
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Create table.
	_, err = client.CreateTable(ctx, &glue.CreateTableInput{
		DatabaseName: aws.String(dbName),
		TableInput: &types.TableInput{
			Name: aws.String(tableName),
		},
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Get tables.
	listOutput, err := client.GetTables(ctx, &glue.GetTablesInput{
		DatabaseName: aws.String(dbName),
		MaxResults:   aws.Int32(10),
	})
	if err != nil {
		t.Fatalf("failed to get tables: %v", err)
	}

	if len(listOutput.TableList) == 0 {
		t.Fatal("no tables returned")
	}

	// Find our table.
	found := false

	for _, table := range listOutput.TableList {
		if table.Name != nil && *table.Name == tableName {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("table %s not found in list", tableName)
	}

	t.Logf("Listed %d tables", len(listOutput.TableList))
}

func TestGlue_DeleteTable(t *testing.T) {
	client := newGlueClient(t)
	ctx := t.Context()

	dbName := "delete_table_database"
	tableName := "delete_test_table"

	// Create database.
	_, err := client.CreateDatabase(ctx, &glue.CreateDatabaseInput{
		DatabaseInput: &types.DatabaseInput{
			Name: aws.String(dbName),
		},
	})
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Create table.
	_, err = client.CreateTable(ctx, &glue.CreateTableInput{
		DatabaseName: aws.String(dbName),
		TableInput: &types.TableInput{
			Name: aws.String(tableName),
		},
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Delete table.
	_, err = client.DeleteTable(ctx, &glue.DeleteTableInput{
		DatabaseName: aws.String(dbName),
		Name:         aws.String(tableName),
	})
	if err != nil {
		t.Fatalf("failed to delete table: %v", err)
	}

	t.Logf("Deleted table: %s", tableName)

	// Verify it's deleted.
	_, err = client.GetTable(ctx, &glue.GetTableInput{
		DatabaseName: aws.String(dbName),
		Name:         aws.String(tableName),
	})
	if err == nil {
		t.Fatal("expected error when getting deleted table")
	}

	t.Logf("Verified table is deleted")
}

func TestGlue_CreateAndDeleteJob(t *testing.T) {
	client := newGlueClient(t)
	ctx := t.Context()

	jobName := "test_job"

	// Create job.
	createOutput, err := client.CreateJob(ctx, &glue.CreateJobInput{
		Name:        aws.String(jobName),
		Description: aws.String("Test ETL job"),
		Role:        aws.String("arn:aws:iam::000000000000:role/GlueRole"),
		Command: &types.JobCommand{
			Name:           aws.String("glueetl"),
			ScriptLocation: aws.String("s3://test-bucket/scripts/etl.py"),
			PythonVersion:  aws.String("3"),
		},
		GlueVersion:     aws.String("3.0"),
		NumberOfWorkers: aws.Int32(10),
		WorkerType:      types.WorkerTypeG1x,
	})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	if createOutput.Name == nil || *createOutput.Name != jobName {
		t.Errorf("job name mismatch: got %v, want %s", createOutput.Name, jobName)
	}

	t.Logf("Created job: %s", jobName)

	// Delete job.
	deleteOutput, err := client.DeleteJob(ctx, &glue.DeleteJobInput{
		JobName: aws.String(jobName),
	})
	if err != nil {
		t.Fatalf("failed to delete job: %v", err)
	}

	if deleteOutput.JobName == nil || *deleteOutput.JobName != jobName {
		t.Errorf("job name mismatch: got %v, want %s", deleteOutput.JobName, jobName)
	}

	t.Logf("Deleted job: %s", jobName)
}

func TestGlue_StartJobRun(t *testing.T) {
	client := newGlueClient(t)
	ctx := t.Context()

	jobName := "run_test_job"

	// Create job first.
	_, err := client.CreateJob(ctx, &glue.CreateJobInput{
		Name: aws.String(jobName),
		Role: aws.String("arn:aws:iam::000000000000:role/GlueRole"),
		Command: &types.JobCommand{
			Name:           aws.String("glueetl"),
			ScriptLocation: aws.String("s3://test-bucket/scripts/etl.py"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	// Start job run.
	runOutput, err := client.StartJobRun(ctx, &glue.StartJobRunInput{
		JobName: aws.String(jobName),
		Arguments: map[string]string{
			"--input":  "s3://test-bucket/input/",
			"--output": "s3://test-bucket/output/",
		},
	})
	if err != nil {
		t.Fatalf("failed to start job run: %v", err)
	}

	if runOutput.JobRunId == nil {
		t.Fatal("job run ID is nil")
	}

	t.Logf("Started job run: %s", *runOutput.JobRunId)
}

func TestGlue_GetNonExistentDatabase(t *testing.T) {
	client := newGlueClient(t)
	ctx := t.Context()

	// Try to get a non-existent database.
	_, err := client.GetDatabase(ctx, &glue.GetDatabaseInput{
		Name: aws.String("non_existent_database"),
	})
	if err == nil {
		t.Fatal("expected error when getting non-existent database")
	}

	t.Logf("Got expected error: %v", err)
}

func TestGlue_CreateDuplicateDatabase(t *testing.T) {
	client := newGlueClient(t)
	ctx := t.Context()

	dbName := "duplicate_test_database"

	// Create database.
	_, err := client.CreateDatabase(ctx, &glue.CreateDatabaseInput{
		DatabaseInput: &types.DatabaseInput{
			Name: aws.String(dbName),
		},
	})
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Try to create the same database again.
	_, err = client.CreateDatabase(ctx, &glue.CreateDatabaseInput{
		DatabaseInput: &types.DatabaseInput{
			Name: aws.String(dbName),
		},
	})
	if err == nil {
		t.Fatal("expected error when creating duplicate database")
	}

	t.Logf("Got expected error: %v", err)
}

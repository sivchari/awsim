//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/sivchari/golden"
)

func newDocDBClient(t *testing.T) *docdb.Client {
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

	return docdb.NewFromConfig(cfg, func(o *docdb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestDocDB_CreateAndDeleteCluster(t *testing.T) {
	client := newDocDBClient(t)
	ctx := t.Context()
	clusterID := "test-docdb-cluster"

	createResult, err := client.CreateDBCluster(ctx, &docdb.CreateDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		Engine:              aws.String("docdb"),
		MasterUsername:      aws.String("admin"),
		MasterUserPassword:  aws.String("password123"),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("DBClusterArn", "DbClusterResourceId", "ClusterCreateTime", "Endpoint", "ReaderEndpoint", "ResultMetadata")).Assert(t.Name()+"_create", createResult)

	// Delete cluster
	deleteResult, err := client.DeleteDBCluster(ctx, &docdb.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		SkipFinalSnapshot:   aws.Bool(true),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("DBClusterArn", "DbClusterResourceId", "ClusterCreateTime", "Endpoint", "ReaderEndpoint", "ResultMetadata")).Assert(t.Name()+"_delete", deleteResult)
}

func TestDocDB_DescribeClusters(t *testing.T) {
	client := newDocDBClient(t)
	ctx := t.Context()
	clusterID := "test-docdb-describe-cluster"

	_, err := client.CreateDBCluster(ctx, &docdb.CreateDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		Engine:              aws.String("docdb"),
		MasterUsername:      aws.String("admin"),
		MasterUserPassword:  aws.String("password123"),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteDBCluster(context.Background(), &docdb.DeleteDBClusterInput{
			DBClusterIdentifier: aws.String(clusterID),
			SkipFinalSnapshot:   aws.Bool(true),
		})
	})

	describeResult, err := client.DescribeDBClusters(ctx, &docdb.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(clusterID),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("DBClusterArn", "DbClusterResourceId", "ClusterCreateTime", "Endpoint", "ReaderEndpoint", "ResultMetadata")).Assert(t.Name(), describeResult)
}

func TestDocDB_ModifyCluster(t *testing.T) {
	client := newDocDBClient(t)
	ctx := t.Context()
	clusterID := "test-docdb-modify-cluster"

	_, err := client.CreateDBCluster(ctx, &docdb.CreateDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		Engine:              aws.String("docdb"),
		MasterUsername:      aws.String("admin"),
		MasterUserPassword:  aws.String("password123"),
		DeletionProtection:  aws.Bool(false),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteDBCluster(context.Background(), &docdb.DeleteDBClusterInput{
			DBClusterIdentifier: aws.String(clusterID),
			SkipFinalSnapshot:   aws.Bool(true),
		})
	})

	modifyResult, err := client.ModifyDBCluster(ctx, &docdb.ModifyDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		DeletionProtection:  aws.Bool(true),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("DBClusterArn", "DbClusterResourceId", "ClusterCreateTime", "Endpoint", "ReaderEndpoint", "ResultMetadata")).Assert(t.Name()+"_modify", modifyResult)

	// Disable deletion protection for cleanup
	_, err = client.ModifyDBCluster(ctx, &docdb.ModifyDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		DeletionProtection:  aws.Bool(false),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDocDB_CreateAndDeleteInstance(t *testing.T) {
	client := newDocDBClient(t)
	ctx := t.Context()
	clusterID := "test-docdb-instance-cluster"
	instanceID := "test-docdb-instance"

	_, err := client.CreateDBCluster(ctx, &docdb.CreateDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		Engine:              aws.String("docdb"),
		MasterUsername:      aws.String("admin"),
		MasterUserPassword:  aws.String("password123"),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteDBCluster(context.Background(), &docdb.DeleteDBClusterInput{
			DBClusterIdentifier: aws.String(clusterID),
			SkipFinalSnapshot:   aws.Bool(true),
		})
	})

	createResult, err := client.CreateDBInstance(ctx, &docdb.CreateDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
		DBInstanceClass:      aws.String("db.r5.large"),
		Engine:               aws.String("docdb"),
		DBClusterIdentifier:  aws.String(clusterID),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("DBInstanceArn", "DbiResourceId", "InstanceCreateTime", "Address", "ResultMetadata")).Assert(t.Name()+"_create", createResult)

	// Delete instance
	deleteResult, err := client.DeleteDBInstance(ctx, &docdb.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("DBInstanceArn", "DbiResourceId", "InstanceCreateTime", "Address", "ResultMetadata")).Assert(t.Name()+"_delete", deleteResult)
}

func TestDocDB_DescribeInstances(t *testing.T) {
	client := newDocDBClient(t)
	ctx := t.Context()
	clusterID := "test-docdb-describe-inst-cluster"
	instanceID := "test-docdb-describe-instance"

	_, err := client.CreateDBCluster(ctx, &docdb.CreateDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		Engine:              aws.String("docdb"),
		MasterUsername:      aws.String("admin"),
		MasterUserPassword:  aws.String("password123"),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteDBInstance(context.Background(), &docdb.DeleteDBInstanceInput{
			DBInstanceIdentifier: aws.String(instanceID),
		})
		_, _ = client.DeleteDBCluster(context.Background(), &docdb.DeleteDBClusterInput{
			DBClusterIdentifier: aws.String(clusterID),
			SkipFinalSnapshot:   aws.Bool(true),
		})
	})

	_, err = client.CreateDBInstance(ctx, &docdb.CreateDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
		DBInstanceClass:      aws.String("db.r5.large"),
		Engine:               aws.String("docdb"),
		DBClusterIdentifier:  aws.String(clusterID),
	})
	if err != nil {
		t.Fatal(err)
	}

	describeResult, err := client.DescribeDBInstances(ctx, &docdb.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instanceID),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("DBInstanceArn", "DbiResourceId", "InstanceCreateTime", "Address", "ResultMetadata")).Assert(t.Name(), describeResult)
}

func TestDocDB_ClusterNotFound(t *testing.T) {
	client := newDocDBClient(t)
	ctx := t.Context()

	_, err := client.DescribeDBClusters(ctx, &docdb.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String("non-existent-cluster"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent cluster")
	}
}

func TestDocDB_DuplicateCluster(t *testing.T) {
	client := newDocDBClient(t)
	ctx := t.Context()
	clusterID := "test-docdb-duplicate-cluster"

	_, err := client.CreateDBCluster(ctx, &docdb.CreateDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		Engine:              aws.String("docdb"),
		MasterUsername:      aws.String("admin"),
		MasterUserPassword:  aws.String("password123"),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteDBCluster(context.Background(), &docdb.DeleteDBClusterInput{
			DBClusterIdentifier: aws.String(clusterID),
			SkipFinalSnapshot:   aws.Bool(true),
		})
	})

	_, err = client.CreateDBCluster(ctx, &docdb.CreateDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		Engine:              aws.String("docdb"),
		MasterUsername:      aws.String("admin"),
		MasterUserPassword:  aws.String("password123"),
	})
	if err == nil {
		t.Fatal("expected error for duplicate cluster")
	}
}

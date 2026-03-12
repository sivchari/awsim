//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
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
		t.Fatalf("failed to create cluster: %v", err)
	}

	if createResult.DBCluster == nil {
		t.Fatal("expected DBCluster in response, got nil")
	}

	if *createResult.DBCluster.DBClusterIdentifier != clusterID {
		t.Errorf("expected cluster identifier %s, got %s", clusterID, *createResult.DBCluster.DBClusterIdentifier)
	}

	if *createResult.DBCluster.Engine != "docdb" {
		t.Errorf("expected engine docdb, got %s", *createResult.DBCluster.Engine)
	}

	// Delete cluster
	deleteResult, err := client.DeleteDBCluster(ctx, &docdb.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		SkipFinalSnapshot:   aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to delete cluster: %v", err)
	}

	if *deleteResult.DBCluster.DBClusterIdentifier != clusterID {
		t.Errorf("expected cluster identifier %s, got %s", clusterID, *deleteResult.DBCluster.DBClusterIdentifier)
	}
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
		t.Fatalf("failed to create cluster: %v", err)
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
		t.Fatalf("failed to describe clusters: %v", err)
	}

	if len(describeResult.DBClusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(describeResult.DBClusters))
	}

	cluster := describeResult.DBClusters[0]

	if *cluster.DBClusterIdentifier != clusterID {
		t.Errorf("expected cluster identifier %s, got %s", clusterID, *cluster.DBClusterIdentifier)
	}

	if *cluster.Status != "available" {
		t.Errorf("expected status available, got %s", *cluster.Status)
	}
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
		t.Fatalf("failed to create cluster: %v", err)
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
		t.Fatalf("failed to modify cluster: %v", err)
	}

	if !*modifyResult.DBCluster.DeletionProtection {
		t.Error("expected deletion protection to be true after modify")
	}

	// Disable deletion protection for cleanup
	_, err = client.ModifyDBCluster(ctx, &docdb.ModifyDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		DeletionProtection:  aws.Bool(false),
	})
	if err != nil {
		t.Fatalf("failed to reset deletion protection: %v", err)
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
		t.Fatalf("failed to create cluster: %v", err)
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
		t.Fatalf("failed to create instance: %v", err)
	}

	if createResult.DBInstance == nil {
		t.Fatal("expected DBInstance in response, got nil")
	}

	if *createResult.DBInstance.DBInstanceIdentifier != instanceID {
		t.Errorf("expected instance identifier %s, got %s", instanceID, *createResult.DBInstance.DBInstanceIdentifier)
	}

	// Delete instance
	deleteResult, err := client.DeleteDBInstance(ctx, &docdb.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
	})
	if err != nil {
		t.Fatalf("failed to delete instance: %v", err)
	}

	if *deleteResult.DBInstance.DBInstanceIdentifier != instanceID {
		t.Errorf("expected instance identifier %s, got %s", instanceID, *deleteResult.DBInstance.DBInstanceIdentifier)
	}
}

func TestDocDB_DescribeInstances(t *testing.T) {
	client := newDocDBClient(t)
	ctx := t.Context()
	instanceID := "test-docdb-describe-instance"

	_, err := client.CreateDBInstance(ctx, &docdb.CreateDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
		DBInstanceClass:      aws.String("db.r5.large"),
		Engine:               aws.String("docdb"),
	})
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteDBInstance(context.Background(), &docdb.DeleteDBInstanceInput{
			DBInstanceIdentifier: aws.String(instanceID),
		})
	})

	describeResult, err := client.DescribeDBInstances(ctx, &docdb.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instanceID),
	})
	if err != nil {
		t.Fatalf("failed to describe instances: %v", err)
	}

	if len(describeResult.DBInstances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(describeResult.DBInstances))
	}

	if *describeResult.DBInstances[0].DBInstanceIdentifier != instanceID {
		t.Errorf("expected instance identifier %s, got %s", instanceID, *describeResult.DBInstances[0].DBInstanceIdentifier)
	}
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
		t.Fatalf("failed to create cluster: %v", err)
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

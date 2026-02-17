//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

func newRDSClient(t *testing.T) *rds.Client {
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

	return rds.NewFromConfig(cfg, func(o *rds.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestRDS_CreateAndDescribeDBInstance(t *testing.T) {
	client := newRDSClient(t)
	ctx := t.Context()

	instanceID := "test-db-instance"

	// Create DB instance
	createResult, err := client.CreateDBInstance(ctx, &rds.CreateDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
		DBInstanceClass:      aws.String("db.t3.micro"),
		Engine:               aws.String("mysql"),
		MasterUsername:       aws.String("admin"),
		MasterUserPassword:   aws.String("password123"),
		AllocatedStorage:     aws.Int32(20),
	})
	if err != nil {
		t.Fatalf("failed to create DB instance: %v", err)
	}

	if createResult.DBInstance == nil {
		t.Fatal("expected DBInstance in response, got nil")
	}

	if *createResult.DBInstance.DBInstanceIdentifier != instanceID {
		t.Errorf("expected identifier %s, got %s", instanceID, *createResult.DBInstance.DBInstanceIdentifier)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteDBInstance(ctx, &rds.DeleteDBInstanceInput{
			DBInstanceIdentifier: aws.String(instanceID),
			SkipFinalSnapshot:    aws.Bool(true),
		})
	})

	// Describe DB instances
	descResult, err := client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instanceID),
	})
	if err != nil {
		t.Fatalf("failed to describe DB instances: %v", err)
	}

	if len(descResult.DBInstances) != 1 {
		t.Errorf("expected 1 instance, got %d", len(descResult.DBInstances))
	}

	if *descResult.DBInstances[0].DBInstanceIdentifier != instanceID {
		t.Errorf("expected identifier %s, got %s", instanceID, *descResult.DBInstances[0].DBInstanceIdentifier)
	}
}

func TestRDS_ModifyDBInstance(t *testing.T) {
	client := newRDSClient(t)
	ctx := t.Context()

	instanceID := "test-modify-db-instance"

	// Create DB instance
	_, err := client.CreateDBInstance(ctx, &rds.CreateDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
		DBInstanceClass:      aws.String("db.t3.micro"),
		Engine:               aws.String("postgres"),
	})
	if err != nil {
		t.Fatalf("failed to create DB instance: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteDBInstance(ctx, &rds.DeleteDBInstanceInput{
			DBInstanceIdentifier: aws.String(instanceID),
			SkipFinalSnapshot:    aws.Bool(true),
		})
	})

	// Modify DB instance
	modifyResult, err := client.ModifyDBInstance(ctx, &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
		DBInstanceClass:      aws.String("db.t3.small"),
		ApplyImmediately:     aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to modify DB instance: %v", err)
	}

	if *modifyResult.DBInstance.DBInstanceClass != "db.t3.small" {
		t.Errorf("expected class db.t3.small, got %s", *modifyResult.DBInstance.DBInstanceClass)
	}
}

func TestRDS_StartAndStopDBInstance(t *testing.T) {
	client := newRDSClient(t)
	ctx := t.Context()

	instanceID := "test-start-stop-db-instance"

	// Create DB instance
	_, err := client.CreateDBInstance(ctx, &rds.CreateDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
		DBInstanceClass:      aws.String("db.t3.micro"),
		Engine:               aws.String("mysql"),
	})
	if err != nil {
		t.Fatalf("failed to create DB instance: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteDBInstance(ctx, &rds.DeleteDBInstanceInput{
			DBInstanceIdentifier: aws.String(instanceID),
			SkipFinalSnapshot:    aws.Bool(true),
		})
	})

	// Stop DB instance
	stopResult, err := client.StopDBInstance(ctx, &rds.StopDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
	})
	if err != nil {
		t.Fatalf("failed to stop DB instance: %v", err)
	}

	if *stopResult.DBInstance.DBInstanceStatus != "stopped" {
		t.Errorf("expected status stopped, got %s", *stopResult.DBInstance.DBInstanceStatus)
	}

	// Start DB instance
	startResult, err := client.StartDBInstance(ctx, &rds.StartDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
	})
	if err != nil {
		t.Fatalf("failed to start DB instance: %v", err)
	}

	if *startResult.DBInstance.DBInstanceStatus != "available" {
		t.Errorf("expected status available, got %s", *startResult.DBInstance.DBInstanceStatus)
	}
}

func TestRDS_DeleteDBInstance(t *testing.T) {
	client := newRDSClient(t)
	ctx := t.Context()

	instanceID := "test-delete-db-instance"

	// Create DB instance
	_, err := client.CreateDBInstance(ctx, &rds.CreateDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
		DBInstanceClass:      aws.String("db.t3.micro"),
		Engine:               aws.String("mysql"),
	})
	if err != nil {
		t.Fatalf("failed to create DB instance: %v", err)
	}

	// Delete DB instance
	deleteResult, err := client.DeleteDBInstance(ctx, &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
		SkipFinalSnapshot:    aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to delete DB instance: %v", err)
	}

	if *deleteResult.DBInstance.DBInstanceIdentifier != instanceID {
		t.Errorf("expected identifier %s, got %s", instanceID, *deleteResult.DBInstance.DBInstanceIdentifier)
	}

	// Verify instance is deleted
	_, err = client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instanceID),
	})
	if err == nil {
		t.Error("expected error when describing deleted instance, got nil")
	}
}

func TestRDS_CreateAndDescribeDBCluster(t *testing.T) {
	client := newRDSClient(t)
	ctx := t.Context()

	clusterID := "test-db-cluster"

	// Create DB cluster
	createResult, err := client.CreateDBCluster(ctx, &rds.CreateDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		Engine:              aws.String("aurora-mysql"),
		MasterUsername:      aws.String("admin"),
		MasterUserPassword:  aws.String("password123"),
	})
	if err != nil {
		t.Fatalf("failed to create DB cluster: %v", err)
	}

	if createResult.DBCluster == nil {
		t.Fatal("expected DBCluster in response, got nil")
	}

	if *createResult.DBCluster.DBClusterIdentifier != clusterID {
		t.Errorf("expected identifier %s, got %s", clusterID, *createResult.DBCluster.DBClusterIdentifier)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteDBCluster(ctx, &rds.DeleteDBClusterInput{
			DBClusterIdentifier: aws.String(clusterID),
			SkipFinalSnapshot:   aws.Bool(true),
		})
	})

	// Describe DB clusters
	descResult, err := client.DescribeDBClusters(ctx, &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(clusterID),
	})
	if err != nil {
		t.Fatalf("failed to describe DB clusters: %v", err)
	}

	if len(descResult.DBClusters) != 1 {
		t.Errorf("expected 1 cluster, got %d", len(descResult.DBClusters))
	}

	if *descResult.DBClusters[0].DBClusterIdentifier != clusterID {
		t.Errorf("expected identifier %s, got %s", clusterID, *descResult.DBClusters[0].DBClusterIdentifier)
	}
}

func TestRDS_DeleteDBCluster(t *testing.T) {
	client := newRDSClient(t)
	ctx := t.Context()

	clusterID := "test-delete-db-cluster"

	// Create DB cluster
	_, err := client.CreateDBCluster(ctx, &rds.CreateDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		Engine:              aws.String("aurora-postgresql"),
	})
	if err != nil {
		t.Fatalf("failed to create DB cluster: %v", err)
	}

	// Delete DB cluster
	deleteResult, err := client.DeleteDBCluster(ctx, &rds.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		SkipFinalSnapshot:   aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to delete DB cluster: %v", err)
	}

	if *deleteResult.DBCluster.DBClusterIdentifier != clusterID {
		t.Errorf("expected identifier %s, got %s", clusterID, *deleteResult.DBCluster.DBClusterIdentifier)
	}

	// Verify cluster is deleted
	_, err = client.DescribeDBClusters(ctx, &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(clusterID),
	})
	if err == nil {
		t.Error("expected error when describing deleted cluster, got nil")
	}
}

func TestRDS_CreateAndDeleteDBSnapshot(t *testing.T) {
	client := newRDSClient(t)
	ctx := t.Context()

	instanceID := "test-snapshot-db-instance"
	snapshotID := "test-db-snapshot"

	// Create DB instance first
	_, err := client.CreateDBInstance(ctx, &rds.CreateDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
		DBInstanceClass:      aws.String("db.t3.micro"),
		Engine:               aws.String("mysql"),
	})
	if err != nil {
		t.Fatalf("failed to create DB instance: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteDBSnapshot(ctx, &rds.DeleteDBSnapshotInput{
			DBSnapshotIdentifier: aws.String(snapshotID),
		})
		_, _ = client.DeleteDBInstance(ctx, &rds.DeleteDBInstanceInput{
			DBInstanceIdentifier: aws.String(instanceID),
			SkipFinalSnapshot:    aws.Bool(true),
		})
	})

	// Create DB snapshot
	createResult, err := client.CreateDBSnapshot(ctx, &rds.CreateDBSnapshotInput{
		DBSnapshotIdentifier: aws.String(snapshotID),
		DBInstanceIdentifier: aws.String(instanceID),
	})
	if err != nil {
		t.Fatalf("failed to create DB snapshot: %v", err)
	}

	if createResult.DBSnapshot == nil {
		t.Fatal("expected DBSnapshot in response, got nil")
	}

	if *createResult.DBSnapshot.DBSnapshotIdentifier != snapshotID {
		t.Errorf("expected snapshot identifier %s, got %s", snapshotID, *createResult.DBSnapshot.DBSnapshotIdentifier)
	}

	if *createResult.DBSnapshot.DBInstanceIdentifier != instanceID {
		t.Errorf("expected instance identifier %s, got %s", instanceID, *createResult.DBSnapshot.DBInstanceIdentifier)
	}

	// Delete DB snapshot
	deleteResult, err := client.DeleteDBSnapshot(ctx, &rds.DeleteDBSnapshotInput{
		DBSnapshotIdentifier: aws.String(snapshotID),
	})
	if err != nil {
		t.Fatalf("failed to delete DB snapshot: %v", err)
	}

	if *deleteResult.DBSnapshot.DBSnapshotIdentifier != snapshotID {
		t.Errorf("expected snapshot identifier %s, got %s", snapshotID, *deleteResult.DBSnapshot.DBSnapshotIdentifier)
	}
}

func TestRDS_DescribeDBInstances_All(t *testing.T) {
	client := newRDSClient(t)
	ctx := t.Context()

	// Create multiple DB instances
	instanceIDs := []string{"test-db-instance-1", "test-db-instance-2"}
	for _, id := range instanceIDs {
		_, err := client.CreateDBInstance(ctx, &rds.CreateDBInstanceInput{
			DBInstanceIdentifier: aws.String(id),
			DBInstanceClass:      aws.String("db.t3.micro"),
			Engine:               aws.String("mysql"),
		})
		if err != nil {
			t.Fatalf("failed to create DB instance %s: %v", id, err)
		}
	}

	t.Cleanup(func() {
		for _, id := range instanceIDs {
			_, _ = client.DeleteDBInstance(ctx, &rds.DeleteDBInstanceInput{
				DBInstanceIdentifier: aws.String(id),
				SkipFinalSnapshot:    aws.Bool(true),
			})
		}
	})

	// Describe all DB instances
	descResult, err := client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		t.Fatalf("failed to describe DB instances: %v", err)
	}

	if len(descResult.DBInstances) < 2 {
		t.Errorf("expected at least 2 instances, got %d", len(descResult.DBInstances))
	}
}

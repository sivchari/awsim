//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/memorydb"
	"github.com/aws/aws-sdk-go-v2/service/memorydb/types"
)

func newMemoryDBClient(t *testing.T) *memorydb.Client {
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

	return memorydb.NewFromConfig(cfg, func(o *memorydb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestMemoryDB_CreateAndDeleteCluster(t *testing.T) {
	client := newMemoryDBClient(t)
	ctx := t.Context()
	clusterName := "test-cluster"

	createResult, err := client.CreateCluster(ctx, &memorydb.CreateClusterInput{
		ClusterName: aws.String(clusterName),
		NodeType:    aws.String("db.r6g.large"),
		ACLName:     aws.String("open-access"),
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	if *createResult.Cluster.Name != clusterName {
		t.Errorf("expected cluster name %s, got %s", clusterName, *createResult.Cluster.Name)
	}

	if createResult.Cluster.ARN == nil || *createResult.Cluster.ARN == "" {
		t.Error("expected cluster ARN to be set")
	}

	if *createResult.Cluster.Status != "available" {
		t.Errorf("expected status available, got %s", *createResult.Cluster.Status)
	}

	// Delete
	deleteResult, err := client.DeleteCluster(ctx, &memorydb.DeleteClusterInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to delete cluster: %v", err)
	}

	if *deleteResult.Cluster.Name != clusterName {
		t.Errorf("expected cluster name %s, got %s", clusterName, *deleteResult.Cluster.Name)
	}
}

func TestMemoryDB_DescribeClusters(t *testing.T) {
	client := newMemoryDBClient(t)
	ctx := t.Context()
	clusterName := "test-describe-cluster"

	_, err := client.CreateCluster(ctx, &memorydb.CreateClusterInput{
		ClusterName: aws.String(clusterName),
		NodeType:    aws.String("db.r6g.large"),
		ACLName:     aws.String("open-access"),
		Description: aws.String("test description"),
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(t.Context(), &memorydb.DeleteClusterInput{
			ClusterName: aws.String(clusterName),
		})
	})

	describeResult, err := client.DescribeClusters(ctx, &memorydb.DescribeClustersInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to describe clusters: %v", err)
	}

	if len(describeResult.Clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(describeResult.Clusters))
	}

	cluster := describeResult.Clusters[0]
	if *cluster.Name != clusterName {
		t.Errorf("expected cluster name %s, got %s", clusterName, *cluster.Name)
	}

	if *cluster.Description != "test description" {
		t.Errorf("expected description 'test description', got %s", *cluster.Description)
	}
}

func TestMemoryDB_UpdateCluster(t *testing.T) {
	client := newMemoryDBClient(t)
	ctx := t.Context()
	clusterName := "test-update-cluster"

	_, err := client.CreateCluster(ctx, &memorydb.CreateClusterInput{
		ClusterName: aws.String(clusterName),
		NodeType:    aws.String("db.r6g.large"),
		ACLName:     aws.String("open-access"),
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(t.Context(), &memorydb.DeleteClusterInput{
			ClusterName: aws.String(clusterName),
		})
	})

	updateResult, err := client.UpdateCluster(ctx, &memorydb.UpdateClusterInput{
		ClusterName: aws.String(clusterName),
		Description: aws.String("updated description"),
	})
	if err != nil {
		t.Fatalf("failed to update cluster: %v", err)
	}

	if *updateResult.Cluster.Description != "updated description" {
		t.Errorf("expected description 'updated description', got %s", *updateResult.Cluster.Description)
	}
}

func TestMemoryDB_CreateAndDeleteUser(t *testing.T) {
	client := newMemoryDBClient(t)
	ctx := t.Context()
	userName := "test-user"

	createResult, err := client.CreateUser(ctx, &memorydb.CreateUserInput{
		UserName:     aws.String(userName),
		AccessString: aws.String("on ~* &* +@all"),
		AuthenticationMode: &types.AuthenticationMode{
			Type:      types.InputAuthenticationTypePassword,
			Passwords: []string{"testpassword123"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if *createResult.User.Name != userName {
		t.Errorf("expected user name %s, got %s", userName, *createResult.User.Name)
	}

	if createResult.User.ARN == nil || *createResult.User.ARN == "" {
		t.Error("expected user ARN to be set")
	}

	// Delete
	deleteResult, err := client.DeleteUser(ctx, &memorydb.DeleteUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	if *deleteResult.User.Name != userName {
		t.Errorf("expected user name %s, got %s", userName, *deleteResult.User.Name)
	}
}

func TestMemoryDB_DescribeUsers(t *testing.T) {
	client := newMemoryDBClient(t)
	ctx := t.Context()
	userName := "test-describe-user"

	_, err := client.CreateUser(ctx, &memorydb.CreateUserInput{
		UserName:     aws.String(userName),
		AccessString: aws.String("on ~* &* +@all"),
		AuthenticationMode: &types.AuthenticationMode{
			Type:      types.InputAuthenticationTypePassword,
			Passwords: []string{"testpassword123"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteUser(t.Context(), &memorydb.DeleteUserInput{
			UserName: aws.String(userName),
		})
	})

	describeResult, err := client.DescribeUsers(ctx, &memorydb.DescribeUsersInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to describe users: %v", err)
	}

	if len(describeResult.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(describeResult.Users))
	}

	if *describeResult.Users[0].Name != userName {
		t.Errorf("expected user name %s, got %s", userName, *describeResult.Users[0].Name)
	}
}

func TestMemoryDB_CreateAndDeleteACL(t *testing.T) {
	client := newMemoryDBClient(t)
	ctx := t.Context()
	aclName := "test-acl"

	createResult, err := client.CreateACL(ctx, &memorydb.CreateACLInput{
		ACLName: aws.String(aclName),
	})
	if err != nil {
		t.Fatalf("failed to create ACL: %v", err)
	}

	if *createResult.ACL.Name != aclName {
		t.Errorf("expected ACL name %s, got %s", aclName, *createResult.ACL.Name)
	}

	if createResult.ACL.ARN == nil || *createResult.ACL.ARN == "" {
		t.Error("expected ACL ARN to be set")
	}

	// Delete
	deleteResult, err := client.DeleteACL(ctx, &memorydb.DeleteACLInput{
		ACLName: aws.String(aclName),
	})
	if err != nil {
		t.Fatalf("failed to delete ACL: %v", err)
	}

	if *deleteResult.ACL.Name != aclName {
		t.Errorf("expected ACL name %s, got %s", aclName, *deleteResult.ACL.Name)
	}
}

func TestMemoryDB_DescribeACLs(t *testing.T) {
	client := newMemoryDBClient(t)
	ctx := t.Context()
	aclName := "test-describe-acl"

	_, err := client.CreateACL(ctx, &memorydb.CreateACLInput{
		ACLName:   aws.String(aclName),
		UserNames: []string{"default"},
	})
	if err != nil {
		t.Fatalf("failed to create ACL: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteACL(t.Context(), &memorydb.DeleteACLInput{
			ACLName: aws.String(aclName),
		})
	})

	describeResult, err := client.DescribeACLs(ctx, &memorydb.DescribeACLsInput{
		ACLName: aws.String(aclName),
	})
	if err != nil {
		t.Fatalf("failed to describe ACLs: %v", err)
	}

	if len(describeResult.ACLs) != 1 {
		t.Fatalf("expected 1 ACL, got %d", len(describeResult.ACLs))
	}

	if *describeResult.ACLs[0].Name != aclName {
		t.Errorf("expected ACL name %s, got %s", aclName, *describeResult.ACLs[0].Name)
	}
}

func TestMemoryDB_ClusterNotFound(t *testing.T) {
	client := newMemoryDBClient(t)
	ctx := t.Context()

	_, err := client.DescribeClusters(ctx, &memorydb.DescribeClustersInput{
		ClusterName: aws.String("non-existent-cluster"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent cluster")
	}
}

func TestMemoryDB_DuplicateCluster(t *testing.T) {
	client := newMemoryDBClient(t)
	ctx := t.Context()
	clusterName := "test-dup-cluster"

	_, err := client.CreateCluster(ctx, &memorydb.CreateClusterInput{
		ClusterName: aws.String(clusterName),
		NodeType:    aws.String("db.r6g.large"),
		ACLName:     aws.String("open-access"),
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(t.Context(), &memorydb.DeleteClusterInput{
			ClusterName: aws.String(clusterName),
		})
	})

	_, err = client.CreateCluster(ctx, &memorydb.CreateClusterInput{
		ClusterName: aws.String(clusterName),
		NodeType:    aws.String("db.r6g.large"),
		ACLName:     aws.String("open-access"),
	})
	if err == nil {
		t.Fatal("expected error for duplicate cluster")
	}
}

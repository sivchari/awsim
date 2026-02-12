//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

func newEKSClient(t *testing.T) *eks.Client {
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

	return eks.NewFromConfig(cfg, func(o *eks.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566/eks")
	})
}

func TestEKS_CreateAndDeleteCluster(t *testing.T) {
	client := newEKSClient(t)
	ctx := t.Context()
	clusterName := "test-cluster"

	// Create cluster
	createResult, err := client.CreateCluster(ctx, &eks.CreateClusterInput{
		Name:    aws.String(clusterName),
		RoleArn: aws.String("arn:aws:iam::123456789012:role/eks-cluster-role"),
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: []string{"subnet-12345678", "subnet-87654321"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	if createResult.Cluster == nil {
		t.Fatal("expected cluster to be created")
	}

	if *createResult.Cluster.Name != clusterName {
		t.Errorf("expected cluster name %s, got %s", clusterName, *createResult.Cluster.Name)
	}

	if createResult.Cluster.Status != types.ClusterStatusActive {
		t.Errorf("expected cluster status ACTIVE, got %s", createResult.Cluster.Status)
	}

	// Delete cluster
	deleteResult, err := client.DeleteCluster(ctx, &eks.DeleteClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to delete cluster: %v", err)
	}

	if deleteResult.Cluster == nil {
		t.Fatal("expected cluster in delete response")
	}

	if deleteResult.Cluster.Status != types.ClusterStatusDeleting {
		t.Errorf("expected cluster status DELETING, got %s", deleteResult.Cluster.Status)
	}
}

func TestEKS_DescribeCluster(t *testing.T) {
	client := newEKSClient(t)
	ctx := t.Context()
	clusterName := "test-describe-cluster"

	// Create cluster
	_, err := client.CreateCluster(ctx, &eks.CreateClusterInput{
		Name:    aws.String(clusterName),
		RoleArn: aws.String("arn:aws:iam::123456789012:role/eks-cluster-role"),
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: []string{"subnet-12345678"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(ctx, &eks.DeleteClusterInput{
			Name: aws.String(clusterName),
		})
	})

	// Describe cluster
	describeResult, err := client.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to describe cluster: %v", err)
	}

	if describeResult.Cluster == nil {
		t.Fatal("expected cluster in describe response")
	}

	if *describeResult.Cluster.Name != clusterName {
		t.Errorf("expected cluster name %s, got %s", clusterName, *describeResult.Cluster.Name)
	}

	if describeResult.Cluster.Endpoint == nil || *describeResult.Cluster.Endpoint == "" {
		t.Error("expected cluster endpoint to be set")
	}

	if describeResult.Cluster.CertificateAuthority == nil || describeResult.Cluster.CertificateAuthority.Data == nil {
		t.Error("expected certificate authority data to be set")
	}
}

func TestEKS_ListClusters(t *testing.T) {
	client := newEKSClient(t)
	ctx := t.Context()
	clusterName := "test-list-cluster"

	// Create cluster
	_, err := client.CreateCluster(ctx, &eks.CreateClusterInput{
		Name:    aws.String(clusterName),
		RoleArn: aws.String("arn:aws:iam::123456789012:role/eks-cluster-role"),
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: []string{"subnet-12345678"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(ctx, &eks.DeleteClusterInput{
			Name: aws.String(clusterName),
		})
	})

	// List clusters
	listResult, err := client.ListClusters(ctx, &eks.ListClustersInput{})
	if err != nil {
		t.Fatalf("failed to list clusters: %v", err)
	}

	found := false
	for _, name := range listResult.Clusters {
		if name == clusterName {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected to find cluster %s in list", clusterName)
	}
}

func TestEKS_CreateAndDeleteNodegroup(t *testing.T) {
	client := newEKSClient(t)
	ctx := t.Context()
	clusterName := "test-nodegroup-cluster"
	nodegroupName := "test-nodegroup"

	// Create cluster first
	_, err := client.CreateCluster(ctx, &eks.CreateClusterInput{
		Name:    aws.String(clusterName),
		RoleArn: aws.String("arn:aws:iam::123456789012:role/eks-cluster-role"),
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: []string{"subnet-12345678", "subnet-87654321"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		// Delete nodegroup first, then cluster
		_, _ = client.DeleteNodegroup(ctx, &eks.DeleteNodegroupInput{
			ClusterName:   aws.String(clusterName),
			NodegroupName: aws.String(nodegroupName),
		})
		_, _ = client.DeleteCluster(ctx, &eks.DeleteClusterInput{
			Name: aws.String(clusterName),
		})
	})

	// Create nodegroup
	createResult, err := client.CreateNodegroup(ctx, &eks.CreateNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodegroupName),
		NodeRole:      aws.String("arn:aws:iam::123456789012:role/eks-nodegroup-role"),
		Subnets:       []string{"subnet-12345678", "subnet-87654321"},
		ScalingConfig: &types.NodegroupScalingConfig{
			MinSize:     aws.Int32(1),
			MaxSize:     aws.Int32(3),
			DesiredSize: aws.Int32(2),
		},
	})
	if err != nil {
		t.Fatalf("failed to create nodegroup: %v", err)
	}

	if createResult.Nodegroup == nil {
		t.Fatal("expected nodegroup to be created")
	}

	if *createResult.Nodegroup.NodegroupName != nodegroupName {
		t.Errorf("expected nodegroup name %s, got %s", nodegroupName, *createResult.Nodegroup.NodegroupName)
	}

	if createResult.Nodegroup.Status != types.NodegroupStatusActive {
		t.Errorf("expected nodegroup status ACTIVE, got %s", createResult.Nodegroup.Status)
	}

	// Delete nodegroup
	deleteResult, err := client.DeleteNodegroup(ctx, &eks.DeleteNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodegroupName),
	})
	if err != nil {
		t.Fatalf("failed to delete nodegroup: %v", err)
	}

	if deleteResult.Nodegroup == nil {
		t.Fatal("expected nodegroup in delete response")
	}

	if deleteResult.Nodegroup.Status != types.NodegroupStatusDeleting {
		t.Errorf("expected nodegroup status DELETING, got %s", deleteResult.Nodegroup.Status)
	}
}

func TestEKS_DescribeNodegroup(t *testing.T) {
	client := newEKSClient(t)
	ctx := t.Context()
	clusterName := "test-describe-nodegroup-cluster"
	nodegroupName := "test-describe-nodegroup"

	// Create cluster first
	_, err := client.CreateCluster(ctx, &eks.CreateClusterInput{
		Name:    aws.String(clusterName),
		RoleArn: aws.String("arn:aws:iam::123456789012:role/eks-cluster-role"),
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: []string{"subnet-12345678"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteNodegroup(ctx, &eks.DeleteNodegroupInput{
			ClusterName:   aws.String(clusterName),
			NodegroupName: aws.String(nodegroupName),
		})
		_, _ = client.DeleteCluster(ctx, &eks.DeleteClusterInput{
			Name: aws.String(clusterName),
		})
	})

	// Create nodegroup
	_, err = client.CreateNodegroup(ctx, &eks.CreateNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodegroupName),
		NodeRole:      aws.String("arn:aws:iam::123456789012:role/eks-nodegroup-role"),
		Subnets:       []string{"subnet-12345678"},
	})
	if err != nil {
		t.Fatalf("failed to create nodegroup: %v", err)
	}

	// Describe nodegroup
	describeResult, err := client.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodegroupName),
	})
	if err != nil {
		t.Fatalf("failed to describe nodegroup: %v", err)
	}

	if describeResult.Nodegroup == nil {
		t.Fatal("expected nodegroup in describe response")
	}

	if *describeResult.Nodegroup.NodegroupName != nodegroupName {
		t.Errorf("expected nodegroup name %s, got %s", nodegroupName, *describeResult.Nodegroup.NodegroupName)
	}

	if *describeResult.Nodegroup.ClusterName != clusterName {
		t.Errorf("expected cluster name %s, got %s", clusterName, *describeResult.Nodegroup.ClusterName)
	}
}

func TestEKS_ListNodegroups(t *testing.T) {
	client := newEKSClient(t)
	ctx := t.Context()
	clusterName := "test-list-nodegroups-cluster"
	nodegroupName := "test-list-nodegroup"

	// Create cluster first
	_, err := client.CreateCluster(ctx, &eks.CreateClusterInput{
		Name:    aws.String(clusterName),
		RoleArn: aws.String("arn:aws:iam::123456789012:role/eks-cluster-role"),
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: []string{"subnet-12345678"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteNodegroup(ctx, &eks.DeleteNodegroupInput{
			ClusterName:   aws.String(clusterName),
			NodegroupName: aws.String(nodegroupName),
		})
		_, _ = client.DeleteCluster(ctx, &eks.DeleteClusterInput{
			Name: aws.String(clusterName),
		})
	})

	// Create nodegroup
	_, err = client.CreateNodegroup(ctx, &eks.CreateNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodegroupName),
		NodeRole:      aws.String("arn:aws:iam::123456789012:role/eks-nodegroup-role"),
		Subnets:       []string{"subnet-12345678"},
	})
	if err != nil {
		t.Fatalf("failed to create nodegroup: %v", err)
	}

	// List nodegroups
	listResult, err := client.ListNodegroups(ctx, &eks.ListNodegroupsInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to list nodegroups: %v", err)
	}

	found := false
	for _, name := range listResult.Nodegroups {
		if name == nodegroupName {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected to find nodegroup %s in list", nodegroupName)
	}
}

func TestEKS_ClusterNotFound(t *testing.T) {
	client := newEKSClient(t)
	ctx := t.Context()

	// Try to describe non-existent cluster
	_, err := client.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String("non-existent-cluster"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent cluster")
	}
}

func TestEKS_NodegroupNotFound(t *testing.T) {
	client := newEKSClient(t)
	ctx := t.Context()
	clusterName := "test-nodegroup-not-found-cluster"

	// Create cluster first
	_, err := client.CreateCluster(ctx, &eks.CreateClusterInput{
		Name:    aws.String(clusterName),
		RoleArn: aws.String("arn:aws:iam::123456789012:role/eks-cluster-role"),
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: []string{"subnet-12345678"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(ctx, &eks.DeleteClusterInput{
			Name: aws.String(clusterName),
		})
	})

	// Try to describe non-existent nodegroup
	_, err = client.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String("non-existent-nodegroup"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent nodegroup")
	}
}

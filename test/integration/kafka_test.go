//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
)

func newKafkaClient(t *testing.T) *kafka.Client {
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

	return kafka.NewFromConfig(cfg, func(o *kafka.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566/kafka")
	})
}

func TestKafka_CreateAndDeleteCluster(t *testing.T) {
	client := newKafkaClient(t)
	ctx := t.Context()
	clusterName := "test-msk-cluster"

	createResult, err := client.CreateCluster(ctx, &kafka.CreateClusterInput{
		ClusterName:         aws.String(clusterName),
		KafkaVersion:        aws.String("3.6.0"),
		NumberOfBrokerNodes: aws.Int32(3),
		BrokerNodeGroupInfo: &types.BrokerNodeGroupInfo{
			ClientSubnets: []string{"subnet-12345678", "subnet-87654321"},
			InstanceType:  aws.String("kafka.m5.large"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	if createResult.ClusterArn == nil || *createResult.ClusterArn == "" {
		t.Fatal("expected cluster ARN to be set")
	}

	if *createResult.ClusterName != clusterName {
		t.Errorf("expected cluster name %s, got %s", clusterName, *createResult.ClusterName)
	}

	// Delete cluster
	deleteResult, err := client.DeleteCluster(ctx, &kafka.DeleteClusterInput{
		ClusterArn: createResult.ClusterArn,
	})
	if err != nil {
		t.Fatalf("failed to delete cluster: %v", err)
	}

	if *deleteResult.ClusterArn != *createResult.ClusterArn {
		t.Errorf("expected cluster ARN %s, got %s", *createResult.ClusterArn, *deleteResult.ClusterArn)
	}
}

func TestKafka_DescribeCluster(t *testing.T) {
	client := newKafkaClient(t)
	ctx := t.Context()
	clusterName := "test-describe-msk-cluster"

	createResult, err := client.CreateCluster(ctx, &kafka.CreateClusterInput{
		ClusterName:         aws.String(clusterName),
		KafkaVersion:        aws.String("3.6.0"),
		NumberOfBrokerNodes: aws.Int32(3),
		BrokerNodeGroupInfo: &types.BrokerNodeGroupInfo{
			ClientSubnets: []string{"subnet-12345678"},
			InstanceType:  aws.String("kafka.m5.large"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(context.Background(), &kafka.DeleteClusterInput{
			ClusterArn: createResult.ClusterArn,
		})
	})

	describeResult, err := client.DescribeCluster(ctx, &kafka.DescribeClusterInput{
		ClusterArn: createResult.ClusterArn,
	})
	if err != nil {
		t.Fatalf("failed to describe cluster: %v", err)
	}

	if describeResult.ClusterInfo == nil {
		t.Fatal("expected cluster info in describe response")
	}

	if *describeResult.ClusterInfo.ClusterName != clusterName {
		t.Errorf("expected cluster name %s, got %s", clusterName, *describeResult.ClusterInfo.ClusterName)
	}

	if describeResult.ClusterInfo.CurrentBrokerSoftwareInfo == nil || *describeResult.ClusterInfo.CurrentBrokerSoftwareInfo.KafkaVersion != "3.6.0" {
		t.Error("expected kafka version 3.6.0 in current broker software info")
	}

	if describeResult.ClusterInfo.State != types.ClusterStateActive {
		t.Errorf("expected cluster state ACTIVE, got %s", describeResult.ClusterInfo.State)
	}
}

func TestKafka_ListClusters(t *testing.T) {
	client := newKafkaClient(t)
	ctx := t.Context()
	clusterName := "test-list-msk-cluster"

	createResult, err := client.CreateCluster(ctx, &kafka.CreateClusterInput{
		ClusterName:         aws.String(clusterName),
		KafkaVersion:        aws.String("3.6.0"),
		NumberOfBrokerNodes: aws.Int32(3),
		BrokerNodeGroupInfo: &types.BrokerNodeGroupInfo{
			ClientSubnets: []string{"subnet-12345678"},
			InstanceType:  aws.String("kafka.m5.large"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(context.Background(), &kafka.DeleteClusterInput{
			ClusterArn: createResult.ClusterArn,
		})
	})

	listResult, err := client.ListClusters(ctx, &kafka.ListClustersInput{})
	if err != nil {
		t.Fatalf("failed to list clusters: %v", err)
	}

	found := false
	for _, c := range listResult.ClusterInfoList {
		if *c.ClusterName == clusterName {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("expected to find cluster %s in list", clusterName)
	}
}

func TestKafka_GetBootstrapBrokers(t *testing.T) {
	client := newKafkaClient(t)
	ctx := t.Context()
	clusterName := "test-bootstrap-msk-cluster"

	createResult, err := client.CreateCluster(ctx, &kafka.CreateClusterInput{
		ClusterName:         aws.String(clusterName),
		KafkaVersion:        aws.String("3.6.0"),
		NumberOfBrokerNodes: aws.Int32(3),
		BrokerNodeGroupInfo: &types.BrokerNodeGroupInfo{
			ClientSubnets: []string{"subnet-12345678"},
			InstanceType:  aws.String("kafka.m5.large"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(context.Background(), &kafka.DeleteClusterInput{
			ClusterArn: createResult.ClusterArn,
		})
	})

	bootstrapResult, err := client.GetBootstrapBrokers(ctx, &kafka.GetBootstrapBrokersInput{
		ClusterArn: createResult.ClusterArn,
	})
	if err != nil {
		t.Fatalf("failed to get bootstrap brokers: %v", err)
	}

	if bootstrapResult.BootstrapBrokerString == nil || *bootstrapResult.BootstrapBrokerString == "" {
		t.Error("expected bootstrap broker string to be set")
	}

	if bootstrapResult.BootstrapBrokerStringTls == nil || *bootstrapResult.BootstrapBrokerStringTls == "" {
		t.Error("expected bootstrap broker TLS string to be set")
	}
}

func TestKafka_ClusterNotFound(t *testing.T) {
	client := newKafkaClient(t)
	ctx := t.Context()

	_, err := client.DescribeCluster(ctx, &kafka.DescribeClusterInput{
		ClusterArn: aws.String("arn:aws:kafka:us-east-1:123456789012:cluster/non-existent/fake-uuid"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent cluster")
	}
}

func TestKafka_DuplicateCluster(t *testing.T) {
	client := newKafkaClient(t)
	ctx := t.Context()
	clusterName := "test-duplicate-msk-cluster"

	createResult, err := client.CreateCluster(ctx, &kafka.CreateClusterInput{
		ClusterName:         aws.String(clusterName),
		KafkaVersion:        aws.String("3.6.0"),
		NumberOfBrokerNodes: aws.Int32(3),
		BrokerNodeGroupInfo: &types.BrokerNodeGroupInfo{
			ClientSubnets: []string{"subnet-12345678"},
			InstanceType:  aws.String("kafka.m5.large"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(context.Background(), &kafka.DeleteClusterInput{
			ClusterArn: createResult.ClusterArn,
		})
	})

	_, err = client.CreateCluster(ctx, &kafka.CreateClusterInput{
		ClusterName:         aws.String(clusterName),
		KafkaVersion:        aws.String("3.6.0"),
		NumberOfBrokerNodes: aws.Int32(3),
		BrokerNodeGroupInfo: &types.BrokerNodeGroupInfo{
			ClientSubnets: []string{"subnet-12345678"},
			InstanceType:  aws.String("kafka.m5.large"),
		},
	})
	if err == nil {
		t.Fatal("expected error for duplicate cluster name")
	}
}

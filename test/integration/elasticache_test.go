//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
)

func newElastiCacheClient(t *testing.T) *elasticache.Client {
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

	return elasticache.NewFromConfig(cfg, func(o *elasticache.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestElastiCache_CreateAndDescribeCacheCluster(t *testing.T) {
	client := newElastiCacheClient(t)
	ctx := t.Context()

	clusterID := "test-cache-cluster"

	// Create cache cluster
	createResult, err := client.CreateCacheCluster(ctx, &elasticache.CreateCacheClusterInput{
		CacheClusterId: aws.String(clusterID),
		CacheNodeType:  aws.String("cache.t3.micro"),
		Engine:         aws.String("redis"),
		NumCacheNodes:  aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to create cache cluster: %v", err)
	}

	if createResult.CacheCluster == nil {
		t.Fatal("expected CacheCluster in response, got nil")
	}

	if *createResult.CacheCluster.CacheClusterId != clusterID {
		t.Errorf("expected identifier %s, got %s", clusterID, *createResult.CacheCluster.CacheClusterId)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCacheCluster(ctx, &elasticache.DeleteCacheClusterInput{
			CacheClusterId: aws.String(clusterID),
		})
	})

	// Describe cache clusters
	descResult, err := client.DescribeCacheClusters(ctx, &elasticache.DescribeCacheClustersInput{
		CacheClusterId: aws.String(clusterID),
	})
	if err != nil {
		t.Fatalf("failed to describe cache clusters: %v", err)
	}

	if len(descResult.CacheClusters) != 1 {
		t.Errorf("expected 1 cluster, got %d", len(descResult.CacheClusters))
	}

	if *descResult.CacheClusters[0].CacheClusterId != clusterID {
		t.Errorf("expected identifier %s, got %s", clusterID, *descResult.CacheClusters[0].CacheClusterId)
	}
}

func TestElastiCache_ModifyCacheCluster(t *testing.T) {
	client := newElastiCacheClient(t)
	ctx := t.Context()

	clusterID := "test-modify-cache-cluster"

	// Create cache cluster
	_, err := client.CreateCacheCluster(ctx, &elasticache.CreateCacheClusterInput{
		CacheClusterId: aws.String(clusterID),
		CacheNodeType:  aws.String("cache.t3.micro"),
		Engine:         aws.String("redis"),
		NumCacheNodes:  aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to create cache cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCacheCluster(ctx, &elasticache.DeleteCacheClusterInput{
			CacheClusterId: aws.String(clusterID),
		})
	})

	// Modify cache cluster
	modifyResult, err := client.ModifyCacheCluster(ctx, &elasticache.ModifyCacheClusterInput{
		CacheClusterId:   aws.String(clusterID),
		CacheNodeType:    aws.String("cache.t3.small"),
		ApplyImmediately: aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to modify cache cluster: %v", err)
	}

	if *modifyResult.CacheCluster.CacheNodeType != "cache.t3.small" {
		t.Errorf("expected node type cache.t3.small, got %s", *modifyResult.CacheCluster.CacheNodeType)
	}
}

func TestElastiCache_DeleteCacheCluster(t *testing.T) {
	client := newElastiCacheClient(t)
	ctx := t.Context()

	clusterID := "test-delete-cache-cluster"

	// Create cache cluster
	_, err := client.CreateCacheCluster(ctx, &elasticache.CreateCacheClusterInput{
		CacheClusterId: aws.String(clusterID),
		CacheNodeType:  aws.String("cache.t3.micro"),
		Engine:         aws.String("redis"),
		NumCacheNodes:  aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to create cache cluster: %v", err)
	}

	// Delete cache cluster
	deleteResult, err := client.DeleteCacheCluster(ctx, &elasticache.DeleteCacheClusterInput{
		CacheClusterId: aws.String(clusterID),
	})
	if err != nil {
		t.Fatalf("failed to delete cache cluster: %v", err)
	}

	if *deleteResult.CacheCluster.CacheClusterId != clusterID {
		t.Errorf("expected identifier %s, got %s", clusterID, *deleteResult.CacheCluster.CacheClusterId)
	}

	// Verify cluster is deleted
	_, err = client.DescribeCacheClusters(ctx, &elasticache.DescribeCacheClustersInput{
		CacheClusterId: aws.String(clusterID),
	})
	if err == nil {
		t.Error("expected error when describing deleted cluster, got nil")
	}
}

func TestElastiCache_CreateAndDescribeReplicationGroup(t *testing.T) {
	client := newElastiCacheClient(t)
	ctx := t.Context()

	groupID := "test-replication-group"

	// Create replication group
	createResult, err := client.CreateReplicationGroup(ctx, &elasticache.CreateReplicationGroupInput{
		ReplicationGroupId:          aws.String(groupID),
		ReplicationGroupDescription: aws.String("Test replication group"),
		CacheNodeType:               aws.String("cache.t3.micro"),
		Engine:                      aws.String("redis"),
		NumCacheClusters:            aws.Int32(2),
	})
	if err != nil {
		t.Fatalf("failed to create replication group: %v", err)
	}

	if createResult.ReplicationGroup == nil {
		t.Fatal("expected ReplicationGroup in response, got nil")
	}

	if *createResult.ReplicationGroup.ReplicationGroupId != groupID {
		t.Errorf("expected identifier %s, got %s", groupID, *createResult.ReplicationGroup.ReplicationGroupId)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteReplicationGroup(ctx, &elasticache.DeleteReplicationGroupInput{
			ReplicationGroupId: aws.String(groupID),
		})
	})

	// Describe replication groups
	descResult, err := client.DescribeReplicationGroups(ctx, &elasticache.DescribeReplicationGroupsInput{
		ReplicationGroupId: aws.String(groupID),
	})
	if err != nil {
		t.Fatalf("failed to describe replication groups: %v", err)
	}

	if len(descResult.ReplicationGroups) != 1 {
		t.Errorf("expected 1 group, got %d", len(descResult.ReplicationGroups))
	}

	if *descResult.ReplicationGroups[0].ReplicationGroupId != groupID {
		t.Errorf("expected identifier %s, got %s", groupID, *descResult.ReplicationGroups[0].ReplicationGroupId)
	}
}

func TestElastiCache_DeleteReplicationGroup(t *testing.T) {
	client := newElastiCacheClient(t)
	ctx := t.Context()

	groupID := "test-delete-replication-group"

	// Create replication group
	_, err := client.CreateReplicationGroup(ctx, &elasticache.CreateReplicationGroupInput{
		ReplicationGroupId:          aws.String(groupID),
		ReplicationGroupDescription: aws.String("Test replication group for deletion"),
		CacheNodeType:               aws.String("cache.t3.micro"),
		Engine:                      aws.String("redis"),
	})
	if err != nil {
		t.Fatalf("failed to create replication group: %v", err)
	}

	// Delete replication group
	deleteResult, err := client.DeleteReplicationGroup(ctx, &elasticache.DeleteReplicationGroupInput{
		ReplicationGroupId: aws.String(groupID),
	})
	if err != nil {
		t.Fatalf("failed to delete replication group: %v", err)
	}

	if *deleteResult.ReplicationGroup.ReplicationGroupId != groupID {
		t.Errorf("expected identifier %s, got %s", groupID, *deleteResult.ReplicationGroup.ReplicationGroupId)
	}

	// Verify group is deleted
	_, err = client.DescribeReplicationGroups(ctx, &elasticache.DescribeReplicationGroupsInput{
		ReplicationGroupId: aws.String(groupID),
	})
	if err == nil {
		t.Error("expected error when describing deleted group, got nil")
	}
}

func TestElastiCache_DescribeCacheClusters_All(t *testing.T) {
	client := newElastiCacheClient(t)
	ctx := t.Context()

	// Create multiple cache clusters
	clusterIDs := []string{"test-cache-cluster-1", "test-cache-cluster-2"}
	for _, id := range clusterIDs {
		_, err := client.CreateCacheCluster(ctx, &elasticache.CreateCacheClusterInput{
			CacheClusterId: aws.String(id),
			CacheNodeType:  aws.String("cache.t3.micro"),
			Engine:         aws.String("redis"),
			NumCacheNodes:  aws.Int32(1),
		})
		if err != nil {
			t.Fatalf("failed to create cache cluster %s: %v", id, err)
		}
	}

	t.Cleanup(func() {
		for _, id := range clusterIDs {
			_, _ = client.DeleteCacheCluster(ctx, &elasticache.DeleteCacheClusterInput{
				CacheClusterId: aws.String(id),
			})
		}
	})

	// Describe all cache clusters
	descResult, err := client.DescribeCacheClusters(ctx, &elasticache.DescribeCacheClustersInput{})
	if err != nil {
		t.Fatalf("failed to describe cache clusters: %v", err)
	}

	if len(descResult.CacheClusters) < 2 {
		t.Errorf("expected at least 2 clusters, got %d", len(descResult.CacheClusters))
	}
}

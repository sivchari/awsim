//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCloudFrontClient(t *testing.T) *cloudfront.Client {
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

	return cloudfront.NewFromConfig(cfg, func(o *cloudfront.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestCloudFront_CreateDistribution(t *testing.T) {
	t.Parallel()

	client := newCloudFrontClient(t)
	ctx := t.Context()

	result, err := client.CreateDistribution(ctx, &cloudfront.CreateDistributionInput{
		DistributionConfig: &types.DistributionConfig{
			CallerReference: aws.String("test-create-distribution"),
			Origins: &types.Origins{
				Quantity: aws.Int32(1),
				Items: []types.Origin{
					{
						Id:         aws.String("myS3Origin"),
						DomainName: aws.String("mybucket.s3.amazonaws.com"),
						S3OriginConfig: &types.S3OriginConfig{
							OriginAccessIdentity: aws.String(""),
						},
					},
				},
			},
			DefaultCacheBehavior: &types.DefaultCacheBehavior{
				TargetOriginId:       aws.String("myS3Origin"),
				ViewerProtocolPolicy: types.ViewerProtocolPolicyAllowAll,
				CachePolicyId:        aws.String("658327ea-f89d-4fab-a63d-7e88639e58f6"),
			},
			Comment: aws.String("Test distribution"),
			Enabled: aws.Bool(true),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Distribution.Id)
	assert.NotEmpty(t, result.Distribution.ARN)
	assert.Equal(t, "InProgress", *result.Distribution.Status)
	assert.NotEmpty(t, result.Distribution.DomainName)
	assert.NotEmpty(t, result.ETag)

	// Clean up.
	_, err = client.DeleteDistribution(ctx, &cloudfront.DeleteDistributionInput{
		Id:      result.Distribution.Id,
		IfMatch: result.ETag,
	})
	require.NoError(t, err)
}

func TestCloudFront_GetDistribution(t *testing.T) {
	t.Parallel()

	client := newCloudFrontClient(t)
	ctx := t.Context()

	// Create distribution first.
	createResult, err := client.CreateDistribution(ctx, &cloudfront.CreateDistributionInput{
		DistributionConfig: &types.DistributionConfig{
			CallerReference: aws.String("test-get-distribution"),
			Origins: &types.Origins{
				Quantity: aws.Int32(1),
				Items: []types.Origin{
					{
						Id:         aws.String("myS3Origin"),
						DomainName: aws.String("mybucket.s3.amazonaws.com"),
						S3OriginConfig: &types.S3OriginConfig{
							OriginAccessIdentity: aws.String(""),
						},
					},
				},
			},
			DefaultCacheBehavior: &types.DefaultCacheBehavior{
				TargetOriginId:       aws.String("myS3Origin"),
				ViewerProtocolPolicy: types.ViewerProtocolPolicyAllowAll,
				CachePolicyId:        aws.String("658327ea-f89d-4fab-a63d-7e88639e58f6"),
			},
			Comment: aws.String("Test distribution"),
			Enabled: aws.Bool(true),
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteDistribution(ctx, &cloudfront.DeleteDistributionInput{
			Id:      createResult.Distribution.Id,
			IfMatch: createResult.ETag,
		})
	})

	// Get distribution.
	getResult, err := client.GetDistribution(ctx, &cloudfront.GetDistributionInput{
		Id: createResult.Distribution.Id,
	})
	require.NoError(t, err)
	assert.Equal(t, *createResult.Distribution.Id, *getResult.Distribution.Id)
	assert.Equal(t, *createResult.Distribution.ARN, *getResult.Distribution.ARN)
	assert.NotEmpty(t, getResult.ETag)
}

func TestCloudFront_ListDistributions(t *testing.T) {
	t.Parallel()

	client := newCloudFrontClient(t)
	ctx := t.Context()

	// Create distribution first.
	createResult, err := client.CreateDistribution(ctx, &cloudfront.CreateDistributionInput{
		DistributionConfig: &types.DistributionConfig{
			CallerReference: aws.String("test-list-distributions"),
			Origins: &types.Origins{
				Quantity: aws.Int32(1),
				Items: []types.Origin{
					{
						Id:         aws.String("myS3Origin"),
						DomainName: aws.String("mybucket.s3.amazonaws.com"),
						S3OriginConfig: &types.S3OriginConfig{
							OriginAccessIdentity: aws.String(""),
						},
					},
				},
			},
			DefaultCacheBehavior: &types.DefaultCacheBehavior{
				TargetOriginId:       aws.String("myS3Origin"),
				ViewerProtocolPolicy: types.ViewerProtocolPolicyAllowAll,
				CachePolicyId:        aws.String("658327ea-f89d-4fab-a63d-7e88639e58f6"),
			},
			Comment: aws.String("Test distribution"),
			Enabled: aws.Bool(true),
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteDistribution(ctx, &cloudfront.DeleteDistributionInput{
			Id:      createResult.Distribution.Id,
			IfMatch: createResult.ETag,
		})
	})

	// List distributions.
	listResult, err := client.ListDistributions(ctx, &cloudfront.ListDistributionsInput{})
	require.NoError(t, err)
	require.NotNil(t, listResult)
	require.NotNil(t, listResult.DistributionList)

	// Find our distribution.
	found := false
	for _, dist := range listResult.DistributionList.Items {
		if *dist.Id == *createResult.Distribution.Id {
			found = true

			break
		}
	}
	assert.True(t, found, "Distribution should be in list")
}

func TestCloudFront_UpdateDistribution(t *testing.T) {
	t.Parallel()

	client := newCloudFrontClient(t)
	ctx := t.Context()

	// Create distribution first.
	createResult, err := client.CreateDistribution(ctx, &cloudfront.CreateDistributionInput{
		DistributionConfig: &types.DistributionConfig{
			CallerReference: aws.String("test-update-distribution"),
			Origins: &types.Origins{
				Quantity: aws.Int32(1),
				Items: []types.Origin{
					{
						Id:         aws.String("myS3Origin"),
						DomainName: aws.String("mybucket.s3.amazonaws.com"),
						S3OriginConfig: &types.S3OriginConfig{
							OriginAccessIdentity: aws.String(""),
						},
					},
				},
			},
			DefaultCacheBehavior: &types.DefaultCacheBehavior{
				TargetOriginId:       aws.String("myS3Origin"),
				ViewerProtocolPolicy: types.ViewerProtocolPolicyAllowAll,
				CachePolicyId:        aws.String("658327ea-f89d-4fab-a63d-7e88639e58f6"),
			},
			Comment: aws.String("Test distribution"),
			Enabled: aws.Bool(true),
		},
	})
	require.NoError(t, err)

	// Update distribution.
	updateResult, err := client.UpdateDistribution(ctx, &cloudfront.UpdateDistributionInput{
		Id:      createResult.Distribution.Id,
		IfMatch: createResult.ETag,
		DistributionConfig: &types.DistributionConfig{
			CallerReference: aws.String("test-update-distribution"),
			Origins: &types.Origins{
				Quantity: aws.Int32(1),
				Items: []types.Origin{
					{
						Id:         aws.String("myS3Origin"),
						DomainName: aws.String("mybucket.s3.amazonaws.com"),
						S3OriginConfig: &types.S3OriginConfig{
							OriginAccessIdentity: aws.String(""),
						},
					},
				},
			},
			DefaultCacheBehavior: &types.DefaultCacheBehavior{
				TargetOriginId:       aws.String("myS3Origin"),
				ViewerProtocolPolicy: types.ViewerProtocolPolicyAllowAll,
				CachePolicyId:        aws.String("658327ea-f89d-4fab-a63d-7e88639e58f6"),
			},
			Comment: aws.String("Updated comment"),
			Enabled: aws.Bool(true),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, updateResult)
	assert.NotEqual(t, *createResult.ETag, *updateResult.ETag, "ETag should change after update")

	// Clean up with new ETag.
	_, err = client.DeleteDistribution(ctx, &cloudfront.DeleteDistributionInput{
		Id:      createResult.Distribution.Id,
		IfMatch: updateResult.ETag,
	})
	require.NoError(t, err)
}

func TestCloudFront_CreateInvalidation(t *testing.T) {
	t.Parallel()

	client := newCloudFrontClient(t)
	ctx := t.Context()

	// Create distribution first.
	createResult, err := client.CreateDistribution(ctx, &cloudfront.CreateDistributionInput{
		DistributionConfig: &types.DistributionConfig{
			CallerReference: aws.String("test-create-invalidation"),
			Origins: &types.Origins{
				Quantity: aws.Int32(1),
				Items: []types.Origin{
					{
						Id:         aws.String("myS3Origin"),
						DomainName: aws.String("mybucket.s3.amazonaws.com"),
						S3OriginConfig: &types.S3OriginConfig{
							OriginAccessIdentity: aws.String(""),
						},
					},
				},
			},
			DefaultCacheBehavior: &types.DefaultCacheBehavior{
				TargetOriginId:       aws.String("myS3Origin"),
				ViewerProtocolPolicy: types.ViewerProtocolPolicyAllowAll,
				CachePolicyId:        aws.String("658327ea-f89d-4fab-a63d-7e88639e58f6"),
			},
			Comment: aws.String("Test distribution"),
			Enabled: aws.Bool(true),
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteDistribution(ctx, &cloudfront.DeleteDistributionInput{
			Id:      createResult.Distribution.Id,
			IfMatch: createResult.ETag,
		})
	})

	// Create invalidation.
	invResult, err := client.CreateInvalidation(ctx, &cloudfront.CreateInvalidationInput{
		DistributionId: createResult.Distribution.Id,
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String("test-invalidation-1"),
			Paths: &types.Paths{
				Quantity: aws.Int32(1),
				Items:    []string{"/*"},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, invResult)
	assert.NotEmpty(t, invResult.Invalidation.Id)
	assert.Equal(t, "InProgress", *invResult.Invalidation.Status)
}

func TestCloudFront_GetInvalidation(t *testing.T) {
	t.Parallel()

	client := newCloudFrontClient(t)
	ctx := t.Context()

	// Create distribution first.
	createResult, err := client.CreateDistribution(ctx, &cloudfront.CreateDistributionInput{
		DistributionConfig: &types.DistributionConfig{
			CallerReference: aws.String("test-get-invalidation"),
			Origins: &types.Origins{
				Quantity: aws.Int32(1),
				Items: []types.Origin{
					{
						Id:         aws.String("myS3Origin"),
						DomainName: aws.String("mybucket.s3.amazonaws.com"),
						S3OriginConfig: &types.S3OriginConfig{
							OriginAccessIdentity: aws.String(""),
						},
					},
				},
			},
			DefaultCacheBehavior: &types.DefaultCacheBehavior{
				TargetOriginId:       aws.String("myS3Origin"),
				ViewerProtocolPolicy: types.ViewerProtocolPolicyAllowAll,
				CachePolicyId:        aws.String("658327ea-f89d-4fab-a63d-7e88639e58f6"),
			},
			Comment: aws.String("Test distribution"),
			Enabled: aws.Bool(true),
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteDistribution(ctx, &cloudfront.DeleteDistributionInput{
			Id:      createResult.Distribution.Id,
			IfMatch: createResult.ETag,
		})
	})

	// Create invalidation.
	invResult, err := client.CreateInvalidation(ctx, &cloudfront.CreateInvalidationInput{
		DistributionId: createResult.Distribution.Id,
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String("test-get-invalidation-1"),
			Paths: &types.Paths{
				Quantity: aws.Int32(1),
				Items:    []string{"/images/*"},
			},
		},
	})
	require.NoError(t, err)

	// Get invalidation.
	getResult, err := client.GetInvalidation(ctx, &cloudfront.GetInvalidationInput{
		DistributionId: createResult.Distribution.Id,
		Id:             invResult.Invalidation.Id,
	})
	require.NoError(t, err)
	assert.Equal(t, *invResult.Invalidation.Id, *getResult.Invalidation.Id)
	assert.Equal(t, "InProgress", *getResult.Invalidation.Status)
}

package cloudfront_test

import (
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sivchari/awsim/internal/server"
	_ "github.com/sivchari/awsim/internal/service/cloudfront"
)

func TestCloudFrontSDK_CreateDistribution(t *testing.T) {
	cfg := server.DefaultConfig()
	srv := server.New(cfg)

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	awsCfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	client := cloudfront.NewFromConfig(awsCfg, func(o *cloudfront.Options) {
		o.BaseEndpoint = aws.String(ts.URL)
	})

	result, err := client.CreateDistribution(t.Context(), &cloudfront.CreateDistributionInput{
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
	assert.NotNil(t, result)
	assert.NotEmpty(t, *result.Distribution.Id)
	assert.NotEmpty(t, *result.Distribution.ARN)
	assert.Equal(t, "InProgress", *result.Distribution.Status)
	assert.NotEmpty(t, *result.ETag)

	// Clean up.
	_, err = client.DeleteDistribution(t.Context(), &cloudfront.DeleteDistributionInput{
		Id:      result.Distribution.Id,
		IfMatch: result.ETag,
	})
	require.NoError(t, err)
}

func TestCloudFrontSDK_GetDistribution(t *testing.T) {
	cfg := server.DefaultConfig()
	srv := server.New(cfg)

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	awsCfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	client := cloudfront.NewFromConfig(awsCfg, func(o *cloudfront.Options) {
		o.BaseEndpoint = aws.String(ts.URL)
	})

	// Create distribution first.
	createResult, err := client.CreateDistribution(t.Context(), &cloudfront.CreateDistributionInput{
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

	// Get distribution.
	getResult, err := client.GetDistribution(t.Context(), &cloudfront.GetDistributionInput{
		Id: createResult.Distribution.Id,
	})
	require.NoError(t, err)
	assert.Equal(t, *createResult.Distribution.Id, *getResult.Distribution.Id)

	// Clean up.
	_, err = client.DeleteDistribution(t.Context(), &cloudfront.DeleteDistributionInput{
		Id:      createResult.Distribution.Id,
		IfMatch: createResult.ETag,
	})
	require.NoError(t, err)
}

func TestCloudFrontSDK_ListDistributions(t *testing.T) {
	cfg := server.DefaultConfig()
	srv := server.New(cfg)

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	awsCfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	client := cloudfront.NewFromConfig(awsCfg, func(o *cloudfront.Options) {
		o.BaseEndpoint = aws.String(ts.URL)
	})

	// Create distribution first.
	createResult, err := client.CreateDistribution(t.Context(), &cloudfront.CreateDistributionInput{
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

	// List distributions.
	listResult, err := client.ListDistributions(t.Context(), &cloudfront.ListDistributionsInput{})
	require.NoError(t, err)
	assert.NotNil(t, listResult.DistributionList)
	assert.GreaterOrEqual(t, len(listResult.DistributionList.Items), 1)

	// Clean up.
	_, err = client.DeleteDistribution(t.Context(), &cloudfront.DeleteDistributionInput{
		Id:      createResult.Distribution.Id,
		IfMatch: createResult.ETag,
	})
	require.NoError(t, err)
}

//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newServiceQuotasClient(t *testing.T) *servicequotas.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	return servicequotas.NewFromConfig(cfg, func(o *servicequotas.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestServiceQuotas_ListServices(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	output, err := client.ListServices(ctx, &servicequotas.ListServicesInput{
		MaxResults: aws.Int32(10),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, output.Services)

	// Check that we have some expected services
	serviceNames := make(map[string]bool)
	for _, svc := range output.Services {
		serviceNames[*svc.ServiceCode] = true
	}

	// We should have some common services
	assert.True(t, len(serviceNames) > 0, "Should have at least one service")
}

func TestServiceQuotas_GetServiceQuota(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	// Get a known quota (EC2 Running On-Demand Standard instances)
	output, err := client.GetServiceQuota(ctx, &servicequotas.GetServiceQuotaInput{
		ServiceCode: aws.String("ec2"),
		QuotaCode:   aws.String("L-1216C47A"),
	})
	require.NoError(t, err)
	require.NotNil(t, output.Quota)
	assert.Equal(t, "ec2", *output.Quota.ServiceCode)
	assert.Equal(t, "L-1216C47A", *output.Quota.QuotaCode)
	assert.True(t, *output.Quota.Value > 0)
}

func TestServiceQuotas_GetServiceQuota_NotFound(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	// Try to get a non-existent quota
	_, err := client.GetServiceQuota(ctx, &servicequotas.GetServiceQuotaInput{
		ServiceCode: aws.String("ec2"),
		QuotaCode:   aws.String("nonexistent-quota"),
	})
	require.Error(t, err)
}

func TestServiceQuotas_ListServiceQuotas(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	output, err := client.ListServiceQuotas(ctx, &servicequotas.ListServiceQuotasInput{
		ServiceCode: aws.String("ec2"),
		MaxResults:  aws.Int32(10),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, output.Quotas)

	// Check that all returned quotas belong to EC2
	for _, quota := range output.Quotas {
		assert.Equal(t, "ec2", *quota.ServiceCode)
	}
}

func TestServiceQuotas_ListServiceQuotas_ServiceNotFound(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	_, err := client.ListServiceQuotas(ctx, &servicequotas.ListServiceQuotasInput{
		ServiceCode: aws.String("nonexistent-service"),
	})
	require.Error(t, err)
}

func TestServiceQuotas_GetAWSDefaultServiceQuota(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	output, err := client.GetAWSDefaultServiceQuota(ctx, &servicequotas.GetAWSDefaultServiceQuotaInput{
		ServiceCode: aws.String("lambda"),
		QuotaCode:   aws.String("L-B99A9384"),
	})
	require.NoError(t, err)
	require.NotNil(t, output.Quota)
	assert.Equal(t, "lambda", *output.Quota.ServiceCode)
	assert.Equal(t, "L-B99A9384", *output.Quota.QuotaCode)
}

func TestServiceQuotas_ListAWSDefaultServiceQuotas(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	output, err := client.ListAWSDefaultServiceQuotas(ctx, &servicequotas.ListAWSDefaultServiceQuotasInput{
		ServiceCode: aws.String("lambda"),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, output.Quotas)

	for _, quota := range output.Quotas {
		assert.Equal(t, "lambda", *quota.ServiceCode)
	}
}

func TestServiceQuotas_RequestServiceQuotaIncrease(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	// Request a quota increase for an adjustable quota
	output, err := client.RequestServiceQuotaIncrease(ctx, &servicequotas.RequestServiceQuotaIncreaseInput{
		ServiceCode:  aws.String("ec2"),
		QuotaCode:    aws.String("L-1216C47A"),
		DesiredValue: aws.Float64(2000),
	})
	require.NoError(t, err)
	require.NotNil(t, output.RequestedQuota)
	assert.NotEmpty(t, *output.RequestedQuota.Id)
	assert.Equal(t, "PENDING", string(output.RequestedQuota.Status))
	assert.Equal(t, float64(2000), *output.RequestedQuota.DesiredValue)
}

func TestServiceQuotas_RequestServiceQuotaIncrease_NonAdjustable(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	// Try to request increase for a non-adjustable quota
	_, err := client.RequestServiceQuotaIncrease(ctx, &servicequotas.RequestServiceQuotaIncreaseInput{
		ServiceCode:  aws.String("sqs"),
		QuotaCode:    aws.String("L-06F64E4A"),
		DesiredValue: aws.Float64(200000),
	})
	require.Error(t, err)
}

func TestServiceQuotas_GetRequestedServiceQuotaChange(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	// First, create a request
	createOutput, err := client.RequestServiceQuotaIncrease(ctx, &servicequotas.RequestServiceQuotaIncreaseInput{
		ServiceCode:  aws.String("ec2"),
		QuotaCode:    aws.String("L-0E3CBAB9"),
		DesiredValue: aws.Float64(10),
	})
	require.NoError(t, err)
	require.NotNil(t, createOutput.RequestedQuota)

	// Then get the request
	getOutput, err := client.GetRequestedServiceQuotaChange(ctx, &servicequotas.GetRequestedServiceQuotaChangeInput{
		RequestId: createOutput.RequestedQuota.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, getOutput.RequestedQuota)
	assert.Equal(t, *createOutput.RequestedQuota.Id, *getOutput.RequestedQuota.Id)
	assert.Equal(t, float64(10), *getOutput.RequestedQuota.DesiredValue)
}

func TestServiceQuotas_GetRequestedServiceQuotaChange_NotFound(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	_, err := client.GetRequestedServiceQuotaChange(ctx, &servicequotas.GetRequestedServiceQuotaChangeInput{
		RequestId: aws.String("nonexistent-request-id"),
	})
	require.Error(t, err)
}

func TestServiceQuotas_ListRequestedServiceQuotaChangeHistory(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	// First, create a request
	_, err := client.RequestServiceQuotaIncrease(ctx, &servicequotas.RequestServiceQuotaIncreaseInput{
		ServiceCode:  aws.String("dynamodb"),
		QuotaCode:    aws.String("L-F98FE922"),
		DesiredValue: aws.Float64(50000),
	})
	require.NoError(t, err)

	// Then list the requests
	output, err := client.ListRequestedServiceQuotaChangeHistory(ctx, &servicequotas.ListRequestedServiceQuotaChangeHistoryInput{
		ServiceCode: aws.String("dynamodb"),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, output.RequestedQuotas)

	// All returned requests should be for DynamoDB
	for _, req := range output.RequestedQuotas {
		assert.Equal(t, "dynamodb", *req.ServiceCode)
	}
}

func TestServiceQuotas_ListRequestedServiceQuotaChangeHistory_ByStatus(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	// Create a request
	_, err := client.RequestServiceQuotaIncrease(ctx, &servicequotas.RequestServiceQuotaIncreaseInput{
		ServiceCode:  aws.String("s3"),
		QuotaCode:    aws.String("L-DC2B2D3D"),
		DesiredValue: aws.Float64(200),
	})
	require.NoError(t, err)

	// List pending requests
	output, err := client.ListRequestedServiceQuotaChangeHistory(ctx, &servicequotas.ListRequestedServiceQuotaChangeHistoryInput{
		Status: "PENDING",
	})
	require.NoError(t, err)

	// All returned requests should be PENDING
	for _, req := range output.RequestedQuotas {
		assert.Equal(t, "PENDING", string(req.Status))
	}
}

func TestServiceQuotas_EndToEnd(t *testing.T) {
	client := newServiceQuotasClient(t)
	ctx := t.Context()

	// 1. List services
	servicesOutput, err := client.ListServices(ctx, &servicequotas.ListServicesInput{})
	require.NoError(t, err)
	assert.NotEmpty(t, servicesOutput.Services)

	// 2. Get quotas for a service
	quotasOutput, err := client.ListServiceQuotas(ctx, &servicequotas.ListServiceQuotasInput{
		ServiceCode: aws.String("ec2"),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, quotasOutput.Quotas)

	// 3. Get a specific quota
	var adjustableQuota *string
	var adjustableQuotaCode *string

	for _, q := range quotasOutput.Quotas {
		if q.Adjustable {
			adjustableQuota = q.QuotaCode
			adjustableQuotaCode = q.QuotaCode

			break
		}
	}

	require.NotNil(t, adjustableQuota, "Should have at least one adjustable quota")

	quotaOutput, err := client.GetServiceQuota(ctx, &servicequotas.GetServiceQuotaInput{
		ServiceCode: aws.String("ec2"),
		QuotaCode:   adjustableQuotaCode,
	})
	require.NoError(t, err)
	require.NotNil(t, quotaOutput.Quota)

	// 4. Request a quota increase
	requestOutput, err := client.RequestServiceQuotaIncrease(ctx, &servicequotas.RequestServiceQuotaIncreaseInput{
		ServiceCode:  aws.String("ec2"),
		QuotaCode:    adjustableQuotaCode,
		DesiredValue: aws.Float64(*quotaOutput.Quota.Value + 100),
	})
	require.NoError(t, err)
	require.NotNil(t, requestOutput.RequestedQuota)
	assert.Equal(t, "PENDING", string(requestOutput.RequestedQuota.Status))

	// 5. Get the request
	getRequestOutput, err := client.GetRequestedServiceQuotaChange(ctx, &servicequotas.GetRequestedServiceQuotaChangeInput{
		RequestId: requestOutput.RequestedQuota.Id,
	})
	require.NoError(t, err)
	assert.Equal(t, *requestOutput.RequestedQuota.Id, *getRequestOutput.RequestedQuota.Id)

	// 6. List request history
	historyOutput, err := client.ListRequestedServiceQuotaChangeHistory(ctx, &servicequotas.ListRequestedServiceQuotaChangeHistoryInput{
		ServiceCode: aws.String("ec2"),
	})
	require.NoError(t, err)

	// Should find our request in the history
	found := false

	for _, req := range historyOutput.RequestedQuotas {
		if *req.Id == *requestOutput.RequestedQuota.Id {
			found = true

			break
		}
	}

	assert.True(t, found, "Should find the created request in history")
}

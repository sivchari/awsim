//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/stretchr/testify/require"
)

func newConfigServiceClient(t *testing.T) *configservice.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	return configservice.NewFromConfig(cfg, func(o *configservice.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestConfigService_PutAndDeleteConfigurationRecorder(t *testing.T) {
	client := newConfigServiceClient(t)
	ctx := t.Context()

	recorderName := "test-recorder-create-delete"
	roleARN := "arn:aws:iam::123456789012:role/config-role"

	// Put configuration recorder.
	_, err := client.PutConfigurationRecorder(ctx, &configservice.PutConfigurationRecorderInput{
		ConfigurationRecorder: &types.ConfigurationRecorder{
			Name:    aws.String(recorderName),
			RoleARN: aws.String(roleARN),
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteConfigurationRecorder(ctx, &configservice.DeleteConfigurationRecorderInput{
			ConfigurationRecorderName: aws.String(recorderName),
		})
	})

	// Verify recorder was created.
	descOutput, err := client.DescribeConfigurationRecorders(ctx, &configservice.DescribeConfigurationRecordersInput{})
	require.NoError(t, err)
	require.Len(t, descOutput.ConfigurationRecorders, 1)
	require.Equal(t, recorderName, *descOutput.ConfigurationRecorders[0].Name)
	require.Equal(t, roleARN, *descOutput.ConfigurationRecorders[0].RoleARN)

	// Delete recorder.
	_, err = client.DeleteConfigurationRecorder(ctx, &configservice.DeleteConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String(recorderName),
	})
	require.NoError(t, err)

	// Verify recorder is deleted.
	descOutput, err = client.DescribeConfigurationRecorders(ctx, &configservice.DescribeConfigurationRecordersInput{})
	require.NoError(t, err)
	require.Empty(t, descOutput.ConfigurationRecorders)
}

func TestConfigService_DescribeConfigurationRecorders(t *testing.T) {
	client := newConfigServiceClient(t)
	ctx := t.Context()

	recorderName := "test-recorder-describe"
	roleARN := "arn:aws:iam::123456789012:role/config-role"

	// Put configuration recorder.
	_, err := client.PutConfigurationRecorder(ctx, &configservice.PutConfigurationRecorderInput{
		ConfigurationRecorder: &types.ConfigurationRecorder{
			Name:    aws.String(recorderName),
			RoleARN: aws.String(roleARN),
			RecordingGroup: &types.RecordingGroup{
				AllSupported: aws.Bool(true),
			},
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteConfigurationRecorder(ctx, &configservice.DeleteConfigurationRecorderInput{
			ConfigurationRecorderName: aws.String(recorderName),
		})
	})

	// Describe all recorders.
	descOutput, err := client.DescribeConfigurationRecorders(ctx, &configservice.DescribeConfigurationRecordersInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(descOutput.ConfigurationRecorders), 1)

	// Describe specific recorder.
	descOutput, err = client.DescribeConfigurationRecorders(ctx, &configservice.DescribeConfigurationRecordersInput{
		ConfigurationRecorderNames: []string{recorderName},
	})
	require.NoError(t, err)
	require.Len(t, descOutput.ConfigurationRecorders, 1)
	require.Equal(t, recorderName, *descOutput.ConfigurationRecorders[0].Name)
}

func TestConfigService_StartAndStopConfigurationRecorder(t *testing.T) {
	client := newConfigServiceClient(t)
	ctx := t.Context()

	recorderName := "test-recorder-start-stop"
	roleARN := "arn:aws:iam::123456789012:role/config-role"

	// Put configuration recorder.
	_, err := client.PutConfigurationRecorder(ctx, &configservice.PutConfigurationRecorderInput{
		ConfigurationRecorder: &types.ConfigurationRecorder{
			Name:    aws.String(recorderName),
			RoleARN: aws.String(roleARN),
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteConfigurationRecorder(ctx, &configservice.DeleteConfigurationRecorderInput{
			ConfigurationRecorderName: aws.String(recorderName),
		})
	})

	// Start recording.
	_, err = client.StartConfigurationRecorder(ctx, &configservice.StartConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String(recorderName),
	})
	require.NoError(t, err)

	// Stop recording.
	_, err = client.StopConfigurationRecorder(ctx, &configservice.StopConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String(recorderName),
	})
	require.NoError(t, err)
}

func TestConfigService_PutAndDeleteConfigRule(t *testing.T) {
	client := newConfigServiceClient(t)
	ctx := t.Context()

	ruleName := "test-rule-create-delete"

	// Put config rule.
	_, err := client.PutConfigRule(ctx, &configservice.PutConfigRuleInput{
		ConfigRule: &types.ConfigRule{
			ConfigRuleName: aws.String(ruleName),
			Source: &types.Source{
				Owner:            types.OwnerAws,
				SourceIdentifier: aws.String("S3_BUCKET_VERSIONING_ENABLED"),
			},
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteConfigRule(ctx, &configservice.DeleteConfigRuleInput{
			ConfigRuleName: aws.String(ruleName),
		})
	})

	// Verify rule was created.
	descOutput, err := client.DescribeConfigRules(ctx, &configservice.DescribeConfigRulesInput{
		ConfigRuleNames: []string{ruleName},
	})
	require.NoError(t, err)
	require.Len(t, descOutput.ConfigRules, 1)
	require.Equal(t, ruleName, *descOutput.ConfigRules[0].ConfigRuleName)

	// Delete rule.
	_, err = client.DeleteConfigRule(ctx, &configservice.DeleteConfigRuleInput{
		ConfigRuleName: aws.String(ruleName),
	})
	require.NoError(t, err)

	// Verify rule is deleted.
	descOutput, err = client.DescribeConfigRules(ctx, &configservice.DescribeConfigRulesInput{
		ConfigRuleNames: []string{ruleName},
	})
	require.NoError(t, err)
	require.Empty(t, descOutput.ConfigRules)
}

func TestConfigService_DescribeConfigRules(t *testing.T) {
	client := newConfigServiceClient(t)
	ctx := t.Context()

	ruleName := "test-rule-describe"

	// Put config rule.
	_, err := client.PutConfigRule(ctx, &configservice.PutConfigRuleInput{
		ConfigRule: &types.ConfigRule{
			ConfigRuleName: aws.String(ruleName),
			Description:    aws.String("Test rule for describe"),
			Source: &types.Source{
				Owner:            types.OwnerAws,
				SourceIdentifier: aws.String("S3_BUCKET_VERSIONING_ENABLED"),
			},
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteConfigRule(ctx, &configservice.DeleteConfigRuleInput{
			ConfigRuleName: aws.String(ruleName),
		})
	})

	// Describe all rules.
	descOutput, err := client.DescribeConfigRules(ctx, &configservice.DescribeConfigRulesInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(descOutput.ConfigRules), 1)

	// Describe specific rule.
	descOutput, err = client.DescribeConfigRules(ctx, &configservice.DescribeConfigRulesInput{
		ConfigRuleNames: []string{ruleName},
	})
	require.NoError(t, err)
	require.Len(t, descOutput.ConfigRules, 1)
	require.Equal(t, ruleName, *descOutput.ConfigRules[0].ConfigRuleName)
	require.Equal(t, "Test rule for describe", *descOutput.ConfigRules[0].Description)
}

func TestConfigService_GetComplianceDetailsByConfigRule(t *testing.T) {
	client := newConfigServiceClient(t)
	ctx := t.Context()

	ruleName := "test-rule-compliance"

	// Put config rule.
	_, err := client.PutConfigRule(ctx, &configservice.PutConfigRuleInput{
		ConfigRule: &types.ConfigRule{
			ConfigRuleName: aws.String(ruleName),
			Source: &types.Source{
				Owner:            types.OwnerAws,
				SourceIdentifier: aws.String("S3_BUCKET_VERSIONING_ENABLED"),
			},
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteConfigRule(ctx, &configservice.DeleteConfigRuleInput{
			ConfigRuleName: aws.String(ruleName),
		})
	})

	// GetComplianceDetailsByConfigRule returns empty list for MVP.
	output, err := client.GetComplianceDetailsByConfigRule(ctx, &configservice.GetComplianceDetailsByConfigRuleInput{
		ConfigRuleName: aws.String(ruleName),
	})
	require.NoError(t, err)
	require.Empty(t, output.EvaluationResults)
}

func TestConfigService_ConfigurationRecorderNotFound(t *testing.T) {
	client := newConfigServiceClient(t)
	ctx := t.Context()

	// Delete non-existent recorder.
	_, err := client.DeleteConfigurationRecorder(ctx, &configservice.DeleteConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String("non-existent-recorder"),
	})
	require.Error(t, err)

	// Start non-existent recorder.
	_, err = client.StartConfigurationRecorder(ctx, &configservice.StartConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String("non-existent-recorder"),
	})
	require.Error(t, err)

	// Stop non-existent recorder.
	_, err = client.StopConfigurationRecorder(ctx, &configservice.StopConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String("non-existent-recorder"),
	})
	require.Error(t, err)
}

func TestConfigService_ConfigRuleNotFound(t *testing.T) {
	client := newConfigServiceClient(t)
	ctx := t.Context()

	// Delete non-existent rule.
	_, err := client.DeleteConfigRule(ctx, &configservice.DeleteConfigRuleInput{
		ConfigRuleName: aws.String("non-existent-rule"),
	})
	require.Error(t, err)

	// Get compliance for non-existent rule.
	_, err = client.GetComplianceDetailsByConfigRule(ctx, &configservice.GetComplianceDetailsByConfigRuleInput{
		ConfigRuleName: aws.String("non-existent-rule"),
	})
	require.Error(t, err)
}

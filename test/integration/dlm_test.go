//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dlm"
	"github.com/aws/aws-sdk-go-v2/service/dlm/types"
	"github.com/stretchr/testify/require"
)

func newDLMClient(t *testing.T) *dlm.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	return dlm.NewFromConfig(cfg, func(o *dlm.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566/dlm")
	})
}

func TestDLM_CreateAndDeleteLifecyclePolicy(t *testing.T) {
	client := newDLMClient(t)
	ctx := t.Context()

	// Create lifecycle policy.
	createOutput, err := client.CreateLifecyclePolicy(ctx, &dlm.CreateLifecyclePolicyInput{
		Description:      aws.String("Test policy for EBS snapshots"),
		ExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/dlm-role"),
		State:            types.SettablePolicyStateValuesEnabled,
		PolicyDetails: &types.PolicyDetails{
			ResourceTypes: []types.ResourceTypeValues{types.ResourceTypeValuesVolume},
			TargetTags: []types.Tag{
				{
					Key:   aws.String("Backup"),
					Value: aws.String("true"),
				},
			},
			Schedules: []types.Schedule{
				{
					Name: aws.String("Daily snapshots"),
					CreateRule: &types.CreateRule{
						Interval:     aws.Int32(24),
						IntervalUnit: types.IntervalUnitValuesHours,
						Times:        []string{"03:00"},
					},
					RetainRule: &types.RetainRule{
						Count: aws.Int32(7),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.PolicyId)

	policyID := *createOutput.PolicyId

	t.Cleanup(func() {
		_, _ = client.DeleteLifecyclePolicy(ctx, &dlm.DeleteLifecyclePolicyInput{
			PolicyId: aws.String(policyID),
		})
	})

	// Get lifecycle policy.
	getOutput, err := client.GetLifecyclePolicy(ctx, &dlm.GetLifecyclePolicyInput{
		PolicyId: aws.String(policyID),
	})
	require.NoError(t, err)
	require.Equal(t, policyID, *getOutput.Policy.PolicyId)
	require.Equal(t, "Test policy for EBS snapshots", *getOutput.Policy.Description)
	require.Equal(t, types.GettablePolicyStateValuesEnabled, getOutput.Policy.State)

	// Delete lifecycle policy.
	_, err = client.DeleteLifecyclePolicy(ctx, &dlm.DeleteLifecyclePolicyInput{
		PolicyId: aws.String(policyID),
	})
	require.NoError(t, err)

	// Verify policy is deleted.
	_, err = client.GetLifecyclePolicy(ctx, &dlm.GetLifecyclePolicyInput{
		PolicyId: aws.String(policyID),
	})
	require.Error(t, err)
}

func TestDLM_UpdateLifecyclePolicy(t *testing.T) {
	client := newDLMClient(t)
	ctx := t.Context()

	// Create lifecycle policy.
	createOutput, err := client.CreateLifecyclePolicy(ctx, &dlm.CreateLifecyclePolicyInput{
		Description:      aws.String("Test policy"),
		ExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/dlm-role"),
		State:            types.SettablePolicyStateValuesEnabled,
		PolicyDetails: &types.PolicyDetails{
			ResourceTypes: []types.ResourceTypeValues{types.ResourceTypeValuesVolume},
			TargetTags: []types.Tag{
				{
					Key:   aws.String("Backup"),
					Value: aws.String("true"),
				},
			},
			Schedules: []types.Schedule{
				{
					Name: aws.String("Daily snapshots"),
					CreateRule: &types.CreateRule{
						Interval:     aws.Int32(24),
						IntervalUnit: types.IntervalUnitValuesHours,
					},
					RetainRule: &types.RetainRule{
						Count: aws.Int32(7),
					},
				},
			},
		},
	})
	require.NoError(t, err)

	policyID := *createOutput.PolicyId

	t.Cleanup(func() {
		_, _ = client.DeleteLifecyclePolicy(ctx, &dlm.DeleteLifecyclePolicyInput{
			PolicyId: aws.String(policyID),
		})
	})

	// Update lifecycle policy.
	_, err = client.UpdateLifecyclePolicy(ctx, &dlm.UpdateLifecyclePolicyInput{
		PolicyId:    aws.String(policyID),
		Description: aws.String("Updated test policy"),
		State:       types.SettablePolicyStateValuesDisabled,
	})
	require.NoError(t, err)

	// Verify update.
	getOutput, err := client.GetLifecyclePolicy(ctx, &dlm.GetLifecyclePolicyInput{
		PolicyId: aws.String(policyID),
	})
	require.NoError(t, err)
	require.Equal(t, "Updated test policy", *getOutput.Policy.Description)
	require.Equal(t, types.GettablePolicyStateValuesDisabled, getOutput.Policy.State)
}

func TestDLM_GetLifecyclePolicies(t *testing.T) {
	client := newDLMClient(t)
	ctx := t.Context()

	// Create multiple lifecycle policies.
	var policyIDs []string

	for i := 0; i < 2; i++ {
		createOutput, err := client.CreateLifecyclePolicy(ctx, &dlm.CreateLifecyclePolicyInput{
			Description:      aws.String("Test policy for listing"),
			ExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/dlm-role"),
			State:            types.SettablePolicyStateValuesEnabled,
			PolicyDetails: &types.PolicyDetails{
				ResourceTypes: []types.ResourceTypeValues{types.ResourceTypeValuesVolume},
				TargetTags: []types.Tag{
					{
						Key:   aws.String("Backup"),
						Value: aws.String("true"),
					},
				},
				Schedules: []types.Schedule{
					{
						Name: aws.String("Daily snapshots"),
						CreateRule: &types.CreateRule{
							Interval:     aws.Int32(24),
							IntervalUnit: types.IntervalUnitValuesHours,
						},
						RetainRule: &types.RetainRule{
							Count: aws.Int32(7),
						},
					},
				},
			},
		})
		require.NoError(t, err)
		policyIDs = append(policyIDs, *createOutput.PolicyId)
	}

	t.Cleanup(func() {
		for _, id := range policyIDs {
			_, _ = client.DeleteLifecyclePolicy(ctx, &dlm.DeleteLifecyclePolicyInput{
				PolicyId: aws.String(id),
			})
		}
	})

	// Get lifecycle policies.
	listOutput, err := client.GetLifecyclePolicies(ctx, &dlm.GetLifecyclePoliciesInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.Policies), 2)
}

func TestDLM_PolicyNotFound(t *testing.T) {
	client := newDLMClient(t)
	ctx := t.Context()

	// Get non-existent policy.
	_, err := client.GetLifecyclePolicy(ctx, &dlm.GetLifecyclePolicyInput{
		PolicyId: aws.String("policy-non-existent"),
	})
	require.Error(t, err)

	// Delete non-existent policy.
	_, err = client.DeleteLifecyclePolicy(ctx, &dlm.DeleteLifecyclePolicyInput{
		PolicyId: aws.String("policy-non-existent"),
	})
	require.Error(t, err)

	// Update non-existent policy.
	_, err = client.UpdateLifecyclePolicy(ctx, &dlm.UpdateLifecyclePolicyInput{
		PolicyId:    aws.String("policy-non-existent"),
		Description: aws.String("Updated description"),
	})
	require.Error(t, err)
}

func TestDLM_FilterPoliciesByState(t *testing.T) {
	client := newDLMClient(t)
	ctx := t.Context()

	// Create an enabled policy.
	createOutput, err := client.CreateLifecyclePolicy(ctx, &dlm.CreateLifecyclePolicyInput{
		Description:      aws.String("Enabled policy"),
		ExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/dlm-role"),
		State:            types.SettablePolicyStateValuesEnabled,
		PolicyDetails: &types.PolicyDetails{
			ResourceTypes: []types.ResourceTypeValues{types.ResourceTypeValuesVolume},
			TargetTags: []types.Tag{
				{
					Key:   aws.String("Backup"),
					Value: aws.String("true"),
				},
			},
			Schedules: []types.Schedule{
				{
					Name: aws.String("Daily"),
					CreateRule: &types.CreateRule{
						Interval:     aws.Int32(24),
						IntervalUnit: types.IntervalUnitValuesHours,
					},
					RetainRule: &types.RetainRule{
						Count: aws.Int32(7),
					},
				},
			},
		},
	})
	require.NoError(t, err)

	policyID := *createOutput.PolicyId

	t.Cleanup(func() {
		_, _ = client.DeleteLifecyclePolicy(ctx, &dlm.DeleteLifecyclePolicyInput{
			PolicyId: aws.String(policyID),
		})
	})

	// Filter by state ENABLED.
	listOutput, err := client.GetLifecyclePolicies(ctx, &dlm.GetLifecyclePoliciesInput{
		State: types.GettablePolicyStateValuesEnabled,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.Policies), 1)

	// Verify all returned policies are enabled.
	for _, policy := range listOutput.Policies {
		require.Equal(t, types.GettablePolicyStateValuesEnabled, policy.State)
	}
}

//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/stretchr/testify/require"
)

func newSchedulerClient(t *testing.T) *scheduler.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	return scheduler.NewFromConfig(cfg, func(o *scheduler.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestScheduler_CreateAndDeleteSchedule(t *testing.T) {
	client := newSchedulerClient(t)
	ctx := t.Context()

	scheduleName := "test-schedule"

	// Create schedule.
	createOutput, err := client.CreateSchedule(ctx, &scheduler.CreateScheduleInput{
		Name:               aws.String(scheduleName),
		ScheduleExpression: aws.String("rate(1 hour)"),
		FlexibleTimeWindow: &types.FlexibleTimeWindow{
			Mode: types.FlexibleTimeWindowModeOff,
		},
		Target: &types.Target{
			Arn:     aws.String("arn:aws:sqs:us-east-1:123456789012:my-queue"),
			RoleArn: aws.String("arn:aws:iam::123456789012:role/scheduler-role"),
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.ScheduleArn)

	t.Cleanup(func() {
		_, _ = client.DeleteSchedule(ctx, &scheduler.DeleteScheduleInput{
			Name: aws.String(scheduleName),
		})
	})

	// Get schedule.
	getOutput, err := client.GetSchedule(ctx, &scheduler.GetScheduleInput{
		Name: aws.String(scheduleName),
	})
	require.NoError(t, err)
	require.Equal(t, scheduleName, *getOutput.Name)
	require.Equal(t, "rate(1 hour)", *getOutput.ScheduleExpression)

	// Delete schedule.
	_, err = client.DeleteSchedule(ctx, &scheduler.DeleteScheduleInput{
		Name: aws.String(scheduleName),
	})
	require.NoError(t, err)

	// Verify schedule is deleted.
	_, err = client.GetSchedule(ctx, &scheduler.GetScheduleInput{
		Name: aws.String(scheduleName),
	})
	require.Error(t, err)
}

func TestScheduler_UpdateSchedule(t *testing.T) {
	client := newSchedulerClient(t)
	ctx := t.Context()

	scheduleName := "test-schedule-update"

	// Create schedule.
	_, err := client.CreateSchedule(ctx, &scheduler.CreateScheduleInput{
		Name:               aws.String(scheduleName),
		ScheduleExpression: aws.String("rate(1 hour)"),
		FlexibleTimeWindow: &types.FlexibleTimeWindow{
			Mode: types.FlexibleTimeWindowModeOff,
		},
		Target: &types.Target{
			Arn:     aws.String("arn:aws:sqs:us-east-1:123456789012:my-queue"),
			RoleArn: aws.String("arn:aws:iam::123456789012:role/scheduler-role"),
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteSchedule(ctx, &scheduler.DeleteScheduleInput{
			Name: aws.String(scheduleName),
		})
	})

	// Update schedule.
	_, err = client.UpdateSchedule(ctx, &scheduler.UpdateScheduleInput{
		Name:               aws.String(scheduleName),
		ScheduleExpression: aws.String("rate(2 hours)"),
		FlexibleTimeWindow: &types.FlexibleTimeWindow{
			Mode: types.FlexibleTimeWindowModeOff,
		},
		Target: &types.Target{
			Arn:     aws.String("arn:aws:sqs:us-east-1:123456789012:my-queue-updated"),
			RoleArn: aws.String("arn:aws:iam::123456789012:role/scheduler-role"),
		},
	})
	require.NoError(t, err)

	// Verify update.
	getOutput, err := client.GetSchedule(ctx, &scheduler.GetScheduleInput{
		Name: aws.String(scheduleName),
	})
	require.NoError(t, err)
	require.Equal(t, "rate(2 hours)", *getOutput.ScheduleExpression)
	require.Equal(t, "arn:aws:sqs:us-east-1:123456789012:my-queue-updated", *getOutput.Target.Arn)
}

func TestScheduler_ListSchedules(t *testing.T) {
	client := newSchedulerClient(t)
	ctx := t.Context()

	// Create multiple schedules.
	scheduleNames := []string{"test-list-schedule-1", "test-list-schedule-2"}

	for _, name := range scheduleNames {
		_, err := client.CreateSchedule(ctx, &scheduler.CreateScheduleInput{
			Name:               aws.String(name),
			ScheduleExpression: aws.String("rate(1 hour)"),
			FlexibleTimeWindow: &types.FlexibleTimeWindow{
				Mode: types.FlexibleTimeWindowModeOff,
			},
			Target: &types.Target{
				Arn:     aws.String("arn:aws:sqs:us-east-1:123456789012:my-queue"),
				RoleArn: aws.String("arn:aws:iam::123456789012:role/scheduler-role"),
			},
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, name := range scheduleNames {
			_, _ = client.DeleteSchedule(ctx, &scheduler.DeleteScheduleInput{
				Name: aws.String(name),
			})
		}
	})

	// List schedules.
	listOutput, err := client.ListSchedules(ctx, &scheduler.ListSchedulesInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.Schedules), 2)
}

func TestScheduler_CreateAndDeleteScheduleGroup(t *testing.T) {
	client := newSchedulerClient(t)
	ctx := t.Context()

	groupName := "test-schedule-group"

	// Create schedule group.
	createOutput, err := client.CreateScheduleGroup(ctx, &scheduler.CreateScheduleGroupInput{
		Name: aws.String(groupName),
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.ScheduleGroupArn)

	t.Cleanup(func() {
		_, _ = client.DeleteScheduleGroup(ctx, &scheduler.DeleteScheduleGroupInput{
			Name: aws.String(groupName),
		})
	})

	// Get schedule group.
	getOutput, err := client.GetScheduleGroup(ctx, &scheduler.GetScheduleGroupInput{
		Name: aws.String(groupName),
	})
	require.NoError(t, err)
	require.Equal(t, groupName, *getOutput.Name)

	// Delete schedule group.
	_, err = client.DeleteScheduleGroup(ctx, &scheduler.DeleteScheduleGroupInput{
		Name: aws.String(groupName),
	})
	require.NoError(t, err)

	// Verify schedule group is deleted.
	_, err = client.GetScheduleGroup(ctx, &scheduler.GetScheduleGroupInput{
		Name: aws.String(groupName),
	})
	require.Error(t, err)
}

func TestScheduler_ListScheduleGroups(t *testing.T) {
	client := newSchedulerClient(t)
	ctx := t.Context()

	// List schedule groups (should include default group).
	listOutput, err := client.ListScheduleGroups(ctx, &scheduler.ListScheduleGroupsInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.ScheduleGroups), 1)

	// Check that default group exists.
	found := false

	for _, group := range listOutput.ScheduleGroups {
		if *group.Name == "default" {
			found = true

			break
		}
	}

	require.True(t, found, "default schedule group should exist")
}

func TestScheduler_ScheduleWithGroup(t *testing.T) {
	client := newSchedulerClient(t)
	ctx := t.Context()

	groupName := "test-schedule-group-with-schedule"
	scheduleName := "test-schedule-in-group"

	// Create schedule group.
	_, err := client.CreateScheduleGroup(ctx, &scheduler.CreateScheduleGroupInput{
		Name: aws.String(groupName),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteSchedule(ctx, &scheduler.DeleteScheduleInput{
			Name:      aws.String(scheduleName),
			GroupName: aws.String(groupName),
		})
		_, _ = client.DeleteScheduleGroup(ctx, &scheduler.DeleteScheduleGroupInput{
			Name: aws.String(groupName),
		})
	})

	// Create schedule in the group.
	_, err = client.CreateSchedule(ctx, &scheduler.CreateScheduleInput{
		Name:               aws.String(scheduleName),
		GroupName:          aws.String(groupName),
		ScheduleExpression: aws.String("rate(1 hour)"),
		FlexibleTimeWindow: &types.FlexibleTimeWindow{
			Mode: types.FlexibleTimeWindowModeOff,
		},
		Target: &types.Target{
			Arn:     aws.String("arn:aws:sqs:us-east-1:123456789012:my-queue"),
			RoleArn: aws.String("arn:aws:iam::123456789012:role/scheduler-role"),
		},
	})
	require.NoError(t, err)

	// Get schedule with group name.
	getOutput, err := client.GetSchedule(ctx, &scheduler.GetScheduleInput{
		Name:      aws.String(scheduleName),
		GroupName: aws.String(groupName),
	})
	require.NoError(t, err)
	require.Equal(t, scheduleName, *getOutput.Name)
	require.Equal(t, groupName, *getOutput.GroupName)
}

func TestScheduler_ScheduleNotFound(t *testing.T) {
	client := newSchedulerClient(t)
	ctx := t.Context()

	// Get non-existent schedule.
	_, err := client.GetSchedule(ctx, &scheduler.GetScheduleInput{
		Name: aws.String("non-existent-schedule"),
	})
	require.Error(t, err)

	// Delete non-existent schedule.
	_, err = client.DeleteSchedule(ctx, &scheduler.DeleteScheduleInput{
		Name: aws.String("non-existent-schedule"),
	})
	require.Error(t, err)
}

func TestScheduler_ScheduleGroupNotFound(t *testing.T) {
	client := newSchedulerClient(t)
	ctx := t.Context()

	// Get non-existent schedule group.
	_, err := client.GetScheduleGroup(ctx, &scheduler.GetScheduleGroupInput{
		Name: aws.String("non-existent-group"),
	})
	require.Error(t, err)

	// Delete non-existent schedule group.
	_, err = client.DeleteScheduleGroup(ctx, &scheduler.DeleteScheduleGroupInput{
		Name: aws.String("non-existent-group"),
	})
	require.Error(t, err)
}

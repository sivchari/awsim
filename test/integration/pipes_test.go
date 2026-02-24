//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/stretchr/testify/require"
)

func newPipesClient(t *testing.T) *pipes.Client {
	t.Helper()

	cfg := newAWSConfig(t)

	return pipes.NewFromConfig(cfg, func(o *pipes.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
}

func TestPipes_CreateAndDescribePipe(t *testing.T) {
	client := newPipesClient(t)
	ctx := t.Context()

	pipeName := "test-pipe-create-describe"
	source := "arn:aws:sqs:us-east-1:123456789012:test-source-queue"
	target := "arn:aws:lambda:us-east-1:123456789012:function:test-target"
	roleArn := "arn:aws:iam::123456789012:role/test-pipe-role"

	// Create pipe.
	createOutput, err := client.CreatePipe(ctx, &pipes.CreatePipeInput{
		Name:    aws.String(pipeName),
		Source:  aws.String(source),
		Target:  aws.String(target),
		RoleArn: aws.String(roleArn),
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.Arn)
	require.Equal(t, pipeName, *createOutput.Name)
	require.Equal(t, types.RequestedPipeStateRunning, createOutput.DesiredState)

	t.Cleanup(func() {
		_, _ = client.DeletePipe(ctx, &pipes.DeletePipeInput{
			Name: aws.String(pipeName),
		})
	})

	// Describe pipe.
	descOutput, err := client.DescribePipe(ctx, &pipes.DescribePipeInput{
		Name: aws.String(pipeName),
	})
	require.NoError(t, err)
	require.Equal(t, pipeName, *descOutput.Name)
	require.Equal(t, source, *descOutput.Source)
	require.Equal(t, target, *descOutput.Target)
	require.Equal(t, roleArn, *descOutput.RoleArn)
}

func TestPipes_CreatePipeWithStoppedState(t *testing.T) {
	client := newPipesClient(t)
	ctx := t.Context()

	pipeName := "test-pipe-stopped"
	source := "arn:aws:sqs:us-east-1:123456789012:test-source-queue"
	target := "arn:aws:lambda:us-east-1:123456789012:function:test-target"
	roleArn := "arn:aws:iam::123456789012:role/test-pipe-role"

	// Create pipe with STOPPED state.
	createOutput, err := client.CreatePipe(ctx, &pipes.CreatePipeInput{
		Name:         aws.String(pipeName),
		Source:       aws.String(source),
		Target:       aws.String(target),
		RoleArn:      aws.String(roleArn),
		DesiredState: types.RequestedPipeStateStopped,
	})
	require.NoError(t, err)
	require.Equal(t, types.RequestedPipeStateStopped, createOutput.DesiredState)

	t.Cleanup(func() {
		_, _ = client.DeletePipe(ctx, &pipes.DeletePipeInput{
			Name: aws.String(pipeName),
		})
	})

	// Describe pipe to verify state.
	descOutput, err := client.DescribePipe(ctx, &pipes.DescribePipeInput{
		Name: aws.String(pipeName),
	})
	require.NoError(t, err)
	require.Equal(t, types.PipeStateStopped, descOutput.CurrentState)
}

func TestPipes_UpdatePipe(t *testing.T) {
	client := newPipesClient(t)
	ctx := t.Context()

	pipeName := "test-pipe-update"
	source := "arn:aws:sqs:us-east-1:123456789012:test-source-queue"
	target := "arn:aws:lambda:us-east-1:123456789012:function:test-target"
	roleArn := "arn:aws:iam::123456789012:role/test-pipe-role"
	newRoleArn := "arn:aws:iam::123456789012:role/test-pipe-role-updated"
	description := "Updated description"

	// Create pipe.
	_, err := client.CreatePipe(ctx, &pipes.CreatePipeInput{
		Name:    aws.String(pipeName),
		Source:  aws.String(source),
		Target:  aws.String(target),
		RoleArn: aws.String(roleArn),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeletePipe(ctx, &pipes.DeletePipeInput{
			Name: aws.String(pipeName),
		})
	})

	// Update pipe.
	updateOutput, err := client.UpdatePipe(ctx, &pipes.UpdatePipeInput{
		Name:        aws.String(pipeName),
		RoleArn:     aws.String(newRoleArn),
		Description: aws.String(description),
	})
	require.NoError(t, err)
	require.Equal(t, pipeName, *updateOutput.Name)

	// Verify update.
	descOutput, err := client.DescribePipe(ctx, &pipes.DescribePipeInput{
		Name: aws.String(pipeName),
	})
	require.NoError(t, err)
	require.Equal(t, newRoleArn, *descOutput.RoleArn)
	require.Equal(t, description, *descOutput.Description)
}

func TestPipes_DeletePipe(t *testing.T) {
	client := newPipesClient(t)
	ctx := t.Context()

	pipeName := "test-pipe-delete"
	source := "arn:aws:sqs:us-east-1:123456789012:test-source-queue"
	target := "arn:aws:lambda:us-east-1:123456789012:function:test-target"
	roleArn := "arn:aws:iam::123456789012:role/test-pipe-role"

	// Create pipe.
	_, err := client.CreatePipe(ctx, &pipes.CreatePipeInput{
		Name:    aws.String(pipeName),
		Source:  aws.String(source),
		Target:  aws.String(target),
		RoleArn: aws.String(roleArn),
	})
	require.NoError(t, err)

	// Delete pipe.
	deleteOutput, err := client.DeletePipe(ctx, &pipes.DeletePipeInput{
		Name: aws.String(pipeName),
	})
	require.NoError(t, err)
	require.Equal(t, pipeName, *deleteOutput.Name)

	// Verify pipe is deleted.
	_, err = client.DescribePipe(ctx, &pipes.DescribePipeInput{
		Name: aws.String(pipeName),
	})
	require.Error(t, err)
}

func TestPipes_ListPipes(t *testing.T) {
	client := newPipesClient(t)
	ctx := t.Context()

	// Create multiple pipes.
	pipeNames := []string{"test-pipe-list-1", "test-pipe-list-2", "test-pipe-list-3"}
	source := "arn:aws:sqs:us-east-1:123456789012:test-source-queue"
	target := "arn:aws:lambda:us-east-1:123456789012:function:test-target"
	roleArn := "arn:aws:iam::123456789012:role/test-pipe-role"

	for _, name := range pipeNames {
		_, err := client.CreatePipe(ctx, &pipes.CreatePipeInput{
			Name:    aws.String(name),
			Source:  aws.String(source),
			Target:  aws.String(target),
			RoleArn: aws.String(roleArn),
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, name := range pipeNames {
			_, _ = client.DeletePipe(ctx, &pipes.DeletePipeInput{
				Name: aws.String(name),
			})
		}
	})

	// List pipes.
	listOutput, err := client.ListPipes(ctx, &pipes.ListPipesInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.Pipes), 3)

	// List pipes with name prefix filter.
	listOutput, err = client.ListPipes(ctx, &pipes.ListPipesInput{
		NamePrefix: aws.String("test-pipe-list-"),
	})
	require.NoError(t, err)
	require.Len(t, listOutput.Pipes, 3)
}

func TestPipes_StartAndStopPipe(t *testing.T) {
	client := newPipesClient(t)
	ctx := t.Context()

	pipeName := "test-pipe-start-stop"
	source := "arn:aws:sqs:us-east-1:123456789012:test-source-queue"
	target := "arn:aws:lambda:us-east-1:123456789012:function:test-target"
	roleArn := "arn:aws:iam::123456789012:role/test-pipe-role"

	// Create pipe with STOPPED state.
	_, err := client.CreatePipe(ctx, &pipes.CreatePipeInput{
		Name:         aws.String(pipeName),
		Source:       aws.String(source),
		Target:       aws.String(target),
		RoleArn:      aws.String(roleArn),
		DesiredState: types.RequestedPipeStateStopped,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeletePipe(ctx, &pipes.DeletePipeInput{
			Name: aws.String(pipeName),
		})
	})

	// Start pipe.
	startOutput, err := client.StartPipe(ctx, &pipes.StartPipeInput{
		Name: aws.String(pipeName),
	})
	require.NoError(t, err)
	require.Equal(t, types.RequestedPipeStateRunning, startOutput.DesiredState)

	// Verify pipe is running.
	descOutput, err := client.DescribePipe(ctx, &pipes.DescribePipeInput{
		Name: aws.String(pipeName),
	})
	require.NoError(t, err)
	require.Equal(t, types.PipeStateRunning, descOutput.CurrentState)

	// Stop pipe.
	stopOutput, err := client.StopPipe(ctx, &pipes.StopPipeInput{
		Name: aws.String(pipeName),
	})
	require.NoError(t, err)
	require.Equal(t, types.RequestedPipeStateStopped, stopOutput.DesiredState)

	// Verify pipe is stopped.
	descOutput, err = client.DescribePipe(ctx, &pipes.DescribePipeInput{
		Name: aws.String(pipeName),
	})
	require.NoError(t, err)
	require.Equal(t, types.PipeStateStopped, descOutput.CurrentState)
}

func TestPipes_TagOperations(t *testing.T) {
	client := newPipesClient(t)
	ctx := t.Context()

	pipeName := "test-pipe-tags"
	source := "arn:aws:sqs:us-east-1:123456789012:test-source-queue"
	target := "arn:aws:lambda:us-east-1:123456789012:function:test-target"
	roleArn := "arn:aws:iam::123456789012:role/test-pipe-role"

	// Create pipe with tags.
	initialTags := map[string]string{
		"Environment": "test",
		"Project":     "awsim",
	}
	createOutput, err := client.CreatePipe(ctx, &pipes.CreatePipeInput{
		Name:    aws.String(pipeName),
		Source:  aws.String(source),
		Target:  aws.String(target),
		RoleArn: aws.String(roleArn),
		Tags:    initialTags,
	})
	require.NoError(t, err)

	pipeArn := createOutput.Arn

	t.Cleanup(func() {
		_, _ = client.DeletePipe(ctx, &pipes.DeletePipeInput{
			Name: aws.String(pipeName),
		})
	})

	// List tags.
	listTagsOutput, err := client.ListTagsForResource(ctx, &pipes.ListTagsForResourceInput{
		ResourceArn: pipeArn,
	})
	require.NoError(t, err)
	require.Equal(t, "test", listTagsOutput.Tags["Environment"])
	require.Equal(t, "awsim", listTagsOutput.Tags["Project"])

	// Add more tags.
	_, err = client.TagResource(ctx, &pipes.TagResourceInput{
		ResourceArn: pipeArn,
		Tags: map[string]string{
			"NewTag": "newvalue",
		},
	})
	require.NoError(t, err)

	// Verify tags were added.
	listTagsOutput, err = client.ListTagsForResource(ctx, &pipes.ListTagsForResourceInput{
		ResourceArn: pipeArn,
	})
	require.NoError(t, err)
	require.Equal(t, "newvalue", listTagsOutput.Tags["NewTag"])

	// Remove a tag.
	_, err = client.UntagResource(ctx, &pipes.UntagResourceInput{
		ResourceArn: pipeArn,
		TagKeys:     []string{"NewTag"},
	})
	require.NoError(t, err)

	// Verify tag was removed.
	listTagsOutput, err = client.ListTagsForResource(ctx, &pipes.ListTagsForResourceInput{
		ResourceArn: pipeArn,
	})
	require.NoError(t, err)
	_, exists := listTagsOutput.Tags["NewTag"]
	require.False(t, exists)
}

func TestPipes_NotFoundErrors(t *testing.T) {
	client := newPipesClient(t)
	ctx := t.Context()

	// Describe non-existent pipe.
	_, err := client.DescribePipe(ctx, &pipes.DescribePipeInput{
		Name: aws.String("non-existent-pipe"),
	})
	require.Error(t, err)

	// Update non-existent pipe.
	_, err = client.UpdatePipe(ctx, &pipes.UpdatePipeInput{
		Name:    aws.String("non-existent-pipe"),
		RoleArn: aws.String("arn:aws:iam::123456789012:role/test-role"),
	})
	require.Error(t, err)

	// Delete non-existent pipe.
	_, err = client.DeletePipe(ctx, &pipes.DeletePipeInput{
		Name: aws.String("non-existent-pipe"),
	})
	require.Error(t, err)

	// Start non-existent pipe.
	_, err = client.StartPipe(ctx, &pipes.StartPipeInput{
		Name: aws.String("non-existent-pipe"),
	})
	require.Error(t, err)

	// Stop non-existent pipe.
	_, err = client.StopPipe(ctx, &pipes.StopPipeInput{
		Name: aws.String("non-existent-pipe"),
	})
	require.Error(t, err)
}

func TestPipes_ConflictErrors(t *testing.T) {
	client := newPipesClient(t)
	ctx := t.Context()

	pipeName := "test-pipe-conflict"
	source := "arn:aws:sqs:us-east-1:123456789012:test-source-queue"
	target := "arn:aws:lambda:us-east-1:123456789012:function:test-target"
	roleArn := "arn:aws:iam::123456789012:role/test-pipe-role"

	// Create pipe.
	_, err := client.CreatePipe(ctx, &pipes.CreatePipeInput{
		Name:    aws.String(pipeName),
		Source:  aws.String(source),
		Target:  aws.String(target),
		RoleArn: aws.String(roleArn),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeletePipe(ctx, &pipes.DeletePipeInput{
			Name: aws.String(pipeName),
		})
	})

	// Try to create duplicate pipe.
	_, err = client.CreatePipe(ctx, &pipes.CreatePipeInput{
		Name:    aws.String(pipeName),
		Source:  aws.String(source),
		Target:  aws.String(target),
		RoleArn: aws.String(roleArn),
	})
	require.Error(t, err)

	// Try to start a running pipe.
	_, err = client.StartPipe(ctx, &pipes.StartPipeInput{
		Name: aws.String(pipeName),
	})
	require.Error(t, err)
}

func TestPipes_CreatePipeWithDescription(t *testing.T) {
	client := newPipesClient(t)
	ctx := t.Context()

	pipeName := "test-pipe-description"
	source := "arn:aws:sqs:us-east-1:123456789012:test-source-queue"
	target := "arn:aws:lambda:us-east-1:123456789012:function:test-target"
	roleArn := "arn:aws:iam::123456789012:role/test-pipe-role"
	description := "Test pipe for EventBridge Pipes integration test"

	// Create pipe with description.
	_, err := client.CreatePipe(ctx, &pipes.CreatePipeInput{
		Name:        aws.String(pipeName),
		Source:      aws.String(source),
		Target:      aws.String(target),
		RoleArn:     aws.String(roleArn),
		Description: aws.String(description),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeletePipe(ctx, &pipes.DeletePipeInput{
			Name: aws.String(pipeName),
		})
	})

	// Verify description.
	descOutput, err := client.DescribePipe(ctx, &pipes.DescribePipeInput{
		Name: aws.String(pipeName),
	})
	require.NoError(t, err)
	require.Equal(t, description, *descOutput.Description)
}

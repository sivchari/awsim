//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func newEC2Client(t *testing.T) *ec2.Client {
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

	return ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestEC2_RunAndDescribeInstances(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()

	// Run instances
	runResult, err := client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-12345678"),
		InstanceType: types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(2),
	})
	if err != nil {
		t.Fatalf("failed to run instances: %v", err)
	}

	if len(runResult.Instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(runResult.Instances))
	}

	instanceIDs := make([]string, 0, len(runResult.Instances))
	for _, inst := range runResult.Instances {
		instanceIDs = append(instanceIDs, *inst.InstanceId)
	}

	t.Cleanup(func() {
		_, _ = client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
			InstanceIds: instanceIDs,
		})
	})

	// Describe instances
	descResult, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	})
	if err != nil {
		t.Fatalf("failed to describe instances: %v", err)
	}

	totalInstances := 0
	for _, reservation := range descResult.Reservations {
		totalInstances += len(reservation.Instances)
	}

	if totalInstances != 2 {
		t.Errorf("expected 2 instances in describe result, got %d", totalInstances)
	}
}

func TestEC2_StartAndStopInstances(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()

	// Run instance
	runResult, err := client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-12345678"),
		InstanceType: types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to run instance: %v", err)
	}

	instanceID := *runResult.Instances[0].InstanceId

	t.Cleanup(func() {
		_, _ = client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
			InstanceIds: []string{instanceID},
		})
	})

	// Stop instance
	stopResult, err := client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		t.Fatalf("failed to stop instance: %v", err)
	}

	if len(stopResult.StoppingInstances) != 1 {
		t.Errorf("expected 1 stopping instance, got %d", len(stopResult.StoppingInstances))
	}

	if stopResult.StoppingInstances[0].CurrentState.Name != types.InstanceStateNameStopped {
		t.Errorf("expected stopped state, got %s", stopResult.StoppingInstances[0].CurrentState.Name)
	}

	// Start instance
	startResult, err := client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		t.Fatalf("failed to start instance: %v", err)
	}

	if len(startResult.StartingInstances) != 1 {
		t.Errorf("expected 1 starting instance, got %d", len(startResult.StartingInstances))
	}

	if startResult.StartingInstances[0].CurrentState.Name != types.InstanceStateNameRunning {
		t.Errorf("expected running state, got %s", startResult.StartingInstances[0].CurrentState.Name)
	}
}

func TestEC2_TerminateInstances(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()

	// Run instance
	runResult, err := client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-12345678"),
		InstanceType: types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to run instance: %v", err)
	}

	instanceID := *runResult.Instances[0].InstanceId

	// Terminate instance
	termResult, err := client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		t.Fatalf("failed to terminate instance: %v", err)
	}

	if len(termResult.TerminatingInstances) != 1 {
		t.Errorf("expected 1 terminating instance, got %d", len(termResult.TerminatingInstances))
	}

	if termResult.TerminatingInstances[0].CurrentState.Name != types.InstanceStateNameTerminated {
		t.Errorf("expected terminated state, got %s", termResult.TerminatingInstances[0].CurrentState.Name)
	}
}

func TestEC2_CreateAndDeleteSecurityGroup(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()
	groupName := "test-security-group"

	// Create security group
	createResult, err := client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(groupName),
		Description: aws.String("Test security group"),
	})
	if err != nil {
		t.Fatalf("failed to create security group: %v", err)
	}

	if createResult.GroupId == nil {
		t.Error("expected group ID to be set")
	}

	groupID := *createResult.GroupId

	// Delete security group
	_, err = client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(groupID),
	})
	if err != nil {
		t.Fatalf("failed to delete security group: %v", err)
	}
}

func TestEC2_AuthorizeSecurityGroupIngress(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()
	groupName := "test-ingress-group"

	// Create security group
	createResult, err := client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(groupName),
		Description: aws.String("Test ingress security group"),
	})
	if err != nil {
		t.Fatalf("failed to create security group: %v", err)
	}

	groupID := *createResult.GroupId

	t.Cleanup(func() {
		_, _ = client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(groupID),
		})
	})

	// Authorize ingress
	_, err = client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(groupID),
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(22),
				ToPort:     aws.Int32(22),
				IpRanges: []types.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("SSH access"),
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to authorize security group ingress: %v", err)
	}
}

func TestEC2_CreateAndDeleteKeyPair(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()
	keyName := "test-key-pair"

	// Create key pair
	createResult, err := client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
	})
	if err != nil {
		t.Fatalf("failed to create key pair: %v", err)
	}

	if createResult.KeyPairId == nil {
		t.Error("expected key pair ID to be set")
	}

	if createResult.KeyMaterial == nil || *createResult.KeyMaterial == "" {
		t.Error("expected key material to be set")
	}

	t.Cleanup(func() {
		_, _ = client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
			KeyName: aws.String(keyName),
		})
	})

	// Delete key pair
	_, err = client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: aws.String(keyName),
	})
	if err != nil {
		t.Fatalf("failed to delete key pair: %v", err)
	}
}

func TestEC2_DescribeKeyPairs(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()
	keyName := "test-describe-key-pair"

	// Create key pair
	_, err := client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
	})
	if err != nil {
		t.Fatalf("failed to create key pair: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
			KeyName: aws.String(keyName),
		})
	})

	// Describe key pairs
	descResult, err := client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
		KeyNames: []string{keyName},
	})
	if err != nil {
		t.Fatalf("failed to describe key pairs: %v", err)
	}

	if len(descResult.KeyPairs) != 1 {
		t.Errorf("expected 1 key pair, got %d", len(descResult.KeyPairs))
	}

	if *descResult.KeyPairs[0].KeyName != keyName {
		t.Errorf("expected key name %s, got %s", keyName, *descResult.KeyPairs[0].KeyName)
	}
}

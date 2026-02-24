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

func TestEC2_CreateAndDeleteVpc(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()

	// Create VPC
	createResult, err := client.CreateVpc(ctx, &ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
	})
	if err != nil {
		t.Fatalf("failed to create VPC: %v", err)
	}

	if createResult.Vpc == nil {
		t.Fatal("expected VPC to be set")
	}

	if createResult.Vpc.VpcId == nil {
		t.Error("expected VPC ID to be set")
	}

	vpcID := *createResult.Vpc.VpcId

	// Delete VPC
	_, err = client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(vpcID),
	})
	if err != nil {
		t.Fatalf("failed to delete VPC: %v", err)
	}
}

func TestEC2_DescribeVpcs(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()

	// Create VPC
	createResult, err := client.CreateVpc(ctx, &ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
	})
	if err != nil {
		t.Fatalf("failed to create VPC: %v", err)
	}

	vpcID := *createResult.Vpc.VpcId

	t.Cleanup(func() {
		_, _ = client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
			VpcId: aws.String(vpcID),
		})
	})

	// Describe VPCs
	descResult, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcID},
	})
	if err != nil {
		t.Fatalf("failed to describe VPCs: %v", err)
	}

	if len(descResult.Vpcs) != 1 {
		t.Errorf("expected 1 VPC, got %d", len(descResult.Vpcs))
	}

	if *descResult.Vpcs[0].VpcId != vpcID {
		t.Errorf("expected VPC ID %s, got %s", vpcID, *descResult.Vpcs[0].VpcId)
	}
}

func TestEC2_CreateAndDeleteSubnet(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()

	// Create VPC first
	vpcResult, err := client.CreateVpc(ctx, &ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
	})
	if err != nil {
		t.Fatalf("failed to create VPC: %v", err)
	}

	vpcID := *vpcResult.Vpc.VpcId

	t.Cleanup(func() {
		_, _ = client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
			VpcId: aws.String(vpcID),
		})
	})

	// Create Subnet
	subnetResult, err := client.CreateSubnet(ctx, &ec2.CreateSubnetInput{
		VpcId:     aws.String(vpcID),
		CidrBlock: aws.String("10.0.1.0/24"),
	})
	if err != nil {
		t.Fatalf("failed to create subnet: %v", err)
	}

	if subnetResult.Subnet == nil {
		t.Fatal("expected subnet to be set")
	}

	subnetID := *subnetResult.Subnet.SubnetId

	// Delete Subnet
	_, err = client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(subnetID),
	})
	if err != nil {
		t.Fatalf("failed to delete subnet: %v", err)
	}
}

func TestEC2_CreateInternetGatewayAndAttach(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()

	// Create VPC first
	vpcResult, err := client.CreateVpc(ctx, &ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
	})
	if err != nil {
		t.Fatalf("failed to create VPC: %v", err)
	}

	vpcID := *vpcResult.Vpc.VpcId

	// Create Internet Gateway
	igwResult, err := client.CreateInternetGateway(ctx, &ec2.CreateInternetGatewayInput{})
	if err != nil {
		t.Fatalf("failed to create internet gateway: %v", err)
	}

	if igwResult.InternetGateway == nil {
		t.Fatal("expected internet gateway to be set")
	}

	igwID := *igwResult.InternetGateway.InternetGatewayId

	t.Cleanup(func() {
		_, _ = client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(igwID),
			VpcId:             aws.String(vpcID),
		})
		_, _ = client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: aws.String(igwID),
		})
		_, _ = client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
			VpcId: aws.String(vpcID),
		})
	})

	// Attach Internet Gateway to VPC
	_, err = client.AttachInternetGateway(ctx, &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(igwID),
		VpcId:             aws.String(vpcID),
	})
	if err != nil {
		t.Fatalf("failed to attach internet gateway: %v", err)
	}

	// Describe Internet Gateways
	descResult, err := client.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []string{igwID},
	})
	if err != nil {
		t.Fatalf("failed to describe internet gateways: %v", err)
	}

	if len(descResult.InternetGateways) != 1 {
		t.Errorf("expected 1 internet gateway, got %d", len(descResult.InternetGateways))
	}

	if len(descResult.InternetGateways[0].Attachments) != 1 {
		t.Errorf("expected 1 attachment, got %d", len(descResult.InternetGateways[0].Attachments))
	}

	if *descResult.InternetGateways[0].Attachments[0].VpcId != vpcID {
		t.Errorf("expected attachment VPC ID %s, got %s", vpcID, *descResult.InternetGateways[0].Attachments[0].VpcId)
	}
}

func TestEC2_CreateRouteTableAndAssociate(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()

	// Create VPC first
	vpcResult, err := client.CreateVpc(ctx, &ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
	})
	if err != nil {
		t.Fatalf("failed to create VPC: %v", err)
	}

	vpcID := *vpcResult.Vpc.VpcId

	// Create Subnet
	subnetResult, err := client.CreateSubnet(ctx, &ec2.CreateSubnetInput{
		VpcId:     aws.String(vpcID),
		CidrBlock: aws.String("10.0.1.0/24"),
	})
	if err != nil {
		t.Fatalf("failed to create subnet: %v", err)
	}

	subnetID := *subnetResult.Subnet.SubnetId

	t.Cleanup(func() {
		_, _ = client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
			SubnetId: aws.String(subnetID),
		})
		_, _ = client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
			VpcId: aws.String(vpcID),
		})
	})

	// Create Route Table
	rtResult, err := client.CreateRouteTable(ctx, &ec2.CreateRouteTableInput{
		VpcId: aws.String(vpcID),
	})
	if err != nil {
		t.Fatalf("failed to create route table: %v", err)
	}

	if rtResult.RouteTable == nil {
		t.Fatal("expected route table to be set")
	}

	rtID := *rtResult.RouteTable.RouteTableId

	// Associate Route Table with Subnet
	assocResult, err := client.AssociateRouteTable(ctx, &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(rtID),
		SubnetId:     aws.String(subnetID),
	})
	if err != nil {
		t.Fatalf("failed to associate route table: %v", err)
	}

	if assocResult.AssociationId == nil {
		t.Error("expected association ID to be set")
	}
}

func TestEC2_CreateRoute(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()

	// Create VPC first
	vpcResult, err := client.CreateVpc(ctx, &ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
	})
	if err != nil {
		t.Fatalf("failed to create VPC: %v", err)
	}

	vpcID := *vpcResult.Vpc.VpcId

	// Create Internet Gateway
	igwResult, err := client.CreateInternetGateway(ctx, &ec2.CreateInternetGatewayInput{})
	if err != nil {
		t.Fatalf("failed to create internet gateway: %v", err)
	}

	igwID := *igwResult.InternetGateway.InternetGatewayId

	// Attach Internet Gateway to VPC
	_, err = client.AttachInternetGateway(ctx, &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(igwID),
		VpcId:             aws.String(vpcID),
	})
	if err != nil {
		t.Fatalf("failed to attach internet gateway: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(igwID),
			VpcId:             aws.String(vpcID),
		})
		_, _ = client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: aws.String(igwID),
		})
		_, _ = client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
			VpcId: aws.String(vpcID),
		})
	})

	// Create Route Table
	rtResult, err := client.CreateRouteTable(ctx, &ec2.CreateRouteTableInput{
		VpcId: aws.String(vpcID),
	})
	if err != nil {
		t.Fatalf("failed to create route table: %v", err)
	}

	rtID := *rtResult.RouteTable.RouteTableId

	// Create Route
	_, err = client.CreateRoute(ctx, &ec2.CreateRouteInput{
		RouteTableId:         aws.String(rtID),
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(igwID),
	})
	if err != nil {
		t.Fatalf("failed to create route: %v", err)
	}

	// Describe Route Tables
	descResult, err := client.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		RouteTableIds: []string{rtID},
	})
	if err != nil {
		t.Fatalf("failed to describe route tables: %v", err)
	}

	if len(descResult.RouteTables) != 1 {
		t.Errorf("expected 1 route table, got %d", len(descResult.RouteTables))
	}

	// Check for the new route (should have local + our new route)
	if len(descResult.RouteTables[0].Routes) < 2 {
		t.Errorf("expected at least 2 routes, got %d", len(descResult.RouteTables[0].Routes))
	}
}

func TestEC2_CreateNatGateway(t *testing.T) {
	client := newEC2Client(t)
	ctx := t.Context()

	// Create VPC first
	vpcResult, err := client.CreateVpc(ctx, &ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
	})
	if err != nil {
		t.Fatalf("failed to create VPC: %v", err)
	}

	vpcID := *vpcResult.Vpc.VpcId

	// Create Subnet
	subnetResult, err := client.CreateSubnet(ctx, &ec2.CreateSubnetInput{
		VpcId:     aws.String(vpcID),
		CidrBlock: aws.String("10.0.1.0/24"),
	})
	if err != nil {
		t.Fatalf("failed to create subnet: %v", err)
	}

	subnetID := *subnetResult.Subnet.SubnetId

	t.Cleanup(func() {
		_, _ = client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
			SubnetId: aws.String(subnetID),
		})
		_, _ = client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
			VpcId: aws.String(vpcID),
		})
	})

	// Create NAT Gateway (private connectivity type - no EIP required)
	natgwResult, err := client.CreateNatGateway(ctx, &ec2.CreateNatGatewayInput{
		SubnetId:         aws.String(subnetID),
		ConnectivityType: types.ConnectivityTypePrivate,
	})
	if err != nil {
		t.Fatalf("failed to create NAT gateway: %v", err)
	}

	if natgwResult.NatGateway == nil {
		t.Fatal("expected NAT gateway to be set")
	}

	natgwID := *natgwResult.NatGateway.NatGatewayId

	// Describe NAT Gateways
	descResult, err := client.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []string{natgwID},
	})
	if err != nil {
		t.Fatalf("failed to describe NAT gateways: %v", err)
	}

	if len(descResult.NatGateways) != 1 {
		t.Errorf("expected 1 NAT gateway, got %d", len(descResult.NatGateways))
	}

	if *descResult.NatGateways[0].NatGatewayId != natgwID {
		t.Errorf("expected NAT gateway ID %s, got %s", natgwID, *descResult.NatGateways[0].NatGatewayId)
	}
}

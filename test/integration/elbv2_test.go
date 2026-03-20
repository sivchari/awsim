//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/sivchari/golden"
)

func newELBv2Client(t *testing.T) *elasticloadbalancingv2.Client {
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

	return elasticloadbalancingv2.NewFromConfig(cfg, func(o *elasticloadbalancingv2.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestELBv2_CreateAndDeleteLoadBalancer(t *testing.T) {
	client := newELBv2Client(t)
	ctx := t.Context()
	lbName := "test-load-balancer"

	// Create load balancer
	createResult, err := client.CreateLoadBalancer(ctx, &elasticloadbalancingv2.CreateLoadBalancerInput{
		Name:    aws.String(lbName),
		Subnets: []string{"subnet-12345678", "subnet-87654321"},
		Type:    types.LoadBalancerTypeEnumApplication,
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("LoadBalancerArn", "DNSName", "CanonicalHostedZoneId", "CreatedTime", "VpcId", "ResultMetadata")).Assert(t.Name()+"_create", createResult)

	lb := createResult.LoadBalancers[0]

	t.Cleanup(func() {
		_, _ = client.DeleteLoadBalancer(context.Background(), &elasticloadbalancingv2.DeleteLoadBalancerInput{
			LoadBalancerArn: lb.LoadBalancerArn,
		})
	})

	// Delete load balancer
	_, err = client.DeleteLoadBalancer(context.Background(), &elasticloadbalancingv2.DeleteLoadBalancerInput{
		LoadBalancerArn: lb.LoadBalancerArn,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestELBv2_DescribeLoadBalancers(t *testing.T) {
	client := newELBv2Client(t)
	ctx := t.Context()
	lbName := "test-describe-lb"

	// Create load balancer
	createResult, err := client.CreateLoadBalancer(ctx, &elasticloadbalancingv2.CreateLoadBalancerInput{
		Name:    aws.String(lbName),
		Subnets: []string{"subnet-12345678"},
	})
	if err != nil {
		t.Fatal(err)
	}

	lbArn := createResult.LoadBalancers[0].LoadBalancerArn

	t.Cleanup(func() {
		_, _ = client.DeleteLoadBalancer(context.Background(), &elasticloadbalancingv2.DeleteLoadBalancerInput{
			LoadBalancerArn: lbArn,
		})
	})

	// Describe load balancers by ARN
	descResult, err := client.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{
		LoadBalancerArns: []string{*lbArn},
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("LoadBalancerArn", "DNSName", "CanonicalHostedZoneId", "CreatedTime", "VpcId", "ResultMetadata")).Assert(t.Name()+"_describe", descResult)
}

func TestELBv2_CreateAndDeleteTargetGroup(t *testing.T) {
	client := newELBv2Client(t)
	ctx := t.Context()
	tgName := "test-target-group"

	// Create target group
	createResult, err := client.CreateTargetGroup(ctx, &elasticloadbalancingv2.CreateTargetGroupInput{
		Name:       aws.String(tgName),
		Protocol:   types.ProtocolEnumHttp,
		Port:       aws.Int32(80),
		VpcId:      aws.String("vpc-12345678"),
		TargetType: types.TargetTypeEnumInstance,
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("TargetGroupArn", "LoadBalancerArns", "ResultMetadata")).Assert(t.Name()+"_create", createResult)

	tg := createResult.TargetGroups[0]

	t.Cleanup(func() {
		_, _ = client.DeleteTargetGroup(context.Background(), &elasticloadbalancingv2.DeleteTargetGroupInput{
			TargetGroupArn: tg.TargetGroupArn,
		})
	})

	// Delete target group
	_, err = client.DeleteTargetGroup(context.Background(), &elasticloadbalancingv2.DeleteTargetGroupInput{
		TargetGroupArn: tg.TargetGroupArn,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestELBv2_DescribeTargetGroups(t *testing.T) {
	client := newELBv2Client(t)
	ctx := t.Context()
	tgName := "test-describe-tg"

	// Create target group
	createResult, err := client.CreateTargetGroup(ctx, &elasticloadbalancingv2.CreateTargetGroupInput{
		Name:       aws.String(tgName),
		Protocol:   types.ProtocolEnumHttp,
		Port:       aws.Int32(80),
		VpcId:      aws.String("vpc-12345678"),
		TargetType: types.TargetTypeEnumInstance,
	})
	if err != nil {
		t.Fatal(err)
	}

	tgArn := createResult.TargetGroups[0].TargetGroupArn

	t.Cleanup(func() {
		_, _ = client.DeleteTargetGroup(context.Background(), &elasticloadbalancingv2.DeleteTargetGroupInput{
			TargetGroupArn: tgArn,
		})
	})

	// Describe target groups
	descResult, err := client.DescribeTargetGroups(ctx, &elasticloadbalancingv2.DescribeTargetGroupsInput{
		TargetGroupArns: []string{*tgArn},
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("TargetGroupArn", "LoadBalancerArns", "ResultMetadata")).Assert(t.Name()+"_describe", descResult)
}

func TestELBv2_RegisterAndDeregisterTargets(t *testing.T) {
	client := newELBv2Client(t)
	ctx := t.Context()
	tgName := "test-register-targets"

	// Create target group
	createResult, err := client.CreateTargetGroup(ctx, &elasticloadbalancingv2.CreateTargetGroupInput{
		Name:       aws.String(tgName),
		Protocol:   types.ProtocolEnumHttp,
		Port:       aws.Int32(80),
		VpcId:      aws.String("vpc-12345678"),
		TargetType: types.TargetTypeEnumInstance,
	})
	if err != nil {
		t.Fatal(err)
	}

	tgArn := createResult.TargetGroups[0].TargetGroupArn

	t.Cleanup(func() {
		_, _ = client.DeleteTargetGroup(context.Background(), &elasticloadbalancingv2.DeleteTargetGroupInput{
			TargetGroupArn: tgArn,
		})
	})

	// Register targets
	_, err = client.RegisterTargets(ctx, &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: tgArn,
		Targets: []types.TargetDescription{
			{Id: aws.String("i-12345678"), Port: aws.Int32(80)},
			{Id: aws.String("i-87654321"), Port: aws.Int32(80)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Deregister targets
	_, err = client.DeregisterTargets(ctx, &elasticloadbalancingv2.DeregisterTargetsInput{
		TargetGroupArn: tgArn,
		Targets: []types.TargetDescription{
			{Id: aws.String("i-12345678")},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestELBv2_CreateAndDeleteListener(t *testing.T) {
	client := newELBv2Client(t)
	ctx := t.Context()
	lbName := "test-listener-lb"
	tgName := "test-listener-tg"

	// Create load balancer
	lbResult, err := client.CreateLoadBalancer(ctx, &elasticloadbalancingv2.CreateLoadBalancerInput{
		Name:    aws.String(lbName),
		Subnets: []string{"subnet-12345678"},
	})
	if err != nil {
		t.Fatal(err)
	}

	lbArn := lbResult.LoadBalancers[0].LoadBalancerArn

	// Create target group
	tgResult, err := client.CreateTargetGroup(ctx, &elasticloadbalancingv2.CreateTargetGroupInput{
		Name:       aws.String(tgName),
		Protocol:   types.ProtocolEnumHttp,
		Port:       aws.Int32(80),
		VpcId:      aws.String("vpc-12345678"),
		TargetType: types.TargetTypeEnumInstance,
	})
	if err != nil {
		t.Fatal(err)
	}

	tgArn := tgResult.TargetGroups[0].TargetGroupArn

	t.Cleanup(func() {
		_, _ = client.DeleteTargetGroup(context.Background(), &elasticloadbalancingv2.DeleteTargetGroupInput{
			TargetGroupArn: tgArn,
		})
		_, _ = client.DeleteLoadBalancer(context.Background(), &elasticloadbalancingv2.DeleteLoadBalancerInput{
			LoadBalancerArn: lbArn,
		})
	})

	// Create listener
	listenerResult, err := client.CreateListener(ctx, &elasticloadbalancingv2.CreateListenerInput{
		LoadBalancerArn: lbArn,
		Port:            aws.Int32(80),
		Protocol:        types.ProtocolEnumHttp,
		DefaultActions: []types.Action{
			{
				Type:           types.ActionTypeEnumForward,
				TargetGroupArn: tgArn,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("ListenerArn", "LoadBalancerArn", "TargetGroupArn", "ResultMetadata")).Assert(t.Name()+"_create", listenerResult)

	listenerArn := listenerResult.Listeners[0].ListenerArn

	// Delete listener
	_, err = client.DeleteListener(context.Background(), &elasticloadbalancingv2.DeleteListenerInput{
		ListenerArn: listenerArn,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestELBv2_LoadBalancerWithTargetGroupAndListener(t *testing.T) {
	client := newELBv2Client(t)
	ctx := t.Context()

	// Create load balancer
	lbResult, err := client.CreateLoadBalancer(ctx, &elasticloadbalancingv2.CreateLoadBalancerInput{
		Name:    aws.String("test-full-lb"),
		Subnets: []string{"subnet-12345678", "subnet-87654321"},
		Type:    types.LoadBalancerTypeEnumApplication,
	})
	if err != nil {
		t.Fatal(err)
	}

	lbArn := lbResult.LoadBalancers[0].LoadBalancerArn

	// Create target group
	tgResult, err := client.CreateTargetGroup(ctx, &elasticloadbalancingv2.CreateTargetGroupInput{
		Name:       aws.String("test-full-tg"),
		Protocol:   types.ProtocolEnumHttp,
		Port:       aws.Int32(80),
		VpcId:      aws.String("vpc-12345678"),
		TargetType: types.TargetTypeEnumInstance,
	})
	if err != nil {
		t.Fatal(err)
	}

	tgArn := tgResult.TargetGroups[0].TargetGroupArn

	// Register targets
	_, err = client.RegisterTargets(ctx, &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: tgArn,
		Targets: []types.TargetDescription{
			{Id: aws.String("i-12345678"), Port: aws.Int32(80)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create listener
	listenerResult, err := client.CreateListener(ctx, &elasticloadbalancingv2.CreateListenerInput{
		LoadBalancerArn: lbArn,
		Port:            aws.Int32(80),
		Protocol:        types.ProtocolEnumHttp,
		DefaultActions: []types.Action{
			{
				Type:           types.ActionTypeEnumForward,
				TargetGroupArn: tgArn,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	listenerArn := listenerResult.Listeners[0].ListenerArn

	// Cleanup in reverse order
	t.Cleanup(func() {
		_, _ = client.DeleteListener(context.Background(), &elasticloadbalancingv2.DeleteListenerInput{
			ListenerArn: listenerArn,
		})
		_, _ = client.DeleteTargetGroup(context.Background(), &elasticloadbalancingv2.DeleteTargetGroupInput{
			TargetGroupArn: tgArn,
		})
		_, _ = client.DeleteLoadBalancer(context.Background(), &elasticloadbalancingv2.DeleteLoadBalancerInput{
			LoadBalancerArn: lbArn,
		})
	})

	// Verify everything is created
	descLbResult, err := client.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{
		LoadBalancerArns: []string{*lbArn},
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("LoadBalancerArn", "DNSName", "CanonicalHostedZoneId", "CreatedTime", "VpcId", "SubnetId", "ResultMetadata")).Assert(t.Name()+"_describe_lb", descLbResult)

	descTgResult, err := client.DescribeTargetGroups(ctx, &elasticloadbalancingv2.DescribeTargetGroupsInput{
		TargetGroupArns: []string{*tgArn},
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("TargetGroupArn", "LoadBalancerArns", "ResultMetadata")).Assert(t.Name()+"_describe_tg", descTgResult)
}

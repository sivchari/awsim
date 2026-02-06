//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func newECSClient(t *testing.T) *ecs.Client {
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

	return ecs.NewFromConfig(cfg, func(o *ecs.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestECS_CreateAndDeleteCluster(t *testing.T) {
	client := newECSClient(t)
	ctx := t.Context()
	clusterName := "test-cluster-create-delete"

	// Create cluster.
	createOutput, err := client.CreateCluster(ctx, &ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	if createOutput.Cluster == nil {
		t.Fatal("cluster is nil")
	}

	if *createOutput.Cluster.ClusterName != clusterName {
		t.Errorf("cluster name mismatch: got %s, want %s", *createOutput.Cluster.ClusterName, clusterName)
	}

	if createOutput.Cluster.Status == nil || *createOutput.Cluster.Status != "ACTIVE" {
		t.Errorf("cluster status mismatch: got %v, want ACTIVE", createOutput.Cluster.Status)
	}

	t.Logf("Created cluster: %s", *createOutput.Cluster.ClusterName)

	// Delete cluster.
	deleteOutput, err := client.DeleteCluster(ctx, &ecs.DeleteClusterInput{
		Cluster: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to delete cluster: %v", err)
	}

	if deleteOutput.Cluster == nil {
		t.Fatal("deleted cluster is nil")
	}

	if *deleteOutput.Cluster.Status != "INACTIVE" {
		t.Errorf("deleted cluster status mismatch: got %s, want INACTIVE", *deleteOutput.Cluster.Status)
	}
}

func TestECS_ListClusters(t *testing.T) {
	client := newECSClient(t)
	ctx := t.Context()
	clusterName := "test-cluster-list"

	// Create cluster.
	_, err := client.CreateCluster(ctx, &ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(ctx, &ecs.DeleteClusterInput{
			Cluster: aws.String(clusterName),
		})
	})

	// List clusters.
	listOutput, err := client.ListClusters(ctx, &ecs.ListClustersInput{})
	if err != nil {
		t.Fatalf("failed to list clusters: %v", err)
	}

	found := false

	for _, arn := range listOutput.ClusterArns {
		if arn == "arn:aws:ecs:us-east-1:000000000000:cluster/"+clusterName {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("cluster %s not found in list", clusterName)
	}
}

func TestECS_DescribeClusters(t *testing.T) {
	client := newECSClient(t)
	ctx := t.Context()
	clusterName := "test-cluster-describe"

	// Create cluster.
	_, err := client.CreateCluster(ctx, &ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(ctx, &ecs.DeleteClusterInput{
			Cluster: aws.String(clusterName),
		})
	})

	// Describe clusters.
	descOutput, err := client.DescribeClusters(ctx, &ecs.DescribeClustersInput{
		Clusters: []string{clusterName},
	})
	if err != nil {
		t.Fatalf("failed to describe clusters: %v", err)
	}

	if len(descOutput.Clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(descOutput.Clusters))
	}

	if *descOutput.Clusters[0].ClusterName != clusterName {
		t.Errorf("cluster name mismatch: got %s, want %s", *descOutput.Clusters[0].ClusterName, clusterName)
	}

	if *descOutput.Clusters[0].Status != "ACTIVE" {
		t.Errorf("cluster status mismatch: got %s, want ACTIVE", *descOutput.Clusters[0].Status)
	}
}

func TestECS_RegisterAndDeregisterTaskDefinition(t *testing.T) {
	client := newECSClient(t)
	ctx := t.Context()
	family := "test-task-definition"

	// Register task definition.
	registerOutput, err := client.RegisterTaskDefinition(ctx, &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(family),
		ContainerDefinitions: []types.ContainerDefinition{
			{
				Name:      aws.String("test-container"),
				Image:     aws.String("nginx:latest"),
				Essential: aws.Bool(true),
				Memory:    aws.Int32(512),
				PortMappings: []types.PortMapping{
					{
						ContainerPort: aws.Int32(80),
						HostPort:      aws.Int32(8080),
						Protocol:      types.TransportProtocolTcp,
					},
				},
			},
		},
		Cpu:         aws.String("256"),
		Memory:      aws.String("512"),
		NetworkMode: types.NetworkModeAwsvpc,
		RequiresCompatibilities: []types.Compatibility{
			types.CompatibilityFargate,
		},
	})
	if err != nil {
		t.Fatalf("failed to register task definition: %v", err)
	}

	if registerOutput.TaskDefinition == nil {
		t.Fatal("task definition is nil")
	}

	if *registerOutput.TaskDefinition.Family != family {
		t.Errorf("family mismatch: got %s, want %s", *registerOutput.TaskDefinition.Family, family)
	}

	if registerOutput.TaskDefinition.Revision != 1 {
		t.Errorf("revision mismatch: got %d, want 1", registerOutput.TaskDefinition.Revision)
	}

	taskDefArn := *registerOutput.TaskDefinition.TaskDefinitionArn
	t.Logf("Registered task definition: %s", taskDefArn)

	// Deregister task definition.
	deregisterOutput, err := client.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefArn),
	})
	if err != nil {
		t.Fatalf("failed to deregister task definition: %v", err)
	}

	if deregisterOutput.TaskDefinition == nil {
		t.Fatal("deregistered task definition is nil")
	}

	if *deregisterOutput.TaskDefinition.Status != "INACTIVE" {
		t.Errorf("task definition status mismatch: got %s, want INACTIVE", *deregisterOutput.TaskDefinition.Status)
	}
}

func TestECS_RunAndStopTask(t *testing.T) {
	client := newECSClient(t)
	ctx := t.Context()
	clusterName := "test-cluster-run-task"
	family := "test-task-run"

	// Create cluster.
	_, err := client.CreateCluster(ctx, &ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(ctx, &ecs.DeleteClusterInput{
			Cluster: aws.String(clusterName),
		})
	})

	// Register task definition.
	registerOutput, err := client.RegisterTaskDefinition(ctx, &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(family),
		ContainerDefinitions: []types.ContainerDefinition{
			{
				Name:      aws.String("test-container"),
				Image:     aws.String("nginx:latest"),
				Essential: aws.Bool(true),
				Memory:    aws.Int32(512),
			},
		},
		Cpu:    aws.String("256"),
		Memory: aws.String("512"),
	})
	if err != nil {
		t.Fatalf("failed to register task definition: %v", err)
	}

	taskDefArn := *registerOutput.TaskDefinition.TaskDefinitionArn

	t.Cleanup(func() {
		_, _ = client.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
			TaskDefinition: aws.String(taskDefArn),
		})
	})

	// Run task.
	runOutput, err := client.RunTask(ctx, &ecs.RunTaskInput{
		Cluster:        aws.String(clusterName),
		TaskDefinition: aws.String(taskDefArn),
		Count:          aws.Int32(1),
		LaunchType:     types.LaunchTypeFargate,
	})
	if err != nil {
		t.Fatalf("failed to run task: %v", err)
	}

	if len(runOutput.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(runOutput.Tasks))
	}

	taskArn := *runOutput.Tasks[0].TaskArn
	t.Logf("Started task: %s", taskArn)

	if *runOutput.Tasks[0].LastStatus != "RUNNING" {
		t.Errorf("task status mismatch: got %s, want RUNNING", *runOutput.Tasks[0].LastStatus)
	}

	// Stop task.
	stopOutput, err := client.StopTask(ctx, &ecs.StopTaskInput{
		Cluster: aws.String(clusterName),
		Task:    aws.String(taskArn),
		Reason:  aws.String("Test stop"),
	})
	if err != nil {
		t.Fatalf("failed to stop task: %v", err)
	}

	if stopOutput.Task == nil {
		t.Fatal("stopped task is nil")
	}

	if *stopOutput.Task.LastStatus != "STOPPED" {
		t.Errorf("stopped task status mismatch: got %s, want STOPPED", *stopOutput.Task.LastStatus)
	}
}

func TestECS_DescribeTasks(t *testing.T) {
	client := newECSClient(t)
	ctx := t.Context()
	clusterName := "test-cluster-describe-tasks"
	family := "test-task-describe"

	// Create cluster.
	_, err := client.CreateCluster(ctx, &ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(ctx, &ecs.DeleteClusterInput{
			Cluster: aws.String(clusterName),
		})
	})

	// Register task definition.
	registerOutput, err := client.RegisterTaskDefinition(ctx, &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(family),
		ContainerDefinitions: []types.ContainerDefinition{
			{
				Name:      aws.String("test-container"),
				Image:     aws.String("nginx:latest"),
				Essential: aws.Bool(true),
				Memory:    aws.Int32(512),
			},
		},
		Cpu:    aws.String("256"),
		Memory: aws.String("512"),
	})
	if err != nil {
		t.Fatalf("failed to register task definition: %v", err)
	}

	taskDefArn := *registerOutput.TaskDefinition.TaskDefinitionArn

	t.Cleanup(func() {
		_, _ = client.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
			TaskDefinition: aws.String(taskDefArn),
		})
	})

	// Run task.
	runOutput, err := client.RunTask(ctx, &ecs.RunTaskInput{
		Cluster:        aws.String(clusterName),
		TaskDefinition: aws.String(taskDefArn),
		Count:          aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to run task: %v", err)
	}

	taskArn := *runOutput.Tasks[0].TaskArn

	t.Cleanup(func() {
		_, _ = client.StopTask(ctx, &ecs.StopTaskInput{
			Cluster: aws.String(clusterName),
			Task:    aws.String(taskArn),
		})
	})

	// Describe tasks.
	descOutput, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(clusterName),
		Tasks:   []string{taskArn},
	})
	if err != nil {
		t.Fatalf("failed to describe tasks: %v", err)
	}

	if len(descOutput.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(descOutput.Tasks))
	}

	if *descOutput.Tasks[0].TaskArn != taskArn {
		t.Errorf("task ARN mismatch: got %s, want %s", *descOutput.Tasks[0].TaskArn, taskArn)
	}
}

func TestECS_CreateAndDeleteService(t *testing.T) {
	client := newECSClient(t)
	ctx := t.Context()
	clusterName := "test-cluster-service"
	serviceName := "test-service"
	family := "test-task-service"

	// Create cluster.
	_, err := client.CreateCluster(ctx, &ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(ctx, &ecs.DeleteClusterInput{
			Cluster: aws.String(clusterName),
		})
	})

	// Register task definition.
	registerOutput, err := client.RegisterTaskDefinition(ctx, &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(family),
		ContainerDefinitions: []types.ContainerDefinition{
			{
				Name:      aws.String("test-container"),
				Image:     aws.String("nginx:latest"),
				Essential: aws.Bool(true),
				Memory:    aws.Int32(512),
			},
		},
		Cpu:    aws.String("256"),
		Memory: aws.String("512"),
	})
	if err != nil {
		t.Fatalf("failed to register task definition: %v", err)
	}

	taskDefArn := *registerOutput.TaskDefinition.TaskDefinitionArn

	t.Cleanup(func() {
		_, _ = client.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
			TaskDefinition: aws.String(taskDefArn),
		})
	})

	// Create service.
	createOutput, err := client.CreateService(ctx, &ecs.CreateServiceInput{
		Cluster:        aws.String(clusterName),
		ServiceName:    aws.String(serviceName),
		TaskDefinition: aws.String(taskDefArn),
		DesiredCount:   aws.Int32(2),
		LaunchType:     types.LaunchTypeFargate,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	if createOutput.Service == nil {
		t.Fatal("service is nil")
	}

	if *createOutput.Service.ServiceName != serviceName {
		t.Errorf("service name mismatch: got %s, want %s", *createOutput.Service.ServiceName, serviceName)
	}

	if createOutput.Service.DesiredCount != 2 {
		t.Errorf("desired count mismatch: got %d, want 2", createOutput.Service.DesiredCount)
	}

	if *createOutput.Service.Status != "ACTIVE" {
		t.Errorf("service status mismatch: got %s, want ACTIVE", *createOutput.Service.Status)
	}

	t.Logf("Created service: %s", *createOutput.Service.ServiceName)

	// Delete service.
	deleteOutput, err := client.DeleteService(ctx, &ecs.DeleteServiceInput{
		Cluster: aws.String(clusterName),
		Service: aws.String(serviceName),
		Force:   aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to delete service: %v", err)
	}

	if deleteOutput.Service == nil {
		t.Fatal("deleted service is nil")
	}

	if *deleteOutput.Service.Status != "DRAINING" {
		t.Errorf("deleted service status mismatch: got %s, want DRAINING", *deleteOutput.Service.Status)
	}
}

func TestECS_UpdateService(t *testing.T) {
	client := newECSClient(t)
	ctx := t.Context()
	clusterName := "test-cluster-update-service"
	serviceName := "test-service-update"
	family := "test-task-update-service"

	// Create cluster.
	_, err := client.CreateCluster(ctx, &ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteCluster(ctx, &ecs.DeleteClusterInput{
			Cluster: aws.String(clusterName),
		})
	})

	// Register task definition.
	registerOutput, err := client.RegisterTaskDefinition(ctx, &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(family),
		ContainerDefinitions: []types.ContainerDefinition{
			{
				Name:      aws.String("test-container"),
				Image:     aws.String("nginx:latest"),
				Essential: aws.Bool(true),
				Memory:    aws.Int32(512),
			},
		},
		Cpu:    aws.String("256"),
		Memory: aws.String("512"),
	})
	if err != nil {
		t.Fatalf("failed to register task definition: %v", err)
	}

	taskDefArn := *registerOutput.TaskDefinition.TaskDefinitionArn

	t.Cleanup(func() {
		_, _ = client.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
			TaskDefinition: aws.String(taskDefArn),
		})
	})

	// Create service.
	_, err = client.CreateService(ctx, &ecs.CreateServiceInput{
		Cluster:        aws.String(clusterName),
		ServiceName:    aws.String(serviceName),
		TaskDefinition: aws.String(taskDefArn),
		DesiredCount:   aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteService(ctx, &ecs.DeleteServiceInput{
			Cluster: aws.String(clusterName),
			Service: aws.String(serviceName),
			Force:   aws.Bool(true),
		})
	})

	// Update service.
	updateOutput, err := client.UpdateService(ctx, &ecs.UpdateServiceInput{
		Cluster:      aws.String(clusterName),
		Service:      aws.String(serviceName),
		DesiredCount: aws.Int32(3),
	})
	if err != nil {
		t.Fatalf("failed to update service: %v", err)
	}

	if updateOutput.Service == nil {
		t.Fatal("updated service is nil")
	}

	if updateOutput.Service.DesiredCount != 3 {
		t.Errorf("desired count mismatch: got %d, want 3", updateOutput.Service.DesiredCount)
	}

	t.Logf("Updated service: %s, desiredCount: %d", *updateOutput.Service.ServiceName, updateOutput.Service.DesiredCount)
}

func TestECS_TaskDefinitionRevision(t *testing.T) {
	client := newECSClient(t)
	ctx := t.Context()
	family := "test-task-revision"

	var taskDefArns []string

	// Register first task definition.
	registerOutput1, err := client.RegisterTaskDefinition(ctx, &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(family),
		ContainerDefinitions: []types.ContainerDefinition{
			{
				Name:      aws.String("container-v1"),
				Image:     aws.String("nginx:1.0"),
				Essential: aws.Bool(true),
				Memory:    aws.Int32(256),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to register first task definition: %v", err)
	}

	taskDefArns = append(taskDefArns, *registerOutput1.TaskDefinition.TaskDefinitionArn)

	if registerOutput1.TaskDefinition.Revision != 1 {
		t.Errorf("first revision mismatch: got %d, want 1", registerOutput1.TaskDefinition.Revision)
	}

	// Register second task definition with same family.
	registerOutput2, err := client.RegisterTaskDefinition(ctx, &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(family),
		ContainerDefinitions: []types.ContainerDefinition{
			{
				Name:      aws.String("container-v2"),
				Image:     aws.String("nginx:2.0"),
				Essential: aws.Bool(true),
				Memory:    aws.Int32(512),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to register second task definition: %v", err)
	}

	taskDefArns = append(taskDefArns, *registerOutput2.TaskDefinition.TaskDefinitionArn)

	if registerOutput2.TaskDefinition.Revision != 2 {
		t.Errorf("second revision mismatch: got %d, want 2", registerOutput2.TaskDefinition.Revision)
	}

	// Cleanup.
	for _, arn := range taskDefArns {
		_, _ = client.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
			TaskDefinition: aws.String(arn),
		})
	}
}

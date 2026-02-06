package ecs

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultAccountID = "000000000000"
	defaultRegion    = "us-east-1"
)

// Storage defines the interface for ECS storage operations.
type Storage interface {
	CreateCluster(ctx context.Context, req *CreateClusterRequest) (*Cluster, error)
	DeleteCluster(ctx context.Context, cluster string) (*Cluster, error)
	DescribeClusters(ctx context.Context, clusters []string) ([]Cluster, []Failure, error)
	ListClusters(ctx context.Context, maxResults int, nextToken string) ([]string, string, error)

	RegisterTaskDefinition(ctx context.Context, req *RegisterTaskDefinitionRequest) (*TaskDefinition, error)
	DeregisterTaskDefinition(ctx context.Context, taskDefinition string) (*TaskDefinition, error)

	RunTask(ctx context.Context, req *RunTaskRequest) ([]Task, []Failure, error)
	StopTask(ctx context.Context, cluster, task, reason string) (*Task, error)
	DescribeTasks(ctx context.Context, cluster string, tasks []string) ([]Task, []Failure, error)

	CreateService(ctx context.Context, req *CreateServiceRequest) (*ECSService, error)
	DeleteService(ctx context.Context, cluster, service string, force bool) (*ECSService, error)
	UpdateService(ctx context.Context, req *UpdateServiceRequest) (*ECSService, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu              sync.RWMutex
	clusters        map[string]*Cluster
	taskDefinitions map[string]*TaskDefinition
	taskDefFamilies map[string][]string // family -> list of task definition ARNs
	tasks           map[string]*Task
	services        map[string]*ECSService
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		clusters:        make(map[string]*Cluster),
		taskDefinitions: make(map[string]*TaskDefinition),
		taskDefFamilies: make(map[string][]string),
		tasks:           make(map[string]*Task),
		services:        make(map[string]*ECSService),
	}
}

func generateID() string {
	return uuid.New().String()[:8]
}

func clusterArn(name string) string {
	return fmt.Sprintf("arn:aws:ecs:%s:%s:cluster/%s", defaultRegion, defaultAccountID, name)
}

func taskDefinitionArn(family string, revision int) string {
	return fmt.Sprintf("arn:aws:ecs:%s:%s:task-definition/%s:%d", defaultRegion, defaultAccountID, family, revision)
}

func taskArn(clusterName, taskID string) string {
	return fmt.Sprintf("arn:aws:ecs:%s:%s:task/%s/%s", defaultRegion, defaultAccountID, clusterName, taskID)
}

func serviceArn(clusterName, serviceName string) string {
	return fmt.Sprintf("arn:aws:ecs:%s:%s:service/%s/%s", defaultRegion, defaultAccountID, clusterName, serviceName)
}

// CreateCluster creates a new ECS cluster.
func (m *MemoryStorage) CreateCluster(_ context.Context, req *CreateClusterRequest) (*Cluster, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := req.ClusterName
	if name == "" {
		name = "default"
	}

	arn := clusterArn(name)

	// Check if cluster already exists.
	if existing, ok := m.clusters[arn]; ok {
		return existing, nil
	}

	cluster := &Cluster{
		ClusterArn:  arn,
		ClusterName: name,
		Status:      "ACTIVE",
		Tags:        req.Tags,
	}

	m.clusters[arn] = cluster

	return cluster, nil
}

// DeleteCluster deletes an ECS cluster.
func (m *MemoryStorage) DeleteCluster(_ context.Context, cluster string) (*Cluster, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	arn := m.resolveClusterArn(cluster)

	existing, ok := m.clusters[arn]
	if !ok {
		return nil, &Error{
			Code:    "ClusterNotFoundException",
			Message: "The specified cluster was not found",
		}
	}

	// Check if cluster has active services or tasks.
	if existing.ActiveServicesCount > 0 {
		return nil, &Error{
			Code:    "ClusterContainsServicesException",
			Message: "The cluster contains active services",
		}
	}

	if existing.RunningTasksCount > 0 {
		return nil, &Error{
			Code:    "ClusterContainsTasksException",
			Message: "The cluster contains running tasks",
		}
	}

	existing.Status = "INACTIVE"
	delete(m.clusters, arn)

	return existing, nil
}

// DescribeClusters describes ECS clusters.
func (m *MemoryStorage) DescribeClusters(_ context.Context, clusters []string) ([]Cluster, []Failure, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Cluster
	var failures []Failure

	// If no clusters specified, return all.
	if len(clusters) == 0 {
		for _, c := range m.clusters {
			result = append(result, *c)
		}

		return result, failures, nil
	}

	for _, cluster := range clusters {
		arn := m.resolveClusterArn(cluster)

		if c, ok := m.clusters[arn]; ok {
			result = append(result, *c)
		} else {
			failures = append(failures, Failure{
				Arn:    arn,
				Reason: "MISSING",
			})
		}
	}

	return result, failures, nil
}

// ListClusters lists ECS cluster ARNs.
func (m *MemoryStorage) ListClusters(_ context.Context, _ int, _ string) ([]string, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var arns []string
	for arn := range m.clusters {
		arns = append(arns, arn)
	}

	return arns, "", nil
}

// RegisterTaskDefinition registers a new task definition.
func (m *MemoryStorage) RegisterTaskDefinition(_ context.Context, req *RegisterTaskDefinitionRequest) (*TaskDefinition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Determine revision number.
	revision := 1
	if existing, ok := m.taskDefFamilies[req.Family]; ok {
		revision = len(existing) + 1
	}

	arn := taskDefinitionArn(req.Family, revision)

	td := &TaskDefinition{
		TaskDefinitionArn:       arn,
		Family:                  req.Family,
		Revision:                revision,
		Status:                  "ACTIVE",
		ContainerDefinitions:    req.ContainerDefinitions,
		CPU:                     req.CPU,
		Memory:                  req.Memory,
		NetworkMode:             req.NetworkMode,
		RequiresCompatibilities: req.RequiresCompatibilities,
		ExecutionRoleArn:        req.ExecutionRoleArn,
		TaskRoleArn:             req.TaskRoleArn,
		Tags:                    req.Tags,
	}

	m.taskDefinitions[arn] = td
	m.taskDefFamilies[req.Family] = append(m.taskDefFamilies[req.Family], arn)

	return td, nil
}

// DeregisterTaskDefinition deregisters a task definition.
func (m *MemoryStorage) DeregisterTaskDefinition(_ context.Context, taskDefinition string) (*TaskDefinition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	arn := m.resolveTaskDefinitionArn(taskDefinition)

	td, ok := m.taskDefinitions[arn]
	if !ok {
		return nil, &Error{
			Code:    "ClientException",
			Message: "The specified task definition was not found",
		}
	}

	td.Status = "INACTIVE"

	return td, nil
}

// RunTask runs a task.
func (m *MemoryStorage) RunTask(_ context.Context, req *RunTaskRequest) ([]Task, []Failure, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clusterArn := m.resolveClusterArn(req.Cluster)

	cluster, ok := m.clusters[clusterArn]
	if !ok {
		// Create default cluster if not exists.
		if req.Cluster == "" || req.Cluster == "default" {
			cluster = &Cluster{
				ClusterArn:  clusterArn,
				ClusterName: "default",
				Status:      "ACTIVE",
			}
			m.clusters[clusterArn] = cluster
		} else {
			return nil, nil, &Error{
				Code:    "ClusterNotFoundException",
				Message: "The specified cluster was not found",
			}
		}
	}

	tdArn := m.resolveTaskDefinitionArn(req.TaskDefinition)

	td, ok := m.taskDefinitions[tdArn]
	if !ok {
		return nil, nil, &Error{
			Code:    "ClientException",
			Message: "The specified task definition was not found",
		}
	}

	count := req.Count
	if count == 0 {
		count = 1
	}

	launchType := req.LaunchType
	if launchType == "" {
		launchType = "EC2"
	}

	var tasks []Task

	for i := 0; i < count; i++ {
		taskID := generateID()
		now := time.Now()

		containers := make([]Container, 0, len(td.ContainerDefinitions))
		for _, cd := range td.ContainerDefinitions {
			containers = append(containers, Container{
				ContainerArn: fmt.Sprintf("arn:aws:ecs:%s:%s:container/%s", defaultRegion, defaultAccountID, generateID()),
				Name:         cd.Name,
				Image:        cd.Image,
				LastStatus:   "RUNNING",
			})
		}

		clusterName := extractClusterName(clusterArn)
		task := Task{
			TaskArn:           taskArn(clusterName, taskID),
			ClusterArn:        clusterArn,
			TaskDefinitionArn: tdArn,
			LastStatus:        "RUNNING",
			DesiredStatus:     "RUNNING",
			CPU:               td.CPU,
			Memory:            td.Memory,
			Containers:        containers,
			StartedAt:         &now,
			Group:             req.Group,
			LaunchType:        launchType,
			Tags:              req.Tags,
		}

		m.tasks[task.TaskArn] = &task
		tasks = append(tasks, task)
	}

	cluster.RunningTasksCount += count

	return tasks, nil, nil
}

// StopTask stops a running task.
func (m *MemoryStorage) StopTask(_ context.Context, cluster, taskID, reason string) (*Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clusterArn := m.resolveClusterArn(cluster)

	task, ok := m.tasks[taskID]
	if !ok {
		// Try to find by short ID.
		for arn, t := range m.tasks {
			if strings.HasSuffix(arn, "/"+taskID) && t.ClusterArn == clusterArn {
				task = t

				break
			}
		}
	}

	if task == nil {
		return nil, &Error{
			Code:    "InvalidParameterException",
			Message: "The specified task was not found",
		}
	}

	now := time.Now()
	task.LastStatus = "STOPPED"
	task.DesiredStatus = "STOPPED"
	task.StoppedAt = &now
	task.StoppedReason = reason

	for i := range task.Containers {
		task.Containers[i].LastStatus = "STOPPED"
	}

	// Update cluster task count.
	if c, ok := m.clusters[task.ClusterArn]; ok && c.RunningTasksCount > 0 {
		c.RunningTasksCount--
	}

	return task, nil
}

// DescribeTasks describes tasks.
func (m *MemoryStorage) DescribeTasks(_ context.Context, cluster string, taskIDs []string) ([]Task, []Failure, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clusterArn := m.resolveClusterArn(cluster)

	var tasks []Task
	var failures []Failure

	for _, taskID := range taskIDs {
		found := false

		for arn, task := range m.tasks {
			if task.ClusterArn != clusterArn {
				continue
			}

			if arn == taskID || strings.HasSuffix(arn, "/"+taskID) {
				tasks = append(tasks, *task)
				found = true

				break
			}
		}

		if !found {
			failures = append(failures, Failure{
				Arn:    taskID,
				Reason: "MISSING",
			})
		}
	}

	return tasks, failures, nil
}

// CreateService creates an ECS service.
func (m *MemoryStorage) CreateService(_ context.Context, req *CreateServiceRequest) (*ECSService, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clusterArn := m.resolveClusterArn(req.Cluster)

	cluster, ok := m.clusters[clusterArn]
	if !ok {
		return nil, &Error{
			Code:    "ClusterNotFoundException",
			Message: "The specified cluster was not found",
		}
	}

	clusterName := extractClusterName(clusterArn)
	arn := serviceArn(clusterName, req.ServiceName)

	// Check if service already exists.
	if _, ok := m.services[arn]; ok {
		return nil, &Error{
			Code:    "ServiceAlreadyExistsException",
			Message: "A service with that name already exists",
		}
	}

	launchType := req.LaunchType
	if launchType == "" {
		launchType = "EC2"
	}

	now := time.Now()
	svc := &ECSService{
		ServiceArn:     arn,
		ServiceName:    req.ServiceName,
		ClusterArn:     clusterArn,
		TaskDefinition: m.resolveTaskDefinitionArn(req.TaskDefinition),
		DesiredCount:   req.DesiredCount,
		RunningCount:   0,
		PendingCount:   req.DesiredCount,
		LaunchType:     launchType,
		Status:         "ACTIVE",
		Deployments: []Deployment{
			{
				ID:             generateID(),
				Status:         "PRIMARY",
				TaskDefinition: m.resolveTaskDefinitionArn(req.TaskDefinition),
				DesiredCount:   req.DesiredCount,
				RunningCount:   0,
				PendingCount:   req.DesiredCount,
				CreatedAt:      &now,
				UpdatedAt:      &now,
			},
		},
		Tags: req.Tags,
	}

	m.services[arn] = svc
	cluster.ActiveServicesCount++

	return svc, nil
}

// DeleteService deletes an ECS service.
func (m *MemoryStorage) DeleteService(_ context.Context, cluster, service string, force bool) (*ECSService, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clusterArn := m.resolveClusterArn(cluster)
	clusterName := extractClusterName(clusterArn)
	svcArn := serviceArn(clusterName, service)

	// Try to find by ARN or name.
	svc, ok := m.services[svcArn]
	if !ok {
		// Try to find by full ARN.
		svc, ok = m.services[service]
		if !ok {
			return nil, &Error{
				Code:    "ServiceNotFoundException",
				Message: "The specified service was not found",
			}
		}

		svcArn = service
	}

	if svc.RunningCount > 0 && !force {
		return nil, &Error{
			Code:    "ServiceNotDrainedException",
			Message: "The service still has running tasks",
		}
	}

	svc.Status = "INACTIVE"
	delete(m.services, svcArn)

	// Update cluster service count.
	if c, ok := m.clusters[svc.ClusterArn]; ok && c.ActiveServicesCount > 0 {
		c.ActiveServicesCount--
	}

	return svc, nil
}

// UpdateService updates an ECS service.
func (m *MemoryStorage) UpdateService(_ context.Context, req *UpdateServiceRequest) (*ECSService, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clusterArn := m.resolveClusterArn(req.Cluster)
	clusterName := extractClusterName(clusterArn)
	svcArn := serviceArn(clusterName, req.Service)

	// Try to find by ARN or name.
	svc, ok := m.services[svcArn]
	if !ok {
		svc, ok = m.services[req.Service]
		if !ok {
			return nil, &Error{
				Code:    "ServiceNotFoundException",
				Message: "The specified service was not found",
			}
		}
	}

	if req.TaskDefinition != "" {
		svc.TaskDefinition = m.resolveTaskDefinitionArn(req.TaskDefinition)
	}

	if req.DesiredCount != nil {
		svc.DesiredCount = *req.DesiredCount
		svc.PendingCount = max(*req.DesiredCount-svc.RunningCount, 0)
	}

	// Update deployment.
	now := time.Now()
	if len(svc.Deployments) > 0 {
		svc.Deployments[0].TaskDefinition = svc.TaskDefinition
		svc.Deployments[0].DesiredCount = svc.DesiredCount
		svc.Deployments[0].UpdatedAt = &now
	}

	return svc, nil
}

// Helper methods.

func (m *MemoryStorage) resolveClusterArn(cluster string) string {
	if cluster == "" {
		return clusterArn("default")
	}

	if strings.HasPrefix(cluster, "arn:") {
		return cluster
	}

	return clusterArn(cluster)
}

func (m *MemoryStorage) resolveTaskDefinitionArn(taskDefinition string) string {
	if strings.HasPrefix(taskDefinition, "arn:") {
		return taskDefinition
	}

	// Try family:revision format.
	parts := strings.Split(taskDefinition, ":")
	if len(parts) == 2 {
		return fmt.Sprintf("arn:aws:ecs:%s:%s:task-definition/%s", defaultRegion, defaultAccountID, taskDefinition)
	}

	// Try to find latest revision.
	if arns, ok := m.taskDefFamilies[taskDefinition]; ok && len(arns) > 0 {
		return arns[len(arns)-1]
	}

	return fmt.Sprintf("arn:aws:ecs:%s:%s:task-definition/%s:1", defaultRegion, defaultAccountID, taskDefinition)
}

func extractClusterName(arn string) string {
	parts := strings.Split(arn, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}

	return arn
}

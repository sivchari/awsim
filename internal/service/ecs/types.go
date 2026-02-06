package ecs

import "time"

// Cluster represents an ECS cluster.
type Cluster struct {
	ClusterArn                        string
	ClusterName                       string
	Status                            string
	RegisteredContainerInstancesCount int
	RunningTasksCount                 int
	PendingTasksCount                 int
	ActiveServicesCount               int
	Tags                              []Tag
}

// TaskDefinition represents an ECS task definition.
type TaskDefinition struct {
	TaskDefinitionArn       string
	Family                  string
	Revision                int
	Status                  string
	ContainerDefinitions    []ContainerDefinition
	CPU                     string
	Memory                  string
	NetworkMode             string
	RequiresCompatibilities []string
	ExecutionRoleArn        string
	TaskRoleArn             string
	Tags                    []Tag
}

// ContainerDefinition represents a container in a task definition.
type ContainerDefinition struct {
	Name         string
	Image        string
	CPU          int
	Memory       int
	Essential    bool
	PortMappings []PortMapping
	Environment  []KeyValuePair
	Command      []string
	EntryPoint   []string
}

// PortMapping represents a port mapping.
type PortMapping struct {
	ContainerPort int
	HostPort      int
	Protocol      string
}

// KeyValuePair represents a key-value pair.
type KeyValuePair struct {
	Name  string
	Value string
}

// Task represents a running task.
type Task struct {
	TaskArn              string
	ClusterArn           string
	TaskDefinitionArn    string
	ContainerInstanceArn string
	LastStatus           string
	DesiredStatus        string
	CPU                  string
	Memory               string
	Containers           []Container
	StartedAt            *time.Time
	StoppedAt            *time.Time
	StoppedReason        string
	Group                string
	LaunchType           string
	Tags                 []Tag
}

// Container represents a container in a task.
type Container struct {
	ContainerArn    string
	Name            string
	Image           string
	LastStatus      string
	ExitCode        *int
	Reason          string
	NetworkBindings []NetworkBinding
}

// NetworkBinding represents a network binding.
type NetworkBinding struct {
	BindIP        string
	ContainerPort int
	HostPort      int
	Protocol      string
}

// ECSService represents an ECS service.
type ECSService struct {
	ServiceArn     string
	ServiceName    string
	ClusterArn     string
	TaskDefinition string
	DesiredCount   int
	RunningCount   int
	PendingCount   int
	LaunchType     string
	Status         string
	Deployments    []Deployment
	Tags           []Tag
}

// Deployment represents a service deployment.
type Deployment struct {
	ID             string
	Status         string
	TaskDefinition string
	DesiredCount   int
	RunningCount   int
	PendingCount   int
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
}

// Tag represents a resource tag.
type Tag struct {
	Key   string
	Value string
}

// Request types.

// CreateClusterRequest represents a CreateCluster request.
type CreateClusterRequest struct {
	ClusterName string `json:"clusterName"`
	Tags        []Tag  `json:"tags,omitempty"`
}

// DeleteClusterRequest represents a DeleteCluster request.
type DeleteClusterRequest struct {
	Cluster string `json:"cluster"`
}

// DescribeClustersRequest represents a DescribeClusters request.
type DescribeClustersRequest struct {
	Clusters []string `json:"clusters,omitempty"`
}

// ListClustersRequest represents a ListClusters request.
type ListClustersRequest struct {
	MaxResults int    `json:"maxResults,omitempty"`
	NextToken  string `json:"nextToken,omitempty"`
}

// RegisterTaskDefinitionRequest represents a RegisterTaskDefinition request.
type RegisterTaskDefinitionRequest struct {
	Family                  string                `json:"family"`
	ContainerDefinitions    []ContainerDefinition `json:"containerDefinitions"`
	CPU                     string                `json:"cpu,omitempty"`
	Memory                  string                `json:"memory,omitempty"`
	NetworkMode             string                `json:"networkMode,omitempty"`
	RequiresCompatibilities []string              `json:"requiresCompatibilities,omitempty"`
	ExecutionRoleArn        string                `json:"executionRoleArn,omitempty"`
	TaskRoleArn             string                `json:"taskRoleArn,omitempty"`
	Tags                    []Tag                 `json:"tags,omitempty"`
}

// DeregisterTaskDefinitionRequest represents a DeregisterTaskDefinition request.
type DeregisterTaskDefinitionRequest struct {
	TaskDefinition string `json:"taskDefinition"`
}

// RunTaskRequest represents a RunTask request.
type RunTaskRequest struct {
	Cluster        string `json:"cluster,omitempty"`
	TaskDefinition string `json:"taskDefinition"`
	Count          int    `json:"count,omitempty"`
	LaunchType     string `json:"launchType,omitempty"`
	Group          string `json:"group,omitempty"`
	Tags           []Tag  `json:"tags,omitempty"`
}

// StopTaskRequest represents a StopTask request.
type StopTaskRequest struct {
	Cluster string `json:"cluster,omitempty"`
	Task    string `json:"task"`
	Reason  string `json:"reason,omitempty"`
}

// DescribeTasksRequest represents a DescribeTasks request.
type DescribeTasksRequest struct {
	Cluster string   `json:"cluster,omitempty"`
	Tasks   []string `json:"tasks"`
}

// CreateServiceRequest represents a CreateService request.
type CreateServiceRequest struct {
	Cluster        string `json:"cluster,omitempty"`
	ServiceName    string `json:"serviceName"`
	TaskDefinition string `json:"taskDefinition"`
	DesiredCount   int    `json:"desiredCount"`
	LaunchType     string `json:"launchType,omitempty"`
	Tags           []Tag  `json:"tags,omitempty"`
}

// DeleteServiceRequest represents a DeleteService request.
type DeleteServiceRequest struct {
	Cluster string `json:"cluster,omitempty"`
	Service string `json:"service"`
	Force   bool   `json:"force,omitempty"`
}

// UpdateServiceRequest represents an UpdateService request.
type UpdateServiceRequest struct {
	Cluster        string `json:"cluster,omitempty"`
	Service        string `json:"service"`
	TaskDefinition string `json:"taskDefinition,omitempty"`
	DesiredCount   *int   `json:"desiredCount,omitempty"`
}

// Response types.

// CreateClusterResponse represents a CreateCluster response.
type CreateClusterResponse struct {
	Cluster *Cluster `json:"cluster"`
}

// DeleteClusterResponse represents a DeleteCluster response.
type DeleteClusterResponse struct {
	Cluster *Cluster `json:"cluster"`
}

// DescribeClustersResponse represents a DescribeClusters response.
type DescribeClustersResponse struct {
	Clusters []Cluster `json:"clusters"`
	Failures []Failure `json:"failures,omitempty"`
}

// ListClustersResponse represents a ListClusters response.
type ListClustersResponse struct {
	ClusterArns []string `json:"clusterArns"`
	NextToken   string   `json:"nextToken,omitempty"`
}

// RegisterTaskDefinitionResponse represents a RegisterTaskDefinition response.
type RegisterTaskDefinitionResponse struct {
	TaskDefinition *TaskDefinition `json:"taskDefinition"`
	Tags           []Tag           `json:"tags,omitempty"`
}

// DeregisterTaskDefinitionResponse represents a DeregisterTaskDefinition response.
type DeregisterTaskDefinitionResponse struct {
	TaskDefinition *TaskDefinition `json:"taskDefinition"`
}

// RunTaskResponse represents a RunTask response.
type RunTaskResponse struct {
	Tasks    []Task    `json:"tasks"`
	Failures []Failure `json:"failures,omitempty"`
}

// StopTaskResponse represents a StopTask response.
type StopTaskResponse struct {
	Task *Task `json:"task"`
}

// DescribeTasksResponse represents a DescribeTasks response.
type DescribeTasksResponse struct {
	Tasks    []Task    `json:"tasks"`
	Failures []Failure `json:"failures,omitempty"`
}

// CreateServiceResponse represents a CreateService response.
type CreateServiceResponse struct {
	Service *ECSService `json:"service"`
}

// DeleteServiceResponse represents a DeleteService response.
type DeleteServiceResponse struct {
	Service *ECSService `json:"service"`
}

// UpdateServiceResponse represents an UpdateService response.
type UpdateServiceResponse struct {
	Service *ECSService `json:"service"`
}

// Failure represents a failure in a batch operation.
type Failure struct {
	Arn    string `json:"arn,omitempty"`
	Reason string `json:"reason"`
}

// Error represents an ECS error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

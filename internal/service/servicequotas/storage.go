package servicequotas

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Error codes.
const (
	errNoSuchResourceException  = "NoSuchResourceException"
	errIllegalArgumentException = "IllegalArgumentException"
	errTooManyRequestsException = "TooManyRequestsException"
)

// Default values.
const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "123456789012"
)

// Quota status values.
const (
	quotaStatusPending = "PENDING"
)

// Quota applied at level values.
const (
	quotaAppliedAtLevelAccount = "ACCOUNT"
)

// quotaDefinition represents a quota definition for initialization.
type quotaDefinition struct {
	code        string
	name        string
	value       float64
	unit        string
	adjustable  bool
	description string
}

// Storage defines the Service Quotas storage interface.
type Storage interface {
	// Service operations
	ListServices(ctx context.Context, maxResults int32, nextToken string) ([]*ServiceInfo, string, error)

	// Quota operations
	GetServiceQuota(ctx context.Context, serviceCode, quotaCode string) (*ServiceQuota, error)
	ListServiceQuotas(ctx context.Context, serviceCode string, maxResults int32, nextToken string) ([]*ServiceQuota, string, error)
	GetAWSDefaultServiceQuota(ctx context.Context, serviceCode, quotaCode string) (*ServiceQuota, error)
	ListAWSDefaultServiceQuotas(ctx context.Context, serviceCode string, maxResults int32, nextToken string) ([]*ServiceQuota, string, error)

	// Quota change request operations
	RequestServiceQuotaIncrease(ctx context.Context, serviceCode, quotaCode string, desiredValue float64) (*QuotaChangeRequest, error)
	GetRequestedServiceQuotaChange(ctx context.Context, requestID string) (*QuotaChangeRequest, error)
	ListRequestedServiceQuotaChangeHistory(ctx context.Context, serviceCode, quotaCode, status string, maxResults int32, nextToken string) ([]*QuotaChangeRequest, string, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu            sync.RWMutex
	services      map[string]*ServiceInfo
	quotas        map[string]map[string]*ServiceQuota // serviceCode -> quotaCode -> quota
	defaultQuotas map[string]map[string]*ServiceQuota // serviceCode -> quotaCode -> quota
	requests      map[string]*QuotaChangeRequest
	region        string
	accountID     string
}

// NewMemoryStorage creates a new MemoryStorage with predefined services and quotas.
func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		services:      make(map[string]*ServiceInfo),
		quotas:        make(map[string]map[string]*ServiceQuota),
		defaultQuotas: make(map[string]map[string]*ServiceQuota),
		requests:      make(map[string]*QuotaChangeRequest),
		region:        defaultRegion,
		accountID:     defaultAccountID,
	}

	storage.initializeDefaultData()

	return storage
}

// initializeDefaultData sets up predefined services and quotas.
func (m *MemoryStorage) initializeDefaultData() {
	m.initializeServices()
	m.initializeEC2Quotas()
	m.initializeS3Quotas()
	m.initializeLambdaQuotas()
	m.initializeDynamoDBQuotas()
	m.initializeSQSQuotas()
}

func (m *MemoryStorage) initializeServices() {
	services := []ServiceInfo{
		{ServiceCode: "ec2", ServiceName: "Amazon Elastic Compute Cloud (Amazon EC2)"},
		{ServiceCode: "s3", ServiceName: "Amazon Simple Storage Service (Amazon S3)"},
		{ServiceCode: "dynamodb", ServiceName: "Amazon DynamoDB"},
		{ServiceCode: "lambda", ServiceName: "AWS Lambda"},
		{ServiceCode: "rds", ServiceName: "Amazon Relational Database Service (Amazon RDS)"},
		{ServiceCode: "sqs", ServiceName: "Amazon Simple Queue Service (Amazon SQS)"},
		{ServiceCode: "sns", ServiceName: "Amazon Simple Notification Service (Amazon SNS)"},
		{ServiceCode: "kinesis", ServiceName: "Amazon Kinesis"},
		{ServiceCode: "elasticache", ServiceName: "Amazon ElastiCache"},
		{ServiceCode: "ecs", ServiceName: "Amazon Elastic Container Service (Amazon ECS)"},
	}

	for i := range services {
		m.services[services[i].ServiceCode] = &services[i]
	}
}

func (m *MemoryStorage) initializeEC2Quotas() {
	m.addServiceQuotas("ec2", "Amazon Elastic Compute Cloud (Amazon EC2)", []quotaDefinition{
		{"L-1216C47A", "Running On-Demand Standard instances", 1920, "None", true, "Max vCPUs for On-Demand Standard instances"},
		{"L-34B43A08", "All Standard Spot Instance Requests", 1920, "None", true, "Max vCPUs for Standard Spot Requests"},
		{"L-0E3CBAB9", "EC2-VPC Elastic IPs", 5, "None", true, "Max Elastic IP addresses for EC2-VPC"},
		{"L-E4BF28E0", "VPCs per Region", 5, "None", true, "Maximum number of VPCs per Region"},
	})
}

func (m *MemoryStorage) initializeS3Quotas() {
	m.addServiceQuotas("s3", "Amazon Simple Storage Service (Amazon S3)", []quotaDefinition{
		{"L-DC2B2D3D", "Buckets", 100, "None", true, "Maximum number of buckets per account"},
	})
}

func (m *MemoryStorage) initializeLambdaQuotas() {
	m.addServiceQuotas("lambda", "AWS Lambda", []quotaDefinition{
		{"L-B99A9384", "Concurrent executions", 1000, "None", true, "Maximum number of concurrent executions"},
		{"L-2ACBD22F", "Function and layer storage", 75, "Gigabytes", true, "Max total storage for functions and layers"},
	})
}

func (m *MemoryStorage) initializeDynamoDBQuotas() {
	m.addServiceQuotas("dynamodb", "Amazon DynamoDB", []quotaDefinition{
		{"L-F98FE922", "Table-level read throughput", 40000, "None", true, "Max read capacity units per table"},
		{"L-82ACEF56", "Table-level write throughput", 40000, "None", true, "Max write capacity units per table"},
	})
}

func (m *MemoryStorage) initializeSQSQuotas() {
	m.addServiceQuotas("sqs", "Amazon Simple Queue Service (Amazon SQS)", []quotaDefinition{
		{"L-06F64E4A", "Messages per queue (backlog)", 120000, "None", false, "Max inflight messages per queue"},
	})
}

// addServiceQuotas adds quotas for a service.
func (m *MemoryStorage) addServiceQuotas(serviceCode, serviceName string, quotas []quotaDefinition) {
	if m.quotas[serviceCode] == nil {
		m.quotas[serviceCode] = make(map[string]*ServiceQuota)
	}

	if m.defaultQuotas[serviceCode] == nil {
		m.defaultQuotas[serviceCode] = make(map[string]*ServiceQuota)
	}

	for _, q := range quotas {
		quota := &ServiceQuota{
			QuotaARN:            generateQuotaARN(m.region, serviceCode, q.code),
			QuotaCode:           q.code,
			QuotaName:           q.name,
			ServiceCode:         serviceCode,
			ServiceName:         serviceName,
			Value:               q.value,
			Unit:                q.unit,
			Adjustable:          q.adjustable,
			GlobalQuota:         false,
			Description:         q.description,
			QuotaAppliedAtLevel: quotaAppliedAtLevelAccount,
		}

		m.quotas[serviceCode][q.code] = quota

		// Default quotas are the same initially
		defaultQuota := *quota
		m.defaultQuotas[serviceCode][q.code] = &defaultQuota
	}
}

// ListServices lists all services.
func (m *MemoryStorage) ListServices(_ context.Context, maxResults int32, _ string) ([]*ServiceInfo, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ServiceInfo, 0, len(m.services))

	for _, svc := range m.services {
		result = append(result, svc)

		//nolint:gosec // len(result) is bounded by the number of services which is limited.
		if maxResults > 0 && int32(len(result)) >= maxResults {
			break
		}
	}

	return result, "", nil
}

// GetServiceQuota gets a service quota.
func (m *MemoryStorage) GetServiceQuota(_ context.Context, serviceCode, quotaCode string) (*ServiceQuota, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	serviceQuotas, exists := m.quotas[serviceCode]
	if !exists {
		return nil, &Error{Code: errNoSuchResourceException, Message: "Service not found: " + serviceCode}
	}

	quota, exists := serviceQuotas[quotaCode]
	if !exists {
		return nil, &Error{Code: errNoSuchResourceException, Message: "Quota not found: " + quotaCode}
	}

	return quota, nil
}

// ListServiceQuotas lists quotas for a service.
func (m *MemoryStorage) ListServiceQuotas(_ context.Context, serviceCode string, maxResults int32, _ string) ([]*ServiceQuota, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	serviceQuotas, exists := m.quotas[serviceCode]
	if !exists {
		return nil, "", &Error{Code: errNoSuchResourceException, Message: "Service not found: " + serviceCode}
	}

	result := make([]*ServiceQuota, 0, len(serviceQuotas))

	for _, quota := range serviceQuotas {
		result = append(result, quota)

		//nolint:gosec // len(result) is bounded by the number of quotas which is limited.
		if maxResults > 0 && int32(len(result)) >= maxResults {
			break
		}
	}

	return result, "", nil
}

// GetAWSDefaultServiceQuota gets a default service quota.
func (m *MemoryStorage) GetAWSDefaultServiceQuota(_ context.Context, serviceCode, quotaCode string) (*ServiceQuota, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	serviceQuotas, exists := m.defaultQuotas[serviceCode]
	if !exists {
		return nil, &Error{Code: errNoSuchResourceException, Message: "Service not found: " + serviceCode}
	}

	quota, exists := serviceQuotas[quotaCode]
	if !exists {
		return nil, &Error{Code: errNoSuchResourceException, Message: "Quota not found: " + quotaCode}
	}

	return quota, nil
}

// ListAWSDefaultServiceQuotas lists default quotas for a service.
func (m *MemoryStorage) ListAWSDefaultServiceQuotas(_ context.Context, serviceCode string, maxResults int32, _ string) ([]*ServiceQuota, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	serviceQuotas, exists := m.defaultQuotas[serviceCode]
	if !exists {
		return nil, "", &Error{Code: errNoSuchResourceException, Message: "Service not found: " + serviceCode}
	}

	result := make([]*ServiceQuota, 0, len(serviceQuotas))

	for _, quota := range serviceQuotas {
		result = append(result, quota)

		//nolint:gosec // len(result) is bounded by the number of quotas which is limited.
		if maxResults > 0 && int32(len(result)) >= maxResults {
			break
		}
	}

	return result, "", nil
}

// RequestServiceQuotaIncrease creates a quota increase request.
func (m *MemoryStorage) RequestServiceQuotaIncrease(_ context.Context, serviceCode, quotaCode string, desiredValue float64) (*QuotaChangeRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if the service exists
	serviceQuotas, exists := m.quotas[serviceCode]
	if !exists {
		return nil, &Error{Code: errNoSuchResourceException, Message: "Service not found: " + serviceCode}
	}

	// Check if the quota exists
	quota, exists := serviceQuotas[quotaCode]
	if !exists {
		return nil, &Error{Code: errNoSuchResourceException, Message: "Quota not found: " + quotaCode}
	}

	// Check if the quota is adjustable
	if !quota.Adjustable {
		return nil, &Error{Code: errIllegalArgumentException, Message: "Quota is not adjustable: " + quotaCode}
	}

	requestID := uuid.New().String()
	now := time.Now()

	request := &QuotaChangeRequest{
		ID:                    requestID,
		ServiceCode:           serviceCode,
		ServiceName:           quota.ServiceName,
		QuotaCode:             quotaCode,
		QuotaName:             quota.QuotaName,
		DesiredValue:          desiredValue,
		Status:                quotaStatusPending,
		Created:               now,
		LastUpdated:           now,
		QuotaARN:              quota.QuotaARN,
		Unit:                  quota.Unit,
		GlobalQuota:           quota.GlobalQuota,
		QuotaRequestedAtLevel: quotaAppliedAtLevelAccount,
	}

	m.requests[requestID] = request

	return request, nil
}

// GetRequestedServiceQuotaChange gets a quota change request.
func (m *MemoryStorage) GetRequestedServiceQuotaChange(_ context.Context, requestID string) (*QuotaChangeRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	request, exists := m.requests[requestID]
	if !exists {
		return nil, &Error{Code: errNoSuchResourceException, Message: "Request not found: " + requestID}
	}

	return request, nil
}

// ListRequestedServiceQuotaChangeHistory lists quota change requests.
func (m *MemoryStorage) ListRequestedServiceQuotaChangeHistory(_ context.Context, serviceCode, quotaCode, status string, maxResults int32, _ string) ([]*QuotaChangeRequest, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*QuotaChangeRequest, 0, len(m.requests))

	for _, request := range m.requests {
		if serviceCode != "" && request.ServiceCode != serviceCode {
			continue
		}

		if quotaCode != "" && request.QuotaCode != quotaCode {
			continue
		}

		if status != "" && request.Status != status {
			continue
		}

		result = append(result, request)

		//nolint:gosec // len(result) is bounded by the number of requests which is limited.
		if maxResults > 0 && int32(len(result)) >= maxResults {
			break
		}
	}

	return result, "", nil
}

// Helper functions.

func generateQuotaARN(region, serviceCode, quotaCode string) string {
	return fmt.Sprintf("arn:aws:servicequotas:%s::%s/%s", region, serviceCode, quotaCode)
}

package lambda

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// Storage defines the Lambda storage interface.
type Storage interface {
	CreateFunction(ctx context.Context, req *CreateFunctionRequest) (*Function, error)
	GetFunction(ctx context.Context, name string) (*Function, error)
	DeleteFunction(ctx context.Context, name string) error
	ListFunctions(ctx context.Context, marker string, maxItems int) ([]*Function, string, error)
	UpdateFunctionCode(ctx context.Context, name string, req *UpdateFunctionCodeRequest) (*Function, error)
	UpdateFunctionConfiguration(ctx context.Context, name string, req *UpdateFunctionConfigurationRequest) (*Function, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu        sync.RWMutex
	functions map[string]*Function
	baseURL   string
	region    string
	accountID string
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		functions: make(map[string]*Function),
		baseURL:   baseURL,
		region:    "us-east-1",
		accountID: "000000000000",
	}
}

// CreateFunction creates a new Lambda function.
func (s *MemoryStorage) CreateFunction(_ context.Context, req *CreateFunctionRequest) (*Function, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.functions[req.FunctionName]; exists {
		return nil, &FunctionError{
			Type:    ErrResourceConflict,
			Message: fmt.Sprintf("Function already exist: %s", req.FunctionName),
		}
	}

	// Calculate code hash.
	codeHash := sha256.Sum256(req.Code.ZipFile)
	codeSha256 := base64.StdEncoding.EncodeToString(codeHash[:])

	// Set defaults.
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 3
	}

	memorySize := req.MemorySize
	if memorySize == 0 {
		memorySize = 128
	}

	packageType := req.PackageType
	if packageType == "" {
		packageType = "Zip"
	}

	architectures := req.Architectures
	if len(architectures) == 0 {
		architectures = []string{"x86_64"}
	}

	fn := &Function{
		FunctionName:  req.FunctionName,
		FunctionArn:   fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", s.region, s.accountID, req.FunctionName),
		Runtime:       req.Runtime,
		Role:          req.Role,
		Handler:       req.Handler,
		Description:   req.Description,
		Timeout:       timeout,
		MemorySize:    memorySize,
		CodeSize:      int64(len(req.Code.ZipFile)),
		CodeSha256:    codeSha256,
		Version:       "$LATEST",
		LastModified:  time.Now().UTC(),
		State:         "Active",
		PackageType:   packageType,
		Architectures: architectures,
		Environment:   req.Environment,
		Code: &FunctionCode{
			ZipFile:         req.Code.ZipFile,
			S3Bucket:        req.Code.S3Bucket,
			S3Key:           req.Code.S3Key,
			S3ObjectVersion: req.Code.S3ObjectVersion,
			ImageURI:        req.Code.ImageURI,
		},
	}

	s.functions[req.FunctionName] = fn

	return fn, nil
}

// GetFunction retrieves a Lambda function by name.
func (s *MemoryStorage) GetFunction(_ context.Context, name string) (*Function, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fn, exists := s.functions[name]
	if !exists {
		return nil, &FunctionError{
			Type:    ErrResourceNotFound,
			Message: fmt.Sprintf("Function not found: %s", name),
		}
	}

	return fn, nil
}

// DeleteFunction deletes a Lambda function.
func (s *MemoryStorage) DeleteFunction(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.functions[name]; !exists {
		return &FunctionError{
			Type:    ErrResourceNotFound,
			Message: fmt.Sprintf("Function not found: %s", name),
		}
	}

	delete(s.functions, name)

	return nil
}

// ListFunctions lists all Lambda functions.
func (s *MemoryStorage) ListFunctions(_ context.Context, marker string, maxItems int) ([]*Function, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxItems == 0 {
		maxItems = 50
	}

	functions := make([]*Function, 0, len(s.functions))
	for _, fn := range s.functions {
		functions = append(functions, fn)
	}

	// Simple pagination (not production-ready).
	start := 0

	if marker != "" {
		for i, fn := range functions {
			if fn.FunctionName == marker {
				start = i + 1

				break
			}
		}
	}

	end := start + maxItems
	if end > len(functions) {
		end = len(functions)
	}

	result := functions[start:end]
	nextMarker := ""

	if end < len(functions) {
		nextMarker = functions[end-1].FunctionName
	}

	return result, nextMarker, nil
}

// UpdateFunctionCode updates the code of a Lambda function.
func (s *MemoryStorage) UpdateFunctionCode(_ context.Context, name string, req *UpdateFunctionCodeRequest) (*Function, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fn, exists := s.functions[name]
	if !exists {
		return nil, &FunctionError{
			Type:    ErrResourceNotFound,
			Message: fmt.Sprintf("Function not found: %s", name),
		}
	}

	// Update code.
	if len(req.ZipFile) > 0 {
		fn.Code.ZipFile = req.ZipFile
		codeHash := sha256.Sum256(req.ZipFile)
		fn.CodeSha256 = base64.StdEncoding.EncodeToString(codeHash[:])
		fn.CodeSize = int64(len(req.ZipFile))
	}

	if req.S3Bucket != "" {
		fn.Code.S3Bucket = req.S3Bucket
	}

	if req.S3Key != "" {
		fn.Code.S3Key = req.S3Key
	}

	if req.S3ObjectVersion != "" {
		fn.Code.S3ObjectVersion = req.S3ObjectVersion
	}

	if req.ImageURI != "" {
		fn.Code.ImageURI = req.ImageURI
	}

	if len(req.Architectures) > 0 {
		fn.Architectures = req.Architectures
	}

	fn.LastModified = time.Now().UTC()

	return fn, nil
}

// UpdateFunctionConfiguration updates the configuration of a Lambda function.
func (s *MemoryStorage) UpdateFunctionConfiguration(_ context.Context, name string, req *UpdateFunctionConfigurationRequest) (*Function, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fn, exists := s.functions[name]
	if !exists {
		return nil, &FunctionError{
			Type:    ErrResourceNotFound,
			Message: fmt.Sprintf("Function not found: %s", name),
		}
	}

	if req.Description != "" {
		fn.Description = req.Description
	}

	if req.Handler != "" {
		fn.Handler = req.Handler
	}

	if req.MemorySize > 0 {
		fn.MemorySize = req.MemorySize
	}

	if req.Role != "" {
		fn.Role = req.Role
	}

	if req.Runtime != "" {
		fn.Runtime = req.Runtime
	}

	if req.Timeout > 0 {
		fn.Timeout = req.Timeout
	}

	if req.Environment != nil {
		fn.Environment = req.Environment
	}

	fn.LastModified = time.Now().UTC()

	return fn, nil
}

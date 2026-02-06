package ssm

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Storage defines the SSM Parameter Store storage interface.
type Storage interface {
	PutParameter(ctx context.Context, req *PutParameterRequest) (*Parameter, error)
	GetParameter(ctx context.Context, name string) (*Parameter, error)
	GetParameters(ctx context.Context, names []string) ([]*Parameter, []string, error)
	GetParametersByPath(ctx context.Context, path string, recursive bool, maxResults int, nextToken string) ([]*Parameter, string, error)
	DeleteParameter(ctx context.Context, name string) error
	DeleteParameters(ctx context.Context, names []string) ([]string, []string, error)
	DescribeParameters(ctx context.Context, maxResults int, nextToken string) ([]*Parameter, string, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu         sync.RWMutex
	parameters map[string]*Parameter
	region     string
	accountID  string
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		parameters: make(map[string]*Parameter),
		region:     "us-east-1",
		accountID:  "000000000000",
	}
}

// PutParameter creates or updates a parameter.
func (s *MemoryStorage) PutParameter(_ context.Context, req *PutParameterRequest) (*Parameter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.parameters[req.Name]

	// Check if parameter exists and overwrite is not set
	if exists && !req.Overwrite {
		return nil, &ParameterError{
			Type:    ErrParameterAlreadyExists,
			Message: "The parameter already exists. To overwrite this value, set the overwrite option in the request to true.",
		}
	}

	// Set defaults
	paramType := req.Type
	if paramType == "" {
		if exists {
			paramType = existing.Type
		} else {
			paramType = ParameterTypeString
		}
	}

	tier := req.Tier
	if tier == "" {
		tier = ParameterTierStandard
	}

	dataType := req.DataType
	if dataType == "" {
		dataType = "text"
	}

	version := int64(1)
	if exists {
		version = existing.Version + 1
	}

	param := &Parameter{
		Name:             req.Name,
		Type:             paramType,
		Value:            req.Value,
		Version:          version,
		LastModifiedDate: time.Now().UTC(),
		ARN:              fmt.Sprintf("arn:aws:ssm:%s:%s:parameter%s", s.region, s.accountID, req.Name),
		DataType:         dataType,
		Tier:             tier,
		Description:      req.Description,
	}

	s.parameters[req.Name] = param

	return param, nil
}

// GetParameter retrieves a parameter by name.
func (s *MemoryStorage) GetParameter(_ context.Context, name string) (*Parameter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	param, exists := s.parameters[name]
	if !exists {
		return nil, &ParameterError{
			Type:    ErrParameterNotFound,
			Message: fmt.Sprintf("Parameter %s not found.", name),
		}
	}

	return param, nil
}

// GetParameters retrieves multiple parameters by name.
func (s *MemoryStorage) GetParameters(_ context.Context, names []string) ([]*Parameter, []string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var params []*Parameter

	var invalidParams []string

	for _, name := range names {
		param, exists := s.parameters[name]
		if exists {
			params = append(params, param)
		} else {
			invalidParams = append(invalidParams, name)
		}
	}

	return params, invalidParams, nil
}

// GetParametersByPath retrieves parameters by path prefix.
func (s *MemoryStorage) GetParametersByPath(_ context.Context, path string, recursive bool, maxResults int, nextToken string) ([]*Parameter, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults == 0 {
		maxResults = 10
	}

	// Normalize path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// Collect matching parameters
	var matches []*Parameter

	for name, param := range s.parameters {
		if matchesPath(name, path, recursive) {
			matches = append(matches, param)
		}
	}

	// Sort by name for consistent pagination
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})

	// Handle pagination
	start := 0

	if nextToken != "" {
		for i, p := range matches {
			if p.Name == nextToken {
				start = i

				break
			}
		}
	}

	end := start + maxResults
	if end > len(matches) {
		end = len(matches)
	}

	result := matches[start:end]
	newNextToken := ""

	if end < len(matches) {
		newNextToken = matches[end].Name
	}

	return result, newNextToken, nil
}

// matchesPath checks if a parameter name matches the given path.
func matchesPath(name, path string, recursive bool) bool {
	if !strings.HasPrefix(name, path) {
		return false
	}

	if recursive {
		return true
	}

	// For non-recursive, check that there are no more slashes after the path prefix
	remainder := strings.TrimPrefix(name, path)

	return !strings.Contains(remainder, "/")
}

// DeleteParameter deletes a parameter by name.
func (s *MemoryStorage) DeleteParameter(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.parameters[name]; !exists {
		return &ParameterError{
			Type:    ErrParameterNotFound,
			Message: fmt.Sprintf("Parameter %s not found.", name),
		}
	}

	delete(s.parameters, name)

	return nil
}

// DeleteParameters deletes multiple parameters.
func (s *MemoryStorage) DeleteParameters(_ context.Context, names []string) ([]string, []string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var deleted []string

	var invalid []string

	for _, name := range names {
		if _, exists := s.parameters[name]; exists {
			delete(s.parameters, name)

			deleted = append(deleted, name)
		} else {
			invalid = append(invalid, name)
		}
	}

	return deleted, invalid, nil
}

// DescribeParameters lists all parameters with metadata.
func (s *MemoryStorage) DescribeParameters(_ context.Context, maxResults int, nextToken string) ([]*Parameter, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults == 0 {
		maxResults = 50
	}

	// Collect all parameters
	params := make([]*Parameter, 0, len(s.parameters))
	for _, p := range s.parameters {
		params = append(params, p)
	}

	// Sort by name for consistent pagination
	sort.Slice(params, func(i, j int) bool {
		return params[i].Name < params[j].Name
	})

	// Handle pagination
	start := 0

	if nextToken != "" {
		for i, p := range params {
			if p.Name == nextToken {
				start = i

				break
			}
		}
	}

	end := start + maxResults
	if end > len(params) {
		end = len(params)
	}

	result := params[start:end]
	newNextToken := ""

	if end < len(params) {
		newNextToken = params[end].Name
	}

	return result, newNextToken, nil
}

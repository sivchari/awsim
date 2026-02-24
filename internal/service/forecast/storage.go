package forecast

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Storage defines the interface for Forecast storage operations.
type Storage interface {
	// Dataset operations
	CreateDataset(ctx context.Context, req *CreateDatasetInput) (string, error)
	DescribeDataset(ctx context.Context, datasetArn string) (*Dataset, error)
	ListDatasets(ctx context.Context, maxResults *int32, nextToken string) ([]*DatasetSummary, string, error)
	DeleteDataset(ctx context.Context, datasetArn string) error

	// DatasetGroup operations
	CreateDatasetGroup(ctx context.Context, req *CreateDatasetGroupInput) (string, error)
	DescribeDatasetGroup(ctx context.Context, datasetGroupArn string) (*DatasetGroup, error)
	ListDatasetGroups(ctx context.Context, maxResults *int32, nextToken string) ([]*DatasetGroupSummary, string, error)
	DeleteDatasetGroup(ctx context.Context, datasetGroupArn string) error
	UpdateDatasetGroup(ctx context.Context, datasetGroupArn string, datasetArns []string) error

	// Predictor operations
	CreatePredictor(ctx context.Context, req *CreatePredictorInput) (string, error)
	DescribePredictor(ctx context.Context, predictorArn string) (*Predictor, error)
	ListPredictors(ctx context.Context, maxResults *int32, nextToken string, filters []Filter) ([]*PredictorSummary, string, error)
	DeletePredictor(ctx context.Context, predictorArn string) error

	// Forecast operations
	CreateForecast(ctx context.Context, req *CreateForecastInput) (string, error)
	DescribeForecast(ctx context.Context, forecastArn string) (*Forecast, error)
	ListForecasts(ctx context.Context, maxResults *int32, nextToken string, filters []Filter) ([]*ForecastSummary, string, error)
	DeleteForecast(ctx context.Context, forecastArn string) error
}

// MemoryStorage implements Storage using in-memory storage.
type MemoryStorage struct {
	mu            sync.RWMutex
	datasets      map[string]*Dataset
	datasetGroups map[string]*DatasetGroup
	predictors    map[string]*Predictor
	forecasts     map[string]*Forecast
	accountID     string
	region        string
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		datasets:      make(map[string]*Dataset),
		datasetGroups: make(map[string]*DatasetGroup),
		predictors:    make(map[string]*Predictor),
		forecasts:     make(map[string]*Forecast),
		accountID:     "123456789012",
		region:        "us-east-1",
	}
}

// Dataset operations.

// CreateDataset creates a new dataset.
func (m *MemoryStorage) CreateDataset(_ context.Context, req *CreateDatasetInput) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate name.
	for _, ds := range m.datasets {
		if ds.DatasetName == req.DatasetName {
			return "", &Error{
				Code:    errResourceAlreadyExistsException,
				Message: fmt.Sprintf("A dataset with name %s already exists", req.DatasetName),
			}
		}
	}

	// Validate dataset type.
	if !isValidDatasetType(req.DatasetType) {
		return "", &Error{
			Code:    errInvalidInputException,
			Message: fmt.Sprintf("Invalid dataset type: %s", req.DatasetType),
		}
	}

	// Validate domain.
	if !isValidDomain(req.Domain) {
		return "", &Error{
			Code:    errInvalidInputException,
			Message: fmt.Sprintf("Invalid domain: %s", req.Domain),
		}
	}

	datasetArn := fmt.Sprintf("arn:aws:forecast:%s:%s:dataset/%s", m.region, m.accountID, req.DatasetName)
	now := time.Now()

	dataset := &Dataset{
		DatasetArn:           datasetArn,
		DatasetName:          req.DatasetName,
		DatasetType:          req.DatasetType,
		Domain:               req.Domain,
		DataFrequency:        req.DataFrequency,
		Schema:               req.Schema,
		Status:               statusActive,
		CreationTime:         ToAWSTimestamp(now),
		LastModificationTime: ToAWSTimestamp(now),
		EncryptionConfig:     req.EncryptionConfig,
	}

	m.datasets[datasetArn] = dataset

	return datasetArn, nil
}

// DescribeDataset returns a dataset.
func (m *MemoryStorage) DescribeDataset(_ context.Context, datasetArn string) (*Dataset, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dataset, exists := m.datasets[datasetArn]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Dataset %s not found", datasetArn),
		}
	}

	return dataset, nil
}

// ListDatasets returns all datasets.
func (m *MemoryStorage) ListDatasets(_ context.Context, maxResults *int32, _ string) ([]*DatasetSummary, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limit := int32(100)
	if maxResults != nil && *maxResults > 0 && *maxResults < limit {
		limit = *maxResults
	}

	summaries := make([]*DatasetSummary, 0, len(m.datasets))

	for _, ds := range m.datasets {
		if int32(len(summaries)) >= limit { //nolint:gosec // G115: len(summaries) is bounded by limit which is int32
			break
		}

		summaries = append(summaries, &DatasetSummary{
			DatasetArn:           ds.DatasetArn,
			DatasetName:          ds.DatasetName,
			DatasetType:          ds.DatasetType,
			Domain:               ds.Domain,
			CreationTime:         ds.CreationTime,
			LastModificationTime: ds.LastModificationTime,
		})
	}

	return summaries, "", nil
}

// DeleteDataset deletes a dataset.
func (m *MemoryStorage) DeleteDataset(_ context.Context, datasetArn string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.datasets[datasetArn]; !exists {
		return &Error{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Dataset %s not found", datasetArn),
		}
	}

	// Check if dataset is in use by any dataset group.
	for _, dg := range m.datasetGroups {
		if slices.Contains(dg.DatasetArns, datasetArn) {
			return &Error{
				Code:    errResourceInUseException,
				Message: fmt.Sprintf("Dataset %s is in use by dataset group %s", datasetArn, dg.DatasetGroupArn),
			}
		}
	}

	delete(m.datasets, datasetArn)

	return nil
}

// DatasetGroup operations.

// CreateDatasetGroup creates a new dataset group.
func (m *MemoryStorage) CreateDatasetGroup(_ context.Context, req *CreateDatasetGroupInput) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate name.
	for _, dg := range m.datasetGroups {
		if dg.DatasetGroupName == req.DatasetGroupName {
			return "", &Error{
				Code:    errResourceAlreadyExistsException,
				Message: fmt.Sprintf("A dataset group with name %s already exists", req.DatasetGroupName),
			}
		}
	}

	// Validate domain.
	if !isValidDomain(req.Domain) {
		return "", &Error{
			Code:    errInvalidInputException,
			Message: fmt.Sprintf("Invalid domain: %s", req.Domain),
		}
	}

	// Validate dataset ARNs exist and have matching domain.
	for _, dsArn := range req.DatasetArns {
		ds, exists := m.datasets[dsArn]
		if !exists {
			return "", &Error{
				Code:    errResourceNotFoundException,
				Message: fmt.Sprintf("Dataset %s not found", dsArn),
			}
		}

		if ds.Domain != req.Domain {
			return "", &Error{
				Code:    errInvalidInputException,
				Message: fmt.Sprintf("Dataset %s has domain %s, but dataset group has domain %s", dsArn, ds.Domain, req.Domain),
			}
		}
	}

	datasetGroupArn := fmt.Sprintf("arn:aws:forecast:%s:%s:dataset-group/%s", m.region, m.accountID, req.DatasetGroupName)
	now := time.Now()

	datasetGroup := &DatasetGroup{
		DatasetGroupArn:      datasetGroupArn,
		DatasetGroupName:     req.DatasetGroupName,
		Domain:               req.Domain,
		DatasetArns:          req.DatasetArns,
		Status:               statusActive,
		CreationTime:         ToAWSTimestamp(now),
		LastModificationTime: ToAWSTimestamp(now),
	}

	m.datasetGroups[datasetGroupArn] = datasetGroup

	return datasetGroupArn, nil
}

// DescribeDatasetGroup returns a dataset group.
func (m *MemoryStorage) DescribeDatasetGroup(_ context.Context, datasetGroupArn string) (*DatasetGroup, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	datasetGroup, exists := m.datasetGroups[datasetGroupArn]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Dataset group %s not found", datasetGroupArn),
		}
	}

	return datasetGroup, nil
}

// ListDatasetGroups returns all dataset groups.
func (m *MemoryStorage) ListDatasetGroups(_ context.Context, maxResults *int32, _ string) ([]*DatasetGroupSummary, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limit := int32(100)
	if maxResults != nil && *maxResults > 0 && *maxResults < limit {
		limit = *maxResults
	}

	summaries := make([]*DatasetGroupSummary, 0, len(m.datasetGroups))

	for _, dg := range m.datasetGroups {
		if int32(len(summaries)) >= limit { //nolint:gosec // G115: len(summaries) is bounded by limit which is int32
			break
		}

		summaries = append(summaries, &DatasetGroupSummary{
			DatasetGroupArn:      dg.DatasetGroupArn,
			DatasetGroupName:     dg.DatasetGroupName,
			CreationTime:         dg.CreationTime,
			LastModificationTime: dg.LastModificationTime,
		})
	}

	return summaries, "", nil
}

// DeleteDatasetGroup deletes a dataset group.
func (m *MemoryStorage) DeleteDatasetGroup(_ context.Context, datasetGroupArn string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.datasetGroups[datasetGroupArn]; !exists {
		return &Error{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Dataset group %s not found", datasetGroupArn),
		}
	}

	// Check if dataset group is in use by any predictor.
	for _, p := range m.predictors {
		if p.InputDataConfig != nil && p.InputDataConfig.DatasetGroupArn == datasetGroupArn {
			return &Error{
				Code:    errResourceInUseException,
				Message: fmt.Sprintf("Dataset group %s is in use by predictor %s", datasetGroupArn, p.PredictorArn),
			}
		}
	}

	delete(m.datasetGroups, datasetGroupArn)

	return nil
}

// UpdateDatasetGroup updates a dataset group.
func (m *MemoryStorage) UpdateDatasetGroup(_ context.Context, datasetGroupArn string, datasetArns []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dg, exists := m.datasetGroups[datasetGroupArn]
	if !exists {
		return &Error{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Dataset group %s not found", datasetGroupArn),
		}
	}

	// Validate dataset ARNs exist and have matching domain.
	for _, dsArn := range datasetArns {
		ds, dsExists := m.datasets[dsArn]
		if !dsExists {
			return &Error{
				Code:    errResourceNotFoundException,
				Message: fmt.Sprintf("Dataset %s not found", dsArn),
			}
		}

		if ds.Domain != dg.Domain {
			return &Error{
				Code:    errInvalidInputException,
				Message: fmt.Sprintf("Dataset %s has domain %s, but dataset group has domain %s", dsArn, ds.Domain, dg.Domain),
			}
		}
	}

	dg.DatasetArns = datasetArns
	dg.LastModificationTime = ToAWSTimestamp(time.Now())

	return nil
}

// Predictor operations.

// CreatePredictor creates a new predictor.
//
//nolint:funlen // validation and struct initialization require more lines
func (m *MemoryStorage) CreatePredictor(_ context.Context, req *CreatePredictorInput) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate name.
	for _, p := range m.predictors {
		if p.PredictorName == req.PredictorName {
			return "", &Error{
				Code:    errResourceAlreadyExistsException,
				Message: fmt.Sprintf("A predictor with name %s already exists", req.PredictorName),
			}
		}
	}

	// Validate dataset group exists.
	if req.InputDataConfig == nil || req.InputDataConfig.DatasetGroupArn == "" {
		return "", &Error{
			Code:    errInvalidInputException,
			Message: "InputDataConfig.DatasetGroupArn is required",
		}
	}

	dg, exists := m.datasetGroups[req.InputDataConfig.DatasetGroupArn]
	if !exists {
		return "", &Error{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Dataset group %s not found", req.InputDataConfig.DatasetGroupArn),
		}
	}

	// Validate forecast horizon.
	if req.ForecastHorizon < 1 || req.ForecastHorizon > 500 {
		return "", &Error{
			Code:    errInvalidInputException,
			Message: "ForecastHorizon must be between 1 and 500",
		}
	}

	predictorID := uuid.New().String()[:8]
	predictorArn := fmt.Sprintf("arn:aws:forecast:%s:%s:predictor/%s_%s", m.region, m.accountID, req.PredictorName, predictorID)
	now := time.Now()

	forecastTypes := req.ForecastTypes
	if len(forecastTypes) == 0 {
		forecastTypes = []string{"0.1", "0.5", "0.9"}
	}

	predictor := &Predictor{
		PredictorArn:    predictorArn,
		PredictorName:   req.PredictorName,
		AlgorithmArn:    req.AlgorithmArn,
		ForecastHorizon: req.ForecastHorizon,
		ForecastTypes:   forecastTypes,
		InputDataConfig: &InputDataConfig{
			DatasetGroupArn: dg.DatasetGroupArn,
		},
		FeaturizationConfig:  req.FeaturizationConfig,
		Status:               statusActive,
		CreationTime:         ToAWSTimestamp(now),
		LastModificationTime: ToAWSTimestamp(now),
	}

	m.predictors[predictorArn] = predictor

	return predictorArn, nil
}

// DescribePredictor returns a predictor.
func (m *MemoryStorage) DescribePredictor(_ context.Context, predictorArn string) (*Predictor, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	predictor, exists := m.predictors[predictorArn]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Predictor %s not found", predictorArn),
		}
	}

	return predictor, nil
}

// ListPredictors returns all predictors.
func (m *MemoryStorage) ListPredictors(_ context.Context, maxResults *int32, _ string, filters []Filter) ([]*PredictorSummary, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limit := int32(100)
	if maxResults != nil && *maxResults > 0 && *maxResults < limit {
		limit = *maxResults
	}

	summaries := make([]*PredictorSummary, 0, len(m.predictors))

	for _, p := range m.predictors {
		if int32(len(summaries)) >= limit { //nolint:gosec // G115: len(summaries) is bounded by limit which is int32
			break
		}

		// Apply filters.
		if !matchesFilters(p, filters) {
			continue
		}

		datasetGroupArn := ""
		if p.InputDataConfig != nil {
			datasetGroupArn = p.InputDataConfig.DatasetGroupArn
		}

		summaries = append(summaries, &PredictorSummary{
			PredictorArn:         p.PredictorArn,
			PredictorName:        p.PredictorName,
			DatasetGroupArn:      datasetGroupArn,
			Status:               p.Status,
			CreationTime:         p.CreationTime,
			LastModificationTime: p.LastModificationTime,
			Message:              p.Message,
		})
	}

	return summaries, "", nil
}

// DeletePredictor deletes a predictor.
func (m *MemoryStorage) DeletePredictor(_ context.Context, predictorArn string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.predictors[predictorArn]; !exists {
		return &Error{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Predictor %s not found", predictorArn),
		}
	}

	// Check if predictor is in use by any forecast.
	for _, f := range m.forecasts {
		if f.PredictorArn == predictorArn {
			return &Error{
				Code:    errResourceInUseException,
				Message: fmt.Sprintf("Predictor %s is in use by forecast %s", predictorArn, f.ForecastArn),
			}
		}
	}

	delete(m.predictors, predictorArn)

	return nil
}

// Forecast operations.

// CreateForecast creates a new forecast.
func (m *MemoryStorage) CreateForecast(_ context.Context, req *CreateForecastInput) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate name.
	for _, f := range m.forecasts {
		if f.ForecastName == req.ForecastName {
			return "", &Error{
				Code:    errResourceAlreadyExistsException,
				Message: fmt.Sprintf("A forecast with name %s already exists", req.ForecastName),
			}
		}
	}

	// Validate predictor exists.
	predictor, exists := m.predictors[req.PredictorArn]
	if !exists {
		return "", &Error{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Predictor %s not found", req.PredictorArn),
		}
	}

	forecastID := uuid.New().String()[:8]
	forecastArn := fmt.Sprintf("arn:aws:forecast:%s:%s:forecast/%s_%s", m.region, m.accountID, req.ForecastName, forecastID)
	now := time.Now()

	forecastTypes := req.ForecastTypes
	if len(forecastTypes) == 0 {
		forecastTypes = predictor.ForecastTypes
	}

	datasetGroupArn := ""
	if predictor.InputDataConfig != nil {
		datasetGroupArn = predictor.InputDataConfig.DatasetGroupArn
	}

	forecast := &Forecast{
		ForecastArn:          forecastArn,
		ForecastName:         req.ForecastName,
		PredictorArn:         req.PredictorArn,
		DatasetGroupArn:      datasetGroupArn,
		ForecastTypes:        forecastTypes,
		Status:               statusActive,
		CreationTime:         ToAWSTimestamp(now),
		LastModificationTime: ToAWSTimestamp(now),
	}

	m.forecasts[forecastArn] = forecast

	return forecastArn, nil
}

// DescribeForecast returns a forecast.
func (m *MemoryStorage) DescribeForecast(_ context.Context, forecastArn string) (*Forecast, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	forecast, exists := m.forecasts[forecastArn]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Forecast %s not found", forecastArn),
		}
	}

	return forecast, nil
}

// ListForecasts returns all forecasts.
func (m *MemoryStorage) ListForecasts(_ context.Context, maxResults *int32, _ string, filters []Filter) ([]*ForecastSummary, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limit := int32(100)
	if maxResults != nil && *maxResults > 0 && *maxResults < limit {
		limit = *maxResults
	}

	summaries := make([]*ForecastSummary, 0, len(m.forecasts))

	for _, f := range m.forecasts {
		if int32(len(summaries)) >= limit { //nolint:gosec // G115: len(summaries) is bounded by limit which is int32
			break
		}

		// Apply filters.
		if !matchesForecastFilters(f, filters) {
			continue
		}

		summaries = append(summaries, &ForecastSummary{
			ForecastArn:          f.ForecastArn,
			ForecastName:         f.ForecastName,
			PredictorArn:         f.PredictorArn,
			DatasetGroupArn:      f.DatasetGroupArn,
			Status:               f.Status,
			CreationTime:         f.CreationTime,
			LastModificationTime: f.LastModificationTime,
			Message:              f.Message,
		})
	}

	return summaries, "", nil
}

// DeleteForecast deletes a forecast.
func (m *MemoryStorage) DeleteForecast(_ context.Context, forecastArn string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.forecasts[forecastArn]; !exists {
		return &Error{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Forecast %s not found", forecastArn),
		}
	}

	delete(m.forecasts, forecastArn)

	return nil
}

// Helper functions.

func isValidDatasetType(datasetType string) bool {
	switch datasetType {
	case datasetTypeTargetTimeSeries, datasetTypeRelatedTimeSeries, datasetTypeItemMetadata:
		return true
	default:
		return false
	}
}

func isValidDomain(domain string) bool {
	switch domain {
	case domainRetail, domainCustom, domainInventoryPlanning, domainEC2Capacity, domainWorkForce, domainWebTraffic, domainMetrics:
		return true
	default:
		return false
	}
}

func matchesFilters(p *Predictor, filters []Filter) bool {
	for _, f := range filters {
		switch f.Key {
		case "DatasetGroupArn":
			if p.InputDataConfig == nil || p.InputDataConfig.DatasetGroupArn != f.Value {
				return false
			}
		case "Status":
			if p.Status != f.Value {
				return false
			}
		}
	}

	return true
}

func matchesForecastFilters(f *Forecast, filters []Filter) bool {
	for _, filter := range filters {
		switch filter.Key {
		case "PredictorArn":
			if f.PredictorArn != filter.Value {
				return false
			}
		case "DatasetGroupArn":
			if f.DatasetGroupArn != filter.Value {
				return false
			}
		case "Status":
			if f.Status != filter.Value {
				return false
			}
		}
	}

	return true
}

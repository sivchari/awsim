package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sivchari/kumo/internal/storage"
)

const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "000000000000"
)

// Storage defines the interface for DynamoDB storage operations.
type Storage interface {
	CreateTable(ctx context.Context, req *CreateTableRequest) (*Table, error)
	DeleteTable(ctx context.Context, tableName string) (*Table, error)
	ListTables(ctx context.Context, exclusiveStartTableName string, limit int) ([]string, string, error)
	DescribeTable(ctx context.Context, tableName string) (*Table, error)
	PutItem(ctx context.Context, tableName string, item Item, returnOld bool, cond ConditionInput) (Item, error)
	GetItem(ctx context.Context, tableName string, key Item) (Item, error)
	DeleteItem(ctx context.Context, tableName string, key Item, returnOld bool, cond ConditionInput) (Item, error)
	UpdateItem(ctx context.Context, tableName string, key Item, updateExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, returnValues string, cond ConditionInput) (Item, error)
	Query(ctx context.Context, tableName, indexName string, keyCondExpr string, filterExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, limit int, exclusiveStartKey Item, scanForward bool) ([]Item, Item, int, error)
	Scan(ctx context.Context, tableName string, filterExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, limit int, exclusiveStartKey Item) ([]Item, Item, int, error)
	TransactWriteItems(ctx context.Context, items []TransactWriteItem) ([]CancellationReason, error)
	TransactGetItems(ctx context.Context, items []TransactGetItem) ([]Item, error)
	BatchWriteItem(ctx context.Context, requestItems map[string][]WriteRequest) (map[string][]WriteRequest, error)
	BatchGetItem(ctx context.Context, requestItems map[string]KeysAndAttributes) (map[string][]Item, error)
	UpdateTimeToLive(ctx context.Context, tableName, attributeName string, enabled bool) error
	DescribeTimeToLive(ctx context.Context, tableName string) (string, bool, error)
}

// Option is a configuration option for MemoryStorage.
type Option func(*MemoryStorage)

// WithDataDir enables persistent storage in the specified directory.
func WithDataDir(dir string) Option {
	return func(s *MemoryStorage) {
		s.dataDir = dir
	}
}

// Compile-time interface checks.
var (
	_ json.Marshaler   = (*MemoryStorage)(nil)
	_ json.Unmarshaler = (*MemoryStorage)(nil)
)

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu      sync.RWMutex          `json:"-"`
	Tables  map[string]*tableData `json:"tables"`
	baseURL string
	dataDir string
}

type tableData struct {
	Table *Table          `json:"table"`
	Items map[string]Item `json:"items"`
}

// NewMemoryStorage creates a new in-memory DynamoDB storage.
func NewMemoryStorage(baseURL string, opts ...Option) *MemoryStorage {
	s := &MemoryStorage{
		Tables:  make(map[string]*tableData),
		baseURL: baseURL,
	}
	for _, o := range opts {
		o(s)
	}

	if s.dataDir != "" {
		_ = storage.Load(s.dataDir, "dynamodb", s)
	}

	return s
}

// MarshalJSON serializes the storage state to JSON.
func (m *MemoryStorage) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type Alias MemoryStorage

	data, err := json.Marshal(&struct{ *Alias }{Alias: (*Alias)(m)})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	return data, nil
}

// UnmarshalJSON restores the storage state from JSON.
func (m *MemoryStorage) UnmarshalJSON(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	type Alias MemoryStorage

	aux := &struct{ *Alias }{Alias: (*Alias)(m)}

	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	if m.Tables == nil {
		m.Tables = make(map[string]*tableData)
	}

	return nil
}

// Close saves the storage state to disk if persistence is enabled.
func (m *MemoryStorage) Close() error {
	if m.dataDir == "" {
		return nil
	}

	if err := storage.Save(m.dataDir, "dynamodb", m); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}

// CreateTable creates a new table.
func (m *MemoryStorage) CreateTable(_ context.Context, req *CreateTableRequest) (*Table, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.Tables[req.TableName]; exists {
		return nil, &TableError{
			Code:    "ResourceInUseException",
			Message: fmt.Sprintf("Table already exists: %s", req.TableName),
		}
	}

	billingMode := req.BillingMode
	if billingMode == "" {
		billingMode = "PROVISIONED"
	}

	table := &Table{
		Name:                   req.TableName,
		KeySchema:              req.KeySchema,
		AttributeDefinitions:   req.AttributeDefinitions,
		ProvisionedThroughput:  req.ProvisionedThroughput,
		GlobalSecondaryIndexes: req.GlobalSecondaryIndexes,
		CreationDateTime:       time.Now(),
		TableStatus:            "ACTIVE",
		ItemCount:              0,
		TableSizeBytes:         0,
		TableARN:               fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/%s", defaultRegion, defaultAccountID, req.TableName),
		BillingMode:            billingMode,
		DeletionProtection:     req.DeletionProtectionEnabled,
	}

	m.Tables[req.TableName] = &tableData{
		Table: table,
		Items: make(map[string]Item),
	}

	return table, nil
}

// DeleteTable deletes a table.
func (m *MemoryStorage) DeleteTable(_ context.Context, tableName string) (*Table, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	td, exists := m.Tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	table := td.Table
	table.TableStatus = "DELETING"

	delete(m.Tables, tableName)

	return table, nil
}

// ListTables lists all tables.
func (m *MemoryStorage) ListTables(_ context.Context, exclusiveStartTableName string, limit int) ([]string, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	names := make([]string, 0, len(m.Tables))
	for name := range m.Tables {
		names = append(names, name)
	}

	sort.Strings(names)

	// Apply exclusive start.
	startIdx := 0

	if exclusiveStartTableName != "" {
		for i, name := range names {
			if name > exclusiveStartTableName {
				startIdx = i

				break
			}
		}
	}

	// Apply limit.
	endIdx := startIdx + limit
	if endIdx > len(names) {
		endIdx = len(names)
	}

	result := names[startIdx:endIdx]

	var lastEvaluated string
	if endIdx < len(names) {
		lastEvaluated = result[len(result)-1]
	}

	return result, lastEvaluated, nil
}

// DescribeTable describes a table.
func (m *MemoryStorage) DescribeTable(_ context.Context, tableName string) (*Table, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	td, exists := m.Tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	// Update item count.
	td.Table.ItemCount = int64(len(td.Items))

	return td.Table, nil
}

// PutItem puts an item into a table.
func (m *MemoryStorage) PutItem(_ context.Context, tableName string, item Item, returnOld bool, cond ConditionInput) (Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	td, exists := m.Tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	key := m.serializeKey(td.Table, item)

	// Evaluate condition against existing item (nil if not exists).
	var existingItem Item
	if existing, ok := td.Items[key]; ok {
		existingItem = existing
	}

	if ok, err := evaluateCondition(existingItem, cond); err != nil {
		return nil, &TableError{
			Code:    "ValidationException",
			Message: fmt.Sprintf("Invalid ConditionExpression: %s", err),
		}
	} else if !ok {
		return nil, &TableError{
			Code:    ErrCodeConditionalCheckFailed,
			Message: "The conditional request failed",
		}
	}

	var oldItem Item

	if returnOld && existingItem != nil {
		oldItem = m.copyItem(existingItem)
	}

	td.Items[key] = m.copyItem(item)

	return oldItem, nil
}

// GetItem gets an item from a table.
func (m *MemoryStorage) GetItem(_ context.Context, tableName string, key Item) (Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	td, exists := m.Tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	keyStr := m.serializeKey(td.Table, key)
	if item, ok := td.Items[keyStr]; ok {
		return m.copyItem(item), nil
	}

	//nolint:nilnil // DynamoDB returns nil item when key not found (valid behavior).
	return nil, nil
}

// DeleteItem deletes an item from a table.
func (m *MemoryStorage) DeleteItem(_ context.Context, tableName string, key Item, returnOld bool, cond ConditionInput) (Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	td, exists := m.Tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	keyStr := m.serializeKey(td.Table, key)

	// Evaluate condition against existing item.
	var existingItem Item
	if existing, ok := td.Items[keyStr]; ok {
		existingItem = existing
	}

	if ok, err := evaluateCondition(existingItem, cond); err != nil {
		return nil, &TableError{
			Code:    "ValidationException",
			Message: fmt.Sprintf("Invalid ConditionExpression: %s", err),
		}
	} else if !ok {
		return nil, &TableError{
			Code:    ErrCodeConditionalCheckFailed,
			Message: "The conditional request failed",
		}
	}

	var oldItem Item

	if existingItem != nil {
		if returnOld {
			oldItem = m.copyItem(existingItem)
		}

		delete(td.Items, keyStr)
	}

	return oldItem, nil
}

// UpdateItem updates an item in a table.
func (m *MemoryStorage) UpdateItem(_ context.Context, tableName string, key Item, updateExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, returnValues string, cond ConditionInput) (Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	td, exists := m.Tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	keyStr := m.serializeKey(td.Table, key)
	item, itemExists := td.Items[keyStr]

	// Evaluate condition against existing item.
	var condItem Item
	if itemExists {
		condItem = item
	}

	if ok, err := evaluateCondition(condItem, cond); err != nil {
		return nil, &TableError{
			Code:    "ValidationException",
			Message: fmt.Sprintf("Invalid ConditionExpression: %s", err),
		}
	} else if !ok {
		return nil, &TableError{
			Code:    ErrCodeConditionalCheckFailed,
			Message: "The conditional request failed",
		}
	}

	var oldItem Item
	if itemExists {
		oldItem = m.copyItem(item)
	} else {
		// Create new item with key attributes.
		item = m.copyItem(key)
	}

	// Parse and apply update expression.
	if updateExpr != "" {
		item = m.applyUpdateExpression(item, updateExpr, exprNames, exprValues)
	}

	td.Items[keyStr] = item

	// Return based on returnValues.
	switch returnValues {
	case ReturnValuesAllOld:
		return oldItem, nil
	case ReturnValuesAllNew:
		return m.copyItem(item), nil
	case ReturnValuesUpdatedOld, ReturnValuesUpdatedNew:
		// Simplified: return all attributes.
		if returnValues == ReturnValuesUpdatedOld {
			return oldItem, nil
		}

		return m.copyItem(item), nil
	default:
		//nolint:nilnil // DynamoDB returns nil when ReturnValues is NONE (valid behavior).
		return nil, nil
	}
}

// Query queries items from a table.
//
//nolint:cyclop,funlen,gocognit // Query has inherent complexity from DynamoDB protocol requirements.
func (m *MemoryStorage) Query(_ context.Context, tableName, indexName, keyCondExpr, filterExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, limit int, exclusiveStartKey Item, scanForward bool) ([]Item, Item, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	td, exists := m.Tables[tableName]
	if !exists {
		return nil, nil, 0, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	// Determine key schema to use (table or GSI).
	keySchema := td.Table.KeySchema

	if indexName != "" {
		found := false

		for _, gsi := range td.Table.GlobalSecondaryIndexes {
			if gsi.IndexName == indexName {
				keySchema = gsi.KeySchema
				found = true

				break
			}
		}

		if !found {
			return nil, nil, 0, &TableError{
				Code:    "ValidationException",
				Message: fmt.Sprintf("The table does not have the specified index: %s", indexName),
			}
		}
	}

	// Get partition key attribute name from the resolved key schema.
	var partitionKeyName string

	for _, ks := range keySchema {
		if ks.KeyType == "HASH" {
			partitionKeyName = ks.AttributeName

			break
		}
	}

	// Parse key condition to extract partition key value.
	partitionKeyValue := m.extractPartitionKeyValue(keyCondExpr, partitionKeyName, exprNames, exprValues)

	// Collect matching items.
	var results []Item

	scannedCount := 0

	for _, item := range td.Items {
		scannedCount++

		// Check partition key match.
		if partitionKeyValue != nil {
			if itemVal, ok := item[partitionKeyName]; ok {
				if !m.attributeValuesEqual(itemVal, *partitionKeyValue) {
					continue
				}
			} else {
				continue
			}
		}

		// Apply filter expression (simplified).
		if filterExpr != "" && !m.evaluateFilterExpression(item, filterExpr, exprNames, exprValues) {
			continue
		}

		results = append(results, m.copyItem(item))
	}

	// Sort results.
	if !scanForward {
		// Reverse order.
		for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
			results[i], results[j] = results[j], results[i]
		}
	}

	// Apply pagination.
	startIdx := 0

	if exclusiveStartKey != nil {
		startKeyStr := m.serializeKey(td.Table, exclusiveStartKey)

		for i, item := range results {
			itemKeyStr := m.serializeKey(td.Table, item)
			if itemKeyStr == startKeyStr {
				startIdx = i + 1

				break
			}
		}
	}

	if startIdx >= len(results) {
		return []Item{}, nil, scannedCount, nil
	}

	results = results[startIdx:]

	var lastEvaluatedKey Item

	if limit > 0 && len(results) > limit {
		results = results[:limit]
		lastEvaluatedKey = m.extractKey(td.Table, results[len(results)-1])
	}

	return results, lastEvaluatedKey, scannedCount, nil
}

// Scan scans items from a table.
//
//nolint:funlen // Scan requires pagination logic that exceeds line limit.
func (m *MemoryStorage) Scan(_ context.Context, tableName, filterExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, limit int, exclusiveStartKey Item) ([]Item, Item, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	td, exists := m.Tables[tableName]
	if !exists {
		return nil, nil, 0, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	// Collect all items.
	var results []Item

	scannedCount := 0

	for _, item := range td.Items {
		scannedCount++

		// Apply filter expression.
		if filterExpr != "" && !m.evaluateFilterExpression(item, filterExpr, exprNames, exprValues) {
			continue
		}

		results = append(results, m.copyItem(item))
	}

	// Sort by key for consistent pagination.
	sort.Slice(results, func(i, j int) bool {
		keyI := m.serializeKey(td.Table, results[i])
		keyJ := m.serializeKey(td.Table, results[j])

		return keyI < keyJ
	})

	// Apply pagination.
	startIdx := 0

	if exclusiveStartKey != nil {
		startKeyStr := m.serializeKey(td.Table, exclusiveStartKey)

		for i, item := range results {
			itemKeyStr := m.serializeKey(td.Table, item)
			if itemKeyStr == startKeyStr {
				startIdx = i + 1

				break
			}
		}
	}

	if startIdx >= len(results) {
		return []Item{}, nil, scannedCount, nil
	}

	results = results[startIdx:]

	var lastEvaluatedKey Item

	if limit > 0 && len(results) > limit {
		results = results[:limit]
		lastEvaluatedKey = m.extractKey(td.Table, results[len(results)-1])
	}

	return results, lastEvaluatedKey, scannedCount, nil
}

// serializeKey creates a string key from the primary key attributes.
func (m *MemoryStorage) serializeKey(table *Table, item Item) string {
	var parts []string

	for _, ks := range table.KeySchema {
		if val, ok := item[ks.AttributeName]; ok {
			parts = append(parts, m.serializeAttributeValue(val))
		}
	}

	return strings.Join(parts, "|")
}

// serializeAttributeValue serializes an attribute value to a string.
//
//nolint:gocritic // hugeParam: AttributeValue must be passed by value to avoid mutation.
func (m *MemoryStorage) serializeAttributeValue(av AttributeValue) string {
	if av.S != nil {
		return "S:" + *av.S
	}

	if av.N != nil {
		return "N:" + *av.N
	}

	if av.B != nil {
		return "B:" + string(av.B)
	}

	return "NULL:" + uuid.New().String()
}

// copyItem creates a deep copy of an item.
//
//nolint:gocritic // rangeValCopy: intentional copy for deep clone operation.
func (m *MemoryStorage) copyItem(item Item) Item {
	if item == nil {
		return nil
	}

	result := make(Item)

	for k, v := range item {
		result[k] = m.copyAttributeValue(v)
	}

	return result
}

// copyAttributeValue creates a deep copy of an attribute value.
//
//nolint:funlen,gocritic // Deep copy of all AttributeValue fields requires many statements.
func (m *MemoryStorage) copyAttributeValue(av AttributeValue) AttributeValue {
	result := AttributeValue{}

	if av.S != nil {
		s := *av.S
		result.S = &s
	}

	if av.N != nil {
		n := *av.N
		result.N = &n
	}

	if av.B != nil {
		b := make([]byte, len(av.B))
		copy(b, av.B)
		result.B = b
	}

	if av.SS != nil {
		ss := make([]string, len(av.SS))
		copy(ss, av.SS)
		result.SS = ss
	}

	if av.NS != nil {
		ns := make([]string, len(av.NS))
		copy(ns, av.NS)
		result.NS = ns
	}

	if av.BS != nil {
		bs := make([][]byte, len(av.BS))
		for i, b := range av.BS {
			bs[i] = make([]byte, len(b))
			copy(bs[i], b)
		}

		result.BS = bs
	}

	if av.M != nil {
		mapCopy := make(map[string]AttributeValue)

		for k, v := range av.M {
			mapCopy[k] = m.copyAttributeValue(v)
		}

		result.M = mapCopy
	}

	if av.L != nil {
		listCopy := make([]AttributeValue, len(av.L))

		for i, v := range av.L {
			listCopy[i] = m.copyAttributeValue(v)
		}

		result.L = listCopy
	}

	if av.NULL != nil {
		n := *av.NULL
		result.NULL = &n
	}

	if av.BOOL != nil {
		b := *av.BOOL
		result.BOOL = &b
	}

	return result
}

// extractKey extracts the primary key from an item.
func (m *MemoryStorage) extractKey(table *Table, item Item) Item {
	key := make(Item)

	for _, ks := range table.KeySchema {
		if val, ok := item[ks.AttributeName]; ok {
			key[ks.AttributeName] = val
		}
	}

	return key
}

// extractPartitionKeyValue extracts the partition key value from a key condition expression.
func (m *MemoryStorage) extractPartitionKeyValue(keyCondExpr, partitionKeyName string, exprNames map[string]string, exprValues map[string]AttributeValue) *AttributeValue {
	if keyCondExpr == "" {
		return nil
	}

	// Simple parsing: look for "attrName = :value" pattern.
	// Replace expression attribute names.
	expr := keyCondExpr
	for placeholder, name := range exprNames {
		expr = strings.ReplaceAll(expr, placeholder, name)
	}

	// Look for partition key equality.
	parts := strings.Split(expr, " AND ")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		//nolint:nestif // Parsing key condition expression requires nested validation.
		if strings.Contains(part, "=") {
			eqParts := strings.SplitN(part, "=", 2)
			if len(eqParts) == 2 {
				attrName := strings.TrimSpace(eqParts[0])
				valuePlaceholder := strings.TrimSpace(eqParts[1])

				if attrName == partitionKeyName {
					if val, ok := exprValues[valuePlaceholder]; ok {
						return &val
					}
				}
			}
		}
	}

	return nil
}

// evaluateFilterExpression evaluates a filter expression against an item.
func (m *MemoryStorage) evaluateFilterExpression(item Item, filterExpr string, exprNames map[string]string, exprValues map[string]AttributeValue) bool {
	result, err := evaluateCondition(item, ConditionInput{
		Expression: filterExpr,
		ExprNames:  exprNames,
		ExprValues: exprValues,
	})
	if err != nil {
		return true
	}

	return result
}

// attributeValuesEqual compares two attribute values for equality.
//
//nolint:gocritic // hugeParam: AttributeValue passed by value for comparison.
func (m *MemoryStorage) attributeValuesEqual(a, b AttributeValue) bool {
	if a.S != nil && b.S != nil {
		return *a.S == *b.S
	}

	if a.N != nil && b.N != nil {
		return *a.N == *b.N
	}

	if a.BOOL != nil && b.BOOL != nil {
		return *a.BOOL == *b.BOOL
	}

	return false
}

// applyUpdateExpression applies an update expression to an item.
func (m *MemoryStorage) applyUpdateExpression(item Item, updateExpr string, exprNames map[string]string, exprValues map[string]AttributeValue) Item {
	// Very simplified update expression parsing.
	// Only supports "SET attr = :val" pattern.
	expr := updateExpr
	for placeholder, name := range exprNames {
		expr = strings.ReplaceAll(expr, placeholder, name)
	}

	// Parse SET clause.
	if strings.HasPrefix(strings.ToUpper(expr), "SET ") {
		setClause := strings.TrimPrefix(expr, "SET ")
		setClause = strings.TrimPrefix(setClause, "set ")

		assignments := strings.Split(setClause, ",")
		for _, assignment := range assignments {
			parts := strings.SplitN(strings.TrimSpace(assignment), "=", 2)
			if len(parts) == 2 {
				attrName := strings.TrimSpace(parts[0])
				valuePlaceholder := strings.TrimSpace(parts[1])

				if val, ok := exprValues[valuePlaceholder]; ok {
					item[attrName] = val
				}
			}
		}
	}

	return item
}

// TransactWriteItems executes a transactional write with all-or-nothing semantics.
func (m *MemoryStorage) TransactWriteItems(_ context.Context, items []TransactWriteItem) ([]CancellationReason, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	reasons := make([]CancellationReason, len(items))
	hasFailure := false

	// Phase 1: Validate all conditions without modifying state.
	for i, twi := range items {
		reason, err := m.validateTransactWriteItem(twi)
		if err != nil {
			return nil, err
		}

		if reason != nil {
			reasons[i] = *reason
			hasFailure = true
		}
	}

	if hasFailure {
		return reasons, &TableError{
			Code:    "TransactionCanceledException",
			Message: "Transaction cancelled, please refer cancellation reasons for specific reasons [CancellationReason]",
		}
	}

	// Phase 2: Apply all mutations atomically.
	for _, twi := range items {
		m.applyTransactWriteItem(twi)
	}

	return nil, nil //nolint:nilnil // Success: nil CancellationReasons means no failures.
}

// validateTransactWriteItem validates a single write item's condition without applying changes.
func (m *MemoryStorage) validateTransactWriteItem(twi TransactWriteItem) (*CancellationReason, error) {
	switch {
	case twi.Put != nil:
		return m.checkTransactCondition(twi.Put.TableName, twi.Put.Item, ConditionInput{
			Expression: twi.Put.ConditionExpression, ExprNames: twi.Put.ExpressionAttributeNames, ExprValues: twi.Put.ExpressionAttributeValues,
		})
	case twi.Delete != nil:
		return m.checkTransactCondition(twi.Delete.TableName, twi.Delete.Key, ConditionInput{
			Expression: twi.Delete.ConditionExpression, ExprNames: twi.Delete.ExpressionAttributeNames, ExprValues: twi.Delete.ExpressionAttributeValues,
		})
	case twi.Update != nil:
		return m.checkTransactCondition(twi.Update.TableName, twi.Update.Key, ConditionInput{
			Expression: twi.Update.ConditionExpression, ExprNames: twi.Update.ExpressionAttributeNames, ExprValues: twi.Update.ExpressionAttributeValues,
		})
	case twi.ConditionCheck != nil:
		return m.checkTransactCondition(twi.ConditionCheck.TableName, twi.ConditionCheck.Key, ConditionInput{
			Expression: twi.ConditionCheck.ConditionExpression, ExprNames: twi.ConditionCheck.ExpressionAttributeNames, ExprValues: twi.ConditionCheck.ExpressionAttributeValues,
		})
	}

	//nolint:nilnil // No action specified is valid (returns success).
	return nil, nil
}

// checkTransactCondition checks a condition against the existing item in a table.
// Must be called under lock.
func (m *MemoryStorage) checkTransactCondition(tableName string, keyOrItem Item, cond ConditionInput) (*CancellationReason, error) {
	td, exists := m.Tables[tableName]
	if !exists {
		return nil, &TableError{Code: "ResourceNotFoundException", Message: fmt.Sprintf("Table: %s not found", tableName)}
	}

	key := m.serializeKey(td.Table, keyOrItem)

	var existing Item
	if e, ok := td.Items[key]; ok {
		existing = e
	}

	ok, err := evaluateCondition(existing, cond)
	if err != nil {
		return &CancellationReason{Code: "ValidationError", Message: err.Error()}, nil //nolint:nilerr // Condition error is returned as cancellation reason, not as error.
	}

	if !ok {
		return &CancellationReason{Code: "ConditionalCheckFailed"}, nil
	}

	return nil, nil //nolint:nilnil // Condition passed, no cancellation reason.
}

// applyTransactWriteItem applies a single write item mutation. Must be called under lock.
func (m *MemoryStorage) applyTransactWriteItem(twi TransactWriteItem) {
	switch {
	case twi.Put != nil:
		td := m.Tables[twi.Put.TableName]
		key := m.serializeKey(td.Table, twi.Put.Item)
		td.Items[key] = m.copyItem(twi.Put.Item)

	case twi.Delete != nil:
		td := m.Tables[twi.Delete.TableName]
		key := m.serializeKey(td.Table, twi.Delete.Key)
		delete(td.Items, key)

	case twi.Update != nil:
		td := m.Tables[twi.Update.TableName]
		key := m.serializeKey(td.Table, twi.Update.Key)

		item, ok := td.Items[key]
		if !ok {
			item = m.copyItem(twi.Update.Key)
		}

		if twi.Update.UpdateExpression != "" {
			item = m.applyUpdateExpression(item, twi.Update.UpdateExpression, twi.Update.ExpressionAttributeNames, twi.Update.ExpressionAttributeValues)
		}

		td.Items[key] = item
	case twi.ConditionCheck != nil:
	}
}

// TransactGetItems retrieves multiple items transactionally.
func (m *MemoryStorage) TransactGetItems(_ context.Context, items []TransactGetItem) ([]Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]Item, len(items))

	for i, tgi := range items {
		if tgi.Get == nil {
			continue
		}

		td, exists := m.Tables[tgi.Get.TableName]
		if !exists {
			return nil, &TableError{
				Code:    "ResourceNotFoundException",
				Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tgi.Get.TableName),
			}
		}

		key := m.serializeKey(td.Table, tgi.Get.Key)
		if item, ok := td.Items[key]; ok {
			results[i] = m.copyItem(item)
		}
	}

	return results, nil
}

// BatchWriteItem writes/deletes multiple items across tables.
func (m *MemoryStorage) BatchWriteItem(_ context.Context, requestItems map[string][]WriteRequest) (map[string][]WriteRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for tableName, requests := range requestItems {
		td, exists := m.Tables[tableName]
		if !exists {
			return nil, &TableError{
				Code:    "ResourceNotFoundException",
				Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
			}
		}

		for _, req := range requests {
			switch {
			case req.PutRequest != nil:
				key := m.serializeKey(td.Table, req.PutRequest.Item)
				td.Items[key] = m.copyItem(req.PutRequest.Item)
			case req.DeleteRequest != nil:
				key := m.serializeKey(td.Table, req.DeleteRequest.Key)
				delete(td.Items, key)
			}
		}
	}

	// kumo processes all items; never returns UnprocessedItems.
	return nil, nil //nolint:nilnil // Intentional: nil UnprocessedItems means all items were processed.
}

// BatchGetItem retrieves multiple items across tables.
func (m *MemoryStorage) BatchGetItem(_ context.Context, requestItems map[string]KeysAndAttributes) (map[string][]Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	responses := make(map[string][]Item)

	for tableName, ka := range requestItems {
		td, exists := m.Tables[tableName]
		if !exists {
			return nil, &TableError{
				Code:    "ResourceNotFoundException",
				Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
			}
		}

		var items []Item

		for _, key := range ka.Keys {
			keyStr := m.serializeKey(td.Table, key)
			if item, ok := td.Items[keyStr]; ok {
				items = append(items, m.copyItem(item))
			}
		}

		if len(items) > 0 {
			responses[tableName] = items
		}
	}

	return responses, nil
}

// UpdateTimeToLive updates the TTL configuration for a table.
func (m *MemoryStorage) UpdateTimeToLive(_ context.Context, tableName, attributeName string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	td, exists := m.Tables[tableName]
	if !exists {
		return &TableError{Code: "ResourceNotFoundException", Message: "Requested resource not found"}
	}

	td.Table.TTLAttributeName = attributeName
	td.Table.TTLEnabled = enabled

	return nil
}

// DescribeTimeToLive returns the TTL configuration for a table.
func (m *MemoryStorage) DescribeTimeToLive(_ context.Context, tableName string) (string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	td, exists := m.Tables[tableName]
	if !exists {
		return "", false, &TableError{Code: "ResourceNotFoundException", Message: "Requested resource not found"}
	}

	return td.Table.TTLAttributeName, td.Table.TTLEnabled, nil
}

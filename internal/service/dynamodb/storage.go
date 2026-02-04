package dynamodb

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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
	PutItem(ctx context.Context, tableName string, item Item, returnOld bool) (Item, error)
	GetItem(ctx context.Context, tableName string, key Item) (Item, error)
	DeleteItem(ctx context.Context, tableName string, key Item, returnOld bool) (Item, error)
	UpdateItem(ctx context.Context, tableName string, key Item, updateExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, returnValues string) (Item, error)
	Query(ctx context.Context, tableName string, keyCondExpr string, filterExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, limit int, exclusiveStartKey Item, scanForward bool) ([]Item, Item, int, error)
	Scan(ctx context.Context, tableName string, filterExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, limit int, exclusiveStartKey Item) ([]Item, Item, int, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu      sync.RWMutex
	tables  map[string]*tableData
	baseURL string
}

type tableData struct {
	table *Table
	items map[string]Item // key is the serialized primary key
}

// NewMemoryStorage creates a new in-memory DynamoDB storage.
func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		tables:  make(map[string]*tableData),
		baseURL: baseURL,
	}
}

// CreateTable creates a new table.
func (m *MemoryStorage) CreateTable(_ context.Context, req *CreateTableRequest) (*Table, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tables[req.TableName]; exists {
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
		Name:                  req.TableName,
		KeySchema:             req.KeySchema,
		AttributeDefinitions:  req.AttributeDefinitions,
		ProvisionedThroughput: req.ProvisionedThroughput,
		CreationDateTime:      time.Now(),
		TableStatus:           "ACTIVE",
		ItemCount:             0,
		TableSizeBytes:        0,
		TableARN:              fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/%s", defaultRegion, defaultAccountID, req.TableName),
		BillingMode:           billingMode,
		DeletionProtection:    req.DeletionProtectionEnabled,
	}

	m.tables[req.TableName] = &tableData{
		table: table,
		items: make(map[string]Item),
	}

	return table, nil
}

// DeleteTable deletes a table.
func (m *MemoryStorage) DeleteTable(_ context.Context, tableName string) (*Table, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	td, exists := m.tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	table := td.table
	table.TableStatus = "DELETING"

	delete(m.tables, tableName)

	return table, nil
}

// ListTables lists all tables.
func (m *MemoryStorage) ListTables(_ context.Context, exclusiveStartTableName string, limit int) ([]string, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	names := make([]string, 0, len(m.tables))
	for name := range m.tables {
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

	td, exists := m.tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	// Update item count.
	td.table.ItemCount = int64(len(td.items))

	return td.table, nil
}

// PutItem puts an item into a table.
func (m *MemoryStorage) PutItem(_ context.Context, tableName string, item Item, returnOld bool) (Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	td, exists := m.tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	key := m.serializeKey(td.table, item)

	var oldItem Item

	if returnOld {
		if existing, ok := td.items[key]; ok {
			oldItem = m.copyItem(existing)
		}
	}

	td.items[key] = m.copyItem(item)

	return oldItem, nil
}

// GetItem gets an item from a table.
func (m *MemoryStorage) GetItem(_ context.Context, tableName string, key Item) (Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	td, exists := m.tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	keyStr := m.serializeKey(td.table, key)
	if item, ok := td.items[keyStr]; ok {
		return m.copyItem(item), nil
	}

	//nolint:nilnil // DynamoDB returns nil item when key not found (valid behavior).
	return nil, nil
}

// DeleteItem deletes an item from a table.
func (m *MemoryStorage) DeleteItem(_ context.Context, tableName string, key Item, returnOld bool) (Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	td, exists := m.tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	keyStr := m.serializeKey(td.table, key)

	var oldItem Item

	if existing, ok := td.items[keyStr]; ok {
		if returnOld {
			oldItem = m.copyItem(existing)
		}

		delete(td.items, keyStr)
	}

	return oldItem, nil
}

// UpdateItem updates an item in a table.
func (m *MemoryStorage) UpdateItem(_ context.Context, tableName string, key Item, updateExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, returnValues string) (Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	td, exists := m.tables[tableName]
	if !exists {
		return nil, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	keyStr := m.serializeKey(td.table, key)
	item, itemExists := td.items[keyStr]

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

	td.items[keyStr] = item

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
//nolint:cyclop,funlen // Query has inherent complexity from DynamoDB protocol requirements.
func (m *MemoryStorage) Query(_ context.Context, tableName, keyCondExpr, filterExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, limit int, exclusiveStartKey Item, scanForward bool) ([]Item, Item, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	td, exists := m.tables[tableName]
	if !exists {
		return nil, nil, 0, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	// Get partition key attribute name.
	var partitionKeyName string

	for _, ks := range td.table.KeySchema {
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

	for _, item := range td.items {
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
		startKeyStr := m.serializeKey(td.table, exclusiveStartKey)

		for i, item := range results {
			itemKeyStr := m.serializeKey(td.table, item)
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
		lastEvaluatedKey = m.extractKey(td.table, results[len(results)-1])
	}

	return results, lastEvaluatedKey, scannedCount, nil
}

// Scan scans items from a table.
//
//nolint:funlen // Scan requires pagination logic that exceeds line limit.
func (m *MemoryStorage) Scan(_ context.Context, tableName, filterExpr string, exprNames map[string]string, exprValues map[string]AttributeValue, limit int, exclusiveStartKey Item) ([]Item, Item, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	td, exists := m.tables[tableName]
	if !exists {
		return nil, nil, 0, &TableError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Requested resource not found: Table: %s not found", tableName),
		}
	}

	// Collect all items.
	var results []Item

	scannedCount := 0

	for _, item := range td.items {
		scannedCount++

		// Apply filter expression.
		if filterExpr != "" && !m.evaluateFilterExpression(item, filterExpr, exprNames, exprValues) {
			continue
		}

		results = append(results, m.copyItem(item))
	}

	// Sort by key for consistent pagination.
	sort.Slice(results, func(i, j int) bool {
		keyI := m.serializeKey(td.table, results[i])
		keyJ := m.serializeKey(td.table, results[j])

		return keyI < keyJ
	})

	// Apply pagination.
	startIdx := 0

	if exclusiveStartKey != nil {
		startKeyStr := m.serializeKey(td.table, exclusiveStartKey)

		for i, item := range results {
			itemKeyStr := m.serializeKey(td.table, item)
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
		lastEvaluatedKey = m.extractKey(td.table, results[len(results)-1])
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

// evaluateFilterExpression evaluates a simple filter expression against an item.
func (m *MemoryStorage) evaluateFilterExpression(item Item, filterExpr string, exprNames map[string]string, exprValues map[string]AttributeValue) bool {
	if filterExpr == "" {
		return true
	}

	// Very simplified filter expression evaluation.
	// Only supports simple equality checks.
	expr := filterExpr
	for placeholder, name := range exprNames {
		expr = strings.ReplaceAll(expr, placeholder, name)
	}

	// Handle simple equality: "attr = :val".
	if strings.Contains(expr, "=") && !strings.Contains(expr, "<>") {
		parts := strings.SplitN(expr, "=", 2)
		if len(parts) == 2 {
			attrName := strings.TrimSpace(parts[0])
			valuePlaceholder := strings.TrimSpace(parts[1])

			itemVal, hasAttr := item[attrName]
			exprVal, hasExpr := exprValues[valuePlaceholder]

			if !hasAttr || !hasExpr {
				return false
			}

			return m.attributeValuesEqual(itemVal, exprVal)
		}
	}

	// Default to true for unsupported expressions.
	return true
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

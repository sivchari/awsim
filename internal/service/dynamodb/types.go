// Package dynamodb provides DynamoDB service emulation for kumo.
package dynamodb

import (
	"time"
)

// ReturnValues constants for DynamoDB operations.
const (
	ReturnValuesAllOld     = "ALL_OLD"
	ReturnValuesAllNew     = "ALL_NEW"
	ReturnValuesUpdatedOld = "UPDATED_OLD"
	ReturnValuesUpdatedNew = "UPDATED_NEW"
)

// Error code constants.
const (
	ErrCodeConditionalCheckFailed = "ConditionalCheckFailedException"
)

// AttributeValue represents a DynamoDB attribute value.
// Only one field should be set at a time.
type AttributeValue struct {
	S    *string                   `json:"S,omitempty"`
	N    *string                   `json:"N,omitempty"`
	B    []byte                    `json:"B,omitempty"`
	SS   []string                  `json:"SS,omitempty"`
	NS   []string                  `json:"NS,omitempty"`
	BS   [][]byte                  `json:"BS,omitempty"`
	M    map[string]AttributeValue `json:"M,omitempty"`
	L    []AttributeValue          `json:"L,omitempty"`
	NULL *bool                     `json:"NULL,omitempty"`
	BOOL *bool                     `json:"BOOL,omitempty"`
}

// KeySchemaElement represents a key schema element.
type KeySchemaElement struct {
	AttributeName string `json:"AttributeName"`
	KeyType       string `json:"KeyType"` // HASH or RANGE
}

// AttributeDefinition represents an attribute definition.
type AttributeDefinition struct {
	AttributeName string `json:"AttributeName"`
	AttributeType string `json:"AttributeType"` // S, N, or B
}

// ProvisionedThroughput represents provisioned throughput settings.
type ProvisionedThroughput struct {
	ReadCapacityUnits  int64 `json:"ReadCapacityUnits"`
	WriteCapacityUnits int64 `json:"WriteCapacityUnits"`
}

// ProvisionedThroughputDescription represents provisioned throughput description.
type ProvisionedThroughputDescription struct {
	ReadCapacityUnits      int64  `json:"ReadCapacityUnits"`
	WriteCapacityUnits     int64  `json:"WriteCapacityUnits"`
	LastIncreaseDateTime   *int64 `json:"LastIncreaseDateTime,omitempty"`
	LastDecreaseDateTime   *int64 `json:"LastDecreaseDateTime,omitempty"`
	NumberOfDecreasesToday int64  `json:"NumberOfDecreasesToday"`
}

// Projection represents a GSI projection.
type Projection struct {
	ProjectionType   string   `json:"ProjectionType"`
	NonKeyAttributes []string `json:"NonKeyAttributes,omitempty"`
}

// GlobalSecondaryIndex represents a GSI definition in CreateTable requests.
type GlobalSecondaryIndex struct {
	IndexName             string                 `json:"IndexName"`
	KeySchema             []KeySchemaElement     `json:"KeySchema"`
	Projection            Projection             `json:"Projection"`
	ProvisionedThroughput *ProvisionedThroughput `json:"ProvisionedThroughput,omitempty"`
}

// LocalSecondaryIndex represents an LSI definition in CreateTable requests.
type LocalSecondaryIndex struct {
	IndexName  string             `json:"IndexName"`
	KeySchema  []KeySchemaElement `json:"KeySchema"`
	Projection Projection         `json:"Projection"`
}

// LocalSecondaryIndexDescription represents an LSI in DescribeTable responses.
type LocalSecondaryIndexDescription struct {
	IndexName      string             `json:"IndexName"`
	KeySchema      []KeySchemaElement `json:"KeySchema"`
	Projection     Projection         `json:"Projection"`
	IndexArn       string             `json:"IndexArn"`
	ItemCount      int64              `json:"ItemCount"`
	IndexSizeBytes int64              `json:"IndexSizeBytes"`
}

// GlobalSecondaryIndexDescription represents a GSI in DescribeTable responses.
type GlobalSecondaryIndexDescription struct {
	IndexName             string                            `json:"IndexName"`
	KeySchema             []KeySchemaElement                `json:"KeySchema"`
	Projection            Projection                        `json:"Projection"`
	IndexStatus           string                            `json:"IndexStatus"`
	IndexArn              string                            `json:"IndexArn"`
	ItemCount             int64                             `json:"ItemCount"`
	IndexSizeBytes        int64                             `json:"IndexSizeBytes"`
	ProvisionedThroughput *ProvisionedThroughputDescription `json:"ProvisionedThroughput,omitempty"`
}

// Table represents a DynamoDB table.
type Table struct {
	Name                   string
	KeySchema              []KeySchemaElement
	AttributeDefinitions   []AttributeDefinition
	ProvisionedThroughput  *ProvisionedThroughput
	GlobalSecondaryIndexes []GlobalSecondaryIndex
	LocalSecondaryIndexes  []LocalSecondaryIndex
	CreationDateTime       time.Time
	TableStatus            string
	ItemCount              int64
	TableSizeBytes         int64
	TableARN               string
	BillingMode            string
	DeletionProtection     bool
	TTLAttributeName       string
	TTLEnabled             bool
}

// TableDescription represents a table description in responses.
type TableDescription struct {
	TableName                 string                            `json:"TableName"`
	TableStatus               string                            `json:"TableStatus"`
	TableARN                  string                            `json:"TableArn"`
	TableID                   string                            `json:"TableId,omitempty"`
	CreationDateTime          float64                           `json:"CreationDateTime"`
	KeySchema                 []KeySchemaElement                `json:"KeySchema"`
	AttributeDefinitions      []AttributeDefinition             `json:"AttributeDefinitions"`
	ProvisionedThroughput     *ProvisionedThroughputDescription `json:"ProvisionedThroughput,omitempty"`
	GlobalSecondaryIndexes    []GlobalSecondaryIndexDescription `json:"GlobalSecondaryIndexes,omitempty"`
	LocalSecondaryIndexes     []LocalSecondaryIndexDescription  `json:"LocalSecondaryIndexes,omitempty"`
	ItemCount                 int64                             `json:"ItemCount"`
	TableSizeBytes            int64                             `json:"TableSizeBytes"`
	BillingModeSummary        *BillingModeSummary               `json:"BillingModeSummary,omitempty"`
	DeletionProtectionEnabled bool                              `json:"DeletionProtectionEnabled"`
}

// BillingModeSummary represents billing mode summary.
type BillingModeSummary struct {
	BillingMode                       string `json:"BillingMode"`
	LastUpdateToPayPerRequestDateTime *int64 `json:"LastUpdateToPayPerRequestDateTime,omitempty"`
}

// Item represents a DynamoDB item.
type Item map[string]AttributeValue

// JSON Request/Response Types for AWS JSON 1.0 Protocol.

// CreateTableRequest is the request for CreateTable.
type CreateTableRequest struct {
	TableName                 string                 `json:"TableName"`
	KeySchema                 []KeySchemaElement     `json:"KeySchema"`
	AttributeDefinitions      []AttributeDefinition  `json:"AttributeDefinitions"`
	ProvisionedThroughput     *ProvisionedThroughput `json:"ProvisionedThroughput,omitempty"`
	GlobalSecondaryIndexes    []GlobalSecondaryIndex `json:"GlobalSecondaryIndexes,omitempty"`
	LocalSecondaryIndexes     []LocalSecondaryIndex  `json:"LocalSecondaryIndexes,omitempty"`
	BillingMode               string                 `json:"BillingMode,omitempty"`
	DeletionProtectionEnabled bool                   `json:"DeletionProtectionEnabled,omitempty"`
}

// CreateTableResponse is the response for CreateTable.
type CreateTableResponse struct {
	TableDescription TableDescription `json:"TableDescription"`
}

// DeleteTableRequest is the request for DeleteTable.
type DeleteTableRequest struct {
	TableName string `json:"TableName"`
}

// DeleteTableResponse is the response for DeleteTable.
type DeleteTableResponse struct {
	TableDescription TableDescription `json:"TableDescription"`
}

// ListTablesRequest is the request for ListTables.
type ListTablesRequest struct {
	ExclusiveStartTableName string `json:"ExclusiveStartTableName,omitempty"`
	Limit                   int    `json:"Limit,omitempty"`
}

// ListTablesResponse is the response for ListTables.
type ListTablesResponse struct {
	TableNames             []string `json:"TableNames"`
	LastEvaluatedTableName string   `json:"LastEvaluatedTableName,omitempty"`
}

// DescribeTableRequest is the request for DescribeTable.
type DescribeTableRequest struct {
	TableName string `json:"TableName"`
}

// DescribeTableResponse is the response for DescribeTable.
type DescribeTableResponse struct {
	Table TableDescription `json:"Table"`
}

// PutItemRequest is the request for PutItem.
type PutItemRequest struct {
	TableName                 string                    `json:"TableName"`
	Item                      Item                      `json:"Item"`
	ConditionExpression       string                    `json:"ConditionExpression,omitempty"`
	ExpressionAttributeNames  map[string]string         `json:"ExpressionAttributeNames,omitempty"`
	ExpressionAttributeValues map[string]AttributeValue `json:"ExpressionAttributeValues,omitempty"`
	ReturnValues              string                    `json:"ReturnValues,omitempty"`
}

// PutItemResponse is the response for PutItem.
type PutItemResponse struct {
	Attributes Item `json:"Attributes,omitempty"`
}

// GetItemRequest is the request for GetItem.
type GetItemRequest struct {
	TableName                string            `json:"TableName"`
	Key                      Item              `json:"Key"`
	ProjectionExpression     string            `json:"ProjectionExpression,omitempty"`
	ExpressionAttributeNames map[string]string `json:"ExpressionAttributeNames,omitempty"`
	ConsistentRead           bool              `json:"ConsistentRead,omitempty"`
}

// GetItemResponse is the response for GetItem.
type GetItemResponse struct {
	Item Item `json:"Item,omitempty"`
}

// DeleteItemRequest is the request for DeleteItem.
type DeleteItemRequest struct {
	TableName                 string                    `json:"TableName"`
	Key                       Item                      `json:"Key"`
	ConditionExpression       string                    `json:"ConditionExpression,omitempty"`
	ExpressionAttributeNames  map[string]string         `json:"ExpressionAttributeNames,omitempty"`
	ExpressionAttributeValues map[string]AttributeValue `json:"ExpressionAttributeValues,omitempty"`
	ReturnValues              string                    `json:"ReturnValues,omitempty"`
}

// DeleteItemResponse is the response for DeleteItem.
type DeleteItemResponse struct {
	Attributes Item `json:"Attributes,omitempty"`
}

// UpdateItemRequest is the request for UpdateItem.
type UpdateItemRequest struct {
	TableName                 string                    `json:"TableName"`
	Key                       Item                      `json:"Key"`
	UpdateExpression          string                    `json:"UpdateExpression,omitempty"`
	ConditionExpression       string                    `json:"ConditionExpression,omitempty"`
	ExpressionAttributeNames  map[string]string         `json:"ExpressionAttributeNames,omitempty"`
	ExpressionAttributeValues map[string]AttributeValue `json:"ExpressionAttributeValues,omitempty"`
	ReturnValues              string                    `json:"ReturnValues,omitempty"`
}

// UpdateItemResponse is the response for UpdateItem.
type UpdateItemResponse struct {
	Attributes Item `json:"Attributes,omitempty"`
}

// QueryRequest is the request for Query.
type QueryRequest struct {
	TableName                 string                    `json:"TableName"`
	IndexName                 string                    `json:"IndexName,omitempty"`
	KeyConditionExpression    string                    `json:"KeyConditionExpression,omitempty"`
	FilterExpression          string                    `json:"FilterExpression,omitempty"`
	ProjectionExpression      string                    `json:"ProjectionExpression,omitempty"`
	ExpressionAttributeNames  map[string]string         `json:"ExpressionAttributeNames,omitempty"`
	ExpressionAttributeValues map[string]AttributeValue `json:"ExpressionAttributeValues,omitempty"`
	Limit                     int                       `json:"Limit,omitempty"`
	ExclusiveStartKey         Item                      `json:"ExclusiveStartKey,omitempty"`
	ScanIndexForward          *bool                     `json:"ScanIndexForward,omitempty"`
	ConsistentRead            bool                      `json:"ConsistentRead,omitempty"`
	Select                    string                    `json:"Select,omitempty"`
}

// QueryResponse is the response for Query.
type QueryResponse struct {
	Items            []Item `json:"Items"`
	Count            int    `json:"Count"`
	ScannedCount     int    `json:"ScannedCount"`
	LastEvaluatedKey Item   `json:"LastEvaluatedKey,omitempty"`
}

// ScanRequest is the request for Scan.
type ScanRequest struct {
	TableName                 string                    `json:"TableName"`
	FilterExpression          string                    `json:"FilterExpression,omitempty"`
	ProjectionExpression      string                    `json:"ProjectionExpression,omitempty"`
	ExpressionAttributeNames  map[string]string         `json:"ExpressionAttributeNames,omitempty"`
	ExpressionAttributeValues map[string]AttributeValue `json:"ExpressionAttributeValues,omitempty"`
	Limit                     int                       `json:"Limit,omitempty"`
	ExclusiveStartKey         Item                      `json:"ExclusiveStartKey,omitempty"`
	ConsistentRead            bool                      `json:"ConsistentRead,omitempty"`
	Select                    string                    `json:"Select,omitempty"`
	Segment                   *int                      `json:"Segment,omitempty"`
	TotalSegments             *int                      `json:"TotalSegments,omitempty"`
}

// ScanResponse is the response for Scan.
type ScanResponse struct {
	Items            []Item `json:"Items"`
	Count            int    `json:"Count"`
	ScannedCount     int    `json:"ScannedCount"`
	LastEvaluatedKey Item   `json:"LastEvaluatedKey,omitempty"`
}

// TimeToLiveSpecification represents a TTL specification for UpdateTimeToLive.
type TimeToLiveSpecification struct {
	AttributeName string `json:"AttributeName"`
	Enabled       bool   `json:"Enabled"`
}

// TimeToLiveDescription represents a TTL description in responses.
type TimeToLiveDescription struct {
	AttributeName    string `json:"AttributeName,omitempty"`
	TimeToLiveStatus string `json:"TimeToLiveStatus"`
}

// UpdateTimeToLiveRequest is the request for UpdateTimeToLive.
type UpdateTimeToLiveRequest struct {
	TableName               string                  `json:"TableName"`
	TimeToLiveSpecification TimeToLiveSpecification `json:"TimeToLiveSpecification"`
}

// UpdateTimeToLiveResponse is the response for UpdateTimeToLive.
type UpdateTimeToLiveResponse struct {
	TimeToLiveSpecification TimeToLiveSpecification `json:"TimeToLiveSpecification"`
}

// DescribeTimeToLiveRequest is the request for DescribeTimeToLive.
type DescribeTimeToLiveRequest struct {
	TableName string `json:"TableName"`
}

// DescribeTimeToLiveResponse is the response for DescribeTimeToLive.
type DescribeTimeToLiveResponse struct {
	TimeToLiveDescription TimeToLiveDescription `json:"TimeToLiveDescription"`
}

// TransactWriteItemsRequest is the request for TransactWriteItems.
type TransactWriteItemsRequest struct {
	TransactItems               []TransactWriteItem `json:"TransactItems"`
	ClientRequestToken          string              `json:"ClientRequestToken,omitempty"`
	ReturnConsumedCapacity      string              `json:"ReturnConsumedCapacity,omitempty"`
	ReturnItemCollectionMetrics string              `json:"ReturnItemCollectionMetrics,omitempty"`
}

// TransactWriteItem represents a single item in a TransactWriteItems request.
type TransactWriteItem struct {
	ConditionCheck *TransactConditionCheck `json:"ConditionCheck,omitempty"`
	Delete         *TransactDelete         `json:"Delete,omitempty"`
	Put            *TransactPut            `json:"Put,omitempty"`
	Update         *TransactUpdate         `json:"Update,omitempty"`
}

// TransactConditionCheck represents a ConditionCheck action in a transaction.
type TransactConditionCheck struct {
	TableName                 string                    `json:"TableName"`
	Key                       Item                      `json:"Key"`
	ConditionExpression       string                    `json:"ConditionExpression"`
	ExpressionAttributeNames  map[string]string         `json:"ExpressionAttributeNames,omitempty"`
	ExpressionAttributeValues map[string]AttributeValue `json:"ExpressionAttributeValues,omitempty"`
}

// TransactDelete represents a Delete action in a transaction.
type TransactDelete struct {
	TableName                 string                    `json:"TableName"`
	Key                       Item                      `json:"Key"`
	ConditionExpression       string                    `json:"ConditionExpression,omitempty"`
	ExpressionAttributeNames  map[string]string         `json:"ExpressionAttributeNames,omitempty"`
	ExpressionAttributeValues map[string]AttributeValue `json:"ExpressionAttributeValues,omitempty"`
}

// TransactPut represents a Put action in a transaction.
type TransactPut struct {
	TableName                 string                    `json:"TableName"`
	Item                      Item                      `json:"Item"`
	ConditionExpression       string                    `json:"ConditionExpression,omitempty"`
	ExpressionAttributeNames  map[string]string         `json:"ExpressionAttributeNames,omitempty"`
	ExpressionAttributeValues map[string]AttributeValue `json:"ExpressionAttributeValues,omitempty"`
}

// TransactUpdate represents an Update action in a transaction.
type TransactUpdate struct {
	TableName                 string                    `json:"TableName"`
	Key                       Item                      `json:"Key"`
	UpdateExpression          string                    `json:"UpdateExpression"`
	ConditionExpression       string                    `json:"ConditionExpression,omitempty"`
	ExpressionAttributeNames  map[string]string         `json:"ExpressionAttributeNames,omitempty"`
	ExpressionAttributeValues map[string]AttributeValue `json:"ExpressionAttributeValues,omitempty"`
}

// TransactWriteItemsResponse is the response for TransactWriteItems.
type TransactWriteItemsResponse struct {
	// Empty on success - DynamoDB returns minimal response.
}

// TransactGetItemsRequest is the request for TransactGetItems.
type TransactGetItemsRequest struct {
	TransactItems          []TransactGetItem `json:"TransactItems"`
	ReturnConsumedCapacity string            `json:"ReturnConsumedCapacity,omitempty"`
}

// TransactGetItem represents a single item in a TransactGetItems request.
type TransactGetItem struct {
	Get *TransactGet `json:"Get"`
}

// TransactGet represents a Get action in a transaction.
type TransactGet struct {
	TableName                string            `json:"TableName"`
	Key                      Item              `json:"Key"`
	ProjectionExpression     string            `json:"ProjectionExpression,omitempty"`
	ExpressionAttributeNames map[string]string `json:"ExpressionAttributeNames,omitempty"`
}

// TransactGetItemsResponse is the response for TransactGetItems.
type TransactGetItemsResponse struct {
	Responses []TransactGetItemResponse `json:"Responses"`
}

// TransactGetItemResponse wraps a single item in the TransactGetItems response.
type TransactGetItemResponse struct {
	Item Item `json:"Item,omitempty"`
}

// CancellationReason represents a reason for transaction cancellation.
type CancellationReason struct {
	Code    string `json:"Code"`
	Message string `json:"Message,omitempty"`
}

// TransactionCanceledResponse represents the error response for canceled transactions.
type TransactionCanceledResponse struct {
	Type                string               `json:"__type"`
	Message             string               `json:"message"`
	CancellationReasons []CancellationReason `json:"CancellationReasons"`
}

// BatchWriteItemRequest is the request for BatchWriteItem.
type BatchWriteItemRequest struct {
	RequestItems map[string][]WriteRequest `json:"RequestItems"`
}

// WriteRequest represents a single write request in a batch.
type WriteRequest struct {
	PutRequest    *BatchPutRequest    `json:"PutRequest,omitempty"`
	DeleteRequest *BatchDeleteRequest `json:"DeleteRequest,omitempty"`
}

// BatchPutRequest represents a put request in a batch write.
type BatchPutRequest struct {
	Item Item `json:"Item"`
}

// BatchDeleteRequest represents a delete request in a batch write.
type BatchDeleteRequest struct {
	Key Item `json:"Key"`
}

// BatchWriteItemResponse is the response for BatchWriteItem.
type BatchWriteItemResponse struct {
	UnprocessedItems map[string][]WriteRequest `json:"UnprocessedItems,omitempty"`
}

// BatchGetItemRequest is the request for BatchGetItem.
type BatchGetItemRequest struct {
	RequestItems map[string]KeysAndAttributes `json:"RequestItems"`
}

// KeysAndAttributes represents keys and optional projection for batch get.
type KeysAndAttributes struct {
	Keys                     []Item            `json:"Keys"`
	ProjectionExpression     string            `json:"ProjectionExpression,omitempty"`
	ExpressionAttributeNames map[string]string `json:"ExpressionAttributeNames,omitempty"`
	ConsistentRead           bool              `json:"ConsistentRead,omitempty"`
}

// BatchGetItemResponse is the response for BatchGetItem.
type BatchGetItemResponse struct {
	Responses       map[string][]Item            `json:"Responses,omitempty"`
	UnprocessedKeys map[string]KeysAndAttributes `json:"UnprocessedKeys,omitempty"`
}

// ErrorResponse represents a DynamoDB error response in JSON format.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}

// TableError represents a DynamoDB table error.
type TableError struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *TableError) Error() string {
	return e.Message
}

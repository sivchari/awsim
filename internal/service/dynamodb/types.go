// Package dynamodb provides DynamoDB service emulation for awsim.
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

// Table represents a DynamoDB table.
type Table struct {
	Name                  string
	KeySchema             []KeySchemaElement
	AttributeDefinitions  []AttributeDefinition
	ProvisionedThroughput *ProvisionedThroughput
	CreationDateTime      time.Time
	TableStatus           string
	ItemCount             int64
	TableSizeBytes        int64
	TableARN              string
	BillingMode           string
	DeletionProtection    bool
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

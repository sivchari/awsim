package dynamodb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// CreateTable handles the CreateTable action.
func (s *Service) CreateTable(w http.ResponseWriter, r *http.Request) {
	var req CreateTableRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeDynamoDBError(w, "SerializationException", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TableName == "" {
		writeDynamoDBError(w, "ValidationException", "TableName is required", http.StatusBadRequest)

		return
	}

	if len(req.KeySchema) == 0 {
		writeDynamoDBError(w, "ValidationException", "KeySchema is required", http.StatusBadRequest)

		return
	}

	table, err := s.storage.CreateTable(r.Context(), &req)
	if err != nil {
		var tErr *TableError
		if errors.As(err, &tErr) {
			writeDynamoDBError(w, tErr.Code, tErr.Message, http.StatusBadRequest)

			return
		}

		writeDynamoDBError(w, "InternalServerError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, CreateTableResponse{
		TableDescription: tableToDescription(table),
	})
}

// DeleteTable handles the DeleteTable action.
func (s *Service) DeleteTable(w http.ResponseWriter, r *http.Request) {
	var req DeleteTableRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeDynamoDBError(w, "SerializationException", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TableName == "" {
		writeDynamoDBError(w, "ValidationException", "TableName is required", http.StatusBadRequest)

		return
	}

	table, err := s.storage.DeleteTable(r.Context(), req.TableName)
	if err != nil {
		var tErr *TableError
		if errors.As(err, &tErr) {
			writeDynamoDBError(w, tErr.Code, tErr.Message, http.StatusBadRequest)

			return
		}

		writeDynamoDBError(w, "InternalServerError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, DeleteTableResponse{
		TableDescription: tableToDescription(table),
	})
}

// ListTables handles the ListTables action.
func (s *Service) ListTables(w http.ResponseWriter, r *http.Request) {
	var req ListTablesRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeDynamoDBError(w, "SerializationException", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	names, lastEvaluated, err := s.storage.ListTables(r.Context(), req.ExclusiveStartTableName, req.Limit)
	if err != nil {
		writeDynamoDBError(w, "InternalServerError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, ListTablesResponse{
		TableNames:             names,
		LastEvaluatedTableName: lastEvaluated,
	})
}

// DescribeTable handles the DescribeTable action.
func (s *Service) DescribeTable(w http.ResponseWriter, r *http.Request) {
	var req DescribeTableRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeDynamoDBError(w, "SerializationException", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TableName == "" {
		writeDynamoDBError(w, "ValidationException", "TableName is required", http.StatusBadRequest)

		return
	}

	table, err := s.storage.DescribeTable(r.Context(), req.TableName)
	if err != nil {
		var tErr *TableError
		if errors.As(err, &tErr) {
			writeDynamoDBError(w, tErr.Code, tErr.Message, http.StatusBadRequest)

			return
		}

		writeDynamoDBError(w, "InternalServerError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, DescribeTableResponse{
		Table: tableToDescription(table),
	})
}

// PutItem handles the PutItem action.
func (s *Service) PutItem(w http.ResponseWriter, r *http.Request) {
	var req PutItemRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeDynamoDBError(w, "SerializationException", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TableName == "" {
		writeDynamoDBError(w, "ValidationException", "TableName is required", http.StatusBadRequest)

		return
	}

	if len(req.Item) == 0 {
		writeDynamoDBError(w, "ValidationException", "Item is required", http.StatusBadRequest)

		return
	}

	returnOld := req.ReturnValues == "ReturnValuesAllOld"

	oldItem, err := s.storage.PutItem(r.Context(), req.TableName, req.Item, returnOld)
	if err != nil {
		var tErr *TableError
		if errors.As(err, &tErr) {
			writeDynamoDBError(w, tErr.Code, tErr.Message, http.StatusBadRequest)

			return
		}

		writeDynamoDBError(w, "InternalServerError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, PutItemResponse{
		Attributes: oldItem,
	})
}

// GetItem handles the GetItem action.
func (s *Service) GetItem(w http.ResponseWriter, r *http.Request) {
	var req GetItemRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeDynamoDBError(w, "SerializationException", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TableName == "" {
		writeDynamoDBError(w, "ValidationException", "TableName is required", http.StatusBadRequest)

		return
	}

	if len(req.Key) == 0 {
		writeDynamoDBError(w, "ValidationException", "Key is required", http.StatusBadRequest)

		return
	}

	item, err := s.storage.GetItem(r.Context(), req.TableName, req.Key)
	if err != nil {
		var tErr *TableError
		if errors.As(err, &tErr) {
			writeDynamoDBError(w, tErr.Code, tErr.Message, http.StatusBadRequest)

			return
		}

		writeDynamoDBError(w, "InternalServerError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, GetItemResponse{
		Item: item,
	})
}

// DeleteItem handles the DeleteItem action.
func (s *Service) DeleteItem(w http.ResponseWriter, r *http.Request) {
	var req DeleteItemRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeDynamoDBError(w, "SerializationException", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TableName == "" {
		writeDynamoDBError(w, "ValidationException", "TableName is required", http.StatusBadRequest)

		return
	}

	if len(req.Key) == 0 {
		writeDynamoDBError(w, "ValidationException", "Key is required", http.StatusBadRequest)

		return
	}

	returnOld := req.ReturnValues == "ReturnValuesAllOld"

	oldItem, err := s.storage.DeleteItem(r.Context(), req.TableName, req.Key, returnOld)
	if err != nil {
		var tErr *TableError
		if errors.As(err, &tErr) {
			writeDynamoDBError(w, tErr.Code, tErr.Message, http.StatusBadRequest)

			return
		}

		writeDynamoDBError(w, "InternalServerError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, DeleteItemResponse{
		Attributes: oldItem,
	})
}

// UpdateItem handles the UpdateItem action.
func (s *Service) UpdateItem(w http.ResponseWriter, r *http.Request) {
	var req UpdateItemRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeDynamoDBError(w, "SerializationException", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TableName == "" {
		writeDynamoDBError(w, "ValidationException", "TableName is required", http.StatusBadRequest)

		return
	}

	if len(req.Key) == 0 {
		writeDynamoDBError(w, "ValidationException", "Key is required", http.StatusBadRequest)

		return
	}

	result, err := s.storage.UpdateItem(
		r.Context(),
		req.TableName,
		req.Key,
		req.UpdateExpression,
		req.ExpressionAttributeNames,
		req.ExpressionAttributeValues,
		req.ReturnValues,
	)
	if err != nil {
		var tErr *TableError
		if errors.As(err, &tErr) {
			writeDynamoDBError(w, tErr.Code, tErr.Message, http.StatusBadRequest)

			return
		}

		writeDynamoDBError(w, "InternalServerError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, UpdateItemResponse{
		Attributes: result,
	})
}

// Query handles the Query action.
func (s *Service) Query(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeDynamoDBError(w, "SerializationException", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TableName == "" {
		writeDynamoDBError(w, "ValidationException", "TableName is required", http.StatusBadRequest)

		return
	}

	scanForward := true
	if req.ScanIndexForward != nil {
		scanForward = *req.ScanIndexForward
	}

	items, lastKey, scannedCount, err := s.storage.Query(
		r.Context(),
		req.TableName,
		req.KeyConditionExpression,
		req.FilterExpression,
		req.ExpressionAttributeNames,
		req.ExpressionAttributeValues,
		req.Limit,
		req.ExclusiveStartKey,
		scanForward,
	)
	if err != nil {
		var tErr *TableError
		if errors.As(err, &tErr) {
			writeDynamoDBError(w, tErr.Code, tErr.Message, http.StatusBadRequest)

			return
		}

		writeDynamoDBError(w, "InternalServerError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, QueryResponse{
		Items:            items,
		Count:            len(items),
		ScannedCount:     scannedCount,
		LastEvaluatedKey: lastKey,
	})
}

// Scan handles the Scan action.
func (s *Service) Scan(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeDynamoDBError(w, "SerializationException", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TableName == "" {
		writeDynamoDBError(w, "ValidationException", "TableName is required", http.StatusBadRequest)

		return
	}

	items, lastKey, scannedCount, err := s.storage.Scan(
		r.Context(),
		req.TableName,
		req.FilterExpression,
		req.ExpressionAttributeNames,
		req.ExpressionAttributeValues,
		req.Limit,
		req.ExclusiveStartKey,
	)
	if err != nil {
		var tErr *TableError
		if errors.As(err, &tErr) {
			writeDynamoDBError(w, tErr.Code, tErr.Message, http.StatusBadRequest)

			return
		}

		writeDynamoDBError(w, "InternalServerError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, ScanResponse{
		Items:            items,
		Count:            len(items),
		ScannedCount:     scannedCount,
		LastEvaluatedKey: lastKey,
	})
}

// tableToDescription converts a Table to TableDescription.
func tableToDescription(table *Table) TableDescription {
	desc := TableDescription{
		TableName:                 table.Name,
		TableStatus:               table.TableStatus,
		TableARN:                  table.TableARN,
		TableID:                   uuid.New().String(),
		CreationDateTime:          float64(table.CreationDateTime.Unix()),
		KeySchema:                 table.KeySchema,
		AttributeDefinitions:      table.AttributeDefinitions,
		ItemCount:                 table.ItemCount,
		TableSizeBytes:            table.TableSizeBytes,
		DeletionProtectionEnabled: table.DeletionProtection,
	}

	if table.ProvisionedThroughput != nil {
		desc.ProvisionedThroughput = &ProvisionedThroughputDescription{
			ReadCapacityUnits:  table.ProvisionedThroughput.ReadCapacityUnits,
			WriteCapacityUnits: table.ProvisionedThroughput.WriteCapacityUnits,
		}
	}

	if table.BillingMode != "" {
		desc.BillingModeSummary = &BillingModeSummary{
			BillingMode: table.BillingMode,
		}
	}

	return desc
}

// readJSONRequest reads and decodes JSON request body.
func readJSONRequest(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	if len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// writeJSONResponse writes a JSON response with HTTP 200 OK.
func writeJSONResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

// writeDynamoDBError writes a DynamoDB error response in JSON format.
func writeDynamoDBError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Type:    code,
		Message: message,
	})
}

// dispatchAction routes the request to the appropriate handler based on X-Amz-Target header.
func (s *Service) dispatchAction(w http.ResponseWriter, r *http.Request) {
	target := r.Header.Get("X-Amz-Target")
	action := strings.TrimPrefix(target, "DynamoDB_20120810.")

	switch action {
	case "CreateTable":
		s.CreateTable(w, r)
	case "DeleteTable":
		s.DeleteTable(w, r)
	case "ListTables":
		s.ListTables(w, r)
	case "DescribeTable":
		s.DescribeTable(w, r)
	case "PutItem":
		s.PutItem(w, r)
	case "GetItem":
		s.GetItem(w, r)
	case "DeleteItem":
		s.DeleteItem(w, r)
	case "UpdateItem":
		s.UpdateItem(w, r)
	case "Query":
		s.Query(w, r)
	case "Scan":
		s.Scan(w, r)
	default:
		writeDynamoDBError(w, "UnknownOperationException", "The action "+action+" is not valid", http.StatusBadRequest)
	}
}

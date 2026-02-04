package sqs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// CreateQueue handles the CreateQueue action.
func (s *Service) CreateQueue(w http.ResponseWriter, r *http.Request) {
	var req CreateQueueRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.QueueName == "" {
		writeSQSError(w, "MissingParameter", "QueueName is required", http.StatusBadRequest)

		return
	}

	queue, err := s.storage.CreateQueue(r.Context(), req.QueueName, req.Attributes)
	if err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, CreateQueueResponse{
		QueueURL: queue.URL,
	})
}

// DeleteQueue handles the DeleteQueue action.
func (s *Service) DeleteQueue(w http.ResponseWriter, r *http.Request) {
	var req DeleteQueueRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.QueueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteQueue(r.Context(), req.QueueURL); err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, struct{}{})
}

// ListQueues handles the ListQueues action.
func (s *Service) ListQueues(w http.ResponseWriter, r *http.Request) {
	var req ListQueuesRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	urls, err := s.storage.ListQueues(r.Context(), req.QueueNamePrefix)
	if err != nil {
		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, ListQueuesResponse{
		QueueUrls: urls,
	})
}

// GetQueueURL handles the GetQueueURL action.
func (s *Service) GetQueueURL(w http.ResponseWriter, r *http.Request) {
	var req GetQueueURLRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.QueueName == "" {
		writeSQSError(w, "MissingParameter", "QueueName is required", http.StatusBadRequest)

		return
	}

	url, err := s.storage.GetQueueURL(r.Context(), req.QueueName)
	if err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, GetQueueURLResponse{
		QueueURL: url,
	})
}

// SendMessage handles the SendMessage action.
func (s *Service) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req SendMessageRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.QueueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	if req.MessageBody == "" {
		writeSQSError(w, "MissingParameter", "MessageBody is required", http.StatusBadRequest)

		return
	}

	msg, err := s.storage.SendMessage(r.Context(), req.QueueURL, req.MessageBody, req.DelaySeconds, req.MessageAttributes)
	if err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, SendMessageResponse{
		MessageID:        msg.MessageID,
		MD5OfMessageBody: msg.MD5OfBody,
	})
}

// ReceiveMessage handles the ReceiveMessage action.
func (s *Service) ReceiveMessage(w http.ResponseWriter, r *http.Request) {
	var req ReceiveMessageRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.QueueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	maxMessages := req.MaxNumberOfMessages
	if maxMessages < 1 {
		maxMessages = 1
	}

	if maxMessages > 10 {
		maxMessages = 10
	}

	messages, err := s.storage.ReceiveMessage(r.Context(), req.QueueURL, maxMessages, req.VisibilityTimeout, req.WaitTimeSeconds)
	if err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, ReceiveMessageResponse{
		Messages: convertMessagesToResponse(messages),
	})
}

// convertMessagesToResponse converts Message slice to MessageResponse slice.
func convertMessagesToResponse(messages []*Message) []MessageResponse {
	result := make([]MessageResponse, len(messages))

	for i, msg := range messages {
		result[i] = MessageResponse{
			MessageID:         msg.MessageID,
			ReceiptHandle:     msg.ReceiptHandle,
			MD5OfBody:         msg.MD5OfBody,
			Body:              msg.Body,
			Attributes:        msg.Attributes,
			MessageAttributes: msg.MessageAttributes,
		}
	}

	return result
}

// DeleteMessage handles the DeleteMessage action.
func (s *Service) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	var req DeleteMessageRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.QueueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	if req.ReceiptHandle == "" {
		writeSQSError(w, "MissingParameter", "ReceiptHandle is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteMessage(r.Context(), req.QueueURL, req.ReceiptHandle); err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, struct{}{})
}

// PurgeQueue handles the PurgeQueue action.
func (s *Service) PurgeQueue(w http.ResponseWriter, r *http.Request) {
	var req PurgeQueueRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.QueueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.PurgeQueue(r.Context(), req.QueueURL); err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, struct{}{})
}

// GetQueueAttributes handles the GetQueueAttributes action.
func (s *Service) GetQueueAttributes(w http.ResponseWriter, r *http.Request) {
	var req GetQueueAttributesRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.QueueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	attrs, err := s.storage.GetQueueAttributes(r.Context(), req.QueueURL, req.AttributeNames)
	if err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, GetQueueAttributesResponse{
		Attributes: attrs,
	})
}

// SetQueueAttributes handles the SetQueueAttributes action.
func (s *Service) SetQueueAttributes(w http.ResponseWriter, r *http.Request) {
	var req SetQueueAttributesRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.QueueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.SetQueueAttributes(r.Context(), req.QueueURL, req.Attributes); err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, struct{}{})
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

// writeSQSError writes an SQS error response in JSON format.
func writeSQSError(w http.ResponseWriter, code, message string, status int) {
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
	action := strings.TrimPrefix(target, "AmazonSQS.")

	switch action {
	case "CreateQueue":
		s.CreateQueue(w, r)
	case "DeleteQueue":
		s.DeleteQueue(w, r)
	case "ListQueues":
		s.ListQueues(w, r)
	case "GetQueueUrl":
		s.GetQueueURL(w, r)
	case "SendMessage":
		s.SendMessage(w, r)
	case "ReceiveMessage":
		s.ReceiveMessage(w, r)
	case "DeleteMessage":
		s.DeleteMessage(w, r)
	case "PurgeQueue":
		s.PurgeQueue(w, r)
	case "GetQueueAttributes":
		s.GetQueueAttributes(w, r)
	case "SetQueueAttributes":
		s.SetQueueAttributes(w, r)
	default:
		writeSQSError(w, "InvalidAction", "The action "+action+" is not valid", http.StatusBadRequest)
	}
}

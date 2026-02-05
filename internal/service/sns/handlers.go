package sns

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Error codes for SNS.
const (
	errNotFound             = "NotFound"
	errInvalidParameter     = "InvalidParameter"
	errInternalServiceError = "InternalError"
	errInvalidAction        = "InvalidAction"
)

// CreateTopic handles the CreateTopic action.
func (s *Service) CreateTopic(w http.ResponseWriter, r *http.Request) {
	var req CreateTopicRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeTopicError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.Name == "" {
		writeTopicError(w, errInvalidParameter, "Topic name is required", http.StatusBadRequest)

		return
	}

	topic, err := s.storage.CreateTopic(r.Context(), req.Name, req.Attributes)
	if err != nil {
		var sErr *TopicError
		if errors.As(err, &sErr) {
			writeTopicError(w, sErr.Code, sErr.Message, http.StatusBadRequest)

			return
		}

		writeTopicError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, CreateTopicResponse{
		TopicARN: topic.ARN,
	})
}

// DeleteTopic handles the DeleteTopic action.
func (s *Service) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	var req DeleteTopicRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeTopicError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TopicARN == "" {
		writeTopicError(w, errInvalidParameter, "TopicArn is required", http.StatusBadRequest)

		return
	}

	err := s.storage.DeleteTopic(r.Context(), req.TopicARN)
	if err != nil {
		var sErr *TopicError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errNotFound {
				status = http.StatusNotFound
			}

			writeTopicError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeTopicError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	// DeleteTopic returns empty response on success.
	writeJSONResponse(w, struct{}{})
}

// ListTopics handles the ListTopics action.
func (s *Service) ListTopics(w http.ResponseWriter, r *http.Request) {
	var req ListTopicsRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeTopicError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	topics, nextToken, err := s.storage.ListTopics(r.Context(), req.NextToken)
	if err != nil {
		writeTopicError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	topicEntries := make([]TopicEntry, 0, len(topics))
	for _, topic := range topics {
		topicEntries = append(topicEntries, TopicEntry{
			TopicARN: topic.ARN,
		})
	}

	writeJSONResponse(w, ListTopicsResponse{
		Topics:    topicEntries,
		NextToken: nextToken,
	})
}

// Subscribe handles the Subscribe action.
func (s *Service) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req SubscribeRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeTopicError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TopicARN == "" {
		writeTopicError(w, errInvalidParameter, "TopicArn is required", http.StatusBadRequest)

		return
	}

	if req.Protocol == "" {
		writeTopicError(w, errInvalidParameter, "Protocol is required", http.StatusBadRequest)

		return
	}

	subscription, err := s.storage.Subscribe(r.Context(), req.TopicARN, req.Protocol, req.Endpoint, req.Attributes)
	if err != nil {
		var sErr *TopicError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errNotFound {
				status = http.StatusNotFound
			}

			writeTopicError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeTopicError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, SubscribeResponse{
		SubscriptionARN: subscription.ARN,
	})
}

// Unsubscribe handles the Unsubscribe action.
func (s *Service) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	var req UnsubscribeRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeTopicError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.SubscriptionARN == "" {
		writeTopicError(w, errInvalidParameter, "SubscriptionArn is required", http.StatusBadRequest)

		return
	}

	err := s.storage.Unsubscribe(r.Context(), req.SubscriptionARN)
	if err != nil {
		var sErr *TopicError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errNotFound {
				status = http.StatusNotFound
			}

			writeTopicError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeTopicError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	// Unsubscribe returns empty response on success.
	writeJSONResponse(w, struct{}{})
}

// Publish handles the Publish action.
func (s *Service) Publish(w http.ResponseWriter, r *http.Request) {
	var req PublishRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeTopicError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	topicARN := req.TopicARN
	if topicARN == "" {
		topicARN = req.TargetARN
	}

	if topicARN == "" {
		writeTopicError(w, errInvalidParameter, "TopicArn or TargetArn is required", http.StatusBadRequest)

		return
	}

	if req.Message == "" {
		writeTopicError(w, errInvalidParameter, "Message is required", http.StatusBadRequest)

		return
	}

	messageID, err := s.storage.Publish(r.Context(), topicARN, req.Message, req.Subject, req.MessageAttributes)
	if err != nil {
		var sErr *TopicError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errNotFound {
				status = http.StatusNotFound
			}

			writeTopicError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeTopicError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, PublishResponse{
		MessageID: messageID,
	})
}

// ListSubscriptions handles the ListSubscriptions action.
func (s *Service) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	var req ListSubscriptionsRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeTopicError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	subscriptions, nextToken, err := s.storage.ListSubscriptions(r.Context(), req.NextToken)
	if err != nil {
		writeTopicError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	entries := convertSubscriptionsToEntries(subscriptions)

	writeJSONResponse(w, ListSubscriptionsResponse{
		Subscriptions: entries,
		NextToken:     nextToken,
	})
}

// ListSubscriptionsByTopic handles the ListSubscriptionsByTopic action.
func (s *Service) ListSubscriptionsByTopic(w http.ResponseWriter, r *http.Request) {
	var req ListSubscriptionsByTopicRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeTopicError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.TopicARN == "" {
		writeTopicError(w, errInvalidParameter, "TopicArn is required", http.StatusBadRequest)

		return
	}

	subscriptions, nextToken, err := s.storage.ListSubscriptionsByTopic(r.Context(), req.TopicARN, req.NextToken)
	if err != nil {
		var sErr *TopicError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errNotFound {
				status = http.StatusNotFound
			}

			writeTopicError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeTopicError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	entries := convertSubscriptionsToEntries(subscriptions)

	writeJSONResponse(w, ListSubscriptionsByTopicResponse{
		Subscriptions: entries,
		NextToken:     nextToken,
	})
}

// convertSubscriptionsToEntries converts subscriptions to list entries.
func convertSubscriptionsToEntries(subscriptions []*Subscription) []SubscriptionEntry {
	entries := make([]SubscriptionEntry, 0, len(subscriptions))

	for _, sub := range subscriptions {
		entries = append(entries, SubscriptionEntry{
			SubscriptionARN: sub.ARN,
			Owner:           sub.Owner,
			Protocol:        sub.Protocol,
			Endpoint:        sub.Endpoint,
			TopicARN:        sub.TopicARN,
		})
	}

	return entries
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

// writeTopicError writes an SNS error response in JSON format.
func writeTopicError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Type:    code,
		Message: message,
	})
}

// DispatchAction routes the request to the appropriate handler based on X-Amz-Target header.
// This method implements the JSONProtocolService interface.
func (s *Service) DispatchAction(w http.ResponseWriter, r *http.Request) {
	target := r.Header.Get("X-Amz-Target")
	action := strings.TrimPrefix(target, "AmazonSimpleNotificationService.")

	switch action {
	case "CreateTopic":
		s.CreateTopic(w, r)
	case "DeleteTopic":
		s.DeleteTopic(w, r)
	case "ListTopics":
		s.ListTopics(w, r)
	case "Subscribe":
		s.Subscribe(w, r)
	case "Unsubscribe":
		s.Unsubscribe(w, r)
	case "Publish":
		s.Publish(w, r)
	case "ListSubscriptions":
		s.ListSubscriptions(w, r)
	case "ListSubscriptionsByTopic":
		s.ListSubscriptionsByTopic(w, r)
	default:
		writeTopicError(w, errInvalidAction, "The action "+action+" is not valid", http.StatusBadRequest)
	}
}

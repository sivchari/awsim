// Package sns provides SNS service emulation for awsim.
package sns

import (
	"time"
)

// Topic represents an SNS topic.
type Topic struct {
	ARN           string
	Name          string
	DisplayName   string
	CreatedTime   time.Time
	Attributes    map[string]string
	Subscriptions map[string]*Subscription
}

// Subscription represents an SNS subscription.
type Subscription struct {
	ARN                          string
	TopicARN                     string
	Protocol                     string
	Endpoint                     string
	Owner                        string
	ConfirmationWasAuthenticated bool
	SubscriptionAttributes       map[string]string
}

// CreateTopicRequest is the request for CreateTopic.
type CreateTopicRequest struct {
	Name       string            `json:"Name"`
	Attributes map[string]string `json:"Attributes,omitempty"`
	Tags       []Tag             `json:"Tags,omitempty"`
}

// CreateTopicResponse is the response for CreateTopic.
type CreateTopicResponse struct {
	TopicARN string `json:"TopicArn"`
}

// DeleteTopicRequest is the request for DeleteTopic.
type DeleteTopicRequest struct {
	TopicARN string `json:"TopicArn"`
}

// ListTopicsRequest is the request for ListTopics.
type ListTopicsRequest struct {
	NextToken string `json:"NextToken,omitempty"`
}

// ListTopicsResponse is the response for ListTopics.
type ListTopicsResponse struct {
	Topics    []TopicEntry `json:"Topics"`
	NextToken string       `json:"NextToken,omitempty"`
}

// TopicEntry represents a topic in list response.
type TopicEntry struct {
	TopicARN string `json:"TopicArn"`
}

// SubscribeRequest is the request for Subscribe.
type SubscribeRequest struct {
	TopicARN              string            `json:"TopicArn"`
	Protocol              string            `json:"Protocol"`
	Endpoint              string            `json:"Endpoint,omitempty"`
	Attributes            map[string]string `json:"Attributes,omitempty"`
	ReturnSubscriptionArn bool              `json:"ReturnSubscriptionArn,omitempty"`
}

// SubscribeResponse is the response for Subscribe.
type SubscribeResponse struct {
	SubscriptionARN string `json:"SubscriptionArn"`
}

// UnsubscribeRequest is the request for Unsubscribe.
type UnsubscribeRequest struct {
	SubscriptionARN string `json:"SubscriptionArn"`
}

// PublishRequest is the request for Publish.
type PublishRequest struct {
	TopicARN               string                      `json:"TopicArn,omitempty"`
	TargetARN              string                      `json:"TargetArn,omitempty"`
	Message                string                      `json:"Message"`
	Subject                string                      `json:"Subject,omitempty"`
	MessageStructure       string                      `json:"MessageStructure,omitempty"`
	MessageAttributes      map[string]MessageAttribute `json:"MessageAttributes,omitempty"`
	MessageDeduplicationID string                      `json:"MessageDeduplicationId,omitempty"`
	MessageGroupID         string                      `json:"MessageGroupId,omitempty"`
}

// PublishResponse is the response for Publish.
type PublishResponse struct {
	MessageID      string `json:"MessageId"`
	SequenceNumber string `json:"SequenceNumber,omitempty"`
}

// ListSubscriptionsRequest is the request for ListSubscriptions.
type ListSubscriptionsRequest struct {
	NextToken string `json:"NextToken,omitempty"`
}

// ListSubscriptionsResponse is the response for ListSubscriptions.
type ListSubscriptionsResponse struct {
	Subscriptions []SubscriptionEntry `json:"Subscriptions"`
	NextToken     string              `json:"NextToken,omitempty"`
}

// ListSubscriptionsByTopicRequest is the request for ListSubscriptionsByTopic.
type ListSubscriptionsByTopicRequest struct {
	TopicARN  string `json:"TopicArn"`
	NextToken string `json:"NextToken,omitempty"`
}

// ListSubscriptionsByTopicResponse is the response for ListSubscriptionsByTopic.
type ListSubscriptionsByTopicResponse struct {
	Subscriptions []SubscriptionEntry `json:"Subscriptions"`
	NextToken     string              `json:"NextToken,omitempty"`
}

// SubscriptionEntry represents a subscription in list response.
type SubscriptionEntry struct {
	SubscriptionARN string `json:"SubscriptionArn"`
	Owner           string `json:"Owner"`
	Protocol        string `json:"Protocol"`
	Endpoint        string `json:"Endpoint"`
	TopicARN        string `json:"TopicArn"`
}

// MessageAttribute represents an SNS message attribute.
type MessageAttribute struct {
	DataType    string `json:"DataType"`
	StringValue string `json:"StringValue,omitempty"`
	BinaryValue []byte `json:"BinaryValue,omitempty"`
}

// Tag represents a tag.
type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// ErrorResponse represents an SNS error response.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"Message"`
}

// TopicError represents an SNS error.
type TopicError struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *TopicError) Error() string {
	return e.Message
}

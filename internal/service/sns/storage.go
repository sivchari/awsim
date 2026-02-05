package sns

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

// SQSPublisher is an interface for publishing messages to SQS.
type SQSPublisher interface {
	PublishToSQS(ctx context.Context, queueURL, messageBody string, attributes map[string]string) error
}

// Storage defines the SNS storage interface.
type Storage interface {
	CreateTopic(ctx context.Context, name string, attributes map[string]string) (*Topic, error)
	DeleteTopic(ctx context.Context, topicARN string) error
	ListTopics(ctx context.Context, nextToken string) ([]*Topic, string, error)
	Subscribe(ctx context.Context, topicARN, protocol, endpoint string, attributes map[string]string) (*Subscription, error)
	Unsubscribe(ctx context.Context, subscriptionARN string) error
	Publish(ctx context.Context, topicARN, message, subject string, attributes map[string]MessageAttribute) (string, error)
	ListSubscriptions(ctx context.Context, nextToken string) ([]*Subscription, string, error)
	ListSubscriptionsByTopic(ctx context.Context, topicARN, nextToken string) ([]*Subscription, string, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu            sync.RWMutex
	topics        map[string]*Topic        // keyed by ARN
	subscriptions map[string]*Subscription // keyed by ARN
	baseURL       string
	sqsPublisher  SQSPublisher
}

// NewMemoryStorage creates a new in-memory SNS storage.
func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		topics:        make(map[string]*Topic),
		subscriptions: make(map[string]*Subscription),
		baseURL:       baseURL,
	}
}

// SetSQSPublisher sets the SQS publisher for SNS to SQS integration.
func (m *MemoryStorage) SetSQSPublisher(publisher SQSPublisher) {
	m.sqsPublisher = publisher
}

// CreateTopic creates a new topic.
func (m *MemoryStorage) CreateTopic(_ context.Context, name string, attributes map[string]string) (*Topic, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	arn := m.buildTopicARN(name)

	// Return existing topic if it exists.
	if topic, exists := m.topics[arn]; exists {
		return topic, nil
	}

	topic := &Topic{
		ARN:           arn,
		Name:          name,
		CreatedTime:   time.Now(),
		Attributes:    attributes,
		Subscriptions: make(map[string]*Subscription),
	}

	if attributes != nil {
		if displayName, ok := attributes["DisplayName"]; ok {
			topic.DisplayName = displayName
		}
	}

	m.topics[arn] = topic

	return topic, nil
}

// DeleteTopic deletes a topic.
func (m *MemoryStorage) DeleteTopic(_ context.Context, topicARN string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	topic, exists := m.topics[topicARN]
	if !exists {
		return &TopicError{
			Code:    "NotFound",
			Message: fmt.Sprintf("Topic does not exist: %s", topicARN),
		}
	}

	// Delete all subscriptions for this topic.
	for subARN := range topic.Subscriptions {
		delete(m.subscriptions, subARN)
	}

	delete(m.topics, topicARN)

	return nil
}

// ListTopics returns all topics.
func (m *MemoryStorage) ListTopics(_ context.Context, nextToken string) ([]*Topic, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all topics.
	allTopics := make([]*Topic, 0, len(m.topics))
	for _, topic := range m.topics {
		allTopics = append(allTopics, topic)
	}

	// Sort by ARN for consistent ordering.
	sort.Slice(allTopics, func(i, j int) bool {
		return allTopics[i].ARN < allTopics[j].ARN
	})

	// Handle pagination.
	startIdx := 0
	maxResults := 100

	if nextToken != "" {
		for i, t := range allTopics {
			if t.ARN == nextToken {
				startIdx = i

				break
			}
		}
	}

	endIdx := min(startIdx+maxResults, len(allTopics))
	result := allTopics[startIdx:endIdx]

	var newNextToken string
	if endIdx < len(allTopics) {
		newNextToken = allTopics[endIdx].ARN
	}

	return result, newNextToken, nil
}

// Subscribe creates a subscription.
func (m *MemoryStorage) Subscribe(_ context.Context, topicARN, protocol, endpoint string, attributes map[string]string) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	topic, exists := m.topics[topicARN]
	if !exists {
		return nil, &TopicError{
			Code:    "NotFound",
			Message: fmt.Sprintf("Topic does not exist: %s", topicARN),
		}
	}

	// Validate protocol.
	validProtocols := map[string]bool{
		"http": true, "https": true, "email": true, "email-json": true,
		"sms": true, "sqs": true, "application": true, "lambda": true,
		"firehose": true,
	}

	if !validProtocols[protocol] {
		return nil, &TopicError{
			Code:    "InvalidParameter",
			Message: fmt.Sprintf("Invalid parameter: Protocol %s is not supported", protocol),
		}
	}

	subscriptionARN := m.buildSubscriptionARN(topicARN)

	subscription := &Subscription{
		ARN:                    subscriptionARN,
		TopicARN:               topicARN,
		Protocol:               protocol,
		Endpoint:               endpoint,
		Owner:                  defaultAccountID,
		SubscriptionAttributes: attributes,
	}

	// For SQS and Lambda protocols, auto-confirm.
	if protocol == "sqs" || protocol == "lambda" {
		subscription.ConfirmationWasAuthenticated = true
	}

	m.subscriptions[subscriptionARN] = subscription
	topic.Subscriptions[subscriptionARN] = subscription

	return subscription, nil
}

// Unsubscribe removes a subscription.
func (m *MemoryStorage) Unsubscribe(_ context.Context, subscriptionARN string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	subscription, exists := m.subscriptions[subscriptionARN]
	if !exists {
		return &TopicError{
			Code:    "NotFound",
			Message: fmt.Sprintf("Subscription does not exist: %s", subscriptionARN),
		}
	}

	// Remove from topic's subscriptions.
	if topic, exists := m.topics[subscription.TopicARN]; exists {
		delete(topic.Subscriptions, subscriptionARN)
	}

	delete(m.subscriptions, subscriptionARN)

	return nil
}

// Publish publishes a message to a topic.
func (m *MemoryStorage) Publish(ctx context.Context, topicARN, message, subject string, attributes map[string]MessageAttribute) (string, error) {
	m.mu.RLock()

	topic, exists := m.topics[topicARN]
	if !exists {
		m.mu.RUnlock()

		return "", &TopicError{
			Code:    "NotFound",
			Message: fmt.Sprintf("Topic does not exist: %s", topicARN),
		}
	}

	// Copy subscriptions while holding read lock.
	subscriptions := make([]*Subscription, 0, len(topic.Subscriptions))
	for _, sub := range topic.Subscriptions {
		subscriptions = append(subscriptions, sub)
	}
	m.mu.RUnlock()

	messageID := uuid.New().String()

	// Deliver to all subscriptions.
	for _, sub := range subscriptions {
		if err := m.deliverMessage(ctx, sub, message, subject, messageID, attributes); err != nil {
			// Log error but continue delivering to other subscriptions.
			continue
		}
	}

	return messageID, nil
}

// deliverMessage delivers a message to a subscription.
func (m *MemoryStorage) deliverMessage(ctx context.Context, sub *Subscription, message, subject, messageID string, _ map[string]MessageAttribute) error {
	switch sub.Protocol {
	case "sqs":
		if m.sqsPublisher != nil {
			attrs := map[string]string{
				"MessageId": messageID,
			}
			if subject != "" {
				attrs["Subject"] = subject
			}

			if err := m.sqsPublisher.PublishToSQS(ctx, sub.Endpoint, message, attrs); err != nil {
				return fmt.Errorf("failed to publish to SQS: %w", err)
			}

			return nil
		}
	case "http", "https":
		// HTTP delivery not implemented in emulator.
		return nil
	default:
		// Other protocols not implemented.
		return nil
	}

	return nil
}

// ListSubscriptions returns all subscriptions.
func (m *MemoryStorage) ListSubscriptions(_ context.Context, nextToken string) ([]*Subscription, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all subscriptions.
	allSubs := make([]*Subscription, 0, len(m.subscriptions))
	for _, sub := range m.subscriptions {
		allSubs = append(allSubs, sub)
	}

	// Sort by ARN for consistent ordering.
	sort.Slice(allSubs, func(i, j int) bool {
		return allSubs[i].ARN < allSubs[j].ARN
	})

	// Handle pagination.
	startIdx := 0
	maxResults := 100

	if nextToken != "" {
		for i, s := range allSubs {
			if s.ARN == nextToken {
				startIdx = i

				break
			}
		}
	}

	endIdx := min(startIdx+maxResults, len(allSubs))
	result := allSubs[startIdx:endIdx]

	var newNextToken string
	if endIdx < len(allSubs) {
		newNextToken = allSubs[endIdx].ARN
	}

	return result, newNextToken, nil
}

// ListSubscriptionsByTopic returns subscriptions for a specific topic.
func (m *MemoryStorage) ListSubscriptionsByTopic(_ context.Context, topicARN, nextToken string) ([]*Subscription, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	topic, exists := m.topics[topicARN]
	if !exists {
		return nil, "", &TopicError{
			Code:    "NotFound",
			Message: fmt.Sprintf("Topic does not exist: %s", topicARN),
		}
	}

	// Collect subscriptions for this topic.
	allSubs := make([]*Subscription, 0, len(topic.Subscriptions))
	for _, sub := range topic.Subscriptions {
		allSubs = append(allSubs, sub)
	}

	// Sort by ARN for consistent ordering.
	sort.Slice(allSubs, func(i, j int) bool {
		return allSubs[i].ARN < allSubs[j].ARN
	})

	// Handle pagination.
	startIdx := 0
	maxResults := 100

	if nextToken != "" {
		for i, s := range allSubs {
			if s.ARN == nextToken {
				startIdx = i

				break
			}
		}
	}

	endIdx := min(startIdx+maxResults, len(allSubs))
	result := allSubs[startIdx:endIdx]

	var newNextToken string
	if endIdx < len(allSubs) {
		newNextToken = allSubs[endIdx].ARN
	}

	return result, newNextToken, nil
}

// buildTopicARN builds an ARN for a topic.
func (m *MemoryStorage) buildTopicARN(name string) string {
	return fmt.Sprintf("arn:aws:sns:%s:%s:%s", defaultRegion, defaultAccountID, name)
}

// buildSubscriptionARN builds an ARN for a subscription.
func (m *MemoryStorage) buildSubscriptionARN(topicARN string) string {
	// Extract topic name from ARN.
	parts := strings.Split(topicARN, ":")
	topicName := parts[len(parts)-1]

	return fmt.Sprintf("arn:aws:sns:%s:%s:%s:%s",
		defaultRegion, defaultAccountID, topicName, uuid.New().String())
}

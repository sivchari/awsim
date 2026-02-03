// Package sqs provides SQS service emulation for awsim.
package sqs

import (
	"encoding/xml"
	"time"
)

// Queue represents an SQS queue.
type Queue struct {
	Name                   string
	URL                    string
	ARN                    string
	CreatedTimestamp       time.Time
	LastModifiedTimestamp  time.Time
	VisibilityTimeout      int
	MessageRetentionPeriod int
	DelaySeconds           int
	MaxMessageSize         int
	ReceiveWaitTimeSeconds int
}

// Message represents an SQS message.
type Message struct {
	MessageID         string
	ReceiptHandle     string
	Body              string
	MD5OfBody         string
	Attributes        map[string]string
	MessageAttributes map[string]MessageAttributeValue
	SentTimestamp     time.Time
	VisibleAt         time.Time
	ReceiveCount      int
}

// MessageAttributeValue represents a message attribute.
type MessageAttributeValue struct {
	DataType    string
	StringValue string
	BinaryValue []byte
}

// XML Response Types

// CreateQueueResult is the response for CreateQueue.
type CreateQueueResult struct {
	XMLName  xml.Name `xml:"CreateQueueResult"`
	QueueURL string   `xml:"QueueUrl"`
}

// CreateQueueResponse wraps CreateQueueResult.
type CreateQueueResponse struct {
	XMLName           xml.Name          `xml:"CreateQueueResponse"`
	Xmlns             string            `xml:"xmlns,attr"`
	CreateQueueResult CreateQueueResult `xml:"CreateQueueResult"`
	ResponseMetadata  ResponseMetadata  `xml:"ResponseMetadata"`
}

// DeleteQueueResponse is the response for DeleteQueue.
type DeleteQueueResponse struct {
	XMLName          xml.Name         `xml:"DeleteQueueResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	ResponseMetadata ResponseMetadata `xml:"ResponseMetadata"`
}

// ListQueuesResult is the result for ListQueues.
type ListQueuesResult struct {
	XMLName  xml.Name `xml:"ListQueuesResult"`
	QueueURL []string `xml:"QueueUrl"`
}

// ListQueuesResponse wraps ListQueuesResult.
type ListQueuesResponse struct {
	XMLName          xml.Name         `xml:"ListQueuesResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	ListQueuesResult ListQueuesResult `xml:"ListQueuesResult"`
	ResponseMetadata ResponseMetadata `xml:"ResponseMetadata"`
}

// GetQueueURLResult is the result for GetQueueUrl.
type GetQueueURLResult struct {
	XMLName  xml.Name `xml:"GetQueueURLResult"`
	QueueURL string   `xml:"QueueUrl"`
}

// GetQueueURLResponse wraps GetQueueURLResult.
type GetQueueURLResponse struct {
	XMLName           xml.Name          `xml:"GetQueueURLResponse"`
	Xmlns             string            `xml:"xmlns,attr"`
	GetQueueURLResult GetQueueURLResult `xml:"GetQueueURLResult"`
	ResponseMetadata  ResponseMetadata  `xml:"ResponseMetadata"`
}

// SendMessageResult is the result for SendMessage.
type SendMessageResult struct {
	XMLName                xml.Name `xml:"SendMessageResult"`
	MessageID              string   `xml:"MessageId"`
	MD5OfMessageBody       string   `xml:"MD5OfMessageBody"`
	MD5OfMessageAttributes string   `xml:"MD5OfMessageAttributes,omitempty"`
}

// SendMessageResponse wraps SendMessageResult.
type SendMessageResponse struct {
	XMLName           xml.Name          `xml:"SendMessageResponse"`
	Xmlns             string            `xml:"xmlns,attr"`
	SendMessageResult SendMessageResult `xml:"SendMessageResult"`
	ResponseMetadata  ResponseMetadata  `xml:"ResponseMetadata"`
}

// ReceiveMessageResult is the result for ReceiveMessage.
type ReceiveMessageResult struct {
	XMLName xml.Name      `xml:"ReceiveMessageResult"`
	Message []MessageInfo `xml:"Message"`
}

// MessageInfo represents message information in XML response.
type MessageInfo struct {
	MessageID        string                 `xml:"MessageId"`
	ReceiptHandle    string                 `xml:"ReceiptHandle"`
	MD5OfBody        string                 `xml:"MD5OfBody"`
	Body             string                 `xml:"Body"`
	Attribute        []AttributeInfo        `xml:"Attribute,omitempty"`
	MessageAttribute []MessageAttributeInfo `xml:"MessageAttribute,omitempty"`
}

// AttributeInfo represents an attribute in XML response.
type AttributeInfo struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
}

// MessageAttributeInfo represents a message attribute in XML response.
type MessageAttributeInfo struct {
	Name  string                    `xml:"Name"`
	Value MessageAttributeValueInfo `xml:"Value"`
}

// MessageAttributeValueInfo represents a message attribute value in XML response.
type MessageAttributeValueInfo struct {
	DataType    string `xml:"DataType"`
	StringValue string `xml:"StringValue,omitempty"`
	BinaryValue []byte `xml:"BinaryValue,omitempty"`
}

// ReceiveMessageResponse wraps ReceiveMessageResult.
type ReceiveMessageResponse struct {
	XMLName              xml.Name             `xml:"ReceiveMessageResponse"`
	Xmlns                string               `xml:"xmlns,attr"`
	ReceiveMessageResult ReceiveMessageResult `xml:"ReceiveMessageResult"`
	ResponseMetadata     ResponseMetadata     `xml:"ResponseMetadata"`
}

// DeleteMessageResponse is the response for DeleteMessage.
type DeleteMessageResponse struct {
	XMLName          xml.Name         `xml:"DeleteMessageResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	ResponseMetadata ResponseMetadata `xml:"ResponseMetadata"`
}

// PurgeQueueResponse is the response for PurgeQueue.
type PurgeQueueResponse struct {
	XMLName          xml.Name         `xml:"PurgeQueueResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	ResponseMetadata ResponseMetadata `xml:"ResponseMetadata"`
}

// GetQueueAttributesResult is the result for GetQueueAttributes.
type GetQueueAttributesResult struct {
	XMLName   xml.Name        `xml:"GetQueueAttributesResult"`
	Attribute []AttributeInfo `xml:"Attribute"`
}

// GetQueueAttributesResponse wraps GetQueueAttributesResult.
type GetQueueAttributesResponse struct {
	XMLName                  xml.Name                 `xml:"GetQueueAttributesResponse"`
	Xmlns                    string                   `xml:"xmlns,attr"`
	GetQueueAttributesResult GetQueueAttributesResult `xml:"GetQueueAttributesResult"`
	ResponseMetadata         ResponseMetadata         `xml:"ResponseMetadata"`
}

// SetQueueAttributesResponse is the response for SetQueueAttributes.
type SetQueueAttributesResponse struct {
	XMLName          xml.Name         `xml:"SetQueueAttributesResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	ResponseMetadata ResponseMetadata `xml:"ResponseMetadata"`
}

// ResponseMetadata contains the request ID.
type ResponseMetadata struct {
	RequestID string `xml:"RequestId"`
}

// ErrorResponse represents an SQS error response.
type ErrorResponse struct {
	XMLName   xml.Name  `xml:"ErrorResponse"`
	Xmlns     string    `xml:"xmlns,attr"`
	Error     ErrorInfo `xml:"Error"`
	RequestID string    `xml:"RequestId"`
}

// ErrorInfo contains error details.
type ErrorInfo struct {
	Type    string `xml:"Type"`
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

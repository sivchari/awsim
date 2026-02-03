package sqs

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

const sqsNamespace = "http://queue.amazonaws.com/doc/2012-11-05/"

// CreateQueue handles the CreateQueue action.
func (s *Service) CreateQueue(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	queueName := r.FormValue("QueueName")
	if queueName == "" {
		writeSQSError(w, "MissingParameter", "QueueName is required", http.StatusBadRequest)

		return
	}

	attributes := parseQueueAttributes(r)

	queue, err := s.storage.CreateQueue(r.Context(), queueName, attributes)
	if err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := CreateQueueResponse{
		Xmlns: sqsNamespace,
		CreateQueueResult: CreateQueueResult{
			QueueURL: queue.URL,
		},
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// DeleteQueue handles the DeleteQueue action.
func (s *Service) DeleteQueue(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	queueURL := r.FormValue("QueueUrl")
	if queueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteQueue(r.Context(), queueURL); err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := DeleteQueueResponse{
		Xmlns: sqsNamespace,
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// ListQueues handles the ListQueues action.
func (s *Service) ListQueues(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	prefix := r.FormValue("QueueNamePrefix")

	urls, err := s.storage.ListQueues(r.Context(), prefix)
	if err != nil {
		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := ListQueuesResponse{
		Xmlns: sqsNamespace,
		ListQueuesResult: ListQueuesResult{
			QueueURL: urls,
		},
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// GetQueueURL handles the GetQueueURL action.
func (s *Service) GetQueueURL(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	queueName := r.FormValue("QueueName")
	if queueName == "" {
		writeSQSError(w, "MissingParameter", "QueueName is required", http.StatusBadRequest)

		return
	}

	url, err := s.storage.GetQueueURL(r.Context(), queueName)
	if err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := GetQueueURLResponse{
		Xmlns: sqsNamespace,
		GetQueueURLResult: GetQueueURLResult{
			QueueURL: url,
		},
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// SendMessage handles the SendMessage action.
func (s *Service) SendMessage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	queueURL := r.FormValue("QueueUrl")
	if queueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	messageBody := r.FormValue("MessageBody")
	if messageBody == "" {
		writeSQSError(w, "MissingParameter", "MessageBody is required", http.StatusBadRequest)

		return
	}

	var delaySeconds int

	if ds := r.FormValue("DelaySeconds"); ds != "" {
		delaySeconds, _ = strconv.Atoi(ds)
	}

	messageAttributes := parseMessageAttributes(r)

	msg, err := s.storage.SendMessage(r.Context(), queueURL, messageBody, delaySeconds, messageAttributes)
	if err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := SendMessageResponse{
		Xmlns: sqsNamespace,
		SendMessageResult: SendMessageResult{
			MessageID:        msg.MessageID,
			MD5OfMessageBody: msg.MD5OfBody,
		},
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// ReceiveMessage handles the ReceiveMessage action.
func (s *Service) ReceiveMessage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	queueURL := r.FormValue("QueueUrl")
	if queueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	maxMessages := parseMaxMessages(r.FormValue("MaxNumberOfMessages"))
	visibilityTimeout := parseIntFormValue(r.FormValue("VisibilityTimeout"))
	waitTimeSeconds := parseIntFormValue(r.FormValue("WaitTimeSeconds"))

	messages, err := s.storage.ReceiveMessage(r.Context(), queueURL, maxMessages, visibilityTimeout, waitTimeSeconds)
	if err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := ReceiveMessageResponse{
		Xmlns: sqsNamespace,
		ReceiveMessageResult: ReceiveMessageResult{
			Message: convertMessagesToInfos(messages),
		},
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// parseMaxMessages parses the MaxNumberOfMessages form value.
func parseMaxMessages(val string) int {
	if val == "" {
		return 1
	}

	maxMessages, _ := strconv.Atoi(val)

	if maxMessages < 1 {
		return 1
	}

	if maxMessages > 10 {
		return 10
	}

	return maxMessages
}

// parseIntFormValue parses an integer form value, returning 0 if empty or invalid.
func parseIntFormValue(val string) int {
	if val == "" {
		return 0
	}

	n, _ := strconv.Atoi(val)

	return n
}

// convertMessagesToInfos converts Message slice to MessageInfo slice.
func convertMessagesToInfos(messages []*Message) []MessageInfo {
	msgInfos := make([]MessageInfo, len(messages))

	for i, msg := range messages {
		attrs := make([]AttributeInfo, 0, len(msg.Attributes))

		for k, v := range msg.Attributes {
			attrs = append(attrs, AttributeInfo{Name: k, Value: v})
		}

		msgInfos[i] = MessageInfo{
			MessageID:     msg.MessageID,
			ReceiptHandle: msg.ReceiptHandle,
			MD5OfBody:     msg.MD5OfBody,
			Body:          msg.Body,
			Attribute:     attrs,
		}
	}

	return msgInfos
}

// DeleteMessage handles the DeleteMessage action.
func (s *Service) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	queueURL := r.FormValue("QueueUrl")
	if queueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	receiptHandle := r.FormValue("ReceiptHandle")
	if receiptHandle == "" {
		writeSQSError(w, "MissingParameter", "ReceiptHandle is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteMessage(r.Context(), queueURL, receiptHandle); err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := DeleteMessageResponse{
		Xmlns: sqsNamespace,
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// PurgeQueue handles the PurgeQueue action.
func (s *Service) PurgeQueue(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	queueURL := r.FormValue("QueueUrl")
	if queueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.PurgeQueue(r.Context(), queueURL); err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := PurgeQueueResponse{
		Xmlns: sqsNamespace,
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// GetQueueAttributes handles the GetQueueAttributes action.
func (s *Service) GetQueueAttributes(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	queueURL := r.FormValue("QueueUrl")
	if queueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	attributeNames := parseAttributeNames(r)

	attrs, err := s.storage.GetQueueAttributes(r.Context(), queueURL, attributeNames)
	if err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	attrInfos := make([]AttributeInfo, 0, len(attrs))

	for k, v := range attrs {
		attrInfos = append(attrInfos, AttributeInfo{Name: k, Value: v})
	}

	resp := GetQueueAttributesResponse{
		Xmlns: sqsNamespace,
		GetQueueAttributesResult: GetQueueAttributesResult{
			Attribute: attrInfos,
		},
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// SetQueueAttributes handles the SetQueueAttributes action.
func (s *Service) SetQueueAttributes(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	queueURL := r.FormValue("QueueUrl")
	if queueURL == "" {
		writeSQSError(w, "MissingParameter", "QueueUrl is required", http.StatusBadRequest)

		return
	}

	attributes := parseQueueAttributes(r)

	if err := s.storage.SetQueueAttributes(r.Context(), queueURL, attributes); err != nil {
		var qErr *QueueError
		if errors.As(err, &qErr) {
			writeSQSError(w, qErr.Code, qErr.Message, http.StatusBadRequest)

			return
		}

		writeSQSError(w, "InternalError", "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := SetQueueAttributesResponse{
		Xmlns: sqsNamespace,
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// writeXMLResponse writes an XML response.
//
//nolint:unparam // status is kept for API consistency
func writeXMLResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(v)
}

// writeSQSError writes an SQS error response.
func writeSQSError(w http.ResponseWriter, code, message string, status int) {
	resp := ErrorResponse{
		Xmlns: sqsNamespace,
		Error: ErrorInfo{
			Type:    "Sender",
			Code:    code,
			Message: message,
		},
		RequestID: uuid.New().String(),
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(resp)
}

// parseQueueAttributes parses Attribute.N.Name and Attribute.N.Value from form.
func parseQueueAttributes(r *http.Request) map[string]string {
	attrs := make(map[string]string)

	for i := 1; ; i++ {
		name := r.FormValue("Attribute." + strconv.Itoa(i) + ".Name")
		if name == "" {
			break
		}

		value := r.FormValue("Attribute." + strconv.Itoa(i) + ".Value")
		attrs[name] = value
	}

	return attrs
}

// parseAttributeNames parses AttributeName.N from form.
func parseAttributeNames(r *http.Request) []string {
	names := make([]string, 0)

	for i := 1; ; i++ {
		name := r.FormValue("AttributeName." + strconv.Itoa(i))
		if name == "" {
			break
		}

		names = append(names, name)
	}

	return names
}

// parseMessageAttributes parses MessageAttribute.N.Name and MessageAttribute.N.Value from form.
func parseMessageAttributes(r *http.Request) map[string]MessageAttributeValue {
	attrs := make(map[string]MessageAttributeValue)

	for i := 1; ; i++ {
		prefix := "MessageAttribute." + strconv.Itoa(i)
		name := r.FormValue(prefix + ".Name")

		if name == "" {
			break
		}

		dataType := r.FormValue(prefix + ".Value.DataType")
		stringValue := r.FormValue(prefix + ".Value.StringValue")

		attrs[name] = MessageAttributeValue{
			DataType:    dataType,
			StringValue: stringValue,
		}
	}

	return attrs
}

// dispatchAction routes the request to the appropriate handler based on Action.
func (s *Service) dispatchAction(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	action := r.FormValue("Action")

	switch action {
	case "CreateQueue":
		s.CreateQueue(w, r)
	case "DeleteQueue":
		s.DeleteQueue(w, r)
	case "ListQueues":
		s.ListQueues(w, r)
	case "GetQueueURL":
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

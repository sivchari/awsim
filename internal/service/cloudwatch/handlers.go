package cloudwatch

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const cloudwatchXMLNS = "http://monitoring.amazonaws.com/doc/2010-08-01/"

// Error codes for CloudWatch.
const (
	errInvalidParameter     = "InvalidParameterValue"
	errMissingParameter     = "MissingParameter"
	errInternalServiceError = "InternalServiceError"
	errInvalidAction        = "InvalidAction"
	errResourceNotFound     = "ResourceNotFound"
)

// PutMetricData handles the PutMetricData action.
func (s *Service) PutMetricData(w http.ResponseWriter, r *http.Request) {
	var req PutMetricDataRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeCloudWatchError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.Namespace == "" {
		writeCloudWatchError(w, errMissingParameter, "The parameter Namespace is required", http.StatusBadRequest)

		return
	}

	if len(req.MetricData) == 0 {
		writeCloudWatchError(w, errMissingParameter, "The parameter MetricData is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.PutMetricData(r.Context(), req.Namespace, req.MetricData); err != nil {
		handleCloudWatchError(w, err)

		return
	}

	writeXMLResponse(w, XMLPutMetricDataResponse{
		Xmlns: cloudwatchXMLNS,
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	})
}

// GetMetricData handles the GetMetricData action.
func (s *Service) GetMetricData(w http.ResponseWriter, r *http.Request) {
	var req GetMetricDataRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeCloudWatchError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if len(req.MetricDataQueries) == 0 {
		writeCloudWatchError(w, errMissingParameter, "The parameter MetricDataQueries is required", http.StatusBadRequest)

		return
	}

	result, err := s.storage.GetMetricData(r.Context(), &req)
	if err != nil {
		handleCloudWatchError(w, err)

		return
	}

	writeXMLResponse(w, XMLGetMetricDataResponse{
		Xmlns:               cloudwatchXMLNS,
		GetMetricDataResult: *result,
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	})
}

// GetMetricStatistics handles the GetMetricStatistics action.
func (s *Service) GetMetricStatistics(w http.ResponseWriter, r *http.Request) {
	var req GetMetricStatisticsRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeCloudWatchError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.Namespace == "" {
		writeCloudWatchError(w, errMissingParameter, "The parameter Namespace is required", http.StatusBadRequest)

		return
	}

	if req.MetricName == "" {
		writeCloudWatchError(w, errMissingParameter, "The parameter MetricName is required", http.StatusBadRequest)

		return
	}

	result, err := s.storage.GetMetricStatistics(r.Context(), &req)
	if err != nil {
		handleCloudWatchError(w, err)

		return
	}

	writeXMLResponse(w, XMLGetMetricStatisticsResponse{
		Xmlns:                     cloudwatchXMLNS,
		GetMetricStatisticsResult: *result,
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	})
}

// ListMetrics handles the ListMetrics action.
func (s *Service) ListMetrics(w http.ResponseWriter, r *http.Request) {
	var req ListMetricsRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeCloudWatchError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	result, err := s.storage.ListMetrics(r.Context(), &req)
	if err != nil {
		handleCloudWatchError(w, err)

		return
	}

	writeXMLResponse(w, XMLListMetricsResponse{
		Xmlns:             cloudwatchXMLNS,
		ListMetricsResult: *result,
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	})
}

// PutMetricAlarm handles the PutMetricAlarm action.
func (s *Service) PutMetricAlarm(w http.ResponseWriter, r *http.Request) {
	var req PutMetricAlarmRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeCloudWatchError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.AlarmName == "" {
		writeCloudWatchError(w, errMissingParameter, "The parameter AlarmName is required", http.StatusBadRequest)

		return
	}

	if req.MetricName == "" {
		writeCloudWatchError(w, errMissingParameter, "The parameter MetricName is required", http.StatusBadRequest)

		return
	}

	if req.Namespace == "" {
		writeCloudWatchError(w, errMissingParameter, "The parameter Namespace is required", http.StatusBadRequest)

		return
	}

	if req.ComparisonOperator == "" {
		writeCloudWatchError(w, errMissingParameter, "The parameter ComparisonOperator is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.PutMetricAlarm(r.Context(), &req); err != nil {
		handleCloudWatchError(w, err)

		return
	}

	writeXMLResponse(w, XMLPutMetricAlarmResponse{
		Xmlns: cloudwatchXMLNS,
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	})
}

// DeleteAlarms handles the DeleteAlarms action.
func (s *Service) DeleteAlarms(w http.ResponseWriter, r *http.Request) {
	var req DeleteAlarmsRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeCloudWatchError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if len(req.AlarmNames) == 0 {
		writeCloudWatchError(w, errMissingParameter, "The parameter AlarmNames is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteAlarms(r.Context(), req.AlarmNames); err != nil {
		handleCloudWatchError(w, err)

		return
	}

	writeXMLResponse(w, XMLDeleteAlarmsResponse{
		Xmlns: cloudwatchXMLNS,
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	})
}

// DescribeAlarms handles the DescribeAlarms action.
func (s *Service) DescribeAlarms(w http.ResponseWriter, r *http.Request) {
	var req DescribeAlarmsRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeCloudWatchError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	result, err := s.storage.DescribeAlarms(r.Context(), &req)
	if err != nil {
		handleCloudWatchError(w, err)

		return
	}

	writeXMLResponse(w, XMLDescribeAlarmsResponse{
		Xmlns:                cloudwatchXMLNS,
		DescribeAlarmsResult: *result,
		ResponseMetadata: ResponseMetadata{
			RequestID: uuid.New().String(),
		},
	})
}

// DispatchAction routes the request to the appropriate handler based on X-Amz-Target header.
func (s *Service) DispatchAction(w http.ResponseWriter, r *http.Request) {
	target := r.Header.Get("X-Amz-Target")
	action := strings.TrimPrefix(target, "GraniteServiceVersion20100801.")

	switch action {
	case "PutMetricData":
		s.PutMetricData(w, r)
	case "GetMetricData":
		s.GetMetricData(w, r)
	case "GetMetricStatistics":
		s.GetMetricStatistics(w, r)
	case "ListMetrics":
		s.ListMetrics(w, r)
	case "PutMetricAlarm":
		s.PutMetricAlarm(w, r)
	case "DeleteAlarms":
		s.DeleteAlarms(w, r)
	case "DescribeAlarms":
		s.DescribeAlarms(w, r)
	default:
		writeCloudWatchError(w, errInvalidAction, "The action "+action+" is not valid", http.StatusBadRequest)
	}
}

// handleCloudWatchError handles CloudWatch errors.
func handleCloudWatchError(w http.ResponseWriter, err error) {
	var cwErr *Error
	if errors.As(err, &cwErr) {
		status := http.StatusBadRequest
		if cwErr.Code == errResourceNotFound {
			status = http.StatusNotFound
		}

		writeCloudWatchError(w, cwErr.Code, cwErr.Message, status)

		return
	}

	writeCloudWatchError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)
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

// writeXMLResponse writes an XML response with HTTP 200 OK.
func writeXMLResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(v)
}

// writeCloudWatchError writes a CloudWatch error response in XML format.
func writeCloudWatchError(w http.ResponseWriter, code, message string, status int) {
	requestID := uuid.New().String()

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("x-amzn-RequestId", requestID)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(XMLErrorResponse{
		Xmlns: cloudwatchXMLNS,
		Error: XMLErrorDetail{
			Type:    "Sender",
			Code:    code,
			Message: message,
		},
		RequestID: requestID,
	})
}

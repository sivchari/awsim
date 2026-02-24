package servicequotas

import "time"

// ServiceInfo represents an AWS service in Service Quotas.
type ServiceInfo struct {
	ServiceCode string
	ServiceName string
}

// ServiceQuota represents a service quota.
type ServiceQuota struct {
	QuotaARN            string
	QuotaCode           string
	QuotaName           string
	ServiceCode         string
	ServiceName         string
	Value               float64
	Unit                string
	Adjustable          bool
	GlobalQuota         bool
	Description         string
	QuotaAppliedAtLevel string
}

// QuotaChangeRequest represents a quota increase request.
type QuotaChangeRequest struct {
	ID                    string
	ServiceCode           string
	ServiceName           string
	QuotaCode             string
	QuotaName             string
	DesiredValue          float64
	Status                string
	Created               time.Time
	LastUpdated           time.Time
	CaseID                string
	Requester             string
	QuotaARN              string
	Unit                  string
	GlobalQuota           bool
	QuotaRequestedAtLevel string
}

// Error represents a Service Quotas error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

// ListServicesRequest represents the ListServices API request.
type ListServicesRequest struct {
	MaxResults *int32 `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// ListServicesResponse represents the ListServices API response.
type ListServicesResponse struct {
	Services  []ServiceInfoOutput `json:"Services,omitempty"`
	NextToken string              `json:"NextToken,omitempty"`
}

// ServiceInfoOutput represents a service in the response.
type ServiceInfoOutput struct {
	ServiceCode string `json:"ServiceCode,omitempty"`
	ServiceName string `json:"ServiceName,omitempty"`
}

// GetServiceQuotaRequest represents the GetServiceQuota API request.
type GetServiceQuotaRequest struct {
	ServiceCode string `json:"ServiceCode"`
	QuotaCode   string `json:"QuotaCode"`
	ContextID   string `json:"ContextId,omitempty"`
}

// GetServiceQuotaResponse represents the GetServiceQuota API response.
type GetServiceQuotaResponse struct {
	Quota *ServiceQuotaOutput `json:"Quota,omitempty"`
}

// ServiceQuotaOutput represents a quota in the response.
type ServiceQuotaOutput struct {
	QuotaARN            string  `json:"QuotaArn,omitempty"`
	QuotaCode           string  `json:"QuotaCode,omitempty"`
	QuotaName           string  `json:"QuotaName,omitempty"`
	ServiceCode         string  `json:"ServiceCode,omitempty"`
	ServiceName         string  `json:"ServiceName,omitempty"`
	Value               float64 `json:"Value,omitempty"`
	Unit                string  `json:"Unit,omitempty"`
	Adjustable          bool    `json:"Adjustable,omitempty"`
	GlobalQuota         bool    `json:"GlobalQuota,omitempty"`
	Description         string  `json:"Description,omitempty"`
	QuotaAppliedAtLevel string  `json:"QuotaAppliedAtLevel,omitempty"`
}

// ListServiceQuotasRequest represents the ListServiceQuotas API request.
type ListServiceQuotasRequest struct {
	ServiceCode         string `json:"ServiceCode"`
	MaxResults          *int32 `json:"MaxResults,omitempty"`
	NextToken           string `json:"NextToken,omitempty"`
	QuotaCode           string `json:"QuotaCode,omitempty"`
	QuotaAppliedAtLevel string `json:"QuotaAppliedAtLevel,omitempty"`
}

// ListServiceQuotasResponse represents the ListServiceQuotas API response.
type ListServiceQuotasResponse struct {
	Quotas    []ServiceQuotaOutput `json:"Quotas,omitempty"`
	NextToken string               `json:"NextToken,omitempty"`
}

// GetAWSDefaultServiceQuotaRequest represents the GetAWSDefaultServiceQuota API request.
type GetAWSDefaultServiceQuotaRequest struct {
	ServiceCode string `json:"ServiceCode"`
	QuotaCode   string `json:"QuotaCode"`
}

// GetAWSDefaultServiceQuotaResponse represents the GetAWSDefaultServiceQuota API response.
type GetAWSDefaultServiceQuotaResponse struct {
	Quota *ServiceQuotaOutput `json:"Quota,omitempty"`
}

// ListAWSDefaultServiceQuotasRequest represents the ListAWSDefaultServiceQuotas API request.
type ListAWSDefaultServiceQuotasRequest struct {
	ServiceCode string `json:"ServiceCode"`
	MaxResults  *int32 `json:"MaxResults,omitempty"`
	NextToken   string `json:"NextToken,omitempty"`
}

// ListAWSDefaultServiceQuotasResponse represents the ListAWSDefaultServiceQuotas API response.
type ListAWSDefaultServiceQuotasResponse struct {
	Quotas    []ServiceQuotaOutput `json:"Quotas,omitempty"`
	NextToken string               `json:"NextToken,omitempty"`
}

// RequestServiceQuotaIncreaseRequest represents the RequestServiceQuotaIncrease API request.
type RequestServiceQuotaIncreaseRequest struct {
	ServiceCode  string  `json:"ServiceCode"`
	QuotaCode    string  `json:"QuotaCode"`
	DesiredValue float64 `json:"DesiredValue"`
	ContextID    string  `json:"ContextId,omitempty"`
}

// RequestServiceQuotaIncreaseResponse represents the RequestServiceQuotaIncrease API response.
type RequestServiceQuotaIncreaseResponse struct {
	RequestedQuota *RequestedServiceQuotaChangeOutput `json:"RequestedQuota,omitempty"`
}

// RequestedServiceQuotaChangeOutput represents a quota change request in the response.
type RequestedServiceQuotaChangeOutput struct {
	ID                    string  `json:"Id,omitempty"`
	ServiceCode           string  `json:"ServiceCode,omitempty"`
	ServiceName           string  `json:"ServiceName,omitempty"`
	QuotaCode             string  `json:"QuotaCode,omitempty"`
	QuotaName             string  `json:"QuotaName,omitempty"`
	DesiredValue          float64 `json:"DesiredValue,omitempty"`
	Status                string  `json:"Status,omitempty"`
	Created               float64 `json:"Created,omitempty"`
	LastUpdated           float64 `json:"LastUpdated,omitempty"`
	CaseID                string  `json:"CaseId,omitempty"`
	Requester             string  `json:"Requester,omitempty"`
	QuotaARN              string  `json:"QuotaArn,omitempty"`
	Unit                  string  `json:"Unit,omitempty"`
	GlobalQuota           bool    `json:"GlobalQuota,omitempty"`
	QuotaRequestedAtLevel string  `json:"QuotaRequestedAtLevel,omitempty"`
}

// GetRequestedServiceQuotaChangeRequest represents the GetRequestedServiceQuotaChange API request.
type GetRequestedServiceQuotaChangeRequest struct {
	RequestID string `json:"RequestId"`
}

// GetRequestedServiceQuotaChangeResponse represents the GetRequestedServiceQuotaChange API response.
type GetRequestedServiceQuotaChangeResponse struct {
	RequestedQuota *RequestedServiceQuotaChangeOutput `json:"RequestedQuota,omitempty"`
}

// ListRequestedServiceQuotaChangeHistoryRequest represents the ListRequestedServiceQuotaChangeHistory API request.
type ListRequestedServiceQuotaChangeHistoryRequest struct {
	ServiceCode           string `json:"ServiceCode,omitempty"`
	QuotaCode             string `json:"QuotaCode,omitempty"`
	Status                string `json:"Status,omitempty"`
	MaxResults            *int32 `json:"MaxResults,omitempty"`
	NextToken             string `json:"NextToken,omitempty"`
	QuotaRequestedAtLevel string `json:"QuotaRequestedAtLevel,omitempty"`
}

// ListRequestedServiceQuotaChangeHistoryResponse represents the ListRequestedServiceQuotaChangeHistory API response.
type ListRequestedServiceQuotaChangeHistoryResponse struct {
	RequestedQuotas []RequestedServiceQuotaChangeOutput `json:"RequestedQuotas,omitempty"`
	NextToken       string                              `json:"NextToken,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}

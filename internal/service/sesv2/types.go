// Package sesv2 provides SES v2 service emulation for awsim.
package sesv2

import "time"

// EmailIdentity represents an email identity (email address or domain).
type EmailIdentity struct {
	IdentityName             string
	IdentityType             string // EMAIL_ADDRESS or DOMAIN
	VerifiedForSendingStatus bool
	DkimAttributes           *DkimAttributes
	CreatedAt                time.Time
}

// DkimAttributes represents DKIM configuration for an identity.
type DkimAttributes struct {
	SigningEnabled          bool
	Status                  string // PENDING, SUCCESS, FAILED, NOT_STARTED, TEMPORARY_FAILURE
	SigningAttributesOrigin string
	Tokens                  []string
}

// ConfigurationSet represents an SES configuration set.
type ConfigurationSet struct {
	Name              string
	DeliveryOptions   *DeliveryOptions
	ReputationOptions *ReputationOptions
	SendingOptions    *SendingOptions
	TrackingOptions   *TrackingOptions
	Tags              []Tag
}

// DeliveryOptions represents delivery options for a configuration set.
type DeliveryOptions struct {
	TlsPolicy       string
	SendingPoolName string
}

// ReputationOptions represents reputation options for a configuration set.
type ReputationOptions struct {
	ReputationMetricsEnabled bool
	LastFreshStart           *time.Time
}

// SendingOptions represents sending options for a configuration set.
type SendingOptions struct {
	SendingEnabled bool
}

// TrackingOptions represents tracking options for a configuration set.
type TrackingOptions struct {
	CustomRedirectDomain string
}

// SentEmail represents a sent email for debugging purposes.
type SentEmail struct {
	MessageID            string
	FromEmailAddress     string
	Destination          *Destination
	Subject              string
	Body                 string
	ConfigurationSetName string
	SentAt               time.Time
}

// Destination represents email destinations.
type Destination struct {
	ToAddresses  []string `json:"ToAddresses,omitempty"`
	CcAddresses  []string `json:"CcAddresses,omitempty"`
	BccAddresses []string `json:"BccAddresses,omitempty"`
}

// Tag represents a tag.
type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// CreateEmailIdentityRequest is the request for CreateEmailIdentity.
type CreateEmailIdentityRequest struct {
	EmailIdentity         string            `json:"EmailIdentity"`
	ConfigurationSetName  string            `json:"ConfigurationSetName,omitempty"`
	DkimSigningAttributes *DkimSigningAttrs `json:"DkimSigningAttributes,omitempty"`
	Tags                  []Tag             `json:"Tags,omitempty"`
}

// DkimSigningAttrs represents DKIM signing attributes in request.
type DkimSigningAttrs struct {
	DomainSigningPrivateKey string `json:"DomainSigningPrivateKey,omitempty"`
	DomainSigningSelector   string `json:"DomainSigningSelector,omitempty"`
	NextSigningKeyLength    string `json:"NextSigningKeyLength,omitempty"`
}

// CreateEmailIdentityResponse is the response for CreateEmailIdentity.
type CreateEmailIdentityResponse struct {
	IdentityType             string          `json:"IdentityType,omitempty"`
	VerifiedForSendingStatus bool            `json:"VerifiedForSendingStatus"`
	DkimAttributes           *DkimAttributes `json:"DkimAttributes,omitempty"`
}

// GetEmailIdentityResponse is the response for GetEmailIdentity.
type GetEmailIdentityResponse struct {
	IdentityType             string          `json:"IdentityType,omitempty"`
	FeedbackForwardingStatus bool            `json:"FeedbackForwardingStatus"`
	VerifiedForSendingStatus bool            `json:"VerifiedForSendingStatus"`
	DkimAttributes           *DkimAttributes `json:"DkimAttributes,omitempty"`
	ConfigurationSetName     string          `json:"ConfigurationSetName,omitempty"`
	Tags                     []Tag           `json:"Tags,omitempty"`
}

// ListEmailIdentitiesRequest is the request for ListEmailIdentities.
type ListEmailIdentitiesRequest struct {
	NextToken string `json:"NextToken,omitempty"`
	PageSize  int32  `json:"PageSize,omitempty"`
}

// ListEmailIdentitiesResponse is the response for ListEmailIdentities.
type ListEmailIdentitiesResponse struct {
	EmailIdentities []EmailIdentitySummary `json:"EmailIdentities,omitempty"`
	NextToken       string                 `json:"NextToken,omitempty"`
}

// EmailIdentitySummary represents an email identity summary.
type EmailIdentitySummary struct {
	IdentityName   string `json:"IdentityName,omitempty"`
	IdentityType   string `json:"IdentityType,omitempty"`
	SendingEnabled bool   `json:"SendingEnabled"`
}

// CreateConfigurationSetRequest is the request for CreateConfigurationSet.
type CreateConfigurationSetRequest struct {
	ConfigurationSetName string             `json:"ConfigurationSetName"`
	DeliveryOptions      *DeliveryOptions   `json:"DeliveryOptions,omitempty"`
	ReputationOptions    *ReputationOptions `json:"ReputationOptions,omitempty"`
	SendingOptions       *SendingOptions    `json:"SendingOptions,omitempty"`
	TrackingOptions      *TrackingOptions   `json:"TrackingOptions,omitempty"`
	Tags                 []Tag              `json:"Tags,omitempty"`
}

// GetConfigurationSetResponse is the response for GetConfigurationSet.
type GetConfigurationSetResponse struct {
	ConfigurationSetName string             `json:"ConfigurationSetName,omitempty"`
	DeliveryOptions      *DeliveryOptions   `json:"DeliveryOptions,omitempty"`
	ReputationOptions    *ReputationOptions `json:"ReputationOptions,omitempty"`
	SendingOptions       *SendingOptions    `json:"SendingOptions,omitempty"`
	TrackingOptions      *TrackingOptions   `json:"TrackingOptions,omitempty"`
	Tags                 []Tag              `json:"Tags,omitempty"`
}

// ListConfigurationSetsRequest is the request for ListConfigurationSets.
type ListConfigurationSetsRequest struct {
	NextToken string `json:"NextToken,omitempty"`
	PageSize  int32  `json:"PageSize,omitempty"`
}

// ListConfigurationSetsResponse is the response for ListConfigurationSets.
type ListConfigurationSetsResponse struct {
	ConfigurationSets []string `json:"ConfigurationSets,omitempty"`
	NextToken         string   `json:"NextToken,omitempty"`
}

// SendEmailRequest is the request for SendEmail.
type SendEmailRequest struct {
	FromEmailAddress               string        `json:"FromEmailAddress,omitempty"`
	FromEmailAddressIdentityArn    string        `json:"FromEmailAddressIdentityArn,omitempty"`
	Destination                    *Destination  `json:"Destination,omitempty"`
	ReplyToAddresses               []string      `json:"ReplyToAddresses,omitempty"`
	FeedbackForwardingEmailAddress string        `json:"FeedbackForwardingEmailAddress,omitempty"`
	Content                        *EmailContent `json:"Content,omitempty"`
	EmailTags                      []MessageTag  `json:"EmailTags,omitempty"`
	ConfigurationSetName           string        `json:"ConfigurationSetName,omitempty"`
}

// EmailContent represents the content of an email.
type EmailContent struct {
	Simple   *SimpleEmail `json:"Simple,omitempty"`
	Raw      *RawEmail    `json:"Raw,omitempty"`
	Template *Template    `json:"Template,omitempty"`
}

// SimpleEmail represents a simple email.
type SimpleEmail struct {
	Subject *Content `json:"Subject,omitempty"`
	Body    *Body    `json:"Body,omitempty"`
}

// RawEmail represents a raw email.
type RawEmail struct {
	Data []byte `json:"Data,omitempty"`
}

// Template represents a template email.
type Template struct {
	TemplateName string `json:"TemplateName,omitempty"`
	TemplateArn  string `json:"TemplateArn,omitempty"`
	TemplateData string `json:"TemplateData,omitempty"`
}

// Body represents the body of an email.
type Body struct {
	Text *Content `json:"Text,omitempty"`
	Html *Content `json:"Html,omitempty"`
}

// Content represents text content.
type Content struct {
	Data    string `json:"Data,omitempty"`
	Charset string `json:"Charset,omitempty"`
}

// MessageTag represents a message tag.
type MessageTag struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

// SendEmailResponse is the response for SendEmail.
type SendEmailResponse struct {
	MessageId string `json:"MessageId,omitempty"`
}

// ErrorResponse represents an SES error response.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}

// IdentityError represents an SES error.
type IdentityError struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *IdentityError) Error() string {
	return e.Message
}

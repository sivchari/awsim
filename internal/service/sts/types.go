package sts

import "encoding/xml"

// AssumeRoleInput represents an AssumeRole request.
type AssumeRoleInput struct {
	RoleArn         string `json:"RoleArn"`
	RoleSessionName string `json:"RoleSessionName"`
	DurationSeconds int32  `json:"DurationSeconds,omitempty"`
	ExternalID      string `json:"ExternalId,omitempty"`
	Policy          string `json:"Policy,omitempty"`
	SerialNumber    string `json:"SerialNumber,omitempty"`
	TokenCode       string `json:"TokenCode,omitempty"`
}

// AssumeRoleWithSAMLInput represents an AssumeRoleWithSAML request.
type AssumeRoleWithSAMLInput struct {
	PrincipalArn    string `json:"PrincipalArn"`
	RoleArn         string `json:"RoleArn"`
	SAMLAssertion   string `json:"SAMLAssertion"`
	DurationSeconds int32  `json:"DurationSeconds,omitempty"`
	Policy          string `json:"Policy,omitempty"`
}

// AssumeRoleWithWebIdentityInput represents an AssumeRoleWithWebIdentity request.
type AssumeRoleWithWebIdentityInput struct {
	RoleArn          string `json:"RoleArn"`
	RoleSessionName  string `json:"RoleSessionName"`
	WebIdentityToken string `json:"WebIdentityToken"`
	DurationSeconds  int32  `json:"DurationSeconds,omitempty"`
	ProviderID       string `json:"ProviderId,omitempty"`
	Policy           string `json:"Policy,omitempty"`
}

// GetSessionTokenInput represents a GetSessionToken request.
type GetSessionTokenInput struct {
	DurationSeconds int32  `json:"DurationSeconds,omitempty"`
	SerialNumber    string `json:"SerialNumber,omitempty"`
	TokenCode       string `json:"TokenCode,omitempty"`
}

// GetFederationTokenInput represents a GetFederationToken request.
type GetFederationTokenInput struct {
	Name            string `json:"Name"`
	DurationSeconds int32  `json:"DurationSeconds,omitempty"`
	Policy          string `json:"Policy,omitempty"`
}

// Credentials represents temporary security credentials.
type Credentials struct {
	AccessKeyID     string `xml:"AccessKeyId"`
	SecretAccessKey string `xml:"SecretAccessKey"`
	SessionToken    string `xml:"SessionToken"`
	Expiration      string `xml:"Expiration"`
}

// AssumedRoleUser represents the assumed role user.
type AssumedRoleUser struct {
	Arn           string `xml:"Arn"`
	AssumedRoleID string `xml:"AssumedRoleId"`
}

// FederatedUser represents the federated user.
type FederatedUser struct {
	Arn             string `xml:"Arn"`
	FederatedUserID string `xml:"FederatedUserId"`
}

// XML response types.

// XMLAssumeRoleResponse is the XML response for AssumeRole.
type XMLAssumeRoleResponse struct {
	XMLName          xml.Name         `xml:"AssumeRoleResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	AssumedRoleUser  *AssumedRoleUser `xml:"AssumeRoleResult>AssumedRoleUser"`
	Credentials      *Credentials     `xml:"AssumeRoleResult>Credentials"`
	PackedPolicySize int32            `xml:"AssumeRoleResult>PackedPolicySize"`
	RequestID        string           `xml:"ResponseMetadata>RequestId"`
}

// XMLAssumeRoleWithSAMLResponse is the XML response for AssumeRoleWithSAML.
type XMLAssumeRoleWithSAMLResponse struct {
	XMLName          xml.Name         `xml:"AssumeRoleWithSAMLResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	AssumedRoleUser  *AssumedRoleUser `xml:"AssumeRoleWithSAMLResult>AssumedRoleUser"`
	Credentials      *Credentials     `xml:"AssumeRoleWithSAMLResult>Credentials"`
	PackedPolicySize int32            `xml:"AssumeRoleWithSAMLResult>PackedPolicySize"`
	RequestID        string           `xml:"ResponseMetadata>RequestId"`
}

// XMLAssumeRoleWithWebIdentityResponse is the XML response for AssumeRoleWithWebIdentity.
type XMLAssumeRoleWithWebIdentityResponse struct {
	XMLName          xml.Name         `xml:"AssumeRoleWithWebIdentityResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	AssumedRoleUser  *AssumedRoleUser `xml:"AssumeRoleWithWebIdentityResult>AssumedRoleUser"`
	Credentials      *Credentials     `xml:"AssumeRoleWithWebIdentityResult>Credentials"`
	PackedPolicySize int32            `xml:"AssumeRoleWithWebIdentityResult>PackedPolicySize"`
	RequestID        string           `xml:"ResponseMetadata>RequestId"`
}

// XMLGetCallerIdentityResponse is the XML response for GetCallerIdentity.
type XMLGetCallerIdentityResponse struct {
	XMLName   xml.Name `xml:"GetCallerIdentityResponse"`
	Xmlns     string   `xml:"xmlns,attr"`
	Account   string   `xml:"GetCallerIdentityResult>Account"`
	Arn       string   `xml:"GetCallerIdentityResult>Arn"`
	UserID    string   `xml:"GetCallerIdentityResult>UserId"`
	RequestID string   `xml:"ResponseMetadata>RequestId"`
}

// XMLGetSessionTokenResponse is the XML response for GetSessionToken.
type XMLGetSessionTokenResponse struct {
	XMLName     xml.Name     `xml:"GetSessionTokenResponse"`
	Xmlns       string       `xml:"xmlns,attr"`
	Credentials *Credentials `xml:"GetSessionTokenResult>Credentials"`
	RequestID   string       `xml:"ResponseMetadata>RequestId"`
}

// XMLGetFederationTokenResponse is the XML response for GetFederationToken.
type XMLGetFederationTokenResponse struct {
	XMLName          xml.Name       `xml:"GetFederationTokenResponse"`
	Xmlns            string         `xml:"xmlns,attr"`
	Credentials      *Credentials   `xml:"GetFederationTokenResult>Credentials"`
	FederatedUser    *FederatedUser `xml:"GetFederationTokenResult>FederatedUser"`
	PackedPolicySize int32          `xml:"GetFederationTokenResult>PackedPolicySize"`
	RequestID        string         `xml:"ResponseMetadata>RequestId"`
}

// XMLErrorResponse is the XML error response.
type XMLErrorResponse struct {
	XMLName   xml.Name `xml:"ErrorResponse"`
	Error     XMLError `xml:"Error"`
	RequestID string   `xml:"RequestId"`
}

// XMLError is an XML error.
type XMLError struct {
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

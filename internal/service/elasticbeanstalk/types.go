package elasticbeanstalk

import "encoding/xml"

const ebXMLNS = "https://elasticbeanstalk.amazonaws.com/docs/2010-12-01/"

// ApplicationDescription represents an Elastic Beanstalk application.
type ApplicationDescription struct {
	ApplicationName string `json:"ApplicationName"`
	Description     string `json:"Description,omitempty"`
	DateCreated     string `json:"DateCreated,omitempty"`
	DateUpdated     string `json:"DateUpdated,omitempty"`
	ApplicationArn  string `json:"ApplicationArn,omitempty"`
}

// EnvironmentDescription represents an Elastic Beanstalk environment.
type EnvironmentDescription struct {
	ApplicationName   string `json:"ApplicationName"`
	EnvironmentID     string `json:"EnvironmentId,omitempty"`
	EnvironmentName   string `json:"EnvironmentName"`
	Description       string `json:"Description,omitempty"`
	SolutionStackName string `json:"SolutionStackName,omitempty"`
	Status            string `json:"Status,omitempty"`
	Health            string `json:"Health,omitempty"`
	DateCreated       string `json:"DateCreated,omitempty"`
	DateUpdated       string `json:"DateUpdated,omitempty"`
	EnvironmentArn    string `json:"EnvironmentArn,omitempty"`
	CNAME             string `json:"CNAME,omitempty"`
	EndpointURL       string `json:"EndpointURL,omitempty"`
}

// CreateApplicationInput represents a CreateApplication request.
type CreateApplicationInput struct {
	ApplicationName string `json:"ApplicationName"`
	Description     string `json:"Description,omitempty"`
}

// CreateEnvironmentInput represents a CreateEnvironment request.
type CreateEnvironmentInput struct {
	ApplicationName   string `json:"ApplicationName"`
	EnvironmentName   string `json:"EnvironmentName"`
	Description       string `json:"Description,omitempty"`
	SolutionStackName string `json:"SolutionStackName,omitempty"`
}

// DescribeApplicationsInput represents a DescribeApplications request.
type DescribeApplicationsInput struct {
	ApplicationNames []string `json:"ApplicationNames,omitempty"`
}

// DescribeEnvironmentsInput represents a DescribeEnvironments request.
type DescribeEnvironmentsInput struct {
	ApplicationName  string   `json:"ApplicationName,omitempty"`
	EnvironmentNames []string `json:"EnvironmentNames,omitempty"`
	EnvironmentIDs   []string `json:"EnvironmentIds,omitempty"`
}

// UpdateApplicationInput represents an UpdateApplication request.
type UpdateApplicationInput struct {
	ApplicationName string `json:"ApplicationName"`
	Description     string `json:"Description,omitempty"`
}

// DeleteApplicationInput represents a DeleteApplication request.
type DeleteApplicationInput struct {
	ApplicationName string `json:"ApplicationName"`
}

// TerminateEnvironmentInput represents a TerminateEnvironment request.
type TerminateEnvironmentInput struct {
	EnvironmentID   string `json:"EnvironmentId,omitempty"`
	EnvironmentName string `json:"EnvironmentName,omitempty"`
}

// XML response types.

// XMLCreateApplicationResponse wraps the CreateApplication XML response.
type XMLCreateApplicationResponse struct {
	XMLName xml.Name                   `xml:"CreateApplicationResponse"`
	Xmlns   string                     `xml:"xmlns,attr"`
	Result  XMLCreateApplicationResult `xml:"CreateApplicationResult"`
	Meta    XMLResponseMetadata        `xml:"ResponseMetadata"`
}

// XMLCreateApplicationResult contains the application in the result.
type XMLCreateApplicationResult struct {
	Application XMLApplicationDescription `xml:"Application"`
}

// XMLDescribeApplicationsResponse wraps the DescribeApplications XML response.
type XMLDescribeApplicationsResponse struct {
	XMLName xml.Name                      `xml:"DescribeApplicationsResponse"`
	Xmlns   string                        `xml:"xmlns,attr"`
	Result  XMLDescribeApplicationsResult `xml:"DescribeApplicationsResult"`
	Meta    XMLResponseMetadata           `xml:"ResponseMetadata"`
}

// XMLDescribeApplicationsResult contains the list of applications.
type XMLDescribeApplicationsResult struct {
	Applications []XMLApplicationDescription `xml:"Applications>member"`
}

// XMLUpdateApplicationResponse wraps the UpdateApplication XML response.
type XMLUpdateApplicationResponse struct {
	XMLName xml.Name                   `xml:"UpdateApplicationResponse"`
	Xmlns   string                     `xml:"xmlns,attr"`
	Result  XMLUpdateApplicationResult `xml:"UpdateApplicationResult"`
	Meta    XMLResponseMetadata        `xml:"ResponseMetadata"`
}

// XMLUpdateApplicationResult contains the updated application.
type XMLUpdateApplicationResult struct {
	Application XMLApplicationDescription `xml:"Application"`
}

// XMLDeleteApplicationResponse wraps the DeleteApplication XML response.
type XMLDeleteApplicationResponse struct {
	XMLName xml.Name            `xml:"DeleteApplicationResponse"`
	Xmlns   string              `xml:"xmlns,attr"`
	Meta    XMLResponseMetadata `xml:"ResponseMetadata"`
}

// XMLCreateEnvironmentResponse wraps the CreateEnvironment XML response.
type XMLCreateEnvironmentResponse struct {
	XMLName xml.Name                   `xml:"CreateEnvironmentResponse"`
	Xmlns   string                     `xml:"xmlns,attr"`
	Result  XMLCreateEnvironmentResult `xml:"CreateEnvironmentResult"`
	Meta    XMLResponseMetadata        `xml:"ResponseMetadata"`
}

// XMLCreateEnvironmentResult contains the environment.
type XMLCreateEnvironmentResult struct {
	XMLEnvironmentDescription
}

// XMLDescribeEnvironmentsResponse wraps the DescribeEnvironments XML response.
type XMLDescribeEnvironmentsResponse struct {
	XMLName xml.Name                      `xml:"DescribeEnvironmentsResponse"`
	Xmlns   string                        `xml:"xmlns,attr"`
	Result  XMLDescribeEnvironmentsResult `xml:"DescribeEnvironmentsResult"`
	Meta    XMLResponseMetadata           `xml:"ResponseMetadata"`
}

// XMLDescribeEnvironmentsResult contains the list of environments.
type XMLDescribeEnvironmentsResult struct {
	Environments []XMLEnvironmentDescription `xml:"Environments>member"`
}

// XMLTerminateEnvironmentResponse wraps the TerminateEnvironment XML response.
type XMLTerminateEnvironmentResponse struct {
	XMLName xml.Name                      `xml:"TerminateEnvironmentResponse"`
	Xmlns   string                        `xml:"xmlns,attr"`
	Result  XMLTerminateEnvironmentResult `xml:"TerminateEnvironmentResult"`
	Meta    XMLResponseMetadata           `xml:"ResponseMetadata"`
}

// XMLTerminateEnvironmentResult contains the terminated environment.
type XMLTerminateEnvironmentResult struct {
	XMLEnvironmentDescription
}

// XMLApplicationDescription represents an application in XML.
type XMLApplicationDescription struct {
	ApplicationName string `xml:"ApplicationName"`
	Description     string `xml:"Description,omitempty"`
	DateCreated     string `xml:"DateCreated"`
	DateUpdated     string `xml:"DateUpdated"`
	ApplicationArn  string `xml:"ApplicationArn"`
}

// XMLEnvironmentDescription represents an environment in XML.
type XMLEnvironmentDescription struct {
	ApplicationName   string `xml:"ApplicationName"`
	EnvironmentID     string `xml:"EnvironmentId"`
	EnvironmentName   string `xml:"EnvironmentName"`
	Description       string `xml:"Description,omitempty"`
	SolutionStackName string `xml:"SolutionStackName,omitempty"`
	Status            string `xml:"Status"`
	Health            string `xml:"Health"`
	DateCreated       string `xml:"DateCreated"`
	DateUpdated       string `xml:"DateUpdated"`
	EnvironmentArn    string `xml:"EnvironmentArn"`
}

// XMLResponseMetadata holds the request ID.
type XMLResponseMetadata struct {
	RequestID string `xml:"RequestId"`
}

// XMLErrorResponse represents an error response.
type XMLErrorResponse struct {
	XMLName   xml.Name `xml:"ErrorResponse"`
	Error     XMLError `xml:"Error"`
	RequestID string   `xml:"RequestId"`
}

// XMLError represents the error detail.
type XMLError struct {
	Type    string `xml:"Type"`
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

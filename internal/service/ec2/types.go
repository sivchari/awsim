package ec2

import (
	"encoding/xml"
	"time"
)

// Instance represents an EC2 instance.
type Instance struct {
	InstanceID       string
	ImageID          string
	InstanceType     string
	State            InstanceState
	PrivateIPAddress string
	PublicIPAddress  string
	SecurityGroups   []GroupIdentifier
	KeyName          string
	LaunchTime       time.Time
	Tags             []Tag
}

// InstanceState represents the state of an instance.
type InstanceState struct {
	Code int
	Name string
}

// GroupIdentifier represents a security group.
type GroupIdentifier struct {
	GroupID   string
	GroupName string
}

// Tag represents a resource tag.
type Tag struct {
	Key   string
	Value string
}

// SecurityGroup represents an EC2 security group.
type SecurityGroup struct {
	GroupID      string
	GroupName    string
	Description  string
	VpcID        string
	IngressRules []IPPermission
	EgressRules  []IPPermission
	Tags         []Tag
}

// IPPermission represents an ingress or egress rule.
type IPPermission struct {
	IPProtocol string
	FromPort   int
	ToPort     int
	IPRanges   []IPRange
}

// IPRange represents a CIDR IP range.
type IPRange struct {
	CidrIP      string
	Description string
}

// KeyPair represents an EC2 key pair.
type KeyPair struct {
	KeyName        string
	KeyFingerprint string
	KeyPairID      string
	KeyMaterial    string
	CreateTime     time.Time
	Tags           []Tag
}

// RunInstancesRequest represents a RunInstances request.
type RunInstancesRequest struct {
	ImageID          string   `json:"ImageId"`
	InstanceType     string   `json:"InstanceType"`
	MinCount         int      `json:"MinCount"`
	MaxCount         int      `json:"MaxCount"`
	KeyName          string   `json:"KeyName,omitempty"`
	SecurityGroupIDs []string `json:"SecurityGroupIds,omitempty"`
	SecurityGroups   []string `json:"SecurityGroups,omitempty"`
}

// TerminateInstancesRequest represents a TerminateInstances request.
type TerminateInstancesRequest struct {
	InstanceIDs []string `json:"InstanceIds"`
}

// DescribeInstancesRequest represents a DescribeInstances request.
type DescribeInstancesRequest struct {
	InstanceIDs []string `json:"InstanceIds,omitempty"`
}

// StartInstancesRequest represents a StartInstances request.
type StartInstancesRequest struct {
	InstanceIDs []string `json:"InstanceIds"`
}

// StopInstancesRequest represents a StopInstances request.
type StopInstancesRequest struct {
	InstanceIDs []string `json:"InstanceIds"`
}

// CreateSecurityGroupRequest represents a CreateSecurityGroup request.
type CreateSecurityGroupRequest struct {
	GroupName        string `json:"GroupName"`
	GroupDescription string `json:"GroupDescription"`
	VpcID            string `json:"VpcId,omitempty"`
}

// DeleteSecurityGroupRequest represents a DeleteSecurityGroup request.
type DeleteSecurityGroupRequest struct {
	GroupID   string `json:"GroupId,omitempty"`
	GroupName string `json:"GroupName,omitempty"`
}

// AuthorizeSecurityGroupIngressRequest represents an AuthorizeSecurityGroupIngress request.
type AuthorizeSecurityGroupIngressRequest struct {
	GroupID       string         `json:"GroupId,omitempty"`
	GroupName     string         `json:"GroupName,omitempty"`
	IPPermissions []IPPermission `json:"IPPermissions"`
}

// AuthorizeSecurityGroupEgressRequest represents an AuthorizeSecurityGroupEgress request.
type AuthorizeSecurityGroupEgressRequest struct {
	GroupID       string         `json:"GroupId"`
	IPPermissions []IPPermission `json:"IPPermissions"`
}

// CreateKeyPairRequest represents a CreateKeyPair request.
type CreateKeyPairRequest struct {
	KeyName string `json:"KeyName"`
	KeyType string `json:"KeyType,omitempty"`
}

// DeleteKeyPairRequest represents a DeleteKeyPair request.
type DeleteKeyPairRequest struct {
	KeyName   string `json:"KeyName,omitempty"`
	KeyPairID string `json:"KeyPairId,omitempty"`
}

// DescribeKeyPairsRequest represents a DescribeKeyPairs request.
type DescribeKeyPairsRequest struct {
	KeyNames   []string `json:"KeyNames,omitempty"`
	KeyPairIDs []string `json:"KeyPairIds,omitempty"`
}

// XML Response types for EC2.

// XMLRunInstancesResponse is the XML response for RunInstances.
type XMLRunInstancesResponse struct {
	XMLName       xml.Name        `xml:"RunInstancesResponse"`
	Xmlns         string          `xml:"xmlns,attr"`
	RequestID     string          `xml:"requestId"`
	ReservationID string          `xml:"reservationId"`
	OwnerID       string          `xml:"ownerId"`
	InstancesSet  XMLInstancesSet `xml:"instancesSet"`
}

// XMLInstancesSet contains a list of instances.
type XMLInstancesSet struct {
	Items []XMLInstance `xml:"item"`
}

// XMLInstance represents an instance in XML format.
type XMLInstance struct {
	InstanceID       string           `xml:"instanceId"`
	ImageID          string           `xml:"imageId"`
	InstanceType     string           `xml:"instanceType"`
	InstanceState    XMLInstanceState `xml:"instanceState"`
	PrivateIPAddress string           `xml:"privateIpAddress,omitempty"`
	IPAddress        string           `xml:"ipAddress,omitempty"`
	KeyName          string           `xml:"keyName,omitempty"`
	LaunchTime       string           `xml:"launchTime"`
	GroupSet         XMLGroupSet      `xml:"groupSet"`
}

// XMLInstanceState represents the state of an instance in XML format.
type XMLInstanceState struct {
	Code int    `xml:"code"`
	Name string `xml:"name"`
}

// XMLGroupSet contains a list of security groups.
type XMLGroupSet struct {
	Items []XMLGroupIdentifier `xml:"item"`
}

// XMLGroupIdentifier represents a security group in XML format.
type XMLGroupIdentifier struct {
	GroupID   string `xml:"groupId"`
	GroupName string `xml:"groupName"`
}

// XMLTerminateInstancesResponse is the XML response for TerminateInstances.
type XMLTerminateInstancesResponse struct {
	XMLName      xml.Name                  `xml:"TerminateInstancesResponse"`
	Xmlns        string                    `xml:"xmlns,attr"`
	RequestID    string                    `xml:"requestId"`
	InstancesSet XMLInstanceStateChangeSet `xml:"instancesSet"`
}

// XMLInstanceStateChangeSet contains a list of instance state changes.
type XMLInstanceStateChangeSet struct {
	Items []XMLInstanceStateChange `xml:"item"`
}

// XMLInstanceStateChange represents an instance state change in XML format.
type XMLInstanceStateChange struct {
	InstanceID    string           `xml:"instanceId"`
	CurrentState  XMLInstanceState `xml:"currentState"`
	PreviousState XMLInstanceState `xml:"previousState"`
}

// XMLDescribeInstancesResponse is the XML response for DescribeInstances.
type XMLDescribeInstancesResponse struct {
	XMLName        xml.Name          `xml:"DescribeInstancesResponse"`
	Xmlns          string            `xml:"xmlns,attr"`
	RequestID      string            `xml:"requestId"`
	ReservationSet XMLReservationSet `xml:"reservationSet"`
}

// XMLReservationSet contains a list of reservations.
type XMLReservationSet struct {
	Items []XMLReservation `xml:"item"`
}

// XMLReservation represents a reservation in XML format.
type XMLReservation struct {
	ReservationID string          `xml:"reservationId"`
	OwnerID       string          `xml:"ownerId"`
	InstancesSet  XMLInstancesSet `xml:"instancesSet"`
}

// XMLStartInstancesResponse is the XML response for StartInstances.
type XMLStartInstancesResponse struct {
	XMLName      xml.Name                  `xml:"StartInstancesResponse"`
	Xmlns        string                    `xml:"xmlns,attr"`
	RequestID    string                    `xml:"requestId"`
	InstancesSet XMLInstanceStateChangeSet `xml:"instancesSet"`
}

// XMLStopInstancesResponse is the XML response for StopInstances.
type XMLStopInstancesResponse struct {
	XMLName      xml.Name                  `xml:"StopInstancesResponse"`
	Xmlns        string                    `xml:"xmlns,attr"`
	RequestID    string                    `xml:"requestId"`
	InstancesSet XMLInstanceStateChangeSet `xml:"instancesSet"`
}

// XMLCreateSecurityGroupResponse is the XML response for CreateSecurityGroup.
type XMLCreateSecurityGroupResponse struct {
	XMLName   xml.Name `xml:"CreateSecurityGroupResponse"`
	Xmlns     string   `xml:"xmlns,attr"`
	RequestID string   `xml:"requestId"`
	Return    bool     `xml:"return"`
	GroupID   string   `xml:"groupId"`
}

// XMLDeleteSecurityGroupResponse is the XML response for DeleteSecurityGroup.
type XMLDeleteSecurityGroupResponse struct {
	XMLName   xml.Name `xml:"DeleteSecurityGroupResponse"`
	Xmlns     string   `xml:"xmlns,attr"`
	RequestID string   `xml:"requestId"`
	Return    bool     `xml:"return"`
}

// XMLAuthorizeSecurityGroupIngressResponse is the XML response for AuthorizeSecurityGroupIngress.
type XMLAuthorizeSecurityGroupIngressResponse struct {
	XMLName   xml.Name `xml:"AuthorizeSecurityGroupIngressResponse"`
	Xmlns     string   `xml:"xmlns,attr"`
	RequestID string   `xml:"requestId"`
	Return    bool     `xml:"return"`
}

// XMLAuthorizeSecurityGroupEgressResponse is the XML response for AuthorizeSecurityGroupEgress.
type XMLAuthorizeSecurityGroupEgressResponse struct {
	XMLName   xml.Name `xml:"AuthorizeSecurityGroupEgressResponse"`
	Xmlns     string   `xml:"xmlns,attr"`
	RequestID string   `xml:"requestId"`
	Return    bool     `xml:"return"`
}

// XMLCreateKeyPairResponse is the XML response for CreateKeyPair.
type XMLCreateKeyPairResponse struct {
	XMLName        xml.Name `xml:"CreateKeyPairResponse"`
	Xmlns          string   `xml:"xmlns,attr"`
	RequestID      string   `xml:"requestId"`
	KeyName        string   `xml:"keyName"`
	KeyFingerprint string   `xml:"keyFingerprint"`
	KeyMaterial    string   `xml:"keyMaterial"`
	KeyPairID      string   `xml:"keyPairId"`
}

// XMLDeleteKeyPairResponse is the XML response for DeleteKeyPair.
type XMLDeleteKeyPairResponse struct {
	XMLName   xml.Name `xml:"DeleteKeyPairResponse"`
	Xmlns     string   `xml:"xmlns,attr"`
	RequestID string   `xml:"requestId"`
	Return    bool     `xml:"return"`
}

// XMLDescribeKeyPairsResponse is the XML response for DescribeKeyPairs.
type XMLDescribeKeyPairsResponse struct {
	XMLName   xml.Name      `xml:"DescribeKeyPairsResponse"`
	Xmlns     string        `xml:"xmlns,attr"`
	RequestID string        `xml:"requestId"`
	KeySet    XMLKeyPairSet `xml:"keySet"`
}

// XMLKeyPairSet contains a list of key pairs.
type XMLKeyPairSet struct {
	Items []XMLKeyPairInfo `xml:"item"`
}

// XMLKeyPairInfo represents a key pair in XML format.
type XMLKeyPairInfo struct {
	KeyName        string `xml:"keyName"`
	KeyFingerprint string `xml:"keyFingerprint"`
	KeyPairID      string `xml:"keyPairId"`
}

// XMLErrorResponse is the XML error response for EC2.
type XMLErrorResponse struct {
	XMLName   xml.Name  `xml:"Response"`
	Errors    XMLErrors `xml:"Errors"`
	RequestID string    `xml:"RequestID"`
}

// XMLErrors contains a list of errors.
type XMLErrors struct {
	Error XMLError `xml:"Error"`
}

// XMLError represents an error in XML format.
type XMLError struct {
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

// Error represents an EC2 error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

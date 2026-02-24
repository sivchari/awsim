// Package appmesh provides the AWS App Mesh service implementation.
package appmesh

import "time"

// Error codes for App Mesh API.
const (
	errBadRequestException       = "BadRequestException"
	errResourceNotFoundException = "NotFoundException"
	errConflictException         = "ConflictException"
	errInternalServerException   = "InternalServerErrorException"
	errResourceInUseException    = "ResourceInUseException"
)

// Resource status values.
const (
	StatusActive   = "ACTIVE"
	StatusInactive = "INACTIVE"
	StatusDeleted  = "DELETED"
)

// Protocol values.
const (
	ProtocolHTTP  = "http"
	ProtocolHTTP2 = "http2"
	ProtocolGRPC  = "grpc"
	ProtocolTCP   = "tcp"
)

// Egress filter types.
const (
	EgressFilterAllowAll = "ALLOW_ALL"
	EgressFilterDropAll  = "DROP_ALL"
)

// IP preference values.
const (
	IPPreferenceIPv4Only      = "IPv4_ONLY"
	IPPreferenceIPv6Only      = "IPv6_ONLY"
	IPPreferenceIPv4Preferred = "IPv4_PREFERRED"
	IPPreferenceIPv6Preferred = "IPv6_PREFERRED"
)

// Error represents an App Mesh API error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// HTTPStatusCode returns the HTTP status code for the error.
func (e *Error) HTTPStatusCode() int {
	switch e.Code {
	case errBadRequestException:
		return 400
	case errResourceNotFoundException:
		return 404
	case errConflictException:
		return 409
	case errResourceInUseException:
		return 409
	default:
		return 500
	}
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"message"`
}

// AWSTimestamp represents a timestamp in Unix epoch format.
type AWSTimestamp struct {
	time.Time
}

// MarshalJSON implements json.Marshaler.
func (t AWSTimestamp) MarshalJSON() ([]byte, error) {
	return []byte(t.Format("1136239445")), nil
}

// ResourceMetadata contains common metadata for all resources.
type ResourceMetadata struct {
	Arn           string       `json:"arn"`
	CreatedAt     AWSTimestamp `json:"createdAt"`
	LastUpdatedAt AWSTimestamp `json:"lastUpdatedAt"`
	MeshOwner     string       `json:"meshOwner"`
	ResourceOwner string       `json:"resourceOwner"`
	UID           string       `json:"uid"`
	Version       int64        `json:"version"`
}

// ResourceStatus represents the status of a resource.
type ResourceStatus struct {
	Status string `json:"status"`
}

// Tag represents a key-value tag.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// --- Mesh Types ---

// MeshSpec represents the specification for a mesh.
type MeshSpec struct {
	EgressFilter     *EgressFilter         `json:"egressFilter,omitempty"`
	ServiceDiscovery *MeshServiceDiscovery `json:"serviceDiscovery,omitempty"`
}

// EgressFilter represents the egress filter for a mesh.
type EgressFilter struct {
	Type string `json:"type"`
}

// MeshServiceDiscovery represents the service discovery configuration for a mesh.
type MeshServiceDiscovery struct {
	IPPreference string `json:"ipPreference,omitempty"`
}

// MeshData represents a mesh resource.
type MeshData struct {
	MeshName string           `json:"meshName"`
	Metadata ResourceMetadata `json:"metadata"`
	Spec     MeshSpec         `json:"spec"`
	Status   ResourceStatus   `json:"status"`
}

// MeshRef represents a reference to a mesh in list responses.
type MeshRef struct {
	Arn           string       `json:"arn"`
	CreatedAt     AWSTimestamp `json:"createdAt"`
	LastUpdatedAt AWSTimestamp `json:"lastUpdatedAt"`
	MeshName      string       `json:"meshName"`
	MeshOwner     string       `json:"meshOwner"`
	ResourceOwner string       `json:"resourceOwner"`
	Version       int64        `json:"version"`
}

// CreateMeshInput represents the input for CreateMesh.
type CreateMeshInput struct {
	ClientToken string   `json:"clientToken,omitempty"`
	MeshName    string   `json:"meshName"`
	Spec        MeshSpec `json:"spec,omitempty"`
	Tags        []Tag    `json:"tags,omitempty"`
}

// CreateMeshOutput represents the output for CreateMesh.
type CreateMeshOutput struct {
	Mesh MeshData `json:"mesh"`
}

// DescribeMeshOutput represents the output for DescribeMesh.
type DescribeMeshOutput struct {
	Mesh MeshData `json:"mesh"`
}

// ListMeshesInput represents the input for ListMeshes.
type ListMeshesInput struct {
	Limit     int32
	NextToken string
}

// ListMeshesOutput represents the output for ListMeshes.
type ListMeshesOutput struct {
	Meshes    []MeshRef `json:"meshes"`
	NextToken string    `json:"nextToken,omitempty"`
}

// UpdateMeshInput represents the input for UpdateMesh.
type UpdateMeshInput struct {
	MeshName    string   `json:"-"`
	ClientToken string   `json:"clientToken,omitempty"`
	Spec        MeshSpec `json:"spec,omitempty"`
}

// UpdateMeshOutput represents the output for UpdateMesh.
type UpdateMeshOutput struct {
	Mesh MeshData `json:"mesh"`
}

// DeleteMeshOutput represents the output for DeleteMesh.
type DeleteMeshOutput struct {
	Mesh MeshData `json:"mesh"`
}

// --- Virtual Node Types ---

// PortMapping represents a port mapping.
type PortMapping struct {
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

// Listener represents a listener for a virtual node.
type Listener struct {
	PortMapping PortMapping `json:"portMapping"`
}

// ServiceDiscoveryAWS represents AWS Cloud Map service discovery.
type ServiceDiscoveryAWS struct {
	NamespaceName string `json:"namespaceName"`
	ServiceName   string `json:"serviceName"`
}

// ServiceDiscoveryDNS represents DNS service discovery.
type ServiceDiscoveryDNS struct {
	Hostname string `json:"hostname"`
}

// ServiceDiscovery represents service discovery configuration.
type ServiceDiscovery struct {
	AWSCloudMap *ServiceDiscoveryAWS `json:"awsCloudMap,omitempty"`
	DNS         *ServiceDiscoveryDNS `json:"dns,omitempty"`
}

// VirtualServiceBackend represents a backend virtual service.
type VirtualServiceBackend struct {
	VirtualServiceName string `json:"virtualServiceName"`
}

// Backend represents a backend for a virtual node.
type Backend struct {
	VirtualService *VirtualServiceBackend `json:"virtualService,omitempty"`
}

// VirtualNodeSpec represents the specification for a virtual node.
type VirtualNodeSpec struct {
	Backends         []Backend         `json:"backends,omitempty"`
	Listeners        []Listener        `json:"listeners,omitempty"`
	ServiceDiscovery *ServiceDiscovery `json:"serviceDiscovery,omitempty"`
}

// VirtualNodeData represents a virtual node resource.
type VirtualNodeData struct {
	MeshName        string           `json:"meshName"`
	Metadata        ResourceMetadata `json:"metadata"`
	Spec            VirtualNodeSpec  `json:"spec"`
	Status          ResourceStatus   `json:"status"`
	VirtualNodeName string           `json:"virtualNodeName"`
}

// VirtualNodeRef represents a reference to a virtual node in list responses.
type VirtualNodeRef struct {
	Arn             string       `json:"arn"`
	CreatedAt       AWSTimestamp `json:"createdAt"`
	LastUpdatedAt   AWSTimestamp `json:"lastUpdatedAt"`
	MeshName        string       `json:"meshName"`
	MeshOwner       string       `json:"meshOwner"`
	ResourceOwner   string       `json:"resourceOwner"`
	Version         int64        `json:"version"`
	VirtualNodeName string       `json:"virtualNodeName"`
}

// CreateVirtualNodeInput represents the input for CreateVirtualNode.
type CreateVirtualNodeInput struct {
	MeshName        string          `json:"-"`
	ClientToken     string          `json:"clientToken,omitempty"`
	Spec            VirtualNodeSpec `json:"spec,omitempty"`
	Tags            []Tag           `json:"tags,omitempty"`
	VirtualNodeName string          `json:"virtualNodeName"`
}

// CreateVirtualNodeOutput represents the output for CreateVirtualNode.
type CreateVirtualNodeOutput struct {
	VirtualNode VirtualNodeData `json:"virtualNode"`
}

// DescribeVirtualNodeOutput represents the output for DescribeVirtualNode.
type DescribeVirtualNodeOutput struct {
	VirtualNode VirtualNodeData `json:"virtualNode"`
}

// ListVirtualNodesInput represents the input for ListVirtualNodes.
type ListVirtualNodesInput struct {
	MeshName  string
	Limit     int32
	NextToken string
}

// ListVirtualNodesOutput represents the output for ListVirtualNodes.
type ListVirtualNodesOutput struct {
	VirtualNodes []VirtualNodeRef `json:"virtualNodes"`
	NextToken    string           `json:"nextToken,omitempty"`
}

// UpdateVirtualNodeInput represents the input for UpdateVirtualNode.
type UpdateVirtualNodeInput struct {
	MeshName        string          `json:"-"`
	VirtualNodeName string          `json:"-"`
	ClientToken     string          `json:"clientToken,omitempty"`
	Spec            VirtualNodeSpec `json:"spec,omitempty"`
}

// UpdateVirtualNodeOutput represents the output for UpdateVirtualNode.
type UpdateVirtualNodeOutput struct {
	VirtualNode VirtualNodeData `json:"virtualNode"`
}

// DeleteVirtualNodeOutput represents the output for DeleteVirtualNode.
type DeleteVirtualNodeOutput struct {
	VirtualNode VirtualNodeData `json:"virtualNode"`
}

// --- Virtual Service Types ---

// VirtualNodeServiceProvider represents a virtual node provider.
type VirtualNodeServiceProvider struct {
	VirtualNodeName string `json:"virtualNodeName"`
}

// VirtualRouterServiceProvider represents a virtual router provider.
type VirtualRouterServiceProvider struct {
	VirtualRouterName string `json:"virtualRouterName"`
}

// VirtualServiceProvider represents a provider for a virtual service.
type VirtualServiceProvider struct {
	VirtualNode   *VirtualNodeServiceProvider   `json:"virtualNode,omitempty"`
	VirtualRouter *VirtualRouterServiceProvider `json:"virtualRouter,omitempty"`
}

// VirtualServiceSpec represents the specification for a virtual service.
type VirtualServiceSpec struct {
	Provider *VirtualServiceProvider `json:"provider,omitempty"`
}

// VirtualServiceData represents a virtual service resource.
type VirtualServiceData struct {
	MeshName           string             `json:"meshName"`
	Metadata           ResourceMetadata   `json:"metadata"`
	Spec               VirtualServiceSpec `json:"spec"`
	Status             ResourceStatus     `json:"status"`
	VirtualServiceName string             `json:"virtualServiceName"`
}

// VirtualServiceRef represents a reference to a virtual service in list responses.
type VirtualServiceRef struct {
	Arn                string       `json:"arn"`
	CreatedAt          AWSTimestamp `json:"createdAt"`
	LastUpdatedAt      AWSTimestamp `json:"lastUpdatedAt"`
	MeshName           string       `json:"meshName"`
	MeshOwner          string       `json:"meshOwner"`
	ResourceOwner      string       `json:"resourceOwner"`
	Version            int64        `json:"version"`
	VirtualServiceName string       `json:"virtualServiceName"`
}

// CreateVirtualServiceInput represents the input for CreateVirtualService.
type CreateVirtualServiceInput struct {
	MeshName           string             `json:"-"`
	ClientToken        string             `json:"clientToken,omitempty"`
	Spec               VirtualServiceSpec `json:"spec,omitempty"`
	Tags               []Tag              `json:"tags,omitempty"`
	VirtualServiceName string             `json:"virtualServiceName"`
}

// CreateVirtualServiceOutput represents the output for CreateVirtualService.
type CreateVirtualServiceOutput struct {
	VirtualService VirtualServiceData `json:"virtualService"`
}

// DescribeVirtualServiceOutput represents the output for DescribeVirtualService.
type DescribeVirtualServiceOutput struct {
	VirtualService VirtualServiceData `json:"virtualService"`
}

// ListVirtualServicesInput represents the input for ListVirtualServices.
type ListVirtualServicesInput struct {
	MeshName  string
	Limit     int32
	NextToken string
}

// ListVirtualServicesOutput represents the output for ListVirtualServices.
type ListVirtualServicesOutput struct {
	VirtualServices []VirtualServiceRef `json:"virtualServices"`
	NextToken       string              `json:"nextToken,omitempty"`
}

// UpdateVirtualServiceInput represents the input for UpdateVirtualService.
type UpdateVirtualServiceInput struct {
	MeshName           string             `json:"-"`
	VirtualServiceName string             `json:"-"`
	ClientToken        string             `json:"clientToken,omitempty"`
	Spec               VirtualServiceSpec `json:"spec,omitempty"`
}

// UpdateVirtualServiceOutput represents the output for UpdateVirtualService.
type UpdateVirtualServiceOutput struct {
	VirtualService VirtualServiceData `json:"virtualService"`
}

// DeleteVirtualServiceOutput represents the output for DeleteVirtualService.
type DeleteVirtualServiceOutput struct {
	VirtualService VirtualServiceData `json:"virtualService"`
}

// --- Virtual Router Types ---

// VirtualRouterListener represents a listener for a virtual router.
type VirtualRouterListener struct {
	PortMapping PortMapping `json:"portMapping"`
}

// VirtualRouterSpec represents the specification for a virtual router.
type VirtualRouterSpec struct {
	Listeners []VirtualRouterListener `json:"listeners,omitempty"`
}

// VirtualRouterData represents a virtual router resource.
type VirtualRouterData struct {
	MeshName          string            `json:"meshName"`
	Metadata          ResourceMetadata  `json:"metadata"`
	Spec              VirtualRouterSpec `json:"spec"`
	Status            ResourceStatus    `json:"status"`
	VirtualRouterName string            `json:"virtualRouterName"`
}

// VirtualRouterRef represents a reference to a virtual router in list responses.
type VirtualRouterRef struct {
	Arn               string       `json:"arn"`
	CreatedAt         AWSTimestamp `json:"createdAt"`
	LastUpdatedAt     AWSTimestamp `json:"lastUpdatedAt"`
	MeshName          string       `json:"meshName"`
	MeshOwner         string       `json:"meshOwner"`
	ResourceOwner     string       `json:"resourceOwner"`
	Version           int64        `json:"version"`
	VirtualRouterName string       `json:"virtualRouterName"`
}

// CreateVirtualRouterInput represents the input for CreateVirtualRouter.
type CreateVirtualRouterInput struct {
	MeshName          string            `json:"-"`
	ClientToken       string            `json:"clientToken,omitempty"`
	Spec              VirtualRouterSpec `json:"spec,omitempty"`
	Tags              []Tag             `json:"tags,omitempty"`
	VirtualRouterName string            `json:"virtualRouterName"`
}

// CreateVirtualRouterOutput represents the output for CreateVirtualRouter.
type CreateVirtualRouterOutput struct {
	VirtualRouter VirtualRouterData `json:"virtualRouter"`
}

// DescribeVirtualRouterOutput represents the output for DescribeVirtualRouter.
type DescribeVirtualRouterOutput struct {
	VirtualRouter VirtualRouterData `json:"virtualRouter"`
}

// ListVirtualRoutersInput represents the input for ListVirtualRouters.
type ListVirtualRoutersInput struct {
	MeshName  string
	Limit     int32
	NextToken string
}

// ListVirtualRoutersOutput represents the output for ListVirtualRouters.
type ListVirtualRoutersOutput struct {
	VirtualRouters []VirtualRouterRef `json:"virtualRouters"`
	NextToken      string             `json:"nextToken,omitempty"`
}

// UpdateVirtualRouterInput represents the input for UpdateVirtualRouter.
type UpdateVirtualRouterInput struct {
	MeshName          string            `json:"-"`
	VirtualRouterName string            `json:"-"`
	ClientToken       string            `json:"clientToken,omitempty"`
	Spec              VirtualRouterSpec `json:"spec,omitempty"`
}

// UpdateVirtualRouterOutput represents the output for UpdateVirtualRouter.
type UpdateVirtualRouterOutput struct {
	VirtualRouter VirtualRouterData `json:"virtualRouter"`
}

// DeleteVirtualRouterOutput represents the output for DeleteVirtualRouter.
type DeleteVirtualRouterOutput struct {
	VirtualRouter VirtualRouterData `json:"virtualRouter"`
}

// --- Route Types ---

// WeightedTarget represents a weighted target for a route.
type WeightedTarget struct {
	VirtualNode string `json:"virtualNode"`
	Weight      int32  `json:"weight"`
}

// HTTPRouteAction represents the action for an HTTP route.
type HTTPRouteAction struct {
	WeightedTargets []WeightedTarget `json:"weightedTargets"`
}

// HTTPRouteMatch represents the match criteria for an HTTP route.
type HTTPRouteMatch struct {
	Prefix string `json:"prefix,omitempty"`
	Path   *struct {
		Exact string `json:"exact,omitempty"`
		Regex string `json:"regex,omitempty"`
	} `json:"path,omitempty"`
	Method string `json:"method,omitempty"`
}

// HTTPRoute represents an HTTP route.
type HTTPRoute struct {
	Action HTTPRouteAction `json:"action"`
	Match  HTTPRouteMatch  `json:"match"`
}

// GRPCRouteAction represents the action for a gRPC route.
type GRPCRouteAction struct {
	WeightedTargets []WeightedTarget `json:"weightedTargets"`
}

// GRPCRouteMatch represents the match criteria for a gRPC route.
type GRPCRouteMatch struct {
	ServiceName string `json:"serviceName,omitempty"`
	MethodName  string `json:"methodName,omitempty"`
}

// GRPCRoute represents a gRPC route.
type GRPCRoute struct {
	Action GRPCRouteAction `json:"action"`
	Match  GRPCRouteMatch  `json:"match"`
}

// TCPRouteAction represents the action for a TCP route.
type TCPRouteAction struct {
	WeightedTargets []WeightedTarget `json:"weightedTargets"`
}

// TCPRoute represents a TCP route.
type TCPRoute struct {
	Action TCPRouteAction `json:"action"`
}

// RouteSpec represents the specification for a route.
type RouteSpec struct {
	HTTPRoute  *HTTPRoute `json:"httpRoute,omitempty"`
	HTTP2Route *HTTPRoute `json:"http2Route,omitempty"`
	GRPCRoute  *GRPCRoute `json:"grpcRoute,omitempty"`
	TCPRoute   *TCPRoute  `json:"tcpRoute,omitempty"`
	Priority   *int32     `json:"priority,omitempty"`
}

// RouteData represents a route resource.
type RouteData struct {
	MeshName          string           `json:"meshName"`
	Metadata          ResourceMetadata `json:"metadata"`
	RouteName         string           `json:"routeName"`
	Spec              RouteSpec        `json:"spec"`
	Status            ResourceStatus   `json:"status"`
	VirtualRouterName string           `json:"virtualRouterName"`
}

// RouteRef represents a reference to a route in list responses.
type RouteRef struct {
	Arn               string       `json:"arn"`
	CreatedAt         AWSTimestamp `json:"createdAt"`
	LastUpdatedAt     AWSTimestamp `json:"lastUpdatedAt"`
	MeshName          string       `json:"meshName"`
	MeshOwner         string       `json:"meshOwner"`
	ResourceOwner     string       `json:"resourceOwner"`
	RouteName         string       `json:"routeName"`
	Version           int64        `json:"version"`
	VirtualRouterName string       `json:"virtualRouterName"`
}

// CreateRouteInput represents the input for CreateRoute.
type CreateRouteInput struct {
	MeshName          string    `json:"-"`
	VirtualRouterName string    `json:"-"`
	ClientToken       string    `json:"clientToken,omitempty"`
	RouteName         string    `json:"routeName"`
	Spec              RouteSpec `json:"spec"`
	Tags              []Tag     `json:"tags,omitempty"`
}

// CreateRouteOutput represents the output for CreateRoute.
type CreateRouteOutput struct {
	Route RouteData `json:"route"`
}

// DescribeRouteOutput represents the output for DescribeRoute.
type DescribeRouteOutput struct {
	Route RouteData `json:"route"`
}

// ListRoutesInput represents the input for ListRoutes.
type ListRoutesInput struct {
	MeshName          string
	VirtualRouterName string
	Limit             int32
	NextToken         string
}

// ListRoutesOutput represents the output for ListRoutes.
type ListRoutesOutput struct {
	Routes    []RouteRef `json:"routes"`
	NextToken string     `json:"nextToken,omitempty"`
}

// UpdateRouteInput represents the input for UpdateRoute.
type UpdateRouteInput struct {
	MeshName          string    `json:"-"`
	VirtualRouterName string    `json:"-"`
	RouteName         string    `json:"-"`
	ClientToken       string    `json:"clientToken,omitempty"`
	Spec              RouteSpec `json:"spec"`
}

// UpdateRouteOutput represents the output for UpdateRoute.
type UpdateRouteOutput struct {
	Route RouteData `json:"route"`
}

// DeleteRouteOutput represents the output for DeleteRoute.
type DeleteRouteOutput struct {
	Route RouteData `json:"route"`
}

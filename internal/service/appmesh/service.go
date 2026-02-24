package appmesh

import "github.com/sivchari/awsim/internal/service"

// Service implements the App Mesh service.
type Service struct {
	storage Storage
}

// New creates a new App Mesh service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Name returns the service name.
func (s *Service) Name() string {
	return "appmesh"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return "/v20190125"
}

// RegisterRoutes registers the service routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Mesh operations.
	r.Handle("PUT", "/meshes", s.CreateMesh)
	r.Handle("GET", "/meshes/{meshName}", s.DescribeMesh)
	r.Handle("GET", "/meshes", s.ListMeshes)
	r.Handle("PUT", "/meshes/{meshName}", s.UpdateMesh)
	r.Handle("DELETE", "/meshes/{meshName}", s.DeleteMesh)

	// Virtual Node operations.
	r.Handle("PUT", "/meshes/{meshName}/virtualNodes", s.CreateVirtualNode)
	r.Handle("GET", "/meshes/{meshName}/virtualNodes/{virtualNodeName}", s.DescribeVirtualNode)
	r.Handle("GET", "/meshes/{meshName}/virtualNodes", s.ListVirtualNodes)
	r.Handle("PUT", "/meshes/{meshName}/virtualNodes/{virtualNodeName}", s.UpdateVirtualNode)
	r.Handle("DELETE", "/meshes/{meshName}/virtualNodes/{virtualNodeName}", s.DeleteVirtualNode)

	// Virtual Service operations.
	r.Handle("PUT", "/meshes/{meshName}/virtualServices", s.CreateVirtualService)
	r.Handle("GET", "/meshes/{meshName}/virtualServices/{virtualServiceName}", s.DescribeVirtualService)
	r.Handle("GET", "/meshes/{meshName}/virtualServices", s.ListVirtualServices)
	r.Handle("PUT", "/meshes/{meshName}/virtualServices/{virtualServiceName}", s.UpdateVirtualService)
	r.Handle("DELETE", "/meshes/{meshName}/virtualServices/{virtualServiceName}", s.DeleteVirtualService)

	// Virtual Router operations.
	r.Handle("PUT", "/meshes/{meshName}/virtualRouters", s.CreateVirtualRouter)
	r.Handle("GET", "/meshes/{meshName}/virtualRouters/{virtualRouterName}", s.DescribeVirtualRouter)
	r.Handle("GET", "/meshes/{meshName}/virtualRouters", s.ListVirtualRouters)
	r.Handle("PUT", "/meshes/{meshName}/virtualRouters/{virtualRouterName}", s.UpdateVirtualRouter)
	r.Handle("DELETE", "/meshes/{meshName}/virtualRouters/{virtualRouterName}", s.DeleteVirtualRouter)

	// Route operations.
	r.Handle("PUT", "/meshes/{meshName}/virtualRouter/{virtualRouterName}/routes", s.CreateRoute)
	r.Handle("GET", "/meshes/{meshName}/virtualRouter/{virtualRouterName}/routes/{routeName}", s.DescribeRoute)
	r.Handle("GET", "/meshes/{meshName}/virtualRouter/{virtualRouterName}/routes", s.ListRoutes)
	r.Handle("PUT", "/meshes/{meshName}/virtualRouter/{virtualRouterName}/routes/{routeName}", s.UpdateRoute)
	r.Handle("DELETE", "/meshes/{meshName}/virtualRouter/{virtualRouterName}/routes/{routeName}", s.DeleteRoute)
}

// Compile-time interface check.
var _ service.Service = (*Service)(nil)

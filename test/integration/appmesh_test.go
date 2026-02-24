//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appmesh"
	"github.com/aws/aws-sdk-go-v2/service/appmesh/types"
	"github.com/stretchr/testify/require"
)

func newAppMeshClient(t *testing.T) *appmesh.Client {
	t.Helper()

	return appmesh.NewFromConfig(testAWSConfig(t), func(o *appmesh.Options) {
		o.BaseEndpoint = aws.String(awsimEndpoint)
	})
}

// --- Mesh Tests ---

func TestAppMesh_CreateAndDescribeMesh(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	meshName := "test-mesh"

	// Create mesh.
	createOutput, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
		Spec: &types.MeshSpec{
			EgressFilter: &types.EgressFilter{
				Type: types.EgressFilterTypeAllowAll,
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createOutput.Mesh)
	require.Equal(t, meshName, *createOutput.Mesh.MeshName)
	require.NotEmpty(t, createOutput.Mesh.Metadata.Arn)

	t.Cleanup(func() {
		_, _ = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
			MeshName: aws.String(meshName),
		})
	})

	// Describe mesh.
	descOutput, err := client.DescribeMesh(ctx, &appmesh.DescribeMeshInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)
	require.Equal(t, meshName, *descOutput.Mesh.MeshName)
	require.Equal(t, types.MeshStatusCodeActive, descOutput.Mesh.Status.Status)
}

func TestAppMesh_ListMeshes(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	// Create multiple meshes.
	for i := 0; i < 3; i++ {
		meshName := "list-mesh-" + string(rune('a'+i))
		_, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
			MeshName: aws.String(meshName),
		})
		require.NoError(t, err)

		t.Cleanup(func() {
			_, _ = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
				MeshName: aws.String(meshName),
			})
		})
	}

	// List meshes.
	listOutput, err := client.ListMeshes(ctx, &appmesh.ListMeshesInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.Meshes), 3)
}

func TestAppMesh_UpdateMesh(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	meshName := "update-mesh"

	// Create mesh.
	_, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
		Spec: &types.MeshSpec{
			EgressFilter: &types.EgressFilter{
				Type: types.EgressFilterTypeAllowAll,
			},
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
			MeshName: aws.String(meshName),
		})
	})

	// Update mesh.
	updateOutput, err := client.UpdateMesh(ctx, &appmesh.UpdateMeshInput{
		MeshName: aws.String(meshName),
		Spec: &types.MeshSpec{
			EgressFilter: &types.EgressFilter{
				Type: types.EgressFilterTypeDropAll,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, types.EgressFilterTypeDropAll, updateOutput.Mesh.Spec.EgressFilter.Type)
}

func TestAppMesh_DeleteMesh(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	meshName := "delete-mesh"

	// Create mesh.
	_, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)

	// Delete mesh.
	deleteOutput, err := client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)
	require.Equal(t, meshName, *deleteOutput.Mesh.MeshName)

	// Verify mesh is deleted.
	_, err = client.DescribeMesh(ctx, &appmesh.DescribeMeshInput{
		MeshName: aws.String(meshName),
	})
	require.Error(t, err)
}

// --- Virtual Node Tests ---

func TestAppMesh_CreateAndDescribeVirtualNode(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	meshName := "vn-mesh"
	virtualNodeName := "test-vn"

	// Create mesh.
	_, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteVirtualNode(ctx, &appmesh.DeleteVirtualNodeInput{
			MeshName:        aws.String(meshName),
			VirtualNodeName: aws.String(virtualNodeName),
		})
		_, _ = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
			MeshName: aws.String(meshName),
		})
	})

	// Create virtual node.
	createOutput, err := client.CreateVirtualNode(ctx, &appmesh.CreateVirtualNodeInput{
		MeshName:        aws.String(meshName),
		VirtualNodeName: aws.String(virtualNodeName),
		Spec: &types.VirtualNodeSpec{
			Listeners: []types.Listener{
				{
					PortMapping: &types.PortMapping{
						Port:     aws.Int32(8080),
						Protocol: types.PortProtocolHttp,
					},
				},
			},
			ServiceDiscovery: &types.ServiceDiscovery{
				Dns: &types.DnsServiceDiscovery{
					Hostname: aws.String("test.local"),
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, virtualNodeName, *createOutput.VirtualNode.VirtualNodeName)
	require.NotEmpty(t, createOutput.VirtualNode.Metadata.Arn)

	// Describe virtual node.
	descOutput, err := client.DescribeVirtualNode(ctx, &appmesh.DescribeVirtualNodeInput{
		MeshName:        aws.String(meshName),
		VirtualNodeName: aws.String(virtualNodeName),
	})
	require.NoError(t, err)
	require.Equal(t, virtualNodeName, *descOutput.VirtualNode.VirtualNodeName)
}

func TestAppMesh_ListVirtualNodes(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	meshName := "vn-list-mesh"

	// Create mesh.
	_, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		for i := 0; i < 3; i++ {
			vnName := "vn-" + string(rune('a'+i))
			_, _ = client.DeleteVirtualNode(ctx, &appmesh.DeleteVirtualNodeInput{
				MeshName:        aws.String(meshName),
				VirtualNodeName: aws.String(vnName),
			})
		}
		_, _ = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
			MeshName: aws.String(meshName),
		})
	})

	// Create multiple virtual nodes.
	for i := 0; i < 3; i++ {
		vnName := "vn-" + string(rune('a'+i))
		_, err := client.CreateVirtualNode(ctx, &appmesh.CreateVirtualNodeInput{
			MeshName:        aws.String(meshName),
			VirtualNodeName: aws.String(vnName),
			Spec:            &types.VirtualNodeSpec{},
		})
		require.NoError(t, err)
	}

	// List virtual nodes.
	listOutput, err := client.ListVirtualNodes(ctx, &appmesh.ListVirtualNodesInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)
	require.Len(t, listOutput.VirtualNodes, 3)
}

// --- Virtual Service Tests ---

func TestAppMesh_CreateAndDescribeVirtualService(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	meshName := "vs-mesh"
	virtualServiceName := "test-vs.local"
	virtualNodeName := "backend-node"

	// Create mesh.
	_, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)

	// Create virtual node.
	_, err = client.CreateVirtualNode(ctx, &appmesh.CreateVirtualNodeInput{
		MeshName:        aws.String(meshName),
		VirtualNodeName: aws.String(virtualNodeName),
		Spec:            &types.VirtualNodeSpec{},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteVirtualService(ctx, &appmesh.DeleteVirtualServiceInput{
			MeshName:           aws.String(meshName),
			VirtualServiceName: aws.String(virtualServiceName),
		})
		_, _ = client.DeleteVirtualNode(ctx, &appmesh.DeleteVirtualNodeInput{
			MeshName:        aws.String(meshName),
			VirtualNodeName: aws.String(virtualNodeName),
		})
		_, _ = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
			MeshName: aws.String(meshName),
		})
	})

	// Create virtual service.
	createOutput, err := client.CreateVirtualService(ctx, &appmesh.CreateVirtualServiceInput{
		MeshName:           aws.String(meshName),
		VirtualServiceName: aws.String(virtualServiceName),
		Spec: &types.VirtualServiceSpec{
			Provider: &types.VirtualServiceProvider{
				VirtualNode: &types.VirtualNodeServiceProvider{
					VirtualNodeName: aws.String(virtualNodeName),
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, virtualServiceName, *createOutput.VirtualService.VirtualServiceName)
	require.NotEmpty(t, createOutput.VirtualService.Metadata.Arn)

	// Describe virtual service.
	descOutput, err := client.DescribeVirtualService(ctx, &appmesh.DescribeVirtualServiceInput{
		MeshName:           aws.String(meshName),
		VirtualServiceName: aws.String(virtualServiceName),
	})
	require.NoError(t, err)
	require.Equal(t, virtualServiceName, *descOutput.VirtualService.VirtualServiceName)
}

// --- Virtual Router Tests ---

func TestAppMesh_CreateAndDescribeVirtualRouter(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	meshName := "vr-mesh"
	virtualRouterName := "test-vr"

	// Create mesh.
	_, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteVirtualRouter(ctx, &appmesh.DeleteVirtualRouterInput{
			MeshName:          aws.String(meshName),
			VirtualRouterName: aws.String(virtualRouterName),
		})
		_, _ = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
			MeshName: aws.String(meshName),
		})
	})

	// Create virtual router.
	createOutput, err := client.CreateVirtualRouter(ctx, &appmesh.CreateVirtualRouterInput{
		MeshName:          aws.String(meshName),
		VirtualRouterName: aws.String(virtualRouterName),
		Spec: &types.VirtualRouterSpec{
			Listeners: []types.VirtualRouterListener{
				{
					PortMapping: &types.PortMapping{
						Port:     aws.Int32(8080),
						Protocol: types.PortProtocolHttp,
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, virtualRouterName, *createOutput.VirtualRouter.VirtualRouterName)
	require.NotEmpty(t, createOutput.VirtualRouter.Metadata.Arn)

	// Describe virtual router.
	descOutput, err := client.DescribeVirtualRouter(ctx, &appmesh.DescribeVirtualRouterInput{
		MeshName:          aws.String(meshName),
		VirtualRouterName: aws.String(virtualRouterName),
	})
	require.NoError(t, err)
	require.Equal(t, virtualRouterName, *descOutput.VirtualRouter.VirtualRouterName)
}

// --- Route Tests ---

func TestAppMesh_CreateAndDescribeRoute(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	meshName := "route-mesh"
	virtualRouterName := "route-vr"
	virtualNodeName := "route-vn"
	routeName := "test-route"

	// Create mesh.
	_, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)

	// Create virtual node.
	_, err = client.CreateVirtualNode(ctx, &appmesh.CreateVirtualNodeInput{
		MeshName:        aws.String(meshName),
		VirtualNodeName: aws.String(virtualNodeName),
		Spec:            &types.VirtualNodeSpec{},
	})
	require.NoError(t, err)

	// Create virtual router.
	_, err = client.CreateVirtualRouter(ctx, &appmesh.CreateVirtualRouterInput{
		MeshName:          aws.String(meshName),
		VirtualRouterName: aws.String(virtualRouterName),
		Spec: &types.VirtualRouterSpec{
			Listeners: []types.VirtualRouterListener{
				{
					PortMapping: &types.PortMapping{
						Port:     aws.Int32(8080),
						Protocol: types.PortProtocolHttp,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteRoute(ctx, &appmesh.DeleteRouteInput{
			MeshName:          aws.String(meshName),
			VirtualRouterName: aws.String(virtualRouterName),
			RouteName:         aws.String(routeName),
		})
		_, _ = client.DeleteVirtualRouter(ctx, &appmesh.DeleteVirtualRouterInput{
			MeshName:          aws.String(meshName),
			VirtualRouterName: aws.String(virtualRouterName),
		})
		_, _ = client.DeleteVirtualNode(ctx, &appmesh.DeleteVirtualNodeInput{
			MeshName:        aws.String(meshName),
			VirtualNodeName: aws.String(virtualNodeName),
		})
		_, _ = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
			MeshName: aws.String(meshName),
		})
	})

	// Create route.
	createOutput, err := client.CreateRoute(ctx, &appmesh.CreateRouteInput{
		MeshName:          aws.String(meshName),
		VirtualRouterName: aws.String(virtualRouterName),
		RouteName:         aws.String(routeName),
		Spec: &types.RouteSpec{
			HttpRoute: &types.HttpRoute{
				Match: &types.HttpRouteMatch{
					Prefix: aws.String("/"),
				},
				Action: &types.HttpRouteAction{
					WeightedTargets: []types.WeightedTarget{
						{
							VirtualNode: aws.String(virtualNodeName),
							Weight:      100,
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, routeName, *createOutput.Route.RouteName)
	require.NotEmpty(t, createOutput.Route.Metadata.Arn)

	// Describe route.
	descOutput, err := client.DescribeRoute(ctx, &appmesh.DescribeRouteInput{
		MeshName:          aws.String(meshName),
		VirtualRouterName: aws.String(virtualRouterName),
		RouteName:         aws.String(routeName),
	})
	require.NoError(t, err)
	require.Equal(t, routeName, *descOutput.Route.RouteName)
}

func TestAppMesh_ListRoutes(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	meshName := "route-list-mesh"
	virtualRouterName := "route-list-vr"
	virtualNodeName := "route-list-vn"

	// Create mesh.
	_, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)

	// Create virtual node.
	_, err = client.CreateVirtualNode(ctx, &appmesh.CreateVirtualNodeInput{
		MeshName:        aws.String(meshName),
		VirtualNodeName: aws.String(virtualNodeName),
		Spec:            &types.VirtualNodeSpec{},
	})
	require.NoError(t, err)

	// Create virtual router.
	_, err = client.CreateVirtualRouter(ctx, &appmesh.CreateVirtualRouterInput{
		MeshName:          aws.String(meshName),
		VirtualRouterName: aws.String(virtualRouterName),
		Spec: &types.VirtualRouterSpec{
			Listeners: []types.VirtualRouterListener{
				{
					PortMapping: &types.PortMapping{
						Port:     aws.Int32(8080),
						Protocol: types.PortProtocolHttp,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		for i := 0; i < 3; i++ {
			rName := "route-" + string(rune('a'+i))
			_, _ = client.DeleteRoute(ctx, &appmesh.DeleteRouteInput{
				MeshName:          aws.String(meshName),
				VirtualRouterName: aws.String(virtualRouterName),
				RouteName:         aws.String(rName),
			})
		}
		_, _ = client.DeleteVirtualRouter(ctx, &appmesh.DeleteVirtualRouterInput{
			MeshName:          aws.String(meshName),
			VirtualRouterName: aws.String(virtualRouterName),
		})
		_, _ = client.DeleteVirtualNode(ctx, &appmesh.DeleteVirtualNodeInput{
			MeshName:        aws.String(meshName),
			VirtualNodeName: aws.String(virtualNodeName),
		})
		_, _ = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
			MeshName: aws.String(meshName),
		})
	})

	// Create multiple routes.
	for i := 0; i < 3; i++ {
		rName := "route-" + string(rune('a'+i))
		_, err := client.CreateRoute(ctx, &appmesh.CreateRouteInput{
			MeshName:          aws.String(meshName),
			VirtualRouterName: aws.String(virtualRouterName),
			RouteName:         aws.String(rName),
			Spec: &types.RouteSpec{
				HttpRoute: &types.HttpRoute{
					Match: &types.HttpRouteMatch{
						Prefix: aws.String("/"),
					},
					Action: &types.HttpRouteAction{
						WeightedTargets: []types.WeightedTarget{
							{
								VirtualNode: aws.String(virtualNodeName),
								Weight:      100,
							},
						},
					},
				},
			},
		})
		require.NoError(t, err)
	}

	// List routes.
	listOutput, err := client.ListRoutes(ctx, &appmesh.ListRoutesInput{
		MeshName:          aws.String(meshName),
		VirtualRouterName: aws.String(virtualRouterName),
	})
	require.NoError(t, err)
	require.Len(t, listOutput.Routes, 3)
}

// --- Error Cases ---

func TestAppMesh_MeshNotFound(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	_, err := client.DescribeMesh(ctx, &appmesh.DescribeMeshInput{
		MeshName: aws.String("non-existent-mesh"),
	})
	require.Error(t, err)
}

func TestAppMesh_DuplicateMesh(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	meshName := "duplicate-mesh"

	_, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
			MeshName: aws.String(meshName),
		})
	})

	// Try to create duplicate mesh.
	_, err = client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
	})
	require.Error(t, err)
}

func TestAppMesh_DeleteMeshWithResources(t *testing.T) {
	client := newAppMeshClient(t)
	ctx := t.Context()

	meshName := "delete-mesh-with-resources"
	virtualNodeName := "blocking-vn"

	// Create mesh.
	_, err := client.CreateMesh(ctx, &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
	})
	require.NoError(t, err)

	// Create virtual node.
	_, err = client.CreateVirtualNode(ctx, &appmesh.CreateVirtualNodeInput{
		MeshName:        aws.String(meshName),
		VirtualNodeName: aws.String(virtualNodeName),
		Spec:            &types.VirtualNodeSpec{},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteVirtualNode(ctx, &appmesh.DeleteVirtualNodeInput{
			MeshName:        aws.String(meshName),
			VirtualNodeName: aws.String(virtualNodeName),
		})
		_, _ = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
			MeshName: aws.String(meshName),
		})
	})

	// Try to delete mesh with virtual node - should fail.
	_, err = client.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
		MeshName: aws.String(meshName),
	})
	require.Error(t, err)
}

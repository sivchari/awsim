{
  "Route": {
    "MeshName": "route-mesh",
    "Metadata": {
      "Arn": "arn:aws:appmesh:us-east-1:123456789012:mesh/route-mesh/virtualRouter/route-vr/route/test-route",
      "CreatedAt": "2026-02-26T15:21:04.392Z",
      "LastUpdatedAt": "2026-02-26T15:21:04.392Z",
      "MeshOwner": "123456789012",
      "ResourceOwner": "123456789012",
      "Uid": "20d6ee06-1a00-45da-a70e-5c6b93c41954",
      "Version": 1
    },
    "RouteName": "test-route",
    "Spec": {
      "GrpcRoute": null,
      "Http2Route": null,
      "HttpRoute": {
        "Action": {
          "WeightedTargets": [
            {
              "VirtualNode": "route-vn",
              "Weight": 100,
              "Port": null
            }
          ]
        },
        "Match": {
          "Headers": null,
          "Method": "",
          "Path": null,
          "Port": null,
          "Prefix": "/",
          "QueryParameters": null,
          "Scheme": ""
        },
        "RetryPolicy": null,
        "Timeout": null
      },
      "Priority": null,
      "TcpRoute": null
    },
    "Status": {
      "Status": "ACTIVE"
    },
    "VirtualRouterName": "route-vr"
  },
  "ResultMetadata": {}
}
{
  "Route": {
    "MeshName": "route-mesh",
    "Metadata": {
      "Arn": "arn:aws:appmesh:us-east-1:123456789012:mesh/route-mesh/virtualRouter/route-vr/route/test-route",
      "CreatedAt": "2026-03-23T07:45:24.909Z",
      "LastUpdatedAt": "2026-03-23T07:45:24.909Z",
      "MeshOwner": "123456789012",
      "ResourceOwner": "123456789012",
      "Uid": "50a4bff3-db75-44fd-b1f0-f8dd51198aec",
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
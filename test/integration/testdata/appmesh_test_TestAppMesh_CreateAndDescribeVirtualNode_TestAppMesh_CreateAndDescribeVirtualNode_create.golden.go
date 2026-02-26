{
  "VirtualNode": {
    "MeshName": "vn-mesh",
    "Metadata": {
      "Arn": "arn:aws:appmesh:us-east-1:123456789012:mesh/vn-mesh/virtualNode/test-vn",
      "CreatedAt": "2026-02-26T15:21:04.372Z",
      "LastUpdatedAt": "2026-02-26T15:21:04.372Z",
      "MeshOwner": "123456789012",
      "ResourceOwner": "123456789012",
      "Uid": "2e4e7d49-f3f3-48f2-af62-f37526bdc53e",
      "Version": 1
    },
    "Spec": {
      "BackendDefaults": null,
      "Backends": null,
      "Listeners": [
        {
          "PortMapping": {
            "Port": 8080,
            "Protocol": "http"
          },
          "ConnectionPool": null,
          "HealthCheck": null,
          "OutlierDetection": null,
          "Timeout": null,
          "Tls": null
        }
      ],
      "Logging": null,
      "ServiceDiscovery": {
        "Value": {
          "Hostname": "test.local",
          "IpPreference": "",
          "ResponseType": ""
        }
      }
    },
    "Status": {
      "Status": "ACTIVE"
    },
    "VirtualNodeName": "test-vn"
  },
  "ResultMetadata": {}
}
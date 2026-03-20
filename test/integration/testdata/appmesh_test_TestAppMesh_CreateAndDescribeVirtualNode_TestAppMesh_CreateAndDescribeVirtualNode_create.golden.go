{
  "VirtualNode": {
    "MeshName": "vn-mesh",
    "Metadata": {
      "Arn": "arn:aws:appmesh:us-east-1:123456789012:mesh/vn-mesh/virtualNode/test-vn",
      "CreatedAt": "2026-03-23T07:45:24.89Z",
      "LastUpdatedAt": "2026-03-23T07:45:24.89Z",
      "MeshOwner": "123456789012",
      "ResourceOwner": "123456789012",
      "Uid": "862257ce-5acf-4bfd-b19a-d413ec3484de",
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
{
  "Cluster": {
    "AccessConfig": null,
    "Arn": "arn:aws:eks:us-east-1:123456789012:cluster/test-describe-cluster",
    "CertificateAuthority": {
      "Data": "ZmFrZS1jZXJ0aWZpY2F0ZS1hdXRob3JpdHktZGF0YQ=="
    },
    "ClientRequestToken": null,
    "ComputeConfig": null,
    "ConnectorConfig": null,
    "ControlPlaneScalingConfig": null,
    "CreatedAt": "2026-03-23T07:45:25.687Z",
    "DeletionProtection": null,
    "EncryptionConfig": null,
    "Endpoint": "https://527ec210.gr7.us-east-1.eks.amazonaws.com",
    "Health": {
      "Issues": null
    },
    "Id": null,
    "Identity": {
      "Oidc": {
        "Issuer": "https://oidc.eks.us-east-1.amazonaws.com/id/046dfca7-ff8c-401b-aa64-edbd065a"
      }
    },
    "KubernetesNetworkConfig": {
      "ElasticLoadBalancing": null,
      "IpFamily": "ipv4",
      "ServiceIpv4Cidr": "10.100.0.0/16",
      "ServiceIpv6Cidr": null
    },
    "Logging": null,
    "Name": "test-describe-cluster",
    "OutpostConfig": null,
    "PlatformVersion": "eks.1",
    "RemoteNetworkConfig": null,
    "ResourcesVpcConfig": {
      "ClusterSecurityGroupId": "sg-86f8a5b5-0b91-43c",
      "EndpointPrivateAccess": false,
      "EndpointPublicAccess": true,
      "PublicAccessCidrs": [
        "0.0.0.0/0"
      ],
      "SecurityGroupIds": null,
      "SubnetIds": [
        "subnet-12345678"
      ],
      "VpcId": "vpc-4ae04196-820b-4e4"
    },
    "RoleArn": "arn:aws:iam::123456789012:role/eks-cluster-role",
    "Status": "ACTIVE",
    "StorageConfig": null,
    "Tags": null,
    "UpgradePolicy": null,
    "Version": "1.29",
    "ZonalShiftConfig": null
  },
  "ResultMetadata": {}
}
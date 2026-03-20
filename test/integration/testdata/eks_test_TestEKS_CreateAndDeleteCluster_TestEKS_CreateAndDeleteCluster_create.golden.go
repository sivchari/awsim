{
  "Cluster": {
    "AccessConfig": null,
    "Arn": "arn:aws:eks:us-east-1:123456789012:cluster/test-cluster",
    "CertificateAuthority": {
      "Data": "ZmFrZS1jZXJ0aWZpY2F0ZS1hdXRob3JpdHktZGF0YQ=="
    },
    "ClientRequestToken": null,
    "ComputeConfig": null,
    "ConnectorConfig": null,
    "ControlPlaneScalingConfig": null,
    "CreatedAt": "2026-03-23T07:45:25.684Z",
    "DeletionProtection": null,
    "EncryptionConfig": null,
    "Endpoint": "https://59201633.gr7.us-east-1.eks.amazonaws.com",
    "Health": {
      "Issues": null
    },
    "Id": null,
    "Identity": {
      "Oidc": {
        "Issuer": "https://oidc.eks.us-east-1.amazonaws.com/id/5bf5d12f-34de-48e5-888a-40318d78"
      }
    },
    "KubernetesNetworkConfig": {
      "ElasticLoadBalancing": null,
      "IpFamily": "ipv4",
      "ServiceIpv4Cidr": "10.100.0.0/16",
      "ServiceIpv6Cidr": null
    },
    "Logging": null,
    "Name": "test-cluster",
    "OutpostConfig": null,
    "PlatformVersion": "eks.1",
    "RemoteNetworkConfig": null,
    "ResourcesVpcConfig": {
      "ClusterSecurityGroupId": "sg-e0840bed-e33c-460",
      "EndpointPrivateAccess": false,
      "EndpointPublicAccess": true,
      "PublicAccessCidrs": [
        "0.0.0.0/0"
      ],
      "SecurityGroupIds": null,
      "SubnetIds": [
        "subnet-12345678",
        "subnet-87654321"
      ],
      "VpcId": "vpc-56d726f5-3222-4ae"
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
{
  "Marker": null,
  "ReplicationGroups": [
    {
      "ARN": "arn:aws:elasticache:us-east-1:000000000000:replicationgroup:test-replication-group",
      "AtRestEncryptionEnabled": false,
      "AuthTokenEnabled": false,
      "AuthTokenLastModifiedDate": null,
      "AutoMinorVersionUpgrade": false,
      "AutomaticFailover": "disabled",
      "CacheNodeType": "cache.t3.micro",
      "ClusterEnabled": false,
      "ClusterMode": "",
      "ConfigurationEndpoint": {
        "Address": "test-replication-group.45dcef12.clustercfg.us-east-1.cache.amazonaws.com",
        "Port": 6379
      },
      "DataTiering": "",
      "Description": "Test replication group",
      "Engine": null,
      "GlobalReplicationGroupInfo": null,
      "IpDiscovery": "",
      "KmsKeyId": null,
      "LogDeliveryConfigurations": null,
      "MemberClusters": [],
      "MemberClustersOutpostArns": null,
      "MultiAZ": "disabled",
      "NetworkType": "",
      "NodeGroups": [
        {
          "NodeGroupId": "0001",
          "NodeGroupMembers": [
            {
              "CacheClusterId": "test-replication-group-0001-0001",
              "CacheNodeId": "0001",
              "CurrentRole": "primary",
              "PreferredAvailabilityZone": "us-east-1a",
              "PreferredOutpostArn": null,
              "ReadEndpoint": {
                "Address": "test-replication-group-0001-0001.067d3adb.us-east-1.cache.amazonaws.com",
                "Port": 6379
              }
            }
          ],
          "PrimaryEndpoint": {
            "Address": "test-replication-group-0001.81a2a63c.us-east-1.cache.amazonaws.com",
            "Port": 6379
          },
          "ReaderEndpoint": {
            "Address": "test-replication-group-0001-ro.8f2d6d7d.us-east-1.cache.amazonaws.com",
            "Port": 6379
          },
          "Slots": null,
          "Status": "available"
        }
      ],
      "PendingModifiedValues": null,
      "ReplicationGroupCreateTime": "2026-03-23T07:45:25.709Z",
      "ReplicationGroupId": "test-replication-group",
      "SnapshotRetentionLimit": 0,
      "SnapshotWindow": null,
      "SnapshottingClusterId": null,
      "Status": "available",
      "TransitEncryptionEnabled": false,
      "TransitEncryptionMode": "",
      "UserGroupIds": null
    }
  ],
  "ResultMetadata": {}
}
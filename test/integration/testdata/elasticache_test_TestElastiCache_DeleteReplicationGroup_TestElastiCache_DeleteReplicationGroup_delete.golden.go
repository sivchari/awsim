{
  "ReplicationGroup": {
    "ARN": "arn:aws:elasticache:us-east-1:000000000000:replicationgroup:test-delete-replication-group",
    "AtRestEncryptionEnabled": false,
    "AuthTokenEnabled": false,
    "AuthTokenLastModifiedDate": null,
    "AutoMinorVersionUpgrade": false,
    "AutomaticFailover": "disabled",
    "CacheNodeType": "cache.t3.micro",
    "ClusterEnabled": false,
    "ClusterMode": "",
    "ConfigurationEndpoint": {
      "Address": "test-delete-replication-group.2eaf4b6f.clustercfg.us-east-1.cache.amazonaws.com",
      "Port": 6379
    },
    "DataTiering": "",
    "Description": "Test replication group for deletion",
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
            "CacheClusterId": "test-delete-replication-group-0001-0001",
            "CacheNodeId": "0001",
            "CurrentRole": "primary",
            "PreferredAvailabilityZone": "us-east-1a",
            "PreferredOutpostArn": null,
            "ReadEndpoint": {
              "Address": "test-delete-replication-group-0001-0001.ffddca10.us-east-1.cache.amazonaws.com",
              "Port": 6379
            }
          }
        ],
        "PrimaryEndpoint": {
          "Address": "test-delete-replication-group-0001.7b42ef79.us-east-1.cache.amazonaws.com",
          "Port": 6379
        },
        "ReaderEndpoint": {
          "Address": "test-delete-replication-group-0001-ro.6c85107e.us-east-1.cache.amazonaws.com",
          "Port": 6379
        },
        "Slots": null,
        "Status": "available"
      }
    ],
    "PendingModifiedValues": null,
    "ReplicationGroupCreateTime": "2026-03-23T07:45:25.712Z",
    "ReplicationGroupId": "test-delete-replication-group",
    "SnapshotRetentionLimit": 0,
    "SnapshotWindow": null,
    "SnapshottingClusterId": null,
    "Status": "deleting",
    "TransitEncryptionEnabled": false,
    "TransitEncryptionMode": "",
    "UserGroupIds": null
  },
  "ResultMetadata": {}
}
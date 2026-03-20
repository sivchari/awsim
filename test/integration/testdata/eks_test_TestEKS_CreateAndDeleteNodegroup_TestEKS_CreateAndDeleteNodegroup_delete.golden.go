{
  "Nodegroup": {
    "AmiType": "AL2_x86_64",
    "CapacityType": "ON_DEMAND",
    "ClusterName": "test-nodegroup-cluster",
    "CreatedAt": "2026-03-23T07:45:25.691Z",
    "DiskSize": null,
    "Health": {
      "Issues": null
    },
    "InstanceTypes": [
      "t3.medium"
    ],
    "Labels": null,
    "LaunchTemplate": null,
    "ModifiedAt": "2026-03-23T07:45:25.691Z",
    "NodeRepairConfig": null,
    "NodeRole": "arn:aws:iam::123456789012:role/eks-nodegroup-role",
    "NodegroupArn": "arn:aws:eks:us-east-1:123456789012:nodegroup/test-nodegroup-cluster/test-nodegroup/3ebe54e1",
    "NodegroupName": "test-nodegroup",
    "ReleaseVersion": "1.29-20231116",
    "RemoteAccess": null,
    "Resources": {
      "AutoScalingGroups": [
        {
          "Name": "eks-test-nodegroup-47ca6304"
        }
      ],
      "RemoteAccessSecurityGroup": null
    },
    "ScalingConfig": {
      "DesiredSize": 2,
      "MaxSize": 3,
      "MinSize": 1
    },
    "Status": "DELETING",
    "Subnets": [
      "subnet-12345678",
      "subnet-87654321"
    ],
    "Tags": null,
    "Taints": null,
    "UpdateConfig": null,
    "Version": "1.29"
  },
  "ResultMetadata": {}
}
{
  "Policy": {
    "DateCreated": "2026-03-23T07:45:25.313853796Z",
    "DateModified": "2026-03-23T07:45:25.313853796Z",
    "DefaultPolicy": null,
    "Description": "Test policy for EBS snapshots",
    "ExecutionRoleArn": "arn:aws:iam::123456789012:role/dlm-role",
    "PolicyArn": "arn:aws:dlm:us-east-1:123456789012:policy/policy-d847a5c1-f376-437",
    "PolicyDetails": {
      "Actions": null,
      "CopyTags": null,
      "CreateInterval": null,
      "CrossRegionCopyTargets": null,
      "EventSource": null,
      "Exclusions": null,
      "ExtendDeletion": null,
      "Parameters": null,
      "PolicyLanguage": "",
      "PolicyType": "",
      "ResourceLocations": null,
      "ResourceType": "",
      "ResourceTypes": [
        "VOLUME"
      ],
      "RetainInterval": null,
      "Schedules": [
        {
          "ArchiveRule": null,
          "CopyTags": null,
          "CreateRule": {
            "CronExpression": null,
            "Interval": 24,
            "IntervalUnit": "HOURS",
            "Location": "",
            "Scripts": null,
            "Times": [
              "03:00"
            ]
          },
          "CrossRegionCopyRules": null,
          "DeprecateRule": null,
          "FastRestoreRule": null,
          "Name": "Daily snapshots",
          "RetainRule": {
            "Count": 7,
            "Interval": null,
            "IntervalUnit": ""
          },
          "ShareRules": null,
          "TagsToAdd": null,
          "VariableTags": null
        }
      ],
      "TargetTags": [
        {
          "Key": "Backup",
          "Value": "true"
        }
      ]
    },
    "PolicyId": "policy-d847a5c1-f376-437",
    "State": "ENABLED",
    "StatusMessage": null,
    "Tags": null
  },
  "ResultMetadata": {}
}
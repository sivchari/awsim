{
  "Policy": {
    "DateCreated": "2026-03-23T07:45:25.317575588Z",
    "DateModified": "2026-03-23T07:45:25.318078004Z",
    "DefaultPolicy": null,
    "Description": "Updated test policy",
    "ExecutionRoleArn": "arn:aws:iam::123456789012:role/dlm-role",
    "PolicyArn": "arn:aws:dlm:us-east-1:123456789012:policy/policy-cb1ca244-6252-4d6",
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
            "Times": null
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
    "PolicyId": "policy-cb1ca244-6252-4d6",
    "State": "DISABLED",
    "StatusMessage": null,
    "Tags": null
  },
  "ResultMetadata": {}
}
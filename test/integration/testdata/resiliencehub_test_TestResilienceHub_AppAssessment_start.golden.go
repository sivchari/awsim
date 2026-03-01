{
  "Assessment": {
    "AssessmentArn": "arn:aws:resiliencehub:us-east-1:123456789012:app-assessment/ffec8a40-b120-4912-a85f-d7dedeb8d747",
    "AssessmentStatus": "Success",
    "Invoker": "User",
    "AppArn": "arn:aws:resiliencehub:us-east-1:123456789012:app/7b1c7225-c06f-4ccb-bb0b-09582c05eabf",
    "AppVersion": "release",
    "AssessmentName": "test-assessment",
    "Compliance": {
      "Hardware": {
        "ComplianceStatus": "PolicyMet",
        "AchievableRpoInSecs": 3600,
        "AchievableRtoInSecs": 3600,
        "CurrentRpoInSecs": 3600,
        "CurrentRtoInSecs": 3600,
        "Message": null,
        "RpoDescription": null,
        "RpoReferenceId": null,
        "RtoDescription": null,
        "RtoReferenceId": null
      },
      "Software": {
        "ComplianceStatus": "PolicyBreached",
        "AchievableRpoInSecs": 3600,
        "AchievableRtoInSecs": 3600,
        "CurrentRpoInSecs": 86400,
        "CurrentRtoInSecs": 86400,
        "Message": null,
        "RpoDescription": null,
        "RpoReferenceId": null,
        "RtoDescription": null,
        "RtoReferenceId": null
      }
    },
    "ComplianceStatus": "PolicyBreached",
    "Cost": null,
    "DriftStatus": "",
    "EndTime": "2026-03-01T14:15:41Z",
    "Message": null,
    "Policy": null,
    "ResiliencyScore": {
      "DisruptionScore": {
        "Hardware": 100,
        "Software": 50
      },
      "Score": 75,
      "ComponentScore": null
    },
    "ResourceErrorsDetails": null,
    "StartTime": "2026-03-01T14:15:41Z",
    "Summary": null,
    "Tags": null,
    "VersionName": null
  },
  "ResultMetadata": {}
}
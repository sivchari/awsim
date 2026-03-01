{
  "Assessment": {
    "AssessmentArn": "arn:aws:resiliencehub:us-east-1:123456789012:app-assessment/a86053d3-5218-440f-9e39-ac751bf19eac",
    "AssessmentStatus": "Success",
    "Invoker": "User",
    "AppArn": "arn:aws:resiliencehub:us-east-1:123456789012:app/ac0727ea-a939-4f47-83b0-48e63bd7a2f5",
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
    "EndTime": "2026-03-01T15:05:15Z",
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
    "StartTime": "2026-03-01T15:05:15Z",
    "Summary": null,
    "Tags": null,
    "VersionName": null
  },
  "ResultMetadata": {}
}
{
  "Assessment": {
    "AssessmentArn": "arn:aws:resiliencehub:us-east-1:123456789012:app-assessment/9919b456-7794-44be-b199-9b9a9b74e9bf",
    "AssessmentStatus": "Success",
    "Invoker": "User",
    "AppArn": "arn:aws:resiliencehub:us-east-1:123456789012:app/ed7409db-79a3-4b7d-8b9e-3e46865aeba8",
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
    "EndTime": "2026-03-23T07:45:26Z",
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
    "StartTime": "2026-03-23T07:45:26Z",
    "Summary": null,
    "Tags": null,
    "VersionName": null
  },
  "ResultMetadata": {}
}
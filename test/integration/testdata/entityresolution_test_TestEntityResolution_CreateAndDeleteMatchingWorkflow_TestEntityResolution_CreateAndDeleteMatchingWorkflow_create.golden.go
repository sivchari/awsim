{
  "InputSourceConfig": [
    {
      "InputSourceARN": null,
      "SchemaName": "test-schema",
      "ApplyNormalization": null
    }
  ],
  "OutputSourceConfig": [
    {
      "Output": [
        {
          "Name": "id",
          "Hashed": null
        }
      ],
      "ApplyNormalization": null,
      "CustomerProfilesIntegrationConfig": null,
      "KMSArn": null,
      "OutputS3Path": "s3://bucket/output/"
    }
  ],
  "ResolutionTechniques": {
    "ResolutionType": "RULE_MATCHING",
    "ProviderProperties": null,
    "RuleBasedProperties": null,
    "RuleConditionProperties": null
  },
  "RoleArn": "arn:aws:iam::000000000000:role/test-role",
  "WorkflowArn": "arn:aws:entityresolution:us-east-1:000000000000:matchingworkflow/test-matching-workflow",
  "WorkflowName": "test-matching-workflow",
  "Description": null,
  "IncrementalRunConfig": null,
  "ResultMetadata": {}
}
{
  "ActionAfterCompletion": "NONE",
  "Arn": "arn:aws:scheduler:us-east-1:123456789012:schedule/default/test-schedule-update",
  "CreationDate": "2026-02-26T15:21:07Z",
  "Description": null,
  "EndDate": null,
  "FlexibleTimeWindow": {
    "Mode": "OFF",
    "MaximumWindowInMinutes": null
  },
  "GroupName": "default",
  "KmsKeyArn": null,
  "LastModificationDate": "2026-02-26T15:21:07Z",
  "Name": "test-schedule-update",
  "ScheduleExpression": "rate(2 hours)",
  "ScheduleExpressionTimezone": "UTC",
  "StartDate": null,
  "State": "ENABLED",
  "Target": {
    "Arn": "arn:aws:sqs:us-east-1:123456789012:my-queue-updated",
    "RoleArn": "arn:aws:iam::123456789012:role/scheduler-role",
    "DeadLetterConfig": null,
    "EcsParameters": null,
    "EventBridgeParameters": null,
    "Input": null,
    "KinesisParameters": null,
    "RetryPolicy": null,
    "SageMakerPipelineParameters": null,
    "SqsParameters": null
  },
  "ResultMetadata": {}
}
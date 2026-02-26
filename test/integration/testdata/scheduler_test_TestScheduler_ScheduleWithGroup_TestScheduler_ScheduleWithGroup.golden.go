{
  "ActionAfterCompletion": "NONE",
  "Arn": "arn:aws:scheduler:us-east-1:123456789012:schedule/test-schedule-group-with-schedule/test-schedule-in-group",
  "CreationDate": "2026-02-26T15:21:07Z",
  "Description": null,
  "EndDate": null,
  "FlexibleTimeWindow": {
    "Mode": "OFF",
    "MaximumWindowInMinutes": null
  },
  "GroupName": "test-schedule-group-with-schedule",
  "KmsKeyArn": null,
  "LastModificationDate": "2026-02-26T15:21:07Z",
  "Name": "test-schedule-in-group",
  "ScheduleExpression": "rate(1 hour)",
  "ScheduleExpressionTimezone": "UTC",
  "StartDate": null,
  "State": "ENABLED",
  "Target": {
    "Arn": "arn:aws:sqs:us-east-1:123456789012:my-queue",
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
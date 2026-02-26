{
  "ActionAfterCompletion": "NONE",
  "Arn": "arn:aws:scheduler:us-east-1:123456789012:schedule/default/test-schedule",
  "CreationDate": "2026-02-26T16:00:39Z",
  "Description": null,
  "EndDate": null,
  "FlexibleTimeWindow": {
    "Mode": "OFF",
    "MaximumWindowInMinutes": null
  },
  "GroupName": "default",
  "KmsKeyArn": null,
  "LastModificationDate": "2026-02-26T16:00:39Z",
  "Name": "test-schedule",
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
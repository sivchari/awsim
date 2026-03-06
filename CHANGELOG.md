# Changelog

## [v0.2.0](https://github.com/sivchari/awsim/compare/v0.1.3...v0.2.0) - 2026-03-06
- feat(mq): implement Amazon MQ service by @sivchari in https://github.com/sivchari/awsim/pull/290
- feat(ds): implement AWS Directory Service by @sivchari in https://github.com/sivchari/awsim/pull/292
- refactor: remove unused Prefix() method from Service interface by @sivchari in https://github.com/sivchari/awsim/pull/294
- feat(s3control): implement AWS S3 Control service by @sivchari in https://github.com/sivchari/awsim/pull/295
- feat(route53resolver): implement AWS Route 53 Resolver service by @sivchari in https://github.com/sivchari/awsim/pull/296
- feat(securitylake): implement AWS Security Lake service by @sivchari in https://github.com/sivchari/awsim/pull/297
- feat(finspace): implement AWS FinSpace service by @sivchari in https://github.com/sivchari/awsim/pull/298
- feat(comprehend): implement AWS Comprehend service by @sivchari in https://github.com/sivchari/awsim/pull/299
- feat(resiliencehub): implement AWS Resilience Hub service by @sivchari in https://github.com/sivchari/awsim/pull/300
- feat(ce): implement AWS Cost Explorer service by @sivchari in https://github.com/sivchari/awsim/pull/301
- feat(rekognition): implement AWS Rekognition service by @sivchari in https://github.com/sivchari/awsim/pull/302
- docs: update README with current service list and examples by @sivchari in https://github.com/sivchari/awsim/pull/303
- Release v0.2.0 by @sivchari in https://github.com/sivchari/awsim/pull/304

## [v0.1.3](https://github.com/sivchari/awsim/compare/v0.1.2...v0.1.3) - 2026-02-26
- fix(kinesis): use NextToken parameter in ListStreams by @sivchari in https://github.com/sivchari/awsim/pull/273
- fix(eks): fix nextToken handling and add maxResults validation by @sivchari in https://github.com/sivchari/awsim/pull/274
- fix(globalaccelerator): implement pagination in ListAccelerators by @sivchari in https://github.com/sivchari/awsim/pull/275
- fix(route53): add pagination support to ListHostedZones by @sivchari in https://github.com/sivchari/awsim/pull/276
- fix(appsync): read nextToken and maxResults parameters in ListGraphqlApis by @sivchari in https://github.com/sivchari/awsim/pull/277
- fix(eventbridge): add ManagedBy field to ListEventBuses response by @sivchari in https://github.com/sivchari/awsim/pull/278
- fix(cloudfront): add missing fields to ListDistributions response by @sivchari in https://github.com/sivchari/awsim/pull/279
- fix(organizations): add State field to ListAccounts response by @sivchari in https://github.com/sivchari/awsim/pull/280
- fix(ecs): add maxResults validation to ListClusters by @sivchari in https://github.com/sivchari/awsim/pull/281
- fix(ecr): implement sorting and pagination in ListImages by @sivchari in https://github.com/sivchari/awsim/pull/282
- fix(secretsmanager): add Type field and rotation metadata to ListSecrets response by @sivchari in https://github.com/sivchari/awsim/pull/283
- fix(cognito): populate LambdaConfig in DescribeUserPool response by @sivchari in https://github.com/sivchari/awsim/pull/284
- fix(acm): add missing fields to ListCertificates response by @sivchari in https://github.com/sivchari/awsim/pull/285
- fix(cloudwatch): add OwningAccounts field to ListMetrics response by @sivchari in https://github.com/sivchari/awsim/pull/286
- fix(s3tables): add missing fields and fix pagination by @sivchari in https://github.com/sivchari/awsim/pull/287
- refactor(test): migrate from testify to golden and separate test dependencies by @sivchari in https://github.com/sivchari/awsim/pull/289

## [v0.1.2](https://github.com/sivchari/awsim/compare/v0.1.1...v0.1.2) - 2026-02-26

## [v0.1.1](https://github.com/sivchari/awsim/compare/v0.1.0...v0.1.1) - 2026-02-26
- fix(s3): align ListBuckets response format with AWS API by @sivchari in https://github.com/sivchari/awsim/pull/255

## [v0.1.0](https://github.com/sivchari/awsim/compare/v0.0.2...v0.1.0) - 2026-02-25
- feat(rds): add RDS service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/233
- feat(elasticache): add ElastiCache service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/235
- feat(route53): add Route 53 service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/237
- feat(ec2): add VPC service implementation by @sivchari in https://github.com/sivchari/awsim/pull/238
- feat(elbv2): add ELB v2 service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/239
- feat(cloudformation): add CloudFormation service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/240
- feat(cloudtrail): add CloudTrail service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/241
- feat(configservice): add AWS Config service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/242
- feat(gamelift): add GameLift service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/243
- feat(servicequotas): add Service Quotas service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/244
- feat(organizations): add Organizations service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/245
- feat(forecast): add Forecast service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/246
- feat(pipes): add EventBridge Pipes service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/247
- feat(emrserverless): add EMR Serverless service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/248
- feat(appmesh): add App Mesh service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/249
- chore: bump version to v0.1.0 by @sivchari in https://github.com/sivchari/awsim/pull/250
- feat(scheduler): add EventBridge Scheduler service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/251
- feat(dlm): add DLM service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/252
- feat(sagemaker): add SageMaker service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/253

## [v0.0.2](https://github.com/sivchari/awsim/compare/v0.0.1...v0.0.2) - 2026-02-17
- feat(eks): add basic EKS service implementation by @sivchari in https://github.com/sivchari/awsim/pull/207
- feat(ecs): add basic ECS service implementation by @sivchari in https://github.com/sivchari/awsim/pull/206
- feat(iam): add IAM service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/209
- feat(s3tables): add S3 Tables service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/210
- feat(athena): add Athena service implementation by @sivchari in https://github.com/sivchari/awsim/pull/211
- feat(codeconnections): add CodeConnections service implementation by @sivchari in https://github.com/sivchari/awsim/pull/212
- feat(kms): add KMS service implementation by @sivchari in https://github.com/sivchari/awsim/pull/213
- feat(globalaccelerator): add Global Accelerator service implementation by @sivchari in https://github.com/sivchari/awsim/pull/214
- feat(cognito): add Cognito Identity Provider service implementation by @sivchari in https://github.com/sivchari/awsim/pull/215
- feat(eventbridge): add EventBridge service implementation by @sivchari in https://github.com/sivchari/awsim/pull/216
- feat(kinesis): add basic Kinesis service implementation by @sivchari in https://github.com/sivchari/awsim/pull/217
- feat(sfn): add basic Step Functions service implementation by @sivchari in https://github.com/sivchari/awsim/pull/218
- feat(ecr): add basic ECR service implementation by @sivchari in https://github.com/sivchari/awsim/pull/219
- feat(apigateway): add API Gateway service implementation by @sivchari in https://github.com/sivchari/awsim/pull/220
- feat(sesv2): add basic SES v2 service implementation by @sivchari in https://github.com/sivchari/awsim/pull/221
- feat(acm): add basic ACM service implementation by @sivchari in https://github.com/sivchari/awsim/pull/222
- feat(glue): add basic Glue service implementation by @sivchari in https://github.com/sivchari/awsim/pull/223
- feat(xray): add X-Ray service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/224
- fix(lint): resolve lint issues in ACM and X-Ray services by @sivchari in https://github.com/sivchari/awsim/pull/228
- fix(acm,glue): use JSONProtocolService interface by @sivchari in https://github.com/sivchari/awsim/pull/229
- fix: ACM and Glue JSON protocol service conflict by @sivchari in https://github.com/sivchari/awsim/pull/230
- feat(batch): add Batch service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/226
- feat(cloudfront): add CloudFront service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/231
- feat(cloudwatch): add CloudWatch metrics service implementation by @sivchari in https://github.com/sivchari/awsim/pull/232
- feat(firehose): add Data Firehose service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/227
- feat(appsync): add AppSync service basic implementation by @sivchari in https://github.com/sivchari/awsim/pull/225

## [v0.0.1](https://github.com/sivchari/awsim/commits/v0.0.1) - 2026-02-06
- feat: add project foundation with base interfaces by @sivchari in https://github.com/sivchari/awsim/pull/186
- refactor(service): add auto-registration via init() by @sivchari in https://github.com/sivchari/awsim/pull/187
- feat(s3): add basic S3 service implementation by @sivchari in https://github.com/sivchari/awsim/pull/188
- feat(sqs): add basic SQS service implementation by @sivchari in https://github.com/sivchari/awsim/pull/189
- feat(dynamodb): add basic DynamoDB service implementation by @sivchari in https://github.com/sivchari/awsim/pull/191
- fix(ci): use integration tests for coverage reporting by @sivchari in https://github.com/sivchari/awsim/pull/192
- feat(secretsmanager): add basic Secrets Manager service implementation by @sivchari in https://github.com/sivchari/awsim/pull/193
- feat(sns): add basic SNS service implementation by @sivchari in https://github.com/sivchari/awsim/pull/194
- feat(lambda): add basic Lambda service implementation by @sivchari in https://github.com/sivchari/awsim/pull/195
- feat(lambda): add HTTP proxy support for Lambda invocations by @sivchari in https://github.com/sivchari/awsim/pull/196
- feat(sqs): add FIFO queue support by @sivchari in https://github.com/sivchari/awsim/pull/198
- feat(s3): add presigned URL support by @sivchari in https://github.com/sivchari/awsim/pull/200
- feat(s3): add versioning support by @sivchari in https://github.com/sivchari/awsim/pull/201
- feat(lambda): add EventSourceMapping API support by @sivchari in https://github.com/sivchari/awsim/pull/202
- feat(ssm): add SSM Parameter Store support by @sivchari in https://github.com/sivchari/awsim/pull/203
- feat(s3): add multipart upload support by @sivchari in https://github.com/sivchari/awsim/pull/199
- feat(cloudwatchlogs): add CloudWatch Logs service implementation by @sivchari in https://github.com/sivchari/awsim/pull/204
- feat(ec2): implement EC2 service with basic instance operations by @sivchari in https://github.com/sivchari/awsim/pull/205

<p align="center">
  <img src="assets/awsim.png" alt="awsim logo" width="480">
  <br><br>
  <a href="https://go.dev/"><img src="https://img.shields.io/github/go-mod/go-version/sivchari/awsim" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/sivchari/awsim" alt="License"></a>
  <a href="https://github.com/sivchari/awsim/releases"><img src="https://img.shields.io/github/v/release/sivchari/awsim" alt="Release"></a>
  <a href="https://github.com/sivchari/awsim/actions/workflows/lint.yaml"><img src="https://github.com/sivchari/awsim/actions/workflows/lint.yaml/badge.svg" alt="Lint"></a>
  <a href="https://github.com/sivchari/awsim/actions/workflows/integration-test.yaml"><img src="https://github.com/sivchari/awsim/actions/workflows/integration-test.yaml/badge.svg" alt="Integration Tests"></a>
</p>

<p align="center">A lightweight AWS service emulator written in Go.<br>Designed for CI/CD environments where authentication-free local AWS testing is needed.</p>

## Features

- **No authentication required** - Perfect for CI environments
- **Single binary** - Easy to distribute and deploy
- **Docker support** - Run as a container
- **Lightweight** - Fast startup, minimal resource usage
- **AWS SDK v2 compatible** - Works seamlessly with Go AWS SDK v2

## Supported Services (69 services)

### Storage
| Service | Description |
|---------|-------------|
| S3 | Object storage |
| S3 Control | S3 account-level operations |
| S3 Tables | S3 table buckets |
| DynamoDB | NoSQL database |
| ElastiCache | In-memory caching |
| MemoryDB | Redis-compatible database |
| Glacier | Archive storage |
| EBS | Block storage |

### Compute
| Service | Description |
|---------|-------------|
| Lambda | Serverless functions |
| Batch | Batch computing |
| EC2 | Virtual machines |
| Elastic Beanstalk | Application deployment |

### Container
| Service | Description |
|---------|-------------|
| ECS | Container orchestration |
| ECR | Container registry |
| EKS | Kubernetes service |

### Database
| Service | Description |
|---------|-------------|
| RDS | Relational database service |

### Messaging & Integration
| Service | Description |
|---------|-------------|
| SQS | Message queuing |
| SNS | Pub/Sub messaging |
| EventBridge | Event bus |
| Kinesis | Real-time streaming |
| Firehose | Data delivery |
| MQ | Message broker (ActiveMQ/RabbitMQ) |
| Pipes | Event-driven integration |
| MSK (Kafka) | Managed streaming for Kafka |

### Security & Identity
| Service | Description |
|---------|-------------|
| IAM | Identity and access management |
| KMS | Key management |
| Secrets Manager | Secret storage |
| ACM | Certificate management |
| Cognito | User authentication |
| Security Lake | Security data lake |
| STS | Security token service |

### Monitoring & Logging
| Service | Description |
|---------|-------------|
| CloudWatch | Metrics and alarms |
| CloudWatch Logs | Log management |
| X-Ray | Distributed tracing |
| CloudTrail | API audit logging |

### Networking & Content Delivery
| Service | Description |
|---------|-------------|
| CloudFront | CDN |
| Global Accelerator | Network acceleration |
| API Gateway | API management |
| Route 53 | DNS service |
| Route 53 Resolver | DNS resolver |
| ELBv2 | Load balancing |
| App Mesh | Service mesh |

### Application Integration
| Service | Description |
|---------|-------------|
| Step Functions | Workflow orchestration |
| AppSync | GraphQL API |
| SES v2 | Email service |
| Scheduler | Task scheduling |
| Amplify | Full-stack application hosting |

### Management & Configuration
| Service | Description |
|---------|-------------|
| SSM | Systems Manager |
| Config | Resource configuration |
| CloudFormation | Infrastructure as code |
| Organizations | Multi-account management |
| Service Quotas | Service limit management |
| CodeConnections | Source code connections |
| Backup | Centralized backup service |

### Analytics & ML
| Service | Description |
|---------|-------------|
| Athena | SQL query service |
| Glue | ETL service |
| Comprehend | NLP service |
| Rekognition | Image/video analysis |
| SageMaker | Machine learning |
| Forecast | Time-series forecasting |
| Data Exchange | Data marketplace |
| Entity Resolution | Entity matching |

### Other Services
| Service | Description |
|---------|-------------|
| Cost Explorer | Cost analysis |
| DLM | Data lifecycle manager |
| Directory Service | Microsoft AD |
| EMR Serverless | Big data processing |
| FinSpace | Financial data management |
| GameLift | Game server hosting |
| Resilience Hub | Application resilience |

## Quick Start

### Docker (Recommended)

```bash
docker run -p 4566:4566 ghcr.io/sivchari/awsim:latest
```

### Binary

```bash
# Build
make build

# Run
./bin/awsim
```

### Docker Compose

```yaml
services:
  awsim:
    image: ghcr.io/sivchari/awsim:latest
    ports:
      - "4566:4566"
```

## Usage Examples

### S3

```go
package main

import (
    "context"
    "strings"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
    cfg, _ := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("us-east-1"),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
    )

    client := s3.NewFromConfig(cfg, func(o *s3.Options) {
        o.BaseEndpoint = aws.String("http://localhost:4566")
        o.UsePathStyle = true
    })

    // Create bucket
    client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
        Bucket: aws.String("my-bucket"),
    })

    // Put object
    client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket: aws.String("my-bucket"),
        Key:    aws.String("hello.txt"),
        Body:   strings.NewReader("Hello, World!"),
    })
}
```

### SQS

```go
package main

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
    cfg, _ := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("us-east-1"),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
    )

    client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
        o.BaseEndpoint = aws.String("http://localhost:4566")
    })

    // Create queue
    result, _ := client.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
        QueueName: aws.String("my-queue"),
    })

    // Send message
    client.SendMessage(context.TODO(), &sqs.SendMessageInput{
        QueueUrl:    result.QueueUrl,
        MessageBody: aws.String("Hello from SQS!"),
    })

    // Receive message
    messages, _ := client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
        QueueUrl: result.QueueUrl,
    })

    for _, msg := range messages.Messages {
        fmt.Println(*msg.Body)
    }
}
```

### DynamoDB

```go
package main

import (
    "context"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func main() {
    cfg, _ := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("us-east-1"),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
    )

    client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
        o.BaseEndpoint = aws.String("http://localhost:4566")
    })

    // Create table
    client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
        TableName: aws.String("users"),
        KeySchema: []types.KeySchemaElement{
            {AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
        },
        AttributeDefinitions: []types.AttributeDefinition{
            {AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
        },
        BillingMode: types.BillingModePayPerRequest,
    })

    // Put item
    client.PutItem(context.TODO(), &dynamodb.PutItemInput{
        TableName: aws.String("users"),
        Item: map[string]types.AttributeValue{
            "id":   &types.AttributeValueMemberS{Value: "user-1"},
            "name": &types.AttributeValueMemberS{Value: "Alice"},
        },
    })
}
```

### Secrets Manager

```go
package main

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func main() {
    cfg, _ := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("us-east-1"),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
    )

    client := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
        o.BaseEndpoint = aws.String("http://localhost:4566")
    })

    // Create secret
    client.CreateSecret(context.TODO(), &secretsmanager.CreateSecretInput{
        Name:         aws.String("my-secret"),
        SecretString: aws.String(`{"username":"admin","password":"secret123"}`),
    })

    // Get secret
    result, _ := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
        SecretId: aws.String("my-secret"),
    })

    fmt.Println(*result.SecretString)
}
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `AWSIM_HOST` | `0.0.0.0` | Server bind address |
| `AWSIM_PORT` | `4566` | Server port |
| `AWSIM_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

## Development

```bash
# Run tests
make test

# Run integration tests
make test-integration

# Lint
make lint

# Build
make build
```

## Contributing

Contributions are welcome! Please see the issues for planned features and improvements.

## License

MIT License

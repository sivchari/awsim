# awsim

A lightweight AWS service emulator written in Go. Designed for CI/CD environments where authentication-free local AWS testing is needed.

## Features

- **No authentication required** - Perfect for CI environments
- **Single binary** - Easy to distribute and deploy
- **Docker support** - Run as a container
- **Lightweight** - Fast startup, minimal resource usage

## Supported Services

### Phase 1 (Current)
- [ ] S3 - Object storage
- [ ] SQS - Message queuing
- [ ] DynamoDB - NoSQL database
- [ ] Secrets Manager - Secret storage
- [ ] SSM Parameter Store - Configuration management

### Phase 2 (Planned)
- [ ] SNS - Pub/Sub messaging
- [ ] CloudWatch Logs - Logging
- [ ] IAM - Identity management
- [ ] Lambda - Serverless functions

### Phase 3 (Future)
- [ ] API Gateway
- [ ] EventBridge
- [ ] Step Functions
- [ ] Kinesis
- [ ] CloudFormation

## Quick Start

### Binary

```bash
# Build
make build

# Run
./bin/awsim
```

### Docker

```bash
# Build image
make docker

# Run container
docker run -p 4566:4566 awsim:0.1.0
```

### Usage with AWS SDK

```go
package main

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
    cfg, _ := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("us-east-1"),
    )

    client := s3.NewFromConfig(cfg, func(o *s3.Options) {
        o.BaseEndpoint = aws.String("http://localhost:4566")
        o.UsePathStyle = true
    })

    // Use client as normal
}
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `AWSIM_HOST` | `0.0.0.0` | Server bind address |
| `AWSIM_PORT` | `4566` | Server port |
| `AWSIM_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `AWSIM_STORAGE` | `memory` | Storage backend (memory, file) |

## Development

```bash
# Run tests
make test

# Run with coverage
make test-cover

# Lint
make lint

# Format
make fmt
```

## License

MIT License

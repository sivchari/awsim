# Contributing to awsim

Thank you for your interest in contributing to awsim!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/awsim.git`
3. Create a feature branch: `git checkout -b feat/your-feature`
4. Make your changes
5. Run tests: `make test`
6. Run linter: `make lint`
7. Commit your changes
8. Push to your fork and submit a pull request

## Development Setup

```bash
# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run integration tests (requires Docker)
make test-integration

# Run linter
make lint
```

## Project Structure

```
awsim/
├── cmd/awsim/          # Application entry point
├── internal/
│   ├── server/         # HTTP server and routing
│   ├── service/        # AWS service implementations
│   │   ├── s3/         # S3 service
│   │   ├── sqs/        # SQS service
│   │   └── ...         # Other services
│   └── errors/         # Error definitions
└── test/
    └── integration/    # Integration tests
```

## Adding a New Service

1. Create a new directory under `internal/service/`
2. Implement the following files:
   - `service.go` - Service registration and routing
   - `handlers.go` - HTTP handlers for API operations
   - `types.go` - Request/response types
   - `storage.go` - In-memory storage (if needed)
3. Register the service in `cmd/awsim/main.go`
4. Add integration tests in `test/integration/`
5. Update README.md with the new service

## Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting
- All exported functions must have documentation comments
- Error messages should be lowercase
- Use `context.Context` for all operations

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test changes
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

Examples:
```
feat(s3): add PutObject operation
fix(sqs): correct message visibility timeout
docs: update README with new services
```

## Pull Request Guidelines

1. **One PR per feature/fix** - Keep changes focused
2. **Write tests** - All new features must have integration tests
3. **Update documentation** - Update README if adding new services
4. **Pass CI** - All tests and lints must pass
5. **Reference issues** - Link related issues with `Closes #123`

## Testing

### Unit Tests
```bash
make test
```

### Integration Tests
Integration tests require the awsim server running:
```bash
# Start awsim with Docker
docker compose up -d

# Run integration tests
make test-integration

# Stop awsim
docker compose down
```

### Golden File Tests
We use golden file testing for API responses:
```bash
# Update golden files (only when intentionally changing responses)
cd test && GOLDEN_UPDATE=true go test -tags=integration -run TestServiceName ./integration/...
```

## Questions?

Feel free to open an issue for any questions or discussions.

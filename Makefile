.PHONY: build run test test-integration clean docker lint lint-fix fmt fmt-diff

BINARY_NAME=awsim
VERSION?=$(shell grep 'const Version' version.go | cut -d'"' -f2)
BUILD_DIR=bin
GOLANGCI_LINT=go tool -modfile tools/go.mod golangci-lint

# Build
build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/awsim

run:
	go run ./cmd/awsim

# Test
test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.out -coverpkg=./... ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration:
	go test -v -tags=integration ./test/integration/...

# Lint
lint:
	$(GOLANGCI_LINT) run ./...

lint-fix:
	$(GOLANGCI_LINT) run --fix ./...

fmt:
	$(GOLANGCI_LINT) fmt ./...

fmt-diff:
	$(GOLANGCI_LINT) fmt ./... --diff

# Docker
docker:
	docker build -t awsim:$(VERSION) -f docker/Dockerfile .

docker-run:
	docker run -p 4566:4566 awsim:$(VERSION)

compose-up:
	docker compose up -d

compose-down:
	docker compose down

# Clean
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Tools
tools:
	cd tools && go mod tidy

# Go mod
mod:
	go mod tidy

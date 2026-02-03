FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /awsim ./cmd/awsim

# Final image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /awsim /usr/local/bin/awsim

EXPOSE 4566

ENTRYPOINT ["awsim"]
CMD ["--host", "0.0.0.0", "--port", "4566"]

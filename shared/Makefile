.PHONY: generate tidy lint test

# Generate protobuf Go code
generate:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go-grpc_out=. proto/*.proto

# Download dependencies
tidy:
	go mod tidy

# Run linter
lint:
	golangci-lint run ./...

# Run tests
test:
	go test ./...

# Build all packages
build:
	go build ./...

# Clean
clean:
	find . -name "*.pb.go" -delete
	find . -name "*_test" -type d -exec rm -rf {} + 2>/dev/null || true

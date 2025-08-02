# Makefile for Racing Service

.PHONY: help build test lint fmt vet clean proto coverage

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build the racing service"
	@echo "  test      - Run all tests"
	@echo "  coverage  - Run tests with coverage report"
	@echo "  lint      - Run linting tools"
	@echo "  fmt       - Format code"
	@echo "  vet       - Run go vet"
	@echo "  proto     - Generate protobuf code"
	@echo "  clean     - Clean build artifacts"
	@echo "  check     - Run all quality checks (fmt, vet, lint, test)"

# Build the service
build:
	go build -o racing ./main.go

# Run all tests
test:
	go test -v ./...

# Run tests with coverage
coverage:
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	go fmt ./...
	@echo "Code formatted"

# Run go vet
vet:
	go vet ./...
	@echo "go vet passed"

# Run linting (requires golint)
lint:
	@which golint > /dev/null || (echo "Installing golint..." && go install golang.org/x/lint/golint@latest)
	golint ./...
	@echo "Linting completed"

# Generate protobuf code
proto:
	go generate ./...
	@echo "Protobuf code generated"

# Clean build artifacts
clean:
	rm -f racing
	rm -f coverage.out
	rm -f coverage.html
	@echo "Clean completed"

# Run all quality checks
check: fmt vet lint test
	@echo "All quality checks passed!"

# Install dependencies for testing and development
deps:
	go mod tidy
	go install golang.org/x/lint/golint@latest
	go get github.com/stretchr/testify/assert
	go get github.com/stretchr/testify/require
	go get github.com/stretchr/testify/mock
	@echo "Dependencies installed"

# Run the service
run: build
	./racing

# Development setup
dev-setup: deps proto check
	@echo "Development environment setup completed!"
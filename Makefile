.PHONY: all help fmt vet lint deps-update install-tools

# Default target
all: help

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  make help          - Display this help message"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Run go vet"
	@echo "  make lint          - Run linter"
	@echo "  make deps-update   - Update dependencies"
	@echo "  make install-tools - Install development tools"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

## lint: Run linter
lint: fmt vet
	@echo "Running linter..."
	golangci-lint run --fix

## deps-update: Update dependencies
deps-update:
	@echo "Updating dependencies..."
	go get -u github.com/natefinch/lumberjack@latest
	go get -u github.com/rs/zerolog@latest
	go mod tidy

tools-install:
	@echo "Installing development tools..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin)
	@echo "Development tools installed"

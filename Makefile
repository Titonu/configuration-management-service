.PHONY: all build run test test-integration test-unit clean lint fmt help docker-build docker-run docker-clean docker-compose-up docker-compose-down docker-compose-dev

# Variables
APP_NAME = config-service
MAIN_PATH = ./cmd/server
BUILD_DIR = ./build
DOCKER_IMAGE = durianpay/config-service:latest

# Go commands
GO = go
GORUN = $(GO) run
GOBUILD = $(GO) build
GOTEST = $(GO) test
GOCLEAN = $(GO) clean
GOFMT = $(GO) fmt
GOLINT = golangci-lint

# Default target
all: clean build

# Help target
help:
	@echo "Available targets:"
	@echo "  all              - Clean and build the application"
	@echo "  build            - Build the application"
	@echo "  run              - Run the application"
	@echo "  test             - Run all tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-unit        - Run unit tests"
	@echo "  clean            - Clean build artifacts"
	@echo "  lint             - Run linters"
	@echo "  fmt              - Format code"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run application in Docker"
	@echo "  docker-clean     - Clean Docker resources"
	@echo "  docker-compose-up    - Start services using docker-compose"
	@echo "  docker-compose-down  - Stop services using docker-compose"
	@echo "  docker-compose-dev   - Start development services with hot reload"
	@echo "  help             - Show this help message"

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	$(GORUN) $(MAIN_PATH)/main.go

# Run all tests
test:
	@echo "Running all tests..."
	$(GOTEST) -v ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v ./tests/integration

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v ./internal/...

# Generate test coverage report
test-coverage:
	@echo "Generating test coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run linters
lint:
	@echo "Running linters..."
	$(GOLINT) run

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

docker-clean:
	@echo "Cleaning Docker resources..."
	docker rmi $(DOCKER_IMAGE) || true

# Docker Compose targets
docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down:
	@echo "Stopping services with docker-compose..."
	docker-compose down

docker-compose-dev:
	@echo "Starting development services with hot reload..."
	docker-compose --profile dev up -d

# App name
APP_NAME ?= go-api

# Timezone
TZ ?= Asia/Shanghai

# Docker image name
IMAGE_NAME ?= $(APP_NAME):latest

# Configuration directory
CONFIG_DIR ?= $(shell pwd)/bin/configs

# Go build flags
GO_FLAGS = -ldflags="-s -w"

# Run environment
RUN_ENV ?= local

# Targets
.PHONY: all test build run docker-build docker-run clean

# Default target that includes formatting, linting, testing, and building
all: fmt test build

# Format the source code
fmt:
	@echo "Running gofmt..."
	@gofmt -w .  # Format all Go files in the current directory
	@echo "Running goimports..."
	@goimports -w .  # Run goimports to organize imports

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...  # Run tests with verbose output

# Build the executable
build: fmt
	@echo "Building binary..."
	@mkdir -p ./bin  # Ensure the bin directory exists
	@go build $(GO_FLAGS) -o ./bin/$(APP_NAME) ./main.go  # Build the Go binary

# Run the application
run:
	@echo "Running application..."
	@./bin/$(APP_NAME)  # Run the compiled binary

# Build the Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build --build-arg TZ=$(TZ) -t $(IMAGE_NAME) .

# Run the Docker container
docker-run: docker-clean
	@echo "Running Docker container..."
	@docker run -d --name $(APP_NAME) \
		-p 8080:8080 \
		-it \
		-v $(CONFIG_DIR):/bin/configs \
		-e APP_NAME=$(APP_NAME) \
		-e RUN_ENV=$(RUN_ENV) \
		--restart always \
		$(IMAGE_NAME)

# Stop and remove existing Docker container with the same name
docker-clean:
	@echo "Stopping and removing existing Docker container..."
	@docker stop $(APP_NAME) 2>/dev/null || true
	@docker rm -f $(APP_NAME) 2>/dev/null || true

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf ./bin/$(APP_NAME)  # Remove the bin directory

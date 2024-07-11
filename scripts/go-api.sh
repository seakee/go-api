#!/bin/sh

# Default values for environment variables
APP_NAME=${APP_NAME:-go-api}
IMAGE_NAME=${IMAGE_NAME:-$APP_NAME:latest}
CONFIG_DIR=${CONFIG_DIR:-$(pwd)/bin/configs}
TZ=${TZ:-Asia/Shanghai}
RUN_ENV=${RUN_ENV:-local}

# Default target that includes formatting, linting, testing, and building
all() {
  fmt
  build
}

# Format the source code
fmt() {
  echo "Running gofmt..."
  gofmt -w .  # Format all Go files in the current directory
  echo "Running goimports..."
  goimports -w .  # Run goimports to organize imports
}

# Build the executable
build() {
  echo "Building binary..."
  mkdir -p ./bin  # Ensure the bin directory exists
  go build -ldflags="-s -w" -o ./bin/$APP_NAME ./main.go  # Build the Go binary
}

# Run the application
run() {
  echo "Running application..."
  RUN_ENV=$RUN_ENV ./bin/$APP_NAME  # Run the compiled binary
}

# Build the Docker image
docker_build() {
  echo "Building Docker image..."
  docker build -t $IMAGE_NAME .
  docker build --build-arg TZ=$TZ -t $IMAGE_NAME .
}

# Run the Docker container
docker_run() {
  docker_clean
  echo "Running Docker container..."
  docker run -d --name $APP_NAME \
    -p 8080:8080 \
    -it \
    -v $CONFIG_DIR:/bin/configs \
    -e APP_NAME=$APP_NAME \
    -e RUN_ENV=$RUN_ENV \
    --restart always \
    $IMAGE_NAME
}

# Stop and remove existing Docker container with the same name
docker_clean() {
  echo "Stopping and removing existing Docker container..."
  docker stop $APP_NAME 2>/dev/null || true
  docker rm -f $APP_NAME 2>/dev/null || true
}

# Clean up build artifacts
clean() {
  echo "Cleaning up..."
  rm -rf ./bin/$APP_NAME  # Remove the bin directory
}

# Parse command line arguments and call corresponding functions
case "$1" in
  all)
    all
    ;;
  fmt)
    fmt
    ;;
  build)
    build
    ;;
  run)
    run
    ;;
  docker-build)
    docker_build
    ;;
  docker-run)
    docker_run
    ;;
  docker-clean)
    docker_clean
    ;;
  clean)
    clean
    ;;
  *)
    echo "Usage: $0 {all|fmt|build|run|docker-build|docker-run|docker-clean|clean}"
    exit 1
    ;;
esac


#!/bin/bash

# Blog System Deployment Script
# This script handles the deployment of the blog system

set -e

# Configuration
APP_NAME="blog-system"
BUILD_DIR="./bin"
BINARY_NAME="blog-server"
LOG_DIR="./logs"
PID_FILE="./blog.pid"
CONFIG_FILE="./configs/config.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    print_status "Go version: $(go version)"
}

# Create necessary directories
create_directories() {
    print_status "Creating necessary directories..."
    mkdir -p "$BUILD_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "./uploads"
    mkdir -p "./configs"
}

# Build the application
build_app() {
    print_status "Building $APP_NAME..."

    # Get dependencies
    go mod tidy

    # Build the binary
    CGO_ENABLED=0 GOOS=linux go build \
        -ldflags="-w -s" \
        -o "$BUILD_DIR/$BINARY_NAME" \
        ./cmd/server/main.go

    if [ $? -eq 0 ]; then
        print_status "Build successful: $BUILD_DIR/$BINARY_NAME"
    else
        print_error "Build failed"
        exit 1
    fi
}

# Run tests
run_tests() {
    print_status "Running tests..."
    go test ./... -v

    if [ $? -eq 0 ]; then
        print_status "All tests passed"
    else
        print_error "Tests failed"
        exit 1
    fi
}

# Stop existing instance
stop_app() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p $PID > /dev/null; then
            print_status "Stopping existing instance (PID: $PID)..."
            kill $PID
            sleep 2

            # Force kill if still running
            if ps -p $PID > /dev/null; then
                print_warning "Force killing process..."
                kill -9 $PID
            fi
        fi
        rm -f "$PID_FILE"
    fi
}

# Start the application
start_app() {
    print_status "Starting $APP_NAME..."

    # Check if config file exists
    if [ ! -f "$CONFIG_FILE" ]; then
        print_warning "Config file not found: $CONFIG_FILE"
        print_status "Creating default config..."
        cat > "$CONFIG_FILE" << EOF
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

database:
  driver: "sqlite"
  dsn: "blog.db"

jwt:
  secret: "your-jwt-secret-key"
  expires_in: 24h

logging:
  level: "info"
  file: "logs/app.log"
EOF
    fi

    # Start the application in background
    nohup "$BUILD_DIR/$BINARY_NAME" > "$LOG_DIR/app.log" 2>&1 &
    echo $! > "$PID_FILE"

    print_status "Application started (PID: $(cat $PID_FILE))"
    print_status "Logs: $LOG_DIR/app.log"
}

# Check application status
check_status() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p $PID > /dev/null; then
            print_status "Application is running (PID: $PID)"

            # Test if the server is responding
            sleep 2
            if curl -f http://localhost:8080/health &> /dev/null; then
                print_status "Health check passed"
            else
                print_warning "Health check failed"
            fi
        else
            print_error "Application is not running"
            rm -f "$PID_FILE"
        fi
    else
        print_error "PID file not found"
    fi
}

# Show usage
usage() {
    echo "Usage: $0 {build|test|start|stop|restart|status|deploy}"
    echo ""
    echo "Commands:"
    echo "  build    - Build the application"
    echo "  test     - Run tests"
    echo "  start    - Start the application"
    echo "  stop     - Stop the application"
    echo "  restart  - Restart the application"
    echo "  status   - Check application status"
    echo "  deploy   - Full deployment (test, build, restart)"
}

# Main script logic
case "${1:-}" in
    build)
        check_go
        create_directories
        build_app
        ;;
    test)
        check_go
        run_tests
        ;;
    start)
        create_directories
        start_app
        ;;
    stop)
        stop_app
        ;;
    restart)
        stop_app
        create_directories
        start_app
        ;;
    status)
        check_status
        ;;
    deploy)
        check_go
        create_directories
        run_tests
        build_app
        stop_app
        start_app
        sleep 3
        check_status
        ;;
    *)
        usage
        exit 1
        ;;
esac

print_status "Operation completed successfully"
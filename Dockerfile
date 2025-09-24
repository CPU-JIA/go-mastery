# Go Mastery - Advanced Go Learning Path
# Multi-stage Dockerfile for development and production environments

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o main ./...

# Development stage
FROM golang:1.24-alpine AS development

# Install development tools
RUN apk add --no-cache \
    git \
    curl \
    make \
    bash \
    vim \
    ca-certificates \
    tzdata \
    build-base

# Install Go development tools
RUN go install honnef.co/go/tools/cmd/staticcheck@latest && \
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest && \
    go install golang.org/x/vuln/cmd/govulncheck@latest && \
    go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install github.com/cosmtrek/air@latest

# Set working directory
WORKDIR /app

# Copy application files
COPY . .

# Download dependencies
RUN go mod download && go mod verify

# Expose development port
EXPOSE 8080 9090 6060

# Default command for development
CMD ["air", "-c", ".air.toml"]

# Production stage
FROM alpine:3.19 AS production

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy configuration files
COPY --from=builder /app/configs/ ./configs/

# Set permissions
RUN chown -R appuser:appgroup /root/

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the binary
ENTRYPOINT ["./main"]

# Demo/Tutorial stage - for running examples
FROM golang:1.24-alpine AS demo

# Install dependencies
RUN apk add --no-cache git ca-certificates make bash

WORKDIR /app

# Copy source code
COPY . .

# Download dependencies
RUN go mod download

# Create scripts directory
RUN mkdir -p /app/scripts

# Create demo runner script
RUN cat > /app/scripts/run-demos.sh << 'EOF'
#!/bin/bash

echo "=== Go Mastery Learning Path Demos ==="
echo "Choose a module to run:"
echo "1. Runtime Internals (07-runtime-internals)"
echo "2. Performance Mastery (08-performance-mastery)"
echo "3. System Programming (09-system-programming)"
echo "4. Compiler Toolchain (10-compiler-toolchain)"
echo "5. Massive Systems (11-massive-systems)"
echo "6. Ecosystem Contribution (12-ecosystem-contribution)"
echo "7. Language Design (13-language-design)"
echo "8. Tech Leadership (14-tech-leadership)"
echo "9. Run all modules"
echo "0. Exit"

read -p "Enter your choice (0-9): " choice

case $choice in
    1) echo "Running Runtime Internals..."; cd 07-runtime-internals && go run main.go ;;
    2) echo "Running Performance Mastery..."; cd 08-performance-mastery && go run main.go ;;
    3) echo "Running System Programming..."; cd 09-system-programming && go run main.go ;;
    4) echo "Running Compiler Toolchain..."; cd 10-compiler-toolchain && go run main.go ;;
    5) echo "Running Massive Systems..."; cd 11-massive-systems && go run main.go ;;
    6) echo "Running Ecosystem Contribution..."; cd 12-ecosystem-contribution && go run main.go ;;
    7) echo "Running Language Design..."; cd 13-language-design && go run main.go ;;
    8) echo "Running Tech Leadership..."; cd 14-tech-leadership && go run main.go ;;
    9)
        echo "Running all modules..."
        for dir in */; do
            if [ -f "$dir/main.go" ]; then
                echo "=== Running $dir ==="
                cd "$dir" && go run main.go && cd ..
                echo ""
            fi
        done
        ;;
    0) echo "Goodbye!"; exit 0 ;;
    *) echo "Invalid choice. Please try again." ;;
esac
EOF

RUN chmod +x /app/scripts/run-demos.sh

# Create module-specific run scripts
RUN for i in {7..14}; do \
        module_name=""; \
        case $i in \
            7) module_name="runtime-internals" ;; \
            8) module_name="performance-mastery" ;; \
            9) module_name="system-programming" ;; \
            10) module_name="compiler-toolchain" ;; \
            11) module_name="massive-systems" ;; \
            12) module_name="ecosystem-contribution" ;; \
            13) module_name="language-design" ;; \
            14) module_name="tech-leadership" ;; \
        esac; \
        cat > /app/scripts/run-$module_name.sh << EOF && \
#!/bin/bash
cd /app/${i}-${module_name}
echo "Running Go ${module_name} module..."
go run main.go
EOF
        chmod +x /app/scripts/run-$module_name.sh; \
    done

EXPOSE 8080

CMD ["/app/scripts/run-demos.sh"]

# Testing stage - for CI/CD
FROM golang:1.24-alpine AS testing

# Install testing tools
RUN apk add --no-cache git ca-certificates make bash curl

# Install Go testing tools
RUN go install honnef.co/go/tools/cmd/staticcheck@latest && \
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest && \
    go install golang.org/x/vuln/cmd/govulncheck@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

WORKDIR /app

COPY . .

RUN go mod download

# Run all quality checks
RUN make quality-check

# Performance stage - for benchmarking
FROM golang:1.24-alpine AS performance

RUN apk add --no-cache git ca-certificates make bash

# Install performance tools
RUN go install github.com/google/pprof@latest

WORKDIR /app

COPY . .

RUN go mod download

# Run benchmarks
RUN make bench

# Default to production stage
FROM production
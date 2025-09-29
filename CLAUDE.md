# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a comprehensive Go mastery learning project structured as a sequential learning path from basic syntax to advanced system programming. The repository contains 16 learning modules (00-15) covering everything from Go basics to open-source contribution.

## Architecture

### Multi-Module Structure
- **Root Module**: `go-mastery` - Main workspace module (Go 1.24)
- **Sub-modules**: Several directories have their own go.mod files for isolated learning environments
  - `00-assessment-system/` - Assessment system
  - `05-microservices/*/` - Multiple microservice examples
  - `06-projects/*/` - Various project implementations
  - `15-opensource-contribution/` - Open source projects

### Learning Path Organization
```
00-assessment-system/     # Skill assessment and tracking
01-basics/               # Basic Go syntax and concepts
02-advanced/             # Advanced Go features
03-concurrency/          # Goroutines, channels, and concurrency patterns
04-web/                  # Web development and HTTP services
05-microservices/        # Microservice architecture and patterns
06-projects/             # Real-world project implementations
06.5-performance-fundamentals/  # Performance optimization basics
07-runtime-internals/    # Go runtime deep dive
08-performance-mastery/  # Advanced performance optimization
09-system-programming/   # System-level programming
10-compiler-toolchain/   # Go toolchain and compiler
11-massive-systems/      # Large-scale system design
12-ecosystem-contribution/  # Go ecosystem contribution
13-language-design/      # Programming language design
14-tech-leadership/      # Technical leadership
15-opensource-contribution/  # Open source project work
```

### Common Infrastructure
- `common/` - Shared utilities and demo code
- `scripts/` - Build and quality automation scripts
- `docs/` - Comprehensive documentation system

## Development Commands

### Essential Commands
```bash
# Setup development environment
make setup                 # Install tools and dependencies
make dev-setup            # Setup with pre-commit hooks

# Development workflow
make dev-check            # Quick pre-commit checks (fmt, vet, short tests)
make build                # Build all modules
make test                 # Run all tests
make test-race            # Run tests with race detection

# Code quality
make fmt                  # Format all Go code
make lint                 # Run golangci-lint with comprehensive rules
make vet                  # Run go vet
make quality-check        # Run all quality checks (CI equivalent)

# Coverage and performance
make coverage             # Generate test coverage report
make coverage-open        # Open coverage report in browser
make bench                # Run benchmark tests

# Complete CI pipeline
make ci                   # Run full CI pipeline locally
```

### Platform-Specific
- **Linux/macOS**: Use `make` commands above
- **Windows**: Use `build.ps1 <command>` (PowerShell equivalent of Make)

### Docker Development
```bash
# Development environment with hot reload
docker-compose up go-mastery-dev

# Production simulation
docker-compose up go-mastery-prod

# Full monitoring stack
docker-compose --profile monitoring up

# Performance testing
docker-compose --profile performance up go-mastery-perf
```

## Key Dependencies

### Core Dependencies
- `github.com/gorilla/mux` - HTTP routing
- `github.com/gorilla/websocket` - WebSocket support
- `gorm.io/gorm` - ORM with PostgreSQL/SQLite drivers
- `github.com/go-redis/redis/v8` - Redis client
- `github.com/golang-jwt/jwt/v4` - JWT authentication
- `github.com/segmentio/kafka-go` - Kafka client
- `github.com/hashicorp/consul/api` - Service discovery

### Development Tools
- `golangci-lint` - Comprehensive linting (configured with 40+ linters)
- `gosec` - Security vulnerability scanner
- `staticcheck` - Advanced static analysis
- `govulncheck` - Go vulnerability database scanner

## Quality Standards

### Enforced Quality Gates
- **Test Coverage**: Minimum 75% coverage required
- **Linting**: Zero warnings policy with golangci-lint
- **Security**: No high/medium vulnerabilities allowed
- **Formatting**: Strict gofmt + goimports compliance
- **Go Version**: Supports Go 1.21-1.24, optimized for Go 1.24

### Pre-commit Hooks
The project uses comprehensive pre-commit hooks including:
- Code formatting (gofmt, goimports)
- Static analysis (go vet, staticcheck)
- Security scanning (gosec, govulncheck)
- Test execution and build verification
- Conventional commit message validation

## Testing Strategy

### Test Types
- **Unit Tests**: Standard Go tests in `*_test.go` files
- **Integration Tests**: Use `-tags=integration` build tag
- **Benchmark Tests**: Performance benchmarks with `-bench=.`
- **Race Detection**: All CI runs include race detection

### Running Specific Tests
```bash
# Single module
cd 01-basics/01-hello && go test

# Integration tests only
go test -tags=integration ./...

# Specific package
go test ./05-microservices/...

# With race detection
go test -race ./...
```

## Module-Specific Notes

### Working with Sub-modules
When working in directories with their own `go.mod`:
1. Change to the specific directory first
2. Use module-local commands: `go test ./...`, `go build ./...`
3. Dependencies are managed independently in each module

### Assessment System (00-assessment-system/)
Contains skill evaluation and progress tracking tools. Has its own module for isolated execution.

### Microservices (05-microservices/)
Multiple independent services with their own go.mod files. Each service can be built and tested independently.

### Projects (06-projects/)
Real-world applications demonstrating complete Go systems. Each project is a separate module with its own dependencies.

## Development Environment Options

### Local Development
- Go 1.24 required
- Make or PowerShell for build automation
- Pre-commit hooks recommended for quality assurance

### Docker Development (Recommended)
- Full containerized development environment
- Includes databases (PostgreSQL, Redis)
- Monitoring stack (Prometheus, Grafana, Jaeger)
- Hot reload support for rapid development

### VS Code Integration
Recommended extensions and settings are configured in `.vscode/` for optimal Go development experience.

## Troubleshooting

### Common Issues
- **Module not found**: Ensure you're in the correct directory with go.mod
- **Tool not found**: Run `make install-tools` to install required development tools
- **Test failures**: Check if Docker services are running for integration tests
- **Permission issues**: On Windows/WSL2, ensure proper file permissions

### Quality Check Failures
- Format issues: Run `make fmt`
- Lint warnings: Check `.golangci.yml` configuration
- Security issues: Review gosec findings, may require code changes
- Coverage too low: Add more comprehensive tests

This is a learning-focused project emphasizing code quality, comprehensive testing, and Go best practices across all skill levels.
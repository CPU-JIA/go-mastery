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
  - `common/security/` - **REQUIRED for all file operations** (see Security Best Practices)
  - Provides secure file I/O wrappers to prevent G301/G304 vulnerabilities
  - Path validation utilities to prevent directory traversal attacks
- `scripts/` - Build and quality automation scripts
- `docs/` - Comprehensive documentation system

**Important**: Always import and use `common/security` for file operations instead of `os` package directly.

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

#### Linux/macOS
Use `make` commands as shown above.

#### Windows (PowerShell)
Use `build.ps1 <command>` for all operations:
```powershell
# Setup and development
.\build.ps1 setup                 # Install tools and dependencies
.\build.ps1 dev-setup            # Setup with pre-commit hooks
.\build.ps1 dev-check            # Quick pre-commit checks

# Build and test
.\build.ps1 build                # Build all modules
.\build.ps1 test                 # Run all tests
.\build.ps1 test-race            # Run tests with race detection

# Code quality
.\build.ps1 fmt                  # Format all Go code
.\build.ps1 lint                 # Run golangci-lint
.\build.ps1 quality-check        # Run all quality checks

# Coverage and CI
.\build.ps1 coverage             # Generate test coverage report
.\build.ps1 ci                   # Run full CI pipeline locally
```

**Windows-Specific Notes**:
- Ensure PowerShell execution policy allows scripts: `Set-ExecutionPolicy RemoteSigned -Scope CurrentUser`
- Some tools (like pre-commit) may require WSL2 or Git Bash
- File paths use backslashes in Windows but Go code should use forward slashes or `filepath.Join()`

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

### Linter Configuration (.golangci.yml)
The project uses **40+ linters** for comprehensive code quality enforcement:

**Key Linters**:
- **Security**: `gosec` (no HIGH/MEDIUM vulnerabilities allowed)
- **Static Analysis**: `staticcheck`, `govet`, `errcheck`
- **Code Quality**: `gocyclo` (complexity ≤15), `gocognit` (≤20), `dupl`, `funlen`
- **Style**: `gofmt`, `gofumpt`, `goimports`, `revive`, `stylecheck`
- **Performance**: `prealloc`, `gocritic`
- **Resource Safety**: `bodyclose`, `sqlclosecheck`, `rowserrcheck`

**Exemption Rules**:
- Test files (`*_test.go`) have relaxed requirements for complexity and duplication
- Generated files (`*.pb.go`, `*_generated.go`) are skipped
- Security exemptions require `#nosec` with detailed justification
- Magic numbers like HTTP status codes (200, 404, 500) are allowed

**Running Linters**:
```bash
# Run all linters
make lint

# Run specific linter
golangci-lint run --disable-all --enable=gosec

# Fix auto-fixable issues
golangci-lint run --fix
```

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

**Commit Message Format** (enforced by pre-commit hooks):
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`
- Use emoji prefixes for clarity: `🎨`, `🐛`, `📝`, `✨`, `♻️`, `⚡`, `✅`, `🔒`
- Example: `🔒 fix(security): 修复G304路径遍历漏洞`

**Setup Pre-commit Hooks**:
```bash
make dev-setup  # Installs and configures pre-commit hooks
```

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

# Single test function
go test -run TestFunctionName ./path/to/package

# Multiple packages matching pattern
go test ./03-concurrency/...

# Verbose output with coverage
go test -v -coverprofile=coverage.out ./...

# Skip specific tests
go test -skip TestLongRunning ./...
```

## Security Best Practices

### Critical Security Standards
This project has undergone comprehensive security hardening (63 vulnerabilities fixed in recent commits). All new code MUST follow these security practices:

### File Operations (G301 & G304 Protection)
**NEVER** use standard library file operations directly. Always use `common/security` package:

```go
// ❌ WRONG - Vulnerable to path traversal and insecure permissions
os.WriteFile(userInput, data, 0777)
os.MkdirAll(path, 0777)

// ✅ CORRECT - Use security package
import "go-mastery/common/security"

security.SecureWriteFile(filename, data, &security.SecureFileOptions{
    Mode: security.DefaultFileMode, // 0600
})
security.SecureMkdirAll(path, security.DefaultDirMode) // 0700
```

### Path Validation
Always validate user-provided paths:
```go
// Validate before using any path from external sources
err := security.ValidateSecurePath(userPath, &security.SecurePathOptions{
    AllowAbsolute: false,
    AllowDotDot:   false,
    MaxDepth:      10,
})
```

### Recommended File Permissions
- **Regular files**: `0600` (owner read/write only)
- **Directories**: `0700` (owner full access only)
- **Config files**: `0400` (owner read-only)
- **Executables**: `0700` (owner execute only)
- **NEVER use**: `0777`, `0666`, or any world-writable permissions

### Gosec Compliance
- The project uses comprehensive security scanning with gosec
- All HIGH and MEDIUM vulnerabilities MUST be fixed before merge
- Use `#nosec` comments ONLY when absolutely necessary with detailed justification:
  ```go
  // #nosec G304 -- Path validated via ValidateSecurePath() before use
  file, err := os.Open(validatedPath)
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
- **Format issues**: Run `make fmt`
- **Lint warnings**: Check `.golangci.yml` configuration, run `make lint` for details
- **Security issues**: Review gosec findings with `gosec -fmt=json ./...`, may require code changes
- **Coverage too low**: Add more comprehensive tests, current threshold is 75%
- **G301 (file permissions)**: Use `common/security.SecureWriteFile()` with appropriate `SecureFileMode`
- **G304 (path traversal)**: Use `common/security.ValidateSecurePath()` before file operations

### Debugging Tips
```bash
# Run tests with verbose output
go test -v ./path/to/package

# Run specific test with debugging
go test -v -run TestName ./path/to/package

# Profile tests
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=.

# Check test coverage for specific package
go test -cover ./path/to/package

# Generate detailed coverage HTML
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## Key Design Principles

### Security-First Architecture
- All file operations MUST use `common/security` package
- Zero tolerance for HIGH/MEDIUM security vulnerabilities
- Defense-in-depth: validation at multiple layers
- Principle of least privilege: restrictive file permissions by default (0600/0700)

### Multi-Module Design
- Each learning phase may have independent `go.mod` for isolation
- Root module acts as workspace for common dependencies
- Enables independent versioning and dependency management per module

### Code Quality Standards
- **Coverage**: Minimum 75% test coverage enforced
- **Complexity**: Max cyclomatic complexity 15, cognitive complexity 20
- **Style**: Automated formatting with gofmt + gofumpt + goimports
- **Security**: Continuous scanning with gosec + govulncheck
- **Performance**: Regular benchmarking, race detection on all tests

This is a learning-focused project emphasizing code quality, comprehensive testing, and Go best practices across all skill levels.
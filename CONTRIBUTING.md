# Contributing to Go Mastery Project

Welcome to the Go Mastery project! This guide will help you contribute effectively while maintaining our high code quality standards.

## Table of Contents

- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Quality Standards](#quality-standards)
- [Testing Guidelines](#testing-guidelines)
- [Commit Convention](#commit-convention)
- [Pull Request Process](#pull-request-process)
- [CI/CD Pipeline](#cicd-pipeline)
- [Troubleshooting](#troubleshooting)

## Development Setup

### Prerequisites

- Go 1.21+ (we test against 1.21, 1.22, 1.23, 1.24)
- Git
- Make (optional but recommended)
- Docker (optional, for containerized builds)

### Quick Start

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd go-mastery
   ```

2. **Set up development environment**:
   ```bash
   make setup
   # or manually:
   go mod download
   make install-tools
   ```

3. **Install pre-commit hooks** (recommended):
   ```bash
   pip install pre-commit
   make dev-setup
   ```

4. **Verify setup**:
   ```bash
   make dev-check
   ```

### Development Tools

Our CI/CD pipeline uses the following tools:

- **gofmt**: Code formatting
- **goimports**: Import management
- **go vet**: Static analysis
- **staticcheck**: Advanced linting
- **gosec**: Security analysis
- **govulncheck**: Vulnerability scanning
- **go test**: Testing with race detection

Install all tools with: `make install-tools`

## Development Workflow

### 1. Before You Start

- Create a feature branch: `git checkout -b feature/your-feature-name`
- Ensure your environment is set up: `make dev-check`

### 2. Making Changes

- Follow our [Code Style Guidelines](#code-style-guidelines)
- Write tests for new functionality
- Run quality checks frequently: `make dev-check`

### 3. Before Committing

- Format your code: `make fmt`
- Run full quality checks: `make quality-check`
- Ensure tests pass: `make test`
- Check coverage: `make coverage-check`

### 4. Committing

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Examples:
```bash
git commit -m "feat: add user authentication middleware"
git commit -m "fix(http): handle timeout errors properly"
git commit -m "docs: update API documentation"
git commit -m "test: add integration tests for auth service"
```

### 5. Submitting Changes

- Push your branch: `git push origin feature/your-feature-name`
- Create a Pull Request
- Ensure CI/CD pipeline passes
- Address review feedback

## Quality Standards

### Code Quality Metrics

Our CI/CD pipeline enforces these quality gates:

| Metric | Requirement | Description |
|--------|-------------|-------------|
| **Build** | ✅ Must Pass | All code must compile successfully |
| **Tests** | ✅ Must Pass | All tests must pass with race detection |
| **Coverage** | ≥ 75% | Test coverage threshold |
| **Formatting** | ✅ Must Pass | Code must be properly formatted |
| **Linting** | ✅ Must Pass | Zero warnings from go vet and staticcheck |
| **Security** | ✅ Must Pass | No HIGH/MEDIUM severity vulnerabilities |
| **Dependencies** | ✅ Must Pass | No known vulnerable dependencies |

### Code Style Guidelines

1. **Formatting**: Use `gofmt` and `goimports`
   ```bash
   make fmt
   ```

2. **Naming Conventions**:
   - Use camelCase for unexported functions/variables
   - Use PascalCase for exported functions/variables
   - Use meaningful, descriptive names

3. **Error Handling**:
   ```go
   // Good
   if err != nil {
       return fmt.Errorf("failed to process user %s: %w", userID, err)
   }

   // Bad
   if err != nil {
       panic(err)
   }
   ```

4. **Package Organization**:
   - Each package should have a single, well-defined purpose
   - Keep packages small and focused
   - Use internal packages for implementation details

5. **Comments**:
   - Comment exported functions, types, and constants
   - Explain complex business logic
   - Use complete sentences

## Testing Guidelines

### Test Structure

Follow the table-driven test pattern:

```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name    string
        user    User
        want    bool
        wantErr bool
    }{
        {
            name: "valid user",
            user: User{Name: "John", Email: "john@example.com"},
            want: true,
            wantErr: false,
        },
        {
            name: "invalid email",
            user: User{Name: "John", Email: "invalid"},
            want: false,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ValidateUser(tt.user)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateUser() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ValidateUser() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Categories

1. **Unit Tests**: Test individual functions/methods
   ```bash
   go test ./...
   ```

2. **Integration Tests**: Test component interactions
   ```bash
   go test -tags=integration ./...
   ```

3. **Benchmark Tests**: Performance testing
   ```bash
   make bench
   ```

### Coverage Requirements

- Minimum 75% test coverage
- Focus on critical paths and error conditions
- Use `make coverage` to generate reports

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/) to automate versioning and changelog generation.

### Commit Types

- **feat**: New features
- **fix**: Bug fixes
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks
- **perf**: Performance improvements
- **ci**: CI/CD changes

### Examples

```bash
# Features
git commit -m "feat(auth): add JWT authentication middleware"
git commit -m "feat: implement user registration endpoint"

# Bug fixes
git commit -m "fix(db): handle connection timeouts properly"
git commit -m "fix: resolve race condition in cache"

# Documentation
git commit -m "docs: update API documentation for v2.0"
git commit -m "docs(readme): add setup instructions"

# Tests
git commit -m "test: add integration tests for payment flow"
git commit -m "test(auth): improve test coverage to 85%"
```

## Pull Request Process

### 1. Before Creating PR

- [ ] Branch is up to date with main
- [ ] All quality checks pass locally
- [ ] Tests have been added/updated
- [ ] Documentation has been updated

### 2. PR Requirements

- [ ] **Title**: Use conventional commit format
- [ ] **Description**: Explain what and why
- [ ] **Tests**: Include test results
- [ ] **Breaking Changes**: Document any breaking changes
- [ ] **Checklist**: Complete the PR template checklist

### 3. PR Template

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that causes existing functionality to not work)
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed
- [ ] Coverage threshold met

## Quality Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Code is commented (particularly complex areas)
- [ ] No new warnings introduced
- [ ] Security considerations addressed
```

### 4. Review Process

1. **Automated Checks**: CI/CD pipeline must pass
2. **Code Review**: At least one approval required
3. **Quality Gate**: All quality metrics must be met
4. **Final Check**: Maintainer approval

## CI/CD Pipeline

### Pipeline Stages

Our GitHub Actions workflow includes:

1. **Static Analysis** (parallel across Go versions 1.21-1.24)
   - Code formatting check
   - go vet analysis
   - staticcheck linting

2. **Security Analysis**
   - gosec security scanning
   - Vulnerability assessment
   - SARIF report generation

3. **Build and Test** (parallel across Go versions)
   - Build verification
   - Race condition testing
   - Coverage analysis

4. **Integration Testing**
   - Database integration tests
   - External service tests

5. **Cross-Platform Build**
   - Linux, Windows, macOS
   - AMD64 and ARM64 architectures

6. **Quality Gate Summary**
   - Aggregate all results
   - Enforce quality thresholds

### Quality Gates

The pipeline will fail if:
- Any build fails
- Tests fail or coverage < 75%
- Security vulnerabilities (HIGH/MEDIUM)
- Code formatting issues
- Linting warnings

### Local Pipeline Simulation

Run the complete CI pipeline locally:

```bash
make ci
```

## Makefile Commands

### Essential Commands

```bash
# Setup
make setup              # Set up development environment
make dev-setup          # Setup with pre-commit hooks

# Development
make dev-check          # Quick checks before commit
make fmt               # Format code
make test              # Run tests
make coverage          # Generate coverage report

# Quality
make quality-check     # Run all quality checks
make security          # Security analysis
make vuln-check        # Vulnerability scanning

# CI/CD
make ci                # Run full CI pipeline
make build-release     # Build release binaries
```

### Complete Command List

Run `make help` for all available commands.

## Project Structure

```
go-mastery/
├── .github/
│   └── workflows/
│       └── ci.yml              # CI/CD pipeline
├── 01-basics/                  # Basic Go concepts
├── 02-advanced/                # Advanced features
├── 03-concurrency/             # Concurrency patterns
├── 04-web/                     # Web development
├── 05-microservices/           # Microservices
├── 06-projects/                # Complete projects
├── common/                     # Shared utilities
├── coverage/                   # Coverage reports (generated)
├── dist/                       # Build artifacts (generated)
├── .pre-commit-config.yaml     # Pre-commit hooks
├── Makefile                    # Build automation
├── go.mod                      # Go module definition
├── README.md                   # Project documentation
└── CONTRIBUTING.md             # This file
```

## Troubleshooting

### Common Issues

1. **Build Failures**
   ```bash
   # Clean and rebuild
   make clean
   make build
   ```

2. **Test Failures**
   ```bash
   # Run tests with verbose output
   make test-verbose

   # Run specific test
   go test -v ./path/to/package -run TestName
   ```

3. **Coverage Issues**
   ```bash
   # Generate detailed coverage report
   make coverage
   make coverage-open
   ```

4. **Import Path Issues**
   ```bash
   # Fix imports
   make fmt
   ```

5. **Pre-commit Hook Issues**
   ```bash
   # Reinstall hooks
   pre-commit uninstall
   pre-commit install
   ```

### Getting Help

- Check the [README.md](README.md) for basic setup
- Run `make help` for available commands
- Check existing issues on GitHub
- Create a new issue with detailed description

### Performance Guidelines

1. **Benchmarking**
   ```bash
   make bench
   ```

2. **Profiling**
   ```bash
   go test -cpuprofile=cpu.prof -bench=.
   go tool pprof cpu.prof
   ```

3. **Memory Analysis**
   ```bash
   go test -memprofile=mem.prof -bench=.
   go tool pprof mem.prof
   ```

## Code Review Guidelines

### For Authors

- Keep PRs focused and small
- Write clear commit messages
- Include tests with your changes
- Update documentation as needed
- Respond promptly to review feedback

### For Reviewers

- Be constructive and specific
- Focus on correctness, maintainability, and performance
- Check test coverage and quality
- Verify documentation updates
- Test locally if possible

## Security Guidelines

1. **Input Validation**: Always validate external inputs
2. **Error Handling**: Don't expose internal details in errors
3. **Dependencies**: Keep dependencies updated
4. **Secrets**: Never commit secrets or credentials
5. **Authentication**: Use secure authentication mechanisms

## Performance Considerations

1. **Avoid Premature Optimization**: Profile first
2. **Memory Management**: Be aware of allocations
3. **Concurrency**: Use appropriate synchronization
4. **I/O Operations**: Handle efficiently with proper timeouts
5. **Database**: Use prepared statements and connection pooling

---

Thank you for contributing to the Go Mastery project! Your efforts help make this a better learning resource for the Go community.
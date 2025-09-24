# Go Mastery Project Build Script for Windows
# PowerShell equivalent of Makefile commands

param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

# Configuration
$BinaryName = "go-mastery"
$BuildDir = "dist"
$CoverageDir = "coverage"
$CoverageFile = "$CoverageDir\coverage.out"
$CoverageHtml = "$CoverageDir\coverage.html"
$CoverageThreshold = 75

# Colors
$Green = "Green"
$Yellow = "Yellow"
$Red = "Red"
$Blue = "Blue"
$Cyan = "Cyan"

function Write-ColorOutput {
    param($Message, $Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Show-Help {
    Write-ColorOutput "Go Mastery Project - PowerShell Build Commands" $Blue
    Write-ColorOutput "================================================" $Blue
    Write-ColorOutput ""
    Write-ColorOutput "Usage: .\build.ps1 <command>" $Cyan
    Write-ColorOutput ""
    Write-ColorOutput "Available Commands:" $Cyan
    Write-ColorOutput "  help              - Show this help message"
    Write-ColorOutput "  setup             - Setup development environment"
    Write-ColorOutput "  deps              - Download dependencies"
    Write-ColorOutput "  build             - Build the application"
    Write-ColorOutput "  test              - Run tests"
    Write-ColorOutput "  test-race         - Run tests with race detection"
    Write-ColorOutput "  coverage          - Generate coverage report"
    Write-ColorOutput "  coverage-check    - Check coverage threshold"
    Write-ColorOutput "  fmt               - Format code"
    Write-ColorOutput "  fmt-check         - Check code formatting"
    Write-ColorOutput "  vet               - Run go vet"
    Write-ColorOutput "  security          - Run security analysis"
    Write-ColorOutput "  vuln-check        - Check vulnerabilities"
    Write-ColorOutput "  quality-check     - Run all quality checks"
    Write-ColorOutput "  clean             - Clean build artifacts"
    Write-ColorOutput "  bench             - Run benchmarks"
    Write-ColorOutput "  info              - Show project information"
}

function Install-Tools {
    Write-ColorOutput "Installing development tools..." $Yellow
    go install honnef.co/go/tools/cmd/staticcheck@latest
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    go install golang.org/x/vuln/cmd/govulncheck@latest
    go install golang.org/x/tools/cmd/goimports@latest
    Write-ColorOutput "Development tools installed successfully!" $Green
}

function Setup {
    Write-ColorOutput "Setting up development environment..." $Yellow
    Install-Tools
    Download-Dependencies
    Write-ColorOutput "Development environment setup complete!" $Green
}

function Download-Dependencies {
    Write-ColorOutput "Downloading dependencies..." $Yellow
    go mod download
    go mod verify
    Write-ColorOutput "Dependencies downloaded and verified!" $Green
}

function Build {
    Write-ColorOutput "Building application..." $Yellow
    go build ./...
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "Build completed successfully!" $Green
    } else {
        Write-ColorOutput "Build failed!" $Red
        exit 1
    }
}

function Test {
    Write-ColorOutput "Running tests..." $Yellow
    go test ./...
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "All tests passed!" $Green
    } else {
        Write-ColorOutput "Tests failed!" $Red
        exit 1
    }
}

function Test-Race {
    Write-ColorOutput "Running tests with race detection..." $Yellow
    go test -race ./...
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "Race tests passed!" $Green
    } else {
        Write-ColorOutput "Race tests failed!" $Red
        exit 1
    }
}

function Generate-Coverage {
    Write-ColorOutput "Generating coverage report..." $Yellow
    if (!(Test-Path $CoverageDir)) {
        New-Item -ItemType Directory -Path $CoverageDir | Out-Null
    }
    go test -race -coverprofile=$CoverageFile -covermode=atomic ./...
    go tool cover -html=$CoverageFile -o $CoverageHtml
    Write-ColorOutput "Coverage report generated: $CoverageHtml" $Green
}

function Check-Coverage {
    Write-ColorOutput "Checking coverage threshold..." $Yellow
    Generate-Coverage
    $coverageOutput = go tool cover -func=$CoverageFile | Select-String "total"
    if ($coverageOutput) {
        $coverageStr = $coverageOutput.ToString().Split()[-1].Replace('%', '')
        $coverage = [double]$coverageStr
        Write-ColorOutput "Current coverage: $coverage%" $Cyan
        if ($coverage -lt $CoverageThreshold) {
            Write-ColorOutput "Coverage $coverage% is below threshold $CoverageThreshold%" $Red
            exit 1
        } else {
            Write-ColorOutput "Coverage $coverage% meets threshold $CoverageThreshold%" $Green
        }
    }
}

function Format-Code {
    Write-ColorOutput "Formatting code..." $Yellow
    gofmt -s -w .
    $goimports = Get-Command goimports -ErrorAction SilentlyContinue
    if ($goimports) {
        goimports -w .
    }
    Write-ColorOutput "Code formatted successfully!" $Green
}

function Check-Format {
    Write-ColorOutput "Checking code formatting..." $Yellow
    $unformatted = gofmt -s -l .
    if ($unformatted) {
        Write-ColorOutput "The following files are not properly formatted:" $Red
        Write-ColorOutput $unformatted $Red
        Write-ColorOutput "Run '.\build.ps1 fmt' to fix formatting issues." $Yellow
        exit 1
    } else {
        Write-ColorOutput "All files are properly formatted!" $Green
    }
}

function Run-Vet {
    Write-ColorOutput "Running go vet..." $Yellow
    go vet ./...
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "Go vet completed successfully!" $Green
    } else {
        Write-ColorOutput "Go vet found issues!" $Red
        exit 1
    }
}

function Run-Security {
    Write-ColorOutput "Running security analysis..." $Yellow
    $gosec = Get-Command gosec -ErrorAction SilentlyContinue
    if ($gosec) {
        gosec ./...
    } else {
        Write-ColorOutput "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest" $Red
    }
    Write-ColorOutput "Security analysis completed!" $Green
}

function Check-Vulnerabilities {
    Write-ColorOutput "Checking for vulnerabilities..." $Yellow
    $govulncheck = Get-Command govulncheck -ErrorAction SilentlyContinue
    if ($govulncheck) {
        govulncheck ./...
    } else {
        Write-ColorOutput "govulncheck not found. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest" $Red
    }
    Write-ColorOutput "Vulnerability check completed!" $Green
}

function Run-QualityCheck {
    Write-ColorOutput "Running all quality checks..." $Yellow
    Check-Format
    Run-Vet
    Run-Security
    Check-Vulnerabilities
    Test-Race
    Check-Coverage
    Write-ColorOutput "All quality checks passed!" $Green
}

function Clean {
    Write-ColorOutput "Cleaning up..." $Yellow
    go clean
    if (Test-Path $BuildDir) { Remove-Item -Recurse -Force $BuildDir }
    if (Test-Path $CoverageDir) { Remove-Item -Recurse -Force $CoverageDir }
    Get-ChildItem -Filter "*.out" | Remove-Item -Force
    Get-ChildItem -Filter "*.html" | Remove-Item -Force
    Get-ChildItem -Filter "*.json" | Remove-Item -Force
    Write-ColorOutput "Cleanup completed!" $Green
}

function Run-Benchmarks {
    Write-ColorOutput "Running benchmarks..." $Yellow
    go test -bench=. -benchmem -count=3 ./... | Tee-Object -FilePath "benchmarks.txt"
    Write-ColorOutput "Benchmarks completed! Results saved to benchmarks.txt" $Green
}

function Show-Info {
    Write-ColorOutput "Go Mastery Project Information" $Blue
    Write-ColorOutput "==============================" $Blue
    $goVersion = go version
    Write-ColorOutput "Go version: $goVersion"
    $goFiles = (Get-ChildItem -Recurse -Filter "*.go" | Where-Object { $_.FullName -notlike "*vendor*" }).Count
    $testFiles = (Get-ChildItem -Recurse -Filter "*_test.go" | Where-Object { $_.FullName -notlike "*vendor*" }).Count
    Write-ColorOutput "Go files: $goFiles"
    Write-ColorOutput "Test files: $testFiles"
    Write-ColorOutput "Build directory: $BuildDir"
    Write-ColorOutput "Coverage directory: $CoverageDir"
    Write-ColorOutput "Coverage threshold: $CoverageThreshold%"
}

# Main command dispatch
switch ($Command.ToLower()) {
    "help" { Show-Help }
    "setup" { Setup }
    "deps" { Download-Dependencies }
    "build" { Build }
    "test" { Test }
    "test-race" { Test-Race }
    "coverage" { Generate-Coverage }
    "coverage-check" { Check-Coverage }
    "fmt" { Format-Code }
    "fmt-check" { Check-Format }
    "vet" { Run-Vet }
    "security" { Run-Security }
    "vuln-check" { Check-Vulnerabilities }
    "quality-check" { Run-QualityCheck }
    "clean" { Clean }
    "bench" { Run-Benchmarks }
    "info" { Show-Info }
    default {
        Write-ColorOutput "Unknown command: $Command" $Red
        Write-ColorOutput "Run '.\build.ps1 help' for available commands." $Yellow
        exit 1
    }
}
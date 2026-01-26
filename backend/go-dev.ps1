<#
.SYNOPSIS
    Run Go commands in a container without installing Go locally.

.DESCRIPTION
    This script wraps Go tooling in a container, mounting your workspace
    to allow seamless development without a local Go installation.

.EXAMPLE
    .\go-dev.ps1 -- build -o bin/api cmd/api/main.go
    .\go-dev.ps1 test ./...
    .\go-dev.ps1 run cmd/api/main.go
    .\go-dev.ps1 mod tidy
    .\go-dev.ps1 fmt ./...

.NOTES
    Uses podman by default. Set $env:CONTAINER_RUNTIME="docker" to use Docker.
#>

param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Command
)

# Detect container runtime (podman or docker)
$runtime = if ($env:CONTAINER_RUNTIME) { $env:CONTAINER_RUNTIME } else { "podman" }

# Determine the backend directory (where this script lives)
$backendDir = $PSScriptRoot
$workspaceRoot = Split-Path -Parent $backendDir

# Convert Windows paths to Unix-style for container mounting
$unixBackendPath = $backendDir -replace '\\', '/' -replace '^([A-Z]):', { "/$($_.Groups[1].Value.ToLower())" }
$unixWorkspacePath = $workspaceRoot -replace '\\', '/' -replace '^([A-Z]):', { "/$($_.Groups[1].Value.ToLower())" }

# Common container options
$containerArgs = @(
    "run"
    "--rm"
    "-it"
    "-v", "${unixBackendPath}:/app"
    "-v", "easi-go-cache:/go/pkg/mod"
    "-v", "easi-go-build-cache:/root/.cache/go-build"
    "-w", "/app"
    "-e", "CGO_ENABLED=0"
    "golang:1.24-alpine"
)

# If no command provided, show usage
if ($null -eq $Command -or $Command.Length -eq 0) {
    Write-Host "Usage: .\go-dev.ps1 [--] <go-command> [args...]" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Cyan
    Write-Host "  .\go-dev.ps1 -- build -o bin/api cmd/api/main.go"
    Write-Host "  .\go-dev.ps1 test -v ./..."
    Write-Host "  .\go-dev.ps1 run cmd/api/main.go"
    Write-Host "  .\go-dev.ps1 mod tidy"
    Write-Host "  .\go-dev.ps1 fmt ./..."
    Write-Host "  .\go-dev.ps1 vet ./..."
    Write-Host "  .\go-dev.ps1 -- tool cover -html=coverage.out"
    Write-Host ""
    Write-Host "Interactive shell:" -ForegroundColor Cyan
    Write-Host "  .\go-dev.ps1 shell"
    Write-Host ""
    Write-Host "Note: Use '--' before commands with dash arguments to prevent PowerShell parameter conflicts" -ForegroundColor Gray
    exit 0
}

# Handle special "shell" command for interactive development
if ($Command[0] -eq "shell") {
    Write-Host "Starting interactive Go development shell..." -ForegroundColor Green
    & $runtime run --rm -it `
        -v "${unixBackendPath}:/app" `
        -v "easi-go-cache:/go/pkg/mod" `
        -v "easi-go-build-cache:/root/.cache/go-build" `
        -w /app `
        -e CGO_ENABLED=0 `
        golang:1.24-alpine sh
    exit $LASTEXITCODE
}

# Run the Go command
Write-Host "Running: go $($Command -join ' ')" -ForegroundColor Cyan
& $runtime @containerArgs go @Command
exit $LASTEXITCODE

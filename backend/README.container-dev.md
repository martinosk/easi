# Containerized Go Development (No Local Go Installation)

This guide shows how to develop EASI backend without installing Go on your machine.

## Prerequisites

- Podman or Docker installed
- PowerShell (for Windows scripts)

## Quick Start

### Option 1: PowerShell Script (Recommended for Windows)

The `go-dev.ps1` script wraps all Go commands in containers:

```powershell
# Navigate to backend directory
cd backend

# Build the application
.\go-dev.ps1 build -o bin/api cmd/api/main.go

# Run tests
.\go-dev.ps1 test -v ./...

# Format code
.\go-dev.ps1 fmt ./...

# Run the app (ensure DB is running via docker-compose)
.\go-dev.ps1 run cmd/api/main.go

# Tidy dependencies
.\go-dev.ps1 mod tidy

# Interactive shell for multiple commands
.\go-dev.ps1 shell
# Inside shell:
# go test ./...
# go build cmd/api/main.go
# exit
```

### Option 2: Container Makefile (Cross-platform)

Use the specialized Makefile for containerized builds:

```bash
# On Windows PowerShell
make -f Makefile.container build
make -f Makefile.container test
make -f Makefile.container coverage

# Create an alias for convenience
Set-Alias gmake "make -f Makefile.container"
gmake build
gmake test
```

### Option 3: Docker Compose Dev Service

Add to `docker-compose.yml` for persistent dev environment:

```yaml
  go-dev:
    image: golang:1.24-alpine
    container_name: easi-go-dev
    volumes:
      - ./backend:/app
      - easi-go-cache:/go/pkg/mod
      - easi-go-build-cache:/root/.cache/go-build
    working_dir: /app
    networks:
      - easi-network
    command: tail -f /dev/null  # Keep container running
```

Then execute commands:
```bash
podman-compose exec go-dev go build cmd/api/main.go
podman-compose exec go-dev go test ./...
```

## Common Development Tasks

### Running Tests

```powershell
# Unit tests
.\go-dev.ps1 test -v ./...

# Integration tests (requires running database)
cd .. && podman-compose up -d postgres && cd backend
.\test_integration.sh  # This script handles containerization
```

### Building

```powershell
# Build binary
.\go-dev.ps1 build -o bin/api cmd/api/main.go

# Build with version
.\go-dev.ps1 build -ldflags "-X 'easi/backend/internal/infrastructure/api.Version=1.0.0'" -o bin/api cmd/api/main.go
```

### Running the Application

```powershell
# Start database first
cd ..
podman-compose up -d postgres

# Run backend
cd backend
.\go-dev.ps1 run cmd/api/main.go
```

### Code Quality

```powershell
# Format code
.\go-dev.ps1 fmt ./...

# Vet code
.\go-dev.ps1 vet ./...

# Run linter (if golangci-lint is in container)
.\go-dev.ps1 shell
# Inside: golangci-lint run
```

### Managing Dependencies

```powershell
# Add a new dependency
.\go-dev.ps1 get github.com/some/package

# Tidy dependencies
.\go-dev.ps1 mod tidy

# Update dependencies
.\go-dev.ps1 get -u ./...
```

## Performance Optimization

### Persistent Volumes for Caching

The setup uses named volumes for Go module and build caches:
- `easi-go-cache`: Stores downloaded Go modules
- `easi-go-build-cache`: Stores compiled packages

This avoids re-downloading dependencies and rebuilding unchanged code.

### First Run

First execution downloads dependencies (slower):
```powershell
.\go-dev.ps1 mod download  # ~1-2 minutes
```

Subsequent runs use cached modules (fast):
```powershell
.\go-dev.ps1 test ./...  # Uses cache
```

## Integration with Existing Workflow

### Update test_integration.sh

The integration test script already handles containerization. Ensure it uses the same approach:

```bash
#!/bin/bash
# Uses container for running tests with database
podman run --rm \
  -v "$(pwd):/app" \
  -v "easi-go-cache:/go/pkg/mod" \
  --network easi_easi-network \
  -w /app \
  golang:1.24-alpine \
  go test -v ./...
```

### CI/CD Compatibility

This approach is CI/CD-friendly since pipelines already use containers. Your Azure Pipelines configuration remains unchanged.

## Troubleshooting

### "Network easi_easi-network not found"

Start docker-compose services first:
```powershell
cd ..
podman-compose up -d postgres
cd backend
```

### "Permission denied" on Linux

Add your user to the docker/podman group or run with sudo.

### Slow builds on Windows

Use WSL2 backend for better file system performance:
```powershell
wsl --set-default-version 2
```

### Cannot connect to database

Ensure the network name matches your docker-compose network. Check with:
```powershell
podman network ls
```

Adjust `--network` flag in `go-dev.ps1` if needed.

## IDE Integration

### VS Code

Install the Go extension and configure it to use the container:

1. Install "Remote - Containers" extension
2. Or use "Dev Containers" configuration (create `.devcontainer/devcontainer.json`)
3. Point Go tools to the container

### GoLand / IntelliJ

Use the "Go on Docker" configuration:
1. Settings → Go → GOROOT
2. Add Docker/Podman-based Go SDK
3. Configure container image: `golang:1.24-alpine`

## Switching Between Local and Container

If you later install Go locally:

```powershell
# Use local Go
go build cmd/api/main.go

# Use containerized Go
.\go-dev.ps1 build cmd/api/main.go
```

Keep both approaches available for flexibility.

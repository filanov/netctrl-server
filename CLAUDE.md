# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

netctrl-server is a Go server for cluster management. This is a new project using Go 1.23+.

## Build Methodology

This project uses a **containerized build approach** with **split Docker containers** - all build, lint, test, and code generation operations run inside Docker containers. This ensures:
- Consistent build environments across all machines
- No need to install Go, buf, golangci-lint, or other tools locally
- Elimination of "works on my machine" issues
- Reproducible builds for CI/CD
- Clear separation between development and production environments

### Docker Architecture

Two dedicated Dockerfiles:
- **`Dockerfile.dev`** (~2.25GB) - Development container with all tools
- **`Dockerfile.prod`** (~22.4MB) - Minimal production container

### Prerequisites

**Required:**
- Docker (20.10+)

**Optional (for native execution):**
- Go 1.25+
- buf CLI
- golangci-lint

If Docker is not available, the Makefile automatically falls back to local tool execution.

## Development Commands

### Quick Start
```bash
# First-time setup - build the development Docker image
make docker-build-dev

# Standard workflow (runs in Docker automatically)
make build                        # Build the binary
make test                         # Run tests
make lint                         # Run linters (golangci-lint + go vet)
make run                          # Build and run the server
```

### Docker-Based Workflow (Recommended)

All standard `make` commands automatically use Docker if available. The Makefile will build the dev image on first use.

#### Building & Running
```bash
make build                        # Build binary in container, output to bin/
make all                          # Clean, deps, proto, and build
make run                          # Build and run the server locally
make docker-run                   # Run server in production container
```

#### Docker Image Management
```bash
make docker-build-dev             # Build development image (with all tools)
make docker-build-prod            # Build production image (multi-stage, minimal)
make docker-shell                 # Open interactive shell in dev container
make docker-clean                 # Remove Docker images
```

#### Protocol Buffers
```bash
make proto                        # Generate Go code from proto files (in container)
```

#### Testing
```bash
make test                         # Run all tests in container
make test-race                    # Run tests with race detector in container
make coverage                     # Generate coverage report in container
```

#### Code Quality
```bash
make lint                         # Run golangci-lint + go vet in container
make fmt                          # Format all Go code in container
```

#### Dependency Management
```bash
make deps                         # Download and tidy dependencies
make tidy                         # Tidy go.mod and go.sum
```

### Native/Local Workflow (Optional)

If you have Go and tools installed locally, you can bypass Docker:

```bash
make build-local                  # Build using local Go
make proto-local                  # Generate proto using local buf
make test-local                   # Run tests using local Go
make lint-local                   # Lint using local tools
make fmt-local                    # Format using local Go
```

Or run Go commands directly:
```bash
go test ./...                     # Run all tests
go test -v ./...                  # Verbose output
go test -race ./...               # Race detector
go test -cover ./...              # Coverage
buf generate                      # Generate proto
buf lint                          # Lint proto files
golangci-lint run ./...           # Run comprehensive linter
```

### Make Targets Reference
```bash
make help                         # Show all available targets with descriptions
```

### Troubleshooting

#### Docker Issues
```bash
# If Docker image is outdated or corrupted
make docker-clean
make docker-build-dev

# Check Docker is running
docker ps

# View container logs during build
make build                        # Will show output from container
```

#### Permission Issues
```bash
# If generated files have wrong permissions
sudo chown -R $USER:$USER .
```

#### Build Cache Issues
```bash
# Clear Docker build cache
docker builder prune

# Clear Go module cache
rm -rf ~/go/pkg
rm -rf ~/.cache/go-build
```

## Project Structure

```
netctrl-server/
├── api/proto/v1/              # Protocol Buffer definitions
│   ├── cluster.proto          # Cluster CRUD operations
│   └── health.proto           # Health check endpoints
├── cmd/server/                # Application entry point
│   └── main.go
├── internal/                  # Private application code
│   ├── server/                # Server orchestration
│   │   ├── server.go          # Main coordinator
│   │   ├── grpc.go            # gRPC server
│   │   └── gateway.go         # HTTP gateway
│   ├── service/               # Business logic
│   │   ├── cluster.go         # Cluster CRUD
│   │   └── health.go          # Health checks
│   ├── storage/               # Data persistence
│   │   ├── interface.go       # Storage contract
│   │   └── memory/
│   │       └── memory.go      # In-memory implementation
│   └── config/
│       └── config.go          # Configuration management
├── pkg/api/v1/                # Generated gRPC code
├── configs/                   # Configuration files
│   └── config.yaml.example    # Configuration template
├── buf.yaml                   # Buf configuration
├── buf.gen.yaml               # Code generation config
└── Makefile                   # Build automation
```

## Architecture

### Dual-Port Server Architecture
- **gRPC Server** (port 9090): Native protocol buffers, high performance
- **HTTP Gateway** (port 8080): REST/JSON API, auto-generated from gRPC definitions
- **Shared Service Layer**: Both servers use the same business logic
- **Storage Backend**: Currently in-memory, designed for easy swapping

### API Endpoints

#### gRPC (localhost:9090)
- `netctrl.v1.ClusterService/CreateCluster`
- `netctrl.v1.ClusterService/GetCluster`
- `netctrl.v1.ClusterService/ListClusters`
- `netctrl.v1.ClusterService/UpdateCluster`
- `netctrl.v1.ClusterService/DeleteCluster`
- `netctrl.v1.HealthService/Check`
- `netctrl.v1.HealthService/Ready`

#### REST (localhost:8080)
- `POST /api/v1/clusters` - Create cluster
- `GET /api/v1/clusters` - List clusters
- `GET /api/v1/clusters/{id}` - Get cluster
- `PATCH /api/v1/clusters/{id}` - Update cluster
- `DELETE /api/v1/clusters/{id}` - Delete cluster
- `GET /api/v1/health` - Health check
- `GET /api/v1/ready` - Readiness check

### Testing the Server

Using grpcurl:
```bash
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext localhost:9090 netctrl.v1.HealthService/Check
```

Using curl:
```bash
curl http://localhost:8080/api/v1/health
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{"name":"test","description":"test cluster"}'
curl http://localhost:8080/api/v1/clusters
```

## Architecture Notes

The server is designed to handle cluster management:

- Cluster IDs are auto-generated UUIDs
- Timestamps are automatically managed
- Graceful shutdown on SIGTERM/SIGINT for container environments
- RESTful API with gRPC backend
- Input validation at the service layer

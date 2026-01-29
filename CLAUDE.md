# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

netctrl-server is a Go server for templating cluster network settings. This is a new project using Go 1.25.4.

## Development Commands

### Protocol Buffers
```bash
make proto                        # Generate Go code from proto files
buf generate                      # Alternative: direct buf command
buf lint                          # Lint proto files
buf breaking --against '.git#branch=main'  # Check for breaking changes
```

### Building
```bash
make build                        # Generate proto and build binary
make all                          # Clean, deps, proto, and build
go build -o bin/netctrl-server ./cmd/server  # Direct build
```

### Testing
```bash
go test ./...                     # Run all tests
go test -v ./...                  # Run all tests with verbose output
go test -race ./...               # Run tests with race detector
go test -cover ./...              # Run tests with coverage
go test -coverprofile=coverage.out ./...  # Generate coverage profile
go tool cover -html=coverage.out         # View coverage in browser
go test ./path/to/package         # Run tests for a specific package
go test -run TestName ./...       # Run a specific test by name
```

### Code Quality
```bash
go fmt ./...                      # Format all Go code
go vet ./...                      # Run static analysis
go mod tidy                       # Clean up go.mod and go.sum
go mod verify                     # Verify dependencies
```

### Dependency Management
```bash
go get <package>                  # Add a dependency
go get -u <package>               # Update a dependency
go mod download                   # Download dependencies
```

### Running
```bash
make run                          # Build and run the server
go run ./cmd/server               # Run without building first (requires proto generation)
./bin/netctrl-server              # Run the built binary directly
```

### Make Targets
```bash
make help                         # Show all available targets
make clean                        # Remove build artifacts
make deps                         # Download and tidy dependencies
make test                         # Run tests
make test-race                    # Run tests with race detector
make coverage                     # Generate and view coverage report
make lint                         # Run linters (go vet)
make fmt                          # Format code
make tidy                         # Tidy go.mod
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
  -d '{"name":"test","description":"test cluster","network_config":{"cidr":"10.0.0.0/16","gateway":"10.0.0.1","dns_servers":["8.8.8.8"],"mtu":1500}}'
curl http://localhost:8080/api/v1/clusters
```

## Architecture Notes

The server is designed to handle cluster network settings templating:

- Network configuration validation is enforced at the service layer
- CIDR, gateway IPs, DNS servers are validated on create/update
- MTU is constrained to 576-9000 range
- Cluster IDs are auto-generated UUIDs
- Timestamps are automatically managed
- Graceful shutdown on SIGTERM/SIGINT for container environments

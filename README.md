# netctrl-server

A high-performance gRPC server with REST API gateway for managing cluster network settings.

## Features

- **Dual API Access**: Native gRPC (port 9090) and REST/JSON gateway (port 8080)
- **Cluster Management**: Full CRUD operations for network cluster configurations
- **Network Validation**: Automatic validation of CIDR, IPs, DNS servers, and MTU
- **Health Checks**: Dedicated health and readiness endpoints
- **Graceful Shutdown**: Container-friendly signal handling
- **Protocol Buffers**: Type-safe, efficient API definitions

## Quick Start

### Prerequisites

**Required:**
- [Docker](https://docs.docker.com/get-docker/) (20.10+)

**Optional (for native development):**
- Go 1.23+
- [buf](https://buf.build/docs/installation)
- [golangci-lint](https://golangci-lint.run/welcome/install/)

### Quick Start with Docker

```bash
# Build the production image
make docker-build-prod

# Run the server
make docker-run

# Or use docker directly
docker run -p 8080:8080 -p 9090:9090 netctrl-server:latest
```

### Development with Docker

```bash
# Build development image (includes all build tools)
make docker-build-dev

# Standard workflow - all commands run in containers
make build        # Build binary
make test         # Run tests
make lint         # Run linters
make proto        # Generate protobuf code

# Open shell in dev container for debugging
make docker-shell
```

### Native Development (Optional)

If you have Go installed locally:

```bash
# Clone the repository
git clone <repository-url>
cd netctrl-server

# Generate protobuf code
make proto

# Build and run
make build
make run
```

The server will start with:
- gRPC server on `localhost:9090`
- HTTP gateway on `localhost:8080`

### Configuration

Copy the example configuration:

```bash
cp configs/config.yaml.example configs/config.yaml
```

Edit `configs/config.yaml` to customize settings. The server uses sensible defaults if the config file is not present.

## API Usage

### REST API Examples

#### Health Check

```bash
curl http://localhost:8080/api/v1/health
```

#### Create a Cluster

```bash
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production-cluster",
    "description": "Production environment cluster"
  }'
```

#### List All Clusters

```bash
curl http://localhost:8080/api/v1/clusters
```

#### Get a Specific Cluster

```bash
curl http://localhost:8080/api/v1/clusters/{cluster-id}
```

#### Update a Cluster

```bash
curl -X PATCH http://localhost:8080/api/v1/clusters/{cluster-id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "updated-cluster-name",
    "description": "Updated description"
  }'
```

#### Delete a Cluster

```bash
curl -X DELETE http://localhost:8080/api/v1/clusters/{cluster-id}
```

### gRPC API Examples

Using [grpcurl](https://github.com/fullstorydev/grpcurl):

```bash
# List available services
grpcurl -plaintext localhost:9090 list

# Health check
grpcurl -plaintext localhost:9090 netctrl.v1.HealthService/Check

# Create a cluster
grpcurl -plaintext -d '{
  "name": "production-cluster",
  "description": "Production environment"
}' localhost:9090 netctrl.v1.ClusterService/CreateCluster

# List clusters
grpcurl -plaintext localhost:9090 netctrl.v1.ClusterService/ListClusters
```

## Build Methodology

This project uses **split containerized builds** for clarity and efficiency:

- **Dedicated containers** - Separate `Dockerfile.dev` and `Dockerfile.prod` for development and production
- **No local tools needed** - All operations run in Docker (Go, buf, golangci-lint)
- **Automatic fallback** - Makefile detects if Docker is unavailable and uses local tools
- **Volume mounts** - Fast development iteration without rebuilding images
- **Comprehensive linting** - golangci-lint with 20+ enabled linters

### Docker Images

Two dedicated images from separate Dockerfiles:

1. **Development image** (`netctrl-server:dev` from `Dockerfile.dev`):
   - Size: ~2.25GB
   - Purpose: Building, testing, linting, proto generation
   - Contains: Go 1.25.4, buf, golangci-lint, all dev tools

2. **Production image** (`netctrl-server:latest` from `Dockerfile.prod`):
   - Size: ~22.4MB (99% smaller!)
   - Purpose: Production deployment only
   - Contains: Minimal Alpine + compiled binary + non-root user

### Workflow

```bash
# All standard commands use Docker automatically
make build        # Runs: docker run ... go build ...
make test         # Runs: docker run ... go test ...
make lint         # Runs: docker run ... golangci-lint ...

# Or use native tools with -local suffix
make build-local  # Uses local Go installation
```

## Development

### Project Structure

```
netctrl-server/
├── api/proto/v1/          # Protocol Buffer definitions
├── cmd/server/            # Application entry point
├── internal/              # Private application code
│   ├── server/            # Server orchestration
│   ├── service/           # Business logic
│   ├── storage/           # Data persistence
│   └── config/            # Configuration management
├── pkg/api/v1/            # Generated gRPC code
└── configs/               # Configuration files
```

### Common Commands

```bash
# Generate code from proto files
make proto

# Build the binary
make build

# Run tests
make test

# Run with race detector
make test-race

# Format code
make fmt

# Run linters
make lint

# View all available commands
make help
```

### Adding New Proto Definitions

1. Create or modify `.proto` files in `api/proto/v1/`
2. Run `make proto` to regenerate Go code
3. Implement the service handlers in `internal/service/`
4. Register the service in `internal/server/grpc.go`

## Architecture

### Dual-Port Setup

The server runs two separate ports:
- **Port 9090**: Native gRPC using protocol buffers
- **Port 8080**: HTTP/JSON gateway auto-generated from gRPC definitions

Both APIs share the same service layer and storage backend, ensuring consistency.

### Storage

Currently uses an in-memory storage implementation. The storage interface (`internal/storage/interface.go`) is designed for easy replacement with persistent backends like PostgreSQL, Redis, etc.

### Graceful Shutdown

The server listens for `SIGTERM` and `SIGINT` signals and performs graceful shutdown:
1. Stop accepting new requests
2. Complete in-flight requests
3. Clean up resources
4. Exit

## API Reference

### Cluster Object

```json
{
  "id": "uuid-string",
  "name": "cluster-name",
  "description": "optional description",
  "created_at": "2025-01-29T10:00:00Z",
  "updated_at": "2025-01-29T10:00:00Z"
}
```

### Error Responses

The API returns standard gRPC status codes (also mapped to HTTP status codes):
- `INVALID_ARGUMENT` (400): Invalid request parameters
- `NOT_FOUND` (404): Cluster not found
- `ALREADY_EXISTS` (409): Cluster with ID already exists
- `INTERNAL` (500): Internal server error

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]

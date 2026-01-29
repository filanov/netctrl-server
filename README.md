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

- Go 1.25.4 or higher
- buf CLI (auto-installed via `go install`)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd netctrl-server

# Download dependencies and build
make all
```

### Running the Server

```bash
# Using make
make run

# Or directly
./bin/netctrl-server
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
    "description": "Production environment cluster",
    "network_config": {
      "cidr": "10.0.0.0/16",
      "gateway": "10.0.0.1",
      "dns_servers": ["8.8.8.8", "8.8.4.4"],
      "mtu": 1500
    }
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
    "network_config": {
      "cidr": "10.1.0.0/16",
      "gateway": "10.1.0.1",
      "dns_servers": ["1.1.1.1"],
      "mtu": 9000
    }
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
  "description": "Production environment",
  "network_config": {
    "cidr": "10.0.0.0/16",
    "gateway": "10.0.0.1",
    "dns_servers": ["8.8.8.8"],
    "mtu": 1500
  }
}' localhost:9090 netctrl.v1.ClusterService/CreateCluster

# List clusters
grpcurl -plaintext localhost:9090 netctrl.v1.ClusterService/ListClusters
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

### Network Configuration Validation

The server automatically validates:
- **CIDR**: Must be valid CIDR notation (e.g., `10.0.0.0/16`)
- **Gateway**: Must be a valid IP address
- **DNS Servers**: Must be valid IP addresses
- **MTU**: Must be between 576 and 9000

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
  "network_config": {
    "cidr": "10.0.0.0/16",
    "gateway": "10.0.0.1",
    "dns_servers": ["8.8.8.8", "8.8.4.4"],
    "mtu": 1500
  },
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

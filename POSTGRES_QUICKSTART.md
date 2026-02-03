# PostgreSQL Quick Start

## Prerequisites
- Docker and Docker Compose
- golang-migrate (optional, for manual migrations)

## 1. Start PostgreSQL

```bash
make db-start
```

This will:
- Start PostgreSQL container
- Wait for it to be ready
- Database will be available at `localhost:5432`

## 2. Run Migrations

```bash
# Install golang-migrate first (if not installed)
brew install golang-migrate  # macOS
# or
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && sudo mv migrate /usr/local/bin/

# Run migrations
make db-migrate
```

## 3. Configure Server

### Option A: Environment Variable
```bash
export DATABASE_URL="postgres://netctrl:netctrl_dev_password@localhost:5432/netctrl?sslmode=disable"
```

### Option B: Config File
Edit `configs/config.yaml`:
```yaml
database:
  url: postgres://netctrl:netctrl_dev_password@localhost:5432/netctrl?sslmode=disable
  max_connections: 100
  min_connections: 20
```

## 4. Run Server

```bash
make run-local
```

## Useful Commands

```bash
# View database logs
make db-logs

# Connect to database
make db-shell

# Stop database
make db-stop

# Reset database (deletes all data!)
make db-reset

# Check migration status
make db-migrate-status

# Rollback last migration
make db-migrate-down
```

## Testing the Database

```bash
# Start database
make db-start

# Wait for ready
sleep 3

# Check connection
docker-compose exec postgres psql -U netctrl -d netctrl -c "\dt"

# Should show: clusters and agents tables

# Query some data
docker-compose exec postgres psql -U netctrl -d netctrl -c "SELECT COUNT(*) FROM agents;"
```

## Architecture

```
┌──────────────┐         ┌────────────────┐
│              │         │                │
│  netctrl-    │────────▶│   PostgreSQL   │
│  server      │         │   Container    │
│              │         │                │
└──────────────┘         └────────────────┘
       │
       │
       ▼
   Connection Pool
   (100 max conns)
```

### Database Schema

**Clusters Table:**
- id (UUID, PK)
- name, description
- created_at, updated_at (auto-managed)

**Agents Table:**
- id (UUID, PK)
- cluster_id (FK → clusters)
- hostname, ip_address, version
- status (enum)
- last_seen, created_at, updated_at
- hardware_collected (boolean)
- network_interfaces (JSONB)

**Indexes:**
- Cluster lookups
- Agent status queries
- Last seen timestamps (for monitor)
- Hardware collection status
- JSONB (GIN) for NIC queries

## Troubleshooting

**Port 5432 already in use:**
```bash
# Check what's using the port
lsof -i :5432

# Stop conflicting service or change port in docker-compose.yml
```

**Migration failed:**
```bash
# Check current version
make db-migrate-status

# Force version (if needed)
migrate -path ./migrations -database "$DB_URL" force VERSION

# Re-run
make db-migrate
```

**Connection refused:**
```bash
# Check if container is running
docker-compose ps

# Check logs
make db-logs

# Restart
docker-compose restart postgres
```

## Performance at Scale (9K agents/cluster)

The PostgreSQL setup is optimized for:
- **1,500+ ops/sec** sustained load
- **45,000+ agents** across clusters
- Efficient queries via strategic indexes
- Connection pooling (100 max connections)

For even higher scale, consider:
- Read replicas for monitoring
- Redis cache for hot data
- Table partitioning by cluster
- PgBouncer for connection pooling

See `DB_SETUP.md` for detailed documentation.

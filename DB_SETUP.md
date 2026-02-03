# Database Setup Guide

This guide explains how to set up and use PostgreSQL for netctrl-server development and production.

## Local Development with Docker Compose

### Prerequisites
- Docker and Docker Compose installed
- Ports 5432 (Postgres) and 5050 (PgAdmin) available

### Quick Start

1. **Start PostgreSQL**
   ```bash
   docker-compose up -d postgres
   ```

2. **Verify Database is Running**
   ```bash
   docker-compose ps
   # Should show postgres as healthy
   ```

3. **Check Database Connection**
   ```bash
   docker-compose exec postgres psql -U netctrl -d netctrl -c "\dt"
   ```

4. **Run Migrations**
   ```bash
   make db-migrate
   ```

5. **Start the Server with Postgres**
   ```bash
   export DATABASE_URL="postgres://netctrl:netctrl_dev_password@localhost:5432/netctrl?sslmode=disable"
   make run-local
   ```

### Database Credentials (Development)

```
Host:     localhost
Port:     5432
Database: netctrl
User:     netctrl
Password: netctrl_dev_password
```

**Connection String:**
```
postgres://netctrl:netctrl_dev_password@localhost:5432/netctrl?sslmode=disable
```

## Database Management

### Using PgAdmin (Optional)

1. **Start PgAdmin**
   ```bash
   docker-compose up -d pgadmin
   ```

2. **Access PgAdmin**
   - URL: http://localhost:5050
   - Email: admin@netctrl.local
   - Password: admin

3. **Add Server in PgAdmin**
   - Right-click "Servers" → "Create" → "Server"
   - **General Tab:**
     - Name: netctrl-local
   - **Connection Tab:**
     - Host: postgres (use container name)
     - Port: 5432
     - Database: netctrl
     - Username: netctrl
     - Password: netctrl_dev_password

### Using psql Command Line

```bash
# Connect to database
docker-compose exec postgres psql -U netctrl -d netctrl

# Common commands
\dt              # List tables
\d agents        # Describe agents table
\di              # List indexes
\l               # List databases
\q               # Quit
```

### Database Operations

**View all clusters:**
```sql
SELECT id, name, description, created_at FROM clusters;
```

**View all agents with hardware status:**
```sql
SELECT
    id,
    hostname,
    status,
    hardware_collected,
    jsonb_array_length(network_interfaces) as nic_count,
    last_seen
FROM agents
ORDER BY last_seen DESC;
```

**Find agents with specific NICs:**
```sql
SELECT
    id,
    hostname,
    network_interfaces
FROM agents
WHERE network_interfaces @> '[{"device_name": "mlx5_0"}]'::jsonb;
```

**Check inactive agents:**
```sql
SELECT id, hostname, last_seen,
       NOW() - last_seen as offline_duration
FROM agents
WHERE status = 'AGENT_STATUS_INACTIVE'
ORDER BY last_seen DESC;
```

## Migrations

### Using golang-migrate

1. **Install migrate tool**
   ```bash
   # macOS
   brew install golang-migrate

   # Linux
   curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
   sudo mv migrate /usr/local/bin/
   ```

2. **Run migrations**
   ```bash
   migrate -path ./migrations \
           -database "postgres://netctrl:netctrl_dev_password@localhost:5432/netctrl?sslmode=disable" \
           up
   ```

3. **Rollback migrations**
   ```bash
   migrate -path ./migrations \
           -database "postgres://netctrl:netctrl_dev_password@localhost:5432/netctrl?sslmode=disable" \
           down 1
   ```

4. **Check migration status**
   ```bash
   migrate -path ./migrations \
           -database "postgres://netctrl:netctrl_dev_password@localhost:5432/netctrl?sslmode=disable" \
           version
   ```

### Creating New Migrations

```bash
# Create new migration files
migrate create -ext sql -dir migrations -seq add_new_feature
```

This creates:
- `migrations/002_add_new_feature.up.sql`
- `migrations/002_add_new_feature.down.sql`

## Production Deployment

### Environment Variables

```bash
export DATABASE_URL="postgres://user:password@host:5432/dbname?sslmode=require"
export DB_MAX_CONNECTIONS=100
export DB_MIN_CONNECTIONS=20
```

### Production Configuration

```yaml
# configs/config.yaml
database:
  url: ${DATABASE_URL}
  max_connections: 100
  min_connections: 20
  max_conn_idle_time: 5m
  connection_timeout: 10s
```

### SSL/TLS Configuration

For production, always use SSL:
```
postgres://user:password@host:5432/dbname?sslmode=require
```

SSL Modes:
- `disable` - No SSL (development only)
- `require` - SSL required, no verification
- `verify-ca` - SSL + verify CA certificate
- `verify-full` - SSL + verify CA + hostname

### Backup and Restore

**Backup:**
```bash
# Full backup
docker-compose exec postgres pg_dump -U netctrl netctrl > backup.sql

# Compressed backup
docker-compose exec postgres pg_dump -U netctrl netctrl | gzip > backup.sql.gz
```

**Restore:**
```bash
# From SQL file
docker-compose exec -T postgres psql -U netctrl netctrl < backup.sql

# From compressed file
gunzip -c backup.sql.gz | docker-compose exec -T postgres psql -U netctrl netctrl
```

## Troubleshooting

### Connection Refused

```bash
# Check if postgres is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Restart postgres
docker-compose restart postgres
```

### Database Locked

```bash
# Check active connections
docker-compose exec postgres psql -U netctrl -d netctrl -c "SELECT * FROM pg_stat_activity;"

# Kill hanging connections (if needed)
docker-compose exec postgres psql -U netctrl -d netctrl -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'netctrl' AND pid != pg_backend_pid();"
```

### Reset Database

```bash
# Stop all services
docker-compose down

# Remove data volumes
docker-compose down -v

# Start fresh
docker-compose up -d postgres
make db-migrate
```

### Performance Issues

```bash
# Check slow queries
docker-compose exec postgres psql -U netctrl -d netctrl -c "SELECT query, calls, total_time, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"

# Check index usage
docker-compose exec postgres psql -U netctrl -d netctrl -c "SELECT schemaname, tablename, indexname, idx_scan FROM pg_stat_user_indexes ORDER BY idx_scan;"

# Check table sizes
docker-compose exec postgres psql -U netctrl -d netctrl -c "SELECT tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
```

## Clean Up

```bash
# Stop services
docker-compose down

# Stop and remove volumes (deletes all data!)
docker-compose down -v

# Remove images
docker-compose down --rmi all -v
```

## Switching Between Storage Backends

The server uses PostgreSQL as its storage backend:

```bash
export DATABASE_URL="postgres://netctrl:netctrl_dev_password@localhost:5432/netctrl?sslmode=disable"
make run-local
```

## References

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [pgx Driver](https://github.com/jackc/pgx)
- [Docker Compose](https://docs.docker.com/compose/)

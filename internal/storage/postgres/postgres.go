package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/filanov/netctrl-server/internal/storage"
)

// Storage implements the storage.Storage interface using PostgreSQL
type Storage struct {
	pool *pgxpool.Pool
}

// Config holds PostgreSQL configuration
type Config struct {
	URL             string
	MaxConnIdleTime string
	ConnectTimeout  string
	MaxConnections  int32
	MinConnections  int32
}

// New creates a new PostgreSQL storage instance
func New(ctx context.Context, cfg Config) (*Storage, error) {
	// Parse connection string
	config, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Apply configuration
	if cfg.MaxConnections > 0 {
		config.MaxConns = cfg.MaxConnections
	}
	if cfg.MinConnections > 0 {
		config.MinConns = cfg.MinConnections
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &Storage{pool: pool}, nil
}

// Close closes the database connection pool
func (s *Storage) Close() {
	s.pool.Close()
}

// Ensure Storage implements storage.Storage interface
var _ storage.Storage = (*Storage)(nil)

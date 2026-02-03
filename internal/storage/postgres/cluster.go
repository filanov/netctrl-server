package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

// CreateCluster creates a new cluster
func (s *Storage) CreateCluster(ctx context.Context, cluster *v1.Cluster) error {
	query := `
		INSERT INTO clusters (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.pool.Exec(ctx, query,
		cluster.Id,
		cluster.Name,
		cluster.Description,
		cluster.CreatedAt.AsTime(),
		cluster.UpdatedAt.AsTime(),
	)

	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	return nil
}

// GetCluster retrieves a cluster by ID
func (s *Storage) GetCluster(ctx context.Context, id string) (*v1.Cluster, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM clusters
		WHERE id = $1
	`

	var cluster v1.Cluster
	var createdAt, updatedAt time.Time

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&cluster.Id,
		&cluster.Name,
		&cluster.Description,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("cluster not found")
		}
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	cluster.CreatedAt = timestamppb.New(createdAt)
	cluster.UpdatedAt = timestamppb.New(updatedAt)

	return &cluster, nil
}

// ListClusters lists all clusters
func (s *Storage) ListClusters(ctx context.Context) ([]*v1.Cluster, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM clusters
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}
	defer rows.Close()

	var clusters []*v1.Cluster
	for rows.Next() {
		var cluster v1.Cluster
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&cluster.Id,
			&cluster.Name,
			&cluster.Description,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cluster: %w", err)
		}

		cluster.CreatedAt = timestamppb.New(createdAt)
		cluster.UpdatedAt = timestamppb.New(updatedAt)

		clusters = append(clusters, &cluster)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clusters: %w", err)
	}

	return clusters, nil
}

// UpdateCluster updates an existing cluster
func (s *Storage) UpdateCluster(ctx context.Context, cluster *v1.Cluster) error {
	query := `
		UPDATE clusters
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := s.pool.Exec(ctx, query,
		cluster.Id,
		cluster.Name,
		cluster.Description,
		cluster.UpdatedAt.AsTime(),
	)

	if err != nil {
		return fmt.Errorf("failed to update cluster: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("cluster not found")
	}

	return nil
}

// DeleteCluster deletes a cluster by ID
func (s *Storage) DeleteCluster(ctx context.Context, id string) error {
	query := `DELETE FROM clusters WHERE id = $1`

	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("cluster not found")
	}

	return nil
}

// ClusterExists checks if a cluster exists
func (s *Storage) ClusterExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM clusters WHERE id = $1)`

	var exists bool
	err := s.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check cluster existence: %w", err)
	}

	return exists, nil
}

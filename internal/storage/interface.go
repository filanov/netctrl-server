package storage

import (
	"context"

	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

// Storage defines the interface for cluster data persistence
type Storage interface {
	// CreateCluster stores a new cluster
	CreateCluster(ctx context.Context, cluster *v1.Cluster) error

	// GetCluster retrieves a cluster by ID
	GetCluster(ctx context.Context, id string) (*v1.Cluster, error)

	// ListClusters returns all clusters
	ListClusters(ctx context.Context) ([]*v1.Cluster, error)

	// UpdateCluster updates an existing cluster
	UpdateCluster(ctx context.Context, cluster *v1.Cluster) error

	// DeleteCluster removes a cluster by ID
	DeleteCluster(ctx context.Context, id string) error

	// ClusterExists checks if a cluster exists by ID
	ClusterExists(ctx context.Context, id string) (bool, error)
}

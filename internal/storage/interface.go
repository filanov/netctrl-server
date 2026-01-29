package storage

import (
	"context"

	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

// Storage defines the interface for cluster and agent data persistence
type Storage interface {
	// Cluster operations
	CreateCluster(ctx context.Context, cluster *v1.Cluster) error
	GetCluster(ctx context.Context, id string) (*v1.Cluster, error)
	ListClusters(ctx context.Context) ([]*v1.Cluster, error)
	UpdateCluster(ctx context.Context, cluster *v1.Cluster) error
	DeleteCluster(ctx context.Context, id string) error
	ClusterExists(ctx context.Context, id string) (bool, error)

	// Agent operations
	CreateAgent(ctx context.Context, agent *v1.Agent) error
	GetAgent(ctx context.Context, id string) (*v1.Agent, error)
	ListAgents(ctx context.Context, clusterID string) ([]*v1.Agent, error)
	UpdateAgent(ctx context.Context, agent *v1.Agent) error
	DeleteAgent(ctx context.Context, id string) error
}

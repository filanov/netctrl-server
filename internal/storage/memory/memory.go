package memory

import (
	"context"
	"fmt"
	"sync"

	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

// Storage implements an in-memory storage backend
type Storage struct {
	mu       sync.RWMutex
	clusters map[string]*v1.Cluster
}

// New creates a new in-memory storage instance
func New() *Storage {
	return &Storage{
		clusters: make(map[string]*v1.Cluster),
	}
}

// CreateCluster stores a new cluster
func (s *Storage) CreateCluster(ctx context.Context, cluster *v1.Cluster) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clusters[cluster.Id]; exists {
		return fmt.Errorf("cluster with ID %s already exists", cluster.Id)
	}

	s.clusters[cluster.Id] = cluster
	return nil
}

// GetCluster retrieves a cluster by ID
func (s *Storage) GetCluster(ctx context.Context, id string) (*v1.Cluster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cluster, exists := s.clusters[id]
	if !exists {
		return nil, fmt.Errorf("cluster with ID %s not found", id)
	}

	return cluster, nil
}

// ListClusters returns all clusters
func (s *Storage) ListClusters(ctx context.Context) ([]*v1.Cluster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clusters := make([]*v1.Cluster, 0, len(s.clusters))
	for _, cluster := range s.clusters {
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

// UpdateCluster updates an existing cluster
func (s *Storage) UpdateCluster(ctx context.Context, cluster *v1.Cluster) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clusters[cluster.Id]; !exists {
		return fmt.Errorf("cluster with ID %s not found", cluster.Id)
	}

	s.clusters[cluster.Id] = cluster
	return nil
}

// DeleteCluster removes a cluster by ID
func (s *Storage) DeleteCluster(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clusters[id]; !exists {
		return fmt.Errorf("cluster with ID %s not found", id)
	}

	delete(s.clusters, id)
	return nil
}

// ClusterExists checks if a cluster exists by ID
func (s *Storage) ClusterExists(ctx context.Context, id string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.clusters[id]
	return exists, nil
}

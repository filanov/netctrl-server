package mock

import (
	"context"
	"fmt"
	"sync"

	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

// Storage is an in-memory storage implementation for testing
type Storage struct {
	clusters map[string]*v1.Cluster
	agents   map[string]*v1.Agent
	mu       sync.RWMutex
}

// New creates a new mock storage instance
func New() *Storage {
	return &Storage{
		clusters: make(map[string]*v1.Cluster),
		agents:   make(map[string]*v1.Agent),
	}
}

// Cluster operations

func (s *Storage) CreateCluster(ctx context.Context, cluster *v1.Cluster) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clusters[cluster.Id] = cluster
	return nil
}

func (s *Storage) GetCluster(ctx context.Context, id string) (*v1.Cluster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cluster, ok := s.clusters[id]
	if !ok {
		return nil, fmt.Errorf("cluster not found")
	}
	return cluster, nil
}

func (s *Storage) ListClusters(ctx context.Context) ([]*v1.Cluster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	clusters := make([]*v1.Cluster, 0, len(s.clusters))
	for _, cluster := range s.clusters {
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

func (s *Storage) UpdateCluster(ctx context.Context, cluster *v1.Cluster) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.clusters[cluster.Id]; !ok {
		return fmt.Errorf("cluster not found")
	}
	s.clusters[cluster.Id] = cluster
	return nil
}

func (s *Storage) DeleteCluster(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.clusters[id]; !ok {
		return fmt.Errorf("cluster not found")
	}
	delete(s.clusters, id)
	// Delete associated agents
	for agentID, agent := range s.agents {
		if agent.ClusterId == id {
			delete(s.agents, agentID)
		}
	}
	return nil
}

func (s *Storage) ClusterExists(ctx context.Context, id string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.clusters[id]
	return ok, nil
}

// Agent operations

func (s *Storage) CreateAgent(ctx context.Context, agent *v1.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agents[agent.Id] = agent
	return nil
}

func (s *Storage) GetAgent(ctx context.Context, id string) (*v1.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	agent, ok := s.agents[id]
	if !ok {
		return nil, fmt.Errorf("agent not found")
	}
	return agent, nil
}

func (s *Storage) ListAgents(ctx context.Context, clusterID string) ([]*v1.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	agents := make([]*v1.Agent, 0)
	for _, agent := range s.agents {
		if clusterID == "" || agent.ClusterId == clusterID {
			agents = append(agents, agent)
		}
	}
	return agents, nil
}

func (s *Storage) UpdateAgent(ctx context.Context, agent *v1.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[agent.Id]; !ok {
		return fmt.Errorf("agent not found")
	}
	s.agents[agent.Id] = agent
	return nil
}

func (s *Storage) DeleteAgent(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[id]; !ok {
		return fmt.Errorf("agent not found")
	}
	delete(s.agents, id)
	return nil
}

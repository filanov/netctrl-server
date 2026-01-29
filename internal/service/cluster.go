package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
	"github.com/filanov/netctrl-server/internal/storage"
)

// ClusterService implements the cluster management service
type ClusterService struct {
	v1.UnimplementedClusterServiceServer
	storage storage.Storage
}

// NewClusterService creates a new cluster service instance
func NewClusterService(storage storage.Storage) *ClusterService {
	return &ClusterService{
		storage: storage,
	}
}

// CreateCluster creates a new cluster
func (s *ClusterService) CreateCluster(ctx context.Context, req *v1.CreateClusterRequest) (*v1.CreateClusterResponse, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Create cluster entity
	now := timestamppb.Now()
	cluster := &v1.Cluster{
		Id:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Store cluster
	if err := s.storage.CreateCluster(ctx, cluster); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create cluster: %v", err)
	}

	return &v1.CreateClusterResponse{
		Cluster: cluster,
	}, nil
}

// GetCluster retrieves a cluster by ID
func (s *ClusterService) GetCluster(ctx context.Context, req *v1.GetClusterRequest) (*v1.GetClusterResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster ID is required")
	}

	cluster, err := s.storage.GetCluster(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cluster not found: %v", err)
	}

	return &v1.GetClusterResponse{
		Cluster: cluster,
	}, nil
}

// ListClusters lists all clusters
func (s *ClusterService) ListClusters(ctx context.Context, req *v1.ListClustersRequest) (*v1.ListClustersResponse, error) {
	clusters, err := s.storage.ListClusters(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list clusters: %v", err)
	}

	return &v1.ListClustersResponse{
		Clusters: clusters,
	}, nil
}

// UpdateCluster updates an existing cluster
func (s *ClusterService) UpdateCluster(ctx context.Context, req *v1.UpdateClusterRequest) (*v1.UpdateClusterResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster ID is required")
	}

	// Get existing cluster
	cluster, err := s.storage.GetCluster(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cluster not found: %v", err)
	}

	// Update fields
	if req.Name != "" {
		cluster.Name = req.Name
	}
	if req.Description != "" {
		cluster.Description = req.Description
	}

	cluster.UpdatedAt = timestamppb.Now()

	// Store updated cluster
	if err := s.storage.UpdateCluster(ctx, cluster); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update cluster: %v", err)
	}

	return &v1.UpdateClusterResponse{
		Cluster: cluster,
	}, nil
}

// DeleteCluster deletes a cluster by ID
func (s *ClusterService) DeleteCluster(ctx context.Context, req *v1.DeleteClusterRequest) (*v1.DeleteClusterResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster ID is required")
	}

	if err := s.storage.DeleteCluster(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.NotFound, "cluster not found: %v", err)
	}

	return &v1.DeleteClusterResponse{
		Success: true,
	}, nil
}

// validateCreateRequest validates the create cluster request
func (s *ClusterService) validateCreateRequest(req *v1.CreateClusterRequest) error {
	if req.Name == "" {
		return fmt.Errorf("cluster name is required")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("cluster name must be less than 255 characters")
	}

	return nil
}

package service

import (
	"context"
	"fmt"
	"net"

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
		Id:            uuid.New().String(),
		Name:          req.Name,
		Description:   req.Description,
		NetworkConfig: req.NetworkConfig,
		CreatedAt:     now,
		UpdatedAt:     now,
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
	if req.NetworkConfig != nil {
		if err := s.validateNetworkConfig(req.NetworkConfig); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		cluster.NetworkConfig = req.NetworkConfig
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

	if req.NetworkConfig == nil {
		return fmt.Errorf("network configuration is required")
	}

	return s.validateNetworkConfig(req.NetworkConfig)
}

// validateNetworkConfig validates network configuration
func (s *ClusterService) validateNetworkConfig(config *v1.NetworkConfig) error {
	if config.Cidr == "" {
		return fmt.Errorf("CIDR is required")
	}

	// Validate CIDR format
	_, _, err := net.ParseCIDR(config.Cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR format: %v", err)
	}

	if config.Gateway == "" {
		return fmt.Errorf("gateway is required")
	}

	// Validate gateway IP
	if net.ParseIP(config.Gateway) == nil {
		return fmt.Errorf("invalid gateway IP address")
	}

	// Validate DNS servers
	for _, dns := range config.DnsServers {
		if net.ParseIP(dns) == nil {
			return fmt.Errorf("invalid DNS server IP address: %s", dns)
		}
	}

	// Validate MTU
	if config.Mtu < 576 || config.Mtu > 9000 {
		return fmt.Errorf("MTU must be between 576 and 9000")
	}

	return nil
}

package service

import (
	"context"

	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

// HealthService implements the health check service
type HealthService struct {
	v1.UnimplementedHealthServiceServer
}

// NewHealthService creates a new health service instance
func NewHealthService() *HealthService {
	return &HealthService{}
}

// Check returns the health status of the service
func (s *HealthService) Check(ctx context.Context, req *v1.HealthCheckRequest) (*v1.HealthCheckResponse, error) {
	return &v1.HealthCheckResponse{
		Status:  v1.HealthStatus_HEALTH_STATUS_HEALTHY,
		Message: "Service is healthy",
	}, nil
}

// Ready returns the readiness status of the service
func (s *HealthService) Ready(ctx context.Context, req *v1.ReadinessCheckRequest) (*v1.ReadinessCheckResponse, error) {
	return &v1.ReadinessCheckResponse{
		Status:  v1.ReadinessStatus_READINESS_STATUS_READY,
		Message: "Service is ready",
	}, nil
}

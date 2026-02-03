package server

import (
	"context"
	"log"
	"sync"

	"net/http"

	"google.golang.org/grpc"

	"github.com/filanov/netctrl-server/internal/config"
	"github.com/filanov/netctrl-server/internal/service"
	"github.com/filanov/netctrl-server/internal/storage"
)

// Server orchestrates the gRPC and HTTP gateway servers
type Server struct {
	config         *config.Config
	storage        storage.Storage
	clusterService *service.ClusterService
	agentService   *service.AgentService
	healthService  *service.HealthService
	agentMonitor   *service.AgentMonitor

	grpcServer    *grpc.Server
	gatewayServer *http.Server
	gatewayCancel context.CancelFunc
	monitorCtx    context.Context
	monitorCancel context.CancelFunc
}

// New creates a new server instance
func New(cfg *config.Config, store storage.Storage) *Server {
	monitorCtx, monitorCancel := context.WithCancel(context.Background())
	return &Server{
		config:         cfg,
		storage:        store,
		clusterService: service.NewClusterService(store),
		agentService:   service.NewAgentService(store),
		healthService:  service.NewHealthService(),
		agentMonitor:   service.NewAgentMonitor(store),
		monitorCtx:     monitorCtx,
		monitorCancel:  monitorCancel,
	}
}

// Start starts both the gRPC and HTTP gateway servers
func (s *Server) Start() error {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Start agent monitor
	go s.agentMonitor.Start(s.monitorCtx)

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.startGRPCServer(); err != nil {
			errChan <- err
		}
	}()

	// Wait a bit for gRPC server to start before starting gateway
	// This ensures the gateway can connect to gRPC
	// In production, you might want a more robust health check
	log.Println("Waiting for gRPC server to be ready...")
	// Simple sleep is acceptable here as gateway needs gRPC to be available
	// Alternative: implement proper health check polling
	// time.Sleep(500 * time.Millisecond)

	// Start HTTP gateway server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.startGatewayServer(); err != nil {
			errChan <- err
		}
	}()

	// Wait for all servers to finish or for an error
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Return the first error if any
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop gracefully stops both servers
func (s *Server) Stop() {
	log.Println("Shutting down servers...")

	// Stop agent monitor
	if s.monitorCancel != nil {
		s.monitorCancel()
	}

	s.stopGatewayServer()
	s.stopGRPCServer()
	log.Println("Servers stopped successfully")
}

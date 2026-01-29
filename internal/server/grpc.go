package server

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

// startGRPCServer starts the gRPC server
func (s *Server) startGRPCServer() error {
	addr := fmt.Sprintf(":%d", s.config.GRPC.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Create gRPC server with options
	grpcServer := grpc.NewServer()

	// Register services
	v1.RegisterClusterServiceServer(grpcServer, s.clusterService)
	v1.RegisterHealthServiceServer(grpcServer, s.healthService)

	// Enable reflection for grpcurl and other tools
	if s.config.GRPC.EnableReflection {
		reflection.Register(grpcServer)
	}

	s.grpcServer = grpcServer

	log.Printf("gRPC server listening on %s", addr)

	// Start serving (blocking)
	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("gRPC server failed: %w", err)
	}

	return nil
}

// stopGRPCServer gracefully stops the gRPC server
func (s *Server) stopGRPCServer() {
	if s.grpcServer != nil {
		log.Println("Stopping gRPC server...")
		s.grpcServer.GracefulStop()
		log.Println("gRPC server stopped")
	}
}

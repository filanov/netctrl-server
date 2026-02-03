package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"

	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

// filteredLogger filters out harmless ServerMetadata errors
type filteredLogger struct{}

func (f *filteredLogger) Write(p []byte) (n int, err error) {
	msg := string(p)
	// Suppress harmless ServerMetadata extraction errors
	if strings.Contains(msg, "Failed to extract ServerMetadata from context") {
		return len(p), nil
	}
	return log.Writer().Write(p)
}

// startGatewayServer starts the HTTP gateway server
func (s *Server) startGatewayServer() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	s.gatewayCancel = cancel

	// Configure grpclog to filter out harmless errors
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(io.Discard, &filteredLogger{}, &filteredLogger{}))

	// Create gRPC-Gateway mux
	mux := runtime.NewServeMux()

	// Connect to gRPC server
	grpcAddr := fmt.Sprintf("localhost:%d", s.config.GRPC.Port)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register service handlers
	if err := v1.RegisterClusterServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register cluster service handler: %w", err)
	}

	if err := v1.RegisterAgentServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register agent service handler: %w", err)
	}

	if err := v1.RegisterHealthServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register health service handler: %w", err)
	}

	// Create HTTP server with middleware
	handler := http.Handler(mux)
	if s.config.Gateway.EnableCORS {
		handler = corsMiddleware(handler)
	}

	addr := fmt.Sprintf(":%d", s.config.Gateway.Port)
	s.gatewayServer = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("HTTP gateway listening on %s", addr)

	// Start serving (blocking)
	if err := s.gatewayServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("gateway server failed: %w", err)
	}

	return nil
}

// stopGatewayServer gracefully stops the HTTP gateway server
func (s *Server) stopGatewayServer() {
	if s.gatewayServer != nil {
		log.Println("Stopping HTTP gateway...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.gatewayServer.Shutdown(ctx); err != nil {
			log.Printf("Gateway shutdown error: %v", err)
		}
		log.Println("HTTP gateway stopped")
	}

	if s.gatewayCancel != nil {
		s.gatewayCancel()
	}
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

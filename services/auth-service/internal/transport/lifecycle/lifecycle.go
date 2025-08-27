package lifecycle

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/internal/config"

	zlog "packages/logger"

	"google.golang.org/grpc"
)

// Manager handles server lifecycle operations
type Manager struct {
	logger     *zlog.Logger
	config     *config.HealthConfig
	grpcServer *grpc.Server
	restServer any // Will be *http.Server
	grpcLis    net.Listener
	restLis    net.Listener
}

// NewManager creates a new lifecycle manager
func NewManager(logger *zlog.Logger, cfg *config.HealthConfig) *Manager {
	return &Manager{
		logger: logger,
		config: cfg,
	}
}

// SetServers sets the gRPC and REST servers
func (lm *Manager) SetServers(grpcServer *grpc.Server, restServer any, grpcLis, restLis net.Listener) {
	lm.grpcServer = grpcServer
	lm.restServer = restServer
	lm.grpcLis = grpcLis
	lm.restLis = restLis
}

// Start starts both gRPC and REST servers
func (lm *Manager) Start(ctx context.Context) error {
	lm.logger.Info(ctx, "Starting gRPC service", map[string]any{
		"port": lm.grpcLis.Addr().String(),
	})

	// Start gRPC server in a goroutine
	go func() {
		if err := lm.grpcServer.Serve(lm.grpcLis); err != nil {
			lm.logger.Error(ctx, err, "gRPC server failed to start", 500)
			os.Exit(1)
		}
	}()

	// Wait for gRPC server to be ready
	if err := lm.waitForGRPCReady(ctx); err != nil {
		return fmt.Errorf("gRPC server readiness check failed: %w", err)
	}

	lm.logger.Info(ctx, "Starting REST gateway", map[string]any{
		"port": lm.restLis.Addr().String(),
	})

	// Start REST server in a goroutine
	go func() {
		if restServer, ok := lm.restServer.(interface{ Serve(net.Listener) error }); ok {
			if err := restServer.Serve(lm.restLis); err != nil {
				lm.logger.Error(ctx, err, "REST server failed to start", 500)
				os.Exit(1)
			}
		}
	}()

	// Wait a moment for REST server to start
	time.Sleep(lm.config.ReadinessDelay)

	lm.logger.Info(ctx, "All servers started successfully", map[string]any{
		"grpc_port": lm.grpcLis.Addr().String(),
		"rest_port": lm.restLis.Addr().String(),
	})

	return nil
}

// waitForGRPCReady waits for the gRPC server to be ready
func (lm *Manager) waitForGRPCReady(ctx context.Context) error {
	ready := make(chan bool, 1)
	go func() {
		maxAttempts := 50 // 5 seconds with 100ms intervals
		attempts := 0
		for attempts < maxAttempts {
			select {
			case <-ctx.Done():
				return
			default:
				// Try to connect to the gRPC server to check if it's ready
				conn, err := grpc.Dial(lm.grpcLis.Addr().String(),
					grpc.WithInsecure(),
					grpc.WithBlock(),
					grpc.WithTimeout(100*time.Millisecond))
				if err == nil {
					conn.Close()
					ready <- true
					return
				}
				attempts++
				time.Sleep(100 * time.Millisecond)
			}
		}
		// If we reach here, the server didn't become ready
		lm.logger.Warn(ctx, "gRPC server readiness check timed out", map[string]any{
			"max_attempts": maxAttempts,
			"timeout_ms":   maxAttempts * 100,
		})
	}()

	// Wait for gRPC server to be ready or timeout
	select {
	case <-ready:
		lm.logger.Info(ctx, "gRPC server is ready")
		return nil
	case <-time.After(10 * time.Second):
		lm.logger.Warn(ctx, "gRPC server readiness check timed out, proceeding anyway")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Shutdown gracefully shuts down both servers
func (lm *Manager) Shutdown(ctx context.Context) error {
	lm.logger.Info(ctx, "Shutting down servers")

	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, lm.config.Timeout)
	defer cancel()

	// Graceful stop the gRPC server
	if lm.grpcServer != nil {
		lm.grpcServer.GracefulStop()
	}

	// Shutdown the REST server
	if lm.restServer != nil {
		if restServer, ok := lm.restServer.(interface{ Shutdown(context.Context) error }); ok {
			if err := restServer.Shutdown(shutdownCtx); err != nil {
				lm.logger.Error(shutdownCtx, err, "Failed to shutdown REST server", 500)
			}
		}
	}

	lm.logger.Info(ctx, "Server shutdown completed")
	return nil
}

// Run starts both servers and handles graceful shutdown
func (lm *Manager) Run(ctx context.Context) error {
	// Start the servers
	if err := lm.Start(ctx); err != nil {
		return fmt.Errorf("failed to start servers: %v", err)
	}

	// Set up signal handling with buffered channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	lm.logger.Info(ctx, "Received shutdown signal")

	// Perform graceful shutdown
	if err := lm.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown servers: %v", err)
	}

	return nil
}

// IsHealthy checks if the servers are healthy
func (lm *Manager) IsHealthy() bool {
	// Check if gRPC server is listening
	if lm.grpcLis == nil {
		return false
	}

	// Check if REST server is listening
	if lm.restLis == nil {
		return false
	}

	return true
}

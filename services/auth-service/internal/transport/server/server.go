package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"time"

	"auth-service/config"
	"auth-service/internal/handler/http"
	"auth-service/internal/repository"
	"auth-service/internal/services"
	"auth-service/internal/transport/lifecycle"

	zlog "packages/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server holds both gRPC and REST servers and their dependencies
type Server struct {
	deps         *Dependencies
	grpcServer   *grpc.Server
	restGateway  *http.RESTGateway
	lifecycle    *lifecycle.Manager
	grpcListener net.Listener
	restListener net.Listener
}

// NewServer initializes both gRPC and REST servers with their dependencies
func NewServer(ctx context.Context) (*Server, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logger := zlog.NewLogger(zlog.Config{
		Level:      cfg.LogLevel,
		Output:     os.Stdout,
		JSONFormat: cfg.LogJSONFormat,
		AddCaller:  true,
		TimeFormat: time.RFC3339,
	})

	// Create a context with correlation ID for initialization
	ctx = zlog.WithCorrelationID(ctx, "")

	// Initialize database
	logger.Info(ctx, "Initializing database")
	db, err := repository.InitDB(ctx, cfg, logger)
	if err != nil {
		logger.Error(ctx, err, "Failed to initialize database", 500)
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize service
	logger.Info(ctx, "Creating service")
	svc := services.NewService(db, logger, cfg)

	// Create dependencies
	deps := NewDependencies(cfg, logger, db, svc)
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("dependency validation failed: %w", err)
	}

	// Setup middleware
	deps.SetupMiddleware()

	// Create gRPC server with middleware
	grpcServer := createGRPCServer(deps)

	// Register services
	registerServices(grpcServer, deps)

	// Enable reflection for development
	if cfg.Environment != config.PRODUCTION_ENV {
		reflection.Register(grpcServer)
	}

	// Create gRPC listener
	grpcListener, err := createGRPCListener(deps)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC listener: %w", err)
	}

	// Create REST gateway
	restGateway := http.NewRESTGateway(&deps.TransportConfig.Gateway, logger)
	// In Docker, both gRPC and REST services run in the same container
	// gRPC service runs on AuthServicePort, REST gateway connects to localhost:AuthServicePort
	grpcAddr := "localhost:" + cfg.AuthServicePort
	if err := restGateway.CreateGateway(ctx, grpcAddr, cfg.TLSEnabled, &deps.TransportConfig.TLS); err != nil {
		return nil, fmt.Errorf("failed to create REST gateway: %w", err)
	}

	// Create lifecycle manager
	lifecycle := lifecycle.NewManager(logger, &deps.TransportConfig.Health)
	lifecycle.SetServers(grpcServer, restGateway.GetServer(), grpcListener, restGateway.GetListener())

	return &Server{
		deps:         deps,
		grpcServer:   grpcServer,
		restGateway:  restGateway,
		lifecycle:    lifecycle,
		grpcListener: grpcListener,
		restListener: restGateway.GetListener(),
	}, nil
}

// createGRPCServer creates a gRPC server with configured middleware
func createGRPCServer(deps *Dependencies) *grpc.Server {
	var serverOptions []grpc.ServerOption

	// Add unary interceptors if available
	if unaryInterceptors := deps.Middleware.GetUnaryInterceptors(); len(unaryInterceptors) > 0 {
		serverOptions = append(serverOptions, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}

	// Add stream interceptors if available
	if streamInterceptors := deps.Middleware.GetStreamInterceptors(); len(streamInterceptors) > 0 {
		serverOptions = append(serverOptions, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	return grpc.NewServer(serverOptions...)
}

// createGRPCListener creates a gRPC listener with TLS if enabled
func createGRPCListener(deps *Dependencies) (net.Listener, error) {
	if deps.TLSManager.IsEnabled() {
		tlsConfig, err := deps.TLSManager.CreateTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS configuration: %w", err)
		}

		// Create TLS listener
		listener, err := tls.Listen("tcp", ":"+deps.Config.AuthServicePort, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS gRPC listener: %w", err)
		}

		deps.Logger.Info(context.Background(), "gRPC server configured with TLS", map[string]any{
			"port":    deps.Config.AuthServicePort,
			"min_tls": deps.Config.MinTLSVersion,
			"max_tls": deps.Config.MaxTLSVersion,
		})

		return listener, nil
	}

	// Create plain TCP listener
	listener, err := net.Listen("tcp", ":"+deps.Config.AuthServicePort)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC listener: %w", err)
	}

	deps.Logger.Info(context.Background(), "gRPC server configured without TLS", map[string]any{
		"port": deps.Config.AuthServicePort,
	})

	return listener, nil
}

// registerServices registers all gRPC services
func registerServices(grpcServer *grpc.Server, deps *Dependencies) {
	// Register all services using the centralized registration function
	RegisterServices(grpcServer, deps.Services, deps.Logger, deps.Database, deps.Config)
}

// Start runs both gRPC and REST servers
func (s *Server) Start(ctx context.Context) error {
	return s.lifecycle.Start(ctx)
}

// Shutdown gracefully shuts down both servers
func (s *Server) Shutdown(ctx context.Context) error {
	// Shutdown lifecycle manager
	if err := s.lifecycle.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown lifecycle manager: %w", err)
	}

	// Close dependencies
	if err := s.deps.Close(ctx); err != nil {
		return fmt.Errorf("failed to close dependencies: %w", err)
	}

	return nil
}

// Run starts both servers and handles graceful shutdown
func (s *Server) Run(ctx context.Context) error {
	return s.lifecycle.Run(ctx)
}

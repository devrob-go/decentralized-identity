package server

import (
	"context"
	"fmt"
	"time"

	"auth-service/config"
	transportConfig "auth-service/internal/config"
	"auth-service/internal/repository"
	"auth-service/internal/services"
	"auth-service/internal/transport/errors"
	"auth-service/internal/transport/middleware"
	"auth-service/internal/transport/tls"

	zlog "packages/logger"
)

// Dependencies holds all server dependencies
type Dependencies struct {
	Config          *config.Config
	TransportConfig *transportConfig.TransportConfig
	Logger          *zlog.Logger
	Database        *repository.DB
	Services        *services.Service
	Middleware      *middleware.Registry
	TLSManager      *tls.Manager
	ErrorMapper     *errors.ErrorMapper
}

// NewDependencies creates a new dependencies instance
func NewDependencies(cfg *config.Config, logger *zlog.Logger, db *repository.DB, svc *services.Service) *Dependencies {
	// Create transport configuration
	transportCfg := transportConfig.DefaultTransportConfig()

	// Override with actual config values
	if cfg.TLSEnabled {
		transportCfg.TLS.Enabled = true
		transportCfg.TLS.CertFile = cfg.TLSCertFile
		transportCfg.TLS.KeyFile = cfg.TLSKeyFile
		transportCfg.TLS.MinVersion = cfg.MinTLSVersion
		transportCfg.TLS.MaxVersion = cfg.MaxTLSVersion
	}

	transportCfg.Server.GRPCPort = cfg.AuthServicePort
	transportCfg.Server.ReadTimeout = time.Duration(cfg.ServerReadTimeout) * time.Second
	transportCfg.Server.WriteTimeout = time.Duration(cfg.ServerWriteTimeout) * time.Second
	transportCfg.Server.ShutdownTimeout = 5 * time.Second

	transportCfg.Gateway.RESTPort = cfg.RestGatewayPort
	transportCfg.Gateway.AllowedOrigins = cfg.AllowedOrigins

	transportCfg.Health.Timeout = time.Duration(cfg.HealthCheckTimeout) * time.Second
	transportCfg.Health.ReadinessDelay = 100 * time.Millisecond

	return &Dependencies{
		Config:          cfg,
		TransportConfig: transportCfg,
		Logger:          logger,
		Database:        db,
		Services:        svc,
		Middleware:      middleware.NewRegistry(),
		TLSManager:      tls.NewManager(&transportCfg.TLS),
		ErrorMapper:     errors.NewErrorMapper(logger),
	}
}

// SetupMiddleware configures all middleware in the correct order
func (d *Dependencies) SetupMiddleware() {
	// Create and register middleware in the correct order

	// 1. Recovery middleware (catches panics)
	recoveryMiddleware := middleware.NewRecoveryMiddleware(d.Logger)
	d.Middleware.AddUnary(recoveryMiddleware.UnaryRecoveryInterceptor())
	d.Middleware.AddStream(recoveryMiddleware.StreamRecoveryInterceptor())

	// 2. Metrics middleware (tracks performance)
	metricsMiddleware := middleware.NewMetricsMiddleware(d.Logger)
	d.Middleware.AddUnary(metricsMiddleware.UnaryMetricsInterceptor())
	d.Middleware.AddStream(metricsMiddleware.StreamMetricsInterceptor())

	// 3. Rate limiting middleware
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(d.Logger, d.Config)
	d.Middleware.AddUnary(rateLimitMiddleware.UnaryRateLimitInterceptor())
	d.Middleware.AddStream(rateLimitMiddleware.StreamRateLimitInterceptor())

	// 4. Security middleware (authentication and authorization)
	securityMiddleware := middleware.NewSecurityMiddleware(d.Logger, d.Config, d.Services)
	d.Middleware.AddUnary(securityMiddleware.UnarySecurityInterceptor())
	d.Middleware.AddStream(securityMiddleware.StreamSecurityInterceptor())

	d.Logger.Info(context.Background(), "Middleware setup completed", map[string]any{
		"middlewares": []string{"recovery", "metrics", "rate_limit", "security"},
	})
}

// Validate checks if all required dependencies are present
func (d *Dependencies) Validate() error {
	if d.Config == nil {
		return fmt.Errorf("config is required")
	}
	if d.Logger == nil {
		return fmt.Errorf("logger is required")
	}
	if d.Database == nil {
		return fmt.Errorf("database is required")
	}
	if d.Services == nil {
		return fmt.Errorf("services are required")
	}
	if d.Middleware == nil {
		return fmt.Errorf("middleware registry is required")
	}
	if d.TLSManager == nil {
		return fmt.Errorf("TLS manager is required")
	}
	if d.ErrorMapper == nil {
		return fmt.Errorf("error mapper is required")
	}
	return nil
}

// Close closes all dependencies that need cleanup
func (d *Dependencies) Close(ctx context.Context) error {
	if d.Database != nil {
		if err := d.Database.Close(ctx); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}
	return nil
}

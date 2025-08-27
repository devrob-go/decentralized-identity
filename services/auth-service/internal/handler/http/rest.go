package http

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"api/auth/v1/proto"
	"auth-service/internal/config"
	"auth-service/internal/transport/errors"

	zlog "packages/logger"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// RESTGateway handles the REST API gateway
type RESTGateway struct {
	config      *config.GatewayConfig
	logger      *zlog.Logger
	errorMapper *errors.ErrorMapper
	server      *http.Server
	listener    net.Listener
	grpcAddr    string
	tlsEnabled  bool
	tlsConfig   any
}

// NewRESTGateway creates a new REST gateway instance
func NewRESTGateway(cfg *config.GatewayConfig, logger *zlog.Logger) *RESTGateway {
	return &RESTGateway{
		config:      cfg,
		logger:      logger,
		errorMapper: errors.NewErrorMapper(logger),
	}
}

// CreateGateway creates the REST gateway server and listener
func (g *RESTGateway) CreateGateway(ctx context.Context, grpcAddr string, tlsEnabled bool, tlsConfig any) error {
	// Create REST listener
	restLis, err := net.Listen("tcp", ":"+g.config.RESTPort)
	if err != nil {
		return fmt.Errorf("failed to create REST listener: %w", err)
	}
	g.listener = restLis

	// Store connection parameters for dynamic connection creation
	g.grpcAddr = grpcAddr
	g.tlsEnabled = tlsEnabled
	g.tlsConfig = tlsConfig

	g.logger.Info(ctx, "Creating REST gateway", map[string]any{
		"rest_port":    g.config.RESTPort,
		"grpc_addr":    grpcAddr,
		"tls_enabled":  tlsEnabled,
		"gateway_type": "grpc_gateway",
	})

	// Create gRPC-Gateway mux with custom options
	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithErrorHandler(g.createErrorHandler()),
	)

	// Create custom HTTP mux to wrap gRPC gateway
	customMux := http.NewServeMux()

	// Register custom health endpoints
	g.registerCustomHealthEndpoints(customMux)

	// Register gRPC gateway handlers
	if err := g.registerHandlers(ctx, gwMux); err != nil {
		return fmt.Errorf("failed to register REST handlers: %w", err)
	}

	// Mount gRPC gateway under the custom mux
	customMux.Handle("/", gwMux)

	// Create HTTP server with proper timeout configurations
	g.server = &http.Server{
		Handler:           g.createMiddleware(customMux),
		Addr:              restLis.Addr().String(),
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	g.logger.Info(ctx, "REST gateway created successfully", map[string]any{
		"rest_port": g.config.RESTPort,
		"endpoints": []string{
			"/v1/health/direct",
			"/v1/health/grpc",
			"/v1/health/test-grpc",
			"/v1/health/service-info",
			"/health",
			"/v1/health",
		},
	})

	return nil
}

// createDialOptions creates gRPC dial options
func (g *RESTGateway) createDialOptions(tlsEnabled bool, tlsConfig any) []grpc.DialOption {
	var dialOptions []grpc.DialOption

	// Configure TLS if enabled
	if tlsEnabled {
		if tlsCfg, ok := tlsConfig.(*config.TLSConfig); ok {
			dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
				MinVersion: tlsCfg.MinVersion,
				MaxVersion: tlsCfg.MaxVersion,
			})))
		}
	} else {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	}

	// Simplified connection options for gRPC gateway
	dialOptions = append(dialOptions,
		// Basic load balancing
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		// Keep connections alive but be more conservative
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                60 * time.Second, // Longer interval for gateway
			Timeout:             10 * time.Second, // Longer timeout
			PermitWithoutStream: false,            // More conservative
		}),
		// Disable retry to avoid conflicts with gateway's own retry logic
		grpc.WithDisableRetry(),
	)

	return dialOptions
}

// createErrorHandler creates a custom error handler
func (g *RESTGateway) createErrorHandler() runtime.ErrorHandlerFunc {
	return func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		// Check for specific gRPC connection errors
		var statusCode int
		var errorMessage string
		var errorType string

		// Log the raw error for debugging
		g.logger.Info(ctx, "Raw error received", map[string]any{
			"error_type": fmt.Sprintf("%T", err),
			"error_msg":  err.Error(),
			"path":       r.URL.Path,
			"method":     r.Method,
			"ctx_done":   ctx.Err(),
		})

		// Check if context was cancelled
		if ctx.Err() != nil {
			g.logger.Warn(ctx, "Context was cancelled", map[string]any{
				"context_error": ctx.Err().Error(),
				"path":          r.URL.Path,
				"method":        r.Method,
			})
		}

		// Map gRPC errors to appropriate HTTP status codes
		if grpcErr, ok := err.(interface{ GRPCStatus() *status.Status }); ok {
			grpcStatus := grpcErr.GRPCStatus()
			g.logger.Info(ctx, "gRPC error details", map[string]any{
				"grpc_code":    grpcStatus.Code().String(),
				"grpc_message": grpcStatus.Message(),
				"grpc_details": grpcStatus.Details(),
			})

			switch grpcStatus.Code() {
			case codes.Canceled:
				// Check if this is actually a client cancellation or context timeout
				if ctx.Err() == context.Canceled {
					statusCode = 499 // Client Closed Request
					errorMessage = "Request cancelled by client"
					errorType = "client_cancelled"
				} else {
					statusCode = 500 // Internal Server Error
					errorMessage = "Request processing cancelled"
					errorType = "request_cancelled"
				}
			case codes.DeadlineExceeded:
				statusCode = 408 // Request Timeout
				errorMessage = "Request timeout"
				errorType = "timeout"
			case codes.Unavailable:
				statusCode = 503 // Service Unavailable
				errorMessage = "Service temporarily unavailable"
				errorType = "service_unavailable"
			case codes.Internal:
				statusCode = 500 // Internal Server Error
				errorMessage = "Internal server error"
				errorType = "internal_error"
			case codes.InvalidArgument:
				statusCode = 400 // Bad Request
				errorMessage = "Invalid request"
				errorType = "invalid_argument"
			case codes.Unauthenticated:
				statusCode = 401 // Unauthorized
				errorMessage = "Authentication required"
				errorType = "unauthenticated"
			case codes.PermissionDenied:
				statusCode = 403 // Forbidden
				errorMessage = "Permission denied"
				errorType = "permission_denied"
			case codes.NotFound:
				statusCode = 404 // Not Found
				errorMessage = "Resource not found"
				errorType = "not_found"
			case codes.AlreadyExists:
				statusCode = 409 // Conflict
				errorMessage = "Resource already exists"
				errorType = "conflict"
			case codes.ResourceExhausted:
				statusCode = 429 // Too Many Requests
				errorMessage = "Resource exhausted"
				errorType = "resource_exhausted"
			default:
				// For unknown gRPC codes, use the original error message
				statusCode = 500
				errorMessage = grpcStatus.Message()
				if errorMessage == "" {
					errorMessage = "gRPC service error"
				}
				errorType = "grpc_error"
			}
		} else {
			// Check if this is a context cancellation error
			if err == context.Canceled || err == context.DeadlineExceeded {
				statusCode = 499
				errorMessage = "Request cancelled by client"
				errorType = "context_cancelled"
			} else {
				// Use error mapper for non-gRPC errors
				statusCode, errorMessage = g.errorMapper.MapToHTTP(ctx, err, r.Method+" "+r.URL.Path)
				errorType = "http_error"
			}
		}

		// Log the error with context
		g.logger.Error(ctx, err, "REST gateway error", statusCode, map[string]any{
			"path":        r.URL.Path,
			"method":      r.Method,
			"status_code": statusCode,
			"error_type":  errorType,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		// Return structured error response
		errorResponse := map[string]interface{}{
			"error":       errorMessage,
			"status_code": statusCode,
			"timestamp":   time.Now().Format(time.RFC3339),
			"path":        r.URL.Path,
		}

		// Add retry information only for specific connection-related errors
		if statusCode == 503 || (statusCode == 499 && errorType == "client_cancelled") {
			errorResponse["retry_after"] = 5
			errorResponse["suggestion"] = "Try the /v1/health/direct endpoint for basic health check"
		}

		// Encode and send error response
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			// Fallback to simple error message if JSON encoding fails
			w.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, errorMessage)))
		}
	}
}

// registerHandlers registers all REST handlers
func (g *RESTGateway) registerHandlers(ctx context.Context, mux *runtime.ServeMux) error {
	// Register health service handlers
	if err := g.registerHealthHandlers(ctx, mux); err != nil {
		return fmt.Errorf("failed to register health handlers: %w", err)
	}

	// Register auth service handlers
	if err := g.registerAuthHandlers(ctx, mux); err != nil {
		return fmt.Errorf("failed to register auth handlers: %w", err)
	}

	return nil
}

// registerCustomHealthEndpoints registers custom health endpoints that don't depend on gRPC
func (g *RESTGateway) registerCustomHealthEndpoints(mux *http.ServeMux) {
	// Add a direct health endpoint that doesn't depend on gRPC
	mux.HandleFunc("/v1/health/direct", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"SERVING","service":"auth-service","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Add a root health endpoint for basic connectivity testing
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"SERVING","service":"auth-service","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Add a gRPC connection health check endpoint
	mux.HandleFunc("/v1/health/grpc", func(w http.ResponseWriter, r *http.Request) {
		status := g.checkGRPCConnectionHealth()
		w.Header().Set("Content-Type", "application/json")
		if status {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"SERVING","grpc_connection":"healthy","service":"auth-service","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"NOT_SERVING","grpc_connection":"unhealthy","service":"auth-service","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
		}
	})

	// Add a test endpoint that directly calls the gRPC health service
	mux.HandleFunc("/v1/health/test-grpc", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Create a direct gRPC connection
		dialOptions := g.createDialOptions(g.tlsEnabled, g.tlsConfig)
		conn, err := grpc.DialContext(ctx, g.grpcAddr, dialOptions...)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":       "Failed to connect to gRPC service",
				"grpc_addr":   g.grpcAddr,
				"tls_enabled": g.tlsEnabled,
				"timestamp":   time.Now().Format(time.RFC3339),
			})
			return
		}
		defer conn.Close()

		// Test the health service directly
		healthClient := proto.NewHealthClient(conn)
		resp, err := healthClient.Check(ctx, &proto.HealthCheckRequest{})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":      "gRPC health service call failed",
				"grpc_addr":  g.grpcAddr,
				"grpc_error": err.Error(),
				"timestamp":  time.Now().Format(time.RFC3339),
			})
			return
		}

		// Success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "SERVING",
			"grpc_status": resp.Status.String(),
			"grpc_addr":   g.grpcAddr,
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	})

	// Add a service discovery test endpoint
	mux.HandleFunc("/v1/health/service-info", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Create a direct gRPC connection
		dialOptions := g.createDialOptions(g.tlsEnabled, g.tlsConfig)
		conn, err := grpc.DialContext(ctx, g.grpcAddr, dialOptions...)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":       "Failed to connect to gRPC service",
				"grpc_addr":   g.grpcAddr,
				"tls_enabled": g.tlsEnabled,
				"timestamp":   time.Now().Format(time.RFC3339),
			})
			return
		}
		defer conn.Close()

		// Get connection state and target info
		state := conn.GetState()
		target := conn.Target()

		// Test service availability
		healthClient := proto.NewHealthClient(conn)
		_, healthErr := healthClient.Check(ctx, &proto.HealthCheckRequest{})

		// Return service information
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"grpc_addr":         g.grpcAddr,
			"connection_target": target,
			"connection_state":  state.String(),
			"tls_enabled":       g.tlsEnabled,
			"health_service": map[string]interface{}{
				"available": healthErr == nil,
				"error": func() string {
					if healthErr != nil {
						return healthErr.Error()
					}
					return ""
				}(),
			},
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})
}

// checkGRPCConnectionHealth checks if the gRPC connection is healthy
func (g *RESTGateway) checkGRPCConnectionHealth() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a temporary connection to test health
	dialOptions := g.createDialOptions(g.tlsEnabled, g.tlsConfig)

	g.logger.Info(ctx, "Attempting gRPC connection test", map[string]any{
		"grpc_addr": g.grpcAddr,
		"tls":       g.tlsEnabled,
		"timeout":   "5s",
	})

	conn, err := grpc.DialContext(ctx, g.grpcAddr, dialOptions...)
	if err != nil {
		g.logger.Error(ctx, err, "gRPC connection health check failed", 500, map[string]any{
			"grpc_addr": g.grpcAddr,
			"tls":       g.tlsEnabled,
			"timeout":   "5s",
		})
		return false
	}
	defer conn.Close()

	// Check connection state
	state := conn.GetState()
	g.logger.Info(ctx, "gRPC connection state", map[string]any{
		"grpc_addr": g.grpcAddr,
		"state":     state.String(),
	})

	if state != connectivity.Idle && state != connectivity.Ready {
		g.logger.Warn(ctx, "gRPC connection not ready", map[string]any{
			"grpc_addr": g.grpcAddr,
			"state":     state.String(),
		})
		return false
	}

	// Try to connect to the health service
	healthClient := proto.NewHealthClient(conn)
	healthCtx, healthCancel := context.WithTimeout(ctx, 3*time.Second)
	defer healthCancel()

	_, err = healthClient.Check(healthCtx, &proto.HealthCheckRequest{})
	if err != nil {
		g.logger.Error(ctx, err, "gRPC health service check failed", 500, map[string]any{
			"grpc_addr": g.grpcAddr,
			"service":   "health",
			"timeout":   "3s",
		})
		return false
	}

	// Log successful health check (only occasionally to avoid spam)
	if time.Now().Unix()%300 == 0 { // Log every 5 minutes
		g.logger.Info(ctx, "gRPC connection health check passed", map[string]any{
			"grpc_addr": g.grpcAddr,
			"state":     state.String(),
		})
	}

	return true
}

// registerHealthHandlers registers health service REST handlers
func (g *RESTGateway) registerHealthHandlers(ctx context.Context, mux *runtime.ServeMux) error {
	g.logger.Info(ctx, "Registering health service handlers", map[string]any{
		"grpc_addr": g.grpcAddr,
		"tls":       g.tlsEnabled,
	})

	// Create a shared connection for the health service
	conn, err := grpc.DialContext(ctx, g.grpcAddr, g.createDialOptions(g.tlsEnabled, g.tlsConfig)...)
	if err != nil {
		g.logger.Error(ctx, err, "Failed to create gRPC connection for health handlers", 500, map[string]any{
			"grpc_addr": g.grpcAddr,
			"service":   "health",
		})
		return fmt.Errorf("failed to create gRPC connection for health handlers: %w", err)
	}

	// Register health service REST handlers with shared connection
	if err := proto.RegisterHealthHandler(ctx, mux, conn); err != nil {
		conn.Close()
		g.logger.Error(ctx, err, "Failed to register health handlers", 500, map[string]any{
			"grpc_addr": g.grpcAddr,
			"service":   "health",
		})
		return fmt.Errorf("failed to register health handlers: %w", err)
	}

	g.logger.Info(ctx, "Health service handlers registered successfully", map[string]any{
		"grpc_addr": g.grpcAddr,
		"service":   "health",
	})
	return nil
}

// registerAuthHandlers registers auth service REST handlers
func (g *RESTGateway) registerAuthHandlers(ctx context.Context, mux *runtime.ServeMux) error {
	g.logger.Info(ctx, "Registering auth service handlers", map[string]any{
		"grpc_addr": g.grpcAddr,
		"tls":       g.tlsEnabled,
	})

	// Create a shared connection for the auth service
	conn, err := grpc.DialContext(ctx, g.grpcAddr, g.createDialOptions(g.tlsEnabled, g.tlsConfig)...)
	if err != nil {
		g.logger.Error(ctx, err, "Failed to create gRPC connection for auth handlers", 500, map[string]any{
			"grpc_addr": g.grpcAddr,
			"service":   "auth",
		})
		return fmt.Errorf("failed to create gRPC connection for auth handlers: %w", err)
	}

	// Register auth service REST handlers with shared connection
	if err := proto.RegisterAuthServiceHandler(ctx, mux, conn); err != nil {
		conn.Close()
		g.logger.Error(ctx, err, "Failed to register auth handlers", 500, map[string]any{
			"grpc_addr": g.grpcAddr,
			"service":   "auth",
		})
		return fmt.Errorf("failed to register auth handlers: %w", err)
	}

	g.logger.Info(ctx, "Auth service handlers registered successfully", map[string]any{
		"grpc_addr": g.grpcAddr,
		"service":   "auth",
	})
	return nil
}

// createMiddleware creates middleware for the REST gateway
func (g *RESTGateway) createMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers with proper origin validation
		origin := r.Header.Get("Origin")
		if origin != "" && g.isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", g.joinStrings(g.config.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", g.joinStrings(g.config.AllowedHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", g.config.MaxAge))

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Add request logging
		g.logger.Info(r.Context(), "REST request", map[string]any{
			"method":     r.Method,
			"path":       r.URL.Path,
			"remote":     r.RemoteAddr,
			"user_agent": r.UserAgent(),
		})

		// Add response logging
		responseWriter := &responseWriter{ResponseWriter: w, statusCode: 200}
		handler.ServeHTTP(responseWriter, r)

		// Log response
		g.logger.Info(r.Context(), "REST response", map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": responseWriter.statusCode,
		})
	})
}

// isAllowedOrigin checks if an origin is in the allowed origins list
func (g *RESTGateway) isAllowedOrigin(origin string) bool {
	for _, allowed := range g.config.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// joinStrings joins a slice of strings with a separator
func (g *RESTGateway) joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for _, str := range strs[1:] {
		result += sep + str
	}
	return result
}

// Stop stops the REST gateway server
func (g *RESTGateway) Stop(ctx context.Context) error {
	if g.server == nil {
		return nil
	}
	return g.server.Shutdown(ctx)
}

// GetServer returns the HTTP server instance
func (g *RESTGateway) GetServer() *http.Server {
	return g.server
}

// GetListener returns the REST listener
func (g *RESTGateway) GetListener() net.Listener {
	return g.listener
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

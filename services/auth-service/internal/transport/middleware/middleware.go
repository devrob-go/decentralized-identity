package middleware

import (
	"context"
	"time"

	zlog "packages/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"auth-service/config"
)

// MetricsMiddleware provides basic metrics collection
type MetricsMiddleware struct {
	requestCount int64
	errorCount   int64
	responseTime time.Duration
	logger       *zlog.Logger
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(logger *zlog.Logger) *MetricsMiddleware {
	return &MetricsMiddleware{
		logger: logger,
	}
}

// UnaryMetricsInterceptor collects metrics for unary RPC calls
func (m *MetricsMiddleware) UnaryMetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		// Extract correlation ID from metadata
		correlationID := extractCorrelationID(ctx)
		ctx = zlog.WithCorrelationID(ctx, correlationID)

		// Log request start
		m.logger.Info(ctx, "gRPC request started", map[string]any{
			"method":         info.FullMethod,
			"correlation_id": correlationID,
		})

		// Handle the request
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Log response
		if err != nil {
			st, _ := status.FromError(err)
			m.logger.Error(ctx, err, "gRPC request failed", int(st.Code()), map[string]any{
				"method":         info.FullMethod,
				"duration":       duration.String(),
				"status_code":    st.Code(),
				"correlation_id": correlationID,
			})
		} else {
			m.logger.Info(ctx, "gRPC request completed", map[string]any{
				"method":         info.FullMethod,
				"duration":       duration.String(),
				"status_code":    codes.OK,
				"correlation_id": correlationID,
			})
		}

		return resp, err
	}
}

// StreamMetricsInterceptor collects metrics for streaming RPC calls
func (m *MetricsMiddleware) StreamMetricsInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Extract correlation ID from metadata
		correlationID := extractCorrelationID(ss.Context())
		ctx := zlog.WithCorrelationID(ss.Context(), correlationID)

		// Create wrapped stream with correlation ID
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Log stream start
		m.logger.Info(ctx, "gRPC stream started", map[string]any{
			"method":         info.FullMethod,
			"correlation_id": correlationID,
		})

		// Handle the stream
		err := handler(srv, wrappedStream)

		// Calculate duration
		duration := time.Since(start)

		// Log stream completion
		if err != nil {
			st, _ := status.FromError(err)
			m.logger.Error(ctx, err, "gRPC stream failed", int(st.Code()), map[string]any{
				"method":         info.FullMethod,
				"duration":       duration.String(),
				"status_code":    st.Code(),
				"correlation_id": correlationID,
			})
		} else {
			m.logger.Info(ctx, "gRPC stream completed", map[string]any{
				"method":         info.FullMethod,
				"duration":       duration.String(),
				"status_code":    codes.OK,
				"correlation_id": correlationID,
			})
		}

		return err
	}
}

// RecoveryMiddleware provides panic recovery for gRPC calls
type RecoveryMiddleware struct {
	logger *zlog.Logger
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(logger *zlog.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

// UnaryRecoveryInterceptor recovers from panics in unary RPC calls
func (r *RecoveryMiddleware) UnaryRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		defer func() {
			if rec := recover(); rec != nil {
				r.logger.Error(ctx, nil, "panic recovered in unary call", 500, map[string]any{
					"method": info.FullMethod,
					"panic":  rec,
				})
			}
		}()

		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor recovers from panics in streaming RPC calls
func (r *RecoveryMiddleware) StreamRecoveryInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		defer func() {
			if rec := recover(); rec != nil {
				r.logger.Error(ss.Context(), nil, "panic recovered in stream call", 500, map[string]any{
					"method": info.FullMethod,
					"panic":  rec,
				})
			}
		}()

		return handler(srv, ss)
	}
}

// RateLimitMiddleware provides basic rate limiting (placeholder for future implementation)
type RateLimitMiddleware struct {
	logger *zlog.Logger
	config *config.Config
	// In-memory rate limiter (for production, use Redis or similar)
	clients map[string]*clientLimiter
}

// clientLimiter tracks rate limiting for a specific client
type clientLimiter struct {
	requests []time.Time
	window   time.Duration
	limit    int
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(logger *zlog.Logger, cfg *config.Config) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		logger:  logger,
		config:  cfg,
		clients: make(map[string]*clientLimiter),
	}
}

// UnaryRateLimitInterceptor provides rate limiting for unary RPC calls
func (rl *RateLimitMiddleware) UnaryRateLimitInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !rl.config.RateLimitEnabled {
			return handler(ctx, req)
		}

		clientID := extractClientID(ctx)
		if !rl.allowRequest(clientID) {
			rl.logger.Warn(ctx, "Rate limit exceeded", map[string]any{
				"client_id": clientID,
				"method":    info.FullMethod,
			})
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// StreamRateLimitInterceptor provides rate limiting for streaming RPC calls
func (rl *RateLimitMiddleware) StreamRateLimitInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !rl.config.RateLimitEnabled {
			return handler(srv, ss)
		}

		clientID := extractClientID(ss.Context())
		if !rl.allowRequest(clientID) {
			rl.logger.Warn(ss.Context(), "Rate limit exceeded for stream", map[string]any{
				"client_id": clientID,
				"method":    info.FullMethod,
			})
			return status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(srv, ss)
	}
}

// allowRequest checks if a request should be allowed based on rate limiting
func (rl *RateLimitMiddleware) allowRequest(clientID string) bool {
	now := time.Now()
	window := time.Duration(rl.config.RateLimitWindow) * time.Second
	limit := rl.config.RateLimitRequests

	// Get or create client limiter
	limiter, exists := rl.clients[clientID]
	if !exists {
		limiter = &clientLimiter{
			requests: make([]time.Time, 0),
			window:   window,
			limit:    limit,
		}
		rl.clients[clientID] = limiter
	}

	// Remove expired requests
	var validRequests []time.Time
	for _, reqTime := range limiter.requests {
		if now.Sub(reqTime) <= window {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if we're under the limit
	if len(validRequests) < limit {
		validRequests = append(validRequests, now)
		limiter.requests = validRequests
		return true
	}

	return false
}

// extractClientID extracts client identifier from context
func extractClientID(ctx context.Context) string {
	// In a real implementation, this would extract from JWT token, IP, or other identifier
	// For now, use a simple approach
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if clientIDs := md.Get("client-id"); len(clientIDs) > 0 {
			return clientIDs[0]
		}
		if userAgents := md.Get("user-agent"); len(userAgents) > 0 {
			return userAgents[0]
		}
	}
	return "unknown"
}

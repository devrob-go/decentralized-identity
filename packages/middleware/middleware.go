package middleware

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTPMiddleware represents HTTP middleware functions
type HTTPMiddleware struct{}

// NewHTTPMiddleware creates a new HTTP middleware instance
func NewHTTPMiddleware() *HTTPMiddleware {
	return &HTTPMiddleware{}
}

// Logging logs HTTP requests with timing information
func (m *HTTPMiddleware) Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			// Log the request details
			logHTTPRequest(r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

// CORS adds CORS headers to HTTP responses
func (m *HTTPMiddleware) CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Recovery recovers from panics in HTTP handlers
func (m *HTTPMiddleware) Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logPanic(r, err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// gRPC Middleware

// LoggingUnary logs gRPC unary requests
func LoggingUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		// Log the gRPC request
		logGRPCRequest(info.FullMethod, err, duration)

		return resp, err
	}
}

// RecoveryUnary recovers from panics in gRPC unary handlers
func RecoveryUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		defer func() {
			if err := recover(); err != nil {
				logPanicGRPC(info.FullMethod, err)
			}
		}()

		return handler(ctx, req)
	}
}

// AuthUnary adds authentication to gRPC unary handlers
func AuthUnary(authFunc func(ctx context.Context) error) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := authFunc(ctx); err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		return handler(ctx, req)
	}
}

// Helper types and functions

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

// Placeholder logging functions - these should be implemented with your logging library
func logHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	// TODO: Implement with your logging library
}

func logGRPCRequest(method string, err error, duration time.Duration) {
	// TODO: Implement with your logging library
}

func logPanic(r *http.Request, err any) {
	// TODO: Implement with your logging library
}

func logPanicGRPC(method string, err any) {
	// TODO: Implement with your logging library
}

package middleware

import (
	"context"
	"testing"
	"time"

	"auth-service/config"
	zlog "packages/logger"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// testConfig returns a minimal configuration for rate limit tests
func testConfig() *config.Config {
	return &config.Config{
		RateLimitEnabled:  true,
		RateLimitRequests: 100,
		RateLimitWindow:   60,
	}
}

func TestNewMetricsMiddleware(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	middleware := NewMetricsMiddleware(logger)

	assert.NotNil(t, middleware)
	assert.Equal(t, logger, middleware.logger)
	assert.Equal(t, int64(0), middleware.requestCount)
	assert.Equal(t, int64(0), middleware.errorCount)
	assert.Equal(t, time.Duration(0), middleware.responseTime)
}

func TestMetricsMiddleware_UnaryMetricsInterceptor(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	middleware := NewMetricsMiddleware(logger)

	interceptor := middleware.UnaryMetricsInterceptor()
	assert.NotNil(t, interceptor)

	// Test that the interceptor can be called
	ctx := context.Background()
	req := "test-request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/TestMethod",
	}

	handler := func(ctx context.Context, req any) (any, error) {
		return "test-response", nil
	}

	// Call the interceptor
	resp, err := interceptor(ctx, req, info, handler)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "test-response", resp)
}

func TestMetricsMiddleware_UnaryMetricsInterceptor_Error(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	middleware := NewMetricsMiddleware(logger)

	interceptor := middleware.UnaryMetricsInterceptor()
	assert.NotNil(t, interceptor)

	// Test that the interceptor can be called with error
	ctx := context.Background()
	req := "test-request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/TestMethod",
	}

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, status.Errorf(codes.Internal, "test error")
	}

	// Call the interceptor
	resp, err := interceptor(ctx, req, info, handler)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Check gRPC status
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestMetricsMiddleware_StreamMetricsInterceptor(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	middleware := NewMetricsMiddleware(logger)

	interceptor := middleware.StreamMetricsInterceptor()
	assert.NotNil(t, interceptor)

	// Test that the interceptor can be called
	ctx := context.Background()
	info := &grpc.StreamServerInfo{
		FullMethod:     "/test.Service/TestStream",
		IsServerStream: true,
		IsClientStream: false,
	}

	handler := func(srv any, stream grpc.ServerStream) error {
		return nil
	}

	// Create a mock stream
	mockStream := &MockServerStream{
		ctx: ctx,
	}

	// Call the interceptor
	err := interceptor(nil, mockStream, info, handler)

	// Assertions
	assert.NoError(t, err)
}

func TestNewRecoveryMiddleware(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	middleware := NewRecoveryMiddleware(logger)

	assert.NotNil(t, middleware)
	assert.Equal(t, logger, middleware.logger)
}

func TestRecoveryMiddleware_UnaryRecoveryInterceptor(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	middleware := NewRecoveryMiddleware(logger)

	interceptor := middleware.UnaryRecoveryInterceptor()
	assert.NotNil(t, interceptor)

	// Test that the interceptor can be called
	ctx := context.Background()
	req := "test-request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/TestMethod",
	}

	handler := func(ctx context.Context, req any) (any, error) {
		return "test-response", nil
	}

	// Call the interceptor
	resp, err := interceptor(ctx, req, info, handler)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "test-response", resp)
}

func TestRecoveryMiddleware_StreamRecoveryInterceptor(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	middleware := NewRecoveryMiddleware(logger)

	interceptor := middleware.StreamRecoveryInterceptor()
	assert.NotNil(t, interceptor)

	// Test that the interceptor can be called
	ctx := context.Background()
	info := &grpc.StreamServerInfo{
		FullMethod:     "/test.Service/TestStream",
		IsServerStream: true,
		IsClientStream: false,
	}

	handler := func(srv any, stream grpc.ServerStream) error {
		return nil
	}

	// Create a mock stream
	mockStream := &MockServerStream{
		ctx: ctx,
	}

	// Call the interceptor
	err := interceptor(nil, mockStream, info, handler)

	// Assertions
	assert.NoError(t, err)
}

func TestNewRateLimitMiddleware(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	middleware := NewRateLimitMiddleware(logger, testConfig())

	assert.NotNil(t, middleware)
	assert.Equal(t, logger, middleware.logger)
}

func TestRateLimitMiddleware_UnaryRateLimitInterceptor(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	middleware := NewRateLimitMiddleware(logger, testConfig())

	interceptor := middleware.UnaryRateLimitInterceptor()
	assert.NotNil(t, interceptor)

	// Test that the interceptor can be called
	ctx := context.Background()
	req := "test-request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/TestMethod",
	}

	handler := func(ctx context.Context, req any) (any, error) {
		return "test-response", nil
	}

	// Call the interceptor
	resp, err := interceptor(ctx, req, info, handler)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "test-response", resp)
}

func TestRateLimitMiddleware_StreamRateLimitInterceptor(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	middleware := NewRateLimitMiddleware(logger, testConfig())

	interceptor := middleware.StreamRateLimitInterceptor()
	assert.NotNil(t, interceptor)

	// Test that the interceptor can be called
	ctx := context.Background()
	info := &grpc.StreamServerInfo{
		FullMethod:     "/test.Service/TestStream",
		IsServerStream: true,
		IsClientStream: false,
	}

	handler := func(srv any, stream grpc.ServerStream) error {
		return nil
	}

	// Create a mock stream
	mockStream := &MockServerStream{
		ctx: ctx,
	}

	// Call the interceptor
	err := interceptor(nil, mockStream, info, handler)

	// Assertions
	assert.NoError(t, err)
}

func TestMiddleware_Structure(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	// Test MetricsMiddleware structure
	metricsMiddleware := NewMetricsMiddleware(logger)
	assert.NotNil(t, metricsMiddleware)
	assert.NotNil(t, metricsMiddleware.logger)
	assert.Equal(t, logger, metricsMiddleware.logger)

	// Test RecoveryMiddleware structure
	recoveryMiddleware := NewRecoveryMiddleware(logger)
	assert.NotNil(t, recoveryMiddleware)
	assert.NotNil(t, recoveryMiddleware.logger)
	assert.Equal(t, logger, recoveryMiddleware.logger)

	// Test RateLimitMiddleware structure
	rateLimitMiddleware := NewRateLimitMiddleware(logger, testConfig())
	assert.NotNil(t, rateLimitMiddleware)
	assert.NotNil(t, rateLimitMiddleware.logger)
	assert.Equal(t, logger, rateLimitMiddleware.logger)
}

func TestMiddleware_MethodSignatures(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	// Test MetricsMiddleware methods
	metricsMiddleware := NewMetricsMiddleware(logger)
	_ = metricsMiddleware.UnaryMetricsInterceptor
	_ = metricsMiddleware.StreamMetricsInterceptor

	// Test RecoveryMiddleware methods
	recoveryMiddleware := NewRecoveryMiddleware(logger)
	_ = recoveryMiddleware.UnaryRecoveryInterceptor
	_ = recoveryMiddleware.StreamRecoveryInterceptor

	// Test RateLimitMiddleware methods
	rateLimitMiddleware := NewRateLimitMiddleware(logger, testConfig())
	_ = rateLimitMiddleware.UnaryRateLimitInterceptor
	_ = rateLimitMiddleware.StreamRateLimitInterceptor

	// If we get here, all methods exist
	assert.True(t, true)
}

// MockServerStream is a mock implementation for testing
type MockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *MockServerStream) Context() context.Context {
	return m.ctx
}

// Benchmark tests for performance
func BenchmarkMetricsMiddleware_UnaryMetricsInterceptor(b *testing.B) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	middleware := NewMetricsMiddleware(logger)
	interceptor := middleware.UnaryMetricsInterceptor()

	ctx := context.Background()
	req := "test-request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/TestMethod",
	}

	handler := func(ctx context.Context, req any) (any, error) {
		return "test-response", nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interceptor(ctx, req, info, handler)
	}
}

func BenchmarkRecoveryMiddleware_UnaryRecoveryInterceptor(b *testing.B) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	middleware := NewRecoveryMiddleware(logger)
	interceptor := middleware.UnaryRecoveryInterceptor()

	ctx := context.Background()
	req := "test-request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/TestMethod",
	}

	handler := func(ctx context.Context, req any) (any, error) {
		return "test-response", nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interceptor(ctx, req, info, handler)
	}
}

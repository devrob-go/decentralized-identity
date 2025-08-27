package middleware

import (
	"context"
	"time"

	zlog "packages/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryLoggingInterceptor provides logging for unary RPC calls
func UnaryLoggingInterceptor(logger *zlog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		// Extract correlation ID from metadata or create new one
		correlationID := extractCorrelationID(ctx)
		ctx = zlog.WithCorrelationID(ctx, correlationID)

		// Log request
		logger.Info(ctx, "gRPC request started", map[string]any{
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
			logger.Error(ctx, err, "gRPC request failed", int(st.Code()), map[string]any{
				"method":         info.FullMethod,
				"duration":       duration.String(),
				"status_code":    st.Code(),
				"correlation_id": correlationID,
			})
		} else {
			logger.Info(ctx, "gRPC request completed", map[string]any{
				"method":         info.FullMethod,
				"duration":       duration.String(),
				"status_code":    codes.OK,
				"correlation_id": correlationID,
			})
		}

		return resp, err
	}
}

// StreamLoggingInterceptor provides logging for streaming RPC calls
func StreamLoggingInterceptor(logger *zlog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Extract correlation ID from metadata or create new one
		correlationID := extractCorrelationID(ss.Context())
		ctx := zlog.WithCorrelationID(ss.Context(), correlationID)

		// Create wrapped stream with correlation ID
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Log stream start
		logger.Info(ctx, "gRPC stream started", map[string]any{
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
			logger.Error(ctx, err, "gRPC stream failed", int(st.Code()), map[string]any{
				"method":         info.FullMethod,
				"duration":       duration.String(),
				"status_code":    st.Code(),
				"correlation_id": correlationID,
			})
		} else {
			logger.Info(ctx, "gRPC stream completed", map[string]any{
				"method":         info.FullMethod,
				"duration":       duration.String(),
				"status_code":    codes.OK,
				"correlation_id": correlationID,
			})
		}

		return err
	}
}

// extractCorrelationID extracts correlation ID from gRPC metadata
func extractCorrelationID(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if correlationIDs := md.Get("x-correlation-id"); len(correlationIDs) > 0 {
			return correlationIDs[0]
		}
	}
	return ""
}

// wrappedServerStream wraps grpc.ServerStream to include correlation ID
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

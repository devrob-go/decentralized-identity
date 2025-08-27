package errors

import (
	"context"

	zlog "packages/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorMapper provides centralized error handling and mapping
type ErrorMapper struct {
	logger *zlog.Logger
}

// NewErrorMapper creates a new error mapper instance
func NewErrorMapper(logger *zlog.Logger) *ErrorMapper {
	return &ErrorMapper{
		logger: logger,
	}
}

// MapToGRPC maps internal errors to appropriate gRPC status codes
func (e *ErrorMapper) MapToGRPC(ctx context.Context, err error, method string) error {
	if err == nil {
		return nil
	}

	// Log the error for debugging
	e.logger.Error(ctx, err, "Error occurred", 500, map[string]any{
		"method": method,
		"error":  err.Error(),
	})

	// Check if it's already a gRPC status error
	if st, ok := status.FromError(err); ok {
		return st.Err()
	}

	// Map common error types to gRPC status codes
	switch {
	case isValidationError(err):
		return status.Error(codes.InvalidArgument, err.Error())
	case isAuthenticationError(err):
		return status.Error(codes.Unauthenticated, err.Error())
	case isAuthorizationError(err):
		return status.Error(codes.PermissionDenied, err.Error())
	case isNotFoundError(err):
		return status.Error(codes.NotFound, err.Error())
	case isConflictError(err):
		return status.Error(codes.AlreadyExists, err.Error())
	case isTimeoutError(err):
		return status.Error(codes.DeadlineExceeded, err.Error())
	case isDatabaseError(err):
		return status.Error(codes.Internal, "database operation failed")
	case isNetworkError(err):
		return status.Error(codes.Unavailable, "service temporarily unavailable")
	default:
		// For unknown errors, return internal error with sanitized message
		return status.Error(codes.Internal, "internal server error")
	}
}

// MapToHTTP maps internal errors to appropriate HTTP status codes and messages
func (e *ErrorMapper) MapToHTTP(ctx context.Context, err error, method string) (int, string) {
	if err == nil {
		return 200, ""
	}

	// Log the error for debugging
	e.logger.Error(ctx, err, "Error occurred", 500, map[string]any{
		"method": method,
		"error":  err.Error(),
	})

	// Check if it's already a gRPC status error
	if st, ok := status.FromError(err); ok {
		return mapGRPCCodeToHTTP(st.Code()), st.Message()
	}

	// Map common error types to HTTP status codes
	switch {
	case isValidationError(err):
		return 400, err.Error()
	case isAuthenticationError(err):
		return 401, err.Error()
	case isAuthorizationError(err):
		return 403, err.Error()
	case isNotFoundError(err):
		return 404, err.Error()
	case isConflictError(err):
		return 409, err.Error()
	case isTimeoutError(err):
		return 408, err.Error()
	case isDatabaseError(err):
		return 500, "internal server error"
	case isNetworkError(err):
		return 503, "service temporarily unavailable"
	default:
		return 500, "internal server error"
	}
}

// mapGRPCCodeToHTTP maps gRPC status codes to HTTP status codes
func mapGRPCCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.OK:
		return 200
	case codes.Canceled:
		return 499
	case codes.Unknown:
		return 500
	case codes.InvalidArgument:
		return 400
	case codes.DeadlineExceeded:
		return 408
	case codes.NotFound:
		return 404
	case codes.AlreadyExists:
		return 409
	case codes.PermissionDenied:
		return 403
	case codes.ResourceExhausted:
		return 429
	case codes.FailedPrecondition:
		return 400
	case codes.Aborted:
		return 409
	case codes.OutOfRange:
		return 400
	case codes.Unimplemented:
		return 501
	case codes.Internal:
		return 500
	case codes.Unavailable:
		return 503
	case codes.DataLoss:
		return 500
	case codes.Unauthenticated:
		return 401
	default:
		return 500
	}
}

// Helper functions to categorize errors
func isValidationError(err error) bool {
	// Add logic to identify validation errors
	return false
}

func isAuthenticationError(err error) bool {
	// Add logic to identify authentication errors
	return false
}

func isAuthorizationError(err error) bool {
	// Add logic to identify authorization errors
	return false
}

func isNotFoundError(err error) bool {
	// Add logic to identify not found errors
	return false
}

func isConflictError(err error) bool {
	// Add logic to identify conflict errors
	return false
}

func isTimeoutError(err error) bool {
	// Add logic to identify timeout errors
	return false
}

func isDatabaseError(err error) bool {
	// Add logic to identify database errors
	return false
}

func isNetworkError(err error) bool {
	// Add logic to identify network errors
	return false
}

// CreateErrorResponse creates a standardized error response
func (e *ErrorMapper) CreateErrorResponse(ctx context.Context, err error, method string) map[string]any {
	statusCode, message := e.MapToHTTP(ctx, err, method)

	return map[string]any{
		"error": map[string]any{
			"code":    statusCode,
			"message": message,
			"method":  method,
		},
		"timestamp": "2024-01-01T00:00:00Z", // This should be actual timestamp
	}
}

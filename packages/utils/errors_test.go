package utils

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		err      error
		expected *AppError
	}{
		{
			name:    "basic validation error",
			code:    "INVALID_INPUT",
			message: "test validation error",
			err:     errors.New("validation failed"),
			expected: &AppError{
				Type:       ErrorTypeValidation,
				Code:       "INVALID_INPUT",
				Message:    "test validation error",
				StatusCode: http.StatusBadRequest,
				Err:        errors.New("validation failed"),
			},
		},
		{
			name:    "validation error without underlying error",
			code:    "MISSING_FIELD",
			message: "field is required",
			err:     nil,
			expected: &AppError{
				Type:       ErrorTypeValidation,
				Code:       "MISSING_FIELD",
				Message:    "field is required",
				StatusCode: http.StatusBadRequest,
				Err:        nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewValidationError(tt.code, tt.message, tt.err)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Code, result.Code)
			assert.Equal(t, tt.expected.Message, result.Message)
			assert.Equal(t, tt.expected.StatusCode, result.StatusCode)
			if tt.err != nil {
				assert.Equal(t, tt.err.Error(), result.Err.Error())
			} else {
				assert.Nil(t, result.Err)
			}
		})
	}
}

func TestNewNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		err      error
		expected *AppError
	}{
		{
			name:    "basic not found error",
			code:    "USER_NOT_FOUND",
			message: "user not found",
			err:     errors.New("user does not exist"),
			expected: &AppError{
				Type:       ErrorTypeNotFound,
				Code:       "USER_NOT_FOUND",
				Message:    "user not found",
				StatusCode: http.StatusNotFound,
				Err:        errors.New("user does not exist"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewNotFoundError(tt.code, tt.message, tt.err)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Code, result.Code)
			assert.Equal(t, tt.expected.Message, result.Message)
			assert.Equal(t, tt.expected.StatusCode, result.StatusCode)
			if tt.err != nil {
				assert.Equal(t, tt.err.Error(), result.Err.Error())
			}
		})
	}
}

func TestNewInternalError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		err      error
		expected *AppError
	}{
		{
			name:    "basic internal error",
			code:    "INTERNAL_ERROR",
			message: "internal server error",
			err:     errors.New("database connection failed"),
			expected: &AppError{
				Type:       ErrorTypeInternal,
				Code:       "INTERNAL_ERROR",
				Message:    "internal server error",
				StatusCode: http.StatusInternalServerError,
				Err:        errors.New("database connection failed"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewInternalError(tt.code, tt.message, tt.err)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Code, result.Code)
			assert.Equal(t, tt.expected.Message, result.Message)
			assert.Equal(t, tt.expected.StatusCode, result.StatusCode)
			if tt.err != nil {
				assert.Equal(t, tt.err.Error(), result.Err.Error())
			}
		})
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "error with underlying error",
			appError: &AppError{
				Type:       ErrorTypeValidation,
				Code:       "INVALID_INPUT",
				Message:    "test error message",
				StatusCode: http.StatusBadRequest,
				Err:        errors.New("underlying error"),
			},
			expected: "INVALID_INPUT: test error message (underlying error)",
		},
		{
			name: "error without underlying error",
			appError: &AppError{
				Type:       ErrorTypeNotFound,
				Code:       "USER_NOT_FOUND",
				Message:    "user not found",
				StatusCode: http.StatusNotFound,
				Err:        nil,
			},
			expected: "USER_NOT_FOUND: user not found",
		},
		{
			name: "error with empty message",
			appError: &AppError{
				Type:       ErrorTypeInternal,
				Code:       "INTERNAL_ERROR",
				Message:    "",
				StatusCode: http.StatusInternalServerError,
				Err:        nil,
			},
			expected: "INTERNAL_ERROR: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")

	tests := []struct {
		name     string
		appError *AppError
		expected error
	}{
		{
			name: "error with underlying error",
			appError: &AppError{
				Type:       ErrorTypeValidation,
				Code:       "INVALID_INPUT",
				Message:    "test error",
				StatusCode: http.StatusBadRequest,
				Err:        underlyingErr,
			},
			expected: underlyingErr,
		},
		{
			name: "error without underlying error",
			appError: &AppError{
				Type:       ErrorTypeNotFound,
				Code:       "USER_NOT_FOUND",
				Message:    "user not found",
				StatusCode: http.StatusNotFound,
				Err:        nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Unwrap()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAppError(t *testing.T) {
	appError := NewValidationError("TEST", "test error", nil)
	regularError := errors.New("regular error")

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "app error",
			err:      appError,
			expected: true,
		},
		{
			name:     "regular error",
			err:      regularError,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "wrapped app error",
			err:      fmt.Errorf("wrapped: %w", appError),
			expected: false, // IsAppError doesn't unwrap errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAppError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAppError(t *testing.T) {
	appError := NewValidationError("TEST", "test error", nil)
	regularError := errors.New("regular error")

	tests := []struct {
		name     string
		err      error
		expected *AppError
	}{
		{
			name:     "app error",
			err:      appError,
			expected: appError,
		},
		{
			name:     "regular error",
			err:      regularError,
			expected: nil,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAppError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorTypeConstants(t *testing.T) {
	// Test that error type constants are properly defined
	assert.Equal(t, ErrorType("validation"), ErrorTypeValidation)
	assert.Equal(t, ErrorType("authentication"), ErrorTypeAuthentication)
	assert.Equal(t, ErrorType("authorization"), ErrorTypeAuthorization)
	assert.Equal(t, ErrorType("not_found"), ErrorTypeNotFound)
	assert.Equal(t, ErrorType("conflict"), ErrorTypeConflict)
	assert.Equal(t, ErrorType("internal"), ErrorTypeInternal)
	assert.Equal(t, ErrorType("database"), ErrorTypeDatabase)
	assert.Equal(t, ErrorType("external"), ErrorTypeExternal)
}

func TestErrorCodeConstants(t *testing.T) {
	// Test that error code constants are properly defined
	assert.Equal(t, "INVALID_INPUT", ErrCodeInvalidInput)
	assert.Equal(t, "USER_NOT_FOUND", ErrCodeUserNotFound)
	assert.Equal(t, "USER_ALREADY_EXISTS", ErrCodeUserAlreadyExists)
	assert.Equal(t, "INVALID_CREDENTIALS", ErrCodeInvalidCredentials)
	assert.Equal(t, "TOKEN_EXPIRED", ErrCodeTokenExpired)
	assert.Equal(t, "TOKEN_INVALID", ErrCodeTokenInvalid)
	assert.Equal(t, "TOKEN_REVOKED", ErrCodeTokenRevoked)
	assert.Equal(t, "INSUFFICIENT_PRIVILEGES", ErrCodeInsufficientPrivileges)
	assert.Equal(t, "DATABASE_CONNECTION", ErrCodeDatabaseConnection)
	assert.Equal(t, "DATABASE_QUERY", ErrCodeDatabaseQuery)
	assert.Equal(t, "EXTERNAL_SERVICE", ErrCodeExternalService)
}

// Benchmark tests for performance
func BenchmarkNewValidationError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewValidationError("TEST_CODE", "test error message", errors.New("underlying error"))
	}
}

func BenchmarkNewNotFoundError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewNotFoundError("USER_NOT_FOUND", "user not found", errors.New("user does not exist"))
	}
}

func BenchmarkNewInternalError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewInternalError("INTERNAL_ERROR", "internal server error", errors.New("database connection failed"))
	}
}

func BenchmarkAppError_Error(b *testing.B) {
	appError := NewValidationError("TEST", "test error message", errors.New("underlying error"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = appError.Error()
	}
}

func BenchmarkAppError_Unwrap(b *testing.B) {
	appError := NewValidationError("TEST", "test error", errors.New("underlying error"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = appError.Unwrap()
	}
}

func BenchmarkIsAppError(b *testing.B) {
	appError := NewValidationError("TEST", "test error", nil)
	regularError := errors.New("regular error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsAppError(appError)
		IsAppError(regularError)
	}
}

func BenchmarkGetAppError(b *testing.B) {
	appError := NewValidationError("TEST", "test error", nil)
	regularError := errors.New("regular error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAppError(appError)
		GetAppError(regularError)
	}
}

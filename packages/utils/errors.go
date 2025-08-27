package utils

import (
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeAuthentication represents authentication errors
	ErrorTypeAuthentication ErrorType = "authentication"
	// ErrorTypeAuthorization represents authorization errors
	ErrorTypeAuthorization ErrorType = "authorization"
	// ErrorTypeNotFound represents not found errors
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeConflict represents conflict errors
	ErrorTypeConflict ErrorType = "conflict"
	// ErrorTypeInternal represents internal server errors
	ErrorTypeInternal ErrorType = "internal"
	// ErrorTypeDatabase represents database errors
	ErrorTypeDatabase ErrorType = "database"
	// ErrorTypeExternal represents external service errors
	ErrorTypeExternal ErrorType = "external"
)

// AppError represents a structured application error
type AppError struct {
	Type       ErrorType `json:"type"`
	Code       string    `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"status_code"`
	Err        error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(code, message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Code:       code,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(code, message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeAuthentication,
		Code:       code,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Err:        err,
	}
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(code, message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeAuthorization,
		Code:       code,
		Message:    message,
		StatusCode: http.StatusForbidden,
		Err:        err,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(code, message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Code:       code,
		Message:    message,
		StatusCode: http.StatusNotFound,
		Err:        err,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(code, message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeConflict,
		Code:       code,
		Message:    message,
		StatusCode: http.StatusConflict,
		Err:        err,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(code, message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Code:       code,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewDatabaseError creates a new database error
func NewDatabaseError(code, message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeDatabase,
		Code:       code,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewExternalError creates a new external service error
func NewExternalError(code, message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeExternal,
		Code:       code,
		Message:    message,
		StatusCode: http.StatusBadGateway,
		Err:        err,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts AppError from an error
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// Common error codes
const (
	ErrCodeInvalidInput           = "INVALID_INPUT"
	ErrCodeUserNotFound           = "USER_NOT_FOUND"
	ErrCodeUserAlreadyExists      = "USER_ALREADY_EXISTS"
	ErrCodeInvalidCredentials     = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired           = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid           = "TOKEN_INVALID"
	ErrCodeTokenRevoked           = "TOKEN_REVOKED"
	ErrCodeInsufficientPrivileges = "INSUFFICIENT_PRIVILEGES"
	ErrCodeDatabaseConnection     = "DATABASE_CONNECTION"
	ErrCodeDatabaseQuery          = "DATABASE_QUERY"
	ErrCodeExternalService        = "EXTERNAL_SERVICE"
)

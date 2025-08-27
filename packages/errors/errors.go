package customerrors

import (
	"fmt"
	"net/http"
)

// Error represents a domain error with additional context
type Error struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	HTTPStatus int               `json:"-"`
	Err        error             `json:"-"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Err
}

// WithDetails adds additional details to the error
func (e *Error) WithDetails(details string) *Error {
	e.Details = details
	return e
}

// WithMetadata adds metadata to the error
func (e *Error) WithMetadata(key, value string) *Error {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// WithHTTPStatus sets the HTTP status code for the error
func (e *Error) WithHTTPStatus(status int) *Error {
	e.HTTPStatus = status
	return e
}

// Common error constructors

// New creates a new error with the given code and message
func New(code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// NewBadRequest creates a new bad request error
func NewBadRequest(code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewUnauthorized creates a new unauthorized error
func NewUnauthorized(code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewForbidden creates a new forbidden error
func NewForbidden(code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

// NewNotFound creates a new not found error
func NewNotFound(code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusNotFound,
	}
}

// NewConflict creates a new conflict error
func NewConflict(code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// NewValidation creates a new validation error
func NewValidation(code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusUnprocessableEntity,
	}
}

// NewTooManyRequests creates a new rate limit error
func NewTooManyRequests(code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusTooManyRequests,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

// WrapBadRequest wraps an error as a bad request
func WrapBadRequest(err error, code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Err:        err,
	}
}

// WrapNotFound wraps an error as not found
func WrapNotFound(err error, code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusNotFound,
		Err:        err,
	}
}

// Common error codes
const (
	// General errors
	ErrInternal      = "INTERNAL_ERROR"
	ErrInvalidInput  = "INVALID_INPUT"
	ErrNotFound      = "NOT_FOUND"
	ErrAlreadyExists = "ALREADY_EXISTS"
	ErrUnauthorized  = "UNAUTHORIZED"
	ErrForbidden     = "FORBIDDEN"
	ErrValidation    = "VALIDATION_ERROR"
	ErrRateLimit     = "RATE_LIMIT_EXCEEDED"
	ErrTimeout       = "TIMEOUT"
	ErrUnavailable   = "SERVICE_UNAVAILABLE"

	// Authentication errors
	ErrInvalidCredentials = "INVALID_CREDENTIALS"
	ErrTokenExpired       = "TOKEN_EXPIRED"
	ErrTokenInvalid       = "TOKEN_INVALID"
	ErrTokenMissing       = "TOKEN_MISSING"
	ErrAccountLocked      = "ACCOUNT_LOCKED"
	ErrAccountDisabled    = "ACCOUNT_DISABLED"

	// Database errors
	ErrDatabaseConnection = "DATABASE_CONNECTION_ERROR"
	ErrDatabaseQuery      = "DATABASE_QUERY_ERROR"
	ErrDatabaseConstraint = "DATABASE_CONSTRAINT_VIOLATION"
	ErrDatabaseTimeout    = "DATABASE_TIMEOUT"

	// External service errors
	ErrExternalService     = "EXTERNAL_SERVICE_ERROR"
	ErrExternalTimeout     = "EXTERNAL_SERVICE_TIMEOUT"
	ErrExternalUnavailable = "EXTERNAL_SERVICE_UNAVAILABLE"
)

// Common error messages
var (
	ErrMsgInternal      = "An internal error occurred"
	ErrMsgInvalidInput  = "Invalid input provided"
	ErrMsgNotFound      = "Resource not found"
	ErrMsgAlreadyExists = "Resource already exists"
	ErrMsgUnauthorized  = "Unauthorized access"
	ErrMsgForbidden     = "Access forbidden"
	ErrMsgValidation    = "Validation failed"
	ErrMsgRateLimit     = "Rate limit exceeded"
	ErrMsgTimeout       = "Request timeout"
	ErrMsgUnavailable   = "Service unavailable"
)

// IsError checks if an error is a domain error
func IsError(err error) bool {
	_, ok := err.(*Error)
	return ok
}

// GetError extracts the domain error from an error
func GetError(err error) *Error {
	if domainErr, ok := err.(*Error); ok {
		return domainErr
	}
	return nil
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	if domainErr := GetError(err); domainErr != nil {
		return domainErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetErrorCode returns the error code for an error
func GetErrorCode(err error) string {
	if domainErr := GetError(err); domainErr != nil {
		return domainErr.Code
	}
	return ErrInternal
}

// GetErrorMessage returns the error message for an error
func GetErrorMessage(err error) string {
	if domainErr := GetError(err); domainErr != nil {
		return domainErr.Message
	}
	return err.Error()
}

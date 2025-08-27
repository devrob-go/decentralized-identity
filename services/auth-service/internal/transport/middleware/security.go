package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"auth-service/config"
	"auth-service/internal/services"

	zlog "packages/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// SecurityMiddleware provides comprehensive security features
type SecurityMiddleware struct {
	logger  *zlog.Logger
	config  *config.Config
	service *services.Service
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(logger *zlog.Logger, cfg *config.Config, service *services.Service) *SecurityMiddleware {
	return &SecurityMiddleware{
		logger:  logger,
		config:  cfg,
		service: service,
	}
}

// UnarySecurityInterceptor provides security for unary RPC calls
func (s *SecurityMiddleware) UnarySecurityInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Add correlation ID for tracking
		correlationID := generateCorrelationID()
		ctx = zlog.WithCorrelationID(ctx, correlationID)

		// Input validation
		if err := s.validateInput(req, info.FullMethod); err != nil {
			s.logger.Warn(ctx, "Input validation failed", map[string]any{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return nil, status.Error(codes.InvalidArgument, "input validation failed")
		}

		// Authentication check for protected methods
		if s.isProtectedMethod(info.FullMethod) {
			if err := s.authenticateRequest(ctx); err != nil {
				s.logger.Warn(ctx, "Authentication failed", map[string]any{
					"method": info.FullMethod,
					"error":  err.Error(),
				})
				return nil, status.Error(codes.Unauthenticated, "authentication required")
			}
		}

		// Authorization check
		if err := s.authorizeRequest(ctx, info.FullMethod); err != nil {
			s.logger.Warn(ctx, "Authorization failed", map[string]any{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
		}

		// Audit logging for sensitive operations
		if s.isSensitiveMethod(info.FullMethod) {
			s.logAuditEvent(ctx, "method_call", info.FullMethod, req)
		}

		return handler(ctx, req)
	}
}

// StreamSecurityInterceptor provides security for streaming RPC calls
func (s *SecurityMiddleware) StreamSecurityInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// Add correlation ID for tracking
		correlationID := generateCorrelationID()
		ctx = zlog.WithCorrelationID(ctx, correlationID)

		// Authentication check for protected methods
		if s.isProtectedMethod(info.FullMethod) {
			if err := s.authenticateRequest(ctx); err != nil {
				s.logger.Warn(ctx, "Authentication failed for stream", map[string]any{
					"method": info.FullMethod,
					"error":  err.Error(),
				})
				return status.Error(codes.Unauthenticated, "authentication required")
			}
		}

		// Create wrapped stream with security context
		wrappedStream := &secureServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrappedStream)
	}
}

// secureServerStream wraps ServerStream with security context
type secureServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *secureServerStream) Context() context.Context {
	return s.ctx
}

// isProtectedMethod checks if a method requires authentication
func (s *SecurityMiddleware) isProtectedMethod(method string) bool {
	protectedMethods := []string{
		"/auth.AuthService/SignOut",
		"/auth.AuthService/RefreshToken",
		"/auth.AuthService/RevokeToken",
		"/auth.AuthService/ListUsers",
		// Add other protected methods here
	}

	for _, protected := range protectedMethods {
		if strings.Contains(method, protected) {
			return true
		}
	}
	return false
}

// isSensitiveMethod checks if a method should be audited
func (s *SecurityMiddleware) isSensitiveMethod(method string) bool {
	sensitiveMethods := []string{
		"/auth.AuthService/SignIn",
		"/auth.AuthService/SignUp",
		"/auth.AuthService/SignOut",
		"/auth.AuthService/Revoke",
	}

	for _, sensitive := range sensitiveMethods {
		if strings.Contains(method, sensitive) {
			return true
		}
	}
	return false
}

// authenticateRequest validates the authentication token
func (s *SecurityMiddleware) authenticateRequest(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("no metadata found")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return fmt.Errorf("no authorization token provided")
	}

	token := tokens[0]
	if !strings.HasPrefix(token, "Bearer ") {
		return fmt.Errorf("invalid token format")
	}

	token = strings.TrimPrefix(token, "Bearer ")

	// Validate JWT token
	if err := s.validateJWTToken(token); err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	return nil
}

// validateJWTToken validates a JWT token using the auth service
func (s *SecurityMiddleware) validateJWTToken(token string) error {
	// Basic format check first
	if len(token) < 10 {
		return fmt.Errorf("token too short")
	}

	// Check if token contains required parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format")
	}

	// Use the auth service to validate the token
	ctx := context.Background()
	_, err := s.service.Auth.ValidateToken(ctx, token, s.config.JWTAccessTokenSecret)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	return nil
}

// authorizeRequest checks if the user has permission to access the method
func (s *SecurityMiddleware) authorizeRequest(ctx context.Context, method string) error {
	// TODO: Implement role-based access control
	// For now, just allow authenticated requests
	return nil
}

// validateInput validates input data for security
func (s *SecurityMiddleware) validateInput(req any, method string) error {
	// Basic input validation
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Method-specific validation
	switch method {
	case "/auth.AuthService/SignUp":
		return s.validateSignUpRequest(req)
	case "/auth.AuthService/SignIn":
		return s.validateSignInRequest(req)
	}

	return nil
}

// validateSignUpRequest validates signup request data
func (s *SecurityMiddleware) validateSignUpRequest(req any) error {
	// TODO: Implement proper request validation using protobuf reflection
	// For now, just basic checks
	return nil
}

// validateSignInRequest validates signin request data
func (s *SecurityMiddleware) validateSignInRequest(req any) error {
	// TODO: Implement proper request validation using protobuf reflection
	// For now, just basic checks
	return nil
}

// logAuditEvent logs security-relevant events
func (s *SecurityMiddleware) logAuditEvent(ctx context.Context, eventType, method string, data any) {
	if s.config.LogSensitiveData {
		s.logger.Info(ctx, "Security audit event", map[string]any{
			"event_type": eventType,
			"method":     method,
			"data":       data,
			"timestamp":  time.Now().UTC(),
		})
	} else {
		s.logger.Info(ctx, "Security audit event", map[string]any{
			"event_type": eventType,
			"method":     method,
			"timestamp":  time.Now().UTC(),
		})
	}
}

// generateCorrelationID generates a unique correlation ID
func generateCorrelationID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateSecurityHeadersMiddleware creates HTTP security headers middleware
func CreateSecurityHeadersMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.SecurityHeadersEnabled {
				// Security Headers
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-Frame-Options", "DENY")
				w.Header().Set("X-XSS-Protection", "1; mode=block")
				w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
				w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

				// HSTS Header
				if cfg.HSTSMaxAge > 0 {
					w.Header().Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d; includeSubDomains; preload", cfg.HSTSMaxAge))
				}

				// Content Security Policy
				if cfg.ContentSecurityPolicy != "" {
					w.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SanitizeInput removes potentially dangerous characters from input
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters
	re := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	input = re.ReplaceAllString(input, "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// ValidatePasswordStrength validates password against security policy
func ValidatePasswordStrength(password string, cfg *config.Config) error {
	if len(password) < cfg.MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", cfg.MinPasswordLength)
	}

	if cfg.RequireUppercase && !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if cfg.RequireLowercase && !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if cfg.RequireNumbers && !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one number")
	}

	if cfg.RequireSpecialChars && !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

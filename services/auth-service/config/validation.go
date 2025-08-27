package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("config validation error for %s: %s", e.Field, e.Message)
}

// ValidationResult holds the result of configuration validation
type ValidationResult struct {
	IsValid bool
	Errors  []ValidationError
}

// AddError adds a validation error to the result
func (r *ValidationResult) AddError(field, message string) {
	r.Errors = append(r.Errors, ValidationError{Field: field, Message: message})
	r.IsValid = false
}

// ValidateConfig performs comprehensive configuration validation
func ValidateConfig(cfg *Config) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	// Validate environment
	if err := validateEnvironment(cfg.Environment); err != nil {
		result.AddError("APP_ENV", err.Error())
	}

	// Validate port numbers
	if err := validatePort(cfg.AuthServicePort, "APP_PORT"); err != nil {
		result.AddError("APP_PORT", err.Error())
	}

	if err := validatePort(cfg.RestGatewayPort, "REST_PORT"); err != nil {
		result.AddError("REST_PORT", err.Error())
	}

	if err := validatePort(cfg.PostgresPort, "POSTGRES_PORT"); err != nil {
		result.AddError("POSTGRES_PORT", err.Error())
	}

	// Validate database configuration
	if err := validateDatabaseConfig(cfg); err != nil {
		result.AddError("database", err.Error())
	}

	// Validate JWT configuration
	if err := validateJWTConfig(cfg); err != nil {
		result.AddError("JWT", err.Error())
	}

	// Validate logging configuration
	if err := validateLoggingConfig(cfg); err != nil {
		result.AddError("logging", err.Error())
	}

	// Validate CORS configuration
	if err := validateCORSConfig(cfg); err != nil {
		result.AddError("CORS", err.Error())
	}

	// Validate security configuration
	if err := validateSecurityConfig(cfg); err != nil {
		result.AddError("security", err.Error())
	}

	// Validate TLS configuration
	if err := validateTLSConfig(cfg); err != nil {
		result.AddError("TLS", err.Error())
	}

	// Validate rate limiting configuration
	if err := validateRateLimitConfig(cfg); err != nil {
		result.AddError("rate_limiting", err.Error())
	}

	// Validate password policy
	if err := validatePasswordPolicy(cfg); err != nil {
		result.AddError("password_policy", err.Error())
	}

	// Validate JWT timing configuration
	if err := validateJWTTimingConfig(cfg); err != nil {
		result.AddError("JWT_timing", err.Error())
	}

	return result
}

// validateEnvironment validates the environment setting
func validateEnvironment(env string) error {
	validEnvs := []string{PRODUCTION_ENV, STAGING_ENV, DEVELOPMENT_ENV}
	for _, validEnv := range validEnvs {
		if env == validEnv {
			return nil
		}
	}
	return fmt.Errorf("must be one of: %s", strings.Join(validEnvs, ", "))
}

// validatePort validates a port number
func validatePort(port, fieldName string) error {
	if port == "" {
		return fmt.Errorf("cannot be empty")
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("must be a valid number")
	}

	if portNum <= 0 || portNum > 65535 {
		return fmt.Errorf("must be between 1 and 65535")
	}

	return nil
}

// validateDatabaseConfig validates database configuration
func validateDatabaseConfig(cfg *Config) error {
	if cfg.PostgresHost == "" {
		return fmt.Errorf("POSTGRES_HOST cannot be empty")
	}

	if cfg.PostgresUser == "" {
		return fmt.Errorf("POSTGRES_USER cannot be empty")
	}

	if cfg.PostgresPassword == "" {
		return fmt.Errorf("POSTGRES_PASSWORD cannot be empty")
	}

	if cfg.PostgresDB == "" {
		return fmt.Errorf("POSTGRES_DB cannot be empty")
	}

	// Validate host format
	if cfg.PostgresHost != "localhost" && cfg.PostgresHost != "127.0.0.1" {
		if net.ParseIP(cfg.PostgresHost) == nil {
			// Check if it's a valid hostname
			if !isValidHostname(cfg.PostgresHost) {
				return fmt.Errorf("POSTGRES_HOST must be a valid IP address or hostname")
			}
		}
	}

	return nil
}

// validateJWTConfig validates JWT configuration
func validateJWTConfig(cfg *Config) error {
	if cfg.JWTAccessTokenSecret == "" {
		return fmt.Errorf("JWT_ACCESS_TOKEN_SECRET cannot be empty")
	}

	if cfg.JWTRefreshTokenSecret == "" {
		return fmt.Errorf("JWT_REFRESH_TOKEN_SECRET cannot be empty")
	}

	if len(cfg.JWTAccessTokenSecret) < 32 {
		return fmt.Errorf("JWT_ACCESS_TOKEN_SECRET must be at least 32 characters long")
	}

	if len(cfg.JWTRefreshTokenSecret) < 32 {
		return fmt.Errorf("JWT_REFRESH_TOKEN_SECRET must be at least 32 characters long")
	}

	// Check for weak secrets in development
	if cfg.Environment == DEVELOPMENT_ENV {
		if cfg.JWTAccessTokenSecret == "default-access" {
			return fmt.Errorf("JWT_ACCESS_TOKEN_SECRET should not use default value in any environment")
		}
		if cfg.JWTRefreshTokenSecret == "default-refresh" {
			return fmt.Errorf("JWT_REFRESH_TOKEN_SECRET should not use default value in any environment")
		}
	}

	return nil
}

// validateLoggingConfig validates logging configuration
func validateLoggingConfig(cfg *Config) error {
	validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	levelValid := false
	for _, validLevel := range validLevels {
		if strings.ToLower(cfg.LogLevel) == validLevel {
			levelValid = true
			break
		}
	}

	if !levelValid {
		return fmt.Errorf("LOG_LEVEL must be one of: %s", strings.Join(validLevels, ", "))
	}

	return nil
}

// validateCORSConfig validates CORS configuration
func validateCORSConfig(cfg *Config) error {
	if len(cfg.AllowedOrigins) == 0 {
		return fmt.Errorf("ALLOWED_ORIGINS cannot be empty")
	}

	for _, origin := range cfg.AllowedOrigins {
		if origin == "" {
			continue // Skip empty origins
		}
		if !isValidOrigin(origin) {
			return fmt.Errorf("invalid origin in ALLOWED_ORIGINS: %s", origin)
		}
	}

	return nil
}

// isValidHostname checks if a string is a valid hostname
func isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// Check if it's a valid domain name
	parts := strings.Split(hostname, ".")
	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
		// Check if part contains only valid characters
		for _, char := range part {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-') {
				return false
			}
		}
		// Check if part doesn't start or end with hyphen
		if part[0] == '-' || part[len(part)-1] == '-' {
			return false
		}
	}

	return true
}

// isValidOrigin checks if a string is a valid origin
func isValidOrigin(origin string) bool {
	if origin == "*" {
		return true
	}

	// Check if it's a valid URL
	if strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://") {
		// Basic URL validation
		if len(origin) > 8 { // http:// or https:// + at least one character
			return true
		}
	}

	return false
}

// validateSecurityConfig validates security-related configuration
func validateSecurityConfig(cfg *Config) error {
	// Validate security headers configuration
	if cfg.SecurityHeadersEnabled {
		if cfg.HSTSMaxAge <= 0 {
			return fmt.Errorf("HSTS_MAX_AGE must be positive when security headers are enabled")
		}
		
		if cfg.ContentSecurityPolicy == "" {
			return fmt.Errorf("CONTENT_SECURITY_POLICY cannot be empty when security headers are enabled")
		}
	}

	// Validate database SSL mode for production
	if cfg.Environment == PRODUCTION_ENV {
		if cfg.DBSSLMode != "require" && cfg.DBSSLMode != "verify-full" {
			return fmt.Errorf("DB_SSL_MODE must be 'require' or 'verify-full' in production environment")
		}
	}

	// Validate database connection limits
	if cfg.DBMaxConnections <= 0 {
		return fmt.Errorf("DB_MAX_CONNECTIONS must be positive")
	}
	
	if cfg.DBMaxIdleConnections <= 0 {
		return fmt.Errorf("DB_MAX_IDLE_CONNECTIONS must be positive")
	}
	
	if cfg.DBMaxIdleConnections > cfg.DBMaxConnections {
		return fmt.Errorf("DB_MAX_IDLE_CONNECTIONS cannot exceed DB_MAX_CONNECTIONS")
	}

	return nil
}

// validateTLSConfig validates TLS configuration
func validateTLSConfig(cfg *Config) error {
	if cfg.TLSEnabled {
		if cfg.TLSCertFile == "" {
			return fmt.Errorf("TLS_CERT_FILE cannot be empty when TLS is enabled")
		}
		
		if cfg.TLSKeyFile == "" {
			return fmt.Errorf("TLS_KEY_FILE cannot be empty when TLS is enabled")
		}
		
		// Validate TLS version range
		if cfg.MinTLSVersion >= cfg.MaxTLSVersion {
			return fmt.Errorf("MIN_TLS_VERSION must be less than MAX_TLS_VERSION")
		}
		
		// Enforce minimum TLS version for production
		if cfg.Environment == PRODUCTION_ENV && cfg.MinTLSVersion < 0x0303 { // TLS 1.2
			return fmt.Errorf("MIN_TLS_VERSION must be at least 1.2 in production environment")
		}
	}

	return nil
}

// validateRateLimitConfig validates rate limiting configuration
func validateRateLimitConfig(cfg *Config) error {
	if cfg.RateLimitEnabled {
		if cfg.RateLimitRequests <= 0 {
			return fmt.Errorf("RATE_LIMIT_REQUESTS must be positive when rate limiting is enabled")
		}
		
		if cfg.RateLimitWindow <= 0 {
			return fmt.Errorf("RATE_LIMIT_WINDOW must be positive when rate limiting is enabled")
		}
		
		// Validate reasonable limits
		if cfg.RateLimitRequests > 10000 {
			return fmt.Errorf("RATE_LIMIT_REQUESTS cannot exceed 10000")
		}
		
		if cfg.RateLimitWindow > 3600 { // 1 hour
			return fmt.Errorf("RATE_LIMIT_WINDOW cannot exceed 3600 seconds")
		}
	}

	return nil
}

// validatePasswordPolicy validates password policy configuration
func validatePasswordPolicy(cfg *Config) error {
	if cfg.MinPasswordLength < 8 {
		return fmt.Errorf("MIN_PASSWORD_LENGTH must be at least 8")
	}
	
	if cfg.MinPasswordLength > 128 {
		return fmt.Errorf("MIN_PASSWORD_LENGTH cannot exceed 128")
	}
	
	// Validate password complexity requirements
	if cfg.RequireUppercase && cfg.RequireLowercase && cfg.RequireNumbers && cfg.RequireSpecialChars {
		// Calculate minimum length based on complexity requirements
		minRequiredLength := 4 // One character for each requirement
		if cfg.MinPasswordLength < minRequiredLength {
			return fmt.Errorf("MIN_PASSWORD_LENGTH must be at least %d when all complexity requirements are enabled", minRequiredLength)
		}
	}

	return nil
}

// validateJWTTimingConfig validates JWT timing configuration
func validateJWTTimingConfig(cfg *Config) error {
	if cfg.JWTExpirationTime <= 0 {
		return fmt.Errorf("JWT_EXPIRATION_TIME must be positive")
	}
	
	if cfg.JWTExpirationTime > 1440 { // 24 hours
		return fmt.Errorf("JWT_EXPIRATION_TIME cannot exceed 1440 minutes (24 hours)")
	}
	
	if cfg.JWTRefreshExpiration <= 0 {
		return fmt.Errorf("JWT_REFRESH_EXPIRATION must be positive")
	}
	
	if cfg.JWTRefreshExpiration > 365 { // 1 year
		return fmt.Errorf("JWT_REFRESH_EXPIRATION cannot exceed 365 days")
	}
	
	// Validate refresh token is longer than access token
	refreshMinutes := cfg.JWTRefreshExpiration * 24 * 60
	if refreshMinutes <= cfg.JWTExpirationTime {
		return fmt.Errorf("JWT_REFRESH_EXPIRATION must be longer than JWT_EXPIRATION_TIME")
	}

	return nil
}

// GetValidationErrors returns a formatted string of all validation errors
func (r *ValidationResult) GetValidationErrors() string {
	if r.IsValid {
		return "Configuration is valid"
	}

	var errorMessages []string
	for _, err := range r.Errors {
		errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}

	return fmt.Sprintf("Configuration validation failed:\n%s", strings.Join(errorMessages, "\n"))
}

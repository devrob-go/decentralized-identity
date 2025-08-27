package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	PRODUCTION_ENV  = "production"
	STAGING_ENV     = "staging"
	DEVELOPMENT_ENV = "development"
)

// Config holds application configuration
type Config struct {
	Environment           string
	AuthServicePort       string
	RestGatewayPort       string
	PostgresUser          string
	PostgresPassword      string
	PostgresDB            string
	PostgresHost          string
	PostgresPort          string
	JWTAccessTokenSecret  string
	JWTRefreshTokenSecret string
	AllowedOrigins        []string
	LogLevel              string
	LogJSONFormat         bool
	HealthCheckTimeout    int // in seconds
	ServerReadTimeout     int // in seconds
	ServerWriteTimeout    int // in seconds

	// Security Configuration
	TLSEnabled    bool
	TLSCertFile   string
	TLSKeyFile    string
	MinTLSVersion uint16
	MaxTLSVersion uint16

	// Rate Limiting
	RateLimitEnabled  bool
	RateLimitRequests int
	RateLimitWindow   int // in seconds

	// Security Headers
	SecurityHeadersEnabled bool
	HSTSMaxAge             int // in seconds
	ContentSecurityPolicy  string

	// Password Policy
	MinPasswordLength   int
	RequireUppercase    bool
	RequireLowercase    bool
	RequireNumbers      bool
	RequireSpecialChars bool

	// JWT Configuration
	JWTExpirationTime    int // in minutes
	JWTRefreshExpiration int // in days

	// Database Security
	DBSSLMode            string
	DBMaxConnections     int
	DBMaxIdleConnections int
	DBConnectionTimeout  int // in seconds

	// Logging Security
	LogSensitiveData  bool
	LogRequestHeaders bool
	LogResponseBody   bool
}

// LoadConfig loads and validates configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file only if it exists, without overwriting existing env vars
	_ = godotenv.Load() // Ignore error if .env doesn't exist
	raw := getEnv("ALLOWED_ORIGINS", "")

	// Parse TLS version strings
	minTLSVersion := parseTLSVersion(getEnv("MIN_TLS_VERSION", "1.2"))
	maxTLSVersion := parseTLSVersion(getEnv("MAX_TLS_VERSION", "1.3"))

	cfg := &Config{
		Environment:           getEnv("APP_ENV", "development"),
		AuthServicePort:       getEnv("APP_PORT", "8081"),
		RestGatewayPort:       getEnv("REST_PORT", "8080"),
		PostgresUser:          getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:      getEnv("POSTGRES_PASSWORD", "password"),
		PostgresDB:            getEnv("POSTGRES_DB", "starter_db"),
		PostgresHost:          getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:          getEnv("POSTGRES_PORT", "5432"),
		JWTAccessTokenSecret:  getEnv("JWT_ACCESS_TOKEN_SECRET", ""),
		JWTRefreshTokenSecret: getEnv("JWT_REFRESH_TOKEN_SECRET", ""),
		AllowedOrigins:        strings.Split(raw, ","),
		LogLevel:              getEnv("LOG_LEVEL", "debug"),
		LogJSONFormat:         getEnv("LOG_JSON_FORMAT", "false") == "true",
		HealthCheckTimeout:    getEnvInt("HEALTH_CHECK_TIMEOUT", 5),
		ServerReadTimeout:     getEnvInt("SERVER_READ_TIMEOUT", 10),
		ServerWriteTimeout:    getEnvInt("SERVER_WRITE_TIMEOUT", 10),

		// Security Configuration
		TLSEnabled:    getEnv("TLS_ENABLED", "false") == "true",
		TLSCertFile:   getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:    getEnv("TLS_KEY_FILE", ""),
		MinTLSVersion: minTLSVersion,
		MaxTLSVersion: maxTLSVersion,

		// Rate Limiting
		RateLimitEnabled:  getEnv("RATE_LIMIT_ENABLED", "true") == "true",
		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvInt("RATE_LIMIT_WINDOW", 60),

		// Security Headers
		SecurityHeadersEnabled: getEnv("SECURITY_HEADERS_ENABLED", "true") == "true",
		HSTSMaxAge:             getEnvInt("HSTS_MAX_AGE", 31536000), // 1 year
		ContentSecurityPolicy:  getEnv("CONTENT_SECURITY_POLICY", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"),

		// Password Policy
		MinPasswordLength:   getEnvInt("MIN_PASSWORD_LENGTH", 12),
		RequireUppercase:    getEnv("REQUIRE_UPPERCASE", "true") == "true",
		RequireLowercase:    getEnv("REQUIRE_LOWERCASE", "true") == "true",
		RequireNumbers:      getEnv("REQUIRE_NUMBERS", "true") == "true",
		RequireSpecialChars: getEnv("REQUIRE_SPECIAL_CHARS", "true") == "true",

		// JWT Configuration
		JWTExpirationTime:    getEnvInt("JWT_EXPIRATION_TIME", 15),   // 15 minutes
		JWTRefreshExpiration: getEnvInt("JWT_REFRESH_EXPIRATION", 7), // 7 days

		// Database Security
		DBSSLMode:            getEnv("DB_SSL_MODE", "require"),
		DBMaxConnections:     getEnvInt("DB_MAX_CONNECTIONS", 25),
		DBMaxIdleConnections: getEnvInt("DB_MAX_IDLE_CONNECTIONS", 5),
		DBConnectionTimeout:  getEnvInt("DB_CONNECTION_TIMEOUT", 30),

		// Logging Security
		LogSensitiveData:  getEnv("LOG_SENSITIVE_DATA", "false") == "true",
		LogRequestHeaders: getEnv("LOG_REQUEST_HEADERS", "false") == "true",
		LogResponseBody:   getEnv("LOG_RESPONSE_BODY", "false") == "true",
	}

	// Validate configuration
	validationResult := ValidateConfig(cfg)
	if !validationResult.IsValid {
		return nil, fmt.Errorf("invalid configuration: %s", validationResult.GetValidationErrors())
	}

	return cfg, nil
}

// parseTLSVersion converts TLS version string to uint16
func parseTLSVersion(version string) uint16 {
	switch strings.ToLower(version) {
	case "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}

// getEnv retrieves an environment variable or returns a fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return fallback
}

// getEnvInt retrieves an environment variable as an integer or returns a fallback
func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

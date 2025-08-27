package config

import (
	"time"
)

// TransportConfig holds all transport-related configuration
type TransportConfig struct {
	TLS      TLSConfig      `json:"tls"`
	Server   ServerConfig   `json:"server"`
	Gateway  GatewayConfig  `json:"gateway"`
	Health   HealthConfig   `json:"health"`
	Security SecurityConfig `json:"security"`
}

// TLSConfig holds TLS-related configuration
type TLSConfig struct {
	Enabled      bool     `json:"enabled"`
	CertFile     string   `json:"cert_file"`
	KeyFile      string   `json:"key_file"`
	MinVersion   uint16   `json:"min_version"`
	MaxVersion   uint16   `json:"max_version"`
	CipherSuites []uint16 `json:"cipher_suites"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	GRPCPort          string        `json:"grpc_port"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	IdleTimeout       time.Duration `json:"idle_timeout"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout"`
	ShutdownTimeout   time.Duration `json:"shutdown_timeout"`
}

// GatewayConfig holds REST gateway configuration
type GatewayConfig struct {
	RESTPort       string   `json:"rest_port"`
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
	MaxAge         int      `json:"max_age"`
}

// HealthConfig holds health check configuration
type HealthConfig struct {
	Timeout        time.Duration `json:"timeout"`
	CheckInterval  time.Duration `json:"check_interval"`
	ReadinessDelay time.Duration `json:"readiness_delay"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	RateLimitEnabled bool   `json:"rate_limit_enabled"`
	RateLimit        int    `json:"rate_limit"`
	JWTSecret        string `json:"jwt_secret"`
	APIKeyRequired   bool   `json:"api_key_required"`
}

// DefaultTransportConfig returns default transport configuration
func DefaultTransportConfig() *TransportConfig {
	return &TransportConfig{
		TLS: TLSConfig{
			Enabled:    false,
			MinVersion: 0x0301, // TLS 1.0
			MaxVersion: 0x0304, // TLS 1.3
			CipherSuites: []uint16{
				0xC02C, // TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
				0xC030, // TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
				0xC02F, // TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
				0xC02B, // TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
				0xCCA9, // TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305
				0xCCA8, // TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
			},
		},
		Server: ServerConfig{
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
			ShutdownTimeout:   5 * time.Second,
		},
		Gateway: GatewayConfig{
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With"},
			MaxAge:         86400, // 24 hours
		},
		Health: HealthConfig{
			Timeout:        5 * time.Second,
			CheckInterval:  30 * time.Second,
			ReadinessDelay: 100 * time.Millisecond,
		},
		Security: SecurityConfig{
			RateLimitEnabled: true,
			RateLimit:        1000,
		},
	}
}

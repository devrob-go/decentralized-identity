package tls

import (
	"crypto/tls"
	"fmt"

	"auth-service/internal/config"
)

// Manager handles TLS configuration and certificate management
type Manager struct {
	config *config.TLSConfig
}

// NewManager creates a new TLS manager
func NewManager(cfg *config.TLSConfig) *Manager {
	return &Manager{
		config: cfg,
	}
}

// CreateTLSConfig creates a TLS configuration for the server
func (m *Manager) CreateTLSConfig() (*tls.Config, error) {
	if !m.config.Enabled {
		return nil, fmt.Errorf("TLS is not enabled")
	}

	// Load certificate and private key
	cert, err := tls.LoadX509KeyPair(m.config.CertFile, m.config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   m.config.MinVersion,
		MaxVersion:   m.config.MaxVersion,
		CipherSuites: m.config.CipherSuites,

		// Security best practices
		PreferServerCipherSuites: true,

		// Curve preferences for ECDHE
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},

		// Additional security settings
		SessionTicketsDisabled: false,
		ClientAuth:             tls.NoClientCert,
	}

	return tlsConfig, nil
}

// CreateClientTLSConfig creates a TLS configuration for client connections
func (m *Manager) CreateClientTLSConfig() (*tls.Config, error) {
	if !m.config.Enabled {
		return nil, fmt.Errorf("TLS is not enabled")
	}

	return &tls.Config{
		MinVersion:   m.config.MinVersion,
		MaxVersion:   m.config.MaxVersion,
		CipherSuites: m.config.CipherSuites,
	}, nil
}

// IsEnabled returns whether TLS is enabled
func (m *Manager) IsEnabled() bool {
	return m.config.Enabled
}

// GetPort returns the TLS port if enabled
func (m *Manager) GetPort() string {
	if m.config.Enabled {
		return ":443" // Default HTTPS port
	}
	return ":80" // Default HTTP port
}

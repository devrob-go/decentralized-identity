package transport

import (
	"net/http"
	"time"
)

// HTTPTransportConfig holds configuration for HTTP transport
type HTTPTransportConfig struct {
	Port         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// NewHTTPTransport creates a new HTTP transport configuration
func NewHTTPTransport(config *HTTPTransportConfig) *http.Server {
	return &http.Server{
		Addr:         ":" + config.Port,
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(config.IdleTimeout) * time.Second,
	}
}

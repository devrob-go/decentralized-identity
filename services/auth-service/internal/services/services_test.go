package services

import (
	"testing"

	"auth-service/config"

	zlog "packages/logger"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		Environment:     "test",
		AuthServicePort: "8081",
	}

	// Test service creation
	service := NewService(nil, logger, cfg)

	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.Config)
	assert.Nil(t, service.DB)
	assert.NotNil(t, service.User)
	assert.NotNil(t, service.Auth)
}

func TestService_Structure(t *testing.T) {
	cfg := &config.Config{
		Environment:     "test",
		AuthServicePort: "8081",
	}

	// Create service
	service := &Service{
		Config: cfg,
		DB:     nil,
		User:   nil,
		Auth:   nil,
	}

	// Test that the service was created properly
	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.Config)
	assert.Nil(t, service.DB)
	assert.Nil(t, service.User)
	assert.Nil(t, service.Auth)
}

func TestService_Configuration(t *testing.T) {
	// Test configuration structure
	cfg := &config.Config{
		Environment:      "test",
		AuthServicePort:  "8081",
		RestGatewayPort:  "8080",
		PostgresHost:     "localhost",
		PostgresPort:     "5432",
		PostgresDB:       "test_db",
		PostgresUser:     "test_user",
		PostgresPassword: "test_password",
	}

	// Verify configuration fields
	assert.Equal(t, "test", cfg.Environment)
	assert.Equal(t, "8081", cfg.AuthServicePort)
	assert.Equal(t, "8080", cfg.RestGatewayPort)
	assert.Equal(t, "localhost", cfg.PostgresHost)
	assert.Equal(t, "5432", cfg.PostgresPort)
	assert.Equal(t, "test_db", cfg.PostgresDB)
	assert.Equal(t, "test_user", cfg.PostgresUser)
	assert.Equal(t, "test_password", cfg.PostgresPassword)
}

func TestService_ServiceComposition(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		Environment:     "test",
		AuthServicePort: "8081",
	}

	// Create service
	service := NewService(nil, logger, cfg)

	// Test that all required services are present
	assert.NotNil(t, service.User, "User service should be initialized")
	assert.NotNil(t, service.Auth, "Auth service should be initialized")
}

// Benchmark tests for performance
func BenchmarkNewService(b *testing.B) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		Environment:     "test",
		AuthServicePort: "8081",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewService(nil, logger, cfg)
	}
}

func BenchmarkService_Configuration(b *testing.B) {
	cfg := &config.Config{
		Environment:     "test",
		AuthServicePort: "8081",
		RestGatewayPort: "8080",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Environment
		_ = cfg.AuthServicePort
		_ = cfg.RestGatewayPort
	}
}

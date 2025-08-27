package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Environment represents the deployment environment
type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

// BaseConfig holds common configuration for all services
type BaseConfig struct {
	Environment Environment
	ServiceName string
	LogLevel    string
	Port        int
	Host        string
}

// LoadBaseConfig loads base configuration from environment variables
func LoadBaseConfig(serviceName string) (*BaseConfig, error) {
	config := &BaseConfig{
		ServiceName: serviceName,
		Environment: Environment(getEnv("ENVIRONMENT", "development")),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Host:        getEnv("HOST", "0.0.0.0"),
	}

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}
	config.Port = port

	return config, nil
}

// IsDevelopment returns true if running in development mode
func (c *BaseConfig) IsDevelopment() bool {
	return c.Environment == Development
}

// IsStaging returns true if running in staging mode
func (c *BaseConfig) IsStaging() bool {
	return c.Environment == Staging
}

// IsProduction returns true if running in production mode
func (c *BaseConfig) IsProduction() bool {
	return c.Environment == Production
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvRequired gets a required environment variable
func getEnvRequired(key string) (string, error) {
	if value := os.Getenv(key); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("required environment variable %s is not set", key)
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) (int, error) {
	if value := os.Getenv(key); value != "" {
		return strconv.Atoi(value)
	}
	return defaultValue, nil
}

// getEnvBool gets a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) (bool, error) {
	if value := os.Getenv(key); value != "" {
		return strconv.ParseBool(strings.ToLower(value))
	}
	return defaultValue, nil
}

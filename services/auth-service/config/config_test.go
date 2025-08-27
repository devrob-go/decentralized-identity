package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment variables
	originalEnv := make(map[string]string)
	envVars := []string{
		"APP_ENV", "APP_PORT", "REST_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD",
		"POSTGRES_DB", "POSTGRES_HOST", "POSTGRES_PORT", "JWT_ACCESS_TOKEN_SECRET",
		"JWT_REFRESH_TOKEN_SECRET", "ALLOWED_ORIGINS", "LOG_LEVEL", "LOG_JSON_FORMAT",
		"HEALTH_CHECK_TIMEOUT", "SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT",
	}

	for _, envVar := range envVars {
		if val := os.Getenv(envVar); val != "" {
			originalEnv[envVar] = val
		}
	}

	// Clean up after test
	defer func() {
		for envVar := range originalEnv {
			os.Setenv(envVar, originalEnv[envVar])
		}
		for _, envVar := range envVars {
			if _, exists := originalEnv[envVar]; !exists {
				os.Unsetenv(envVar)
			}
		}
	}()

	tests := []struct {
		name           string
		envVars        map[string]string
		expectedConfig *Config
		expectError    bool
	}{
		{
			name: "default configuration",
			envVars: map[string]string{
				"JWT_ACCESS_TOKEN_SECRET":  "this-is-a-very-long-secret-key-for-access-tokens-32",
				"JWT_REFRESH_TOKEN_SECRET": "this-is-a-very-long-secret-key-for-refresh-tokens-32",
			},
			expectedConfig: &Config{
				Environment:           "development",
				AuthServicePort:       "8081",
				RestGatewayPort:       "8080",
				PostgresUser:          "postgres",
				PostgresPassword:      "password",
				PostgresDB:            "starter_db",
				PostgresHost:          "localhost",
				PostgresPort:          "5432",
				JWTAccessTokenSecret:  "this-is-a-very-long-secret-key-for-access-tokens-32",
				JWTRefreshTokenSecret: "this-is-a-very-long-secret-key-for-refresh-tokens-32",
				AllowedOrigins:        []string{""},
				LogLevel:              "debug",
				LogJSONFormat:         false,
				HealthCheckTimeout:    5,
				ServerReadTimeout:     10,
				ServerWriteTimeout:    10,
			},
			expectError: false,
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
				"APP_ENV":                  "production",
				"APP_PORT":                 "9091",
				"REST_PORT":                "9090",
				"POSTGRES_USER":            "custom_user",
				"POSTGRES_PASSWORD":        "custom_password",
				"POSTGRES_DB":              "custom_db",
				"POSTGRES_HOST":            "localhost",
				"POSTGRES_PORT":            "5433",
				"JWT_ACCESS_TOKEN_SECRET":  "this-is-a-very-long-secret-key-for-access-tokens-32",
				"JWT_REFRESH_TOKEN_SECRET": "this-is-a-very-long-secret-key-for-refresh-tokens-32",
				"ALLOWED_ORIGINS":          "https://example.com,https://api.example.com",
				"LOG_LEVEL":                "info",
				"LOG_JSON_FORMAT":          "true",
				"HEALTH_CHECK_TIMEOUT":     "30",
				"SERVER_READ_TIMEOUT":      "60",
				"SERVER_WRITE_TIMEOUT":     "60",
			},
			expectedConfig: &Config{
				Environment:           "production",
				AuthServicePort:       "9091",
				RestGatewayPort:       "9090",
				PostgresUser:          "custom_user",
				PostgresPassword:      "custom_password",
				PostgresDB:            "custom_db",
				PostgresHost:          "localhost",
				PostgresPort:          "5433",
				JWTAccessTokenSecret:  "this-is-a-very-long-secret-key-for-access-tokens-32",
				JWTRefreshTokenSecret: "this-is-a-very-long-secret-key-for-refresh-tokens-32",
				AllowedOrigins:        []string{"https://example.com", "https://api.example.com"},
				LogLevel:              "info",
				LogJSONFormat:         true,
				HealthCheckTimeout:    30,
				ServerReadTimeout:     60,
				ServerWriteTimeout:    60,
			},
			expectError: false,
		},
		{
			name: "partial custom configuration",
			envVars: map[string]string{
				"APP_ENV":                  "staging",
				"APP_PORT":                 "8082",
				"POSTGRES_USER":            "staging_user",
				"LOG_LEVEL":                "warn",
				"LOG_JSON_FORMAT":          "false",
				"HEALTH_CHECK_TIMEOUT":     "15",
				"JWT_ACCESS_TOKEN_SECRET":  "this-is-a-very-long-secret-key-for-access-tokens-32",
				"JWT_REFRESH_TOKEN_SECRET": "this-is-a-very-long-secret-key-for-refresh-tokens-32",
			},
			expectedConfig: &Config{
				Environment:           "staging",
				AuthServicePort:       "8082",
				RestGatewayPort:       "8080", // default
				PostgresUser:          "staging_user",
				PostgresPassword:      "password",   // default
				PostgresDB:            "starter_db", // default
				PostgresHost:          "localhost",  // default
				PostgresPort:          "5432",       // default
				JWTAccessTokenSecret:  "this-is-a-very-long-secret-key-for-access-tokens-32",
				JWTRefreshTokenSecret: "this-is-a-very-long-secret-key-for-refresh-tokens-32",
				AllowedOrigins:        []string{""}, // default
				LogLevel:              "warn",
				LogJSONFormat:         false,
				HealthCheckTimeout:    15,
				ServerReadTimeout:     10, // default
				ServerWriteTimeout:    10, // default
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			for _, envVar := range envVars {
				os.Unsetenv(envVar)
			}

			// Set environment variables for this test
			for envVar, value := range tt.envVars {
				os.Setenv(envVar, value)
			}

			// Load configuration
			config, err := LoadConfig()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, config)

			// Verify configuration values
			assert.Equal(t, tt.expectedConfig.Environment, config.Environment)
			assert.Equal(t, tt.expectedConfig.AuthServicePort, config.AuthServicePort)
			assert.Equal(t, tt.expectedConfig.RestGatewayPort, config.RestGatewayPort)
			assert.Equal(t, tt.expectedConfig.PostgresUser, config.PostgresUser)
			assert.Equal(t, tt.expectedConfig.PostgresPassword, config.PostgresPassword)
			assert.Equal(t, tt.expectedConfig.PostgresDB, config.PostgresDB)
			assert.Equal(t, tt.expectedConfig.PostgresHost, config.PostgresHost)
			assert.Equal(t, tt.expectedConfig.PostgresPort, config.PostgresPort)
			assert.Equal(t, tt.expectedConfig.JWTAccessTokenSecret, config.JWTAccessTokenSecret)
			assert.Equal(t, tt.expectedConfig.JWTRefreshTokenSecret, config.JWTRefreshTokenSecret)
			assert.Equal(t, tt.expectedConfig.LogLevel, config.LogLevel)
			assert.Equal(t, tt.expectedConfig.LogJSONFormat, config.LogJSONFormat)
			assert.Equal(t, tt.expectedConfig.HealthCheckTimeout, config.HealthCheckTimeout)
			assert.Equal(t, tt.expectedConfig.ServerReadTimeout, config.ServerReadTimeout)
			assert.Equal(t, tt.expectedConfig.ServerWriteTimeout, config.ServerWriteTimeout)

			// Verify allowed origins
			if len(tt.expectedConfig.AllowedOrigins) == 1 && tt.expectedConfig.AllowedOrigins[0] == "" {
				assert.Equal(t, []string{""}, config.AllowedOrigins)
			} else {
				assert.ElementsMatch(t, tt.expectedConfig.AllowedOrigins, config.AllowedOrigins)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	// Save original environment variable
	originalValue := os.Getenv("TEST_ENV_VAR")
	defer os.Setenv("TEST_ENV_VAR", originalValue)

	tests := []struct {
		name        string
		key         string
		fallback    string
		envValue    string
		expected    string
		description string
	}{
		{
			name:        "environment variable exists",
			key:         "TEST_ENV_VAR",
			fallback:    "fallback_value",
			envValue:    "actual_value",
			expected:    "actual_value",
			description: "should return the environment variable value",
		},
		{
			name:        "environment variable does not exist",
			key:         "NONEXISTENT_VAR",
			fallback:    "fallback_value",
			envValue:    "",
			expected:    "fallback_value",
			description: "should return the fallback value",
		},
		{
			name:        "environment variable is empty",
			key:         "TEST_ENV_VAR",
			fallback:    "fallback_value",
			envValue:    "",
			expected:    "fallback_value",
			description: "should return the fallback value for empty env var",
		},
		{
			name:        "environment variable has whitespace",
			key:         "TEST_ENV_VAR",
			fallback:    "fallback_value",
			envValue:    "  value_with_whitespace  ",
			expected:    "value_with_whitespace",
			description: "should trim whitespace from environment variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.fallback)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	// Save original environment variable
	originalValue := os.Getenv("TEST_INT_VAR")
	defer os.Setenv("TEST_INT_VAR", originalValue)

	tests := []struct {
		name        string
		key         string
		fallback    int
		envValue    string
		expected    int
		description string
	}{
		{
			name:        "valid integer environment variable",
			key:         "TEST_INT_VAR",
			fallback:    100,
			envValue:    "42",
			expected:    42,
			description: "should return the parsed integer value",
		},
		{
			name:        "environment variable does not exist",
			key:         "NONEXISTENT_INT_VAR",
			fallback:    100,
			envValue:    "",
			expected:    100,
			description: "should return the fallback value",
		},
		{
			name:        "invalid integer environment variable",
			key:         "TEST_INT_VAR",
			fallback:    100,
			envValue:    "not_a_number",
			expected:    100,
			description: "should return the fallback value for invalid integer",
		},
		{
			name:        "zero environment variable",
			key:         "TEST_INT_VAR",
			fallback:    100,
			envValue:    "0",
			expected:    0,
			description: "should return zero when environment variable is zero",
		},
		{
			name:        "negative integer environment variable",
			key:         "TEST_INT_VAR",
			fallback:    100,
			envValue:    "-42",
			expected:    -42,
			description: "should return negative integer value",
		},
		{
			name:        "large integer environment variable",
			key:         "TEST_INT_VAR",
			fallback:    100,
			envValue:    "999999",
			expected:    999999,
			description: "should return large integer value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvInt(tt.key, tt.fallback)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestEnvironmentConstants(t *testing.T) {
	// Test that environment constants are properly defined
	assert.Equal(t, "production", PRODUCTION_ENV)
	assert.Equal(t, "staging", STAGING_ENV)
	assert.Equal(t, "development", DEVELOPMENT_ENV)
}

func TestConfigValidation(t *testing.T) {
	// Test that LoadConfig calls validation
	// This test ensures that the validation is integrated with config loading

	// Save original environment variables
	originalEnv := make(map[string]string)
	envVars := []string{"APP_ENV", "APP_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"}

	for _, envVar := range envVars {
		if val := os.Getenv(envVar); val != "" {
			originalEnv[envVar] = val
		}
	}

	// Clean up after test
	defer func() {
		for envVar := range originalEnv {
			os.Setenv(envVar, originalEnv[envVar])
		}
		for _, envVar := range envVars {
			if _, exists := originalEnv[envVar]; !exists {
				os.Unsetenv(envVar)
			}
		}
	}()

	// Set valid environment variables
	os.Setenv("APP_ENV", "development")
	os.Setenv("APP_PORT", "8081")
	os.Setenv("POSTGRES_USER", "postgres")
	os.Setenv("POSTGRES_PASSWORD", "password")
	os.Setenv("POSTGRES_DB", "starter_db")
	os.Setenv("JWT_ACCESS_TOKEN_SECRET", "this-is-a-very-long-secret-key-for-access-tokens-32")
	os.Setenv("JWT_REFRESH_TOKEN_SECRET", "this-is-a-very-long-secret-key-for-refresh-tokens-32")

	// Load configuration should succeed with valid values
	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
}

// Benchmark tests for performance
func BenchmarkLoadConfig(b *testing.B) {
	// Save original environment variables
	originalEnv := make(map[string]string)
	envVars := []string{"APP_ENV", "APP_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"}

	for _, envVar := range envVars {
		if val := os.Getenv(envVar); val != "" {
			originalEnv[envVar] = val
		}
	}

	// Clean up after benchmark
	defer func() {
		for envVar := range originalEnv {
			os.Setenv(envVar, originalEnv[envVar])
		}
		for _, envVar := range envVars {
			if _, exists := originalEnv[envVar]; !exists {
				os.Unsetenv(envVar)
			}
		}
	}()

	// Set test environment variables
	os.Setenv("APP_ENV", "development")
	os.Setenv("APP_PORT", "8081")
	os.Setenv("POSTGRES_USER", "postgres")
	os.Setenv("POSTGRES_PASSWORD", "password")
	os.Setenv("POSTGRES_DB", "starter_db")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadConfig()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetEnv(b *testing.B) {
	// Set a test environment variable
	os.Setenv("BENCHMARK_TEST_VAR", "test_value")
	defer os.Unsetenv("BENCHMARK_TEST_VAR")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getEnv("BENCHMARK_TEST_VAR", "fallback")
	}
}

func BenchmarkGetEnvInt(b *testing.B) {
	// Set a test environment variable
	os.Setenv("BENCHMARK_TEST_INT_VAR", "42")
	defer os.Unsetenv("BENCHMARK_TEST_INT_VAR")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getEnvInt("BENCHMARK_TEST_INT_VAR", 100)
	}
}

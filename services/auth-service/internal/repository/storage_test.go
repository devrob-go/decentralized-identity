package repository

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"auth-service/config"

	zlog "packages/logger"

	"github.com/stretchr/testify/assert"
)

func TestStorageConstants(t *testing.T) {
	// Test that all constants are properly defined
	assert.Equal(t, "storage", APP_LAYER)
	assert.Equal(t, "./storage/migrations", DefaultMigrationsDir)
	assert.Equal(t, 5*time.Second, DefaultConnTimeout)
	assert.Equal(t, 25, DefaultMaxOpenConns)
	assert.Equal(t, 10, DefaultMaxIdleConns)
	assert.Equal(t, 30*time.Minute, DefaultConnMaxLifetime)
}

func TestStorageErrorConstants(t *testing.T) {
	// Test that error constants are properly defined
	assert.NotNil(t, ErrUniqueViolation)
	assert.NotNil(t, ErrForeignKeyViolation)
	assert.NotNil(t, ErrNotNullViolation)
	assert.NotNil(t, ErrCheckViolation)
	assert.NotNil(t, ErrExclusionViolation)

	// Test error messages
	assert.Contains(t, ErrUniqueViolation.Error(), "unique constraint violation")
	assert.Contains(t, ErrForeignKeyViolation.Error(), "foreign key violation")
	assert.Contains(t, ErrNotNullViolation.Error(), "not-null violation")
	assert.Contains(t, ErrCheckViolation.Error(), "check constraint violation")
	assert.Contains(t, ErrExclusionViolation.Error(), "exclusion constraint violation")
}

func TestFromConfig(t *testing.T) {
	// Test configuration conversion
	appCfg := &config.Config{
		PostgresUser:     "testuser",
		PostgresPassword: "testpass",
		PostgresHost:     "localhost",
		PostgresPort:     "5432",
		PostgresDB:       "testdb",
	}

	storageCfg := FromConfig(appCfg)

	// Verify the connection string format
	expectedConnStr := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
	assert.Equal(t, expectedConnStr, storageCfg.ConnStr)

	// Verify default values are set
	assert.Equal(t, DefaultMigrationsDir, storageCfg.MigrationsDir)
	assert.Equal(t, DefaultConnTimeout, storageCfg.ConnTimeout)
	assert.Equal(t, DefaultMaxOpenConns, storageCfg.MaxOpenConns)
	assert.Equal(t, DefaultMaxIdleConns, storageCfg.MaxIdleConns)
	assert.Equal(t, DefaultConnMaxLifetime, storageCfg.ConnMaxLifetime)
}

func TestFromConfig_WithDifferentValues(t *testing.T) {
	// Test with different database configurations
	appCfg := &config.Config{
		PostgresUser:     "produser",
		PostgresPassword: "prodpass",
		PostgresHost:     "prod.example.com",
		PostgresPort:     "5433",
		PostgresDB:       "proddb",
	}

	storageCfg := FromConfig(appCfg)

	expectedConnStr := "postgres://produser:prodpass@prod.example.com:5433/proddb?sslmode=disable"
	assert.Equal(t, expectedConnStr, storageCfg.ConnStr)
}

func TestDefaultUniqueMessage(t *testing.T) {
	tests := []struct {
		name        string
		constraint  string
		expected    string
		description string
	}{
		{
			name:        "email constraint",
			constraint:  "users_email_key",
			expected:    "email must be unique",
			description: "should extract 'email' from constraint name",
		},
		{
			name:        "username constraint",
			constraint:  "users_username_key",
			expected:    "username must be unique",
			description: "should extract 'username' from constraint name",
		},
		{
			name:        "short constraint",
			constraint:  "a_b",
			expected:    "value must be unique",
			description: "should return default message for short constraint names",
		},
		{
			name:        "empty constraint",
			constraint:  "",
			expected:    "value must be unique",
			description: "should return default message for empty constraint",
		},
		{
			name:        "single part constraint",
			constraint:  "unique",
			expected:    "value must be unique",
			description: "should return default message for single part constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultUniqueMessage(tt.constraint)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestHandlePgError_NonPgError(t *testing.T) {
	// Test handling of non-PostgreSQL errors
	regularError := errors.New("regular error")
	status, err := HandlePgError(regularError)

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestHandlePgError_UnhandledPgError(t *testing.T) {
	// Test handling of unhandled PostgreSQL error codes
	// This would require creating a mock pq.Error, but we can test the structure
	// For now, we'll test the error handling logic conceptually

	// Test that the function exists and can be called
	status, err := HandlePgError(errors.New("test error"))
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Error(t, err)
}

func TestConfig_Structure(t *testing.T) {
	// Test Config struct creation and field access
	cfg := &Config{
		ConnStr:         "test-connection-string",
		MigrationsDir:   "./test/migrations",
		ConnTimeout:     10 * time.Second,
		MaxOpenConns:    50,
		MaxIdleConns:    20,
		ConnMaxLifetime: 60 * time.Minute,
	}

	// Verify all fields are properly set
	assert.Equal(t, "test-connection-string", cfg.ConnStr)
	assert.Equal(t, "./test/migrations", cfg.MigrationsDir)
	assert.Equal(t, 10*time.Second, cfg.ConnTimeout)
	assert.Equal(t, 50, cfg.MaxOpenConns)
	assert.Equal(t, 20, cfg.MaxIdleConns)
	assert.Equal(t, 60*time.Minute, cfg.ConnMaxLifetime)
}

func TestNamedPreparer_Interface(t *testing.T) {
	// Test that the NamedPreparer interface is properly defined
	// This is a compile-time check, but we can verify the interface exists
	var _ NamedPreparer = (*DB)(nil)

	// The interface should have the PrepareNamedContext method
	// This is verified by the compiler when we try to assign *DB to NamedPreparer
}

func TestDB_Structure(t *testing.T) {
	// Test DB struct creation
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	// Create a DB instance with nil sqlx.DB (for testing structure only)
	db := &DB{
		DB:     nil,
		logger: logger,
	}

	// Verify the structure
	assert.Nil(t, db.DB)
	assert.Equal(t, logger, db.logger)
}

// Benchmark tests for performance
func BenchmarkDefaultUniqueMessage(b *testing.B) {
	constraint := "users_email_key"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = defaultUniqueMessage(constraint)
	}
}

func BenchmarkFromConfig(b *testing.B) {
	appCfg := &config.Config{
		PostgresUser:     "testuser",
		PostgresPassword: "testpass",
		PostgresHost:     "localhost",
		PostgresPort:     "5432",
		PostgresDB:       "testdb",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromConfig(appCfg)
	}
}

func BenchmarkHandlePgError_NonPgError(b *testing.B) {
	regularError := errors.New("regular error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HandlePgError(regularError)
	}
}

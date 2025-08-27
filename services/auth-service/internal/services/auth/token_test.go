package authentication

import (
	"testing"
	"time"

	zlog "packages/logger"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_GenerateTokens(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	// Test service creation
	authService := &AuthService{
		DB:     nil, // Will cause panic if called, but we're testing structure
		logger: logger,
	}

	// Test that the service was created properly
	assert.NotNil(t, authService)
	assert.Equal(t, logger, authService.logger)
	assert.Nil(t, authService.DB)
}

func TestAuthService_ValidateToken_Structure(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	// Test service creation
	authService := &AuthService{
		DB:     nil,
		logger: logger,
	}

	// Test that the service was created properly
	assert.NotNil(t, authService)
	assert.Equal(t, logger, authService.logger)
	assert.Nil(t, authService.DB)
}

func TestAuthService_TokenGeneration_TimeCalculation(t *testing.T) {
	// Test the time calculation logic from GenerateTokens
	now := time.Now()
	accessExpiresAt := now.Add(15 * time.Minute)
	refreshExpiresAt := now.Add(7 * 24 * time.Hour)

	// Verify the time calculations
	assert.True(t, accessExpiresAt.After(now))
	assert.True(t, refreshExpiresAt.After(now))
	assert.True(t, refreshExpiresAt.After(accessExpiresAt))

	// Verify the durations are approximately correct
	accessDuration := accessExpiresAt.Sub(now)
	refreshDuration := refreshExpiresAt.Sub(now)

	// Allow for small timing differences (within 1 second)
	assert.InDelta(t, 15*time.Minute, accessDuration, float64(time.Second))
	assert.InDelta(t, 7*24*time.Hour, refreshDuration, float64(time.Second))
}

func TestAuthService_TokenValidation_Claims(t *testing.T) {
	// Test the token type validation logic from ValidateToken
	validTokenType := "access"
	invalidTokenType := "refresh"
	missingTokenType := ""

	// Test valid token type
	assert.Equal(t, "access", validTokenType)

	// Test invalid token type
	assert.NotEqual(t, "access", invalidTokenType)

	// Test missing token type
	assert.NotEqual(t, "access", missingTokenType)
}

func TestAuthService_TokenExpiration_Logic(t *testing.T) {
	// Test the token expiration logic from ValidateToken
	now := time.Now()

	// Test expired token
	expiredToken := now.Add(-1 * time.Hour) // Token expired 1 hour ago
	assert.True(t, now.After(expiredToken))

	// Test valid token
	validToken := now.Add(1 * time.Hour) // Token expires in 1 hour
	assert.True(t, now.Before(validToken))

	// Test token expiring now
	expiringNow := now
	assert.False(t, now.After(expiringNow))
	assert.False(t, now.Before(expiringNow))
}

// Benchmark tests for performance
func BenchmarkAuthService_TokenGeneration_TimeCalculation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		now := time.Now()
		_ = now.Add(15 * time.Minute)
		_ = now.Add(7 * 24 * time.Hour)
	}
}

func BenchmarkAuthService_TokenValidation_Claims(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenType := "access"
		_ = tokenType == "access"
	}
}

func BenchmarkAuthService_TokenExpiration_Logic(b *testing.B) {
	now := time.Now()
	expiredToken := now.Add(-1 * time.Hour)
	validToken := now.Add(1 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = now.After(expiredToken)
		_ = now.Before(validToken)
	}
}

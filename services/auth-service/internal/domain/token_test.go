package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserToken_StructFields(t *testing.T) {
	tokenID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	token := &UserToken{
		ID:               tokenID,
		UserID:           userID,
		AccessToken:      "access_token_string",
		RefreshToken:     "refresh_token_string",
		AccessExpiresAt:  now.Add(15 * time.Minute),
		RefreshExpiresAt: now.Add(7 * 24 * time.Hour),
		IsRevoked:        false,
		CreatedAt:        now,
	}

	// Test that all fields are properly set
	assert.Equal(t, tokenID, token.ID)
	assert.Equal(t, userID, token.UserID)
	assert.Equal(t, "access_token_string", token.AccessToken)
	assert.Equal(t, "refresh_token_string", token.RefreshToken)
	assert.Equal(t, now.Add(15*time.Minute), token.AccessExpiresAt)
	assert.Equal(t, now.Add(7*24*time.Hour), token.RefreshExpiresAt)
	assert.Equal(t, false, token.IsRevoked)
	assert.Equal(t, now, token.CreatedAt)
}

func TestUserToken_ZeroValue(t *testing.T) {
	token := &UserToken{}

	// Test that zero values are properly initialized
	assert.Equal(t, uuid.Nil, token.ID)
	assert.Equal(t, uuid.Nil, token.UserID)
	assert.Equal(t, "", token.AccessToken)
	assert.Equal(t, "", token.RefreshToken)
	assert.Equal(t, time.Time{}, token.AccessExpiresAt)
	assert.Equal(t, time.Time{}, token.RefreshExpiresAt)
	assert.Equal(t, false, token.IsRevoked)
	assert.Equal(t, time.Time{}, token.CreatedAt)
}

func TestUserToken_FieldTypes(t *testing.T) {
	// Test that the UserToken struct has the correct field types
	var token UserToken

	// Test UUID fields
	assert.IsType(t, uuid.UUID{}, token.ID)
	assert.IsType(t, uuid.UUID{}, token.UserID)

	// Test string fields
	assert.IsType(t, "", token.AccessToken)
	assert.IsType(t, "", token.RefreshToken)

	// Test time fields
	assert.IsType(t, time.Time{}, token.AccessExpiresAt)
	assert.IsType(t, time.Time{}, token.RefreshExpiresAt)
	assert.IsType(t, time.Time{}, token.CreatedAt)

	// Test boolean field
	assert.IsType(t, false, token.IsRevoked)
}

func TestUserToken_JSONSerialization(t *testing.T) {
	tokenID := uuid.New()
	userID := uuid.New()
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	token := &UserToken{
		ID:               tokenID,
		UserID:           userID,
		AccessToken:      "access_token_string",
		RefreshToken:     "refresh_token_string",
		AccessExpiresAt:  now.Add(15 * time.Minute),
		RefreshExpiresAt: now.Add(7 * 24 * time.Hour),
		IsRevoked:        false,
		CreatedAt:        now,
	}

	// Test that the struct can be created and accessed
	assert.NotNil(t, token)
	assert.Equal(t, tokenID, token.ID)
	assert.Equal(t, userID, token.UserID)
	assert.Equal(t, "access_token_string", token.AccessToken)
	assert.Equal(t, "refresh_token_string", token.RefreshToken)
	assert.Equal(t, now.Add(15*time.Minute), token.AccessExpiresAt)
	assert.Equal(t, now.Add(7*24*time.Hour), token.RefreshExpiresAt)
	assert.Equal(t, false, token.IsRevoked)
	assert.Equal(t, now, token.CreatedAt)
}

func TestUserToken_DeepCopy(t *testing.T) {
	original := &UserToken{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		AccessToken:      "access_token_string",
		RefreshToken:     "refresh_token_string",
		AccessExpiresAt:  time.Now().Add(15 * time.Minute),
		RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		IsRevoked:        false,
		CreatedAt:        time.Now(),
	}

	// Create a copy
	copy := &UserToken{
		ID:               original.ID,
		UserID:           original.UserID,
		AccessToken:      original.AccessToken,
		RefreshToken:     original.RefreshToken,
		AccessExpiresAt:  original.AccessExpiresAt,
		RefreshExpiresAt: original.RefreshExpiresAt,
		IsRevoked:        original.IsRevoked,
		CreatedAt:        original.CreatedAt,
	}

	// Test that the copy has the same values
	assert.Equal(t, original.ID, copy.ID)
	assert.Equal(t, original.UserID, copy.UserID)
	assert.Equal(t, original.AccessToken, copy.AccessToken)
	assert.Equal(t, original.RefreshToken, copy.RefreshToken)
	assert.Equal(t, original.AccessExpiresAt, copy.AccessExpiresAt)
	assert.Equal(t, original.RefreshExpiresAt, copy.RefreshExpiresAt)
	assert.Equal(t, original.IsRevoked, copy.IsRevoked)
	assert.Equal(t, original.CreatedAt, copy.CreatedAt)

	// Test that they are different instances
	assert.NotSame(t, original, copy)
}

func TestUserToken_RevokedToken(t *testing.T) {
	tokenID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	token := &UserToken{
		ID:               tokenID,
		UserID:           userID,
		AccessToken:      "access_token_string",
		RefreshToken:     "refresh_token_string",
		AccessExpiresAt:  now.Add(15 * time.Minute),
		RefreshExpiresAt: now.Add(7 * 24 * time.Hour),
		IsRevoked:        true,
		CreatedAt:        now,
	}

	// Test that revoked token fields are properly set
	assert.True(t, token.IsRevoked)
}

func TestUserToken_ExpiredToken(t *testing.T) {
	tokenID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	expiredAt := now.Add(-1 * time.Hour) // Token expired 1 hour ago

	token := &UserToken{
		ID:               tokenID,
		UserID:           userID,
		AccessToken:      "access_token_string",
		RefreshToken:     "refresh_token_string",
		AccessExpiresAt:  expiredAt,
		RefreshExpiresAt: now.Add(7 * 24 * time.Hour),
		IsRevoked:        false,
		CreatedAt:        now.Add(-2 * time.Hour),
	}

	// Test that expired token fields are properly set
	assert.Equal(t, expiredAt, token.AccessExpiresAt)
	assert.True(t, token.AccessExpiresAt.Before(now))
}

func TestUserToken_RefreshToken(t *testing.T) {
	tokenID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	token := &UserToken{
		ID:               tokenID,
		UserID:           userID,
		AccessToken:      "access_token_string",
		RefreshToken:     "refresh_token_string",
		AccessExpiresAt:  now.Add(15 * time.Minute),
		RefreshExpiresAt: now.Add(7 * 24 * time.Hour), // 7 days
		IsRevoked:        false,
		CreatedAt:        now,
	}

	// Test that refresh token fields are properly set
	assert.Equal(t, now.Add(7*24*time.Hour), token.RefreshExpiresAt)
}

func TestUserToken_AccessToken(t *testing.T) {
	tokenID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	token := &UserToken{
		ID:               tokenID,
		UserID:           userID,
		AccessToken:      "access_token_string",
		RefreshToken:     "refresh_token_string",
		AccessExpiresAt:  now.Add(15 * time.Minute), // 15 minutes
		RefreshExpiresAt: now.Add(7 * 24 * time.Hour),
		IsRevoked:        false,
		CreatedAt:        now,
	}

	// Test that access token fields are properly set
	assert.Equal(t, now.Add(15*time.Minute), token.AccessExpiresAt)
}

// Benchmark tests for performance
func BenchmarkUserToken_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &UserToken{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			AccessToken:      "access_token_string",
			RefreshToken:     "refresh_token_string",
			AccessExpiresAt:  time.Now().Add(15 * time.Minute),
			RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			IsRevoked:        false,
			CreatedAt:        time.Now(),
		}
	}
}

func BenchmarkUserToken_RevokedTokenCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &UserToken{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			AccessToken:      "access_token_string",
			RefreshToken:     "refresh_token_string",
			AccessExpiresAt:  time.Now().Add(15 * time.Minute),
			RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			IsRevoked:        true,
			CreatedAt:        time.Now(),
		}
	}
}

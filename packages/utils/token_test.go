package utils

import (
	"testing"
	"time"

	"auth-service/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAccessTokenSimple(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		email    string
		userName string
		role     string
		secret   string
		wantErr  bool
	}{
		{
			name:     "valid token generation",
			userID:   "user123",
			email:    "test@example.com",
			userName: "Test User",
			role:     "user",
			secret:   "test-secret",
			wantErr:  false,
		},
		{
			name:     "empty secret",
			userID:   "user123",
			email:    "test@example.com",
			userName: "Test User",
			role:     "user",
			secret:   "",
			wantErr:  false,
		},
		{
			name:     "empty user data",
			userID:   "",
			email:    "",
			userName: "",
			role:     "",
			secret:   "test-secret",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateAccessTokenSimple(tt.userID, tt.email, tt.userName, tt.role, tt.secret)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		})
	}
}

func TestGenerateRefreshTokenSimple(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		secret  string
		wantErr bool
	}{
		{
			name:    "valid refresh token generation",
			userID:  "user123",
			secret:  "test-secret",
			wantErr: false,
		},
		{
			name:    "empty secret",
			userID:  "user123",
			secret:  "",
			wantErr: false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			secret:  "test-secret",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateRefreshTokenSimple(tt.userID, tt.secret)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		})
	}
}

func TestGenerateAccessToken(t *testing.T) {
	user := &models.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	}

	tests := []struct {
		name    string
		user    *models.User
		secret  string
		wantErr bool
	}{
		{
			name:    "valid user token generation",
			user:    user,
			secret:  "test-secret",
			wantErr: false,
		},
		{
			name:    "empty secret",
			user:    user,
			secret:  "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateAccessToken(tt.user, tt.secret)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	user := &models.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	}

	tests := []struct {
		name    string
		user    *models.User
		secret  string
		wantErr bool
	}{
		{
			name:    "valid user refresh token generation",
			user:    user,
			secret:  "test-secret",
			wantErr: false,
		},
		{
			name:    "empty secret",
			user:    user,
			secret:  "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateRefreshToken(tt.user, tt.secret)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		})
	}
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret"
	userID := "user123"
	email := "test@example.com"
	name := "Test User"
	role := "user"

	// Generate a valid token
	accessToken, err := GenerateAccessTokenSimple(userID, email, name, role, secret)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	refreshToken, err := GenerateRefreshTokenSimple(userID, secret)
	if err != nil {
		t.Fatalf("Failed to generate test refresh token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		secret  string
		wantErr bool
	}{
		{
			name:    "valid access token",
			token:   accessToken,
			secret:  secret,
			wantErr: false,
		},
		{
			name:    "valid refresh token",
			token:   refreshToken,
			secret:  secret,
			wantErr: false,
		},
		{
			name:    "invalid token",
			token:   "invalid.token.here",
			secret:  secret,
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  secret,
			wantErr: true,
		},
		{
			name:    "wrong secret",
			token:   accessToken,
			secret:  "wrong-secret",
			wantErr: true,
		},
		{
			name:    "empty secret",
			token:   accessToken,
			secret:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token, tt.secret)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, claims)

			// Verify claims for access token
			if tt.token == accessToken {
				assert.Equal(t, userID, claims["user_id"])
				assert.Equal(t, email, claims["email"])
				assert.Equal(t, name, claims["name"])
				assert.Equal(t, role, claims["role"])
				assert.Equal(t, "access", claims["type"])
			}

			// Verify claims for refresh token
			if tt.token == refreshToken {
				assert.Equal(t, userID, claims["user_id"])
				assert.Equal(t, "refresh", claims["type"])
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	secret := "test-secret"
	userID := "user123"
	email := "test@example.com"
	name := "Test User"
	role := "user"

	// Generate tokens
	accessToken, err := GenerateAccessTokenSimple(userID, email, name, role, secret)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	refreshToken, err := GenerateRefreshTokenSimple(userID, secret)
	if err != nil {
		t.Fatalf("Failed to generate test refresh token: %v", err)
	}

	// Validate tokens immediately
	accessClaims, err := ValidateToken(accessToken, secret)
	assert.NoError(t, err)
	assert.NotNil(t, accessClaims)

	refreshClaims, err := ValidateToken(refreshToken, secret)
	assert.NoError(t, err)
	assert.NotNil(t, refreshClaims)

	// Check expiration times
	accessExp, ok := accessClaims["exp"].(float64)
	assert.True(t, ok)
	refreshExp, ok := refreshClaims["exp"].(float64)
	assert.True(t, ok)

	now := time.Now().Unix()

	// Access token should expire in ~15 minutes
	assert.Greater(t, accessExp, float64(now))
	assert.Less(t, accessExp, float64(now+16*60)) // Should be less than 16 minutes from now

	// Refresh token should expire in ~7 days
	assert.Greater(t, refreshExp, float64(now))
	assert.Less(t, refreshExp, float64(now+8*24*60*60)) // Should be less than 8 days from now
}

func TestTokenClaims(t *testing.T) {
	secret := "test-secret"
	userID := "user123"
	email := "test@example.com"
	name := "Test User"
	role := "user"

	// Generate access token
	accessToken, err := GenerateAccessTokenSimple(userID, email, name, role, secret)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Validate and check claims
	claims, err := ValidateToken(accessToken, secret)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// Check required claims
	assert.Equal(t, userID, claims["user_id"])
	assert.Equal(t, email, claims["email"])
	assert.Equal(t, name, claims["name"])
	assert.Equal(t, role, claims["role"])
	assert.Equal(t, "access", claims["type"])

	// Check timestamp claims
	assert.NotNil(t, claims["iat"]) // issued at
	assert.NotNil(t, claims["exp"]) // expiration
}

// Benchmark tests for performance
func BenchmarkGenerateAccessTokenSimple(b *testing.B) {
	userID := "user123"
	email := "test@example.com"
	name := "Test User"
	role := "user"
	secret := "test-secret"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateAccessTokenSimple(userID, email, name, role, secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateRefreshTokenSimple(b *testing.B) {
	userID := "user123"
	secret := "test-secret"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateRefreshTokenSimple(userID, secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateToken(b *testing.B) {
	secret := "test-secret"
	userID := "user123"
	email := "test@example.com"
	name := "Test User"
	role := "user"

	token, err := GenerateAccessTokenSimple(userID, email, name, role, secret)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValidateToken(token, secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}

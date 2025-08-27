package utils

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "testpassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "verylongpasswordwithspecialchars!@#$%^&*()",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			if err != nil {
				t.Fatalf("HashPassword() error = %v", err)
			}
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash)
			assert.Len(t, hash, 60) // bcrypt hash length
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password and hash",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPasswordHash(tt.password, tt.hash)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// Test that VerifyPassword is an alias for CheckPasswordHash
	result1 := CheckPasswordHash(password, hash)
	result2 := VerifyPassword(password, hash)
	assert.Equal(t, result1, result2)
	assert.True(t, result1)
	assert.True(t, result2)
}

func TestComparePasswords(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	tests := []struct {
		name          string
		storedHash    sql.NullString
		plainPassword string
		want          bool
	}{
		{
			name: "valid hash and password",
			storedHash: sql.NullString{
				String: hash,
				Valid:  true,
			},
			plainPassword: password,
			want:          true,
		},
		{
			name: "invalid hash and password",
			storedHash: sql.NullString{
				String: hash,
				Valid:  true,
			},
			plainPassword: "wrongpassword",
			want:          false,
		},
		{
			name: "null hash",
			storedHash: sql.NullString{
				String: "",
				Valid:  false,
			},
			plainPassword: password,
			want:          false,
		},
		{
			name: "empty hash string",
			storedHash: sql.NullString{
				String: "",
				Valid:  true,
			},
			plainPassword: password,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComparePasswords(tt.storedHash, tt.plainPassword)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestHashToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "simple token",
			token: "testtoken",
		},
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "long token",
			token: "verylongtokenwithspecialchars!@#$%^&*()",
		},
		{
			name:  "token with spaces",
			token: "token with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashToken(tt.token)
			assert.NotEmpty(t, hash)
			assert.Len(t, hash, 64) // SHA256 hex string length

			// Test that the same token always produces the same hash
			hash2 := HashToken(tt.token)
			assert.Equal(t, hash, hash2)

			// Test that different tokens produce different hashes
			if tt.token != "" {
				differentHash := HashToken(tt.token + "different")
				assert.NotEqual(t, hash, differentHash)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkHashPassword(b *testing.B) {
	password := "testpassword123"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckPasswordHash(b *testing.B) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPasswordHash(password, hash)
	}
}

func BenchmarkHashToken(b *testing.B) {
	token := "testtoken123"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HashToken(token)
	}
}

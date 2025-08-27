package repository

import (
	"testing"

	"auth-service/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidateUserCreate_ValidUser(t *testing.T) {
	// Test with valid user data
	user := &models.User{
		ID:       uuid.New(),
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
	}

	err := ValidateUserCreate(user)
	assert.NoError(t, err, "valid user should pass validation")
}

func TestValidateUserCreate_InvalidName(t *testing.T) {
	tests := []struct {
		name        string
		userName    string
		description string
	}{
		{
			name:        "empty name",
			userName:    "",
			description: "should fail validation for empty name",
		},
		{
			name:        "name too long",
			userName:    "This is a very long name that exceeds the maximum allowed length for a user name in the system and should fail validation",
			description: "should fail validation for name longer than 100 characters",
		},
		{
			name:        "single character name",
			userName:    "J",
			description: "should pass validation for single character name",
		},
		{
			name:        "exactly 100 characters",
			userName:    "This is exactly one hundred characters long name that should pass validation in the system",
			description: "should pass validation for exactly 100 character name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				ID:       uuid.New(),
				Name:     tt.userName,
				Email:    "test@example.com",
				Password: "password123",
			}

			err := ValidateUserCreate(user)
			if tt.userName == "" || len(tt.userName) > 100 {
				assert.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), "validation failed")
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestValidateUserCreate_InvalidEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		description string
	}{
		{
			name:        "empty email",
			email:       "",
			description: "should fail validation for empty email",
		},
		{
			name:        "invalid email format",
			email:       "invalid-email",
			description: "should fail validation for invalid email format",
		},
		{
			name:        "email too long",
			email:       "this.is.a.very.long.email.address.that.exceeds.the.maximum.allowed.length.for.an.email.address.in.the.system@example.com",
			description: "should fail validation for email longer than 100 characters",
		},
		{
			name:        "valid email",
			email:       "test@example.com",
			description: "should pass validation for valid email",
		},
		{
			name:        "email with subdomain",
			email:       "test@sub.example.com",
			description: "should pass validation for email with subdomain",
		},
		{
			name:        "email with plus",
			email:       "test+tag@example.com",
			description: "should pass validation for email with plus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				ID:       uuid.New(),
				Name:     "Test User",
				Email:    tt.email,
				Password: "password123",
			}

			err := ValidateUserCreate(user)
			if tt.email == "" || len(tt.email) > 100 || tt.email == "invalid-email" {
				assert.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), "validation failed")
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestValidateUserCreate_InvalidPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		description string
	}{
		{
			name:        "empty password",
			password:    "",
			description: "should fail validation for empty password",
		},
		{
			name:        "password too short",
			password:    "12345",
			description: "should fail validation for password shorter than 6 characters",
		},
		{
			name:        "password exactly 6 characters",
			password:    "123456",
			description: "should pass validation for exactly 6 character password",
		},
		{
			name:        "password exactly 255 characters",
			password:    repeatString("a", 255),
			description: "should pass validation for exactly 255 character password",
		},
		{
			name:        "password too long",
			password:    repeatString("a", 256),
			description: "should fail validation for password longer than 255 characters",
		},
		{
			name:        "valid password",
			password:    "password123",
			description: "should pass validation for valid password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				ID:       uuid.New(),
				Name:     "Test User",
				Email:    "test@example.com",
				Password: tt.password,
			}

			err := ValidateUserCreate(user)
			if tt.password == "" || len(tt.password) < 6 || len(tt.password) > 255 {
				assert.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), "validation failed")
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestValidateUserCreate_MultipleValidationErrors(t *testing.T) {
	// Test user with multiple validation errors
	user := &models.User{
		ID:       uuid.New(),
		Name:     "",              // Invalid: empty name
		Email:    "invalid-email", // Invalid: wrong format
		Password: "123",           // Invalid: too short
	}

	err := ValidateUserCreate(user)
	assert.Error(t, err, "user with multiple validation errors should fail")
	assert.Contains(t, err.Error(), "validation failed")
}

func TestValidateUserCreate_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		user        *models.User
		description string
	}{
		{
			name: "minimum valid user",
			user: &models.User{
				ID:       uuid.New(),
				Name:     "A",
				Email:    "a@b.com",
				Password: "123456",
			},
			description: "should pass validation for minimum valid values",
		},
		{
			name: "maximum valid user",
			user: &models.User{
				ID:       uuid.New(),
				Name:     repeatString("a", 100),
				Email:    repeatString("a", 80) + "@example.com", // 80 + 13 = 93 characters
				Password: repeatString("a", 255),
			},
			description: "should pass validation for maximum valid values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserCreate(tt.user)
			assert.NoError(t, err, tt.description)
		})
	}
}

// Helper function to repeat a string
func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// Benchmark tests for performance
func BenchmarkValidateUserCreate_ValidUser(b *testing.B) {
	user := &models.User{
		ID:       uuid.New(),
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateUserCreate(user)
	}
}

func BenchmarkValidateUserCreate_InvalidUser(b *testing.B) {
	user := &models.User{
		ID:       uuid.New(),
		Name:     "",
		Email:    "invalid-email",
		Password: "123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateUserCreate(user)
	}
}

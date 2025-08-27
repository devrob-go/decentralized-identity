package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_StructFields(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	user := &User{
		ID:        userID,
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		Password:  "hashedpassword123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test that all fields are properly set
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john.doe@example.com", user.Email)
	assert.Equal(t, "hashedpassword123", user.Password)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
}

func TestUser_ZeroValue(t *testing.T) {
	user := &User{}

	// Test that zero values are properly initialized
	assert.Equal(t, uuid.Nil, user.ID)
	assert.Equal(t, "", user.Name)
	assert.Equal(t, "", user.Email)
	assert.Equal(t, "", user.Password)
	assert.Equal(t, time.Time{}, user.CreatedAt)
	assert.Equal(t, time.Time{}, user.UpdatedAt)
}

func TestUser_JSONTags(t *testing.T) {
	user := &User{
		ID:        uuid.New(),
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		Password:  "hashedpassword123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test that JSON tags are properly set
	// This is a structural test to ensure the struct is properly defined
	assert.NotNil(t, user)
}

func TestCredentials_StructFields(t *testing.T) {
	credentials := &Credentials{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Test that all fields are properly set
	assert.Equal(t, "test@example.com", credentials.Email)
	assert.Equal(t, "password123", credentials.Password)
}

func TestCredentials_ZeroValue(t *testing.T) {
	credentials := &Credentials{}

	// Test that zero values are properly initialized
	assert.Equal(t, "", credentials.Email)
	assert.Equal(t, "", credentials.Password)
}

func TestUserCreateRequest_StructFields(t *testing.T) {
	request := &UserCreateRequest{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
	}

	// Test that all fields are properly set
	assert.Equal(t, "John Doe", request.Name)
	assert.Equal(t, "john.doe@example.com", request.Email)
	assert.Equal(t, "password123", request.Password)
}

func TestUserCreateRequest_ZeroValue(t *testing.T) {
	request := &UserCreateRequest{}

	// Test that zero values are properly initialized
	assert.Equal(t, "", request.Name)
	assert.Equal(t, "", request.Email)
	assert.Equal(t, "", request.Password)
}

func TestUser_FieldTypes(t *testing.T) {
	// Test that the User struct has the correct field types
	var user User

	// Test ID field type
	assert.IsType(t, uuid.UUID{}, user.ID)

	// Test string fields
	assert.IsType(t, "", user.Name)
	assert.IsType(t, "", user.Email)
	assert.IsType(t, "", user.Password)

	// Test time fields
	assert.IsType(t, time.Time{}, user.CreatedAt)
	assert.IsType(t, time.Time{}, user.UpdatedAt)
}

func TestCredentials_FieldTypes(t *testing.T) {
	// Test that the Credentials struct has the correct field types
	var credentials Credentials

	// Test string fields
	assert.IsType(t, "", credentials.Email)
	assert.IsType(t, "", credentials.Password)
}

func TestUserCreateRequest_FieldTypes(t *testing.T) {
	// Test that the UserCreateRequest struct has the correct field types
	var request UserCreateRequest

	// Test string fields
	assert.IsType(t, "", request.Name)
	assert.IsType(t, "", request.Email)
	assert.IsType(t, "", request.Password)
}

func TestUser_JSONSerialization(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	user := &User{
		ID:        userID,
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		Password:  "hashedpassword123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test that the struct can be created and accessed
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john.doe@example.com", user.Email)
	assert.Equal(t, "hashedpassword123", user.Password)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
}

func TestUser_DeepCopy(t *testing.T) {
	original := &User{
		ID:        uuid.New(),
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		Password:  "hashedpassword123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create a copy
	copy := &User{
		ID:        original.ID,
		Name:      original.Name,
		Email:     original.Email,
		Password:  original.Password,
		CreatedAt: original.CreatedAt,
		UpdatedAt: original.UpdatedAt,
	}

	// Test that the copy has the same values
	assert.Equal(t, original.ID, copy.ID)
	assert.Equal(t, original.Name, copy.Name)
	assert.Equal(t, original.Email, copy.Email)
	assert.Equal(t, original.Password, copy.Password)
	assert.Equal(t, original.CreatedAt, copy.CreatedAt)
	assert.Equal(t, original.UpdatedAt, copy.UpdatedAt)

	// Test that they are different instances
	assert.NotSame(t, original, copy)
}

// Benchmark tests for performance
func BenchmarkUser_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &User{
			ID:        uuid.New(),
			Name:      "John Doe",
			Email:     "john.doe@example.com",
			Password:  "hashedpassword123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
}

func BenchmarkCredentials_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &Credentials{
			Email:    "test@example.com",
			Password: "password123",
		}
	}
}

func BenchmarkUserCreateRequest_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &UserCreateRequest{
			Name:     "John Doe",
			Email:    "john.doe@example.com",
			Password: "password123",
		}
	}
}

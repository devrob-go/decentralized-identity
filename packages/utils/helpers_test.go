package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "valid password with digit and uppercase",
			password: "Password123",
			expected: true,
		},
		{
			name:     "valid password with digit and uppercase, different order",
			password: "123Password",
			expected: true,
		},
		{
			name:     "valid password with special characters",
			password: "Pass@word123",
			expected: true,
		},
		{
			name:     "password too short",
			password: "Pass1",
			expected: false,
		},
		{
			name:     "password missing digit",
			password: "Password",
			expected: false,
		},
		{
			name:     "password missing uppercase",
			password: "password123",
			expected: false,
		},
		{
			name:     "password missing both digit and uppercase",
			password: "password",
			expected: false,
		},
		{
			name:     "empty password",
			password: "",
			expected: false,
		},
		{
			name:     "exactly 8 characters with requirements",
			password: "Pass1word",
			expected: true,
		},
		{
			name:     "long password with requirements",
			password: "VeryLongPassword123",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePassword(tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidDate(t *testing.T) {
	now := time.Now()
	zeroTime := time.Time{}

	tests := []struct {
		name     string
		value    any
		expected error
	}{
		{
			name:     "valid time.Time",
			value:    now,
			expected: nil,
		},
		{
			name:     "valid *time.Time",
			value:    &now,
			expected: nil,
		},
		{
			name:     "zero time.Time",
			value:    zeroTime,
			expected: fmt.Errorf("date_of_birth is required and must be a valid date"),
		},
		{
			name:     "nil *time.Time",
			value:    (*time.Time)(nil),
			expected: fmt.Errorf("date_of_birth is required and must be a valid date"),
		},
		{
			name:     "zero *time.Time",
			value:    &zeroTime,
			expected: fmt.Errorf("date_of_birth is required and must be a valid date"),
		},
		{
			name:     "string value",
			value:    "2023-01-01",
			expected: fmt.Errorf("date_of_birth must be a valid date"),
		},
		{
			name:     "int value",
			value:    123,
			expected: fmt.Errorf("date_of_birth must be a valid date"),
		},
		{
			name:     "nil value",
			value:    nil,
			expected: fmt.Errorf("date_of_birth must be a valid date"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidDate(tt.value)
			if tt.expected == nil {
				assert.NoError(t, result)
			} else {
				assert.Error(t, result)
				assert.Equal(t, tt.expected.Error(), result.Error())
			}
		})
	}
}

func TestValidateAlphaNumericSpace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "alphanumeric with spaces",
			input:    "Hello World 123",
			expected: true,
		},
		{
			name:     "only letters and spaces",
			input:    "Hello World",
			expected: true,
		},
		{
			name:     "only numbers and spaces",
			input:    "123 456 789",
			expected: true,
		},
		{
			name:     "mixed alphanumeric with spaces",
			input:    "User123 Name456",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "single character",
			input:    "a",
			expected: true,
		},
		{
			name:     "single number",
			input:    "1",
			expected: true,
		},
		{
			name:     "single space",
			input:    " ",
			expected: true,
		},
		{
			name:     "multiple spaces",
			input:    "   ",
			expected: true,
		},
		{
			name:     "with special characters",
			input:    "Hello@World",
			expected: false,
		},
		{
			name:     "with punctuation",
			input:    "Hello, World!",
			expected: false,
		},
		{
			name:     "with underscore",
			input:    "Hello_World",
			expected: false,
		},
		{
			name:     "with dash",
			input:    "Hello-World",
			expected: false,
		},
		{
			name:     "with newline",
			input:    "Hello\nWorld",
			expected: false,
		},
		{
			name:     "with tab",
			input:    "Hello\tWorld",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAlphaNumericSpace(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests for performance
func BenchmarkValidatePassword(b *testing.B) {
	password := "Password123"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidatePassword(password)
	}
}

func BenchmarkValidDate(b *testing.B) {
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidDate(now)
	}
}

func BenchmarkValidateAlphaNumericSpace(b *testing.B) {
	input := "Hello World 123"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateAlphaNumericSpace(input)
	}
}

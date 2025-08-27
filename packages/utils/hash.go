package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// PasswordStrength represents password strength levels
type PasswordStrength int

const (
	PasswordStrengthWeak PasswordStrength = iota
	PasswordStrengthMedium
	PasswordStrengthStrong
	PasswordStrengthVeryStrong
)

// HashPassword creates a secure hash of a password using bcrypt
func HashPassword(password string) (string, error) {
	// Use higher cost for production (12 is default, 14+ for high security)
	cost := 14
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// HashPasswordWithCost creates a hash with a specific cost factor
func HashPasswordWithCost(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return "", fmt.Errorf("cost must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// CheckPasswordHash verifies a password against its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// VerifyPassword is an alias for CheckPasswordHash for better readability
func VerifyPassword(password, hash string) bool {
	return CheckPasswordHash(password, hash)
}

// ComparePasswords compares a stored hash with a plain password
func ComparePasswords(storedHash sql.NullString, plainPassword string) bool {
	if !storedHash.Valid {
		fmt.Println("Stored hash invalid (null)")
		return false
	}
	fmt.Println("Comparing plain:", plainPassword, "with hash:", storedHash.String)
	err := bcrypt.CompareHashAndPassword([]byte(storedHash.String), []byte(plainPassword))
	if err != nil {
		fmt.Println("bcrypt compare error:", err)
	} else {
		fmt.Println("Passwords match!")
	}
	return err == nil
}

// ValidatePasswordStrength validates password strength according to security policy
func ValidatePasswordStrength(password string, minLength int, requireUppercase, requireLowercase, requireNumbers, requireSpecialChars bool) error {
	if len(password) < minLength {
		return fmt.Errorf("password must be at least %d characters long", minLength)
	}

	if requireUppercase && !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if requireLowercase && !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if requireNumbers && !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one number")
	}

	if requireSpecialChars && !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check for common weak patterns
	if isCommonWeakPassword(password) {
		return fmt.Errorf("password is too common or follows a predictable pattern")
	}

	return nil
}

// GetPasswordStrength calculates password strength score
func GetPasswordStrength(password string) PasswordStrength {
	score := 0

	// Length bonus
	if len(password) >= 8 {
		score += 1
	}
	if len(password) >= 12 {
		score += 1
	}
	if len(password) >= 16 {
		score += 1
	}

	// Character variety bonus
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		score += 1
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		score += 1
	}
	if regexp.MustCompile(`[0-9]`).MatchString(password) {
		score += 1
	}
	if regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		score += 1
	}

	// Penalty for common patterns
	if isCommonWeakPassword(password) {
		score -= 2
	}

	// Penalty for repeated characters
	if hasRepeatedChars(password) {
		score -= 1
	}

	// Penalty for sequential characters
	if hasSequentialChars(password) {
		score -= 1
	}

	// Determine strength level
	switch {
	case score <= 2:
		return PasswordStrengthWeak
	case score <= 4:
		return PasswordStrengthMedium
	case score <= 6:
		return PasswordStrengthStrong
	default:
		return PasswordStrengthVeryStrong
	}
}

// isCommonWeakPassword checks for common weak password patterns
func isCommonWeakPassword(password string) bool {
	password = strings.ToLower(password)
	
	// Common weak passwords
	weakPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome", "monkey",
		"dragon", "master", "sunshine", "princess", "shadow",
		"football", "baseball", "basketball", "superman", "batman",
	}

	for _, weak := range weakPasswords {
		if password == weak {
			return true
		}
	}

	// Check for keyboard patterns
	keyboardPatterns := []string{
		"qwerty", "asdfgh", "zxcvbn", "123456", "654321",
	}

	for _, pattern := range keyboardPatterns {
		if strings.Contains(password, pattern) {
			return true
		}
	}

	return false
}

// hasRepeatedChars checks for excessive character repetition
func hasRepeatedChars(password string) bool {
	if len(password) < 3 {
		return false
	}

	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i] == password[i+2] {
			return true
		}
	}
	return false
}

// hasSequentialChars checks for sequential character patterns
func hasSequentialChars(password string) bool {
	if len(password) < 3 {
		return false
	}

	for i := 0; i < len(password)-2; i++ {
		// Check for sequential letters
		if password[i]+1 == password[i+1] && password[i+1]+1 == password[i+2] {
			return true
		}
		// Check for sequential numbers
		if password[i] >= '0' && password[i] <= '7' &&
			password[i+1] == password[i]+1 &&
			password[i+2] == password[i]+2 {
			return true
		}
	}
	return false
}

// GenerateSecurePassword generates a cryptographically secure random password
func GenerateSecurePassword(length int, includeUppercase, includeLowercase, includeNumbers, includeSpecialChars bool) (string, error) {
	if length < 8 {
		return "", fmt.Errorf("password length must be at least 8")
	}

	var charset string
	if includeLowercase {
		charset += "abcdefghijklmnopqrstuvwxyz"
	}
	if includeUppercase {
		charset += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	if includeNumbers {
		charset += "0123456789"
	}
	if includeSpecialChars {
		charset += "!@#$%^&*()_+-=[]{}|;:,.<>?"
	}

	if charset == "" {
		return "", fmt.Errorf("at least one character set must be selected")
	}

	password := make([]byte, length)
	for i := range password {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random password: %w", err)
		}
		password[i] = charset[randomIndex.Int64()]
	}

	return string(password), nil
}

// hashToken returns a SHA256 hash of a token
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	if length < 16 {
		return "", fmt.Errorf("token length must be at least 16")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

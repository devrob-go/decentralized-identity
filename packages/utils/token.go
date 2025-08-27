package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"auth-service/models"

	"github.com/golang-jwt/jwt"
)

// TokenConfig holds JWT configuration
type TokenConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	EncryptionKey      string
}

// GenerateAccessTokenSimple creates a new access token for a user (simplified version)
func GenerateAccessTokenSimple(userID string, email string, name string, role string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"name":    name,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshTokenSimple creates a new refresh token for a user (simplified version)
func GenerateRefreshTokenSimple(userID string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateAccessToken creates a new access token for a user
func GenerateAccessToken(user *models.User, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"name":    user.Name,
		"email":   user.Email,
		"exp":     time.Now().Add(15 * time.Minute).Unix(), // Reduced from 7 days to 15 minutes for security
		"iat":     time.Now().Unix(),
		"type":    "access",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken creates a new refresh token for a user
func GenerateRefreshToken(user *models.User, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"name":    user.Name,
		"email":   user.Email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GenerateTimedCode creates a time-based encrypted code
func GenerateTimedCode(encryptionKey string) (string, error) {
	if len(encryptionKey) < 32 {
		return "", fmt.Errorf("encryption key must be at least 32 bytes")
	}

	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	randomPart := base64.URLEncoding.EncodeToString(randomBytes)
	randomPart = strings.TrimRight(randomPart, "=")
	raw := fmt.Sprintf("%d.%s", timestamp, randomPart)

	return EncryptCode(raw, encryptionKey)
}

// IsCodeExpired checks if a time-based code has expired
func IsCodeExpired(encryptedCode string, validDuration time.Duration, encryptionKey string) (bool, error) {
	decrypted, err := DecryptCode(encryptedCode, encryptionKey)
	if err != nil {
		return true, fmt.Errorf("failed to decrypt code: %w", err)
	}

	parts := strings.Split(decrypted, ".")
	if len(parts) != 2 {
		return true, fmt.Errorf("invalid decrypted code format")
	}

	timestampInt, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return true, fmt.Errorf("invalid timestamp in code")
	}

	codeTime := time.Unix(timestampInt, 0)
	if time.Since(codeTime) > validDuration {
		return true, nil
	}

	return false, nil
}

// EncryptCode encrypts plain text using AES-GCM
func EncryptCode(plainText string, encryptionKey string) (string, error) {
	if len(encryptionKey) < 32 {
		return "", fmt.Errorf("encryption key must be at least 32 bytes")
	}

	key := []byte(encryptionKey)
	if len(key) > 32 {
		key = key[:32] // Truncate to 32 bytes for AES-256
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.URLEncoding.EncodeToString(cipherText), nil
}

// DecryptCode decrypts encrypted text using AES-GCM
func DecryptCode(encrypted string, encryptionKey string) (string, error) {
	if len(encryptionKey) < 32 {
		return "", fmt.Errorf("encryption key must be at least 32 bytes")
	}

	key := []byte(encryptionKey)
	if len(key) > 32 {
		key = key[:32] // Truncate to 32 bytes for AES-256
	}

	cipherText, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(cipherText) < nonceSize {
		return "", fmt.Errorf("invalid ciphertext length")
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plainText), nil
}

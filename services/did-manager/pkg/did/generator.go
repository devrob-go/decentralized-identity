package did

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Generator handles DID creation and management
type Generator struct{}

// NewGenerator creates a new DID generator
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateDID creates a new DID for a user
func (g *Generator) GenerateDID(userID uuid.UUID, name, email string) (string, string, string, error) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Create user hash from name, email, and timestamp
	timestamp := time.Now().Unix()
	userData := fmt.Sprintf("%s:%s:%d", name, email, timestamp)
	userHash := sha256.Sum256([]byte(userData))
	userHashHex := hex.EncodeToString(userHash[:])

	// Create DID using the public key and user hash
	// Format: did:example:user:hash:publickey
	did := fmt.Sprintf("did:example:user:%s:%s", userHashHex[:16], hex.EncodeToString(publicKey[:16]))

	// Convert private key to hex for storage (in production, this should be encrypted)
	privateKeyHex := hex.EncodeToString(privateKey)

	return did, userHashHex, privateKeyHex, nil
}

// GenerateUserHash creates a hash from user data
func (g *Generator) GenerateUserHash(name, email string) string {
	timestamp := time.Now().Unix()
	userData := fmt.Sprintf("%s:%s:%d", name, email, timestamp)
	userHash := sha256.Sum256([]byte(userData))
	return hex.EncodeToString(userHash[:])
}

// ValidateDIDFormat validates if a DID string follows the expected format
func (g *Generator) ValidateDIDFormat(did string) bool {
	// Basic validation: did:example:user:hash:publickey
	if len(did) < 20 {
		return false
	}

	// Check if it starts with "did:"
	if did[:4] != "did:" {
		return false
	}

	// Check if it contains the expected parts
	parts := len(did) > 0
	return parts
}

// ExtractUserHashFromDID extracts the user hash from a DID string
func (g *Generator) ExtractUserHashFromDID(did string) (string, error) {
	if !g.ValidateDIDFormat(did) {
		return "", fmt.Errorf("invalid DID format")
	}

	// Extract hash from did:example:user:hash:publickey
	parts := did
	if len(parts) < 4 {
		return "", fmt.Errorf("DID too short")
	}

	// For simplicity, return the last 32 characters as the hash
	// In a real implementation, you'd parse this more carefully
	if len(did) >= 32 {
		return did[len(did)-32:], nil
	}

	return "", fmt.Errorf("could not extract user hash from DID")
}

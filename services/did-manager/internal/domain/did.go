package domain

import (
	"time"

	"github.com/google/uuid"
)

// DID represents a Decentralized Identifier
type DID struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Did          string    `json:"did" db:"did"`
	UserHash     string    `json:"user_hash" db:"user_hash"`
	PublicKey    string    `json:"public_key" db:"public_key"`
	Status       string    `json:"status" db:"status"` // active, revoked, expired
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	BlockchainTx string    `json:"blockchain_tx" db:"blockchain_tx"`
}

// DIDCreateRequest represents a request to create a new DID
type DIDCreateRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	Name     string    `json:"name" binding:"required"`
	Email    string    `json:"email" binding:"required,email"`
	Password string    `json:"password" binding:"required"`
}

// DIDResponse represents the response after DID creation
type DIDResponse struct {
	DID      *DID   `json:"did"`
	UserHash string `json:"user_hash"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// DIDVerificationRequest represents a request to verify a DID
type DIDVerificationRequest struct {
	DID      string `json:"did" binding:"required"`
	UserHash string `json:"user_hash"`
}

// DIDVerificationResponse represents the response after DID verification
type DIDVerificationResponse struct {
	IsValid      bool   `json:"is_valid"`
	DID          string `json:"did"`
	UserHash     string `json:"user_hash"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	BlockchainTx string `json:"blockchain_tx"`
}

// DIDStatus represents the current status of a DID
type DIDStatus string

const (
	DIDStatusPending DIDStatus = "pending"
	DIDStatusActive  DIDStatus = "active"
	DIDStatusRevoked DIDStatus = "revoked"
	DIDStatusExpired DIDStatus = "expired"
	DIDStatusFailed  DIDStatus = "failed"
)

// DIDRepository defines the interface for DID data operations
type DIDRepository interface {
	Create(did *DID) error
	GetByID(id uuid.UUID) (*DID, error)
	GetByDID(did string) (*DID, error)
	GetByUserID(userID uuid.UUID) (*DID, error)
	GetByUserHash(userHash string) (*DID, error)
	Update(did *DID) error
	UpdateStatus(id uuid.UUID, status string, txHash string) error
	ListByStatus(status string) ([]*DID, error)
}

// DIDService defines the interface for DID business logic
type DIDService interface {
	CreateDID(req *DIDCreateRequest) (*DIDResponse, error)
	VerifyDID(req *DIDVerificationRequest) (*DIDVerificationResponse, error)
	GetDIDByUserID(userID uuid.UUID) (*DID, error)
	UpdateDIDStatus(didID uuid.UUID, status string, txHash string) error
	ProcessBlockchainQueue() error
}

package repository

import (
	"database/sql"
	"fmt"
	"log"

	"did-manager/internal/domain"

	"github.com/google/uuid"
)

// DIDRepository implements the DID repository interface
type DIDRepository struct {
	db *sql.DB
}

// NewDIDRepository creates a new DID repository
func NewDIDRepository(db *sql.DB) *DIDRepository {
	return &DIDRepository{db: db}
}

// Create creates a new DID record
func (r *DIDRepository) Create(did *domain.DID) error {
	query := `
		INSERT INTO dids (id, user_id, did, user_hash, public_key, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(query,
		did.ID,
		did.UserID,
		did.Did,
		did.UserHash,
		did.PublicKey,
		did.Status,
		did.CreatedAt,
		did.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create DID: %w", err)
	}

	return nil
}

// GetByID retrieves a DID by ID
func (r *DIDRepository) GetByID(id uuid.UUID) (*domain.DID, error) {
	query := `
		SELECT id, user_id, did, user_hash, public_key, status, created_at, updated_at, blockchain_tx
		FROM dids WHERE id = $1
	`

	var did domain.DID
	err := r.db.QueryRow(query, id).Scan(
		&did.ID,
		&did.UserID,
		&did.Did,
		&did.UserHash,
		&did.PublicKey,
		&did.Status,
		&did.CreatedAt,
		&did.UpdatedAt,
		&did.BlockchainTx,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("DID not found")
		}
		return nil, fmt.Errorf("failed to get DID: %w", err)
	}

	return &did, nil
}

// GetByDID retrieves a DID by DID string
func (r *DIDRepository) GetByDID(didString string) (*domain.DID, error) {
	query := `
		SELECT id, user_id, did, user_hash, public_key, status, created_at, updated_at, COALESCE(blockchain_tx, '') as blockchain_tx
		FROM dids WHERE did = $1
	`

	log.Printf("DEBUG: Searching for DID: %s", didString)

	var did domain.DID
	err := r.db.QueryRow(query, didString).Scan(
		&did.ID,
		&did.UserID,
		&did.Did,
		&did.UserHash,
		&did.PublicKey,
		&did.Status,
		&did.CreatedAt,
		&did.UpdatedAt,
		&did.BlockchainTx,
	)

	if err != nil {
		log.Printf("DEBUG: Query error: %v", err)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("DID not found")
		}
		return nil, fmt.Errorf("failed to get DID: %w", err)
	}

	log.Printf("DEBUG: Found DID: %+v", did)
	return &did, nil
}

// GetByUserID retrieves a DID by user ID
func (r *DIDRepository) GetByUserID(userID uuid.UUID) (*domain.DID, error) {
	query := `
		SELECT id, user_id, did, user_hash, public_key, status, created_at, updated_at, blockchain_tx
		FROM dids WHERE user_id = $1
	`

	var did domain.DID
	err := r.db.QueryRow(query, userID).Scan(
		&did.ID,
		&did.UserID,
		&did.Did,
		&did.UserHash,
		&did.PublicKey,
		&did.Status,
		&did.CreatedAt,
		&did.UpdatedAt,
		&did.BlockchainTx,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("DID not found")
		}
		return nil, fmt.Errorf("failed to get DID: %w", err)
	}

	return &did, nil
}

// GetByUserHash retrieves a DID by user hash
func (r *DIDRepository) GetByUserHash(userHash string) (*domain.DID, error) {
	query := `
		SELECT id, user_id, did, user_hash, public_key, status, created_at, updated_at, blockchain_tx
		FROM dids WHERE user_hash = $1
	`

	var did domain.DID
	err := r.db.QueryRow(query, userHash).Scan(
		&did.ID,
		&did.UserID,
		&did.Did,
		&did.UserHash,
		&did.PublicKey,
		&did.Status,
		&did.CreatedAt,
		&did.UpdatedAt,
		&did.BlockchainTx,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("DID not found")
		}
		return nil, fmt.Errorf("failed to get DID: %w", err)
	}

	return &did, nil
}

// Update updates a DID record
func (r *DIDRepository) Update(did *domain.DID) error {
	query := `
		UPDATE dids 
		SET user_id = $2, did = $3, user_hash = $4, public_key = $5, status = $6, updated_at = $7, blockchain_tx = $8
		WHERE id = $1
	`

	_, err := r.db.Exec(query,
		did.ID,
		did.UserID,
		did.Did,
		did.UserHash,
		did.PublicKey,
		did.Status,
		did.UpdatedAt,
		did.BlockchainTx,
	)

	if err != nil {
		return fmt.Errorf("failed to update DID: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of a DID
func (r *DIDRepository) UpdateStatus(id uuid.UUID, status string, txHash string) error {
	query := `
		UPDATE dids 
		SET status = $2, blockchain_tx = $3, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(query, id, status, txHash)
	if err != nil {
		return fmt.Errorf("failed to update DID status: %w", err)
	}

	return nil
}

// ListByStatus retrieves DIDs by status
func (r *DIDRepository) ListByStatus(status string) ([]*domain.DID, error) {
	query := `
		SELECT id, user_id, did, user_hash, public_key, status, created_at, updated_at, blockchain_tx
		FROM dids WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query DIDs: %w", err)
	}
	defer rows.Close()

	var dids []*domain.DID
	for rows.Next() {
		var did domain.DID
		err := rows.Scan(
			&did.ID,
			&did.UserID,
			&did.Did,
			&did.UserHash,
			&did.PublicKey,
			&did.Status,
			&did.CreatedAt,
			&did.UpdatedAt,
			&did.BlockchainTx,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan DID: %w", err)
		}
		dids = append(dids, &did)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return dids, nil
}

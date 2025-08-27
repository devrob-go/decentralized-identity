package models

import (
	"time"

	"github.com/google/uuid"
)

type UserToken struct {
	ID               uuid.UUID `db:"id" json:"id"`
	UserID           uuid.UUID `db:"user_id" json:"user_id"`
	AccessToken      string    `db:"access_token" json:"access_token"`
	RefreshToken     string    `db:"refresh_token" json:"refresh_token"`
	AccessExpiresAt  time.Time `db:"access_expires_at" json:"access_expires_at"`
	RefreshExpiresAt time.Time `db:"refresh_expires_at" json:"refresh_expires_at"`
	IsRevoked        bool      `db:"is_revoked" json:"is_revoked"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

package authentication

import (
	"auth-service/internal/repository"

	zlog "packages/logger"
)

// AuthService handles authentication operations
type AuthService struct {
	DB     *repository.DB
	logger *zlog.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(db *repository.DB, logger *zlog.Logger) *AuthService {
	return &AuthService{
		DB:     db,
		logger: logger,
	}
}

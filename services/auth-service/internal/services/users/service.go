package users

import (
	"auth-service/internal/repository"

	zlog "packages/logger"
)

// UserService handles user operations
type UserService struct {
	DB     *repository.DB
	logger *zlog.Logger
}

// NewUserService creates a new user service
func NewUserService(db *repository.DB, logger *zlog.Logger) *UserService {
	return &UserService{
		DB:     db,
		logger: logger,
	}
}

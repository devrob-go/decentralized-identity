package services

import (
	"auth-service/config"
	auth "auth-service/internal/services/auth"
	"auth-service/internal/services/users"
	"auth-service/internal/repository"

	zlog "packages/logger"
)

// Service encapsulates all business logic services
type Service struct {
	Config *config.Config
	DB     *repository.DB
	User   *users.UserService
	Auth   *auth.AuthService
}

// NewService creates a new service instance
func NewService(db *repository.DB, logger *zlog.Logger, cfg *config.Config) *Service {
	return &Service{
		Config: cfg,
		DB:     db,
		User:   users.NewUserService(db, logger),
		Auth:   auth.NewAuthService(db, logger),
	}
}

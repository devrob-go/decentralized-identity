package services

import (
	"auth-service/config"
	"auth-service/internal/clients"
	"auth-service/internal/repository"
	auth "auth-service/internal/services/auth"
	"auth-service/internal/services/users"
	"os"

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
	// Initialize DID client if URL is provided
	var didClient *clients.DIDClient
	didManagerURL := os.Getenv("DID_MANAGER_URL")
	if didManagerURL != "" {
		didClient = clients.NewDIDClient(didManagerURL)
		logger.Info(nil, "DID Manager client initialized", map[string]any{
			"url": didManagerURL,
		})
	} else {
		logger.Warn(nil, "DID_MANAGER_URL not set, DID integration disabled")
	}

	return &Service{
		Config: cfg,
		DB:     db,
		User:   users.NewUserService(db, logger),
		Auth:   auth.NewAuthService(db, logger, didClient),
	}
}

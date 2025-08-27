package authentication

import (
	"context"
	"errors"
	"net/http"
	"time"

	"auth-service/internal/clients"
	"auth-service/models"
	"auth-service/utils"
)

// SignUp registers a new user
func (s *AuthService) SignUp(ctx context.Context, req *models.UserCreateRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, _ := s.DB.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		err := errors.New("user already exists")
		s.logger.Error(ctx, err, "user already exists", http.StatusConflict, map[string]any{
			"email": req.Email,
		})
		return nil, err
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Error(ctx, err, "failed to hash password", http.StatusInternalServerError, nil)
		return nil, err
	}

	// Create user
	user := &models.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user, err = s.DB.CreateUser(ctx, user)
	if err != nil {
		s.logger.Error(ctx, err, "failed to create user", http.StatusInternalServerError, nil)
		return nil, err
	}

	// Create DID for the user if DID client is available
	if s.didClient != nil {
		didRequest := &clients.DIDCreateRequest{
			UserID:   user.ID.String(),
			Name:     user.Name,
			Email:    user.Email,
			Password: req.Password, // Use original password for DID hash
		}

		didResponse, err := s.didClient.CreateDID(didRequest)
		if err != nil {
			s.logger.Warn(ctx, "failed to create DID for user", map[string]any{
				"user_id": user.ID.String(),
				"error":   err.Error(),
			})
			// Don't fail user creation if DID creation fails
		} else {
			// Update user with DID information
			user.DID = didResponse.Data.DIDRecord.DID
			user.UserHash = didResponse.Data.UserHash

			s.logger.Info(ctx, "DID created successfully for user", map[string]any{
				"user_id": user.ID.String(),
				"did":     didResponse.Data.DIDRecord.DID,
				"status":  didResponse.Data.Status,
			})
		}
	}

	s.logger.Info(ctx, "user registered successfully", map[string]any{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"did":     user.DID,
	})
	return user, nil
}

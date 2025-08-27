package authentication

import (
	"context"
	"errors"
	"net/http"
	"time"

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

	s.logger.Info(ctx, "user registered successfully", map[string]any{
		"user_id": user.ID.String(),
		"email":   user.Email,
	})
	return user, nil
}

package authentication

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"auth-service/models"
	"auth-service/utils"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

// SignIn authenticates a user and returns tokens
func (s *AuthService) SignIn(ctx context.Context, credentials *models.Credentials, accessSecret, refreshSecret string) (*models.User, string, string, error) {
	if err := validation.ValidateStruct(credentials,
		validation.Field(&credentials.Email, validation.Required, is.Email),
		validation.Field(&credentials.Password, validation.Required, validation.Length(8, 60)),
	); err != nil {
		s.logger.Error(ctx, err, "validation error", http.StatusBadRequest)
		return nil, "", "", fmt.Errorf("validation error: %w", err)
	}

	// Get user by email
	user, err := s.DB.GetUserByEmail(ctx, credentials.Email)
	if err != nil {
		s.logger.Error(ctx, err, "failed to fetch user", http.StatusInternalServerError, nil)
		return nil, "", "", errors.New("invalid credentials")
	}

	// Verify password
	if !utils.CheckPasswordHash(credentials.Password, user.Password) {
		err := fmt.Errorf("invalid email or password")
		s.logger.Error(ctx, err, "password mismatch", http.StatusUnauthorized, nil)
		return nil, "", "", errors.New("invalid credentials")
	}

	// Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, user, accessSecret, refreshSecret)
	if err != nil {
		s.logger.Error(ctx, err, "failed to generate token", http.StatusInternalServerError, nil)
		return nil, "", "", err
	}

	s.logger.Info(ctx, "sign in successful", map[string]any{
		"user_id": user.ID.String(),
		"email":   user.Email,
	})
	return user, accessToken, refreshToken, nil
}

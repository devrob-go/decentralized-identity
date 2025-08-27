package authentication

import (
	"auth-service/models"
	"auth-service/utils"
	"context"
	"errors"
	"net/http"
	"time"
)

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string, accessSecret, refreshSecret string) (*models.UserToken, error) {
	if refreshToken == "" {
		err := errors.New("refresh token cannot be empty")
		s.logger.Error(ctx, err, "validation error", http.StatusBadRequest)
		return nil, err
	}

	// Validate the refresh token JWT
	claims, err := utils.ValidateToken(refreshToken, refreshSecret)
	if err != nil {
		s.logger.Error(ctx, err, "invalid refresh token", http.StatusUnauthorized)
		return nil, errors.New("invalid refresh token")
	}

	// Validate token type
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "refresh" {
		s.logger.Error(ctx, errors.New("invalid token type"), "invalid refresh token claims", http.StatusUnauthorized)
		return nil, errors.New("invalid refresh token")
	}

	// Get token from database
	token, err := s.DB.GetTokenByRefreshToken(ctx, refreshToken)
	if err != nil {
		s.logger.Error(ctx, err, "refresh token not found in database", http.StatusUnauthorized)
		return nil, errors.New("invalid refresh token")
	}

	// Check if token is revoked
	if token.IsRevoked {
		err := errors.New("refresh token revoked")
		s.logger.Error(ctx, err, "refresh token revoked", http.StatusUnauthorized, map[string]any{
			"token_id": token.ID.String(),
		})
		return nil, err
	}

	// Check if refresh token is expired
	if time.Now().After(token.RefreshExpiresAt) {
		err := errors.New("refresh token expired")
		s.logger.Error(ctx, err, "refresh token expired", http.StatusUnauthorized, map[string]any{
			"token_id": token.ID.String(),
		})
		return nil, err
	}

	// Get user
	user, err := s.DB.GetUserByID(ctx, token.UserID)
	if err != nil {
		s.logger.Error(ctx, err, "user not found for refresh token", http.StatusNotFound, map[string]any{
			"user_id": token.UserID.String(),
		})
		return nil, errors.New("user not found")
	}

	// Generate new access token
	now := time.Now()
	accessExpiresAt := now.Add(15 * time.Minute)

	newAccessToken, err := utils.GenerateAccessToken(user, accessSecret)
	if err != nil {
		s.logger.Error(ctx, err, "failed to generate new access token", http.StatusInternalServerError, map[string]any{
			"user_id": user.ID.String(),
		})
		return nil, err
	}

	// Update the token in database with new access token
	if err := s.DB.UpdateAccessToken(ctx, token.ID, newAccessToken, accessExpiresAt); err != nil {
		s.logger.Error(ctx, err, "failed to update access token", http.StatusInternalServerError, map[string]any{
			"token_id": token.ID.String(),
		})
		return nil, err
	}

	// Create updated token model
	updatedToken := &models.UserToken{
		ID:               token.ID,
		UserID:           token.UserID,
		AccessToken:      newAccessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: token.RefreshExpiresAt,
		IsRevoked:        false,
		CreatedAt:        token.CreatedAt,
	}

	s.logger.Info(ctx, "token refreshed successfully", map[string]any{
		"user_id":  user.ID.String(),
		"token_id": token.ID.String(),
	})

	return updatedToken, nil
}

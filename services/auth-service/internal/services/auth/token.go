package authentication

import (
	"context"
	"errors"
	"net/http"
	"time"

	"auth-service/models"
	"auth-service/utils"
)

// GenerateTokens creates access and refresh tokens for a user
func (s *AuthService) GenerateTokens(ctx context.Context, user *models.User, accessSecret, refreshSecret string) (string, string, error) {
	now := time.Now()
	accessExpiresAt := now.Add(15 * time.Minute)
	refreshExpiresAt := now.Add(7 * 24 * time.Hour)

	accessToken, err := utils.GenerateAccessToken(user, accessSecret)
	if err != nil {
		s.logger.Error(ctx, err, "failed to generate access token", http.StatusInternalServerError, map[string]any{
			"user_id": user.ID.String(),
		})
		return "", "", err
	}

	refreshToken, err := utils.GenerateRefreshToken(user, refreshSecret)
	if err != nil {
		s.logger.Error(ctx, err, "failed to generate refresh token", http.StatusInternalServerError, map[string]any{
			"user_id": user.ID.String(),
		})
		return "", "", err
	}

	if err := s.DB.StoreTokens(ctx, user.ID, accessToken, refreshToken, accessExpiresAt, refreshExpiresAt); err != nil {
		s.logger.Error(ctx, err, "failed to store tokens", http.StatusInternalServerError, map[string]any{
			"user_id": user.ID.String(),
		})
		return "", "", err
	}

	s.logger.Info(ctx, "tokens generated successfully", map[string]any{
		"user_id": user.ID.String(),
	})
	return accessToken, refreshToken, nil
}

// ValidateToken validates an access token
func (s *AuthService) ValidateToken(ctx context.Context, accessToken string, secret string) (*models.User, error) {
	// First validate the JWT token
	claims, err := utils.ValidateToken(accessToken, secret)
	if err != nil {
		s.logger.Error(ctx, err, "invalid JWT token", http.StatusUnauthorized, nil)
		return nil, errors.New("invalid token")
	}

	// Validate token type
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "access" {
		s.logger.Error(ctx, errors.New("invalid token type"), "invalid token claims", http.StatusUnauthorized, nil)
		return nil, errors.New("invalid token")
	}

	// Check if token exists in database and is not revoked
	token, err := s.DB.GetTokenByAccessToken(ctx, accessToken)
	if err != nil {
		s.logger.Error(ctx, err, "token not found in database", http.StatusUnauthorized, nil)
		return nil, errors.New("invalid token")
	}

	if token.IsRevoked {
		err := errors.New("token revoked")
		s.logger.Error(ctx, err, "token revoked", http.StatusUnauthorized, map[string]any{
			"token_id": token.ID.String(),
		})
		return nil, err
	}

	if time.Now().After(token.AccessExpiresAt) {
		err := errors.New("token expired")
		s.logger.Error(ctx, err, "token expired", http.StatusUnauthorized, map[string]any{
			"token_id": token.ID.String(),
		})
		return nil, err
	}

	user, err := s.DB.GetUserByID(ctx, token.UserID)
	if err != nil {
		s.logger.Error(ctx, err, "user not found for token", http.StatusNotFound, map[string]any{
			"user_id": token.UserID.String(),
		})
		return nil, errors.New("user not found")
	}

	s.logger.Info(ctx, "token validated successfully", map[string]any{
		"user_id": user.ID.String(),
	})
	return user, nil
}

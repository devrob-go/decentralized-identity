package authentication

import (
	"context"
	"errors"
	"net/http"

	authpkg "packages/auth"
)

// RevokeToken revokes an access token
func (s *AuthService) RevokeToken(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		err := errors.New("access token can not be empty")
		s.logger.Error(ctx, err, "validation error", http.StatusBadRequest)
		return err
	}

	// First revoke in database
	if err := s.DB.RevokeToken(ctx, accessToken); err != nil {
		s.logger.Error(ctx, err, "failed to revoke token", http.StatusInternalServerError, nil)
		return err
	}

	// Also revoke in memory for immediate effect
	authpkg.RevokeToken(accessToken)

	s.logger.Info(ctx, "token revoked successfully", map[string]any{
		"access_token": accessToken,
	})
	return nil
}

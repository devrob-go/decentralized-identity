package authentication

import (
	"context"
	"errors"
	"net/http"
)

// Signout revokes all tokens for a user
func (s *AuthService) Signout(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		err := errors.New("access token can not be empty")
		s.logger.Error(ctx, err, "validation error", http.StatusBadRequest)
		return err
	}
	if err := s.RevokeToken(ctx, accessToken); err != nil {
		s.logger.Error(ctx, err, "failed to signout user", http.StatusInternalServerError, nil)
		return err
	}

	s.logger.Info(ctx, "user signed out successfully", map[string]any{
		"access_token": accessToken,
	})
	return nil
}

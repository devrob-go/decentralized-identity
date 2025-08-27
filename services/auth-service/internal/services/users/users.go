package users

import (
	"context"
	"net/http"

	"auth-service/models"
)

// GetAllUsers retrieves all active users with pagination
func (s *UserService) GetAllUsers(ctx context.Context, page, limit int) ([]models.User, int, error) {
	// Calculate offset from page and limit
	offset := (page - 1) * limit
	if page <= 0 {
		offset = 0
	}

	users, err := s.DB.ListUsers(ctx, limit, offset)
	if err != nil {
		s.logger.Error(ctx, err, "failed to retrieve users", http.StatusInternalServerError, map[string]any{
			"page":   page,
			"limit":  limit,
			"offset": offset,
		})
		return nil, 0, err
	}

	// Get total count
	total, err := s.DB.CountUsers(ctx)
	if err != nil {
		s.logger.Error(ctx, err, "failed to count users", http.StatusInternalServerError, nil)
		return users, 0, err
	}

	s.logger.Info(ctx, "retrieved users successfully", map[string]any{
		"page":   page,
		"limit":  limit,
		"offset": offset,
		"count":  len(users),
		"total":  total,
	})
	return users, total, nil
}

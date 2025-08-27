package repository

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"auth-service/models"

	"github.com/google/uuid"
)

const (
	storeTokensQuery = `
		INSERT INTO user_tokens (
			user_id, 
			access_token, 
			refresh_token, 
			access_expires_at, 
			refresh_expires_at, 
			is_revoked
		) VALUES (
			:user_id,
			:access_token,
			:refresh_token,
			:access_expires_at,
			:refresh_expires_at,
			false
		)
	`

	revokeTokenQuery = `
		UPDATE user_tokens
		SET is_revoked = true
		WHERE access_token = :access_token
	`

	getTokenByAccessTokenQuery = `
		SELECT 
			id, 
			user_id, 
			access_token, 
			refresh_token, 
			access_expires_at, 
			refresh_expires_at, 
			is_revoked, 
			created_at
		FROM user_tokens
		WHERE access_token = :access_token
	`

	getTokenByRefreshTokenQuery = `
		SELECT 
			id, 
			user_id, 
			access_token, 
			refresh_token, 
			access_expires_at, 
			refresh_expires_at, 
			is_revoked, 
			created_at
		FROM user_tokens
		WHERE refresh_token = :refresh_token
	`

	updateAccessTokenQuery = `
		UPDATE user_tokens
		SET access_token = :access_token, access_expires_at = :access_expires_at
		WHERE id = :id
	`
)

// StoreTokens stores access and refresh tokens for a user
func (db *DB) StoreTokens(ctx context.Context, userID uuid.UUID, accessToken, refreshToken string, accessExpiresAt, refreshExpiresAt time.Time) error {
	params := map[string]any{
		"user_id":            userID,
		"access_token":       accessToken,
		"refresh_token":      refreshToken,
		"access_expires_at":  accessExpiresAt,
		"refresh_expires_at": refreshExpiresAt,
	}

	stmt, err := db.PrepareNamedContext(ctx, storeTokensQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare insert token failed", http.StatusInternalServerError)
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, params)
	if err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "insert token failed", status)
		return mappedErr
	}

	db.logger.Info(ctx, "tokens stored successfully", map[string]any{
		"user_id": userID,
	})

	return nil
}

// RevokeToken marks a token as revoked
func (db *DB) RevokeToken(ctx context.Context, accessToken string) error {
	params := map[string]any{
		"access_token": accessToken,
	}

	stmt, err := db.PrepareNamedContext(ctx, revokeTokenQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare revoke token failed", http.StatusInternalServerError)
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "revoke token failed", status)
		return mappedErr
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		db.logger.Error(ctx, err, "failed to get rows affected", http.StatusInternalServerError)
		return err
	}

	if rowsAffected == 0 {
		db.logger.Info(ctx, "token not found to revoke", map[string]any{
			"access_token": accessToken,
		})
		return errors.New("token not found")
	}

	db.logger.Info(ctx, "token revoked successfully", map[string]any{
		"access_token": accessToken,
	})

	return nil
}

// GetTokenByAccessToken retrieves a token by access token
func (db *DB) GetTokenByAccessToken(ctx context.Context, accessToken string) (*models.UserToken, error) {
	params := map[string]any{
		"access_token": accessToken,
	}

	var token models.UserToken
	stmt, err := db.PrepareNamedContext(ctx, getTokenByAccessTokenQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare select token failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &token, params); err != nil {
		if err == sql.ErrNoRows {
			db.logger.Info(ctx, "token not found", map[string]any{
				"access_token": accessToken,
			})
			return nil, errors.New("token not found")
		}
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "select token failed", status)
		return nil, mappedErr
	}

	return &token, nil
}

// GetTokenByRefreshToken retrieves a token by refresh token
func (db *DB) GetTokenByRefreshToken(ctx context.Context, refreshToken string) (*models.UserToken, error) {
	params := map[string]any{
		"refresh_token": refreshToken,
	}

	var token models.UserToken
	stmt, err := db.PrepareNamedContext(ctx, getTokenByRefreshTokenQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare select refresh token failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &token, params); err != nil {
		if err == sql.ErrNoRows {
			db.logger.Info(ctx, "refresh token not found", map[string]any{
				"refresh_token": refreshToken,
			})
			return nil, errors.New("refresh token not found")
		}
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "select refresh token failed", status)
		return nil, mappedErr
	}

	return &token, nil
}

// UpdateAccessToken updates the access token and its expiration time
func (db *DB) UpdateAccessToken(ctx context.Context, tokenID uuid.UUID, newAccessToken string, newExpiresAt time.Time) error {
	params := map[string]any{
		"id":                tokenID,
		"access_token":      newAccessToken,
		"access_expires_at": newExpiresAt,
	}

	stmt, err := db.PrepareNamedContext(ctx, updateAccessTokenQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare update access token failed", http.StatusInternalServerError)
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "update access token failed", status)
		return mappedErr
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		db.logger.Error(ctx, err, "failed to get rows affected", http.StatusInternalServerError)
		return err
	}

	if rowsAffected == 0 {
		db.logger.Info(ctx, "token not found to update", map[string]any{
			"token_id": tokenID,
		})
		return errors.New("token not found")
	}

	db.logger.Info(ctx, "access token updated successfully", map[string]any{
		"token_id": tokenID,
	})

	return nil
}

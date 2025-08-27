package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"auth-service/models"

	"github.com/google/uuid"
)

// Named queries
const (
	insertUserQuery = `
		INSERT INTO users (
			name,
			email,
			password,
			created_at,
			updated_at
		) VALUES (
			:name,
			:email,
			:password,
			:created_at,
			:updated_at
		)
		RETURNING id, name, email, created_at, updated_at
	`

	getUserByEmailQuery = `
		SELECT 
			id,
			name,
			email,
			password,
			created_at,
			updated_at
		FROM users
		WHERE email = :email
	`

	getUserByIDQuery = `
		SELECT 
			id,
			name,
			email,
			password,
			created_at,
			updated_at
		FROM users
		WHERE id = :id
	`

	listUsersQuery = `
		SELECT 
			id,
			name,
			email,
			created_at,
			updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset
	`

	countUsersQuery = `
		SELECT COUNT(*) FROM users
	`
)

// CreateUser inserts a new user into the database
func (db *DB) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if err := ValidateUserCreate(user); err != nil {
		db.logger.Error(ctx, err, "validation failed", http.StatusBadRequest)
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	stmt, err := db.PrepareNamedContext(ctx, insertUserQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare insert failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	var newUser models.User
	if err := stmt.GetContext(ctx, &newUser, user); err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "insert failed", status)
		return nil, mappedErr
	}

	db.logger.Info(ctx, "user created successfully", map[string]any{
		"user_id": newUser.ID,
		"email":   newUser.Email,
	})

	return &newUser, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	params := map[string]any{
		"email": email,
	}

	var user models.User
	stmt, err := db.PrepareNamedContext(ctx, getUserByEmailQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare select failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &user, params); err != nil {
		if err == sql.ErrNoRows {
			db.logger.Info(ctx, "user not found", map[string]any{
				"email": email,
			})
			return nil, errors.New("user not found")
		}
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "select failed", status)
		return nil, mappedErr
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	params := map[string]any{
		"id": id,
	}

	var user models.User
	stmt, err := db.PrepareNamedContext(ctx, getUserByIDQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare select failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &user, params); err != nil {
		if err == sql.ErrNoRows {
			db.logger.Info(ctx, "user not found", map[string]any{
				"id": id,
			})
			return nil, errors.New("user not found")
		}
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "select failed", status)
		return nil, mappedErr
	}

	return &user, nil
}

// ListUsers retrieves a list of users with pagination
func (db *DB) ListUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	params := map[string]any{
		"limit":  limit,
		"offset": offset,
	}

	var users []models.User
	stmt, err := db.PrepareNamedContext(ctx, listUsersQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare select failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &users, params); err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "select failed", status)
		return nil, mappedErr
	}

	db.logger.Info(ctx, "users retrieved successfully", map[string]any{
		"count":  len(users),
		"limit":  limit,
		"offset": offset,
	})

	return users, nil
}

// CountUsers returns the total number of users
func (db *DB) CountUsers(ctx context.Context) (int, error) {
	var count int
	stmt, err := db.PrepareNamedContext(ctx, countUsersQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare count failed", http.StatusInternalServerError)
		return 0, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &count, map[string]any{}); err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "count failed", status)
		return 0, mappedErr
	}

	return count, nil
}

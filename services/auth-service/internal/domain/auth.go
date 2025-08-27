package domain

import (
	"context"
)

// AuthResponse represents authentication response
type AuthResponse struct {
	User   *User      `json:"user"`
	Tokens *UserToken `json:"tokens"`
}

// TokenResponse represents token-only response
type TokenResponse struct {
	Tokens *UserToken `json:"tokens"`
}

// AuthService defines the interface for authentication business logic
type AuthService interface {
	SignUp(name, email, password string) (*AuthResponse, error)
	SignIn(email, password string) (*AuthResponse, error)
	SignOut(accessToken string) error
	RefreshToken(refreshToken string) (*TokenResponse, error)
	RevokeToken(accessToken string) error
	ValidateToken(ctx context.Context, accessToken string, secret string) (*User, error)
}

// UserService defines the interface for user management business logic
type UserService interface {
	ListUsers(page, limit int) ([]*User, int, error)
}

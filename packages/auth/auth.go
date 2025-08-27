// Since all the services will share the same middleware, instead of repeating same code we can define it here
// and use it in the main function to apply to all routes it is applicable to.
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"password,omitempty" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// role enumeration
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	// RoleSuperAdmin role is off limit except bootstrapping
	RoleSystemAdmin Role = "system_admin"
)

// get func for role
func (r Role) Get() string { return string(r) }

func LogError(ctx context.Context, err error, msg string, code int) {
	log.Printf("[ERROR] %s | code: %d | err: %v", msg, code, err)
}

func LogInfo(ctx context.Context, msg string, fields map[string]any) {
	log.Printf("[INFO] %s | context: %v", msg, fields)
}

// In-memory token storage for starter package
var (
	revokedTokens = make(map[string]bool)
	tokenMutex    sync.RWMutex
)

// AuthMiddleware validates JWT access token and injects user into context.
func AuthMiddleware(JWTAccessTokenSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		token, err := extractToken(c.GetHeader("Authorization"))
		if err != nil {
			LogError(ctx, err, "Failed to extract token", http.StatusUnauthorized)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Check if token is revoked in memory
		if isTokenRevoked(token) {
			LogError(ctx, nil, "Token is revoked", http.StatusUnauthorized)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is revoked"})
			c.Abort()
			return
		}

		user, err := parseUserFromToken(token, JWTAccessTokenSecret, ctx)
		if err != nil {
			LogError(ctx, err, "Failed to validate token", http.StatusUnauthorized)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		newCtx := context.WithValue(c.Request.Context(), ctxKeyUser{}, user)
		c.Request = c.Request.WithContext(newCtx)
		c.Set("user", user)
		c.Next()
	}
}

// SignoutMiddleware validates token format but allows expired tokens for signout
func SignoutMiddleware(JWTAccessTokenSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		token, err := extractToken(c.GetHeader("Authorization"))
		if err != nil {
			LogError(ctx, err, "Failed to extract token", http.StatusUnauthorized)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// For signout, we only check if the token format is valid
		// We don't validate if it's expired or revoked, as users should be able to sign out
		// even with expired tokens to clear their session

		// Check if token is revoked in memory (if it's a valid token that was previously revoked)
		if isTokenRevoked(token) {
			LogError(ctx, nil, "Token is revoked", http.StatusUnauthorized)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is revoked"})
			c.Abort()
			return
		}

		// Try to parse the token to get user info if possible, but don't fail if expired
		// We'll use a more lenient parsing approach for signout
		if user, err := parseUserFromTokenLenient(token, JWTAccessTokenSecret, ctx); err == nil && user != nil {
			// Token is valid, add user to context
			newCtx := context.WithValue(c.Request.Context(), ctxKeyUser{}, user)
			c.Request = c.Request.WithContext(newCtx)
			c.Set("user", user)
		}

		c.Next()
	}
}

// extractToken extracts Bearer token from Authorization header
func extractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid token format")
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

// parseUserFromToken extracts and validates user details from access token
func parseUserFromToken(tokenStr string, secret string, ctx context.Context) (*User, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		LogError(ctx, nil, "Invalid token claims", http.StatusUnauthorized)
		return nil, errors.New("invalid token claims")
	}

	if exp, ok := claims["exp"].(float64); !ok || int64(exp) < time.Now().Unix() {
		LogError(ctx, nil, "Token expired or missing expiration", http.StatusUnauthorized)
		return nil, errors.New("token expired or missing expiration")
	}

	user := &User{}
	var errMsg string
	if id, ok := claims["user_id"].(string); ok && id != "" {
		uid, err := uuid.Parse(id)
		if err != nil {
			errMsg = "invalid user ID format in token"
		} else {
			user.ID = uid
		}
	} else {
		errMsg = "missing or invalid user ID in token"
	}

	if errMsg != "" {
		LogError(ctx, nil, errMsg, http.StatusUnauthorized)
		return nil, errors.New(errMsg)
	}

	user.Name = claims["name"].(string)
	if user.Name == "" {
		user.Name = "Unknown User"
	}

	user.Email = claims["email"].(string)
	if user.Email == "" {
		user.Email = "Unknown Email"
	}

	LogInfo(ctx, "User authenticated successfully", map[string]any{"user_id": user.ID})
	return user, nil
}

// parseUserFromTokenLenient is a more lenient version of parseUserFromToken that allows expired tokens.
func parseUserFromTokenLenient(tokenStr string, secret string, ctx context.Context) (*User, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		LogError(ctx, nil, "Invalid token claims", http.StatusUnauthorized)
		return nil, errors.New("invalid token claims")
	}

	// Allow expired tokens for signout
	// if exp, ok := claims["exp"].(float64); !ok || int64(exp) < time.Now().Unix() {
	// 	LogError(ctx, nil, "Token expired or missing expiration", http.StatusUnauthorized)
	// 	return nil, errors.New("token expired or missing expiration")
	// }

	user := &User{}
	var errMsg string
	if id, ok := claims["user_id"].(string); ok && id != "" {
		uid, err := uuid.Parse(id)
		if err != nil {
			errMsg = "invalid user ID format in token"
		} else {
			user.ID = uid
		}
	} else {
		errMsg = "missing or invalid user ID in token"
	}

	if errMsg != "" {
		LogError(ctx, nil, errMsg, http.StatusUnauthorized)
		return nil, errors.New(errMsg)
	}

	user.Name = claims["name"].(string)
	if user.Name == "" {
		user.Name = "Unknown User"
	}

	user.Email = claims["email"].(string)
	if user.Email == "" {
		user.Email = "Unknown Email"
	}

	LogInfo(ctx, "User authenticated successfully (lenient)", map[string]any{"user_id": user.ID})
	return user, nil
}

// isTokenRevoked checks if a token is revoked in memory
func isTokenRevoked(token string) bool {
	tokenHash := HashToken(token)
	tokenMutex.RLock()
	defer tokenMutex.RUnlock()
	return revokedTokens[tokenHash]
}

// RevokeToken adds a token to the revoked tokens list
func RevokeToken(token string) {
	tokenHash := HashToken(token)
	tokenMutex.Lock()
	defer tokenMutex.Unlock()
	revokedTokens[tokenHash] = true
}

// ctxKeyUser is the type used for storing user in context to avoid key collisions.
type ctxKeyUser struct{}

// GetUserFromContext retrieves the authenticated user from the context.Context.
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user := ctx.Value(ctxKeyUser{})
	if user == nil {
		return nil, false
	}
	u, ok := user.(*User)
	return u, ok
}

// hashToken returns a SHA256 hash of a token
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

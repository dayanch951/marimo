package database

import (
	"errors"
	"time"

	"github.com/dayanch951/marimo/shared/models"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrTokenNotFound        = errors.New("refresh token not found")
	ErrTokenExpired         = errors.New("refresh token expired")
	ErrTokenRevoked         = errors.New("refresh token revoked")
)

// Database defines the interface for database operations
type Database interface {
	// User operations
	CreateUser(email, password, name, role string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	UpdateUser(id, name, email string) error
	AssignRole(userID, role string) error
	ValidatePassword(email, password string) (*models.User, error)
	ListUsers(page, limit int) ([]*models.User, int, error)

	// Refresh token operations
	CreateRefreshToken(userID, token string, expiresAt time.Time) (*models.RefreshToken, error)
	GetRefreshToken(token string) (*models.RefreshToken, error)
	RevokeRefreshToken(token string) error
	RevokeAllUserTokens(userID string) error
	CleanupExpiredTokens() error
}

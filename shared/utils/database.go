package utils

import (
	"errors"
	"sync"
	"time"

	"github.com/dayanch951/marimo/shared/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrTokenNotFound     = errors.New("refresh token not found")
	ErrTokenExpired      = errors.New("refresh token expired")
	ErrTokenRevoked      = errors.New("refresh token revoked")
)

// MemoryDB is a shared in-memory database
type MemoryDB struct {
	users         map[string]*models.User
	emails        map[string]string
	refreshTokens map[string]*models.RefreshToken
	mu            sync.RWMutex
}

// NewMemoryDB creates a new in-memory database
func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		users:         make(map[string]*models.User),
		emails:        make(map[string]string),
		refreshTokens: make(map[string]*models.RefreshToken),
	}
}

// CreateUser creates a new user
func (db *MemoryDB) CreateUser(email, password, name, role string) (*models.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.emails[email]; exists {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if role == "" {
		role = models.RoleUser
	}

	user := &models.User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		Password:  string(hashedPassword),
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	db.users[user.ID] = user
	db.emails[email] = user.ID

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (db *MemoryDB) GetUserByEmail(email string) (*models.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	userID, exists := db.emails[email]
	if !exists {
		return nil, ErrUserNotFound
	}

	user, exists := db.users[userID]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (db *MemoryDB) GetUserByID(id string) (*models.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, exists := db.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// UpdateUser updates user information
func (db *MemoryDB) UpdateUser(id, name, email string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	user, exists := db.users[id]
	if !exists {
		return ErrUserNotFound
	}

	if email != user.Email {
		if _, exists := db.emails[email]; exists {
			return ErrUserAlreadyExists
		}
		delete(db.emails, user.Email)
		db.emails[email] = id
		user.Email = email
	}

	user.Name = name
	user.UpdatedAt = time.Now()

	return nil
}

// AssignRole assigns a role to a user
func (db *MemoryDB) AssignRole(userID, role string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	user, exists := db.users[userID]
	if !exists {
		return ErrUserNotFound
	}

	user.Role = role
	user.UpdatedAt = time.Now()

	return nil
}

// ValidatePassword validates a user's password
func (db *MemoryDB) ValidatePassword(email, password string) (*models.User, error) {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

// ListUsers returns all users
func (db *MemoryDB) ListUsers(page, limit int) ([]*models.User, int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	users := make([]*models.User, 0, len(db.users))
	for _, user := range db.users {
		users = append(users, user)
	}

	total := len(users)
	start := (page - 1) * limit
	end := start + limit

	if start > total {
		return []*models.User{}, total, nil
	}

	if end > total {
		end = total
	}

	return users[start:end], total, nil
}

// CreateRefreshToken creates a new refresh token
func (db *MemoryDB) CreateRefreshToken(userID, token string, expiresAt time.Time) (*models.RefreshToken, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	db.refreshTokens[token] = refreshToken
	return refreshToken, nil
}

// GetRefreshToken retrieves a refresh token
func (db *MemoryDB) GetRefreshToken(token string) (*models.RefreshToken, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	refreshToken, exists := db.refreshTokens[token]
	if !exists {
		return nil, ErrTokenNotFound
	}

	// Check if token is expired
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Check if token is revoked
	if refreshToken.Revoked {
		return nil, ErrTokenRevoked
	}

	return refreshToken, nil
}

// RevokeRefreshToken revokes a specific refresh token
func (db *MemoryDB) RevokeRefreshToken(token string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	refreshToken, exists := db.refreshTokens[token]
	if !exists {
		return ErrTokenNotFound
	}

	refreshToken.Revoked = true
	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (db *MemoryDB) RevokeAllUserTokens(userID string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, token := range db.refreshTokens {
		if token.UserID == userID && !token.Revoked {
			token.Revoked = true
		}
	}

	return nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (db *MemoryDB) CleanupExpiredTokens() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	now := time.Now()
	for token, refreshToken := range db.refreshTokens {
		if now.After(refreshToken.ExpiresAt) {
			delete(db.refreshTokens, token)
		}
	}

	return nil
}

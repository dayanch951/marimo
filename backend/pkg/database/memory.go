package database

import (
	"errors"
	"sync"
	"time"

	"github.com/dayanch951/marimo/backend/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
)

// MemoryDB is an in-memory database for demonstration purposes
type MemoryDB struct {
	users map[string]*models.User // key: user ID
	emails map[string]string      // key: email, value: user ID
	mu    sync.RWMutex
}

// NewMemoryDB creates a new in-memory database
func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		users:  make(map[string]*models.User),
		emails: make(map[string]string),
	}
}

// CreateUser creates a new user
func (db *MemoryDB) CreateUser(email, password, name string) (*models.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check if user already exists
	if _, exists := db.emails[email]; exists {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		Password:  string(hashedPassword),
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

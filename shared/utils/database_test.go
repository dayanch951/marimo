package utils

import (
	"testing"
	"time"
)

func TestMemoryDB_CreateUser(t *testing.T) {
	db := NewMemoryDB()

	user, err := db.CreateUser("test@example.com", "password123", "Test User", "user")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Email = %v, want test@example.com", user.Email)
	}
	if user.Name != "Test User" {
		t.Errorf("Name = %v, want Test User", user.Name)
	}
	if user.Role != "user" {
		t.Errorf("Role = %v, want user", user.Role)
	}
	if user.ID == "" {
		t.Error("User ID is empty")
	}
	if user.Password == "password123" {
		t.Error("Password should be hashed")
	}
}

func TestMemoryDB_CreateUser_Duplicate(t *testing.T) {
	db := NewMemoryDB()

	email := "duplicate@example.com"
	_, err := db.CreateUser(email, "password123", "User 1", "user")
	if err != nil {
		t.Fatalf("First CreateUser() error = %v", err)
	}

	// Try to create duplicate
	_, err = db.CreateUser(email, "password456", "User 2", "admin")
	if err != ErrUserAlreadyExists {
		t.Errorf("CreateUser() error = %v, want %v", err, ErrUserAlreadyExists)
	}
}

func TestMemoryDB_GetUserByEmail(t *testing.T) {
	db := NewMemoryDB()

	email := "test@example.com"
	createdUser, _ := db.CreateUser(email, "password123", "Test User", "user")

	user, err := db.GetUserByEmail(email)
	if err != nil {
		t.Fatalf("GetUserByEmail() error = %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("User ID = %v, want %v", user.ID, createdUser.ID)
	}
	if user.Email != email {
		t.Errorf("Email = %v, want %v", user.Email, email)
	}
}

func TestMemoryDB_GetUserByEmail_NotFound(t *testing.T) {
	db := NewMemoryDB()

	_, err := db.GetUserByEmail("nonexistent@example.com")
	if err != ErrUserNotFound {
		t.Errorf("GetUserByEmail() error = %v, want %v", err, ErrUserNotFound)
	}
}

func TestMemoryDB_GetUserByID(t *testing.T) {
	db := NewMemoryDB()

	createdUser, _ := db.CreateUser("test@example.com", "password123", "Test User", "user")

	user, err := db.GetUserByID(createdUser.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("User ID = %v, want %v", user.ID, createdUser.ID)
	}
}

func TestMemoryDB_ValidatePassword(t *testing.T) {
	db := NewMemoryDB()

	email := "test@example.com"
	password := "password123"
	db.CreateUser(email, password, "Test User", "user")

	// Valid password
	user, err := db.ValidatePassword(email, password)
	if err != nil {
		t.Fatalf("ValidatePassword() error = %v", err)
	}
	if user.Email != email {
		t.Errorf("Email = %v, want %v", user.Email, email)
	}

	// Invalid password
	_, err = db.ValidatePassword(email, "wrongpassword")
	if err != ErrInvalidPassword {
		t.Errorf("ValidatePassword() error = %v, want %v", err, ErrInvalidPassword)
	}

	// Nonexistent user
	_, err = db.ValidatePassword("nonexistent@example.com", password)
	if err != ErrUserNotFound {
		t.Errorf("ValidatePassword() error = %v, want %v", err, ErrUserNotFound)
	}
}

func TestMemoryDB_UpdateUser(t *testing.T) {
	db := NewMemoryDB()

	user, _ := db.CreateUser("test@example.com", "password123", "Test User", "user")

	newName := "Updated Name"
	newEmail := "updated@example.com"

	err := db.UpdateUser(user.ID, newName, newEmail)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}

	updatedUser, _ := db.GetUserByID(user.ID)
	if updatedUser.Name != newName {
		t.Errorf("Name = %v, want %v", updatedUser.Name, newName)
	}
	if updatedUser.Email != newEmail {
		t.Errorf("Email = %v, want %v", updatedUser.Email, newEmail)
	}
}

func TestMemoryDB_AssignRole(t *testing.T) {
	db := NewMemoryDB()

	user, _ := db.CreateUser("test@example.com", "password123", "Test User", "user")

	err := db.AssignRole(user.ID, "admin")
	if err != nil {
		t.Fatalf("AssignRole() error = %v", err)
	}

	updatedUser, _ := db.GetUserByID(user.ID)
	if updatedUser.Role != "admin" {
		t.Errorf("Role = %v, want admin", updatedUser.Role)
	}
}

func TestMemoryDB_ListUsers(t *testing.T) {
	db := NewMemoryDB()

	// Create test users
	for i := 0; i < 15; i++ {
		db.CreateUser("user"+string(rune(i))+"@example.com", "password", "User", "user")
	}

	// Get first page
	users, total, err := db.ListUsers(1, 10)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	if total != 15 {
		t.Errorf("Total = %d, want 15", total)
	}
	if len(users) != 10 {
		t.Errorf("Users count = %d, want 10", len(users))
	}

	// Get second page
	users, total, err = db.ListUsers(2, 10)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	if len(users) != 5 {
		t.Errorf("Users count on page 2 = %d, want 5", len(users))
	}
}

func TestMemoryDB_RefreshToken_Create(t *testing.T) {
	db := NewMemoryDB()

	userID := "test-user-id"
	token := "test-refresh-token"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	rt, err := db.CreateRefreshToken(userID, token, expiresAt)
	if err != nil {
		t.Fatalf("CreateRefreshToken() error = %v", err)
	}

	if rt.UserID != userID {
		t.Errorf("UserID = %v, want %v", rt.UserID, userID)
	}
	if rt.Token != token {
		t.Errorf("Token = %v, want %v", rt.Token, token)
	}
	if rt.Revoked {
		t.Error("Token should not be revoked")
	}
}

func TestMemoryDB_RefreshToken_GetAndValidate(t *testing.T) {
	db := NewMemoryDB()

	token := "test-refresh-token"
	expiresAt := time.Now().Add(1 * time.Hour)

	db.CreateRefreshToken("user-id", token, expiresAt)

	// Get valid token
	rt, err := db.GetRefreshToken(token)
	if err != nil {
		t.Fatalf("GetRefreshToken() error = %v", err)
	}
	if rt.Token != token {
		t.Errorf("Token = %v, want %v", rt.Token, token)
	}

	// Get nonexistent token
	_, err = db.GetRefreshToken("nonexistent-token")
	if err != ErrTokenNotFound {
		t.Errorf("GetRefreshToken() error = %v, want %v", err, ErrTokenNotFound)
	}
}

func TestMemoryDB_RefreshToken_Expired(t *testing.T) {
	db := NewMemoryDB()

	token := "expired-token"
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired

	db.CreateRefreshToken("user-id", token, expiresAt)

	_, err := db.GetRefreshToken(token)
	if err != ErrTokenExpired {
		t.Errorf("GetRefreshToken() error = %v, want %v", err, ErrTokenExpired)
	}
}

func TestMemoryDB_RefreshToken_Revoke(t *testing.T) {
	db := NewMemoryDB()

	token := "test-token"
	expiresAt := time.Now().Add(1 * time.Hour)

	db.CreateRefreshToken("user-id", token, expiresAt)

	// Revoke token
	err := db.RevokeRefreshToken(token)
	if err != nil {
		t.Fatalf("RevokeRefreshToken() error = %v", err)
	}

	// Try to get revoked token
	_, err = db.GetRefreshToken(token)
	if err != ErrTokenRevoked {
		t.Errorf("GetRefreshToken() error = %v, want %v", err, ErrTokenRevoked)
	}
}

func TestMemoryDB_RefreshToken_RevokeAll(t *testing.T) {
	db := NewMemoryDB()

	userID := "test-user"
	expiresAt := time.Now().Add(1 * time.Hour)

	// Create multiple tokens for same user
	db.CreateRefreshToken(userID, "token1", expiresAt)
	db.CreateRefreshToken(userID, "token2", expiresAt)
	db.CreateRefreshToken("other-user", "token3", expiresAt)

	// Revoke all tokens for user
	err := db.RevokeAllUserTokens(userID)
	if err != nil {
		t.Fatalf("RevokeAllUserTokens() error = %v", err)
	}

	// Check that user's tokens are revoked
	_, err = db.GetRefreshToken("token1")
	if err != ErrTokenRevoked {
		t.Error("token1 should be revoked")
	}

	_, err = db.GetRefreshToken("token2")
	if err != ErrTokenRevoked {
		t.Error("token2 should be revoked")
	}

	// Other user's token should not be revoked
	_, err = db.GetRefreshToken("token3")
	if err != nil {
		t.Errorf("token3 should not be revoked, error = %v", err)
	}
}

func TestMemoryDB_RefreshToken_Cleanup(t *testing.T) {
	db := NewMemoryDB()

	// Create expired and valid tokens
	db.CreateRefreshToken("user1", "expired-token", time.Now().Add(-1*time.Hour))
	db.CreateRefreshToken("user2", "valid-token", time.Now().Add(1*time.Hour))

	// Cleanup
	err := db.CleanupExpiredTokens()
	if err != nil {
		t.Fatalf("CleanupExpiredTokens() error = %v", err)
	}

	// Expired token should not exist
	_, err = db.GetRefreshToken("expired-token")
	if err != ErrTokenNotFound {
		t.Error("Expired token should be deleted")
	}

	// Valid token should still exist
	_, err = db.GetRefreshToken("valid-token")
	if err != nil {
		t.Errorf("Valid token should exist, error = %v", err)
	}
}

func BenchmarkMemoryDB_CreateUser(b *testing.B) {
	db := NewMemoryDB()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.CreateUser("user"+string(rune(i))+"@example.com", "password", "User", "user")
	}
}

func BenchmarkMemoryDB_GetUserByEmail(b *testing.B) {
	db := NewMemoryDB()
	db.CreateUser("test@example.com", "password", "User", "user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.GetUserByEmail("test@example.com")
	}
}

func BenchmarkMemoryDB_ValidatePassword(b *testing.B) {
	db := NewMemoryDB()
	db.CreateUser("test@example.com", "password", "User", "user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.ValidatePassword("test@example.com", "password")
	}
}

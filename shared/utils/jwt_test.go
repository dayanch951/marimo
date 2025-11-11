package utils

import (
	"testing"
	"time"

	"github.com/dayanch951/marimo/shared/models"
)

func TestGenerateAccessToken(t *testing.T) {
	user := &models.User{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  "user",
	}

	token, err := GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if token == "" {
		t.Error("Generated token is empty")
	}

	// Validate the token
	claims, err := ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("UserID = %v, want %v", claims.UserID, user.ID)
	}
	if claims.Email != user.Email {
		t.Errorf("Email = %v, want %v", claims.Email, user.Email)
	}
	if claims.Role != user.Role {
		t.Errorf("Role = %v, want %v", claims.Role, user.Role)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token1, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if token1 == "" {
		t.Error("Generated refresh token is empty")
	}

	// Generate another token and ensure they're different
	token2, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if token1 == token2 {
		t.Error("Refresh tokens are not unique")
	}

	// Check length (base64 of 32 bytes should be around 43-44 chars)
	if len(token1) < 40 {
		t.Errorf("Refresh token too short: %d characters", len(token1))
	}
}

func TestGenerateTokenPair(t *testing.T) {
	user := &models.User{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  "admin",
	}

	pair, refreshToken, expiry, err := GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	// Check access token
	if pair.AccessToken == "" {
		t.Error("Access token is empty")
	}

	// Check refresh token
	if pair.RefreshToken == "" {
		t.Error("Refresh token is empty")
	}

	if refreshToken != pair.RefreshToken {
		t.Error("Returned refresh token doesn't match pair")
	}

	// Check token type
	if pair.TokenType != "Bearer" {
		t.Errorf("TokenType = %v, want Bearer", pair.TokenType)
	}

	// Check expires_in
	expectedExpiry := int64(AccessTokenDuration.Seconds())
	if pair.ExpiresIn != expectedExpiry {
		t.Errorf("ExpiresIn = %v, want %v", pair.ExpiresIn, expectedExpiry)
	}

	// Check expiry time
	expectedExpiryTime := time.Now().Add(RefreshTokenDuration)
	if expiry.Before(expectedExpiryTime.Add(-time.Minute)) || expiry.After(expectedExpiryTime.Add(time.Minute)) {
		t.Error("Refresh token expiry time is incorrect")
	}

	// Validate access token
	claims, err := ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("UserID = %v, want %v", claims.UserID, user.ID)
	}
}

func TestValidateAccessToken_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"Invalid format", "not.a.token"},
		{"Random string", "random-string-not-jwt"},
		{"Malformed JWT", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateAccessToken(tt.token)
			if err == nil {
				t.Error("ValidateAccessToken() should return error for invalid token")
			}
		})
	}
}

func TestAccessTokenExpiration(t *testing.T) {
	// This test would require manipulating time, which is complex
	// In a real scenario, you might use a time mocking library
	// For now, we'll just test that the expiration is set correctly
	user := &models.User{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  "user",
	}

	token, err := GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}

	// Check that expiration is in the future
	if time.Now().After(claims.ExpiresAt.Time) {
		t.Error("Token is already expired")
	}

	// Check that expiration is approximately AccessTokenDuration from now
	expectedExpiry := time.Now().Add(AccessTokenDuration)
	actualExpiry := claims.ExpiresAt.Time

	diff := actualExpiry.Sub(expectedExpiry)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Token expiration time is off by %v", diff)
	}
}

func BenchmarkGenerateAccessToken(b *testing.B) {
	user := &models.User{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  "user",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateAccessToken(user)
	}
}

func BenchmarkValidateAccessToken(b *testing.B) {
	user := &models.User{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  "user",
	}

	token, _ := GenerateAccessToken(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateAccessToken(token)
	}
}

func BenchmarkGenerateRefreshToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateRefreshToken()
	}
}

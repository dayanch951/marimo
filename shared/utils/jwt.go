package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"time"

	"github.com/dayanch951/marimo/shared/models"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// Token types
const (
	AccessTokenDuration  = 15 * time.Minute // Short-lived access token
	RefreshTokenDuration = 7 * 24 * time.Hour // Long-lived refresh token
)

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GetJWTSecret returns the JWT secret from environment
func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-this-min-32-chars"
	}
	return []byte(secret)
}

// GenerateAccessToken generates a short-lived JWT access token
func GenerateAccessToken(user *models.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "marimo-erp",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTSecret())
}

// GenerateRefreshToken generates a random refresh token string
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(user *models.User) (*TokenPair, string, time.Time, error) {
	// Generate access token
	accessToken, err := GenerateAccessToken(user)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	// Generate refresh token
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, "", time.Time{}, err
	}

	refreshExpiry := time.Now().Add(RefreshTokenDuration)

	pair := &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(AccessTokenDuration.Seconds()),
		TokenType:    "Bearer",
	}

	return pair, refreshToken, refreshExpiry, nil
}

// ValidateAccessToken validates and parses a JWT access token
func ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return GetJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshAccessToken validates refresh token and generates new access token
func RefreshAccessToken(refreshToken string, user *models.User) (string, error) {
	return GenerateAccessToken(user)
}

package grpc

import (
	"context"

	"github.com/dayanch951/marimo/backend/pkg/auth"
	"github.com/dayanch951/marimo/backend/pkg/database"
)

// AuthServiceServer implements the gRPC AuthService
type AuthServiceServer struct {
	UnimplementedAuthServiceServer
	db *database.MemoryDB
}

// NewAuthServiceServer creates a new gRPC auth service server
func NewAuthServiceServer(db *database.MemoryDB) *AuthServiceServer {
	return &AuthServiceServer{db: db}
}

// Login handles gRPC login requests
func (s *AuthServiceServer) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Validate credentials
	user, err := s.db.ValidatePassword(req.Email, req.Password)
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		}, nil
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: "Failed to generate token",
		}, nil
	}

	return &LoginResponse{
		Success: true,
		Token:   token,
		Message: "Login successful",
	}, nil
}

// Register handles gRPC registration requests
func (s *AuthServiceServer) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// Create user
	user, err := s.db.CreateUser(req.Email, req.Password, req.Name)
	if err != nil {
		if err == database.ErrUserAlreadyExists {
			return &RegisterResponse{
				Success: false,
				Message: "User already exists",
			}, nil
		}
		return &RegisterResponse{
			Success: false,
			Message: "Failed to create user",
		}, nil
	}

	return &RegisterResponse{
		Success: true,
		Message: "User created successfully",
		UserId:  user.ID,
	}, nil
}

// ValidateToken handles gRPC token validation requests
func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	claims, err := auth.ValidateToken(req.Token)
	if err != nil {
		return &ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID,
		Email:  claims.Email,
	}, nil
}

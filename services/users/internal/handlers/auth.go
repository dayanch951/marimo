package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dayanch951/marimo/shared/database"
	"github.com/dayanch951/marimo/shared/middleware"
	"github.com/dayanch951/marimo/shared/models"
	"github.com/dayanch951/marimo/shared/utils"
	"github.com/dayanch951/marimo/shared/validator"
)

type AuthHandler struct {
	db database.Database
}

func NewAuthHandler(db database.Database) *AuthHandler {
	return &AuthHandler{db: db}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type AuthResponse struct {
	Success      bool               `json:"success"`
	Message      string             `json:"message"`
	Token        string             `json:"token,omitempty"` // Deprecated: use TokenPair
	User         *models.User       `json:"user,omitempty"`
	AccessToken  string             `json:"access_token,omitempty"`
	RefreshToken string             `json:"refresh_token,omitempty"`
	ExpiresIn    int64              `json:"expires_in,omitempty"`
	TokenType    string             `json:"token_type,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate email
	if err := validator.ValidateEmail(req.Email); err != nil {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid email format",
		})
		return
	}

	// Validate password
	if err := validator.ValidatePassword(req.Password, validator.DefaultPasswordRequirements()); err != nil {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Validate name
	if err := validator.ValidateName(req.Name); err != nil {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid name format",
		})
		return
	}

	user, err := h.db.CreateUser(req.Email, req.Password, req.Name, models.RoleUser)
	if err != nil {
		if err == database.ErrUserAlreadyExists {
			respondJSON(w, http.StatusConflict, AuthResponse{
				Success: false,
				Message: "User already exists",
			})
			return
		}
		respondJSON(w, http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to create user",
		})
		return
	}

	respondJSON(w, http.StatusCreated, AuthResponse{
		Success: true,
		Message: "User created successfully",
		User:    user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate email
	if err := validator.ValidateEmail(req.Email); err != nil {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid email format",
		})
		return
	}

	// Basic password check (don't reveal requirements on login)
	if req.Password == "" || len(req.Password) > 128 {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid password",
		})
		return
	}

	user, err := h.db.ValidatePassword(req.Email, req.Password)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Generate token pair (access + refresh)
	tokenPair, refreshToken, refreshExpiry, err := utils.GenerateTokenPair(user)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to generate tokens",
		})
		return
	}

	// Store refresh token in database
	_, err = h.db.CreateRefreshToken(user.ID, refreshToken, refreshExpiry)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to store refresh token",
		})
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Success:      true,
		Message:      "Login successful",
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
	})
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	user, err := h.db.GetUserByID(claims.UserID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, total, err := h.db.ListUsers(1, 100)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to list users",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"users":   users,
		"total":   total,
	})
}

func (h *AuthHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	err := h.db.AssignRole(req.UserID, req.Role)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to assign role",
		})
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Success: true,
		Message: "Role assigned successfully",
	})
}

// RefreshToken refreshes an access token using a refresh token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.RefreshToken == "" {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Refresh token is required",
		})
		return
	}

	// Validate refresh token
	storedToken, err := h.db.GetRefreshToken(req.RefreshToken)
	if err != nil {
		if err == database.ErrTokenNotFound || err == database.ErrTokenExpired || err == database.ErrTokenRevoked {
			respondJSON(w, http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Invalid or expired refresh token",
			})
			return
		}
		respondJSON(w, http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to validate refresh token",
		})
		return
	}

	// Get user
	user, err := h.db.GetUserByID(storedToken.UserID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Generate new token pair
	tokenPair, newRefreshToken, refreshExpiry, err := utils.GenerateTokenPair(user)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to generate tokens",
		})
		return
	}

	// Revoke old refresh token
	if err := h.db.RevokeRefreshToken(req.RefreshToken); err != nil {
		// Log error but don't fail the request
	}

	// Store new refresh token
	_, err = h.db.CreateRefreshToken(user.ID, newRefreshToken, refreshExpiry)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to store refresh token",
		})
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Success:      true,
		Message:      "Token refreshed successfully",
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
	})
}

// Logout revokes a refresh token
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.RefreshToken != "" {
		// Revoke the specific refresh token
		if err := h.db.RevokeRefreshToken(req.RefreshToken); err != nil {
			// Log error but don't fail - token might already be revoked
		}
	}

	// Optionally revoke all user tokens if user is authenticated
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if ok {
		if err := h.db.RevokeAllUserTokens(claims.UserID); err != nil {
			// Log error but don't fail
		}
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

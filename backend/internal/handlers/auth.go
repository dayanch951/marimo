package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dayanch951/marimo/backend/internal/middleware"
	"github.com/dayanch951/marimo/backend/internal/models"
	"github.com/dayanch951/marimo/backend/pkg/auth"
	"github.com/dayanch951/marimo/backend/pkg/database"
)

type AuthHandler struct {
	db *database.MemoryDB
}

func NewAuthHandler(db *database.MemoryDB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, models.AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondWithJSON(w, http.StatusBadRequest, models.AuthResponse{
			Success: false,
			Message: "Email, password, and name are required",
		})
		return
	}

	// Create user
	user, err := h.db.CreateUser(req.Email, req.Password, req.Name)
	if err != nil {
		if err == database.ErrUserAlreadyExists {
			respondWithJSON(w, http.StatusConflict, models.AuthResponse{
				Success: false,
				Message: "User already exists",
			})
			return
		}
		respondWithJSON(w, http.StatusInternalServerError, models.AuthResponse{
			Success: false,
			Message: "Failed to create user",
		})
		return
	}

	respondWithJSON(w, http.StatusCreated, models.AuthResponse{
		Success: true,
		Message: "User created successfully",
		UserID:  user.ID,
	})
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, models.AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		respondWithJSON(w, http.StatusBadRequest, models.AuthResponse{
			Success: false,
			Message: "Email and password are required",
		})
		return
	}

	// Validate credentials
	user, err := h.db.ValidatePassword(req.Email, req.Password)
	if err != nil {
		respondWithJSON(w, http.StatusUnauthorized, models.AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, models.AuthResponse{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}

	respondWithJSON(w, http.StatusOK, models.AuthResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		UserID:  user.ID,
	})
}

// Profile returns the current user's profile (protected route)
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		respondWithJSON(w, http.StatusUnauthorized, models.AuthResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	user, err := h.db.GetUserByID(claims.UserID)
	if err != nil {
		respondWithJSON(w, http.StatusNotFound, models.AuthResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// Helper function to respond with JSON
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

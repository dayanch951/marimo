package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dayanch951/marimo/shared/middleware"
	"github.com/dayanch951/marimo/shared/models"
	"github.com/dayanch951/marimo/shared/utils"
)

type AuthHandler struct {
	db *utils.MemoryDB
}

func NewAuthHandler(db *utils.MemoryDB) *AuthHandler {
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
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Token   string       `json:"token,omitempty"`
	User    *models.User `json:"user,omitempty"`
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

	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Email, password, and name are required",
		})
		return
	}

	user, err := h.db.CreateUser(req.Email, req.Password, req.Name, models.RoleUser)
	if err != nil {
		if err == utils.ErrUserAlreadyExists {
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

	user, err := h.db.ValidatePassword(req.Email, req.Password)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		User:    user,
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

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

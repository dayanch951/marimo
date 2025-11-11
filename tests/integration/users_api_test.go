package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dayanch951/marimo/services/users/internal/handlers"
	"github.com/dayanch951/marimo/shared/middleware"
	"github.com/dayanch951/marimo/shared/models"
	"github.com/dayanch951/marimo/shared/utils"
	"github.com/gorilla/mux"
)

func setupTestServer() *httptest.Server {
	db := utils.NewMemoryDB()
	authHandler := handlers.NewAuthHandler(db)

	router := mux.NewRouter()
	router.HandleFunc("/api/users/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/users/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/users/refresh", authHandler.RefreshToken).Methods("POST")
	router.HandleFunc("/api/users/logout", authHandler.Logout).Methods("POST")

	protected := router.PathPrefix("/api/users").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/profile", authHandler.GetProfile).Methods("GET")
	protected.HandleFunc("/list", authHandler.ListUsers).Methods("GET")

	admin := router.PathPrefix("/api/users/admin").Subrouter()
	admin.Use(middleware.AuthMiddleware)
	admin.Use(middleware.RoleMiddleware(models.RoleAdmin))
	admin.HandleFunc("/assign-role", authHandler.AssignRole).Methods("POST")

	handler := middleware.CORS(router)
	return httptest.NewServer(handler)
}

func TestRegisterEndpoint(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Test successful registration
	payload := map[string]string{
		"email":    "test@example.com",
		"password": "TestPass123!",
		"name":     "Test User",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(server.URL+"/api/users/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if !result["success"].(bool) {
		t.Error("Registration should succeed")
	}
}

func TestRegisterEndpoint_InvalidEmail(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	payload := map[string]string{
		"email":    "invalid-email",
		"password": "TestPass123!",
		"name":     "Test User",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(server.URL+"/api/users/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRegisterEndpoint_WeakPassword(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	payload := map[string]string{
		"email":    "test@example.com",
		"password": "weak",
		"name":     "Test User",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(server.URL+"/api/users/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestLoginEndpoint(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Register a user first
	regPayload := map[string]string{
		"email":    "login@example.com",
		"password": "LoginPass123!",
		"name":     "Login User",
	}
	body, _ := json.Marshal(regPayload)
	http.Post(server.URL+"/api/users/register", "application/json", bytes.NewBuffer(body))

	// Test login
	loginPayload := map[string]string{
		"email":    "login@example.com",
		"password": "LoginPass123!",
	}
	body, _ = json.Marshal(loginPayload)
	resp, err := http.Post(server.URL+"/api/users/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["access_token"] == nil {
		t.Error("Should return access_token")
	}
	if result["refresh_token"] == nil {
		t.Error("Should return refresh_token")
	}
}

func TestLoginEndpoint_InvalidCredentials(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	payload := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "WrongPass123!",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(server.URL+"/api/users/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestProtectedEndpoint_NoAuth(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Try to access protected endpoint without token
	resp, err := http.Get(server.URL + "/api/users/profile")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestProtectedEndpoint_WithAuth(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Register and login
	regPayload := map[string]string{
		"email":    "auth@example.com",
		"password": "AuthPass123!",
		"name":     "Auth User",
	}
	body, _ := json.Marshal(regPayload)
	http.Post(server.URL+"/api/users/register", "application/json", bytes.NewBuffer(body))

	loginPayload := map[string]string{
		"email":    "auth@example.com",
		"password": "AuthPass123!",
	}
	body, _ = json.Marshal(loginPayload)
	loginResp, _ := http.Post(server.URL+"/api/users/login", "application/json", bytes.NewBuffer(body))

	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	loginResp.Body.Close()

	accessToken := loginResult["access_token"].(string)

	// Access protected endpoint
	req, _ := http.NewRequest("GET", server.URL+"/api/users/profile", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRefreshTokenEndpoint(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Register and login
	regPayload := map[string]string{
		"email":    "refresh@example.com",
		"password": "RefreshPass123!",
		"name":     "Refresh User",
	}
	body, _ := json.Marshal(regPayload)
	http.Post(server.URL+"/api/users/register", "application/json", bytes.NewBuffer(body))

	loginPayload := map[string]string{
		"email":    "refresh@example.com",
		"password": "RefreshPass123!",
	}
	body, _ = json.Marshal(loginPayload)
	loginResp, _ := http.Post(server.URL+"/api/users/login", "application/json", bytes.NewBuffer(body))

	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	loginResp.Body.Close()

	refreshToken := loginResult["refresh_token"].(string)

	// Refresh token
	refreshPayload := map[string]string{
		"refresh_token": refreshToken,
	}
	body, _ = json.Marshal(refreshPayload)
	resp, err := http.Post(server.URL+"/api/users/refresh", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["access_token"] == nil {
		t.Error("Should return new access_token")
	}
	if result["refresh_token"] == nil {
		t.Error("Should return new refresh_token")
	}
}

package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dayanch951/marimo/shared/middleware"
	"github.com/gorilla/mux"
)

const port = ":8086"

type DashboardStats struct {
	TotalUsers    int     `json:"total_users"`
	TotalOrders   int     `json:"total_orders"`
	TotalRevenue  float64 `json:"total_revenue"`
	ActiveProducts int    `json:"active_products"`
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Protected routes
	api := router.PathPrefix("/api/main").Subrouter()
	api.Use(middleware.AuthMiddleware)
	api.HandleFunc("/dashboard", getDashboard).Methods("GET")
	api.HandleFunc("/stats", getStats).Methods("GET")

	handler := middleware.CORS(router)

	log.Printf("Main service starting on port %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getDashboard(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)

	dashboard := map[string]interface{}{
		"welcome": "Welcome to Marimo ERP",
		"user": map[string]string{
			"id":    claims.UserID,
			"email": claims.Email,
			"role":  claims.Role,
		},
		"modules": []string{
			"users",
			"config",
			"accounting",
			"factory",
			"shop",
		},
	}

	respondJSON(w, http.StatusOK, dashboard)
}

func getStats(w http.ResponseWriter, r *http.Request) {
	// Mock stats - in production, this would aggregate from other services
	stats := DashboardStats{
		TotalUsers:     10,
		TotalOrders:    25,
		TotalRevenue:   1250.50,
		ActiveProducts: 15,
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Main Service OK"))
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

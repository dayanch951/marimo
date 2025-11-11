package main

import (
	"log"
	"net/http"

	"github.com/dayanch951/marimo/services/users/internal/handlers"
	"github.com/dayanch951/marimo/shared/middleware"
	"github.com/dayanch951/marimo/shared/models"
	"github.com/dayanch951/marimo/shared/utils"
	"github.com/gorilla/mux"
)

const port = ":8081"

func main() {
	// Initialize database
	db := utils.NewMemoryDB()

	// Create default admin user
	db.CreateUser("admin@example.com", "admin123", "Admin User", models.RoleAdmin)
	log.Println("Default admin user created: admin@example.com / admin123")

	// Create handlers
	authHandler := handlers.NewAuthHandler(db)

	// Create router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/api/users/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/users/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Protected routes
	protected := router.PathPrefix("/api/users").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/profile", authHandler.GetProfile).Methods("GET")
	protected.HandleFunc("/list", authHandler.ListUsers).Methods("GET")

	// Admin only routes
	admin := router.PathPrefix("/api/users/admin").Subrouter()
	admin.Use(middleware.AuthMiddleware)
	admin.Use(middleware.RoleMiddleware(models.RoleAdmin))
	admin.HandleFunc("/assign-role", authHandler.AssignRole).Methods("POST")

	// Apply CORS
	handler := middleware.CORS(router)

	log.Printf("Users service starting on port %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Users Service OK"))
}

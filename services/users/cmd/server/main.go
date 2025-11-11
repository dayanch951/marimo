package main

import (
	"context"
	"net/http"
	"time"

	"github.com/dayanch951/marimo/services/users/internal/handlers"
	"github.com/dayanch951/marimo/shared/logger"
	"github.com/dayanch951/marimo/shared/middleware"
	"github.com/dayanch951/marimo/shared/models"
	"github.com/dayanch951/marimo/shared/utils"
	"github.com/gorilla/mux"
)

const port = ":8081"

func main() {
	// Initialize logger
	log := logger.New("users-service")
	log.Info("Starting Users Service...")

	// Initialize database
	db := utils.NewMemoryDB()

	// Create default admin user
	_, err := db.CreateUser("admin@example.com", "admin123", "Admin User", models.RoleAdmin)
	if err != nil {
		log.Infof("Admin user already exists or error: %v", err)
	} else {
		log.Info("Default admin user created: admin@example.com / admin123")
	}

	// Create handlers
	authHandler := handlers.NewAuthHandler(db)

	// Create router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/api/users/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/users/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/health", healthCheck(log)).Methods("GET")

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

	// Create HTTP server
	server := &http.Server{
		Addr:         port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Users service listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Setup graceful shutdown
	utils.GracefulShutdown(server, 30*time.Second, func() {
		log.Info("Shutting down Users Service gracefully...")
	})

	log.Info("Users Service stopped")
}

func healthCheck(log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Health check requested")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Users Service OK"))
	}
}

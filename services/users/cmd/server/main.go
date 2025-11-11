package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/dayanch951/marimo/services/users/internal/handlers"
	"github.com/dayanch951/marimo/shared/database"
	"github.com/dayanch951/marimo/shared/logger"
	"github.com/dayanch951/marimo/shared/middleware"
	"github.com/dayanch951/marimo/shared/models"
	"github.com/dayanch951/marimo/shared/utils"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize logger
	log := logger.New("users-service")
	log.Info("Starting Users Service...")

	// Get configuration from environment
	port := getEnv("USERS_PORT", "8081")
	if port[0] != ':' {
		port = ":" + port
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "marimo_dev")
	dbSSLMode := getEnv("DB_SSL_MODE", "disable")
	usePostgres := getEnv("USE_POSTGRES", "false")

	// Initialize database
	var db database.Database
	var err error

	if usePostgres == "true" {
		log.Info("Initializing PostgreSQL database...")
		pgDB, err := database.NewPostgresDB(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		db = pgDB
		log.Info("PostgreSQL database connected successfully")

		// Cleanup on shutdown
		defer func() {
			if pgDB != nil {
				pgDB.Close()
				log.Info("PostgreSQL connection closed")
			}
		}()
	} else {
		log.Info("Initializing in-memory database...")
		db = utils.NewMemoryDB()
		log.Info("In-memory database initialized")
	}

	// Create default admin user
	_, err = db.CreateUser("admin@example.com", "admin123", "Admin User", models.RoleAdmin)
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

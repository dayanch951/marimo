package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dayanch951/marimo/backend/internal/grpc"
	"github.com/dayanch951/marimo/backend/internal/handlers"
	"github.com/dayanch951/marimo/backend/internal/middleware"
	"github.com/dayanch951/marimo/backend/pkg/database"
	"github.com/gorilla/mux"
	grpcLib "google.golang.org/grpc"
)

const (
	httpPort = "8080"
	grpcPort = "50051"
)

func main() {
	// Initialize database
	db := database.NewMemoryDB()
	log.Println("Database initialized")

	// Start gRPC server in a goroutine
	go startGRPCServer(db)

	// Start HTTP REST server
	startHTTPServer(db)
}

func startHTTPServer(db *database.MemoryDB) {
	// Create handlers
	authHandler := handlers.NewAuthHandler(db)

	// Create router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

	// Protected routes
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/profile", authHandler.Profile).Methods("GET")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Apply CORS middleware
	handler := middleware.CORS(router)

	// Start server
	log.Printf("HTTP REST server starting on port %s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, handler); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func startGRPCServer(db *database.MemoryDB) {
	// Create listener
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Create gRPC server
	grpcServer := grpcLib.NewServer()
	authService := grpc.NewAuthServiceServer(db)
	grpc.RegisterAuthServiceServer(grpcServer, authService)

	log.Printf("gRPC server starting on port %s", grpcPort)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}

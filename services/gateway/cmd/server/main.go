package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/dayanch951/marimo/shared/middleware"
	"github.com/gorilla/mux"
)

const port = ":8080"

// Service URLs - in production, use service discovery
var services = map[string]string{
	"users":      "http://users:8081",
	"config":     "http://config:8082",
	"accounting": "http://accounting:8083",
	"factory":    "http://factory:8084",
	"shop":       "http://shop:8085",
	"main":       "http://main:8086",
}

// For local development without Docker
var localServices = map[string]string{
	"users":      "http://localhost:8081",
	"config":     "http://localhost:8082",
	"accounting": "http://localhost:8083",
	"factory":    "http://localhost:8084",
	"shop":       "http://localhost:8085",
	"main":       "http://localhost:8086",
}

func main() {
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// API Gateway routes
	router.PathPrefix("/api/users").HandlerFunc(proxyHandler("users"))
	router.PathPrefix("/api/config").HandlerFunc(proxyHandler("config"))
	router.PathPrefix("/api/accounting").HandlerFunc(proxyHandler("accounting"))
	router.PathPrefix("/api/factory").HandlerFunc(proxyHandler("factory"))
	router.PathPrefix("/api/shop").HandlerFunc(proxyHandler("shop"))
	router.PathPrefix("/api/main").HandlerFunc(proxyHandler("main"))

	// Apply CORS
	handler := middleware.CORS(router)

	log.Printf("API Gateway starting on port %s", port)
	log.Println("Available services:")
	for name, url := range services {
		log.Printf("  - %s: %s", name, url)
	}

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}

func proxyHandler(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Try Docker service URL first, fall back to local
		serviceURL := services[serviceName]
		targetURL, err := url.Parse(serviceURL)
		if err != nil {
			// Try local service
			serviceURL = localServices[serviceName]
			targetURL, err = url.Parse(serviceURL)
			if err != nil {
				http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
				return
			}
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Update the request
		r.URL.Host = targetURL.Host
		r.URL.Scheme = targetURL.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = targetURL.Host

		// Log the request
		log.Printf("Proxying %s %s to %s", r.Method, r.URL.Path, serviceURL)

		// Proxy the request
		proxy.ServeHTTP(w, r)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	// Check all services
	statuses := make(map[string]string)
	allHealthy := true

	for name, serviceURL := range localServices {
		resp, err := http.Get(serviceURL + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			statuses[name] = "unhealthy"
			allHealthy = false
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			statuses[name] = strings.TrimSpace(string(body))
		}
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"gateway": "OK",
		"services": statuses,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health check response: %v", err)
	}
}

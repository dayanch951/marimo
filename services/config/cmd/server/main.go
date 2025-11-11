package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/dayanch951/marimo/shared/middleware"
	"github.com/gorilla/mux"
)

const port = ":8082"

type ConfigItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"` // system, user, app
}

var (
	configs = make(map[string]*ConfigItem)
	mu      sync.RWMutex
)

func main() {
	// Initialize default configs
	initDefaultConfigs()

	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Protected routes
	api := router.PathPrefix("/api/config").Subrouter()
	api.Use(middleware.AuthMiddleware)
	api.HandleFunc("", listConfigs).Methods("GET")
	api.HandleFunc("/{key}", getConfig).Methods("GET")
	api.HandleFunc("", setConfig).Methods("POST")
	api.HandleFunc("/{key}", deleteConfig).Methods("DELETE")

	handler := middleware.CORS(router)

	log.Printf("Config service starting on port %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDefaultConfigs() {
	configs["app_name"] = &ConfigItem{Key: "app_name", Value: "Marimo ERP", Type: "system"}
	configs["currency"] = &ConfigItem{Key: "currency", Value: "USD", Type: "system"}
	configs["timezone"] = &ConfigItem{Key: "timezone", Value: "UTC", Type: "system"}
	log.Println("Default configs initialized")
}

func listConfigs(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	items := make([]*ConfigItem, 0, len(configs))
	for _, item := range configs {
		items = append(items, item)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"configs": items,
	})
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	mu.RLock()
	item, exists := configs[key]
	mu.RUnlock()

	if !exists {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "Config not found",
		})
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func setConfig(w http.ResponseWriter, r *http.Request) {
	var item ConfigItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	mu.Lock()
	configs[item.Key] = &item
	mu.Unlock()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Config saved",
	})
}

func deleteConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	mu.Lock()
	delete(configs, key)
	mu.Unlock()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Config deleted",
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Config Service OK"))
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

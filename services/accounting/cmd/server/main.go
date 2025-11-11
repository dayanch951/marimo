package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dayanch951/marimo/shared/middleware"
	"github.com/dayanch951/marimo/shared/models"
	"github.com/gorilla/mux"
)

const port = ":8083"

type Transaction struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // income, expense
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

var (
	transactions = make(map[string]*Transaction)
	mu           sync.RWMutex
	counter      = 0
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Protected routes - accountant or admin only
	api := router.PathPrefix("/api/accounting").Subrouter()
	api.Use(middleware.AuthMiddleware)
	api.Use(middleware.RoleMiddleware(models.RoleAccountant, models.RoleAdmin))
	api.HandleFunc("/transactions", listTransactions).Methods("GET")
	api.HandleFunc("/transactions", createTransaction).Methods("POST")
	api.HandleFunc("/transactions/{id}", getTransaction).Methods("GET")
	api.HandleFunc("/balance", getBalance).Methods("GET")

	handler := middleware.CORS(router)

	log.Printf("Accounting service starting on port %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createTransaction(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)

	var tx Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	mu.Lock()
	counter++
	tx.ID = fmt.Sprintf("TXN-%d", counter)
	tx.CreatedBy = claims.UserID
	tx.CreatedAt = time.Now()
	transactions[tx.ID] = &tx
	mu.Unlock()

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success":     true,
		"message":     "Transaction created",
		"transaction": tx,
	})
}

func listTransactions(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	txList := make([]*Transaction, 0, len(transactions))
	for _, tx := range transactions {
		txList = append(txList, tx)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"transactions": txList,
	})
}

func getTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	mu.RLock()
	tx, exists := transactions[id]
	mu.RUnlock()

	if !exists {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "Transaction not found",
		})
		return
	}

	respondJSON(w, http.StatusOK, tx)
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	var income, expense float64
	for _, tx := range transactions {
		if tx.Type == "income" {
			income += tx.Amount
		} else if tx.Type == "expense" {
			expense += tx.Amount
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"balance": income - expense,
		"income":  income,
		"expense": expense,
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Accounting Service OK"))
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

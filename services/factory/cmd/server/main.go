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

const port = ":8084"

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	SKU         string    `json:"sku"`
	Quantity    int       `json:"quantity"`
	Status      string    `json:"status"` // in_production, completed, pending
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type ProductionOrder struct {
	ID         string    `json:"id"`
	ProductID  string    `json:"product_id"`
	Quantity   int       `json:"quantity"`
	Status     string    `json:"status"` // pending, in_progress, completed
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

var (
	products   = make(map[string]*Product)
	orders     = make(map[string]*ProductionOrder)
	mu         sync.RWMutex
	productCounter = 0
	orderCounter   = 0
)

func main() {
	initDefaultProducts()

	router := mux.NewRouter()

	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Protected routes
	api := router.PathPrefix("/api/factory").Subrouter()
	api.Use(middleware.AuthMiddleware)
	api.Use(middleware.RoleMiddleware(models.RoleManager, models.RoleAdmin))

	// Products
	api.HandleFunc("/products", listProducts).Methods("GET")
	api.HandleFunc("/products", createProduct).Methods("POST")
	api.HandleFunc("/products/{id}", getProduct).Methods("GET")
	api.HandleFunc("/products/{id}/status", updateProductStatus).Methods("PUT")

	// Production Orders
	api.HandleFunc("/orders", listOrders).Methods("GET")
	api.HandleFunc("/orders", createOrder).Methods("POST")
	api.HandleFunc("/orders/{id}", getOrder).Methods("GET")

	handler := middleware.CORS(router)

	log.Printf("Factory service starting on port %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDefaultProducts() {
	products["PROD-1"] = &Product{
		ID:        "PROD-1",
		Name:      "Widget A",
		SKU:       "WGT-A-001",
		Quantity:  100,
		Status:    "completed",
		CreatedBy: "system",
		CreatedAt: time.Now(),
	}
	log.Println("Default products initialized")
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)

	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	mu.Lock()
	productCounter++
	product.ID = fmt.Sprintf("PROD-%d", productCounter)
	product.CreatedBy = claims.UserID
	product.CreatedAt = time.Now()
	product.Status = "pending"
	products[product.ID] = &product
	mu.Unlock()

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Product created",
		"product": product,
	})
}

func listProducts(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	productList := make([]*Product, 0, len(products))
	for _, p := range products {
		productList = append(productList, p)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"products": productList,
	})
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	mu.RLock()
	product, exists := products[id]
	mu.RUnlock()

	if !exists {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "Product not found",
		})
		return
	}

	respondJSON(w, http.StatusOK, product)
}

func updateProductStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	mu.Lock()
	product, exists := products[id]
	if !exists {
		mu.Unlock()
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "Product not found",
		})
		return
	}

	product.Status = req.Status
	if req.Status == "completed" {
		now := time.Now()
		product.CompletedAt = &now
	}
	mu.Unlock()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Product status updated",
	})
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)

	var order ProductionOrder
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	mu.Lock()
	orderCounter++
	order.ID = fmt.Sprintf("ORD-%d", orderCounter)
	order.CreatedBy = claims.UserID
	order.CreatedAt = time.Now()
	order.Status = "pending"
	orders[order.ID] = &order
	mu.Unlock()

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Order created",
		"order":   order,
	})
}

func listOrders(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	orderList := make([]*ProductionOrder, 0, len(orders))
	for _, o := range orders {
		orderList = append(orderList, o)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"orders":  orderList,
	})
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	mu.RLock()
	order, exists := orders[id]
	mu.RUnlock()

	if !exists {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "Order not found",
		})
		return
	}

	respondJSON(w, http.StatusOK, order)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Factory Service OK"))
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

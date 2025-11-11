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

const port = ":8085"

type ShopProduct struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
	ImageURL    string  `json:"image_url"`
}

type Order struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Items      []OrderItem `json:"items"`
	Total      float64   `json:"total"`
	Status     string    `json:"status"` // pending, processing, shipped, delivered
	CreatedAt  time.Time `json:"created_at"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

var (
	shopProducts = make(map[string]*ShopProduct)
	orders       = make(map[string]*Order)
	mu           sync.RWMutex
	orderCounter = 0
)

func main() {
	initDefaultProducts()

	router := mux.NewRouter()

	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Public routes
	router.HandleFunc("/api/shop/products", listProducts).Methods("GET")
	router.HandleFunc("/api/shop/products/{id}", getProduct).Methods("GET")

	// Protected routes
	protected := router.PathPrefix("/api/shop").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/orders", createOrder).Methods("POST")
	protected.HandleFunc("/orders", listUserOrders).Methods("GET")
	protected.HandleFunc("/orders/{id}", getOrder).Methods("GET")

	// Admin routes
	admin := router.PathPrefix("/api/shop/admin").Subrouter()
	admin.Use(middleware.AuthMiddleware)
	admin.Use(middleware.RoleMiddleware(models.RoleAdmin, models.RoleShopManager))
	admin.HandleFunc("/products", createProduct).Methods("POST")
	admin.HandleFunc("/products/{id}", updateProduct).Methods("PUT")
	admin.HandleFunc("/products/{id}", deleteProduct).Methods("DELETE")
	admin.HandleFunc("/orders", listAllOrders).Methods("GET")

	handler := middleware.CORS(router)

	log.Printf("Shop service starting on port %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDefaultProducts() {
	shopProducts["SHOP-1"] = &ShopProduct{
		ID:          "SHOP-1",
		Name:        "Premium Widget",
		Description: "High quality widget for all your needs",
		Price:       29.99,
		Stock:       50,
		Category:    "Electronics",
		ImageURL:    "/images/widget.jpg",
	}
	shopProducts["SHOP-2"] = &ShopProduct{
		ID:          "SHOP-2",
		Name:        "Deluxe Gadget",
		Description: "Amazing gadget with advanced features",
		Price:       49.99,
		Stock:       30,
		Category:    "Electronics",
		ImageURL:    "/images/gadget.jpg",
	}
	log.Println("Default shop products initialized")
}

func listProducts(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	products := make([]*ShopProduct, 0, len(shopProducts))
	for _, p := range shopProducts {
		products = append(products, p)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"products": products,
	})
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	mu.RLock()
	product, exists := shopProducts[id]
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

func createProduct(w http.ResponseWriter, r *http.Request) {
	var product ShopProduct
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	mu.Lock()
	if product.ID == "" {
		product.ID = fmt.Sprintf("SHOP-%d", len(shopProducts)+1)
	}
	shopProducts[product.ID] = &product
	mu.Unlock()

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Product created",
		"product": product,
	})
}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var updates ShopProduct
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	mu.Lock()
	product, exists := shopProducts[id]
	if !exists {
		mu.Unlock()
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "Product not found",
		})
		return
	}

	product.Name = updates.Name
	product.Description = updates.Description
	product.Price = updates.Price
	product.Stock = updates.Stock
	product.Category = updates.Category
	mu.Unlock()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Product updated",
	})
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	mu.Lock()
	delete(shopProducts, id)
	mu.Unlock()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Product deleted",
	})
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	mu.Lock()
	orderCounter++
	order.ID = fmt.Sprintf("ORDER-%d", orderCounter)
	order.UserID = claims.UserID
	order.CreatedAt = time.Now()
	order.Status = "pending"

	// Calculate total
	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}
	order.Total = total

	orders[order.ID] = &order
	mu.Unlock()

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Order created",
		"order":   order,
	})
}

func listUserOrders(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)

	mu.RLock()
	defer mu.RUnlock()

	userOrders := make([]*Order, 0)
	for _, order := range orders {
		if order.UserID == claims.UserID {
			userOrders = append(userOrders, order)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"orders":  userOrders,
	})
}

func listAllOrders(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	allOrders := make([]*Order, 0, len(orders))
	for _, order := range orders {
		allOrders = append(allOrders, order)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"orders":  allOrders,
	})
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
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

	// Check if user owns the order or is admin
	if order.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		respondJSON(w, http.StatusForbidden, map[string]interface{}{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	respondJSON(w, http.StatusOK, order)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Shop Service OK"))
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

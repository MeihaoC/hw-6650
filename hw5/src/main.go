package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

// Product represents a product in our store (matches OpenAPI schema)
type Product struct {
	ProductID    int32  `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int32  `json:"category_id"`
	Weight       int32  `json:"weight"`
	SomeOtherID  int32  `json:"some_other_id"`
}

// Error represents an error response (matches OpenAPI schema)
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ProductStore holds our in-memory product data
type ProductStore struct {
	mu       sync.RWMutex
	products map[int32]Product
}

// Global store instance
var store = &ProductStore{
	products: make(map[int32]Product),
}

// GetProduct handles GET /products/{productId}
func GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productIDStr := vars["productId"]

	// Parse and validate productId
	productID64, err := strconv.ParseInt(productIDStr, 10, 32)
	if err != nil || productID64 < 1 {
		sendError(w, http.StatusBadRequest, "INVALID_INPUT", "Product ID must be a positive integer", "")
		return
	}
	productID := int32(productID64)

	// Retrieve product from store
	store.mu.RLock()
	product, exists := store.products[productID]
	store.mu.RUnlock()

	if !exists {
		sendError(w, http.StatusNotFound, "NOT_FOUND", "Product not found", "")
		return
	}

	// Return 200 OK with product
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

// AddProductDetails handles POST /products/{productId}/details
func AddProductDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productIDStr := vars["productId"]

	// Parse and validate productId from path
	productID64, err := strconv.ParseInt(productIDStr, 10, 32)
	if err != nil || productID64 < 1 {
		sendError(w, http.StatusBadRequest, "INVALID_INPUT", "Product ID must be a positive integer", "")
		return
	}
	productID := int32(productID64)

	// Decode request body
	var product Product
	err = json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		sendError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON format", err.Error())
		return
	}

	// Validate required fields
	if err := validateProduct(product); err != nil {
		sendError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid product data", err.Error())
		return
	}

	// Ensure the product_id in body matches the path parameter
	if product.ProductID != productID {
		sendError(w, http.StatusBadRequest, "INVALID_INPUT", "Product ID in body must match path parameter", "")
		return
	}

	// Store the product (upsert behavior: add or update)
	// Note: Strictly following the spec would require checking if product exists first
	// and returning 404 if not found. For this assignment, we allow creation.
	store.mu.Lock()
	store.products[productID] = product
	store.mu.Unlock()

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// validateProduct checks if all required fields are present and valid
func validateProduct(p Product) error {
	if p.ProductID < 1 {
		return &ValidationError{"product_id must be at least 1"}
	}
	if p.SKU == "" || len(p.SKU) > 100 {
		return &ValidationError{"sku is required and must be 1-100 characters"}
	}
	if p.Manufacturer == "" || len(p.Manufacturer) > 200 {
		return &ValidationError{"manufacturer is required and must be 1-200 characters"}
	}
	if p.CategoryID < 1 {
		return &ValidationError{"category_id must be at least 1"}
	}
	if p.Weight < 0 {
		return &ValidationError{"weight must be 0 or greater"}
	}
	if p.SomeOtherID < 1 {
		return &ValidationError{"some_other_id must be at least 1"}
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string {
	return e.msg
}

// sendError sends an error response in the format specified by OpenAPI
func sendError(w http.ResponseWriter, statusCode int, errorCode string, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorCode,
		Message: message,
		Details: details,
	})
}

// HealthCheck handles GET /health for container health checks
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	router := mux.NewRouter()

	// Product API routes (matching OpenAPI spec)
	router.HandleFunc("/products/{productId}", GetProduct).Methods("GET")
	router.HandleFunc("/products/{productId}/details", AddProductDetails).Methods("POST")

	// Health check endpoint
	router.HandleFunc("/health", HealthCheck).Methods("GET")

	// Start server
	port := "8080"
	log.Printf("Product API Server starting on port %s...", port)
	log.Printf("Endpoints available:")
	log.Printf("  GET  /products/{productId}")
	log.Printf("  POST /products/{productId}/details")
	log.Printf("  GET  /health")
	log.Fatal(http.ListenAndServe(":"+port, router))
}

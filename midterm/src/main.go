package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

// Product structure
type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
}

// SearchResponse structure
type SearchResponse struct {
	Products   []Product `json:"products"`
	TotalFound int       `json:"total_found"`
	SearchTime string    `json:"search_time"`
}

// Global product storage using sync.Map
var productStore sync.Map

// Sample data for generation
var brands = []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta"}
var categories = []string{"Electronics", "Books", "Home", "Clothing", "Sports", "Toys", "Garden", "Automotive"}
var descriptions = []string{
	"High quality product",
	"Best seller item",
	"Premium choice",
	"Customer favorite",
	"Top rated product",
}

func main() {
	log.Println("Initializing product data...")
	initProducts()
	log.Println("Successfully loaded 100,000 products")

	// Start server
	log.Println("Starting server on port 8080...")
	if err := fasthttp.ListenAndServe(":8080", requestHandler); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

// Initialize 100,000 products at startup
func initProducts() {
	for i := 1; i <= 100000; i++ {
		product := Product{
			ID:          i,
			Name:        fmt.Sprintf("Product %s %d", brands[i%len(brands)], i), // Use modulo to rotate through brands. e.g., "Product Alpha 1"
			Category:    categories[i%len(categories)],                          // Rotate through categories
			Description: descriptions[i%len(descriptions)],                      // Rotate through descriptions
			Brand:       brands[i%len(brands)],                                  // Rotate through brands
		}
		productStore.Store(i, product)
	}
}

// Main request handler
func requestHandler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path()) // Get path as bytes and convert to string

	switch path {
	case "/products/search":
		searchProducts(ctx) // Handle search requests
	case "/health":
		ctx.SetStatusCode(fasthttp.StatusOK)
	default:
		ctx.Error("Not Found", fasthttp.StatusNotFound) // 404 for other paths
	}
}

// Search products endpoint - checks exactly 100 products
func searchProducts(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()

	// Get query parameter
	query := strings.ToLower(string(ctx.QueryArgs().Peek("q"))) // Convert to string and lowercase for case-insensitive search
	if query == "" {
		ctx.Error("Query parameter 'q' is required", fasthttp.StatusBadRequest) // 400 if 'q' is missing
		return
	}

	// Initialize variables
	var matches []Product // Slice to hold matching products
	checked := 0          // Counter for how many products we've examined
	maxCheck := 100       // Stop after checking 100 products
	maxResults := 20      // Return max 20 products in response

	// Iterate through productStore and check exactly 100 products
	// sync.Map doesn't allow regular for loops, so we use .Range(). It iterates over all key-value pairs.
	productStore.Range(func(key, value interface{}) bool {
		if checked >= maxCheck {
			return false // Stop iteration after checking 100 products
		}

		checked++
		product := value.(Product) // Type assertion: value is of type interface{}, so we assert it to Product

		// Search in name and category (case-insensitive)
		if strings.Contains(strings.ToLower(product.Name), query) ||
			strings.Contains(strings.ToLower(product.Category), query) {
			if len(matches) < maxResults {
				matches = append(matches, product) // Add product to matches slice
			}
		}

		return true // Continue iteration
	})

	// Calculate search time
	searchTime := time.Since(startTime)

	// Prepare response
	response := SearchResponse{
		Products:   matches,
		TotalFound: len(matches),
		SearchTime: fmt.Sprintf("%.3fms", float64(searchTime.Microseconds())/1000.0),
	}

	// Send JSON response
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetStatusCode(fasthttp.StatusOK) // 200 OK

	jsonData, err := json.Marshal(response) // Convert response struct to JSON
	// Example output: {"products":[{"id":1,"name":"Product Alpha 1","category":"Electronics","description":"High quality product","brand":"Alpha"},...],"total_found":20,"search_time":"1.234ms"}
	if err != nil {
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetBody(jsonData) // Sends the JSON back to the client
}

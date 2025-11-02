package main

import (
	"log"
	"os"

	"shopping-cart-service/database"
	"shopping-cart-service/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 Starting Shopping Cart Service...")

	// Initialize database connection
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Initialize database schema (creates tables if they don't exist)
	if err := database.InitSchema(); err != nil {
		log.Fatalf("❌ Failed to initialize schema: %v", err)
	}

	// Set up Gin router
	gin.SetMode(gin.ReleaseMode) // Use release mode for production
	r := gin.Default()

	// Health check endpoint (required for ALB health checks)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "shopping-cart-service",
		})
	})

	// Shopping cart endpoints (per HW8 OpenAPI spec)
	r.POST("/shopping-carts", handlers.CreateCart)
	r.GET("/shopping-carts/:shoppingCartId", handlers.GetCart)
	r.POST("/shopping-carts/:shoppingCartId/items", handlers.AddItemToCart)

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("✅ Server starting on port %s", port)
	log.Println("📍 Endpoints:")
	log.Println("   GET  /health")
	log.Println("   POST /shopping-carts")
	log.Println("   GET  /shopping-carts/:id")
	log.Println("   POST /shopping-carts/:id/items")

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

package main

import (
	"log"
	"os"

	"shopping-cart-service/database"
	"shopping-cart-service/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting Shopping Cart Service...")

	// Check which database to use
	useDynamoDB := os.Getenv("USE_DYNAMODB") == "true"

	if useDynamoDB {
		log.Println("Using DynamoDB")
		if err := database.InitDynamoDB(); err != nil {
			log.Fatalf("Failed to initialize DynamoDB: %v", err)
		}
	} else {
		log.Println("Using MySQL")
		if err := database.InitDB(); err != nil {
			log.Fatalf("Failed to initialize MySQL: %v", err)
		}
		defer database.CloseDB()

		if err := database.InitSchema(); err != nil {
			log.Fatalf("Failed to initialize schema: %v", err)
		}
	}

	// Set up Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		dbType := "MySQL"
		if useDynamoDB {
			dbType = "DynamoDB"
		}
		c.JSON(200, gin.H{
			"status":   "healthy",
			"service":  "shopping-cart-service",
			"database": dbType,
		})
	})

	// Route to appropriate handlers based on database
	if useDynamoDB {
		r.POST("/shopping-carts", handlers.CreateCartDynamo)
		r.GET("/shopping-carts/:shoppingCartId", handlers.GetCartDynamo)
		r.POST("/shopping-carts/:shoppingCartId/items", handlers.AddItemToCartDynamo)
	} else {
		r.POST("/shopping-carts", handlers.CreateCart)
		r.GET("/shopping-carts/:shoppingCartId", handlers.GetCart)
		r.POST("/shopping-carts/:shoppingCartId/items", handlers.AddItemToCart)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

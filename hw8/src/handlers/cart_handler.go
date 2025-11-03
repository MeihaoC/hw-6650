package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"shopping-cart-service/database"
	"shopping-cart-service/models"

	"github.com/gin-gonic/gin"
)

// CreateCart creates a new shopping cart
// POST /shopping-carts
func CreateCart(c *gin.Context) {
	var req models.CreateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Insert into shopping_carts table
	result, err := database.DB.Exec(
		"INSERT INTO shopping_carts (customer_id, created_at, updated_at) VALUES (?, ?, ?)",
		req.CustomerID, time.Now(), time.Now(),
	)
	if err != nil {
		log.Printf("Error creating cart: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to create shopping cart",
		})
		return
	}

	cartID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting cart ID: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to get cart ID",
		})
		return
	}

	log.Printf("Created cart %d for customer %d", cartID, req.CustomerID)
	c.JSON(http.StatusCreated, models.CreateCartResponse{
		ShoppingCartID: strconv.FormatInt(cartID, 10),
	})
}

// GetCart retrieves a shopping cart with all items using efficient JOIN
// GET /shopping-carts/{shoppingCartId}
func GetCart(c *gin.Context) {
	cartIDStr := c.Param("shoppingCartId")
	cartID, err := strconv.Atoi(cartIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid cart ID",
		})
		return
	}

	// Get cart with items using single LEFT JOIN (per HW8 requirements)
	query := `
		SELECT 
			sc.cart_id, sc.customer_id, sc.created_at, sc.updated_at,
			ci.product_id, ci.quantity, ci.added_at, ci.updated_at
		FROM shopping_carts sc
		LEFT JOIN cart_items ci ON sc.cart_id = ci.cart_id
		WHERE sc.cart_id = ?
		ORDER BY ci.product_id
	`

	rows, err := database.DB.Query(query, cartID)
	if err != nil {
		log.Printf("Error retrieving cart: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to retrieve cart",
		})
		return
	}
	defer rows.Close()

	var cart models.ShoppingCart
	cart.Items = []models.CartItem{}
	cartFound := false

	for rows.Next() {
		var productID sql.NullInt64
		var quantity sql.NullInt64
		var addedAt sql.NullTime
		var updatedAt sql.NullTime

		err := rows.Scan(
			&cart.CartID, &cart.CustomerID, &cart.CreatedAt, &cart.UpdatedAt,
			&productID, &quantity, &addedAt, &updatedAt,
		)
		if err != nil {
			log.Printf("Error scanning cart row: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "DATABASE_ERROR",
				Message: "Failed to scan cart data",
			})
			return
		}

		cartFound = true

		// If cart has items (productID is not null), add them
		if productID.Valid && quantity.Valid {
			item := models.CartItem{
				ProductID: int(productID.Int64),
				Quantity:  int(quantity.Int64),
			}
			if addedAt.Valid {
				item.AddedAt = addedAt.Time
			}
			if updatedAt.Valid {
				item.UpdatedAt = updatedAt.Time
			}
			cart.Items = append(cart.Items, item)
		}
	}

	if !cartFound {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "NOT_FOUND",
			Message: "Shopping cart not found",
		})
		return
	}

	log.Printf("Retrieved cart %d with %d items (using JOIN)", cartID, len(cart.Items))
	c.JSON(http.StatusOK, cart)
}

// AddItemToCart adds or updates an item in the cart
// POST /shopping-carts/{shoppingCartId}/items
func AddItemToCart(c *gin.Context) {
	cartIDStr := c.Param("shoppingCartId")
	cartID, err := strconv.Atoi(cartIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid cart ID",
		})
		return
	}

	var req models.AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Check if cart exists
	var exists int
	err = database.DB.QueryRow("SELECT 1 FROM shopping_carts WHERE cart_id = ?", cartID).Scan(&exists)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "NOT_FOUND",
			Message: "Shopping cart not found",
		})
		return
	}
	if err != nil {
		log.Printf("Error checking cart existence: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to verify cart",
		})
		return
	}

	// Insert or update cart item (MySQL UPSERT using ON DUPLICATE KEY UPDATE)
	_, err = database.DB.Exec(`
		INSERT INTO cart_items (cart_id, product_id, quantity, added_at, updated_at) 
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			quantity = quantity + VALUES(quantity),
			updated_at = VALUES(updated_at)
	`, cartID, req.ProductID, req.Quantity, time.Now(), time.Now())

	if err != nil {
		log.Printf("Error adding item to cart: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to add item to cart",
		})
		return
	}

	// Update cart's updated_at timestamp
	_, err = database.DB.Exec("UPDATE shopping_carts SET updated_at = ? WHERE cart_id = ?", time.Now(), cartID)
	if err != nil {
		log.Printf("Warning: Failed to update cart timestamp: %v", err)
	}

	log.Printf("Added %d x product %d to cart %d", req.Quantity, req.ProductID, cartID)
	c.Status(http.StatusNoContent)
}

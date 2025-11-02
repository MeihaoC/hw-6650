package models

import "time"

// ShoppingCart represents a shopping cart
type ShoppingCart struct {
	CartID     int        `json:"cart_id"`
	CustomerID int        `json:"customer_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Items      []CartItem `json:"items,omitempty"`
}

// CartItem represents an item in a shopping cart
type CartItem struct {
	CartID    int       `json:"cart_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	AddedAt   time.Time `json:"added_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCartRequest for POST /shopping-carts
type CreateCartRequest struct {
	CustomerID int `json:"customer_id" binding:"required,min=1"`
}

// CreateCartResponse for POST /shopping-carts
type CreateCartResponse struct {
	ShoppingCartID int `json:"shopping_cart_id"`
}

// AddItemRequest for POST /shopping-carts/{id}/items
type AddItemRequest struct {
	ProductID int `json:"product_id" binding:"required,min=1"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

// ErrorResponse for error handling
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

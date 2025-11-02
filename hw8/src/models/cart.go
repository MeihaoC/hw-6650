package models

import "time"

// ShoppingCart represents a shopping cart (works for both MySQL and DynamoDB)
type ShoppingCart struct {
	CartID     string     `json:"cart_id" dynamodbav:"cart_id"`
	CustomerID int        `json:"customer_id" dynamodbav:"customer_id"`
	CreatedAt  time.Time  `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" dynamodbav:"updated_at"`
	Items      []CartItem `json:"items,omitempty" dynamodbav:"items,omitempty"`
}

// CartItem for DynamoDB (embedded in cart)
type CartItem struct {
	ProductID int       `json:"product_id" dynamodbav:"product_id"`
	Quantity  int       `json:"quantity" dynamodbav:"quantity"`
	AddedAt   time.Time `json:"added_at" dynamodbav:"added_at"`
	UpdatedAt time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

// CreateCartRequest for POST /shopping-carts
type CreateCartRequest struct {
	CustomerID int `json:"customer_id" binding:"required,min=1"`
}

// CreateCartResponse for POST /shopping-carts
type CreateCartResponse struct {
	ShoppingCartID string `json:"shopping_cart_id"` // Changed to string for DynamoDB UUID
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

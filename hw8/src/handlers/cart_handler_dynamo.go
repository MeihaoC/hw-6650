package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"shopping-cart-service/database"
	"shopping-cart-service/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateCartDynamo creates a new shopping cart in DynamoDB
// POST /shopping-carts
func CreateCartDynamo(c *gin.Context) {
	var req models.CreateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Generate UUID for cart_id
	cartID := uuid.New().String()

	// Create cart object
	cart := models.ShoppingCart{
		CartID:     cartID,
		CustomerID: req.CustomerID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Items:      []models.CartItem{}, // Empty items array
	}

	// Marshal to DynamoDB attribute values
	item, err := attributevalue.MarshalMap(cart)
	if err != nil {
		log.Printf("Error marshaling cart: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to create shopping cart",
		})
		return
	}

	// Put item in DynamoDB
	_, err = database.DynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(database.TableName),
		Item:      item,
	})

	if err != nil {
		log.Printf("Error putting item to DynamoDB: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to create shopping cart",
		})
		return
	}

	log.Printf("Created cart %s for customer %d (DynamoDB)", cartID, req.CustomerID)
	c.JSON(http.StatusCreated, models.CreateCartResponse{
		ShoppingCartID: cartID,
	})
}

// GetCartDynamo retrieves a shopping cart from DynamoDB
// GET /shopping-carts/{shoppingCartId}
func GetCartDynamo(c *gin.Context) {
	cartID := c.Param("shoppingCartId")
	if cartID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid cart ID",
		})
		return
	}

	// Get item from DynamoDB
	result, err := database.DynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(database.TableName),
		Key: map[string]types.AttributeValue{
			"cart_id": &types.AttributeValueMemberS{Value: cartID},
		},
	})

	if err != nil {
		log.Printf("Error getting item from DynamoDB: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to retrieve cart",
		})
		return
	}

	// Check if item exists
	if result.Item == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "NOT_FOUND",
			Message: "Shopping cart not found",
		})
		return
	}

	// Unmarshal DynamoDB item to cart struct
	var cart models.ShoppingCart
	err = attributevalue.UnmarshalMap(result.Item, &cart)
	if err != nil {
		log.Printf("Error unmarshaling cart: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to parse cart data",
		})
		return
	}

	// Ensure items is not nil (return empty array instead)
	if cart.Items == nil {
		cart.Items = []models.CartItem{}
	}

	log.Printf("Retrieved cart %s with %d items (DynamoDB)", cartID, len(cart.Items))
	c.JSON(http.StatusOK, cart)
}

// AddItemToCartDynamo adds or updates an item in the cart (DynamoDB)
// POST /shopping-carts/{shoppingCartId}/items
func AddItemToCartDynamo(c *gin.Context) {
	cartID := c.Param("shoppingCartId")
	if cartID == "" {
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

	// First, get the existing cart
	result, err := database.DynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(database.TableName),
		Key: map[string]types.AttributeValue{
			"cart_id": &types.AttributeValueMemberS{Value: cartID},
		},
	})

	if err != nil {
		log.Printf("Error getting cart: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to retrieve cart",
		})
		return
	}

	if result.Item == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "NOT_FOUND",
			Message: "Shopping cart not found",
		})
		return
	}

	// Unmarshal existing cart
	var cart models.ShoppingCart
	err = attributevalue.UnmarshalMap(result.Item, &cart)
	if err != nil {
		log.Printf("Error unmarshaling cart: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to parse cart",
		})
		return
	}

	// Initialize items array if nil
	if cart.Items == nil {
		cart.Items = []models.CartItem{}
	}

	// Check if product already exists in cart
	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == req.ProductID {
			// Update existing item quantity
			cart.Items[i].Quantity += req.Quantity
			cart.Items[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	// If product not found, add new item
	if !found {
		newItem := models.CartItem{
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
			AddedAt:   time.Now(),
			UpdatedAt: time.Now(),
		}
		cart.Items = append(cart.Items, newItem)
	}

	// Update cart timestamp
	cart.UpdatedAt = time.Now()

	// Marshal updated cart
	item, err := attributevalue.MarshalMap(cart)
	if err != nil {
		log.Printf("Error marshaling updated cart: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to update cart",
		})
		return
	}

	// Put updated cart back to DynamoDB
	_, err = database.DynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(database.TableName),
		Item:      item,
	})

	if err != nil {
		log.Printf("Error updating cart: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Failed to add item to cart",
		})
		return
	}

	log.Printf("Added %d x product %d to cart %s (DynamoDB)", req.Quantity, req.ProductID, cartID)
	c.Status(http.StatusNoContent)
}

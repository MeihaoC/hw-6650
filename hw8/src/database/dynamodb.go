package database

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var DynamoClient *dynamodb.Client
var TableName string

// InitDynamoDB initializes DynamoDB client
func InitDynamoDB() error {
	ctx := context.Background()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return err
	}

	// Create DynamoDB client
	DynamoClient = dynamodb.NewFromConfig(cfg)

	// Get table name from environment
	TableName = os.Getenv("DYNAMODB_TABLE_NAME")
	if TableName == "" {
		log.Fatal("DYNAMODB_TABLE_NAME environment variable not set")
	}

	log.Printf("DynamoDB client initialized (table: %s)", TableName)
	return nil
}

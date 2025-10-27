package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"hw7/pkg/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Simulate payment processing (same 3-second delay)
func simulatePaymentProcessing() {
	time.Sleep(3 * time.Second)
}

// Handler processes SNS events
func handler(ctx context.Context, snsEvent events.SNSEvent) error {
	log.Printf("Received %d SNS messages", len(snsEvent.Records))

	for _, record := range snsEvent.Records {
		snsMessage := record.SNS.Message

		// Parse the order from SNS message
		var order models.Order
		if err := json.Unmarshal([]byte(snsMessage), &order); err != nil {
			log.Printf("Failed to parse order: %v", err)
			continue
		}

		log.Printf("Processing order: %s (customer: %d)", order.OrderID, order.CustomerID)

		// Simulate payment processing
		simulatePaymentProcessing()

		log.Printf("Completed order: %s", order.OrderID)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}

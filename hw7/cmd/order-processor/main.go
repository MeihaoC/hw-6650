package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"hw7/pkg/models"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var (
	sqsClient *sqs.Client
	queueURL  string
)

func init() {
	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-west-2"),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	sqsClient = sqs.NewFromConfig(cfg)

	queueURL = os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL environment variable not set")
	}
}

// Simulate payment processing (same 3-second delay)
func simulatePaymentProcessing() {
	done := make(chan bool, 1)
	go func() {
		time.Sleep(3 * time.Second)
		done <- true
	}()
	<-done
}

// Process a single order
func processOrder(order models.Order, receiptHandle string, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("Processing order: %s (customer: %d)", order.OrderID, order.CustomerID)

	// Simulate payment processing
	simulatePaymentProcessing()

	log.Printf("Completed order: %s", order.OrderID)

	// Delete message from queue after successful processing
	_, err := sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: &receiptHandle,
	})

	if err != nil {
		log.Printf("Failed to delete message: %v", err)
	}
}

// Poll SQS continuously
func pollQueue(numWorkers int) {
	log.Printf("Starting order processor with %d worker(s)", numWorkers)

	var wg sync.WaitGroup

	for {
		// Long polling: wait up to 20 seconds for messages
		result, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 10, // Receive up to 10 messages
			WaitTimeSeconds:     20, // Long polling (20 seconds)
			VisibilityTimeout:   30, // 30 seconds to process
		})

		if err != nil {
			log.Printf("Error receiving messages: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if len(result.Messages) == 0 {
			log.Println("No messages received, continuing to poll...")
			continue
		}

		log.Printf("Received %d messages from queue", len(result.Messages))

		// Process each message in a separate goroutine
		for _, msg := range result.Messages {
			// Parse the SNS message wrapper
			var snsMessage struct {
				Message string `json:"Message"`
			}

			if err := json.Unmarshal([]byte(*msg.Body), &snsMessage); err != nil {
				log.Printf("Failed to parse SNS wrapper: %v", err)
				continue
			}

			// Parse the actual order
			var order models.Order
			if err := json.Unmarshal([]byte(snsMessage.Message), &order); err != nil {
				log.Printf("Failed to parse order: %v", err)
				continue
			}

			// Spawn goroutine to process this order
			wg.Add(1)
			go processOrder(order, *msg.ReceiptHandle, &wg)
		}
	}
}

func main() {
	// Default to 1 worker (Phase 3 requirement)
	numWorkers := 1
	if workers := os.Getenv("NUM_WORKERS"); workers != "" {
		// We'll use this in Phase 5 for scaling
		log.Printf("NUM_WORKERS set to: %s (not used in this simple version)", workers)
	}

	log.Printf("Order Processor starting...")
	log.Printf("Queue URL: %s", queueURL)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		os.Exit(0)
	}()

	// Start polling
	pollQueue(numWorkers)
}

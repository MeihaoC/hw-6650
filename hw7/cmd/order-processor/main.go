package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"hw7/pkg/models"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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

// Worker goroutine - processes messages from a channel
func worker(id int, jobs <-chan types.Message, wg *sync.WaitGroup) {
	for msg := range jobs {
		// Parse the SNS message wrapper
		var snsMessage struct {
			Message string `json:"Message"`
		}

		if err := json.Unmarshal([]byte(*msg.Body), &snsMessage); err != nil {
			log.Printf("Worker %d: Failed to parse SNS wrapper: %v", id, err)
			wg.Done()
			continue
		}

		// Parse the actual order
		var order models.Order
		if err := json.Unmarshal([]byte(snsMessage.Message), &order); err != nil {
			log.Printf("Worker %d: Failed to parse order: %v", id, err)
			wg.Done()
			continue
		}

		log.Printf("Worker %d: Processing order %s (customer: %d)", id, order.OrderID, order.CustomerID)

		// Simulate payment processing
		simulatePaymentProcessing()

		log.Printf("Worker %d: Completed order %s", id, order.OrderID)

		// Delete message from queue after successful processing
		_, err := sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
			QueueUrl:      &queueURL,
			ReceiptHandle: msg.ReceiptHandle,
		})

		if err != nil {
			log.Printf("Worker %d: Failed to delete message: %v", id, err)
		}

		wg.Done()
	}
}

// Poll SQS continuously with worker pool
func pollQueue(numWorkers int) {
	log.Printf("Starting order processor with %d worker(s)", numWorkers)

	// Create a buffered channel for jobs
	jobs := make(chan types.Message, 100)

	var wg sync.WaitGroup

	// Start worker goroutines
	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs, &wg)
	}

	// Main polling loop
	for {
		// Long polling: wait up to 20 seconds for messages
		result, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20,
			VisibilityTimeout:   30,
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

		// Send messages to workers
		for _, msg := range result.Messages {
			wg.Add(1)
			jobs <- msg
		}
	}
}

func main() {
	// Read NUM_WORKERS from environment variable
	numWorkers := 1
	if workers := os.Getenv("NUM_WORKERS"); workers != "" {
		if w, err := strconv.Atoi(workers); err == nil && w > 0 {
			numWorkers = w
		} else {
			log.Printf("Invalid NUM_WORKERS value: %s, defaulting to 1", workers)
		}
	}

	log.Printf("Order Processor starting...")
	log.Printf("Queue URL: %s", queueURL)
	log.Printf("Number of workers: %d", numWorkers)

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

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"hw7/pkg/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/google/uuid"
)

var (
	snsClient   *sns.SNS
	snsTopicARN string
)

func init() {
	// Initialize AWS session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))

	snsClient = sns.New(sess)

	// Get SNS topic ARN from environment variable
	snsTopicARN = os.Getenv("SNS_TOPIC_ARN")
	if snsTopicARN == "" {
		log.Println("Warning: SNS_TOPIC_ARN not set, async endpoint will not work")
	}
}

// Simulate payment processing bottleneck
func simulatePaymentProcessing() {
	// Create a buffered channel to actually block the goroutine
	// (time.Sleep doesn't block the thread in Go)
	done := make(chan bool, 1)
	go func() {
		time.Sleep(3 * time.Second)
		done <- true
	}()
	<-done
}

func handleSyncOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate order ID and timestamp
	order.OrderID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.Status = "processing"

	log.Printf("Processing sync order: %s", order.OrderID)

	// Simulate 3-second payment processing (THE BOTTLENECK!)
	simulatePaymentProcessing()

	order.Status = "completed"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func handleAsyncOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order.OrderID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.Status = "pending" // Note: pending, not processing!

	log.Printf("Accepting async order: %s", order.OrderID)

	// Publish to SNS (fast!)
	orderJSON, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Failed to serialize order", http.StatusInternalServerError)
		return
	}

	_, err = snsClient.Publish(&sns.PublishInput{
		TopicArn: aws.String(snsTopicARN),
		Message:  aws.String(string(orderJSON)),
	})

	if err != nil {
		log.Printf("Failed to publish to SNS: %v", err)
		http.Error(w, "Failed to queue order", http.StatusInternalServerError)
		return
	}

	// Return immediately with 202 Accepted
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202, not 200!
	json.NewEncoder(w).Encode(order)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/orders/sync", handleSyncOrder)
	http.HandleFunc("/orders/async", handleAsyncOrder)
	http.HandleFunc("/health", handleHealth)

	port := "8080"
	log.Printf("Order API starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

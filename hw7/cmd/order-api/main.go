package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"hw7/pkg/models"

	"github.com/google/uuid"
)

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

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/orders/sync", handleSyncOrder)
	http.HandleFunc("/health", handleHealth)

	port := "8080"
	log.Printf("Order API starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// SplitRequest represents the incoming HTTP request
// Client will send: {"s3_url": "s3://bucket-name/shakespeare-hamlet.txt"}
type SplitRequest struct {
	S3URL string `json:"s3_url"`
}

// SplitResponse represents what we send back to the client
// We'll return: {"chunk_urls": ["s3://bucket/chunk-0.txt", "s3://bucket/chunk-1.txt", "s3://bucket/chunk-2.txt"]}
type SplitResponse struct {
	ChunkURLs []string `json:"chunk_urls"`
}

func main() {
	// Register the /split endpoint to handle incoming requests
	// When someone POSTs to /split, handleSplit function will be called
	http.HandleFunc("/split", handleSplit)

	// Get the port from environment variable (ECS will set this)
	// If not set, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the HTTP server
	log.Printf("Splitter service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// handleSplit is the main function that processes split requests
// It: 1) Downloads file from S3, 2) Splits it, 3) Uploads chunks, 4) Returns URLs
func handleSplit(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON request body to get the S3 URL
	var req SplitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received split request for: %s", req.S3URL)

	// Extract bucket name and file key from the S3 URL
	// Example: "s3://my-bucket/shakespeare.txt" -> bucket="my-bucket", key="shakespeare.txt"
	bucket, key := parseS3URL(req.S3URL)
	if bucket == "" || key == "" {
		http.Error(w, "Invalid S3 URL", http.StatusBadRequest)
		return
	}

	// Create a new AWS session using credentials from environment
	// ECS task role will automatically provide these credentials
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")), // e.g., "us-east-1"
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create S3 service client
	svc := s3.New(sess)

	// STEP 1: Download the original file from S3
	log.Printf("Downloading file from S3: bucket=%s, key=%s", bucket, key)
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get object: %v", err), http.StatusInternalServerError)
		return
	}
	defer result.Body.Close()

	// Read the entire file content into memory
	content, err := io.ReadAll(result.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Downloaded file size: %d bytes", len(content))

	// STEP 2: Split the content into 3 equal chunks
	chunks := splitIntoChunks(string(content), 3)
	log.Printf("Split into %d chunks", len(chunks))

	// STEP 3: Upload each chunk back to S3
	var chunkURLs []string
	timestamp := time.Now().Unix() // Use timestamp to make filenames unique

	for i, chunk := range chunks {
		// Create unique key for each chunk
		// Example: "chunks/1701234567-chunk-0.txt"
		chunkKey := fmt.Sprintf("chunks/%d-chunk-%d.txt", timestamp, i)

		// Upload chunk to S3
		_, err := svc.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucket), // Same bucket as source
			Key:    aws.String(chunkKey),
			Body:   bytes.NewReader([]byte(chunk)),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to upload chunk: %v", err), http.StatusInternalServerError)
			return
		}

		// Create S3 URL for this chunk and add to response
		chunkURL := fmt.Sprintf("s3://%s/%s", bucket, chunkKey)
		chunkURLs = append(chunkURLs, chunkURL)
		log.Printf("Uploaded chunk %d to %s (size: %d bytes)", i, chunkURL, len(chunk))
	}

	// STEP 4: Send response with chunk URLs back to client
	resp := SplitResponse{ChunkURLs: chunkURLs}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	log.Printf("Successfully split file into %d chunks", len(chunkURLs))
}

// parseS3URL extracts bucket and key from an S3 URL
// Supports two formats:
// - s3://bucket-name/path/to/file.txt
// - https://bucket-name.s3.amazonaws.com/path/to/file.txt
func parseS3URL(s3URL string) (bucket, key string) {
	if strings.HasPrefix(s3URL, "s3://") {
		// Format: s3://bucket/key
		// Remove "s3://" prefix
		trimmed := strings.TrimPrefix(s3URL, "s3://")
		// Split on first "/" to separate bucket from key
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	} else if strings.Contains(s3URL, ".s3.amazonaws.com/") {
		// Format: https://bucket.s3.amazonaws.com/key
		parts := strings.SplitN(s3URL, ".s3.amazonaws.com/", 2)
		if len(parts) == 2 {
			bucket = strings.TrimPrefix(parts[0], "https://")
			key = parts[1]
			return bucket, key
		}
	}
	// Return empty strings if URL format is not recognized
	return "", ""
}

// splitIntoChunks divides text into equal parts based on word count
// This ensures we don't split in the middle of words
func splitIntoChunks(content string, numChunks int) []string {
	// Split the entire text into individual words
	// strings.Fields() splits on any whitespace and removes empty strings
	words := strings.Fields(content)
	totalWords := len(words)

	// Calculate how many words should go in each chunk
	// For example: 9000 words / 3 chunks = 3000 words per chunk
	wordsPerChunk := totalWords / numChunks

	log.Printf("Total words: %d, words per chunk: %d", totalWords, wordsPerChunk)

	// Create array to store our chunks
	chunks := make([]string, 0, numChunks)

	// Create each chunk
	for i := 0; i < numChunks; i++ {
		// Calculate start and end indices for this chunk
		start := i * wordsPerChunk
		end := start + wordsPerChunk

		// The last chunk gets any remaining words
		// This handles cases where totalWords isn't perfectly divisible by numChunks
		if i == numChunks-1 {
			end = totalWords
		}

		// Make sure we don't go past the array bounds
		if start < totalWords {
			// Get the words for this chunk and join them back with spaces
			chunkWords := words[start:end]
			chunk := strings.Join(chunkWords, " ")
			chunks = append(chunks, chunk)

			log.Printf("Chunk %d: words %d-%d (total: %d words)",
				i, start, end-1, len(chunkWords))
		}
	}

	return chunks
}

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

// MapRequest contains the S3 URL of the chunk to process
type MapRequest struct {
	ChunkURL string `json:"chunk_url"`
}

// MapResponse contains the S3 URL of the word count results
type MapResponse struct {
	ResultURL string `json:"result_url"`
}

// WordCount holds the count results
type WordCount map[string]int

func main() {
	// Register the /map endpoint
	http.HandleFunc("/map", handleMap)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Different port from splitter
	}

	log.Printf("Mapper service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// handleMap processes a chunk of text and counts word occurrences
func handleMap(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request to get chunk URL
	var req MapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received map request for chunk: %s", req.ChunkURL)

	// Parse S3 URL to get bucket and key
	bucket, key := parseS3URL(req.ChunkURL)
	if bucket == "" || key == "" {
		http.Error(w, "Invalid S3 URL", http.StatusBadRequest)
		return
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)

	// STEP 1: Download chunk from S3
	log.Printf("Downloading chunk from S3: bucket=%s, key=%s", bucket, key)
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get chunk: %v", err), http.StatusInternalServerError)
		return
	}
	defer result.Body.Close()

	// Read chunk content
	content, err := io.ReadAll(result.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Downloaded chunk size: %d bytes", len(content))

	// STEP 2: Count words in the chunk
	wordCounts := countWords(string(content))
	log.Printf("Counted %d unique words", len(wordCounts))

	// STEP 3: Convert counts to JSON
	jsonData, err := json.MarshalIndent(wordCounts, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// STEP 4: Upload results to S3
	timestamp := time.Now().Unix()
	// Extract chunk number from original key for naming
	chunkNum := extractChunkNumber(key)
	resultKey := fmt.Sprintf("results/%d-mapper-%s.json", timestamp, chunkNum)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(resultKey),
		Body:        bytes.NewReader(jsonData),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload results: %v", err), http.StatusInternalServerError)
		return
	}

	resultURL := fmt.Sprintf("s3://%s/%s", bucket, resultKey)
	log.Printf("Uploaded results to %s", resultURL)

	// STEP 5: Return response
	resp := MapResponse{ResultURL: resultURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// countWords counts occurrences of each word in the text
func countWords(text string) WordCount {
	counts := make(WordCount)

	// Convert to lowercase for case-insensitive counting
	text = strings.ToLower(text)

	// Split into words
	words := strings.Fields(text)

	for _, word := range words {
		// Clean the word - remove common punctuation
		word = cleanWord(word)

		// Skip empty words
		if word == "" {
			continue
		}

		// Increment count for this word
		counts[word]++
	}

	return counts
}

// cleanWord removes common punctuation from start and end of words
func cleanWord(word string) string {
	// Remove common punctuation
	punctuation := ".,;:!?\"'()[]{}—–-"

	// Trim punctuation from both ends
	word = strings.Trim(word, punctuation)

	// Additional cleaning for common cases
	// Remove possessive 's
	if strings.HasSuffix(word, "'s") {
		word = strings.TrimSuffix(word, "'s")
	}

	return word
}

// parseS3URL extracts bucket and key from S3 URL
func parseS3URL(s3URL string) (bucket, key string) {
	if strings.HasPrefix(s3URL, "s3://") {
		trimmed := strings.TrimPrefix(s3URL, "s3://")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return "", ""
}

// extractChunkNumber gets the chunk number from the key
// e.g., "chunks/1234-chunk-0.txt" -> "0"
func extractChunkNumber(key string) string {
	parts := strings.Split(key, "-")
	for i, part := range parts {
		if part == "chunk" && i+1 < len(parts) {
			// Remove .txt extension
			num := strings.TrimSuffix(parts[i+1], ".txt")
			return num
		}
	}
	return "unknown"
}

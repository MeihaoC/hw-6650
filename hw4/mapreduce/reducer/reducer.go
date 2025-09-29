package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ReduceRequest contains the S3 URLs of mapper outputs to aggregate
type ReduceRequest struct {
	ResultURLs []string `json:"result_urls"`
}

// ReduceResponse contains the final aggregated results URL
type ReduceResponse struct {
	FinalResultURL string `json:"final_result_url"`
	TotalWords     int    `json:"total_words"`
	UniqueWords    int    `json:"unique_words"`
}

// WordCount holds word counts
type WordCount map[string]int

// WordFrequency for sorting results
type WordFrequency struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

func main() {
	// Register the /reduce endpoint
	http.HandleFunc("/reduce", handleReduce)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Different port from splitter and mapper
	}

	log.Printf("Reducer service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// handleReduce aggregates word counts from multiple mapper outputs
func handleReduce(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request to get mapper result URLs
	var req ReduceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received reduce request for %d mapper results", len(req.ResultURLs))

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)

	// STEP 1: Aggregate counts from all mapper outputs
	aggregatedCounts := make(WordCount)

	for i, resultURL := range req.ResultURLs {
		log.Printf("Processing mapper result %d: %s", i+1, resultURL)

		// Parse S3 URL
		bucket, key := parseS3URL(resultURL)
		if bucket == "" || key == "" {
			http.Error(w, fmt.Sprintf("Invalid S3 URL: %s", resultURL), http.StatusBadRequest)
			return
		}

		// Download mapper result from S3
		result, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get mapper result: %v", err), http.StatusInternalServerError)
			return
		}

		// Read and parse JSON
		content, err := io.ReadAll(result.Body)
		result.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Parse the mapper's word counts
		var mapperCounts WordCount
		if err := json.Unmarshal(content, &mapperCounts); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse mapper results: %v", err), http.StatusInternalServerError)
			return
		}

		// Aggregate counts
		for word, count := range mapperCounts {
			aggregatedCounts[word] += count
		}

		log.Printf("Aggregated %d words from mapper %d", len(mapperCounts), i+1)
	}

	// STEP 2: Calculate statistics
	totalWords := 0
	uniqueWords := len(aggregatedCounts)

	for _, count := range aggregatedCounts {
		totalWords += count
	}

	log.Printf("Final aggregation: %d total words, %d unique words", totalWords, uniqueWords)

	// STEP 3: Sort words by frequency (optional but useful)
	sortedWords := sortByFrequency(aggregatedCounts)

	// STEP 4: Create final result structure
	finalResult := map[string]interface{}{
		"total_words":  totalWords,
		"unique_words": uniqueWords,
		"word_counts":  aggregatedCounts,
		"top_50_words": getTopN(sortedWords, 50),
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(finalResult, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// STEP 5: Upload final results to S3
	// Use the bucket from first result URL
	bucket, _ := parseS3URL(req.ResultURLs[0])
	timestamp := time.Now().Unix()
	finalKey := fmt.Sprintf("final/%d-word-count-final.json", timestamp)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(finalKey),
		Body:        bytes.NewReader(jsonData),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload final results: %v", err), http.StatusInternalServerError)
		return
	}

	finalURL := fmt.Sprintf("s3://%s/%s", bucket, finalKey)
	log.Printf("Uploaded final results to %s", finalURL)

	// STEP 6: Also create a simple CSV for easy viewing
	csvData := createCSV(sortedWords)
	csvKey := fmt.Sprintf("final/%d-word-count-final.csv", timestamp)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(csvKey),
		Body:        bytes.NewReader([]byte(csvData)),
		ContentType: aws.String("text/csv"),
	})
	if err != nil {
		log.Printf("Warning: Failed to upload CSV: %v", err)
		// Don't fail the request if CSV upload fails
	}

	// STEP 7: Return response
	resp := ReduceResponse{
		FinalResultURL: finalURL,
		TotalWords:     totalWords,
		UniqueWords:    uniqueWords,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.Printf("Reduction complete! Total: %d words, Unique: %d words", totalWords, uniqueWords)
}

// sortByFrequency sorts words by their count (descending)
func sortByFrequency(counts WordCount) []WordFrequency {
	result := make([]WordFrequency, 0, len(counts))

	for word, count := range counts {
		result = append(result, WordFrequency{
			Word:  word,
			Count: count,
		})
	}

	// Sort by count (descending), then by word (ascending) for ties
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].Word < result[j].Word
	})

	return result
}

// getTopN returns the top N words by frequency
func getTopN(sorted []WordFrequency, n int) []WordFrequency {
	if len(sorted) < n {
		return sorted
	}
	return sorted[:n]
}

// createCSV creates a CSV string from word frequencies
func createCSV(frequencies []WordFrequency) string {
	var buffer bytes.Buffer

	// Write header
	buffer.WriteString("word,count\n")

	// Write data (limit to top 1000 for readability)
	limit := len(frequencies)
	if limit > 1000 {
		limit = 1000
	}

	for i := 0; i < limit; i++ {
		buffer.WriteString(fmt.Sprintf("%s,%d\n",
			frequencies[i].Word,
			frequencies[i].Count))
	}

	return buffer.String()
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

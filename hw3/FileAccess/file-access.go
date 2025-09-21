package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func unbufferedWrite(filename string, iterations int) time.Duration {
	// Create/truncate file
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Record start time
	start := time.Now()

	// Write directly to file (unbuffered)
	for i := 0; i < iterations; i++ {
		line := fmt.Sprintf("Line %d: This is some test data for unbuffered writing\n", i)
		file.Write([]byte(line)) // Direct system call each time!
	}

	// File automatically flushed on close
	elapsed := time.Since(start)
	return elapsed
}

func bufferedWrite(filename string, iterations int) time.Duration {
	// Create/truncate file
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Wrap in buffered writer (default 4KB buffer)
	writer := bufio.NewWriter(file)

	// Record start time
	start := time.Now()

	// Write to buffer
	for i := 0; i < iterations; i++ {
		line := fmt.Sprintf("Line %d: This is some test data for buffered writing\n", i)
		writer.WriteString(line) // Goes to memory buffer first
	}

	// Flush buffer to disk
	writer.Flush() // This is crucial! Writes everything to disk

	elapsed := time.Since(start)
	return elapsed
}

func main() {
	iterations := 100000

	fmt.Println("=== File I/O Buffering Experiment ===")
	fmt.Printf("Writing %d lines to file\n\n", iterations)

	// Test unbuffered writes
	fmt.Println("1. UNBUFFERED writes (direct to disk each line):")
	unbufferedTime := unbufferedWrite("unbuffered.txt", iterations)
	fmt.Printf("   Time: %v\n", unbufferedTime)
	fmt.Printf("   Per write: %v\n\n", unbufferedTime/time.Duration(iterations))

	// Test buffered writes
	fmt.Println("2. BUFFERED writes (accumulate in memory, then flush):")
	bufferedTime := bufferedWrite("buffered.txt", iterations)
	fmt.Printf("   Time: %v\n", bufferedTime)
	fmt.Printf("   Per write: %v\n\n", bufferedTime/time.Duration(iterations))

	// Calculate speedup
	speedup := float64(unbufferedTime) / float64(bufferedTime)
	fmt.Printf("📊 RESULTS:\n")
	fmt.Printf("Buffered is %.2fx faster!\n", speedup)
	fmt.Printf("Time saved: %v\n", unbufferedTime-bufferedTime)

	// Show file sizes to confirm both wrote same amount
	unbufferedInfo, _ := os.Stat("unbuffered.txt")
	bufferedInfo, _ := os.Stat("buffered.txt")
	fmt.Printf("\nFile sizes (should be identical):\n")
	fmt.Printf("  unbuffered.txt: %d bytes\n", unbufferedInfo.Size())
	fmt.Printf("  buffered.txt:   %d bytes\n", bufferedInfo.Size())
}

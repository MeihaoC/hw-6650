package main

import (
	"fmt"
	"runtime"
	"time"
)

func pingPongSingleThread(iterations int) time.Duration {
	// Force Go to use only 1 OS thread
	runtime.GOMAXPROCS(1)

	// Create unbuffered channels for ping-pong
	ping := make(chan struct{})
	pong := make(chan struct{})
	done := make(chan bool)

	// Start timing
	start := time.Now()

	// Goroutine 1: Ping sender
	go func() {
		for i := 0; i < iterations; i++ {
			ping <- struct{}{} // Send ping
			<-pong             // Wait for pong
		}
		done <- true
	}()

	// Goroutine 2: Pong responder
	go func() {
		for i := 0; i < iterations; i++ {
			<-ping             // Wait for ping
			pong <- struct{}{} // Send pong back
		}
	}()

	// Wait for completion
	<-done

	return time.Since(start)
}

func pingPongMultiThread(iterations int) time.Duration {
	// Allow Go to use all available CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Create unbuffered channels
	ping := make(chan struct{})
	pong := make(chan struct{})
	done := make(chan bool)

	// Start timing
	start := time.Now()

	// Goroutine 1: Ping sender
	go func() {
		for i := 0; i < iterations; i++ {
			ping <- struct{}{} // Send ping
			<-pong             // Wait for pong
		}
		done <- true
	}()

	// Goroutine 2: Pong responder
	go func() {
		for i := 0; i < iterations; i++ {
			<-ping             // Wait for ping
			pong <- struct{}{} // Send pong back
		}
	}()

	// Wait for completion
	<-done

	return time.Since(start)
}

func main() {
	iterations := 1000000 // 1 million ping-pongs

	fmt.Println("=== Context Switching Experiment ===")
	fmt.Printf("Performing %d ping-pong exchanges (2 context switches each)\n", iterations)
	fmt.Printf("Total context switches: %d\n\n", iterations*2)

	// Warm-up runs to stabilize CPU
	fmt.Println("Warming up...")
	pingPongSingleThread(10000)
	pingPongMultiThread(10000)

	fmt.Println("\n1. SINGLE OS THREAD (GOMAXPROCS=1):")
	fmt.Println("   Both goroutines must run on the same thread")

	var singleTotal time.Duration
	for i := 1; i <= 3; i++ {
		duration := pingPongSingleThread(iterations)
		fmt.Printf("   Trial %d: %v\n", i, duration)
		singleTotal += duration
	}
	singleAvg := singleTotal / 3
	switchTimeSingle := singleAvg / time.Duration(iterations*2)

	fmt.Printf("   Average total: %v\n", singleAvg)
	fmt.Printf("   Per context switch: %v\n\n", switchTimeSingle)

	fmt.Printf("2. MULTIPLE OS THREADS (GOMAXPROCS=%d):\n", runtime.NumCPU())
	fmt.Println("   Goroutines can run on different threads")

	var multiTotal time.Duration
	for i := 1; i <= 3; i++ {
		duration := pingPongMultiThread(iterations)
		fmt.Printf("   Trial %d: %v\n", i, duration)
		multiTotal += duration
	}
	multiAvg := multiTotal / 3
	switchTimeMulti := multiAvg / time.Duration(iterations*2)

	fmt.Printf("   Average total: %v\n", multiAvg)
	fmt.Printf("   Per context switch: %v\n\n", switchTimeMulti)

	// Analysis
	fmt.Println("📊 RESULTS ANALYSIS:")
	if singleAvg < multiAvg {
		speedup := float64(multiAvg) / float64(singleAvg)
		fmt.Printf("✅ Single-thread is %.2fx FASTER\n", speedup)
		fmt.Printf("Reason: No cross-thread coordination needed\n")
	} else {
		speedup := float64(singleAvg) / float64(multiAvg)
		fmt.Printf("✅ Multi-thread is %.2fx FASTER\n", speedup)
		fmt.Printf("Reason: True parallelism possible\n")
	}

	fmt.Printf("\nContext switch costs:\n")
	fmt.Printf("  Same thread:      %v per switch\n", switchTimeSingle)
	fmt.Printf("  Cross-thread:     %v per switch\n", switchTimeMulti)
	fmt.Printf("  Difference:       %v\n", switchTimeMulti-switchTimeSingle)
}

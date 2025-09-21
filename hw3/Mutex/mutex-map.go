package main

import (
	"fmt"
	"sync"
	"time"
)

// SafeMap wraps a map with a mutex for thread-safe access
type SafeMap struct {
	mu sync.Mutex
	m  map[int]int
}

// Set safely writes a key-value pair
func (sm *SafeMap) Set(key, value int) {
	sm.mu.Lock()         // Lock before modifying
	defer sm.mu.Unlock() // Unlock when function returns
	sm.m[key] = value
}

// Get safely retrieves a value
func (sm *SafeMap) Get(key int) (int, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	val, ok := sm.m[key]
	return val, ok
}

// Len safely returns the map length
func (sm *SafeMap) Len() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return len(sm.m)
}

func main() {
	fmt.Println("=== Mutex-Protected Map with Reads and Writes ===")
	fmt.Println("25 writer goroutines × 1000 writes = 25,000 writes")
	fmt.Println("25 reader goroutines × 2000 reads = 50,000 reads")
	fmt.Println("Total operations: 75,000\n")

	// Run the experiment 3 times and calculate mean
	var totalTime time.Duration

	for trial := 1; trial <= 3; trial++ {
		fmt.Printf("Trial %d: ", trial)

		// Create a new SafeMap
		safeMap := &SafeMap{
			m: make(map[int]int),
		}

		var wg sync.WaitGroup
		start := time.Now()

		// First, populate some initial data
		for i := 0; i < 1000; i++ {
			safeMap.Set(i, i*10)
		}

		// Spawn 25 WRITER goroutines
		for g := 0; g < 25; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				// Each writes 1000 entries
				for i := 0; i < 1000; i++ {
					key := goroutineID*1000 + i
					safeMap.Set(key, i) // Thread-safe WRITE operation
				}
			}(g)
		}

		// Spawn 25 READER goroutines
		for g := 0; g < 25; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				successfulReads := 0
				for i := 0; i < 2000; i++ {
					// Try to read various keys
					key := (goroutineID*100 + i) % 25000
					if _, ok := safeMap.Get(key); ok { // READ operation
						successfulReads++
					}
				}
				// Optionally track successful reads
			}(g)
		}

		wg.Wait()
		elapsed := time.Since(start)
		totalTime += elapsed

		mapLen := safeMap.Len() // Final READ operation
		fmt.Printf("Final map length: %d, Time: %v\n", mapLen, elapsed)
	}

	// Calculate and print mean
	meanTime := totalTime / 3
	fmt.Printf("\nMean time for 75,000 operations: %v\n", meanTime)
	fmt.Printf("Mutex ensures thread-safety for BOTH reads and writes!\n")
}

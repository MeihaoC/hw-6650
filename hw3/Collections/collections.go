package main

import (
	"fmt"
	"sync"
)

func main() {
	// Plain map - NOT thread-safe!
	m := make(map[int]int)
	var wg sync.WaitGroup

	fmt.Println("Starting concurrent map writes...")
	fmt.Println("Expected: 50 goroutines × 1000 entries = 50,000 unique keys")

	// Spawn 50 goroutines
	for g := 0; g < 50; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			// Each goroutine writes 1000 entries
			for i := 0; i < 1000; i++ {
				// Key formula: g*1000 + i ensures unique keys
				// Goroutine 0: keys 0-999
				// Goroutine 1: keys 1000-1999
				// Goroutine 2: keys 2000-2999, etc.
				m[goroutineID*1000+i] = i
			}
		}(g)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	fmt.Printf("Map length: %d\n", len(m))
	fmt.Printf("Missing entries: %d\n", 50000-len(m))
}

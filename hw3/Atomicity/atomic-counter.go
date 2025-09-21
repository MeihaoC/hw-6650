package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {

	// Atomic integer counter
	var ops atomic.Uint64

	var wg sync.WaitGroup

	// Start 50 goroutines
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Each increments 1000 times
			for range 1000 {
				ops.Add(1)
			}
		}()
	}

	wg.Wait()

	fmt.Println("ops:", ops.Load())

	// Regular integer counter for comparison
	var regularOps uint64

	var wg2 sync.WaitGroup

	// Start 50 goroutines for regular counter
	for range 50 {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			// Each increments 1000 times
			for range 1000 {
				regularOps++
			}
		}()
	}

	wg2.Wait()
	fmt.Println("Regular ops:", regularOps)

	// Show the comparison
	fmt.Println("\nExpected value: 50,000")
	fmt.Printf("Atomic counter accuracy: %d (lost: %d)\n",
		ops.Load(), 50000-int(ops.Load()))
	fmt.Printf("Regular counter accuracy: %d (lost: %d)\n",
		regularOps, 50000-int(regularOps))
}

package main

import (
	"fmt"
	"sync"
	"time"
)

// RWMap wraps a map with a RWMutex
type RWMap struct {
	mu sync.RWMutex // RWMutex instead of Mutex
	m  map[int]int
}

// Set uses write lock for writing
func (rwm *RWMap) Set(key, value int) {
	rwm.mu.Lock() // Exclusive write lock
	defer rwm.mu.Unlock()
	rwm.m[key] = value
}

// Get uses read lock for reading
func (rwm *RWMap) Get(key int) (int, bool) {
	rwm.mu.RLock() // Shared read lock - multiple readers allowed!
	defer rwm.mu.RUnlock()
	val, ok := rwm.m[key]
	return val, ok
}

// Len uses read lock since it only reads
func (rwm *RWMap) Len() int {
	rwm.mu.RLock()
	defer rwm.mu.RUnlock()
	return len(rwm.m)
}

func main() {
	fmt.Println("=== RWMutex Map with Reads and Writes ===")
	fmt.Println("25 writer goroutines × 1000 writes = 25,000 writes")
	fmt.Println("25 reader goroutines × 2000 reads = 50,000 reads")
	fmt.Println("Total operations: 75,000\n")

	var totalTime time.Duration

	for trial := 1; trial <= 3; trial++ {
		fmt.Printf("Trial %d: ", trial)

		rwMap := &RWMap{
			m: make(map[int]int),
		}

		var wg sync.WaitGroup
		start := time.Now()

		// Initial data
		for i := 0; i < 1000; i++ {
			rwMap.Set(i, i*10)
		}

		// 25 WRITER goroutines
		for g := 0; g < 25; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for i := 0; i < 1000; i++ {
					key := goroutineID*1000 + i
					rwMap.Set(key, i) // Write lock - exclusive
				}
			}(g)
		}

		// 25 READER goroutines
		for g := 0; g < 25; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for i := 0; i < 2000; i++ {
					key := (goroutineID*100 + i) % 25000
					rwMap.Get(key) // Read lock - can be shared!
				}
			}(g)
		}

		wg.Wait()
		elapsed := time.Since(start)
		totalTime += elapsed

		mapLen := rwMap.Len()
		fmt.Printf("Final map length: %d, Time: %v\n", mapLen, elapsed)
	}

	fmt.Printf("\nMean time for 75,000 operations: %v\n", totalTime/3)
	fmt.Println("RWMutex allows multiple concurrent readers while still protecting against writes!")
}

package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Test structures for each approach

type MutexMap struct {
	mu sync.Mutex
	m  map[int]int
}

type RWMutexMap struct {
	mu sync.RWMutex
	m  map[int]int
}

// Test 1: Balanced Read/Write (50/50)
func testMutexBalanced() time.Duration {
	mm := &MutexMap{m: make(map[int]int)}
	var wg sync.WaitGroup

	// Pre-populate
	for i := 0; i < 1000; i++ {
		mm.m[i] = i
	}

	start := time.Now()

	// 25 writers, 25 readers
	for w := 0; w < 25; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				mm.mu.Lock()
				mm.m[id*1000+i] = i
				mm.mu.Unlock()
			}
		}(w)
	}

	for r := 0; r < 25; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				mm.mu.Lock()
				_ = mm.m[i%26000]
				mm.mu.Unlock()
			}
		}(r)
	}

	wg.Wait()
	return time.Since(start)
}

func testRWMutexBalanced() time.Duration {
	rwm := &RWMutexMap{m: make(map[int]int)}
	var wg sync.WaitGroup

	// Pre-populate
	for i := 0; i < 1000; i++ {
		rwm.m[i] = i
	}

	start := time.Now()

	// 25 writers
	for w := 0; w < 25; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				rwm.mu.Lock()
				rwm.m[id*1000+i] = i
				rwm.mu.Unlock()
			}
		}(w)
	}

	// 25 readers
	for r := 0; r < 25; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				rwm.mu.RLock()
				_ = rwm.m[i%26000]
				rwm.mu.RUnlock()
			}
		}(r)
	}

	wg.Wait()
	return time.Since(start)
}

func testSyncMapBalanced() time.Duration {
	var m sync.Map
	var wg sync.WaitGroup

	// Pre-populate
	for i := 0; i < 1000; i++ {
		m.Store(i, i)
	}

	start := time.Now()

	// 25 writers
	for w := 0; w < 25; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				m.Store(id*1000+i, i)
			}
		}(w)
	}

	// 25 readers
	for r := 0; r < 25; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				m.Load(i % 26000)
			}
		}(r)
	}

	wg.Wait()
	return time.Since(start)
}

// Test 2: Read-Heavy (90% reads, 10% writes)
func testMutexReadHeavy() time.Duration {
	mm := &MutexMap{m: make(map[int]int)}
	var wg sync.WaitGroup

	// Pre-populate
	for i := 0; i < 5000; i++ {
		mm.m[i] = i
	}

	start := time.Now()

	// 5 writers (500 writes each)
	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				mm.mu.Lock()
				mm.m[id*1000+i] = i
				mm.mu.Unlock()
			}
		}(w)
	}

	// 45 readers (500 reads each)
	for r := 0; r < 45; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				mm.mu.Lock()
				_ = mm.m[i%5000]
				mm.mu.Unlock()
			}
		}(r)
	}

	wg.Wait()
	return time.Since(start)
}

func testRWMutexReadHeavy() time.Duration {
	rwm := &RWMutexMap{m: make(map[int]int)}
	var wg sync.WaitGroup

	// Pre-populate
	for i := 0; i < 5000; i++ {
		rwm.m[i] = i
	}

	start := time.Now()

	// 5 writers
	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				rwm.mu.Lock()
				rwm.m[id*1000+i] = i
				rwm.mu.Unlock()
			}
		}(w)
	}

	// 45 readers
	for r := 0; r < 45; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				rwm.mu.RLock()
				_ = rwm.m[i%5000]
				rwm.mu.RUnlock()
			}
		}(r)
	}

	wg.Wait()
	return time.Since(start)
}

func testSyncMapReadHeavy() time.Duration {
	var m sync.Map
	var wg sync.WaitGroup

	// Pre-populate
	for i := 0; i < 5000; i++ {
		m.Store(i, i)
	}

	start := time.Now()

	// 5 writers
	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				m.Store(id*1000+i, i)
			}
		}(w)
	}

	// 45 readers
	for r := 0; r < 45; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				m.Load(i % 5000)
			}
		}(r)
	}

	wg.Wait()
	return time.Since(start)
}

func runBenchmark(name string, fn func() time.Duration) time.Duration {
	var total time.Duration
	for i := 0; i < 3; i++ {
		duration := fn()
		fmt.Printf("      Trial %d: %v\n", i+1, duration)
		total += duration
	}
	avg := total / 3
	fmt.Printf("      Average: %v\n", avg)
	return avg
}

func main() {
	fmt.Println("=== Comprehensive Map Synchronization Comparison ===\n")

	// Test 1: Balanced workload
	fmt.Println("SCENARIO 1: Balanced Read/Write (50/50)")
	fmt.Println("25 writers (1000 writes each) + 25 readers (1000 reads each)")
	fmt.Println("Total: 25,000 writes + 25,000 reads = 50,000 operations\n")

	fmt.Println("  1. Mutex:")
	mutexBalanced := runBenchmark("Mutex Balanced", testMutexBalanced)

	fmt.Println("\n  2. RWMutex:")
	rwMutexBalanced := runBenchmark("RWMutex Balanced", testRWMutexBalanced)

	fmt.Println("\n  3. sync.Map:")
	syncMapBalanced := runBenchmark("sync.Map Balanced", testSyncMapBalanced)

	// Test 2: Read-heavy workload
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("\nSCENARIO 2: Read-Heavy (90% reads, 10% writes)")
	fmt.Println("5 writers (500 writes each) + 45 readers (500 reads each)")
	fmt.Println("Total: 2,500 writes + 22,500 reads = 25,000 operations\n")

	fmt.Println("  1. Mutex:")
	mutexReadHeavy := runBenchmark("Mutex Read-Heavy", testMutexReadHeavy)

	fmt.Println("\n  2. RWMutex:")
	rwMutexReadHeavy := runBenchmark("RWMutex Read-Heavy", testRWMutexReadHeavy)

	fmt.Println("\n  3. sync.Map:")
	syncMapReadHeavy := runBenchmark("sync.Map Read-Heavy", testSyncMapReadHeavy)

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("\n📊 PERFORMANCE SUMMARY\n")

	fmt.Println("Balanced Workload (50% reads, 50% writes):")
	fmt.Printf("  🥇 Winner: ")
	if rwMutexBalanced < mutexBalanced && rwMutexBalanced < syncMapBalanced {
		fmt.Printf("RWMutex (%.2fms)\n", float64(rwMutexBalanced.Microseconds())/1000)
	} else if mutexBalanced < syncMapBalanced {
		fmt.Printf("Mutex (%.2fms)\n", float64(mutexBalanced.Microseconds())/1000)
	} else {
		fmt.Printf("sync.Map (%.2fms)\n", float64(syncMapBalanced.Microseconds())/1000)
	}
	fmt.Printf("  - Mutex:     %.2fms\n", float64(mutexBalanced.Microseconds())/1000)
	fmt.Printf("  - RWMutex:   %.2fms\n", float64(rwMutexBalanced.Microseconds())/1000)
	fmt.Printf("  - sync.Map:  %.2fms\n", float64(syncMapBalanced.Microseconds())/1000)

	fmt.Println("\nRead-Heavy Workload (90% reads, 10% writes):")
	fmt.Printf("  🥇 Winner: ")
	if rwMutexReadHeavy < mutexReadHeavy && rwMutexReadHeavy < syncMapReadHeavy {
		fmt.Printf("RWMutex (%.2fms)\n", float64(rwMutexReadHeavy.Microseconds())/1000)
	} else if syncMapReadHeavy < mutexReadHeavy {
		fmt.Printf("sync.Map (%.2fms)\n", float64(syncMapReadHeavy.Microseconds())/1000)
	} else {
		fmt.Printf("Mutex (%.2fms)\n", float64(mutexReadHeavy.Microseconds())/1000)
	}
	fmt.Printf("  - Mutex:     %.2fms (baseline)\n", float64(mutexReadHeavy.Microseconds())/1000)
	fmt.Printf("  - RWMutex:   %.2fms (%.1f%% improvement)\n",
		float64(rwMutexReadHeavy.Microseconds())/1000,
		float64(mutexReadHeavy-rwMutexReadHeavy)/float64(mutexReadHeavy)*100)
	fmt.Printf("  - sync.Map:  %.2fms (%.1f%% improvement)\n",
		float64(syncMapReadHeavy.Microseconds())/1000,
		float64(mutexReadHeavy-syncMapReadHeavy)/float64(mutexReadHeavy)*100)

	fmt.Println("\n📝 KEY INSIGHTS:")
	fmt.Println("• Mutex: Consistent but forces all operations to serialize")
	fmt.Println("• RWMutex: Excels when reads can happen concurrently")
	fmt.Println("• sync.Map: Best for stable keys with rare updates")
}

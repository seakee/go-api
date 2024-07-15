// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package trace

import (
	"runtime"
	"sync"
	"testing"
)

const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
	GiB // 1073741824
)

var curMem uint64

// TestTraceID_New is a function that tests the concurrency safety of the NewTraceID function.
// It simulates a high concurrency scenario by creating a large number of concurrent goroutines and
// checks whether the generated TraceIDs remain unique under this scenario.
// Note: This function can consume significant amounts of memory.
// Please evaluate your computer's hardware resources before executing.
func TestTraceID_New(t *testing.T) {
	// Create a new instance of TraceID
	tid := NewTraceID()

	// Use a mutex to ensure thread-safe access to the uniqueIDs map
	var mu sync.Mutex
	// Store generated TraceIDs to check for duplicates
	uniqueIDs := make(map[string]struct{})
	// Use a wait group to wait for all concurrent test goroutines to complete
	var wg sync.WaitGroup

	// Define the concurrency level, i.e., the number of goroutines running simultaneously
	const concurrency = 10000
	// Launch the specified number of concurrent goroutines
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// In each goroutine, generate a large number of TraceIDs in a loop
			for j := 0; j < 10000; j++ {
				id := tid.New()
				mu.Lock()
				// Check if the generated ID already exists; if so, record an error and exit the current goroutine
				if _, exists := uniqueIDs[id]; exists {
					mu.Unlock()
					t.Errorf("Duplicate ID found: %s", id)
					return
				}
				// Add the newly generated ID to the uniqueIDs map
				uniqueIDs[id] = struct{}{}
				mu.Unlock()
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	// Record the count of unique IDs generated
	t.Logf("Unique IDs count: %d", len(uniqueIDs))
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func BenchmarkTraceID_New(b *testing.B) {
	id := NewTraceID()
	for i := 0; i < b.N; i++ {
		id.New()
	}
}

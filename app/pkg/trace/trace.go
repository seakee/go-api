// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package trace provides functionality for generating unique trace identifiers.
// It is designed to be used in distributed systems for tracking requests or operations
// across multiple services or components.
package trace

import (
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sk-pkg/util"
)

const (
	// used to initialize the index to a large value to avoid collisions
	initIndex = 10000000
	// used to pad the timestamp in the trace ID
	digits = "0123456789abcdefghijklmnopqrstuvwxyz"
)

var (
	hostnameOnce sync.Once // Ensures hostname is retrieved only once
	hostname     string    // Cached hostname
)

// ID represents a structure for generating unique identifiers.
// It uses a combination of hostname, timestamp, and a sequence number
// to ensure uniqueness across distributed systems.
type ID struct {
	index  uint64     // Sequence number, accessed atomically
	prefix string     // Prefix containing timestamp and hostname
	mu     sync.Mutex // Mutex to ensure thread-safety when updating the prefix
}

// NewTraceID creates and returns a new ID instance.
// It initializes the ID with the current hostname and timestamp.
//
// Returns:
//   - *ID: A pointer to the newly created ID instance.
//
// Example:
//
//	traceID := NewTraceID()
//	uniqueIdentifier := traceID.New()
func NewTraceID() *ID {
	t := &ID{
		index: initIndex,
	}
	t.updatePrefix()
	return t
}

// updatePrefix combines the current timestamp and cached hostname to form the prefix.
// This method should be called with external synchronization.
func (t *ID) updatePrefix() {
	var err error

	t.mu.Lock()
	defer t.mu.Unlock()

	// Retrieve hostname only once
	hostnameOnce.Do(func() {
		hostname, err = os.Hostname()
		if err != nil {
			log.Printf("Failed to get hostname: %v", err)
			// Use a default value if hostname retrieval fails
			hostname = "unknown"
		}
	})

	// Construct the prefix using hostname and current timestamp
	t.prefix = util.SpliceStr(hostname, "-", formatBase36(uint64(time.Now().Unix())), "-")
	t.index = initIndex
}

// New generates and returns a new unique identifier.
//
// Returns:
//   - string: A unique identifier combining the prefix and a sequence number.
//
// Example:
//
//	traceID := NewTraceID()
//	id1 := traceID.New() // e.g., "hostname-timestamp-sequence1"
//	id2 := traceID.New() // e.g., "hostname-timestamp-sequence2"
func (t *ID) New() string {
	// Atomically increment the sequence number
	newIndex := atomic.AddUint64(&t.index, 1)

	// If the sequence number overflows, update the prefix and reset the sequence
	if newIndex == 0 {
		t.mu.Lock()
		defer t.mu.Unlock()
		if atomic.LoadUint64(&t.index) == 0 {
			t.updatePrefix()
		}
	}

	// Convert the sequence number to a base-36 string
	id := formatBase36(newIndex)

	// Combine the prefix and the sequence number
	return util.SpliceStr(t.prefix, id)
}

// formatBase36 converts a number to a base-36 string.
//
// Parameters:
//   - n: The number to be converted.
//
// Returns:
//   - string: The base-36 representation of the number.
func formatBase36(n uint64) string {
	if n == 0 {
		return "0"
	}

	var buf [13]byte // enough for 64-bit numbers
	i := len(buf)

	for n > 0 {
		i--
		buf[i] = digits[n%36]
		n /= 36
	}

	return string(buf[i:])
}

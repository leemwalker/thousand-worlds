package debug

import (
	"log"
	"sync/atomic"
	"time"
)

// Debug flags (bitmask)
const (
	None      uint64 = 0
	Perf      uint64 = 1 << 0 // Performance timing
	Logic     uint64 = 1 << 1 // Logic sanity checks
	Geology   uint64 = 1 << 2 // Geology-specific info
	Tectonics uint64 = 1 << 3 // Tectonics-specific info
	Weather   uint64 = 1 << 4 // Weather-specific info
	All       uint64 = 0xFFFFFFFFFFFFFFFF
)

// activeFlags stores the currently enabled debug flags
// Accessed atomically for thread safety without locks
var activeFlags uint64

// SetFlags sets the active debug flags
func SetFlags(flags uint64) {
	atomic.StoreUint64(&activeFlags, flags)
}

// Enable adds a flag to the active set
func Enable(flag uint64) {
	current := atomic.LoadUint64(&activeFlags)
	atomic.StoreUint64(&activeFlags, current|flag)
}

// Disable removes a flag from the active set
func Disable(flag uint64) {
	current := atomic.LoadUint64(&activeFlags)
	atomic.StoreUint64(&activeFlags, current&^flag)
}

// Is checks if a specific flag is enabled
// Designed to be inlineable and zero-overhead if false
func Is(flag uint64) bool {
	return (atomic.LoadUint64(&activeFlags) & flag) != 0
}

// Log prints a message if the specific flag is enabled
func Log(flag uint64, format string, args ...interface{}) {
	if Is(flag) {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// Time returns a function that, when called, logs the elapsed time.
// Usage: defer debug.Time(debug.Perf, "Operation Name")()
// Returns nil func if flag is not set, minimizing overhead.
func Time(flag uint64, name string) func() {
	if !Is(flag) {
		return func() {}
	}
	start := time.Now()
	return func() {
		log.Printf("[DEBUG] [Perf] %s took %v", name, time.Since(start))
	}
}

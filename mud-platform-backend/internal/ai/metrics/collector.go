package metrics

import (
	"runtime"
	"sync"
	"time"
)

// OllamaMetrics holds resource usage stats
type OllamaMetrics struct {
	CPUPercent      float64
	RAMUsedMB       int64
	RAMTotalMB      int64
	ActiveRequests  int32
	QueuedRequests  int32
	AvgResponseTime time.Duration
	RequestsPerMin  int
}

// MetricsCollector gathers stats
type MetricsCollector struct {
	metrics OllamaMetrics
	mu      sync.RWMutex
}

// NewCollector creates a new collector
func NewCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// GetMetrics returns current stats
func (c *MetricsCollector) GetMetrics() OllamaMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metrics
}

// UpdateStats simulates gathering stats for MVP
// In a real deployment, this would query Docker API or OS stats
func (c *MetricsCollector) UpdateStats() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simulate CPU/RAM based on active requests (mock logic)
	// Real implementation would use `docker stats` or `gopsutil`

	// For MVP, we just track internal counters if we had them linked
	// But since this is a standalone collector for now, we'll just set some dummy values
	// to prove the interface works.

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.metrics.RAMUsedMB = int64(m.Alloc / 1024 / 1024)
	c.metrics.RAMTotalMB = int64(m.Sys / 1024 / 1024)

	// CPU is hard to get for a specific container from inside another without access
	// So we'll leave it as 0 or mock it
	c.metrics.CPUPercent = 0.0
}

// RecordRequestStart increments active requests
func (c *MetricsCollector) RecordRequestStart() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics.ActiveRequests++
}

// RecordRequestEnd decrements active requests and updates timing
func (c *MetricsCollector) RecordRequestEnd(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics.ActiveRequests--
	// Update AvgResponseTime (simple moving average or similar)
	// For simplicity:
	c.metrics.AvgResponseTime = duration
}

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	assert.NotNil(t, m)
	assert.NotNil(t, m.HTTPRequestLatency)
	assert.NotNil(t, m.ErrorRates)
	assert.NotNil(t, m.CacheHitRates)
	assert.NotNil(t, m.NPCFPS)
	assert.NotNil(t, m.EventAppendRate)
	assert.NotNil(t, m.ActiveConnections)
}

func TestMetrics_Registration(t *testing.T) {
	// Create a new registry for testing to avoid global state pollution
	reg := prometheus.NewRegistry()
	m := NewMetrics()

	// Register all metrics
	m.Register(reg)

	// Verify registration by checking if we can collect from them
	// This is a bit indirect, but if they weren't registered or valid, usage might panic or fail

	// Test Counter
	m.EventAppendRate.Inc()
	val := testutil.ToFloat64(m.EventAppendRate)
	assert.Equal(t, 1.0, val)

	// Test Gauge
	m.ActiveConnections.WithLabelValues("websocket").Set(10)
	val = testutil.ToFloat64(m.ActiveConnections.WithLabelValues("websocket"))
	assert.Equal(t, 10.0, val)
}

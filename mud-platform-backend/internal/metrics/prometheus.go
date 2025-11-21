package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds all the prometheus collectors for the application.
type Metrics struct {
	HTTPRequestLatency *prometheus.HistogramVec
	ErrorRates         *prometheus.CounterVec
	CacheHitRates      *prometheus.GaugeVec
	NPCFPS             *prometheus.GaugeVec
	EventAppendRate    prometheus.Counter
	ActiveConnections  *prometheus.GaugeVec
}

// NewMetrics initializes and returns a new Metrics struct.
func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		}, []string{"method", "path", "status"}),
		ErrorRates: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "error_rate_total",
			Help: "Total number of errors",
		}, []string{"service", "endpoint", "error_type"}),
		CacheHitRates: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cache_hit_rate",
			Help: "Cache hit rate (0.0-1.0)",
		}, []string{"cache_type"}), // L1, L2
		NPCFPS: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "npc_simulation_fps",
			Help: "NPC simulation ticks per second",
		}, []string{"world_id"}),
		EventAppendRate: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "event_store_append_total",
			Help: "Total number of events appended",
		}),
		ActiveConnections: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		}, []string{"type"}), // websocket, database
	}
}

// Register registers all metrics with the provided registry.
func (m *Metrics) Register(reg prometheus.Registerer) {
	reg.MustRegister(
		m.HTTPRequestLatency,
		m.ErrorRates,
		m.CacheHitRates,
		m.NPCFPS,
		m.EventAppendRate,
		m.ActiveConnections,
	)
}

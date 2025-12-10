package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	cpuGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ollama_cpu_percent",
		Help: "Current CPU usage of Ollama container",
	})
	ramGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ollama_ram_used_mb",
		Help: "Current RAM usage of Ollama container in MB",
	})
	activeRequestsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ollama_active_requests",
		Help: "Number of currently active requests",
	})
	responseTimeHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "ollama_response_time_seconds",
		Help:    "Response time of Ollama requests",
		Buckets: prometheus.DefBuckets,
	})
)

// UpdatePrometheusMetrics updates the gauges with latest values
func UpdatePrometheusMetrics(m OllamaMetrics) {
	cpuGauge.Set(m.CPUPercent)
	ramGauge.Set(float64(m.RAMUsedMB))
	activeRequestsGauge.Set(float64(m.ActiveRequests))
}

// RecordResponseTime observes the duration
func RecordResponseTime(seconds float64) {
	responseTimeHistogram.Observe(seconds)
}

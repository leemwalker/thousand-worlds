package gamemap

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricRenderDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "game_map_render_duration_seconds",
		Help: "Duration of map render operations",
		// Buckets tailored for render times (50ms to 5s)
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
	}, []string{"status"}) // status: success, timeout, failure

	metricConcurrentRenders = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "game_map_render_concurrent_count",
		Help: "Number of map renders currently in progress",
	})

	metricRenderRejected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "game_map_render_rejected_total",
		Help: "Total number of map render requests rejected due to concurrency limit",
	})

	metricImageSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "game_map_render_image_bytes",
		Help:    "Size of generated map images in bytes",
		Buckets: prometheus.ExponentialBuckets(1024*100, 2, 10), // 100KB to ~100MB
	})

	// Cache metrics (Sprint 1)
	metricCacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "game_map_cache_hits_total",
		Help: "Total number of render cache hits",
	})

	metricCacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "game_map_cache_misses_total",
		Help: "Total number of render cache misses",
	})

	// Render error breakdown (Sprint 1)
	metricRenderErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "game_map_render_errors_total",
			Help: "Total render errors by type",
		},
		[]string{"error_type"}, // timeout, encoding, loading, concurrency
	)
)

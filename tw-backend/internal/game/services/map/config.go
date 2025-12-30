package gamemap

import "time"

// RenderConfig holds configuration for the map rendering service.
type RenderConfig struct {
	// Max dimensions to prevent OOM
	MaxWidth  int
	MaxHeight int

	// Default dimensions if none specified
	DefaultWidth  int
	DefaultHeight int

	// WebP quality (0-100)
	WebPQuality float32

	// RenderTimeout is the hard limit for a single render operation
	RenderTimeout time.Duration

	// ConcurrencyLimit bounds the number of simultaneous renders
	ConcurrencyLimit int
}

// DefaultRenderConfig returns the standard configuration
func DefaultRenderConfig() RenderConfig {
	return RenderConfig{
		MaxWidth:         4096,
		MaxHeight:        4096,
		DefaultWidth:     2048,
		DefaultHeight:    2048,
		WebPQuality:      80.0,
		RenderTimeout:    5 * time.Second,
		ConcurrencyLimit: 2,
	}
}

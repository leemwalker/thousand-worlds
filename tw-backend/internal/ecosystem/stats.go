package ecosystem

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// GlobalStats represents a point-in-time snapshot of world statistics
type GlobalStats struct {
	WorldID          uuid.UUID
	Year             int64
	Population       int64
	AvgTemperature   float64
	MaxTemperature   float64
	MinTemperature   float64
	AvgElevation     float64
	SeaLevel         float64
	AtmosphereCarbon float64 // CO2 ppm
	AtmosphereOxygen float64 // O2 percentage
	RecordedAt       time.Time
}

// MetricsCollector defines the interface for recording simulation statistics
type MetricsCollector interface {
	RecordStats(ctx context.Context, stats GlobalStats) error
}

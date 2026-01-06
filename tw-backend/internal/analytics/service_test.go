package analytics

import (
	"context"
	"testing"
	"time"

	"tw-backend/internal/ecosystem"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestService_RecordStats_Unit tests the RecordStats method without DB
// This is a unit test that verifies the struct implements the interface
func TestService_ImplementsMetricsCollector(t *testing.T) {
	// Compile-time interface check
	var _ ecosystem.MetricsCollector = (*Service)(nil)
}

// TestGlobalStats_Defaults tests GlobalStats struct initialization
func TestGlobalStats_Defaults(t *testing.T) {
	stats := ecosystem.GlobalStats{
		WorldID:          uuid.New(),
		Year:             1000000,
		Population:       5000,
		AvgTemperature:   14.5,
		MaxTemperature:   45.0,
		MinTemperature:   -40.0,
		AvgElevation:     250.0,
		SeaLevel:         0.0,
		AtmosphereCarbon: 400.0,
		AtmosphereOxygen: 21.0,
	}

	assert.NotEqual(t, uuid.Nil, stats.WorldID)
	assert.Equal(t, int64(1000000), stats.Year)
	assert.Equal(t, 14.5, stats.AvgTemperature)
	assert.Equal(t, 21.0, stats.AtmosphereOxygen)
}

// TestNewService_InvalidURL tests error handling for bad connection strings
func TestNewService_InvalidURL(t *testing.T) {
	// Invalid URL should fail
	_, err := NewService("postgres://invalid:invalid@localhost:9999/nonexistent?connect_timeout=1")
	require.Error(t, err)
}

// TestNewService_EmptyURL tests error handling for empty URL
func TestNewService_EmptyURL(t *testing.T) {
	_, err := NewService("")
	require.Error(t, err)
}

// Integration test - requires TimescaleDB
// Skip in short mode
func TestService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires a running TimescaleDB instance
	// Set ANALYTICS_TEST_URL environment variable to run
	testURL := "postgres://admin:password123@localhost:5433/mud_metrics?sslmode=disable"

	service, err := NewService(testURL)
	if err != nil {
		t.Skipf("TimescaleDB not available: %v", err)
	}
	defer service.Close()

	// Test recording stats
	ctx := context.Background()
	stats := ecosystem.GlobalStats{
		WorldID:          uuid.New(),
		Year:             1000000,
		Population:       5000,
		AvgTemperature:   14.5,
		MaxTemperature:   45.0,
		MinTemperature:   -40.0,
		AvgElevation:     250.0,
		SeaLevel:         0.0,
		AtmosphereCarbon: 400.0,
		AtmosphereOxygen: 21.0,
		RecordedAt:       time.Now(),
	}

	err = service.RecordStats(ctx, stats)
	assert.NoError(t, err)
}

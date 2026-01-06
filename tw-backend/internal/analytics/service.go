package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tw-backend/internal/circuitbreaker"
	"tw-backend/internal/ecosystem"

	_ "github.com/jackc/pgx/v5/stdlib" // Use pgx stdlib for database/sql compatibility
)

// Service handles analytics and metric storage using TimescaleDB
type Service struct {
	db *sql.DB
	cb *circuitbreaker.CircuitBreaker
}

// Ensure Service implements ecosystem.MetricsCollector
var _ ecosystem.MetricsCollector = (*Service)(nil)

// NewService creates a new analytics service
func NewService(dbURL string) (*Service, error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, fmt.Errorf("open analytics db: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping analytics db: %w", err)
	}

	// Configure circuit breaker for TimescaleDB calls
	cbConfig := circuitbreaker.Config{
		Name:             "timescaledb",
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
	}
	cb := circuitbreaker.New(cbConfig)

	// Log state transitions
	cb.OnStateChange(func(name string, from, to circuitbreaker.State) {
		fmt.Printf("[CircuitBreaker] %s: %s -> %s\n", name, from, to)
	})

	s := &Service{db: db, cb: cb}

	if err := s.initializeSchema(); err != nil {
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return s, nil
}

// Close closes the database connection
func (s *Service) Close() error {
	return s.db.Close()
}

// CircuitBreakerStats returns the current circuit breaker statistics.
func (s *Service) CircuitBreakerStats() circuitbreaker.Stats {
	return s.cb.Stats()
}

// initializeSchema sets up the TimescaleDB extension and hypertables
func (s *Service) initializeSchema() error {
	// Enable TimescaleDB extension
	_, err := s.db.Exec(`CREATE EXTENSION IF NOT EXISTS timescaledb;`)
	if err != nil {
		return fmt.Errorf("create timescaledb extension: %w", err)
	}

	// Create metrics table
	query := `
	CREATE TABLE IF NOT EXISTS world_metrics (
		time             TIMESTAMPTZ NOT NULL,
		world_id         UUID NOT NULL,
		year             BIGINT NOT NULL,
		population       BIGINT,
		avg_temp         DOUBLE PRECISION,
		max_temp         DOUBLE PRECISION,
		min_temp         DOUBLE PRECISION,
		avg_elevation    DOUBLE PRECISION,
		sea_level        DOUBLE PRECISION,
		co2_ppm          DOUBLE PRECISION,
		o2_percent       DOUBLE PRECISION
	);
	`
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("create metrics table: %w", err)
	}

	// Convert to hypertable (partitioned by time)
	// We use 'if not exists' logic by checking if it's already a hypertable
	// Note: create_hypertable fails if table is already one, unless if_not_exists is true (TimescaleDB 2.0+)
	_, err = s.db.Exec(`SELECT create_hypertable('world_metrics', 'time', if_not_exists => TRUE);`)
	if err != nil {
		return fmt.Errorf("create hypertable: %w", err)
	}

	// Create index on world_id and time for fast queries per world
	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_world_metrics_world_time ON world_metrics (world_id, time DESC);`)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	return nil
}

// RecordStats persists a snapshot of world statistics.
// Uses circuit breaker to prevent cascading failures when TimescaleDB is unavailable.
func (s *Service) RecordStats(ctx context.Context, stats ecosystem.GlobalStats) error {
	return s.cb.Execute(ctx, func(ctx context.Context) error {
		return s.recordStatsInternal(ctx, stats)
	})
}

// recordStatsInternal performs the actual database insert.
func (s *Service) recordStatsInternal(ctx context.Context, stats ecosystem.GlobalStats) error {
	query := `
	INSERT INTO world_metrics (
		time, world_id, year, population, 
		avg_temp, max_temp, min_temp, 
		avg_elevation, sea_level, 
		co2_ppm, o2_percent
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Use current time if RecordedAt is zero
	ts := stats.RecordedAt
	if ts.IsZero() {
		ts = time.Now()
	}

	_, err := s.db.ExecContext(ctx, query,
		ts, stats.WorldID, stats.Year, stats.Population,
		stats.AvgTemperature, stats.MaxTemperature, stats.MinTemperature,
		stats.AvgElevation, stats.SeaLevel,
		stats.AtmosphereCarbon, stats.AtmosphereOxygen,
	)

	if err != nil {
		return fmt.Errorf("insert metrics: %w", err)
	}

	return nil
}

// Package ecosystem provides database persistence for simulation events and checkpoints.
package ecosystem

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// PostgresEventLogger implements DBEventLogger for PostgreSQL
type PostgresEventLogger struct {
	db *sql.DB
}

// NewPostgresEventLogger creates a new PostgreSQL event logger
func NewPostgresEventLogger(db *sql.DB) *PostgresEventLogger {
	return &PostgresEventLogger{db: db}
}

// LogEvent stores a simulation event in the database
func (p *PostgresEventLogger) LogEvent(ctx context.Context, event *SimulationEvent) error {
	query := `
		INSERT INTO simulation_events (id, world_id, year, event_type, severity, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := p.db.ExecContext(ctx, query,
		event.ID,
		event.WorldID,
		event.Year,
		event.EventType,
		event.Severity,
		event.Details,
		event.Timestamp,
	)
	return err
}

// GetEvents retrieves simulation events for a world within a year range
func (p *PostgresEventLogger) GetEvents(ctx context.Context, worldID uuid.UUID, fromYear, toYear int64) ([]SimulationEvent, error) {
	query := `
		SELECT id, world_id, year, event_type, severity, details, created_at
		FROM simulation_events
		WHERE world_id = $1 AND year >= $2 AND year <= $3
		ORDER BY year ASC
	`
	rows, err := p.db.QueryContext(ctx, query, worldID, fromYear, toYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []SimulationEvent
	for rows.Next() {
		var e SimulationEvent
		if err := rows.Scan(&e.ID, &e.WorldID, &e.Year, &e.EventType, &e.Severity, &e.Details, &e.Timestamp); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// GetEventsByType retrieves simulation events of a specific type
func (p *PostgresEventLogger) GetEventsByType(ctx context.Context, worldID uuid.UUID, eventType SimulationEventType) ([]SimulationEvent, error) {
	query := `
		SELECT id, world_id, year, event_type, severity, details, created_at
		FROM simulation_events
		WHERE world_id = $1 AND event_type = $2
		ORDER BY year ASC
	`
	rows, err := p.db.QueryContext(ctx, query, worldID, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []SimulationEvent
	for rows.Next() {
		var e SimulationEvent
		if err := rows.Scan(&e.ID, &e.WorldID, &e.Year, &e.EventType, &e.Severity, &e.Details, &e.Timestamp); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// PostgresCheckpointStore implements CheckpointStore for PostgreSQL
type PostgresCheckpointStore struct {
	db *sql.DB
}

// NewPostgresCheckpointStore creates a new PostgreSQL checkpoint store
func NewPostgresCheckpointStore(db *sql.DB) *PostgresCheckpointStore {
	return &PostgresCheckpointStore{db: db}
}

// Save stores a checkpoint in the database
func (p *PostgresCheckpointStore) Save(ctx context.Context, checkpoint *Checkpoint) error {
	query := `
		INSERT INTO world_checkpoints (id, world_id, year, checkpoint_type, state_data, species_count, population_sum, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (world_id, year) DO UPDATE SET
			checkpoint_type = EXCLUDED.checkpoint_type,
			state_data = EXCLUDED.state_data,
			species_count = EXCLUDED.species_count,
			population_sum = EXCLUDED.population_sum,
			created_at = EXCLUDED.created_at
	`
	_, err := p.db.ExecContext(ctx, query,
		checkpoint.ID,
		checkpoint.WorldID,
		checkpoint.Year,
		checkpoint.Type,
		checkpoint.StateData,
		checkpoint.SpeciesCount,
		checkpoint.PopulationSum,
		checkpoint.CreatedAt,
	)
	return err
}

// Load retrieves a checkpoint for a specific world and year
func (p *PostgresCheckpointStore) Load(ctx context.Context, worldID uuid.UUID, year int64) (*Checkpoint, error) {
	query := `
		SELECT id, world_id, year, checkpoint_type, state_data, species_count, population_sum, created_at
		FROM world_checkpoints
		WHERE world_id = $1 AND year = $2
	`
	var c Checkpoint
	err := p.db.QueryRowContext(ctx, query, worldID, year).Scan(
		&c.ID, &c.WorldID, &c.Year, &c.Type, &c.StateData, &c.SpeciesCount, &c.PopulationSum, &c.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// LoadLatest retrieves the most recent checkpoint for a world
func (p *PostgresCheckpointStore) LoadLatest(ctx context.Context, worldID uuid.UUID) (*Checkpoint, error) {
	query := `
		SELECT id, world_id, year, checkpoint_type, state_data, species_count, population_sum, created_at
		FROM world_checkpoints
		WHERE world_id = $1
		ORDER BY year DESC
		LIMIT 1
	`
	var c Checkpoint
	err := p.db.QueryRowContext(ctx, query, worldID).Scan(
		&c.ID, &c.WorldID, &c.Year, &c.Type, &c.StateData, &c.SpeciesCount, &c.PopulationSum, &c.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// LoadNearestBefore retrieves the checkpoint closest to but not after the given year
func (p *PostgresCheckpointStore) LoadNearestBefore(ctx context.Context, worldID uuid.UUID, year int64) (*Checkpoint, error) {
	query := `
		SELECT id, world_id, year, checkpoint_type, state_data, species_count, population_sum, created_at
		FROM world_checkpoints
		WHERE world_id = $1 AND year <= $2
		ORDER BY year DESC
		LIMIT 1
	`
	var c Checkpoint
	err := p.db.QueryRowContext(ctx, query, worldID, year).Scan(
		&c.ID, &c.WorldID, &c.Year, &c.Type, &c.StateData, &c.SpeciesCount, &c.PopulationSum, &c.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// List retrieves all checkpoints for a world, ordered by year
func (p *PostgresCheckpointStore) List(ctx context.Context, worldID uuid.UUID) ([]Checkpoint, error) {
	query := `
		SELECT id, world_id, year, checkpoint_type, species_count, population_sum, created_at
		FROM world_checkpoints
		WHERE world_id = $1
		ORDER BY year ASC
	`
	rows, err := p.db.QueryContext(ctx, query, worldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checkpoints []Checkpoint
	for rows.Next() {
		var c Checkpoint
		if err := rows.Scan(&c.ID, &c.WorldID, &c.Year, &c.Type, &c.SpeciesCount, &c.PopulationSum, &c.CreatedAt); err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, c)
	}
	return checkpoints, rows.Err()
}

// Delete removes a checkpoint by ID
func (p *PostgresCheckpointStore) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM world_checkpoints WHERE id = $1`
	_, err := p.db.ExecContext(ctx, query, id)
	return err
}

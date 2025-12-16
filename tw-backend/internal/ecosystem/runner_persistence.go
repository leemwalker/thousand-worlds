// Package ecosystem provides persistence for the simulation runner state.
package ecosystem

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// PersistedRunnerState represents the saved state of a runner
type PersistedRunnerState struct {
	WorldID     uuid.UUID       `json:"world_id"`
	CurrentYear int64           `json:"current_year"`
	Speed       SimulationSpeed `json:"speed"`
	State       RunnerState     `json:"state"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// RunnerStateRepository handles persistence of runner state
type RunnerStateRepository struct {
	db *sql.DB
}

// NewRunnerStateRepository creates a new repository
func NewRunnerStateRepository(db *sql.DB) *RunnerStateRepository {
	return &RunnerStateRepository{db: db}
}

// Save persists the runner state to the database
func (r *RunnerStateRepository) Save(ctx context.Context, state *PersistedRunnerState) error {
	query := `
		INSERT INTO world_runner_state (world_id, current_year, speed, state, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (world_id) DO UPDATE SET
			current_year = EXCLUDED.current_year,
			speed = EXCLUDED.speed,
			state = EXCLUDED.state,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		state.WorldID,
		state.CurrentYear,
		int(state.Speed),
		string(state.State),
		time.Now(),
	)
	return err
}

// Load retrieves the runner state from the database
func (r *RunnerStateRepository) Load(ctx context.Context, worldID uuid.UUID) (*PersistedRunnerState, error) {
	query := `
		SELECT world_id, current_year, speed, state, updated_at
		FROM world_runner_state
		WHERE world_id = $1
	`
	var state PersistedRunnerState
	var speedInt int
	var stateStr string
	err := r.db.QueryRowContext(ctx, query, worldID).Scan(
		&state.WorldID,
		&state.CurrentYear,
		&speedInt,
		&stateStr,
		&state.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	state.Speed = SimulationSpeed(speedInt)
	state.State = RunnerState(stateStr)
	return &state, nil
}

// Delete removes the runner state from the database
func (r *RunnerStateRepository) Delete(ctx context.Context, worldID uuid.UUID) error {
	query := `DELETE FROM world_runner_state WHERE world_id = $1`
	_, err := r.db.ExecContext(ctx, query, worldID)
	return err
}

// SaveState saves the runner's current state to the database
func (sr *SimulationRunner) SaveState(repo *RunnerStateRepository) error {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	state := &PersistedRunnerState{
		WorldID:     sr.config.WorldID,
		CurrentYear: sr.currentYear,
		Speed:       sr.config.Speed,
		State:       sr.state,
		UpdatedAt:   time.Now(),
	}
	return repo.Save(context.Background(), state)
}

// RestoreFromState restores the runner from a persisted state
func (sr *SimulationRunner) RestoreFromState(state *PersistedRunnerState) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.currentYear = state.CurrentYear
	sr.config.Speed = state.Speed
	// Don't restore running state - require explicit start
	if state.State == RunnerRunning {
		sr.state = RunnerPaused
	} else {
		sr.state = state.State
	}
}

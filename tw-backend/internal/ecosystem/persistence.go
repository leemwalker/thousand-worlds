package ecosystem

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"tw-backend/internal/ecosystem/population"

	"github.com/google/uuid"
)

// SimulationSnapshotRepository handles persistence of full simulation snapshots
type SimulationSnapshotRepository struct {
	db *sql.DB
}

// NewSimulationSnapshotRepository creates a new repository
func NewSimulationSnapshotRepository(db *sql.DB) *SimulationSnapshotRepository {
	return &SimulationSnapshotRepository{db: db}
}

// SaveSnapshot perists the full population simulator state
func (r *SimulationSnapshotRepository) SaveSnapshot(ctx context.Context, worldID uuid.UUID, sim *population.PopulationSimulator) error {
	data, err := json.Marshal(sim)
	if err != nil {
		return fmt.Errorf("failed to marshal simulation state: %w", err)
	}

	query := `
		INSERT INTO world_simulation_snapshot (world_id, data, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (world_id) DO UPDATE SET
			data = EXCLUDED.data,
			updated_at = EXCLUDED.updated_at
	`
	_, err = r.db.ExecContext(ctx, query, worldID, data, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save simulation snapshot: %w", err)
	}
	return nil
}

// LoadSnapshot retrieves the full population simulator state
func (r *SimulationSnapshotRepository) LoadSnapshot(ctx context.Context, worldID uuid.UUID) (*population.PopulationSimulator, error) {
	query := `
		SELECT data
		FROM world_simulation_snapshot
		WHERE world_id = $1
	`
	var data []byte
	err := r.db.QueryRowContext(ctx, query, worldID).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil // No snapshot exists
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load simulation snapshot: %w", err)
	}

	sim := &population.PopulationSimulator{}
	if err := json.Unmarshal(data, sim); err != nil {
		return nil, fmt.Errorf("failed to unmarshal simulation state: %w", err)
	}

	// Important: We must re-seed the RNG and re-initialize systems that weren't serialized
	// This should be done by the caller (sim runner) because it knows the seed logic

	return sim, nil
}

// DeleteSnapshot removes the snapshot
func (r *SimulationSnapshotRepository) DeleteSnapshot(ctx context.Context, worldID uuid.UUID) error {
	query := `DELETE FROM world_simulation_snapshot WHERE world_id = $1`
	_, err := r.db.ExecContext(ctx, query, worldID)
	return err
}

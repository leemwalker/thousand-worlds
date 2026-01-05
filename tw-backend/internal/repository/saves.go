package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrSaveNotFound is returned when a requested save doesn't exist.
var ErrSaveNotFound = errors.New("save not found")

// WorldSave represents a player's world save with snapshot reference.
type WorldSave struct {
	ID            uuid.UUID      `json:"id"`
	WorldID       uuid.UUID      `json:"world_id"`
	PlayerID      *uuid.UUID     `json:"player_id,omitempty"`
	SnapshotKey   string         `json:"snapshot_key"`   // MinIO object key
	EventSequence uint64         `json:"event_sequence"` // NATS JetStream sequence
	Year          int64          `json:"year"`           // Simulation year at save
	CreatedAt     time.Time      `json:"created_at"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// SaveRepository handles persistence of world saves.
type SaveRepository interface {
	CreateSave(ctx context.Context, save *WorldSave) error
	GetLatestSave(ctx context.Context, worldID uuid.UUID) (*WorldSave, error)
	GetSavesForWorld(ctx context.Context, worldID uuid.UUID, limit int) ([]*WorldSave, error)
	GetSavesForPlayer(ctx context.Context, playerID uuid.UUID, limit int) ([]*WorldSave, error)
	DeleteSave(ctx context.Context, saveID uuid.UUID) error
	DeleteOldSaves(ctx context.Context, worldID uuid.UUID, keepCount int) (int64, error)
}

// PostgresSaveRepository implements SaveRepository using PostgreSQL.
type PostgresSaveRepository struct {
	db *sql.DB
}

// NewPostgresSaveRepository creates a new PostgreSQL-backed save repository.
func NewPostgresSaveRepository(db *sql.DB) *PostgresSaveRepository {
	return &PostgresSaveRepository{db: db}
}

// CreateSave inserts a new world save.
func (r *PostgresSaveRepository) CreateSave(ctx context.Context, save *WorldSave) error {
	if save.ID == uuid.Nil {
		save.ID = uuid.New()
	}
	if save.CreatedAt.IsZero() {
		save.CreatedAt = time.Now()
	}

	metadataJSON, err := json.Marshal(save.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO world_saves (id, world_id, player_id, snapshot_key, event_sequence, year, created_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.ExecContext(ctx, query,
		save.ID,
		save.WorldID,
		save.PlayerID,
		save.SnapshotKey,
		save.EventSequence,
		save.Year,
		save.CreatedAt,
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("insert save: %w", err)
	}

	return nil
}

// GetLatestSave returns the most recent save for a world.
func (r *PostgresSaveRepository) GetLatestSave(ctx context.Context, worldID uuid.UUID) (*WorldSave, error) {
	query := `
		SELECT id, world_id, player_id, snapshot_key, event_sequence, year, created_at, metadata
		FROM world_saves
		WHERE world_id = $1
		ORDER BY year DESC, created_at DESC
		LIMIT 1
	`

	save := &WorldSave{}
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, worldID).Scan(
		&save.ID,
		&save.WorldID,
		&save.PlayerID,
		&save.SnapshotKey,
		&save.EventSequence,
		&save.Year,
		&save.CreatedAt,
		&metadataJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSaveNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan save: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &save.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	return save, nil
}

// GetSavesForWorld returns saves for a world, ordered by year descending.
func (r *PostgresSaveRepository) GetSavesForWorld(ctx context.Context, worldID uuid.UUID, limit int) ([]*WorldSave, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, world_id, player_id, snapshot_key, event_sequence, year, created_at, metadata
		FROM world_saves
		WHERE world_id = $1
		ORDER BY year DESC, created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, worldID, limit)
	if err != nil {
		return nil, fmt.Errorf("query saves: %w", err)
	}
	defer rows.Close()

	return r.scanSaves(rows)
}

// GetSavesForPlayer returns saves for a player, ordered by creation time descending.
func (r *PostgresSaveRepository) GetSavesForPlayer(ctx context.Context, playerID uuid.UUID, limit int) ([]*WorldSave, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, world_id, player_id, snapshot_key, event_sequence, year, created_at, metadata
		FROM world_saves
		WHERE player_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("query saves: %w", err)
	}
	defer rows.Close()

	return r.scanSaves(rows)
}

// DeleteSave removes a save by ID.
func (r *PostgresSaveRepository) DeleteSave(ctx context.Context, saveID uuid.UUID) error {
	query := `DELETE FROM world_saves WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, saveID)
	if err != nil {
		return fmt.Errorf("delete save: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrSaveNotFound
	}
	return nil
}

// DeleteOldSaves removes old saves for a world, keeping the most recent keepCount.
// Returns the number of saves deleted.
func (r *PostgresSaveRepository) DeleteOldSaves(ctx context.Context, worldID uuid.UUID, keepCount int) (int64, error) {
	query := `
		DELETE FROM world_saves
		WHERE world_id = $1
		AND id NOT IN (
			SELECT id FROM world_saves
			WHERE world_id = $1
			ORDER BY year DESC, created_at DESC
			LIMIT $2
		)
	`

	result, err := r.db.ExecContext(ctx, query, worldID, keepCount)
	if err != nil {
		return 0, fmt.Errorf("delete old saves: %w", err)
	}

	return result.RowsAffected()
}

func (r *PostgresSaveRepository) scanSaves(rows *sql.Rows) ([]*WorldSave, error) {
	var saves []*WorldSave

	for rows.Next() {
		save := &WorldSave{}
		var metadataJSON []byte

		if err := rows.Scan(
			&save.ID,
			&save.WorldID,
			&save.PlayerID,
			&save.SnapshotKey,
			&save.EventSequence,
			&save.Year,
			&save.CreatedAt,
			&metadataJSON,
		); err != nil {
			return nil, fmt.Errorf("scan save: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &save.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}

		saves = append(saves, save)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return saves, nil
}

// Verify PostgresSaveRepository implements SaveRepository at compile time
var _ SaveRepository = (*PostgresSaveRepository)(nil)

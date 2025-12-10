package worldentity

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for world entities
type Repository interface {
	Create(ctx context.Context, entity *WorldEntity) error
	GetByID(ctx context.Context, id uuid.UUID) (*WorldEntity, error)
	GetByWorldID(ctx context.Context, worldID uuid.UUID) ([]*WorldEntity, error)
	GetByWorldAndType(ctx context.Context, worldID uuid.UUID, entityType EntityType) ([]*WorldEntity, error)
	GetAtPosition(ctx context.Context, worldID uuid.UUID, x, y, radius float64) ([]*WorldEntity, error)
	GetByName(ctx context.Context, worldID uuid.UUID, name string) (*WorldEntity, error)
	Update(ctx context.Context, entity *WorldEntity) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgresRepository
func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create inserts a new world entity
func (r *PostgresRepository) Create(ctx context.Context, entity *WorldEntity) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}

	query := `
		INSERT INTO world_entities (id, world_id, entity_type, name, description, details, x, y, z, collision, locked, interactable, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		entity.ID,
		entity.WorldID,
		entity.EntityType,
		entity.Name,
		entity.Description,
		entity.Details,
		entity.X,
		entity.Y,
		entity.Z,
		entity.Collision,
		entity.Locked,
		entity.Interactable,
		entity.Metadata,
	).Scan(&entity.CreatedAt, &entity.UpdatedAt)
}

// GetByID retrieves an entity by its ID
func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*WorldEntity, error) {
	query := `
		SELECT id, world_id, entity_type, name, description, details, x, y, z, collision, locked, interactable, metadata, created_at, updated_at
		FROM world_entities
		WHERE id = $1
	`

	entity := &WorldEntity{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&entity.ID,
		&entity.WorldID,
		&entity.EntityType,
		&entity.Name,
		&entity.Description,
		&entity.Details,
		&entity.X,
		&entity.Y,
		&entity.Z,
		&entity.Collision,
		&entity.Locked,
		&entity.Interactable,
		&entity.Metadata,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// GetByWorldID retrieves all entities in a world
func (r *PostgresRepository) GetByWorldID(ctx context.Context, worldID uuid.UUID) ([]*WorldEntity, error) {
	query := `
		SELECT id, world_id, entity_type, name, description, details, x, y, z, collision, locked, interactable, metadata, created_at, updated_at
		FROM world_entities
		WHERE world_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, worldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*WorldEntity
	for rows.Next() {
		entity := &WorldEntity{}
		err := rows.Scan(
			&entity.ID,
			&entity.WorldID,
			&entity.EntityType,
			&entity.Name,
			&entity.Description,
			&entity.Details,
			&entity.X,
			&entity.Y,
			&entity.Z,
			&entity.Collision,
			&entity.Locked,
			&entity.Interactable,
			&entity.Metadata,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, nil
}

// GetByWorldAndType retrieves entities of a specific type in a world
func (r *PostgresRepository) GetByWorldAndType(ctx context.Context, worldID uuid.UUID, entityType EntityType) ([]*WorldEntity, error) {
	query := `
		SELECT id, world_id, entity_type, name, description, details, x, y, z, collision, locked, interactable, metadata, created_at, updated_at
		FROM world_entities
		WHERE world_id = $1 AND entity_type = $2
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, worldID, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*WorldEntity
	for rows.Next() {
		entity := &WorldEntity{}
		err := rows.Scan(
			&entity.ID,
			&entity.WorldID,
			&entity.EntityType,
			&entity.Name,
			&entity.Description,
			&entity.Details,
			&entity.X,
			&entity.Y,
			&entity.Z,
			&entity.Collision,
			&entity.Locked,
			&entity.Interactable,
			&entity.Metadata,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, nil
}

// GetAtPosition retrieves entities within a radius of a position
func (r *PostgresRepository) GetAtPosition(ctx context.Context, worldID uuid.UUID, x, y, radius float64) ([]*WorldEntity, error) {
	// Use bounding box for initial filter, then calculate exact distance
	query := `
		SELECT id, world_id, entity_type, name, description, details, x, y, z, collision, locked, interactable, metadata, created_at, updated_at
		FROM world_entities
		WHERE world_id = $1
		  AND x BETWEEN $2 - $4 AND $2 + $4
		  AND y BETWEEN $3 - $4 AND $3 + $4
		  AND SQRT(POWER(x - $2, 2) + POWER(y - $3, 2)) <= $4
		ORDER BY SQRT(POWER(x - $2, 2) + POWER(y - $3, 2))
	`

	rows, err := r.db.Query(ctx, query, worldID, x, y, radius)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*WorldEntity
	for rows.Next() {
		entity := &WorldEntity{}
		err := rows.Scan(
			&entity.ID,
			&entity.WorldID,
			&entity.EntityType,
			&entity.Name,
			&entity.Description,
			&entity.Details,
			&entity.X,
			&entity.Y,
			&entity.Z,
			&entity.Collision,
			&entity.Locked,
			&entity.Interactable,
			&entity.Metadata,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, nil
}

// GetByName retrieves an entity by name in a world (case-insensitive)
func (r *PostgresRepository) GetByName(ctx context.Context, worldID uuid.UUID, name string) (*WorldEntity, error) {
	query := `
		SELECT id, world_id, entity_type, name, description, details, x, y, z, collision, locked, interactable, metadata, created_at, updated_at
		FROM world_entities
		WHERE world_id = $1 AND LOWER(name) = LOWER($2)
		LIMIT 1
	`

	entity := &WorldEntity{}
	err := r.db.QueryRow(ctx, query, worldID, name).Scan(
		&entity.ID,
		&entity.WorldID,
		&entity.EntityType,
		&entity.Name,
		&entity.Description,
		&entity.Details,
		&entity.X,
		&entity.Y,
		&entity.Z,
		&entity.Collision,
		&entity.Locked,
		&entity.Interactable,
		&entity.Metadata,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("entity '%s' not found: %w", name, err)
	}
	return entity, nil
}

// Update updates an existing world entity
func (r *PostgresRepository) Update(ctx context.Context, entity *WorldEntity) error {
	query := `
		UPDATE world_entities
		SET name = $2, description = $3, details = $4, x = $5, y = $6, z = $7,
		    collision = $8, locked = $9, interactable = $10, metadata = $11, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		entity.ID,
		entity.Name,
		entity.Description,
		entity.Details,
		entity.X,
		entity.Y,
		entity.Z,
		entity.Collision,
		entity.Locked,
		entity.Interactable,
		entity.Metadata,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("entity not found: %s", entity.ID)
	}
	return nil
}

// Delete removes a world entity
func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM world_entities WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("entity not found: %s", id)
	}
	return nil
}

// Ensure PostgresRepository implements Repository
var _ Repository = (*PostgresRepository)(nil)

// Helper to check if string contains substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

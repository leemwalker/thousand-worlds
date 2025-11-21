package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Entity represents a spatial entity.
type Entity struct {
	ID       uuid.UUID
	WorldID  uuid.UUID
	X, Y, Z  float64
	Metadata map[string]interface{}
}

// SpatialRepository defines methods for spatial operations.
type SpatialRepository interface {
	CreateEntity(ctx context.Context, worldID, entityID uuid.UUID, x, y, z float64) error
	UpdateEntityLocation(ctx context.Context, entityID uuid.UUID, x, y, z float64) error
	GetEntity(ctx context.Context, entityID uuid.UUID) (*Entity, error)
	GetEntitiesNearby(ctx context.Context, worldID uuid.UUID, x, y, z, radius float64) ([]Entity, error)
	GetEntitiesInBounds(ctx context.Context, worldID uuid.UUID, minX, minY, maxX, maxY float64) ([]Entity, error)
	CalculateDistance(ctx context.Context, entity1ID, entity2ID uuid.UUID) (float64, error)
}

// PostgresSpatialRepository implements SpatialRepository using PostGIS.
type PostgresSpatialRepository struct {
	db *pgxpool.Pool
}

// NewPostgresSpatialRepository creates a new PostgresSpatialRepository.
func NewPostgresSpatialRepository(db *pgxpool.Pool) *PostgresSpatialRepository {
	return &PostgresSpatialRepository{db: db}
}

func (r *PostgresSpatialRepository) CreateEntity(ctx context.Context, worldID, entityID uuid.UUID, x, y, z float64) error {
	query := `
		INSERT INTO entities (id, world_id, position, metadata)
		VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4, $5), 4326), '{}')
	`
	_, err := r.db.Exec(ctx, query, entityID, worldID, x, y, z)
	return err
}

func (r *PostgresSpatialRepository) UpdateEntityLocation(ctx context.Context, entityID uuid.UUID, x, y, z float64) error {
	query := `
		UPDATE entities
		SET position = ST_SetSRID(ST_MakePoint($2, $3, $4), 4326)
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, entityID, x, y, z)
	return err
}

func (r *PostgresSpatialRepository) GetEntity(ctx context.Context, entityID uuid.UUID) (*Entity, error) {
	query := `
		SELECT id, world_id, ST_X(position::geometry), ST_Y(position::geometry), ST_Z(position::geometry), metadata
		FROM entities
		WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, entityID)

	var e Entity
	err := row.Scan(&e.ID, &e.WorldID, &e.X, &e.Y, &e.Z, &e.Metadata)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *PostgresSpatialRepository) GetEntitiesNearby(ctx context.Context, worldID uuid.UUID, x, y, z, radius float64) ([]Entity, error) {
	// ST_DWithin with GEOGRAPHY uses meters on sphere surface
	query := `
		SELECT id, world_id, ST_X(position::geometry), ST_Y(position::geometry), ST_Z(position::geometry), metadata
		FROM entities
		WHERE world_id = $1
		AND ST_DWithin(position, ST_SetSRID(ST_MakePoint($2, $3, $4), 4326), $5)
	`
	rows, err := r.db.Query(ctx, query, worldID, x, y, z, radius)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.WorldID, &e.X, &e.Y, &e.Z, &e.Metadata); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, nil
}

func (r *PostgresSpatialRepository) GetEntitiesInBounds(ctx context.Context, worldID uuid.UUID, minX, minY, maxX, maxY float64) ([]Entity, error) {
	// Using bounding box for spherical geography (lon/lat bounds)
	query := `
		SELECT id, world_id, ST_X(position::geometry), ST_Y(position::geometry), ST_Z(position::geometry), metadata
		FROM entities
		WHERE world_id = $1
		AND position::geometry && ST_MakeEnvelope($2, $3, $4, $5, 4326)
	`
	rows, err := r.db.Query(ctx, query, worldID, minX, minY, maxX, maxY)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.WorldID, &e.X, &e.Y, &e.Z, &e.Metadata); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, nil
}

func (r *PostgresSpatialRepository) CalculateDistance(ctx context.Context, entity1ID, entity2ID uuid.UUID) (float64, error) {
	query := `
		SELECT ST_Distance(e1.position, e2.position)
		FROM entities e1, entities e2
		WHERE e1.id = $1 AND e2.id = $2
	`
	var distance float64
	err := r.db.QueryRow(ctx, query, entity1ID, entity2ID).Scan(&distance)
	return distance, err
}

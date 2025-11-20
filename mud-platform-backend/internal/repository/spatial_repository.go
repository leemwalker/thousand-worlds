package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SpatialRepository struct {
	db *pgxpool.Pool
}

func NewSpatialRepository(db *pgxpool.Pool) *SpatialRepository {
	return &SpatialRepository{
		db: db,
	}
}

// GetEntitiesInRadius retrieves entity IDs within a given radius (in meters) of a 3D coordinate.
// It uses PostGIS ST_3DDWithin function.
func (r *SpatialRepository) GetEntitiesInRadius(ctx context.Context, worldID string, x, y, z float64, radius float64) ([]string, error) {
	query := `
		SELECT id 
		FROM entities 
		WHERE world_id = $1 
		AND ST_3DDWithin(location, ST_MakePoint($2, $3, $4), $5)
	`

	rows, err := r.db.Query(ctx, query, worldID, x, y, z, radius)
	if err != nil {
		return nil, fmt.Errorf("spatial.GetEntitiesInRadius: query failed: %w", err)
	}
	defer rows.Close()

	var entityIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("spatial.GetEntitiesInRadius: scan failed: %w", err)
		}
		entityIDs = append(entityIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("spatial.GetEntitiesInRadius: rows iteration failed: %w", err)
	}

	return entityIDs, nil
}

// UpdateEntityLocation updates the 3D location of an entity in the database.
func (r *SpatialRepository) UpdateEntityLocation(ctx context.Context, entityID string, worldID string, x, y, z float64) error {
	query := `
		UPDATE entities 
		SET location = ST_MakePoint($1, $2, $3) 
		WHERE id = $4 AND world_id = $5
	`
	commandTag, err := r.db.Exec(ctx, query, x, y, z, entityID, worldID)
	if err != nil {
		return fmt.Errorf("spatial.UpdateEntityLocation: exec failed: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("spatial.UpdateEntityLocation: no rows affected (entity not found?)")
	}

	return nil
}

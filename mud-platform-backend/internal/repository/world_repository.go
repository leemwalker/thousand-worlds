package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorldShape represents the type of a world.
type WorldShape string

const (
	WorldShapeSphere   WorldShape = "sphere"
	WorldShapeCube     WorldShape = "cube"
	WorldShapeInfinite WorldShape = "infinite"
)

// World represents a game world with its spatial properties.
type World struct {
	ID        uuid.UUID
	Name      string
	Shape     WorldShape
	Radius    *float64 // for sphere worlds (meters)
	BoundsMin *Vector3 // for cube worlds
	BoundsMax *Vector3 // for cube worlds
	Metadata  map[string]interface{}
	CreatedAt time.Time
}

// Vector3 represents a 3D vector.
type Vector3 struct {
	X, Y, Z float64
}

// WorldRepository defines methods for world management.
type WorldRepository interface {
	CreateWorld(ctx context.Context, world *World) error
	GetWorld(ctx context.Context, worldID uuid.UUID) (*World, error)
	ListWorlds(ctx context.Context) ([]World, error)
	UpdateWorld(ctx context.Context, world *World) error
	DeleteWorld(ctx context.Context, worldID uuid.UUID) error
}

// PostgresWorldRepository implements WorldRepository using PostgreSQL.
type PostgresWorldRepository struct {
	db *pgxpool.Pool
}

// NewPostgresWorldRepository creates a new PostgresWorldRepository.
func NewPostgresWorldRepository(db *pgxpool.Pool) *PostgresWorldRepository {
	return &PostgresWorldRepository{db: db}
}

func (r *PostgresWorldRepository) CreateWorld(ctx context.Context, world *World) error {
	query := `
		INSERT INTO worlds (id, name, shape, radius, bounds_min_x, bounds_min_y, bounds_min_z, bounds_max_x, bounds_max_y, bounds_max_z, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	var boundsMinX, boundsMinY, boundsMinZ, boundsMaxX, boundsMaxY, boundsMaxZ *float64
	if world.BoundsMin != nil {
		boundsMinX = &world.BoundsMin.X
		boundsMinY = &world.BoundsMin.Y
		boundsMinZ = &world.BoundsMin.Z
	}
	if world.BoundsMax != nil {
		boundsMaxX = &world.BoundsMax.X
		boundsMaxY = &world.BoundsMax.Y
		boundsMaxZ = &world.BoundsMax.Z
	}

	_, err := r.db.Exec(ctx, query,
		world.ID, world.Name, world.Shape, world.Radius,
		boundsMinX, boundsMinY, boundsMinZ,
		boundsMaxX, boundsMaxY, boundsMaxZ,
		world.Metadata,
	)
	return err
}

func (r *PostgresWorldRepository) GetWorld(ctx context.Context, worldID uuid.UUID) (*World, error) {
	query := `
		SELECT id, name, shape, radius, bounds_min_x, bounds_min_y, bounds_min_z, bounds_max_x, bounds_max_y, bounds_max_z, metadata, created_at
		FROM worlds
		WHERE id = $1
	`

	var world World
	var boundsMinX, boundsMinY, boundsMinZ, boundsMaxX, boundsMaxY, boundsMaxZ *float64

	err := r.db.QueryRow(ctx, query, worldID).Scan(
		&world.ID, &world.Name, &world.Shape, &world.Radius,
		&boundsMinX, &boundsMinY, &boundsMinZ,
		&boundsMaxX, &boundsMaxY, &boundsMaxZ,
		&world.Metadata, &world.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Reconstruct Vector3 pointers from components
	if boundsMinX != nil {
		world.BoundsMin = &Vector3{X: *boundsMinX, Y: *boundsMinY, Z: *boundsMinZ}
	}
	if boundsMaxX != nil {
		world.BoundsMax = &Vector3{X: *boundsMaxX, Y: *boundsMaxY, Z: *boundsMaxZ}
	}

	return &world, nil
}

func (r *PostgresWorldRepository) ListWorlds(ctx context.Context) ([]World, error) {
	query := `
		SELECT id, name, shape, radius, bounds_min_x, bounds_min_y, bounds_min_z, bounds_max_x, bounds_max_y, bounds_max_z, metadata, created_at
		FROM worlds
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var worlds []World
	for rows.Next() {
		var world World
		var boundsMinX, boundsMinY, boundsMinZ, boundsMaxX, boundsMaxY, boundsMaxZ *float64

		err := rows.Scan(
			&world.ID, &world.Name, &world.Shape, &world.Radius,
			&boundsMinX, &boundsMinY, &boundsMinZ,
			&boundsMaxX, &boundsMaxY, &boundsMaxZ,
			&world.Metadata, &world.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if boundsMinX != nil {
			world.BoundsMin = &Vector3{X: *boundsMinX, Y: *boundsMinY, Z: *boundsMinZ}
		}
		if boundsMaxX != nil {
			world.BoundsMax = &Vector3{X: *boundsMaxX, Y: *boundsMaxY, Z: *boundsMaxZ}
		}

		worlds = append(worlds, world)
	}

	return worlds, nil
}

func (r *PostgresWorldRepository) UpdateWorld(ctx context.Context, world *World) error {
	query := `
		UPDATE worlds
		SET name = $2, shape = $3, radius = $4, 
		    bounds_min_x = $5, bounds_min_y = $6, bounds_min_z = $7,
		    bounds_max_x = $8, bounds_max_y = $9, bounds_max_z = $10,
		    metadata = $11
		WHERE id = $1
	`

	var boundsMinX, boundsMinY, boundsMinZ, boundsMaxX, boundsMaxY, boundsMaxZ *float64
	if world.BoundsMin != nil {
		boundsMinX = &world.BoundsMin.X
		boundsMinY = &world.BoundsMin.Y
		boundsMinZ = &world.BoundsMin.Z
	}
	if world.BoundsMax != nil {
		boundsMaxX = &world.BoundsMax.X
		boundsMaxY = &world.BoundsMax.Y
		boundsMaxZ = &world.BoundsMax.Z
	}

	_, err := r.db.Exec(ctx, query,
		world.ID, world.Name, world.Shape, world.Radius,
		boundsMinX, boundsMinY, boundsMinZ,
		boundsMaxX, boundsMaxY, boundsMaxZ,
		world.Metadata,
	)
	return err
}

func (r *PostgresWorldRepository) DeleteWorld(ctx context.Context, worldID uuid.UUID) error {
	query := `DELETE FROM worlds WHERE id = $1`
	_, err := r.db.Exec(ctx, query, worldID)
	return err
}

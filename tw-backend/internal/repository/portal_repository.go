package repository

import (
	"context"

	"mud-platform-backend/internal/world/spatial"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PortalRepository defines methods for interacting with portals
type PortalRepository interface {
	CreatePortal(ctx context.Context, portal *spatial.Portal) error
	GetPortalsByWorldID(ctx context.Context, worldID uuid.UUID) ([]spatial.Portal, error)
	GetPortal(ctx context.Context, portalID uuid.UUID) (*spatial.Portal, error)
	DeletePortal(ctx context.Context, portalID uuid.UUID) error
}

// PostgresPortalRepository implements PortalRepository
type PostgresPortalRepository struct {
	db *pgxpool.Pool
}

func NewPostgresPortalRepository(db *pgxpool.Pool) *PostgresPortalRepository {
	return &PostgresPortalRepository{db: db}
}

func (r *PostgresPortalRepository) CreatePortal(ctx context.Context, portal *spatial.Portal) error {
	query := `
		INSERT INTO portals (portal_id, world_id, location_x, location_y, side, description)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		portal.ID, portal.WorldID, portal.LocationX, portal.LocationY, portal.Side, portal.Description,
	)
	return err
}

func (r *PostgresPortalRepository) GetPortalsByWorldID(ctx context.Context, worldID uuid.UUID) ([]spatial.Portal, error) {
	query := `
		SELECT portal_id, world_id, location_x, location_y, side, description
		FROM portals
		WHERE world_id = $1
	`
	rows, err := r.db.Query(ctx, query, worldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var portals []spatial.Portal
	for rows.Next() {
		var p spatial.Portal
		if err := rows.Scan(&p.ID, &p.WorldID, &p.LocationX, &p.LocationY, &p.Side, &p.Description); err != nil {
			return nil, err
		}
		portals = append(portals, p)
	}
	return portals, nil
}

func (r *PostgresPortalRepository) GetPortal(ctx context.Context, portalID uuid.UUID) (*spatial.Portal, error) {
	query := `
		SELECT portal_id, world_id, location_x, location_y, side, description
		FROM portals
		WHERE portal_id = $1
	`
	var p spatial.Portal
	err := r.db.QueryRow(ctx, query, portalID).Scan(
		&p.ID, &p.WorldID, &p.LocationX, &p.LocationY, &p.Side, &p.Description,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresPortalRepository) DeletePortal(ctx context.Context, portalID uuid.UUID) error {
	query := `DELETE FROM portals WHERE portal_id = $1`
	_, err := r.db.Exec(ctx, query, portalID)
	return err
}

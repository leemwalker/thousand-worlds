package minerals

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// Repository defines the interface for mineral persistence
type Repository interface {
	SaveDeposit(ctx context.Context, deposit *MineralDeposit) error
	GetDeposit(ctx context.Context, id uuid.UUID) (*MineralDeposit, error)
	GetDepositsInRegion(ctx context.Context, minX, minY, maxX, maxY float64) ([]*MineralDeposit, error)
	UpdateDepletion(ctx context.Context, history *DepletionHistory) error
}

// PostgresRepository implements Repository for PostgreSQL
type PostgresRepository struct {
	DB *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{DB: db}
}

func (r *PostgresRepository) SaveDeposit(ctx context.Context, d *MineralDeposit) error {
	query := `
		INSERT INTO mineral_deposits (
			deposit_id, mineral_type, formation_type, 
			location_x, location_y, depth, 
			quantity, concentration, vein_size, geological_age,
			vein_shape, vein_orientation_x, vein_orientation_y, 
			vein_length, vein_width, surface_visible, required_depth
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	_, err := r.DB.ExecContext(ctx, query,
		d.DepositID, d.MineralType.Name, d.FormationType,
		d.Location.X, d.Location.Y, d.Depth,
		d.Quantity, d.Concentration, d.VeinSize, d.GeologicalAge,
		d.VeinShape, d.VeinOrientation.X, d.VeinOrientation.Y,
		d.VeinLength, d.VeinWidth, d.SurfaceVisible, d.RequiredDepth,
	)
	return err
}

func (r *PostgresRepository) GetDeposit(ctx context.Context, id uuid.UUID) (*MineralDeposit, error) {
	// Implementation omitted for brevity, but would select and scan
	return nil, errors.New("not implemented")
}

func (r *PostgresRepository) GetDepositsInRegion(ctx context.Context, minX, minY, maxX, maxY float64) ([]*MineralDeposit, error) {
	// Implementation omitted for brevity
	return nil, errors.New("not implemented")
}

func (r *PostgresRepository) UpdateDepletion(ctx context.Context, h *DepletionHistory) error {
	query := `
		INSERT INTO mineral_depletion (
			deposit_id, original_quantity, current_quantity, 
			first_extracted, depleted_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (deposit_id) DO UPDATE SET
			current_quantity = EXCLUDED.current_quantity,
			depleted_at = EXCLUDED.depleted_at
	`
	_, err := r.DB.ExecContext(ctx, query,
		h.DepositID, h.OriginalQuantity, h.CurrentQuantity,
		h.FirstExtracted, h.DepletedAt,
	)
	return err
}

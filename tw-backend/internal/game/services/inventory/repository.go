package inventory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles inventory persistence
type Repository interface {
	AddItem(ctx context.Context, charID uuid.UUID, itemID uuid.UUID, quantity int, metadata map[string]interface{}) error
	RemoveItem(ctx context.Context, charID uuid.UUID, itemID uuid.UUID, quantity int) error
	GetInventory(ctx context.Context, charID uuid.UUID) ([]InventoryItem, error)
}

// InventoryItem represents an item in an inventory
type InventoryItem struct {
	ID          uuid.UUID              `json:"id"`
	CharacterID uuid.UUID              `json:"character_id"`
	ItemID      uuid.UUID              `json:"item_id"`
	Quantity    int                    `json:"quantity"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`

	// Enriched data
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) AddItem(ctx context.Context, charID uuid.UUID, itemID uuid.UUID, quantity int, metadata map[string]interface{}) error {
	query := `
		INSERT INTO character_inventory (id, character_id, item_id, quantity, metadata)
		VALUES ($1, $2, $3, $4, $5)
	`
	// Use passed metadata directly
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	_, err := r.db.Exec(ctx, query,
		uuid.New(),
		charID,
		itemID,
		quantity,
		metadata,
	)
	return err
}

func (r *PostgresRepository) RemoveItem(ctx context.Context, charID uuid.UUID, itemID uuid.UUID, quantity int) error {
	// Simple removal: delete row if exists.
	// Future: decrement quantity and delete if <= 0
	query := `
		DELETE FROM character_inventory 
		WHERE character_id = $1 AND item_id = $2
	`
	// Note: limits removal to fully matching item_id
	_, err := r.db.Exec(ctx, query, charID, itemID)
	return err
}

func (r *PostgresRepository) GetInventory(ctx context.Context, charID uuid.UUID) ([]InventoryItem, error) {
	query := `
		SELECT id, character_id, item_id, quantity, metadata, created_at, updated_at
		FROM character_inventory
		WHERE character_id = $1
	`
	rows, err := r.db.Query(ctx, query, charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []InventoryItem
	for rows.Next() {
		var i InventoryItem
		// Metadata fetch might need specific scanning for map[string]interface{}
		// pgx usually handles JSONB to map
		err := rows.Scan(
			&i.ID,
			&i.CharacterID,
			&i.ItemID,
			&i.Quantity,
			&i.Metadata,
			&i.CreatedAt,
			&i.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Enrich from metadata
		if name, ok := i.Metadata["name"].(string); ok {
			i.Name = name
		}
		if desc, ok := i.Metadata["description"].(string); ok {
			i.Description = desc
		}

		items = append(items, i)
	}
	return items, nil
}

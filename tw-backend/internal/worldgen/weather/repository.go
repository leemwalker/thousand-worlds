package weather

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// Repository defines the interface for weather persistence
type Repository interface {
	SaveWeatherState(ctx context.Context, state *WeatherState) error
	GetWeatherState(ctx context.Context, cellID uuid.UUID, timestamp int64) (*WeatherState, error)
	GetWeatherHistory(ctx context.Context, cellID uuid.UUID, days int) ([]*WeatherState, error)
	GetAnnualPrecipitation(ctx context.Context, cellID uuid.UUID, year int) (float64, error)
}

// PostgresRepository implements Repository for PostgreSQL
type PostgresRepository struct {
	DB *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{DB: db}
}

func (r *PostgresRepository) SaveWeatherState(ctx context.Context, state *WeatherState) error {
	query := `
		INSERT INTO weather_states (
			state_id, cell_id, timestamp, state_type, 
			temperature, precipitation, wind_direction, wind_speed,
			humidity, visibility
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.DB.ExecContext(ctx, query,
		uuid.New(), state.CellID, state.Timestamp, state.State,
		state.Temperature, state.Precipitation, state.Wind.Direction, state.Wind.Speed,
		state.Humidity, state.Visibility,
	)
	return err
}

func (r *PostgresRepository) GetWeatherState(ctx context.Context, cellID uuid.UUID, timestamp int64) (*WeatherState, error) {
	// Implementation omitted for brevity
	return nil, nil
}

func (r *PostgresRepository) GetWeatherHistory(ctx context.Context, cellID uuid.UUID, days int) ([]*WeatherState, error) {
	// Implementation omitted for brevity
	return nil, nil
}

func (r *PostgresRepository) GetAnnualPrecipitation(ctx context.Context, cellID uuid.UUID, year int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(precipitation), 0)
		FROM weather_states
		WHERE cell_id = $1 
		AND EXTRACT(YEAR FROM timestamp) = $2
	`

	var total float64
	err := r.DB.QueryRowContext(ctx, query, cellID, year).Scan(&total)
	return total, err
}

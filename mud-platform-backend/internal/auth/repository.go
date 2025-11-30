package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrCharacterNotFound = errors.New("character not found")
	ErrDuplicateEmail    = errors.New("email already exists")
)

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *PostgresRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (user_id, email, password_hash, created_at)
		VALUES ($1, LOWER($2), $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
	)

	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return ErrDuplicateEmail
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (r *PostgresRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	query := `
		SELECT user_id, email, password_hash, created_at, last_login
		FROM users
		WHERE user_id = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.LastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	return &user, err
}

// GetUserByEmail retrieves a user by email
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT user_id, email, password_hash, created_at, last_login
		FROM users
		WHERE email = LOWER($1)
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.UserID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.LastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	return &user, err
}

// UpdateUser updates an existing user
func (r *PostgresRepository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET email = LOWER($2), password_hash = $3, last_login = $4
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Email,
		user.PasswordHash,
		user.LastLogin,
	)

	return err
}

// CreateCharacter creates a new character
func (r *PostgresRepository) CreateCharacter(ctx context.Context, char *Character) error {
	query := `
		INSERT INTO characters (character_id, user_id, world_id, name, role, appearance, description, occupation, position, created_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::jsonb, $7, $8, ST_SetSRID(ST_MakePoint($9, $10), 4326), $11)
	`

	var lat, lon float64
	if char.Position != nil {
		lat = char.Position.Latitude
		lon = char.Position.Longitude
	}

	_, err := r.db.ExecContext(ctx, query,
		char.CharacterID,
		char.UserID,
		char.WorldID,
		char.Name,
		char.Role,
		char.Appearance,
		char.Description,
		char.Occupation,
		lon, lat, // PostGIS expects (longitude, latitude)
		char.CreatedAt,
	)

	return err
}

// GetCharacter retrieves a character by ID
func (r *PostgresRepository) GetCharacter(ctx context.Context, characterID uuid.UUID) (*Character, error) {
	query := `
		SELECT 
			c.character_id, c.user_id, c.world_id, c.name, 
			COALESCE(c.role, ''), COALESCE(c.appearance::text, ''), COALESCE(c.description, ''), COALESCE(c.occupation, ''),
			ST_Y(c.position::geometry) as latitude,
			ST_X(c.position::geometry) as longitude,
			c.created_at, c.last_played
		FROM characters c
		WHERE c.character_id = $1
	`

	var char Character
	var lat, lon sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, characterID).Scan(
		&char.CharacterID,
		&char.UserID,
		&char.WorldID,
		&char.Name,
		&char.Role,
		&char.Appearance,
		&char.Description,
		&char.Occupation,
		&lat,
		&lon,
		&char.CreatedAt,
		&char.LastPlayed,
	)

	if err == sql.ErrNoRows {
		return nil, ErrCharacterNotFound
	}
	if err != nil {
		return nil, err
	}

	if lat.Valid && lon.Valid {
		char.Position = &Position{
			Latitude:  lat.Float64,
			Longitude: lon.Float64,
		}
	}

	return &char, nil
}

// GetUserCharacters retrieves all characters for a user
func (r *PostgresRepository) GetUserCharacters(ctx context.Context, userID uuid.UUID) ([]*Character, error) {
	query := `
		SELECT 
			c.character_id, c.user_id, c.world_id, c.name,
			COALESCE(c.role, ''), COALESCE(c.appearance::text, ''), COALESCE(c.description, ''), COALESCE(c.occupation, ''),
			ST_Y(c.position::geometry) as latitude,
			ST_X(c.position::geometry) as longitude,
			c.created_at, c.last_played
		FROM characters c
		WHERE c.user_id = $1
		ORDER BY c.last_played DESC NULLS LAST, c.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var characters []*Character
	for rows.Next() {
		var char Character
		var lat, lon sql.NullFloat64

		err := rows.Scan(
			&char.CharacterID,
			&char.UserID,
			&char.WorldID,
			&char.Name,
			&char.Role,
			&char.Appearance,
			&char.Description,
			&char.Occupation,
			&lat,
			&lon,
			&char.CreatedAt,
			&char.LastPlayed,
		)
		if err != nil {
			return nil, err
		}

		if lat.Valid && lon.Valid {
			char.Position = &Position{
				Latitude:  lat.Float64,
				Longitude: lon.Float64,
			}
		}

		characters = append(characters, &char)
	}

	return characters, rows.Err()
}

// GetCharacterByUserAndWorld retrieves a character by user and world
func (r *PostgresRepository) GetCharacterByUserAndWorld(ctx context.Context, userID, worldID uuid.UUID) (*Character, error) {
	query := `
		SELECT 
			c.character_id, c.user_id, c.world_id, c.name,
			COALESCE(c.role, ''), COALESCE(c.appearance::text, ''), COALESCE(c.description, ''), COALESCE(c.occupation, ''),
			ST_Y(c.position::geometry) as latitude,
			ST_X(c.position::geometry) as longitude,
			c.created_at, c.last_played
		FROM characters c
		WHERE c.user_id = $1 AND c.world_id = $2
	`

	var char Character
	var lat, lon sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, userID, worldID).Scan(
		&char.CharacterID,
		&char.UserID,
		&char.WorldID,
		&char.Name,
		&char.Role,
		&char.Appearance,
		&char.Description,
		&char.Occupation,
		&lat,
		&lon,
		&char.CreatedAt,
		&char.LastPlayed,
	)

	if err == sql.ErrNoRows {
		return nil, ErrCharacterNotFound
	}
	if err != nil {
		return nil, err
	}

	if lat.Valid && lon.Valid {
		char.Position = &Position{
			Latitude:  lat.Float64,
			Longitude: lon.Float64,
		}
	}

	return &char, nil
}

// UpdateCharacter updates an existing character
func (r *PostgresRepository) UpdateCharacter(ctx context.Context, char *Character) error {
	query := `
		UPDATE characters
		SET name = $2, position = ST_SetSRID(ST_MakePoint($3, $4), 4326), last_played = $5
		WHERE character_id = $1
	`

	var lat, lon float64
	if char.Position != nil {
		lat = char.Position.Latitude
		lon = char.Position.Longitude
	}

	_, err := r.db.ExecContext(ctx, query,
		char.CharacterID,
		char.Name,
		lon, lat, // PostGIS expects (longitude, latitude)
		char.LastPlayed,
	)

	return err
}

// ConnectDB establishes a database connection
func ConnectDB(connectionString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

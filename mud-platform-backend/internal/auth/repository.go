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
		INSERT INTO users (user_id, email, username, password_hash, created_at, last_world_id)
		VALUES ($1, LOWER($2), $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.CreatedAt,
		user.LastWorldID,
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
		SELECT user_id, email, username, password_hash, created_at, last_login, last_world_id
		FROM users
		WHERE user_id = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.LastLogin,
		&user.LastWorldID,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	return &user, err
}

// GetUserByEmail retrieves a user by email
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT user_id, email, username, password_hash, created_at, last_login, last_world_id
		FROM users
		WHERE email = LOWER($1)
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.UserID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.LastLogin,
		&user.LastWorldID,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	return &user, err
}

// GetUserByUsername retrieves a user by username
func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT user_id, email, username, password_hash, created_at, last_login, last_world_id
		FROM users
		WHERE username = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.UserID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.LastLogin,
		&user.LastWorldID,
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
		SET email = LOWER($2), password_hash = $3, last_login = $4, last_world_id = $5
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Email,
		user.PasswordHash,
		user.LastLogin,
		user.LastWorldID,
	)

	return err
}

// CreateCharacter creates a new character
func (r *PostgresRepository) CreateCharacter(ctx context.Context, char *Character) error {
	query := `
		INSERT INTO characters (character_id, user_id, world_id, name, role, appearance, description, occupation, position, position_x, position_y, position_z, created_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::jsonb, $7, $8, ST_SetSRID(ST_MakePoint($9, $10), 4326), $9, $10, $11, $12)
	`

	// Use PositionX/Y if set, or fall back to Position.Longitude/Latitude
	lon := char.PositionX
	lat := char.PositionY
	if char.Position != nil {
		if lon == 0 && lat == 0 {
			lon = char.Position.Longitude
			lat = char.Position.Latitude
		}
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
		lon, lat, // for position (geometry) and x, y
		char.PositionZ,
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
			COALESCE(c.position_x, 0), COALESCE(c.position_y, 0), COALESCE(c.position_z, 0),
			c.created_at, c.last_played
		FROM characters c
		WHERE c.character_id = $1
	`
	// log.Printf("[AUTH] GetCharacter looking for ID: %s", characterID)

	var char Character
	err := r.db.QueryRowContext(ctx, query, characterID).Scan(
		&char.CharacterID,
		&char.UserID,
		&char.WorldID,
		&char.Name,
		&char.Role,
		&char.Appearance,
		&char.Description,
		&char.Occupation,
		&char.PositionX,
		&char.PositionY,
		&char.PositionZ,
		&char.CreatedAt,
		&char.LastPlayed,
	)

	if err == sql.ErrNoRows {
		return nil, ErrCharacterNotFound
	}
	if err != nil {
		return nil, err
	}

	// Backfill Position for compatibility
	char.Position = &Position{
		Latitude:  char.PositionY,
		Longitude: char.PositionX,
	}

	return &char, nil
}

// GetUserCharacters retrieves all characters for a user
func (r *PostgresRepository) GetUserCharacters(ctx context.Context, userID uuid.UUID) ([]*Character, error) {
	query := `
		SELECT 
			c.character_id, c.user_id, c.world_id, c.name,
			COALESCE(c.role, ''), COALESCE(c.appearance::text, ''), COALESCE(c.description, ''), COALESCE(c.occupation, ''),
			COALESCE(c.position_x, 0), COALESCE(c.position_y, 0), COALESCE(c.position_z, 0),
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

		err := rows.Scan(
			&char.CharacterID,
			&char.UserID,
			&char.WorldID,
			&char.Name,
			&char.Role,
			&char.Appearance,
			&char.Description,
			&char.Occupation,
			&char.PositionX,
			&char.PositionY,
			&char.PositionZ,
			&char.CreatedAt,
			&char.LastPlayed,
		)
		if err != nil {
			return nil, err
		}

		char.Position = &Position{
			Latitude:  char.PositionY,
			Longitude: char.PositionX,
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
			COALESCE(c.position_x, 0), COALESCE(c.position_y, 0), COALESCE(c.position_z, 0),
			c.created_at, c.last_played
		FROM characters c
		WHERE c.user_id = $1 AND c.world_id = $2
	`

	var char Character
	err := r.db.QueryRowContext(ctx, query, userID, worldID).Scan(
		&char.CharacterID,
		&char.UserID,
		&char.WorldID,
		&char.Name,
		&char.Role,
		&char.Appearance,
		&char.Description,
		&char.Occupation,
		&char.PositionX,
		&char.PositionY,
		&char.PositionZ,
		&char.CreatedAt,
		&char.LastPlayed,
	)

	if err == sql.ErrNoRows {
		return nil, ErrCharacterNotFound
	}
	if err != nil {
		return nil, err
	}

	char.Position = &Position{
		Latitude:  char.PositionY,
		Longitude: char.PositionX,
	}

	return &char, nil
}

// UpdateCharacter updates an existing character
func (r *PostgresRepository) UpdateCharacter(ctx context.Context, char *Character) error {
	query := `
		UPDATE characters
		SET name = $2, 
		    position = ST_SetSRID(ST_MakePoint($3, $4), 4326), 
		    position_x = $3, position_y = $4, position_z = $5,
		    last_played = $6
		WHERE character_id = $1
	`

	// Use PositionX/Y if set, or fall back to Position.Longitude/Latitude
	lon := char.PositionX
	lat := char.PositionY
	if char.Position != nil {
		// If explicit X/Y are zero but Position is set, might want to use Position?
		// But 0,0 is a valid coordinate.
		// Let's assume the service sets PositionX/Y.
		// If migrating from old code, PositionX/Y might be 0.
		// But we just populated them in GetCharacter.
	}

	_, err := r.db.ExecContext(ctx, query,
		char.CharacterID,
		char.Name,
		lon, lat, char.PositionZ,
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

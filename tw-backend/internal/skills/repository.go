package skills

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for skills persistence
type Repository interface {
	GetSkills(ctx context.Context, characterID uuid.UUID) ([]Skill, error)
	UpdateSkill(ctx context.Context, characterID uuid.UUID, skillName string, xp float64) error
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new PostgreSQL repository
func NewRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetSkills retrieves all persistent skills for a character
func (r *PostgresRepository) GetSkills(ctx context.Context, characterID uuid.UUID) ([]Skill, error) {
	query := `
		SELECT skill_name, xp
		FROM character_skills
		WHERE character_id = $1
	`

	rows, err := r.db.Query(ctx, query, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		var s Skill
		if err := rows.Scan(&s.Name, &s.XP); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}

	return skills, nil
}

// UpdateSkill updates or inserts a skill's XP
func (r *PostgresRepository) UpdateSkill(ctx context.Context, characterID uuid.UUID, skillName string, xp float64) error {
	query := `
		INSERT INTO character_skills (character_id, skill_name, xp, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (character_id, skill_name)
		DO UPDATE SET xp = $3, updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.Exec(ctx, query, characterID, skillName, xp)
	return err
}

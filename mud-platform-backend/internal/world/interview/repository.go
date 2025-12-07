package interview

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements Repository for PostgreSQL
type PostgresRepository struct {
	db *pgxpool.Pool
}

// IsWorldNameTaken checks if a world name is already taken
func (r *PostgresRepository) IsWorldNameTaken(ctx context.Context, name string) (bool, error) {
	// Check if name exists in worlds table (case insensitive)
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM worlds 
			WHERE name ILIKE $1
		)
	`
	var exists bool
	err := r.db.QueryRow(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check world name: %w", err)
	}
	return exists, nil
}

// NewRepository creates a new PostgresRepository
func NewRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetInterview retrieves an interview by user ID
func (r *PostgresRepository) GetInterview(ctx context.Context, userID uuid.UUID) (*Interview, error) {
	query := `
		SELECT id, user_id, status, current_question_index, created_at, updated_at
		FROM world_interviews
		WHERE user_id = $1
	`
	row := r.db.QueryRow(ctx, query, userID)

	var i Interview
	var statusStr string
	err := row.Scan(&i.ID, &i.UserID, &statusStr, &i.CurrentQuestionIndex, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to get interview: %w", err)
	}
	i.Status = Status(statusStr)
	return &i, nil
}

// CreateInterview creates a new interview for a user
func (r *PostgresRepository) CreateInterview(ctx context.Context, userID uuid.UUID) (*Interview, error) {
	query := `
		INSERT INTO world_interviews (user_id, status, current_question_index)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	var i Interview
	i.UserID = userID
	i.Status = StatusNotStarted
	i.CurrentQuestionIndex = 0

	err := r.db.QueryRow(ctx, query, userID, i.Status, i.CurrentQuestionIndex).Scan(&i.ID, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create interview: %w", err)
	}
	return &i, nil
}

// UpdateInterviewStatus updates the status of an interview
func (r *PostgresRepository) UpdateInterviewStatus(ctx context.Context, id uuid.UUID, status Status) error {
	query := `
		UPDATE world_interviews
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update interview status: %w", err)
	}
	return nil
}

// UpdateQuestionIndex updates the current question index
func (r *PostgresRepository) UpdateQuestionIndex(ctx context.Context, id uuid.UUID, index int) error {
	query := `
		UPDATE world_interviews
		SET current_question_index = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, index, id)
	if err != nil {
		return fmt.Errorf("failed to update question index: %w", err)
	}
	return nil
}

// SaveAnswer saves an answer to a question
func (r *PostgresRepository) SaveAnswer(ctx context.Context, interviewID uuid.UUID, index int, text string) error {
	query := `
		INSERT INTO interview_answers (interview_id, question_index, answer_text)
		VALUES ($1, $2, $3)
		ON CONFLICT (interview_id, question_index)
		DO UPDATE SET answer_text = EXCLUDED.answer_text, created_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, interviewID, index, text)
	if err != nil {
		return fmt.Errorf("failed to save answer: %w", err)
	}
	return nil
}

// GetAnswers retrieves all answers for an interview
func (r *PostgresRepository) GetAnswers(ctx context.Context, interviewID uuid.UUID) ([]Answer, error) {
	query := `
		SELECT id, interview_id, question_index, answer_text, created_at
		FROM interview_answers
		WHERE interview_id = $1
		ORDER BY question_index ASC
	`
	rows, err := r.db.Query(ctx, query, interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get answers: %w", err)
	}
	defer rows.Close()

	var answers []Answer
	for rows.Next() {
		var a Answer
		err := rows.Scan(&a.ID, &a.InterviewID, &a.QuestionIndex, &a.AnswerText, &a.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan answer: %w", err)
		}
		answers = append(answers, a)
	}
	return answers, nil
}

// SaveConfiguration saves a world configuration
func (r *PostgresRepository) SaveConfiguration(ctx context.Context, config *WorldConfiguration) error {
	query := `
		INSERT INTO world_configurations (interview_id, world_id, created_by, configuration)
		VALUES ($1, $2, $3, $4)
	`
	// We need to serialize the config to JSONB
	// But wait, WorldConfiguration struct has fields that map to columns in the OLD table.
	// In the NEW table, we have a single 'configuration' JSONB column.
	// So we should marshal the whole struct (or relevant parts) to JSON.

	// However, WorldConfiguration struct in types.go still has the old fields.
	// I should probably just store the whole struct as JSON.

	_, err := r.db.Exec(ctx, query, config.InterviewID, config.WorldID, config.CreatedBy, config)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	return nil
}

// GetConfigurationByWorldID retrieves a configuration by world ID
func (r *PostgresRepository) GetConfigurationByWorldID(ctx context.Context, worldID uuid.UUID) (*WorldConfiguration, error) {
	query := `
		SELECT configuration
		FROM world_configurations
		WHERE world_id = $1
	`
	var config WorldConfiguration
	err := r.db.QueryRow(ctx, query, worldID).Scan(&config)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}
	return &config, nil
}

// GetConfigurationByUserID retrieves a configuration by user ID
func (r *PostgresRepository) GetConfigurationByUserID(ctx context.Context, userID uuid.UUID) (*WorldConfiguration, error) {
	query := `
		SELECT configuration
		FROM world_configurations
		WHERE created_by = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var config WorldConfiguration
	err := r.db.QueryRow(ctx, query, userID).Scan(&config)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}
	return &config, nil
}

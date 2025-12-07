package interview

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for interview repository operations
type Repository interface {
	// Interview Lifecycle
	GetInterview(ctx context.Context, userID uuid.UUID) (*Interview, error)
	CreateInterview(ctx context.Context, userID uuid.UUID) (*Interview, error)
	UpdateInterviewStatus(ctx context.Context, id uuid.UUID, status Status) error
	UpdateQuestionIndex(ctx context.Context, id uuid.UUID, index int) error

	// Answers
	SaveAnswer(ctx context.Context, interviewID uuid.UUID, index int, text string) error
	GetAnswers(ctx context.Context, interviewID uuid.UUID) ([]Answer, error)

	// Configuration (for world generation and statue description)
	SaveConfiguration(ctx context.Context, config *WorldConfiguration) error
	GetConfigurationByWorldID(ctx context.Context, worldID uuid.UUID) (*WorldConfiguration, error)
	GetConfigurationByUserID(ctx context.Context, userID uuid.UUID) (*WorldConfiguration, error)

	// World Name Validation
	IsWorldNameTaken(ctx context.Context, name string) (bool, error)
}

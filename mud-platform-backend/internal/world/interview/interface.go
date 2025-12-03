package interview

import (
	"context"

	"github.com/google/uuid"
)

// RepositoryInterface defines the interface for interview repository operations
type RepositoryInterface interface {
	// Configuration methods
	GetConfigurationByWorldID(worldID uuid.UUID) (*WorldConfiguration, error)
	// Name uniqueness check
	IsWorldNameTaken(name string) (bool, error)
	SaveConfiguration(config *WorldConfiguration) error

	// Session methods
	GetSessionByID(sessionID string) (*InterviewSession, error)
	SaveSession(session *InterviewSession) error
	GetActiveSessionForUser(ctx context.Context, userID uuid.UUID) (*InterviewSession, error)

	// Interview methods (for backward compatibility with service)
	SaveInterview(session *InterviewSession) error
	GetInterview(id uuid.UUID) (*InterviewSession, error)
	UpdateInterview(session *InterviewSession) error
	GetActiveInterviewByPlayer(playerID uuid.UUID) (*InterviewSession, error)
}

// Ensure Repository implements RepositoryInterface
var _ RepositoryInterface = (*Repository)(nil)

// Ensure MockRepository implements RepositoryInterface
var _ RepositoryInterface = (*MockRepository)(nil)

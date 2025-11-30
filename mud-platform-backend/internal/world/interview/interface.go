package interview

import (
	"context"

	"github.com/google/uuid"
)

// RepositoryInterface defines the interface for interview repository operations
type RepositoryInterface interface {
	GetConfigurationByWorldID(worldID uuid.UUID) (*WorldConfiguration, error)
	GetSessionByID(sessionID string) (*InterviewSession, error)
	SaveSession(session *InterviewSession) error
	GetActiveSessionForUser(ctx context.Context, userID uuid.UUID) (*InterviewSession, error)
}

// Ensure Repository implements RepositoryInterface
var _ RepositoryInterface = (*Repository)(nil)

// Ensure MockRepository implements RepositoryInterface
var _ RepositoryInterface = (*MockRepository)(nil)

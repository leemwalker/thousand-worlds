package interview

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// MockRepository is a mock implementation of Repository for testing
type MockRepository struct {
	mu            sync.RWMutex
	configurations map[uuid.UUID]*WorldConfiguration
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		configurations: make(map[uuid.UUID]*WorldConfiguration),
	}
}

// SaveConfiguration stores a configuration (for testing)
func (m *MockRepository) SaveConfiguration(worldID uuid.UUID, config *WorldConfiguration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configurations[worldID] = config
}

// GetConfigurationByWorldID retrieves a configuration
func (m *MockRepository) GetConfigurationByWorldID(worldID uuid.UUID) (*WorldConfiguration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.configurations[worldID]
	if !exists {
		return nil, fmt.Errorf("configuration not found for world %s", worldID)
	}
	return config, nil
}

// GetSessionByID mock implementation
func (m *MockRepository) GetSessionByID(sessionID string) (*InterviewSession, error) {
	return nil, fmt.Errorf("not implemented")
}

// SaveSession mock implementation
func (m *MockRepository) SaveSession(session *InterviewSession) error {
	return nil
}

// GetActiveSessionForUser mock implementation
func (m *MockRepository) GetActiveSessionForUser(ctx context.Context, userID uuid.UUID) (*InterviewSession, error) {
	return nil, fmt.Errorf("no active session")
}

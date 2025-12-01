package interview

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// MockRepository is a mock implementation of Repository for testing
type MockRepository struct {
	mu             sync.RWMutex
	configurations map[uuid.UUID]*WorldConfiguration
	sessions       map[uuid.UUID]*InterviewSession
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		configurations: make(map[uuid.UUID]*WorldConfiguration),
		sessions:       make(map[uuid.UUID]*InterviewSession),
	}
}

// SaveConfiguration stores a configuration (for testing)
func (m *MockRepository) SaveConfiguration(config *WorldConfiguration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if config.WorldID != nil {
		m.configurations[*config.WorldID] = config
	}
	return nil
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

// SaveInterview saves an interview session
func (m *MockRepository) SaveInterview(session *InterviewSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	return nil
}

// GetInterview retrieves an interview by ID
func (m *MockRepository) GetInterview(id uuid.UUID) (*InterviewSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}

// UpdateInterview updates an existing interview
func (m *MockRepository) UpdateInterview(session *InterviewSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	return nil
}

// GetActiveInterviewByPlayer gets the active interview for a player
func (m *MockRepository) GetActiveInterviewByPlayer(playerID uuid.UUID) (*InterviewSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, session := range m.sessions {
		if session.PlayerID == playerID && !session.State.IsComplete {
			return session, nil
		}
	}
	return nil, nil // No active session
}

// GetSessionByID retrieves a session by string ID (interface method)
func (m *MockRepository) GetSessionByID(sessionID string) (*InterviewSession, error) {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}
	return m.GetInterview(id)
}

// SaveSession saves a session (interface method)
func (m *MockRepository) SaveSession(session *InterviewSession) error {
	return m.SaveInterview(session)
}

// GetActiveSessionForUser retrieves active session for a user (interface method)
func (m *MockRepository) GetActiveSessionForUser(ctx context.Context, userID uuid.UUID) (*InterviewSession, error) {
	return m.GetActiveInterviewByPlayer(userID)
}


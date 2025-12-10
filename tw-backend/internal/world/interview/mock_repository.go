package interview

import (
	"context"
	"fmt"
	"tw-backend/internal/repository"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockRepository is a mock implementation of Repository for testing
type MockRepository struct {
	mu             sync.RWMutex
	interviews     map[uuid.UUID]*Interview
	answers        map[uuid.UUID][]Answer
	configurations map[uuid.UUID]*WorldConfiguration
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		interviews:     make(map[uuid.UUID]*Interview),
		answers:        make(map[uuid.UUID][]Answer),
		configurations: make(map[uuid.UUID]*WorldConfiguration),
	}
}

// GetInterview retrieves an interview by user ID
func (m *MockRepository) GetInterview(ctx context.Context, userID uuid.UUID) (*Interview, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, i := range m.interviews {
		if i.UserID == userID {
			return i, nil
		}
	}
	return nil, nil
}

// CreateInterview creates a new interview
func (m *MockRepository) CreateInterview(ctx context.Context, userID uuid.UUID) (*Interview, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New()
	interview := &Interview{
		ID:                   id,
		UserID:               userID,
		Status:               StatusNotStarted,
		CurrentQuestionIndex: 0,
		CreatedAt:            time.Now(), // Use time.Now() or passed context time?
		UpdatedAt:            time.Now(),
	}
	m.interviews[id] = interview
	return interview, nil
}

// UpdateInterviewStatus updates status
func (m *MockRepository) UpdateInterviewStatus(ctx context.Context, id uuid.UUID, status Status) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if i, ok := m.interviews[id]; ok {
		i.Status = status
		i.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("interview not found")
}

// UpdateQuestionIndex updates index
func (m *MockRepository) UpdateQuestionIndex(ctx context.Context, id uuid.UUID, index int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if i, ok := m.interviews[id]; ok {
		i.CurrentQuestionIndex = index
		i.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("interview not found")
}

// SaveAnswer saves an answer
func (m *MockRepository) SaveAnswer(ctx context.Context, interviewID uuid.UUID, index int, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ans := Answer{
		ID:            uuid.New(),
		InterviewID:   interviewID,
		QuestionIndex: index,
		AnswerText:    text,
		CreatedAt:     time.Now(),
	}

	// Remove existing answer for this index if any
	existing := m.answers[interviewID]
	var newAnswers []Answer
	for _, a := range existing {
		if a.QuestionIndex != index {
			newAnswers = append(newAnswers, a)
		}
	}
	newAnswers = append(newAnswers, ans)
	m.answers[interviewID] = newAnswers
	return nil
}

// GetAnswers retrieves answers
func (m *MockRepository) GetAnswers(ctx context.Context, interviewID uuid.UUID) ([]Answer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.answers[interviewID], nil
}

// SaveConfiguration saves config
func (m *MockRepository) SaveConfiguration(ctx context.Context, config *WorldConfiguration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configurations[config.ID] = config
	return nil
}

// GetConfigurationByWorldID gets config by world ID
func (m *MockRepository) GetConfigurationByWorldID(ctx context.Context, worldID uuid.UUID) (*WorldConfiguration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.configurations {
		if c.WorldID != nil && *c.WorldID == worldID {
			return c, nil
		}
	}
	return nil, nil
}

// GetConfigurationByUserID gets config by user ID
func (m *MockRepository) GetConfigurationByUserID(ctx context.Context, userID uuid.UUID) (*WorldConfiguration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.configurations {
		if c.CreatedBy == userID {
			return c, nil
		}
	}
	return nil, nil
}

// IsWorldNameTaken checks if name is taken
func (m *MockRepository) IsWorldNameTaken(ctx context.Context, name string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.configurations {
		// Simple case-insensitive check
		if len(c.WorldName) == len(name) {
			match := true
			for i := 0; i < len(name); i++ {
				c1 := c.WorldName[i]
				c2 := name[i]
				if c1 >= 'A' && c1 <= 'Z' {
					c1 += 'a' - 'A'
				}
				if c2 >= 'A' && c2 <= 'Z' {
					c2 += 'a' - 'A'
				}
				if c1 != c2 {
					match = false
					break
				}
			}
			if match {
				return true, nil
			}
		}
	}
	return false, nil
}

// MockWorldRepository mocks repository.WorldRepository
type MockWorldRepository struct {
	mu     sync.RWMutex
	Worlds map[uuid.UUID]*repository.World
}

func NewMockWorldRepository() *MockWorldRepository {
	return &MockWorldRepository{
		Worlds: make(map[uuid.UUID]*repository.World),
	}
}

func (m *MockWorldRepository) CreateWorld(ctx context.Context, world *repository.World) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Worlds[world.ID] = world
	return nil
}

func (m *MockWorldRepository) GetWorld(ctx context.Context, worldID uuid.UUID) (*repository.World, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if w, ok := m.Worlds[worldID]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("world not found")
}

func (m *MockWorldRepository) ListWorlds(ctx context.Context) ([]repository.World, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var worlds []repository.World
	for _, w := range m.Worlds {
		worlds = append(worlds, *w)
	}
	return worlds, nil
}

func (m *MockWorldRepository) GetWorldsByOwner(ctx context.Context, ownerID uuid.UUID) ([]repository.World, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var worlds []repository.World
	for _, w := range m.Worlds {
		if w.OwnerID == ownerID {
			worlds = append(worlds, *w)
		}
	}
	return worlds, nil
}

func (m *MockWorldRepository) UpdateWorld(ctx context.Context, world *repository.World) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Worlds[world.ID]; ok {
		m.Worlds[world.ID] = world
		return nil
	}
	return fmt.Errorf("world not found")
}

func (m *MockWorldRepository) DeleteWorld(ctx context.Context, worldID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Worlds, worldID)
	return nil
}

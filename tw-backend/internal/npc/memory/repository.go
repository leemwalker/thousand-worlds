package memory

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for memory storage
type Repository interface {
	CreateMemory(memory Memory) (uuid.UUID, error)
	GetMemory(id uuid.UUID) (Memory, error)
	GetMemoriesByNPC(npcID uuid.UUID, limit, offset int) ([]Memory, error)
	GetMemoriesByType(npcID uuid.UUID, memoryType string, limit int) ([]Memory, error)
	GetMemoriesByTimeframe(npcID uuid.UUID, startTime, endTime time.Time) ([]Memory, error)
	GetMemoriesByEntity(npcID uuid.UUID, entityID uuid.UUID) ([]Memory, error)
	GetMemoriesByEmotion(npcID uuid.UUID, minEmotionalWeight float64, limit int) ([]Memory, error)
	GetMemoriesByTags(npcID uuid.UUID, tags []string, matchAll bool) ([]Memory, error)
	GetAllMemories(npcID uuid.UUID) ([]Memory, error)
	UpdateMemory(memory Memory) error
	DeleteMemory(id uuid.UUID) error
}

// MockRepository is an in-memory implementation for testing
type MockRepository struct {
	memories map[uuid.UUID]Memory
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		memories: make(map[uuid.UUID]Memory),
	}
}

func (r *MockRepository) CreateMemory(memory Memory) (uuid.UUID, error) {
	if memory.ID == uuid.Nil {
		memory.ID = uuid.New()
	}
	r.memories[memory.ID] = memory
	return memory.ID, nil
}

func (r *MockRepository) GetMemory(id uuid.UUID) (Memory, error) {
	mem, ok := r.memories[id]
	if !ok {
		return Memory{}, errors.New("memory not found")
	}
	return mem, nil
}

func (r *MockRepository) GetMemoriesByNPC(npcID uuid.UUID, limit, offset int) ([]Memory, error) {
	var result []Memory
	for _, mem := range r.memories {
		if mem.NPCID == npcID {
			result = append(result, mem)
		}
	}
	// Sort by timestamp desc (mock behavior)
	// ... skipping sort for mock simplicity unless needed

	// Apply offset/limit
	if offset >= len(result) {
		return []Memory{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (r *MockRepository) GetMemoriesByType(npcID uuid.UUID, memoryType string, limit int) ([]Memory, error) {
	var result []Memory
	for _, mem := range r.memories {
		if mem.NPCID == npcID && mem.Type == memoryType {
			result = append(result, mem)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *MockRepository) GetMemoriesByTimeframe(npcID uuid.UUID, startTime, endTime time.Time) ([]Memory, error) {
	var result []Memory
	for _, mem := range r.memories {
		if mem.NPCID == npcID && (mem.Timestamp.Equal(startTime) || mem.Timestamp.After(startTime)) && (mem.Timestamp.Equal(endTime) || mem.Timestamp.Before(endTime)) {
			result = append(result, mem)
		}
	}
	return result, nil
}

func (r *MockRepository) GetMemoriesByEntity(npcID uuid.UUID, entityID uuid.UUID) ([]Memory, error) {
	var result []Memory
	for _, mem := range r.memories {
		if mem.NPCID != npcID {
			continue
		}

		// Check content for entity ID
		found := false
		switch c := mem.Content.(type) {
		case ObservationContent:
			for _, e := range c.EntitiesPresent {
				if e == entityID {
					found = true
					break
				}
			}
		case ConversationContent:
			for _, p := range c.Participants {
				if p == entityID {
					found = true
					break
				}
			}
		case EventContent:
			for _, p := range c.Participants {
				if p == entityID {
					found = true
					break
				}
			}
		case RelationshipContent:
			if c.TargetEntityID == entityID {
				found = true
			}
		}
		if found {
			result = append(result, mem)
		}
	}
	return result, nil
}

func (r *MockRepository) GetMemoriesByEmotion(npcID uuid.UUID, minEmotionalWeight float64, limit int) ([]Memory, error) {
	var result []Memory
	for _, mem := range r.memories {
		if mem.NPCID == npcID && mem.EmotionalWeight >= minEmotionalWeight {
			result = append(result, mem)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *MockRepository) GetMemoriesByTags(npcID uuid.UUID, tags []string, matchAll bool) ([]Memory, error) {
	var result []Memory
	for _, mem := range r.memories {
		if mem.NPCID != npcID {
			continue
		}

		matches := 0
		for _, tag := range tags {
			for _, memTag := range mem.Tags {
				if tag == memTag {
					matches++
					break
				}
			}
		}

		if matchAll {
			if matches == len(tags) {
				result = append(result, mem)
			}
		} else {
			if matches > 0 {
				result = append(result, mem)
			}
		}
	}
	return result, nil
}

func (r *MockRepository) GetAllMemories(npcID uuid.UUID) ([]Memory, error) {
	var result []Memory
	for _, mem := range r.memories {
		if mem.NPCID == npcID {
			result = append(result, mem)
		}
	}
	return result, nil
}

func (r *MockRepository) UpdateMemory(memory Memory) error {
	if _, ok := r.memories[memory.ID]; !ok {
		return errors.New("memory not found")
	}
	r.memories[memory.ID] = memory
	return nil
}

func (r *MockRepository) DeleteMemory(id uuid.UUID) error {
	delete(r.memories, id)
	return nil
}

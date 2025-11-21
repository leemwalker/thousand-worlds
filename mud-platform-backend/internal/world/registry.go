package world

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Registry is a thread-safe in-memory registry of world states
type Registry struct {
	mu     sync.RWMutex
	worlds map[uuid.UUID]*WorldState
}

// NewRegistry creates a new world registry
func NewRegistry() *Registry {
	return &Registry{
		worlds: make(map[uuid.UUID]*WorldState),
	}
}

// RegisterWorld adds a new world to the registry
// Returns error if world with same ID already exists
func (r *Registry) RegisterWorld(state *WorldState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.worlds[state.ID]; exists {
		return fmt.Errorf("world %s already registered", state.ID)
	}

	// Store a copy to prevent external mutation
	worldCopy := *state
	r.worlds[state.ID] = &worldCopy

	return nil
}

// GetWorld retrieves a world by ID
// Returns error if world not found
func (r *Registry) GetWorld(worldID uuid.UUID) (*WorldState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	world, exists := r.worlds[worldID]
	if !exists {
		return nil, fmt.Errorf("world %s not found", worldID)
	}

	// Return a copy to prevent external mutation
	worldCopy := *world
	return &worldCopy, nil
}

// UpdateWorld atomically updates a world's state
// Returns error if world not found
func (r *Registry) UpdateWorld(worldID uuid.UUID, updateFn func(*WorldState)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	world, exists := r.worlds[worldID]
	if !exists {
		return fmt.Errorf("world %s not found", worldID)
	}

	// Apply update function
	updateFn(world)

	return nil
}

// ListWorlds returns all registered worlds
func (r *Registry) ListWorlds() []*WorldState {
	r.mu.RLock()
	defer r.mu.RUnlock()

	worlds := make([]*WorldState, 0, len(r.worlds))
	for _, world := range r.worlds {
		// Return copies to prevent external mutation
		worldCopy := *world
		worlds = append(worlds, &worldCopy)
	}

	return worlds
}

// RemoveWorld removes a world from the registry
// Returns error if world not found
func (r *Registry) RemoveWorld(worldID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.worlds[worldID]; !exists {
		return fmt.Errorf("world %s not found", worldID)
	}

	delete(r.worlds, worldID)

	return nil
}

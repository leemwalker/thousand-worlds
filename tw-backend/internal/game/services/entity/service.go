package entity

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// EntityType defines the type of entity
type EntityType string

const (
	EntityTypeItem EntityType = "item"
	EntityTypeNPC  EntityType = "npc"
)

// Entity represents an object in the world
type Entity struct {
	ID           uuid.UUID
	Type         EntityType
	Name         string
	Description  string
	WorldID      uuid.UUID
	X            float64
	Y            float64
	Z            float64
	Interactable bool
	Properties   map[string]interface{}
}

// Service manages entities in the game world
type Service struct {
	entities map[uuid.UUID]*Entity
	mutex    sync.RWMutex
}

// NewService creates a new entity service
func NewService() *Service {
	return &Service{
		entities: make(map[uuid.UUID]*Entity),
	}
}

// AddEntity adds an entity to the world
func (s *Service) AddEntity(ctx context.Context, entity *Entity) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	s.entities[entity.ID] = entity
	return nil
}

// RemoveEntity removes an entity from the world
func (s *Service) RemoveEntity(ctx context.Context, id uuid.UUID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.entities[id]; !exists {
		return fmt.Errorf("entity not found")
	}
	delete(s.entities, id)
	return nil
}

// GetEntitiesAt returns all entities within a radius of a location
func (s *Service) GetEntitiesAt(ctx context.Context, worldID uuid.UUID, x, y, radius float64) ([]*Entity, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var result []*Entity
	for _, e := range s.entities {
		if e.WorldID != worldID {
			continue
		}

		// Simple distance check (ignoring Z for now or assuming flat plane logic mostly)
		dx := e.X - x
		dy := e.Y - y
		distSq := dx*dx + dy*dy

		if distSq <= radius*radius {
			result = append(result, e)
		}
	}
	return result, nil
}

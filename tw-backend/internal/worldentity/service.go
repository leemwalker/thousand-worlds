package worldentity

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/google/uuid"
)

// Service manages world entities with caching
type Service struct {
	repo  Repository
	cache map[uuid.UUID][]*WorldEntity // In-memory cache per world
	mu    sync.RWMutex
}

// NewService creates a new WorldEntity service
func NewService(repo Repository) *Service {
	return &Service{
		repo:  repo,
		cache: make(map[uuid.UUID][]*WorldEntity),
	}
}

// Create adds a new entity to the world
func (s *Service) Create(ctx context.Context, entity *WorldEntity) error {
	if err := s.repo.Create(ctx, entity); err != nil {
		return err
	}
	// Invalidate cache for this world
	s.invalidateCache(entity.WorldID)
	return nil
}

// GetByID retrieves an entity by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*WorldEntity, error) {
	return s.repo.GetByID(ctx, id)
}

// GetEntitiesInWorld returns all entities in a world (cached)
func (s *Service) GetEntitiesInWorld(ctx context.Context, worldID uuid.UUID) ([]*WorldEntity, error) {
	s.mu.RLock()
	if cached, ok := s.cache[worldID]; ok {
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	// Load from database
	entities, err := s.repo.GetByWorldID(ctx, worldID)
	if err != nil {
		return nil, err
	}

	// Cache the results
	s.mu.Lock()
	s.cache[worldID] = entities
	s.mu.Unlock()

	return entities, nil
}

// GetEntitiesAt returns entities within radius of position
func (s *Service) GetEntitiesAt(ctx context.Context, worldID uuid.UUID, x, y, radius float64) ([]*WorldEntity, error) {
	// Use cached entities if available for small radius queries
	entities, err := s.GetEntitiesInWorld(ctx, worldID)
	if err != nil {
		// Fallback to direct DB query
		return s.repo.GetAtPosition(ctx, worldID, x, y, radius)
	}

	// Filter cached entities by distance
	var result []*WorldEntity
	for _, e := range entities {
		dist := math.Sqrt(math.Pow(e.X-x, 2) + math.Pow(e.Y-y, 2))
		if dist <= radius {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetEntityByName finds an entity by name in the world (case-insensitive)
func (s *Service) GetEntityByName(ctx context.Context, worldID uuid.UUID, name string) (*WorldEntity, error) {
	return s.repo.GetByName(ctx, worldID, name)
}

// CheckCollision checks if movement to (x,y) is blocked by an entity
// Returns true if blocked, the blocking entity, and any error
func (s *Service) CheckCollision(ctx context.Context, worldID uuid.UUID, x, y float64) (bool, *WorldEntity, error) {
	// Get all entities in the world (cached if available)
	entities, err := s.GetEntitiesInWorld(ctx, worldID)
	if err != nil {
		fmt.Printf("[COLLISION] Error getting entities: %v\n", err)
		return false, nil, err
	}

	for _, entity := range entities {
		if !entity.Collision {
			continue
		}

		// Calculate distance to entity center
		dist := math.Sqrt(math.Pow(entity.X-x, 2) + math.Pow(entity.Y-y, 2))
		radius := entity.CollisionRadius()

		if dist < radius {
			return true, entity, nil
		}
	}

	return false, nil, nil
}

// CanInteract checks if an entity can be interacted with for a given action
// Returns (allowed, error message)
func (s *Service) CanInteract(entity *WorldEntity, action string) (bool, string) {
	if entity == nil {
		return false, "Entity not found."
	}

	// Check if entity is locked
	if entity.Locked {
		switch action {
		case "get", "take", "grab", "pick":
			return false, fmt.Sprintf("You cannot move the %s.", entity.Name)
		case "push", "pull", "move":
			return false, fmt.Sprintf("You cannot move the %s.", entity.Name)
		}
	}

	// Check if entity type is gettable
	if action == "get" || action == "take" || action == "grab" || action == "pick" {
		if entity.EntityType != EntityTypeItem {
			return false, fmt.Sprintf("You cannot pick up the %s.", entity.Name)
		}
	}

	return true, ""
}

// Update updates an entity
func (s *Service) Update(ctx context.Context, entity *WorldEntity) error {
	if err := s.repo.Update(ctx, entity); err != nil {
		return err
	}
	s.invalidateCache(entity.WorldID)
	return nil
}

// Delete removes an entity
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	// Get entity first to know which world cache to invalidate
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.invalidateCache(entity.WorldID)
	return nil
}

// invalidateCache clears the cache for a world
func (s *Service) invalidateCache(worldID uuid.UUID) {
	s.mu.Lock()
	delete(s.cache, worldID)
	s.mu.Unlock()
}

// ClearCache clears all cached data
func (s *Service) ClearCache() {
	s.mu.Lock()
	s.cache = make(map[uuid.UUID][]*WorldEntity)
	s.mu.Unlock()
}

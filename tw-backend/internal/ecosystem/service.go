package ecosystem

import (
	"log"
	"math/rand"
	"sync"
	"tw-backend/internal/ai/behaviortree"
	goap "tw-backend/internal/ai/goap"
	"tw-backend/internal/ecosystem/state"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// Service manages all ecosystem entities
type Service struct {
	Entities map[uuid.UUID]*state.LivingEntityState
	mu       sync.RWMutex

	Spawner          *Spawner
	Needs            *state.NeedSystem
	Planner          *goap.Planner
	EvolutionManager *EvolutionManager

	// Map of entity ID to its current behavior tree
	Behaviors map[uuid.UUID]behaviortree.Node
}

func NewService(seed int64) *Service {
	return &Service{
		Entities:         make(map[uuid.UUID]*state.LivingEntityState),
		Spawner:          NewSpawner(seed),
		Needs:            &state.NeedSystem{},
		Planner:          goap.NewPlanner(),
		EvolutionManager: NewEvolutionManager(),
		Behaviors:        make(map[uuid.UUID]behaviortree.Node),
	}
}

// GetEvolutionManager returns the evolution manager for reproduction
func (s *Service) GetEvolutionManager() *EvolutionManager {
	return s.EvolutionManager
}

// SpawnBiomes populates the world based on biomes
// This would be called by WorldGen or a periodic spawner
func (s *Service) SpawnBiomes(worldID uuid.UUID, biomes []geography.Biome) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cap total entities to prevent memory issues
	const maxEntities = 1000

	// If we have many biomes, sample them rather than spawning for all
	biomesToProcess := biomes
	if len(biomes) > 200 {
		// Sample ~200 biomes randomly
		sampled := make([]geography.Biome, 200)
		step := len(biomes) / 200
		for i := 0; i < 200; i++ {
			sampled[i] = biomes[i*step]
		}
		biomesToProcess = sampled
	}

	for _, b := range biomesToProcess {
		if len(s.Entities) >= maxEntities {
			break // Cap reached
		}

		// Calculate density based on biome type
		count := 5 // default
		if b.Type == geography.BiomeDesert {
			count = 2
		}
		if b.Type == geography.BiomeRainforest {
			count = 10
		}

		// Don't exceed max
		remaining := maxEntities - len(s.Entities)
		if count > remaining {
			count = remaining
		}

		newEntities := s.Spawner.SpawnEntitiesForBiome(b.Type, count)
		for _, e := range newEntities {
			e.WorldID = worldID
			s.Entities[e.EntityID] = e

			// Assign AI based on diet
			switch e.Diet {
			case state.DietPhotosynthetic:
				s.Behaviors[e.EntityID] = behaviortree.NewFloraTree()
			case state.DietHerbivore:
				s.Behaviors[e.EntityID] = behaviortree.NewHerbivoreTree()
			default:
				// Carnivores/Omnivores default to herbivore tree for now
				s.Behaviors[e.EntityID] = behaviortree.NewHerbivoreTree()
			}
		}
	}
}

// Tick advances the simulation for all entities
func (s *Service) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Pre-categorize entities by diet for O(1) prey lookup
	var flora, herbivores []*state.LivingEntityState
	var floraIDs, herbivoreIDs []uuid.UUID
	for id, e := range s.Entities {
		switch e.Diet {
		case state.DietPhotosynthetic:
			flora = append(flora, e)
			floraIDs = append(floraIDs, id)
		case state.DietHerbivore:
			herbivores = append(herbivores, e)
			herbivoreIDs = append(herbivoreIDs, id)
		}
	}

	toRemove := make(map[uuid.UUID]bool)

	// Single pass: aging, feeding, death check
	for id, entity := range s.Entities {
		entity.Age++
		s.Needs.Tick(entity, nil)

		// Feeding (only if hungry)
		if entity.Diet != state.DietPhotosynthetic && entity.Needs.Hunger >= 50 {
			var preyList []*state.LivingEntityState
			var preyIDs []uuid.UUID

			switch entity.Diet {
			case state.DietHerbivore:
				preyList, preyIDs = flora, floraIDs
			case state.DietCarnivore:
				preyList, preyIDs = herbivores, herbivoreIDs
			case state.DietOmnivore:
				// Combine flora and herbivores
				preyList = append(flora, herbivores...)
				preyIDs = append(floraIDs, herbivoreIDs...)
			}

			// Random prey selection (O(1) instead of O(n))
			if len(preyList) > 0 {
				idx := rand.Intn(len(preyList))
				preyID := preyIDs[idx]
				if !toRemove[preyID] && preyID != id {
					entity.Needs.Hunger = 0
					entity.Needs.Thirst *= 0.5
					toRemove[preyID] = true
				}
			}
		}

		// Death check
		shouldDie := toRemove[id]
		if !shouldDie {
			maxAge := int64(18250) // Flora: 50 years
			switch entity.Diet {
			case state.DietHerbivore:
				maxAge = 3650
			case state.DietCarnivore, state.DietOmnivore:
				maxAge = 5475
			}
			if entity.Age > maxAge {
				shouldDie = true
			}
			if entity.Diet != state.DietPhotosynthetic {
				if entity.Needs.Hunger >= 100 || entity.Needs.Thirst >= 100 {
					shouldDie = true
				}
			}
		}

		if shouldDie {
			delete(s.Entities, id)
			delete(s.Behaviors, id)
		}
	}
}

// TickWithStats runs a tick and returns statistics for logging
func (s *Service) TickWithStats() (flora, herbivores, carnivores, deaths int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Pre-categorize
	var floraList, herbivoreList []*state.LivingEntityState
	var floraIDs, herbivoreIDs []uuid.UUID
	for id, e := range s.Entities {
		switch e.Diet {
		case state.DietPhotosynthetic:
			floraList = append(floraList, e)
			floraIDs = append(floraIDs, id)
		case state.DietHerbivore:
			herbivoreList = append(herbivoreList, e)
			herbivoreIDs = append(herbivoreIDs, id)
		}
	}

	toRemove := make(map[uuid.UUID]bool)

	for id, entity := range s.Entities {
		entity.Age++
		s.Needs.Tick(entity, nil)

		if entity.Diet != state.DietPhotosynthetic && entity.Needs.Hunger >= 50 {
			var preyList []*state.LivingEntityState
			var preyIDs []uuid.UUID
			switch entity.Diet {
			case state.DietHerbivore:
				preyList, preyIDs = floraList, floraIDs
			case state.DietCarnivore:
				preyList, preyIDs = herbivoreList, herbivoreIDs
			case state.DietOmnivore:
				preyList = append(floraList, herbivoreList...)
				preyIDs = append(floraIDs, herbivoreIDs...)
			}
			if len(preyList) > 0 {
				idx := rand.Intn(len(preyList))
				preyID := preyIDs[idx]
				if !toRemove[preyID] && preyID != id {
					entity.Needs.Hunger = 0
					entity.Needs.Thirst *= 0.5
					toRemove[preyID] = true
				}
			}
		}

		shouldDie := toRemove[id]
		if !shouldDie {
			maxAge := int64(18250)
			switch entity.Diet {
			case state.DietHerbivore:
				maxAge = 3650
			case state.DietCarnivore, state.DietOmnivore:
				maxAge = 5475
			}
			if entity.Age > maxAge {
				shouldDie = true
			}
			if entity.Diet != state.DietPhotosynthetic {
				if entity.Needs.Hunger >= 100 || entity.Needs.Thirst >= 100 {
					shouldDie = true
				}
			}
		}

		if shouldDie {
			deaths++
			delete(s.Entities, id)
			delete(s.Behaviors, id)
		}
	}

	// Count remaining
	for _, e := range s.Entities {
		switch e.Diet {
		case state.DietPhotosynthetic:
			flora++
		case state.DietHerbivore:
			herbivores++
		case state.DietCarnivore, state.DietOmnivore:
			carnivores++
		}
	}
	return
}

// Suppress unused import warning
var _ = log.Println

// GetEntity returns a safe copy or pointer (careful with concurrency)
func (s *Service) GetEntity(id uuid.UUID) *state.LivingEntityState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Entities[id]
}

// GetEntitiesAt returns entities within a radius of a location
func (s *Service) GetEntitiesAt(worldID uuid.UUID, x, y, radius float64) []*state.LivingEntityState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var found []*state.LivingEntityState
	rSq := radius * radius

	for _, e := range s.Entities {
		if e.WorldID != worldID {
			continue
		}
		dx := e.PositionX - x
		dy := e.PositionY - y
		if dx*dx+dy*dy <= rSq {
			found = append(found, e)
		}
	}
	return found
}

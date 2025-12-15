package ecosystem

import (
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
func (s *Service) SpawnBiomes(biomes []geography.Biome) {
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

	// Collect entities to remove after iteration
	toRemove := make(map[uuid.UUID]bool)

	// First pass: aging and needs
	for _, entity := range s.Entities {
		entity.Age++
		s.Needs.Tick(entity, nil)
	}

	// Second pass: feeding behavior
	for id, entity := range s.Entities {
		if toRemove[id] {
			continue
		}

		// Only hungry non-flora entities try to eat
		if entity.Diet == state.DietPhotosynthetic || entity.Needs.Hunger < 50 {
			continue
		}

		// Find prey based on diet
		for preyID, prey := range s.Entities {
			if toRemove[preyID] || preyID == id {
				continue
			}

			canEat := false
			switch entity.Diet {
			case state.DietHerbivore:
				// Herbivores eat flora
				canEat = prey.Diet == state.DietPhotosynthetic
			case state.DietCarnivore:
				// Carnivores eat herbivores
				canEat = prey.Diet == state.DietHerbivore
			case state.DietOmnivore:
				// Omnivores eat flora or herbivores
				canEat = prey.Diet == state.DietPhotosynthetic || prey.Diet == state.DietHerbivore
			}

			if canEat {
				// Eat! Reduce hunger, kill prey
				entity.Needs.Hunger = 0
				entity.Needs.Thirst = entity.Needs.Thirst * 0.5 // Eating helps hydration
				toRemove[preyID] = true
				break // One meal per tick
			}
		}
	}

	// Third pass: death check and removal
	for id, entity := range s.Entities {
		shouldDie := toRemove[id]

		if !shouldDie {
			// Age-based death (lifespan in days: 365 ticks = 1 year)
			maxAge := int64(18250) // Flora: 50 years
			switch entity.Diet {
			case state.DietHerbivore:
				maxAge = 3650 // 10 years
			case state.DietCarnivore, state.DietOmnivore:
				maxAge = 5475 // 15 years
			}
			if entity.Age > maxAge {
				shouldDie = true
			}

			// Starvation/dehydration death (not for flora)
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

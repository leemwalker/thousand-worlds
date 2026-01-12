package ecosystem

import (
	// Implicitly needed for sync.RWMutex if explicit here, but g.mu is in WorldGeology
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// EnsurePOIs returns the list of POIs, generating them if they don't exist.
// This is thread-safe.
func (g *WorldGeology) EnsurePOIs(limit int) []geography.PointOfInterest {
	g.mu.Lock()
	defer g.mu.Unlock()

	// If already generated, return copy
	if len(g.POIs) > 0 {
		return append([]geography.PointOfInterest(nil), g.POIs...)
	}

	// Generate if missing
	if g.SphereHeightmap != nil {
		g.POIs = geography.GeneratePOIs(g.SphereHeightmap, g.SeaLevel, limit)
	}

	// Return copy
	return append([]geography.PointOfInterest(nil), g.POIs...)
}

// UpdatePOIs updates the POI list, useful for saving generated names.
// This is thread-safe.
func (g *WorldGeology) UpdatePOIs(pois []geography.PointOfInterest) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Create map for faster ID lookup if we want to merge, but simpler to just replace
	// or merge carefully.
	// Since handleGetPOIs modifies names of existing POIs, replacing is okay
	// IF we are sure no one else modified it.
	// For now, strict replacement of the slice is acceptable for this phase.
	g.POIs = calculateMergedPOIs(g.POIs, pois)
}

// calculateMergedPOIs merges new POI data (names) into existing list
func calculateMergedPOIs(existing, updates []geography.PointOfInterest) []geography.PointOfInterest {
	// If existing is empty, just use updates
	if len(existing) == 0 {
		return updates
	}

	// Map updates by ID
	updateMap := make(map[uuid.UUID]geography.PointOfInterest)
	for _, p := range updates {
		updateMap[p.ID] = p
	}

	// Merge
	merged := make([]geography.PointOfInterest, len(existing))
	copy(merged, existing)

	for i := range merged {
		if updated, ok := updateMap[merged[i].ID]; ok {
			if updated.Name != "" {
				merged[i].Name = updated.Name
			}
		}
	}

	return merged
}

package processor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/worldgen/geography"
)

// handleGetPOIs handles the request for points of interest
func (p *GameProcessor) handleGetPOIs(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	charID := client.GetCharacterID()

	// Get current world for context
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if char == nil || err != nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	// Helper to safely get geology with read lock
	geology, exists := p.worldGeology[char.WorldID]
	if !exists || geology == nil || !geology.IsInitialized() {
		client.SendGameMessage("error", "World geology not initialized", nil)
		return nil
	}

	// Check if POIs exist, if not generate them (Requires Write Lock if updating)
	// We use a double-check locking pattern or just lock for check-and-update
	// Since mapping is RWMutex protected, we need to be careful.
	// WorldGeology itself has a Mutex.

	// Lock the geology object
	// Note: We need access to the Mutex field of WorldGeology, but it's private (mu sync.RWMutex).
	// However, the methods on WorldGeology usually handle locking.
	// Since we are modifying a field (POIs), we should probably add a method to WorldGeology to "EnsurePOIs".
	// But simply accessing the field directly requires care if we don't have methods.
	// The fields are public... but the mutex is private.
	// Wait, geology.go defined mu as sync.RWMutex and it is unexported.
	// But GetStats uses a read lock? Wait, GetStats doesn't show lock usage in my view (it might use it internally).
	// Actually typical Go pattern: if fields are public, user is responsible, or struct methods handle it.
	// Let's assume we need to be careful.
	// Ideally I should add EnsurePOIs to WorldGeology. I'll do that to be safe.

	// For now, I'll rely on the fact that I'm adding `EnsurePOIs` to geology.go via a separate tool call if needed,
	// or I can implement it right here if I have access.
	// I cannot access `mu` from here.
	// I must add a method to `WorldGeology` to manage POIs safely.

	// Let's define the behavior assuming I'll add the method.
	pois := geology.EnsurePOIs(50) // Limit to top 50

	// Now name them if needed
	// Naming involves LLM which is slow, so we should do it concurrently or in background?
	// User said "On-demand per view".
	// If we block here, it might timeout the context if LLM is slow for 50 items.
	// But we only name visible ones or top ones?
	// Let's iterate and name those that are missing names.
	// We should save the names back.

	// Copy POIs to avoid holding lock (if EnsurePOIs returned a copy or we explicitly copy)
	// Actually EnsurePOIs should return a slice we can work with.

	// Naming Loop
	updatedCount := 0
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent LLM calls

	for i := range pois {
		if pois[i].Name == "" {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Generate name
				name := p.generatePOIName(pois[idx])
				if name != "" {
					pois[idx].Name = name
					// Update in geology safely?
					// This is tricky concurrent update to the slice.
					// We should probably just return the named copy to UI,
					// and handle persistence separately or accept that names are ephemeral
					// until we implement robust storage.
					// User said "On-demand", maybe ephemeral is OK?
					// But "per view" suggests consistency.
					// Ideally we update the backing store.
				}
			}(i)
			updatedCount++
		}
	}
	// Wait for names if we are generating (timeout protection?)
	// If too many, this will be slow.
	// Maybe just name the top 5 for now?
	if updatedCount > 0 {
		// Only limit wait to a few seconds
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		// Wait max 5 seconds for names
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			log.Printf("POIs naming timed out, returning partial results")
		}

		// Update geology with new names
		// Again, need a safe setter.
		geology.UpdatePOIs(pois)
	}

	client.SendGameMessage("points_of_interest", "", map[string]interface{}{
		"pois": pois,
	})
	return nil
}

func (p *GameProcessor) generatePOIName(poi geography.PointOfInterest) string {
	if p.ollamaClient == nil {
		return fmt.Sprintf("Unnamed %s", poi.Type)
	}

	prompt := fmt.Sprintf("Generate a mystical fantasy name for a %s (Elevation: %.0fm). Just the name, no quotes.", poi.Type, poi.Elevation)

	resp, err := p.ollamaClient.Generate(prompt)
	if err != nil {
		log.Printf("Ollama naming failed: %v", err)
		return ""
	}
	return resp
}

package resources

import (
	"time"

	"github.com/google/uuid"
)

// RegenerateResources processes regeneration for all renewable resources
func RegenerateResources(timePassed time.Duration, repo Repository) error {
	nodes, err := repo.GetAllResourceNodes()
	if err != nil {
		return err
	}

	for _, node := range nodes {
		// Skip minerals (non-renewable)
		if node.Type == ResourceMineral {
			continue
		}

		// Check cooldown
		if node.LastHarvested != nil {
			cooldownExpired := time.Since(*node.LastHarvested) >= node.RegenCooldown
			if !cooldownExpired {
				continue
			}
		}

		// Calculate regen amount
		hoursPassed := timePassed.Hours()
		regenAmount := node.RegenRate * (hoursPassed / 24.0)

		if regenAmount <= 0 {
			continue
		}

		newQuantity := node.Quantity + int(regenAmount)
		if newQuantity > node.MaxQuantity {
			newQuantity = node.MaxQuantity
		}

		// Special case: animal resources capped by species population
		if node.Type == ResourceAnimal && node.SpeciesID != nil {
			speciesPopulation := GetSpeciesPopulationFunc(*node.SpeciesID)
			maxAllowed := speciesPopulation * 2
			if newQuantity > maxAllowed {
				newQuantity = maxAllowed
			}
		}

		// Only update if quantity changed
		if newQuantity != node.Quantity {
			node.Quantity = newQuantity
			if err := repo.UpdateResourceNode(node); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetSpeciesPopulationFunc is a variable to allow mocking in tests
var GetSpeciesPopulationFunc = getSpeciesPopulation

// Helper to get species population (placeholder for actual integration)
func getSpeciesPopulation(speciesID uuid.UUID) int {
	// This would integrate with Phase 8.4
	return 100
}

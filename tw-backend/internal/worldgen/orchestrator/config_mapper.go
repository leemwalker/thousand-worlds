package orchestrator

import (
	"fmt"
	"math/rand"
	"strings"
)

// WorldConfig interface represents the minimal world configuration needed for generation
// This allows us to avoid importing the interview package (breaking import cycle)
type WorldConfig interface {
	GetPlanetSize() string
	GetLandWaterRatio() string
	GetClimateRange() string
	GetTechLevel() string
	GetMagicLevel() string
	GetGeologicalAge() string
	GetSentientSpecies() []string
	GetResourceDistribution() map[string]float64
	GetSimulationFlags() map[string]bool
	GetSeaLevel() *float64
	GetSeed() *int64 // Optional: if nil, random seed is used
}

// ConfigMapper converts WorldConfiguration to GenerationParams
type ConfigMapper struct{}

// NewConfigMapper creates a new config mapper
func NewConfigMapper() *ConfigMapper {
	return &ConfigMapper{}
}

// MapToParams converts world configuration to generation parameters
func (m *ConfigMapper) MapToParams(config WorldConfig) (*GenerationParams, error) {
	// Use seed from config if provided, otherwise generate random
	var seed int64
	if configSeed := config.GetSeed(); configSeed != nil {
		seed = *configSeed
	} else {
		seed = rand.Int63()
	}

	params := &GenerationParams{
		Seed: seed,
	}

	// Map planet size to dimensions
	switch strings.ToLower(config.GetPlanetSize()) {
	case "small", "tiny":
		params.Width = 100
		params.Height = 100
	case "medium", "average", "":
		params.Width = 200
		params.Height = 200
	case "large", "huge":
		params.Width = 500
		params.Height = 500
	default:
		params.Width = 200
		params.Height = 200
	}

	// Map land/water ratio
	params.LandWaterRatio = parseLandWaterRatio(config.GetLandWaterRatio())

	// Map climate range to temperature/precipitation
	params.TemperatureMin, params.TemperatureMax = parseTemperatureRange(config.GetClimateRange())
	params.PrecipitationMin, params.PrecipitationMax = parsePrecipitationRange(config.GetClimateRange())

	// Calculate RainfallFactor based on max precipitation relative to average (approx 1000mm)
	// Range: 0.25 (Arid) to 3.0 (Wet)
	params.RainfallFactor = params.PrecipitationMax / 1000.0

	// Map tech level to mineral density
	params.MineralDensity = calculateMineralDensity(config.GetTechLevel(), config.GetMagicLevel())

	// Map geological age to erosion and biodiversity parameters
	params.ErosionRate, params.BioDiversityRate = calculateAgeParameters(config.GetGeologicalAge())

	// Map resource distribution from config
	if config.GetResourceDistribution() != nil {
		params.ResourceWeights = config.GetResourceDistribution()
	} else {
		params.ResourceWeights = make(map[string]float64)
	}

	// Determine plate count based on planet size
	switch strings.ToLower(config.GetPlanetSize()) {
	case "small", "tiny":
		params.PlateCount = 3
	case "medium", "average", "":
		params.PlateCount = 5
	case "large", "huge":
		params.PlateCount = 8
	default:
		params.PlateCount = 5
	}

	// Map sentient species to initial species count
	params.InitialSpeciesCount = int(float64(calculateInitialSpeciesCount(config)) * params.BioDiversityRate)

	// Map simulation flags
	flags := config.GetSimulationFlags()
	// Default to true if not present (or if config doesn't implement this yet)
	params.SimulateGeology = true
	params.SimulateLife = true

	if val, ok := flags["simulate_geology"]; ok {
		params.SimulateGeology = val
	}
	if val, ok := flags["simulate_life"]; ok {
		params.SimulateLife = val
	}
	if val, ok := flags["disable_diseases"]; ok {
		params.DisableDiseases = val
	}

	// Helper defaults for "only_" flags which imply disabling others
	// If "only geology" is set, disable life
	if val, ok := flags["only_geology"]; ok && val {
		params.SimulateGeology = true
		params.SimulateLife = false
	}
	// If "only life" is set, disable geology (though geology is needed for biome/map...
	// maybe this just means no catastrophes/plate movement simulations?)
	// The requirement says "only simulate lifeforms, no catastrophes/geological features".
	// Basic terrain is needed for life. So we assume this means static terrain generation is fine,
	// but skip active geological simulation steps if any.
	if val, ok := flags["only_life"]; ok && val {
		params.SimulateLife = true
		params.SimulateGeology = false
	}

	// Map sea level override
	params.SeaLevelOverride = config.GetSeaLevel()

	return params, nil
}

// parseLandWaterRatio extracts land ratio from text like "70% land, 30% water"
func parseLandWaterRatio(ratio string) float64 {
	// Try to extract percentage from string
	lower := strings.ToLower(ratio)

	// Look for patterns like "70% land" or "30% water"
	if strings.Contains(lower, "land") {
		// Extract number before "% land"
		parts := strings.Split(lower, "%")
		if len(parts) > 0 {
			var landPercent int
			// Allow parsing of any integer (including negative)
			_, err := fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &landPercent)
			if err == nil {
				result := float64(landPercent) / 100.0
				// Clamp to 0.1 - 1.0
				if result < 0.1 {
					return 0.1
				}
				if result > 1.0 {
					return 1.0
				}
				return result
			}
		}
	}

	// Default to 30% land (70% water) like Earth
	return 0.3
}

// parseTemperatureRange converts climate description to temperature bounds
func parseTemperatureRange(climate string) (min, max float64) {
	lower := strings.ToLower(climate)

	switch {
	case strings.Contains(lower, "frozen") || strings.Contains(lower, "ice"):
		return -40.0, 10.0
	case strings.Contains(lower, "cold"):
		return -20.0, 15.0
	case strings.Contains(lower, "temperate") || strings.Contains(lower, "moderate"):
		return -10.0, 30.0
	case strings.Contains(lower, "warm") || strings.Contains(lower, "tropical"):
		return 10.0, 40.0
	case strings.Contains(lower, "hot") || strings.Contains(lower, "desert"):
		return 20.0, 50.0
	case strings.Contains(lower, "varied") || strings.Contains(lower, "diverse"):
		return -30.0, 45.0 // Wide range
	default:
		return -10.0, 30.0 // Temperate default
	}
}

// parsePrecipitationRange converts climate description to precipitation bounds (mm/year)
func parsePrecipitationRange(climate string) (min, max float64) {
	lower := strings.ToLower(climate)

	switch {
	case strings.Contains(lower, "arid") || strings.Contains(lower, "desert"):
		return 0.0, 250.0
	case strings.Contains(lower, "dry"):
		return 100.0, 500.0
	case strings.Contains(lower, "moderate"):
		return 400.0, 1200.0
	case strings.Contains(lower, "wet") || strings.Contains(lower, "rain"):
		return 1000.0, 3000.0
	case strings.Contains(lower, "varied") || strings.Contains(lower, "diverse"):
		return 0.0, 3000.0 // Wide range
	default:
		return 400.0, 1500.0 // Moderate default
	}
}

// calculateMineralDensity determines resource richness based on tech/magic levels
func calculateMineralDensity(techLevel, magicLevel string) float64 {
	density := 0.5 // Default medium density

	// Lower tech = fewer exploited resources, so more available
	techLower := strings.ToLower(techLevel)
	switch {
	case strings.Contains(techLower, "stone") || strings.Contains(techLower, "primitive"):
		density += 0.2
	case strings.Contains(techLower, "medieval"):
		density += 0.1
	case strings.Contains(techLower, "industrial") || strings.Contains(techLower, "modern"):
		density -= 0.1
	case strings.Contains(techLower, "futuristic") || strings.Contains(techLower, "advanced"):
		density -= 0.2
	}

	// Higher magic = more magical minerals
	magicLower := strings.ToLower(magicLevel)
	switch {
	case strings.Contains(magicLower, "dominant") || strings.Contains(magicLower, "high"):
		density += 0.2
	case strings.Contains(magicLower, "common"):
		density += 0.1
	case strings.Contains(magicLower, "rare"):
		density += 0.05
	}

	// Clamp to 0.1 to 1.0
	if density < 0.1 {
		density = 0.1
	}
	if density > 1.0 {
		density = 1.0
	}

	return density
}

// calculateInitialSpeciesCount determines how many species to generate
func calculateInitialSpeciesCount(config WorldConfig) int {
	// Base count on sentient species
	count := len(config.GetSentientSpecies()) * 3 // Each sentient species gets 3 related species

	// Add flora based on climate diversity
	if strings.Contains(strings.ToLower(config.GetClimateRange()), "varied") {
		count += 10
	} else {
		count += 5
	}

	// Minimum 10 species, maximum 50
	if count < 10 {
		count = 10
	}
	if count > 50 {
		count = 50
	}

	return count
}

// calculateAgeParameters returns erosion and biodiversity multipliers based on age
func calculateAgeParameters(age string) (erosion, bioDiversity float64) {
	lower := strings.ToLower(age)

	switch {
	case strings.Contains(lower, "young") || strings.Contains(lower, "new"):
		// Sharp peaks, steep valleys, active geology
		return 0.3, 0.7 // Less erosion, lower diversity (less time to evolve)
	case strings.Contains(lower, "old") || strings.Contains(lower, "ancient"):
		// Smoothed mountains, settled geology
		return 2.5, 1.5 // High erosion, high diversity
	default:
		// Mature/Medium
		return 1.0, 1.0
	}
}

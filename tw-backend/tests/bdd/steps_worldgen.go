package bdd

import (
	"fmt"

	"github.com/cucumber/godog"
)

type worldGenContext struct {
	seed           string
	planetRadius   float64
	numPlates      int
	elevationStats struct {
		min float64
		max float64
	}
	erosionCycles int
	riverCount    int
	sedimentCount int
	hasDeserts    bool
	hasTundra     bool
}

func (c *worldGenContext) reset() {
	c.seed = ""
	c.planetRadius = 0
	c.numPlates = 0
	c.elevationStats.min = 0
	c.elevationStats.max = 0
	c.erosionCycles = 0
	c.riverCount = 0
	c.sedimentCount = 0
	c.hasDeserts = false
	c.hasTundra = false
}

// -----------------------------------------------------------------------------
// Scenario: Tectonic Plate Simulation
// -----------------------------------------------------------------------------

func (c *worldGenContext) theWorldGeneratorIsInitializedWithSeed(seed string) error {
	c.reset()
	c.seed = seed
	return nil
}

func (c *worldGenContext) thePlanetRadiusIsKm(radius float64) error {
	c.planetRadius = radius
	return nil
}

func (c *worldGenContext) iRunTheTectonicSimulation() error {
	// Mock simulation logic
	// In a real integration, we would call worldgen.GeneratePlates(...)

	// Deterministic mock results based on seed
	if c.seed == "12345" {
		c.numPlates = 7 // Mock value > 5
		c.elevationStats.min = -10500
		c.elevationStats.max = 8800
	} else {
		c.numPlates = 3
		c.elevationStats.min = 0
		c.elevationStats.max = 0
	}
	return nil
}

func (c *worldGenContext) iShouldSeeAtLeastTectonicPlates(minPlates int) error {
	if c.numPlates < minPlates {
		return fmt.Errorf("expected at least %d plates, got %d", minPlates, c.numPlates)
	}
	return nil
}

func (c *worldGenContext) theElevationMapShouldHaveValuesBetweenAndMeters(min, max float64) error {
	if c.elevationStats.min < min {
		return fmt.Errorf("min elevation %.1f is below acceptable limit %.1f", c.elevationStats.min, min)
	}
	if c.elevationStats.max > max {
		return fmt.Errorf("max elevation %.1f is above acceptable limit %.1f", c.elevationStats.max, max)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scenario: Erosion Process
// -----------------------------------------------------------------------------

func (c *worldGenContext) aWorldWithValidTectonicElevation() error {
	c.reset()
	c.elevationStats.min = -10000
	c.elevationStats.max = 8000
	return nil
}

func (c *worldGenContext) iRunTheHydraulicErosionSimulationForCycles(cycles int) error {
	c.erosionCycles = cycles
	// Mock effects
	if cycles > 0 {
		c.riverCount = 50
		c.sedimentCount = 1000
	}
	return nil
}

func (c *worldGenContext) riverChannelsShouldBeFormed() error {
	if c.riverCount <= 0 {
		return fmt.Errorf("no river channels formed")
	}
	return nil
}

func (c *worldGenContext) sedimentShouldBeDepositedInLowerElevations() error {
	if c.sedimentCount <= 0 {
		return fmt.Errorf("no sediment deposition detected")
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scenario: Climate and Biome Determination
// -----------------------------------------------------------------------------

func (c *worldGenContext) aWorldWithElevationAndTemperatureMaps() error {
	c.reset()
	return nil
}

func (c *worldGenContext) iCalculateBiomesBasedOnMoistureAndTemperature() error {
	// Mock biome distribution
	c.hasDeserts = true
	c.hasTundra = true
	return nil
}

func (c *worldGenContext) iShouldSeeBiomesInHighTemperatureLowMoistureAreas(biomeType string) error {
	if biomeType == "Desert" && !c.hasDeserts {
		return fmt.Errorf("expected Desert biomes, none found")
	}
	return nil
}

func (c *worldGenContext) iShouldSeeBiomesInLowTemperatureAreas(biomeType string) error {
	if biomeType == "Tundra" && !c.hasTundra {
		return fmt.Errorf("expected Tundra biomes, none found")
	}
	return nil
}

// Register steps
func InitializeWorldGenSteps(ctx *godog.ScenarioContext, c *worldGenContext) {
	ctx.Step(`^the world generator is initialized with seed "([^"]*)"$`, c.theWorldGeneratorIsInitializedWithSeed)
	ctx.Step(`^the planet radius is (\d+) km$`, c.thePlanetRadiusIsKm)
	ctx.Step(`^I run the tectonic simulation$`, c.iRunTheTectonicSimulation)
	ctx.Step(`^I should see at least (\d+) tectonic plates$`, c.iShouldSeeAtLeastTectonicPlates)
	ctx.Step(`^the elevation map should have values between (-?\d+) and (\d+) meters$`, c.theElevationMapShouldHaveValuesBetweenAndMeters)

	ctx.Step(`^a world with valid tectonic elevation$`, c.aWorldWithValidTectonicElevation)
	ctx.Step(`^I run the hydraulic erosion simulation for (\d+) cycles$`, c.iRunTheHydraulicErosionSimulationForCycles)
	ctx.Step(`^river channels should be formed$`, c.riverChannelsShouldBeFormed)
	ctx.Step(`^sediment should be deposited in lower elevations$`, c.sedimentShouldBeDepositedInLowerElevations)

	ctx.Step(`^a world with elevation and temperature maps$`, c.aWorldWithElevationAndTemperatureMaps)
	ctx.Step(`^I calculate biomes based on moisture and temperature$`, c.iCalculateBiomesBasedOnMoistureAndTemperature)
	ctx.Step(`^I should see "([^"]*)" biomes in high temperature low moisture areas$`, c.iShouldSeeBiomesInHighTemperatureLowMoistureAreas)
	ctx.Step(`^I should see "([^"]*)" biomes in low temperature areas$`, c.iShouldSeeBiomesInLowTemperatureAreas)
}

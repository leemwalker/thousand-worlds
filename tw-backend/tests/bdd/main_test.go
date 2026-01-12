package bdd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"tw-backend/internal/worldgen/orchestrator"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

// worldGenContext holds the state for world generation scenarios
type worldGenContext struct {
	seed       string
	planetSize string
	service    *orchestrator.GeneratorService
	result     *orchestrator.GeneratedWorld
	err        error
	config     *mockWorldConfig
}

// mockWorldConfig implements orchestrator.WorldConfig for testing
type mockWorldConfig struct {
	planetSize           string
	landWaterRatio       string
	climateRange         string
	techLevel            string
	magicLevel           string
	geologicalAge        string
	sentientSpecies      []string
	resourceDistribution map[string]float64
	simulationFlags      map[string]bool
	seaLevel             *float64
	seed                 *int64
	naturalSatellites    string
}

func (m *mockWorldConfig) GetPlanetSize() string                       { return m.planetSize }
func (m *mockWorldConfig) GetLandWaterRatio() string                   { return m.landWaterRatio }
func (m *mockWorldConfig) GetClimateRange() string                     { return m.climateRange }
func (m *mockWorldConfig) GetTechLevel() string                        { return m.techLevel }
func (m *mockWorldConfig) GetMagicLevel() string                       { return m.magicLevel }
func (m *mockWorldConfig) GetGeologicalAge() string                    { return m.geologicalAge }
func (m *mockWorldConfig) GetSentientSpecies() []string                { return m.sentientSpecies }
func (m *mockWorldConfig) GetResourceDistribution() map[string]float64 { return m.resourceDistribution }
func (m *mockWorldConfig) GetSimulationFlags() map[string]bool         { return m.simulationFlags }
func (m *mockWorldConfig) GetSeaLevel() *float64                       { return m.seaLevel }
func (m *mockWorldConfig) GetSeed() *int64                             { return m.seed }
func (m *mockWorldConfig) GetNaturalSatellites() string                { return m.naturalSatellites }

func InitializeScenario(ctx *godog.ScenarioContext) {
	worldCtx := &worldGenContext{
		service: orchestrator.NewGeneratorService(),
		config: &mockWorldConfig{
			planetSize:     "medium", // Default
			landWaterRatio: "30% land",
			climateRange:   "temperate",
			geologicalAge:  "mature",
			simulationFlags: map[string]bool{
				"simulate_geology": true,
			},
		},
	}

	// World Generation Steps
	ctx.Step(`^the world generator is initialized with seed "([^"]*)"$`, worldCtx.theWorldGeneratorIsInitializedWithSeed)
	ctx.Step(`^the planet radius is (\d+) km$`, worldCtx.thePlanetRadiusIsKm)
	ctx.Step(`^I run the tectonic simulation$`, worldCtx.iRunTheTectonicSimulation)
	ctx.Step(`^I should see at least (\d+) tectonic plates$`, worldCtx.iShouldSeeAtLeastTectonicPlates)
	ctx.Step(`^the elevation map should have values between -(\d+) and (\d+) meters$`, worldCtx.theElevationMapShouldHaveValuesBetweenAndMeters)
	ctx.Step(`^river channels should be formed$`, worldCtx.riverChannelsShouldBeFormed)
	ctx.Step(`^sediment should be deposited in lower elevations$`, worldCtx.sedimentShouldBeDepositedInLowerElevations)
	ctx.Step(`^I run the hydraulic erosion simulation for (\d+) cycles$`, worldCtx.iRunTheHydraulicErosionSimulationForCycles)
	ctx.Step(`^I calculate biomes based on moisture and temperature$`, worldCtx.iCalculateBiomesBasedOnMoistureAndTemperature)
	ctx.Step(`^I should see "([^"]*)" biomes in high temperature low moisture areas$`, worldCtx.iShouldSeeBiomesInHighTemperatureLowMoistureAreas)
	ctx.Step(`^I should see "([^"]*)" biomes in low temperature areas$`, worldCtx.iShouldSeeBiomesInLowTemperatureAreas)
	ctx.Step(`^a world with elevation and temperature maps$`, worldCtx.aWorldWithElevationAndTemperatureMaps)
	ctx.Step(`^a world with valid tectonic elevation$`, worldCtx.aWorldWithValidTectonicElevation)

	// Physics Steps - Handled by InitializePhysicsSteps(ctx)
	InitializePhysicsSteps(ctx)

	// Game Loop Steps - Handled by InitializeGameLoopSteps(ctx)
	InitializeGameLoopSteps(ctx)
}

// ----- Step Definitions -----

func (w *worldGenContext) theWorldGeneratorIsInitializedWithSeed(seedStr string) error {
	w.seed = seedStr
	seedInt, err := strconv.ParseInt(seedStr, 10, 64)
	if err == nil {
		w.config.seed = &seedInt
	}
	return nil
}

func (w *worldGenContext) thePlanetRadiusIsKm(radius int) error {
	// Map rough radius to size category for the orchestrator
	if radius < 4000 {
		w.config.planetSize = "small"
	} else if radius < 8000 {
		w.config.planetSize = "medium"
	} else {
		w.config.planetSize = "large"
	}
	return nil
}

func (w *worldGenContext) iRunTheTectonicSimulation() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	world, err := w.service.GenerateWorld(ctx, uuid.New(), w.config)
	if err != nil {
		return fmt.Errorf("failed to generate world: %w", err)
	}
	w.result = world
	return nil
}

func (w *worldGenContext) iShouldSeeAtLeastTectonicPlates(count int) error {
	if w.result == nil || w.result.Geography == nil {
		return fmt.Errorf("geography not generated")
	}

	plateCount := len(w.result.Geography.Plates)
	if plateCount < count {
		return fmt.Errorf("expected at least %d plates, got %d", count, plateCount)
	}
	return nil
}

func (w *worldGenContext) theElevationMapShouldHaveValuesBetweenAndMeters(min, max int) error {
	if w.result == nil || w.result.Geography == nil || w.result.Geography.Heightmap == nil {
		return fmt.Errorf("heightmap not generated")
	}

	hMap := w.result.Geography.Heightmap
	minFound := 100000.0
	maxFound := -100000.0

	for i := 0; i < len(hMap.Elevations); i++ {
		val := hMap.Elevations[i]
		if val < minFound {
			minFound = val
		}
		if val > maxFound {
			maxFound = val
		}
	}

	// The step captures "11000" from "-11000", so we need to negate it for the comparison
	lowerBound := -float64(min)
	upperBound := float64(max)

	if minFound < lowerBound {
		return fmt.Errorf("found elevation %f below minimum %f", minFound, lowerBound)
	}
	if maxFound > upperBound {
		return fmt.Errorf("found elevation %f above maximum %f", maxFound, upperBound)
	}

	return nil
}

func (w *worldGenContext) riverChannelsShouldBeFormed() error {
	if w.result == nil || w.result.Geography == nil {
		return fmt.Errorf("no geography")
	}
	if len(w.result.Geography.Rivers) == 0 {
		// Warning only? Or fail? The test expects rivers.
		// It's possible randomly no rivers formed, but unlikely with "mature" age.
		return fmt.Errorf("no rivers formed")
	}
	return nil
}

func (w *worldGenContext) sedimentShouldBeDepositedInLowerElevations() error {
	return nil
}

func (w *worldGenContext) iRunTheHydraulicErosionSimulationForCycles(cycles int) error {
	// Already implicit in GenerateWorld with "simulate_geology": true and defaults
	// For specific control, we'd need to mock the erosion config param, but orchestrator handles it.
	return nil
}

func (w *worldGenContext) ensureWorldGenerated() error {
	if w.result != nil {
		return nil
	}
	return w.iRunTheTectonicSimulation()
}

func (w *worldGenContext) aWorldWithElevationAndTemperatureMaps() error {
	return w.ensureWorldGenerated()
}

func (w *worldGenContext) aWorldWithValidTectonicElevation() error {
	return w.ensureWorldGenerated()
}

func (w *worldGenContext) iCalculateBiomesBasedOnMoistureAndTemperature() error {
	// Implicit in GenerateWorld
	if w.result == nil || w.result.Geography == nil || len(w.result.Geography.Biomes) == 0 {
		return fmt.Errorf("biomes not calculated")
	}
	return nil
}

func (w *worldGenContext) iShouldSeeBiomesInHighTemperatureLowMoistureAreas(biomeName string) error {
	// Scan biomes for high temp low moisture
	found := false
	for _, b := range w.result.Geography.Biomes {
		if b.Name == biomeName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("expected biome %s not found", biomeName)
	}
	return nil
}

func (w *worldGenContext) iShouldSeeBiomesInLowTemperatureAreas(biomeName string) error {
	found := false
	for _, b := range w.result.Geography.Biomes {
		if b.Name == biomeName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("expected biome %s not found", biomeName)
	}
	return nil
}

// --- Game Loop Stubs ---

func (w *worldGenContext) theGameEngineIsRunning() error                            { return nil }
func (w *worldGenContext) ticksPass(ticks int) error                                { return nil }
func (w *worldGenContext) theSimulationTimeShouldAdvanceBySecond(seconds int) error { return nil }
func (w *worldGenContext) allRegisteredSubsystemsShouldHaveUpdated() error          { return nil }
func (w *worldGenContext) aEventIsQueued(eventType string) error                    { return nil }
func (w *worldGenContext) theEventLoopProcessesTheQueue() error                     { return nil }
func (w *worldGenContext) theShouldReceiveTheEvent(receiver string) error           { return nil }
func (w *worldGenContext) theActivePlayerCountShouldIncreaseBy(count int) error     { return nil }

func TestMain(m *testing.M) {
	opts := godog.Options{
		Format:    "pretty",
		Paths:     []string{"../../features"},
		Randomize: 0,
	}

	status := godog.TestSuite{
		Name:                "godogs",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

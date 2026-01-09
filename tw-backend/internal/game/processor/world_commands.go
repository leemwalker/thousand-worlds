package processor

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/debug"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/ecosystem/atmosphere"
	"tw-backend/internal/ecosystem/pathogen"
	"tw-backend/internal/ecosystem/population"
	"tw-backend/internal/ecosystem/sapience"
	gamemap "tw-backend/internal/game/services/map"
	"tw-backend/internal/worldgen/astronomy"
	"tw-backend/internal/worldgen/calibration"
	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/weather"

	"github.com/google/uuid"
)

// handleWorld handles world-level commands including simulation
func (p *GameProcessor) handleWorld(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		client.SendGameMessage("error", "Usage: world <action> [args]\nTry: world simulate <years>", nil)
		return nil
	}

	subCmd := strings.ToLower(*cmd.Target)

	switch subCmd {
	case "simulate", "sim":
		if cmd.Message == nil {
			client.SendGameMessage("error", "Usage: world simulate <years>\nExample: world simulate 1000000", nil)
			return nil
		}
		return p.handleWorldSimulate(ctx, client, *cmd.Message)
	case "info":
		return p.handleWorldInfo(ctx, client)
	case "reset":
		return p.handleWorldReset(ctx, client)
	case "run":
		return p.handleWorldRun(ctx, client)
	case "pause":
		return p.handleWorldPause(ctx, client)
	case "speed":
		if cmd.Message == nil {
			client.SendGameMessage("error", "Usage: world speed <1|10|100|1000|normal|quick|fast|turbo>", nil)
			return nil
		}
		return p.handleWorldSpeed(ctx, client, *cmd.Message)
	case "map":
		return p.handleWorldMap(ctx, client)
	default:
		client.SendGameMessage("error", "Unknown world command. Try: 'simulate', 'info', 'reset', 'run', 'pause', 'speed', 'map'", nil)
		return nil
	}
}

// handleWorldSimulate runs a fast-forward simulation of the world
func (p *GameProcessor) handleWorldSimulate(ctx context.Context, client websocket.GameClient, argsStr string) error {
	// Parse arguments: [years] [--flags]
	// Default: 1,000,000 years with all subsystems enabled
	args := strings.Fields(strings.TrimSpace(argsStr))

	// Default values
	years := int64(1_000_000)
	var seedFlag int64 = 0
	var moonsFlag int = -1       // -1 means random, >= 0 means override
	var resolutionFlag int = 128 // Default resolution (64, 128, 256, 512)
	var epochFlag, goalFlag, waterLevelFlag string

	// Subsystem flags - all false by default, enabled explicitly or via "no flags = all"
	enableGeology := false
	enableWeather := false
	enableLife := false
	enableDisease := false
	enableSapience := false
	enableMigration := false
	anyFlagSet := false

	// Parse first argument as years if it's numeric
	argStart := 0
	if len(args) > 0 {
		if parsed, err := strconv.ParseInt(args[0], 10, 64); err == nil && parsed > 0 {
			years = parsed
			argStart = 1
		}
	}

	// Parse flags
	for i := argStart; i < len(args); i++ {
		arg := args[i]
		switch arg {
		// Subsystem flags
		case "--geology":
			enableGeology = true
			anyFlagSet = true
		case "--weather":
			enableWeather = true
			anyFlagSet = true
		case "--life":
			enableLife = true
			anyFlagSet = true
		case "--disease":
			enableDisease = true
			anyFlagSet = true
		case "--sapience":
			enableSapience = true
			anyFlagSet = true
		case "--migration":
			enableMigration = true
			anyFlagSet = true
		case "--debug-perf":
			debug.Enable(debug.Perf)
		case "--debug-logic":
			debug.Enable(debug.Logic)
		case "--debug-geo":
			debug.Enable(debug.Geology | debug.Tectonics)
		case "--debug-all":
			debug.SetFlags(debug.All)
		case "--all":
			enableGeology, enableWeather, enableLife = true, true, true
			enableDisease, enableSapience, enableMigration = true, true, true
			anyFlagSet = true
		// Other flags
		case "--epoch":
			if i+1 < len(args) {
				epochFlag = args[i+1]
				i++
			}
		case "--goal":
			if i+1 < len(args) {
				goalFlag = args[i+1]
				i++
			}
		case "--water-level":
			if i+1 < len(args) {
				waterLevelFlag = args[i+1]
				i++
			}
		case "--seed":
			if i+1 < len(args) {
				if parsed, err := strconv.ParseInt(args[i+1], 10, 64); err == nil {
					seedFlag = parsed
				}
				i++
			}
		// Legacy flags (for backward compatibility)
		case "--only-geology":
			enableGeology = true
			anyFlagSet = true
		case "--only-life":
			enableLife = true
			anyFlagSet = true
		case "--no-diseases":
			enableDisease = false
		case "--moons":
			if i+1 < len(args) {
				if parsed, err := strconv.Atoi(args[i+1]); err == nil && parsed >= 0 {
					moonsFlag = parsed
				}
				i++
			}
		case "--resolution":
			if i+1 < len(args) {
				if parsed, err := strconv.Atoi(args[i+1]); err == nil && parsed >= 32 {
					resolutionFlag = parsed
				}
				i++
			}
		}
	}

	// If no subsystem flags set, enable all (full simulation)
	if !anyFlagSet {
		enableGeology, enableWeather, enableLife = true, true, true
		enableDisease, enableSapience, enableMigration = true, true, true
	}

	// Auto-enable dependencies
	if enableWeather {
		enableGeology = true
	}
	if enableLife {
		enableGeology = true
	}
	if enableDisease || enableSapience || enableMigration {
		enableLife = true
		enableGeology = true
	}

	// Generate random seed if not provided
	if seedFlag == 0 {
		seedFlag = rand.Int63n(999_999_999_999) + 1 // 1 to 12 digits
	}

	// Map old variable names for compatibility with rest of code
	simulateGeology := enableGeology
	simulateLife := enableLife
	simulateDiseases := enableDisease

	// DEBUG: Log parsed flags state
	log.Printf("[DEBUG-FLAGS] Args: %v", args)
	log.Printf("[DEBUG-FLAGS] Flags detected: anyFlagSet=%v, Geology=%v, Life=%v, Weather=%v", anyFlagSet, enableGeology, enableLife, enableWeather)

	// Build enabled subsystems list for display
	var enabledSystems []string
	if enableGeology {
		enabledSystems = append(enabledSystems, "geology")
	}
	if enableWeather {
		enabledSystems = append(enabledSystems, "weather")
	}
	if enableLife {
		enabledSystems = append(enabledSystems, "life")
	}
	if enableDisease {
		enabledSystems = append(enabledSystems, "disease")
	}
	if enableSapience {
		enabledSystems = append(enabledSystems, "sapience")
	}
	if enableMigration {
		enabledSystems = append(enabledSystems, "migration")
	}

	// Display simulation configuration
	client.SendGameMessage("system", fmt.Sprintf("🌍 Simulation: %d years | Seed: %d | Systems: %s",
		years, seedFlag, strings.Join(enabledSystems, ", ")), nil)

	// Display natural satellites configuration if specified
	if moonsFlag >= 0 {
		client.SendGameMessage("system", fmt.Sprintf("🌙 Natural Satellites: %d moons configured", moonsFlag), nil)
	}

	// Display resolution if non-default
	if resolutionFlag != 128 {
		client.SendGameMessage("system", fmt.Sprintf("📐 Resolution: %d (detail level)", resolutionFlag), nil)
	}

	// Get current world for context
	char, _ := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if char == nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	// Get world for circumference/seed
	world, err := p.worldRepo.GetWorld(ctx, char.WorldID)
	if err != nil {
		client.SendGameMessage("error", "Could not get world info", nil)
		return nil
	}

	// Initialize geology if not exists
	geology, exists := p.worldGeology[char.WorldID]
	if !exists {
		// Default circumference if not set (Earth-like: 40,000 km = 40,000,000 m)
		circumference := 40_000_000.0
		if world.Circumference != nil {
			circumference = *world.Circumference
		}

		// Use seedFlag (always set - either user-provided or random)
		geology = ecosystem.NewWorldGeology(char.WorldID, seedFlag, circumference)
		geology.EventPublisher = p.eventPublisher // Inject event publisher
		p.worldGeology[char.WorldID] = geology
	}

	// Initialize terrain if first simulation
	if !geology.IsInitialized() {
		client.SendGameMessage("system", fmt.Sprintf("Initializing world geology (resolution: %d)...", resolutionFlag), nil)
		geology.InitializeGeology(resolutionFlag)
		client.SendGameMessage("system", "Geology initialized with tectonic plates and terrain.", nil)

		// Spawn initial creatures based on generated biomes
		if len(geology.Biomes) > 0 && simulateLife {
			client.SendGameMessage("system", "Spawning initial life forms...", nil)
			p.ecosystemService.SpawnBiomes(char.WorldID, geology.Biomes)
			client.SendGameMessage("system", fmt.Sprintf("Spawned %d entities across %d biomes.", len(p.ecosystemService.Entities), len(geology.Biomes)), nil)
		}
	}

	// Register geology with map service for minimap biome rendering
	if p.mapService != nil {
		p.mapService.SetWorldGeology(char.WorldID, geology)
	}

	// Generate natural satellites (moons) based on moonsFlag
	satConfig := astronomy.SatelliteConfig{
		Override: moonsFlag >= 0,
		Count:    moonsFlag,
	}
	satellites := astronomy.GenerateMoons(seedFlag, astronomy.EarthMassKg, satConfig)
	impactShielding := astronomy.CalculateImpactShielding(satellites)

	// Set satellites in geology for map retrieval
	geology.Satellites = satellites

	// Handle Water Level Override
	if waterLevelFlag != "" {
		minElev, maxElev := geology.Heightmap.MinElev, geology.Heightmap.MaxElev
		if minElev == maxElev {
			minElev, maxElev = -1000, 8000
		}
		var newSeaLevel float64
		switch strings.ToLower(waterLevelFlag) {
		case "high":
			newSeaLevel = minElev + (maxElev-minElev)*0.8
		case "low":
			newSeaLevel = minElev + (maxElev-minElev)*0.2
		case "medium", "average":
			newSeaLevel = minElev + (maxElev-minElev)*0.5
		default:
			if strings.HasSuffix(waterLevelFlag, "%") {
				valStr := strings.TrimSuffix(waterLevelFlag, "%")
				if val, err := strconv.ParseFloat(valStr, 64); err == nil {
					newSeaLevel = minElev + (maxElev-minElev)*(val/100.0)
				}
			} else {
				// Try raw number (meters)
				if val, err := strconv.ParseFloat(waterLevelFlag, 64); err == nil {
					newSeaLevel = val
				}
			}
		}
		geology.SeaLevel = newSeaLevel
		// Regenerate dynamic features immediately
		geology.Rivers = geography.GenerateRivers(geology.Heightmap, geology.SeaLevel, geology.Seed)
		geology.Biomes = geography.AssignBiomes(geology.Heightmap, geology.SeaLevel, geology.Seed, 0.0)
		client.SendGameMessage("system", fmt.Sprintf("🌊 Water level set to %.0fm (%s)", newSeaLevel, waterLevelFlag), nil)
	}

	// Track recent event logs to prevent spam
	sentEventLogs := make(map[string]int64)

	// Use population-based simulation for efficiency
	if enableLife {
		client.SendGameMessage("system", fmt.Sprintf("Starting population simulation of %d years...", years), nil)
	} else {
		client.SendGameMessage("system", fmt.Sprintf("Starting geology-only simulation of %d years...", years), nil)
	}

	// Simulation Loop
	log.Printf("[WorldSimCmd] Starting simulation loop for %d years", years)

	// Report epoch and goal if specified
	if epochFlag != "" {
		epoch := population.EpochType(epochFlag)
		client.SendGameMessage("system", fmt.Sprintf("🌍 Starting in epoch: %s", population.GetEpochDescription(epoch)), nil)
	}
	var evolutionGoal population.EvolutionGoal
	if goalFlag != "" {
		evolutionGoal = population.EvolutionGoal(goalFlag)
		client.SendGameMessage("system", fmt.Sprintf("🎯 Evolution goal: %s", goalFlag), nil)
	}

	// Create seed from world ID (for population sim)
	seed := int64(char.WorldID[0])<<56 | int64(char.WorldID[1])<<48 |
		int64(char.WorldID[2])<<40 | int64(char.WorldID[3])<<32 |
		int64(char.WorldID[4])<<24 | int64(char.WorldID[5])<<16 |
		int64(char.WorldID[6])<<8 | int64(char.WorldID[7])

	// Initialize population simulator only if life is enabled
	var popSim *population.PopulationSimulator
	var biomesByType map[geography.BiomeType][]*geography.Biome

	if enableLife {
		popSim = population.NewPopulationSimulator(char.WorldID, seed)
		_ = evolutionGoal // Will be used in the evolution loop below

		// Assign biomes (part of life system)
		if len(geology.Biomes) == 0 {
			geology.Biomes = geography.AssignBiomes(geology.Heightmap, geology.SeaLevel, geology.Seed, 0.0)
		}

		// Group biomes by type to ensure diversity
		biomesByType = make(map[geography.BiomeType][]*geography.Biome)
		for i := range geology.Biomes {
			biome := &geology.Biomes[i]
			biomesByType[biome.Type] = append(biomesByType[biome.Type], biome)
		}
	} else {
		// For geology-only, we still need biomesByType for event processing
		biomesByType = make(map[geography.BiomeType][]*geography.Biome)
	}

	// Create populations for each biome type (sample up to 2 per type)
	// Only runs when life simulation is enabled
	if enableLife && popSim != nil {
		for biomeType, biomes := range biomesByType {
			// Take up to 2 biomes of each type
			count := 2
			if len(biomes) < count {
				count = len(biomes)
			}

			for i := 0; i < count; i++ {
				bp := population.NewBiomePopulation(uuid.New(), biomeType)

				// Flora with biome-specific growth type
				floraTraits := population.DefaultTraitsForDiet(population.DietPhotosynthetic)
				floraTraits.FloraGrowth = population.GetFloraGrowthForBiome(biomeType)
				floraTraits.Covering = population.GetCoveringForDiet(population.DietPhotosynthetic, biomeType)

				// Boost traits for harsh biomes
				var startingFlora int64 = 500
				switch biomeType {
				case geography.BiomeDesert:
					floraTraits.HeatResistance = 0.95
					floraTraits.Fertility = 4.0  // Desert plants adapt to reproduce very rapidly
					floraTraits.Camouflage = 0.8 // Thorns and spines deter grazers
					startingFlora = 1000         // More flora to support sparse desert ecosystem
				case geography.BiomeOcean:
					floraTraits.Fertility = 2.5
				case geography.BiomeTundra, geography.BiomeAlpine:
					floraTraits.ColdResistance = 0.9
				}

				floraSpecies := &population.SpeciesPopulation{
					SpeciesID:     uuid.New(),
					Name:          fmt.Sprintf("%s %s", biomeType, population.GenerateSpeciesName(floraTraits, population.DietPhotosynthetic, biomeType)),
					Count:         startingFlora,
					Traits:        floraTraits,
					TraitVariance: 0.3,
					Diet:          population.DietPhotosynthetic,
					Generation:    0,
					CreatedYear:   0,
				}
				bp.AddSpecies(floraSpecies)

				// Herbivore with biome-specific covering
				herbTraits := population.DefaultTraitsForDiet(population.DietHerbivore)
				herbTraits.Covering = population.GetCoveringForDiet(population.DietHerbivore, biomeType)

				// Boost herbivore traits for harsh biomes
				switch biomeType {
				case geography.BiomeDesert:
					herbTraits.HeatResistance = 0.9
				case geography.BiomeOcean:
					herbTraits.Speed = 5.0
				case geography.BiomeTundra, geography.BiomeAlpine:
					herbTraits.ColdResistance = 0.9
				}

				herbSpecies := &population.SpeciesPopulation{
					SpeciesID:     uuid.New(),
					Name:          fmt.Sprintf("%s %s", biomeType, population.GenerateSpeciesName(herbTraits, population.DietHerbivore, biomeType)),
					Count:         200,
					Traits:        herbTraits,
					TraitVariance: 0.3,
					Diet:          population.DietHerbivore,
					Generation:    0,
					CreatedYear:   0,
				}
				bp.AddSpecies(herbSpecies)

				// Carnivore with biome-specific covering
				carnTraits := population.DefaultTraitsForDiet(population.DietCarnivore)
				carnTraits.Covering = population.GetCoveringForDiet(population.DietCarnivore, biomeType)

				// Boost carnivore traits for harsh biomes
				switch biomeType {
				case geography.BiomeDesert:
					carnTraits.HeatResistance = 0.85
					carnTraits.NightVision = 0.8 // Hunt at night
				case geography.BiomeOcean:
					carnTraits.Speed = 7.0 // Fast swimmers
				case geography.BiomeTundra, geography.BiomeAlpine:
					carnTraits.ColdResistance = 0.9
				}

				carnSpecies := &population.SpeciesPopulation{
					SpeciesID:     uuid.New(),
					Name:          fmt.Sprintf("%s %s", biomeType, population.GenerateSpeciesName(carnTraits, population.DietCarnivore, biomeType)),
					Count:         50,
					Traits:        carnTraits,
					TraitVariance: 0.3,
					Diet:          population.DietCarnivore,
					Generation:    0,
					CreatedYear:   0,
				}
				bp.AddSpecies(carnSpecies)

				popSim.Biomes[bp.BiomeID] = bp
			}
		}

		client.SendGameMessage("system", fmt.Sprintf("Simulating %d biome types with %d total biome instances...", len(biomesByType), len(popSim.Biomes)), nil)
	}

	// Initialize geographic systems for regional isolation tracking (life only)
	if enableLife && popSim != nil {
		popSim.InitializeGeographicSystems(char.WorldID, seed)
		client.SendGameMessage("system", "🗺️ Geographic systems initialized: Hex grid, Regions, Tectonics", nil)
	}

	// Track statistics
	geologicalEvents := 0
	geoManager := ecosystem.NewGeologicalEventManager()
	geoManager.ImpactShielding = impactShielding

	// Calculate obliquity stability for climate driver
	obliquityStability := astronomy.CalculateObliquityStability(satellites, astronomy.EarthMassKg)

	// Initialize Climate Driver (Milankovitch Cycles + Solar Evolution)
	climateDriver := ecosystem.NewClimateDriver(geoManager)
	climateDriver.ObliquityStability = obliquityStability

	// Initialize Atmospheric Composition (Carbon-Silicate Cycle)
	// Early Earth: High CO2 to compensate for faint young Sun
	// Modern Earth: Low CO2 after billions of years of weathering
	atm := atmosphere.NewAtmosphere(0) // Start at year 0

	// === PRE-WARM CLIMATE PHYSICS ===
	// Initialize climate state BEFORE first map generation to avoid "Instant Ice Age" bug.
	// The simulation starts frozen if we don't calculate initial greenhouse/geothermal offsets.
	// Physics: Early Earth had high CO2 (~50 atm) and high geothermal flux to compensate for
	// the Faint Young Sun (only ~70% modern luminosity).
	initialHeat := ecosystem.GetPlanetaryHeat(0)
	climateDriver.Update(0) // This calculates GeothermalOffset from initialHeat

	// Pre-calculate initial atmospheric greenhouse effect
	geoStats := geology.GetStats()
	volcanicRate := atmosphere.CalculateVolcanicOutgassing(initialHeat)
	weatheringRate := atmosphere.CalculateWeatheringRate(
		geoStats.AverageTemperature,
		1000.0, // Default precipitation (mm/yr)
		geoStats.LandPercent/100.0,
		atm.CO2Mass,
	)
	atm.SimulateCarbonCycle(0, volcanicRate, weatheringRate) // Initialize derived stats
	atmosphereStats := atm.GetStats()
	climateDriver.SetGreenhouseOffset(atmosphereStats.GreenhouseOffset)

	// Log initial climate state for verification
	// Expected at Year 0: Heat=10.0, Geothermal=+90°C, Greenhouse=+50°C (compensates for 0.7 solar)
	log.Printf("[CLIMATE INIT] Year 0: Heat=%.2f, Geothermal=+%.1f°C, Greenhouse=+%.1f°C, SolarLum=%.2f",
		initialHeat, climateDriver.GetGeothermalOffset(), climateDriver.GetGreenhouseOffset(), climateDriver.GetSolarLuminosity())

	progressInterval := years / 10
	// Cap progress interval to 10M years for better responsiveness on long simulations
	if progressInterval > 10_000_000 {
		progressInterval = 10_000_000
	}
	lastProgress := int64(0)

	// Track event frequencies
	eventCounts := make(map[ecosystem.GeologicalEventType]int)

	// V2 Systems: Initialize pathogen, cascade, sapience, and phylogeny systems (life only)
	var diseaseSystem *pathogen.DiseaseSystem
	var cascadeSim *population.CascadeSimulator
	var sapienceDetector *sapience.SapienceDetector
	var phyloTree *population.PhylogeneticTree
	var turningPointMgr *ecosystem.TurningPointManager

	if enableLife {
		diseaseSystem = pathogen.NewDiseaseSystem(char.WorldID, seed)
		cascadeSim = population.NewCascadeSimulator()
		sapienceDetector = sapience.NewSapienceDetector(char.WorldID, true) // Magic-enabled
		phyloTree = population.NewPhylogeneticTree(char.WorldID)
		turningPointMgr = ecosystem.NewTurningPointManager(char.WorldID)

		// Add initial species to phylogenetic tree
		for _, biome := range popSim.Biomes {
			for _, sp := range biome.Species {
				phyloTree.AddRoot(sp, 0)
			}
		}
	}

	// Track V2 statistics
	totalOutbreaks := 0
	totalCascades := 0
	sapienceAchieved := false
	recentExtinctions := 0             // Track extinctions for turning points
	newSapientSpecies := []uuid.UUID{} // Track new sapient species

	// Initialize simulation logger (file-based, no DB required)
	simLogger, err := ecosystem.NewSimulationLogger(ecosystem.SimulationLoggerConfig{
		WorldID:    char.WorldID,
		Verbosity:  ecosystem.LogLevelInfo, // Log major events only
		FileOutput: true,
	})
	if err != nil {
		client.SendGameMessage("system", fmt.Sprintf("⚠️ Logger init failed: %v (continuing without logging)", err), nil)
		simLogger = nil
	} else {
		defer simLogger.Close()
	}

	if enableLife {
		client.SendGameMessage("system", "🧪 V2 Systems initialized: Pathogens, Cascades, Sapience, Phylogeny", nil)
	}

	// Run simulation year by year (fast!)
	// Run simulation year by year (fast!) or with larger steps
	year := int64(0)
	iterationCount := int64(0) // Debug counter

	// Performance profiling
	var totalCarbonTime, totalEventTime, totalGeologyTime, totalOtherTime time.Duration
	var profileSamples int64

	for year < years {
		// Calculate adaptive step size at the START of the loop
		// Default to 1 year (required if life is enabled for reproduction/death cycles)
		stepSize := int64(1)

		if simulateGeology && !simulateLife {
			heat := ecosystem.GetPlanetaryHeat(year)

			// GEOLOGY-ONLY OPTIMIZATION: Use aggressive stepping throughout
			// Since we don't need year-by-year resolution for biology,
			// we can use much larger steps even in later eons
			if heat > 4.0 {
				// Hadean era (year 0-500M): Molten/Violent
				// AGGRESSIVE: Use 100k year steps
				stepSize = 100_000
			} else if heat > 1.5 {
				// Archean/Proterozoic (year 500M-3B): Cooling
				// AGGRESSIVE for geology-only: Use 100k year steps (was 10k)
				// Surface features still don't need fine resolution
				stepSize = 100_000
			} else {
				// Phanerozoic/Modern (year 3B+): Stable
				// Use 10k year steps for geology-only (was 100)
				// Only need fine resolution when biology is active
				stepSize = 10_000
			}

			// Ensure we don't overshoot the end
			if year+stepSize > years {
				stepSize = years - year
			}

			// Debug logging (first iteration only)
			if year == 0 {
				log.Printf("[ADAPTIVE STEPPING] Year 0: heat=%.2f, stepSize=%d, simulateLife=%v", heat, stepSize, simulateLife)
			}
		}

		// Progress reporting
		if year-lastProgress >= progressInterval && progressInterval > 0 {
			percent := (year * 100) / years
			if popSim != nil {
				totalPop, totalSpecies, totalExtinct := popSim.GetStats()
				client.SendGameMessage("system", fmt.Sprintf("⏳ Progress: %d%% (Year %d, Pop: %d, Species: %d, Extinct: %d)",
					percent, year, totalPop, totalSpecies, totalExtinct), nil)
			} else {
				client.SendGameMessage("system", fmt.Sprintf("⏳ Progress: %d%% (Year %d)", percent, year), nil)
			}
			lastProgress = year

			// Send map update to client at 25%, 50%, 75% milestones for visual feedback
			if percent == 25 || percent == 50 || percent == 75 {
				client.SendGameMessage("system", fmt.Sprintf("🗺️ Updating map at %d%%...", percent), nil)
				pct := int(percent)
				processor := p
				wID := char.WorldID
				cli := client // Capture client interface for goroutine
				go func() {
					updateCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
					defer cancel()
					if err := processor.sendMapUpdateToClient(updateCtx, cli, wID); err != nil {
						log.Printf("[WorldSimCmd] Map update at %d%% failed: %v", pct, err)
					}
				}()
			}
		}

		// Simulate population dynamics + evolution + speciation
		if simulateLife {
			popSim.SimulateYear()
		}

		// Update Climate Driver (Milankovitch Cycles)
		// Triggers ice ages/interglacials based on orbital mechanics
		// Only check every 100,000 years as designed (orbital cycles are very slow)
		if year%100_000 == 0 {
			climateDriver.Update(year)
		}

		// Apply evolution every 1000 years
		if simulateLife && popSim.CurrentYear%1000 == 0 {
			popSim.ApplyEvolution()

			// Apply co-evolution (predator-prey arms race) every 1000 years
			popSim.ApplyCoEvolution()

			// Apply genetic drift (stronger effect on small populations)
			popSim.ApplyGeneticDrift()

			// Apply sexual selection (display traits affect reproduction)
			popSim.ApplySexualSelection()
		}

		// Check for speciation every 10000 years
		if simulateLife && popSim.CurrentYear%10000 == 0 {
			// Update atmospheric oxygen levels
			oldO2 := popSim.OxygenLevel
			newO2 := popSim.UpdateOxygenLevel()
			popSim.ApplyOxygenEffects()

			// Report significant O2 changes (>2% shift)
			o2Change := (newO2 - oldO2) * 100
			if math.Abs(o2Change) > 0.5 {
				direction := "rising"
				if o2Change < 0 {
					direction = "falling"
				}
				client.SendGameMessage("system", fmt.Sprintf("🌬️ Atmospheric oxygen %s: %.1f%%", direction, newO2*100), nil)
			}

			newSpecies := popSim.CheckSpeciation()
			if newSpecies > 0 {
				client.SendGameMessage("system", fmt.Sprintf("🧬 %d new species evolved through speciation", newSpecies), nil)
				// TODO: Add speciation events to phylogenetic tree when CheckSpeciation returns parent/child info
			}

			// Allow species to migrate between biomes
			migrants := popSim.ApplyMigrationCycle()
			if migrants > 100 {
				client.SendGameMessage("system", fmt.Sprintf("🦋 %d individuals migrated to new biomes", migrants), nil)
			}

			// V2: Pathogen simulation - check for outbreaks every 10k years
			if simulateDiseases && simulateLife {
				speciesData := make(map[uuid.UUID]pathogen.SpeciesInfo)
				for _, biome := range popSim.Biomes {
					for _, sp := range biome.Species {
						if sp.Count > 0 {
							speciesData[sp.SpeciesID] = pathogen.SpeciesInfo{
								Population:        sp.Count,
								DiseaseResistance: float32(sp.Traits.DiseaseResistance),
								DietType:          string(sp.Diet),
								Density:           float64(sp.Count) / float64(biome.CarryingCapacity+1),
							}
							// Check for spontaneous outbreaks
							newPathogen, outbreak := diseaseSystem.CheckSpontaneousOutbreak(
								sp.SpeciesID, sp.Name, sp.Count,
								float64(sp.Count)/float64(biome.CarryingCapacity+1),
							)
							if outbreak != nil {
								totalOutbreaks++
								// CalculateR0 needs density and resistance params
								density := float32(sp.Count) / float32(biome.CarryingCapacity+1)
								r0 := newPathogen.CalculateR0(density, float32(sp.Traits.DiseaseResistance))
								client.SendGameMessage("system", fmt.Sprintf("🦠 OUTBREAK: %s (%s) in %s! R₀: %.1f",
									newPathogen.Name, newPathogen.Type, sp.Name, r0), nil)
								// Log to simulation logger
								if simLogger != nil {
									simLogger.LogPathogenOutbreakV2(ctx, popSim.CurrentYear, newPathogen.Name, string(newPathogen.Type), string(newPathogen.Transmission), sp.Name, r0, newPathogen.Virulence, outbreak.PeakInfected)
								}
							}
						}
					}
				}
				// Update all active outbreaks
				diseaseSystem.Update(popSim.CurrentYear, speciesData)
				// Report pandemic events
				for _, pandemic := range diseaseSystem.GetPandemics() {
					// Report if this is a large pandemic
					if pandemic.TotalDeaths > 1000 && pandemic.EndYear == popSim.CurrentYear {
						client.SendGameMessage("system", fmt.Sprintf("☠️ PANDEMIC: %s killed %d across multiple populations",
							pandemic.PathogenID, pandemic.TotalDeaths), nil)
					}
				}
			}

			// V2: Sapience detection - check species for proto-sapience and sapience
			if !sapienceAchieved && simulateLife {
				for _, biome := range popSim.Biomes {
					for _, sp := range biome.Species {
						if sp.Count > 1000 && sp.Traits.Intelligence > 0.5 { // Only check intelligent species
							// Map available traits, use fallbacks for missing ones
							traits := sapience.SpeciesTraits{
								Intelligence:  sp.Traits.Intelligence,
								Social:        sp.Traits.Social,
								ToolUse:       sp.Traits.Intelligence * 0.8, // Infer tool use from intelligence
								Communication: sp.Traits.Social * 0.7,       // Infer from social
								MagicAffinity: 0.0,                          // Default, no magic affinity trait
								Population:    sp.Count,
								Generation:    sp.Generation,
							}
							candidate := sapienceDetector.Evaluate(sp.SpeciesID, sp.Name, traits, popSim.CurrentYear)
							if candidate != nil {
								if candidate.Level == sapience.SapienceSapient {
									sapienceAchieved = true
									newSapientSpecies = append(newSapientSpecies, sp.SpeciesID) // Track for turning points
									client.SendGameMessage("system", fmt.Sprintf("🧠 SAPIENCE ACHIEVED! %s has become sapient! (Score: %.2f)",
										sp.Name, candidate.Score), nil)
								} else if candidate.Level == sapience.SapienceProtoSapient {
									client.SendGameMessage("system", fmt.Sprintf("🔮 Proto-sapience detected: %s shows early signs (Score: %.2f)",
										sp.Name, candidate.Score), nil)
								}
							}
						}
					}
				}
			}

			// V2: Extinction cascade - check for cascades when species go extinct
			// Build ecological relationships from population data (simplified)
			for _, biome := range popSim.Biomes {
				for _, sp := range biome.Species {
					if sp.Count == 0 {
						continue
					}
					// Infer relationships from diet
					switch sp.Diet {
					case population.DietCarnivore:
						// Carnivores depend on herbivores
						for _, prey := range biome.Species {
							if prey.Diet == population.DietHerbivore && prey.Count > 0 {
								cascadeSim.AddRelationship(population.EcologicalRelationship{
									SourceSpeciesID: sp.SpeciesID,
									TargetSpeciesID: prey.SpeciesID,
									Type:            population.RelationshipPredation,
									Strength:        0.5,
									IsObligate:      false,
								})
							}
						}
					case population.DietHerbivore:
						// Herbivores depend on flora
						for _, flora := range biome.Species {
							if flora.Diet == population.DietPhotosynthetic && flora.Count > 0 {
								cascadeSim.AddRelationship(population.EcologicalRelationship{
									SourceSpeciesID: sp.SpeciesID,
									TargetSpeciesID: flora.SpeciesID,
									Type:            population.RelationshipPredation,
									Strength:        0.3,
									IsObligate:      false,
								})
							}
						}
					}
				}
			}

			// Check for new extinctions and calculate cascades
			if simulateLife {
				for _, biome := range popSim.Biomes {
					for _, sp := range biome.Species {
						if sp.Count == 0 && sp.Generation > 0 { // Newly extinct
							recentExtinctions++ // Track for turning points
							result := cascadeSim.CalculateCascade(sp.SpeciesID, sp.Name, popSim.CurrentYear, 3)
							if result != nil && result.TotalAffected > 0 {
								totalCascades++
								client.SendGameMessage("system", fmt.Sprintf("💀 EXTINCTION CASCADE: %s extinction affects %d other species",
									sp.Name, result.TotalAffected), nil)

								// Apply cascade effects to populations
								for affectedID, impact := range result.PopulationChanges {
									for _, b := range popSim.Biomes {
										if affected, ok := b.Species[affectedID]; ok {
											deaths := int64(float32(affected.Count) * impact)
											affected.Count -= deaths
											if affected.Count < 0 {
												affected.Count = 0
											}
										}
									}
								}

								// Update phylogenetic tree
								phyloTree.MarkExtinct(sp.SpeciesID, popSim.CurrentYear)
							}
						}
					}
				}
			}
		}

		// Check for theological/geological events (standardized tick rate)
		if simulateGeology {
			// Standardize to 365 ticks per year
			currentTick := year * 365

			// stepSize is calculated at the beginning of the loop

			iterationStart := time.Now()

			// === CARBON-SILICATE CYCLE ===
			// Update atmospheric composition every 100,000 years
			// The carbon-silicate cycle operates on million-year timescales
			// More frequent updates don't improve accuracy but waste CPU
			carbonStart := time.Now()
			if year%100_000 == 0 || year == 0 {
				// Simulate atmospheric composition changes
				// This creates a self-regulating climate thermostat:
				// - Volcanism adds CO2 (proportional to planetary heat)
				// - Weathering removes CO2 (proportional to temp × precipitation × CO2)
				// - Negative feedback: Warming → More weathering → Less CO2 → Cooling

				// Calculate volcanic CO2 emissions (source)
				heat := ecosystem.GetPlanetaryHeat(year)
				volcanicRate := atmosphere.CalculateVolcanicOutgassing(heat)

				// Calculate weathering CO2 removal (sink)
				geoStats := geology.GetStats()
				weatheringRate := atmosphere.CalculateWeatheringRate(
					geoStats.AverageTemperature,
					1000.0, // TODO: Get actual global average precipitation from weather system
					geoStats.LandPercent/100.0,
					atm.CO2Mass,
				)

				// Update atmospheric CO2 (mass balance)
				// Apply the rates for 100k years of accumulated change
				atmosphereStepSize := int64(100_000)
				if year == 0 {
					atmosphereStepSize = stepSize // First iteration uses actual stepSize
				}
				atm.SimulateCarbonCycle(atmosphereStepSize, volcanicRate, weatheringRate)

				// Update climate driver with greenhouse effect from atmosphere
				atmosphereStats := atm.GetStats()
				climateDriver.SetGreenhouseOffset(atmosphereStats.GreenhouseOffset)
			}
			carbonTime := time.Since(carbonStart)
			totalCarbonTime += carbonTime

			// === GEOLOGICAL EVENTS ===
			eventStart := time.Now()

			// Trigger random events
			// We pass currentTick and stepSize (dt)
			geoManager.CheckForNewEvents(currentTick, stepSize)
			geoManager.UpdateActiveEvents(currentTick) // Clean up expired events

			eventTime := time.Since(eventStart)
			totalEventTime += eventTime

			// === GEOLOGY SIMULATION ===
			geologyStart := time.Now()

			// Update Geology state (Tectonics, Erosion, etc)
			// Apply combined temperature modifiers:
			// - Geological events (volcanic winter, etc)
			// - Geothermal offset (internal heat)
			// - Greenhouse offset (atmospheric CO2)
			eventTempMod, _, _ := geoManager.GetEnvironmentModifiers()
			totalTempMod := eventTempMod + climateDriver.GetGeothermalOffset() + climateDriver.GetGreenhouseOffset()
			phaseEvent := geology.SimulateGeology(stepSize, totalTempMod)

			// MANUALLY TRIGGER BIOME GENERATION
			// Refactored to occur here instead of inside SimulateGeology to prevent memory leaks in geology-only runs.
			// Only update biomes if life is being simulated (to feed populations), or very rarely.
			// 10M year interval matches the previous internal logic but is now conditional.
			if simulateLife && year%10_000_000 == 0 {
				geology.UpdateBiomes(totalTempMod)
			}

			// Log phase transition events (e.g., Great Deluge)
			if phaseEvent != nil {
				client.SendGameMessage("system", fmt.Sprintf("🌊 %s: %s (Year %d)",
					phaseEvent.Type, phaseEvent.Description, phaseEvent.Year), nil)
			}

			geologyTime := time.Since(geologyStart)
			totalGeologyTime += geologyTime

			// === PROFILING ===
			otherTime := time.Since(iterationStart) - carbonTime - eventTime - geologyTime
			totalOtherTime += otherTime
			profileSamples++

			// Log performance breakdown every 100 iterations
			if profileSamples > 0 && profileSamples%100 == 0 {
				avgCarbon := totalCarbonTime / time.Duration(profileSamples)
				avgEvent := totalEventTime / time.Duration(profileSamples)
				avgGeology := totalGeologyTime / time.Duration(profileSamples)
				avgOther := totalOtherTime / time.Duration(profileSamples)
				avgTotal := avgCarbon + avgEvent + avgGeology + avgOther

				log.Printf("[PERF] Avg/Iter: %v | Geo: %v (%.0f%%) | Carbon: %v (%.0f%%) | Event: %v (%.0f%%) | Other: %v (%.0f%%)",
					avgTotal,
					avgGeology, float64(avgGeology)/float64(avgTotal)*100,
					avgCarbon, float64(avgCarbon)/float64(avgTotal)*100,
					avgEvent, float64(avgEvent)/float64(avgTotal)*100,
					avgOther, float64(avgOther)/float64(avgTotal)*100)
			}

			// Process ALL active events for biome transitions and effects
			// This ensures warming events (climate recovery) are properly handled
			for _, e := range geoManager.ActiveEvents {
				eventType := population.ExtinctionEventType(e.Type)

				// Check if this event started recently (within this check period)
				eventAge := currentTick - e.StartTick
				isNewEvent := eventAge < stepSize*365 // Started this year

				if isNewEvent {
					// Rate limit notifications for frequent events to prevent spam
					lastLog, known := sentEventLogs[string(e.Type)]
					shouldLog := true
					if known && currentTick-lastLog < 500000*365 { // Don't spam same event type within 500k years
						shouldLog = false
					}

					if shouldLog {
						geologicalEvents++
						eventCounts[e.Type]++
						// Log the event
						client.SendGameMessage("system", fmt.Sprintf("⚠️ GEOLOGICAL EVENT: %s (severity: %.0f%%)", e.Type, e.Severity*100), nil)
						sentEventLogs[string(e.Type)] = currentTick
					}

					geology.ApplyEvent(e)

					// Apply extinction event to populations based on event type
					if simulateLife && isNewEvent && shouldLog { // Only apply sudden extinction shock if it's new
						deaths := popSim.ApplyExtinctionEvent(eventType, e.Severity)
						if deaths > 100 {
							client.SendGameMessage("system", fmt.Sprintf("   💀 %d organisms perished", deaths), nil)
						}
					}
				}

				// Apply biome transitions for ALL active events (cooling AND warming)
				// This is what allows climate recovery to work!
				if simulateLife && popSim != nil {
					transitioned := popSim.ApplyBiomeTransitions(eventType, e.Severity)
					if transitioned > 0 && isNewEvent { // Only log transitions for new events to reduce spam
						// Rate limit these details too
						lastLog, _ := sentEventLogs[string(e.Type)]
						if currentTick-lastLog < 1000*365 { // Slight grace period matching the main log
							if e.Type == ecosystem.EventWarming || e.Type == ecosystem.EventGreenhouseSpike {
								client.SendGameMessage("system", fmt.Sprintf("   🌡️ %d biomes warming! Climate recovery in progress", transitioned), nil)
							} else {
								client.SendGameMessage("system", fmt.Sprintf("   🌍 %d biomes shifted due to climate change", transitioned), nil)
							}
						}
					}
				}

				// Update continental configuration for drift events
				if eventType == population.EventContinentalDrift && isNewEvent && popSim != nil {
					oldFrag := popSim.ContinentalFragmentation
					newFrag := popSim.UpdateContinentalConfiguration(true, e.Severity)
					popSim.ApplyContinentalEffects()

					// Report significant configuration changes
					fragChange := math.Abs(newFrag - oldFrag)
					if fragChange > 0.05 {
						var status string
						if newFrag > 0.7 {
							status = "fragmented (high endemism)"
						} else if newFrag < 0.3 {
							status = "unified (supercontinent forming)"
						} else {
							status = "moderate"
						}
						client.SendGameMessage("system", fmt.Sprintf("   🗺️ Continental configuration: %s (%.0f%%)", status, newFrag*100), nil)
					}
				}
			}

			// Update geographic systems (hex grid, regions, tectonics)
			if simulateLife && popSim != nil {
				popSim.UpdateGeographicSystems(10000)

				// Apply isolation effects (gigantism/dwarfism) to isolated regions
				isolationAffected := popSim.ApplyIsolationEffects()
				if isolationAffected > 0 && year%100000 == 0 {
					client.SendGameMessage("system", fmt.Sprintf("🏝️ Island effects: %d species affected by isolation", isolationAffected), nil)
				}
			}

			// Regional migration every 100,000 years
			if simulateLife && popSim != nil && year%100000 == 0 && year > 0 {
				migrations := popSim.ApplyRegionalMigration()
				if migrations > 0 {
					client.SendGameMessage("system", fmt.Sprintf("🌍 Regional migration: %d species expanded to new regions", migrations), nil)
				}
			}

			// Check for turning points every 100,000 years
			if simulateLife && popSim != nil && year%100000 == 0 && year > 0 {
				totalPop, totalSpecies, _ := popSim.GetStats()

				// Determine significant event string based on recent activity
				significantEvent := ""
				if len(geoManager.ActiveEvents) > 0 {
					for _, e := range geoManager.ActiveEvents {
						if e.Severity > 0.5 {
							significantEvent = string(e.Type)
							break
						}
					}
				}

				// Check for turning point
				tp := turningPointMgr.CheckForTurningPoint(
					popSim.CurrentYear,
					int(totalSpecies),
					recentExtinctions,
					newSapientSpecies,
					significantEvent,
				)

				if tp != nil {
					client.SendGameMessage("system", fmt.Sprintf("🔮 TURNING POINT: %s - %s", tp.Title, tp.Description), nil)
					if simLogger != nil {
						simLogger.LogTurningPoint(ctx, popSim.CurrentYear, string(tp.Trigger), "auto_resolved")
					}
					// For sync simulation, auto-resolve with first option (observe only)
					if len(tp.Interventions) > 0 {
						turningPointMgr.ResolveTurningPoint(tp.ID, tp.Interventions[0].ID)
					}
				}

				// Reset periodic counters
				recentExtinctions = 0
				newSapientSpecies = []uuid.UUID{}
				_ = totalPop // Silence unused variable warning
			}
		}

		year += stepSize

		// Debug: Log iteration count every 100 iterations
		if year == 0 {
			iterationCount = 0
		}
		iterationCount++
		if iterationCount%100 == 0 {
			log.Printf("[PERF] Iteration #%d: year=%d, increment=%d, heat=%.2f",
				iterationCount, year, stepSize, ecosystem.GetPlanetaryHeat(year))
		}
	}

	// Update biomes one last time to ensure final map state is correct
	// Calculate final temp mod
	eventTempMod, _, _ := geoManager.GetEnvironmentModifiers()
	finalTempMod := eventTempMod + climateDriver.GetGeothermalOffset() + climateDriver.GetGreenhouseOffset()
	geology.UpdateBiomes(finalTempMod)

	// Get final statistics
	geoStats = geology.GetStats()
	var totalPop, totalSpecies, totalExtinct int64
	if popSim != nil {
		totalPop, totalSpecies, totalExtinct = popSim.GetStats()
	}

	// Build summary
	var sb strings.Builder
	sb.WriteString("=== Simulation Complete ===\n")
	sb.WriteString(fmt.Sprintf("Years Simulated: %d\n", years))
	if popSim != nil {
		sb.WriteString(fmt.Sprintf("Total Population: %d\n", totalPop))
		sb.WriteString(fmt.Sprintf("Living Species: %d\n", totalSpecies))
		sb.WriteString(fmt.Sprintf("Extinct Species: %d\n", totalExtinct))
	}
	sb.WriteString(fmt.Sprintf("Geological Events: %d\n", geologicalEvents))

	// Event Breakdown
	sb.WriteString("--- Event Frequency ---\n")
	for eventType, count := range eventCounts {
		sb.WriteString(fmt.Sprintf("%s: %d\n", string(eventType), count))
	}

	// V2 Statistics
	if simulateLife {
		sb.WriteString("--- V2 Features ---\n")
		sb.WriteString(fmt.Sprintf("Disease Outbreaks: %d\n", totalOutbreaks))
		sb.WriteString(fmt.Sprintf("Extinction Cascades: %d\n", totalCascades))
		if sapienceAchieved {
			sb.WriteString("Sapience: ACHIEVED! 🧠\n")
		} else {
			progress := sapienceDetector.CalculateSapienceProgress()
			sb.WriteString(fmt.Sprintf("Sapience Progress: %.0f%%\n", progress*100))
		}
		sb.WriteString(fmt.Sprintf("Species in Tree of Life: %d\n", len(phyloTree.Nodes)))
	}

	sb.WriteString("--- Terrain Stats ---\n")
	sb.WriteString(fmt.Sprintf("Tectonic Plates: %d\n", geoStats.PlateCount))
	sb.WriteString(fmt.Sprintf("Avg Temperature: %.1f°C\n", geoStats.AverageTemperature))
	sb.WriteString(fmt.Sprintf("Avg Elevation: %.0fm\n", geoStats.AverageElevation))
	sb.WriteString(fmt.Sprintf("Max Elevation: %.0fm\n", geoStats.MaxElevation))
	sb.WriteString(fmt.Sprintf("Sea Level: %.0fm\n", geoStats.SeaLevel))
	sb.WriteString(fmt.Sprintf("Land Coverage: %.1f%%\n", geoStats.LandPercent))

	// Earth-Like Habitability Score (era-aware)
	calStats := calibration.CollectStats(geology)
	eraBenchmarks, eraName := calibration.BenchmarksForEra(years)
	habitability := calibration.CalculateHabitabilityScore(calStats, eraBenchmarks)
	sb.WriteString("--- Habitability Score ---\n")
	sb.WriteString(fmt.Sprintf("Era: %s Earth (%.1f Ga)\n", eraName, float64(years)/1e9))
	sb.WriteString(fmt.Sprintf("Earth-Like Score: %.0f/100 %s\n", habitability.Score, habitability.Emoji()))
	sb.WriteString(fmt.Sprintf("Ocean: %s | Land: %s | Climate: %s\n",
		habitability.OceanGrade(), habitability.LandGrade(), habitability.ClimateGrade()))
	if habitability.BimodalOK {
		sb.WriteString("Bimodal Hypsometry: ✓\n")
	} else {
		sb.WriteString("Bimodal Hypsometry: ✗ (crustal differentiation incomplete)\n")
	}

	// Land Formation Analysis
	sb.WriteString("--- Land Formations ---\n")
	sb.WriteString(fmt.Sprintf("Continents: %d\n", calStats.ContinentCount))

	// Describe continent formation stage
	switch {
	case calStats.ContinentCount == 0:
		sb.WriteString("  🌊 Water World - No stable continental crust\n")
	case calStats.ContinentCount == 1:
		sb.WriteString("  🏝️ Supercontinent - All land unified\n")
	case calStats.ContinentCount <= 3:
		sb.WriteString("  🌍 Proto-continents - Early continental assembly\n")
	case calStats.ContinentCount <= 7:
		sb.WriteString("  🌎 Modern-style - Multiple distinct landmasses\n")
	default:
		sb.WriteString("  🗾 Fragmented - Many small continents/islands\n")
	}

	// Geological provinces (mineral potential)
	sb.WriteString(fmt.Sprintf("Geological Provinces: %d\n", calStats.ProvinceCount))
	if calStats.CratonCount > 0 || calStats.OrogenCount > 0 || calStats.BasinCount > 0 {
		sb.WriteString(fmt.Sprintf("  ⛏️ Cratons: %d (ancient stable cores - diamond potential)\n", calStats.CratonCount))
		sb.WriteString(fmt.Sprintf("  ⛰️ Orogens: %d (fold belts - gold/copper potential)\n", calStats.OrogenCount))
		sb.WriteString(fmt.Sprintf("  🪨 Basins: %d (sedimentary - iron/coal potential)\n", calStats.BasinCount))
	}

	// Hydrology features
	if calStats.RiverCount > 0 || calStats.LakeCount > 0 {
		sb.WriteString(fmt.Sprintf("Rivers: %d | Lakes: %d\n", calStats.RiverCount, calStats.LakeCount))
	}

	// Cave systems
	if calStats.CaveCount > 0 {
		sb.WriteString(fmt.Sprintf("Cave Systems: %d\n", calStats.CaveCount))
	}

	// Natural Satellites section
	sb.WriteString("--- Natural Satellites ---\n")
	if len(satellites) == 0 {
		sb.WriteString("Moons: None\n")
		sb.WriteString("Climate Stability: Chaotic (no stabilizing moon)\n")
	} else {
		sb.WriteString(fmt.Sprintf("Moons: %d\n", len(satellites)))
		for _, sat := range satellites {
			// Mass in Luna units (Earth's Moon = 1.0 Luna)
			massInLunas := sat.Mass / astronomy.MoonMassKg
			// Distance in thousands of km
			distanceKm := sat.Distance / 1000.0
			sb.WriteString(fmt.Sprintf("  🌙 %s: %.2fx Luna, %.0f km\n", sat.Name, massInLunas, distanceKm))
		}

		// Calculate effects
		tidalStress := astronomy.CalculateTidalStress(satellites)
		obliquityStability := astronomy.CalculateObliquityStability(satellites, astronomy.EarthMassKg)

		sb.WriteString(fmt.Sprintf("Tidal Stress: %.2fx Earth\n", tidalStress))
		if obliquityStability > 0.5 {
			sb.WriteString("Climate Stability: Stable (large moon stabilizes axis)\n")
		} else {
			sb.WriteString("Climate Stability: Variable (small moons)\n")
		}
		sb.WriteString(fmt.Sprintf("Impact Shielding: %.0f%%\n", impactShielding*100))

		// Calculate asteroid impacts prevented
		// Formula: actual_impacts × (shielding / (1 - shielding))
		// This represents impacts that WOULD have occurred without moons
		actualImpacts := eventCounts[ecosystem.EventAsteroidImpact]
		if actualImpacts > 0 && impactShielding > 0 {
			// Inverse calculation: if shielding = 10%, and we had 100 impacts,
			// then without moons we'd have had 100 / (1 - 0.10) = 111 impacts
			// So prevented = 111 - 100 = 11
			unshieldedImpacts := float64(actualImpacts) / (1.0 - impactShielding)
			impactsPrevented := int(unshieldedImpacts) - actualImpacts
			if impactsPrevented > 0 {
				sb.WriteString(fmt.Sprintf("Asteroids Deflected: %d (would have hit without moons)\n", impactsPrevented))
			}
		}
	}

	// Species breakdown grouped by biome type
	sb.WriteString("--- Species by Biome Type ---\n")

	// Aggregate by biome type
	type biomeTypeStats struct {
		count      int
		population int64
		species    map[string]struct {
			count      int64
			generation int64
		}
	}
	biomeTypeMap := make(map[string]*biomeTypeStats)

	if popSim != nil {
		for _, biome := range popSim.Biomes {
			biomeTypeName := string(biome.BiomeType)
			if _, exists := biomeTypeMap[biomeTypeName]; !exists {
				biomeTypeMap[biomeTypeName] = &biomeTypeStats{
					species: make(map[string]struct {
						count      int64
						generation int64
					}),
				}
			}
			stats := biomeTypeMap[biomeTypeName]
			stats.count++
			stats.population += biome.TotalPopulation()

			for _, sp := range biome.Species {
				// Use base species name (without biome prefix for cleaner display)
				existing := stats.species[sp.Name]
				existing.count += sp.Count
				if sp.Generation > existing.generation {
					existing.generation = sp.Generation
				}
				stats.species[sp.Name] = existing
			}
		}
	} // end if popSim != nil

	// Output grouped stats
	for biomeType, stats := range biomeTypeMap {
		sb.WriteString(fmt.Sprintf("%s (%d biomes, Pop: %d):\n", biomeType, stats.count, stats.population))
		speciesShown := 0
		for name, sp := range stats.species {
			if speciesShown >= 5 {
				sb.WriteString(fmt.Sprintf("  ...and %d more species\n", len(stats.species)-5))
				break
			}
			sb.WriteString(fmt.Sprintf("  %s: %d (Gen %d)\n", name, sp.count, sp.generation))
			speciesShown++
		}
	}

	// Fossil record
	if popSim != nil && len(popSim.FossilRecord.Extinct) > 0 {
		sb.WriteString("--- Fossil Record ---\n")
		shown := 0
		for _, ext := range popSim.FossilRecord.Extinct {
			if shown >= 5 {
				sb.WriteString(fmt.Sprintf("...and %d more extinct species\n", len(popSim.FossilRecord.Extinct)-5))
				break
			}
			duration := ext.ExistedUntil - ext.ExistedFrom
			sb.WriteString(fmt.Sprintf("† %s (existed %d years, cause: %s)\n", ext.Name, duration, ext.ExtinctionCause))
			shown++
		}
	}

	client.SendGameMessage("system", sb.String(), nil)

	// Send map update directly to client so they can see the simulated world
	client.SendGameMessage("system", "🗺️ Rendering world map...", nil)
	if err := p.sendMapUpdateToClient(ctx, client, char.WorldID); err != nil {
		log.Printf("[WorldSimCmd] Failed to send map update to client: %v", err)
		client.SendGameMessage("warning", "Map update failed - use 'world map' to load manually", nil)
	} else {
		client.SendGameMessage("system", "✅ World map updated!", nil)
	}

	return nil
}

// handleWorldInfo shows current world state
func (p *GameProcessor) handleWorldInfo(ctx context.Context, client websocket.GameClient) error {
	char, err := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if err != nil {
		client.SendGameMessage("error", "Could not get character info", nil)
		return nil
	}

	world, err := p.worldRepo.GetWorld(ctx, char.WorldID)
	if err != nil {
		client.SendGameMessage("error", "Could not get world info", nil)
		return nil
	}

	var sb strings.Builder
	sb.WriteString("=== World Info ===\n")
	sb.WriteString(fmt.Sprintf("Name: %s\n", world.Name))
	sb.WriteString(fmt.Sprintf("ID: %s\n", world.ID))
	if world.Circumference != nil {
		circumKm := *world.Circumference / 1000
		sb.WriteString(fmt.Sprintf("Circumference: %.0f km\n", circumKm))
	}
	sb.WriteString(fmt.Sprintf("Entities: %d\n", len(p.ecosystemService.Entities)))

	// Show terrain stats if geology has been simulated
	if geology, exists := p.worldGeology[char.WorldID]; exists && geology.IsInitialized() {
		geoStats := geology.GetStats()
		sb.WriteString("--- Terrain ---\n")
		sb.WriteString(fmt.Sprintf("Tectonic Plates: %d\n", geoStats.PlateCount))
		sb.WriteString(fmt.Sprintf("Avg Elevation: %.0fm\n", geoStats.AverageElevation))
		sb.WriteString(fmt.Sprintf("Max Elevation: %.0fm\n", geoStats.MaxElevation))
		sb.WriteString(fmt.Sprintf("Min Elevation: %.0fm\n", geoStats.MinElevation))
		sb.WriteString(fmt.Sprintf("Sea Level: %.0fm\n", geoStats.SeaLevel))
		sb.WriteString(fmt.Sprintf("Land Coverage: %.1f%%\n", geoStats.LandPercent))
		sb.WriteString(fmt.Sprintf("Years Simulated: %d\n", geoStats.YearsSimulated))
	} else {
		sb.WriteString("--- Terrain ---\n")
		sb.WriteString("Not yet simulated. Use 'world simulate <years>' to generate terrain.\n")
	}

	// Show async runner status if one exists
	if runner := p.getRunner(char.WorldID); runner != nil {
		stats := runner.GetStats()
		speed := runner.GetSpeed()
		sb.WriteString("--- Async Simulation ---\n")
		var stateIcon string
		switch stats.State {
		case ecosystem.RunnerRunning:
			stateIcon = "▶️"
		case ecosystem.RunnerPaused:
			stateIcon = "⏸️"
		case ecosystem.RunnerIdle:
			stateIcon = "⏹️"
		default:
			stateIcon = "❓"
		}
		sb.WriteString(fmt.Sprintf("State: %s %s\n", stateIcon, stats.State))
		sb.WriteString(fmt.Sprintf("Current Year: %d\n", stats.CurrentYear))
		sb.WriteString(fmt.Sprintf("Years Simulated: %d\n", stats.YearsSimulated))
		sb.WriteString(fmt.Sprintf("Speed: %d years/tick\n", speed))
		sb.WriteString(fmt.Sprintf("Avg Rate: %.1f years/sec\n", stats.YearsPerSecond))
		sb.WriteString(fmt.Sprintf("Ticks: %d | Snapshots: %d\n", stats.TickCount, stats.SnapshotCount))
	}

	client.SendGameMessage("system", sb.String(), nil)
	return nil
}

// handleWorldReset resets the world simulation to default state
func (p *GameProcessor) handleWorldReset(ctx context.Context, client websocket.GameClient) error {
	char, err := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if err != nil {
		client.SendGameMessage("error", "Could not get character info", nil)
		return nil
	}

	worldID := char.WorldID

	// Stop and remove async runner if it exists
	if runner := p.getRunner(worldID); runner != nil {
		runner.Stop()
		delete(p.worldRunners, worldID)
		client.SendGameMessage("system", "⏹️ Async simulation stopped.", nil)
	}

	// Clear geology for this world
	delete(p.worldGeology, worldID)

	// Clear map service geology cache
	if p.mapService != nil {
		p.mapService.SetWorldGeology(worldID, nil)
	}

	// Clear all entities for this world
	for id, entity := range p.ecosystemService.Entities {
		if entity.WorldID == worldID {
			delete(p.ecosystemService.Entities, id)
			delete(p.ecosystemService.Behaviors, id)
		}
	}

	client.SendGameMessage("system", "🔄 World reset complete. Geology, entities, and simulation state cleared.\nUse 'world simulate <years>' or 'world run' to start fresh.", nil)

	// Send world_reset message so frontend knows to switch to molten planet view
	// (geology is now nil, so we can't render a map)
	client.SendGameMessage("world_reset", "Geology cleared - returning to molten planet view", nil)

	return nil
}

// handleWorldRun starts or resumes the async simulation runner
func (p *GameProcessor) handleWorldRun(ctx context.Context, client websocket.GameClient) error {
	char, _ := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if char == nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	// Get or create runner for this world
	runner := p.getOrCreateRunner(char.WorldID)
	if runner == nil {
		client.SendGameMessage("error", "Failed to create simulation runner", nil)
		return nil
	}

	switch runner.GetState() {
	case ecosystem.RunnerRunning:
		client.SendGameMessage("system", "⏯️ Simulation already running. Use 'world pause' to stop.", nil)
	case ecosystem.RunnerPaused:
		runner.Resume()
		client.SendGameMessage("system", "▶️ Simulation resumed.", nil)
	default:
		if err := runner.Start(0); err != nil {
			client.SendGameMessage("error", fmt.Sprintf("Failed to start runner: %v", err), nil)
			return nil
		}
		client.SendGameMessage("system", "▶️ Simulation started.", nil)
	}
	return nil
}

// handleWorldPause pauses the async simulation runner
func (p *GameProcessor) handleWorldPause(ctx context.Context, client websocket.GameClient) error {
	char, _ := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if char == nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	runner := p.getRunner(char.WorldID)
	if runner == nil {
		client.SendGameMessage("system", "⏸️ No simulation running.", nil)
		return nil
	}

	runner.Pause()
	client.SendGameMessage("system", "⏸️ Simulation paused. Use 'world run' to resume.", nil)
	return nil
}

// handleWorldSpeed changes the simulation speed
func (p *GameProcessor) handleWorldSpeed(ctx context.Context, client websocket.GameClient, speedStr string) error {
	char, _ := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if char == nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	// Parse speed from string or alias
	var speed ecosystem.SimulationSpeed
	speedLower := strings.ToLower(speedStr)
	switch speedLower {
	case "normal", "1":
		speed = ecosystem.SpeedSlow // 1 year/sec
	case "quick", "10":
		speed = ecosystem.SpeedNormal // 10 years/sec
	case "fast", "100":
		speed = ecosystem.SpeedFast // 100 years/sec
	case "turbo", "1000":
		speed = ecosystem.SpeedTurbo // 1000 years/sec
	default:
		client.SendGameMessage("error", "Invalid speed. Use: normal, quick, fast, turbo (or 1, 10, 100, 1000)", nil)
		return nil
	}

	runner := p.getRunner(char.WorldID)
	if runner == nil {
		client.SendGameMessage("system", fmt.Sprintf("🏃 Speed set to %s. Start simulation with 'world run'.", speedLower), nil)
		return nil
	}

	runner.SetSpeed(speed)
	client.SendGameMessage("system", fmt.Sprintf("🏃 Simulation speed set to %s (%d years/sec).", speedLower, speed), nil)
	return nil
}

// handleWorldMap sends full world map data to the client for the world map modal
func (p *GameProcessor) handleWorldMap(ctx context.Context, client websocket.GameClient) error {
	// LEGACY HANDLER: Used for JSON data (low-res).
	// New clients should use `graphic_mode="image"` (handled by handleWorldMapImage) for high-res.

	char, err := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if err != nil || char == nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	if p.mapService == nil {
		client.SendGameMessage("error", "Map service not available", nil)
		return nil
	}

	// Get aggregated world map data (64x64 grid by default)
	mapData, err := p.mapService.GetWorldMapData(ctx, char, 64)
	if err != nil {
		client.SendGameMessage("error", fmt.Sprintf("Failed to generate world map: %v", err), nil)
		return nil
	}

	// Debug logging
	sampleBiome := "none"
	if len(mapData.Tiles) > 0 {
		sampleBiome = mapData.Tiles[len(mapData.Tiles)/2].Biome
	}
	log.Printf("[WORLDMAP] Sending world_map_data: tiles=%d, grid=%dx%d, worldSize=%.0fx%.0f, sampleBiome=%s",
		len(mapData.Tiles), mapData.GridWidth, mapData.GridHeight, mapData.WorldWidth, mapData.WorldHeight, sampleBiome)

	// Convert to map[string]interface{} for JSON serialization
	payload := map[string]interface{}{
		"tiles":        mapData.Tiles,
		"grid_width":   mapData.GridWidth,
		"grid_height":  mapData.GridHeight,
		"world_width":  mapData.WorldWidth,
		"world_height": mapData.WorldHeight,
		"player_x":     mapData.PlayerX,
		"player_y":     mapData.PlayerY,
		"world_id":     mapData.WorldID.String(),
		"world_name":   mapData.WorldName,
		"is_simulated": mapData.IsSimulated,

		// Planetary stats
		"simulated_years": mapData.SimulatedYears,
		"avg_temperature": mapData.AvgTemperature,
		"max_elevation":   mapData.MaxElevation,
		"sea_level":       mapData.SeaLevel,
		"land_coverage":   mapData.LandCoverage,
		"seed":            mapData.Seed,

		// Add basic history log (events not available in sync mode)
		"history": []string{"Simulation data retrieved successfully."},
	}

	// Add satellites if available (Natural Satellites Phase 4)
	// First try lookService cache (orchestrator flow)
	if p.lookService != nil {
		if worldData, ok := p.lookService.GetCachedWorldData(char.WorldID); ok && worldData != nil {
			payload["satellites"] = worldData.Satellites
		}
	}

	// If not found in lookService cache, check mapService geology (world simulate flow)
	if payload["satellites"] == nil && p.mapService != nil {
		// Geology cache in map service is populated after simulation
		if geo := p.mapService.GetWorldGeology(char.WorldID); geo != nil {
			payload["satellites"] = geo.Satellites
		}
	}

	// Add overlays if geology data is available
	if geo := p.mapService.GetWorldGeology(char.WorldID); geo != nil {
		overlays := make(map[string]interface{})

		// Tectonic plate overlay - array of plate IDs for each cell
		if tectonicMap, plateMeta := geo.GetTectonicMap(mapData.GridWidth, mapData.GridHeight); tectonicMap != nil {
			overlays["tectonics"] = tectonicMap
			overlays["plate_info"] = plateMeta
			log.Printf("[WORLDMAP] Added tectonics overlay: %d cells, %d plates", len(tectonicMap), len(plateMeta))
		} else {
			log.Printf("[WORLDMAP] No tectonic map available (plates=%d, topology=%v)", len(geo.Plates), geo.Topology != nil)
		}

		// Calculate environmental data for overlays
		// We calculate these on demand for the requested grid size
		width, height := mapData.GridWidth, mapData.GridHeight

		tempMap := geo.GetTemperatureMap(width, height)
		overlays["temp"] = tempMap

		moistureMap := geo.GetMoistureMap(width, height)
		overlays["moisture"] = moistureMap

		elevMap := geo.GetElevationMap(width, height)
		overlays["elevation"] = elevMap

		// Sediment map for satellite-style rendering (Phase 6b)
		sedimentMap := geo.GetSedimentMap(width, height)
		overlays["sediment"] = sedimentMap

		biomeMap := geo.GetBiomeMap(width, height, tempMap, moistureMap)
		overlays["biome"] = biomeMap

		resourceMap := geo.GetResourceMap(width, height, elevMap)
		overlays["resources"] = resourceMap

		featuresMap := geo.GetTerrainFeaturesMap(width, height)
		overlays["features"] = featuresMap

		log.Printf("[WORLDMAP] Added env overlays: Temp/Moist/Elev/Sediment/Biome/Res (grid %dx%d)", width, height)

		// Mineral deposit overlay - list of discovered and undiscovered deposits
		minerals := geo.GetMineralDeposits()
		if len(minerals) > 0 {
			overlays["minerals"] = minerals
			log.Printf("[WORLDMAP] Added minerals overlay: %d deposits", len(minerals))
		}

		// River network overlay - Phase C
		if len(geo.Rivers) > 0 {
			overlays["rivers"] = geo.Rivers
			log.Printf("[WORLDMAP] Added rivers overlay: %d rivers", len(geo.Rivers))
		}

		if len(overlays) > 0 {
			payload["overlays"] = overlays
		}
	} else {
		log.Printf("[WORLDMAP] No geology data available for world %s", char.WorldID)
	}

	client.SendGameMessage("world_map_data", "", payload)
	return nil
}

// getOrCreateRunner gets an existing runner or creates a new one for the world
// now initialized with V2 population simulator and persistence
func (p *GameProcessor) getOrCreateRunner(worldID uuid.UUID) *ecosystem.SimulationRunner {
	if p.worldRunners == nil {
		p.worldRunners = make(map[uuid.UUID]*ecosystem.SimulationRunner)
	}
	if runner, ok := p.worldRunners[worldID]; ok {
		return runner
	}

	// Create config
	config := ecosystem.DefaultConfig(worldID)
	// Pass repositories
	runner := ecosystem.NewSimulationRunner(config, p.simSnapshotRepo, p.runnerStateRepo)

	// Initialize Simulator (Load from DB or Create New)
	// Use world ID as seed part 1
	seed := int64(worldID[0])<<56 | int64(worldID[1])<<48 |
		int64(worldID[2])<<40 | int64(worldID[3])<<32 |
		int64(worldID[4])<<24 | int64(worldID[5])<<16 |
		int64(worldID[6])<<8 | int64(worldID[7])

	// Initialize (this handles loading snapshot if available)
	runner.InitializePopulationSimulator(seed)

	// Configure satellite physics (Natural Satellites Phase 4)
	// Look up cached world data to get satellites
	if p.lookService != nil {
		if worldData, ok := p.lookService.GetCachedWorldData(worldID); ok && worldData != nil {
			runner.ConfigureSatellitePhysics(worldData.Satellites)
		}
	}

	// Set handlers
	runner.SetEventBroadcastHandler(func(event ecosystem.RunnerEvent) {
		// New: Trigger Map Update on visualization event
		if event.Type == "visualization_update" {
			go func() {
				// Use specific context for map generation independent of current request
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cancel()
				if err := p.broadcastMapUpdate(ctx, worldID); err != nil {
					log.Printf("[MAP] Failed to broadcast update: %v", err)
				}
			}()
		}

		// Broadcast to all clients in this world
		if p.Hub != nil {
			clients := p.Hub.GetClientsByWorldID(worldID)
			for _, c := range clients {
				// Using "sim_event" message type for notifications
				c.SendGameMessage("sim_event", event.Description, map[string]interface{}{
					"year":       event.Year,
					"type":       event.Type,
					"importance": event.Importance,
				})
			}
		}
	})

	p.worldRunners[worldID] = runner
	return runner
}

// getRunner retrieves an existing runner for the world (nil if not exists)
func (p *GameProcessor) getRunner(worldID uuid.UUID) *ecosystem.SimulationRunner {
	if p.worldRunners == nil {
		return nil
	}
	return p.worldRunners[worldID]
}

// getSeasonFromYear calculates season from simulated year for weather simulation
func (p *GameProcessor) getSeasonFromYear(simulatedYear int64) weather.Season {
	// Cycle through seasons: 4 seasons per year
	seasonIndex := simulatedYear % 4
	switch seasonIndex {
	case 0:
		return weather.SeasonSpring
	case 1:
		return weather.SeasonSummer
	case 2:
		return weather.SeasonFall
	default:
		return weather.SeasonWinter
	}
}

// legacy runEcosystemTick removed - replaced by V2 logic in runner.tick()

// handleWorldMapImage handles requests for high-resolution map images (Option 5/Hybrid)
// handleWorldMapImage handles requests for high-resolution map images (Option 5/Hybrid)
func (p *GameProcessor) handleWorldMapImage(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	char, err := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if err != nil || char == nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	if p.mapService == nil {
		client.SendGameMessage("error", "Map service not available", nil)
		return nil
	}

	geo := p.mapService.GetWorldGeology(char.WorldID)
	if geo == nil {
		client.SendGameMessage("error", "World geology not initialized. Run simulation first.", nil)
		return nil
	}

	// Parse optional parameters from client
	// { "width": 4096, "height": 2048 }
	var params struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	// Attempt to parse payload if present
	if len(cmd.Payload) > 0 {
		if err := json.Unmarshal(cmd.Payload, &params); err != nil {
			log.Printf("[MAP] Failed to parse params: %v", err)
			// Continue with defaults
		}
	}

	// Determine resolution
	width := 2048
	height := 1024
	if params.Width > 0 && params.Height > 0 {
		width = params.Width
		height = params.Height
		// Limit to safe maximums
		if width > 8192 {
			width = 8192
		}
		if height > 4096 {
			height = 4096
		}
	}

	// Grid constant
	const GridSize = 256

	// Create a new context with timeout for this specific operation
	// The renderer has its own timeout logic, but we enforce it here too
	renderCtx, cancel := context.WithTimeout(ctx, 125*time.Second) // Slightly longer than renderer timeout
	defer cancel()

	// 1. Render Image (Visuals)
	imageBytes, err := p.mapService.RenderMap(renderCtx, char.WorldID, geo, width, height)
	if err != nil {
		log.Printf("[MAP] Render failed: %v", err)
		client.SendGameMessage("error", fmt.Sprintf("Map render failed: %v", err), nil)
		return nil
	}

	// 2. Generate Binary Grid Data (Logic layer for tooltips)
	// Grid is 256x256 regardless of image resolution
	gridData := p.mapService.BuildBinaryGrid(geo, GridSize, GridSize/2)
	var gridBytes []byte
	if gridData != nil {
		gridBytes = gridData.Serialize()
	}

	// 3. Render Heightmap PNG (Displacement for 3D Globe)
	// Resolution matches visual map for 1:1 displacement
	heightmapBytes, err := p.mapService.RenderHeightmapPNG(renderCtx, char.WorldID, geo, width, height)
	if err != nil {
		log.Printf("[MAP] Heightmap render failed: %v", err)
		// Non-critical, continue without it
	}

	// 4. Render Material PNG (Rock hardness, continental, sediment for data-driven coloring)
	materialBytes, err := p.mapService.RenderMaterialPNG(renderCtx, char.WorldID, geo, width, height)
	if err != nil {
		log.Printf("[MAP] Material render failed: %v", err)
		// Non-critical, continue without it
	}

	// 5. Render Ice PNG (Ice sheet coverage for polar/glacier visualization)
	iceBytes, err := p.mapService.RenderIcePNG(renderCtx, char.WorldID, geo, width, height)
	if err != nil {
		log.Printf("[MAP] Ice render failed: %v", err)
		// Non-critical, continue without it
	}

	// 6. Construct Binary Message
	// Protocol: [Type:1][JSONLen:4][JSON][BinLen:4][Image][GridLen:4][Grid][HeightMapLen:4][HeightMap][MaterialLen:4][Material][IceLen:4][Ice]

	// Construct JSON Metadata
	type MapImageMetadata struct {
		Width            int     `json:"width"`
		Height           int     `json:"height"`
		GridWidth        int     `json:"grid_width"`
		GridHeight       int     `json:"grid_height"`
		WorldWidth       float64 `json:"world_width"`
		WorldHeight      float64 `json:"world_height"`
		PlayerX          float64 `json:"player_x"`
		PlayerY          float64 `json:"player_y"`
		SimulatedYears   int64   `json:"simulated_years"`
		AvgTemperature   float64 `json:"avg_temperature"`
		MaxElevation     float64 `json:"max_elevation"`
		MinElevation     float64 `json:"min_elevation"`
		SeaLevel         float64 `json:"sea_level"`
		LandCoverage     float64 `json:"land_coverage"`
		Seed             int64   `json:"seed"`
		IsSimulated      bool    `json:"is_simulated"`
		HasGridData      bool    `json:"has_grid_data"`
		HasHeightmapData bool    `json:"has_heightmap_data"`
		HasMaterialData  bool    `json:"has_material_data"`
		HasIceData       bool    `json:"has_ice_data"`
	}

	stats := geo.GetStats()

	meta := MapImageMetadata{
		Width:            width,
		Height:           height,
		GridWidth:        GridSize,
		GridHeight:       GridSize / 2,
		WorldWidth:       geo.Circumference,
		WorldHeight:      geo.Circumference / 2,
		PlayerX:          char.PositionX,
		PlayerY:          char.PositionY,
		SimulatedYears:   stats.YearsSimulated,
		AvgTemperature:   stats.AverageTemperature,
		MaxElevation:     stats.MaxElevation,
		MinElevation:     stats.MinElevation,
		SeaLevel:         stats.SeaLevel,
		LandCoverage:     stats.LandPercent * 100, // Convert to percentage
		Seed:             geo.Seed,
		IsSimulated:      true,
		HasGridData:      len(gridBytes) > 0,
		HasHeightmapData: len(heightmapBytes) > 0,
		HasMaterialData:  len(materialBytes) > 0,
		HasIceData:       len(iceBytes) > 0,
	}

	jsonBytes, err := json.Marshal(meta)
	if err != nil {
		log.Printf("[MAP] JSON marshal failed: %v", err)
		return nil
	}

	// Binary section: [ImageLen:4][Image][GridLen:4][Grid][HeightMapLen:4][HeightMap][MaterialLen:4][Material][IceLen:4][Ice]
	binSectionSize := 4 + len(imageBytes) + 4 + len(gridBytes) + 4 + len(heightmapBytes) + 4 + len(materialBytes) + 4 + len(iceBytes)
	binSection := bytes.NewBuffer(make([]byte, 0, binSectionSize))

	// Write Image Length and Data (Big Endian)
	binary.Write(binSection, binary.BigEndian, uint32(len(imageBytes)))
	binSection.Write(imageBytes)

	// Write Grid Length and Data (Big Endian)
	binary.Write(binSection, binary.BigEndian, uint32(len(gridBytes)))
	if len(gridBytes) > 0 {
		binSection.Write(gridBytes)
	}

	// Write Heightmap Length and Data (Big Endian)
	binary.Write(binSection, binary.BigEndian, uint32(len(heightmapBytes)))
	if len(heightmapBytes) > 0 {
		binSection.Write(heightmapBytes)
	}

	// Write Material Length and Data (Big Endian)
	binary.Write(binSection, binary.BigEndian, uint32(len(materialBytes)))
	if len(materialBytes) > 0 {
		binSection.Write(materialBytes)
	}

	// Write Ice Length and Data (Big Endian)
	binary.Write(binSection, binary.BigEndian, uint32(len(iceBytes)))
	if len(iceBytes) > 0 {
		binSection.Write(iceBytes)
	}

	binBytes := binSection.Bytes()

	// Calculate total size: 1 + 4 + len(json) + 4 + len(binSection)
	totalSize := 1 + 4 + len(jsonBytes) + 4 + len(binBytes)
	buf := bytes.NewBuffer(make([]byte, 0, totalSize))

	// Write Header (Type 0x01)
	buf.WriteByte(0x01)

	// Write JSON Length (Big Endian)
	binary.Write(buf, binary.BigEndian, uint32(len(jsonBytes)))
	// Write JSON Data
	buf.Write(jsonBytes)

	// Write Binary Section Length (Big Endian)
	binary.Write(buf, binary.BigEndian, uint32(len(binBytes)))
	// Write Binary Section Data
	buf.Write(binBytes)

	log.Printf("[MAP] Sending world map image: %dx%d, grid %dx%d, image=%d bytes, grid=%d bytes, heightmap=%d bytes, material=%d bytes, ice=%d bytes",
		width, height, GridSize, GridSize/2, len(imageBytes), len(gridBytes), len(heightmapBytes), len(materialBytes), len(iceBytes))

	// Send Raw Binary Message
	client.SendRawBytes(buf.Bytes())

	return nil
}

// handleWorldTile handles individual tile requests for the cube-face tile system.
// Message format: "face,level,x,y,size" (e.g., "0,2,1,3,256")
func (p *GameProcessor) handleWorldTile(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	char, err := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if err != nil || char == nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	if p.mapService == nil {
		client.SendGameMessage("error", "Map service not available", nil)
		return nil
	}

	geo := p.mapService.GetWorldGeology(char.WorldID)
	if geo == nil {
		client.SendGameMessage("error", "World geology not initialized. Run simulation first.", nil)
		return nil
	}

	// Parse tile request from message
	if cmd.Message == nil {
		client.SendGameMessage("error", "Usage: world_tile face,level,x,y,size", nil)
		return nil
	}

	// Parse comma-separated values
	parts := strings.Split(*cmd.Message, ",")
	if len(parts) != 5 {
		client.SendGameMessage("error", "Invalid tile request format. Expected: face,level,x,y,size", nil)
		return nil
	}

	face, err := strconv.Atoi(parts[0])
	if err != nil || face < 0 || face > 5 {
		client.SendGameMessage("error", "Invalid face (must be 0-5)", nil)
		return nil
	}

	level, err := strconv.Atoi(parts[1])
	if err != nil || level < 0 {
		client.SendGameMessage("error", "Invalid level (must be >= 0)", nil)
		return nil
	}

	tileX, err := strconv.Atoi(parts[2])
	if err != nil || tileX < 0 {
		client.SendGameMessage("error", "Invalid x coordinate", nil)
		return nil
	}

	tileY, err := strconv.Atoi(parts[3])
	if err != nil || tileY < 0 {
		client.SendGameMessage("error", "Invalid y coordinate", nil)
		return nil
	}

	size, err := strconv.Atoi(parts[4])
	if err != nil || size < 64 || size > 1024 {
		client.SendGameMessage("error", "Invalid size (must be 64-1024)", nil)
		return nil
	}

	// Render the tile
	tileRenderer := p.mapService.TileRenderer()
	if tileRenderer == nil {
		client.SendGameMessage("error", "Tile renderer not available", nil)
		return nil
	}

	req := gamemap.TileRequest{
		Face:  gamemap.CubeFace(face),
		Level: level,
		X:     tileX,
		Y:     tileY,
		Size:  size,
	}

	tileData, err := tileRenderer.RenderTile(ctx, req, geo)
	if err != nil {
		client.SendGameMessage("error", fmt.Sprintf("Failed to render tile: %v", err), nil)
		return err
	}

	// Build binary response
	// Format: [type:1][json_len:4][json][bin_len:4][image][heightmap_len:4][heightmap]
	metadata := map[string]interface{}{
		"type":          "world_tile",
		"face":          face,
		"level":         level,
		"x":             tileX,
		"y":             tileY,
		"width":         tileData.Width,
		"height":        tileData.Height,
		"imageSize":     len(tileData.Image),
		"heightmapSize": len(tileData.Heightmap),
	}

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal tile metadata: %w", err)
	}

	// Build binary payload
	binSection := bytes.NewBuffer(nil)
	binSection.Write(tileData.Image)
	binSection.Write(tileData.Heightmap)

	// Calculate total size
	totalSize := 1 + 4 + len(jsonBytes) + 4 + len(binSection.Bytes())
	buf := bytes.NewBuffer(make([]byte, 0, totalSize))

	// Write Header (Type 0x02 for tiles)
	buf.WriteByte(0x02)

	// Write JSON Length (Big Endian)
	binary.Write(buf, binary.BigEndian, uint32(len(jsonBytes)))
	buf.Write(jsonBytes)

	// Write Binary Section Length
	binary.Write(buf, binary.BigEndian, uint32(len(binSection.Bytes())))
	buf.Write(binSection.Bytes())

	log.Printf("[TILE] Sending tile f=%d l=%d (%d,%d): %dx%d, image=%d bytes, heightmap=%d bytes",
		face, level, tileX, tileY, tileData.Width, tileData.Height, len(tileData.Image), len(tileData.Heightmap))

	client.SendRawBytes(buf.Bytes())
	return nil
}

// broadcastMapUpdate generates and sends a map update to all clients in the world
func (p *GameProcessor) broadcastMapUpdate(ctx context.Context, worldID uuid.UUID) error {
	if p.mapService == nil || p.Hub == nil {
		return nil
	}

	// Check if there are any clients in this world to avoid wasting CPU
	clients := p.Hub.GetClientsByWorldID(worldID)
	if len(clients) == 0 {
		return nil
	}

	geo := p.mapService.GetWorldGeology(worldID)
	if geo == nil {
		return fmt.Errorf("geology not found")
	}

	// Standard update resolution
	width := 2048
	height := 1024
	const GridSize = 256

	// 1. Render Image
	imageBytes, err := p.mapService.RenderMap(ctx, worldID, geo, width, height)
	if err != nil {
		return err
	}

	// 2. Generate Binary Grid Data
	gridData := p.mapService.BuildBinaryGrid(geo, GridSize, GridSize/2)
	var gridBytes []byte
	if gridData != nil {
		gridBytes = gridData.Serialize()
	}

	// 3. Render Heightmap PNG
	heightmapBytes, err := p.mapService.RenderHeightmapPNG(ctx, worldID, geo, width, height)
	if err != nil {
		log.Printf("[MAP] Heightmap render failed in broadcast: %v", err)
	}

	// 4. Render Material PNG
	materialBytes, err := p.mapService.RenderMaterialPNG(ctx, worldID, geo, width, height)
	if err != nil {
		log.Printf("[MAP] Material render failed in broadcast: %v", err)
	}

	// 5. Render Ice PNG
	iceBytes, err := p.mapService.RenderIcePNG(ctx, worldID, geo, width, height)
	if err != nil {
		log.Printf("[MAP] Ice render failed in broadcast: %v", err)
	}

	// 6. Construct Binary Message (Protocol 0x01)
	stats := geo.GetStats()
	type MapImageMetadata struct {
		Width            int     `json:"width"`
		Height           int     `json:"height"`
		GridWidth        int     `json:"grid_width"`
		GridHeight       int     `json:"grid_height"`
		WorldWidth       float64 `json:"world_width"`
		WorldHeight      float64 `json:"world_height"`
		PlayerX          float64 `json:"player_x"`
		PlayerY          float64 `json:"player_y"`
		SimulatedYears   int64   `json:"simulated_years"`
		AvgTemperature   float64 `json:"avg_temperature"`
		MaxElevation     float64 `json:"max_elevation"`
		MinElevation     float64 `json:"min_elevation"`
		SeaLevel         float64 `json:"sea_level"`
		LandCoverage     float64 `json:"land_coverage"`
		Seed             int64   `json:"seed"`
		IsSimulated      bool    `json:"is_simulated"`
		HasGridData      bool    `json:"has_grid_data"`
		HasHeightmapData bool    `json:"has_heightmap_data"`
		HasMaterialData  bool    `json:"has_material_data"`
		HasIceData       bool    `json:"has_ice_data"`
	}

	meta := MapImageMetadata{
		Width:            width,
		Height:           height,
		GridWidth:        GridSize,
		GridHeight:       GridSize / 2,
		WorldWidth:       geo.Circumference,
		WorldHeight:      geo.Circumference / 2,
		PlayerX:          0, // Origin for broadcast
		PlayerY:          0,
		SimulatedYears:   stats.YearsSimulated,
		AvgTemperature:   stats.AverageTemperature,
		MaxElevation:     stats.MaxElevation,
		MinElevation:     stats.MinElevation,
		SeaLevel:         stats.SeaLevel,
		LandCoverage:     stats.LandPercent * 100,
		Seed:             geo.Seed,
		IsSimulated:      true,
		HasGridData:      len(gridBytes) > 0,
		HasHeightmapData: len(heightmapBytes) > 0,
		HasMaterialData:  len(materialBytes) > 0,
		HasIceData:       len(iceBytes) > 0,
	}

	jsonBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	// Binary section
	binSectionSize := 4 + len(imageBytes) + 4 + len(gridBytes) + 4 + len(heightmapBytes) + 4 + len(materialBytes) + 4 + len(iceBytes)
	binSection := bytes.NewBuffer(make([]byte, 0, binSectionSize))

	binary.Write(binSection, binary.BigEndian, uint32(len(imageBytes)))
	binSection.Write(imageBytes)

	binary.Write(binSection, binary.BigEndian, uint32(len(gridBytes)))
	if len(gridBytes) > 0 {
		binSection.Write(gridBytes)
	}

	binary.Write(binSection, binary.BigEndian, uint32(len(heightmapBytes)))
	if len(heightmapBytes) > 0 {
		binSection.Write(heightmapBytes)
	}

	binary.Write(binSection, binary.BigEndian, uint32(len(materialBytes)))
	if len(materialBytes) > 0 {
		binSection.Write(materialBytes)
	}

	binary.Write(binSection, binary.BigEndian, uint32(len(iceBytes)))
	if len(iceBytes) > 0 {
		binSection.Write(iceBytes)
	}

	binBytes := binSection.Bytes()
	totalSize := 1 + 4 + len(jsonBytes) + 4 + len(binBytes)
	buf := bytes.NewBuffer(make([]byte, 0, totalSize))

	buf.WriteByte(0x01)
	binary.Write(buf, binary.BigEndian, uint32(len(jsonBytes)))
	buf.Write(jsonBytes)
	binary.Write(buf, binary.BigEndian, uint32(len(binBytes)))
	buf.Write(binBytes)

	payload := buf.Bytes()

	log.Printf("[MAP] Broadcasting world map update to %d clients", len(clients))

	for _, client := range clients {
		client.SendRawBytes(payload)
		client.SendGameMessage("system", "🌍 Planetary simulation update received.", nil)
	}

	return nil
}

// sendMapUpdateToClient generates and sends a map update directly to a single client
// This is used when the world ID may be nil/empty and broadcast won't work
func (p *GameProcessor) sendMapUpdateToClient(ctx context.Context, client websocket.GameClient, worldID uuid.UUID) error {
	if p.mapService == nil {
		return fmt.Errorf("map service not available")
	}

	geo := p.mapService.GetWorldGeology(worldID)
	if geo == nil {
		return fmt.Errorf("geology not found for world %s", worldID)
	}

	// Standard update resolution
	width := 2048
	height := 1024
	const GridSize = 256

	// 1. Render Image
	imageBytes, err := p.mapService.RenderMap(ctx, worldID, geo, width, height)
	if err != nil {
		return err
	}

	// 2. Generate Binary Grid Data
	gridData := p.mapService.BuildBinaryGrid(geo, GridSize, GridSize/2)
	var gridBytes []byte
	if gridData != nil {
		gridBytes = gridData.Serialize()
	}

	// 3. Render Heightmap PNG
	heightmapBytes, err := p.mapService.RenderHeightmapPNG(ctx, worldID, geo, width, height)
	if err != nil {
		log.Printf("[MAP] Heightmap render failed: %v", err)
		heightmapBytes = nil
	}

	// 4. Render Material PNG (rock/soil types)
	materialBytes, err := p.mapService.RenderMaterialPNG(ctx, worldID, geo, width, height)
	if err != nil {
		log.Printf("[MAP] Material render failed: %v", err)
		materialBytes = nil
	}

	// 5. Render Ice PNG (glaciers/snow)
	iceBytes, err := p.mapService.RenderIcePNG(ctx, worldID, geo, width, height)
	if err != nil {
		log.Printf("[MAP] Ice render failed: %v", err)
		iceBytes = nil
	}

	// Construct JSON Metadata
	type MapImageMetadata struct {
		Width            int     `json:"width"`
		Height           int     `json:"height"`
		GridWidth        int     `json:"grid_width"`
		GridHeight       int     `json:"grid_height"`
		SeaLevel         float64 `json:"sea_level"`
		MaxElevation     float64 `json:"max_elevation"`
		MinElevation     float64 `json:"min_elevation"`
		HasGridData      bool    `json:"has_grid_data"`
		HasHeightmapData bool    `json:"has_heightmap_data"`
		HasMaterialData  bool    `json:"has_material_data"`
		HasIceData       bool    `json:"has_ice_data"`
	}

	stats := geo.GetStats()
	metadata := MapImageMetadata{
		Width:            width,
		Height:           height,
		GridWidth:        GridSize,
		GridHeight:       GridSize / 2,
		SeaLevel:         stats.SeaLevel,
		MaxElevation:     stats.MaxElevation,
		MinElevation:     stats.MinElevation,
		HasGridData:      len(gridBytes) > 0,
		HasHeightmapData: len(heightmapBytes) > 0,
		HasMaterialData:  len(materialBytes) > 0,
		HasIceData:       len(iceBytes) > 0,
	}

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	// Build binary section
	binSection := bytes.NewBuffer(nil)

	binary.Write(binSection, binary.BigEndian, uint32(len(imageBytes)))
	if len(imageBytes) > 0 {
		binSection.Write(imageBytes)
	}

	binary.Write(binSection, binary.BigEndian, uint32(len(gridBytes)))
	if len(gridBytes) > 0 {
		binSection.Write(gridBytes)
	}

	binary.Write(binSection, binary.BigEndian, uint32(len(heightmapBytes)))
	if len(heightmapBytes) > 0 {
		binSection.Write(heightmapBytes)
	}

	binary.Write(binSection, binary.BigEndian, uint32(len(materialBytes)))
	if len(materialBytes) > 0 {
		binSection.Write(materialBytes)
	}

	binary.Write(binSection, binary.BigEndian, uint32(len(iceBytes)))
	if len(iceBytes) > 0 {
		binSection.Write(iceBytes)
	}

	binBytes := binSection.Bytes()
	totalSize := 1 + 4 + len(jsonBytes) + 4 + len(binBytes)
	buf := bytes.NewBuffer(make([]byte, 0, totalSize))

	buf.WriteByte(0x01)
	binary.Write(buf, binary.BigEndian, uint32(len(jsonBytes)))
	buf.Write(jsonBytes)
	binary.Write(buf, binary.BigEndian, uint32(len(binBytes)))
	buf.Write(binBytes)

	payload := buf.Bytes()

	log.Printf("[MAP] Sending map update directly to client (world %s)", worldID)
	client.SendRawBytes(payload)

	return nil
}

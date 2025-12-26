package ecosystem

import (
	"math"
	"math/rand"
	"sync"
	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/astronomy"
	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/underground"
	"tw-backend/internal/worldgen/weather"

	"github.com/google/uuid"
)

// WorldGeology manages terrain evolution over geological time
type WorldGeology struct {
	mu sync.RWMutex

	WorldID       uuid.UUID
	Seed          int64
	Circumference float64 // meters

	// Core geographic data
	Heightmap       *geography.Heightmap       // Flat heightmap for legacy consumers
	SphereHeightmap *geography.SphereHeightmap // Spherical heightmap for proper 3D operations
	Plates          []geography.TectonicPlate
	SeaLevel        float64          // meters (0 = baseline, positive = higher sea level)
	Topology        spatial.Topology // Spherical topology for plate operations

	// Underground data (Phase 3)
	Columns     *underground.ColumnGrid // Per-column underground data
	Caves       []*underground.Cave     // Cave networks
	Composition string                  // "volcanic", "continental", "oceanic", "ancient"

	// Dynamic geographic features
	Hotspots   []geography.Point // Fixed mantle plume locations
	Rivers     [][]geography.Point
	Biomes     []geography.Biome
	Satellites []astronomy.Satellite // Natural satellites

	// Simulation state
	TotalYearsSimulated int64
	rng                 *rand.Rand

	// Scale factors (pixels to real-world)
	PixelsPerKm float64 // How many heightmap pixels per real km

	// Time Accumulators for variable step simulation
	TectonicStressAccumulator float64 // Years of accumulated tectonic stress
	ErosionAccumulator        float64 // Years of accumulated erosion potential
	DepositAccumulator        float64 // Years of accumulated organic deposit time
	RiverAccumulator          float64 // Years of accumulated river/biome update time
	MaintenanceAccumulator    float64 // Years of accumulated maintenance time (subsidence, clamping, stats)
	GeneralAccumulator        float64 // Years of accumulated time for lower frequency events
}

// GeologyStats contains summary statistics for world info display
type GeologyStats struct {
	AverageElevation   float64
	AverageTemperature float64
	MaxElevation       float64
	MinElevation       float64
	SeaLevel           float64
	LandPercent        float64
	PlateCount         int
	HotspotCount       int
	RiverCount         int
	BiomeCount         int
	YearsSimulated     int64
}

// NewWorldGeology creates a new geology manager for a world
// composition: "volcanic", "continental", "oceanic", or "ancient"
func NewWorldGeology(worldID uuid.UUID, seed int64, circumferenceMeters float64) *WorldGeology {
	return &WorldGeology{
		WorldID:       worldID,
		Seed:          seed,
		Circumference: circumferenceMeters,
		SeaLevel:      0,             // Baseline sea level
		Composition:   "continental", // Default composition
		rng:           rand.New(rand.NewSource(seed)),
	}
}

// SetComposition sets the world's geological composition.
// Valid values: "volcanic", "continental", "oceanic", "ancient"
func (g *WorldGeology) SetComposition(composition string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Composition = composition
}

// InitializeGeology creates the baseline terrain from scratch
// This should be called when a world is first simulated
func (g *WorldGeology) InitializeGeology() {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Calculate map dimensions based on circumference
	// Circumference in meters -> convert to km for our scale
	circumKm := g.Circumference / 1000.0

	// Target: ~10 km per pixel for reasonable detail
	// For Earth-like (40,000 km), this gives 4000x2000 (too large for memory)
	// Let's cap at 512x256 and adjust scale
	maxWidth := 512
	maxHeight := 256

	// Calculate pixels per km based on circumference
	// width = circumference, height = circumference/2 (latitude)
	width := int(circumKm / 10)  // 10km per pixel
	height := int(circumKm / 20) // latitude is half

	if width > maxWidth {
		width = maxWidth
	}
	if height > maxHeight {
		height = maxHeight
	}
	if width < 64 {
		width = 64
	}
	if height < 32 {
		height = 32
	}

	g.PixelsPerKm = float64(width) / circumKm

	// Create spherical topology for all plate operations
	g.Topology = spatial.NewCubeSphereTopology(height)

	// Generate tectonic plates using spherical topology
	plateCount := 6 + g.rng.Intn(4) // 6-9 plates for variety
	g.Plates = geography.GeneratePlates(plateCount, g.Topology, g.Seed)

	// Generate initial heightmap using spherical topology
	// Create sphere heightmap and convert to flat for legacy consumers
	g.SphereHeightmap = geography.NewSphereHeightmap(g.Topology)
	g.SphereHeightmap = geography.GenerateHeightmap(g.Plates, g.SphereHeightmap, g.Topology, g.Seed, 1.0, 1.0)
	g.Heightmap = g.SphereHeightmap.ToFlatHeightmap(width, height)

	// Initialize hotspots (2-5 fixed mantle plume locations)
	numHotspots := 2 + g.rng.Intn(4)
	g.Hotspots = make([]geography.Point, numHotspots)
	for i := 0; i < numHotspots; i++ {
		g.Hotspots[i] = geography.Point{
			X: float64(g.rng.Intn(width)),
			Y: float64(g.rng.Intn(height)),
		}
	}

	// Calculate initial sea level (target ~30% land coverage)
	g.SeaLevel = geography.AssignOceanLand(g.Heightmap, 0.3)

	// Generate initial rivers using spherical algorithm
	if g.SphereHeightmap != nil {
		sphereRivers := geography.GenerateRiversSpherical(g.SphereHeightmap, g.SeaLevel, g.Seed)
		g.Rivers = geography.ConvertSphericalRiversToFlat(sphereRivers, g.Topology.Resolution())
		// Sync sphere heightmap changes from river erosion
		g.syncSphereToFlat()
	} else {
		g.Rivers = geography.GenerateRivers(g.Heightmap, g.SeaLevel, g.Seed)
	}

	// Initialize biomes using Weather→Biome pipeline (no latitude coupling)
	g.Biomes = g.generateBiomesFromClimate(0.0) // No global temp modifier initially

	// Initialize underground column grid (Phase 3)
	g.initializeColumns(width, height)
}

// syncSphereToFlat updates the flat Heightmap from the SphereHeightmap
// Call this after making changes to SphereHeightmap to keep both in sync
func (g *WorldGeology) syncSphereToFlat() {
	if g.SphereHeightmap == nil || g.Heightmap == nil {
		return
	}
	g.Heightmap = g.SphereHeightmap.ToFlatHeightmap(g.Heightmap.Width, g.Heightmap.Height)
}

// initializeColumns creates the underground column grid and generates strata
func (g *WorldGeology) initializeColumns(width, height int) {
	g.Columns = underground.NewColumnGrid(width, height)
	g.Caves = []*underground.Cave{}

	// Initialize each column with surface from heightmap and strata based on composition
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			col := g.Columns.Get(x, y)
			surface := g.Heightmap.Get(x, y)
			col.Surface = surface

			// Generate strata based on world composition
			g.generateStrata(col, surface)

			// Add magma layer at hotspots
			for _, hotspot := range g.Hotspots {
				dist := math.Sqrt(math.Pow(float64(x)-hotspot.X, 2) + math.Pow(float64(y)-hotspot.Y, 2))
				if dist < 5 { // Within hotspot radius
					col.Magma = &underground.MagmaInfo{
						TopZ:        surface - 1000,
						BottomZ:     surface - 5000,
						Temperature: 1500,
						Pressure:    100,
						Viscosity:   0.5,
					}
				}
			}
		}
	}
}

// generateStrata creates geological layers for a column based on composition
func (g *WorldGeology) generateStrata(col *underground.WorldColumn, surface float64) {
	bedrock := col.Bedrock

	switch g.Composition {
	case "volcanic":
		// Volcanic worlds: basalt dominant, frequent lava tubes
		col.AddStratum("soil", surface, surface-5, 2, 0, 0.4)
		col.AddStratum("basalt", surface-5, surface-200, 6, 1000, 0.1)
		col.AddStratum("gabbro", surface-200, surface-2000, 7, 100000, 0.05)
		col.AddStratum("mantle", surface-2000, bedrock, 9, 1000000, 0.01)

	case "oceanic":
		// Oceanic worlds: limestone rich, extensive caves
		if surface < g.SeaLevel {
			// Underwater: thick limestone from marine deposits
			col.AddStratum("sediment", surface, surface-20, 2, 100, 0.5)
			col.AddStratum("limestone", surface-20, surface-500, 4, 10000, 0.3)
			col.AddStratum("chalk", surface-500, surface-1000, 3, 50000, 0.2)
			col.AddStratum("granite", surface-1000, bedrock, 8, 500000, 0.05)
		} else {
			// Coastal land
			col.AddStratum("soil", surface, surface-10, 2, 0, 0.4)
			col.AddStratum("limestone", surface-10, surface-300, 4, 10000, 0.3)
			col.AddStratum("granite", surface-300, bedrock, 8, 500000, 0.05)
		}

	case "ancient":
		// Ancient worlds: deep erosion, mineral-rich, extensive caves
		col.AddStratum("soil", surface, surface-15, 2, 0, 0.4)
		col.AddStratum("sandstone", surface-15, surface-100, 4, 100000, 0.25)
		col.AddStratum("limestone", surface-100, surface-400, 5, 500000, 0.3)
		col.AddStratum("schist", surface-400, surface-1500, 6, 1000000, 0.1)
		col.AddStratum("granite", surface-1500, bedrock, 9, 2000000, 0.02)

	default: // "continental"
		// Continental: balanced mix
		col.AddStratum("soil", surface, surface-10, 2, 0, 0.4)
		col.AddStratum("sedimentary", surface-10, surface-100, 4, 10000, 0.2)
		col.AddStratum("limestone", surface-100, surface-300, 5, 100000, 0.25)
		col.AddStratum("granite", surface-300, surface-2000, 8, 500000, 0.05)
		col.AddStratum("basalt", surface-2000, bedrock, 7, 1000000, 0.03)
	}
}

// simulateCaveFormation generates caves through limestone dissolution
// Called during SimulateGeology every 100,000+ years
func (g *WorldGeology) simulateCaveFormation(yearsElapsed int64) {
	if g.Columns == nil {
		return
	}

	// Build rainfall array from biomes (moisture affects dissolution)
	rainfall := make([]float64, len(g.Biomes))
	for i, biome := range g.Biomes {
		// Estimate rainfall from biome type
		switch biome.Type {
		case "rainforest", "swamp":
			rainfall[i] = 1.0
		case "grassland", "savanna":
			rainfall[i] = 0.6
		case "forest", "taiga":
			rainfall[i] = 0.7
		case "tundra":
			rainfall[i] = 0.3
		case "desert", "volcanic":
			rainfall[i] = 0.1
		case "ocean", "beach":
			rainfall[i] = 0.8
		default:
			rainfall[i] = 0.5
		}
	}

	// Configure cave formation
	config := underground.DefaultCaveConfig()
	// Adjust based on composition
	switch g.Composition {
	case "oceanic":
		config.DissolutionRate *= 2.0 // More limestone = faster caves
	case "ancient":
		config.DissolutionRate *= 3.0 // Very old = extensive caves
	case "volcanic":
		config.DissolutionRate *= 0.5 // Less limestone
	}

	// Run cave formation simulation
	newCaves := underground.SimulateCaveFormation(
		g.Columns,
		rainfall,
		yearsElapsed,
		g.Seed+g.TotalYearsSimulated,
		config,
	)

	// Register new caves
	g.Caves = append(g.Caves, newCaves...)
}

// simulateMagmaChambers processes magma chamber evolution and tectonic volcanism
func (g *WorldGeology) simulateMagmaChambers(yearsElapsed int64) {
	if g.Columns == nil || len(g.Plates) == 0 {
		return
	}

	// Extract tectonic boundaries from plate data
	// Use 3D Position and Velocity projected to 2D for legacy underground API
	plateCentroids := make([]underground.Vector3, len(g.Plates))
	plateMovements := make([]underground.Vector3, len(g.Plates))
	for i, plate := range g.Plates {
		// Convert spherical coordinate to flat x,y
		plateCentroids[i] = underground.Vector3{
			X: float64(plate.Centroid.Face*g.Heightmap.Width/6 + plate.Centroid.X),
			Y: float64(plate.Centroid.Y),
			Z: 0,
		}
		plateMovements[i] = underground.Vector3{
			X: plate.Velocity.X,
			Y: plate.Velocity.Y,
			Z: plate.Velocity.Z,
		}
	}

	boundaries := underground.GetTectonicBoundaries(
		g.Heightmap.Width,
		g.Heightmap.Height,
		plateCentroids,
		plateMovements,
	)

	// Get existing magma chambers from columns
	chambers := g.collectMagmaChambers()

	config := underground.DefaultMagmaConfig()
	// Adjust for composition
	if g.Composition == "volcanic" {
		config.EruptionThreshold = 60 // More frequent eruptions
		config.LavaTubeFormationProb = 0.9
	}

	// Run magma simulation
	erupted, newTubes, _ := underground.SimulateMagmaChambers(
		g.Columns,
		chambers,
		boundaries,
		yearsElapsed,
		g.Seed+g.TotalYearsSimulated,
		config,
	)

	// Handle eruptions - apply surface effects
	for _, chamber := range erupted {
		x, y := int(chamber.Center.X), int(chamber.Center.Y)
		if x >= 0 && x < g.Heightmap.Width && y >= 0 && y < g.Heightmap.Height {
			// Apply volcano to surface
			height := 500 + g.rng.Float64()*1500 // 500-2000m
			radius := 2.0 + g.rng.Float64()*3.0
			geography.ApplyVolcanoFlat(g.Heightmap, float64(x), float64(y), radius, height)
		}
	}

	// Register new lava tubes as caves
	g.Caves = append(g.Caves, newTubes...)
}

// collectMagmaChambers gathers magma chambers from column data
func (g *WorldGeology) collectMagmaChambers() []*underground.MagmaChamber {
	chambers := []*underground.MagmaChamber{}

	for _, col := range g.Columns.AllColumns() {
		if col.Magma != nil {
			chambers = append(chambers, &underground.MagmaChamber{
				Center: underground.Vector3{
					X: float64(col.X),
					Y: float64(col.Y),
					Z: (col.Magma.TopZ + col.Magma.BottomZ) / 2,
				},
				Temperature: col.Magma.Temperature,
				Pressure:    col.Magma.Pressure,
				Viscosity:   col.Magma.Viscosity,
			})
		}
	}

	return chambers
}

// simulateDepositEvolution processes organic deposit transformation
func (g *WorldGeology) simulateDepositEvolution(yearsElapsed int64) {
	if g.Columns == nil {
		return
	}

	// Build rainfall map from biomes for sedimentation calculation
	rainfall := make([]float64, len(g.Biomes))
	for i, biome := range g.Biomes {
		switch biome.Type {
		case "rainforest", "swamp":
			rainfall[i] = 1.0
		case "grassland", "savanna":
			rainfall[i] = 0.6
		case "forest", "taiga":
			rainfall[i] = 0.7
		case "tundra":
			rainfall[i] = 0.3
		case "desert":
			rainfall[i] = 0.1
		case "ocean", "beach":
			rainfall[i] = 0.8
		default:
			rainfall[i] = 0.5
		}
	}

	config := underground.DefaultDepositConfig()

	underground.SimulateDepositEvolution(
		g.Columns,
		g.TotalYearsSimulated,
		config,
		rainfall,
		g.Seed+g.TotalYearsSimulated,
	)
}

// SimulateGeology advances geological processes over time
// dt is the time step in years (Delta Time)
// globalTempMod is the current global temperature offset (e.g. from volcanic winter)
func (g *WorldGeology) SimulateGeology(dt int64, globalTempMod float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Heightmap == nil {
		return // Not initialized
	}

	g.TotalYearsSimulated += dt

	// Accumulate time for variable step processing
	dtFloat := float64(dt)
	g.TectonicStressAccumulator += dtFloat
	g.ErosionAccumulator += dtFloat
	g.DepositAccumulator += dtFloat
	g.RiverAccumulator += dtFloat
	g.MaintenanceAccumulator += dtFloat
	g.GeneralAccumulator += dtFloat

	// Plate movement: ~2cm/year = 0.00002 km/year
	// Over 1 million years = 20 km of movement
	// We accumulate movement and apply tectonic effects periodically

	// Apply plate tectonics every 100,000 years for efficiency
	tectonicInterval := 100_000.0
	if g.TectonicStressAccumulator >= tectonicInterval {
		// Calculate how many intervals passed
		intervals := int64(g.TectonicStressAccumulator / tectonicInterval)

		// Run tectonic updates
		for i := int64(0); i < intervals; i++ {
			g.advancePlates(tectonicInterval)

			// Fix 1: Re-enable Equilibrium Tectonics
			// Uses asymptotic approach to prevent runaway elevation
			if g.SphereHeightmap != nil && g.Topology != nil {
				g.SphereHeightmap = geography.SimulateTectonics(g.Plates, g.SphereHeightmap, g.Topology)
				g.syncSphereToFlat()
			}
		}

		// Keep remainder
		g.TectonicStressAccumulator -= float64(intervals) * tectonicInterval
	}

	// Apply erosion (more frequent)
	// Thermal erosion: 1 iteration per 10,000 years
	// We map the continuous erosion potential to discrete iterations
	erosionInterval := 10_000.0
	if g.ErosionAccumulator >= erosionInterval {
		intervals := int(g.ErosionAccumulator / erosionInterval)

		// Thermal erosion iterations
		// Cap iterations per frame to avoid lag spikes on huge updates, but for normal sim it's fine
		// 1 iteration per 10k years
		iterations := intervals
		if iterations > 0 {
			if iterations > 10 {
				iterations = 10
			} // Reasonable cap per frame
			geography.ApplyThermalErosion(g.Heightmap, iterations, g.Seed+g.TotalYearsSimulated)
		}

		// Hydraulic erosion: proportional to time but capped
		// 1000 drops per 10,000 years
		drops := int((float64(intervals) * 1000))
		if drops > 0 {
			if drops > 5000 {
				drops = 5000
			} // Cap
			geography.ApplyHydraulicErosion(g.Heightmap, drops, g.Seed+g.TotalYearsSimulated)
		}

		// Decrement accumulator
		// Note: We subtract what we actually processed (or intended to).
		// If we capped it, we technically "lost" some erosion, which improves stability.
		g.ErosionAccumulator -= float64(intervals) * erosionInterval
	}

	// Apply hotspot activity
	// This function already handles partial years probabilistically if needed,
	// or we can pass dtFloat.
	g.applyHotspotActivity(dtFloat)

	// Low frequency events using GeneralAccumulator
	// We can check multiple intervals

	// Cave formation (every 100,000 years)
	if g.GeneralAccumulator >= 100_000 {
		// Just run it once if we crossed the threshold, or loop if huge time jump?
		// For huge jumps (e.g. 1M years), running 10 times might be slow.
		// Let's run it once but scale the effect if supported, otherwise just once.
		// Detailed cave sim is expensive. Let's trigger it once if threshold crossed.
		// TODO: Better scaling for huge time jumps
		if g.Columns != nil {
			g.simulateCaveFormation(int64(g.GeneralAccumulator))
		}
	}

	// Tectonic Volcanism / Magma Chambers (every 10,000 years)
	// We'll use a modulo logic or just run it if enough time passed
	// This part is tricky with a single GeneralAccumulator.
	// Let's rely on the fact that if GeneralAccumulator is big, we execute these.

	// Ideally we'd have accumulators for each, but let's approximate:
	if dt >= 10_000 && g.Columns != nil {
		g.simulateMagmaChambers(dt)
	} else if g.GeneralAccumulator >= 10_000 && g.Columns != nil {
		// Run with accumulated time
		// Ideally we subtract from a specific accumulator.
		// Let's just pass dt for now if it's small steps summing up?
		// No, if dt=1, we call this every 10,000th call?
		// Let's simplify: only run these expensive detailed subsurface sims if dt is large
		// OR strictly periodically.
		// For standardized ticks, dt=1. We need strict periodic triggers.
		// We'll use the TotalYearsSimulated for modulo checks for these "optional" details,
		// OR separate accumulators.
		// Modulo checks on TotalYearsSimulated works for dt=1 steps.
		// For variable steps (dt=10000), modulo might skip.
		// Let's use the explicit checks on the total time logic which handles arbitrary jumps effectively IF we check ranges.
		// BUT easier approach:

		if g.GeneralAccumulator >= 10_000 {
			if g.Columns != nil {
				g.simulateMagmaChambers(10_000)
			}
			// Only subtract if we assume this is the main consumer of GeneralAcc
			// But we have multiple consumers.
		}
	}

	// Reset GeneralAccumulator if it gets too big (periodic cleanup)
	// or use it as a 10k year clock
	if g.GeneralAccumulator >= 100_000 {
		g.GeneralAccumulator = 0 // Reset after the longest cycle (Cave formation)
	}

	// Simulate organic deposit evolution (sedimentation and transformation)
	// Simulate organic deposit evolution (sedimentation and transformation)
	// Every 1,000 years
	depositInterval := 1_000.0
	if g.DepositAccumulator >= depositInterval && g.Columns != nil {
		intervals := int64(g.DepositAccumulator / depositInterval)
		// Run deposit sim
		// We pass the accumulated time
		accumulatedTime := int64(float64(intervals) * depositInterval)
		g.simulateDepositEvolution(accumulatedTime)

		// Decrement accumulator
		g.DepositAccumulator -= float64(accumulatedTime)
	}

	// Regenerate dynamic features using spherical algorithms
	// Rivers and biomes change as terrain evolves
	// We throttle this to every 1,000 years to avoid massive performance cost
	riverInterval := 1_000.0
	if g.RiverAccumulator >= riverInterval {
		if g.SphereHeightmap != nil {
			sphereRivers := geography.GenerateRiversSpherical(g.SphereHeightmap, g.SeaLevel, g.Seed+g.TotalYearsSimulated)
			g.Rivers = geography.ConvertSphericalRiversToFlat(sphereRivers, g.Topology.Resolution())
			g.syncSphereToFlat() // Sync river erosion to flat heightmap
		} else {
			g.Rivers = geography.GenerateRivers(g.Heightmap, g.SeaLevel, g.Seed+g.TotalYearsSimulated)
		}
		// Pass global temperature modifier to biome assignment
		// Uses new Weather→Biome pipeline
		g.Biomes = g.generateBiomesFromClimate(globalTempMod)

		// Decrement accumulator using modulo to keep phase but prevent buildup
		g.RiverAccumulator = math.Mod(g.RiverAccumulator, riverInterval)
	}

	// Fix 3: Apply isostatic adjustment & Maintenance
	// We throttle this to every 1,000 years
	maintenanceInterval := 1_000.0
	if g.MaintenanceAccumulator >= maintenanceInterval {
		// Calculate how much time this maintenance step represents
		accumulatedTime := g.MaintenanceAccumulator

		// Subside mountains
		// Rate: 0.01% per 10k years.
		// Scale by accumulatedTime.
		subsidenceRate := 1e-8 * accumulatedTime
		for i, elev := range g.Heightmap.Elevations {
			if elev > 8000 {
				excess := elev - 8000
				g.Heightmap.Elevations[i] -= excess * subsidenceRate
			}
		}

		// Fix 5: Global elevation clamping on SphereHeightmap
		if g.SphereHeightmap != nil {
			g.SphereHeightmap.ClampElevations(geography.MinElevation, geography.MaxElevation)
			g.syncSphereToFlat()
		} else {
			for i, elev := range g.Heightmap.Elevations {
				if elev > geography.MaxElevation {
					g.Heightmap.Elevations[i] = geography.MaxElevation
				} else if elev < geography.MinElevation {
					g.Heightmap.Elevations[i] = geography.MinElevation
				}
			}
		}

		// Update heightmap min/max stats
		g.updateHeightmapStats()

		// Reset accumulator (modulo)
		g.MaintenanceAccumulator = math.Mod(g.MaintenanceAccumulator, maintenanceInterval)
	}

	// Fix 4: Sea level equilibrium model - sea level recovers toward baseline
	// Recovery rate: 1% per 10k years = 0.01 / 10000 = 1e-6 per year
	targetSeaLevel := 0.0 // Baseline sea level
	recoveryRatePerYear := 1e-6
	seaLevelChange := (targetSeaLevel - g.SeaLevel) * recoveryRatePerYear * dtFloat
	g.SeaLevel += seaLevelChange

	// Fix 5: Global elevation clamping on SphereHeightmap
	// Ensures no runaway elevation accumulation over geological time
	if g.SphereHeightmap != nil {
		g.SphereHeightmap.ClampElevations(geography.MinElevation, geography.MaxElevation)
		// Sync to flat heightmap after clamping
		g.syncSphereToFlat()
	} else {
		// Fallback: clamp flat heightmap directly
		for i, elev := range g.Heightmap.Elevations {
			if elev > geography.MaxElevation {
				g.Heightmap.Elevations[i] = geography.MaxElevation
			} else if elev < geography.MinElevation {
				g.Heightmap.Elevations[i] = geography.MinElevation
			}
		}
	}

	// Update heightmap min/max
	g.updateHeightmapStats()
}

// applyHotspotActivity adds volcanic material at hotspot locations
func (g *WorldGeology) applyHotspotActivity(years float64) {
	// Probability of eruption per year at a hotspot
	// say 1 eruption every 1000 years
	numEruptions := int(years / 1000.0)
	if numEruptions == 0 && g.rng.Float64() < (years/1000.0) {
		numEruptions = 1
	}

	for _, hotspot := range g.Hotspots {
		for i := 0; i < numEruptions; i++ {
			// Small eruption
			// Jitter location slightly (within 2-3 pixels) to create a cluster/shield volcano
			jx := hotspot.X + (g.rng.Float64()*4 - 2)
			jy := hotspot.Y + (g.rng.Float64()*4 - 2)

			// Height addition (small, builds up over time)
			// 10-30m per eruption
			height := 10.0 + g.rng.Float64()*20.0
			radius := 1.5 // Small distinct cones

			geography.ApplyVolcanoFlat(g.Heightmap, jx, jy, radius, height)
		}
	}
}

// advancePlates moves tectonic plates and recalculates boundaries
// Uses great circle rotation on the sphere to move plate positions
func (g *WorldGeology) advancePlates(years float64) {
	// Planet radius in km (circumference / 2π)
	planetRadius := g.Circumference / (2 * math.Pi * 1000) // Convert m to km

	// Movement rate: ~2cm/year = 0.00002 km/year (average plate speed)
	plateSpeed := 0.00002 // km/year

	for i := range g.Plates {
		// Age the plate
		g.Plates[i].Age += years / 1_000_000 // Age in million years

		// Calculate rotation angle: θ = distance / radius = (speed * time) / radius
		distance := plateSpeed * years   // km moved
		theta := distance / planetRadius // radians

		// Get current position and velocity
		pos := g.Plates[i].Position
		vel := g.Plates[i].Velocity

		// Rotation axis = Position × Velocity (perpendicular to both)
		axis := pos.Cross(vel)
		if axis.Length() < 1e-9 {
			// Velocity is parallel to position - no meaningful rotation
			continue
		}

		// Rotate position around the axis
		newPos := pos.RotateAround(axis, theta)
		g.Plates[i].Position = newPos.Normalize() // Keep on unit sphere

		// Update centroid from new position
		if g.Topology != nil {
			g.Plates[i].Centroid = g.Topology.FromVector(newPos.X, newPos.Y, newPos.Z)
		}
	}

	// Re-assign plate regions after significant movement
	// Only do this every 10M years to avoid excessive computation
	if int64(years)%10_000_000 == 0 && g.Topology != nil {
		geography.ReassignPlateRegions(g.Plates, g.Topology)
	}
}

// ApplyEvent handles geological events that affect terrain
func (g *WorldGeology) ApplyEvent(event GeologicalEvent) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Heightmap == nil {
		return
	}

	switch event.Type {
	case EventVolcanicWinter:
		g.applyVolcanicMountains(event.Severity)

	case EventAsteroidImpact:
		g.applyImpactCrater(event.Severity)

	case EventIceAge:
		g.applyIceAgeEffects(event.Severity)

	case EventContinentalDrift:
		g.applyContinentalDrift(event.Severity)

	case EventFloodBasalt:
		g.applyFloodBasalt(event.Severity)

	// Ocean anoxia doesn't affect terrain
	case EventOceanAnoxia:
		// No terrain effect
	}

	g.updateHeightmapStats()
}

// applyVolcanicMountains adds volcanic features during volcanic winter
func (g *WorldGeology) applyVolcanicMountains(severity float64) {
	// Number of volcanoes based on severity
	numVolcanoes := 1 + int(severity*3)

	// Use spherical operations if available
	if g.SphereHeightmap != nil && g.Topology != nil {
		resolution := g.Topology.Resolution()
		for i := 0; i < numVolcanoes; i++ {
			// Random location on sphere
			face := g.rng.Intn(6)
			x := g.rng.Intn(resolution)
			y := g.rng.Intn(resolution)
			center := spatial.Coordinate{Face: face, X: x, Y: y}

			// Volcano height based on severity (200-500m per event)
			height := 200 + severity*300
			radius := 2.0 + g.rng.Float64()*2.0

			geography.ApplyVolcanoSpherical(g.SphereHeightmap, center, g.Topology, radius, height)
		}
		// Sync to flat heightmap
		g.syncSphereToFlat()
	} else {
		// Fallback to flat heightmap
		for i := 0; i < numVolcanoes; i++ {
			x := float64(g.rng.Intn(g.Heightmap.Width))
			y := float64(g.rng.Intn(g.Heightmap.Height))
			height := 200 + severity*300
			radius := 2.0 + g.rng.Float64()*2.0
			geography.ApplyVolcanoFlat(g.Heightmap, x, y, radius, height)
		}
	}
}

// applyImpactCrater creates a crater from asteroid impact
func (g *WorldGeology) applyImpactCrater(severity float64) {
	// Crater size based on severity (10-50 cells radius)
	radius := int(10 + severity*40)

	// Depth based on severity (500-3000m)
	depth := 500 + severity*2500

	// Rim height (15% of depth)
	rimHeight := depth * 0.15

	// Use spherical operations if available
	if g.SphereHeightmap != nil && g.Topology != nil {
		resolution := g.Topology.Resolution()
		// Random impact location on sphere
		centerFace := g.rng.Intn(6)
		centerX := g.rng.Intn(resolution)
		centerY := g.rng.Intn(resolution)
		center := spatial.Coordinate{Face: centerFace, X: centerX, Y: centerY}

		// Use BFS to apply crater with proper cross-face handling
		visited := make(map[spatial.Coordinate]bool)
		queue := []struct {
			coord spatial.Coordinate
			dist  int
		}{{center, 0}}
		visited[center] = true

		directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			dist := float64(current.dist)
			if dist < float64(radius) {
				// Inside crater - depression
				factor := 1.0 - (dist / float64(radius))
				currentElev := g.SphereHeightmap.Get(current.coord)
				g.SphereHeightmap.Set(current.coord, currentElev-depth*factor*factor)
			} else if dist < float64(radius)*1.3 {
				// Crater rim - raised
				t := (dist - float64(radius)) / (float64(radius) * 0.3)
				factor := 1.0 - t
				currentElev := g.SphereHeightmap.Get(current.coord)
				g.SphereHeightmap.Set(current.coord, currentElev+rimHeight*factor)
			}

			// Only expand if within extended radius
			if current.dist < int(float64(radius)*1.5) {
				for _, dir := range directions {
					neighbor := g.Topology.GetNeighbor(current.coord, dir)
					if !visited[neighbor] {
						visited[neighbor] = true
						queue = append(queue, struct {
							coord spatial.Coordinate
							dist  int
						}{neighbor, current.dist + 1})
					}
				}
			}
		}
		g.syncSphereToFlat()
	} else {
		// Fallback to flat heightmap
		centerX := g.rng.Intn(g.Heightmap.Width)
		centerY := g.rng.Intn(g.Heightmap.Height)

		for dy := -radius * 2; dy <= radius*2; dy++ {
			for dx := -radius * 2; dx <= radius*2; dx++ {
				px, py := centerX+dx, centerY+dy
				if px >= 0 && px < g.Heightmap.Width && py >= 0 && py < g.Heightmap.Height {
					dist := math.Sqrt(float64(dx*dx + dy*dy))

					if dist < float64(radius) {
						factor := 1.0 - (dist / float64(radius))
						current := g.Heightmap.Get(px, py)
						g.Heightmap.Set(px, py, current-depth*factor*factor)
					} else if dist < float64(radius)*1.3 {
						t := (dist - float64(radius)) / (float64(radius) * 0.3)
						factor := 1.0 - t
						current := g.Heightmap.Get(px, py)
						g.Heightmap.Set(px, py, current+rimHeight*factor)
					}
				}
			}
		}
	}
}

// applyIceAgeEffects lowers sea level and applies glacial erosion
func (g *WorldGeology) applyIceAgeEffects(severity float64) {
	// Sea level drop (50-120m based on severity)
	g.SeaLevel -= 50 + severity*70

	// Glacial erosion - carve U-shaped valleys in high-elevation areas
	if g.SphereHeightmap != nil && g.Topology != nil {
		// Apply to sphere heightmap
		threshold := g.SphereHeightmap.MaxElev * 0.6 // Top 40% of elevation
		resolution := g.Topology.Resolution()

		for face := 0; face < 6; face++ {
			for y := 0; y < resolution; y++ {
				for x := 0; x < resolution; x++ {
					coord := spatial.Coordinate{Face: face, X: x, Y: y}
					elev := g.SphereHeightmap.Get(coord)
					if elev > threshold {
						erosion := (elev - threshold) * 0.1 * severity
						g.SphereHeightmap.Set(coord, elev-erosion)
					}
				}
			}
		}
		g.syncSphereToFlat()
	} else {
		// Fallback to flat heightmap
		threshold := g.Heightmap.MaxElev * 0.6
		for y := 0; y < g.Heightmap.Height; y++ {
			for x := 0; x < g.Heightmap.Width; x++ {
				elev := g.Heightmap.Get(x, y)
				if elev > threshold {
					erosion := (elev - threshold) * 0.1 * severity
					g.Heightmap.Set(x, y, elev-erosion)
				}
			}
		}
	}
}

// applyContinentalDrift accelerates plate movement and simulates tectonic effects
// Note: Removed direct SimulateTectonics call to prevent additive elevation accumulation.
// Tectonic effects now use equilibrium-based approach applied during normal simulation.
func (g *WorldGeology) applyContinentalDrift(severity float64) {
	// Enhanced plate movement (accelerated by severity)
	extraYears := 50_000 + int64(severity*100_000)
	g.advancePlates(float64(extraYears))

	// Minor elevation adjustment at convergent boundaries
	// Instead of full SimulateTectonics, apply small equilibrium-based changes
	if g.SphereHeightmap != nil && g.Topology != nil {
		// Apply minor boundary uplift based on severity (max 100m per event)
		g.applyMinorBoundaryUplift(severity * 100)
		// Sync to flat heightmap for legacy consumers
		g.syncSphereToFlat()
	} else {
		// Fallback: simple uplift for when spherical data isn't available
		// Capped at 50m per event to prevent runaway growth
		uplift := 50 * severity
		if uplift > 50 {
			uplift = 50
		}
		for i := range g.Heightmap.Elevations {
			if g.Heightmap.Elevations[i] > g.SeaLevel {
				g.Heightmap.Elevations[i] += uplift
				// Apply cap
				if g.Heightmap.Elevations[i] > geography.MaxElevation {
					g.Heightmap.Elevations[i] = geography.MaxElevation
				}
			}
		}
	}
}

// applyMinorBoundaryUplift applies small elevation changes at plate boundaries.
// Uses equilibrium-based approach: moves toward target elevation rather than adding fixed amounts.
// maxChange limits the maximum elevation change per call to prevent runaway accumulation.
func (g *WorldGeology) applyMinorBoundaryUplift(maxChange float64) {
	if g.SphereHeightmap == nil || g.Topology == nil || len(g.Plates) == 0 {
		return
	}

	// Build reverse lookup: coordinate -> plate index
	coordToPlate := make(map[spatial.Coordinate]int)
	for i, p := range g.Plates {
		for coord := range p.Region {
			coordToPlate[coord] = i
		}
	}

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}
	resolution := g.Topology.Resolution()

	// Process all cells to detect boundaries
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				currentPlateIdx, exists := coordToPlate[coord]
				if !exists {
					continue
				}
				currentPlate := g.Plates[currentPlateIdx]

				// Check neighbors for boundary
				for _, dir := range directions {
					neighbor := g.Topology.GetNeighbor(coord, dir)
					neighborPlateIdx, exists := coordToPlate[neighbor]
					if !exists || neighborPlateIdx == currentPlateIdx {
						continue
					}

					// Found a boundary between two plates
					neighborPlate := g.Plates[neighborPlateIdx]
					boundaryType := geography.CalculateBoundaryType(currentPlate, neighborPlate)

					// Get target and current elevation
					targetElev := geography.GetTargetElevation(currentPlate, neighborPlate, boundaryType)
					currentElev := g.SphereHeightmap.Get(coord)

					// Calculate equilibrium change (10% of difference)
					delta := (targetElev - currentElev) * 0.1

					// Cap the change to prevent large swings
					if delta > maxChange {
						delta = maxChange
					} else if delta < -maxChange {
						delta = -maxChange
					}

					// Apply change with clamping
					newElev := currentElev + delta
					if newElev > geography.MaxElevation {
						newElev = geography.MaxElevation
					} else if newElev < geography.MinElevation {
						newElev = geography.MinElevation
					}
					g.SphereHeightmap.Set(coord, newElev)
				}
			}
		}
	}
}

// applyFloodBasalt creates large volcanic provinces
func (g *WorldGeology) applyFloodBasalt(severity float64) {
	// Radius based on severity (30-100 cells)
	radius := 30 + int(severity*70)

	// Height of basalt layers (100-500m)
	height := 100 + severity*400

	// Use spherical operations if available
	if g.SphereHeightmap != nil && g.Topology != nil {
		resolution := g.Topology.Resolution()
		// Random center on sphere
		centerFace := g.rng.Intn(6)
		centerX := g.rng.Intn(resolution)
		centerY := g.rng.Intn(resolution)
		center := spatial.Coordinate{Face: centerFace, X: centerX, Y: centerY}

		// Use BFS to apply basalt with proper cross-face handling
		visited := make(map[spatial.Coordinate]bool)
		queue := []struct {
			coord spatial.Coordinate
			dist  int
		}{{center, 0}}
		visited[center] = true

		directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if current.dist < radius {
				dist := float64(current.dist)
				factor := 1.0 - (dist / float64(radius))
				factor = factor * factor // Smoother falloff

				currentElev := g.SphereHeightmap.Get(current.coord)
				g.SphereHeightmap.Set(current.coord, currentElev+height*factor)

				// Expand to neighbors
				for _, dir := range directions {
					neighbor := g.Topology.GetNeighbor(current.coord, dir)
					if !visited[neighbor] {
						visited[neighbor] = true
						queue = append(queue, struct {
							coord spatial.Coordinate
							dist  int
						}{neighbor, current.dist + 1})
					}
				}
			}
		}
		g.syncSphereToFlat()
	} else {
		// Fallback to flat heightmap
		centerX := g.rng.Intn(g.Heightmap.Width)
		centerY := g.rng.Intn(g.Heightmap.Height)

		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				px, py := centerX+dx, centerY+dy
				if px >= 0 && px < g.Heightmap.Width && py >= 0 && py < g.Heightmap.Height {
					dist := math.Sqrt(float64(dx*dx + dy*dy))
					if dist < float64(radius) {
						factor := 1.0 - (dist / float64(radius))
						factor = factor * factor
						current := g.Heightmap.Get(px, py)
						g.Heightmap.Set(px, py, current+height*factor)
					}
				}
			}
		}
	}
}

// updateHeightmapStats recalculates min/max elevation
func (g *WorldGeology) updateHeightmapStats() {
	minElev, maxElev := math.MaxFloat64, -math.MaxFloat64
	for _, val := range g.Heightmap.Elevations {
		if val < minElev {
			minElev = val
		}
		if val > maxElev {
			maxElev = val
		}
	}
	g.Heightmap.MinElev = minElev
	g.Heightmap.MaxElev = maxElev
}

// GetStats returns current geological statistics
func (g *WorldGeology) GetStats() GeologyStats {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.Heightmap == nil {
		return GeologyStats{PlateCount: len(g.Plates)}
	}

	// Calculate average elevation
	sum := 0.0
	landCount := 0
	for _, elev := range g.Heightmap.Elevations {
		sum += elev
		if elev > g.SeaLevel {
			landCount++
		}
	}

	totalPixels := float64(len(g.Heightmap.Elevations))
	avgElev := sum / totalPixels
	landPercent := float64(landCount) / totalPixels * 100

	// Calculate average temperature
	avgTemp := 0.0
	if len(g.Biomes) > 0 {
		totalTemp := 0.0
		for _, b := range g.Biomes {
			totalTemp += b.Temperature
		}
		avgTemp = totalTemp / float64(len(g.Biomes))
	}

	return GeologyStats{
		AverageElevation:   avgElev,
		AverageTemperature: avgTemp,
		MaxElevation:       g.Heightmap.MaxElev,
		MinElevation:       g.Heightmap.MinElev,
		SeaLevel:           g.SeaLevel,
		LandPercent:        landPercent,
		PlateCount:         len(g.Plates),
		HotspotCount:       len(g.Hotspots),
		RiverCount:         len(g.Rivers),
		BiomeCount:         len(g.Biomes),
		YearsSimulated:     g.TotalYearsSimulated,
	}
}

// IsInitialized returns whether geology has been set up
func (g *WorldGeology) IsInitialized() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Heightmap != nil
}

// TriggerTectonicCollision player-triggered plate collision forming mountain range
// magnitude 0.0-1.0 controls mountain height (2000-6000m)
func (g *WorldGeology) TriggerTectonicCollision(x, y float64, magnitude float32) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Heightmap == nil {
		return
	}

	// Mountain range height based on magnitude
	height := 2000 + float64(magnitude)*4000 // 2000-6000m
	length := 5.0 + float64(magnitude)*10.0  // 5-15 cells long

	// Create mountain range at specified location
	angle := g.rng.Float64() * math.Pi // Random orientation

	for i := 0.0; i < length; i++ {
		// Calculate position along range
		px := x + math.Cos(angle)*i*2
		py := y + math.Sin(angle)*i*2

		// Wrap coordinates
		if int(px) < 0 || int(px) >= g.Heightmap.Width ||
			int(py) < 0 || int(py) >= g.Heightmap.Height {
			continue
		}

		// Apply mountain with some variation
		peakHeight := height * (1.0 + (g.rng.Float64()-0.5)*0.4)
		radius := 2.0 + g.rng.Float64()*1.5

		geography.ApplyVolcanoFlat(g.Heightmap, px, py, radius, peakHeight)
	}

	g.updateHeightmapStats()
}

// TriggerCatastrophe triggers a player-initiated catastrophic event
// eventType: "volcano", "asteroid", "flood_basalt", "ice_age"
// magnitude 0.0-1.0 controls severity
func (g *WorldGeology) TriggerCatastrophe(eventType string, magnitude float32) {
	g.mu.Lock()
	defer g.mu.Unlock()

	severity := float64(magnitude)

	switch eventType {
	case "volcano":
		g.applyVolcanicMountains(severity)
	case "asteroid":
		g.applyImpactCrater(severity)
	case "flood_basalt":
		g.applyFloodBasalt(severity)
	case "ice_age":
		g.applyIceAgeEffects(severity)
	case "continental_drift":
		g.applyContinentalDrift(severity)
	}

	g.updateHeightmapStats()
}

// ShiftTemperature applies a global temperature change to all biomes
// shift is in degrees Celsius (positive = warming, negative = cooling)
func (g *WorldGeology) ShiftTemperature(shift float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for i := range g.Biomes {
		g.Biomes[i].Temperature += shift
	}
}

// generateBiomesFromClimate uses the Weather→Biome pipeline.
// This is the correct causal chain: Weather determines temperature,
// which determines biome type (no latitude math in biomes.go).
func (g *WorldGeology) generateBiomesFromClimate(globalTempMod float64) []geography.Biome {
	seed := g.Seed + g.TotalYearsSimulated

	// 1. Generate climate data from Weather service
	climateData := weather.GenerateInitialClimate(g.Heightmap, g.SeaLevel, seed, globalTempMod)

	// 2. Classify biomes using climate data
	biomes := make([]geography.Biome, g.Heightmap.Width*g.Heightmap.Height)
	for y := 0; y < g.Heightmap.Height; y++ {
		for x := 0; x < g.Heightmap.Width; x++ {
			idx := y*g.Heightmap.Width + x
			elev := g.Heightmap.Get(x, y)
			climate := weather.GetClimateAt(climateData, g.Heightmap.Width, x, y)

			biomeType := geography.ClassifyBiome(
				climate.Temperature,
				climate.AnnualRainfall,
				climate.SoilDrainage,
				elev,
				g.SeaLevel,
			)

			biomes[idx] = geography.Biome{
				BiomeID:       uuid.New(),
				Name:          string(biomeType),
				Type:          biomeType,
				Temperature:   climate.Temperature,
				Precipitation: climate.AnnualRainfall,
			}
		}
	}

	return biomes
}

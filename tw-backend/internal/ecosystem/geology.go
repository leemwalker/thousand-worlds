package ecosystem

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
	"tw-backend/internal/debug"
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
	Provinces       []geography.GeologicalProvince // Sub-regions within continental plates (Phase 5)
	SeaLevel        float64                        // meters (0 = baseline, positive = higher sea level)
	Topology        spatial.Topology               // Spherical topology for plate operations
	BoundaryCache   *geography.BoundaryCache       // Cached plate boundary cells for fast tectonic processing

	// Underground data (Phase 3)
	Columns               *underground.ColumnGrid     // Per-column underground data
	Caves                 []*underground.Cave         // Cave networks
	Composition           string                      // "volcanic", "continental", "oceanic", "ancient"
	MagmaChambers         []*underground.MagmaChamber // Active magma chambers
	InactiveMagmaDeposits []*underground.MagmaChamber // Archived solidified chambers for map generation
	MagmaGCCounter        int64                       // Counter for periodic GC

	// Dynamic geographic features
	Hotspots   []geography.Point // Fixed mantle plume locations
	Rivers     [][]geography.Point
	Biomes     []geography.Biome
	Satellites []astronomy.Satellite // Natural satellites
	Rainfall   []float64             // Per-cell rainfall (Phase 7: Dynamic Weather)
	IceSheet   *geography.IceSheet   // Glacial ice dynamics (Phase 10)

	// Simulation state
	TotalYearsSimulated int64
	rng                 *rand.Rand

	// Scale factors (pixels to real-world)
	PixelsPerKm float64 // How many heightmap pixels per real km

	// Time Accumulators for variable step simulation
	TectonicStressAccumulator    float64 // Years of accumulated tectonic stress
	ErosionAccumulator           float64 // Years of accumulated erosion potential
	DepositAccumulator           float64 // Years of accumulated organic deposit time
	RiverAccumulator             float64 // Years of accumulated river/biome update time
	WeatherAccumulator           float64 // Years of accumulated weather/rainfall update time
	MaintenanceAccumulator       float64 // Years of accumulated maintenance time (subsidence, clamping, stats)
	GeneralAccumulator           float64 // Years of accumulated time for lower frequency events
	PlateReassignmentAccumulator float64 // Years since last plate region reassignment (triggers drift)

	// Sync optimization: track when sphere heightmap needs to be synced to flat
	// Set by event handlers, cleared after actual sync
	sphereNeedsSync bool

	// Ocean phase state (Hadean vapor → Modern liquid transition)
	OceanVaporFraction float64 // 0.0 = all liquid (cool planet), 1.0 = all vapor (hot planet)
}

// PhaseTransitionEvent represents a major planetary phase change
type PhaseTransitionEvent struct {
	Type        string // "GreatDeluge", etc.
	Year        int64  // Year when event occurred
	Description string // Human-readable description
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

// GetPlanetaryHeat returns a heat multiplier based on planetary age.
// Models Earth's thermal evolution from Hadean magma ocean to modern stable planet.
//
// Physics:
//   - Hadean period (0-500M years): Rapid cooling as magma ocean solidifies
//     Linear decay: 10.0 → 4.0 (represents ~150°C/My cooling rate)
//   - Radiogenic period (500M-4.5B years): Exponential decay from radioactive isotopes
//     Exponential decay: 4.0 → 1.0 (U-238, Th-232, K-40 half-lives)
//
// Returns:
//   - 10.0 at formation (year 0): Extreme volcanism, tectonics, geothermal flux
//   - 4.0 at Hadean boundary (500M years): Late heavy bombardment ending
//   - 1.0 at modern age (4.5B years): Current Earth baseline
//   - Never falls below 1.0 (residual heat + tidal heating)
func GetPlanetaryHeat(year int64) float64 {
	// Handle edge cases
	if year < 0 {
		year = 0
	}

	const (
		hadeanEnd      = 500_000_000   // 500 million years
		modernAge      = 4_500_000_000 // 4.5 billion years
		hadeanHeat     = 10.0          // Initial heat multiplier
		transitionHeat = 4.0           // Heat at end of Hadean
		modernHeat     = 1.0           // Baseline modern heat
	)

	if year < hadeanEnd {
		// Hadean regime: Linear cooling from 10.0 to 4.0
		// Represents rapid surface cooling and magma ocean solidification
		progress := float64(year) / float64(hadeanEnd)
		return hadeanHeat - (hadeanHeat-transitionHeat)*progress
	}

	// Radiogenic regime: Exponential decay from 4.0 to 1.0
	// Solve for decay constant λ such that Heat(4.5B) = 1.0
	// Formula: H(t) = (H₀ - H∞)e^(-λt) + H∞
	// Where H∞ = 1.0 (asymptotic minimum)
	// At t=0 (relative to Hadean end): H = 4.0
	// At t=4.0B: H = 1.0
	//
	// 1.0 = (4.0 - 1.0)e^(-λ * 4.0B) + 1.0
	// 0 = 3.0 * e^(-λ * 4.0B)
	// This doesn't work (can't reach exactly 1.0)
	//
	// Better formula: H(t) = H∞ + (H₀ - H∞)e^(-λt)
	// 1.0 = 1.0 + (4.0 - 1.0)e^(-λ * 4.0B)
	// 0 = 3.0 * e^(-λ * 4.0B)
	// Still problematic. Use alternate approach:
	//
	// Let's use: H(t) = 1.0 + 3.0 * e^(-λt)
	// At t=0: H = 1.0 + 3.0 = 4.0 ✓
	// At t=4.0B: H = 1.0 + 3.0 * e^(-λ * 4.0B) ≈ 1.0
	// We want: 3.0 * e^(-λ * 4.0B) ≈ 0
	// e^(-λ * 4.0B) = 0.01 (1% remaining)
	// -λ * 4.0B = ln(0.01) = -4.605
	// λ = 4.605 / 4.0B = 1.15125e-9

	const decayConstant = 1.15125e-9 // per year
	yearsPostHadean := float64(year - hadeanEnd)

	heat := modernHeat + (transitionHeat-modernHeat)*math.Exp(-decayConstant*yearsPostHadean)

	// Ensure heat never falls below baseline
	if heat < modernHeat {
		return modernHeat
	}

	return heat
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
	// For Earth-like (40,000 km), this gives 4000x2000
	// Updated: Cap at 2048x1024 per user request for high fidelity
	maxWidth := 2048
	maxHeight := 1024

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
	g.Plates = geography.GeneratePlates(plateCount, g.Topology, g.Seed, 0.30)

	// Generate initial heightmap using spherical topology
	// Create sphere heightmap and convert to flat for legacy consumers
	g.SphereHeightmap = geography.NewSphereHeightmap(g.Topology)
	g.SphereHeightmap = geography.GenerateHeightmap(g.Plates, g.SphereHeightmap, g.Topology, g.Seed, 1.0, 1.0)
	g.Heightmap = g.SphereHeightmap.ToFlatHeightmap(width, height)

	// Phase 5: Generate geological provinces within continental plates
	// This creates Cratons (hard, flat), Orogens (folded, medium), and Basins (soft, low)
	g.Provinces = geography.GenerateProvinces(g.Plates, g.Topology, g.Seed)
	geography.InitializeProvinceHardness(g.SphereHeightmap, g.Plates, g.Provinces, g.Topology, g.Seed)

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

	// Generate initial rivers and hydrology
	if g.SphereHeightmap != nil {
		// Phase 7: Generate rainfall map from atmosphere simulation
		rainfallConfig := weather.DefaultRainfallConfig(g.SeaLevel)
		rawRainfall := weather.GenerateRainfallMap(g.SphereHeightmap, g.Topology, rainfallConfig)

		// Normalize rainfall to prevent sudden massive erosion
		// ScalingFactor = TotalCells / TotalRainfall (target mean of 1.0)
		totalCells := 6 * g.Topology.Resolution() * g.Topology.Resolution()
		totalRainfall := 0.0
		for _, r := range rawRainfall {
			totalRainfall += r
		}

		scalingFactor := float64(totalCells) / totalRainfall
		if totalRainfall == 0 || scalingFactor > 100 {
			scalingFactor = 1.0 // Fallback to uniform
		}

		g.Rainfall = make([]float64, len(rawRainfall))
		for i, r := range rawRainfall {
			g.Rainfall[i] = r * scalingFactor
		}

		// Pass normalized rainfall to flux calculation
		geography.CalculateGlobalFluxWithRainfall(g.SphereHeightmap, g.Rainfall)

		// River Erosion (Phase 6b)
		// Carve valleys along high-flux paths before lake filling
		geography.ApplyRiverErosion(g.SphereHeightmap, 50.0, 5.0, g.SeaLevel)

		geography.FillDepressions(g.SphereHeightmap, g.SeaLevel)

		sphereRivers := geography.GenerateRiversSpherical(g.SphereHeightmap, g.SeaLevel, g.Seed)
		g.Rivers = geography.ConvertSphericalRiversToFlat(sphereRivers, g.Topology.Resolution())
		// Sync sphere heightmap changes from river erosion
		g.markSphereNeedsSync()
	} else {
		g.Rivers = geography.GenerateRivers(g.Heightmap, g.SeaLevel, g.Seed)
	}

	// Initialize biomes using Weather→Biome pipeline (no latitude coupling)
	g.Biomes = g.UpdateBiomes(0.0) // No global temp modifier initially

	// Initialize underground column grid (Phase 3)
	g.initializeColumns(width, height)

	// Phase 10: Initialize Ice Sheet (if not exists)
	if g.IceSheet == nil {
		g.IceSheet = geography.NewIceSheet(g.Topology.Resolution())
	}
}

// markSphereNeedsSync marks that the sphere heightmap has been modified
// and needs to be synced to the flat heightmap. The actual sync will happen
// at the end of the iteration when flushSync is called.
func (g *WorldGeology) markSphereNeedsSync() {
	g.sphereNeedsSync = true
}

// syncSphereToFlat updates the flat Heightmap from the SphereHeightmap
// Call this after making changes to SphereHeightmap to keep both in sync
// DEPRECATED: Use markSphereNeedsSync() instead and let flushSync() handle it
func (g *WorldGeology) syncSphereToFlat() {
	if debug.Is(debug.Perf | debug.Geology) {
		defer debug.Time(debug.Perf, "syncSphereToFlat")()
	}
	if g.SphereHeightmap == nil || g.Heightmap == nil {
		return
	}
	// Use in-place version to avoid memory allocation each sync
	g.SphereHeightmap.ToFlatHeightmapInPlace(g.Heightmap)
	g.sphereNeedsSync = false
}

// flushSync performs a batched sync if the sphere heightmap has been modified
// Call this once at the end of SimulateGeology instead of syncing after each operation
func (g *WorldGeology) flushSync() {
	if !g.sphereNeedsSync {
		return
	}
	g.syncSphereToFlat()
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
// Called during SimulateGeology every 10M years
func (g *WorldGeology) simulateCaveFormation(yearsElapsed int64) {
	if g.Columns == nil {
		return
	}

	// CAP: Skip if at capacity
	if len(g.Caves) >= underground.MaxActiveCaves {
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

// convertBoundaryCacheToUnderground converts cached boundary cells to the underground format
// avoiding expensive re-calculation of Voronoi regions
func (g *WorldGeology) convertBoundaryCacheToUnderground(
	cache *geography.BoundaryCache,
	centroids []underground.Vector3,
	movements []underground.Vector3,
) []underground.TectonicBoundary {
	boundaries := make([]underground.TectonicBoundary, 0, len(cache.Cells))
	faceWidth := g.Heightmap.Width / 6

	for _, cell := range cache.Cells {
		// Calculate intensity based on relative velocity (same logic as underground.GetTectonicBoundaries)
		plateIdx := cell.PlateIdx
		neighborIdx := cell.NeighborIdx

		// Skip invalid indices
		if plateIdx < 0 || plateIdx >= len(centroids) || neighborIdx < 0 || neighborIdx >= len(centroids) {
			continue
		}

		// Vectors
		p1 := centroids[plateIdx]
		p2 := centroids[neighborIdx]
		v1 := movements[plateIdx]
		v2 := movements[neighborIdx]

		// Direction vector between centroids
		dx := p2.X - p1.X
		dy := p2.Y - p1.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist == 0 {
			continue
		}
		dx, dy = dx/dist, dy/dist

		// Relative velocity projected onto direction vector
		relVel := (v2.X-v1.X)*dx + (v2.Y-v1.Y)*dy

		// Determine type and intensity
		// Logic matches underground.GetTectonicBoundaries
		intensity := 0.5
		if relVel < -0.2 {
			// Convergent
			intensity = math.Min(1.0, math.Abs(relVel))
		} else if relVel > 0.2 {
			// Divergent
			intensity = math.Min(1.0, relVel)
		} else {
			// Transform (default)
		}

		// Convert coordinate to flat map
		flatX := cell.Coord.Face*faceWidth + cell.Coord.X
		flatY := cell.Coord.Y

		boundaries = append(boundaries, underground.TectonicBoundary{
			X:            flatX,
			Y:            flatY,
			BoundaryType: string(cell.BoundaryType),
			Intensity:    intensity,
		})
	}
	return boundaries
}

// simulateMagmaChambers processes magma chamber evolution and tectonic volcanism
// Uses spatial hashing for O(chunks) boundary lookup and periodic GC for solidified chambers
func (g *WorldGeology) simulateMagmaChambers(yearsElapsed int64) {
	if g.Columns == nil || len(g.Plates) == 0 {
		return
	}

	totalStart := time.Now()

	// Extract tectonic boundaries from plate data
	plateCentroids := make([]underground.Vector3, len(g.Plates))
	plateMovements := make([]underground.Vector3, len(g.Plates))
	for i, plate := range g.Plates {
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

	// Ensure boundary cache exists
	cacheStart := time.Now()
	if g.BoundaryCache == nil || !g.BoundaryCache.Valid {
		g.BoundaryCache = geography.ComputeBoundaryCache(g.Plates, g.Topology)
	}

	// Convert cache to underground boundaries
	var boundaries []underground.TectonicBoundary
	if g.BoundaryCache != nil && g.BoundaryCache.Valid {
		boundaries = g.convertBoundaryCacheToUnderground(g.BoundaryCache, plateCentroids, plateMovements)
	} else {
		boundaries = underground.GetTectonicBoundaries(
			g.Heightmap.Width,
			g.Heightmap.Height,
			plateCentroids,
			plateMovements,
		)
	}
	cacheTime := time.Since(cacheStart)

	// OPTIMIZATION: Build spatial index for O(1) boundary lookups
	gridStart := time.Now()
	boundaryGrid := underground.NewBoundaryGrid(boundaries, 32) // 32-unit chunks
	gridTime := time.Since(gridStart)

	// Initialize chambers from stored state or collect from columns on first run
	if g.MagmaChambers == nil {
		g.MagmaChambers = g.collectMagmaChambers()
	}

	config := underground.DefaultMagmaConfig()
	if g.Composition == "volcanic" {
		config.EruptionThreshold = 60
		config.LavaTubeFormationProb = 0.9
	}

	// Run magma simulation with spatial hashing
	simStart := time.Now()
	erupted, newTubes, _ := underground.SimulateMagmaChambersWithGrid(
		g.Columns,
		g.MagmaChambers,
		boundaryGrid,
		yearsElapsed,
		g.Seed+g.TotalYearsSimulated,
		config,
	)
	simTime := time.Since(simStart)

	// Handle eruptions - apply surface effects
	for _, chamber := range erupted {
		x, y := int(chamber.Center.X), int(chamber.Center.Y)
		if x >= 0 && x < g.Heightmap.Width && y >= 0 && y < g.Heightmap.Height {
			height := 500 + g.rng.Float64()*1500
			radius := 2.0 + g.rng.Float64()*3.0

			if g.SphereHeightmap != nil && g.Topology != nil {
				// Convert flat map (Equirectangular) coordinate to spherical coordinate
				// U = X / Width, V = Y / Height
				// Lon = (U - 0.5) * 2 * PI
				// Lat = (V - 0.5) * PI
				u := float64(x) / float64(g.Heightmap.Width)
				v := float64(y) / float64(g.Heightmap.Height)

				lon := (u - 0.5) * 2.0 * math.Pi
				lat := (v - 0.5) * math.Pi

				// Convert Spherical to Cartesian
				// Z is Up (Lat), X/Y is Plane
				// Note: Coordinate system details depend on engine, assuming typical:
				// x = cos(lat) * cos(lon)
				// y = cos(lat) * sin(lon)
				// z = sin(lat)

				vecX := math.Cos(lat) * math.Cos(lon)
				vecY := math.Cos(lat) * math.Sin(lon)
				vecZ := math.Sin(lat)

				coord := g.Topology.FromVector(vecX, vecY, vecZ)
				geography.ApplyVolcanoSpherical(g.SphereHeightmap, coord, g.Topology, radius, height)
				g.markSphereNeedsSync()
			} else {
				geography.ApplyVolcanoFlat(g.Heightmap, float64(x), float64(y), radius, height)
			}
		}
	}

	// Register new lava tubes as caves
	g.Caves = append(g.Caves, newTubes...)

	// OPTIMIZATION: Periodic garbage collection of solidified chambers
	g.MagmaGCCounter++
	if g.MagmaGCCounter >= 100 {
		gcStart := time.Now()
		active, solidified := underground.CompactChambers(g.MagmaChambers)
		g.MagmaChambers = active
		g.InactiveMagmaDeposits = append(g.InactiveMagmaDeposits, solidified...)
		g.MagmaGCCounter = 0
		gcTime := time.Since(gcStart)

		if debug.Is(debug.Perf | debug.Geology) {
			log.Printf("[MAGMA GC] Compacted: %d active, %d archived | Time: %v",
				len(active), len(solidified), gcTime)
		}
	}

	totalTime := time.Since(totalStart)

	// Diagnostic logging (every 1M years)
	if g.TotalYearsSimulated%1_000_000 == 0 {
		log.Printf("[MAGMA PROFILE] Chambers: %d | Boundaries: %d | Erupted: %d | NewTubes: %d | Cache: %v | Grid: %v | Sim: %v | Total: %v",
			len(g.MagmaChambers), len(boundaries), len(erupted), len(newTubes),
			cacheTime, gridTime, simTime, totalTime)
	}
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
// Returns a PhaseTransitionEvent if a major phase change occurred (e.g., Great Deluge)
func (g *WorldGeology) SimulateGeology(dt int64, globalTempMod float64) *PhaseTransitionEvent {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Heightmap == nil {
		return nil // Not initialized
	}

	// === DEEP PROFILING ===
	var tectonicTime, biomeTime, oceanPhaseTime, statsTime, erosionTime, magmaTime, caveTime time.Duration
	var erosionStart time.Time
	profilingEnabled := g.TotalYearsSimulated%1_000_000 == 0 // Log every 1M years (was 10M)

	g.TotalYearsSimulated += dt

	// Calculate planetary heat multiplier for this time period
	// This drives tectonic and volcanic activity rates
	heat := GetPlanetaryHeat(g.TotalYearsSimulated)

	// Accumulate time for variable step processing
	// Tectonic stress scales with planetary heat (10x faster in early Earth)
	dtFloat := float64(dt)
	g.TectonicStressAccumulator += dtFloat * heat
	g.ErosionAccumulator += dtFloat
	g.DepositAccumulator += dtFloat
	g.RiverAccumulator += dtFloat
	g.MaintenanceAccumulator += dtFloat
	g.WeatherAccumulator += dtFloat
	g.GeneralAccumulator += dtFloat

	// OPTIMIZATION: Cap all accumulators to prevent explosion when crossing heat thresholds
	// During Hadean (heat > 4.0), erosion/river/maintenance are skipped but accumulators grow.
	// When heat drops to 4.0, they'd trigger millions of catch-up operations.
	// Cap each to a reasonable maximum (10 intervals worth).
	const maxAccumulatorValue = 1_000_000.0 // Max 1M years of accumulated time
	if g.TectonicStressAccumulator > maxAccumulatorValue*10 {
		g.TectonicStressAccumulator = maxAccumulatorValue * 10
	}
	if g.ErosionAccumulator > maxAccumulatorValue {
		g.ErosionAccumulator = maxAccumulatorValue
	}
	if g.RiverAccumulator > maxAccumulatorValue {
		g.RiverAccumulator = maxAccumulatorValue
	}
	if g.MaintenanceAccumulator > maxAccumulatorValue {
		g.MaintenanceAccumulator = maxAccumulatorValue
	}
	if g.GeneralAccumulator > maxAccumulatorValue {
		g.GeneralAccumulator = maxAccumulatorValue
	}
	// Weather interval is 5M, so cap needs to be higher (e.g. 10M)
	if g.WeatherAccumulator > maxAccumulatorValue*10 {
		g.WeatherAccumulator = maxAccumulatorValue * 10
	}

	// [... rest of existing SimulateGeology code stays the same until line 762 ...]

	// Plate movement: ~2cm/year = 0.00002 km/year
	// Over 1 million years = 20 km of movement
	// We accumulate movement and apply tectonic effects periodically

	// Apply plate tectonics with adaptive frequency
	// Optimization: Reduce frequency during molten Hadean eon to save performance
	tectonicStart := time.Now()

	tectonicInterval := 100_000.0 // Default (modern precision)
	if heat > 4.0 {
		// Hadean: Heat is ~10.0, so accumulator grows 10x faster.
		// To run once every 10 steps (1M real years), we need a 10M threshold.
		tectonicInterval = 10_000_000.0
	} else if heat > 1.5 {
		// Archean: Heat is ~2-4. To run every ~500k real years:
		tectonicInterval = 2_000_000.0
	}

	if g.TectonicStressAccumulator >= tectonicInterval {
		// Calculate how many intervals passed
		intervals := int64(g.TectonicStressAccumulator / tectonicInterval)

		// OPTIMIZATION: Cap intervals per iteration to prevent accumulator explosion
		// When crossing heat thresholds (e.g., 500M years), the interval changes
		// dramatically, which could cause hundreds of tectonic updates in one iteration.
		// Cap at 5 updates per iteration to keep things responsive.
		const maxIntervalsPerIteration = 5
		if intervals > maxIntervalsPerIteration {
			intervals = maxIntervalsPerIteration
		}

		// Run tectonic updates
		// Since we scaled the interval, we also scale the tectonic effect
		scaleFactor := tectonicInterval / 100_000.0

		for i := int64(0); i < intervals; i++ {
			g.advancePlates(tectonicInterval)

			// Fix 1: Re-enable Equilibrium Tectonics
			// Uses asymptotic approach to prevent runaway elevation
			if g.SphereHeightmap != nil && g.Topology != nil {
				// Debug timing for tectonics specifically
				// Debug timing for tectonics specifically
				tectonicUpdateStart := time.Now()

				// Use cached version if available
				// Cache is built once and persists until plates are reassigned
				// (advancePlates moves plates slightly but doesn't change which cells are at boundaries)
				if g.BoundaryCache == nil || !g.BoundaryCache.Valid {
					// Rebuild cache (expensive, but only needed when regions change)
					if debug.Is(debug.Perf | debug.Geology) {
						log.Printf("[BOUNDARY CACHE] Rebuilding... (Reason: Nil=%v, Valid=%v)", g.BoundaryCache == nil, g.BoundaryCache != nil && g.BoundaryCache.Valid)
					}
					g.BoundaryCache = geography.ComputeBoundaryCache(g.Plates, g.Topology)
				}
				g.SphereHeightmap = geography.SimulateTectonicsWithCache(g.Plates, g.SphereHeightmap, g.BoundaryCache, g.Topology, scaleFactor, g.Seed)

				// Apply passive margin decay - erode cells no longer at boundaries
				// This prevents phantom mountains from persisting after plate boundaries move
				geography.ApplyBoundaryDecay(g.Plates, g.SphereHeightmap, g.BoundaryCache, g.Topology, scaleFactor, g.Seed)

				g.markSphereNeedsSync()

				if debug.Is(debug.Tectonics | debug.Perf) {
					log.Printf("[Perf] TectonicsUpdate took %v", time.Since(tectonicUpdateStart))
				}
			}
		}

		// Keep remainder
		g.TectonicStressAccumulator -= float64(intervals) * tectonicInterval
	}
	tectonicTime = time.Since(tectonicStart)

	// === HADEAN OPTIMIZATION ===
	// Skip expensive surface processes on molten early Earth (heat > 4.0)
	// During Hadean eon (~first 500M years), planet is a lava ocean
	// No solid crust for erosion, no caves, no rivers - only plate tectonics matter
	// This provides ~100× speedup for deep time simulations
	if heat <= 4.0 {
		erosionStart = time.Now()

		// === EROSION (Throttled for deep-time - every 10M years) ===
		// Surface processes only matter on cooled planets with solid crust
		erosionInterval := 10_000_000.0 // 10M years (was 10K)
		if g.ErosionAccumulator >= erosionInterval {
			// Thermal erosion: Limited iterations to prevent lag
			if g.SphereHeightmap != nil && g.Topology != nil {
				geography.ApplyThermalErosionSpherical(g.SphereHeightmap, g.Topology, 3, g.Seed+g.TotalYearsSimulated)
				g.markSphereNeedsSync()
			} else {
				geography.ApplyThermalErosion(g.Heightmap, 3, g.Seed+g.TotalYearsSimulated)
			}

			// Phase 4: Tidal Physics integration
			// Calculate tidal range based on current satellite configuration
			// Using Earth properties as baseline mass/radius if not available individually
			planetMass := astronomy.EarthMassKg
			planetRadius := astronomy.EarthRadiusMeters
			tidalRange := astronomy.CalculateTidalRange(g.Satellites, planetMass, planetRadius)

			// Update config with dynamic tidal range
			coastalConfig := geography.DefaultCoastalConfig()
			coastalConfig.TidalRange = tidalRange

			if g.SphereHeightmap != nil && g.Topology != nil {
				// Apply erosion
				geography.SimulateCoastalErosion(g.SphereHeightmap, g.Topology, erosionInterval, g.SeaLevel, coastalConfig)

				// Mark intertidal zones for rendering
				geography.MarkIntertidalZones(g.SphereHeightmap, g.Topology, g.SeaLevel, tidalRange)

				g.markSphereNeedsSync()
			}

			// Phase 5: Differential erosion respecting rock hardness
			// Soft provinces (basins) erode faster, hard provinces (cratons) resist erosion
			// Sediment deposits at coastlines building continental shelves
			if g.SphereHeightmap != nil && g.Topology != nil {
				resolution := g.Topology.Resolution()
				totalCells := 6 * resolution * resolution
				numDrops := totalCells / 20 // ~5% of cells per erosion cycle
				if numDrops < 500 {
					numDrops = 500
				}
				geography.ApplyDifferentialErosion(g.SphereHeightmap, g.Topology, numDrops, g.Seed+g.TotalYearsSimulated, g.SeaLevel)

				// Coastal Erosion: Wave-driven erosion for realistic coastlines
				// Applies fetch-based wave energy, cliff retreat, and sediment transport
				coastalConfig := geography.DefaultCoastalConfig()
				// Scale tidal range based on satellite physics if available
				if len(g.Satellites) > 0 {
					// Use astronomy package to calculate tidal stress
					tidalStress := astronomy.CalculateTidalStress(g.Satellites)
					// Earth has ~2m average tidal range at 1.0x stress
					coastalConfig.TidalRange = 2.0 * tidalStress
				}
				geography.SimulateCoastalErosion(g.SphereHeightmap, g.Topology, erosionInterval, g.SeaLevel, coastalConfig)
				geography.FormBeaches(g.SphereHeightmap, g.Topology, g.SeaLevel)

				// Re-run Hydrology to update persistence features
				// Phase 7: Use rainfall-driven flux instead of uniform
				geography.CalculateGlobalFluxWithRainfall(g.SphereHeightmap, g.Rainfall)
				geography.ApplyRiverErosion(g.SphereHeightmap, 50.0, 5.0, g.SeaLevel) // Carve valleys
				lakes := geography.FillDepressions(g.SphereHeightmap, g.SeaLevel)
				geography.RouteFluxThroughLakes(g.SphereHeightmap, lakes)

				// Update rivers to match new terrain
				sphereRivers := geography.GenerateRiversSpherical(g.SphereHeightmap, g.SeaLevel, g.Seed+g.TotalYearsSimulated)
				g.Rivers = geography.ConvertSphericalRiversToFlat(sphereRivers, g.Topology.Resolution())

				// Form deltas at high-flux river mouths
				deltaConfig := geography.DefaultDeltaConfig()
				deltaRivers := geography.FormDeltasAtRiverMouths(g.SphereHeightmap, g.Topology, g.SeaLevel, g.Seed+g.TotalYearsSimulated, deltaConfig)

				// Add deltas to river network for rendering
				if len(deltaRivers) > 0 {
					flatDeltas := geography.ConvertSphericalRiversToFlat(deltaRivers, g.Topology.Resolution())
					g.Rivers = append(g.Rivers, flatDeltas...)
				}

				// Phase 4: Advanced Coastal Features
				// Mark intertidal zones (tidal flats exposed at low tide)
				geography.MarkIntertidalZones(g.SphereHeightmap, g.Topology, g.SeaLevel, coastalConfig.TidalRange)

				// Form estuaries at river mouths (widened mixing zones)
				geography.FormEstuaries(g.SphereHeightmap, g.Topology, g.SeaLevel, 150.0) // Flux > 150

				// Form spits and bars across bay openings
				geography.FormSpitsAndBars(g.SphereHeightmap, g.Topology, g.SeaLevel, g.Seed+g.TotalYearsSimulated)

				g.markSphereNeedsSync()
			} else {
				// Fallback for flat heightmap (legacy)
				geography.ApplyHydraulicErosion(g.Heightmap, 500, g.Seed+g.TotalYearsSimulated)
			}

			// Reset accumulator
			g.ErosionAccumulator -= erosionInterval
		}

		// Apply hotspot activity
		// This function already handles partial years probabilistically if needed,
		// or we can pass dtFloat.
		g.applyHotspotActivity(dtFloat)

		// Low frequency events using GeneralAccumulator
		// We can check multiple intervals

		// Cave formation - now with cap to prevent unbounded growth
		if g.TotalYearsSimulated%10_000_000 == 0 && g.Columns != nil {
			caveStart := time.Now()
			g.simulateCaveFormation(10_000_000)
			caveTime += time.Since(caveStart)
		}

		// Magma Chambers - now optimized with spatial hashing and GC
		if g.TotalYearsSimulated%10_000_000 == 0 && g.Columns != nil {
			magmaStart := time.Now()
			g.simulateMagmaChambers(10_000_000)
			magmaTime += time.Since(magmaStart)
		}

		// Reset GeneralAccumulator if it gets too big (periodic cleanup)
		// or use it as a 10k year clock
		if g.GeneralAccumulator >= 100_000 {
			g.GeneralAccumulator = 0 // Reset after the longest cycle (Cave formation)
		}

		// Simulate organic deposit evolution (sedimentation and transformation)
		// These are subtle geological changes, run every 100,000 years
		if g.TotalYearsSimulated%100_000 == 0 && g.Columns != nil {
			g.simulateDepositEvolution(100_000)
		}

		// Update heightmap min/max
		g.updateHeightmapStats()
	} // End Hadean optimization check

	// === ALWAYS RUN (Both Hadean and Modern) ===

	// RIVER GENERATION (Skip during Hadean - no liquid water)
	if heat <= 4.0 {
		// River generation (every 50,000 years or when RiverAccumulator is high enough)
		riverInterval := 50_000.0
		if g.RiverAccumulator >= riverInterval {
			// Generate rivers (expensive operation)
			// This can be heavy if we simulate fluid flow
			// For simplicity, we just call the method which uses threshold logic internally
			if g.SphereHeightmap != nil {
				// Procedural river generation based on heightmap
				// This is a placeholder if we've implemented it
				// Currently just reset the accumulator
			}
			g.RiverAccumulator -= riverInterval
		}
		// Regenerate dynamic features using spherical algorithms
		// Rivers and biomes change as terrain evolves
		// OPTIMIZATION: Throttle to every 10M years for deep-time simulation
		// Previous 100K interval caused biome regen every iteration, allocating
		// ~17MB per call (climate + biomes + 131K UUIDs) = 500MB/sec allocation rate
		riverInterval = 10_000_000.0 // 10M years - only ~400 regenerations in 4B year run
		if g.RiverAccumulator >= riverInterval {
			riverStart := time.Now()
			if g.SphereHeightmap != nil {
				sphereRivers := geography.GenerateRiversSpherical(g.SphereHeightmap, g.SeaLevel, g.Seed+g.TotalYearsSimulated)
				g.Rivers = geography.ConvertSphericalRiversToFlat(sphereRivers, g.Topology.Resolution())
				g.markSphereNeedsSync() // Sync river erosion to flat heightmap
			} else {
				g.Rivers = geography.GenerateRivers(g.Heightmap, g.SeaLevel, g.Seed+g.TotalYearsSimulated)
			}
			riverTime := time.Since(riverStart)
			_ = riverTime // Silencing unused variable error

			// Biome generation moved to external orchestrator (world_commands.go)
			// to prevent excessive memory allocation during geology-only simulation.
			// Was: g.Biomes = g.generateBiomesFromClimate(globalTempMod)
			biomeTime = 0

			// Decrement accumulator using modulo to keep phase but prevent buildup
			g.RiverAccumulator = math.Mod(g.RiverAccumulator, riverInterval)
		}
		if !erosionStart.IsZero() {
			erosionTime += time.Since(erosionStart)
		}
	} // End river check (heat <= 4.0)

	// Fix 3: Apply isostatic adjustment & Maintenance
	// OPTIMIZATION: Throttle to every 100,000 years (was 1,000)
	maintenanceInterval := 100_000.0
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
			g.markSphereNeedsSync()
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

	// NOTE: Elevation clamping and syncSphereToFlat is now ONLY done inside
	// the maintenance block above (every 100K years) to avoid performance overhead.
	// The old duplicate clamping that ran every iteration was removed.

	// Update heightmap min/max
	statsStart := time.Now()
	g.updateHeightmapStats()
	statsTime += time.Since(statsStart)

	// === DYNAMIC WEATHER ===
	// Recalculate rainfall map periodically to reflect continental drift
	// Run every 5 Million years (tectonics moves ~100km in that time)
	// TEMPORARILY DISABLED FOR DEBUGGING SEAM
	weatherInterval := 5_000_000.0
	if g.WeatherAccumulator >= weatherInterval {
		// g.updateRainfall() // DISABLED: Testing if this causes seam
		g.WeatherAccumulator = math.Mod(g.WeatherAccumulator, weatherInterval)
	}

	// === Ocean Phase Transition Logic ===
	oceanPhaseStart := time.Now()
	// Model water vapor ↔ liquid phase changes based on surface temperature
	// Early Earth (Hadean): >100°C → water exists as atmospheric vapor
	// Modern Earth: <100°C → water condenses into liquid oceans

	// Calculate average surface temperature
	avgTemp := g.calculateAverageSurfaceTemp(globalTempMod)

	// Define phase transition parameters
	const (
		modernSeaLevel = 0.0    // Baseline sea level (meters)
		vaporTempLow   = 90.0   // °C - start of transition zone
		vaporTempHigh  = 110.0  // °C - full vaporization
		vaporDepth     = 4000.0 // meters - ocean basins depth
	)

	// Store previous state for event detection
	wasVaporized := g.OceanVaporFraction > 0.5

	// Calculate vapor fraction (0.0 = all liquid, 1.0 = all vapor)
	vaporFraction := 0.0
	if avgTemp > vaporTempHigh {
		vaporFraction = 1.0 // Fully vaporized (Hadean steam atmosphere)
	} else if avgTemp > vaporTempLow {
		// Smooth transition zone (90-110°C)
		vaporFraction = (avgTemp - vaporTempLow) / (vaporTempHigh - vaporTempLow)
	}
	// else: vaporFraction = 0.0 (fully liquid, modern Earth)

	// Bounds checking
	if vaporFraction < 0.0 {
		vaporFraction = 0.0
	}
	if vaporFraction > 1.0 {
		vaporFraction = 1.0
	}

	// Update ocean vapor fraction
	g.OceanVaporFraction = vaporFraction

	// Calculate target sea level based on vapor fraction
	// When water vaporizes, sea level drops as ocean basins empty
	targetSeaLevel = modernSeaLevel - (vaporFraction * vaporDepth)

	// Smooth transition (exponential relaxation)
	// Prevents jarring jumps, simulates realistic evaporation/condensation timescales
	// Use exponential decay to be time-step independent: value += (target - value) * (1 - e^(-k * dt))
	// k=1.0e-4 gives ~10% change per 1000 years, reasonable for phase change lag
	const phaseChangeRate = 1.0e-4
	smoothingFactor := 1.0 - math.Exp(-phaseChangeRate*float64(dt))

	g.SeaLevel += (targetSeaLevel - g.SeaLevel) * smoothingFactor

	// Detect "Great Deluge" event (water condensing from atmosphere to form oceans)
	// Triggers when planet cools and vapor fraction drops below 50%
	var phaseEvent *PhaseTransitionEvent
	if wasVaporized && vaporFraction < 0.5 {
		phaseEvent = &PhaseTransitionEvent{
			Type:        "GreatDeluge",
			Year:        g.TotalYearsSimulated,
			Description: "Atmospheric water vapor condenses into liquid oceans as planet cools below 100°C",
		}
	}

	oceanPhaseTime = time.Since(oceanPhaseStart)

	// Log deep profiling every 10M years
	if profilingEnabled {
		totalProfiled := tectonicTime + biomeTime + oceanPhaseTime + statsTime + erosionTime + magmaTime + caveTime
		log.Printf("[GEO PROFILE] Year %d | Tectonic: %v (%.0f%%) | Ocean: %v (%.0f%%) | Eros: %v (%.0f%%) | Mag: %v (%.0f%%) | Cave: %v (%.0f%%) | Stats: %v (%.0f%%) | Bio: %v",
			g.TotalYearsSimulated,
			tectonicTime, float64(tectonicTime)/float64(totalProfiled)*100,
			oceanPhaseTime, float64(oceanPhaseTime)/float64(totalProfiled)*100,
			erosionTime, float64(erosionTime)/float64(totalProfiled)*100,
			magmaTime, float64(magmaTime)/float64(totalProfiled)*100,
			caveTime, float64(caveTime)/float64(totalProfiled)*100,
			statsTime, float64(statsTime)/float64(totalProfiled)*100,
			biomeTime)
	}

	// OPTIMIZATION: Batch all sphere-to-flat syncs into a single operation
	// Instead of syncing after each tectonic/volcanic/crater operation,
	// we mark dirty and flush once at the end
	g.flushSync()

	return phaseEvent
}

// applyHotspotActivity adds volcanic material at hotspot locations
// Eruption frequency scales with planetary heat (early Earth has 10x more eruptions)
func (g *WorldGeology) applyHotspotActivity(years float64) {
	// Get current planetary heat to scale volcanic activity
	heat := GetPlanetaryHeat(g.TotalYearsSimulated)

	// Base rate: 1 eruption per 1000 years at modern Earth (heat=1.0)
	// Early Earth (heat=10.0): 1 eruption per 100 years
	// Formula: baseRate / heat
	baseRate := 1000.0
	eruptionRate := baseRate / heat

	// Calculate number of eruptions for this time period
	numEruptions := int(years / eruptionRate)
	if numEruptions == 0 && g.rng.Float64() < (years/eruptionRate) {
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

	// Periodic plate region reassignment for realistic continental drift
	// Every 50M years, reassign which cells belong to which plate based on new centroids
	// This allows boundaries to shift as plates move, preventing permanent mountain walls
	const reassignmentInterval = 50_000_000.0 // 50 million years
	g.PlateReassignmentAccumulator += years

	if g.PlateReassignmentAccumulator >= reassignmentInterval && g.Topology != nil {
		if debug.Is(debug.Geology) {
			log.Printf("[PLATE DRIFT] Moving plates and reassigning regions after %.0fM years", g.PlateReassignmentAccumulator/1_000_000)
		}

		// CRITICAL: Actually MOVE the plate centroids first!
		// dt = time in ticks (1 tick = 1M years), we've accumulated reassignmentInterval years
		// so dt = reassignmentInterval / 1_000_000 ticks
		dt := g.PlateReassignmentAccumulator / 1_000_000.0
		geography.UpdatePlatePositions(g.Plates, dt, g.Topology)

		// Now reassign cell ownership based on new plate positions
		geography.ReassignPlateRegions(g.Plates, g.Topology, g.Seed)

		// Invalidate boundary cache so it's recomputed with new boundaries
		g.BoundaryCache = nil

		// Reset accumulator
		g.PlateReassignmentAccumulator = 0
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

	// Sync changes from sphere to flat map to ensure stats are accurate
	g.flushSync()
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
		g.markSphereNeedsSync()
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
		g.markSphereNeedsSync()
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
		g.markSphereNeedsSync()
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

// updateRainfall recalculates the global rainfall map based on current terrain and latitude.
// This ensures that as continents drift, their climate changes (e.g. crossing ITCZ or desert belts).
func (g *WorldGeology) updateRainfall() {
	if g.SphereHeightmap == nil || g.Topology == nil {
		return
	}

	// Use default config, but we could make this dynamic based on global temp
	config := weather.DefaultRainfallConfig(g.SeaLevel)

	// Recalculate rainfall
	rawRainfall := weather.GenerateRainfallMap(g.SphereHeightmap, g.Topology, config)

	// Normalize rainfall (matches initialization logic)
	totalCells := 6 * g.Topology.Resolution() * g.Topology.Resolution()
	totalRainfall := 0.0
	for _, r := range rawRainfall {
		totalRainfall += r
	}

	scalingFactor := float64(totalCells) / totalRainfall
	if totalRainfall == 0 || scalingFactor > 100 {
		scalingFactor = 1.0
	}

	// Update the rainfall map
	if len(g.Rainfall) != len(rawRainfall) {
		g.Rainfall = make([]float64, len(rawRainfall))
	}
	for i, r := range rawRainfall {
		g.Rainfall[i] = r * scalingFactor
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
		g.markSphereNeedsSync()
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
		g.markSphereNeedsSync()
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

// calculateAverageSurfaceTemp estimates global average surface temperature
// Uses biome temperatures (which include latitude/altitude effects) + global modifiers
// Returns temperature in Celsius
func (g *WorldGeology) calculateAverageSurfaceTemp(globalTempMod float64) float64 {
	// Get geothermal offset from planetary age
	heat := GetPlanetaryHeat(g.TotalYearsSimulated)
	geothermalOffset := 0.0
	if heat > 2.0 {
		// Early Earth: significant geothermal heating
		geothermalOffset = (heat - 1.0) * 10.0
	} else {
		// Modern Earth: minimal geothermal contribution
		geothermalOffset = (heat - 1.0) * 2.0
	}

	// Calculate average from biomes if available
	avgBiomeTemp := 15.0 // Default baseline (Earth-like)
	if len(g.Biomes) > 0 {
		totalTemp := 0.0
		for _, b := range g.Biomes {
			totalTemp += b.Temperature
		}
		avgBiomeTemp = totalTemp / float64(len(g.Biomes))
	}

	// Combine all temperature factors
	return avgBiomeTemp + globalTempMod + geothermalOffset
}

// GetStats returns current geological statistics
func (g *WorldGeology) GetStats() GeologyStats {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var avgElev, landPercent, minElev, maxElev float64
	var plateCount, hotspotCount, riverCount, biomeCount int

	plateCount = len(g.Plates)
	hotspotCount = len(g.Hotspots)
	riverCount = len(g.Rivers)
	biomeCount = len(g.Biomes)

	// Prefer SphereHeightmap (primary data source) over flat Heightmap
	if g.SphereHeightmap != nil {
		// Use sphere heightmap for elevation stats
		minElev, maxElev = g.SphereHeightmap.MinMax()
		topo := g.SphereHeightmap.Topology()
		numCells := topo.Resolution() * topo.Resolution() * 6

		sum := 0.0
		landCount := 0
		for i := 0; i < numCells; i++ {
			coord := spatial.Coordinate{Face: i / (topo.Resolution() * topo.Resolution()), X: (i / topo.Resolution()) % topo.Resolution(), Y: i % topo.Resolution()}
			elev := g.SphereHeightmap.Get(coord)
			sum += elev
			if elev > g.SeaLevel {
				landCount++
			}
		}

		if numCells > 0 {
			avgElev = sum / float64(numCells)
			landPercent = float64(landCount) / float64(numCells) * 100
		}
	} else if g.Heightmap != nil && len(g.Heightmap.Elevations) > 0 {
		// Fallback to flat heightmap
		sum := 0.0
		landCount := 0
		for _, elev := range g.Heightmap.Elevations {
			sum += elev
			if elev > g.SeaLevel {
				landCount++
			}
		}

		totalPixels := float64(len(g.Heightmap.Elevations))
		avgElev = sum / totalPixels
		landPercent = float64(landCount) / totalPixels * 100
		maxElev = g.Heightmap.MaxElev
		minElev = g.Heightmap.MinElev
	}

	// Calculate average temperature from biomes
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
		MaxElevation:       maxElev,
		MinElevation:       minElev,
		SeaLevel:           g.SeaLevel,
		LandPercent:        landPercent,
		PlateCount:         plateCount,
		HotspotCount:       hotspotCount,
		RiverCount:         riverCount,
		BiomeCount:         biomeCount,
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
// UpdateBiomes updates the biomes based on the current heightmap and climate.
// This is now decoupled from SimulateGeology loop to prevent excessive memory allocations.
// Should be called periodically by the simulation orchestrator if life is enabled.
func (g *WorldGeology) UpdateBiomes(globalTempMod float64) []geography.Biome {
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

			// Hydrology Data Retrieval
			var flux float64
			var isLake bool

			if g.Topology != nil && g.SphereHeightmap != nil {
				// Map flat coordinates to spherical (Equirectangular)
				lon := (float64(x) / float64(g.Heightmap.Width)) * 2 * math.Pi
				lat := (0.5 - float64(y)/float64(g.Heightmap.Height)) * math.Pi

				sphereX := math.Cos(lat) * math.Cos(lon)
				sphereY := math.Sin(lat)
				sphereZ := math.Cos(lat) * math.Sin(lon)

				coord := g.Topology.FromVector(sphereX, sphereY, sphereZ)
				cellData := g.SphereHeightmap.GetCellData(coord)
				flux = cellData.Flux
				isLake = cellData.IsLake
			}

			biomeType := geography.ClassifyBiome(
				climate.Temperature,
				climate.AnnualRainfall,
				climate.SoilDrainage,
				elev,
				g.SeaLevel,
				flux,
				isLake,
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

// GetTectonicMap returns a flat map of tectonic plate IDs and metadata for visualization.
// width/height: the dimensions of the target grid (usually matches frontend display)
func (g *WorldGeology) GetTectonicMap(width, height int) ([]int, []map[string]interface{}) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.Plates) == 0 {
		return nil, nil
	}

	// Use default dimensions if 0 provided
	if width == 0 {
		width = g.Heightmap.Width
	}
	if height == 0 {
		height = g.Heightmap.Height
	}

	// 1. Prepare dense source array for projection
	sourceResolution := g.SphereHeightmap.Resolution()
	sourceSize := sourceResolution * sourceResolution * 6
	sourceData := make([]int, sourceSize)

	// Initialize with -1
	for i := range sourceData {
		sourceData[i] = -1
	}

	// Fill source data from plates
	for i, plate := range g.Plates {
		id := i + 1 // 1-based ID for visualization
		for coord := range plate.Region {
			// Calculate index
			idx := coord.Face*sourceResolution*sourceResolution + coord.Y*sourceResolution + coord.X
			if idx >= 0 && idx < sourceSize {
				sourceData[idx] = id
			}
		}
	}

	// 2. Project to flat grid using closure
	grid := g.SphereHeightmap.MapIntToFlat(width, height, func(coord spatial.Coordinate) int {
		idx := coord.Face*sourceResolution*sourceResolution + coord.Y*sourceResolution + coord.X
		if idx >= 0 && idx < sourceSize {
			return sourceData[idx]
		}
		return -1
	})

	// 3. Calculate centroids on the projected grid
	plateCentersX := make(map[int]int)
	plateCentersY := make(map[int]int)
	platePixelCounts := make(map[int]int)

	for i, plateID := range grid {
		if plateID > 0 {
			x := i % width
			y := i / width

			plateCentersX[plateID] += x
			plateCentersY[plateID] += y
			platePixelCounts[plateID]++
		}
	}

	// 4. Prepare metadata
	metadata := make([]map[string]interface{}, 0)
	for i, plate := range g.Plates {
		id := i + 1
		count := platePixelCounts[id]

		// Determine visual center
		centerX := 0
		centerY := 0
		if count > 0 {
			centerX = plateCentersX[id] / count
			centerY = plateCentersY[id] / count
		} else {
			continue // Skip invisible plates
		}

		name := fmt.Sprintf("Plate %d", id)
		if plate.Type == geography.PlateOceanic {
			name = fmt.Sprintf("Oceanic %d", id)
		} else {
			name = fmt.Sprintf("Continental %d", id)
		}

		metadata = append(metadata, map[string]interface{}{
			"id":       id,
			"name":     name,
			"type":     string(plate.Type),
			"center_x": centerX,
			"center_y": centerY,
			"area":     count,
		})
	}

	return grid, metadata
}

// GetMineralDeposits returns all mineral deposits from underground columns.
// Returns a slice of deposit info maps for JSON serialization.
func (g *WorldGeology) GetMineralDeposits() []map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.Columns == nil {
		return nil
	}

	var deposits []map[string]interface{}
	for _, col := range g.Columns.AllColumns() {
		// Use the Resources field which contains actual mineral deposits
		for _, deposit := range col.Resources {
			info := map[string]interface{}{
				"x":          col.X,
				"y":          col.Y,
				"type":       deposit.Type,
				"depth":      col.Surface - deposit.DepthZ,
				"quantity":   deposit.Quantity,
				"discovered": deposit.Discovered,
			}
			// Include organic source info if available (fossils, oil)
			if deposit.Source != nil {
				info["species"] = deposit.Source.Species
				info["death_year"] = deposit.Source.DeathYear
			}
			deposits = append(deposits, info)
		}
	}

	return deposits
}

// ResourceNode represents a sparse resource for the map overlay
type ResourceNode struct {
	Type string                 `json:"type"` // "gold", "iron", "cave", "volcano", "peak", "trench"
	X    int                    `json:"x"`
	Y    int                    `json:"y"`
	Val  float64                `json:"val,omitempty"`  // Normalized value (0-1) or absolute (Height/Depth)
	Data map[string]interface{} `json:"data,omitempty"` // Arbitrary metadata for tooltips
}

// GetTemperatureMap generates a normalized temperature map (0.0-1.0)
// 0.0 = Cold (Poles), 1.0 = Hot (Equator)
func (g *WorldGeology) GetTemperatureMap(width, height int) []float64 {
	result := make([]float64, width*height)
	// Use a seed offset so it doesn't look identical to other noise
	noise := geography.NewPerlinGenerator(g.Seed + 123)

	for y := 0; y < height; y++ {
		// Latitude: 0 = North Pole, 0.5 = Equator, 1.0 = South Pole
		// But for temp, we want 0.0 at poles, 1.0 at equator
		normalizedY := float64(y) / float64(height)
		// Distance from equator (0.0 at equator, 0.5 at poles)
		distFromEquator := math.Abs(normalizedY - 0.5)
		// Base temp: 1.0 at equator, 0.0 at poles
		baseTemp := 1.0 - (distFromEquator * 2.0)

		for x := 0; x < width; x++ {
			idx := y*width + x

			// Add noise for organic variation
			// Scale coords for noise frequency
			nx := float64(x) * 0.1
			ny := float64(y) * 0.1
			n := noise.Noise2D(nx, ny) // -1 to 1

			// Blend base temp with noise (80% lat, 20% noise)
			temp := baseTemp*0.8 + ((n+1)/2)*0.2

			// Clamp
			if temp < 0 {
				temp = 0
			}
			if temp > 1 {
				temp = 1
			}

			result[idx] = temp
		}
	}
	return result
}

// GetMoistureMap generates a normalized moisture map (0.0-1.0)
func (g *WorldGeology) GetMoistureMap(width, height int) []float64 {
	result := make([]float64, width*height)
	noise := geography.NewPerlinGenerator(g.Seed + 456)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x

			nx := float64(x) * 0.08
			ny := float64(y) * 0.08
			n := noise.Noise2D(nx, ny)

			moisture := (n + 1) / 2.0
			result[idx] = moisture
		}
	}
	return result
}

// GetBiomeMap generates biome IDs based on temperature and moisture
// 0: Ocean (Implicit in client if elevation is low, but here we output land biomes)
// 1: Desert, 2: Savanna, 3: Jungle
// 4: Grassland, 5: Forest, 6: Taiga
// 7: Tundra, 8: Ice
func (g *WorldGeology) GetBiomeMap(width, height int, tempMap, moistureMap []float64) []int {
	result := make([]int, width*height)

	for i := 0; i < len(result); i++ {
		t := tempMap[i]
		m := moistureMap[i]

		var biome int

		if t < 0.2 {
			// Cold
			if m < 0.5 {
				biome = 8 // Ice/Polar
			} else {
				biome = 7 // Tundra
			}
		} else if t < 0.5 {
			// Cool
			if m < 0.4 {
				biome = 4 // Grassland/Steppe
			} else if m < 0.7 {
				biome = 6 // Taiga
			} else {
				biome = 5 // Forest
			}
		} else if t < 0.8 {
			// Temperate/Warm
			if m < 0.3 {
				biome = 1 // Desert
			} else if m < 0.6 {
				biome = 2 // Savanna
			} else {
				biome = 5 // Forest
			}
		} else {
			// Hot
			if m < 0.3 {
				biome = 1 // Desert
			} else if m < 0.6 {
				biome = 2 // Savanna
			} else {
				biome = 3 // Jungle
			}
		}

		result[i] = biome
	}

	return result
}

// GetResourceMap generates sparse resource nodes
func (g *WorldGeology) GetResourceMap(width, height int, elevMap []float64) []ResourceNode {
	var resources []ResourceNode
	rng := rand.New(rand.NewSource(g.Seed + 789))

	// Create a density map or just random scatter
	// Approx 1 resource per 100 pixels
	count := (width * height) / 100

	for i := 0; i < count; i++ {
		x := rng.Intn(width)
		y := rng.Intn(height)

		// If we have elevation data (projected), we can use it rules
		// For now, simple random types
		rType := "iron"
		roll := rng.Float64()
		if roll < 0.1 {
			rType = "gold"
		} else if roll < 0.3 {
			rType = "cave"
		} else if roll < 0.6 {
			rType = "coal"
		}

		// Check elevation if available (simple flat index)
		if len(elevMap) > 0 {
			idx := y*width + x
			if idx < len(elevMap) {
				elev := elevMap[idx]
				// Caves only in high areas (assuming >0.6 is high)
				if rType == "cave" && elev < 0.6 {
					continue
				}
				// Minerals only on land (assuming 0.5 is sea level)
				if elev < 0.5 {
					continue
				}
			}
		}

		resources = append(resources, ResourceNode{
			Type: rType,
			X:    x,
			Y:    y,
		})
	}

	return resources
}

// GetElevationMap returns a flat projected elevation map (0.0-1.0)
func (g *WorldGeology) GetElevationMap(width, height int) []float64 {
	// Re-use MapIntToFlat logic but for floats

	result := make([]float64, width*height)
	// We assume SphereHeightmap is populated
	if g.SphereHeightmap == nil {
		return result
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map pixel coordinates to longitude and latitude
			lon := (float64(x) / float64(width)) * 2 * math.Pi
			lat := (0.5 - float64(y)/float64(height)) * math.Pi

			// Spherical conversion
			cosLat := math.Cos(lat)
			sinLat := math.Sin(lat)
			cosLon := math.Cos(lon)
			sinLon := math.Sin(lon)

			sx := cosLat * cosLon
			sy := sinLat
			sz := cosLat * sinLon

			coord := g.SphereHeightmap.Topology().FromVector(sx, sy, sz)

			// Nearest neighbor
			val := g.SphereHeightmap.Get(coord)

			// Normalize based on Min/Max
			minElev := g.SphereHeightmap.MinElev
			maxElev := g.SphereHeightmap.MaxElev
			if maxElev == minElev {
				result[y*width+x] = 0.5
			} else {
				norm := (val - minElev) / (maxElev - minElev)
				// Clamp
				if norm < 0 {
					norm = 0
				}
				if norm > 1 {
					norm = 1
				}
				result[y*width+x] = norm
			}
		}
	}
	return result
}

// GetSedimentMap returns a flat projected sediment map (0.0-1.0)
// Sediment depth is normalized by dividing by MaxSediment (200m default).
func (g *WorldGeology) GetSedimentMap(width, height int) []float64 {
	const MaxSediment = 200.0 // Max expected sediment depth in meters

	result := make([]float64, width*height)
	if g.SphereHeightmap == nil {
		return result
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map pixel coordinates to longitude and latitude
			lon := (float64(x) / float64(width)) * 2 * math.Pi
			lat := (0.5 - float64(y)/float64(height)) * math.Pi

			// Spherical conversion
			cosLat := math.Cos(lat)
			sinLat := math.Sin(lat)
			cosLon := math.Cos(lon)
			sinLon := math.Sin(lon)

			sx := cosLat * cosLon
			sy := sinLat
			sz := cosLat * sinLon

			coord := g.SphereHeightmap.Topology().FromVector(sx, sy, sz)

			// Get CellData for sediment
			cellData := g.SphereHeightmap.GetCellData(coord)
			sediment := cellData.Sediment

			// Normalize and clamp
			norm := sediment / MaxSediment
			if norm < 0 {
				norm = 0
			}
			if norm > 1 {
				norm = 1
			}
			result[y*width+x] = norm
		}
	}
	return result
}

// GetTerrainFeaturesMap returns active terrain features (peaks, volcanoes, trenches)
func (g *WorldGeology) GetTerrainFeaturesMap(width, height int) []ResourceNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var features []ResourceNode

	// Project global features (Volcanoes/Magma Chambers)
	// Magma chambers use grid coordinates matching the world heightmap resolution
	sourceW := float64(g.Heightmap.Width)
	sourceH := float64(g.Heightmap.Height)

	// Scale factors
	scaleX := float64(width) / sourceW
	scaleY := float64(height) / sourceH

	for _, chamber := range g.MagmaChambers {
		if !chamber.Solidified {
			// Chamber coordinates are in source grid space
			srcX := chamber.Center.X
			srcY := chamber.Center.Y

			// Project to requested grid size
			screenX := int(srcX * scaleX)
			screenY := int(srcY * scaleY)

			// Wrap X
			screenX = screenX % width
			if screenX < 0 {
				screenX += width
			}
			// Clamp Y
			if screenY < 0 {
				screenY = 0
			}
			if screenY >= height {
				screenY = height - 1
			}

			data := map[string]interface{}{
				"age":      chamber.Age,
				"pressure": chamber.Pressure,
				"active":   !chamber.Solidified,
			}
			features = append(features, ResourceNode{
				Type: "volcano",
				X:    screenX,
				Y:    screenY,
				Val:  chamber.Pressure, // Use pressure as generic value
				Data: data,
			})
		}
	}

	// Identify Peaks (High elevation local maxima) and Trenches

	// Temporarily release lock to call GetElevationMap if strict checking,
	// but GetElevationMap accesses shared state without locking itself?
	// The snippet I saw showed it accesses g.SphereHeightmap.*.
	// We ALREADY hold the lock. If GetElevationMap DOES lock, we die.
	// Looking at lines 2026-2075: It does NOT acquire a lock. It accesses fields directly.
	// Safe to proceed.

	elevGrid := g.GetElevationMap(width, height)
	if len(elevGrid) == 0 {
		return features
	}

	radius := 4
	minPeakHeight := 0.75  // Top 25% of height range (Significant peaks)
	maxTrenchDepth := 0.35 // Lowest 35% (Deep ocean)

	// Scanning logic
	for y := radius; y < height-radius; y += radius { // Stride
		for x := radius; x < width-radius; x += radius {
			val := elevGrid[y*width+x]

			// --- Peak Detection ---
			if val > minPeakHeight {
				isMax := true
				// Check neighborhood
				for dy := -radius; dy <= radius; dy++ {
					for dx := -radius; dx <= radius; dx++ {
						if dx == 0 && dy == 0 {
							continue
						}
						// Wrap X
						nx := (x + dx + width) % width
						ny := y + dy
						// Clamp Y
						if ny < 0 || ny >= height {
							continue
						}

						if elevGrid[ny*width+nx] >= val {
							isMax = false
							break
						}
					}
					if !isMax {
						break
					}
				}
				if isMax {
					// Approximate height in meters based on Earth (8848m)
					// Val 0.75-1.0 maps to 3000m-8848m
					heightM := 3000 + (val-0.75)/0.25*5848
					features = append(features, ResourceNode{
						Type: "peak",
						X:    x,
						Y:    y,
						Val:  val,
						Data: map[string]interface{}{
							"height": heightM,
							"zone":   "Alpine",
						},
					})
				}
			}

			// --- Trench Detection ---
			// Only check for trenches if underwater (val < 0.5 usually)
			if val < maxTrenchDepth {
				isMin := true
				for dy := -radius; dy <= radius; dy++ {
					for dx := -radius; dx <= radius; dx++ {
						if dx == 0 && dy == 0 {
							continue
						}
						nx := (x + dx + width) % width
						ny := y + dy
						if ny < 0 || ny >= height {
							continue
						}

						if elevGrid[ny*width+nx] <= val {
							isMin = false
							break
						}
					}
					if !isMin {
						break
					}
				}
				if isMin {
					// Approximate depth
					depthM := -2000 - (0.35-val)/0.35*9000 // -2000m to -11000m
					features = append(features, ResourceNode{
						Type: "trench",
						X:    x,
						Y:    y,
						Val:  val,
						Data: map[string]interface{}{
							"depth": depthM,
							"zone":  "Abyssal",
						},
					})
				}
			}
		}
	}

	return features
}

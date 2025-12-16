package ecosystem

import (
	"math"
	"math/rand"
	"sync"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// WorldGeology manages terrain evolution over geological time
type WorldGeology struct {
	mu sync.RWMutex

	WorldID       uuid.UUID
	Seed          int64
	Circumference float64 // meters

	// Core geographic data
	Heightmap *geography.Heightmap
	Plates    []geography.TectonicPlate
	SeaLevel  float64 // meters (0 = baseline, positive = higher sea level)

	// Dynamic geographic features
	Hotspots []geography.Point // Fixed mantle plume locations
	Rivers   [][]geography.Point
	Biomes   []geography.Biome

	// Simulation state
	TotalYearsSimulated int64
	rng                 *rand.Rand

	// Scale factors (pixels to real-world)
	PixelsPerKm float64 // How many heightmap pixels per real km
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
func NewWorldGeology(worldID uuid.UUID, seed int64, circumferenceMeters float64) *WorldGeology {
	return &WorldGeology{
		WorldID:       worldID,
		Seed:          seed,
		Circumference: circumferenceMeters,
		SeaLevel:      0, // Baseline sea level
		rng:           rand.New(rand.NewSource(seed)),
	}
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

	// Generate tectonic plates
	// Plate count scales with world size (more surface = more plates)
	plateCount := 6 + g.rng.Intn(4) // 6-9 plates for variety
	g.Plates = geography.GeneratePlates(plateCount, width, height, g.Seed)

	// Generate initial heightmap using existing worldgen
	// Default erosion rate 1.0, rainfall 1.0 for balanced terrain
	g.Heightmap = geography.GenerateHeightmap(width, height, g.Plates, g.Seed, 1.0, 1.0)

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

	// Generate initial rivers
	g.Rivers = geography.GenerateRivers(g.Heightmap, g.SeaLevel, g.Seed)

	// Initialize biomes with default temp (0 offset)
	g.Biomes = geography.AssignBiomes(g.Heightmap, g.SeaLevel, g.Seed, 0.0)
}

// SimulateGeology advances geological processes over time
// yearsElapsed is the number of years to simulate
// globalTempMod is the current global temperature offset (e.g. from volcanic winter)
func (g *WorldGeology) SimulateGeology(yearsElapsed int64, globalTempMod float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Heightmap == nil {
		return // Not initialized
	}

	g.TotalYearsSimulated += yearsElapsed

	// Plate movement: ~2cm/year = 0.00002 km/year
	// Over 1 million years = 20 km of movement
	// We accumulate movement and apply tectonic effects periodically

	// Apply plate tectonics every 100,000 years for efficiency
	tectonicInterval := int64(100_000)
	if yearsElapsed >= tectonicInterval {
		intervals := yearsElapsed / tectonicInterval
		for i := int64(0); i < intervals; i++ {
			g.advancePlates(float64(tectonicInterval))
		}
	}

	// Apply erosion (more frequent)
	// Thermal erosion: 1 iteration per 10,000 years
	thermalIterations := int(yearsElapsed / 10_000)
	if thermalIterations > 0 && thermalIterations <= 10 { // Cap iterations
		geography.ApplyThermalErosion(g.Heightmap, thermalIterations, g.Seed+g.TotalYearsSimulated)
	}

	// Hydraulic erosion: proportional to time but capped
	// 1000 drops per 10,000 years
	drops := int((yearsElapsed * 1000) / 10_000)
	if drops > 0 && drops <= 5000 {
		geography.ApplyHydraulicErosion(g.Heightmap, drops, g.Seed+g.TotalYearsSimulated)
	}

	// Apply hotspot activity
	g.applyHotspotActivity(float64(yearsElapsed))

	// Regenerate dynamic features
	// Rivers and biomes change as terrain evolves
	g.Rivers = geography.GenerateRivers(g.Heightmap, g.SeaLevel, g.Seed+g.TotalYearsSimulated)
	// Pass global temperature modifier to biome assignment
	g.Biomes = geography.AssignBiomes(g.Heightmap, g.SeaLevel, g.Seed+g.TotalYearsSimulated, globalTempMod)

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

			geography.ApplyVolcano(g.Heightmap, jx, jy, radius, height)
		}
	}
}

// advancePlates moves tectonic plates and recalculates boundaries
func (g *WorldGeology) advancePlates(years float64) {
	// Movement rate: 2cm/year = 0.00002 km/year
	movementRate := 0.00002 * years // km

	// Convert to pixels
	pixelMovement := movementRate * g.PixelsPerKm

	// Move plate centroids
	for i := range g.Plates {
		g.Plates[i].Centroid.X += g.Plates[i].MovementVector.X * pixelMovement
		g.Plates[i].Centroid.Y += g.Plates[i].MovementVector.Y * pixelMovement

		// Wrap around (toroidal topology)
		if g.Plates[i].Centroid.X < 0 {
			g.Plates[i].Centroid.X += float64(g.Heightmap.Width)
		}
		if g.Plates[i].Centroid.X >= float64(g.Heightmap.Width) {
			g.Plates[i].Centroid.X -= float64(g.Heightmap.Width)
		}
		// Y doesn't wrap, clamp instead
		if g.Plates[i].Centroid.Y < 0 {
			g.Plates[i].Centroid.Y = 0
		}
		if g.Plates[i].Centroid.Y >= float64(g.Heightmap.Height) {
			g.Plates[i].Centroid.Y = float64(g.Heightmap.Height - 1)
		}

		// Age plates
		g.Plates[i].Age += years / 1_000_000 // Age in million years
	}

	// Recalculate tectonic effects on boundaries
	tectonicMods := geography.SimulateTectonics(g.Plates, g.Heightmap.Width, g.Heightmap.Height)

	// Apply a small fraction of tectonic modification (gradual buildup)
	scaleFactor := 0.01 // Only 1% per interval for gradual change
	for i := range g.Heightmap.Elevations {
		g.Heightmap.Elevations[i] += tectonicMods.Elevations[i] * scaleFactor
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

	for i := 0; i < numVolcanoes; i++ {
		// Random location, preferring plate boundaries
		x := float64(g.rng.Intn(g.Heightmap.Width))
		y := float64(g.rng.Intn(g.Heightmap.Height))

		// Volcano height based on severity (2000-5000m)
		height := 2000 + severity*3000
		radius := 2.0 + g.rng.Float64()*2.0

		geography.ApplyVolcano(g.Heightmap, x, y, radius, height)
	}
}

// applyImpactCrater creates a crater from asteroid impact
func (g *WorldGeology) applyImpactCrater(severity float64) {
	// Random impact location
	centerX := g.rng.Intn(g.Heightmap.Width)
	centerY := g.rng.Intn(g.Heightmap.Height)

	// Crater size based on severity (10-50 pixel radius)
	radius := int(10 + severity*40)

	// Depth based on severity (500-3000m)
	depth := 500 + severity*2500

	// Rim height (15% of depth)
	rimHeight := depth * 0.15

	// Apply crater depression with raised rim
	for dy := -radius * 2; dy <= radius*2; dy++ {
		for dx := -radius * 2; dx <= radius*2; dx++ {
			px, py := centerX+dx, centerY+dy
			if px >= 0 && px < g.Heightmap.Width && py >= 0 && py < g.Heightmap.Height {
				dist := math.Sqrt(float64(dx*dx + dy*dy))

				if dist < float64(radius) {
					// Inside crater - depression
					factor := 1.0 - (dist / float64(radius))
					current := g.Heightmap.Get(px, py)
					g.Heightmap.Set(px, py, current-depth*factor*factor)

				} else if dist < float64(radius)*1.3 {
					// Crater rim - raised
					t := (dist - float64(radius)) / (float64(radius) * 0.3)
					factor := 1.0 - t
					current := g.Heightmap.Get(px, py)
					g.Heightmap.Set(px, py, current+rimHeight*factor)
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
	// Find high-elevation pixels and erode downward
	threshold := g.Heightmap.MaxElev * 0.6 // Top 40% of elevation

	for y := 0; y < g.Heightmap.Height; y++ {
		for x := 0; x < g.Heightmap.Width; x++ {
			elev := g.Heightmap.Get(x, y)
			if elev > threshold {
				// Glacial carving - erode proportionally
				erosion := (elev - threshold) * 0.1 * severity
				g.Heightmap.Set(x, y, elev-erosion)
			}
		}
	}
}

// applyContinentalDrift accelerates plate movement and creates mountains
func (g *WorldGeology) applyContinentalDrift(severity float64) {
	// Enhanced plate movement
	extraYears := 50_000 + int64(severity*100_000)
	g.advancePlates(float64(extraYears))

	// Additional mountain building at convergent boundaries
	// Recalculate tectonics with higher effect
	tectonicMods := geography.SimulateTectonics(g.Plates, g.Heightmap.Width, g.Heightmap.Height)

	scaleFactor := 0.05 * severity // 5% per event, scaled by severity
	for i := range g.Heightmap.Elevations {
		g.Heightmap.Elevations[i] += tectonicMods.Elevations[i] * scaleFactor
	}
}

// applyFloodBasalt creates large volcanic provinces
func (g *WorldGeology) applyFloodBasalt(severity float64) {
	// Large area volcanic activity
	centerX := g.rng.Intn(g.Heightmap.Width)
	centerY := g.rng.Intn(g.Heightmap.Height)

	// Radius based on severity (30-100 pixels)
	radius := 30 + int(severity*70)

	// Height of basalt layers (500-2000m)
	height := 500 + severity*1500

	// Apply basalt layers with gradual falloff
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			px, py := centerX+dx, centerY+dy
			if px >= 0 && px < g.Heightmap.Width && py >= 0 && py < g.Heightmap.Height {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist < float64(radius) {
					factor := 1.0 - (dist / float64(radius))
					factor = factor * factor // Smoother falloff

					current := g.Heightmap.Get(px, py)
					g.Heightmap.Set(px, py, current+height*factor)
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

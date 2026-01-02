package geography

import (
	"math"

	"tw-backend/internal/spatial"
)

// =============================================================================
// Phase 10: Glacial Geomorphology
// =============================================================================

// IceData represents ice sheet properties for a single grid cell.
type IceData struct {
	Thickness float64          // Ice thickness in meters (0 = no ice)
	FlowDir   spatial.Vector3D // Flow direction (normalized tangent vector)
	FlowSpeed float64          // Flow speed in m/year
	Age       float64          // Ice age in years (for tracking)
}

// IceSheet represents a continental or regional ice sheet system.
type IceSheet struct {
	// Grid of ice data indexed by coordinate
	Ice map[spatial.Coordinate]*IceData

	// Aggregate statistics
	TotalVolume  float64 // km³ of ice
	TotalArea    float64 // km² covered
	MaxThickness float64 // meters

	// Sediment tracking for moraines
	Sediment map[spatial.Coordinate]float64 // Accumulated sediment load
}

// NewIceSheet creates an empty ice sheet system.
func NewIceSheet() *IceSheet {
	return &IceSheet{
		Ice:      make(map[spatial.Coordinate]*IceData),
		Sediment: make(map[spatial.Coordinate]float64),
	}
}

// Constants for ice physics
const (
	// IceAccumulationRate is m/year per precipitation unit (simplified)
	IceAccumulationRate = 0.5

	// IceAblationRate is m/year per degree above 0°C
	IceAblationRate = 2.0

	// IceDensity in kg/m³
	IceDensity = 917.0

	// IceFlowConstant (simplified SIA) - relates thickness to flow speed
	// v ~ A * τ^n where τ ~ ρgh * slope, n=3 for ice
	// Simplified: v = C * H^3 * slope
	IceFlowConstant = 1e-16

	// MinIceThickness below which ice disappears
	MinIceThickness = 1.0

	// ErosionCoefficient relates ice flux to bedrock erosion (mm/year per unit flux)
	ErosionCoefficient = 0.001

	// SedimentCapacity is max sediment load per unit ice thickness (arbitrary units)
	SedimentCapacity = 0.1
)

// Update processes ice accumulation, flow, and ablation for one time step.
// dt is in years, temperature is global average, precipitation is mm/year equivalent.
// topology is used for neighbor lookups, heightmap provides bedrock elevation.
func (is *IceSheet) Update(dt float64, tempGrid []float64, precipGrid []float64,
	heightmap *SphereHeightmap, topology spatial.Topology) {

	resolution := topology.Resolution()
	totalCells := 6 * resolution * resolution

	// Phase 1: Accumulation (where cold + precipitation)
	for idx := 0; idx < totalCells; idx++ {
		coord := iceIndexToCoord(idx, resolution)
		temp := tempGrid[idx]
		precip := precipGrid[idx]

		// Ice accumulates where temp < 0 and precipitation > 0
		if temp < 0 && precip > 0 {
			accumulation := precip * IceAccumulationRate * dt

			ice, exists := is.Ice[coord]
			if !exists {
				ice = &IceData{}
				is.Ice[coord] = ice
			}
			ice.Thickness += accumulation
		}
	}

	// Phase 2: Flow (SIA - ice flows downhill from high to low pressure)
	// Pressure ~ ice thickness, flows toward lower elevation + thinner ice
	newIce := make(map[spatial.Coordinate]*IceData)

	for coord, ice := range is.Ice {
		if ice.Thickness < MinIceThickness {
			continue
		}

		// Calculate gradient (slope) - ice flows toward lowest neighbor
		baseElev := heightmap.Get(coord) + ice.Thickness
		lowestNeighbor := coord
		lowestElev := baseElev

		for _, dir := range []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West} {
			neighbor := topology.GetNeighbor(coord, dir)
			neighborIce := 0.0
			if ni, ok := is.Ice[neighbor]; ok {
				neighborIce = ni.Thickness
			}
			neighborElev := heightmap.Get(neighbor) + neighborIce
			if neighborElev < lowestElev {
				lowestElev = neighborElev
				lowestNeighbor = neighbor
			}
		}

		// Calculate flow speed using simplified SIA
		slope := (baseElev - lowestElev) / 100000.0 // Assume 100km grid spacing
		if slope < 0 {
			slope = 0
		}

		// v = C * H^3 * slope (Glen's flow law)
		flowSpeed := IceFlowConstant * math.Pow(ice.Thickness, 3) * slope * 1e12 // Scale to m/year

		// Limit flow to available ice
		flowFrac := math.Min(flowSpeed*dt/ice.Thickness, 0.5) // Max 50% flows out per step

		// Transfer ice to neighbor
		iceTransfer := ice.Thickness * flowFrac

		// Keep remaining ice at current cell
		remainingThickness := ice.Thickness - iceTransfer
		if remainingThickness >= MinIceThickness {
			if _, exists := newIce[coord]; !exists {
				newIce[coord] = &IceData{}
			}
			newIce[coord].Thickness += remainingThickness
		}

		// Add transferred ice to neighbor
		if iceTransfer > 0 && lowestNeighbor != coord {
			if _, exists := newIce[lowestNeighbor]; !exists {
				newIce[lowestNeighbor] = &IceData{}
			}
			newIce[lowestNeighbor].Thickness += iceTransfer

			// Store flow direction for erosion calculation
			newIce[lowestNeighbor].FlowSpeed = flowSpeed
		}
	}

	// Phase 3: Ablation (melting at warm edges)
	for coord, ice := range newIce {
		idx := iceCoordToIndex(coord, resolution)
		if idx < 0 || idx >= len(tempGrid) {
			continue
		}
		temp := tempGrid[idx]

		if temp > 0 {
			meltRate := temp * IceAblationRate * dt
			ice.Thickness -= meltRate
			if ice.Thickness < 0 {
				ice.Thickness = 0
			}
		}
	}

	// Update the ice map
	is.Ice = newIce

	// Update stats
	is.updateStats(resolution)
}

// ApplyErosion applies glacial erosion to the heightmap based on ice flow.
// Returns total erosion volume for tracking.
func (is *IceSheet) ApplyErosion(heightmap *SphereHeightmap, dt float64, resolution int) float64 {
	totalErosion := 0.0

	for coord, ice := range is.Ice {
		if ice.Thickness < MinIceThickness {
			continue
		}

		// Erosion rate = coefficient * ice flux (thickness * velocity)
		flux := ice.Thickness * ice.FlowSpeed
		erosionDepth := flux * ErosionCoefficient * dt / 1000.0 // Convert mm to m

		// Apply erosion to bedrock
		currentElev := heightmap.Get(coord)
		newElev := currentElev - erosionDepth
		heightmap.Set(coord, newElev)
		totalErosion += erosionDepth

		// Accumulate sediment (for moraine deposition)
		is.Sediment[coord] += erosionDepth * SedimentCapacity
	}

	return totalErosion
}

// DepositMoraines deposits accumulated sediment at ice margins.
// Called when ice retreats.
func (is *IceSheet) DepositMoraines(heightmap *SphereHeightmap, topology spatial.Topology, resolution int) {
	for coord, sediment := range is.Sediment {
		ice, hasIce := is.Ice[coord]

		// Deposit sediment where ice is thin or absent (margins)
		if !hasIce || ice.Thickness < 10.0 {
			if sediment > 0.1 {
				// Raise terrain by depositing sediment
				currentElev := heightmap.Get(coord)
				heightmap.Set(coord, currentElev+sediment*10.0) // Scale factor
				is.Sediment[coord] = 0
			}
		}
	}
}

// updateStats recalculates aggregate statistics.
func (is *IceSheet) updateStats(resolution int) {
	is.TotalVolume = 0
	is.TotalArea = 0
	is.MaxThickness = 0

	cellAreaKm2 := 100.0 * 100.0 // Assume 100km grid cells

	for _, ice := range is.Ice {
		if ice.Thickness >= MinIceThickness {
			is.TotalArea += cellAreaKm2
			is.TotalVolume += ice.Thickness / 1000.0 * cellAreaKm2 // m to km * km²
			if ice.Thickness > is.MaxThickness {
				is.MaxThickness = ice.Thickness
			}
		}
	}
}

// Helper: Convert flat index to coordinate
func iceIndexToCoord(idx, resolution int) spatial.Coordinate {
	resSq := resolution * resolution
	face := idx / resSq
	rem := idx % resSq
	y := rem / resolution
	x := rem % resolution
	return spatial.Coordinate{Face: face, X: x, Y: y}
}

// Helper: Convert coordinate to flat index
func iceCoordToIndex(coord spatial.Coordinate, resolution int) int {
	return coord.Face*resolution*resolution + coord.Y*resolution + coord.X
}

// GlacialFeatureType represents the type of glacial landform
type GlacialFeatureType string

const (
	FeatureUValley GlacialFeatureType = "u_valley"
	FeatureCirque  GlacialFeatureType = "cirque"
	FeatureFjord   GlacialFeatureType = "fjord"
	FeatureArete   GlacialFeatureType = "arete"
	FeatureMoraine GlacialFeatureType = "moraine"
)

// GlacialFeature represents a detected glacial landform
type GlacialFeature struct {
	Type     GlacialFeatureType
	Location spatial.Coordinate
	Size     float64 // Relative size/importance
}

// DetectGlacialFeatures identifies glacial landforms in the terrain.
// Call after ice retreat to identify U-valleys, cirques, and fjords.
func (is *IceSheet) DetectGlacialFeatures(heightmap *SphereHeightmap, topology spatial.Topology, seaLevel float64) []GlacialFeature {
	features := make([]GlacialFeature, 0)
	resolution := topology.Resolution()

	for coord, sediment := range is.Sediment {
		if sediment < 0.1 {
			continue
		}

		elev := heightmap.Get(coord)

		// Check for moraine deposits
		if sediment > 0.5 {
			features = append(features, GlacialFeature{
				Type:     FeatureMoraine,
				Location: coord,
				Size:     sediment,
			})
		}

		// Check for U-valley characteristics:
		// - Eroded (had ice)
		// - Neighbors are higher (valley walls)
		// - Below the ice extent
		neighborElevs := make([]float64, 0, 4)
		for _, dir := range []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West} {
			neighbor := topology.GetNeighbor(coord, dir)
			neighborElevs = append(neighborElevs, heightmap.Get(neighbor))
		}

		avgNeighbor := 0.0
		for _, ne := range neighborElevs {
			avgNeighbor += ne
		}
		avgNeighbor /= float64(len(neighborElevs))

		// U-valley: center is lower than neighbors (trough)
		if elev < avgNeighbor-100 && elev > seaLevel {
			features = append(features, GlacialFeature{
				Type:     FeatureUValley,
				Location: coord,
				Size:     avgNeighbor - elev,
			})
		}

		// Fjord: U-valley that's below sea level
		if elev < seaLevel && sediment > 0.3 {
			features = append(features, GlacialFeature{
				Type:     FeatureFjord,
				Location: coord,
				Size:     seaLevel - elev,
			})
		}
	}

	_ = resolution // Used for potential future calculations
	return features
}

// CreateGlacialLakes identifies locations where moraine dams may form lakes.
// Returns coordinates where lakes should form.
func (is *IceSheet) CreateGlacialLakes(heightmap *SphereHeightmap, topology spatial.Topology) []spatial.Coordinate {
	lakes := make([]spatial.Coordinate, 0)
	resolution := topology.Resolution()

	for coord, sediment := range is.Sediment {
		if sediment < 1.0 {
			continue // Need significant moraine deposit
		}

		// Check if this moraine is damming water (neighbors behind it are lower)
		elev := heightmap.Get(coord)
		dammed := false

		for _, dir := range []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West} {
			neighbor := topology.GetNeighbor(coord, dir)
			neighborElev := heightmap.Get(neighbor)

			// If a neighbor is significantly lower, it may be a lake bed
			if neighborElev < elev-50 {
				lakes = append(lakes, neighbor)
				dammed = true
			}
		}

		if dammed {
			// Mark the moraine location too
			_ = coord
		}
	}

	_ = resolution
	return lakes
}

// ApplyIsostaticRebound simulates post-glacial rebound after ice removal.
// The crust rises as the weight of ice is removed.
func (is *IceSheet) ApplyIsostaticRebound(heightmap *SphereHeightmap, previousIce map[spatial.Coordinate]*IceData, reboundRate float64, dt float64) {
	for coord, oldIce := range previousIce {
		currentIce, stillHasIce := is.Ice[coord]
		var iceRemoved float64

		if stillHasIce {
			iceRemoved = oldIce.Thickness - currentIce.Thickness
		} else {
			iceRemoved = oldIce.Thickness
		}

		if iceRemoved > 0 {
			// Rebound proportional to ice removed
			// Full rebound takes ~10,000 years
			rebound := iceRemoved * reboundRate * dt / 10000.0
			currentElev := heightmap.Get(coord)
			heightmap.Set(coord, currentElev+rebound)
		}
	}
}

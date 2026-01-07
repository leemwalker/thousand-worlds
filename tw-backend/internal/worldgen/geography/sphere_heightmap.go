package geography

import (
	"runtime"
	"sync"

	"tw-backend/internal/spatial"
)

// =============================================================================
// Cell Data Layer (Phase 5: Geological Provinces)
// =============================================================================

// CellData stores geological properties for a single cell.
// Used for differential erosion based on rock hardness and sediment tracking.
type CellData struct {
	RockHardness float64 // 0.0 (Soft sandstone/sediment) to 1.0 (Hard granite)
	Sediment     float64 // Depth of loose material in meters on top of bedrock
	ProvinceID   int     // ID of the geological province (Craton, FoldBelt, Basin)
	Flux         float64 // Water flow accumulation (Phase 6)
	IsLake       bool    // Part of a lake (Phase 6)
	LakeID       int     // ID of the lake (Phase 6)
	LakeDepth    float64 // Depth of water in meters (Surface - Bedrock) (Phase 6)
	// Coastal features (Phase 4)
	IsIntertidal bool // In tidal zone (exposed at low tide, submerged at high)
	IsEstuary    bool // River-ocean mixing zone
	IsSpit       bool // Extended sediment bar

	// Crustal Properties (Phase 8b)
	IsContinental bool // True if this cell is Continental Crust (Granite/Island Arc), False if Oceanic (Basalt)

	// Sediment Mineral Content (Phase 9: Mineral Tracking)
	// Maps mineral name to amount in the sediment layer (e.g., {"Gold": 0.001, "Iron": 0.5})
	// Used for placer deposit formation at river deltas and coastal shelves.
	SedimentMinerals map[string]float64
}

// SphereHeightmap wraps 6 flat Heightmaps into a spherical surface
// using the cube-sphere topology for neighbor lookups.
// The elevation stored is the Total Surface Height (Bedrock + Sediment).
// Bedrock height can be derived as: BedrockHeight = TotalHeight - Sediment.
type SphereHeightmap struct {
	topology spatial.Topology
	faces    [6]*Heightmap
	cellData [6][]CellData // Parallel array for geological cell properties
	MinElev  float64
	MaxElev  float64

	// Delta tracking for event streaming (Phase 1: Event Infrastructure)
	deltaEnabled bool
	deltas       []CellChange
}

// CellChange represents a single cell's elevation/sediment change.
// Used for delta streaming in simulation events.
type CellChange struct {
	Face           int
	X, Y           int
	ElevationDelta float64
	SedimentDelta  float64
}

// NewSphereHeightmap creates a new spherical heightmap using the given topology.
// Each face is initialized with a flat Heightmap of size Resolution x Resolution.
// Cell data is initialized with zero values (0.0 hardness, 0.0 sediment, 0 province).
func NewSphereHeightmap(topology spatial.Topology) *SphereHeightmap {
	res := topology.Resolution()
	shm := &SphereHeightmap{
		topology: topology,
	}

	for i := 0; i < 6; i++ {
		shm.faces[i] = NewHeightmap(res, res)
		shm.cellData[i] = make([]CellData, res*res)
	}

	return shm
}

// Resolution returns the grid size of each face
func (s *SphereHeightmap) Resolution() int {
	return s.topology.Resolution()
}

// Get returns the elevation at the given spherical coordinate
func (s *SphereHeightmap) Get(coord spatial.Coordinate) float64 {
	if coord.Face < 0 || coord.Face >= 6 {
		return 0
	}
	return s.faces[coord.Face].Get(coord.X, coord.Y)
}

// Set sets the elevation at the given spherical coordinate
func (s *SphereHeightmap) Set(coord spatial.Coordinate, val float64) {
	if coord.Face < 0 || coord.Face >= 6 {
		return
	}
	// Record delta if tracking is enabled
	if s.deltaEnabled {
		oldVal := s.faces[coord.Face].Get(coord.X, coord.Y)
		s.recordDelta(coord.Face, coord.X, coord.Y, val-oldVal, 0)
	}
	s.faces[coord.Face].Set(coord.X, coord.Y, val)
}

// =============================================================================
// Cell Data Accessors (Phase 5: Geological Provinces)
// =============================================================================

// GetCellData returns the geological properties for a cell
func (s *SphereHeightmap) GetCellData(coord spatial.Coordinate) CellData {
	if coord.Face < 0 || coord.Face >= 6 {
		return CellData{}
	}
	res := s.topology.Resolution()
	idx := coord.Y*res + coord.X
	if idx < 0 || idx >= len(s.cellData[coord.Face]) {
		return CellData{}
	}
	return s.cellData[coord.Face][idx]
}

// SetCellData sets the geological properties for a cell
func (s *SphereHeightmap) SetCellData(coord spatial.Coordinate, data CellData) {
	if coord.Face < 0 || coord.Face >= 6 {
		return
	}
	res := s.topology.Resolution()
	idx := coord.Y*res + coord.X
	if idx < 0 || idx >= len(s.cellData[coord.Face]) {
		return
	}
	s.cellData[coord.Face][idx] = data
}

// GetRockHardness returns the rock hardness (0.0-1.0) for a cell
// This is a convenience method for erosion calculations.
func (s *SphereHeightmap) GetRockHardness(coord spatial.Coordinate) float64 {
	return s.GetCellData(coord).RockHardness
}

// AddSediment adds loose sediment material to a cell.
// This increases BOTH the Sediment depth AND the TotalHeight (surface elevation).
// Use this when depositing eroded material.
func (s *SphereHeightmap) AddSediment(coord spatial.Coordinate, amount float64) {
	if coord.Face < 0 || coord.Face >= 6 || amount <= 0 {
		return
	}
	res := s.topology.Resolution()
	idx := coord.Y*res + coord.X
	if idx < 0 || idx >= len(s.cellData[coord.Face]) {
		return
	}
	// Record delta if tracking is enabled
	s.recordDelta(coord.Face, coord.X, coord.Y, amount, amount)
	// Increase sediment depth
	s.cellData[coord.Face][idx].Sediment += amount
	// Increase total surface height
	currentElev := s.Get(coord)
	s.faces[coord.Face].Set(coord.X, coord.Y, currentElev+amount) // Direct set to avoid double recording
}

// Erode removes material from a cell, sediment first, then bedrock.
// Returns the actual amount of material removed (may be less than requested).
// This decreases TotalHeight by the removed amount.
// Sediment is removed before bedrock to simulate realistic erosion.
func (s *SphereHeightmap) Erode(coord spatial.Coordinate, amount float64) float64 {
	if coord.Face < 0 || coord.Face >= 6 || amount <= 0 {
		return 0
	}
	res := s.topology.Resolution()
	idx := coord.Y*res + coord.X
	if idx < 0 || idx >= len(s.cellData[coord.Face]) {
		return 0
	}

	currentElev := s.Get(coord)
	sediment := s.cellData[coord.Face][idx].Sediment
	totalRemoved := 0.0
	sedimentRemoved := 0.0

	// Phase 1: Remove sediment first
	if sediment > 0 {
		sedimentToRemove := amount
		if sedimentToRemove > sediment {
			sedimentToRemove = sediment
		}
		s.cellData[coord.Face][idx].Sediment -= sedimentToRemove
		sedimentRemoved = sedimentToRemove
		totalRemoved += sedimentToRemove
		amount -= sedimentToRemove
	}

	// Phase 2: Remove bedrock if sediment exhausted
	if amount > 0 {
		totalRemoved += amount
	}

	// Record delta if tracking is enabled (negative values for erosion)
	s.recordDelta(coord.Face, coord.X, coord.Y, -totalRemoved, -sedimentRemoved)

	// Update total surface elevation (direct set to avoid double recording)
	s.faces[coord.Face].Set(coord.X, coord.Y, currentElev-totalRemoved)

	return totalRemoved
}

// AddSedimentWithMinerals adds sediment material with associated mineral content.
// This increases both the sediment depth and elevation, while accumulating minerals.
// Used for delta deposition and coastal shelf building.
func (s *SphereHeightmap) AddSedimentWithMinerals(coord spatial.Coordinate, amount float64, minerals map[string]float64) {
	if coord.Face < 0 || coord.Face >= 6 || amount <= 0 {
		return
	}
	res := s.topology.Resolution()
	idx := coord.Y*res + coord.X
	if idx < 0 || idx >= len(s.cellData[coord.Face]) {
		return
	}

	// Increase sediment depth
	s.cellData[coord.Face][idx].Sediment += amount

	// Initialize mineral map if needed
	if s.cellData[coord.Face][idx].SedimentMinerals == nil {
		s.cellData[coord.Face][idx].SedimentMinerals = make(map[string]float64)
	}

	// Accumulate minerals
	for mineral, mineralAmount := range minerals {
		s.cellData[coord.Face][idx].SedimentMinerals[mineral] += mineralAmount
	}

	// Increase total surface height
	currentElev := s.Get(coord)
	s.Set(coord, currentElev+amount)
}

// ErodeWithMinerals removes material from a cell and returns extracted minerals.
// Minerals are extracted proportionally to the amount of sediment eroded.
// Returns (totalEroded, extractedMinerals).
func (s *SphereHeightmap) ErodeWithMinerals(coord spatial.Coordinate, amount float64) (float64, map[string]float64) {
	if coord.Face < 0 || coord.Face >= 6 || amount <= 0 {
		return 0, nil
	}
	res := s.topology.Resolution()
	idx := coord.Y*res + coord.X
	if idx < 0 || idx >= len(s.cellData[coord.Face]) {
		return 0, nil
	}

	currentElev := s.Get(coord)
	sediment := s.cellData[coord.Face][idx].Sediment
	totalRemoved := 0.0
	extractedMinerals := make(map[string]float64)

	// Phase 1: Remove sediment first (with proportional mineral extraction)
	if sediment > 0 {
		sedimentToRemove := amount
		if sedimentToRemove > sediment {
			sedimentToRemove = sediment
		}

		// Calculate extraction ratio
		extractionRatio := sedimentToRemove / sediment

		// Extract proportional minerals from sediment
		cellMinerals := s.cellData[coord.Face][idx].SedimentMinerals
		if cellMinerals != nil {
			for mineral, mineralAmount := range cellMinerals {
				extracted := mineralAmount * extractionRatio
				extractedMinerals[mineral] = extracted
				s.cellData[coord.Face][idx].SedimentMinerals[mineral] -= extracted
			}
		}

		s.cellData[coord.Face][idx].Sediment -= sedimentToRemove
		totalRemoved += sedimentToRemove
		amount -= sedimentToRemove
	}

	// Phase 2: Remove bedrock if sediment exhausted (no minerals from bedrock erosion here)
	if amount > 0 {
		totalRemoved += amount
	}

	// Update total surface elevation
	s.Set(coord, currentElev-totalRemoved)

	return totalRemoved, extractedMinerals
}

// GetMineralDensities returns a map of mineral names to their densities (g/cm³).
// Used for density sorting in placer deposit formation - heavier minerals deposit first.
func GetMineralDensities() map[string]float64 {
	return map[string]float64{
		// Heavy minerals (deposit first at river mouths)
		"Platinum": 21.5,
		"Gold":     19.3,
		"Silver":   10.5,

		// Medium density (deposit in mid-delta)
		"Copper": 8.9,
		"Iron":   7.9,
		"Tin":    7.3,

		// Light minerals (transport further downstream)
		"Quartz":    2.65,
		"Limestone": 2.7,
		"Coal":      1.3,
	}
}

// GetNeighborElevation returns the elevation of the neighboring cell in the given direction.
// Handles cross-face transitions automatically using the topology.
func (s *SphereHeightmap) GetNeighborElevation(coord spatial.Coordinate, dir spatial.Direction) float64 {
	neighborCoord := s.topology.GetNeighbor(coord, dir)
	return s.Get(neighborCoord)
}

// UpdateMinMax recalculates the minimum and maximum elevations across all faces
func (s *SphereHeightmap) UpdateMinMax() {
	first := true
	for _, face := range s.faces {
		for _, elev := range face.Elevations {
			if first {
				s.MinElev = elev
				s.MaxElev = elev
				first = false
			} else {
				if elev < s.MinElev {
					s.MinElev = elev
				}
				if elev > s.MaxElev {
					s.MaxElev = elev
				}
			}
		}
	}
}

// MinMax returns the minimum and maximum elevations
func (s *SphereHeightmap) MinMax() (min, max float64) {
	return s.MinElev, s.MaxElev
}

// GetFace returns the underlying Heightmap for a specific face.
// Useful for bulk operations or serialization.
func (s *SphereHeightmap) GetFace(face int) *Heightmap {
	if face < 0 || face >= 6 {
		return nil
	}
	return s.faces[face]
}

// Topology returns the underlying topology for neighbor lookups
func (s *SphereHeightmap) Topology() spatial.Topology {
	return s.topology
}

// ToFlatHeightmap converts this spherical heightmap to a flat equirectangular projection.
// Uses latitude/longitude mapping for proper global coverage of all 6 cube-sphere faces.
// width and height specify the dimensions of the output heightmap.
// NOTE: This allocates a new heightmap. For repeated calls, use ToFlatHeightmapInPlace.
func (s *SphereHeightmap) ToFlatHeightmap(width, height int) *Heightmap {
	flat := NewHeightmap(width, height)
	s.ToFlatHeightmapInPlace(flat)
	return flat
}

// ToFlatHeightmapInPlace converts this spherical heightmap to a flat equirectangular projection,
// writing directly into the provided heightmap to avoid memory allocation.
// The destination heightmap must already be the correct size.
// Parallelized for performance on large heightmaps.
func (s *SphereHeightmap) ToFlatHeightmapInPlace(dest *Heightmap) {
	width := dest.Width
	height := dest.Height

	// Parallelize by row chunks using worker pool
	workers := runtime.NumCPU()
	rowsPerWorker := height / workers
	if rowsPerWorker < 1 {
		rowsPerWorker = 1
		workers = height
	}

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		startY := w * rowsPerWorker
		endY := startY + rowsPerWorker
		if w == workers-1 {
			endY = height // Last worker takes remaining rows
		}

		wg.Add(1)
		go func(sy, ey int) {
			defer wg.Done()
			for y := sy; y < ey; y++ {
				for x := 0; x < width; x++ {
					// Map pixel coordinates to longitude and latitude
					lon := (float64(x) / float64(width)) * 2 * 3.141592653589793
					lat := (0.5 - float64(y)/float64(height)) * 3.141592653589793

					// Convert lat/lon to 3D unit sphere coordinates
					cosLat := cosineApprox(lat)
					sinLat := sineApprox(lat)
					cosLon := cosineApprox(lon)
					sinLon := sineApprox(lon)

					sphereX := cosLat * cosLon
					sphereY := sinLat
					sphereZ := cosLat * sinLon

					// Use topology to find the correct cube-sphere face and coordinate
					coord := s.topology.FromVector(sphereX, sphereY, sphereZ)
					elev := s.Get(coord)
					dest.Set(x, y, elev)
				}
			}
		}(startY, endY)
	}
	wg.Wait()

	dest.MinElev = s.MinElev
	dest.MaxElev = s.MaxElev
}

// MapIntToFlat creates a flat equirectangular projection of integer data associated with spherical coordinates.
// It uses the same projection logic as ToFlatHeightmap to ensure alignment.
// inputFunc returns the integer value for a given coordinate (or specific sentinel if not found).
func (s *SphereHeightmap) MapIntToFlat(width, height int, inputFunc func(spatial.Coordinate) int) []int {
	result := make([]int, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map pixel coordinates to longitude and latitude
			// Longitude: 0 to 2π (left to right)
			// Latitude: π/2 to -π/2 (top to bottom, north pole to south pole)
			lon := (float64(x) / float64(width)) * 2 * 3.141592653589793  // 0 to 2π
			lat := (0.5 - float64(y)/float64(height)) * 3.141592653589793 // π/2 to -π/2

			// Convert lat/lon to 3D unit sphere coordinates
			cosLat := cosineApprox(lat)
			sinLat := sineApprox(lat)
			cosLon := cosineApprox(lon)
			sinLon := sineApprox(lon)

			sphereX := cosLat * cosLon
			sphereY := sinLat
			sphereZ := cosLat * sinLon

			// Use topology to find the correct cube-sphere face and coordinate
			coord := s.topology.FromVector(sphereX, sphereY, sphereZ)

			// Get value from input function
			val := inputFunc(coord)

			// Set in flat array
			idx := y*width + x
			result[idx] = val
		}
	}
	return result
}

// cosineApprox provides cosine using math package
func cosineApprox(x float64) float64 {
	// Using Taylor series approximation to avoid import cycle
	// cos(x) = 1 - x²/2! + x⁴/4! - x⁶/6! + ...
	// For better accuracy, normalize x to [-π, π]
	const pi = 3.141592653589793
	const twoPi = 2 * pi

	// Normalize to [-π, π]
	for x > pi {
		x -= twoPi
	}
	for x < -pi {
		x += twoPi
	}

	x2 := x * x
	x4 := x2 * x2
	x6 := x4 * x2
	x8 := x6 * x2

	return 1 - x2/2 + x4/24 - x6/720 + x8/40320
}

// sineApprox provides sine using Taylor series
func sineApprox(x float64) float64 {
	const pi = 3.141592653589793
	const twoPi = 2 * pi

	// Normalize to [-π, π]
	for x > pi {
		x -= twoPi
	}
	for x < -pi {
		x += twoPi
	}

	x2 := x * x
	x3 := x2 * x
	x5 := x3 * x2
	x7 := x5 * x2
	x9 := x7 * x2

	return x - x3/6 + x5/120 - x7/5040 + x9/362880
}

// ClampElevations constrains all elevation values to be within [minElev, maxElev].
// This prevents runaway elevation accumulation over geological time.
func (s *SphereHeightmap) ClampElevations(minElev, maxElev float64) {
	for _, face := range s.faces {
		for i, elev := range face.Elevations {
			if elev > maxElev {
				face.Elevations[i] = maxElev
			} else if elev < minElev {
				face.Elevations[i] = minElev
			}
		}
	}
	// Update min/max after clamping
	s.UpdateMinMax()
}

// =============================================================================
// Delta Tracking (Phase 1: Event Infrastructure)
// =============================================================================

// EnableDeltaTracking starts recording elevation/sediment changes.
// Changes are accumulated until FlushDeltas is called.
// This is designed for event streaming - enable before a simulation cycle,
// then flush and emit the delta as an event.
func (s *SphereHeightmap) EnableDeltaTracking() {
	s.deltaEnabled = true
	s.deltas = nil // Reset any previous deltas
}

// DisableDeltaTracking stops recording changes without flushing.
func (s *SphereHeightmap) DisableDeltaTracking() {
	s.deltaEnabled = false
	s.deltas = nil
}

// FlushDeltas returns accumulated changes and resets the delta buffer.
// Returns nil if no changes were recorded.
func (s *SphereHeightmap) FlushDeltas() []CellChange {
	result := s.deltas
	s.deltas = nil
	return result
}

// DeltaCount returns the number of accumulated deltas.
// Useful for deciding whether to emit an event.
func (s *SphereHeightmap) DeltaCount() int {
	return len(s.deltas)
}

// IsDeltaTrackingEnabled returns whether delta tracking is active.
func (s *SphereHeightmap) IsDeltaTrackingEnabled() bool {
	return s.deltaEnabled
}

// recordDelta records a cell change if delta tracking is enabled.
// Called internally by Set, AddSediment, Erode, etc.
func (s *SphereHeightmap) recordDelta(face, x, y int, elevDelta, sedDelta float64) {
	if !s.deltaEnabled {
		return
	}
	// Only record if there's a meaningful change
	if elevDelta == 0 && sedDelta == 0 {
		return
	}
	s.deltas = append(s.deltas, CellChange{
		Face:           face,
		X:              x,
		Y:              y,
		ElevationDelta: elevDelta,
		SedimentDelta:  sedDelta,
	})
}

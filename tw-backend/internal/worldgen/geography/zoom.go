package geography

import (
	"encoding/binary"
	"hash/fnv"
	"math"

	"tw-backend/internal/spatial"
)

// =============================================================================
// Infinite Zoom: Macro-to-Micro Coordinate System
// =============================================================================

// MacroLocation represents a position on the global grid with sub-cell precision.
// Face/X/Y identify the macro cell, U/V are the 0-1 offset within that cell.
type MacroLocation struct {
	Face int     // Cube-sphere face [0-5]
	X    int     // Cell X coordinate within face
	Y    int     // Cell Y coordinate within face
	U    float64 // Horizontal offset within cell [0, 1)
	V    float64 // Vertical offset within cell [0, 1)
}

// LODLevel defines the level of detail for terrain generation
type LODLevel int

const (
	LODGlobal   LODLevel = 0 // 300km - Macro cell
	LODRegional LODLevel = 1 // 50km - Major features
	LODLocal    LODLevel = 2 // 1km - Hills, valleys
	LODDetailed LODLevel = 3 // 100m - Boulders, paths
	LODFine     LODLevel = 4 // 10m - Individual trees
	LODNPC      LODLevel = 5 // 1m - Ground detail
)

// LODScales maps each LOD level to its approximate meters-per-unit scale
var LODScales = map[LODLevel]float64{
	LODGlobal:   312000, // 312 km
	LODRegional: 50000,  // 50 km
	LODLocal:    1000,   // 1 km
	LODDetailed: 100,    // 100 m
	LODFine:     10,     // 10 m
	LODNPC:      1,      // 1 m
}

// =============================================================================
// Coordinate Conversion
// =============================================================================

// LatLonToSphere converts latitude/longitude (degrees) to unit sphere coordinates.
// Convention: Y is up (north pole), XZ is equatorial plane.
func LatLonToSphere(latDeg, lonDeg float64) (x, y, z float64) {
	lat := latDeg * math.Pi / 180.0
	lon := lonDeg * math.Pi / 180.0

	cosLat := math.Cos(lat)
	sinLat := math.Sin(lat)
	cosLon := math.Cos(lon)
	sinLon := math.Sin(lon)

	x = cosLat * cosLon
	y = sinLat
	z = cosLat * sinLon

	return x, y, z
}

// SphereToLatLon converts unit sphere coordinates to latitude/longitude (degrees).
func SphereToLatLon(x, y, z float64) (latDeg, lonDeg float64) {
	// Normalize just in case
	mag := math.Sqrt(x*x + y*y + z*z)
	if mag > 0 {
		x, y, z = x/mag, y/mag, z/mag
	}

	lat := math.Asin(y)
	lon := math.Atan2(z, x)

	return lat * 180.0 / math.Pi, lon * 180.0 / math.Pi
}

// GlobalToMacro converts latitude/longitude to a macro cell location with sub-cell offset.
// Uses the CubeSphere topology for face detection and coordinate mapping.
func GlobalToMacro(latDeg, lonDeg float64, topology spatial.Topology) MacroLocation {
	// Convert to sphere coordinates
	x, y, z := LatLonToSphere(latDeg, lonDeg)

	// Get macro cell from topology
	coord := topology.FromVector(x, y, z)

	// Calculate sub-cell offset
	u, v := calculateCellOffset(x, y, z, coord, topology)

	return MacroLocation{
		Face: coord.Face,
		X:    coord.X,
		Y:    coord.Y,
		U:    u,
		V:    v,
	}
}

// calculateCellOffset computes the [0,1) offset within a macro cell.
// This is the fractional position of the point within the cell boundaries.
func calculateCellOffset(x, y, z float64, coord spatial.Coordinate, topology spatial.Topology) (u, v float64) {
	resolution := topology.Resolution()

	// Get the cell's center in sphere coordinates
	sx, sy, sz := topology.ToSphere(coord)

	// Project the point onto the local tangent plane at cell center
	// Use the difference vector and project onto face-aligned axes

	// For simplicity, use the UV coordinates from the cube projection
	// The FromVector already computed which face we're on; now find fractional position

	// Re-project to cube face to get precise u,v
	absX, absY, absZ := math.Abs(x), math.Abs(y), math.Abs(z)

	var cubeU, cubeV float64

	switch {
	case absZ >= absX && absZ >= absY: // Front or Back face
		if z > 0 {
			cubeU, cubeV = x/z, -y/z
		} else {
			cubeU, cubeV = -x/(-z), -y/(-z)
		}
	case absX >= absY: // Left or Right face
		if x > 0 {
			cubeU, cubeV = -z/x, -y/x
		} else {
			cubeU, cubeV = z/(-x), -y/(-x)
		}
	default: // Top or Bottom face
		if y > 0 {
			cubeU, cubeV = x/y, z/y
		} else {
			cubeU, cubeV = x/(-y), -z/(-y)
		}
	}

	// Convert from [-1, 1] to [0, resolution]
	gridX := (cubeU + 1) / 2 * float64(resolution)
	gridY := (cubeV + 1) / 2 * float64(resolution)

	// Extract fractional part as offset
	u = gridX - float64(coord.X)
	v = gridY - float64(coord.Y)

	// Clamp to [0, 1)
	u = math.Max(0, math.Min(0.9999, u))
	v = math.Max(0, math.Min(0.9999, v))

	// Avoid using sx, sy, sz directly in offsets
	_ = sx
	_ = sy
	_ = sz

	return u, v
}

// MacroToGlobal converts a macro location back to latitude/longitude.
func MacroToGlobal(loc MacroLocation, topology spatial.Topology) (latDeg, lonDeg float64) {
	resolution := topology.Resolution()

	// Compute interpolated grid position
	gridX := float64(loc.X) + loc.U
	gridY := float64(loc.Y) + loc.V

	// Convert to [-1, 1] cube coordinates
	cubeU := gridX/float64(resolution)*2 - 1
	cubeV := gridY/float64(resolution)*2 - 1

	// Map to sphere based on face
	var x, y, z float64
	switch loc.Face {
	case spatial.FaceFront:
		x, y, z = cubeU, -cubeV, 1
	case spatial.FaceBack:
		x, y, z = -cubeU, -cubeV, -1
	case spatial.FaceLeft:
		x, y, z = -1, -cubeV, cubeU
	case spatial.FaceRight:
		x, y, z = 1, -cubeV, -cubeU
	case spatial.FaceTop:
		x, y, z = cubeU, 1, cubeV
	case spatial.FaceBottom:
		x, y, z = cubeU, -1, -cubeV
	}

	// Normalize to unit sphere
	mag := math.Sqrt(x*x + y*y + z*z)
	x, y, z = x/mag, y/mag, z/mag

	return SphereToLatLon(x, y, z)
}

// =============================================================================
// Deterministic Seeding
// =============================================================================

// GenerateLocalSeed creates a unique, reproducible seed for a specific location and detail level.
// Uses FNV-1a hash to combine inputs into a single int64 seed.
// This ensures returning players see the exact same procedural content.
func GenerateLocalSeed(globalSeed int64, face, x, y int, detailLevel LODLevel) int64 {
	h := fnv.New64a()

	// Write all components in fixed order
	binary.Write(h, binary.LittleEndian, globalSeed)
	binary.Write(h, binary.LittleEndian, int32(face))
	binary.Write(h, binary.LittleEndian, int32(x))
	binary.Write(h, binary.LittleEndian, int32(y))
	binary.Write(h, binary.LittleEndian, int32(detailLevel))

	return int64(h.Sum64())
}

// GenerateSubCellSeed creates a seed for a specific sub-cell within a macro cell.
// Used for LOD > 0 where we subdivide cells into smaller patches.
func GenerateSubCellSeed(parentSeed int64, subX, subY int) int64 {
	h := fnv.New64a()

	binary.Write(h, binary.LittleEndian, parentSeed)
	binary.Write(h, binary.LittleEndian, int32(subX))
	binary.Write(h, binary.LittleEndian, int32(subY))

	return int64(h.Sum64())
}

// =============================================================================
// Biome-Influenced Micro Generation
// =============================================================================

// GetMicroNoiseConfig returns FBM configuration tuned for a specific biome type.
// Mountain biomes get high roughness, plains get smooth terrain.
func GetMicroNoiseConfig(biome BiomeType) FBMConfig {
	switch biome {
	case BiomeMountain, BiomeHighMountain, BiomeAlpine:
		return FBMConfig{
			Octaves:      8,
			Frequency:    0.03,
			Lacunarity:   2.1,
			Persistence:  0.6, // High roughness
			WarpStrength: 0.5, // Strong distortion
		}

	case BiomeHighland:
		return FBMConfig{
			Octaves:      7,
			Frequency:    0.025,
			Lacunarity:   2.0,
			Persistence:  0.5,
			WarpStrength: 0.4,
		}

	case BiomeGrassland, BiomeLowland:
		return FBMConfig{
			Octaves:      4,
			Frequency:    0.02,
			Lacunarity:   2.0,
			Persistence:  0.3, // Gentle rolling
			WarpStrength: 0.2,
		}

	case BiomeDesert:
		return FBMConfig{
			Octaves:      5,
			Frequency:    0.04,
			Lacunarity:   2.2,
			Persistence:  0.4, // Dune-like
			WarpStrength: 0.6, // Wavy patterns
		}

	case BiomeRainforest, BiomeDeciduousForest:
		return FBMConfig{
			Octaves:      6,
			Frequency:    0.025,
			Lacunarity:   2.0,
			Persistence:  0.45,
			WarpStrength: 0.35,
		}

	case BiomeTaiga, BiomeTundra:
		return FBMConfig{
			Octaves:      5,
			Frequency:    0.02,
			Lacunarity:   2.0,
			Persistence:  0.35,
			WarpStrength: 0.25,
		}

	case BiomeOcean:
		return FBMConfig{
			Octaves:      4,
			Frequency:    0.015,
			Lacunarity:   2.0,
			Persistence:  0.25, // Mostly flat seafloor
			WarpStrength: 0.15,
		}

	default:
		return DefaultTerrainFBMConfig()
	}
}

// =============================================================================
// Micro Terrain Generation
// =============================================================================

// MicroTerrainParams contains all parameters needed to generate micro terrain at a location.
type MicroTerrainParams struct {
	GlobalSeed    int64
	Location      MacroLocation
	BaseElevation float64
	Biome         BiomeType
	LOD           LODLevel
}

// GenerateMicroTerrain creates local terrain detail for a specific point.
// Returns the elevation offset to add to the base macro elevation.
func GenerateMicroTerrain(params MicroTerrainParams) float64 {
	// Generate deterministic seed for this location
	localSeed := GenerateLocalSeed(
		params.GlobalSeed,
		params.Location.Face,
		params.Location.X,
		params.Location.Y,
		params.LOD,
	)

	// Get biome-appropriate noise config
	config := GetMicroNoiseConfig(params.Biome)

	// Create FBM generator with local seed
	fbm := NewFBMGenerator(localSeed, config)

	// Scale based on LOD
	scale := LODScales[params.LOD]
	if scale == 0 {
		scale = 1000 // Default 1km
	}

	// Sample at sub-cell position, scaled for detail level
	sampleX := params.Location.U * scale
	sampleY := params.Location.V * scale

	// Generate normalized noise and scale to appropriate elevation range
	// Higher LODs generate smaller variations
	maxVariation := 100.0 / float64(params.LOD+1) // 100m at LOD0, 50m at LOD1, etc.

	return fbm.FBM2D(sampleX, sampleY) * maxVariation
}

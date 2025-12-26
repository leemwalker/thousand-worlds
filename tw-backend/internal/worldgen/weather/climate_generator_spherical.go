package weather

import (
	"math"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

// SphereClimateMap is climate data organized by cube-sphere faces.
// Access: [face][y*faceSize + x]
type SphereClimateMap [][]ClimateData

// GenerateInitialClimateSpherical creates climate data for a spherical world.
// Uses proper 3D position to derive latitude, applying accurate temperature
// gradients based on distance from equator.
//
// Parameters:
//   - sphereMap: Spherical heightmap data
//   - topology: The cube-sphere topology for coordinate conversions
//   - seaLevel: Current sea level in meters
//   - seed: Random seed for moisture noise
//   - globalTempMod: Global temperature modifier (e.g., volcanic winter = -10)
//   - geothermalOffset: Temperature increase from planetary internal heat (early Earth = +90°C)
//
// Returns climate data organized by face, with each face indexed row-major.
func GenerateInitialClimateSpherical(
	sphereMap *geography.SphereHeightmap,
	topology spatial.Topology,
	seaLevel float64,
	seed int64,
	globalTempMod float64,
	geothermalOffset float64,
) SphereClimateMap {
	faceSize := sphereMap.Resolution()
	climateData := make(SphereClimateMap, 6)

	// Initialize each face's climate data
	for face := 0; face < 6; face++ {
		climateData[face] = make([]ClimateData, faceSize*faceSize)
	}

	// Use Perlin noise for moisture patterns
	noise := geography.NewPerlinGenerator(seed)

	// Generate climate for each cell on each face
	for face := 0; face < 6; face++ {
		for y := 0; y < faceSize; y++ {
			for x := 0; x < faceSize; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				idx := y*faceSize + x

				// Get elevation from spherical heightmap
				elevation := sphereMap.Get(coord)

				// Get latitude from 3D position (the key spherical calculation)
				latitude := GetLatitudeFromCoord(topology, coord)

				// Normalize latitude to 0-1 range (0 at equator, 1 at poles)
				latitudeNormalized := math.Abs(latitude) / 90.0

				// Calculate base temperature using physics model
				baseTemp := calculateTemperatureFromLatitude(latitudeNormalized, elevation, seaLevel, globalTempMod)

				// Add geothermal contribution from planetary internal heat
				// Early Earth (Hadean): +90°C from mantle heat flux
				// Modern Earth: ~0°C (radiogenic heat depleted)
				temp := baseTemp + geothermalOffset

				// Get longitude for noise variation
				longitude := GetLongitudeFromCoord(topology, coord)

				// Moisture from Perlin noise using lat/lon for consistency across faces
				// Scale coordinates to get reasonable noise frequency
				noiseX := (longitude + 180) / 360.0 * float64(faceSize) * 0.1
				noiseY := (latitude + 90) / 180.0 * float64(faceSize) * 0.1
				n := noise.Noise2D(noiseX, noiseY)
				moisture := (n + 1.0) / 2.0 // Normalize to 0-1

				// Convert moisture to rainfall
				rainfall := moisture * 2000.0

				// Seasonality: higher at poles
				seasonality := latitudeNormalized * 0.8

				// Soil drainage
				drainage := 0.5
				if elevation <= seaLevel {
					drainage = 0.0
				} else {
					altitudeAboveSea := elevation - seaLevel
					drainage = math.Min(1.0, 0.3+altitudeAboveSea/5000.0)
				}

				climateData[face][idx] = ClimateData{
					Temperature:    temp,
					AnnualRainfall: rainfall,
					Seasonality:    seasonality,
					SoilDrainage:   drainage,
				}
			}
		}
	}

	return climateData
}

// GetClimateAtSpherical returns climate data for a specific spherical coordinate.
func GetClimateAtSpherical(climateMap SphereClimateMap, faceSize int, coord spatial.Coordinate) ClimateData {
	if coord.Face < 0 || coord.Face >= len(climateMap) {
		return ClimateData{}
	}

	idx := coord.Y*faceSize + coord.X
	if idx < 0 || idx >= len(climateMap[coord.Face]) {
		return ClimateData{}
	}

	return climateMap[coord.Face][idx]
}

// CalculateWindSpherical computes wind direction and speed for a spherical coordinate.
// Uses the coordinate's true latitude from 3D position and applies Coriolis effect.
func CalculateWindSpherical(topology spatial.Topology, coord spatial.Coordinate, season Season) Wind {
	latitude := GetLatitudeFromCoord(topology, coord)
	longitude := GetLongitudeFromCoord(topology, coord)
	return CalculateWind(latitude, longitude, season)
}

// Get3DWindVector converts latitude-based wind to a 3D world-space vector.
// This is useful for advection across cube-sphere faces.
func Get3DWindVector(topology spatial.Topology, coord spatial.Coordinate, season Season) spatial.Vector3D {
	wind := CalculateWindSpherical(topology, coord, season)

	// Get position on sphere
	px, py, pz := topology.ToSphere(coord)
	pos := spatial.Vector3D{X: px, Y: py, Z: pz}

	// Calculate tangent basis vectors
	// Normal is the position (for unit sphere)
	normal := pos.Normalize()

	// "Up" in tangent space points towards +Y axis, projected onto tangent
	worldUp := spatial.Vector3D{X: 0, Y: 1, Z: 0}
	upDot := worldUp.Dot(normal)
	tangentUp := worldUp.Add(normal.Scale(-upDot)).Normalize()

	// "East" is perpendicular to up and normal (right-hand rule: up × normal = east)
	tangentEast := spatial.Vector3D{
		X: tangentUp.Y*normal.Z - tangentUp.Z*normal.Y,
		Y: tangentUp.Z*normal.X - tangentUp.X*normal.Z,
		Z: tangentUp.X*normal.Y - tangentUp.Y*normal.X,
	}.Normalize()

	// Convert wind direction (degrees) to vector in tangent space
	// 0° = North, 90° = East, etc.
	dirRad := wind.Direction * math.Pi / 180.0
	northComponent := math.Cos(dirRad) * wind.Speed
	eastComponent := math.Sin(dirRad) * wind.Speed

	// Combine tangent vectors scaled by wind components
	windVec := tangentUp.Scale(northComponent).Add(tangentEast.Scale(eastComponent))

	return windVec
}

// OceanTemperatureProvider is an interface for getting ocean temperature data.
// This allows the climate generator to use ocean data without importing the ocean package.
type OceanTemperatureProvider interface {
	// GetAverageOceanTemp returns the average temperature of neighboring ocean cells.
	// Returns (temperature, hasOceanNeighbor).
	GetAverageOceanTemp(coord spatial.Coordinate) (float64, bool)
	// IsOcean returns true if the coordinate is an ocean cell.
	IsOcean(coord spatial.Coordinate) bool
}

// ApplyOceanModeration modifies coastal land temperatures based on nearby ocean temperatures.
// This implements the "Gulf Stream effect" where warm currents heat high-latitude coasts.
//
// Physics:
//   - Coastal land cells have their temperature moderated by neighboring ocean
//   - Formula: LandTemp = (LandTemp * 0.7) + (OceanTemp * 0.3)
//   - Effect: Cold land near warm water heats up; hot land near cold water cools down
func ApplyOceanModeration(
	climateData SphereClimateMap,
	topology spatial.Topology,
	oceanProvider OceanTemperatureProvider,
	sphereMap *geography.SphereHeightmap,
	seaLevel float64,
) {
	faceSize := topology.Resolution()
	directions := []spatial.Direction{
		spatial.North, spatial.South, spatial.East, spatial.West,
	}

	for face := 0; face < 6; face++ {
		for y := 0; y < faceSize; y++ {
			for x := 0; x < faceSize; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				idx := y*faceSize + x

				// Skip ocean cells
				if sphereMap.Get(coord) <= seaLevel {
					continue
				}

				// Check if this land cell has an ocean neighbor
				oceanTemp, hasOceanNeighbor := oceanProvider.GetAverageOceanTemp(coord)
				if !hasOceanNeighbor {
					// Also check if any direct neighbor is ocean
					for _, dir := range directions {
						neighbor := topology.GetNeighbor(coord, dir)
						if oceanProvider.IsOcean(neighbor) {
							hasOceanNeighbor = true
							break
						}
					}
					if !hasOceanNeighbor {
						continue
					}
				}

				// Apply maritime moderation: 70% land influence, 30% ocean influence
				landTemp := climateData[face][idx].Temperature
				moderatedTemp := (landTemp * 0.7) + (oceanTemp * 0.3)

				// Also reduce temperature extremes (seasonality) for coastal areas
				moderatedSeasonality := climateData[face][idx].Seasonality * 0.7

				climateData[face][idx].Temperature = moderatedTemp
				climateData[face][idx].Seasonality = moderatedSeasonality
			}
		}
	}
}

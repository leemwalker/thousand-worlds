package weather

import (
	"testing"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

func TestGenerateInitialClimateSpherical(t *testing.T) {
	const faceSize = 32

	// Create sphere heightmap with varied terrain
	topology := spatial.NewCubeSphereTopology(faceSize)
	sphereMap := geography.NewSphereHeightmap(topology)

	// Set up terrain: create a mountain at top face (polar) and low land at front face (equatorial)
	for face := 0; face < 6; face++ {
		for y := 0; y < faceSize; y++ {
			for x := 0; x < faceSize; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				if face == 4 { // Top face (polar)
					sphereMap.Set(coord, 1000) // 1000m elevation
				} else if face == 5 { // Bottom face (polar)
					sphereMap.Set(coord, 500)
				} else { // Equatorial faces
					sphereMap.Set(coord, 100) // Low elevation
				}
			}
		}
	}

	seaLevel := 0.0
	seed := int64(42)
	globalTempMod := 0.0

	t.Run("generates climate for all face coordinates", func(t *testing.T) {
		climate := GenerateInitialClimateSpherical(sphereMap, topology, seaLevel, seed, globalTempMod, 0.0)

		// Should have data for all 6 faces
		if len(climate) != 6 {
			t.Errorf("Expected 6 faces, got %d", len(climate))
		}

		// Each face should have faceSize*faceSize entries
		for face := 0; face < 6; face++ {
			if len(climate[face]) != faceSize*faceSize {
				t.Errorf("Face %d: expected %d entries, got %d", face, faceSize*faceSize, len(climate[face]))
			}
		}
	})

	t.Run("polar regions are colder than equator", func(t *testing.T) {
		climate := GenerateInitialClimateSpherical(sphereMap, topology, seaLevel, seed, globalTempMod, 0.0)

		// Get average temp at equatorial face (face 0) center
		equatorTemp := climate[0][faceSize/2*faceSize+faceSize/2].Temperature

		// Get average temp at polar face (face 4) center
		polarTemp := climate[4][faceSize/2*faceSize+faceSize/2].Temperature

		// Polar should be colder (accounting for elevation effect too)
		if polarTemp >= equatorTemp {
			t.Errorf("Polar temp (%f) should be colder than equator temp (%f)", polarTemp, equatorTemp)
		}
	})

	t.Run("higher elevation is colder due to lapse rate", func(t *testing.T) {
		// Create a mountain on an equatorial face
		mountainMap := geography.NewSphereHeightmap(topology)
		for face := 0; face < 6; face++ {
			for y := 0; y < faceSize; y++ {
				for x := 0; x < faceSize; x++ {
					coord := spatial.Coordinate{Face: face, X: x, Y: y}
					mountainMap.Set(coord, 100) // Base elevation
				}
			}
		}
		// Add a mountain at center of face 0
		mountainMap.Set(spatial.Coordinate{Face: 0, X: faceSize / 2, Y: faceSize / 2}, 5000)

		climate := GenerateInitialClimateSpherical(mountainMap, topology, seaLevel, seed, globalTempMod, 0.0)

		mountainIdx := faceSize/2*faceSize + faceSize/2
		neighborIdx := faceSize/2*faceSize + faceSize/2 + 1

		mountainTemp := climate[0][mountainIdx].Temperature
		lowlandTemp := climate[0][neighborIdx].Temperature

		// Mountain should be significantly colder
		if mountainTemp >= lowlandTemp-10 {
			t.Errorf("Mountain (%f°C) should be much colder than lowland (%f°C)", mountainTemp, lowlandTemp)
		}
	})

	t.Run("seasonality is higher at poles", func(t *testing.T) {
		climate := GenerateInitialClimateSpherical(sphereMap, topology, seaLevel, seed, globalTempMod, 0.0)

		// Equator center seasonality
		equatorSeasonality := climate[0][faceSize/2*faceSize+faceSize/2].Seasonality

		// Polar center seasonality
		polarSeasonality := climate[4][faceSize/2*faceSize+faceSize/2].Seasonality

		// Poles should have higher seasonality
		if polarSeasonality <= equatorSeasonality {
			t.Errorf("Polar seasonality (%f) should be higher than equator (%f)",
				polarSeasonality, equatorSeasonality)
		}
	})

	t.Run("global temp modifier affects all cells", func(t *testing.T) {
		normalClimate := GenerateInitialClimateSpherical(sphereMap, topology, seaLevel, seed, 0, 0.0)
		coldClimate := GenerateInitialClimateSpherical(sphereMap, topology, seaLevel, seed, -10, 0.0)

		// Sample cell
		normalTemp := normalClimate[0][0].Temperature
		coldTemp := coldClimate[0][0].Temperature

		diff := normalTemp - coldTemp
		if diff < 9 || diff > 11 {
			t.Errorf("Temperature difference should be ~10°C, got %f (normal: %f, cold: %f)",
				diff, normalTemp, coldTemp)
		}
	})
}

func TestGetClimateAtSpherical(t *testing.T) {
	const faceSize = 16
	topology := spatial.NewCubeSphereTopology(faceSize)
	sphereMap := geography.NewSphereHeightmap(topology)

	// Set uniform elevation
	for face := 0; face < 6; face++ {
		for y := 0; y < faceSize; y++ {
			for x := 0; x < faceSize; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				sphereMap.Set(coord, 100)
			}
		}
	}

	climate := GenerateInitialClimateSpherical(sphereMap, topology, 0, 42, 0, 0.0)

	t.Run("returns valid climate for valid coordinate", func(t *testing.T) {
		coord := spatial.Coordinate{Face: 0, X: 5, Y: 5}
		data := GetClimateAtSpherical(climate, faceSize, coord)

		// Should have non-zero temperature (unless very cold)
		// Actually temperature could be any value, just verify structure
		if data.SoilDrainage < 0 || data.SoilDrainage > 1 {
			t.Errorf("SoilDrainage should be 0-1, got %f", data.SoilDrainage)
		}
	})

	t.Run("returns empty climate for invalid face", func(t *testing.T) {
		coord := spatial.Coordinate{Face: 10, X: 5, Y: 5}
		data := GetClimateAtSpherical(climate, faceSize, coord)

		if data.Temperature != 0 && data.AnnualRainfall != 0 {
			t.Errorf("Expected empty climate for invalid face, got %+v", data)
		}
	})
}

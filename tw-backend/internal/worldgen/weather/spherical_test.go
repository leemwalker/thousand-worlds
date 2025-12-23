package weather

import (
	"math"
	"testing"

	"tw-backend/internal/spatial"
)

func TestGetLatitudeFromCoord(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	center := 32

	tests := []struct {
		name       string
		coord      spatial.Coordinate
		wantLatMin float64
		wantLatMax float64
	}{
		// Top face center should be near +90° (North Pole)
		{"top face center", spatial.Coordinate{Face: 4, X: center, Y: center}, 80, 90},
		// Bottom face center should be near -90° (South Pole)
		{"bottom face center", spatial.Coordinate{Face: 5, X: center, Y: center}, -90, -80},
		// Front face center should be near 0° (equator)
		{"front face center", spatial.Coordinate{Face: 0, X: center, Y: center}, -10, 10},
		// Back face center should also be near 0° (equator)
		{"back face center", spatial.Coordinate{Face: 1, X: center, Y: center}, -10, 10},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetLatitudeFromCoord(topo, tc.coord)
			if got < tc.wantLatMin || got > tc.wantLatMax {
				t.Errorf("GetLatitudeFromCoord(%v) = %f, want in range [%f, %f]",
					tc.coord, got, tc.wantLatMin, tc.wantLatMax)
			}
		})
	}
}

func TestGetLongitudeFromCoord(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	center := 32

	tests := []struct {
		name       string
		coord      spatial.Coordinate
		wantLonMin float64
		wantLonMax float64
	}{
		// Front face center should be ~0° longitude
		{"front face center", spatial.Coordinate{Face: 0, X: center, Y: center}, -10, 10},
		// Right face center should be ~90° East
		{"right face center", spatial.Coordinate{Face: 3, X: center, Y: center}, 80, 100},
		// Back face center should be ~180° or -180°
		{"back face center", spatial.Coordinate{Face: 1, X: center, Y: center}, 170, 180},
		// Left face center should be ~-90° (270° or -90°)
		{"left face center", spatial.Coordinate{Face: 2, X: center, Y: center}, -100, -80},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetLongitudeFromCoord(topo, tc.coord)
			// Handle wrap-around for back face
			if tc.name == "back face center" {
				if !(got > 170 || got < -170) {
					t.Errorf("GetLongitudeFromCoord(%v) = %f, want near ±180", tc.coord, got)
				}
			} else if got < tc.wantLonMin || got > tc.wantLonMax {
				t.Errorf("GetLongitudeFromCoord(%v) = %f, want in range [%f, %f]",
					tc.coord, got, tc.wantLonMin, tc.wantLonMax)
			}
		})
	}
}

func TestWindToLocalDirection(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)

	tests := []struct {
		name    string
		coord   spatial.Coordinate
		windVec spatial.Vector3D
		want    spatial.Direction
	}{
		// Wind blowing North (+Y in world space) on front face
		{"north wind on front", spatial.Coordinate{Face: 0, X: 32, Y: 32}, spatial.Vector3D{X: 0, Y: 1, Z: 0}, spatial.North},
		// Wind blowing East (+X in world space) on front face
		{"east wind on front", spatial.Coordinate{Face: 0, X: 32, Y: 32}, spatial.Vector3D{X: 1, Y: 0, Z: 0}, spatial.East},
		// Wind blowing South (-Y in world space) on front face
		{"south wind on front", spatial.Coordinate{Face: 0, X: 32, Y: 32}, spatial.Vector3D{X: 0, Y: -1, Z: 0}, spatial.South},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := WindToLocalDirection(topo, tc.coord, tc.windVec)
			if got != tc.want {
				t.Errorf("WindToLocalDirection(%v, %v) = %v, want %v", tc.coord, tc.windVec, got, tc.want)
			}
		})
	}
}

func TestSimulateAdvectionSpherical(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)

	// Test within-face advection
	t.Run("within face", func(t *testing.T) {
		coord := spatial.Coordinate{Face: 0, X: 32, Y: 32}
		windVec := spatial.Vector3D{X: 1, Y: 0, Z: 0} // East wind
		moisture := 50.0

		newCoord, _, newMoisture := SimulateAdvectionSpherical(topo, coord, windVec, moisture)

		// Should move east (X increases)
		if newCoord.Face != 0 || newCoord.X <= coord.X {
			t.Errorf("Expected eastward movement, got %v", newCoord)
		}
		if newMoisture != moisture {
			t.Errorf("Moisture changed unexpectedly: %f -> %f", moisture, newMoisture)
		}
	})

	// Test cross-face advection (polar crossing)
	t.Run("polar crossing", func(t *testing.T) {
		// North edge of front face
		coord := spatial.Coordinate{Face: 0, X: 32, Y: 0}
		windVec := spatial.Vector3D{X: 0, Y: 1, Z: 0} // North wind
		moisture := 75.0

		newCoord, _, _ := SimulateAdvectionSpherical(topo, coord, windVec, moisture)

		// Should cross to top face (Face 4)
		if newCoord.Face != 4 {
			t.Errorf("Expected face 4 (Top), got face %d", newCoord.Face)
		}
	})
}

func TestGetLatitudeFromCoord_PolarValues(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)

	// Very top of top face should be close to 90°
	topCorner := spatial.Coordinate{Face: 4, X: 32, Y: 0}
	lat := GetLatitudeFromCoord(topo, topCorner)
	if lat < 45 {
		t.Errorf("Top face should have high latitude, got %f", lat)
	}

	// Very bottom of bottom face should be close to -90°
	bottomCorner := spatial.Coordinate{Face: 5, X: 32, Y: 63}
	lat = GetLatitudeFromCoord(topo, bottomCorner)
	if lat > -45 {
		t.Errorf("Bottom face should have low latitude, got %f", lat)
	}
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

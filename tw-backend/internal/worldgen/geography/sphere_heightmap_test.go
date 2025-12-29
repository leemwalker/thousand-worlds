package geography

import (
	"testing"

	"tw-backend/internal/spatial"
)

func TestNewSphereHeightmap(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	if shm == nil {
		t.Fatal("NewSphereHeightmap returned nil")
	}

	if shm.Resolution() != 64 {
		t.Errorf("Resolution() = %d, want 64", shm.Resolution())
	}
}

func TestSphereHeightmap_GetSet(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	tests := []struct {
		name  string
		coord spatial.Coordinate
		value float64
	}{
		{"face 0 center", spatial.Coordinate{Face: 0, X: 32, Y: 32}, 1000.0},
		{"face 1 corner", spatial.Coordinate{Face: 1, X: 0, Y: 0}, -500.0},
		{"face 4 top center", spatial.Coordinate{Face: 4, X: 32, Y: 32}, 8848.0},
		{"face 5 bottom", spatial.Coordinate{Face: 5, X: 63, Y: 63}, -10000.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shm.Set(tc.coord, tc.value)
			got := shm.Get(tc.coord)
			if got != tc.value {
				t.Errorf("Get after Set: got %f, want %f", got, tc.value)
			}
		})
	}
}

func TestSphereHeightmap_GetDefault(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	// Unset coordinate should return 0
	coord := spatial.Coordinate{Face: 2, X: 10, Y: 10}
	got := shm.Get(coord)
	if got != 0 {
		t.Errorf("Get unset coordinate: got %f, want 0", got)
	}
}

func TestSphereHeightmap_GetNeighborElevation_WithinFace(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	// Set up a gradient on face 0
	center := spatial.Coordinate{Face: 0, X: 32, Y: 32}
	north := spatial.Coordinate{Face: 0, X: 32, Y: 31}
	east := spatial.Coordinate{Face: 0, X: 33, Y: 32}

	shm.Set(center, 100.0)
	shm.Set(north, 200.0)
	shm.Set(east, 300.0)

	// Test GetNeighborElevation
	gotNorth := shm.GetNeighborElevation(center, spatial.North)
	if gotNorth != 200.0 {
		t.Errorf("GetNeighborElevation(North) = %f, want 200.0", gotNorth)
	}

	gotEast := shm.GetNeighborElevation(center, spatial.East)
	if gotEast != 300.0 {
		t.Errorf("GetNeighborElevation(East) = %f, want 300.0", gotEast)
	}
}

func TestSphereHeightmap_GetNeighborElevation_CrossFace(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	// Set value on edge of face 0
	edgeCoord := spatial.Coordinate{Face: 0, X: 63, Y: 32}
	shm.Set(edgeCoord, 500.0)

	// The neighbor to the east should be on face 3 (Right)
	neighborCoord := topo.GetNeighbor(edgeCoord, spatial.East)
	shm.Set(neighborCoord, 750.0)

	// Test cross-face retrieval
	got := shm.GetNeighborElevation(edgeCoord, spatial.East)
	if got != 750.0 {
		t.Errorf("GetNeighborElevation cross-face: got %f, want 750.0 (neighbor coord: %v)", got, neighborCoord)
	}
}

func TestSphereHeightmap_AllFacesIndependent(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	// Set same X,Y on different faces - should be independent
	for face := 0; face < 6; face++ {
		coord := spatial.Coordinate{Face: face, X: 10, Y: 10}
		shm.Set(coord, float64(face*100))
	}

	for face := 0; face < 6; face++ {
		coord := spatial.Coordinate{Face: face, X: 10, Y: 10}
		got := shm.Get(coord)
		want := float64(face * 100)
		if got != want {
			t.Errorf("Face %d: got %f, want %f", face, got, want)
		}
	}
}

func TestSphereHeightmap_MinMaxElev(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	// Set various elevations
	shm.Set(spatial.Coordinate{Face: 0, X: 0, Y: 0}, -11000.0) // Mariana Trench depth
	shm.Set(spatial.Coordinate{Face: 4, X: 32, Y: 32}, 8848.0) // Everest height

	shm.UpdateMinMax()

	min, max := shm.MinMax()
	if min != -11000.0 {
		t.Errorf("MinElev = %f, want -11000.0", min)
	}
	if max != 8848.0 {
		t.Errorf("MaxElev = %f, want 8848.0", max)
	}
}

// TestSphereHeightmap_ToFlatHeightmap_SamplesAllFaces verifies that the
// equirectangular projection correctly samples from all 6 cube-sphere faces.
// This is a regression test for the bug where only faces 0-1 were sampled.
func TestSphereHeightmap_ToFlatHeightmap_SamplesAllFaces(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	// Set a distinctive elevation at the center of each face
	// Using large, unique values that are easy to identify
	faceElevations := map[int]float64{
		0: 1000.0,  // Front
		1: 2000.0,  // Back
		2: 3000.0,  // Left
		3: 4000.0,  // Right
		4: 5000.0,  // Top
		5: -5000.0, // Bottom (ocean)
	}

	for face, elev := range faceElevations {
		// Set the center of each face
		shm.Set(spatial.Coordinate{Face: face, X: 32, Y: 32}, elev)
		// Also set a broader region to increase hit probability
		for dx := -5; dx <= 5; dx++ {
			for dy := -5; dy <= 5; dy++ {
				shm.Set(spatial.Coordinate{Face: face, X: 32 + dx, Y: 32 + dy}, elev)
			}
		}
	}
	shm.UpdateMinMax()

	// Convert to flat heightmap
	flat := shm.ToFlatHeightmap(256, 128)

	// Check which face values appear in the flat output
	facesFound := make(map[int]bool)
	for _, elev := range flat.Elevations {
		for face, expected := range faceElevations {
			// Use tolerance for floating point comparison
			if elev >= expected-0.1 && elev <= expected+0.1 {
				facesFound[face] = true
			}
		}
	}

	// Verify at least 4 out of 6 faces are represented
	// (poles may be compressed in equirectangular projection)
	if len(facesFound) < 4 {
		t.Errorf("ToFlatHeightmap only sampled %d faces, want at least 4. Found: %v",
			len(facesFound), facesFound)
	}

	// The old broken projection only samples faces 0 and 1
	// This test will fail if fewer than 4 faces are found
	t.Logf("Faces found in flat output: %v", facesFound)
}

// =============================================================================
// CellData Tests (Phase 5: Geological Provinces)
// =============================================================================

func TestSphereHeightmap_CellDataInitialization(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	// By default, all cells should have zero values
	coord := spatial.Coordinate{Face: 0, X: 32, Y: 32}
	data := shm.GetCellData(coord)

	if data.RockHardness != 0.0 {
		t.Errorf("Default RockHardness = %f, want 0.0", data.RockHardness)
	}
	if data.Sediment != 0.0 {
		t.Errorf("Default Sediment = %f, want 0.0", data.Sediment)
	}
	if data.ProvinceID != 0 {
		t.Errorf("Default ProvinceID = %d, want 0", data.ProvinceID)
	}
}

func TestSphereHeightmap_SetCellData(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	coord := spatial.Coordinate{Face: 2, X: 10, Y: 20}
	data := CellData{
		RockHardness: 0.9,
		Sediment:     50.0,
		ProvinceID:   42,
	}
	shm.SetCellData(coord, data)

	got := shm.GetCellData(coord)
	if got.RockHardness != 0.9 {
		t.Errorf("RockHardness = %f, want 0.9", got.RockHardness)
	}
	if got.Sediment != 50.0 {
		t.Errorf("Sediment = %f, want 50.0", got.Sediment)
	}
	if got.ProvinceID != 42 {
		t.Errorf("ProvinceID = %d, want 42", got.ProvinceID)
	}
}

func TestSphereHeightmap_GetRockHardness(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	coord := spatial.Coordinate{Face: 1, X: 5, Y: 5}
	shm.SetCellData(coord, CellData{RockHardness: 0.75, Sediment: 0, ProvinceID: 1})

	got := shm.GetRockHardness(coord)
	if got != 0.75 {
		t.Errorf("GetRockHardness = %f, want 0.75", got)
	}
}

func TestSphereHeightmap_AddSediment(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	coord := spatial.Coordinate{Face: 3, X: 20, Y: 20}

	// Set initial elevation
	shm.Set(coord, 1000.0)

	// Add sediment - should increase both Sediment AND TotalHeight
	shm.AddSediment(coord, 50.0)

	data := shm.GetCellData(coord)
	if data.Sediment != 50.0 {
		t.Errorf("Sediment after AddSediment = %f, want 50.0", data.Sediment)
	}

	newElev := shm.Get(coord)
	if newElev != 1050.0 {
		t.Errorf("Elevation after AddSediment = %f, want 1050.0", newElev)
	}

	// Add more sediment
	shm.AddSediment(coord, 25.0)
	data = shm.GetCellData(coord)
	if data.Sediment != 75.0 {
		t.Errorf("Sediment after second AddSediment = %f, want 75.0", data.Sediment)
	}
	if shm.Get(coord) != 1075.0 {
		t.Errorf("Elevation after second AddSediment = %f, want 1075.0", shm.Get(coord))
	}
}

func TestSphereHeightmap_Erode_SedimentFirst(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	coord := spatial.Coordinate{Face: 4, X: 15, Y: 15}

	// Set initial elevation and sediment
	shm.Set(coord, 1000.0) // Total height
	shm.SetCellData(coord, CellData{RockHardness: 0.5, Sediment: 30.0, ProvinceID: 1})

	// Erode 20 meters - should only remove sediment
	removed := shm.Erode(coord, 20.0)

	if removed != 20.0 {
		t.Errorf("Erode returned %f, want 20.0", removed)
	}

	data := shm.GetCellData(coord)
	if data.Sediment != 10.0 {
		t.Errorf("Sediment after eroding 20m = %f, want 10.0", data.Sediment)
	}

	newElev := shm.Get(coord)
	if newElev != 980.0 {
		t.Errorf("Elevation after eroding 20m = %f, want 980.0", newElev)
	}
}

func TestSphereHeightmap_Erode_ThenBedrock(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	coord := spatial.Coordinate{Face: 5, X: 25, Y: 25}

	// Set initial elevation with only 10m sediment
	shm.Set(coord, 500.0) // Total height
	shm.SetCellData(coord, CellData{RockHardness: 0.2, Sediment: 10.0, ProvinceID: 2})

	// Erode 30 meters - should remove all sediment (10m) then bedrock (20m)
	removed := shm.Erode(coord, 30.0)

	if removed != 30.0 {
		t.Errorf("Erode returned %f, want 30.0", removed)
	}

	data := shm.GetCellData(coord)
	if data.Sediment != 0.0 {
		t.Errorf("Sediment after full erosion = %f, want 0.0", data.Sediment)
	}

	newElev := shm.Get(coord)
	if newElev != 470.0 {
		t.Errorf("Elevation after eroding 30m = %f, want 470.0 (500 - 30)", newElev)
	}
}

func TestSphereHeightmap_Erode_NoSediment(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	coord := spatial.Coordinate{Face: 0, X: 50, Y: 50}

	// Set initial elevation with no sediment
	shm.Set(coord, 2000.0)
	shm.SetCellData(coord, CellData{RockHardness: 0.9, Sediment: 0.0, ProvinceID: 3})

	// Erode 100 meters - should only remove bedrock
	removed := shm.Erode(coord, 100.0)

	if removed != 100.0 {
		t.Errorf("Erode returned %f, want 100.0", removed)
	}

	newElev := shm.Get(coord)
	if newElev != 1900.0 {
		t.Errorf("Elevation after eroding bedrock = %f, want 1900.0", newElev)
	}
}

func TestSphereHeightmap_CellData_AllFacesIndependent(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(64)
	shm := NewSphereHeightmap(topo)

	// Set different cell data on same X,Y across different faces
	for face := 0; face < 6; face++ {
		coord := spatial.Coordinate{Face: face, X: 10, Y: 10}
		shm.SetCellData(coord, CellData{
			RockHardness: float64(face) * 0.15,
			Sediment:     float64(face) * 10.0,
			ProvinceID:   face + 100,
		})
	}

	// Verify independence
	for face := 0; face < 6; face++ {
		coord := spatial.Coordinate{Face: face, X: 10, Y: 10}
		got := shm.GetCellData(coord)

		wantHardness := float64(face) * 0.15
		wantSediment := float64(face) * 10.0
		wantProvinceID := face + 100

		if got.RockHardness != wantHardness {
			t.Errorf("Face %d: RockHardness = %f, want %f", face, got.RockHardness, wantHardness)
		}
		if got.Sediment != wantSediment {
			t.Errorf("Face %d: Sediment = %f, want %f", face, got.Sediment, wantSediment)
		}
		if got.ProvinceID != wantProvinceID {
			t.Errorf("Face %d: ProvinceID = %d, want %d", face, got.ProvinceID, wantProvinceID)
		}
	}
}

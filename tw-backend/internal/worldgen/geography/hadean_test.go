package geography

import (
	"testing"

	"tw-backend/internal/spatial"
)

// TestHadeanGrowth verifies that starting from a "Water World" (0% continental crust),
// the simulation naturally accretes continental crust over time via subduction.
func TestHadeanGrowth(t *testing.T) {
	// 1. Setup Hadean Earth (0% Continental)
	res := 16
	topology := spatial.NewCubeSphereTopology(res)
	seed := int64(999) // Fixed seed

	// Create 10 plates, all Oceanic
	plates := GeneratePlates(10, topology, seed, 0.0)

	// Confirm initial state
	for _, p := range plates {
		if p.Type == PlateContinental {
			t.Fatal("Expected all plates to be Oceanic initially")
		}
	}
	t.Log("Initialized Hadean Earth: 10 Oceanic Plates")

	// Initialize Heightmap (Deep Ocean everywhere)
	shm := NewSphereHeightmap(topology)
	shm = GenerateHeightmap(plates, shm, topology, seed, 1.0, 1.0)

	// Double check no continental cells
	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				cell := shm.GetCellData(spatial.Coordinate{Face: face, X: x, Y: y})
				if cell.IsContinental {
					t.Fatal("Found continental cell in Hadean initialization")
				}
			}
		}
	}

	// 2. Simulate Deep Time (e.g., 50 Ticks ~ 50-100 Million Years)
	// We need enough time for plates to move, collide, and accumulate mass
	ticks := 50
	dt := 1.0 // 1 Million Years per tick

	// Initialize plate motion if not already (GeneratePlates calls it, but re-calling is safe/random check)
	// Actually GeneratePlates(..., 0.0) calls InitializePlateMotion internally.
	// But let's ensure velocities are non-zero.

	continentalCellsBytes := 0
	_ = continentalCellsBytes

	for i := 0; i < ticks; i++ {
		// A. Move Plates
		UpdatePlatePositions(plates, dt, topology)

		// B. Reassign Regions (expensive but necessary for collision detection)
		ReassignPlateRegions(plates, topology)

		// C. Compute Cache
		cache := ComputeBoundaryCache(plates, topology)

		// D. Simulate Tectonics (Accretion happens here)
		// scaleFactor represents time duration for flux
		_ = SimulateTectonicsWithCache(plates, shm, cache, topology, dt)

		// Optional: Log progress every 10 ticks
		if i%10 == 0 {
			totalAccreted := 0.0
			for _, p := range plates {
				totalAccreted += p.AccretedMass
			}
			t.Logf("Tick %d: Total Accreted Mass = %.2f", i, totalAccreted)
		}
	}

	// 3. Verify Growth
	// Count continental cells
	continentalCells := 0
	totalCells := 6 * res * res

	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				cell := shm.GetCellData(spatial.Coordinate{Face: face, X: x, Y: y})
				if cell.IsContinental {
					continentalCells++
				}
			}
		}
	}

	growthPerc := float64(continentalCells) / float64(totalCells) * 100.0
	t.Logf("Final State: %d Continental Cells (%.2f%% coverage)", continentalCells, growthPerc)

	if continentalCells == 0 {
		t.Error("Failed to grow any continental crust after simulation")
	}

	// Check plate stats
	totalAccreted := 0.0
	for _, p := range plates {
		totalAccreted += p.AccretedMass
	}
	if totalAccreted <= 0 {
		t.Error("No mass accreted on plates")
	}
}

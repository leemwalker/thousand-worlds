package geography

import (
	"tw-backend/internal/spatial"
)

// SphereHeightmap wraps 6 flat Heightmaps into a spherical surface
// using the cube-sphere topology for neighbor lookups.
type SphereHeightmap struct {
	topology spatial.Topology
	faces    [6]*Heightmap
	MinElev  float64
	MaxElev  float64
}

// NewSphereHeightmap creates a new spherical heightmap using the given topology.
// Each face is initialized with a flat Heightmap of size Resolution x Resolution.
func NewSphereHeightmap(topology spatial.Topology) *SphereHeightmap {
	res := topology.Resolution()
	shm := &SphereHeightmap{
		topology: topology,
	}

	for i := 0; i < 6; i++ {
		shm.faces[i] = NewHeightmap(res, res)
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
	s.faces[coord.Face].Set(coord.X, coord.Y, val)
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
// Uses a simple face-wrapping projection for compatibility with legacy systems.
// width and height specify the dimensions of the output heightmap.
func (s *SphereHeightmap) ToFlatHeightmap(width, height int) *Heightmap {
	flat := NewHeightmap(width, height)
	resolution := s.topology.Resolution()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map flat coordinates to sphere coordinate
			// Use modulo to wrap around the face grid
			face := (x / resolution) % 6
			fx := x % resolution
			fy := y % resolution

			if fx >= resolution {
				fx = resolution - 1
			}
			if fy >= resolution {
				fy = resolution - 1
			}

			coord := spatial.Coordinate{Face: face, X: fx, Y: fy}
			elev := s.Get(coord)
			flat.Set(x, y, elev)
		}
	}

	flat.MinElev = s.MinElev
	flat.MaxElev = s.MaxElev

	return flat
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

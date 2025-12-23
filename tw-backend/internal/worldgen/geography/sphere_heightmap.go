package geography

import (
	"tw-backend/internal/spatial"
)

// SphereHeightmap wraps 6 flat Heightmaps into a spherical surface
// using the cube-sphere topology for neighbor lookups.
type SphereHeightmap struct {
	topology spatial.Topology
	faces    [6]*Heightmap
	minElev  float64
	maxElev  float64
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
				s.minElev = elev
				s.maxElev = elev
				first = false
			} else {
				if elev < s.minElev {
					s.minElev = elev
				}
				if elev > s.maxElev {
					s.maxElev = elev
				}
			}
		}
	}
}

// MinMax returns the minimum and maximum elevations
func (s *SphereHeightmap) MinMax() (min, max float64) {
	return s.minElev, s.maxElev
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

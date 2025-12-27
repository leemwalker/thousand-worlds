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
func (s *SphereHeightmap) ToFlatHeightmapInPlace(dest *Heightmap) {
	width := dest.Width
	height := dest.Height

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map pixel coordinates to longitude and latitude
			// Longitude: 0 to 2π (left to right)
			// Latitude: π/2 to -π/2 (top to bottom, north pole to south pole)
			lon := (float64(x) / float64(width)) * 2 * 3.141592653589793  // 0 to 2π
			lat := (0.5 - float64(y)/float64(height)) * 3.141592653589793 // π/2 to -π/2

			// Convert lat/lon to 3D unit sphere coordinates
			// Standard spherical coordinate conversion:
			// x = cos(lat) * cos(lon)
			// y = sin(lat)           (Y is up/down axis)
			// z = cos(lat) * sin(lon)
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

	dest.MinElev = s.MinElev
	dest.MaxElev = s.MaxElev
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

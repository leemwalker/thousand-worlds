package gamemap

// CubeFace represents one of the six faces of a cube map projection.
type CubeFace int

const (
	FaceFront CubeFace = iota
	FaceBack
	FaceLeft
	FaceRight
	FaceTop
	FaceBottom
)

// String returns a human-readable name for the cube face.
func (f CubeFace) String() string {
	names := []string{"Front", "Back", "Left", "Right", "Top", "Bottom"}
	if int(f) >= 0 && int(f) < len(names) {
		return names[f]
	}
	return "Unknown"
}

// TileCoord identifies a tile in the quadtree pyramid.
type TileCoord struct {
	Face  CubeFace
	Level int // 0 = 1 tile per face, N = 4^N tiles per face
	X     int // Column within level grid
	Y     int // Row within level grid
}

// TileData contains the rendered tile image and heightmap.
type TileData struct {
	Coord     TileCoord
	Image     []byte // WebP or PNG encoded texture
	Heightmap []byte // 16-bit grayscale PNG for displacement
	Width     int
	Height    int
}

// TileRequest is used to request a specific tile from the renderer.
type TileRequest struct {
	Face  CubeFace
	Level int
	X     int
	Y     int
	Size  int // Pixel dimensions (e.g., 256, 512)
}

// TilesPerSide returns the number of tiles per side at a given level.
// Level 0: 1, Level 1: 2, Level 2: 4, Level N: 2^N
func TilesPerSide(level int) int {
	if level < 0 {
		return 1
	}
	return 1 << level // 2^level
}

// TotalTiles returns the total number of tiles at a given level (for one face).
// Level 0: 1, Level 1: 4, Level 2: 16, Level N: 4^N
func TotalTiles(level int) int {
	n := TilesPerSide(level)
	return n * n
}

// RawTileData contains the raw float/byte data for a tile, before serialization/compression.
type RawTileData struct {
	Coord     TileCoord
	Heightmap []float32 // Row-major height values
	Biomes    []uint8   // Row-major biome IDs
	Water     []float32 // Row-major water levels (optional, can be inferred if uniform)
	Width     int
	Height    int
}

package gamemap

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"

	"tw-backend/internal/ecosystem"
	"tw-backend/internal/spatial"
)

// TileRenderer generates tile images for cube-projected globe rendering.
type TileRenderer struct {
	config RenderConfig
}

// NewTileRenderer creates a new tile renderer with the given configuration.
func NewTileRenderer(config RenderConfig) *TileRenderer {
	return &TileRenderer{
		config: config,
	}
}

// RenderTile generates a single tile image and heightmap for the given tile request.
// If worldGeo is nil, generates a placeholder/test tile.
func (r *TileRenderer) RenderTile(ctx context.Context, req TileRequest, worldGeo *ecosystem.WorldGeology) (*TileData, error) {
	// Validate tile coordinates
	if err := r.validateRequest(req); err != nil {
		return nil, err
	}

	size := req.Size
	if size <= 0 {
		size = 256
	}

	// Generate tile image
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	heightImg := image.NewGray16(image.Rect(0, 0, size, size))

	tilesPerSide := TilesPerSide(req.Level)
	tileSize := 1.0 / float64(tilesPerSide)

	// UV bounds for this tile within the cube face
	uMin := float64(req.X) * tileSize
	vMin := float64(req.Y) * tileSize

	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			// Map pixel to UV within tile (0-1)
			u := uMin + (float64(px)+0.5)/float64(size)*tileSize
			v := vMin + (float64(py)+0.5)/float64(size)*tileSize

			// Convert cube UV to spherical lat/lon
			lat, lon := CubeToSphere(req.Face, u, v)

			// Get elevation and color from world data (or placeholder)
			var elevation float64
			var c color.RGBA

			if worldGeo != nil && worldGeo.SphereHeightmap != nil {
				elevation = r.sampleElevation(worldGeo, req.Face, u, v)
				c = r.sampleColor(worldGeo, lat, lon, elevation)
			} else {
				// Placeholder: checkerboard pattern with elevation gradient
				elevation = r.placeholderElevation(lat, lon)
				c = r.placeholderColor(req.Face, u, v, elevation)
			}

			img.Set(px, py, c)

			// Encode heightmap as 16-bit grayscale (0 = min, 65535 = max)
			normalizedH := (elevation + 11000) / (8848 + 11000) // Normalize -11km to +8.8km
			normalizedH = math.Max(0, math.Min(1, normalizedH))
			heightImg.SetGray16(px, py, color.Gray16{Y: uint16(normalizedH * 65535)})
		}
	}

	// Encode images
	var imgBuf, heightBuf bytes.Buffer

	if err := png.Encode(&imgBuf, img); err != nil {
		return nil, fmt.Errorf("failed to encode tile image: %w", err)
	}
	if err := png.Encode(&heightBuf, heightImg); err != nil {
		return nil, fmt.Errorf("failed to encode heightmap: %w", err)
	}

	return &TileData{
		Coord: TileCoord{
			Face:  req.Face,
			Level: req.Level,
			X:     req.X,
			Y:     req.Y,
		},
		Image:     imgBuf.Bytes(),
		Heightmap: heightBuf.Bytes(),
		Width:     size,
		Height:    size,
	}, nil
}

// validateRequest checks if tile coordinates are valid for the given level.
func (r *TileRenderer) validateRequest(req TileRequest) error {
	if req.Level < 0 {
		return fmt.Errorf("invalid level: %d (must be >= 0)", req.Level)
	}

	tilesPerSide := TilesPerSide(req.Level)

	if req.X < 0 || req.X >= tilesPerSide {
		return fmt.Errorf("X coordinate %d out of range [0, %d) for level %d", req.X, tilesPerSide, req.Level)
	}
	if req.Y < 0 || req.Y >= tilesPerSide {
		return fmt.Errorf("Y coordinate %d out of range [0, %d) for level %d", req.Y, tilesPerSide, req.Level)
	}

	if req.Face < FaceFront || req.Face > FaceBottom {
		return fmt.Errorf("invalid cube face: %d", req.Face)
	}

	return nil
}

// CubeToSphere converts cube face UV coordinates to spherical lat/lon.
// u, v are in range [0, 1] within the face.
// Returns lat in [-90, 90] and lon in [-180, 180].
func CubeToSphere(face CubeFace, u, v float64) (lat, lon float64) {
	// Convert UV to [-1, 1] range for cube coordinates
	x := 2*u - 1
	y := 2*v - 1

	// Map cube face coordinates to 3D unit cube position
	var cx, cy, cz float64
	switch face {
	case FaceFront:
		cx, cy, cz = x, -y, 1 // Z+
	case FaceBack:
		cx, cy, cz = -x, -y, -1 // Z-
	case FaceLeft:
		cx, cy, cz = -1, -y, -x // X-
	case FaceRight:
		cx, cy, cz = 1, -y, x // X+
	case FaceTop:
		cx, cy, cz = x, 1, y // Y+
	case FaceBottom:
		cx, cy, cz = x, -1, -y // Y-
	}

	// Normalize to unit sphere
	length := math.Sqrt(cx*cx + cy*cy + cz*cz)
	cx /= length
	cy /= length
	cz /= length

	// Convert to spherical coordinates
	lat = math.Asin(cy) * 180 / math.Pi
	lon = math.Atan2(cz, cx) * 180 / math.Pi

	return lat, lon
}

// sampleElevation gets elevation from world geology data using cube face coordinates.
func (r *TileRenderer) sampleElevation(geo *ecosystem.WorldGeology, face CubeFace, u, v float64) float64 {
	if geo == nil || geo.SphereHeightmap == nil {
		return 0
	}

	res := geo.SphereHeightmap.Resolution()
	if res <= 0 {
		return 0
	}

	// Convert UV (0-1) to grid coordinates
	x := int(u * float64(res))
	y := int(v * float64(res))

	// Clamp to valid range
	if x < 0 {
		x = 0
	}
	if x >= res {
		x = res - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= res {
		y = res - 1
	}

	// Use spatial.Coordinate to access the sphere heightmap
	coord := spatial.Coordinate{
		Face: int(face),
		X:    x,
		Y:    y,
	}

	return geo.SphereHeightmap.Get(coord)
}

// sampleColor gets color from world geology data based on elevation and biome.
func (r *TileRenderer) sampleColor(geo *ecosystem.WorldGeology, lat, lon, elevation float64) color.RGBA {
	// Use the same color logic as renderer.go (simplified)
	seaLevel := geo.SeaLevel

	if elevation < seaLevel {
		// Water
		depth := seaLevel - elevation
		if depth > 4000 {
			return color.RGBA{10, 30, 80, 255} // Deep ocean
		}
		return color.RGBA{30, 80, 140, 255} // Shallow ocean
	}

	// Land - elevation-based coloring
	heightAboveSea := elevation - seaLevel
	if heightAboveSea > 4000 {
		return color.RGBA{255, 255, 255, 255} // Snow
	}
	if heightAboveSea > 2000 {
		return color.RGBA{139, 137, 137, 255} // Rock
	}
	if heightAboveSea > 500 {
		return color.RGBA{34, 139, 34, 255} // Forest
	}
	return color.RGBA{124, 252, 0, 255} // Lowland
}

// placeholderElevation generates test elevation based on lat/lon.
func (r *TileRenderer) placeholderElevation(lat, lon float64) float64 {
	// Simple sine wave pattern for testing
	return math.Sin(lat*math.Pi/30)*math.Cos(lon*math.Pi/30)*4000 + 1000
}

// placeholderColor generates a distinctive test pattern per face.
func (r *TileRenderer) placeholderColor(face CubeFace, u, v, elevation float64) color.RGBA {
	// Base color per face
	faceColors := []color.RGBA{
		{255, 100, 100, 255}, // Front - Red
		{100, 255, 100, 255}, // Back - Green
		{100, 100, 255, 255}, // Left - Blue
		{255, 255, 100, 255}, // Right - Yellow
		{255, 100, 255, 255}, // Top - Magenta
		{100, 255, 255, 255}, // Bottom - Cyan
	}

	base := faceColors[int(face)%len(faceColors)]

	// Checkerboard overlay
	gridX := int(u * 8)
	gridY := int(v * 8)
	if (gridX+gridY)%2 == 0 {
		base.R = uint8(float64(base.R) * 0.8)
		base.G = uint8(float64(base.G) * 0.8)
		base.B = uint8(float64(base.B) * 0.8)
	}

	// Elevation shading
	shade := (elevation + 11000) / 20000
	base.R = uint8(math.Min(255, float64(base.R)*shade*1.5))
	base.G = uint8(math.Min(255, float64(base.G)*shade*1.5))
	base.B = uint8(math.Min(255, float64(base.B)*shade*1.5))

	return base
}

// RenderRawTile generates a single tile's raw data for WebGPU rendering.
// Returns heightmap and biome data as flat arrays.
func (r *TileRenderer) RenderRawTile(ctx context.Context, req TileRequest, worldGeo *ecosystem.WorldGeology) (*RawTileData, error) {
	// Validate tile coordinates
	if err := r.validateRequest(req); err != nil {
		return nil, err
	}

	size := req.Size
	if size <= 0 {
		return nil, fmt.Errorf("invalid size %d", size)
	}

	// Prepare data arrays
	heightmap := make([]float32, size*size)
	biomes := make([]uint8, size*size)
	water := make([]float32, size*size)

	tilesPerSide := TilesPerSide(req.Level)
	tileSize := 1.0 / float64(tilesPerSide)

	// UV bounds for this tile within the cube face
	uMin := float64(req.X) * tileSize
	vMin := float64(req.Y) * tileSize

	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			// Map pixel to UV within tile (0-1)
			u := uMin + (float64(px)+0.5)/float64(size)*tileSize
			v := vMin + (float64(py)+0.5)/float64(size)*tileSize

			// Convert cube UV to spherical lat/lon
			// lat, lon := CubeToSphere(req.Face, u, v)

			// Get elevation and data from world geology
			var elevation float64
			var biomeID uint8
			var waterLevel float64

			if worldGeo != nil && worldGeo.SphereHeightmap != nil {
				elevation = r.sampleElevation(worldGeo, req.Face, u, v)
				waterLevel = worldGeo.SeaLevel
				// Sample biome
				// Note: accessing Biomes via sphere mapping would need a helper
				// For now, we'll placeholder biome ID based on elevation/lat/lon if map is missing
				// TODO: Add direct biome sampling to WorldGeology
				biomeID = 0 // Default
			} else {
				// Placeholder
				elevation = 0
				waterLevel = 0
			}

			// Flatten index: row-major (y * width + x)
			idx := py*size + px
			heightmap[idx] = float32(elevation)
			biomes[idx] = biomeID
			water[idx] = float32(waterLevel)
		}
	}

	return &RawTileData{
		Coord: TileCoord{
			Face:  req.Face,
			Level: req.Level,
			X:     req.X,
			Y:     req.Y,
		},
		Heightmap: heightmap,
		Biomes:    biomes,
		Water:     water,
		Width:     size,
		Height:    size,
	}, nil
}

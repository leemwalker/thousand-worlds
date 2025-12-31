package gamemap

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
)

// Binary Grid Format (Sprint 2)
//
// Header: 9 bytes
//   - Magic:    "WMAP" (4 bytes)
//   - Version:  1 (uint8)
//   - Width:    uint16 little-endian
//   - Height:   uint16 little-endian
//
// Elevation Section: width * height * 4 bytes
//   - float32 little-endian for each cell
//
// Biome Section: width * height * 1 byte
//   - uint8 BiomeID for each cell

const (
	binaryGridMagic   = "WMAP"
	binaryGridVersion = 1
	headerSize        = 9 // 4 (magic) + 1 (version) + 2 (width) + 2 (height)
)

// BiomeID is a compact uint8 representation of biome types
type BiomeID uint8

// BiomeID constants - maps to geography.BiomeType strings
const (
	BiomeIDUnknown BiomeID = iota
	BiomeIDOcean
	BiomeIDLowland
	BiomeIDHighland
	BiomeIDMountain
	BiomeIDHighMountain
	BiomeIDRainforest
	BiomeIDDesert
	BiomeIDGrassland
	BiomeIDDeciduousForest
	BiomeIDTaiga
	BiomeIDTundra
	BiomeIDAlpine
	BiomeIDLake
	BiomeIDWetland
)

// biomeStringToID maps biome type strings to compact IDs
var biomeStringToIDMap = map[string]BiomeID{
	"Ocean":            BiomeIDOcean,
	"Lowland":          BiomeIDLowland,
	"Highland":         BiomeIDHighland,
	"Mountain":         BiomeIDMountain,
	"High Mountain":    BiomeIDHighMountain,
	"Rainforest":       BiomeIDRainforest,
	"Desert":           BiomeIDDesert,
	"Grassland":        BiomeIDGrassland,
	"Deciduous Forest": BiomeIDDeciduousForest,
	"Taiga":            BiomeIDTaiga,
	"Tundra":           BiomeIDTundra,
	"Alpine":           BiomeIDAlpine,
	"Lake":             BiomeIDLake,
	"Wetland":          BiomeIDWetland,
}

// biomeIDToString maps compact IDs back to strings
var biomeIDToStringMap = map[BiomeID]string{
	BiomeIDUnknown:         "Unknown",
	BiomeIDOcean:           "Ocean",
	BiomeIDLowland:         "Lowland",
	BiomeIDHighland:        "Highland",
	BiomeIDMountain:        "Mountain",
	BiomeIDHighMountain:    "High Mountain",
	BiomeIDRainforest:      "Rainforest",
	BiomeIDDesert:          "Desert",
	BiomeIDGrassland:       "Grassland",
	BiomeIDDeciduousForest: "Deciduous Forest",
	BiomeIDTaiga:           "Taiga",
	BiomeIDTundra:          "Tundra",
	BiomeIDAlpine:          "Alpine",
	BiomeIDLake:            "Lake",
	BiomeIDWetland:         "Wetland",
}

// BiomeStringToID converts a biome type string to a compact BiomeID
func BiomeStringToID(biomeType string) BiomeID {
	if id, ok := biomeStringToIDMap[biomeType]; ok {
		return id
	}
	return BiomeIDUnknown
}

// BiomeIDToString converts a BiomeID back to a string
func BiomeIDToString(id BiomeID) string {
	if s, ok := biomeIDToStringMap[id]; ok {
		return s
	}
	return "Unknown"
}

// BinaryGrid holds a compact binary representation of map data
type BinaryGrid struct {
	width     int
	height    int
	elevation []float32 // width * height float32 values
	biome     []BiomeID // width * height biome IDs
}

// NewBinaryGrid creates a new binary grid with the given dimensions
func NewBinaryGrid(width, height int) *BinaryGrid {
	size := width * height
	return &BinaryGrid{
		width:     width,
		height:    height,
		elevation: make([]float32, size),
		biome:     make([]BiomeID, size),
	}
}

// Width returns the grid width
func (g *BinaryGrid) Width() int {
	return g.width
}

// Height returns the grid height
func (g *BinaryGrid) Height() int {
	return g.height
}

// SetElevation sets the elevation at the given grid coordinates
func (g *BinaryGrid) SetElevation(x, y int, elev float64) {
	if x >= 0 && x < g.width && y >= 0 && y < g.height {
		g.elevation[y*g.width+x] = float32(elev)
	}
}

// GetElevation returns the elevation at the given grid coordinates
func (g *BinaryGrid) GetElevation(x, y int) float64 {
	if x >= 0 && x < g.width && y >= 0 && y < g.height {
		return float64(g.elevation[y*g.width+x])
	}
	return 0
}

// SetBiome sets the biome ID at the given grid coordinates
func (g *BinaryGrid) SetBiome(x, y int, biome BiomeID) {
	if x >= 0 && x < g.width && y >= 0 && y < g.height {
		g.biome[y*g.width+x] = biome
	}
}

// GetBiome returns the biome ID at the given grid coordinates
func (g *BinaryGrid) GetBiome(x, y int) BiomeID {
	if x >= 0 && x < g.width && y >= 0 && y < g.height {
		return g.biome[y*g.width+x]
	}
	return BiomeIDUnknown
}

// Serialize converts the grid to a binary byte slice
func (g *BinaryGrid) Serialize() []byte {
	size := g.width * g.height
	totalSize := headerSize + size*4 + size // header + elevation (float32) + biome (uint8)

	buf := bytes.NewBuffer(make([]byte, 0, totalSize))

	// Write header
	buf.WriteString(binaryGridMagic)
	buf.WriteByte(binaryGridVersion)
	binary.Write(buf, binary.LittleEndian, uint16(g.width))
	binary.Write(buf, binary.LittleEndian, uint16(g.height))

	// Write elevation data (float32 little-endian)
	for _, elev := range g.elevation {
		binary.Write(buf, binary.LittleEndian, elev)
	}

	// Write biome data (uint8)
	for _, biome := range g.biome {
		buf.WriteByte(byte(biome))
	}

	return buf.Bytes()
}

// ParseBinaryGrid parses a binary byte slice into a BinaryGrid
func ParseBinaryGrid(data []byte) (*BinaryGrid, error) {
	if len(data) < headerSize {
		return nil, errors.New("binary grid data too short")
	}

	// Verify magic
	if string(data[0:4]) != binaryGridMagic {
		return nil, errors.New("invalid binary grid magic")
	}

	// Read header
	version := data[4]
	if version != binaryGridVersion {
		return nil, errors.New("unsupported binary grid version")
	}

	width := int(binary.LittleEndian.Uint16(data[5:7]))
	height := int(binary.LittleEndian.Uint16(data[7:9]))

	size := width * height
	expectedSize := headerSize + size*4 + size

	if len(data) < expectedSize {
		return nil, errors.New("binary grid data truncated")
	}

	grid := NewBinaryGrid(width, height)

	// Read elevation data
	offset := headerSize
	for i := 0; i < size; i++ {
		bits := binary.LittleEndian.Uint32(data[offset : offset+4])
		grid.elevation[i] = math.Float32frombits(bits)
		offset += 4
	}

	// Read biome data
	for i := 0; i < size; i++ {
		grid.biome[i] = BiomeID(data[offset])
		offset++
	}

	return grid, nil
}

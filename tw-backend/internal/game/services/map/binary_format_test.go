package gamemap

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBinaryGridFormat_HeaderMagic verifies the header structure.
func TestBinaryGridFormat_HeaderMagic(t *testing.T) {
	grid := NewBinaryGrid(256, 256)
	require.NotNil(t, grid)

	data := grid.Serialize()

	// Header should start with magic bytes "WMAP"
	assert.Equal(t, []byte("WMAP"), data[:4], "Magic bytes should be WMAP")

	// Version byte
	assert.Equal(t, byte(1), data[4], "Version should be 1")

	// Width (little-endian uint16)
	width := uint16(data[5]) | uint16(data[6])<<8
	assert.Equal(t, uint16(256), width, "Width should be 256")

	// Height (little-endian uint16)
	height := uint16(data[7]) | uint16(data[8])<<8
	assert.Equal(t, uint16(256), height, "Height should be 256")
}

// TestBinaryGridFormat_RoundTrip verifies serialize/deserialize cycle.
func TestBinaryGridFormat_RoundTrip(t *testing.T) {
	// Create grid with known values
	grid := NewBinaryGrid(64, 64)
	require.NotNil(t, grid)

	// Set some values
	grid.SetElevation(0, 0, 100.5)
	grid.SetElevation(63, 63, -500.0)
	grid.SetBiome(0, 0, BiomeIDOcean)
	grid.SetBiome(10, 10, BiomeIDGrassland)

	// Serialize
	data := grid.Serialize()

	// Deserialize
	parsed, err := ParseBinaryGrid(data)
	require.NoError(t, err)
	require.NotNil(t, parsed)

	// Verify dimensions
	assert.Equal(t, 64, parsed.Width())
	assert.Equal(t, 64, parsed.Height())

	// Verify values
	assert.InDelta(t, 100.5, parsed.GetElevation(0, 0), 0.01)
	assert.InDelta(t, -500.0, parsed.GetElevation(63, 63), 0.01)
	assert.Equal(t, BiomeIDOcean, parsed.GetBiome(0, 0))
	assert.Equal(t, BiomeIDGrassland, parsed.GetBiome(10, 10))
}

// TestBinaryGridFormat_Size verifies the expected binary size.
func TestBinaryGridFormat_Size(t *testing.T) {
	scenarios := []struct {
		name           string
		width, height  int
		expectedHeader int
		expectedTotal  int
	}{
		{"256x256", 256, 256, 9, 9 + 256*256*4 + 256*256}, // 9 + 262144 + 65536 = 327689
		{"64x64", 64, 64, 9, 9 + 64*64*4 + 64*64},         // 9 + 16384 + 4096 = 20489
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			grid := NewBinaryGrid(sc.width, sc.height)
			data := grid.Serialize()

			// Header: 4 (magic) + 1 (version) + 2 (width) + 2 (height) = 9 bytes
			// Elevation: width * height * 4 (float32)
			// Biome: width * height * 1 (uint8)
			expectedTotal := sc.expectedHeader + sc.width*sc.height*4 + sc.width*sc.height
			assert.Equal(t, expectedTotal, len(data), "Total binary size")
		})
	}
}

// TestBinaryGridFormat_InvalidHeader verifies error handling.
func TestBinaryGridFormat_InvalidHeader(t *testing.T) {
	// Too short
	_, err := ParseBinaryGrid([]byte{0, 1, 2})
	assert.Error(t, err, "Should error on short data")

	// Wrong magic
	wrongMagic := bytes.Repeat([]byte{0}, 1000)
	_, err = ParseBinaryGrid(wrongMagic)
	assert.Error(t, err, "Should error on wrong magic")
}

// TestBiomeIDMapping verifies biome string to ID conversion.
func TestBiomeIDMapping(t *testing.T) {
	// Test all known biome types
	scenarios := []struct {
		biomeString string
		expectedID  BiomeID
	}{
		{"Ocean", BiomeIDOcean},
		{"Grassland", BiomeIDGrassland},
		{"Desert", BiomeIDDesert},
		{"Rainforest", BiomeIDRainforest},
		{"Tundra", BiomeIDTundra},
		{"Unknown", BiomeIDUnknown},
		{"", BiomeIDUnknown},
	}

	for _, sc := range scenarios {
		t.Run(sc.biomeString, func(t *testing.T) {
			id := BiomeStringToID(sc.biomeString)
			assert.Equal(t, sc.expectedID, id)
		})
	}
}

// TestBiomeIDToString verifies ID to string conversion.
func TestBiomeIDToString(t *testing.T) {
	assert.Equal(t, "Ocean", BiomeIDToString(BiomeIDOcean))
	assert.Equal(t, "Grassland", BiomeIDToString(BiomeIDGrassland))
	assert.Equal(t, "Unknown", BiomeIDToString(BiomeIDUnknown))
	assert.Equal(t, "Unknown", BiomeIDToString(255)) // Invalid ID
}

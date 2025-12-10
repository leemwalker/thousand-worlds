package geography

import (
	"github.com/google/uuid"
)

// Point represents a 2D coordinate
type Point struct {
	X, Y float64
}

// Vector represents a 2D direction and magnitude
type Vector struct {
	X, Y float64
}

// PlateType distinguishes between continental and oceanic plates
type PlateType string

const (
	PlateContinental PlateType = "continental"
	PlateOceanic     PlateType = "oceanic"
)

// BoundaryType represents the type of interaction between plates
type BoundaryType string

const (
	BoundaryDivergent  BoundaryType = "divergent"
	BoundaryConvergent BoundaryType = "convergent"
	BoundaryTransform  BoundaryType = "transform"
)

// TectonicPlate represents a piece of the planet's crust
type TectonicPlate struct {
	PlateID        uuid.UUID
	Type           PlateType
	BoundaryPoints []Point // Polygon vertices defining the plate shape
	Centroid       Point
	MovementVector Vector
	Thickness      float64 // km
	Age            float64 // million years
}

// Heightmap represents the elevation grid of the world
type Heightmap struct {
	Width      int
	Height     int
	Elevations []float64 // 1D array mapped to 2D grid
	MinElev    float64
	MaxElev    float64
}

// NewHeightmap creates a new heightmap
func NewHeightmap(width, height int) *Heightmap {
	return &Heightmap{
		Width:      width,
		Height:     height,
		Elevations: make([]float64, width*height),
	}
}

// Get returns elevation at x,y
func (h *Heightmap) Get(x, y int) float64 {
	if x < 0 || x >= h.Width || y < 0 || y >= h.Height {
		return 0
	}
	return h.Elevations[y*h.Width+x]
}

// Set sets elevation at x,y
func (h *Heightmap) Set(x, y int, val float64) {
	if x >= 0 && x < h.Width && y >= 0 && y < h.Height {
		h.Elevations[y*h.Width+x] = val
	}
}

// BiomeType represents the classification of a region
type BiomeType string

const (
	BiomeOcean           BiomeType = "Ocean"
	BiomeLowland         BiomeType = "Lowland"
	BiomeHighland        BiomeType = "Highland"
	BiomeMountain        BiomeType = "Mountain"
	BiomeHighMountain    BiomeType = "High Mountain"
	BiomeRainforest      BiomeType = "Rainforest"
	BiomeDesert          BiomeType = "Desert"
	BiomeGrassland       BiomeType = "Grassland"
	BiomeDeciduousForest BiomeType = "Deciduous Forest"
	BiomeTaiga           BiomeType = "Taiga"
	BiomeTundra          BiomeType = "Tundra"
	BiomeAlpine          BiomeType = "Alpine"
)

// Biome represents a specific ecological region
type Biome struct {
	BiomeID       uuid.UUID
	Name          string
	Type          BiomeType
	Temperature   float64 // Average Celsius
	Precipitation float64 // mm/year
	Vegetation    []string
	NativeSpecies []string
	Resources     []string
}

// WorldMap holds all generated geographic data
type WorldMap struct {
	Heightmap *Heightmap
	Plates    []TectonicPlate
	Biomes    []Biome
	Rivers    [][]Point // List of river paths
}

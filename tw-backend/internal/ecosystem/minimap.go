// Package ecosystem provides minimap rendering with biome and elevation palettes.
package ecosystem

import (
	"github.com/google/uuid"
)

// BiomeVisual defines the visual representation of a biome
type BiomeVisual struct {
	Emoji    string `json:"emoji"`
	Char     string `json:"char"`
	Color    string `json:"color"`    // Hex for char rendering
	Tailwind string `json:"tailwind"` // Tailwind class for emoji bg
}

// ElevationVisual defines the visual representation of an elevation tier
type ElevationVisual struct {
	Name     string `json:"name"`
	Color    string `json:"color"`    // Hex for char background
	Tailwind string `json:"tailwind"` // Tailwind class
}

// CatastropheVisual defines the visual overlay for catastrophic events
type CatastropheVisual struct {
	Emoji    string `json:"emoji"`
	Char     string `json:"char"`
	Color    string `json:"color"`
	Tailwind string `json:"tailwind"` // Includes animation
}

// BiomeMap maps biome types to their visual representation
var BiomeMap = map[string]BiomeVisual{
	"ocean":      {Emoji: "🌊", Char: "~", Color: "#1d4ed8", Tailwind: "bg-blue-700 text-blue-200"},
	"rainforest": {Emoji: "🌴", Char: "%", Color: "#065f46", Tailwind: "bg-emerald-800 text-emerald-300"},
	"grassland":  {Emoji: "🌾", Char: "\"", Color: "#84cc16", Tailwind: "bg-lime-500 text-lime-900"},
	"deciduous":  {Emoji: "🌳", Char: "&", Color: "#16a34a", Tailwind: "bg-green-600 text-green-100"},
	"alpine":     {Emoji: "🏔️", Char: "^", Color: "#a8a29e", Tailwind: "bg-stone-400 text-stone-800"},
	"taiga":      {Emoji: "🌲", Char: "*", Color: "#134e4a", Tailwind: "bg-teal-900 text-teal-200"},
	"desert":     {Emoji: "🌵", Char: ".", Color: "#fcd34d", Tailwind: "bg-amber-300 text-amber-800"},
	"tundra":     {Emoji: "❄️", Char: "-", Color: "#e2e8f0", Tailwind: "bg-slate-200 text-slate-600"},
}

// ElevationMap maps elevation ranges to visual representation
var ElevationMap = []struct {
	MaxElevation float64
	Visual       ElevationVisual
}{
	{-1000, ElevationVisual{Name: "deep_ocean", Color: "#1e3a5f", Tailwind: "bg-blue-900"}},
	{0, ElevationVisual{Name: "shallow_water", Color: "#3b82f6", Tailwind: "bg-blue-500"}},
	{500, ElevationVisual{Name: "lowland", Color: "#22c55e", Tailwind: "bg-green-500"}},
	{2000, ElevationVisual{Name: "highland", Color: "#a16207", Tailwind: "bg-amber-700"}},
	{10000, ElevationVisual{Name: "peak", Color: "#f5f5f4", Tailwind: "bg-stone-100"}},
}

// CatastropheMap maps catastrophe types to visual overlays
var CatastropheMap = map[string]CatastropheVisual{
	"volcano":      {Emoji: "🌋", Char: "A", Color: "#dc2626", Tailwind: "bg-red-600 animate-pulse"},
	"asteroid":     {Emoji: "☄️", Char: "@", Color: "#ea580c", Tailwind: "bg-orange-600 animate-bounce"},
	"flood_basalt": {Emoji: "♨️", Char: "#", Color: "#171717", Tailwind: "bg-neutral-900 text-red-500"},
	"anoxia":       {Emoji: "🦠", Char: "~", Color: "#6b21a8", Tailwind: "bg-purple-800 text-purple-300"},
	"ice_age":      {Emoji: "🧊", Char: "=", Color: "#cffafe", Tailwind: "bg-cyan-100 text-cyan-800"},
}

// MinimapCell represents a single cell in the minimap
type MinimapCell struct {
	Q         int     `json:"q"`
	R         int     `json:"r"`
	BiomeType string  `json:"biome_type"`
	Elevation float64 `json:"elevation"`

	// Biome visual data
	BiomeEmoji    string `json:"biome_emoji"`
	BiomeChar     string `json:"biome_char"`
	BiomeColor    string `json:"biome_color"`
	BiomeTailwind string `json:"biome_tailwind"`

	// Elevation visual data
	ElevName     string `json:"elev_name"`
	ElevColor    string `json:"elev_color"`
	ElevTailwind string `json:"elev_tailwind"`

	// Catastrophe overlay (if active)
	Catastrophe         string `json:"catastrophe,omitempty"`
	CatastropheEmoji    string `json:"catastrophe_emoji,omitempty"`
	CatastropheChar     string `json:"catastrophe_char,omitempty"`
	CatastropheColor    string `json:"catastrophe_color,omitempty"`
	CatastropheTailwind string `json:"catastrophe_tailwind,omitempty"`
}

// GetBiomeVisual returns the visual representation for a biome type
func GetBiomeVisual(biomeType string) BiomeVisual {
	if visual, ok := BiomeMap[biomeType]; ok {
		return visual
	}
	// Default to grassland if unknown
	return BiomeMap["grassland"]
}

// GetElevationVisual returns the visual representation for an elevation
func GetElevationVisual(elevation float64) ElevationVisual {
	for _, tier := range ElevationMap {
		if elevation <= tier.MaxElevation {
			return tier.Visual
		}
	}
	// Default to peak for extremely high elevations
	return ElevationMap[len(ElevationMap)-1].Visual
}

// GetCatastropheVisual returns the visual overlay for a catastrophe
func GetCatastropheVisual(catastropheType string) *CatastropheVisual {
	if visual, ok := CatastropheMap[catastropheType]; ok {
		return &visual
	}
	return nil
}

// NewMinimapCell creates a fully populated minimap cell
func NewMinimapCell(q, r int, biomeType string, elevation float64, catastrophe string) MinimapCell {
	biomeVis := GetBiomeVisual(biomeType)
	elevVis := GetElevationVisual(elevation)

	cell := MinimapCell{
		Q:             q,
		R:             r,
		BiomeType:     biomeType,
		Elevation:     elevation,
		BiomeEmoji:    biomeVis.Emoji,
		BiomeChar:     biomeVis.Char,
		BiomeColor:    biomeVis.Color,
		BiomeTailwind: biomeVis.Tailwind,
		ElevName:      elevVis.Name,
		ElevColor:     elevVis.Color,
		ElevTailwind:  elevVis.Tailwind,
	}

	if catastrophe != "" {
		if catVis := GetCatastropheVisual(catastrophe); catVis != nil {
			cell.Catastrophe = catastrophe
			cell.CatastropheEmoji = catVis.Emoji
			cell.CatastropheChar = catVis.Char
			cell.CatastropheColor = catVis.Color
			cell.CatastropheTailwind = catVis.Tailwind
		}
	}

	return cell
}

// MinimapUpdate contains a batch of cell updates for broadcast
type MinimapBatch struct {
	WorldID uuid.UUID     `json:"world_id"`
	Year    int64         `json:"year"`
	Cells   []MinimapCell `json:"cells"`
}

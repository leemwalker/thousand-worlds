package minerals

import (
	"math/rand"

	"tw-backend/internal/worldgen/geography"
)

// DetermineVeinSize calculates the size of the vein based on mineral type and context
func DetermineVeinSize(mineralType MineralType, context *TectonicContext) VeinSize {
	// Base size probability
	r := rand.Float64()

	// Adjust based on formation compatibility
	// e.g. Igneous minerals in volcanic areas are larger
	if mineralType.FormationType == FormationIgneous && context.IsVolcanic {
		if r > 0.7 {
			return VeinSizeMassive
		} else if r > 0.4 {
			return VeinSizeLarge
		}
		return VeinSizeMedium
	}

	// Sedimentary in basins
	if mineralType.FormationType == FormationSedimentary && context.IsSedimentaryBasin {
		if r > 0.6 {
			return VeinSizeMassive
		} else if r > 0.3 {
			return VeinSizeLarge
		}
		return VeinSizeMedium
	}

	// Default distribution
	if r > 0.95 {
		return VeinSizeMassive
	} else if r > 0.8 {
		return VeinSizeLarge
	} else if r > 0.5 {
		return VeinSizeMedium
	}
	return VeinSizeSmall
}

// GenerateVeinGeometry determines the shape and dimensions of the vein
func GenerateVeinGeometry(mineralType MineralType, context *TectonicContext, size VeinSize) (VeinShape, geography.Vector, float64, float64) {
	var shape VeinShape
	var orientation geography.Vector
	var length, width float64

	// Base dimensions multiplier based on size
	sizeMult := 1.0
	switch size {
	case VeinSizeSmall:
		sizeMult = 0.5
	case VeinSizeMedium:
		sizeMult = 1.0
	case VeinSizeLarge:
		sizeMult = 2.0
	case VeinSizeMassive:
		sizeMult = 5.0
	}

	switch mineralType.FormationType {
	case FormationIgneous:
		// Follows magma intrusion paths
		shape = VeinShapeLinear
		orientation = context.MagmaFlowDirection
		length = (500 + rand.Float64()*2000) * sizeMult
		width = (10 + rand.Float64()*50) * sizeMult

	case FormationSedimentary:
		// Horizontal layers
		shape = VeinShapePlanar
		orientation = geography.Vector{X: 0, Y: 0} // Horizontal plane (conceptually)
		length = (1000 + rand.Float64()*5000) * sizeMult
		width = (500 + rand.Float64()*2000) * sizeMult

	case FormationMetamorphic:
		// Follows fault lines
		shape = VeinShapeLinear
		orientation = context.FaultLineDirection
		length = (200 + rand.Float64()*1500) * sizeMult
		width = (5 + rand.Float64()*30) * sizeMult
	}

	return shape, orientation, length, width
}

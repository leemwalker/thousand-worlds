package minerals

import (
	"math/rand"

	"mud-platform-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// GenerateMineralVein creates a single mineral deposit based on context
func GenerateMineralVein(
	tectonicContext *TectonicContext,
	mineralType MineralType,
	epicenter geography.Point,
) *MineralDeposit {

	// Generate unique ID
	depositID := uuid.New()

	// 1. Determine vein size based on mineral type and geological conditions
	veinSize := DetermineVeinSize(mineralType, tectonicContext)

	// 2. Calculate quantity based on size and concentration
	concentration := CalculateConcentration(mineralType, tectonicContext)
	baseQuantity := GetBaseQuantity(mineralType, veinSize)
	quantity := int(float64(baseQuantity) * concentration)

	// 3. Determine vein shape based on formation type
	veinShape, veinOrientation, veinLength, veinWidth := GenerateVeinGeometry(mineralType, tectonicContext, veinSize)

	// 4. Determine depth based on formation conditions
	depth := CalculateDepositDepth(mineralType, tectonicContext, epicenter)

	// 5. Surface visibility (only if depth < 50m and in mountainous/eroded area)
	surfaceVisible := depth < 50 && tectonicContext.ErosionLevel > 0.7

	return &MineralDeposit{
		DepositID:       depositID,
		MineralType:     mineralType,
		FormationType:   mineralType.FormationType,
		Location:        epicenter,
		Depth:           depth,
		Quantity:        quantity,
		Concentration:   concentration,
		VeinSize:        veinSize,
		VeinShape:       veinShape,
		VeinOrientation: veinOrientation,
		VeinLength:      veinLength,
		VeinWidth:       veinWidth,
		SurfaceVisible:  surfaceVisible,
		RequiredDepth:   depth,
		GeologicalAge:   tectonicContext.Age,
	}
}

// CalculateDepositDepth determines the depth of the deposit
func CalculateDepositDepth(mineralType MineralType, context *TectonicContext, epicenter geography.Point) float64 {
	baseDepth := 0.0

	switch mineralType.FormationType {
	case FormationIgneous:
		// Can be deep or shallow depending on volcanic activity
		if context.IsVolcanic {
			baseDepth = 0 + rand.Float64()*1000 // Surface to 1km
		} else {
			baseDepth = 500 + rand.Float64()*2000 // 500m to 2.5km
		}
	case FormationSedimentary:
		// Usually shallower, in layers
		baseDepth = 100 + rand.Float64()*900 // 100m to 1km
	case FormationMetamorphic:
		// Deep, high pressure
		baseDepth = 500 + rand.Float64()*2500 // 500m to 3km
	}

	// Adjust for specific minerals
	if mineralType.Name == "Diamond" || mineralType.Name == "Platinum" {
		baseDepth += 1500 // Much deeper
	}

	// Adjust for erosion (erosion brings deep things closer to surface)
	erosionOffset := context.ErosionLevel * 500
	depth := baseDepth - erosionOffset
	if depth < 0 {
		depth = 0
	}

	return depth
}

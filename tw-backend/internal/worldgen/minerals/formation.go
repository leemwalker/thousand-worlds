package minerals

import (
	"math/rand"

	"tw-backend/internal/worldgen/geography"

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

// Point represents a 2D coordinate (for placer deposit river locations)
type Point struct {
	X, Y float64
}

// GenerateBIFDeposits generates Banded Iron Formation deposits at ancient oceanic locations.
// BIF forms when rising O2 precipitates dissolved iron (Great Oxygenation Event).
// RED STATE: Returns nil - not yet implemented.
func GenerateBIFDeposits(oceanLocations []Point, oxygenSpike float64) []*MineralDeposit {
	// TODO: Implement BIF formation
	// Higher oxygenSpike should create more deposits
	// Deposits should contain alternating chert and hematite layers
	return nil
}

// GeneratePlacerDeposits generates alluvial deposits along river paths.
// Gold and heavy minerals accumulate at river bends through erosion.
// RED STATE: Returns nil - not yet implemented.
func GeneratePlacerDeposits(riverPaths [][]Point, mineralType string, erosionRate float64) []*MineralDeposit {
	// TODO: Implement placer formation
	// Deposits should form at bend in river paths
	// Concentration should increase downstream
	return nil
}

// GenerateHydrothermalDeposits creates sulfide deposits at ocean ridges.
// RED STATE: Returns nil - not yet implemented.
func GenerateHydrothermalDeposits(ridgeLocations []Point) []*MineralDeposit {
	// TODO: Implement hydrothermal vent deposits
	// Should contain copper, zinc, and gold
	return nil
}

// GenerateKimberlitePipe creates diamond-bearing volcanic pipes.
// Only forms in ancient cratons with deep mantle eruptions.
// RED STATE: Returns nil - not yet implemented.
func GenerateKimberlitePipe(cratonAge float64, depth float64) *MineralDeposit {
	// TODO: Implement kimberlite pipe formation
	// Craton must be > 2.5B years old
	// Depth must be > 150km for diamond stability
	return nil
}

// ExtractResource extracts minerals from a deposit, reducing its quantity.
// Returns the amount actually extracted.
// RED STATE: Returns 0 - not yet implemented.
func ExtractResource(deposit *MineralDeposit, amount int) int {
	// TODO: Implement extraction logic
	// Should respect tool hardness requirements
	// Should reduce deposit.Quantity
	return 0
}

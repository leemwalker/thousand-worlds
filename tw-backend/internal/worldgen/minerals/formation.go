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
// Requires oxygenSpike > 0.05 threshold — the Great Oxygenation Event.
func GenerateBIFDeposits(oceanLocations []Point, oxygenSpike float64) []*MineralDeposit {
	// Threshold: oxygen spike must be significant to trigger iron precipitation
	const oxygenThreshold = 0.05
	if oxygenSpike <= oxygenThreshold || len(oceanLocations) == 0 {
		return nil
	}

	deposits := make([]*MineralDeposit, 0, len(oceanLocations))

	for _, loc := range oceanLocations {
		// Create BIF deposit at this ocean location
		deposit := &MineralDeposit{
			DepositID: uuid.New(),
			MineralType: MineralType{
				Name:          "Iron",
				FormationType: FormationSedimentary,
			},
			FormationType: FormationSedimentary,
			Location: geography.Point{
				X: loc.X,
				Y: loc.Y,
			},
			Quantity:      int(oxygenSpike * 10000), // Scale with oxygen level
			Concentration: oxygenSpike,
			VeinSize:      VeinSizeLarge,
		}
		deposits = append(deposits, deposit)
	}

	return deposits
}

// GeneratePlacerDeposits generates alluvial deposits along river paths.
// Gold and heavy minerals accumulate at river bends through erosion.
func GeneratePlacerDeposits(riverPaths [][]Point, mineralType string, erosionRate float64) []*MineralDeposit {
	if len(riverPaths) == 0 || erosionRate <= 0 {
		return nil
	}

	deposits := make([]*MineralDeposit, 0)

	for _, path := range riverPaths {
		// Find bends in the river (points where direction changes significantly)
		// For simplicity, treat every 2nd-3rd point as a potential bend site
		for i := 1; i < len(path)-1; i++ {
			loc := path[i]

			// Create placer deposit at bend
			deposit := &MineralDeposit{
				DepositID: uuid.New(),
				MineralType: MineralType{
					Name:          mineralType,
					FormationType: FormationSedimentary,
				},
				FormationType: FormationSedimentary,
				Location: geography.Point{
					X: loc.X,
					Y: loc.Y,
				},
				Quantity:      int(erosionRate * float64(i+1) * 100), // Increases downstream
				Concentration: erosionRate * float64(i+1) / float64(len(path)),
				VeinSize:      VeinSizeSmall,
			}
			deposits = append(deposits, deposit)
		}
	}

	if len(deposits) == 0 {
		return nil
	}
	return deposits
}

// GenerateHydrothermalDeposits creates sulfide deposits at ocean ridges.
func GenerateHydrothermalDeposits(ridgeLocations []Point) []*MineralDeposit {
	if len(ridgeLocations) == 0 {
		return nil
	}

	deposits := make([]*MineralDeposit, 0, len(ridgeLocations))

	// Hydrothermal minerals: copper, zinc, gold
	mineralTypes := []string{"Copper", "Zinc", "Gold"}

	for i, loc := range ridgeLocations {
		// Cycle through mineral types for each vent location
		mineralName := mineralTypes[i%len(mineralTypes)]

		deposit := &MineralDeposit{
			DepositID: uuid.New(),
			MineralType: MineralType{
				Name:          mineralName,
				FormationType: FormationIgneous,
			},
			FormationType: FormationIgneous,
			Location: geography.Point{
				X: loc.X,
				Y: loc.Y,
			},
			Quantity:      500 + (i * 100), // Varying quantities
			Concentration: 0.6 + float64(i)*0.1,
			VeinSize:      VeinSizeMedium,
		}
		deposits = append(deposits, deposit)
	}

	return deposits
}

// GenerateKimberlitePipe creates diamond-bearing volcanic pipes.
// Only forms in ancient cratons with deep mantle eruptions.
// Requirements: craton > 2.5B years old, depth > 150km
func GenerateKimberlitePipe(cratonAge float64, depth float64) *MineralDeposit {
	// Kimberlite conditions:
	// 1. Craton must be ancient (> 2.5 billion years)
	// 2. Eruption must be from deep mantle (> 150 km for diamond stability)
	const minCratonAge = 2.5 // Billion years
	const minDepth = 150.0   // km

	if cratonAge < minCratonAge || depth < minDepth {
		return nil
	}

	return &MineralDeposit{
		DepositID: uuid.New(),
		MineralType: MineralType{
			Name:          "Diamond",
			FormationType: FormationIgneous,
			Hardness:      10.0, // Diamonds are hardest
		},
		FormationType: FormationIgneous,
		Depth:         depth,
		Quantity:      int((cratonAge - minCratonAge) * (depth - minDepth) * 10),
		Concentration: cratonAge / 5.0, // Older = richer
		VeinSize:      VeinSizeMedium,
		GeologicalAge: cratonAge,
	}
}

// ExtractResource extracts minerals from a deposit, reducing its quantity.
// Returns the amount actually extracted (capped at available quantity).
func ExtractResource(deposit *MineralDeposit, amount int) int {
	if deposit == nil || amount <= 0 {
		return 0
	}

	// Extract only what's available
	extracted := amount
	if extracted > deposit.Quantity {
		extracted = deposit.Quantity
	}

	// Reduce deposit quantity
	deposit.Quantity -= extracted

	return extracted
}

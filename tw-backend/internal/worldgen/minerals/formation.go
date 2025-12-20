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

// CoalFormationConfig holds parameters for coal seam generation
type CoalFormationConfig struct {
	OrganicMatter float64 // Amount of plant matter buried (0-1)
	BurialDepth   float64 // Meters
	BurialAge     int64   // Years of burial
}

// GenerateCoalDeposits creates coal seams from buried organic matter.
// Coal rank increases with depth and age (peat → lignite → bituminous → anthracite).
func GenerateCoalDeposits(config CoalFormationConfig) []*MineralDeposit {
	// Coal requires significant organic matter and burial
	if config.OrganicMatter < 0.3 || config.BurialAge < 1_000_000 {
		return nil
	}

	deposits := make([]*MineralDeposit, 0)

	// Determine coal rank based on burial depth and age
	// Deeper/older = higher rank
	var coalType string
	rank := (config.BurialDepth / 1000) + float64(config.BurialAge)/100_000_000

	switch {
	case rank < 1:
		coalType = "Peat"
	case rank < 2:
		coalType = "Lignite"
	case rank < 5:
		coalType = "Bituminous"
	default:
		coalType = "Anthracite"
	}

	deposit := &MineralDeposit{
		DepositID: uuid.New(),
		MineralType: MineralType{
			Name:          coalType,
			FormationType: FormationSedimentary,
			BaseValue:     int(rank * 3),
		},
		FormationType: FormationSedimentary,
		Depth:         config.BurialDepth,
		Quantity:      int(config.OrganicMatter * 10000),
		Concentration: config.OrganicMatter,
		VeinSize:      VeinSizeLarge,
		GeologicalAge: float64(config.BurialAge) / 1_000_000,
	}
	deposits = append(deposits, deposit)

	return deposits
}

// EvaporiteFormationConfig holds parameters for evaporite generation
type EvaporiteFormationConfig struct {
	WaterVolume   float64 // Initial water in basin
	EvaporateRate float64 // Rate of evaporation (0-1)
	Climate       string  // "arid", "semi-arid", etc.
}

// GenerateEvaporiteDeposits creates salt and gypsum deposits from evaporating water.
// Forms in sequence: carbonates → gypsum → halite (salt) → potash
func GenerateEvaporiteDeposits(config EvaporiteFormationConfig) []*MineralDeposit {
	// Evaporites require arid climate and significant evaporation
	if config.EvaporateRate < 0.5 || config.Climate != "arid" {
		return nil
	}

	deposits := make([]*MineralDeposit, 0)

	// Evaporation sequence based on solubility (least soluble precipitates first)
	evaporites := []struct {
		name       string
		solubility float64
	}{
		{"Gypsum", 0.3},
		{"Halite", 0.6},
		{"Potash", 0.9},
	}

	for _, evap := range evaporites {
		if config.EvaporateRate >= evap.solubility {
			deposit := &MineralDeposit{
				DepositID: uuid.New(),
				MineralType: MineralType{
					Name:          evap.name,
					FormationType: FormationSedimentary,
				},
				FormationType: FormationSedimentary,
				Quantity:      int(config.WaterVolume * (1 - evap.solubility) * 1000),
				Concentration: config.EvaporateRate,
				VeinSize:      VeinSizeLarge,
			}
			deposits = append(deposits, deposit)
		}
	}

	if len(deposits) == 0 {
		return nil
	}
	return deposits
}

// GenerateToolStoneDeposits creates obsidian/flint deposits suitable for tool-making.
// Obsidian from volcanic flows, flint from chalk/limestone nodules.
func GenerateToolStoneDeposits(volcanicContext bool, hasChalk bool) []*MineralDeposit {
	if !volcanicContext && !hasChalk {
		return nil
	}

	deposits := make([]*MineralDeposit, 0)

	if volcanicContext {
		// Obsidian - volcanic glass
		deposit := &MineralDeposit{
			DepositID: uuid.New(),
			MineralType: MineralType{
				Name:          "Obsidian",
				FormationType: FormationIgneous,
				Hardness:      5.5,
				BaseValue:     20,
			},
			FormationType: FormationIgneous,
			Quantity:      500,
			Concentration: 0.8,
			VeinSize:      VeinSizeSmall,
		}
		deposits = append(deposits, deposit)
	}

	if hasChalk {
		// Flint - silica nodules in chalk
		deposit := &MineralDeposit{
			DepositID: uuid.New(),
			MineralType: MineralType{
				Name:          "Flint",
				FormationType: FormationSedimentary,
				Hardness:      7.0,
				BaseValue:     10,
			},
			FormationType: FormationSedimentary,
			Quantity:      1000,
			Concentration: 0.6,
			VeinSize:      VeinSizeSmall,
		}
		deposits = append(deposits, deposit)
	}

	return deposits
}

// DiscoverDeposits searches a region for mineral deposits.
// Returns deposits found based on survey skill and surface visibility.
func DiscoverDeposits(deposits []*MineralDeposit, searchDepth float64, surveySkill float64) []*MineralDeposit {
	if len(deposits) == 0 || surveySkill <= 0 {
		return nil
	}

	discovered := make([]*MineralDeposit, 0)

	for _, dep := range deposits {
		// Surface-visible deposits are always found
		if dep.SurfaceVisible {
			discovered = append(discovered, dep)
			continue
		}

		// Hidden deposits require depth access and skill
		if dep.Depth <= searchDepth && surveySkill >= 0.5 {
			// Probability based on skill and depth
			discoveryChance := surveySkill * (1 - dep.Depth/searchDepth*0.5)
			if discoveryChance > 0.3 {
				discovered = append(discovered, dep)
			}
		}
	}

	if len(discovered) == 0 {
		return nil
	}
	return discovered
}

// SampleConcentration returns the ore grade at a specific point in a deposit.
// Concentration varies from center (richest) to edges.
func SampleConcentration(deposit *MineralDeposit, sampleX, sampleY float64) float64 {
	if deposit == nil {
		return 0
	}

	// Calculate distance from deposit center
	dx := sampleX - deposit.Location.X
	dy := sampleY - deposit.Location.Y
	distance := dx*dx + dy*dy

	// Use vein dimensions for falloff (approximate radius)
	radiusSq := deposit.VeinLength * deposit.VeinWidth / 4
	if radiusSq == 0 {
		radiusSq = 100 // Default 10m radius
	}

	// Concentration decreases with distance from center
	falloff := 1 - (distance / (radiusSq * 4))
	if falloff < 0 {
		falloff = 0
	}

	return deposit.Concentration * falloff
}

// TinFormationContext holds parameters for tin deposit generation
type TinFormationContext struct {
	HasGranite     bool    // Granitic intrusion present
	HasSedimentary bool    // Sedimentary contact zone
	Temperature    float64 // Hydrothermal temperature
}

// GenerateTinDeposits creates cassiterite (tin ore) deposits at granite-sedimentary contacts.
// Tin commonly co-occurs with copper in hydrothermal systems (important for Bronze Age).
func GenerateTinDeposits(ctx TinFormationContext, copperLocations []Point) []*MineralDeposit {
	// Tin requires granite-sedimentary contact and hydrothermal activity
	if !ctx.HasGranite || !ctx.HasSedimentary || ctx.Temperature < 300 {
		return nil
	}

	deposits := make([]*MineralDeposit, 0)

	// Create tin deposits near copper locations (co-location for bronze production)
	for i, copperLoc := range copperLocations {
		// Cassiterite (tin ore) deposit
		tinDeposit := &MineralDeposit{
			DepositID: uuid.New(),
			MineralType: MineralType{
				Name:          "Cassiterite",
				FormationType: FormationIgneous,
				BaseValue:     30, // Tin is valuable for bronze
			},
			FormationType: FormationIgneous,
			Location: geography.Point{
				X: copperLoc.X + float64(i%3-1)*10, // Slight offset from copper
				Y: copperLoc.Y + float64(i%2)*10,
			},
			Quantity:      200 + i*50,
			Concentration: 0.6,
			VeinSize:      VeinSizeSmall,
		}
		deposits = append(deposits, tinDeposit)
	}

	if len(deposits) == 0 {
		return nil
	}
	return deposits
}

// SaltpeterFormationContext holds parameters for saltpeter generation
type SaltpeterFormationContext struct {
	HasCaves      bool    // Cave environment present
	HasDesert     bool    // Desert soil conditions
	OrganicMatter float64 // Organic source (bat guano, etc.) 0-1
}

// GenerateSaltpeterDeposits creates potassium nitrate deposits for gunpowder.
// Saltpeter forms in caves (bat guano) or arid soils with nitrogen fixation.
func GenerateSaltpeterDeposits(ctx SaltpeterFormationContext) []*MineralDeposit {
	// Saltpeter requires either cave with organic matter or desert conditions
	if !ctx.HasCaves && !ctx.HasDesert {
		return nil
	}
	if ctx.OrganicMatter < 0.1 && !ctx.HasDesert {
		return nil
	}

	deposits := make([]*MineralDeposit, 0)

	if ctx.HasCaves && ctx.OrganicMatter > 0.1 {
		// Cave saltpeter from bat guano - higher quality
		deposit := &MineralDeposit{
			DepositID: uuid.New(),
			MineralType: MineralType{
				Name:          "Saltpeter",
				FormationType: FormationSedimentary,
				BaseValue:     15,
			},
			FormationType: FormationSedimentary,
			Quantity:      int(ctx.OrganicMatter * 500),
			Concentration: ctx.OrganicMatter,
			VeinSize:      VeinSizeSmall,
			Depth:         0, // Surface of cave floor
		}
		deposits = append(deposits, deposit)
	}

	if ctx.HasDesert {
		// Desert saltpeter - lower concentration but larger areas
		deposit := &MineralDeposit{
			DepositID: uuid.New(),
			MineralType: MineralType{
				Name:          "Saltpeter",
				FormationType: FormationSedimentary,
				BaseValue:     10,
			},
			FormationType: FormationSedimentary,
			Quantity:      300,
			Concentration: 0.3,
			VeinSize:      VeinSizeMedium,
			Depth:         0,
		}
		deposits = append(deposits, deposit)
	}

	return deposits
}

// ManaCrystalFormationContext holds parameters for magical crystal generation
type ManaCrystalFormationContext struct {
	LeyLineStrength float64 // Magical energy level (0-1)
	IsIntersection  bool    // Multiple ley lines cross here
	Depth           float64 // Underground depth
}

// GenerateManaCrystals creates magical crystals at ley line locations.
// Crystal potency depends on ley line strength and intersection multiplier.
func GenerateManaCrystals(ctx ManaCrystalFormationContext) []*MineralDeposit {
	// Requires minimum magical energy level
	if ctx.LeyLineStrength < 0.1 {
		return nil
	}

	deposits := make([]*MineralDeposit, 0)

	// Calculate crystal potency based on ley line strength
	potency := ctx.LeyLineStrength
	if ctx.IsIntersection {
		potency *= 2.0 // Intersections amplify magical energy
	}
	if potency > 1.0 {
		potency = 1.0
	}

	// Determine crystal type based on potency
	var crystalName string
	switch {
	case potency >= 0.8:
		crystalName = "Greater Mana Crystal"
	case potency >= 0.5:
		crystalName = "Mana Crystal"
	default:
		crystalName = "Minor Mana Crystal"
	}

	deposit := &MineralDeposit{
		DepositID: uuid.New(),
		MineralType: MineralType{
			Name:          crystalName,
			FormationType: FormationMetamorphic, // Magically transformed
			BaseValue:     int(potency * 100),
		},
		FormationType: FormationMetamorphic,
		Quantity:      int(potency*10) + 1,
		Concentration: potency,
		VeinSize:      VeinSizeSmall,
		Depth:         ctx.Depth,
	}
	deposits = append(deposits, deposit)

	return deposits
}

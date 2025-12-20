package underground

import (
	"math/rand"

	"github.com/google/uuid"
)

// DepositType represents the stage of organic deposit transformation
type DepositType string

const (
	DepositRemains      DepositType = "remains"      // Fresh remains
	DepositMineralizing DepositType = "mineralizing" // Being replaced by minerals
	DepositFossil       DepositType = "fossil"       // Fully fossilized
	DepositOil          DepositType = "oil"          // Transformed to oil
	DepositCoal         DepositType = "coal"         // Transformed to coal (from plants)
)

// OrganicDepositConfig controls deposit transformation timeline
// Timeline is reality÷10 as per user spec
type OrganicDepositConfig struct {
	BurialDepthForMineralization float64 // Meters below surface to start mineralizing
	MineralizationAge            int64   // Years to complete mineralization (reality: 10K-1M, config: 1K-100K)
	FossilizationAge             int64   // Years to become fossil (reality: 10M+, config: 1M+)
	OilFormationAge              int64   // Years for oil (reality: 50M+, config: 5M+)
	OilFormationDepth            float64 // Meters depth required for oil
	OilFormationTemp             float64 // Temperature required (Kelvin)
	SedimentRatePerYear          float64 // Meters of sediment per year (average)
}

// DefaultDepositConfig returns reality÷10 timeline configuration
func DefaultDepositConfig() OrganicDepositConfig {
	return OrganicDepositConfig{
		BurialDepthForMineralization: 10,        // 10m burial to start mineralizing
		MineralizationAge:            1_000,     // 1K years (reality: 10K)
		FossilizationAge:             100_000,   // 100K years (reality: 1M+)
		OilFormationAge:              5_000_000, // 5M years (reality: 50M+)
		OilFormationDepth:            3000,      // 3km deep
		OilFormationTemp:             373,       // 100°C (373K)
		SedimentRatePerYear:          0.0001,    // 0.1mm/year (0.1m per 1000 years)
	}
}

// CreateOrganicDeposit creates a new deposit when an entity dies
func CreateOrganicDeposit(
	entityID uuid.UUID,
	species string,
	x, y int,
	surfaceZ float64,
	deathYear int64,
	isPlant bool,
) Deposit {
	deposit := Deposit{
		ID:         uuid.New(),
		Type:       string(DepositRemains),
		DepthZ:     surfaceZ, // Starts at surface
		Quantity:   1.0,      // Base unit
		Discovered: false,
		Source: &OrganicSource{
			OriginalEntityID: entityID,
			Species:          species,
			DeathYear:        deathYear,
			BurialYear:       0, // Not buried yet
		},
	}

	// Plants become coal, animals become fossils/oil
	if isPlant {
		deposit.Type = string(DepositRemains) + "_plant"
	}

	return deposit
}

// SimulateDepositEvolution processes organic deposit transformation
func SimulateDepositEvolution(
	columns *ColumnGrid,
	currentYear int64,
	config OrganicDepositConfig,
	rainfallMap []float64, // For variable sedimentation
	seed int64,
) {
	rng := rand.New(rand.NewSource(seed))

	for _, col := range columns.AllColumns() {
		// Apply sedimentation to column
		sedimentThisYear := calculateSedimentation(col, rainfallMap, config, rng)

		// Process each deposit
		for i := range col.Resources {
			deposit := &col.Resources[i]
			if deposit.Source == nil {
				continue // Not an organic deposit
			}

			// Bury deposits deeper with sediment accumulation
			deposit.DepthZ -= sedimentThisYear

			// Check for burial
			if deposit.Source.BurialYear == 0 && deposit.DepthZ < col.Surface-1 {
				deposit.Source.BurialYear = currentYear
			}

			// Transform based on age, depth, and conditions
			transformDeposit(deposit, col, currentYear, config)
		}
	}
}

// calculateSedimentation determines sediment rate for a column
func calculateSedimentation(col *WorldColumn, rainfallMap []float64, config OrganicDepositConfig, rng *rand.Rand) float64 {
	baseRate := config.SedimentRatePerYear

	// More sediment in areas with water runoff
	idx := col.Y*100 + col.X // Approximate - should use actual grid width
	if idx < len(rainfallMap) {
		baseRate *= (1 + rainfallMap[idx])
	}

	// Variation
	baseRate *= (0.8 + rng.Float64()*0.4) // 80-120% of base

	return baseRate
}

// transformDeposit handles the transformation pipeline
func transformDeposit(deposit *Deposit, col *WorldColumn, currentYear int64, config OrganicDepositConfig) {
	if deposit.Source == nil {
		return
	}

	age := deposit.Source.Age(currentYear)
	burialDuration := deposit.Source.BurialDuration(currentYear)
	depth := col.Surface - deposit.DepthZ // Depth below surface

	currentType := DepositType(deposit.Type)
	isPlant := deposit.Type == string(DepositRemains)+"_plant"

	// Get temperature at depth (geothermal gradient: ~25°C per km depth)
	tempAtDepth := 288.0 + (depth/1000.0)*25.0 // 288K = 15°C surface temp

	switch currentType {
	case DepositRemains, DepositType(string(DepositRemains) + "_plant"):
		// Start mineralization if buried deep enough and old enough
		if burialDuration > 0 && depth >= config.BurialDepthForMineralization {
			if age >= config.MineralizationAge/10 { // Early stage
				if isPlant {
					deposit.Type = string(DepositMineralizing) + "_plant"
				} else {
					deposit.Type = string(DepositMineralizing)
				}
			}
		}

	case DepositMineralizing, DepositType(string(DepositMineralizing) + "_plant"):
		// Complete fossilization
		if age >= config.FossilizationAge {
			if isPlant || deposit.Type == string(DepositMineralizing)+"_plant" {
				deposit.Type = string(DepositCoal)
			} else {
				deposit.Type = string(DepositFossil)
			}
		}

	case DepositFossil:
		// Fossil to oil transformation (only animals, not plants)
		if age >= config.OilFormationAge &&
			depth >= config.OilFormationDepth &&
			tempAtDepth >= config.OilFormationTemp {
			// Check if organic-rich (originally larger creatures)
			if isOrganicRich(deposit.Source.Species) {
				deposit.Type = string(DepositOil)
				deposit.Quantity *= 100 // Oil quantity multiplier
			}
		}
	}
}

// isOrganicRich determines if a species produces oil vs just fossils
func isOrganicRich(species string) bool {
	// Marine life and larger organisms produce more organic material
	oilProducers := map[string]bool{
		"fish":      true,
		"whale":     true,
		"plankton":  true,
		"algae":     true,
		"kelp":      true,
		"shark":     true,
		"squid":     true,
		"jellyfish": true,
		"dinosaur":  true,
		"mammoth":   true,
	}
	return oilProducers[species]
}

// GetDepositByType retrieves deposits of a specific type from a column
func (col *WorldColumn) GetDepositByType(depositType DepositType) []Deposit {
	result := []Deposit{}
	for _, d := range col.Resources {
		if d.Type == string(depositType) {
			result = append(result, d)
		}
	}
	return result
}

// AddOrganicDeposit adds a new organic deposit to a column
func (col *WorldColumn) AddOrganicDeposit(deposit Deposit) {
	col.Resources = append(col.Resources, deposit)
}

// GetFossils returns all fossilized deposits
func (col *WorldColumn) GetFossils() []Deposit {
	return col.GetDepositByType(DepositFossil)
}

// GetOil returns all oil deposits
func (col *WorldColumn) GetOil() []Deposit {
	return col.GetDepositByType(DepositOil)
}

// GetCoal returns all coal deposits
func (col *WorldColumn) GetCoal() []Deposit {
	return col.GetDepositByType(DepositCoal)
}

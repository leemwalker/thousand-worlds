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

// SimulateSedimentDeposition accumulates sediment layers on a grid over time.
// Given a delta deposition rate and time period, new sedimentary strata form.
// Returns the total sediment thickness deposited.
func SimulateSedimentDeposition(grid *ColumnGrid, deltaDeposit float64, years int64) float64 {
	// Calculate total thickness: rate * time
	totalThickness := deltaDeposit * float64(years)

	// Apply to all columns
	for _, col := range grid.AllColumns() {
		// Create new sediment layer on top
		newLayer := StrataLayer{
			TopZ:     col.Surface,
			BottomZ:  col.Surface - totalThickness,
			Material: "sediment",
			Hardness: 2.0, // Soft sediment
			Age:      0,   // Freshly deposited
			Porosity: 0.4, // High porosity for fresh sediment
		}

		// Add layer to strata (push down existing layers conceptually)
		if len(col.Strata) > 0 {
			col.Strata = append([]StrataLayer{newLayer}, col.Strata...)
		} else {
			col.Strata = []StrataLayer{newLayer}
		}
	}

	return totalThickness
}

// GeodeType represents different geode varieties
type GeodeType string

const (
	GeodeAmethyst   GeodeType = "amethyst"
	GeodeQuartz     GeodeType = "quartz"
	GeodeAgate      GeodeType = "agate"
	GeodeChalcedony GeodeType = "chalcedony"
)

// Geode represents a crystal-filled void
type Geode struct {
	ID       uuid.UUID
	Type     GeodeType
	Location Vector3
	Radius   float64
	Quality  float64 // 0-1, crystal quality
}

// GenerateGeodes creates geodes in volcanic strata with fluid-filled voids.
// Geodes form when mineral-rich water slowly crystallizes inside gas bubbles.
func GenerateGeodes(grid *ColumnGrid, seed int64) []Geode {
	geodes := make([]Geode, 0)
	rng := rand.New(rand.NewSource(seed))

	for _, col := range grid.AllColumns() {
		// Check if column has volcanic material
		hasVolcanic := false
		for _, layer := range col.Strata {
			if layer.Material == "basalt" || layer.Material == "volcanic" {
				hasVolcanic = true
				break
			}
		}

		if !hasVolcanic {
			continue
		}

		// Check for fluid conditions (porosity > 0.2)
		hasFluid := false
		for _, layer := range col.Strata {
			if layer.Porosity > 0.2 {
				hasFluid = true
				break
			}
		}

		if !hasFluid {
			continue
		}

		// Generate geodes based on conditions
		numGeodes := 1 + rng.Intn(3)
		for i := 0; i < numGeodes; i++ {
			geodeTypes := []GeodeType{GeodeAmethyst, GeodeQuartz, GeodeAgate, GeodeChalcedony}
			geode := Geode{
				ID:       uuid.New(),
				Type:     geodeTypes[rng.Intn(len(geodeTypes))],
				Location: Vector3{X: float64(col.X), Y: float64(col.Y), Z: col.Surface - 10 - rng.Float64()*50},
				Radius:   0.1 + rng.Float64()*0.5,
				Quality:  0.3 + rng.Float64()*0.7,
			}
			geodes = append(geodes, geode)
		}
	}

	return geodes
}

// SimulateFaulting shifts strata layers at a fault location by a slip amount.
// This simulates earthquake-induced vertical displacement.
// Returns the number of affected columns.
func SimulateFaulting(grid *ColumnGrid, faultX int, slip float64) int {
	affected := 0

	for _, col := range grid.AllColumns() {
		// Apply fault to columns at or beyond fault line
		if col.X >= faultX {
			// Shift all strata by slip amount
			for i := range col.Strata {
				col.Strata[i].TopZ += slip
				col.Strata[i].BottomZ += slip
			}
			col.Surface += slip
			affected++
		}
	}

	return affected
}

// LeyLineNode represents a point where magical ley lines intersect underground
type LeyLineNode struct {
	ID          uuid.UUID
	Location    Vector3
	Power       float64 // Magical energy level (0-100)
	Connections int     // Number of ley lines meeting here
}

// GenerateLeyLineNodes creates magical nodes at underground intersection points.
// Nodes form where multiple ley lines cross, accumulating magical energy.
func GenerateLeyLineNodes(grid *ColumnGrid, magicLevel float64, seed int64) []LeyLineNode {
	nodes := make([]LeyLineNode, 0)

	// Only generate nodes if magic level is sufficient
	if magicLevel < 0.1 {
		return nodes
	}

	rng := rand.New(rand.NewSource(seed))

	// Generate nodes based on grid intersection patterns
	gridWidth := grid.Width
	gridHeight := grid.Height

	// Create nodes at regular intervals, adjusted by magic level
	interval := int(10 / (magicLevel + 0.1))
	if interval < 1 {
		interval = 1
	}

	for x := interval; x < gridWidth; x += interval {
		for y := interval; y < gridHeight; y += interval {
			col := grid.Get(x, y)
			if col == nil {
				continue
			}

			node := LeyLineNode{
				ID:          uuid.New(),
				Location:    Vector3{X: float64(x), Y: float64(y), Z: col.Surface - 50 - rng.Float64()*100},
				Power:       magicLevel * (50 + rng.Float64()*50),
				Connections: 2 + rng.Intn(4),
			}
			nodes = append(nodes, node)
		}
	}

	return nodes
}

package minerals

import (
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// FormationType defines how a mineral is formed
type FormationType string

const (
	FormationIgneous     FormationType = "igneous"
	FormationSedimentary FormationType = "sedimentary"
	FormationMetamorphic FormationType = "metamorphic"
)

// VeinShape defines the geometric shape of the deposit
type VeinShape string

const (
	VeinShapeLinear    VeinShape = "linear"
	VeinShapePlanar    VeinShape = "planar"
	VeinShapeSpherical VeinShape = "spherical"
)

// VeinSize defines the relative size category
type VeinSize string

const (
	VeinSizeSmall   VeinSize = "small"
	VeinSizeMedium  VeinSize = "medium"
	VeinSizeLarge   VeinSize = "large"
	VeinSizeMassive VeinSize = "massive"
)

// MineralType represents a specific type of mineral
type MineralType struct {
	Name          string
	FormationType FormationType
	BaseValue     int // Relative value per unit
	Hardness      float64
}

// Common Mineral Types
var (
	// Igneous
	MineralGold    = MineralType{Name: "Gold", FormationType: FormationIgneous, BaseValue: 100, Hardness: 2.5}
	MineralSilver  = MineralType{Name: "Silver", FormationType: FormationIgneous, BaseValue: 50, Hardness: 2.5}
	MineralCopper  = MineralType{Name: "Copper", FormationType: FormationIgneous, BaseValue: 20, Hardness: 3.0}
	MineralDiamond = MineralType{Name: "Diamond", FormationType: FormationIgneous, BaseValue: 500, Hardness: 10.0}
	MineralBasalt  = MineralType{Name: "Basalt", FormationType: FormationIgneous, BaseValue: 1, Hardness: 6.0}
	MineralGranite = MineralType{Name: "Granite", FormationType: FormationIgneous, BaseValue: 2, Hardness: 7.0}

	// Sedimentary
	MineralCoal      = MineralType{Name: "Coal", FormationType: FormationSedimentary, BaseValue: 5, Hardness: 2.0}
	MineralLimestone = MineralType{Name: "Limestone", FormationType: FormationSedimentary, BaseValue: 2, Hardness: 3.0}
	MineralSandstone = MineralType{Name: "Sandstone", FormationType: FormationSedimentary, BaseValue: 2, Hardness: 6.0}
	MineralSalt      = MineralType{Name: "Salt", FormationType: FormationSedimentary, BaseValue: 3, Hardness: 2.5}

	// Metamorphic
	MineralMarble   = MineralType{Name: "Marble", FormationType: FormationMetamorphic, BaseValue: 10, Hardness: 3.0}
	MineralIron     = MineralType{Name: "Iron", FormationType: FormationMetamorphic, BaseValue: 15, Hardness: 4.0}
	MineralRuby     = MineralType{Name: "Ruby", FormationType: FormationMetamorphic, BaseValue: 200, Hardness: 9.0}
	MineralSapphire = MineralType{Name: "Sapphire", FormationType: FormationMetamorphic, BaseValue: 200, Hardness: 9.0}
	MineralEmerald  = MineralType{Name: "Emerald", FormationType: FormationMetamorphic, BaseValue: 250, Hardness: 7.5}
	MineralPlatinum = MineralType{Name: "Platinum", FormationType: FormationIgneous, BaseValue: 300, Hardness: 4.5} // Often associated with deep igneous
)

// MineralDeposit represents a generated mineral deposit in the world
type MineralDeposit struct {
	DepositID     uuid.UUID
	MineralType   MineralType
	FormationType FormationType
	Location      geography.Point // Surface location
	Depth         float64         // Meters below surface
	Quantity      int             // Total extractable units
	Concentration float64         // 0.0 to 1.0 (ore grade)
	VeinSize      VeinSize
	GeologicalAge float64 // Million years old

	// Spatial extent
	VeinShape       VeinShape
	VeinOrientation geography.Vector // Direction of vein (for linear)
	VeinLength      float64          // Meters
	VeinWidth       float64          // Meters

	// Discovery
	SurfaceVisible bool    // Can be seen without mining
	RequiredDepth  float64 // Min mining depth to reach
}

// TectonicContext provides geological context for mineral formation
type TectonicContext struct {
	PlateBoundaryType  geography.BoundaryType
	MagmaFlowDirection geography.Vector
	FaultLineDirection geography.Vector
	ErosionLevel       float64 // 0.0 to 1.0
	Age                float64 // Million years
	Elevation          float64 // Meters
	IsVolcanic         bool
	IsSedimentaryBasin bool
}

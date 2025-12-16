// Package ecosystem defines supernatural powers that can be granted to species.
package ecosystem

// PowerType categorizes supernatural abilities
type PowerType string

const (
	PowerTypePsychic    PowerType = "psychic"
	PowerTypeBiological PowerType = "biological"
	PowerTypeElemental  PowerType = "elemental"
	PowerTypeMagical    PowerType = "magical"
	PowerTypePlant      PowerType = "plant"
)

// Power represents a supernatural ability that can be granted to a species
type Power struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        PowerType `json:"type"`
	Description string    `json:"description"`
	Cost        int       `json:"cost"` // Divine Energy cost to grant
}

// DefaultPowers returns all available supernatural powers
var DefaultPowers = []Power{
	// Psychic Powers
	{ID: "telekinesis", Name: "Telekinesis", Type: PowerTypePsychic, Description: "Move objects without physical touch", Cost: 60},
	{ID: "telepathy", Name: "Telepathy", Type: PowerTypePsychic, Description: "Mind-to-mind communication", Cost: 50},
	{ID: "empathy", Name: "Empathy", Type: PowerTypePsychic, Description: "Sense emotions of others", Cost: 30},
	{ID: "precognition", Name: "Precognition", Type: PowerTypePsychic, Description: "Danger sense and future glimpses", Cost: 70},
	{ID: "teleportation", Name: "Teleportation", Type: PowerTypePsychic, Description: "Short-range instant spatial travel", Cost: 80},

	// Biological Powers
	{ID: "fast_healing", Name: "Fast Healing", Type: PowerTypeBiological, Description: "Rapid wound recovery", Cost: 40},
	{ID: "regeneration", Name: "Regeneration", Type: PowerTypeBiological, Description: "Regrow lost limbs and organs", Cost: 70},
	{ID: "healing_aura", Name: "Healing Aura", Type: PowerTypeBiological, Description: "Heal nearby organisms", Cost: 55},
	{ID: "camouflage", Name: "Camouflage", Type: PowerTypeBiological, Description: "Active invisibility through skin adaptation", Cost: 45},
	{ID: "bioluminescence", Name: "Bioluminescence", Type: PowerTypeBiological, Description: "Generate light from body", Cost: 25},
	{ID: "venom", Name: "Venom", Type: PowerTypeBiological, Description: "Produce toxic secretions", Cost: 35},
	{ID: "echolocation", Name: "Echolocation", Type: PowerTypeBiological, Description: "Navigate and hunt via sound waves", Cost: 30},
	{ID: "flight", Name: "Flight", Type: PowerTypeBiological, Description: "Powered flight without wings", Cost: 65},
	{ID: "aquatic_breathing", Name: "Aquatic Breathing", Type: PowerTypeBiological, Description: "Breathe underwater", Cost: 35},

	// Elemental Powers
	{ID: "pyrokinesis", Name: "Pyrokinesis", Type: PowerTypeElemental, Description: "Generate and control fire", Cost: 60},
	{ID: "cryokinesis", Name: "Cryokinesis", Type: PowerTypeElemental, Description: "Generate and control ice/cold", Cost: 60},
	{ID: "electrokinesis", Name: "Electrokinesis", Type: PowerTypeElemental, Description: "Generate and control electricity", Cost: 65},

	// Plant Powers
	{ID: "medicinal", Name: "Medicinal Properties", Type: PowerTypePlant, Description: "Cures disease when consumed", Cost: 40},

	// Magical Powers
	{ID: "magic_affinity", Name: "Magic Affinity", Type: PowerTypeMagical, Description: "Harness ambient magical energy", Cost: 75},
}

// GetPowerByID returns a power by its ID
func GetPowerByID(id string) *Power {
	for _, p := range DefaultPowers {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// GetPowersByType returns all powers of a given type
func GetPowersByType(powerType PowerType) []Power {
	var powers []Power
	for _, p := range DefaultPowers {
		if p.Type == powerType {
			powers = append(powers, p)
		}
	}
	return powers
}

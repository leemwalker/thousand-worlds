package appearance

import (
	"tw-backend/internal/character"
	"strings"
)

// ApplySpeciesTemplate modifies the description based on species traits
func ApplySpeciesTemplate(desc string, species string) string {
	var features []string

	switch species {
	case character.SpeciesElf:
		features = append(features, "pointed ears", "angular features")
		// Elves tend to be graceful/elegant
		if !strings.Contains(desc, "graceful") {
			features = append(features, "graceful bearing")
		}
	case character.SpeciesDwarf:
		features = append(features, "broad shoulders")
		// Dwarves often have beards (assuming male default for generic desc, or add "if male")
		// For now, generic dwarf traits
		features = append(features, "sturdy frame")
	case character.SpeciesHuman:
		// Humans are versatile, maybe no specific extra features enforced
	}

	if len(features) > 0 {
		return desc + " with " + strings.Join(features, ", ")
	}
	return desc
}

// GetSpeciesHeightRange returns min/max height in cm
func GetSpeciesHeightRange(species string) (int, int) {
	switch species {
	case character.SpeciesElf:
		return 160, 210 // Tall
	case character.SpeciesDwarf:
		return 120, 150 // Short
	case character.SpeciesHuman:
		return 150, 200
	}
	return 150, 180
}

package skills

import "math"

// Quality Tiers
const (
	QualityPoor       = 0
	QualityCommon     = 1
	QualityFine       = 2
	QualityMasterwork = 3
	QualityLegendary  = 4
)

// GetQualityTier returns the quality tier based on skill level
// Formula: floor(skill / 20)
func GetQualityTier(skillLevel int) int {
	if skillLevel < 0 {
		return QualityPoor
	}
	tier := int(math.Floor(float64(skillLevel) / 20.0))
	if tier > QualityLegendary {
		return QualityLegendary
	}
	return tier
}

// GetQualityName returns the string representation of a quality tier
func GetQualityName(tier int) string {
	switch tier {
	case QualityPoor:
		return "Poor"
	case QualityCommon:
		return "Common"
	case QualityFine:
		return "Fine"
	case QualityMasterwork:
		return "Masterwork"
	case QualityLegendary:
		return "Legendary"
	default:
		return "Unknown"
	}
}

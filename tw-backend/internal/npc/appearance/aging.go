package appearance

import (
	"fmt"
)

// Age Categories
const (
	AgeChild      = "Child"
	AgeYoungAdult = "Young Adult"
	AgeAdult      = "Adult"
	AgeMiddleAged = "Middle-Aged"
	AgeElder      = "Elder"
)

// GetAgeCategory determines category based on age percentage
func GetAgeCategory(age, lifespan int) string {
	pct := float64(age) / float64(lifespan)
	switch {
	case pct < 0.2:
		return AgeChild
	case pct < 0.4:
		return AgeYoungAdult
	case pct < 0.6:
		return AgeAdult
	case pct < 0.8:
		return AgeMiddleAged
	default:
		return AgeElder
	}
}

// ApplyAgeModifiers adds age-specific details
func ApplyAgeModifiers(baseDesc string, category string) string {
	var details string

	switch category {
	case AgeChild:
		details = "youthful with smooth skin and bright eyes"
	case AgeYoungAdult:
		details = "in their prime with clear features"
	case AgeAdult:
		details = "mature with some weathering"
	case AgeMiddleAged:
		details = "middle-aged with lines forming and some gray"
	case AgeElder:
		details = "elderly with deeply lined skin and stooped posture"
	}

	return fmt.Sprintf("%s. %s.", baseDesc, details)
}

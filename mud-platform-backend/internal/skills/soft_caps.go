package skills

// ApplySoftCap reduces XP gain if skill level exceeds soft cap
// Soft Cap = Attribute * 1.5
// If Level > Soft Cap, XP gain is halved (or reduced significantly)
func ApplySoftCap(xpGain float64, skillLevel int, attributeVal int) float64 {
	softCap := float64(attributeVal) * 1.5

	if float64(skillLevel) > softCap {
		// Requirement says "XP required doubles after soft cap"
		// This is equivalent to "XP gain halved"
		return xpGain * 0.5
	}

	return xpGain
}

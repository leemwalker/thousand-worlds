package evolution

// UpdateAtmosphere calculates O2 changes from flora biomass
func UpdateAtmosphere(currentO2 float64, floraBiomass float64) float64 {
	// O2 production proportional to flora biomass
	delta := floraBiomass * 0.001

	newO2 := currentO2 + delta

	// Clamp to 0.0 - 0.40 (0% - 40%)
	if newO2 < 0.0 {
		return 0.0
	}
	if newO2 > 0.40 {
		return 0.40
	}
	return newO2
}

// CalculateSolarLuminosity returns luminosity factor based on age of system
func CalculateSolarLuminosity(ageInBillionYears float64) float64 {
	// Linear increase: 8% per billion years
	return 1.0 + (ageInBillionYears * 0.08)
}

// TrophicDynamicsResult holds population changes after trophic simulation
type TrophicDynamicsResult struct {
	HerbivorePop       int
	FloraBiomass       float64
	StarvationOccurred bool
}

// SimulateTrophicDynamics calculates carrying capacity feedback
func SimulateTrophicDynamics(herbivores []*Species, floraBiomass float64, rainfall float64, temperature float64) TrophicDynamicsResult {
	// Calculate biomass capacity based on rainfall/temp
	biomassCapacity := rainfall * 0.01 * (1.0 - (temperature-20)/100.0)
	if biomassCapacity < 0 {
		biomassCapacity = 0
	}

	totalHerbivorePop := 0
	for _, h := range herbivores {
		totalHerbivorePop += h.Population
	}

	starvation := false
	if float64(totalHerbivorePop) > biomassCapacity*10000 { // Scaled capacity
		// Overpopulation -> starvation
		starvation = true
		for _, h := range herbivores {
			h.Population = h.Population / 2 // 50% die-off
		}
		totalHerbivorePop = totalHerbivorePop / 2
	}

	return TrophicDynamicsResult{
		HerbivorePop:       totalHerbivorePop,
		FloraBiomass:       floraBiomass,
		StarvationOccurred: starvation,
	}
}

// CheckSapienceEmergence evaluates if a species has become sapient
func CheckSapienceEmergence(intelligence float64, social float64) bool {
	return intelligence > 0.9 && social > 0.8
}

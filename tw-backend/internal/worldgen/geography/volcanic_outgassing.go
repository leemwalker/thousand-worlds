package geography

// VolcanicGasComposition represents the gases released by an eruption
type VolcanicGasComposition struct {
	CO2 float64 // Carbon Dioxide (Greenhouse gas)
	SO2 float64 // Sulfur Dioxide (Cooling aerosol)
	H2O float64 // Water Vapor (Greenhouse gas + Ocean source)
}

// CalculateVolcanicEmissions estimates the gas output of an eruption.
// magnitude: 0.0-1.0 (approximates VEI or volume scale)
// magmaType: "basaltic" (low gas), "andesitic" (med), "rhyolitic" (high)
func CalculateVolcanicEmissions(magnitude float64, magmaType string) VolcanicGasComposition {
	// Base volume in cubic km (DRE - Dense Rock Equivalent)
	// Magnitude 0.1 ~ VEI 3 (0.01 km3)
	// Magnitude 1.0 ~ VEI 8 (1000 km3)
	// Linear scaling for simplicity in simulation:
	// Volume = Base * 10^(magnitude * 4)
	// Actually, let's use a simpler linear-exponential proxy for game balance.
	volume := magnitude * 100.0 // Simplified volume unit

	comp := VolcanicGasComposition{}

	switch magmaType {
	case "basaltic": // Hotspots, Shield Volcanoes
		// Basalt: High CO2, Low SO2
		comp.CO2 = volume * 0.5
		comp.SO2 = volume * 0.1
		comp.H2O = volume * 0.4
	case "andesitic", "rhyolitic": // Subduction zones, Explosive
		// High SO2 (Explosive), Medium CO2
		comp.CO2 = volume * 0.3
		comp.SO2 = volume * 0.5
		comp.H2O = volume * 0.6
	default:
		comp.CO2 = volume * 0.4
		comp.SO2 = volume * 0.2
		comp.H2O = volume * 0.5
	}

	return comp
}

package damage

// ArmorEffectiveness maps WeaponType -> ArmorType -> Reduction Percentage (0.0 to 1.0)
var ArmorEffectiveness = map[WeaponType]map[ArmorType]float64{
	WeaponSlashing: {
		ArmorLeather:   0.15,
		ArmorChainMail: 0.35,
		ArmorPlate:     0.50,
	},
	WeaponPiercing: {
		ArmorLeather:   0.10,
		ArmorChainMail: 0.20,
		ArmorPlate:     0.40,
	},
	WeaponBludgeoning: {
		ArmorLeather:   0.05,
		ArmorChainMail: 0.15,
		ArmorPlate:     0.25,
	},
	WeaponRanged: {
		// Assuming ranged behaves similar to piercing for now, or defined separately
		// Prompt didn't explicitly specify Ranged vs Armor chart, but usually similar to Piercing
		// Let's use Piercing values for now as a safe default or define specific if needed.
		// Prompt said "Piercing weapons balanced", let's assume Ranged is similar or slightly less effective vs plate.
		// Actually, let's stick to the prompt's explicit list. It listed Slashing, Piercing, Bludgeoning.
		// It didn't list Ranged in the chart. I will add reasonable defaults for Ranged.
		ArmorLeather:   0.10,
		ArmorChainMail: 0.20,
		ArmorPlate:     0.30, // Arrows struggle against plate
	},
}

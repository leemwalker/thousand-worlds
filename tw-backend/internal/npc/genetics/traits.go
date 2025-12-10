package genetics

import (
	"mud-platform-backend/internal/character"
)

// AttributeModifiers holds bonuses to be applied to base stats
type AttributeModifiers struct {
	Attributes character.Attributes
}

// CalculateAttributeBonuses derives stats from DNA
func CalculateAttributeBonuses(dna DNA) AttributeModifiers {
	mods := AttributeModifiers{}

	// Helper to get bonus from gene
	// SS (+10), Ss (+5), ss (+0)
	getBonus := func(geneName string, homozygousBonus, heterozygousBonus int) int {
		if g, ok := dna.Genes[geneName]; ok {
			// Check genotype
			if g.Allele1 == g.Allele2 {
				if g.IsDominant1 {
					return homozygousBonus // AA
				}
				return 0 // aa
			}
			return heterozygousBonus // Aa
		}
		return 0
	}

	// Physical
	// Might: Strength (S/s) + Muscle (M/m)
	// S: +10/+5, M: +8/+4
	mods.Attributes.Might += getBonus(GeneStrength, 10, 5)
	mods.Attributes.Might += getBonus(GeneMuscle, 8, 4)

	// Agility: Reflex (R/r) + Coord (C/c)
	mods.Attributes.Agility += getBonus(GeneReflex, 10, 5)
	mods.Attributes.Agility += getBonus(GeneCoord, 8, 4)

	// Endurance: Stamina (E/e) + Resilience (L/l)
	mods.Attributes.Endurance += getBonus(GeneStamina, 10, 5)
	mods.Attributes.Endurance += getBonus(GeneResilience, 8, 4)

	// Vitality: Health (H/h) + Recovery (V/v)
	mods.Attributes.Vitality += getBonus(GeneHealth, 10, 5)
	mods.Attributes.Vitality += getBonus(GeneRecovery, 8, 4)

	// Mental
	// Intellect: Cognition (I/i) + Learning (K/k)
	mods.Attributes.Intellect += getBonus(GeneCognition, 10, 5)
	mods.Attributes.Intellect += getBonus(GeneLearning, 8, 4)

	// Cunning: Perception (P/p) + Analysis (A/a)
	mods.Attributes.Cunning += getBonus(GenePerception, 10, 5)
	mods.Attributes.Cunning += getBonus(GeneAnalysis, 8, 4)

	// Sensory
	// Sight: Vision (Vi/vi) + Color (Co/co)
	mods.Attributes.Sight += getBonus(GeneVision, 10, 5)
	mods.Attributes.Sight += getBonus(GeneColor, 5, 2) // Color less impact on raw sight?

	// Hearing: Auditory (Au/au) + Range (Ra/ra)
	mods.Attributes.Hearing += getBonus(GeneAuditory, 10, 5)
	mods.Attributes.Hearing += getBonus(GeneRange, 8, 4)

	return mods
}

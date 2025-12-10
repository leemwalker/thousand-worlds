package personality

import (
	"math/rand"
	"mud-platform-backend/internal/npc/genetics"
)

// Personality Gene Keys
const (
	GeneOpenness          = "gene_openness"
	GeneConscientiousness = "gene_conscientiousness"
	GeneExtraversion      = "gene_extraversion"
	GeneAgreeableness     = "gene_agreeableness"
	GeneNeuroticism       = "gene_neuroticism"
)

// DeriveFromGenetics calculates the genetic baseline for personality
func DeriveFromGenetics(dna genetics.DNA) *Personality {
	p := NewPersonality()

	// Helper to roll from range
	roll := func(min, max float64) float64 {
		return min + rand.Float64()*(max-min)
	}

	// Helper to get score from gene
	// AA (Dominant/Dominant) = 70-90
	// Aa (Dominant/Recessive) = 50-70
	// aa (Recessive/Recessive) = 20-50
	getScore := func(geneName string) float64 {
		if g, ok := dna.Genes[geneName]; ok {
			if g.Allele1 == g.Allele2 {
				if g.IsDominant1 {
					// Homozygous Dominant (AA)
					return roll(70, 90)
				}
				// Homozygous Recessive (aa)
				return roll(20, 50)
			}
			// Heterozygous (Aa)
			return roll(50, 70)
		}
		// Default if gene missing: Average
		return roll(40, 60)
	}

	p.Openness.Value = getScore(GeneOpenness)
	p.Conscientiousness.Value = getScore(GeneConscientiousness)
	p.Extraversion.Value = getScore(GeneExtraversion)
	p.Agreeableness.Value = getScore(GeneAgreeableness)
	p.Neuroticism.Value = getScore(GeneNeuroticism)

	return p
}

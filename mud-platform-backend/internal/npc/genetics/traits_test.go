package genetics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateAttributeBonuses(t *testing.T) {
	dna := NewDNA()
	// Strength: SS (+10)
	dna.Genes[GeneStrength] = NewGene(GeneStrength, "S", "S")
	// Muscle: Mm (+4)
	dna.Genes[GeneMuscle] = NewGene(GeneMuscle, "M", "m")

	mods := CalculateAttributeBonuses(dna)

	assert.Equal(t, 14, mods.Attributes.Might) // 10 + 4
	assert.Equal(t, 0, mods.Attributes.Agility)
}

func TestGenerateAppearance(t *testing.T) {
	dna := NewDNA()
	dna.Genes[GeneHeight] = NewGene(GeneHeight, "T", "T") // Tall
	dna.Genes[GeneBuild] = NewGene(GeneBuild, "b", "b")   // Lean
	dna.Genes[GeneHair] = NewGene(GeneHair, "h", "h")     // Blonde
	dna.Genes[GeneEye] = NewGene(GeneEye, "e", "e")       // Blue

	desc := GenerateAppearance(dna)

	assert.Contains(t, desc, "Tall")
	assert.Contains(t, desc, "lean")
	assert.Contains(t, desc, "blonde hair")
	assert.Contains(t, desc, "blue eyes")
}

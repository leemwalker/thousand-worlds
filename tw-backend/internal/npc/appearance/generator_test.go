package appearance

import (
	"testing"

	"tw-backend/internal/character"
	"tw-backend/internal/npc/genetics"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAppearance(t *testing.T) {
	dna := genetics.NewDNA()
	// Setup Genes
	dna.Genes[genetics.GeneHeight] = genetics.NewGene(genetics.GeneHeight, "T", "T")   // Tall
	dna.Genes[genetics.GeneBuild] = genetics.NewGene(genetics.GeneBuild, "B", "B")     // Broad
	dna.Genes[genetics.GeneMuscle] = genetics.NewGene(genetics.GeneMuscle, "M", "M")   // Muscular
	dna.Genes[genetics.GeneHair] = genetics.NewGene(genetics.GeneHair, "H", "H")       // Dark
	dna.Genes[genetics.GenePigment] = genetics.NewGene(genetics.GenePigment, "P", "P") // Strong -> Black
	dna.Genes[genetics.GeneEye] = genetics.NewGene(genetics.GeneEye, "e", "e")         // Blue
	dna.Genes[genetics.GeneMelanin] = genetics.NewGene(genetics.GeneMelanin, "m", "m") // Low -> Pale Blue

	desc := GenerateAppearance(dna, 30, 100, character.SpeciesHuman)

	assert.Contains(t, desc.FullDescription, "Tall")
	assert.Contains(t, desc.FullDescription, "muscular")
	assert.Contains(t, desc.FullDescription, "black hair")
	assert.Contains(t, desc.FullDescription, "pale blue eyes")
	assert.Contains(t, desc.FullDescription, "in their prime") // Young Adult (30/100)
}

func TestGenerateAppearance_Elf(t *testing.T) {
	dna := genetics.NewDNA()
	// Defaults
	dna.Genes[genetics.GeneHeight] = genetics.NewGene(genetics.GeneHeight, "t", "T")   // Average
	dna.Genes[genetics.GeneBuild] = genetics.NewGene(genetics.GeneBuild, "b", "b")     // Lean
	dna.Genes[genetics.GeneMuscle] = genetics.NewGene(genetics.GeneMuscle, "m", "m")   // Low muscle
	dna.Genes[genetics.GeneHair] = genetics.NewGene(genetics.GeneHair, "h", "h")       // Light
	dna.Genes[genetics.GenePigment] = genetics.NewGene(genetics.GenePigment, "p", "p") // Weak -> Platinum Blonde
	dna.Genes[genetics.GeneEye] = genetics.NewGene(genetics.GeneEye, "E", "E")         // Brown
	dna.Genes[genetics.GeneMelanin] = genetics.NewGene(genetics.GeneMelanin, "M", "M") // High -> Dark Brown

	desc := GenerateAppearance(dna, 100, 500, character.SpeciesElf) // 100/500 = 0.2 -> Young Adult

	assert.Contains(t, desc.FullDescription, "pointed ears")
	assert.Contains(t, desc.FullDescription, "graceful bearing")
	assert.Contains(t, desc.FullDescription, "platinum blonde hair")
}

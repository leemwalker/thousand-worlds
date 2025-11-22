package genetics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInheritGene_PunnettSquare(t *testing.T) {
	// Parent 1: AA (Homozygous Dominant)
	p1 := NewGene("test", "A", "A")
	// Parent 2: aa (Homozygous Recessive)
	p2 := NewGene("test", "a", "a")

	// Child should ALWAYS be Aa
	child := InheritGene(p1, p2)
	assert.True(t, (child.Allele1 == "A" && child.Allele2 == "a") || (child.Allele1 == "a" && child.Allele2 == "A"))
	assert.True(t, child.IsDominant1 || child.IsDominant2)
	assert.Equal(t, "A", child.Phenotype)
}

func TestInheritGene_Heterozygous(t *testing.T) {
	// P1: Aa, P2: Aa
	p1 := NewGene("test", "A", "a")
	p2 := NewGene("test", "A", "a")

	// Run simulation to verify ratios (approximate)
	aaCount := 0
	AaCount := 0
	AACount := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		c := InheritGene(p1, p2)
		if c.Allele1 == "a" && c.Allele2 == "a" {
			aaCount++
		} else if c.Allele1 == "A" && c.Allele2 == "A" {
			AACount++
		} else {
			AaCount++
		}
	}

	// Expected: 25% AA, 50% Aa, 25% aa
	assert.InDelta(t, 0.25, float64(AACount)/float64(iterations), 0.05)
	assert.InDelta(t, 0.50, float64(AaCount)/float64(iterations), 0.05)
	assert.InDelta(t, 0.25, float64(aaCount)/float64(iterations), 0.05)
}

func TestInherit_FullDNA(t *testing.T) {
	d1 := NewDNA()
	d1.Genes["height"] = NewGene("height", "T", "T")

	d2 := NewDNA()
	d2.Genes["height"] = NewGene("height", "t", "t")

	child, err := Inherit(d1, d2)
	assert.NoError(t, err)
	assert.Contains(t, child.Genes, "height")
	assert.Equal(t, "T", child.Genes["height"].Phenotype)
}

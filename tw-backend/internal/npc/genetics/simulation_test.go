package genetics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimulation_MultiGeneration(t *testing.T) {
	// Gen 0: 100 AA, 100 aa
	popSize := 200
	population := make([]DNA, popSize)

	for i := 0; i < 100; i++ {
		d := NewDNA()
		d.Genes["trait"] = NewGene("trait", "A", "A")
		population[i] = d
	}
	for i := 100; i < 200; i++ {
		d := NewDNA()
		d.Genes["trait"] = NewGene("trait", "a", "a")
		population[i] = d
	}

	// Since I can't easily do a full statistical simulation in a unit test without flakiness,
	// I will test a specific scenario:
	// AA x aa -> All Aa (Gen 1)
	// Aa x Aa -> 1:2:1 (Gen 2)

	// Gen 1
	p1 := NewDNA()
	p1.Genes["trait"] = NewGene("trait", "A", "A")
	p2 := NewDNA()
	p2.Genes["trait"] = NewGene("trait", "a", "a")

	gen1 := make([]DNA, 100)
	for i := 0; i < 100; i++ {
		child, _ := Inherit(p1, p2)
		gen1[i] = child
		// Assert all are Aa
		// Note: Mutation might change this! 5% rate.
		// So we expect ~95% Aa.
	}

	aaCount := 0
	AaCount := 0
	AACount := 0

	for _, d := range gen1 {
		g := d.Genes["trait"]
		if g.Allele1 == "a" && g.Allele2 == "a" {
			aaCount++
		} else if g.Allele1 == "A" && g.Allele2 == "A" {
			AACount++
		} else {
			AaCount++
		}
	}

	// With mutation, we might get some AA or aa from A->a or a->A flips.
	// But mostly Aa.
	assert.True(t, AaCount > 80, "Expected mostly Aa in Gen 1")

	// Gen 2: Breed Gen 1 (Aa) with Gen 1 (Aa)
	gen2 := make([]DNA, 1000)
	parentAa := NewDNA()
	parentAa.Genes["trait"] = NewGene("trait", "A", "a")

	for i := 0; i < 1000; i++ {
		child, _ := Inherit(parentAa, parentAa)
		gen2[i] = child
	}

	aaCount = 0
	AaCount = 0
	AACount = 0

	for _, d := range gen2 {
		g := d.Genes["trait"]
		if g.Allele1 == "a" && g.Allele2 == "a" {
			aaCount++
		} else if g.Allele1 == "A" && g.Allele2 == "A" {
			AACount++
		} else {
			AaCount++
		}
	}

	// Expect 1:2:1 approx
	// AA ~ 250, Aa ~ 500, aa ~ 250
	// Allow wide margin for randomness + mutation
	assert.InDelta(t, 250, AACount, 50)
	assert.InDelta(t, 500, AaCount, 60)
	assert.InDelta(t, 250, aaCount, 50)
}

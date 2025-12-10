package genetics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMutate_Rate(t *testing.T) {
	gene := NewGene("test", "A", "A")
	mutations := 0
	iterations := 10000

	for i := 0; i < iterations; i++ {
		mutated := Mutate(gene)
		if mutated.Allele1 != "A" || mutated.Allele2 != "A" {
			mutations++
		}
	}

	// Expected ~5%
	rate := float64(mutations) / float64(iterations)
	assert.InDelta(t, 0.05, rate, 0.01)
}

func TestMutate_Types(t *testing.T) {
	gene := NewGene("test", "A", "A")

	// Force mutation by running until change
	changed := false
	for i := 0; i < 1000; i++ {
		m := Mutate(gene)
		if m.Allele1 != "A" || m.Allele2 != "A" {
			changed = true
			// Check if it's 'a' (flip) or 'M' (new)
			isFlip := (m.Allele1 == "a" || m.Allele2 == "a")
			isNew := (m.Allele1 == "M" || m.Allele2 == "M")
			assert.True(t, isFlip || isNew)
			break
		}
	}
	assert.True(t, changed)
}

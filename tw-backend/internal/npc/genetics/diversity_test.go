package genetics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateGeneticDistance(t *testing.T) {
	d1 := NewDNA()
	d1.Genes["A"] = NewGene("A", "A", "A")

	d2 := NewDNA()
	d2.Genes["A"] = NewGene("A", "a", "a")

	// AA vs aa -> 0 matches -> Distance 1.0
	dist := CalculateGeneticDistance(d1, d2)
	assert.Equal(t, 1.0, dist)

	d3 := NewDNA()
	d3.Genes["A"] = NewGene("A", "A", "a")

	// AA vs Aa -> 1 match (A) -> Distance 0.5
	dist = CalculateGeneticDistance(d1, d3)
	assert.Equal(t, 0.5, dist)

	d4 := NewDNA()
	d4.Genes["A"] = NewGene("A", "A", "A")

	// AA vs AA -> 2 matches -> Distance 0.0
	dist = CalculateGeneticDistance(d1, d4)
	assert.Equal(t, 0.0, dist)
}

func TestCheckCompatibility(t *testing.T) {
	d1 := NewDNA()
	d1.Genes["A"] = NewGene("A", "A", "A")

	// Identical -> Incompatible
	assert.False(t, CheckCompatibility(d1, d1))

	d2 := NewDNA()
	d2.Genes["A"] = NewGene("A", "a", "a")

	// Different -> Compatible
	assert.True(t, CheckCompatibility(d1, d2))
}

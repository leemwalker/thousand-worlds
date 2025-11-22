package genetics

import (
	"math/rand"
	"unicode"
)

// Inherit creates a child DNA profile from two parents
func Inherit(parent1, parent2 DNA) (DNA, error) {
	child := NewDNA()

	// Iterate over all genes in parent1 (assuming parents have same gene set)
	// In a real system, we might need to handle disparate gene sets, but for now assume compatibility.
	for trait, gene1 := range parent1.Genes {
		if gene2, ok := parent2.Genes[trait]; ok {
			childGene := InheritGene(gene1, gene2)
			child.Genes[trait] = childGene
		}
	}

	return child, nil
}

// InheritGene combines two parent genes using Punnett square logic
func InheritGene(g1, g2 Gene) Gene {
	// Select one allele from each parent randomly
	a1 := g1.Allele1
	if rand.Float64() < 0.5 {
		a1 = g1.Allele2
	}

	a2 := g2.Allele1
	if rand.Float64() < 0.5 {
		a2 = g2.Allele2
	}

	return NewGene(g1.TraitName, a1, a2)
}

// NewGene creates a gene and determines phenotype
func NewGene(trait, a1, a2 string) Gene {
	g := Gene{
		TraitName: trait,
		Allele1:   a1,
		Allele2:   a2,
	}

	g.IsDominant1 = isDominant(a1)
	g.IsDominant2 = isDominant(a2)

	// Determine Phenotype
	// Simple dominance: If either is dominant, express dominant.
	// If both recessive, express recessive.
	// If co-dominance is needed, we can add logic here.
	// For now, standard dominance.

	if g.IsDominant1 {
		g.Phenotype = a1
	} else if g.IsDominant2 {
		g.Phenotype = a2
	} else {
		g.Phenotype = a1 // Both recessive, they should be same trait type usually (e.g. 'a' and 'a')
	}

	return g
}

func isDominant(allele string) bool {
	if len(allele) == 0 {
		return false
	}
	r := rune(allele[0])
	return unicode.IsUpper(r)
}

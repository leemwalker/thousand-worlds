package genetics

import (
	"math/rand"
	"unicode"
)

const (
	MutationRate = 0.05 // 5% chance per gene
)

// Mutate applies random mutations to a gene
func Mutate(g Gene) Gene {
	if rand.Float64() > MutationRate {
		return g
	}

	// Mutation Types:
	// 1. Flip (50%): A -> a or a -> A
	// 2. New (30%): Introduce rare variant (e.g. B)
	// 3. Amplification (20%): A -> A+ (Not strictly supported by string schema, maybe just flip for now or ignore)
	// Let's stick to Flip and New for string alleles.

	roll := rand.Float64()

	// Apply mutation to ONE allele randomly
	targetAllele := &g.Allele1
	if rand.Float64() < 0.5 {
		targetAllele = &g.Allele2
	}

	if roll < 0.5 {
		// Flip Case
		*targetAllele = flipCase(*targetAllele)
	} else if roll < 0.8 {
		// New Variant (Rare)
		// Generate a new letter? Or just a different char.
		// Let's assume standard traits use A/a. Rare might be B/b.
		// For simplicity, let's just flip to a specific "mutated" char 'M' or 'm' depending on case.
		if unicode.IsUpper(rune((*targetAllele)[0])) {
			*targetAllele = "M"
		} else {
			*targetAllele = "m"
		}
	} else {
		// Amplification / Other
		// For now, treat as Flip
		*targetAllele = flipCase(*targetAllele)
	}

	// Re-evaluate dominance and phenotype
	g.IsDominant1 = isDominant(g.Allele1)
	g.IsDominant2 = isDominant(g.Allele2)

	if g.IsDominant1 {
		g.Phenotype = g.Allele1
	} else if g.IsDominant2 {
		g.Phenotype = g.Allele2
	} else {
		g.Phenotype = g.Allele1
	}

	return g
}

func flipCase(s string) string {
	if len(s) == 0 {
		return s
	}
	r := rune(s[0])
	if unicode.IsUpper(r) {
		return string(unicode.ToLower(r))
	}
	return string(unicode.ToUpper(r))
}

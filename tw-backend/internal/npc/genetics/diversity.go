package genetics

// CalculateGeneticDistance computes the difference between two DNA profiles
// Returns 0.0 (identical) to 1.0 (completely different)
func CalculateGeneticDistance(dna1, dna2 DNA) float64 {
	if len(dna1.Genes) == 0 || len(dna2.Genes) == 0 {
		return 1.0 // Assume different if empty? Or incompatible.
	}

	diffCount := 0
	totalGenes := 0

	// Iterate over union of keys? Or just assume same schema.
	// Let's iterate over dna1 keys.
	for trait, g1 := range dna1.Genes {
		totalGenes++
		if g2, ok := dna2.Genes[trait]; ok {
			// Compare alleles
			// Distance is 0 if identical alleles, 0.5 if one differs, 1.0 if both differ
			// Order doesn't matter for distance (Aa == aA)

			matchCount := 0
			// Check g1.A1 against g2.A1/A2
			if g1.Allele1 == g2.Allele1 {
				matchCount++
			} else if g1.Allele1 == g2.Allele2 {
				matchCount++
			}

			// Check g1.A2 against remaining
			// This is tricky. Simple set intersection.
			// Bag of alleles: {A, a} vs {A, A}
			// Intersection: {A} -> 1 match.
			// {A, a} vs {a, a} -> {a} -> 1 match.
			// {A, B} vs {C, D} -> 0 matches.

			// Let's use a simpler metric:
			// 0 matches = 1.0 dist
			// 1 match = 0.5 dist
			// 2 matches = 0.0 dist

			// Bag approach
			bag2 := map[string]int{}
			bag2[g2.Allele1]++
			bag2[g2.Allele2]++

			matches := 0
			if bag2[g1.Allele1] > 0 {
				matches++
				bag2[g1.Allele1]--
			}
			if bag2[g1.Allele2] > 0 {
				matches++
				bag2[g1.Allele2]--
			}

			if matches == 0 {
				diffCount += 2
			} else if matches == 1 {
				diffCount += 1
			}
			// 2 matches -> 0 diff
		} else {
			// Missing gene in dna2 -> max difference
			diffCount += 2
		}
	}

	// Normalize: Max diff is 2 * totalGenes
	if totalGenes == 0 {
		return 0.0
	}
	return float64(diffCount) / float64(2*totalGenes)
}

// CheckCompatibility determines if two individuals can breed healthily
// Requires genetic distance >= 0.2 (arbitrary threshold for "not siblings")
// Prompt says: "Minimum genetic distance: geneticSimilarity < 0.8"
// Similarity = 1 - Distance. So Distance > 0.2.
func CheckCompatibility(dna1, dna2 DNA) bool {
	dist := CalculateGeneticDistance(dna1, dna2)
	return dist > 0.2
}

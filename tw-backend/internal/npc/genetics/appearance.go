package genetics

import (
	"fmt"
	"strings"
)

// GenerateAppearance creates a descriptive string from DNA
func GenerateAppearance(dna DNA) string {
	var parts []string

	// Height
	// TT = Tall, Tt = Average, tt = Short
	if g, ok := dna.Genes[GeneHeight]; ok {
		if g.Allele1 == "T" && g.Allele2 == "T" {
			parts = append(parts, "Tall")
		} else if g.Allele1 == "t" && g.Allele2 == "t" {
			parts = append(parts, "Short")
		} else {
			parts = append(parts, "Average height")
		}
	}

	// Build
	// BB = Muscular, Bb = Average, bb = Lean
	if g, ok := dna.Genes[GeneBuild]; ok {
		if g.Allele1 == "B" && g.Allele2 == "B" {
			parts = append(parts, "muscular")
		} else if g.Allele1 == "b" && g.Allele2 == "b" {
			parts = append(parts, "lean")
		} else {
			parts = append(parts, "average build")
		}
	}

	// Hair Color
	// HrHr = Black, Hrhr = Brown, hrhr = Blonde
	// Pigment modifies? Let's stick to simple dominant/recessive for now.
	hairColor := "brown"
	if g, ok := dna.Genes[GeneHair]; ok {
		if g.Allele1 == "H" && g.Allele2 == "H" {
			hairColor = "black"
		} else if g.Allele1 == "h" && g.Allele2 == "h" {
			hairColor = "blonde"
		}
	}
	parts = append(parts, fmt.Sprintf("with %s hair", hairColor))

	// Eye Color
	// EyEy = Brown, Eyey = Hazel, eyey = Blue
	eyeColor := "hazel"
	if g, ok := dna.Genes[GeneEye]; ok {
		if g.Allele1 == "E" && g.Allele2 == "E" {
			eyeColor = "brown"
		} else if g.Allele1 == "e" && g.Allele2 == "e" {
			eyeColor = "blue"
		}
	}
	parts = append(parts, fmt.Sprintf("%s eyes", eyeColor))

	// Jaw
	if g, ok := dna.Genes[GeneJaw]; ok {
		if g.IsDominant1 || g.IsDominant2 {
			parts = append(parts, "a strong jaw")
		}
	}

	return strings.Join(parts, ", ")
}

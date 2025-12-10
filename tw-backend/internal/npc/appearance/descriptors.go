package appearance

import (
	"tw-backend/internal/npc/genetics"
)

// GetHeightDescriptor returns a string description based on height gene
// Assumes T/t alleles. TT=Tall, Tt=Average, tt=Short.
// We can add numeric ranges if we had actual height values, but for now based on genotype.
func GetHeightDescriptor(g genetics.Gene) string {
	// TT -> Very Tall / Towering
	// Tt -> Average / Tall
	// tt -> Short / Diminutive

	// Let's use allele combinations
	if g.Allele1 == "T" && g.Allele2 == "T" {
		return "tall"
	} else if g.Allele1 == "t" && g.Allele2 == "t" {
		return "short"
	}
	return "average height"
}

// GetBuildDescriptor combines Muscle and Build genes
func GetBuildDescriptor(buildGene, muscleGene genetics.Gene) string {
	// Build: B (Broad/Large frame), b (Narrow/Small frame)
	// Muscle: M (Muscular), m (Low muscle)

	// Let's assume B is dominant for Broad frame.
	// Actually, let's stick to the prompt:
	// BB/MM = muscular, Bb/Mm = average, bb/mm = lean

	// Let's score it:
	// B/M = 2 points, b/m = 0 points.
	// Score 0-4?

	score := 0
	if buildGene.Allele1 == "B" {
		score++
	}
	if buildGene.Allele2 == "B" {
		score++
	}
	if muscleGene.Allele1 == "M" {
		score++
	}
	if muscleGene.Allele2 == "M" {
		score++
	}

	switch {
	case score >= 3:
		return "muscular"
	case score == 2:
		return "average build"
	case score <= 1:
		return "lean"
	}
	return "average build"
}

// GetHairDescriptor returns texture and color
func GetHairDescriptor(hairGene, pigmentGene genetics.Gene) string {
	// Hair: H (Dark?), h (Light?)
	// Pigment: P (Strong), p (Weak)

	// Prompt: HrHr/PiPi = black, HrHr/Pipi = brown, hrhr/pipi = blonde

	isDarkHair := (hairGene.Allele1 == "H" && hairGene.Allele2 == "H")
	isLightHair := (hairGene.Allele1 == "h" && hairGene.Allele2 == "h")

	isStrongPigment := (pigmentGene.Allele1 == "P" && pigmentGene.Allele2 == "P")
	isWeakPigment := (pigmentGene.Allele1 == "p" && pigmentGene.Allele2 == "p")

	if isDarkHair {
		if isStrongPigment {
			return "black"
		}
		return "dark brown"
	}
	if isLightHair {
		if isWeakPigment {
			return "platinum blonde"
		}
		return "blonde"
	}

	// Heterozygous / Mixed
	if isStrongPigment {
		return "chestnut"
	}
	return "light brown"
}

// GetEyeDescriptor returns eye color
func GetEyeDescriptor(eyeGene, melaninGene genetics.Gene) string {
	// EyEy/MeMe = brown
	// eyey/meme = blue

	isBrown := (eyeGene.Allele1 == "E" && eyeGene.Allele2 == "E")
	isBlue := (eyeGene.Allele1 == "e" && eyeGene.Allele2 == "e")

	isHighMelanin := (melaninGene.Allele1 == "M" && melaninGene.Allele2 == "M")
	isLowMelanin := (melaninGene.Allele1 == "m" && melaninGene.Allele2 == "m")

	if isBrown {
		if isHighMelanin {
			return "dark brown"
		}
		return "brown"
	}
	if isBlue {
		if isLowMelanin {
			return "pale blue"
		}
		return "blue"
	}

	// Mixed
	if isHighMelanin {
		return "hazel"
	}
	return "green"
}

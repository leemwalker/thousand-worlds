package appearance

import (
	"fmt"
	"strings"

	"mud-platform-backend/internal/npc/genetics"
)

// AppearanceDescription holds the generated description components
type AppearanceDescription struct {
	FullDescription string
	Height          string
	Build           string
	Hair            string
	Eyes            string
	AgeCategory     string
}

// GenerateAppearance creates a full physical description
func GenerateAppearance(dna genetics.DNA, age, lifespan int, species string) AppearanceDescription {
	// 1. Basic Descriptors from Genetics
	heightDesc := GetHeightDescriptor(dna.Genes[genetics.GeneHeight])
	buildDesc := GetBuildDescriptor(dna.Genes[genetics.GeneBuild], dna.Genes[genetics.GeneMuscle])
	hairDesc := GetHairDescriptor(dna.Genes[genetics.GeneHair], dna.Genes[genetics.GenePigment])
	eyeDesc := GetEyeDescriptor(dna.Genes[genetics.GeneEye], dna.Genes[genetics.GeneMelanin])

	// 2. Construct Base Sentence
	// "[Height], [Build] [Species] with [Hair] hair and [Eyes] eyes"
	base := fmt.Sprintf("%s, %s %s with %s hair and %s eyes",
		strings.Title(heightDesc), buildDesc, strings.ToLower(species), hairDesc, eyeDesc)

	// 3. Apply Species Template
	base = ApplySpeciesTemplate(base, species)

	// 4. Apply Aging
	ageCat := GetAgeCategory(age, lifespan)
	fullDesc := ApplyAgeModifiers(base, ageCat)

	return AppearanceDescription{
		FullDescription: fullDesc,
		Height:          heightDesc,
		Build:           buildDesc,
		Hair:            hairDesc,
		Eyes:            eyeDesc,
		AgeCategory:     ageCat,
	}
}

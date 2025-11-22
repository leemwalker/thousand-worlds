package personality

import (
	"mud-platform-backend/internal/npc/genetics"
	"testing"
)

func TestDeriveFromGenetics(t *testing.T) {
	// Setup DNA with known genes
	dna := genetics.NewDNA()

	// Openness: Homozygous Dominant (AA) -> 70-90
	dna.Genes[GeneOpenness] = genetics.Gene{
		Allele1: "O", Allele2: "O", IsDominant1: true,
	}

	// Conscientiousness: Heterozygous (Aa) -> 50-70
	dna.Genes[GeneConscientiousness] = genetics.Gene{
		Allele1: "C", Allele2: "c", IsDominant1: true,
	}

	// Extraversion: Homozygous Recessive (aa) -> 20-50
	dna.Genes[GeneExtraversion] = genetics.Gene{
		Allele1: "e", Allele2: "e", IsDominant1: false,
	}

	p := DeriveFromGenetics(dna)

	// Verify Openness
	if p.Openness.Value < 70 || p.Openness.Value > 90 {
		t.Errorf("Expected Openness 70-90, got %f", p.Openness.Value)
	}

	// Verify Conscientiousness
	if p.Conscientiousness.Value < 50 || p.Conscientiousness.Value > 70 {
		t.Errorf("Expected Conscientiousness 50-70, got %f", p.Conscientiousness.Value)
	}

	// Verify Extraversion
	if p.Extraversion.Value < 20 || p.Extraversion.Value > 50 {
		t.Errorf("Expected Extraversion 20-50, got %f", p.Extraversion.Value)
	}

	// Verify Defaults (Agreeableness/Neuroticism) -> 40-60
	if p.Agreeableness.Value < 40 || p.Agreeableness.Value > 60 {
		t.Errorf("Expected Agreeableness 40-60 (default), got %f", p.Agreeableness.Value)
	}
}

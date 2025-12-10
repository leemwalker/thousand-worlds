package interview

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateConfiguration_Valid(t *testing.T) {
	config := &WorldConfiguration{
		ID:          uuid.New(),
		InterviewID: uuid.New(),
		CreatedBy:   uuid.New(),
		// Required fields
		WorldName:       "Test World",
		Theme:           "high fantasy",
		TechLevel:       "medieval",
		PlanetSize:      "Earth-sized",
		SentientSpecies: []string{"humans", "elves"},
		// Optional fields
		Tone:               "epic",
		MagicLevel:         "common",
		LandWaterRatio:     "70% land, 30% water",
		PoliticalStructure: "feudal kingdoms",
	}

	errors := ValidateConfiguration(config)
	if len(errors) != 0 {
		t.Errorf("Expected no validation errors for valid config, got %d errors: %v", len(errors), errors)
	}
}

func TestValidateConfiguration_MissingTheme(t *testing.T) {
	config := &WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     uuid.New(),
		CreatedBy:       uuid.New(),
		WorldName:       "Test World",
		Theme:           "", // Missing
		TechLevel:       "medieval",
		PlanetSize:      "Earth-sized",
		SentientSpecies: []string{"humans"},
	}

	errors := ValidateConfiguration(config)
	if len(errors) == 0 {
		t.Error("Expected validation error for missing theme")
	}

	found := false
	for _, err := range errors {
		if err.Field == "Theme" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for Theme field")
	}
}

func TestValidateConfiguration_MissingTechLevel(t *testing.T) {
	config := &WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     uuid.New(),
		CreatedBy:       uuid.New(),
		WorldName:       "Test World",
		Theme:           "sci-fi",
		TechLevel:       "", // Missing
		PlanetSize:      "Earth-sized",
		SentientSpecies: []string{"humans"},
	}

	errors := ValidateConfiguration(config)
	if len(errors) == 0 {
		t.Error("Expected validation error for missing tech level")
	}

	found := false
	for _, err := range errors {
		if err.Field == "TechLevel" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for TechLevel field")
	}
}

func TestValidateConfiguration_MissingPlanetSize(t *testing.T) {
	config := &WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     uuid.New(),
		CreatedBy:       uuid.New(),
		WorldName:       "Test World",
		Theme:           "fantasy",
		TechLevel:       "medieval",
		PlanetSize:      "", // Missing
		SentientSpecies: []string{"humans"},
	}

	errors := ValidateConfiguration(config)
	if len(errors) == 0 {
		t.Error("Expected validation error for missing planet size")
	}

	found := false
	for _, err := range errors {
		if err.Field == "PlanetSize" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for PlanetSize field")
	}
}

func TestValidateConfiguration_MissingSentientSpecies(t *testing.T) {
	config := &WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     uuid.New(),
		CreatedBy:       uuid.New(),
		WorldName:       "Test World",
		Theme:           "fantasy",
		TechLevel:       "medieval",
		PlanetSize:      "Earth-sized",
		SentientSpecies: nil, // Missing
	}

	errors := ValidateConfiguration(config)
	if len(errors) == 0 {
		t.Error("Expected validation error for missing sentient species")
	}

	found := false
	for _, err := range errors {
		if err.Field == "SentientSpecies" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for SentientSpecies field")
	}
}

func TestValidateConfiguration_EmptySentientSpecies(t *testing.T) {
	config := &WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     uuid.New(),
		CreatedBy:       uuid.New(),
		WorldName:       "Test World",
		Theme:           "fantasy",
		TechLevel:       "medieval",
		PlanetSize:      "Earth-sized",
		SentientSpecies: []string{}, // Empty array
	}

	errors := ValidateConfiguration(config)
	if len(errors) == 0 {
		t.Error("Expected validation error for empty sentient species array")
	}

	found := false
	for _, err := range errors {
		if err.Field == "SentientSpecies" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for SentientSpecies field")
	}
}

func TestValidateConfiguration_MultipleErrors(t *testing.T) {
	config := &WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     uuid.New(),
		CreatedBy:       uuid.New(),
		Theme:           "", // Missing
		TechLevel:       "", // Missing
		PlanetSize:      "", // Missing
		SentientSpecies: []string{},
	}

	errors := ValidateConfiguration(config)
	if len(errors) < 4 {
		t.Errorf("Expected at least 4 validation errors, got %d", len(errors))
	}
}

func TestValidateTechLevel_Valid(t *testing.T) {
	validTechLevels := []string{
		"stone_age",
		"medieval",
		"renaissance",
		"industrial",
		"modern",
		"futuristic",
		"mixed",
	}

	for _, techLevel := range validTechLevels {
		config := &WorldConfiguration{
			ID:              uuid.New(),
			InterviewID:     uuid.New(),
			CreatedBy:       uuid.New(),
			WorldName:       "Test World",
			Theme:           "fantasy",
			TechLevel:       techLevel,
			PlanetSize:      "Earth-sized",
			SentientSpecies: []string{"humans"},
		}

		errors := ValidateConfiguration(config)
		for _, err := range errors {
			if err.Field == "TechLevel" {
				t.Errorf("Tech level '%s' should be valid, got error: %s", techLevel, err.Message)
			}
		}
	}
}

func TestValidateTechLevel_Invalid(t *testing.T) {
	config := &WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     uuid.New(),
		CreatedBy:       uuid.New(),
		WorldName:       "Test World",
		Theme:           "fantasy",
		TechLevel:       "super_advanced_alien_tech", // Invalid
		PlanetSize:      "Earth-sized",
		SentientSpecies: []string{"humans"},
	}

	errors := ValidateConfiguration(config)
	found := false
	for _, err := range errors {
		if err.Field == "TechLevel" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected validation error for invalid tech level")
	}
}

func TestValidateMagicLevel_Valid(t *testing.T) {
	validMagicLevels := []string{
		"", // Optional field
		"none",
		"rare",
		"common",
		"dominant",
	}

	for _, magicLevel := range validMagicLevels {
		config := &WorldConfiguration{
			ID:              uuid.New(),
			InterviewID:     uuid.New(),
			CreatedBy:       uuid.New(),
			WorldName:       "Test World",
			Theme:           "fantasy",
			TechLevel:       "medieval",
			MagicLevel:      magicLevel,
			PlanetSize:      "Earth-sized",
			SentientSpecies: []string{"humans"},
		}

		errors := ValidateConfiguration(config)
		for _, err := range errors {
			if err.Field == "MagicLevel" {
				t.Errorf("Magic level '%s' should be valid, got error: %s", magicLevel, err.Message)
			}
		}
	}
}

func TestValidateMagicLevel_Invalid(t *testing.T) {
	config := &WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     uuid.New(),
		CreatedBy:       uuid.New(),
		WorldName:       "Test World",
		Theme:           "fantasy",
		TechLevel:       "medieval",
		MagicLevel:      "super_magic", // Invalid
		PlanetSize:      "Earth-sized",
		SentientSpecies: []string{"humans"},
	}

	errors := ValidateConfiguration(config)
	found := false
	for _, err := range errors {
		if err.Field == "MagicLevel" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected validation error for invalid magic level")
	}
}

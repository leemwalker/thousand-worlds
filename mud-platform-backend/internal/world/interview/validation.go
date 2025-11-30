package interview

import (
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

// ValidTechLevels defines accepted technology levels
var ValidTechLevels = []string{
	"stone_age",
	"medieval",
	"renaissance",
	"industrial",
	"modern",
	"futuristic",
	"mixed",
}

// ValidMagicLevels defines accepted magic levels
var ValidMagicLevels = []string{
	"none",
	"rare",
	"common",
	"dominant",
}

// ValidateConfiguration checks if configuration is valid for world generation
func ValidateConfiguration(config *WorldConfiguration) []ValidationError {
	var errors []ValidationError

	// Validate required fields
	errors = append(errors, validateRequiredFields(config)...)

	// Validate tech level if provided
	if config.TechLevel != "" {
		errors = append(errors, validateTechLevel(config)...)
	}

	// Validate magic level if provided
	if config.MagicLevel != "" {
		errors = append(errors, validateMagicLevel(config)...)
	}

	// Validate species
	errors = append(errors, validateSpecies(config)...)

	return errors
}

// validateRequiredFields checks that all required fields are present
func validateRequiredFields(config *WorldConfiguration) []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(config.Theme) == "" {
		errors = append(errors, ValidationError{
			Field:   "Theme",
			Message: "Theme is required",
		})
	}

	if strings.TrimSpace(config.TechLevel) == "" {
		errors = append(errors, ValidationError{
			Field:   "TechLevel",
			Message: "Tech level is required",
		})
	}

	if strings.TrimSpace(config.PlanetSize) == "" {
		errors = append(errors, ValidationError{
			Field:   "PlanetSize",
			Message: "Planet size is required",
		})
	}

	return errors
}

// validateTechLevel checks if tech level is valid
func validateTechLevel(config *WorldConfiguration) []ValidationError {
	var errors []ValidationError

	valid := false
	for _, validLevel := range ValidTechLevels {
		if config.TechLevel == validLevel {
			valid = true
			break
		}
	}

	if !valid {
		errors = append(errors, ValidationError{
			Field:   "TechLevel",
			Message: "Invalid tech level. Must be one of: " + strings.Join(ValidTechLevels, ", "),
		})
	}

	return errors
}

// validateMagicLevel checks if magic level is valid
func validateMagicLevel(config *WorldConfiguration) []ValidationError {
	var errors []ValidationError

	valid := false
	for _, validLevel := range ValidMagicLevels {
		if config.MagicLevel == validLevel {
			valid = true
			break
		}
	}

	if !valid {
		errors = append(errors, ValidationError{
			Field:   "MagicLevel",
			Message: "Invalid magic level. Must be one of: " + strings.Join(ValidMagicLevels, ", "),
		})
	}

	return errors
}

// validateSpecies checks that at least one sentient species is defined
func validateSpecies(config *WorldConfiguration) []ValidationError {
	var errors []ValidationError

	if config.SentientSpecies == nil || len(config.SentientSpecies) == 0 {
		errors = append(errors, ValidationError{
			Field:   "SentientSpecies",
			Message: "At least one sentient species is required",
		})
	}

	return errors
}

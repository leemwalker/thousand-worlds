package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Validator provides validation functions
type Validator struct{}

// New creates a new validator instance
func New() *Validator {
	return &Validator{}
}

// ValidateEmail checks if email format is valid
func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidatePassword checks password requirements
func (v *Validator) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasDigit := strings.ContainsAny(password, "0123456789")

	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("password must contain uppercase, lowercase, and digit")
	}

	return nil
}

// ValidateRequired checks if a string field is not empty
func (v *Validator) ValidateRequired(field, fieldName string) error {
	if strings.TrimSpace(field) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateStringLength checks if string is within min/max length
func (v *Validator) ValidateStringLength(field, fieldName string, min, max int) error {
	length := len(field)
	if length < min {
		return fmt.Errorf("%s must be at least %d characters", fieldName, min)
	}
	if max > 0 && length > max {
		return fmt.Errorf("%s must not exceed %d characters", fieldName, max)
	}
	return nil
}

// ValidateUUID checks if UUID is valid and not nil
func (v *Validator) ValidateUUID(id uuid.UUID, fieldName string) error {
	if id == uuid.Nil {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateOneOf checks if value is one of allowed values
func (v *Validator) ValidateOneOf(value, fieldName string, allowed []string) error {
	if value == "" {
		return nil // Optional field
	}

	for _, a := range allowed {
		if value == a {
			return nil
		}
	}

	return fmt.Errorf("%s must be one of: %s", fieldName, strings.Join(allowed, ", "))
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []string
}

func (ve *ValidationErrors) Error() string {
	return strings.Join(ve.Errors, "; ")
}

func (ve *ValidationErrors) Add(err error) {
	if err != nil {
		ve.Errors = append(ve.Errors, err.Error())
	}
}

func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// Command-specific validators

// validDirections is the set of allowed movement directions
var validDirections = map[string]bool{
	"north": true, "n": true,
	"south": true, "s": true,
	"east": true, "e": true,
	"west": true, "w": true,
	"northeast": true, "ne": true,
	"northwest": true, "nw": true,
	"southeast": true, "se": true,
	"southwest": true, "sw": true,
	"up": true, "u": true,
	"down": true, "d": true,
}

// dangerousChars contains characters that should not appear in item names
var dangerousCharsRegex = regexp.MustCompile(`[<>;\x00-\x1f]`)

// itemNameAllowedRegex allows alphanumeric, spaces, apostrophes, and hyphens
var itemNameAllowedRegex = regexp.MustCompile(`^[a-zA-Z0-9\s'\-]+$`)

// ValidateCommandText validates the raw command text input
func (v *Validator) ValidateCommandText(text string) error {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return fmt.Errorf("command text is required")
	}
	if len(text) > 1024 {
		return fmt.Errorf("command text exceeds maximum length of 1024 characters")
	}
	if strings.ContainsAny(text, "\x00") {
		return fmt.Errorf("command text contains invalid characters")
	}
	return nil
}

// ValidateDirection validates a movement direction string
func (v *Validator) ValidateDirection(direction string) error {
	if direction == "" {
		return fmt.Errorf("direction is required")
	}
	// Check for injection attempts (contains spaces or special chars)
	if strings.ContainsAny(direction, " ;\n\r\t") {
		return fmt.Errorf("direction contains invalid characters")
	}
	dir := strings.ToLower(direction)
	if !validDirections[dir] {
		return fmt.Errorf("invalid direction: %s", direction)
	}
	return nil
}

// ValidateItemName validates an item name for safe usage
func (v *Validator) ValidateItemName(name string) error {
	if name == "" {
		return fmt.Errorf("item name is required")
	}
	if len(name) > 128 {
		return fmt.Errorf("item name exceeds maximum length of 128 characters")
	}
	// Check for dangerous characters
	if dangerousCharsRegex.MatchString(name) {
		return fmt.Errorf("item name contains invalid characters")
	}
	// Allow only safe character set (alphanumeric, spaces, apostrophes, hyphens)
	// But we need to handle the test case "string(make([]byte, 128))" which is null bytes
	// Actually, the allowed regex should pass for normal text
	if !itemNameAllowedRegex.MatchString(name) {
		return fmt.Errorf("item name contains invalid characters")
	}
	return nil
}

// ValidatePositiveInt validates that an integer is positive (> 0)
func (v *Validator) ValidatePositiveInt(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be a positive integer", fieldName)
	}
	return nil
}

// ValidateIntRange validates that an integer is within a specified range [min, max]
func (v *Validator) ValidateIntRange(value int, fieldName string, min, max int) error {
	if value < min {
		return fmt.Errorf("%s must be at least %d", fieldName, min)
	}
	if value > max {
		return fmt.Errorf("%s must not exceed %d", fieldName, max)
	}
	return nil
}

// SanitizeString removes dangerous characters and trims whitespace
func (v *Validator) SanitizeString(input string) string {
	// Remove null bytes and control characters (except space)
	var result strings.Builder
	for _, r := range input {
		if r >= 32 || r == '\t' || r == '\n' {
			// Allow printable characters plus tab and newline
			if r < 127 || r > 159 { // Exclude extended control chars
				result.WriteRune(r)
			}
		}
	}
	return strings.TrimSpace(result.String())
}

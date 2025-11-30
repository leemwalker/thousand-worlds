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

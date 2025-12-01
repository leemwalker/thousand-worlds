package validation

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	v := New()

	tests := []struct {
		email    string
		hasError bool
	}{
		{"test@example.com", false},
		{"user.name+tag@example.co.uk", false},
		{"", true},
		{"invalid-email", true},
		{"@example.com", true},
		{"user@", true},
	}

	for _, tt := range tests {
		err := v.ValidateEmail(tt.email)
		if tt.hasError {
			assert.Error(t, err, "Expected error for email: %s", tt.email)
		} else {
			assert.NoError(t, err, "Expected no error for email: %s", tt.email)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	v := New()

	tests := []struct {
		password string
		hasError bool
	}{
		{"Password123", false},
		{"Pass1", true},       // Too short
		{"password123", true}, // No upper
		{"PASSWORD123", true}, // No lower
		{"Password", true},    // No digit
		{"", true},
	}

	for _, tt := range tests {
		err := v.ValidatePassword(tt.password)
		if tt.hasError {
			assert.Error(t, err, "Expected error for password: %s", tt.password)
		} else {
			assert.NoError(t, err, "Expected no error for password: %s", tt.password)
		}
	}
}

func TestValidateRequired(t *testing.T) {
	v := New()
	assert.NoError(t, v.ValidateRequired("value", "field"))
	assert.Error(t, v.ValidateRequired("", "field"))
	assert.Error(t, v.ValidateRequired("   ", "field"))
}

func TestValidateStringLength(t *testing.T) {
	v := New()
	assert.NoError(t, v.ValidateStringLength("abc", "field", 1, 5))
	assert.Error(t, v.ValidateStringLength("", "field", 1, 5))
	assert.Error(t, v.ValidateStringLength("abcdef", "field", 1, 5))
}

func TestValidateUUID(t *testing.T) {
	v := New()
	assert.NoError(t, v.ValidateUUID(uuid.New(), "field"))
	assert.Error(t, v.ValidateUUID(uuid.Nil, "field"))
}

func TestValidateOneOf(t *testing.T) {
	v := New()
	allowed := []string{"A", "B"}
	assert.NoError(t, v.ValidateOneOf("A", "field", allowed))
	assert.NoError(t, v.ValidateOneOf("", "field", allowed)) // Optional
	assert.Error(t, v.ValidateOneOf("C", "field", allowed))
}

func TestValidationErrors(t *testing.T) {
	ve := &ValidationErrors{}
	assert.False(t, ve.HasErrors())

	ve.Add(nil)
	assert.False(t, ve.HasErrors())

	ve.Add(assert.AnError)
	assert.True(t, ve.HasErrors())
	assert.Equal(t, assert.AnError.Error(), ve.Error())
}

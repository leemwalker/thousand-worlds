package validation

import (
	"strings"
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

// Command-specific validation tests

func TestValidateCommandText(t *testing.T) {
	v := New()

	tests := []struct {
		name     string
		text     string
		hasError bool
	}{
		{"valid short command", "look", false},
		{"valid command with args", "say hello world", false},
		{"empty command", "", true},
		{"whitespace only", "   ", true},
		{"too long command", strings.Repeat("a", 1025), true}, // Max 1024 chars
		{"contains null byte", "hello\x00world", true},
		{"valid max length", strings.Repeat("a", 1024), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateCommandText(tt.text)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDirection(t *testing.T) {
	v := New()

	tests := []struct {
		name      string
		direction string
		hasError  bool
	}{
		{"north", "north", false},
		{"south", "south", false},
		{"east", "east", false},
		{"west", "west", false},
		{"northeast", "northeast", false},
		{"northwest", "northwest", false},
		{"southeast", "southeast", false},
		{"southwest", "southwest", false},
		{"up", "up", false},
		{"down", "down", false},
		{"short n", "n", false},
		{"short s", "s", false},
		{"short e", "e", false},
		{"short w", "w", false},
		{"short ne", "ne", false},
		{"short nw", "nw", false},
		{"short se", "se", false},
		{"short sw", "sw", false},
		{"short u", "u", false},
		{"short d", "d", false},
		{"invalid direction", "sideways", true},
		{"empty", "", true},
		{"injection attempt", "north; drop table", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateDirection(tt.direction)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateItemName(t *testing.T) {
	v := New()

	tests := []struct {
		name     string
		itemName string
		hasError bool
	}{
		{"valid simple", "sword", false},
		{"valid with spaces", "iron sword", false},
		{"valid with apostrophe", "wizard's staff", false},
		{"valid with numbers", "potion42", false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 129), true}, // Max 128 chars
		{"contains special chars", "sword<script>", true},
		{"contains semicolon", "sword; drop", true},
		{"valid max length", strings.Repeat("a", 128), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateItemName(tt.itemName)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	v := New()

	tests := []struct {
		name     string
		value    int
		max      int
		hasError bool
	}{
		{"valid positive", 5, 100, false},
		{"valid at max", 100, 100, false},
		{"zero", 0, 100, true},
		{"negative", -5, 100, true},
		{"exceeds max", 101, 100, false}, // Note: max validation separate
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePositiveInt(tt.value, "test_field")
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateIntRange(t *testing.T) {
	v := New()

	tests := []struct {
		name     string
		value    int
		min      int
		max      int
		hasError bool
	}{
		{"valid in range", 50, 1, 100, false},
		{"at min", 1, 1, 100, false},
		{"at max", 100, 1, 100, false},
		{"below min", 0, 1, 100, true},
		{"above max", 101, 1, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateIntRange(tt.value, "test_field", tt.min, tt.max)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	v := New()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal text", "hello world", "hello world"},
		{"trim whitespace", "  hello  ", "hello"},
		{"remove null bytes", "hello\x00world", "helloworld"},
		{"remove control chars", "hello\x07world", "helloworld"},
		{"preserve apostrophe", "wizard's staff", "wizard's staff"},
		// XSS prevention tests
		{"strip script tags", "<script>alert('xss')</script>", "alert('xss')"},
		{"strip img onerror", `<img src=x onerror="alert('xss')">`, ""},
		{"strip nested tags", "<div><span>text</span></div>", "text"},
		{"strip self-closing", "hello<br/>world", "helloworld"},
		{"strip malformed tag", "hello<script>world</script>", "helloworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

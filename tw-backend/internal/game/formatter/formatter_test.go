package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		style    Style
		expected string
	}{
		{
			name:     "Bold text",
			input:    "Hello",
			style:    StyleBold,
			expected: `<span class="font-bold">Hello</span>`,
		},
		{
			name:     "Italic text",
			input:    "World",
			style:    StyleItalic,
			expected: `<span class="italic">World</span>`,
		},
		{
			name:     "Colored text",
			input:    "Error",
			style:    StyleRed,
			expected: `<span class="text-red-400">Error</span>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Format(tt.input, tt.style)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatItem(t *testing.T) {
	// Test item formatting with rarity
	result := Item("Excalibur", "legendary")
	assert.Contains(t, result, "Excalibur")
	assert.Contains(t, result, "text-orange-500") // Legendary color
	assert.Contains(t, result, "font-bold")
}

func TestFormatRoom(t *testing.T) {
	// Test room title formatting
	result := RoomTitle("Grand Lobby")
	assert.Contains(t, result, "Grand Lobby")
	assert.Contains(t, result, "text-blue-400")
	assert.Contains(t, result, "text-xl")
}

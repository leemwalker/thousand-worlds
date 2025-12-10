package formatter

import "fmt"

// Style represents a text style (CSS class)
type Style string

const (
	StyleBold   Style = "font-bold"
	StyleItalic Style = "italic"
	StyleRed    Style = "text-red-400"
	StyleGreen  Style = "text-green-400"
	StyleBlue   Style = "text-blue-400"
	StyleYellow Style = "text-yellow-400"
	StyleCyan   Style = "text-cyan-400"
	StylePurple Style = "text-purple-400"
	StyleOrange Style = "text-orange-500"
	StyleGray   Style = "text-gray-300"
	StyleDark   Style = "text-gray-500"
)

// Format wraps text in a span with the given style
func Format(text string, style Style) string {
	return fmt.Sprintf(`<span class="%s">%s</span>`, style, text)
}

// Item formats an item name based on rarity
func Item(name string, rarity string) string {
	var color Style
	switch rarity {
	case "uncommon":
		color = StyleGreen
	case "rare":
		color = StyleBlue
	case "very_rare":
		color = StylePurple
	case "legendary":
		color = StyleOrange
	default: // common
		color = StyleGray
	}
	return fmt.Sprintf(`<span class="%s font-bold">%s</span>`, color, name)
}

// RoomTitle formats a room title
func RoomTitle(title string) string {
	return fmt.Sprintf(`<span class="%s text-xl font-bold">%s</span>`, StyleBlue, title)
}

// Target formats a target name (e.g. for combat)
func Target(name string) string {
	return fmt.Sprintf(`<span class="%s font-bold">%s</span>`, StyleYellow, name)
}

// Damage formats a damage number
func Damage(amount int) string {
	return fmt.Sprintf(`<span class="%s font-bold">%d</span>`, StyleOrange, amount)
}

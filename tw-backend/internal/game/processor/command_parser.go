package processor

import (
	"regexp"
	"strconv"
	"strings"

	"tw-backend/cmd/game-server/websocket"
)

// CommandParser parses raw text commands into structured CommandData
type CommandParser struct {
	// Command aliases map
	aliases map[string][]string
}

// NewCommandParser creates a new command parser
func NewCommandParser() *CommandParser {
	return &CommandParser{
		aliases: map[string][]string{
			"north":     {"n"},
			"northeast": {"ne"},
			"east":      {"e"},
			"southeast": {"se"},
			"south":     {"s"},
			"southwest": {"sw"},
			"west":      {"w"},
			"northwest": {"nw"},
			"up":        {"u"},
			"down":      {"d", "dn"},
			"look":      {"l", "examine", "inspect", "view", "ex"},
			"say":       {"speak"},
			"whisper":   {"psst"},
			"tell":      {"message", "msg", "pm"},
			"who":       {"players", "online"},
			"get":       {"take", "grab", "pick", "pickup"},
			"push":      {"pull", "move"},
			"drop":      {"release", "discard", "throw"},
			"attack":    {"hit", "fight", "strike", "kill"},
			"talk":      {"chat"},
			"inventory": {"inv", "i", "items", "bag"},
			"craft":     {"make", "build", "forge"},
			"use":       {"consume", "activate", "apply"},
			"reply":     {"r"},
			"lobby":     {"exit", "leave", "hub"},
			"create":    nil,
			"weather":   {"climate", "forecast"},
			"ecosystem": {"eco"},
			"world":     nil,
			"fly":       nil,
		},
	}
}

// ParseText parses raw text input into a CommandData struct
func (p *CommandParser) ParseText(text string) *websocket.CommandData {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	// Split into words
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	// Get the command (first word)
	commandWord := strings.ToLower(words[0])
	args := words[1:]

	// Resolve command alias
	action := p.resolveAlias(commandWord)

	// Build CommandData based on action type
	cmd := &websocket.CommandData{Action: action}

	switch action {
	case "say":
		// Everything after "say" is the message
		if len(args) > 0 {
			message := strings.Join(args, " ")
			cmd.Message = &message
		}

	case "whisper", "tell":
		// Format: whisper/tell <recipient> <message>
		if len(args) >= 2 {
			recipient := args[0]
			message := strings.Join(args[1:], " ")
			cmd.Recipient = &recipient
			cmd.Message = &message
		} else if len(args) == 1 {
			// Just recipient, no message (error will be handled by processor)
			recipient := args[0]
			cmd.Recipient = &recipient
		}

	case "reply":
		// Format: reply <message>
		// Recipient will be resolved by the processor using last tell sender
		if len(args) > 0 {
			message := strings.Join(args, " ")
			cmd.Message = &message
		}

	case "create":
		// Format: create world, create <target>
		if len(args) > 0 {
			target := strings.Join(args, " ")
			cmd.Target = &target
		}

	case "enter":
		// Format: enter <world_id/portal_name/target>
		if len(args) > 0 {
			target := strings.Join(args, " ")
			cmd.Target = &target
		}

	case "look", "get", "push", "drop", "attack", "talk", "craft", "use", "open", "face":
		// Format: <action> <target>
		// Join all args as target (handles multi-word targets like "iron sword")
		if len(args) > 0 {
			target := strings.Join(args, " ")
			// Remove common articles/prepositions
			target = p.cleanTarget(target)
			cmd.Target = &target
		}

	// Direction commands don't need additional processing
	case "north", "northeast", "east", "southeast", "south", "southwest", "west", "northwest", "up", "down":
		if len(args) > 0 {
			target := strings.Join(args, " ")
			cmd.Target = &target
		}

	// Commands without arguments
	case "who", "inventory", "lobby":
		// No additional fields needed

	case "help":
		// Format: help [args]
		if len(args) > 0 {
			target := strings.Join(args, " ")
			cmd.Target = &target
		}

	case "watcher":
		// Format: watcher <world_id>
		if len(args) > 0 {
			target := strings.Join(args, " ")
			cmd.Target = &target
		}

	case "ecosystem":
		// Format: ecosystem <subcommand> <args>
		// e.g. ecosystem spawn rabbit -> Target="spawn", Message="rabbit"
		if len(args) >= 2 {
			target := args[0]
			message := strings.Join(args[1:], " ")
			cmd.Target = &target
			cmd.Message = &message
		} else if len(args) == 1 {
			target := args[0]
			cmd.Target = &target
		}

	case "world":
		// Format: world <subcommand> <args>
		// e.g. world simulate 1000000 -> Target="simulate", Message="1000000"
		if len(args) >= 2 {
			target := args[0]
			message := strings.Join(args[1:], " ")
			cmd.Target = &target
			cmd.Message = &message
		} else if len(args) == 1 {
			target := args[0]
			cmd.Target = &target
		}

	case "fly":
		// Format: fly <height>
		if len(args) >= 1 {
			target := strings.Join(args, " ")
			cmd.Target = &target
		}

	default:
		// Unknown command - keep the action as-is
		// Error will be handled by processor
	}

	return cmd
}

// resolveAlias converts an alias to its canonical action name
func (p *CommandParser) resolveAlias(word string) string {
	// Check if it's already a known action
	for action, aliases := range p.aliases {
		if word == action {
			return action
		}
		// Check aliases
		for _, alias := range aliases {
			if word == alias {
				return action
			}
		}
	}

	// Not found - return as-is (will be handled as unknown command)
	return word
}

// cleanTarget removes common articles and prepositions from targets
func (p *CommandParser) cleanTarget(target string) string {
	target = strings.ToLower(target)

	// Remove leading articles/prepositions
	re := regexp.MustCompile(`^(at|to|with|up|down|the|a|an)\s+`)
	target = re.ReplaceAllString(target, "")

	// Run again for cases like "up the sword" -> "the sword" -> "sword"
	re = regexp.MustCompile(`^(the|a|an)\s+`)
	target = re.ReplaceAllString(target, "")

	return strings.TrimSpace(target)
}

// SimulationConfig holds parsed simulation command arguments
type SimulationConfig struct {
	Years            int64
	Epoch            string
	Goal             string
	WaterLevel       string
	SimulateGeology  bool
	SimulateLife     bool
	SimulateDiseases bool
}

// ParseSimulationArgs parses simulation command arguments into a config struct.
// Input format: "<years> [--flag] [--key value]..."
// Supported flags:
//   - --only-geology: Enable only geology simulation
//   - --only-life: Enable only life simulation (includes diseases)
//   - --no-diseases: Disable disease simulation
//   - --water-level <value>: Set water level (e.g., "high", "90%")
//   - --epoch <name>: Label the epoch (e.g., "Jurassic")
//   - --goal <name>: Set simulation goal (e.g., "sapience")
func ParseSimulationArgs(argsStr string) *SimulationConfig {
	argsStr = strings.TrimSpace(argsStr)
	if argsStr == "" {
		return nil
	}

	parts := strings.Fields(argsStr)
	if len(parts) == 0 {
		return nil
	}

	// Parse years (first argument)
	yearsStr := parts[0]
	years, err := parseYears(yearsStr)
	if err != nil || years <= 0 {
		return nil
	}

	// Cap years at 10 billion
	const maxYears int64 = 10_000_000_000
	if years > maxYears {
		years = maxYears
	}

	// Initialize config with defaults
	config := &SimulationConfig{
		Years:            years,
		SimulateGeology:  true,
		SimulateLife:     true,
		SimulateDiseases: true,
	}

	// Parse flags (remaining arguments)
	for i := 1; i < len(parts); i++ {
		arg := parts[i]

		switch arg {
		case "--only-geology":
			config.SimulateGeology = true
			config.SimulateLife = false
			config.SimulateDiseases = false

		case "--only-life":
			config.SimulateGeology = false
			config.SimulateLife = true
			config.SimulateDiseases = true

		case "--no-diseases":
			config.SimulateDiseases = false

		case "--water-level":
			if i+1 < len(parts) {
				i++
				config.WaterLevel = parts[i]
			}

		case "--epoch":
			if i+1 < len(parts) {
				i++
				config.Epoch = parts[i]
			}

		case "--goal":
			if i+1 < len(parts) {
				i++
				config.Goal = parts[i]
			}
		}
	}

	return config
}

// parseYears converts a year string to int64, handling potential overflow
func parseYears(s string) (int64, error) {
	// Use ParseInt for proper int64 parsing
	years, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return years, nil
}

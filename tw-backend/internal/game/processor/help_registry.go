package processor

import (
	"fmt"
	"sort"
	"strings"
)

// CommandMetadata defines the structure for command help documentation
type CommandMetadata struct {
	Name        string
	Description string
	Usage       string
	Aliases     []string
	Flags       map[string]string          // Flag name -> description
	SubCommands map[string]CommandMetadata // For nested commands like 'world simulate'
	Category    string
}

// CommandRegistry holds all available commands
var CommandRegistry = map[string]CommandMetadata{
	// Movement
	"north": {
		Name:        "north",
		Description: "Move north.",
		Usage:       "north",
		Aliases:     []string{"n"},
		Category:    "Movement",
	},
	"northeast": {
		Name:        "northeast",
		Description: "Move northeast.",
		Usage:       "northeast",
		Aliases:     []string{"ne"},
		Category:    "Movement",
	},
	"east": {
		Name:        "east",
		Description: "Move east.",
		Usage:       "east",
		Aliases:     []string{"e"},
		Category:    "Movement",
	},
	"southeast": {
		Name:        "southeast",
		Description: "Move southeast.",
		Usage:       "southeast",
		Aliases:     []string{"se"},
		Category:    "Movement",
	},
	"south": {
		Name:        "south",
		Description: "Move south.",
		Usage:       "south",
		Aliases:     []string{"s"},
		Category:    "Movement",
	},
	"southwest": {
		Name:        "southwest",
		Description: "Move southwest.",
		Usage:       "southwest",
		Aliases:     []string{"sw"},
		Category:    "Movement",
	},
	"west": {
		Name:        "west",
		Description: "Move west.",
		Usage:       "west",
		Aliases:     []string{"w"},
		Category:    "Movement",
	},
	"northwest": {
		Name:        "northwest",
		Description: "Move northwest.",
		Usage:       "northwest",
		Aliases:     []string{"nw"},
		Category:    "Movement",
	},
	"up": {
		Name:        "up",
		Description: "Move up.",
		Usage:       "up",
		Aliases:     []string{"u"},
		Category:    "Movement",
	},
	"down": {
		Name:        "down",
		Description: "Move down.",
		Usage:       "down",
		Aliases:     []string{"d", "dn"},
		Category:    "Movement",
	},
	"fly": {
		Name:        "fly",
		Description: "Fly to a specific height.",
		Usage:       "fly <height>",
		Category:    "Movement",
	},

	// Interaction
	"look": {
		Name:        "look",
		Description: "Look around or examine a specific target.",
		Usage:       "look [target]",
		Aliases:     []string{"l", "examine", "inspect", "view", "ex"},
		Category:    "Interaction",
	},
	"get": {
		Name:        "get",
		Description: "Pick up an item.",
		Usage:       "get <item>",
		Aliases:     []string{"take", "grab", "pick", "pickup"},
		Category:    "Interaction",
	},
	"drop": {
		Name:        "drop",
		Description: "Drop an item from your inventory.",
		Usage:       "drop <item>",
		Aliases:     []string{"release", "discard", "throw"},
		Category:    "Interaction",
	},
	"open": {
		Name:        "open",
		Description: "Open a door or container.",
		Usage:       "open <target>",
		Category:    "Interaction",
	},
	"enter": {
		Name:        "enter",
		Description: "Enter a portal or doorway.",
		Usage:       "enter <target>",
		Category:    "Interaction",
	},
	"push": {
		Name:        "push",
		Description: "Push or move an object.",
		Usage:       "push <object>",
		Aliases:     []string{"pull", "move"},
		Category:    "Interaction",
	},
	"use": {
		Name:        "use",
		Description: "Use an item or object.",
		Usage:       "use <item>",
		Aliases:     []string{"consume", "activate", "apply"},
		Category:    "Interaction",
	},
	"craft": {
		Name:        "craft",
		Description: "Craft an item.",
		Usage:       "craft <item>",
		Aliases:     []string{"make", "build", "forge"},
		Category:    "Interaction",
	},
	"inventory": {
		Name:        "inventory",
		Description: "View your current inventory.",
		Usage:       "inventory",
		Aliases:     []string{"inv", "i", "items", "bag"},
		Category:    "Interaction",
	},

	// Communication
	"say": {
		Name:        "say",
		Description: "Say something to everyone in the room.",
		Usage:       "say <message>",
		Aliases:     []string{"speak"},
		Category:    "Communication",
	},
	"whisper": {
		Name:        "whisper",
		Description: "Whisper to a nearby player.",
		Usage:       "whisper <player> <message>",
		Aliases:     []string{"psst"},
		Category:    "Communication",
	},
	"tell": {
		Name:        "tell",
		Description: "Send a private message to any online player.",
		Usage:       "tell <player> <message>",
		Aliases:     []string{"message", "msg", "pm"},
		Category:    "Communication",
	},
	"reply": {
		Name:        "reply",
		Description: "Reply to the last person who messaged you.",
		Usage:       "reply <message>",
		Aliases:     []string{"r"},
		Category:    "Communication",
	},
	"talk": {
		Name:        "talk",
		Description: "Talk to an NPC.",
		Usage:       "talk <npc>",
		Aliases:     []string{"chat"},
		Category:    "Communication",
	},

	// Social / Meta
	"who": {
		Name:        "who",
		Description: "List online players.",
		Usage:       "who",
		Aliases:     []string{"players", "online"},
		Category:    "Social",
	},
	"lobby": {
		Name:        "lobby",
		Description: "Return to the lobby.",
		Usage:       "lobby",
		Aliases:     []string{"exit", "leave", "hub"},
		Category:    "Social",
	},
	"help": {
		Name:        "help",
		Description: "Show available commands or help for a specific command.",
		Usage:       "help [command]",
		Category:    "Social",
	},

	// World Management
	"world": {
		Name:        "world",
		Description: "World management commands.",
		Usage:       "world <subcommand> [args]",
		Category:    "World Management",
		SubCommands: map[string]CommandMetadata{
			"simulate": {
				Name:        "simulate",
				Description: "Run a fast-forward simulation of the world.",
				Usage:       "world simulate <years> [flags]",
				Flags: map[string]string{
					"--epoch <name>":        "Start in specific epoch",
					"--goal <name>":         "Set evolution goal",
					"--only-geology":        "Simulate only geology",
					"--only-life":           "Simulate only life",
					"--no-diseases":         "Disable disease simulation",
					"--water-level <level>": "Set water level (high, low, medium, %, or meters)",
				},
			},
			"info": {
				Name:        "info",
				Description: "Show world information.",
				Usage:       "world info",
			},
			"reset": {
				Name:        "reset",
				Description: "Reset the world state.",
				Usage:       "world reset",
			},
			"run": {
				Name:        "run",
				Description: "Start the continuous simulation.",
				Usage:       "world run",
			},
			"pause": {
				Name:        "pause",
				Description: "Pause the continuous simulation.",
				Usage:       "world pause",
			},
			"speed": {
				Name:        "speed",
				Description: "Set the simulation speed.",
				Usage:       "world speed <speed>",
			},
		},
	},
	"ecosystem": {
		Name:        "ecosystem",
		Description: "Ecosystem management commands.",
		Usage:       "ecosystem <subcommand> [args]",
		Aliases:     []string{"eco"},
		Category:    "World Management",
		SubCommands: map[string]CommandMetadata{
			"spawn": {
				Name:        "spawn",
				Description: "Spawn an entity.",
				Usage:       "ecosystem spawn <entity>",
			},
			"status": {
				Name:        "status",
				Description: "Show ecosystem status.",
				Usage:       "ecosystem status",
			},
		},
	},
	"weather": {
		Name:        "weather",
		Description: "Check the weather.",
		Usage:       "weather",
		Aliases:     []string{"climate", "forecast"},
		Category:    "World Management",
	},
	"create": {
		Name:        "create",
		Description: "Create a new world or character.",
		Usage:       "create <world|character>",
		Category:    "World Management",
	},
}

// GetHelpText generates help text based on arguments
func GetHelpText(args []string) string {
	// Case 1: No arguments - show all commands grouped by category
	if len(args) == 0 {
		return formatGeneralHelp()
	}

	search := strings.ToLower(args[0])

	// Case 2: Specific command help
	if cmd, exists := CommandRegistry[search]; exists {
		// If there are more args, check for subcommands
		if len(args) > 1 {
			subSearch := strings.ToLower(args[1])
			if subCmd, subExists := cmd.SubCommands[subSearch]; subExists {
				return formatCommandHelp(subCmd, cmd.Name)
			}
			return fmt.Sprintf("Unknown subcommand '%s' for '%s'.\n%s", subSearch, cmd.Name, formatCommandHelp(cmd, ""))
		}
		return formatCommandHelp(cmd, "")
	}

	// Case 3: Search by prefix
	matches := make([]string, 0)
	for name := range CommandRegistry {
		if strings.HasPrefix(name, search) {
			matches = append(matches, name)
		}
	}

	if len(matches) > 0 {
		sort.Strings(matches)
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Commands starting with '%s':\n", search))
		for _, name := range matches {
			cmd := CommandRegistry[name]
			sb.WriteString(fmt.Sprintf("  %-15s - %s\n", cmd.Name, cmd.Description))
		}
		return sb.String()
	}

	return fmt.Sprintf("Unknown command '%s'. Type 'help' to see all commands.", search)
}

func formatGeneralHelp() string {
	var sb strings.Builder
	sb.WriteString("Available Commands:\n")

	// Group by category
	commandsByCategory := make(map[string][]CommandMetadata)
	for _, cmd := range CommandRegistry {
		commandsByCategory[cmd.Category] = append(commandsByCategory[cmd.Category], cmd)
	}

	// Sort categories
	categories := make([]string, 0, len(commandsByCategory))
	for cat := range commandsByCategory {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("\n  %s:\n", cat))
		cmds := commandsByCategory[cat]
		// Sort commands in category
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name < cmds[j].Name
		})

		for _, cmd := range cmds {
			sb.WriteString(fmt.Sprintf("    %-15s - %s\n", cmd.Name, cmd.Description))
		}
	}

	sb.WriteString("\nType 'help <command>' for more details (e.g., 'help world').")
	return sb.String()
}

func formatCommandHelp(cmd CommandMetadata, parentName string) string {
	var sb strings.Builder
	fullName := cmd.Name
	if parentName != "" {
		fullName = fmt.Sprintf("%s %s", parentName, cmd.Name)
	}

	sb.WriteString(fmt.Sprintf("Command: %s\n", fullName))
	sb.WriteString(fmt.Sprintf("Description: %s\n", cmd.Description))
	sb.WriteString(fmt.Sprintf("Usage: %s\n", cmd.Usage))

	if len(cmd.Aliases) > 0 {
		sb.WriteString(fmt.Sprintf("Aliases: %s\n", strings.Join(cmd.Aliases, ", ")))
	}

	if len(cmd.Flags) > 0 {
		sb.WriteString("\nFlags:\n")
		// Sort flags
		flags := make([]string, 0, len(cmd.Flags))
		for f := range cmd.Flags {
			flags = append(flags, f)
		}
		sort.Strings(flags)
		for _, f := range flags {
			sb.WriteString(fmt.Sprintf("  %-25s %s\n", f, cmd.Flags[f]))
		}
	}

	if len(cmd.SubCommands) > 0 {
		sb.WriteString("\nSubcommands:\n")
		// Sort subcommands
		subs := make([]string, 0, len(cmd.SubCommands))
		for s := range cmd.SubCommands {
			subs = append(subs, s)
		}
		sort.Strings(subs)
		for _, s := range subs {
			sub := cmd.SubCommands[s]
			sb.WriteString(fmt.Sprintf("  %-15s - %s\n", sub.Name, sub.Description))
		}
	}

	return sb.String()
}

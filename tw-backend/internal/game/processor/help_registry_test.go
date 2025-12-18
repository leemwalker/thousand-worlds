package processor

import (
	"strings"
	"testing"
)

func TestGetHelpText(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name:     "General Help",
			args:     []string{},
			contains: []string{"Available Commands:", "Movement:", "Interaction:", "Show available commands"},
		},
		{
			name:     "Command Help - Simple",
			args:     []string{"look"},
			contains: []string{"Command: look", "Usage: look [target]", "Aliases: l, examine"},
		},
		{
			name:     "Command Help - With Subcommand",
			args:     []string{"world", "simulate"},
			contains: []string{"Command: world simulate", "Usage: world simulate <years>", "Flags:", "--epoch"},
		},
		{
			name:     "Prefix Search - Matches found",
			args:     []string{"wo"},
			contains: []string{"Commands starting with 'wo':", "world", "World management commands"},
		},
		{
			name:     "Prefix Search - No matches",
			args:     []string{"xyz"},
			contains: []string{"Unknown command 'xyz'"},
		},
		{
			name:     "Subcommand not found",
			args:     []string{"world", "xyz"},
			contains: []string{"Unknown subcommand 'xyz' for 'world'", "Usage: world <subcommand>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetHelpText(tt.args)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("GetHelpText() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}

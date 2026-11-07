package cli

import (
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/theme"
)

var consoleCommands = []string{"about", "exit", "help", "man", "quit", "theme", "version"}

type completionState struct {
	completed string
	choices   []string
}

func (s *completionState) tab(input string) (string, []string, bool) {
	if input == s.completed && len(s.choices) > 0 {
		_, choices := completeLine(input)
		if len(choices) > 1 {
			s.completed, s.choices = input, choices
			return input, choices, true
		}
	}
	completed, choices := completeLine(input)
	s.completed, s.choices = completed, choices
	return completed, choices, false
}

func completeLine(input string) (string, []string) {
	separator := strings.LastIndex(input, " ")
	var prefix, partial string
	var available []string
	if separator < 0 {
		partial = input
		available = consoleCommands
	} else {
		prefix = input[:separator+1]
		partial = input[separator+1:]
		if strings.TrimSpace(input[:separator]) != "theme" {
			return input, nil
		}
		available = cliThemeAliases(theme.CatalogFor(theme.ModeCLI))
	}

	matches := make([]string, 0, len(available))
	for _, candidate := range available {
		if strings.HasPrefix(strings.ToLower(candidate), strings.ToLower(partial)) {
			matches = append(matches, candidate)
		}
	}
	if len(matches) == 0 {
		return input, nil
	}
	if len(matches) == 1 {
		return prefix + matches[0] + " ", matches
	}
	common := commonPrefix(matches)
	if len(common) > len(partial) {
		return prefix + common, matches
	}
	return input, matches
}

func commonPrefix(values []string) string {
	if len(values) == 0 {
		return ""
	}
	prefix := values[0]
	for _, value := range values[1:] {
		for !strings.HasPrefix(strings.ToLower(value), strings.ToLower(prefix)) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}
	return prefix
}

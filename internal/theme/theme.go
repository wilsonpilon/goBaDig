package theme

import (
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	ModeGUI = "gui"
	ModeTUI = "tui"
	ModeCLI = "cli"
)

type Theme struct {
	ID          string  `json:"id"`
	Mode        string  `json:"mode"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Palette     Palette `json:"palette"`
}

type Palette struct {
	Background string `json:"background"`
	Foreground string `json:"foreground"`
	Surface    string `json:"surface"`
	Primary    string `json:"primary"`
	Hover      string `json:"hover"`
	Selection  string `json:"selection"`
	Muted      string `json:"muted"`
	Border     string `json:"border"`
	Error      string `json:"error"`
}

func Catalog() []Theme {
	return CatalogFor(ModeGUI)
}

func CatalogAll() []Theme {
	var catalog []Theme
	for _, mode := range []string{ModeGUI, ModeTUI, ModeCLI} {
		catalog = append(catalog, CatalogFor(mode)...)
	}
	return catalog
}

func CatalogFor(mode string) []Theme {
	guiThemes := []Theme{
		{
			ID: "vscode-dark", Mode: ModeGUI, Name: "VS Code Dark",
			Description: "Escuro moderno com azul de destaque.",
			Palette: Palette{
				Background: "#1E1E1E", Foreground: "#D4D4D4", Surface: "#252526",
				Primary: "#007ACC", Hover: "#2A2D2E", Selection: "#264F78",
				Muted: "#858585", Border: "#3C3C3C", Error: "#F14C4C",
			},
		},
		{
			ID: "dracula", Mode: ModeGUI, Name: "Dracula",
			Description: "Escuro vibrante em roxo, rosa e ciano.",
			Palette: Palette{
				Background: "#282A36", Foreground: "#F8F8F2", Surface: "#343746",
				Primary: "#BD93F9", Hover: "#44475A", Selection: "#44475A",
				Muted: "#6272A4", Border: "#44475A", Error: "#FF5555",
			},
		},
		{
			ID: "monokai", Mode: ModeGUI, Name: "Monokai",
			Description: "Escuro contrastante com verde e laranja.",
			Palette: Palette{
				Background: "#272822", Foreground: "#F8F8F2", Surface: "#3E3D32",
				Primary: "#A6E22E", Hover: "#49483E", Selection: "#49483E",
				Muted: "#75715E", Border: "#49483E", Error: "#F92672",
			},
		},
		{
			ID: "nord", Mode: ModeGUI, Name: "Nord",
			Description: "Tons frios de azul e gelo.",
			Palette: Palette{
				Background: "#2E3440", Foreground: "#ECEFF4", Surface: "#3B4252",
				Primary: "#88C0D0", Hover: "#434C5E", Selection: "#4C566A",
				Muted: "#D8DEE9", Border: "#4C566A", Error: "#BF616A",
			},
		},
		{
			ID: "solarized-light", Mode: ModeGUI, Name: "Solarized Light",
			Description: "Claro e suave com acentos teal e dourado.",
			Palette: Palette{
				Background: "#FDF6E3", Foreground: "#586E75", Surface: "#EEE8D5",
				Primary: "#268BD2", Hover: "#E6DFC9", Selection: "#D6E6EB",
				Muted: "#93A1A1", Border: "#D3CBB7", Error: "#DC322F",
			},
		},
		{
			ID: "gruvbox-dark", Mode: ModeGUI, Name: "Gruvbox Dark",
			Description: "Escuro quente com tons terrosos e amarelo.",
			Palette: Palette{
				Background: "#282828", Foreground: "#EBDBB2", Surface: "#3C3836",
				Primary: "#FABD2F", Hover: "#504945", Selection: "#665C54",
				Muted: "#A89984", Border: "#504945", Error: "#FB4934",
			},
		},
	}

	catalog := make([]Theme, len(guiThemes))
	for index, item := range guiThemes {
		if mode != ModeGUI {
			item.ID = mode + "-" + item.ID
		}
		item.Mode = mode
		catalog[index] = item
	}
	return catalog
}

func (t Theme) Validate() error {
	if strings.TrimSpace(t.ID) == "" || strings.TrimSpace(t.Mode) == "" || strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("theme id, mode, and name are required")
	}
	fields := map[string]string{
		"background": t.Palette.Background,
		"foreground": t.Palette.Foreground,
		"surface":    t.Palette.Surface,
		"primary":    t.Palette.Primary,
		"hover":      t.Palette.Hover,
		"selection":  t.Palette.Selection,
		"muted":      t.Palette.Muted,
		"border":     t.Palette.Border,
		"error":      t.Palette.Error,
	}
	for name, value := range fields {
		if len(value) != 7 || value[0] != '#' {
			return fmt.Errorf("theme %q has an invalid %s color %q", t.ID, name, value)
		}
		if _, err := hex.DecodeString(value[1:]); err != nil {
			return fmt.Errorf("theme %q has an invalid %s color %q: %w", t.ID, name, value, err)
		}
	}
	return nil
}

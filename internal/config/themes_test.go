package config

import (
	"path/filepath"
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/theme"
)

func TestThemeCatalogAndSelectionPersist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.InstallThemes(theme.Catalog()); err != nil {
		t.Fatal(err)
	}
	if err := store.InstallThemes(theme.Catalog()); err != nil {
		t.Fatal(err)
	}

	themes, err := store.ListThemes(theme.ModeGUI)
	if err != nil {
		t.Fatal(err)
	}
	if len(themes) < 5 || len(themes) > 7 {
		t.Fatalf("catalog has %d themes; want 5 to 7", len(themes))
	}
	active, err := store.EnsureActiveTheme(theme.ModeGUI, "vscode-dark")
	if err != nil {
		t.Fatal(err)
	}
	if active.ID != "vscode-dark" {
		t.Fatalf("initial active theme = %q; want vscode-dark", active.ID)
	}
	if err := store.SetActiveTheme(theme.ModeGUI, "dracula"); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	active, err = store.EnsureActiveTheme(theme.ModeGUI, "vscode-dark")
	if err != nil {
		t.Fatal(err)
	}
	if active.ID != "dracula" {
		t.Fatalf("active theme after reopening = %q; want dracula", active.ID)
	}
}

func TestSetActiveThemeRejectsWrongModeOrUnknownTheme(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.InstallThemes(theme.Catalog()); err != nil {
		t.Fatal(err)
	}
	if err := store.SetActiveTheme(theme.ModeGUI, "missing"); err == nil {
		t.Fatal("expected an error for an uninstalled theme")
	}
	if err := store.SetActiveTheme("cli", "dracula"); err == nil {
		t.Fatal("expected an error when selecting a theme from another mode")
	}
}

func TestThemeCatalogCoversGUIAndTerminalModes(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.InstallThemes(theme.CatalogAll()); err != nil {
		t.Fatal(err)
	}

	for _, mode := range []string{theme.ModeGUI, theme.ModeTUI, theme.ModeCLI} {
		items, err := store.ListThemes(mode)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != 6 {
			t.Fatalf("%s theme count = %d, want 6", mode, len(items))
		}
	}
	for mode, id := range map[string]string{
		theme.ModeGUI: "vscode-dark",
		theme.ModeTUI: "tui-vscode-dark",
		theme.ModeCLI: "cli-vscode-dark",
	} {
		active, err := store.EnsureActiveTheme(mode, id)
		if err != nil {
			t.Fatal(err)
		}
		if active.ID != id {
			t.Errorf("default %s theme = %q, want %q", mode, active.ID, id)
		}
	}
}

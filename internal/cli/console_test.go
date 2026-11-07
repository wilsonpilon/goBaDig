package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/theme"
)

func TestRunConsole(t *testing.T) {
	in := strings.NewReader("about\nversion\nnope\nexit\n")
	var out bytes.Buffer

	RunConsole(in, &out)

	got := out.String()
	for _, want := range []string{
		"Modo console",
		"goBaDig> ",
		"Basic Dignified Suite, em Go.", // from "about"
		"bloco NEW, feature themes",     // from "version"
		`comando desconhecido: "nope"`,  // unknown command
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("console output missing %q; got:\n%s", want, got)
		}
	}
}

func TestRunConsoleStopsOnEOF(t *testing.T) {
	// Empty input hits EOF on the first read; RunConsole must return
	// instead of blocking forever waiting for an "exit"/"quit" line.
	in := strings.NewReader("")
	var out bytes.Buffer

	RunConsole(in, &out)

	if !strings.Contains(out.String(), "goBaDig> ") {
		t.Fatalf("expected the prompt to be printed at least once; got:\n%s", out.String())
	}
}

func TestCompleteLine(t *testing.T) {
	cases := []struct {
		input string
		want  string
		count int
	}{
		{input: "ver", want: "version ", count: 1},
		{input: "theme dr", want: "theme dracula ", count: 1},
		{input: "theme ", want: "theme ", count: 6},
		{input: "the", want: "theme ", count: 1},
		{input: "about x", want: "about x", count: 0},
	}
	for _, test := range cases {
		got, candidates := completeLine(test.input)
		if got != test.want {
			t.Errorf("completeLine(%q) = %q; want %q", test.input, got, test.want)
		}
		if len(candidates) != test.count {
			t.Errorf("completeLine(%q) returned %d candidates; want %d", test.input, len(candidates), test.count)
		}
	}
}

func TestThemeCompletionListsThemesOnSecondTab(t *testing.T) {
	var completion completionState
	line, _, show := completion.tab("theme")
	if line != "theme " || show {
		t.Fatalf("first Tab = (%q, show=%t), want completed command without list", line, show)
	}
	line, choices, show := completion.tab(line)
	if !show || line != "theme " || len(choices) != 6 {
		t.Fatalf("second Tab = (%q, %d choices, show=%t), want theme list", line, len(choices), show)
	}
}

func TestThemeCommandAcceptsDisplayNameAndPersistsSelection(t *testing.T) {
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.InstallThemes(theme.CatalogFor(theme.ModeCLI)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.EnsureActiveTheme(theme.ModeCLI, defaultCLIThemeID); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	state := &consoleState{out: &out, store: store}
	state.executeTheme("Solarized Light")
	active, err := store.ActiveTheme(theme.ModeCLI)
	if err != nil {
		t.Fatal(err)
	}
	if active.ID != "cli-solarized-light" {
		t.Fatalf("active theme = %q; want cli-solarized-light", active.ID)
	}
	if !strings.Contains(out.String(), "Solarized Light") {
		t.Fatalf("theme confirmation missing from output: %s", out.String())
	}
}

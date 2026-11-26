package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/options"
	"github.com/wilsonpilon/goBaDig/internal/theme"
	"github.com/wilsonpilon/goBaDig/internal/version"
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
		fmt.Sprintf("bloco %s, feature %s", version.Block, version.Feature), // from "version"
		`comando desconhecido: "nope"`,                                      // unknown command
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
		{input: "dia", want: "dialect ", count: 1},
		{input: "dialect ms", want: "dialect msx ", count: 1},
		{input: "com", want: "compile ", count: 1},
		{input: "exe", want: "execute ", count: 1},
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

func TestSetCommandShowsAndUpdatesACoreOption(t *testing.T) {
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var out bytes.Buffer
	state := &consoleState{out: &out, store: store}

	state.executeSet("INITIAL_LINE")
	if !strings.Contains(out.String(), "INITIAL_LINE = 10") {
		t.Fatalf("executeSet(INITIAL_LINE) without a value should show the catalog default; got: %s", out.String())
	}

	out.Reset()
	state.executeSet("INITIAL_LINE 1")
	if !strings.Contains(out.String(), "INITIAL_LINE = 1") {
		t.Fatalf("executeSet(INITIAL_LINE 1) confirmation missing; got: %s", out.String())
	}

	out.Reset()
	state.executeSet("initial_line")
	if !strings.Contains(out.String(), "INITIAL_LINE = 1") {
		t.Fatalf("executeSet(initial_line) (lowercase key) should still read the value just set; got: %s", out.String())
	}
}

func TestSetCommandIsCaseInsensitiveOnTheCommandItself(t *testing.T) {
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var out bytes.Buffer
	state := &consoleState{out: &out, store: store}
	state.execute("SET INITIAL_LINE 7")

	value, err := store.GetOption(options.CategoryCore, "INITIAL_LINE")
	if err != nil {
		t.Fatal(err)
	}
	if value != "7" {
		t.Fatalf(`execute("SET INITIAL_LINE 7"): INITIAL_LINE = %q, want "7" (the user's own worked example, in caps)`, value)
	}
}

func TestSetCommandResolvesABareKeyAgainstTheCurrentDialect(t *testing.T) {
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var out bytes.Buffer
	state := &consoleState{out: &out, store: store} // dialect unset -> defaults to MSX
	state.executeSet("TOKENIZE true")

	if _, err := store.GetOption(options.CategoryZX81, "TOKENIZE"); err == nil {
		t.Fatalf("SET TOKENIZE with the default (MSX) dialect must not touch ZX81's own TOKENIZE")
	}
	value, err := store.GetOption(options.CategoryMSX, "TOKENIZE")
	if err != nil {
		t.Fatal(err)
	}
	if value != "true" {
		t.Fatalf("msx.TOKENIZE = %q, want true", value)
	}
}

func TestDialectCommandSwitchesWhichCategorySetResolvesAgainst(t *testing.T) {
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var out bytes.Buffer
	state := &consoleState{out: &out, store: store}
	state.executeDialect("zx81")
	if state.dialect != options.CategoryZX81 {
		t.Fatalf("after executeDialect(zx81), state.dialect = %q, want %q", state.dialect, options.CategoryZX81)
	}

	out.Reset()
	// ZX81's own TOKENIZE defaults to true (unlike MSX's false) -
	// confirms "set" is reading from zx81, not msx, after the switch.
	state.executeSet("TOKENIZE")
	if !strings.Contains(out.String(), "TOKENIZE = true") {
		t.Fatalf("after switching to zx81, executeSet(TOKENIZE) = %q, want it to report zx81's own default (true)", out.String())
	}
}

func TestSetCommandRejectsUnknownOption(t *testing.T) {
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var out bytes.Buffer
	state := &consoleState{out: &out, store: store}
	state.executeSet("NOT_A_REAL_OPTION 1")
	if !strings.Contains(out.String(), "opção desconhecida") {
		t.Fatalf("executeSet(NOT_A_REAL_OPTION 1): want an unknown-option error, got: %s", out.String())
	}
}

func TestSetCommandRejectsInvalidValueWithoutChangingTheStoredOne(t *testing.T) {
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var out bytes.Buffer
	state := &consoleState{out: &out, store: store}
	state.executeSet("REM_HEADER maybe")
	if !strings.Contains(out.String(), "erro") {
		t.Fatalf("executeSet(REM_HEADER maybe): want an error, got: %s", out.String())
	}
	value, err := store.GetOption(options.CategoryCore, "REM_HEADER")
	if err != nil {
		t.Fatal(err)
	}
	if value != "true" {
		t.Fatalf("REM_HEADER after a rejected SET = %q, want the untouched default %q", value, "true")
	}
}

func TestSettingsCommandListsCoreAndCurrentDialectOptions(t *testing.T) {
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var out bytes.Buffer
	state := &consoleState{out: &out, store: store}
	state.executeSettings()

	got := out.String()
	for _, want := range []string{"Basic Dignified:", "INITIAL_LINE = 10", "msx:", "CONVERT_PRINT = NONE"} {
		if !strings.Contains(got, want) {
			t.Fatalf("executeSettings() output missing %q; got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "SPECTRUM_MODEL") {
		t.Fatalf("executeSettings() with the default (MSX) dialect must not list ZX Spectrum's own options; got:\n%s", got)
	}
}

func TestCompileCommandWritesAmxAndBmxNextToTheSourceFile(t *testing.T) {
	dir := t.TempDir()
	dmxPath := filepath.Join(dir, "programa.dmx")
	if err := os.WriteFile(dmxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	state := &consoleState{out: &out}
	state.executeCompile(dmxPath)

	if !strings.Contains(out.String(), "Convertido") {
		t.Fatalf("executeCompile(%s) output missing success confirmation; got: %s", dmxPath, out.String())
	}
	for _, ext := range []string{".amx", ".bmx"} {
		path := strings.TrimSuffix(dmxPath, ".dmx") + ext
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}

// TestCompileCommandWritesAzxAndPNextToTheSourceFile mirrors
// TestCompileCommandWritesAmxAndBmxNextToTheSourceFile for the ZX81
// side of compileByExtension's dispatch.
func TestCompileCommandWritesAzxAndPNextToTheSourceFile(t *testing.T) {
	dir := t.TempDir()
	dzxPath := filepath.Join(dir, "programa.dzx")
	if err := os.WriteFile(dzxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	state := &consoleState{out: &out}
	state.executeCompile(dzxPath)

	if !strings.Contains(out.String(), "Convertido") {
		t.Fatalf("executeCompile(%s) output missing success confirmation; got: %s", dzxPath, out.String())
	}
	for _, ext := range []string{".azx", ".p"} {
		path := strings.TrimSuffix(dzxPath, ".dzx") + ext
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}

// TestCompileCommandWritesAzsAndTapNextToTheSourceFile mirrors
// TestCompileCommandWritesAmxAndBmxNextToTheSourceFile for the ZX
// Spectrum side of compileByExtension's dispatch.
func TestCompileCommandWritesAzsAndTapNextToTheSourceFile(t *testing.T) {
	dir := t.TempDir()
	dzsPath := filepath.Join(dir, "programa.dzs")
	if err := os.WriteFile(dzsPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	state := &consoleState{out: &out}
	state.executeCompile(dzsPath)

	if !strings.Contains(out.String(), "Convertido") {
		t.Fatalf("executeCompile(%s) output missing success confirmation; got: %s", dzsPath, out.String())
	}
	for _, ext := range []string{".azs", ".tap"} {
		path := strings.TrimSuffix(dzsPath, ".dzs") + ext
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}

func TestCompileCommandWithoutAnArgumentShowsUsage(t *testing.T) {
	var out bytes.Buffer
	state := &consoleState{out: &out}
	state.executeCompile("")
	if !strings.Contains(out.String(), "uso:") {
		t.Fatalf("executeCompile(\"\") output missing usage hint; got: %s", out.String())
	}
}

func TestCompileCommandReportsErrorForMissingFile(t *testing.T) {
	var out bytes.Buffer
	state := &consoleState{out: &out}
	state.executeCompile(filepath.Join(t.TempDir(), "nope.dmx"))
	if !strings.Contains(out.String(), "erro") {
		t.Fatalf("executeCompile(missing file) output missing error; got: %s", out.String())
	}
}

// TestExecuteCommandBuildsTheDiskEvenWhenTheEmulatorIsMissing points
// PATH at an empty directory so openmsx can never be found - this
// still exercises the real msx.BuildTestDisk step (which doesn't
// touch PATH at all) without this test launching an actual emulator
// window every time the suite runs. Confirms the disk gets built
// (and stays on disk) even though the emulator-launch step fails,
// and that the failure is reported through the console the same way
// every other error here is.
func TestExecuteCommandBuildsTheDiskEvenWhenTheEmulatorIsMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	dmxPath := filepath.Join(dir, "programa.dmx")
	if err := os.WriteFile(dmxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	state := &consoleState{out: &out}
	state.executeExecute(dmxPath)

	if !strings.Contains(out.String(), "erro") {
		t.Fatalf("executeExecute with no emulator on PATH: want an error reported, got: %s", out.String())
	}
	diskPath := strings.TrimSuffix(dmxPath, ".dmx") + ".dsk"
	if _, err := os.Stat(diskPath); err != nil {
		t.Fatalf("expected the disk to still have been built at %s: %v", diskPath, err)
	}
}

// TestExecuteCommandBuildsTheZX81FileEvenWhenTheEmulatorIsMissing
// mirrors TestExecuteCommandBuildsTheDiskEvenWhenTheEmulatorIsMissing
// for the ZX81 side - there is no disk to build, just the ".p" file
// itself, from the real zx81.CompileFile step.
func TestExecuteCommandBuildsTheZX81FileEvenWhenTheEmulatorIsMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	dzxPath := filepath.Join(dir, "programa.dzx")
	if err := os.WriteFile(dzxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	state := &consoleState{out: &out}
	state.executeExecute(dzxPath)

	if !strings.Contains(out.String(), "erro") {
		t.Fatalf("executeExecute with no emulator on PATH: want an error reported, got: %s", out.String())
	}
	pPath := strings.TrimSuffix(dzxPath, ".dzx") + ".p"
	if _, err := os.Stat(pPath); err != nil {
		t.Fatalf("expected the .p file to still have been built at %s: %v", pPath, err)
	}
}

// TestExecuteCommandBuildsTheZXSpectrumFileEvenWhenTheEmulatorIsMissing
// mirrors TestExecuteCommandBuildsTheDiskEvenWhenTheEmulatorIsMissing
// for the ZX Spectrum side - there is no disk to build, just the
// ".tap" file itself, from the real zxspectrum.CompileFile step.
func TestExecuteCommandBuildsTheZXSpectrumFileEvenWhenTheEmulatorIsMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	dzsPath := filepath.Join(dir, "programa.dzs")
	if err := os.WriteFile(dzsPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	state := &consoleState{out: &out}
	state.executeExecute(dzsPath)

	if !strings.Contains(out.String(), "erro") {
		t.Fatalf("executeExecute with no emulator on PATH: want an error reported, got: %s", out.String())
	}
	tapPath := strings.TrimSuffix(dzsPath, ".dzs") + ".tap"
	if _, err := os.Stat(tapPath); err != nil {
		t.Fatalf("expected the .tap file to still have been built at %s: %v", tapPath, err)
	}
}

func TestExecuteCommandWithoutAnArgumentShowsUsage(t *testing.T) {
	var out bytes.Buffer
	state := &consoleState{out: &out}
	state.executeExecute("")
	if !strings.Contains(out.String(), "uso:") {
		t.Fatalf("executeExecute(\"\") output missing usage hint; got: %s", out.String())
	}
}

func TestExecuteCommandReportsErrorForMissingFile(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	var out bytes.Buffer
	state := &consoleState{out: &out}
	state.executeExecute(filepath.Join(t.TempDir(), "nope.dmx"))
	if !strings.Contains(out.String(), "erro") {
		t.Fatalf("executeExecute(missing file) output missing error; got: %s", out.String())
	}
}

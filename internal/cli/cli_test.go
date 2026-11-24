package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want Mode
	}{
		{"no args", []string{}, ModeBanner},
		{"help long", []string{"--help"}, ModeHelp},
		{"help short", []string{"-h"}, ModeHelp},
		{"man", []string{"--man"}, ModeMan},
		{"tui", []string{"--tui"}, ModeTUI},
		{"gui", []string{"--gui"}, ModeGUI},
		{"cli", []string{"--cli"}, ModeConsole},
		{"unknown flag falls back to help", []string{"--nope"}, ModeHelp},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Parse(c.args); got != c.want {
				t.Fatalf("Parse(%v) = %v, want %v", c.args, got, c.want)
			}
		})
	}
}

func TestRunHandlesNonInteractiveModes(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"--man"}, {}} {
		handled, code := Run(args)
		if !handled {
			t.Fatalf("Run(%v): expected handled=true", args)
		}
		if code != 0 {
			t.Fatalf("Run(%v): expected exit code 0, got %d", args, code)
		}
	}
}

func TestRunLeavesInteractiveModesToCaller(t *testing.T) {
	for _, args := range [][]string{{"--tui"}, {"--gui"}, {"--cli"}} {
		handled, _ := Run(args)
		if handled {
			t.Fatalf("Run(%v): expected handled=false, caller must launch the mode", args)
		}
	}
}

func TestCompileFileArg(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
		ok   bool
	}{
		{"single bare file", []string{"programa.dmx"}, "programa.dmx", true},
		{"no args", []string{}, "", false},
		{"only a flag", []string{"--tui"}, "", false},
		{"two bare files - ambiguous", []string{"a.dmx", "b.dmx"}, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := compileFileArg(c.args)
			if ok != c.ok || got != c.want {
				t.Fatalf("compileFileArg(%v) = (%q, %v), want (%q, %v)", c.args, got, ok, c.want, c.ok)
			}
		})
	}
}

// TestRunCompilesABareFileArgument confirms "goBaDig programa.dmx"
// (no flag at all) is the first real "do something to a user's own
// file" command in the project - Run must recognize a single
// positional file argument, call through to msx.CompileFile, and
// report success once the real .amx/.bmx files exist on disk.
func TestRunCompilesABareFileArgument(t *testing.T) {
	dir := t.TempDir()
	dmxPath := filepath.Join(dir, "programa.dmx")
	if err := os.WriteFile(dmxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	handled, code := Run([]string{dmxPath})
	if !handled {
		t.Fatalf("Run(%v): expected handled=true", dmxPath)
	}
	if code != 0 {
		t.Fatalf("Run(%v): expected exit code 0, got %d", dmxPath, code)
	}

	amxPath := filepath.Join(dir, "programa.amx")
	bmxPath := filepath.Join(dir, "programa.bmx")
	if _, err := os.Stat(amxPath); err != nil {
		t.Fatalf("expected %s to exist: %v", amxPath, err)
	}
	if _, err := os.Stat(bmxPath); err != nil {
		t.Fatalf("expected %s to exist: %v", bmxPath, err)
	}
}

func TestRunReportsAnErrorForAMissingFileArgument(t *testing.T) {
	handled, code := Run([]string{filepath.Join(t.TempDir(), "nope.dmx")})
	if !handled {
		t.Fatalf("Run(missing file): expected handled=true")
	}
	if code == 0 {
		t.Fatalf("Run(missing file): expected a non-zero exit code")
	}
}

// TestRunPrefersAnExplicitModeFlagOverABareFileArgument confirms
// Parse's own flag matching still wins when both a recognized flag
// and a bare file argument are present - compileFileArg is only
// consulted in Run's ModeHelp branch (ver its own doc comment),
// which an explicit flag like --tui never reaches.
func TestRunPrefersAnExplicitModeFlagOverABareFileArgument(t *testing.T) {
	handled, _ := Run([]string{"--tui", "programa.dmx"})
	if handled {
		t.Fatalf(`Run(["--tui", "programa.dmx"]): expected handled=false (the --tui flag must win, not trigger a compile)`)
	}
}

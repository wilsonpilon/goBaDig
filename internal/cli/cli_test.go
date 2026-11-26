package cli

import (
	"os"
	"path/filepath"
	"strings"
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

func TestExecuteFileArg(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
		ok   bool
	}{
		{"flag with file", []string{"--execute", "programa.dmx"}, "programa.dmx", true},
		{"short flag", []string{"-execute", "programa.dmx"}, "programa.dmx", true},
		{"flag without a following file", []string{"--execute"}, "", false},
		{"no flag at all", []string{"programa.dmx"}, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := executeFileArg(c.args)
			if ok != c.ok || got != c.want {
				t.Fatalf("executeFileArg(%v) = (%q, %v), want (%q, %v)", c.args, got, ok, c.want, c.ok)
			}
		})
	}
}

// TestEmulatorArg mirrors TestExecuteFileArg's own table style - the
// "--emulator" flag added 2026-11-26, independent of --execute's own
// filename argument (ver emulatorArg's own doc comment on why it
// scans the whole args slice rather than only looking right after
// --execute).
func TestEmulatorArg(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"flag before --execute", []string{"--emulator", "zesarux", "--execute", "p.dzs"}, "zesarux"},
		{"flag after the file", []string{"--execute", "p.dzs", "--emulator", "zesarux"}, "zesarux"},
		{"short flag", []string{"--execute", "p.dzs", "-emulator", "fmsx"}, "fmsx"},
		{"flag without a following name", []string{"--execute", "p.dzs", "--emulator"}, ""},
		{"no flag at all", []string{"--execute", "p.dzs"}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := emulatorArg(c.args); got != c.want {
				t.Fatalf("emulatorArg(%v) = %q, want %q", c.args, got, c.want)
			}
		})
	}
}

// TestRunExecuteHonorsAnExplicitEmulatorFlag confirms "--emulator"
// actually reaches runInEmulatorByExtension end to end through Run()
// - PATH points at an empty directory, same no-real-emulator-window
// pattern as every other test here, but ZEsarUX is a valid ZX81
// backend (dialect.go's emulatorBackends), so this still exercises
// the full --emulator plumbing, not just a rejection.
func TestRunExecuteHonorsAnExplicitEmulatorFlag(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	dzxPath := filepath.Join(dir, "programa.dzx")
	if err := os.WriteFile(dzxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	handled, code := Run([]string{"--execute", dzxPath, "--emulator", "zesarux"})
	if !handled {
		t.Fatalf("Run(--execute %s --emulator zesarux): expected handled=true", dzxPath)
	}
	if code == 0 {
		t.Fatalf("Run(--execute %s --emulator zesarux) with no emulator on PATH: expected a non-zero exit code", dzxPath)
	}
	if _, err := os.Stat(filepath.Join(dir, "programa.p")); err != nil {
		t.Fatalf("expected the .p file to still have been built: %v", err)
	}
}

// TestRunExecuteHonorsAnExplicitEmulatorFlagBeforeTheFile is a
// regression test: "--emulator" BETWEEN "--execute" and the filename
// (exactly the form documented in HelpText) used to make
// executeFileArg mistake the emulator NAME for the file to execute,
// since it always took whatever token came right after "--execute"
// literally - fixed by stripEmulatorFlag, called from Run before any
// other argument parsing runs.
func TestRunExecuteHonorsAnExplicitEmulatorFlagBeforeTheFile(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	dzsPath := filepath.Join(dir, "programa.dzs")
	if err := os.WriteFile(dzsPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	handled, code := Run([]string{"--execute", "--emulator", "zesarux", dzsPath})
	if !handled {
		t.Fatalf("Run(--execute --emulator zesarux %s): expected handled=true", dzsPath)
	}
	if code == 0 {
		t.Fatalf("Run(--execute --emulator zesarux %s) with no emulator on PATH: expected a non-zero exit code", dzsPath)
	}
	if _, err := os.Stat(filepath.Join(dir, "programa.tap")); err != nil {
		t.Fatalf("expected the .tap file to still have been built: %v", err)
	}
}

// TestRunExecuteRejectsAnUnknownEmulatorFlag confirms a backend name
// that doesn't belong to the dispatched dialect is rejected with a
// clear error, not silently ignored or falling back to the default.
func TestRunExecuteRejectsAnUnknownEmulatorFlag(t *testing.T) {
	dir := t.TempDir()
	dmxPath := filepath.Join(dir, "programa.dmx")
	if err := os.WriteFile(dmxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	handled, code := Run([]string{"--execute", dmxPath, "--emulator", "zesarux"})
	if !handled {
		t.Fatalf("Run(--execute %s --emulator zesarux): expected handled=true", dmxPath)
	}
	if code == 0 {
		t.Fatal("Run(--execute ... --emulator zesarux) for MSX: expected a non-zero exit code (zesarux isn't an MSX backend)")
	}
}

// TestRunExecuteBuildsTheDiskEvenWhenTheEmulatorIsMissing mirrors
// console_test.go's own equivalent test for the --cli entry point -
// points PATH at an empty directory so this never actually launches
// a real emulator window, while still confirming the real
// msx.BuildTestDisk step ran (the disk exists) and the emulator's
// own failure is reported with a non-zero exit code.
func TestRunExecuteBuildsTheDiskEvenWhenTheEmulatorIsMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	dmxPath := filepath.Join(dir, "programa.dmx")
	if err := os.WriteFile(dmxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	handled, code := Run([]string{"--execute", dmxPath})
	if !handled {
		t.Fatalf("Run(--execute %s): expected handled=true", dmxPath)
	}
	if code == 0 {
		t.Fatalf("Run(--execute %s) with no emulator on PATH: expected a non-zero exit code", dmxPath)
	}
	diskPath := strings.TrimSuffix(dmxPath, ".dmx") + ".dsk"
	if _, err := os.Stat(diskPath); err != nil {
		t.Fatalf("expected the disk to still have been built at %s: %v", diskPath, err)
	}
}

// TestRunCompilesABareZX81FileArgument mirrors
// TestRunCompilesABareFileArgument for the ZX81 side of
// compileByExtension's dispatch.
func TestRunCompilesABareZX81FileArgument(t *testing.T) {
	dir := t.TempDir()
	dzxPath := filepath.Join(dir, "programa.dzx")
	if err := os.WriteFile(dzxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	handled, code := Run([]string{dzxPath})
	if !handled {
		t.Fatalf("Run(%v): expected handled=true", dzxPath)
	}
	if code != 0 {
		t.Fatalf("Run(%v): expected exit code 0, got %d", dzxPath, code)
	}

	if _, err := os.Stat(filepath.Join(dir, "programa.azx")); err != nil {
		t.Fatalf("expected programa.azx to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "programa.p")); err != nil {
		t.Fatalf("expected programa.p to exist: %v", err)
	}
}

// TestRunExecuteBuildsTheZX81FileEvenWhenTheEmulatorIsMissing mirrors
// TestRunExecuteBuildsTheDiskEvenWhenTheEmulatorIsMissing for the
// ZX81 side - there is no disk to build, just the ".p" file itself.
func TestRunExecuteBuildsTheZX81FileEvenWhenTheEmulatorIsMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	dzxPath := filepath.Join(dir, "programa.dzx")
	if err := os.WriteFile(dzxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	handled, code := Run([]string{"--execute", dzxPath})
	if !handled {
		t.Fatalf("Run(--execute %s): expected handled=true", dzxPath)
	}
	if code == 0 {
		t.Fatalf("Run(--execute %s) with no emulator on PATH: expected a non-zero exit code", dzxPath)
	}
	pPath := strings.TrimSuffix(dzxPath, ".dzx") + ".p"
	if _, err := os.Stat(pPath); err != nil {
		t.Fatalf("expected the .p file to still have been built at %s: %v", pPath, err)
	}
}

// TestRunCompilesABareZXSpectrumFileArgument mirrors
// TestRunCompilesABareFileArgument for the ZX Spectrum side of
// compileByExtension's dispatch.
func TestRunCompilesABareZXSpectrumFileArgument(t *testing.T) {
	dir := t.TempDir()
	dzsPath := filepath.Join(dir, "programa.dzs")
	if err := os.WriteFile(dzsPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	handled, code := Run([]string{dzsPath})
	if !handled {
		t.Fatalf("Run(%v): expected handled=true", dzsPath)
	}
	if code != 0 {
		t.Fatalf("Run(%v): expected exit code 0, got %d", dzsPath, code)
	}

	if _, err := os.Stat(filepath.Join(dir, "programa.azs")); err != nil {
		t.Fatalf("expected programa.azs to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "programa.tap")); err != nil {
		t.Fatalf("expected programa.tap to exist: %v", err)
	}
}

// TestRunExecuteBuildsTheZXSpectrumFileEvenWhenTheEmulatorIsMissing
// mirrors TestRunExecuteBuildsTheDiskEvenWhenTheEmulatorIsMissing for
// the ZX Spectrum side - there is no disk to build, just the ".tap"
// file itself.
func TestRunExecuteBuildsTheZXSpectrumFileEvenWhenTheEmulatorIsMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	dzsPath := filepath.Join(dir, "programa.dzs")
	if err := os.WriteFile(dzsPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	handled, code := Run([]string{"--execute", dzsPath})
	if !handled {
		t.Fatalf("Run(--execute %s): expected handled=true", dzsPath)
	}
	if code == 0 {
		t.Fatalf("Run(--execute %s) with no emulator on PATH: expected a non-zero exit code", dzsPath)
	}
	tapPath := strings.TrimSuffix(dzsPath, ".dzs") + ".tap"
	if _, err := os.Stat(tapPath); err != nil {
		t.Fatalf("expected the .tap file to still have been built at %s: %v", tapPath, err)
	}
}

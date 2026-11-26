package dialect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileByExtensionDispatchesByFileExtension(t *testing.T) {
	dir := t.TempDir()

	dmxPath := filepath.Join(dir, "hello.dmx")
	if err := os.WriteFile(dmxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	classicPath, binaryPath, err := CompileByExtension(dmxPath)
	if err != nil {
		t.Fatalf("CompileByExtension(.dmx): %v", err)
	}
	if want := filepath.Join(dir, "hello.amx"); classicPath != want {
		t.Fatalf("classicPath = %q, want %q", classicPath, want)
	}
	if want := filepath.Join(dir, "hello.bmx"); binaryPath != want {
		t.Fatalf("binaryPath = %q, want %q", binaryPath, want)
	}

	dzxPath := filepath.Join(dir, "hello.dzx")
	if err := os.WriteFile(dzxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	classicPath, binaryPath, err = CompileByExtension(dzxPath)
	if err != nil {
		t.Fatalf("CompileByExtension(.dzx): %v", err)
	}
	if want := filepath.Join(dir, "hello.azx"); classicPath != want {
		t.Fatalf("classicPath = %q, want %q", classicPath, want)
	}
	if want := filepath.Join(dir, "hello.p"); binaryPath != want {
		t.Fatalf("binaryPath = %q, want %q", binaryPath, want)
	}

	dzsPath := filepath.Join(dir, "hello.dzs")
	if err := os.WriteFile(dzsPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	classicPath, binaryPath, err = CompileByExtension(dzsPath)
	if err != nil {
		t.Fatalf("CompileByExtension(.dzs): %v", err)
	}
	if want := filepath.Join(dir, "hello.azs"); classicPath != want {
		t.Fatalf("classicPath = %q, want %q", classicPath, want)
	}
	if want := filepath.Join(dir, "hello.tap"); binaryPath != want {
		t.Fatalf("binaryPath = %q, want %q", binaryPath, want)
	}
}

func TestCompileByExtensionRejectsAnUnknownExtension(t *testing.T) {
	_, _, err := CompileByExtension("programa.txt")
	if err == nil || !strings.Contains(err.Error(), "unrecognized dialect") {
		t.Fatalf("CompileByExtension(.txt): want an 'unrecognized dialect' error, got %v", err)
	}
}

// TestRunInEmulatorByExtensionDispatchesByFileExtension confirms both
// dialects build their own ready-to-run artifact even when the
// emulator itself can't be found - PATH is overridden to an empty
// directory so this never actually launches a real emulator window
// during a test run (same pattern already used by
// msx/console_test.go's own equivalent coverage).
func TestRunInEmulatorByExtensionDispatchesByFileExtension(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	dir := t.TempDir()

	dmxPath := filepath.Join(dir, "hello.dmx")
	if err := os.WriteFile(dmxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := RunInEmulatorByExtension(dmxPath, ""); err == nil {
		t.Fatal("RunInEmulatorByExtension(.dmx) with no emulator on PATH: want error, got nil")
	}
	if _, err := os.Stat(filepath.Join(dir, "hello.dsk")); err != nil {
		t.Fatalf("expected the MSX test disk to still have been built: %v", err)
	}

	dzxPath := filepath.Join(dir, "hello.dzx")
	if err := os.WriteFile(dzxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := RunInEmulatorByExtension(dzxPath, ""); err == nil {
		t.Fatal("RunInEmulatorByExtension(.dzx) with no emulator on PATH: want error, got nil")
	}
	if _, err := os.Stat(filepath.Join(dir, "hello.p")); err != nil {
		t.Fatalf("expected the ZX81 .p file to still have been built: %v", err)
	}

	dzsPath := filepath.Join(dir, "hello.dzs")
	if err := os.WriteFile(dzsPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := RunInEmulatorByExtension(dzsPath, ""); err == nil {
		t.Fatal("RunInEmulatorByExtension(.dzs) with no emulator on PATH: want error, got nil")
	}
	if _, err := os.Stat(filepath.Join(dir, "hello.tap")); err != nil {
		t.Fatalf("expected the ZX Spectrum .tap file to still have been built: %v", err)
	}
}

func TestRunInEmulatorByExtensionRejectsAnUnknownExtension(t *testing.T) {
	_, err := RunInEmulatorByExtension("programa.txt", "")
	if err == nil || !strings.Contains(err.Error(), "unrecognized dialect") {
		t.Fatalf("RunInEmulatorByExtension(.txt): want an 'unrecognized dialect' error, got %v", err)
	}
}

// TestRunInEmulatorByExtensionHonorsAnExplicitBackend confirms each
// dialect's own alternative backends (EmulatorBackends) are reachable,
// and that the artifact still gets built even when that specific
// backend can't be found on PATH - same no-emulator-on-PATH pattern as
// TestRunInEmulatorByExtensionDispatchesByFileExtension.
func TestRunInEmulatorByExtensionHonorsAnExplicitBackend(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	dir := t.TempDir()

	cases := []struct {
		ext, backend, artifact string
	}{
		{".dmx", "fmsx", "hello.dsk"},
		{".dmx", "bluemsx+", "hello.dsk"},
		{".dzx", "zesarux", "hello.p"},
		{".dzs", "zesarux", "hello.tap"},
	}
	for _, c := range cases {
		path := filepath.Join(dir, "hello"+c.ext)
		if err := os.WriteFile(path, []byte("print 1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := RunInEmulatorByExtension(path, c.backend); err == nil {
			t.Fatalf("RunInEmulatorByExtension(%s, %q) with no emulator on PATH: want error, got nil", c.ext, c.backend)
		}
		if _, err := os.Stat(filepath.Join(dir, c.artifact)); err != nil {
			t.Fatalf("%s/%s: expected %s to still have been built: %v", c.ext, c.backend, c.artifact, err)
		}
	}
}

func TestRunInEmulatorByExtensionRejectsAnUnknownBackend(t *testing.T) {
	dir := t.TempDir()
	dzsPath := filepath.Join(dir, "hello.dzs")
	if err := os.WriteFile(dzsPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := RunInEmulatorByExtension(dzsPath, "nope"); err == nil || !strings.Contains(err.Error(), "unknown --emulator") {
		t.Fatalf(`RunInEmulatorByExtension(.dzs, "nope"): want an "unknown --emulator" error, got %v`, err)
	}

	// A backend valid for a DIFFERENT dialect is still rejected - each
	// dialect's own list is exactly what it supports, not shared
	// vocabulary.
	if _, err := RunInEmulatorByExtension(dzsPath, "fmsx"); err == nil || !strings.Contains(err.Error(), "unknown --emulator") {
		t.Fatalf(`RunInEmulatorByExtension(.dzs, "fmsx"): want an "unknown --emulator" error, got %v`, err)
	}
}

package cli

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
	classicPath, binaryPath, err := compileByExtension(dmxPath)
	if err != nil {
		t.Fatalf("compileByExtension(.dmx): %v", err)
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
	classicPath, binaryPath, err = compileByExtension(dzxPath)
	if err != nil {
		t.Fatalf("compileByExtension(.dzx): %v", err)
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
	classicPath, binaryPath, err = compileByExtension(dzsPath)
	if err != nil {
		t.Fatalf("compileByExtension(.dzs): %v", err)
	}
	if want := filepath.Join(dir, "hello.azs"); classicPath != want {
		t.Fatalf("classicPath = %q, want %q", classicPath, want)
	}
	if want := filepath.Join(dir, "hello.tap"); binaryPath != want {
		t.Fatalf("binaryPath = %q, want %q", binaryPath, want)
	}
}

func TestCompileByExtensionRejectsAnUnknownExtension(t *testing.T) {
	_, _, err := compileByExtension("programa.txt")
	if err == nil || !strings.Contains(err.Error(), "unrecognized dialect") {
		t.Fatalf("compileByExtension(.txt): want an 'unrecognized dialect' error, got %v", err)
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
	if _, err := runInEmulatorByExtension(dmxPath); err == nil {
		t.Fatal("runInEmulatorByExtension(.dmx) with no emulator on PATH: want error, got nil")
	}
	if _, err := os.Stat(filepath.Join(dir, "hello.dsk")); err != nil {
		t.Fatalf("expected the MSX test disk to still have been built: %v", err)
	}

	dzxPath := filepath.Join(dir, "hello.dzx")
	if err := os.WriteFile(dzxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runInEmulatorByExtension(dzxPath); err == nil {
		t.Fatal("runInEmulatorByExtension(.dzx) with no emulator on PATH: want error, got nil")
	}
	if _, err := os.Stat(filepath.Join(dir, "hello.p")); err != nil {
		t.Fatalf("expected the ZX81 .p file to still have been built: %v", err)
	}

	dzsPath := filepath.Join(dir, "hello.dzs")
	if err := os.WriteFile(dzsPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runInEmulatorByExtension(dzsPath); err == nil {
		t.Fatal("runInEmulatorByExtension(.dzs) with no emulator on PATH: want error, got nil")
	}
	if _, err := os.Stat(filepath.Join(dir, "hello.tap")); err != nil {
		t.Fatalf("expected the ZX Spectrum .tap file to still have been built: %v", err)
	}
}

func TestRunInEmulatorByExtensionRejectsAnUnknownExtension(t *testing.T) {
	_, err := runInEmulatorByExtension("programa.txt")
	if err == nil || !strings.Contains(err.Error(), "unrecognized dialect") {
		t.Fatalf("runInEmulatorByExtension(.txt): want an 'unrecognized dialect' error, got %v", err)
	}
}

package msx

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/options"
)

func TestResolveAsmsxPathReturnsAnErrorWhenNotConfiguredAndNotOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, err := ResolveAsmsxPath(nil); err == nil {
		t.Fatal("ResolveAsmsxPath(nil) with an empty PATH: want error, got nil")
	}
}

func TestResolveAsmsxPathUsesTheConfiguredPathWhenItExists(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	configured := filepath.Join(t.TempDir(), "myasmsx.exe")
	if err := os.WriteFile(configured, []byte("not a real binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := store.SetOption(options.CategoryMSX, "ASMSX_PATH", configured); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveAsmsxPath(store)
	if err != nil {
		t.Fatalf("ResolveAsmsxPath: %v", err)
	}
	if got != configured {
		t.Fatalf("ResolveAsmsxPath = %q, want %q", got, configured)
	}
}

func TestResolveAsmsxPathIgnoresAConfiguredPathThatDoesNotExist(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.SetOption(options.CategoryMSX, "ASMSX_PATH", filepath.Join(t.TempDir(), "nope.exe")); err != nil {
		t.Fatal(err)
	}

	if _, err := ResolveAsmsxPath(store); err == nil {
		t.Fatal("ResolveAsmsxPath with a configured-but-missing file and an empty PATH: want error, got nil")
	}
}

func TestResolveAsmsxPathFallsBackToSystemPath(t *testing.T) {
	binDir := t.TempDir()
	name := "asmsx"
	if runtime.GOOS == "windows" {
		name = "asmsx.exe"
	}
	fakePath := filepath.Join(binDir, name)
	if err := os.WriteFile(fakePath, []byte("not a real binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	got, err := ResolveAsmsxPath(nil)
	if err != nil {
		t.Fatalf("ResolveAsmsxPath(nil): %v", err)
	}
	if got != fakePath {
		t.Fatalf("ResolveAsmsxPath(nil) = %q, want %q", got, fakePath)
	}
}

// buildFakeAsmsx compiles a tiny real Go program that stands in for
// asmsx: it reads its own single argument (the ".asm" filename, same
// as CompileAsm passes), prints a fixed message, and - unless mode is
// "fail" or "nobin" - writes "<base>.bin" next to it before exiting.
// This exercises CompileAsm's actual subprocess/output/file-detection
// logic without needing a real asmsx install (ver OUTLINE.md: "se
// houver um asmsx real disponível... caso contrário só o caminho de
// erro é testado automaticamente" - a built fake covers the success
// path too, deterministically, same effort as the plan's own
// fallback would have needed anyway).
func buildFakeAsmsx(t *testing.T, mode string) string {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "fakeasmsx.go")
	source := `package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	mode := "` + mode + `"
	if len(os.Args) < 2 {
		fmt.Println("missing asm filename")
		os.Exit(1)
	}
	asm := os.Args[1]
	if mode == "fail" {
		fmt.Println("erro fake: linha 1: sintaxe invalida")
		os.Exit(1)
	}
	fmt.Println("compilado via fake asmsx")
	if mode == "nobin" {
		os.Exit(0)
	}
	base := strings.TrimSuffix(asm, ".asm")
	if err := os.WriteFile(base+".bin", []byte{0xC3, 0x00, 0x40}, 0o644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
`
	if err := os.WriteFile(src, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(dir, "fakeasmsx")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	build := exec.Command("go", "build", "-o", exe, src)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building fake asmsx: %v: %s", err, output)
	}
	return exe
}

func TestCompileAsmSucceedsAndFindsTheGeneratedBin(t *testing.T) {
	fake := buildFakeAsmsx(t, "ok")
	dir := t.TempDir()
	asmPath := filepath.Join(dir, "game.asm")
	if err := os.WriteFile(asmPath, []byte("; nothing"), 0o644); err != nil {
		t.Fatal(err)
	}

	binPath, output, err := CompileAsm(asmPath, fake)
	if err != nil {
		t.Fatalf("CompileAsm: %v (output: %s)", err, output)
	}
	if !strings.Contains(output, "compilado via fake asmsx") {
		t.Fatalf("output = %q, want it to contain the fake tool's message", output)
	}
	wantBin := filepath.Join(dir, "game.bin")
	if binPath != wantBin {
		t.Fatalf("binPath = %q, want %q", binPath, wantBin)
	}
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected %s to exist: %v", binPath, err)
	}
}

func TestCompileAsmWrapsTheToolsOutputOnFailure(t *testing.T) {
	fake := buildFakeAsmsx(t, "fail")
	dir := t.TempDir()
	asmPath := filepath.Join(dir, "game.asm")
	if err := os.WriteFile(asmPath, []byte("; nothing"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := CompileAsm(asmPath, fake)
	if err == nil {
		t.Fatal("CompileAsm with a failing tool: want error, got nil")
	}
	if !strings.Contains(err.Error(), "sintaxe invalida") {
		t.Fatalf("error = %v, want it to contain the tool's own message", err)
	}
}

func TestCompileAsmReportsAMissingBinEvenWhenTheToolExitsCleanly(t *testing.T) {
	fake := buildFakeAsmsx(t, "nobin")
	dir := t.TempDir()
	asmPath := filepath.Join(dir, "game.asm")
	if err := os.WriteFile(asmPath, []byte("; nothing"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := CompileAsm(asmPath, fake)
	if err == nil {
		t.Fatal("CompileAsm when the tool never writes a .bin: want error, got nil")
	}
	if !strings.Contains(err.Error(), "game.bin") {
		t.Fatalf("error = %v, want it to name the missing game.bin", err)
	}
}

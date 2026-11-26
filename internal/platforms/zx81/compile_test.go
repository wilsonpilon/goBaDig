package zx81

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestCompileFileWritesTheAzxAndPNextToTheSourceFile is the first
// end-to-end test of the real "read a .dzx file on disk, write
// .azx/.p next to it" path - mirrors
// msx.TestCompileFileWritesTheAmxAndBmxNextToTheSourceFile exactly,
// ZX81's own REM wording ("REM", the word, not "'") and extensions.
func TestCompileFileWritesTheAzxAndPNextToTheSourceFile(t *testing.T) {
	dir := t.TempDir()
	dzxPath := filepath.Join(dir, "hello.dzx")
	if err := os.WriteFile(dzxPath, []byte("counter=0\nprint counter\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	azxPath, pPath, err := CompileFile(dzxPath, nil)
	if err != nil {
		t.Fatalf("CompileFile: %v", err)
	}
	if want := filepath.Join(dir, "hello.azx"); azxPath != want {
		t.Fatalf("azxPath = %q, want %q", azxPath, want)
	}
	if want := filepath.Join(dir, "hello.p"); pPath != want {
		t.Fatalf("pPath = %q, want %q", pPath, want)
	}

	azxContent, err := os.ReadFile(azxPath)
	if err != nil {
		t.Fatal(err)
	}
	wantAzx := "10 REM Converted by goBaDig 2026\r\n" +
		"20 REM https://github.com/wilsonpilon/goBaDig\r\n" +
		"30 let z=0\r\n" +
		"40 print z\r\n"
	if string(azxContent) != wantAzx {
		t.Fatalf("azx content = %q, want %q", azxContent, wantAzx)
	}

	pContent, err := os.ReadFile(pPath)
	if err != nil {
		t.Fatal(err)
	}
	wantP, err := Tokenize([]string{
		"10 REM Converted by goBaDig 2026",
		"20 REM https://github.com/wilsonpilon/goBaDig",
		"30 let z=0",
		"40 print z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pContent, wantP) {
		t.Fatalf("p content = %x, want %x", pContent, wantP)
	}
}

func TestCompileFileReturnsAnErrorForAMissingSourceFile(t *testing.T) {
	if _, _, err := CompileFile(filepath.Join(t.TempDir(), "nope.dzx"), nil); err == nil {
		t.Fatalf("CompileFile(missing file): want error, got nil")
	}
}

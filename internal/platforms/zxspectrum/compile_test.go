package zxspectrum

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestCompileFileWritesTheAzsAndTapNextToTheSourceFile mirrors
// msx.TestCompileFileWritesTheAmxAndBmxNextToTheSourceFile/
// zx81.TestCompileFileWritesTheAzxAndPNextToTheSourceFile - ZX
// Spectrum's own wording ("PRINT"/"LET" always uppercase) and
// extensions (.azs/.tap).
func TestCompileFileWritesTheAzsAndTapNextToTheSourceFile(t *testing.T) {
	dir := t.TempDir()
	dzsPath := filepath.Join(dir, "hello.dzs")
	if err := os.WriteFile(dzsPath, []byte("counter=0\nprint counter\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	azsPath, tapPath, err := CompileFile(dzsPath, nil)
	if err != nil {
		t.Fatalf("CompileFile: %v", err)
	}
	if want := filepath.Join(dir, "hello.azs"); azsPath != want {
		t.Fatalf("azsPath = %q, want %q", azsPath, want)
	}
	if want := filepath.Join(dir, "hello.tap"); tapPath != want {
		t.Fatalf("tapPath = %q, want %q", tapPath, want)
	}

	azsContent, err := os.ReadFile(azsPath)
	if err != nil {
		t.Fatal(err)
	}
	wantAzs := "10 REM Converted with Basic Dignified\r\n" +
		"20 REM https://github.com/farique1/basic-dignified\r\n" +
		"30 LET counter=0\r\n" +
		"40 PRINT counter\r\n"
	if string(azsContent) != wantAzs {
		t.Fatalf("azs content = %q, want %q", azsContent, wantAzs)
	}

	tapContent, err := os.ReadFile(tapPath)
	if err != nil {
		t.Fatal(err)
	}
	program, err := Tokenize(
		"10 REM Converted with Basic Dignified\n"+
			"20 REM https://github.com/farique1/basic-dignified\n"+
			"30 LET counter=0\n"+
			"40 PRINT counter\n",
		"48")
	if err != nil {
		t.Fatal(err)
	}
	wantTap, err := MakeTap(program, "hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tapContent, wantTap) {
		t.Fatalf("tap content = %x, want %x", tapContent, wantTap)
	}
}

func TestCompileFileReturnsAnErrorForAMissingSourceFile(t *testing.T) {
	if _, _, err := CompileFile(filepath.Join(t.TempDir(), "nope.dzs"), nil); err == nil {
		t.Fatalf("CompileFile(missing file): want error, got nil")
	}
}

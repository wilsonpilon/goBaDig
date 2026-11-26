package msx

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestCompileToBinaryTokenizesTheGeneratedClassicListing confirms the
// wiring itself: CompileToBinary's bytes for a small Dignified
// program must be EXACTLY what Tokenize produces from that same
// program's own Classic text (core.Run's GenerateResult.Code) -
// catches a wiring bug (wrong lines passed through, an extra/missing
// line, etc.) even though Tokenize's own byte-level correctness
// already has its dedicated golden-vector tests (tokenizer_test.go).
func TestCompileToBinaryTokenizesTheGeneratedClassicListing(t *testing.T) {
	ctx, platform := contextPlatform(t, "counter=0", "print counter")

	got, err := CompileToBinary(ctx, platform)
	if err != nil {
		t.Fatalf("CompileToBinary: %v", err)
	}

	want, err := Tokenize([]string{"10 zz=0", "20 print zz"})
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("CompileToBinary = %x, want %x", got, want)
	}
}

// TestCompileToBinaryPropagatesAParseError confirms a Dignified-level
// error (not a Tokenize-level one) still surfaces through
// CompileToBinary - e.g. a FUNC never closed by a matching RET.
func TestCompileToBinaryPropagatesAParseError(t *testing.T) {
	ctx, platform := contextPlatform(t, "func .broken()")

	if _, err := CompileToBinary(ctx, platform); err == nil {
		t.Fatalf("CompileToBinary: want an error for an unclosed FUNC, got nil")
	}
}

// TestCompileFileWritesTheAmxAndBmxNextToTheSourceFile is the first
// end-to-end test of the real "read a .dmx file on disk, write
// .amx/.bmx next to it" path (badig.py's own dignified(), ver
// CompileFile's own doc comment) - every piece it chains together
// (core.Run, Tokenize, core.LoadLines) already had its own tests;
// this one is about the WIRING and the real files produced.
func TestCompileFileWritesTheAmxAndBmxNextToTheSourceFile(t *testing.T) {
	dir := t.TempDir()
	dmxPath := filepath.Join(dir, "hello.dmx")
	if err := os.WriteFile(dmxPath, []byte("counter=0\nprint counter\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	amxPath, bmxPath, err := CompileFile(dmxPath, nil)
	if err != nil {
		t.Fatalf("CompileFile: %v", err)
	}
	if want := filepath.Join(dir, "hello.amx"); amxPath != want {
		t.Fatalf("amxPath = %q, want %q", amxPath, want)
	}
	if want := filepath.Join(dir, "hello.bmx"); bmxPath != want {
		t.Fatalf("bmxPath = %q, want %q", bmxPath, want)
	}

	amxContent, err := os.ReadFile(amxPath)
	if err != nil {
		t.Fatal(err)
	}
	// DefaultCompileSettings has RemHeader=true, so the 2-line REM
	// header comes before the real content, at lines 30/40 (header
	// at 10/20, LineStep=10) - CRLF line endings, matching the real
	// badig.py's own output on Windows (ver CompileFile's doc
	// comment).
	wantAmx := "10 'Converted by goBaDig 2026\r\n" +
		"20 'https://github.com/wilsonpilon/goBaDig\r\n" +
		"30 zz=0\r\n" +
		"40 print zz\r\n"
	if string(amxContent) != wantAmx {
		t.Fatalf("amx content = %q, want %q", amxContent, wantAmx)
	}

	bmxContent, err := os.ReadFile(bmxPath)
	if err != nil {
		t.Fatal(err)
	}
	wantBmx, err := Tokenize([]string{
		"10 'Converted by goBaDig 2026",
		"20 'https://github.com/wilsonpilon/goBaDig",
		"30 zz=0",
		"40 print zz",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bmxContent, wantBmx) {
		t.Fatalf("bmx content = %x, want %x", bmxContent, wantBmx)
	}
}

func TestCompileFileReturnsAnErrorForAMissingSourceFile(t *testing.T) {
	if _, _, err := CompileFile(filepath.Join(t.TempDir(), "nope.dmx"), nil); err == nil {
		t.Fatalf("CompileFile(missing file): want error, got nil")
	}
}

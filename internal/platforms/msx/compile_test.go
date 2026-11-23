package msx

import (
	"bytes"
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

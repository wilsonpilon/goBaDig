package msx

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
)

// These 13 cases are the golden vectors from docs/golden-vectors/
// msx-tokenizer.md, transcribed 1:1 (same inputs, same expected hex) -
// they were captured by running the real Python tokenizer
// (badig/msx/msxbatoken.py) and are the ground truth this port must
// match byte-exact. See that file for what each case is pinning down.

func hexBytes(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(strings.ReplaceAll(s, " ", ""))
	if err != nil {
		t.Fatalf("invalid hex literal %q: %v", s, err)
	}
	return b
}

func tokenizeOrFatal(t *testing.T, lines ...string) []byte {
	t.Helper()
	got, err := Tokenize(lines)
	if err != nil {
		t.Fatalf("Tokenize(%v): %v", lines, err)
	}
	return got
}

func assertHex(t *testing.T, got []byte, want string) {
	t.Helper()
	wantBytes := hexBytes(t, want)
	if !bytes.Equal(got, wantBytes) {
		t.Fatalf("got  %x\nwant %x", got, wantBytes)
	}
}

func TestNumberedBasicLineHasExpectedMSXTokens(t *testing.T) {
	assertHex(t, tokenizeOrFatal(t, "10 PRINT 1"), "ff 09 80 0a 00 91 20 12 00 00 00")
}

func TestPrintAndQuestionMarkTokenizeIdentically(t *testing.T) {
	a := tokenizeOrFatal(t, "10 PRINT 1")
	b := tokenizeOrFatal(t, "10 ? 1")
	if !bytes.Equal(a, b) {
		t.Fatalf("PRINT and ? tokenized differently:\nPRINT: %x\n?:     %x", a, b)
	}
}

func TestSmallIntegerConstantsUseTheDedicatedSingleByteRange(t *testing.T) {
	cases := []struct {
		value int
		want  string
	}{
		{0, "ff 09 80 0a 00 91 20 11 00 00 00"},
		{1, "ff 09 80 0a 00 91 20 12 00 00 00"},
		{9, "ff 09 80 0a 00 91 20 1a 00 00 00"},
	}
	for _, c := range cases {
		got, err := Tokenize([]string{"10 PRINT " + strconv.Itoa(c.value)})
		if err != nil {
			t.Fatalf("Tokenize: %v", err)
		}
		assertHex(t, got, c.want)
	}
}

func TestOneByteIntegerConstantAboveNine(t *testing.T) {
	assertHex(t, tokenizeOrFatal(t, "10 PRINT 255"), "ff 0a 80 0a 00 91 20 0f ff 00 00 00")
}

func TestTwoByteIntegerConstant(t *testing.T) {
	assertHex(t, tokenizeOrFatal(t, "10 PRINT 32767"), "ff 0b 80 0a 00 91 20 1c ff 7f 00 00 00")
}

func TestUnaryMinusIsTheMinusOperatorNotANegativeLiteral(t *testing.T) {
	assertHex(t, tokenizeOrFatal(t, "10 PRINT -1"), "ff 0a 80 0a 00 91 20 f2 12 00 00 00")
}

func TestSinglePrecisionFloatUsesDecimalDigitPackingNotBinary(t *testing.T) {
	assertHex(t, tokenizeOrFatal(t, "10 PRINT 3.14"), "ff 0d 80 0a 00 91 20 1d 41 31 40 00 00 00 00")
}

func TestDoublePrecisionFloatSuffixWidensTheEncoding(t *testing.T) {
	assertHex(t, tokenizeOrFatal(t, "10 PRINT 3.14#"), "ff 11 80 0a 00 91 20 1f 41 31 40 00 00 00 00 00 00 00 00")
}

func TestStringLiteralIsStoredAsLiteralQuotedASCII(t *testing.T) {
	assertHex(t, tokenizeOrFatal(t, `10 PRINT "HI"`), "ff 0c 80 0a 00 91 20 22 48 49 22 00 00 00")
}

func TestRemKeywordAndApostropheShorthandEncodeDifferently(t *testing.T) {
	rem := tokenizeOrFatal(t, "10 REM hello")
	apostrophe := tokenizeOrFatal(t, "10 ' hello")

	assertHex(t, rem, "ff 0d 80 0a 00 8f 20 68 65 6c 6c 6f 00 00 00")
	assertHex(t, apostrophe, "ff 0f 80 0a 00 3a 8f e6 20 68 65 6c 6c 6f 00 00 00")
	if bytes.Equal(rem, apostrophe) {
		t.Fatal("REM and ' tokenized identically; they must differ")
	}
}

func TestColonSeparatesMultipleStatementsOnOneLine(t *testing.T) {
	assertHex(t, tokenizeOrFatal(t, "10 PRINT 1:PRINT 2"), "ff 0d 80 0a 00 91 20 12 3a 91 20 13 00 00 00")
}

func TestExtendedFunctionUsesTheFFPrefix(t *testing.T) {
	assertHex(t, tokenizeOrFatal(t, "10 PRINT PEEK(0)"), "ff 0d 80 0a 00 91 20 ff 97 28 11 29 00 00 00")
}

func TestThenBranchTargetIsARawTwoByteValueNotATypedConstant(t *testing.T) {
	got := tokenizeOrFatal(t, "10 IF A=1 THEN 20", "20 PRINT 1")
	assertHex(t, got, "ff 11 80 0a 00 8b 20 41 ef 12 20 da 20 0e 14 00 00 19 80 14 00 91 20 12 00 00 00")
}

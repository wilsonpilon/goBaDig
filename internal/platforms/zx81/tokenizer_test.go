package zx81

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

// lineBody scans a complete ".p" image for the line with the given
// number and returns its tokenized body (everything between the
// line's own 4-byte record header and the next one) - mirrors
// TokenizerTests.line_body in test_tokenizer_zx81.py. Returns nil if
// the line isn't found.
func lineBody(program []byte, number int) []byte {
	offset := 116
	end := int(program[3]) | int(program[4])<<8
	for offset < end {
		length := int(program[offset+2]) | int(program[offset+3])<<8
		lineNumber := int(program[offset])<<8 | int(program[offset+1])
		if lineNumber == number {
			return program[offset+4 : offset+4+length]
		}
		offset += 4 + length
	}
	return nil
}

func hexBytes(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(strings.ReplaceAll(s, " ", ""))
	if err != nil {
		t.Fatalf("hexBytes(%q): %v", s, err)
	}
	return b
}

// TestNumberedBasicLineHasExpectedZx81Tokens mirrors
// test_numbered_basic_line_has_expected_zx81_tokens: confirms the
// base case end to end (keyword token, hidden-float number, line
// terminator) at the real, fixed program-area offset.
func TestNumberedBasicLineHasExpectedZx81Tokens(t *testing.T) {
	program, err := Tokenize([]string{"10 PRINT 1"})
	if err != nil {
		t.Fatal(err)
	}
	got := program[116:129]
	want := hexBytes(t, "00 0a 09 00 f5 1d 7e 81 00 00 00 00 76")
	if !bytes.Equal(got, want) {
		t.Fatalf("program[116:129] = % x, want % x", got, want)
	}
}

func TestInverseVideoAndGraphicsEscape(t *testing.T) {
	program, err := Tokenize([]string{`10 PRINT "%A \:'"`})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(program, []byte{0xA6}) {
		t.Fatalf("expected inverse-video byte 0xA6 in %x", program)
	}
	if !bytes.Contains(program, []byte{0x07}) {
		t.Fatalf("expected graphics escape byte 0x07 in %x", program)
	}
}

func TestRemAbsorbsKeywordSeparator(t *testing.T) {
	program, err := Tokenize([]string{"10 REM raw text"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(program, hexBytes(t, "ea 37 26 3c")) {
		t.Fatalf("expected ea 37 26 3c in %x", program)
	}
}

// TestLiteralHexEscapesHaveExactLineBytes mirrors
// test_literal_hex_escapes_have_exact_line_bytes - every case is a
// golden vector captured from the real Python tokenizer (see
// docs/golden-vectors/zx81-tokenizer.md).
func TestLiteralHexEscapesHaveExactLineBytes(t *testing.T) {
	cases := []struct {
		source string
		want   string
	}{
		{`600 REM \{D8}\{D8}\{D8}END OF GAME\{D8}\{D8}\{D8}`,
			"EA D8 D8 D8 2A 33 29 00 34 2B 00 2C 26 32 2A D8 D8 D8 76"},
		{`1 REM \{C9}\{00}`, "EA C9 00 76"},
		{`10 PRINT "\{F5}"`, "F5 0B F5 0B 76"},
		{`20 LET \{35}OKE=6000`, "F1 35 34 30 2A 14 22 1C 1C 1C 7E 8D 3B 80 00 00 76"},
		{`30 LET A=0\{7E}\{81}\{00}\{00}\{00}\{00}`, "F1 26 14 1C 7E 81 00 00 00 00 76"},
		{`30 LET A=1e999\{7e}\{81}\{00}\{00}\{00}\{00}`, "F1 26 14 1D 2A 25 25 25 7E 81 00 00 00 00 76"},
		{`10 PRINT "\{0b}1":\{ea}PRINT 2`, "F5 0B 0B 1D 0B 0E EA F5 1E 7E 82 00 00 00 00 76"},
		{`10 PRINT \{35}123`, "F5 35 1D 1E 1F 7E 87 76 00 00 00 76"},
		{`10 LET A=0 \{7e}\{81}\{00}\{00}\{00}\{00}`, "F1 26 14 1C 7E 00 00 00 00 00 7E 81 00 00 00 00 76"},
	}
	for _, c := range cases {
		t.Run(c.source, func(t *testing.T) {
			program, err := Tokenize([]string{c.source})
			if err != nil {
				t.Fatal(err)
			}
			length := int(program[118]) | int(program[119])<<8
			got := program[120 : 120+length]
			want := hexBytes(t, c.want)
			if !bytes.Equal(got, want) {
				t.Fatalf("body = % x, want % x", got, want)
			}
		})
	}
}

// TestHiddenNumbersRoundAndRespectTheExponentRange mirrors
// test_hidden_numbers_round_and_respect_the_exponent_range - the core
// golden vectors for the ZX81's "hidden floating point" number
// encoding (Sinclair bias-128 exponent + 32-bit mantissa, not IEEE754).
func TestHiddenNumbersRoundAndRespectTheExponentRange(t *testing.T) {
	cases := map[string]string{
		"0":            "00 00 00 00 00",
		"1":            "81 00 00 00 00",
		"0.5":          "80 00 00 00 00",
		"10":           "84 20 00 00 00",
		"16514":        "8F 01 04 00 00",
		"1.9999999999": "82 00 00 00 00",
		"3.14159265":   "82 49 0F DA 9E",
	}
	for text, want := range cases {
		t.Run(text, func(t *testing.T) {
			encoded, err := numberBytes(text, 10)
			if err != nil {
				t.Fatal(err)
			}
			got := encoded[len(encoded)-5:]
			if !bytes.Equal(got, hexBytes(t, want)) {
				t.Fatalf("numberBytes(%q) tail = % x, want % x", text, got, hexBytes(t, want))
			}
		})
	}

	for _, text := range []string{"1.8E38", "1E-39"} {
		t.Run(text, func(t *testing.T) {
			_, err := numberBytes(text, 10)
			if err == nil || !strings.Contains(err.Error(), "out of range") {
				t.Fatalf("numberBytes(%q): want an 'out of range' error, got %v", text, err)
			}
		})
	}
}

func TestAllHexByteValuesAreLiteralInRem(t *testing.T) {
	source := "10 REM "
	for v := range 256 {
		source += "\\{" + hexByteText(byte(v)) + "}"
	}

	program, err := Tokenize([]string{source})
	if err != nil {
		t.Fatal(err)
	}
	length := int(program[118]) | int(program[119])<<8
	got := program[120 : 120+length]

	want := append([]byte{0xEA}, allBytes()...)
	want = append(want, 0x76)
	if !bytes.Equal(got, want) {
		t.Fatalf("body = % x, want % x", got, want)
	}
}

func hexByteText(v byte) string {
	const digits = "0123456789abcdef"
	return string([]byte{digits[v>>4], digits[v&0xF]})
}

func allBytes() []byte {
	b := make([]byte, 256)
	for i := range b {
		b[i] = byte(i)
	}
	return b
}

func TestRejectsMalformedHexEscapesInEveryContext(t *testing.T) {
	escapes := []string{`\{D}`, `\{GG}`, `\{D8`, `\{}`, `\{D80}`}
	bodies := []string{"REM ", `PRINT "`, "LET A="}
	for _, escape := range escapes {
		for _, prefix := range bodies {
			suffix := ""
			if prefix == `PRINT "` {
				suffix = `"`
			}
			source := "40 " + prefix + escape + suffix
			t.Run(source, func(t *testing.T) {
				_, err := Tokenize([]string{source})
				if err == nil || !strings.Contains(err.Error(), "line 40:") {
					t.Fatalf("Tokenize(%q): want a 'line 40:' error, got %v", source, err)
				}
			})
		}
	}
}

func TestProgramPointersFollowProgramAndDisplayFile(t *testing.T) {
	program, err := Tokenize([]string{"10 PRINT 1"})
	if err != nil {
		t.Fatal(err)
	}
	dataFile := int(program[3]) | int(program[4])<<8
	variables := int(program[7]) | int(program[8])<<8

	if want := 0x407D + 13; dataFile != want {
		t.Fatalf("dataFile = %d, want %d", dataFile, want)
	}
	if want := dataFile + 793; variables != want {
		t.Fatalf("variables = %d, want %d", variables, want)
	}
	if tail := program[len(program)-2:]; !bytes.Equal(tail, []byte{0x80, 0x76}) {
		t.Fatalf("final 2 bytes = % x, want 80 76", tail)
	}
}

func TestCompoundOperatorsUseTheirZx81Tokens(t *testing.T) {
	program, err := Tokenize([]string{"10 1<>2 1<=2 1>=2 2**3"})
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []byte{221, 219, 220, 216} {
		if !bytes.Contains(program, []byte{code}) {
			t.Fatalf("expected byte %d in %x", code, program)
		}
	}
}

// TestRemxGeneratesMachineCodeRemLines mirrors the pure-tokenizer
// slice of test_remx_generates_machine_code_rem_lines (the Python
// test runs it through the full Dignified pipeline; this feeds the
// same already-numbered lines straight to Tokenize, since remx is
// recognized at this level regardless).
func TestRemxGeneratesMachineCodeRemLines(t *testing.T) {
	program, err := Tokenize([]string{"remx 1,6", "10 PRINT 1"})
	if err != nil {
		t.Fatal(err)
	}
	want := append([]byte{0xEA}, bytes.Repeat([]byte{0x1C}, 6)...)
	want = append(want, 0x76)
	if got := lineBody(program, 1); !bytes.Equal(got, want) {
		t.Fatalf("lineBody(1) = % x, want % x", got, want)
	}
}

func TestRemxDefaultFillByteIsTheDigitZero(t *testing.T) {
	program, err := Tokenize([]string{"remx 9000,3"})
	if err != nil {
		t.Fatal(err)
	}
	want := hexBytes(t, "EA 1C 1C 1C 76")
	if got := lineBody(program, 9000); !bytes.Equal(got, want) {
		t.Fatalf("lineBody(9000) = % x, want % x", got, want)
	}
}

func TestRemxQuotedFillByte(t *testing.T) {
	program, err := Tokenize([]string{`remx 9000,3,"A"`})
	if err != nil {
		t.Fatal(err)
	}
	// 'A' -> characterCode('a') == lowerByte('A')-59 == 0x26.
	want := hexBytes(t, "EA 26 26 26 76")
	if got := lineBody(program, 9000); !bytes.Equal(got, want) {
		t.Fatalf("lineBody(9000) = % x, want % x", got, want)
	}
}

func TestRemxRejectsInvalidArguments(t *testing.T) {
	cases := []struct {
		line    string
		message string
	}{
		{"remx 1", "remx expects"},
		{"remx 1,0", "count must be 1..49152"},
		{"remx 1,49153", "count must be 1..49152"},
		{"remx 1,5,256", "byte must be 0..255"},
		{"remx 10000,5", "line number must be 0..9999"},
	}
	for _, c := range cases {
		t.Run(c.line, func(t *testing.T) {
			_, err := Tokenize([]string{c.line})
			if err == nil || !strings.Contains(err.Error(), c.message) {
				t.Fatalf("Tokenize(%q): want error containing %q, got %v", c.line, c.message, err)
			}
		})
	}
}

func TestLineNumbersMustStrictlyIncrease(t *testing.T) {
	_, err := Tokenize([]string{"10 PRINT 1", "10 PRINT 2"})
	if err == nil || !strings.Contains(err.Error(), "must increase") {
		t.Fatalf("want a 'must increase' error, got %v", err)
	}
}

func TestBlankAndHashCommentLinesAreSkipped(t *testing.T) {
	program, err := Tokenize([]string{"# a book listing", "", "10 PRINT 1"})
	if err != nil {
		t.Fatal(err)
	}
	if lineBody(program, 10) == nil {
		t.Fatalf("expected line 10 to exist in %x", program)
	}
}

func TestUnrepresentableCharactersAreRejected(t *testing.T) {
	for _, ch := range []string{"!", "%", "&", "'", "@"} {
		t.Run(ch, func(t *testing.T) {
			_, err := Tokenize([]string{"10 REM " + ch})
			// '%' and "'" are handled as escapes elsewhere in
			// tokenizeLine (graphics/inverse-video), not rejected on
			// their own - only bare, un-escaped use inside REM text
			// (which skips every other branch) reaches characterCode
			// and fails for '!'/'&'/"'"/'@'. '%' alone at end of text
			// (no following character) also falls through to
			// characterCode.
			if ch == "!" || ch == "&" || ch == "'" || ch == "@" || ch == "%" {
				if err == nil {
					t.Fatalf("Tokenize(%q): want an error, got none", ch)
				}
			}
		})
	}
}

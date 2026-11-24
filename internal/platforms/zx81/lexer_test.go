package zx81

import (
	"strings"
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// This is the first real test of core.Lexer/Description against ZX81
// (not the fake platform from internal/core/pipeline_test.go) - it
// drives core.Lex() itself, not the full Lex->Parse->Generate
// pipeline, since ZX81's Parser doesn't exist yet (see OUTLINE.md,
// Bloco 3: Description/Lexer land this round, Parser is the next
// one) - same shape as internal/platforms/msx/lexer_test.go's own
// first round.

func lexLines(t *testing.T, lines ...string) []core.Token {
	t.Helper()
	source := make([]core.Line, 0, len(lines)+1)
	source = append(source, core.Line{Number: 0, Text: "PROGRAM", File: "test"})
	for i, l := range lines {
		source = append(source, core.Line{Number: i + 1, Text: l, File: "test.dzx"})
	}
	ctx := &core.CompileContext{
		Source:       source,
		Settings:     &core.CompileSettings{LineStart: 10, LineStep: 10},
		Symbols:      core.NewSymbolTables(),
		IncludeStack: []string{"test.dzx"},
	}
	platform := core.Platform{Description: Description, Lexer: Lexer{}}

	tokens, err := core.Lex(ctx, platform)
	if err != nil {
		t.Fatalf("core.Lex(%v): %v", lines, err)
	}
	return tokens
}

func lexLinesExpectError(t *testing.T, lines ...string) error {
	t.Helper()
	source := make([]core.Line, 0, len(lines)+1)
	source = append(source, core.Line{Number: 0, Text: "PROGRAM", File: "test"})
	for i, l := range lines {
		source = append(source, core.Line{Number: i + 1, Text: l, File: "test.dzx"})
	}
	ctx := &core.CompileContext{
		Source:       source,
		Settings:     &core.CompileSettings{LineStart: 10, LineStep: 10},
		Symbols:      core.NewSymbolTables(),
		IncludeStack: []string{"test.dzx"},
	}
	platform := core.Platform{Description: Description, Lexer: Lexer{}}

	_, err := core.Lex(ctx, platform)
	if err == nil {
		t.Fatalf("core.Lex(%v): expected an error, got none", lines)
	}
	return err
}

// kindsAndValues drops the PROGRAM/NEWLINE/EOF structural tokens core
// adds, keeping only what the ZX81 Lexer (or its identifier/string
// fallback) actually classified - easier to assert against.
func kindsAndValues(tokens []core.Token) [][2]string {
	var out [][2]string
	for _, tok := range tokens {
		switch tok.Kind {
		case core.KindProgram, core.KindNewline, core.KindEOF:
			continue
		}
		out = append(out, [2]string{string(tok.Kind), tok.String()})
	}
	return out
}

// TestLexKeywordsOperatorsAndSymbols: "x+1" (no space) merges into
// "x" + "+1" (ONE signed-number token), NOT "x"/"+"/"1" - confirmed
// against the real badig.py (same quirk as
// TestSignedNumbersOnlyMergeWithAPlainDigitRun below; this is the
// first, most basic case of it).
func TestLexKeywordsOperatorsAndSymbols(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "print x+1"))
	want := [][2]string{
		{string(KindInstruction), "print"},
		{string(core.KindIdentifier), "x"},
		{string(KindNumber), "+1"},
	}
	assertTokens(t, tokens, want)
}

func TestExtendedSymbolsWinOverTheirSinglleCharPrefix(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "x+=1"))
	want := [][2]string{
		{string(core.KindIdentifier), "x"},
		{string(KindSymbol), "+="},
		{string(KindNumber), "1"},
	}
	assertTokens(t, tokens, want)
}

// TestCompoundSymbolsWinOverTheirSingleCharPrefix confirms the 4
// two-character entries inside c_symbol itself ("<>", ">=", "<=",
// "**") are checked before their single-character prefixes - listed
// first in badig_zx81.py's own c_symbol, and this project's symbols
// slice preserves that order deliberately (ver description.go).
func TestCompoundSymbolsWinOverTheirSingleCharPrefix(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "x<>1 x>=1 x<=1 x**2"))
	want := [][2]string{
		{string(core.KindIdentifier), "x"}, {string(KindSymbol), "<>"}, {string(KindNumber), "1"},
		{string(core.KindIdentifier), "x"}, {string(KindSymbol), ">="}, {string(KindNumber), "1"},
		{string(core.KindIdentifier), "x"}, {string(KindSymbol), "<="}, {string(KindNumber), "1"},
		{string(core.KindIdentifier), "x"}, {string(KindSymbol), "**"}, {string(KindNumber), "2"},
	}
	assertTokens(t, tokens, want)
}

func TestHashFollowedByLetterIsRefusedToTheSharedDignifiedFallback(t *testing.T) {
	// "#a" is a KEEP/toggle-rem tag (shared Dignified vocabulary), not
	// ZX81's own "#" symbol - same guard as MSX's lexer.
	tokens := kindsAndValues(lexLines(t, "#a"))
	want := [][2]string{{string(core.KindToggleRem), "#a"}}
	assertTokens(t, tokens, want)
}

func TestHashAloneIsStillZx81sOwnSymbol(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "poke #,1"))
	want := [][2]string{
		{string(KindInstruction), "poke"},
		{string(KindSymbol), "#"},
		{string(KindSymbol), ","},
		{string(KindNumber), "1"},
	}
	assertTokens(t, tokens, want)
}

// TestSignedNumbersOnlyMergeWithAPlainDigitRun documents (and locks
// in) a genuinely surprising, empirically-confirmed real-engine quirk
// (ver lexer.go's own doc comment on the '+'/'-' case): a leading sign
// merges into a following number ONLY when immediately followed by a
// plain digit - never through a decimal point, and never through an
// exponent. Confirmed in this session by running the real badig.py
// with `-lex -prr` against each of these exact lines.
func TestSignedNumbersOnlyMergeWithAPlainDigitRun(t *testing.T) {
	cases := []struct {
		line string
		want [][2]string
	}{
		{"print -5", [][2]string{{string(KindInstruction), "print"}, {string(KindNumber), "-5"}}},
		{"print 5-3", [][2]string{{string(KindInstruction), "print"}, {string(KindNumber), "5"}, {string(KindNumber), "-3"}}},
		{"print y - 5", [][2]string{
			{string(KindInstruction), "print"}, {string(core.KindIdentifier), "y"},
			{string(KindSymbol), "-"}, {string(KindNumber), "5"},
		}},
		{"print y-5", [][2]string{
			{string(KindInstruction), "print"}, {string(core.KindIdentifier), "y"}, {string(KindNumber), "-5"},
		}},
		{"let x=-05", [][2]string{
			{string(KindInstruction), "let"}, {string(core.KindIdentifier), "x"},
			{string(KindSymbol), "="}, {string(KindNumber), "-05"},
		}},
		{"let x=-5.5", [][2]string{
			{string(KindInstruction), "let"}, {string(core.KindIdentifier), "x"},
			{string(KindSymbol), "="}, {string(KindNumber), "-5"}, {string(KindNumber), ".5"},
		}},
		{"print -.5", [][2]string{
			{string(KindInstruction), "print"}, {string(KindSymbol), "-"}, {string(KindNumber), ".5"},
		}},
		{"print +.5", [][2]string{
			{string(KindInstruction), "print"}, {string(KindSymbol), "+"}, {string(KindNumber), ".5"},
		}},
		{"let x=-55", [][2]string{
			{string(KindInstruction), "let"}, {string(core.KindIdentifier), "x"},
			{string(KindSymbol), "="}, {string(KindNumber), "-55"},
		}},
	}
	for _, c := range cases {
		t.Run(c.line, func(t *testing.T) {
			tokens := kindsAndValues(lexLines(t, c.line))
			assertTokens(t, tokens, c.want)
		})
	}
}

func TestUnsignedNumbersKeepTheirFullDecimalAndExponentForm(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{"print .5", ".5"},
		{"print 4E4", "4E4"},
		{"print 4e-4", "4e-4"},
		{"print 4e+4", "4e+4"},
	}
	for _, c := range cases {
		t.Run(c.line, func(t *testing.T) {
			tokens := kindsAndValues(lexLines(t, c.line))
			want := [][2]string{{string(KindInstruction), "print"}, {string(KindNumber), c.want}}
			assertTokens(t, tokens, want)
		})
	}
}

func TestHexByteEscapeIsItsOwnToken(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, `let \{35}oke=6000`))
	want := [][2]string{
		{string(KindInstruction), "let"},
		{string(KindHexByte), `\{35}`},
		{string(core.KindIdentifier), "oke"},
		{string(KindSymbol), "="},
		{string(KindNumber), "6000"},
	}
	assertTokens(t, tokens, want)
}

func TestMalformedHexByteEscapeIsFatal(t *testing.T) {
	for _, line := range []string{`print \{D}`, `print \{GG}`, `print \{D8`} {
		t.Run(line, func(t *testing.T) {
			err := lexLinesExpectError(t, line)
			if !strings.Contains(err.Error(), "hexadecimal escape") {
				t.Fatalf("lex(%q): want a hexadecimal-escape error, got %v", line, err)
			}
		})
	}
}

func TestStringLiteralSupportsBothQuoteEscapeStyles(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{`print "he said ""hi"""`, `"he said ""hi"""`},
		{`print "he said \"hi\""`, `"he said \"hi\""`},
	}
	for _, c := range cases {
		t.Run(c.line, func(t *testing.T) {
			tokens := kindsAndValues(lexLines(t, c.line))
			want := [][2]string{{string(KindInstruction), "print"}, {string(core.KindString), c.want}}
			assertTokens(t, tokens, want)
		})
	}
}

func TestRemClaimsTheRestOfTheLine(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, `rem hello "world" \{41}`))
	want := [][2]string{{string(KindComment), `rem hello "world" \{41}`}}
	assertTokens(t, tokens, want)
}

func TestPrintShorthandQuestionMark(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, `?1`))
	want := [][2]string{{string(KindInstruction), "?"}, {string(KindNumber), "1"}}
	assertTokens(t, tokens, want)
}

func TestJumpKeywordsAndFunctions(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "goto 10", "gosub 20", `print chr$(65)`))
	want := [][2]string{
		{string(KindJump), "goto"}, {string(KindNumber), "10"},
		{string(KindJump), "gosub"}, {string(KindNumber), "20"},
		{string(KindInstruction), "print"}, {string(KindFunction), "chr$"},
		{string(KindSymbol), "("}, {string(KindNumber), "65"}, {string(KindSymbol), ")"},
	}
	assertTokens(t, tokens, want)
}

// TestKeywordPrefixDoesNotSwallowALongerIdentifier mirrors the real
// "copy_VRAM"/"color1" bugs MSX's own lexer fixed (ver msx/lexer.go):
// "told" starts with the reserved word "TO" but must still lex as one
// plain identifier.
func TestKeywordPrefixDoesNotSwallowALongerIdentifier(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "let told=1"))
	want := [][2]string{
		{string(KindInstruction), "let"},
		{string(core.KindIdentifier), "told"},
		{string(KindSymbol), "="},
		{string(KindNumber), "1"},
	}
	assertTokens(t, tokens, want)
}

func assertTokens(t *testing.T, got, want [][2]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d tokens %v, want %d tokens %v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d = %v, want %v (full: got=%v want=%v)", i, got[i], want[i], got, want)
		}
	}
}

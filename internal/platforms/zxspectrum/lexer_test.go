package zxspectrum

import (
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

func lexLines(t *testing.T, lines ...string) []core.Token {
	t.Helper()
	source := make([]core.Line, 0, len(lines)+1)
	source = append(source, core.Line{Number: 0, Text: "PROGRAM", File: "test"})
	for i, l := range lines {
		source = append(source, core.Line{Number: i + 1, Text: l, File: "test.dzs"})
	}
	ctx := &core.CompileContext{
		Source:       source,
		Settings:     &core.CompileSettings{LineStart: 10, LineStep: 10},
		Symbols:      core.NewSymbolTables(),
		IncludeStack: []string{"test.dzs"},
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
		source = append(source, core.Line{Number: i + 1, Text: l, File: "test.dzs"})
	}
	ctx := &core.CompileContext{
		Source:       source,
		Settings:     &core.CompileSettings{LineStart: 10, LineStep: 10},
		Symbols:      core.NewSymbolTables(),
		IncludeStack: []string{"test.dzs"},
	}
	platform := core.Platform{Description: Description, Lexer: Lexer{}}

	_, err := core.Lex(ctx, platform)
	if err == nil {
		t.Fatalf("core.Lex(%v): expected an error, got none", lines)
	}
	return err
}

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

func assertTokens(t *testing.T, got, want [][2]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d tokens %v, want %d tokens %v", len(got), got, len(want), want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("token %d: got %v, want %v (full got %v)", i, got[i], want[i], got)
		}
	}
}

func TestLexKeywordsOperatorsAndSymbols(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, `print x+1`))
	want := [][2]string{
		{string(KindInstruction), "print"},
		{string(core.KindIdentifier), "x"},
		{string(KindSymbol), "+"},
		{string(KindNumber), "1"},
	}
	assertTokens(t, tokens, want)
}

func TestExtendedSymbolsWinOverTheirSingleCharPrefix(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "x+=1"))
	want := [][2]string{
		{string(core.KindIdentifier), "x"},
		{string(KindExtendedSymbol), "+="},
		{string(KindNumber), "1"},
	}
	assertTokens(t, tokens, want)
}

func TestCompoundSymbolsWinOverTheirSingleCharPrefix(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "x<>1 x>=1 x<=1"))
	want := [][2]string{
		{string(core.KindIdentifier), "x"}, {string(KindSymbol), "<>"}, {string(KindNumber), "1"},
		{string(core.KindIdentifier), "x"}, {string(KindSymbol), ">="}, {string(KindNumber), "1"},
		{string(core.KindIdentifier), "x"}, {string(KindSymbol), "<="}, {string(KindNumber), "1"},
	}
	assertTokens(t, tokens, want)
}

func TestApostropheIsAPlainSymbolNotAComment(t *testing.T) {
	// Confirmed empirically (ver tokenizer.go's own doc comment):
	// Sinclair BASIC's PRINT-list newline separator, not a Dignified
	// comment marker.
	tokens := kindsAndValues(lexLines(t, `print "A"'"B"`))
	want := [][2]string{
		{string(KindInstruction), "print"},
		{string(core.KindString), `"A"`},
		{string(KindSymbol), "'"},
		{string(core.KindString), `"B"`},
	}
	assertTokens(t, tokens, want)
}

func TestKeywordDoesNotCutAShorterPrefixOutOfALongerIdentifier(t *testing.T) {
	// "total" must not split into "to"+"tal" - same boundary guard as
	// MSX/ZX81's own lexer.
	tokens := kindsAndValues(lexLines(t, "let total=1"))
	want := [][2]string{
		{string(KindInstruction), "let"},
		{string(core.KindIdentifier), "total"},
		{string(KindSymbol), "="},
		{string(KindNumber), "1"},
	}
	assertTokens(t, tokens, want)
}

func TestMultiWordJumpsAndInstructionsWithVariableSpacing(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "goto 10", "go  to 10", "go sub 20", "def fn f(x)=x", "open#1", "close #1"))
	want := [][2]string{
		{string(KindJump), "goto"}, {string(KindNumber), "10"},
		{string(KindJump), "go  to"}, {string(KindNumber), "10"},
		{string(KindJump), "go sub"}, {string(KindNumber), "20"},
		{string(KindInstruction), "def fn"}, {string(core.KindIdentifier), "f"},
		{string(KindSymbol), "("}, {string(core.KindIdentifier), "x"}, {string(KindSymbol), ")"}, {string(KindSymbol), "="}, {string(core.KindIdentifier), "x"},
		{string(KindInstruction), "open#"}, {string(KindNumber), "1"},
		{string(KindInstruction), "close #"}, {string(KindNumber), "1"},
	}
	assertTokens(t, tokens, want)
}

func TestHashFollowedByLetterIsRefusedToTheSharedDignifiedFallback(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "#a"))
	want := [][2]string{{string(core.KindToggleRem), "#a"}}
	assertTokens(t, tokens, want)
}

func TestHashAloneIsStillSpectrumsOwnSymbol(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "print #1;1"))
	want := [][2]string{
		{string(KindInstruction), "print"},
		{string(KindSymbol), "#"},
		{string(KindNumber), "1"},
		{string(KindSymbol), ";"},
		{string(KindNumber), "1"},
	}
	assertTokens(t, tokens, want)
}

func TestHexEscapeDecodesAndRejectsMalformedInput(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, `\{41}`))
	want := [][2]string{{string(KindHexByte), `\{41}`}}
	assertTokens(t, tokens, want)

	lexLinesExpectError(t, `\{4`)
	lexLinesExpectError(t, `\{GG}`)
}

// TestRemHexEscapeTurnsTheRestOfTheLineIntoLiteralRemText mirrors the
// real behavior confirmed in tokenizer_test.go's
// TestRawBytesAreLiteralInCodeStringsAndRem ("\{EA} 123:FN..." -> the
// ":"/"FN"/quotes after it are never tokenized as code).
func TestRemHexEscapeTurnsTheRestOfTheLineIntoLiteralRemText(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, `\{EA} 123:FN "A"`))
	want := [][2]string{
		{string(KindHexByte), `\{EA}`},
		{string(KindComment), ` 123:FN "A"`},
	}
	assertTokens(t, tokens, want)
}

// TestRemClaimsTheRestOfTheLine confirms REM splits into its OWN
// keyword token (eligible for uppercase-forcing/spacing like any other
// instruction, ver parser.go's Pass5) plus a separate content token
// kept byte-for-byte - NOT one merged blob like MSX/ZX81's own,
// simpler REM handling (ver lexer.go's own doc comment on this case).
func TestRemClaimsTheRestOfTheLine(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, `rem foo: "bar" 'x`))
	want := [][2]string{
		{string(KindInstruction), "rem"},
		{string(KindComment), ` foo: "bar" 'x`},
	}
	assertTokens(t, tokens, want)
}

func TestSpectrumStringsOnlyEscapeQuotesByDoublingAndAreFatalIfUnterminated(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, `print "say ""hi"""`))
	want := [][2]string{
		{string(KindInstruction), "print"},
		{string(core.KindString), `"say ""hi"""`},
	}
	assertTokens(t, tokens, want)

	lexLinesExpectError(t, `print "unterminated`)
}

func TestNumberPatternHasNoLeadingSign(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "let x=-1"))
	want := [][2]string{
		{string(KindInstruction), "let"},
		{string(core.KindIdentifier), "x"},
		{string(KindSymbol), "="},
		{string(KindSymbol), "-"},
		{string(KindNumber), "1"},
	}
	assertTokens(t, tokens, want)
}

func TestBinAndRemxAreRecognizedAsOperatorAndInstruction(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "let x=bin 101"))
	want := [][2]string{
		{string(KindInstruction), "let"},
		{string(core.KindIdentifier), "x"},
		{string(KindSymbol), "="},
		{string(KindOperator), "bin"},
		{string(KindNumber), "101"},
	}
	assertTokens(t, tokens, want)

	tokens = kindsAndValues(lexLines(t, "remx 100,4"))
	want = [][2]string{
		{string(KindInstruction), "remx"},
		{string(KindNumber), "100"},
		{string(KindSymbol), ","},
		{string(KindNumber), "4"},
	}
	assertTokens(t, tokens, want)
}

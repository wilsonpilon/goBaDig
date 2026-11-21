package msx

import (
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// This is the first real test of core.Lexer/Description against an
// actual dialect (not the fake platform from internal/core/
// pipeline_test.go) - it drives core.Lex() itself, not the full
// Lex->Parse->Generate pipeline, since MSX's Parser doesn't exist yet
// (see OUTLINE.md, Bloco 2: Description/Lexer landed this round,
// Parser is the next one).

func lexLines(t *testing.T, lines ...string) []core.Token {
	t.Helper()
	source := make([]core.Line, 0, len(lines)+1)
	source = append(source, core.Line{Number: 0, Text: "PROGRAM", File: "test"})
	for i, l := range lines {
		source = append(source, core.Line{Number: i + 1, Text: l, File: "test.dmx"})
	}
	ctx := &core.CompileContext{
		Source:       source,
		Settings:     &core.CompileSettings{LineStart: 10, LineStep: 10},
		Symbols:      core.NewSymbolTables(),
		IncludeStack: []string{"test.dmx"},
	}
	platform := core.Platform{Description: Description, Lexer: Lexer{}}

	tokens, err := core.Lex(ctx, platform)
	if err != nil {
		t.Fatalf("core.Lex(%v): %v", lines, err)
	}
	return tokens
}

// kindsAndValues drops the PROGRAM/NEWLINE/EOF structural tokens core
// adds, keeping only what the MSX Lexer (or its identifier/string
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

func TestLexKeywordsOperatorsAndSymbols(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "print x+1"))
	want := [][2]string{
		{string(KindInstruction), "print"},
		{string(core.KindIdentifier), "x"},
		{string(KindSymbol), "+"},
		{string(KindNumber), "1"},
	}
	assertTokens(t, tokens, want)
}

func TestLexCaseInsensitiveKeywordMatchesUppercaseAndLowercase(t *testing.T) {
	upper := kindsAndValues(lexLines(t, "PRINT 1"))
	lower := kindsAndValues(lexLines(t, "print 1"))
	assertTokens(t, upper, [][2]string{{string(KindInstruction), "PRINT"}, {string(KindNumber), "1"}})
	assertTokens(t, lower, [][2]string{{string(KindInstruction), "print"}, {string(KindNumber), "1"}})
}

func TestLexCompoundAssignmentIsOneSymbolNotTwo(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "x+=5"))
	want := [][2]string{
		{string(core.KindIdentifier), "x"},
		{string(KindSymbol), "+="},
		{string(KindNumber), "5"},
	}
	assertTokens(t, tokens, want)
}

func TestLexJumpKeywordAndStringLiteral(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, `if x=1 then print "hi"`))
	want := [][2]string{
		{string(KindInstruction), "if"},
		{string(core.KindIdentifier), "x"},
		{string(KindSymbol), "="},
		{string(KindNumber), "1"},
		{string(KindJump), "then"},
		{string(KindInstruction), "print"},
		{string(core.KindString), `"hi"`},
	}
	assertTokens(t, tokens, want)
}

func TestLexNumberFormats(t *testing.T) {
	tokens := kindsAndValues(lexLines(t, "print 3.14, &H1A, &O17, &B101"))
	want := [][2]string{
		{string(KindInstruction), "print"},
		{string(KindNumber), "3.14"},
		{string(KindSymbol), ","},
		{string(KindNumber), "&H1A"},
		{string(KindSymbol), ","},
		{string(KindNumber), "&O17"},
		{string(KindSymbol), ","},
		{string(KindNumber), "&B101"},
	}
	assertTokens(t, tokens, want)
}

func TestLexCommentAndDataKeyword(t *testing.T) {
	remTokens := kindsAndValues(lexLines(t, "rem a comment"))
	assertTokens(t, remTokens, [][2]string{{string(KindComment), "rem a comment"}})

	apostropheTokens := kindsAndValues(lexLines(t, "' a comment"))
	assertTokens(t, apostropheTokens, [][2]string{{string(KindComment), "' a comment"}})

	dataTokens := kindsAndValues(lexLines(t, "data 1,2,3"))
	want := [][2]string{
		{string(KindData), "data"},
		{string(KindNumber), "1"},
		{string(KindSymbol), ","},
		{string(KindNumber), "2"},
		{string(KindSymbol), ","},
		{string(KindNumber), "3"},
	}
	assertTokens(t, dataTokens, want)
}

func TestLexUnrecognizedCharacterFallsBackTolerantly(t *testing.T) {
	// "`" isn't an MSX symbol/keyword, nor a shared Dignified symbol
	// (~ [ ] { } @), nor an identifier-start or quote - it must still
	// produce a token (KindUnknown) instead of aborting lexing,
	// mirroring the original engine's tolerant "character not
	// recognized" handling.
	tokens := kindsAndValues(lexLines(t, "print `"))
	want := [][2]string{
		{string(KindInstruction), "print"},
		{string(core.KindUnknown), "`"},
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
			t.Fatalf("token %d = %v, want %v\nfull got: %v\nfull want: %v", i, got[i], want[i], got, want)
		}
	}
}

package core

import (
	"strconv"
	"strings"
	"testing"
)

// This file implements a minimal fake platform - NOT MSX, NOT any
// real dialect - that only recognizes PRINT, integers, and the word
// "IGNORE". Its only job is to prove the Description/Lexer/Parser/
// Platform contract in contract.go/platform.go is actually usable end
// to end, driving a real token through Lex -> Parse -> Generate,
// before any real platform (Bloco 2) is asked to depend on it. Ver
// OUTLINE.md, Bloco 1, Fase 3.

const (
	fakeKindPrint  TokenKind = "FAKE_PRINT"
	fakeKindNumber TokenKind = "FAKE_NUMBER"
)

type fakeDescription struct{}

func (fakeDescription) SysInfo() SystemInfo {
	return SystemInfo{Name: "Fake", DignifiedExt: ".fake", ClassicExt: ".fakec", BinaryExt: ".fakeb"}
}

func (fakeDescription) QuoteChar() byte { return '"' }

func (fakeDescription) StatementSeparator() string { return ":" }

func (fakeDescription) LoopBackInstruction() string { return "goto" }

type fakeLexer struct{}

func (fakeLexer) Init(*CompileContext) {}

// Lex only recognizes three things, by prefix on the remaining line
// text (the v1 Lexer contract, validated for real against MSX in a
// later round - see OUTLINE.md, Bloco 2): "PRINT", "IGNORE" (consumed
// and discarded - zero tokens), and a run of digits.
func (fakeLexer) Lex(_ *CompileContext, remaining string) (int, []Token, bool, error) {
	upper := strings.ToUpper(remaining)
	switch {
	case strings.HasPrefix(upper, "PRINT"):
		return len("PRINT"), []Token{{Kind: fakeKindPrint, Value: remaining[:len("PRINT")]}}, true, nil
	case strings.HasPrefix(upper, "IGNORE"):
		return len("IGNORE"), nil, true, nil
	}
	digits := 0
	for digits < len(remaining) && remaining[digits] >= '0' && remaining[digits] <= '9' {
		digits++
	}
	if digits > 0 {
		n, err := strconv.Atoi(remaining[:digits])
		if err != nil {
			return 0, nil, false, nil
		}
		return digits, []Token{{Kind: fakeKindNumber, Value: n}}, true, nil
	}
	// Not recognized: let the driver fall back to its own generic
	// handling (identifier/string/raw character).
	return 0, nil, false, nil
}

// fakeParser counts how many times Pass4/Pass5 ran, so a test can
// confirm Parse() really does gate them to the main file only. Every
// other pass just lets the driver keep each token unchanged (Continue).
type fakeParser struct {
	pass4Calls *int
	pass5Calls *int
}

func (fakeParser) Pass1(_ *CompileContext, tokens []Token, index int) (HookAction, []Token, error) {
	return Continue, nil, nil
}
func (fakeParser) Pass2(_ *CompileContext, tokens []Token, index int) (HookAction, []Token, error) {
	return Continue, nil, nil
}
func (fakeParser) Pass3(_ *CompileContext, tokens []Token, index int) (HookAction, []Token, error) {
	return Continue, nil, nil
}
func (p fakeParser) Pass4(_ *CompileContext, tokens []Token, index int) (HookAction, []Token, error) {
	if p.pass4Calls != nil {
		*p.pass4Calls++
	}
	return Continue, nil, nil
}
func (p fakeParser) Pass5(_ *CompileContext, tokens []Token, index int) (HookAction, []Token, error) {
	if p.pass5Calls != nil {
		*p.pass5Calls++
	}
	return Continue, nil, nil
}
func (fakeParser) Generate(_ *CompileContext, tokens []Token, index int) (HookAction, []Token, error) {
	return Continue, nil, nil
}

type fakeSettings struct{}

func (fakeSettings) Apply(*CompileSettings) error { return nil }

func newFakeContext(sourceLine string, includeStack []string) *CompileContext {
	return &CompileContext{
		Source: []Line{
			{Number: 0, Text: "PROGRAM", File: "fake"},
			{Number: 1, Text: sourceLine, File: "fake.dmx"},
		},
		Settings:     &CompileSettings{LineStart: 10, LineStep: 10},
		Symbols:      NewSymbolTables(),
		IncludeStack: includeStack,
	}
}

// newFakeMultilineContext is newFakeContext but for several source
// lines - needed to test label resolution/line numbering, which only
// show up across multiple lines.
func newFakeMultilineContext(lines []string, includeStack []string) *CompileContext {
	source := make([]Line, 0, len(lines)+1)
	source = append(source, Line{Number: 0, Text: "PROGRAM", File: "fake"})
	for i, l := range lines {
		source = append(source, Line{Number: i + 1, Text: l, File: "fake.dmx"})
	}
	return &CompileContext{
		Source:       source,
		Settings:     &CompileSettings{LineStart: 10, LineStep: 10},
		Symbols:      NewSymbolTables(),
		IncludeStack: includeStack,
	}
}

func runFake(t *testing.T, lines ...string) (GenerateResult, error) {
	t.Helper()
	ctx := newFakeMultilineContext(lines, []string{"fake.dmx"})
	platform := Platform{
		Description: fakeDescription{},
		Lexer:       fakeLexer{},
		Parser:      fakeParser{},
		Settings:    fakeSettings{},
	}
	return Run(ctx, platform)
}

func TestPass4ResolvesLabelToTheLineItAttachesTo(t *testing.T) {
	result, err := runFake(t, "goto {dest}", "{dest}", "PRINT 1")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 goto 20", "20 PRINT 1"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (label attaches to the line that follows it, wasting none)", result.Code, want)
	}
}

func TestPass4ResolvesAtSignToTheCurrentLine(t *testing.T) {
	result, err := runFake(t, "PRINT 1", "goto {@}")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1", "20 goto 20"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v ({@} resolves to its own line)", result.Code, want)
	}
}

func TestPass3CollapsesBlankLinesWithoutWastingLineNumbers(t *testing.T) {
	result, err := runFake(t, "PRINT 1", "", "", "PRINT 2")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1", "20 PRINT 2"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (blank lines must not consume line numbers)", result.Code, want)
	}
}

func TestPass3MergesLineContinuationIntoOneNumberedLine(t *testing.T) {
	result, err := runFake(t, "PRINT 1 _", "PRINT 2")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1 PRINT 2"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (\"_\" at end of line merges into one classic line)", result.Code, want)
	}
}

func TestPass4FailsOnAJumpToAnUndeclaredLabel(t *testing.T) {
	_, err := runFake(t, "goto {missing}")
	if err == nil {
		t.Fatal("expected a fatal error jumping to an undeclared label, got none")
	}
}

func TestPass2ClosesALoopLabelWithAnImplicitGoto(t *testing.T) {
	result, err := runFake(t, "keys{", "PRINT 1", "}")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1", "20 goto 10"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (opener attaches to the next real line; closer jumps back to it)", result.Code, want)
	}
}

func TestPass2InsertsASeparatorWhenTheCloserSharesItsLine(t *testing.T) {
	result, err := runFake(t, "keys{", "PRINT 1}")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1 : goto 10"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (\"}\" not alone on its line needs a separator before the goto)", result.Code, want)
	}
}

func TestPass2ClosesNestedLoopLabelsInnermostFirst(t *testing.T) {
	result, err := runFake(t, "outer{", "inner{", "PRINT 1", "}", "}")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1", "20 goto 10", "30 goto 10"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (both closers must resolve to the same shared opener line)", result.Code, want)
	}
}

func TestPass4ResolvesExitToTheLineAfterTheLoopCloses(t *testing.T) {
	result, err := runFake(t, "keys{", "exit", "PRINT 1", "}", "PRINT 2", "PRINT 3")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 goto 40", "20 PRINT 1", "30 goto 10", "40 PRINT 2", "50 PRINT 3"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (exit jumps past the loop's closing goto)", result.Code, want)
	}
}

func TestPass4FailsWhenExitHasNowhereRealToLand(t *testing.T) {
	_, err := runFake(t, "keys{", "exit", "PRINT 1", "}", "PRINT 2")
	if err == nil {
		t.Fatal("expected a fatal error - the loop is the last thing before end of program, exit has nowhere to land")
	}
}

func TestPass2FailsOnACloserWithoutAMatchingOpener(t *testing.T) {
	_, err := runFake(t, "}")
	if err == nil {
		t.Fatal("expected a fatal error closing a loop label that was never opened, got none")
	}
}

func TestPass2FailsOnAnUnclosedLoopLabel(t *testing.T) {
	_, err := runFake(t, "keys{", "PRINT 1")
	if err == nil {
		t.Fatal("expected a fatal error for a loop label never closed, got none")
	}
}

func TestPass2FailsOnExitOutsideAnyLoopLabel(t *testing.T) {
	_, err := runFake(t, "exit")
	if err == nil {
		t.Fatal("expected a fatal error for exit outside of any loop label, got none")
	}
}

// TestLexIdentifierPatternIsCaseInsensitive guards against a real bug
// found while researching DEFINE: badig_dignified.py's d_idnttp looks
// lowercase-only but is compiled inside the original engine's single
// combined regex under re.I (badig.py:127-132), so it actually
// matches any case - real Dignified source routinely uses mixed-case
// names ("gameName$", "fill_VRAM", confirmed in
// badig/examples/msx/CGK173.dmx). A strictly-lowercase pattern here
// would cut "gameName" at the "N" instead of keeping it as one token.
func TestLexIdentifierPatternIsCaseInsensitive(t *testing.T) {
	result, err := runFake(t, "PRINT gameName")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT gameName"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (\"gameName\" must survive as one identifier)", result.Code, want)
	}
}

func TestPipelineEndToEnd(t *testing.T) {
	ctx := newFakeContext("PRINT 1 IGNORE foo", []string{"fake.dmx"})
	platform := Platform{
		Description: fakeDescription{},
		Lexer:       fakeLexer{},
		Parser:      fakeParser{},
		Settings:    fakeSettings{},
	}

	result, err := Run(ctx, platform)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Code) != 1 {
		t.Fatalf("expected 1 generated line, got %d: %v", len(result.Code), result.Code)
	}
	// "IGNORE" (recognized by fakeLexer, zero tokens) is gone; "foo"
	// (not recognized by fakeLexer, falls back to core's generic
	// Dignified-identifier matching) survives as one identifier token;
	// PRINT/1 (recognized by fakeLexer) survive reclassified; "10 " is
	// core Pass 4's line numbering (added this round - ver OUTLINE.md,
	// "labels e numeração").
	const want = "10 PRINT 1 foo"
	if result.Code[0] != want {
		t.Fatalf("Code[0] = %q, want %q", result.Code[0], want)
	}
}

func TestParseSkipsPass4And5ForIncludedFiles(t *testing.T) {
	var pass4Calls, pass5Calls int
	platform := Platform{
		Parser: fakeParser{pass4Calls: &pass4Calls, pass5Calls: &pass5Calls},
	}
	tokens := []Token{{Kind: KindEOF}}

	// IncludeStack length 2: this is a file pulled in via INCLUDE, not
	// the main file - Pass4/Pass5 must not run.
	ctx := newFakeContext("", []string{"main.dmx", "included.dmx"})
	if _, err := Parse(ctx, tokens, platform); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pass4Calls != 0 || pass5Calls != 0 {
		t.Fatalf("included file ran Pass4/Pass5: pass4Calls=%d pass5Calls=%d, want 0/0", pass4Calls, pass5Calls)
	}

	// IncludeStack length 1: this is the main file - Pass4/Pass5 must run.
	ctx = newFakeContext("", []string{"main.dmx"})
	if _, err := Parse(ctx, tokens, platform); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pass4Calls == 0 || pass5Calls == 0 {
		t.Fatalf("main file skipped Pass4/Pass5: pass4Calls=%d pass5Calls=%d, want >0/>0", pass4Calls, pass5Calls)
	}
}

func TestVariableAllocatorDescendsAndExhausts(t *testing.T) {
	category := VariableCategory{Name: "scalar", AliasChars: "ab", AliasLen: 1}
	allocator := NewDescendingAllocator()

	first, err := allocator.Allocate(category, "counter", nil)
	if err != nil {
		t.Fatalf("Allocate(counter): %v", err)
	}
	if first != "b" {
		t.Fatalf("first alias = %q, want %q (descending order)", first, "b")
	}

	second, err := allocator.Allocate(category, "total", nil)
	if err != nil {
		t.Fatalf("Allocate(total): %v", err)
	}
	if second != "a" {
		t.Fatalf("second alias = %q, want %q", second, "a")
	}

	// Re-allocating an already-known long name must return the same alias.
	again, err := allocator.Allocate(category, "counter", nil)
	if err != nil || again != first {
		t.Fatalf("Allocate(counter) again = %q, %v; want %q, nil", again, err, first)
	}

	if _, err := allocator.Allocate(category, "overflow", nil); err == nil {
		t.Fatal("Allocate did not report exhaustion once the category's alias space was used up")
	} else if !strings.Contains(err.Error(), ErrAliasesExhausted.Error()) {
		t.Fatalf("Allocate error = %v, want it to wrap ErrAliasesExhausted", err)
	}
}

func TestDiagnosticErrorFormatsPointerUnderColumn(t *testing.T) {
	source := []Line{{Number: 1, Text: "  LET A=1", File: "fake.dmx"}}
	pos := &Position{Source: source, Line: 0, Column: 8, Offset: 2}
	d := Diagnostic{Severity: SeverityError, Message: "unexpected token", Pos: pos}

	got := d.Error()
	const want = "unexpected token\n    LET A=1\n    -----^"
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

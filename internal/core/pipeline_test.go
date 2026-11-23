package core

import (
	"fmt"
	"os"
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

func (fakeDescription) CommentChar() string { return "'" }

func (fakeDescription) AssignmentOperator() string { return "=" }

func (fakeDescription) FunctionCallInstruction() string { return "gosub" }

func (fakeDescription) FunctionReturnInstruction() string { return "return" }

func (fakeDescription) FunctionCallBoundaryWords() []string { return []string{"PRINT"} }

func (fakeDescription) VariableAliasLength() int { return 2 }

// ReservedKeywords includes the fake's own declared instruction words
// (LoopBackInstruction/FunctionCallInstruction/
// FunctionReturnInstruction below) alongside PRINT/FOO - any real
// dialect's "goto"/"gosub"/"return" equivalents would genuinely be
// reserved keywords too, so leaving them out here would just be an
// incomplete fake, not a more honest one.
func (fakeDescription) ReservedKeywords() []string {
	return []string{"PRINT", "FOO", "GOTO", "GOSUB", "RETURN"}
}

func (fakeDescription) CompoundSymbols() []string { return []string{"(", ")", ",", ":", "_"} }

type fakeLexer struct{}

func (fakeLexer) Init(*CompileContext) {}

// Lex only recognizes three things, by prefix on the remaining line
// text (the v1 Lexer contract, validated for real against MSX in a
// later round - see OUTLINE.md, Bloco 2): "PRINT", "IGNORE" (consumed
// and discarded - zero tokens), and a run of digits.
func (fakeLexer) Lex(_ *CompileContext, remaining string, _ bool) (int, []Token, bool, error) {
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
func (fakeParser) BuiltinDefines() map[string]BuiltinDefine { return nil }

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
	want := []string{"10 PRINT 1:goto 10"}
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

func TestPass4InsertsRemHeaderWhenEnabled(t *testing.T) {
	ctx := newFakeMultilineContext([]string{"PRINT 1"}, []string{"fake.dmx"})
	ctx.Settings.RemHeader = true
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
	want := []string{
		"10 'Converted with Basic Dignified",
		"20 'https://github.com/farique1/basic-dignified",
		"30 PRINT 1",
	}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (RemHeader=true must insert the 2-line REM header before any real content, numbered normally)", result.Code, want)
	}
}

func TestPass4OmitsRemHeaderByDefault(t *testing.T) {
	result, err := runFake(t, "PRINT 1")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (RemHeader's Go zero-value is false - no header unless explicitly opted in)", result.Code, want)
	}
}

func TestExpandTabsAdvancesToTheNextTabStop(t *testing.T) {
	cases := []struct {
		in   string
		size int
		want string
	}{
		{"a\tb", 4, "a   b"},          // 1 char then tab -> 3 spaces to reach column 4
		{"\tb", 4, "    b"},           // tab at column 0 -> 4 spaces
		{"ab\tcd\te", 4, "ab  cd  e"}, // two tabs, each to the next 4-column stop
		{"no tabs here", 4, "no tabs here"},
		{"a\tb", 0, "a\tb"}, // tabSize<=0: left unexpanded, documented simplification
	}
	for _, c := range cases {
		if got := expandTabs(c.in, c.size); got != c.want {
			t.Errorf("expandTabs(%q, %d) = %q, want %q", c.in, c.size, got, c.want)
		}
	}
}

func TestLoadLinesExpandsTabsInARealFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/tabs.dmx"
	if err := os.WriteFile(path, []byte("print 1\n\tprint 2\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	lines, err := LoadLines(path, 4)
	if err != nil {
		t.Fatalf("LoadLines: %v", err)
	}
	if len(lines) < 3 || lines[2].Text != "    print 2" {
		t.Fatalf("lines = %#v, want line 2's text to be \"    print 2\" (tab expanded to 4 spaces)", lines)
	}
}

func TestLexDignifiedLineCommentWorksAnywhereOnALine(t *testing.T) {
	result, err := runFake(t,
		"PRINT 1 ## trailing comment, same line as real code",
		"## a whole line that's just a comment",
		"PRINT 2",
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1", "20 PRINT 2"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (\"##\" must consume to end of line with zero tokens, wherever it appears)", result.Code, want)
	}
}

func TestLexDignifiedBlockCommentDiscardsEverythingInside(t *testing.T) {
	result, err := runFake(t,
		"PRINT 1",
		"###",
		"this line is gone",
		"",
		"so is this blank one",
		"###",
		"PRINT 2",
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1", "20 PRINT 2"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (a \"###\" block, including every NEWLINE inside it, must vanish without wasting a line number)", result.Code, want)
	}
}

func TestLexFailsOnAnUnclosedDignifiedBlockComment(t *testing.T) {
	_, err := runFake(t, "PRINT 1", "###", "never closes")
	if err == nil {
		t.Fatal("expected a fatal error for a \"###\" block never closed, got none")
	}
}

func TestLexClassicCommentBlockExpandsIntoSeparateNumberedLines(t *testing.T) {
	result, err := runFake(t,
		"PRINT 1",
		"''",
		"hello",
		"",
		"world",
		"''",
		"PRINT 2",
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1", "20 'hello", "30 '", "40 'world", "50 PRINT 2"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (each line inside a \"''\" block, even a blank one, becomes its own real classic comment line)", result.Code, want)
	}
}

func TestLexFailsOnAnUnclosedClassicCommentBlock(t *testing.T) {
	_, err := runFake(t, "PRINT 1", "''", "never closes")
	if err == nil {
		t.Fatal("expected a fatal error for a \"''\" block never closed, got none")
	}
}

func TestPass3CollapsesTwoAdjacentSeparatorsIntoOne(t *testing.T) {
	result, err := runFake(t, "PRINT 1::PRINT 2")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1:PRINT 2"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (two adjacent \":\" must collapse into one)", result.Code, want)
	}
}

func TestPass3MergesLinesAcrossATrailingSeparatorKeepingIt(t *testing.T) {
	result, err := runFake(t, "PRINT 1:", "PRINT 2")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1:PRINT 2"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (unlike \"_\", a trailing \":\" merges the next line in WITHOUT disappearing itself)", result.Code, want)
	}
}

func TestPass3MergesALineStartingWithASeparatorIntoThePreviousOne(t *testing.T) {
	result, err := runFake(t, "PRINT 1", ":PRINT 2")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1:PRINT 2"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (a line starting with \":\" merges into the previous line)", result.Code, want)
	}
}

func TestPass3JoinsTwoAdjacentStringLiteralsOnTheSameLine(t *testing.T) {
	result, err := runFake(t, `PRINT "a""b"`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{`10 PRINT "ab"`}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (two adjacent string literals with nothing between them must merge into one)", result.Code, want)
	}
}

func TestPass3JoinsTwoAdjacentStringLiteralsAcrossANewline(t *testing.T) {
	result, err := runFake(t, `PRINT "a"`, `"b"`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{`10 PRINT "ab"`}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (the NEWLINE between two lines that are each just a string literal must vanish too, merging them into one)", result.Code, want)
	}
}

func TestPass1ExpandsASimpleDefineWithNoPlaceholder(t *testing.T) {
	result, err := runFake(t, "define [a][PRINT 1]", "[a]")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (the define's own line must vanish entirely)", result.Code, want)
	}
}

func TestPass1ExpandsADefineWithPlaceholderUsingExplicitArgOrDefault(t *testing.T) {
	result, err := runFake(t, "define [v][PRINT [0]]", "[v](5)", "[v]")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 5", "20 PRINT 0"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v ([v](5) substitutes 5; bare [v] falls back to the [0] default)", result.Code, want)
	}
}

func TestPass1ReadsSeveralDefinesOnOneLine(t *testing.T) {
	result, err := runFake(t, "define [a][PRINT 1], [b][PRINT 2]", "[a]", "[b]")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1", "20 PRINT 2"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (comma-separated defines on one DEFINE line)", result.Code, want)
	}
}

func TestPass1ExpandsANestedDefineCallAsAnArgument(t *testing.T) {
	result, err := runFake(t, "define [a][PRINT [0]]", "define [b][9]", "[a]([b])")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 9"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v ([b] expands first, its result becomes [a]'s argument)", result.Code, want)
	}
}

func TestPass1FailsCallingAnUndefinedName(t *testing.T) {
	_, err := runFake(t, "[missing]")
	if err == nil {
		t.Fatal("expected a fatal error calling an undefined name, got none")
	}
}

func TestPass1FailsOnADuplicateDefineName(t *testing.T) {
	_, err := runFake(t, "define [a][PRINT 1], [a][PRINT 2]")
	if err == nil {
		t.Fatal("expected a fatal error for a duplicate define name, got none")
	}
}

func TestPass1FailsWhenDefineIsNotAtTheStartOfALine(t *testing.T) {
	_, err := runFake(t, "PRINT 1 define [a][PRINT 2]")
	if err == nil {
		t.Fatal("expected a fatal error for DEFINE not at the start of a line, got none")
	}
}

func TestPass2ExpandsAFunctionCallWithAnExplicitArgument(t *testing.T) {
	result, err := runFake(t, "func .f(x=x)", "PRINT x", "ret y", ".f(5)")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// "x=5:gosub <f's line>" - x differs from the default (x itself),
	// so an explicit assignment is emitted before the gosub. Each
	// line is prefixed with its own number, so this checks Contains,
	// not HasPrefix. No spaces around "=" or ":" - neither is a
	// reserved keyword, only "gosub" gets padding (real spacing
	// rules, ver Generate's own doc comment).
	want := "x=5:gosub "
	if len(result.Code) == 0 || !strings.Contains(result.Code[len(result.Code)-1], want) {
		t.Fatalf("Code = %v, want last line to contain %q", result.Code, want)
	}
}

func TestPass2ExpandsAFunctionCallCapturingReturnValues(t *testing.T) {
	result, err := runFake(t, "func .f(x=x)", "ret y", "a = .f(1)")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// "x=1:gosub <line>:a=y" - the call captured "y" (the function's
	// internal return variable) into the caller's "a". No spaces
	// around "=" or ":" - real spacing rules only pad reserved
	// keywords (ver Generate's own doc comment).
	want := "x=1:gosub "
	last := result.Code[len(result.Code)-1]
	if !strings.Contains(last, want) || !strings.HasSuffix(last, ":a=y") {
		t.Fatalf("Code = %v, want last line containing %q and ending in %q", result.Code, want, ":a=y")
	}
}

func TestPass2SkipsARedundantSelfAssignmentInAFunctionCall(t *testing.T) {
	result, err := runFake(t, "func .f(x=x)", "ret", ".f(x)")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Calling .f(x) with the SAME text as the parameter's own name -
	// compare_func_args finds them equal, so no "x=x" assignment is
	// emitted at all, straight to "gosub".
	last := result.Code[len(result.Code)-1]
	if strings.Contains(last, "=") {
		t.Fatalf("Code = %v, want no redundant self-assignment in %q", result.Code, last)
	}
	if !strings.Contains(last, "gosub ") {
		t.Fatalf("Code = %v, want %q to contain \"gosub \"", result.Code, last)
	}
}

func TestPass2LeavesAnOmittedTrailingParameterUntouched(t *testing.T) {
	result, err := runFake(t, "func .f(x=x, y=y)", "ret", ".f(9)")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Calling .f(9) gives only ONE value for a 2-parameter function -
	// "y" is never reached at all (not even assigned its own default)
	// since the call has no second comma-separated slot for it -
	// mirrors replace_func_calls' zip() stopping at the shorter list.
	last := result.Code[len(result.Code)-1]
	if strings.Contains(last, "y") {
		t.Fatalf("Code = %v, want the omitted trailing parameter \"y\" untouched in %q", result.Code, last)
	}
}

func TestPass1AllowsRetWithNoReturnValues(t *testing.T) {
	_, err := runFake(t, "func .f()", "ret", ".f()")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestPass2FailsCallingAnUndefinedFunction(t *testing.T) {
	_, err := runFake(t, ".missing()")
	if err == nil {
		t.Fatal("expected a fatal error calling an undefined function, got none")
	}
}

func TestPass1FailsOnRetWithoutFunc(t *testing.T) {
	_, err := runFake(t, "ret")
	if err == nil {
		t.Fatal("expected a fatal error for ret without func, got none")
	}
}

func TestPass1FailsOnANestedFunc(t *testing.T) {
	_, err := runFake(t, "func .a()", "func .b()", "ret", "ret")
	if err == nil {
		t.Fatal("expected a fatal error for a func nested inside another, got none")
	}
}

func TestPass1FailsWhenFuncIsNotAtTheStartOfALine(t *testing.T) {
	_, err := runFake(t, "PRINT 1 func .f()")
	if err == nil {
		t.Fatal("expected a fatal error for FUNC not at the start of a line, got none")
	}
}

// runFakeWithSymbols is runFake but also returns the CompileContext,
// so a test can inspect ctx.Symbols afterward - DECLARE writes
// straight into those 3 tables (Declares/HardShortVars/
// HardLongVars), and the fake platform has no real variable-renaming
// Pass 3/4 of its own to observe the effect through rendered output
// the way the MSX end-to-end tests do.
func runFakeWithSymbols(t *testing.T, lines ...string) (*CompileContext, error) {
	t.Helper()
	ctx := newFakeMultilineContext(lines, []string{"fake.dmx"})
	platform := Platform{
		Description: fakeDescription{},
		Lexer:       fakeLexer{},
		Parser:      fakeParser{},
		Settings:    fakeSettings{},
	}
	_, err := Run(ctx, platform)
	return ctx, err
}

func TestPass1DeclareAssignsAnExplicitAlias(t *testing.T) {
	ctx, err := runFakeWithSymbols(t, "declare longname:zz")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := ctx.Symbols.Declares["longname@fake.dmx"]; got != "zz" {
		t.Fatalf("Declares[longname@fake.dmx] = %q, want %q", got, "zz")
	}
}

func TestPass1DeclareBareKeepsALongNameVerbatim(t *testing.T) {
	ctx, err := runFakeWithSymbols(t, "declare longname")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !ctx.Symbols.HardLongVars["longname@fake.dmx"] {
		t.Fatal("\"longname\" was not registered in HardLongVars (same effect as ~longname)")
	}
}

func TestPass1DeclareBareReservesAnAlreadyShortName(t *testing.T) {
	ctx, err := runFakeWithSymbols(t, "declare ab")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !ctx.Symbols.HardShortVars["ab"] {
		t.Fatal("\"ab\" was not registered in HardShortVars")
	}
}

func TestPass1ReadsSeveralDeclaresOnOneLine(t *testing.T) {
	ctx, err := runFakeWithSymbols(t, "declare a1:aa, b2:bb")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if ctx.Symbols.Declares["a1@fake.dmx"] != "aa" || ctx.Symbols.Declares["b2@fake.dmx"] != "bb" {
		t.Fatalf("Declares = %v, want a1->aa and b2->bb", ctx.Symbols.Declares)
	}
}

func TestPass1FailsWhenDeclareIsNotAtTheStartOfALine(t *testing.T) {
	_, err := runFake(t, "PRINT 1 declare a1:aa")
	if err == nil {
		t.Fatal("expected a fatal error for DECLARE not at the start of a line, got none")
	}
}

func TestPass1FailsDeclaringAReservedKeywordAsAVariable(t *testing.T) {
	_, err := runFake(t, "declare foo:aa")
	if err == nil {
		t.Fatal("expected a fatal error declaring a reserved keyword as a variable, got none")
	}
}

func TestPass1FailsDeclaringATooShortVariableName(t *testing.T) {
	_, err := runFake(t, "declare a:aa")
	if err == nil {
		t.Fatal("expected a fatal error for a 1-character declared variable, got none")
	}
}

func TestPass1FailsOnADuplicateDeclaredAlias(t *testing.T) {
	_, err := runFake(t, "declare a1:xx, b2:xx")
	if err == nil {
		t.Fatal("expected a fatal error reusing the same alias for two different variables, got none")
	}
}

func TestPass1StripsAToggleBlockByDefaultWithNoKeepDeclared(t *testing.T) {
	result, err := runFake(t, "PRINT 2", "#a", "PRINT 1", "#a", "PRINT 3")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 2", "20 PRINT 3"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (the whole #a block must vanish, wasting no line number)", result.Code, want)
	}
}

func TestPass1KeepsAToggleBlockWhenItsOwnTagIsKept(t *testing.T) {
	result, err := runFake(t, "keep #a", "PRINT 2", "#a", "PRINT 1", "#a", "PRINT 3")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 2", "20 PRINT 1", "30 PRINT 3"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (KEEP #a must keep the block's content)", result.Code, want)
	}
}

func TestPass1KeepAllKeepsAnyTag(t *testing.T) {
	result, err := runFake(t, "keep #all", "PRINT 2", "#x", "PRINT 1", "#x", "PRINT 3")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 2", "20 PRINT 1", "30 PRINT 3"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (#ALL must keep any tag's block)", result.Code, want)
	}
}

func TestPass1KeepNoneOverridesAnyOtherKeptTag(t *testing.T) {
	result, err := runFake(t, "keep #a #none", "PRINT 2", "#a", "PRINT 1", "#a", "PRINT 3")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 2", "20 PRINT 3"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (#NONE must override #a being kept too)", result.Code, want)
	}
}

func TestPass1StripsOnlyTheSingleLineForAOneLineToggle(t *testing.T) {
	result, err := runFake(t, "PRINT 2", "#a PRINT 999", "PRINT 3")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 2", "20 PRINT 3"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (only the one tagged line should vanish)", result.Code, want)
	}
}

func TestPass1FailsOnAnUnclosedToggleBlock(t *testing.T) {
	_, err := runFake(t, "#a", "PRINT 1")
	if err == nil {
		t.Fatal("expected a fatal error for a toggle block never closed, got none")
	}
}

func TestPass1FailsOnAnInvalidKeepTag(t *testing.T) {
	_, err := runFake(t, "keep somevar")
	if err == nil {
		t.Fatal("expected a fatal error for an invalid keep tag, got none")
	}
}

func TestPass1FailsOnADuplicateKeepTag(t *testing.T) {
	_, err := runFake(t, "keep #a #a")
	if err == nil {
		t.Fatal("expected a fatal error for a duplicate keep tag, got none")
	}
}

// fakeLoader is a CompileContext.LoadFile double that serves Lines
// from an in-memory map instead of touching disk - lines map keys are
// paths exactly as resolveInclude would ask for them; requested
// records every path actually requested, for tests asserting on
// resolution itself (Achado 2 - relative paths resolve against the
// MAIN file's own directory, not the including file's).
func fakeLoader(lines map[string][]string, requested *[]string) func(string) ([]Line, error) {
	return func(path string) ([]Line, error) {
		if requested != nil {
			*requested = append(*requested, path)
		}
		content, ok := lines[path]
		if !ok {
			return nil, fmt.Errorf("fakeLoader: no file registered for %q", path)
		}
		source := make([]Line, 0, len(content)+1)
		source = append(source, Line{Number: 0, Text: "PROGRAM", File: path})
		for i, text := range content {
			source = append(source, Line{Number: i + 1, Text: text, File: path})
		}
		return source, nil
	}
}

func runFakeWithInclude(t *testing.T, mainLines []string, included map[string][]string, requested *[]string) (*CompileContext, GenerateResult, error) {
	t.Helper()
	ctx := newFakeMultilineContext(mainLines, []string{"fake.dmx"})
	ctx.LoadFile = fakeLoader(included, requested)
	platform := Platform{
		Description: fakeDescription{},
		Lexer:       fakeLexer{},
		Parser:      fakeParser{},
		Settings:    fakeSettings{},
	}
	result, err := Run(ctx, platform)
	return ctx, result, err
}

func TestPass3IncludeSplicesContentBetweenSurroundingLines(t *testing.T) {
	_, result, err := runFakeWithInclude(t,
		[]string{"PRINT 1", `include "sub.fake"`, "PRINT 3"},
		map[string][]string{"sub.fake": {"PRINT 2"}},
		nil,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1", "20 PRINT 2", "30 PRINT 3"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (included content must splice in with no doubled blank line)", result.Code, want)
	}
}

func TestPass3IncludeSharesSymbolTablesByPointer(t *testing.T) {
	ctx, result, err := runFakeWithInclude(t,
		[]string{"PRINT 0", `include "vars.fake"`, "PRINT 1"},
		map[string][]string{"vars.fake": {"declare longname:zz"}},
		nil,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 0", "20 PRINT 1"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (DECLARE's own line vanishes same as always)", result.Code, want)
	}
	if got := ctx.Symbols.Declares["longname@vars.fake"]; got != "zz" {
		t.Fatalf("ctx.Symbols.Declares[%q] = %q, want %q (an included file's DECLARE must be visible on the including file's own Symbols - shared pointer, Achado 3)", "longname@vars.fake", got, "zz")
	}
}

func TestPass3IncludeResolvesRelativePathsAgainstTheMainFileNotTheIncludingFile(t *testing.T) {
	var requested []string
	_, result, err := runFakeWithInclude(t,
		[]string{`include "mid.fake"`, "PRINT 9"},
		map[string][]string{
			"mid.fake":  {`include "leaf.fake"`},
			"leaf.fake": {"PRINT 1"},
		},
		&requested,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT 1", "20 PRINT 9"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v", result.Code, want)
	}
	wantRequested := []string{"mid.fake", "leaf.fake"}
	if strings.Join(requested, "|") != strings.Join(wantRequested, "|") {
		t.Fatalf("requested paths = %v, want %v (bare \"leaf.fake\" inside mid.fake must resolve against the MAIN file's own directory, Achado 2 - not against mid.fake's)", requested, wantRequested)
	}
}

func TestPass3IncludeFailsWhenTheFilenameIsNotAQuotedString(t *testing.T) {
	_, _, err := runFakeWithInclude(t,
		[]string{"include 5"},
		nil,
		nil,
	)
	if err == nil {
		t.Fatal("expected a fatal error for INCLUDE without a quoted filename, got none")
	}
}

func TestPass3IncludeFailsWithNoLoaderConfigured(t *testing.T) {
	ctx := newFakeMultilineContext([]string{`include "sub.fake"`}, []string{"fake.dmx"})
	platform := Platform{
		Description: fakeDescription{},
		Lexer:       fakeLexer{},
		Parser:      fakeParser{},
		Settings:    fakeSettings{},
	}
	_, err := Run(ctx, platform)
	if err == nil {
		t.Fatal("expected a fatal error for INCLUDE with no LoadFile configured, got none")
	}
}

func TestPass3IncludeFailsWhenTheLoaderErrors(t *testing.T) {
	_, _, err := runFakeWithInclude(t,
		[]string{`include "missing.fake"`},
		map[string][]string{},
		nil,
	)
	if err == nil {
		t.Fatal("expected a fatal error when the loader can't find the file, got none")
	}
}

func TestGenerateAddsSpaceAroundAReservedKeywordWhenNeighborsArentCompound(t *testing.T) {
	result, err := runFake(t, "PRINT FOO")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT FOO"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (two reserved words in a row get exactly one space, not a double space)", result.Code, want)
	}
}

func TestGenerateSuppressesSpacingAgainstACompoundSymbol(t *testing.T) {
	result, err := runFake(t, "PRINT(1)")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT(1)"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (\"(\" is a compound symbol - no space between a reserved keyword and it)", result.Code, want)
	}
}

func TestGenerateStillPadsAHashAfterAReservedKeyword(t *testing.T) {
	result, err := runFake(t, "PRINT#1")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 PRINT #1"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (\"#\" is deliberately NOT a compound symbol, badig_msx.py's own c_symbol_str.remove('#') - PRINT still gets its trailing space)", result.Code, want)
	}
}

func TestGenerateAlwaysPadsAfterTheLineNumberEvenWithoutAnyReservedWord(t *testing.T) {
	result, err := runFake(t, "a")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"10 a"}
	if strings.Join(result.Code, "|") != strings.Join(want, "|") {
		t.Fatalf("Code = %v, want %v (the line number's own trailing space is unconditional, not tied to the reserved-keyword rule)", result.Code, want)
	}
}

func TestGenerateFailsOnALineOverTheCharacterLimit(t *testing.T) {
	_, err := runFake(t, `PRINT "`+strings.Repeat("x", 260)+`"`)
	if err == nil {
		t.Fatal("expected a fatal error for a line over 256 characters, got none")
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
		Description: fakeDescription{},
		Parser:      fakeParser{pass4Calls: &pass4Calls, pass5Calls: &pass5Calls},
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

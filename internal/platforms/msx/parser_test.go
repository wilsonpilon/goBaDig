package msx

import (
	"strings"
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// parseLines lexes and parses (passes 1-5, gated the same way
// core.Parse always gates 4/5 to the main file) a small Dignified MSX
// program, returning the final token stream.
func parseLines(t *testing.T, lines ...string) ([]core.Token, *core.CompileContext) {
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
	platform := core.Platform{Description: Description, Lexer: Lexer{}, Parser: NewParser()}

	tokens, err := core.Lex(ctx, platform)
	if err != nil {
		t.Fatalf("core.Lex(%v): %v", lines, err)
	}
	tokens, err = core.Parse(ctx, tokens, platform)
	if err != nil {
		t.Fatalf("core.Parse(%v): %v", lines, err)
	}
	return tokens, ctx
}

// renamedValues returns the (possibly renamed) Value of every
// identifier-shaped token left in the stream, skipping structural
// tokens (including the line numbers core Pass 4 now inserts - these
// tests are about variable renaming/Pass 5, not numbering, which has
// its own tests in internal/core/pipeline_test.go), in source order.
func renamedValues(tokens []core.Token) []string {
	var out []string
	for _, tok := range tokens {
		switch tok.Kind {
		case core.KindProgram, core.KindNewline, core.KindEOF, core.KindLineNumber:
			continue
		}
		out = append(out, tok.String())
	}
	return out
}

// parseLinesWithParser is parseLines, but letting the caller pass a
// pre-configured *Parser (set ConvertPrint/StripThenGoto/Translate
// before calling) instead of always getting NewParser()'s defaults.
func parseLinesWithParser(t *testing.T, parser *Parser, lines ...string) []core.Token {
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
	platform := core.Platform{Description: Description, Lexer: Lexer{}, Parser: parser}

	tokens, err := core.Lex(ctx, platform)
	if err != nil {
		t.Fatalf("core.Lex(%v): %v", lines, err)
	}
	tokens, err = core.Parse(ctx, tokens, platform)
	if err != nil {
		t.Fatalf("core.Parse(%v): %v", lines, err)
	}
	return tokens
}

func TestParseRenamesLongVariableToDescendingAlias(t *testing.T) {
	tokens, _ := parseLines(t, "print counter")
	got := renamedValues(tokens)
	want := []string{"print", "zz"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseReusesTheSameAliasForARepeatedLongName(t *testing.T) {
	tokens, _ := parseLines(t, "counter=counter+1")
	got := renamedValues(tokens)
	// Both occurrences of "counter" must get the SAME alias:
	// [alias, "=", alias, "+", "1"].
	if len(got) != 5 || got[0] != got[2] {
		t.Fatalf("got %v, want [alias = alias + 1] with both aliases equal", got)
	}
}

func TestParseGivesDifferentLongNamesDifferentDescendingAliases(t *testing.T) {
	tokens, _ := parseLines(t, "print first", "print second")
	got := renamedValues(tokens)
	want := []string{"print", "zz", "print", "zy"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v (descending from zz)", got, want)
	}
}

func TestParseLeavesAnAlreadyShortNameUnchangedAndReservesIt(t *testing.T) {
	tokens, ctx := parseLines(t, "print ab")
	got := renamedValues(tokens)
	want := []string{"print", "ab"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v", got, want)
	}
	if !ctx.Symbols.HardShortVars["ab"] {
		t.Fatal("short name \"ab\" was not reserved in HardShortVars")
	}
}

func TestParseKeepsATildePrefixedNameVerbatimAndDropsTheTilde(t *testing.T) {
	tokens, ctx := parseLines(t, "print ~longname")
	got := renamedValues(tokens)
	want := []string{"print", "longname"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v (no \"~\" token, name untouched)", got, want)
	}
	if !ctx.Symbols.HardLongVars["longname@test.dmx"] {
		t.Fatal("\"longname\" was not registered in HardLongVars")
	}
}

// TestEndToEndLexParseGenerateRenamesVariables is the full pipeline
// (Lex -> Parse -> Generate), on a small but real Dignified MSX
// program - not just Lex+Parse like the tests above. It's the actual
// proof that Bloco 1's "v1" interfaces and Bloco 2's MSX Description/
// Lexer/Parser fit together end to end, same spirit as the fake-
// platform test in internal/core/pipeline_test.go, now with real
// vocabulary instead of a stand-in.
func TestEndToEndLexParseGenerateRenamesVariables(t *testing.T) {
	ctx, platform := contextPlatform(t, "counter=0", "print counter", "print ~kept")
	result, err := core.Run(ctx, platform)
	if err != nil {
		t.Fatalf("core.Run: %v", err)
	}
	want := []string{"10 zz = 0", "20 print zz", "30 print kept"}
	if len(result.Code) != len(want) {
		t.Fatalf("got %d lines %v, want %d lines %v", len(result.Code), result.Code, len(want), want)
	}
	for i := range want {
		if result.Code[i] != want[i] {
			t.Fatalf("line %d = %q, want %q (full: %v)", i, result.Code[i], want[i], result.Code)
		}
	}
}

func TestPass5ExpandsCompoundAssignment(t *testing.T) {
	tokens := parseLinesWithParser(t, NewParser(), "ab+=5")
	got := renamedValues(tokens)
	want := []string{"ab", "=", "ab", "+", "5"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPass5ExpandsIncrementWithTrailingOne(t *testing.T) {
	tokens := parseLinesWithParser(t, NewParser(), "ab++")
	got := renamedValues(tokens)
	want := []string{"ab", "=", "ab", "+", "1"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPass5ReplacesTrueAndFalse(t *testing.T) {
	tokens := parseLinesWithParser(t, NewParser(), "print true", "print false")
	got := renamedValues(tokens)
	want := []string{"print", "-", "1", "print", "0"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPass5ConvertsPrintToQuestionMark(t *testing.T) {
	parser := NewParser()
	parser.ConvertPrint = "?"
	tokens := parseLinesWithParser(t, parser, "print ab")
	got := renamedValues(tokens)
	want := []string{"?", "ab"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPass5ConvertsQuestionMarkToPrint(t *testing.T) {
	parser := NewParser()
	parser.ConvertPrint = "p"
	tokens := parseLinesWithParser(t, parser, "? ab")
	got := renamedValues(tokens)
	want := []string{"print", "ab"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPass5StripsThenBeforeGoto(t *testing.T) {
	parser := NewParser()
	parser.StripThenGoto = "T"
	tokens := parseLinesWithParser(t, parser, "if ab=1 then goto 100")
	got := renamedValues(tokens)
	want := []string{"if", "ab", "=", "1", "goto", "100"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v (no \"then\")", got, want)
	}
}

func TestPass5StripsGotoAfterThen(t *testing.T) {
	parser := NewParser()
	parser.StripThenGoto = "G"
	tokens := parseLinesWithParser(t, parser, "if ab=1 then goto 100")
	got := renamedValues(tokens)
	want := []string{"if", "ab", "=", "1", "then", "100"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v (no \"goto\")", got, want)
	}
}

func TestPass5TranslatesUnicodeInsideAString(t *testing.T) {
	parser := NewParser()
	parser.Translate = true
	tokens := parseLinesWithParser(t, parser, `print "☺"`)
	got := renamedValues(tokens)
	want := []string{"print", `"A"`}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestEndToEndLabelAndVariableRenamingCoexist confirms core's new
// label resolution/line numbering (ver internal/core/pipeline_test.go
// for the mechanism's own tests, not repeated here) and MSX's
// existing variable renaming don't interfere with each other when
// they run in the same compile.
func TestEndToEndLabelAndVariableRenamingCoexist(t *testing.T) {
	ctx, platform := contextPlatform(t, "goto {dest}", "{dest}", "print longvar")
	result, err := core.Run(ctx, platform)
	if err != nil {
		t.Fatalf("core.Run: %v", err)
	}
	want := []string{"10 goto 20", "20 print zz"}
	if len(result.Code) != len(want) {
		t.Fatalf("got %d lines %v, want %d lines %v", len(result.Code), result.Code, len(want), want)
	}
	for i := range want {
		if result.Code[i] != want[i] {
			t.Fatalf("line %d = %q, want %q (full: %v)", i, result.Code[i], want[i], result.Code)
		}
	}
}

// TestEndToEndLoopLabelAndVariableRenamingCoexist is
// TestEndToEndLabelAndVariableRenamingCoexist but for a loop label
// (core's own mechanism tests are in internal/core/pipeline_test.go)
// - confirms the implicit goto a loop's closing `}` emits doesn't
// interfere with MSX's variable renaming either.
func TestEndToEndLoopLabelAndVariableRenamingCoexist(t *testing.T) {
	ctx, platform := contextPlatform(t, "keys{", "print longvar", "}")
	result, err := core.Run(ctx, platform)
	if err != nil {
		t.Fatalf("core.Run: %v", err)
	}
	want := []string{"10 print zz", "20 goto 10"}
	if len(result.Code) != len(want) {
		t.Fatalf("got %d lines %v, want %d lines %v", len(result.Code), result.Code, len(want), want)
	}
	for i := range want {
		if result.Code[i] != want[i] {
			t.Fatalf("line %d = %q, want %q (full: %v)", i, result.Code[i], want[i], result.Code)
		}
	}
}

// TestEndToEndDefineLabelAndVariableRenamingCoexist confirms a
// DEFINE's content can itself hold a label reference (real example:
// badig/examples/msx/CGK173.dmx's `pauseKEY`/`pauseANY` defines do
// exactly this with `{@}`) and that core's DEFINE expansion (Pass 1)
// runs early enough for the spliced-in `{here}` to still be resolved
// correctly by Pass 2/4, together with MSX's own variable renaming.
func TestEndToEndDefineLabelAndVariableRenamingCoexist(t *testing.T) {
	ctx, platform := contextPlatform(t,
		"define [chk][if a=0 then{here}]",
		"[chk]",
		"{here}",
		"print longvar",
	)
	result, err := core.Run(ctx, platform)
	if err != nil {
		t.Fatalf("core.Run: %v", err)
	}
	want := []string{"10 if a = 0 then 20", "20 print zz"}
	if len(result.Code) != len(want) {
		t.Fatalf("got %d lines %v, want %d lines %v", len(result.Code), result.Code, len(want), want)
	}
	for i := range want {
		if result.Code[i] != want[i] {
			t.Fatalf("line %d = %q, want %q (full: %v)", i, result.Code[i], want[i], result.Code)
		}
	}
}

// TestEndToEndFuncRetAndVariableRenamingCoexist confirms a FUNC's
// parameter/default (just an ordinary identifier at Pass 1 time, when
// FUNC/RET are parsed) still gets renamed like any other long
// variable once Pass 3/4 run afterward - both in the function's own
// body AND in the assignment a call site's expansion inserts.
func TestEndToEndFuncRetAndVariableRenamingCoexist(t *testing.T) {
	ctx, platform := contextPlatform(t,
		"func .f(longvar=longvar)",
		"print longvar",
		"ret",
		".f(5)",
	)
	result, err := core.Run(ctx, platform)
	if err != nil {
		t.Fatalf("core.Run: %v", err)
	}
	want := []string{"10", "20 print zz", "30 return", "40 zz = 5 : gosub 10"}
	if len(result.Code) != len(want) {
		t.Fatalf("got %d lines %v, want %d lines %v", len(result.Code), result.Code, len(want), want)
	}
	for i := range want {
		if result.Code[i] != want[i] {
			t.Fatalf("line %d = %q, want %q (full: %v)", i, result.Code[i], want[i], result.Code)
		}
	}
}

func TestParseFailsUsingTildeOnAnAlreadyShortName(t *testing.T) {
	_, err := core.Run(contextPlatform(t, "print ~ab"))
	if err == nil {
		t.Fatal("expected a fatal error using ~ on a 2-letter name, got none")
	}
}

// get_hard_variable also fatally rejects re-using `~` on a name
// already present in Declares - but Declares is only ever populated
// by Pass 4 (which always runs strictly after Pass 3 finishes for
// every token) or by an explicit `declare` statement, which isn't
// implemented yet (core-level Dignified instruction handling, out of
// scope this round - see OUTLINE.md, Bloco 2). That branch is ported
// faithfully in parser.go for when `declare` lands, but isn't
// reachable from this round's actual capability, so it has no test
// here - a misleading "reaches this code" test would be worse than
// none.

// contextPlatform builds the (ctx, platform) pair core.Run wants,
// same shape as parseLines above but exposed separately for tests
// that need to call core.Run directly instead of just core.Lex+core.Parse.
func contextPlatform(t *testing.T, lines ...string) (*core.CompileContext, core.Platform) {
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
	return ctx, core.Platform{Description: Description, Lexer: Lexer{}, Parser: NewParser()}
}

package zx81

import (
	"strings"
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// contextPlatform mirrors internal/platforms/msx/parser_test.go's own
// helper - same shape, ZX81 vocabulary.
func contextPlatform(t *testing.T, lines ...string) (*core.CompileContext, core.Platform) {
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
	return ctx, core.Platform{Description: Description, Lexer: Lexer{}, Parser: NewParser()}
}

func runLines(t *testing.T, lines ...string) core.GenerateResult {
	t.Helper()
	ctx, platform := contextPlatform(t, lines...)
	result, err := core.Run(ctx, platform)
	if err != nil {
		t.Fatalf("core.Run(%v): %v", lines, err)
	}
	return result
}

func assertCode(t *testing.T, got core.GenerateResult, want []string) {
	t.Helper()
	if len(got.Code) != len(want) {
		t.Fatalf("got %d lines %v, want %d lines %v", len(got.Code), got.Code, len(want), want)
	}
	for i := range want {
		if got.Code[i] != want[i] {
			t.Fatalf("line %d = %q, want %q (full: %v)", i, got.Code[i], want[i], got.Code)
		}
	}
}

// TestEndToEndAllocatesOneLetterVariablesDescendingFromZ confirms
// process_variable's own allocation order (badig_zx81.py:350:
// `reversed(self.des.c_var_chr)`) - the FIRST long name seen gets "z",
// the second gets "y", and so on.
func TestEndToEndAllocatesOneLetterVariablesDescendingFromZ(t *testing.T) {
	got := runLines(t, "counter=0", "let other=1", "print counter", "print other")
	assertCode(t, got, []string{"10 let z=0", "20 let y=1", "30 print z", "40 print y"})
}

// TestEndToEndAlreadyShortVariablesAreLeftAlone confirms a
// single-letter name used directly in the source is never touched -
// nor handed out to a different long name later.
func TestEndToEndAlreadyShortVariablesAreLeftAlone(t *testing.T) {
	got := runLines(t, "a=1", "longname=2")
	assertCode(t, got, []string{"10 let a=1", "20 let z=2"})
}

// TestEndToEndDollarSuffixSharesTheSameAliasSlot confirms the "$"
// string-type suffix rides along with whatever alias its base name
// gets, same as MSX's own %/!/#/$ handling (ver Pass4's own doc
// comment - ZX81 only has "$", no numeric-type suffixes).
func TestEndToEndDollarSuffixSharesTheSameAliasSlot(t *testing.T) {
	got := runLines(t, `name$="hi"`, "print name$")
	assertCode(t, got, []string{`10 let z$="hi"`, "20 print z$"})
}

// TestEndToEndTildeRejectsZX81KeepingALongName confirms the big
// architectural difference from MSX documented in Pass4's own doc
// comment: "~" can never actually be honored for ZX81 (classic BASIC
// has no way to represent a name longer than 1 char), so using it on
// a genuinely long name is always a fatal error - for EITHER the
// marked occurrence or a later, unmarked one.
func TestEndToEndTildeRejectsZX81KeepingALongName(t *testing.T) {
	ctx, platform := contextPlatform(t, "print ~longname")
	_, err := core.Run(ctx, platform)
	if err == nil || !strings.Contains(err.Error(), "must be one letter") {
		t.Fatalf("want a 'must be one letter' error, got %v", err)
	}
}

func TestEndToEndTildeOnAnAlreadyShortNameIsStillRejected(t *testing.T) {
	ctx, platform := contextPlatform(t, "print ~a")
	_, err := core.Run(ctx, platform)
	if err == nil || !strings.Contains(err.Error(), "can't use ~") {
		t.Fatalf("want a \"can't use ~\" error, got %v", err)
	}
}

// TestEndToEndImplicitLetAfterLineStart confirms generate()'s own
// ZX-81-specific rule (badig_zx81.py:380-383): a bare assignment at
// the start of a statement gets "let " inserted automatically -
// ZX-81 classic BASIC has no bare-assignment syntax at all.
func TestEndToEndImplicitLetAfterLineStart(t *testing.T) {
	got := runLines(t, "a=1")
	assertCode(t, got, []string{"10 let a=1"})
}

// TestEndToEndImplicitLetAfterThen confirms the SAME rule also fires
// right after "THEN" (the other place a ZX-81 statement can start
// mid-line), not just right after a line number.
func TestEndToEndImplicitLetAfterThen(t *testing.T) {
	got := runLines(t, "if a=1 then b=2")
	assertCode(t, got, []string{"10 if a=1 then let b=2"})
}

// TestEndToEndImplicitLetDoesNotFireMidStatement confirms the rule is
// narrow: an identifier anywhere else (e.g. the right-hand side of an
// assignment) must NOT get a spurious "let" inserted in front of it.
func TestEndToEndImplicitLetDoesNotFireMidStatement(t *testing.T) {
	got := runLines(t, "a=b")
	assertCode(t, got, []string{"10 let a=b"})
}

// TestEndToEndPrintAtWithBothCoordinates confirms the PRINT AT
// builtin define's main case: `[?](row,col) "text"` expands to
// `print at row,col;"text"` - confirmed against the real engine's own
// behavior during this session's research (`[?](10,5) "ZX-81"` ->
// `print at 10,5;"zx-81"`).
func TestEndToEndPrintAtWithBothCoordinates(t *testing.T) {
	got := runLines(t, `[?](10,5) "hi"`)
	assertCode(t, got, []string{`10 print at 10,5;"hi"`})
}

// TestEndToEndPrintAtDefaultsMissingCoordinatesToZero covers the
// three argument-count cases replace_print_at itself distinguishes:
// no call at all (bare "[?]"), one coordinate only, and neither.
func TestEndToEndPrintAtDefaultsMissingCoordinatesToZero(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{`[?] "hi"`, `10 print at 0,0;"hi"`},
		{`[?](7) "hi"`, `10 print at 7,0;"hi"`},
		{`[?](,7) "hi"`, `10 print at 0,7;"hi"`},
	}
	for _, c := range cases {
		t.Run(c.line, func(t *testing.T) {
			got := runLines(t, c.line)
			assertCode(t, got, []string{c.want})
		})
	}
}

// TestEndToEndPrintAtKeepsNestedParensOutOfTheSplit confirms the
// comma-splitting respects parentheses nesting (a coordinate
// expression can itself contain a comma-using call/expression),
// mirroring replace_print_at's own `nesting` tracking: the comma
// inside "peek(1,2)" must stay part of the column side, not be
// mistaken for the row/column separator.
func TestEndToEndPrintAtKeepsNestedParensOutOfTheSplit(t *testing.T) {
	got := runLines(t, `[?](row,peek(1,2)) "hi"`)
	assertCode(t, got, []string{`10 print at z,peek(1,2);"hi"`})
}

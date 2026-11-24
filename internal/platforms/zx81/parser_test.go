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

// TestEndToEndStringLiteralEscapesSurviveToGeneratedCode ports
// test_doubled_and_backslash_quotes_inside_dignified_strings
// (test_tokenizer_zx81.py) - confirms both quote-escape styles
// (ver lexer_test.go's own TestStringLiteralSupportsBothQuoteEscapeStyles
// for the LEXER-level check) survive unchanged all the way through
// Generate(), since nothing between lexing and rendering ever decodes
// them - that only happens later, at the tokenizer stage (already
// ported, Etapa 1). Confirmed against the real badig.py.
func TestEndToEndStringLiteralEscapesSurviveToGeneratedCode(t *testing.T) {
	got := runLines(t,
		`print "TESTE ""123"" OK"`,
		`print "LEGACY \"123\" OK"`,
	)
	assertCode(t, got, []string{
		`10 print "TESTE ""123"" OK"`,
		`20 print "LEGACY \"123\" OK"`,
	})
}

// TestEndToEndDeclareExplicitAliasInteractsWithAutoAllocation ports
// test_long_variables_allocate_from_z_down_and_honor_declare: a
// DECLAREd alias is reserved up front (so the auto-allocator must
// skip it), and a later long name's own auto-allocated alias must
// still descend correctly from "z" regardless. Confirmed against the
// real badig.py.
func TestEndToEndDeclareExplicitAliasInteractsWithAutoAllocation(t *testing.T) {
	got := runLines(t,
		"declare message:c",
		`temperature$ = "warm"`,
		"counter = 1",
		"print temperature$",
		"print message$",
	)
	assertCode(t, got, []string{
		`10 let z$="warm"`,
		"20 let y=1",
		"30 print z$",
		"40 print c$",
	})
}

// TestEndToEndRealIntroExample is a real, substantial ZX81 Dignified
// program - badig/examples/zx81/zx81-intro.DZX (minus its own
// "##BB:arguments=-ca" remtag/comment header, since remtag extraction
// and capitalise_all are both out of scope - ver OUTLINE.md), run
// through this whole package's Lex->Parse->Generate pipeline.
// Confirmed BYTE FOR BYTE against a real `python badig.py
// examples/zx81/zx81-intro.DZX -id zx81 --tk_ascii_only` run on this
// machine (with the remtag/comments stripped, same reason) - same
// validation discipline the MSX port used against CGK173.dmx.
// Exercises DECLARE with two aliases in one statement, two different
// FOR/NEXT loop variables, raw classic "print at R,C;" syntax typed
// directly (not through the "[?]" shorthand), backslash-escaped
// graphics characters inside strings, a loop label that targets its
// OWN line (an infinite "wait for a key" loop), and STOP.
func TestEndToEndRealIntroExample(t *testing.T) {
	got := runLines(t,
		"declare column:c, row:r",
		"",
		"cls",
		"",
		"for column = 2 to 29",
		`    print at 1,column; "\::"`,
		`    print at 20,column; "\::"`,
		"next column",
		"",
		"for row = 2 to 19",
		`    print at row,1; "\::"`,
		`    print at row,30; "\::"`,
		"next row",
		"",
		`print at 1,1; "\##"`,
		`print at 1,30; "\##"`,
		`print at 20,1; "\##"`,
		`print at 20,30; "\##"`,
		"",
		`print at 3,10; "\~~ ZX-81 \~~"`,
		`print at 5,6; "BASIC DIGNIFIED SUITE"`,
		`print at 9,8; "AUTHOR: FARIQUE"`,
		`print at 12,6; "ZX-81 MODULE: BARNEY"`,
		`print at 16,3; "COPYLEFT: Cybernostra, Inc."`,
		`print at 18,12; "2026"`,
		`print at 19,9; "PRESS ANY KEY"`,
		"",
		"{wait_key}",
		`if inkey$ = "" then goto {wait_key}`,
		"stop",
	)
	assertCode(t, got, []string{
		"10 cls",
		"20 for c=2 to 29",
		`30 print at 1,c;"\::"`,
		`40 print at 20,c;"\::"`,
		"50 next c",
		"60 for r=2 to 19",
		`70 print at r,1;"\::"`,
		`80 print at r,30;"\::"`,
		"90 next r",
		`100 print at 1,1;"\##"`,
		`110 print at 1,30;"\##"`,
		`120 print at 20,1;"\##"`,
		`130 print at 20,30;"\##"`,
		`140 print at 3,10;"\~~ ZX-81 \~~"`,
		`150 print at 5,6;"BASIC DIGNIFIED SUITE"`,
		`160 print at 9,8;"AUTHOR: FARIQUE"`,
		`170 print at 12,6;"ZX-81 MODULE: BARNEY"`,
		`180 print at 16,3;"COPYLEFT: Cybernostra, Inc."`,
		`190 print at 18,12;"2026"`,
		`200 print at 19,9;"PRESS ANY KEY"`,
		`210 if inkey$="" then goto 210`,
		"220 stop",
	})
}

// TestEndToEndRealMcodeExample is the OTHER real example,
// badig/examples/zx81/mcode-ok.DZX, minus its own remtag/comment
// header AND its "remx 0,256,"0"" line (remx-as-a-Dignified-directive
// is deliberately deferred - ver OUTLINE.md, Bloco 3) - confirmed
// byte for byte the same way as TestEndToEndRealIntroExample above.
// Exercises DECLARE with an EXPLICIT alias per variable, a "$"-suffix
// variable sharing its base name's alias, string indexing
// (`h$(i)`), arithmetic expressions, and the "len"/"code"/"rand"/
// "usr" functions.
func TestEndToEndRealMcodeExample(t *testing.T) {
	got := runLines(t,
		"declare routine:h, address:a",
		"",
		"cls",
		`print "LOADING MACHINE CODE..."`,
		`routine$ = "3E34D73E30D7C9"`,
		"address = 16514",
		"for i = 1 to len routine$ step 2",
		"    poke address, 16 * (code routine$(i) - 28) + code routine$(i + 1) - 28",
		"    address = address + 1",
		"next i",
		"",
		"print",
		`print "CALLING USR 16514: ";`,
		"rand usr 16514",
	)
	assertCode(t, got, []string{
		"10 cls",
		`20 print "LOADING MACHINE CODE..."`,
		`30 let h$="3E34D73E30D7C9"`,
		"40 let a=16514",
		"50 for i=1 to len h$ step 2",
		"60 poke a,16*(code h$(i)-28)+code h$(i+1)-28",
		"70 let a=a+1",
		"80 next i",
		"90 print",
		`100 print "CALLING USR 16514: ";`,
		"110 rand usr 16514",
	})
}

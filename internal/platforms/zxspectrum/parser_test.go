package zxspectrum

import (
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

func contextPlatform(t *testing.T, lines ...string) (*core.CompileContext, core.Platform) {
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

func runLinesExpectError(t *testing.T, lines ...string) error {
	t.Helper()
	ctx, platform := contextPlatform(t, lines...)
	_, err := core.Run(ctx, platform)
	if err == nil {
		t.Fatalf("core.Run(%v): expected an error, got none", lines)
	}
	return err
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

// TestEndToEndPrintAtSplicesOneArgumentLiterally confirms the real,
// confirmed-by-running-badig.py behavior: "[?](5)" splices its ONE
// argument verbatim into the single placeholder, producing "PRINT AT
// 5;" with no comma/second coordinate - NOT independently defaulted
// per-coordinate the way ZX81's own "[?](row,col)" is (ver parser.go's
// BuiltinDefines doc comment).
func TestEndToEndPrintAtSplicesOneArgumentLiterally(t *testing.T) {
	got := runLines(t, `[?](5) "HI"`)
	assertCode(t, got, []string{`10 PRINT AT 5;"HI"`})
}

func TestEndToEndPrintAtWithBothCoordinatesAndDefault(t *testing.T) {
	got := runLines(t, `[?](5,10) "HI"`, `[?] "HI"`)
	assertCode(t, got, []string{`10 PRINT AT 5,10;"HI"`, `20 PRINT AT 0,0;"HI"`})
}

// TestEndToEndApostropheIsUntouched confirms the real Sinclair
// PRINT-list separator passes straight through (ver description.go's
// own doc comment on symbolWords).
func TestEndToEndApostropheIsUntouched(t *testing.T) {
	got := runLines(t, `PRINT "A"'"B"`)
	assertCode(t, got, []string{`10 PRINT "A"'"B"`})
}

func TestEndToEndImplicitLetAfterLineStartAndColonAndThen(t *testing.T) {
	got := runLines(t, "CLS:x=1", "IF x=1 THEN y=2")
	assertCode(t, got, []string{"10 CLS:LET x=1", "20 IF x=1 THEN LET y=2"})
}

func TestEndToEndImplicitLetDoesNotFireMidStatement(t *testing.T) {
	got := runLines(t, "LET x=1")
	assertCode(t, got, []string{"10 LET x=1"})
}

func TestEndToEndCompoundAssignmentExpandsFullRhsInParens(t *testing.T) {
	got := runLines(t, "CLS:x=1", "x+=5*2")
	assertCode(t, got, []string{"10 CLS:LET x=1", "20 LET x=x+(5*2)"})
}

func TestEndToEndIncrementAndDecrement(t *testing.T) {
	got := runLines(t, "x=1", "x++", "x--")
	assertCode(t, got, []string{"10 LET x=1", "20 LET x=x+(1)", "30 LET x=x-(1)"})
}

func TestEndToEndStringCompoundAssignmentOnlyAllowsPlusEquals(t *testing.T) {
	got := runLines(t, `a$="x"`, `a$+="y"`)
	assertCode(t, got, []string{`10 LET a$="x"`, `20 LET a$=a$+("y")`})

	runLinesExpectError(t, `a$="x"`, `a$-="y"`)
}

func TestEndToEndIncrementRejectsAnyRightOperand(t *testing.T) {
	runLinesExpectError(t, "x=1", "x++1")
}

func TestEndToEndNestedCompoundAssignmentIsRejected(t *testing.T) {
	runLinesExpectError(t, "x=1", "y=1", "x+=y+=1")
}

func TestEndToEndCompoundAssignmentRequiresAPrecedingStatementStart(t *testing.T) {
	// "x" here is mid-expression (the RHS of "y="), not a statement's
	// own target - no implicit LET precedes it.
	runLinesExpectError(t, "x=1", "y=x+=1")
}

func TestEndToEndTrueFalseAreOneAndZero(t *testing.T) {
	got := runLines(t, "x=TRUE", "y=FALSE")
	assertCode(t, got, []string{"10 LET x=1", "20 LET y=0"})
}

func TestEndToEndGotoAndGosubAlwaysRenderWithASpaceRegardlessOfSpelling(t *testing.T) {
	got := runLines(t, "GOTO {lbl}", "GO    TO {lbl}", "GOSUB {lbl}", "GO SUB {lbl}", "{lbl}", "STOP")
	assertCode(t, got, []string{
		"10 GO TO 50",
		"20 GO TO 50",
		"30 GO SUB 50",
		"40 GO SUB 50",
		"50 STOP",
	})
}

func TestEndToEndReservedWordsAlwaysRenderUppercase(t *testing.T) {
	got := runLines(t, "print sin(1)+cos(2)")
	assertCode(t, got, []string{"10 PRINT SIN(1)+COS(2)"})
}

func TestEndToEndRemContentStaysAsTypedWithDoubleSpaceAfterRem(t *testing.T) {
	// Confirmed empirically (running badig.py): the keyword "REM" is
	// forced uppercase and gets core.Generate's own automatic
	// reserved-word space, PLUS whatever leading space the user's own
	// comment text already had - two spaces total, not a port bug.
	got := runLines(t, `print "hi": rem lowercase stays`)
	assertCode(t, got, []string{`10 PRINT "hi":REM  lowercase stays`})
}

// TestEndToEndStringVariableNamesMustBeExactlyOneLetterPlusDollar
// confirms validate_identifier's own `len(name)==2` check counts the
// WHOLE string INCLUDING the "$" - so a valid Sinclair string name is
// exactly one letter plus "$" ("a$"), not two letters plus "$".
func TestEndToEndStringVariableNamesMustBeExactlyOneLetterPlusDollar(t *testing.T) {
	got := runLines(t, `a$="hi"`)
	assertCode(t, got, []string{`10 LET a$="hi"`})
	runLinesExpectError(t, `ab$="hi"`)
}

func TestEndToEndArrayAndForControlVariablesMustBeOneLetter(t *testing.T) {
	got := runLines(t, "DIM a(5)", "FOR i=1 TO 5", "NEXT i")
	assertCode(t, got, []string{"10 DIM a(5)", "20 FOR i=1 TO 5", "30 NEXT i"})
	runLinesExpectError(t, "DIM array(5)")
	runLinesExpectError(t, "FOR counter=1 TO 5", "NEXT counter")
}

func TestEndToEndScalarNumericNamesAcceptAnyLength(t *testing.T) {
	got := runLines(t, "totalscore=1")
	assertCode(t, got, []string{"10 LET totalscore=1"})
}

// TestEndToEndCollidingWithARomKeywordIsRejected uses "cat" - a real
// ROM 48K token (tokenCodes, tokenizer.go) that is deliberately NOT
// part of this module's own Dignified `instructions` vocabulary (CAT/
// FORMAT/MOVE/ERASE are disk/microdrive commands this tape-focused
// module doesn't support at all - confirmed by tokenizer_test.go's own
// TestRejectsUnrepresentableOrUnsupportedSource, which rejects "CAT"
// as a STATEMENT too) - so it lexes as a plain identifier, reaching
// Pass4's collision check. "print"/"let"/etc. wouldn't exercise this
// at all: they're already Dignified keywords, never lexed as
// identifiers in the first place.
func TestEndToEndCollidingWithARomKeywordIsRejected(t *testing.T) {
	runLinesExpectError(t, "cat=1")
}

func TestEndToEndTildeIsRemovedWithoutAnyAliasingEffect(t *testing.T) {
	got := runLines(t, "~a=1")
	assertCode(t, got, []string{"10 LET a=1"})
}

func TestEndToEndDeclareSubstitutesTheAliasAndKeepsTheDollarSuffix(t *testing.T) {
	got := runLines(t, "DECLARE longname:x", "longname=1", "longname$=\"hi\"")
	assertCode(t, got, []string{"10 LET x=1", "20 LET x$=\"hi\""})
}

func TestEndToEndLineNumberBeyond9999IsRejected(t *testing.T) {
	ctx, platform := contextPlatform(t, "stop")
	ctx.Settings.LineStart = 10000
	_, err := core.Run(ctx, platform)
	if err == nil {
		t.Fatal("expected a line-number-range error")
	}
}

func TestEndToEndRemxIsRejectedAsNotImplementedYet(t *testing.T) {
	runLinesExpectError(t, "REMX 100,4")
}

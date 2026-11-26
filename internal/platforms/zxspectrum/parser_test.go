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

// --- Etapa 4: real examples, badig/examples/zxspectrum/ ---
//
// Each of the following mirrors one real .DZS file from the project's
// own local reference (badig/ - gitignored, so its own text is
// embedded here directly, no file read), minus its own leading "##"
// Dignified-comment header line (cosmetic, produces no token either
// way) - confirmed BYTE FOR BYTE against the real badig.py (not just
// these unit tests): both the rendered Classic listing (.AZS) AND the
// tokenized .tap this package's own Tokenize/MakeTap produce from it
// match the real Python's own output exactly, for ALL 9 real examples
// in that directory (hello/data/fn/colours/functions/native/streams/
// graphics/play128 - graphics' only with its "REMX 1000,32" line
// removed, same deferred-scope decision as ZX81's own real-example
// tests; play128 tokenized with model="128"). These three were picked
// to commit as the most structurally distinct: the simplest full
// program (labels/implicit LET/GOTO), compound assignment + FUNC/RET,
// and DEF FN/FN. RemHeader is left at its default (false) here, same
// as runLines everywhere else in this file - the real header/REM-
// numbering shift this produces vs. the real file's own output
// (RemHeader=true) was manually re-derived from the confirmed-correct
// run, not guessed.

// TestEndToEndRealHelloExample mirrors badig/examples/zxspectrum/hello.DZS.
func TestEndToEndRealHelloExample(t *testing.T) {
	got := runLines(t,
		"REM FIRST SPECTRUM MILESTONE",
		"CLS",
		"LET counter=1",
		"{again}",
		`PRINT "SPECTRUM ";counter`,
		"LET counter=counter+1",
		"IF counter<=3 THEN GO TO {again}",
		`PRINT "DONE"`,
		"STOP",
	)
	assertCode(t, got, []string{
		"10 REM  FIRST SPECTRUM MILESTONE",
		"20 CLS",
		"30 LET counter=1",
		`40 PRINT "SPECTRUM ";counter`,
		"50 LET counter=counter+1",
		"60 IF counter<=3 THEN GO TO 40",
		`70 PRINT "DONE"`,
		"80 STOP",
	})
}

// TestEndToEndRealFunctionsExample mirrors
// badig/examples/zxspectrum/functions.DZS - exercises `*=`/`++` compound
// assignment, a Dignified FUNC/RET call (core-level, shared machinery),
// and `[?](row,col)` PRINT AT with both coordinates given.
func TestEndToEndRealFunctionsExample(t *testing.T) {
	got := runLines(t,
		"REM DIGNIFIED CONVENIENCES",
		"CLS",
		"score=2",
		"score*=1+2",
		"score++",
		"IF score=7 THEN score+=3",
		`[?](2,0) "SCORE ";score`,
		"answer=.double(score)",
		"PRINT answer",
		"STOP",
		"FUNC .double(value)",
		"    result=value*2",
		"RET result",
	)
	assertCode(t, got, []string{
		"10 REM  DIGNIFIED CONVENIENCES",
		"20 CLS",
		"30 LET score=2",
		"40 LET score=score*(1+2)",
		"50 LET score=score+(1)",
		"60 IF score=7 THEN LET score=score+(3)",
		`70 PRINT AT 2,0;"SCORE ";score`,
		"80 LET value=score:GO SUB 110:LET answer=result",
		"90 PRINT answer",
		"100 STOP",
		"110 LET result=value*2",
		"120 RETURN",
	})
}

// TestEndToEndRealFnExample mirrors badig/examples/zxspectrum/fn.DZS -
// exercises DEF FN (numeric, string-named, no-arg, and one calling
// another FN internally) and FN calls, including a nested one.
func TestEndToEndRealFnExample(t *testing.T) {
	got := runLines(t,
		"CLS",
		"DEF FN f(x,a$)=x+LEN a$",
		`DEF FN s$(a$)=a$+"!"`,
		"DEF FN z()=42",
		`DEF FN g(x)=FN f(x,"AB")*2`,
		"x=99",
		`PRINT "FN NUMERIC ";FN f(3,"AB")`,
		`PRINT "FN STRING ";FN s$("SPECTRUM")`,
		`PRINT "FN NO ARGS ";FN z()`,
		`PRINT "FN NESTED ";FN g(3)`,
		`PRINT "GLOBAL X ";x`,
		"STOP",
	)
	assertCode(t, got, []string{
		"10 CLS",
		"20 DEF FN f(x,a$)=x+LEN a$",
		`30 DEF FN s$(a$)=a$+"!"`,
		"40 DEF FN z()=42",
		`50 DEF FN g(x)=FN f(x,"AB")*2`,
		"60 LET x=99",
		`70 PRINT "FN NUMERIC ";FN f(3,"AB")`,
		`80 PRINT "FN STRING ";FN s$("SPECTRUM")`,
		`90 PRINT "FN NO ARGS ";FN z()`,
		`100 PRINT "FN NESTED ";FN g(3)`,
		`110 PRINT "GLOBAL X ";x`,
		"120 STOP",
	})
}

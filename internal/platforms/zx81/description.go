package zx81

import "github.com/wilsonpilon/goBaDig/internal/core"

// Vocabulary tables, transcribed verbatim (same entries, same order
// where order matters) from badig/zx81/badig_zx81.py's Description
// class (__init__, read in full this session, not summarized).
//
// These are TEXT-matching tables for the Dignified-level Lexer - they
// classify what a piece of source IS, they do not encode it to bytes.
// Byte encoding is a later, separate stage (the already-ported
// tokenizer, tokenizer.go in this same package), which works on the
// generated Classic ASCII text, not on these Lexer tokens.
var (
	// instructions mirrors c_instrc (badig_zx81.py:146-150): 31 plain
	// keywords. Deliberately smaller than MSX's own instructions list -
	// ZX81 BASIC has far fewer statements, and several MSX-style
	// keywords (RUN, LIST, LLIST) are classified here as plain
	// instructions for ZX81, not as jump keywords (see jumpKeywords).
	instructions = []string{
		"COPY", "CLEAR", "UNPLOT", "CLS", "IF", "RANDOMIZE", "RAND",
		"SAVE", "RUN", "PLOT", "PRINT", "POKE", "NEXT", "PAUSE",
		"LET", "LIST", "LOAD", "INPUT", "FOR", "TO", "STEP", "DIM",
		"CONTINUE", "CONT", "SCROLL", "NEW", "FAST", "SLOW", "STOP",
		"LLIST", "LPRINT",
	}

	// functionsDollar mirrors c_funcdl (badig_zx81.py:151): functions
	// whose name ends in "$".
	functionsDollar = []string{"CHR$", "STR$", "INKEY$"}

	// functionsPlain mirrors c_funcnm (badig_zx81.py:152-154). Unlike
	// MSX, ZX81's "USR" takes no optional trailing device digit - it's
	// a plain entry here, no special-cased scanning needed for it.
	functionsPlain = []string{
		"USR", "PEEK", "ABS", "SGN", "SQR", "INT", "EXP", "LN",
		"ATN", "ACS", "ASN", "TAN", "COS", "SIN", "LEN", "VAL",
		"CODE", "TAB", "AT", "PI", "RND",
	}

	// jumpKeywords mirrors c_jumpin (badig_zx81.py:155) - deliberately
	// small (unlike MSX's own larger jumpKeywords): RUN/LIST/LLIST are
	// plain instructions for ZX81 (see instructions above), not jump
	// keywords, even though they take a line-number-shaped argument in
	// classic BASIC too.
	jumpKeywords = []string{"GOTO", "GOSUB", "RETURN", "THEN"}

	// operators mirrors c_operat (badig_zx81.py:156) - no MOD/XOR,
	// unlike MSX.
	operators = []string{"AND", "OR", "NOT"}

	// symbols mirrors c_symbol (badig_zx81.py:157-158), already
	// unescaped and in the SAME order as the Python source - order
	// matters here: the 4 two-character entries ("<>", ">=", "<=",
	// "**") must be checked before any of their single-character
	// prefixes, or e.g. "<>" would wrongly match "<" first and leave a
	// dangling ">". Confirmed empirically (ver lexer.go's own doc
	// comment on signed numbers for the same kind of empirical check).
	symbols = []string{
		"<>", ">=", "<=", "**",
		">", "=", "<", "+", "*", "/", "-", ",", "(", ")", ";", "?", "#", `\`,
	}

	// extendedSymbols mirrors c_e_symb (badig_zx81.py:159) - compound
	// assignment/increment/decrement operators, Dignified-only syntax
	// (ZX81 classic BASIC has no such operators itself). Must be
	// checked BEFORE symbols above, same reason as MSX's own
	// extendedSymbols (ver msx/description.go).
	extendedSymbols = []string{"++", "--", "+=", "-=", "*=", "/=", "^="}
)

// description implements core.Description for ZX81.
type description struct{}

// Description is the ZX81 core.Description singleton.
var Description core.Description = description{}

func (description) SysInfo() core.SystemInfo {
	return core.SystemInfo{
		Name:         "Sinclair ZX-81",
		DignifiedExt: ".dzx",
		ClassicExt:   ".azx",
		BinaryExt:    ".p",
	}
}

func (description) QuoteChar() byte { return '"' }

// StatementSeparator mirrors c_instsp_str (badig_zx81.py:179) - same
// as MSX.
func (description) StatementSeparator() string { return ":" }

// LoopBackInstruction mirrors c_loop_back (badig_zx81.py:180) - same
// as MSX.
func (description) LoopBackInstruction() string { return "goto" }

// CommentChar mirrors c_altrem (badig_zx81.py:168): "REM", the WHOLE
// WORD, unlike MSX's single punctuation character ("'"). ZX81 classic
// BASIC has no punctuation shortcut for REM at all - confirmed reading
// badig_zx81.py's own classic_groups construction (no clc_brem entry,
// unlike MSX's badig_msx.py:205/221), so ZX81 genuinely has no "”"
// verbatim-comment-BLOCK syntax either (core's own shared mechanism,
// pipeline.go, derives that marker from CommentChar()+CommentChar() -
// "REMREM" never appears as real source text, so that path correctly
// never fires for this platform; no extra code needed here for that).
func (description) CommentChar() string { return "REM" }

// AssignmentOperator mirrors c_equal, inherited unchanged from
// MSXDescription (badig_zx81.py's Description never overrides it) -
// same value as MSX.
func (description) AssignmentOperator() string { return "=" }

// FunctionCallInstruction mirrors c_func_call (badig_zx81.py:181) -
// same as MSX.
func (description) FunctionCallInstruction() string { return "gosub" }

// FunctionReturnInstruction mirrors c_func_ret (badig_zx81.py:182) -
// same as MSX.
func (description) FunctionReturnInstruction() string { return "return" }

// FunctionCallBoundaryWords mirrors c_func_stop_kw (badig_zx81.py:
// 183-184: "c_instrc + c_jumpin + c_funcdl + c_funcnm").
func (description) FunctionCallBoundaryWords() []string {
	words := make([]string, 0, len(instructions)+len(jumpKeywords)+len(functionsDollar)+len(functionsPlain))
	words = append(words, instructions...)
	words = append(words, jumpKeywords...)
	words = append(words, functionsDollar...)
	words = append(words, functionsPlain...)
	return words
}

// VariableAliasLength mirrors c_var_valid_chars (badig_zx81.py:176):
// 1, not 2 like MSX - every ZX81 variable alias is a single letter.
func (description) VariableAliasLength() int { return variableAliasLen }

// ReservedKeywords mirrors c_reserved_kw (badig_zx81.py:165: "c_instrc
// + c_jumpin + c_funcdl + c_funcnm + c_operat + ['REM']") - adds the
// literal word "REM" (ZX81 has no DATA statement, unlike MSX, which
// adds "DATA" here instead).
func (description) ReservedKeywords() []string {
	words := make([]string, 0, len(instructions)+len(jumpKeywords)+len(functionsDollar)+len(functionsPlain)+len(operators)+1)
	words = append(words, instructions...)
	words = append(words, jumpKeywords...)
	words = append(words, functionsDollar...)
	words = append(words, functionsPlain...)
	words = append(words, operators...)
	words = append(words, "REM")
	return words
}

// CompoundSymbols mirrors c_symb_comp (badig_zx81.py:164: "c_symbol +
// [':', '_']") - unlike MSX, nothing is removed from symbols here
// (ZX81 has no "print #1 needs a space after print" concern the way
// MSX does; "#" stays in c_symb_comp unlike MSX's c_symb_comp).
// extendedSymbols (c_e_symb) is NOT included here - confirmed reading
// badig_zx81.py's own c_symb_comp assignment, which is built from
// c_symbol alone.
func (description) CompoundSymbols() []string {
	words := make([]string, 0, len(symbols)+2)
	words = append(words, symbols...)
	words = append(words, ":", "_")
	return words
}

// Variable allocation constants, confirmed reading process_variable
// (badig_zx81.py:330-356) and its surrounding Description fields
// (badig_zx81.py:172-178) directly this session. Not used yet by this
// package - the Parser (next phase) is what actually allocates
// aliases - kept here now so they aren't rediscovered/retranscribed
// later. internal/core/allocator.go's own doc comment already names
// this exact scheme ("ZX81 (everything 1 letter)") as something its
// generic VariableAllocator was designed to support.
const (
	// variableAliasChars is c_var_chr: ZX81's own alias alphabet -
	// happens to be the same 26 lowercase letters as MSX's, but used
	// 1 at a time here, not 2.
	variableAliasChars = "abcdefghijklmnopqrstuvwxyz"
	// variableAliasLen is c_var_valid_chars: every ZX81 alias is a
	// single character (unlike MSX's fixed 2).
	variableAliasLen = 1
)

// variableCategory is the single ZX81 variable category (core's
// VariableAllocator, from Bloco 1, is parametrized by category - ZX81
// just has one, like MSX, since it uses the same 1-letter scheme for
// every variable regardless of type). Not used yet; see the const
// block above. core.NewDescendingAllocator() hands out aliases
// DESCENDING through AliasChars (last index first), which for this
// alphabet means z, y, x, ..., a - exactly process_variable's own
// `reversed(self.des.c_var_chr)` (badig_zx81.py:350).
var variableCategory = core.VariableCategory{
	Name:       "zx81-variable",
	AliasChars: variableAliasChars,
	AliasLen:   variableAliasLen,
}

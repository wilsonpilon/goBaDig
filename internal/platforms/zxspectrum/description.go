package zxspectrum

import "github.com/wilsonpilon/goBaDig/internal/core"

// Vocabulary tables, transcribed verbatim (same entries, same order
// where order matters) from badig/zxspectrum/badig_zxspectrum.py's
// Description class (__init__, read in full this session).
//
// These are DIGNIFIED-LEVEL tables (what the Lexer recognizes in
// source text) - a different, earlier stage than the tokenizer
// (tokenizer.go, Etapa 1), which encodes already-generated Classic
// ASCII text into bytes. The two don't fully coincide: `instructions`
// here includes Dignified-only entries ("REMX", "?") that are never
// ROM keywords, and the tokenizer's own `keywords`/`supported`
// restricts which of the REAL ROM words this module's 48K/128K subset
// accepts - two different concerns, deliberately not merged.
var (
	// instructions mirrors badig_zxspectrum.py:59-67 (51 words).
	instructionWords = []string{
		"PRINT", "LET", "IF", "CLS", "STOP", "BORDER", "INK",
		"PAPER", "PAUSE", "BEEP", "INPUT", "RETURN", "FOR", "NEXT",
		"TO", "STEP", "DIM", "DATA", "READ", "RESTORE", "RANDOMIZE",
		"AT", "TAB", "FLASH", "BRIGHT", "INVERSE", "OVER",
		"PLOT", "DRAW", "CIRCLE", "POKE", "CLEAR",
		"RUN", "LIST", "NEW", "CONTINUE", "OUT", "REMX", "?",
		"DEF FN", "LOAD", "SAVE", "VERIFY", "MERGE", "LINE",
		"OPEN #", "CLOSE #", "LPRINT", "LLIST", "COPY",
		"PLAY", "SPECTRUM",
	}

	// jumpWords mirrors badig_zxspectrum.py:68 - unlike MSX/ZX81, BOTH
	// spellings of GOTO/GOSUB are listed explicitly: "GO TO"/"GO SUB"
	// (with a space) are Lexer-level vocabulary here, not just a
	// tokenizer concern.
	jumpWords = []string{"GOTO", "GO TO", "GOSUB", "GO SUB", "THEN"}

	// operatorWords mirrors badig_zxspectrum.py:69 - BIN is an operator
	// here (a binary-literal prefix), unlike MSX/ZX81.
	operatorWords = []string{"AND", "OR", "NOT", "BIN"}

	// symbolWords mirrors badig_zxspectrum.py:70-71, in the SAME order
	// (longest/most-specific first matters: "<=">">="/"<>" must be
	// checked before their single-character prefixes "<"/">"). Includes
	// `'` (apostrophe) - confirmed empirically (running badig.py) to be
	// the real Sinclair BASIC PRINT-list "new line" separator
	// (`PRINT "A"'"B"`), a plain symbol with no Dignified-specific
	// meaning - it passes straight through untouched.
	symbolWords = []string{
		"<=", ">=", "<>", ">", "=", "<", "+", "*", "/", "^",
		"-", ",", "(", ")", ";", "'", "#",
	}
)

// Token kinds produced by the ZX Spectrum Lexer.
const (
	KindOperator       core.TokenKind = "SPECTRUM_OPERATOR"
	KindSymbol         core.TokenKind = "SPECTRUM_SYMBOL"
	KindExtendedSymbol core.TokenKind = "SPECTRUM_EXTENDED_SYMBOL"
	KindFunction       core.TokenKind = "SPECTRUM_FUNCTION"
	KindJump           core.TokenKind = "SPECTRUM_JUMP"
	KindInstruction    core.TokenKind = "SPECTRUM_INSTRUCTION"
	KindNumber         core.TokenKind = "SPECTRUM_NUMBER"
	KindComment        core.TokenKind = "SPECTRUM_COMMENT"
	KindSeparator      core.TokenKind = "SPECTRUM_SEPARATOR"
	// KindHexByte is a Dignified-exclusive `\{XX}` raw-byte escape -
	// same mechanism as ZX81's own (ver zx81/lexer.go), decoded later
	// by the tokenizer (tokenizer.go, already ported in Etapa 1).
	KindHexByte core.TokenKind = "SPECTRUM_HEXBYTE"
)

// extendedSymbols mirrors c_e_symb (badig_zxspectrum.py:96): compound
// assignment/increment/decrement operators, Dignified-only syntax.
// Must be checked BEFORE symbolWords, same reasoning as MSX/ZX81's own
// extendedSymbols (ver msx/lexer.go, zx81/lexer.go).
var extendedSymbols = []string{"++", "--", "+=", "-=", "*=", "/=", "^="}

// description implements core.Description for ZX Spectrum.
type description struct{}

// Description is the ZX Spectrum core.Description singleton.
var Description core.Description = description{}

func (description) SysInfo() core.SystemInfo {
	return core.SystemInfo{
		Name:         "ZX Spectrum",
		DignifiedExt: ".dzs",
		ClassicExt:   ".azs",
		BinaryExt:    ".tap",
	}
}

func (description) QuoteChar() byte { return '"' }

// StatementSeparator mirrors c_instsp (badig_zxspectrum.py's
// classic_groups, `^:$`) - same as MSX/ZX81.
func (description) StatementSeparator() string { return ":" }

// LoopBackInstruction mirrors c_loop_back (badig_zxspectrum.py:84) -
// same as MSX/ZX81.
func (description) LoopBackInstruction() string { return "goto" }

// CommentChar mirrors c_altrem (badig_zxspectrum.py:74): "REM", the
// WHOLE WORD - same as ZX81, unlike MSX's single punctuation `'`
// (Spectrum's own `'` is real Sinclair BASIC vocabulary, the PRINT-list
// separator - see symbolWords' own doc comment - never a comment
// marker).
func (description) CommentChar() string { return "REM" }

// AssignmentOperator mirrors c_equal, inherited unchanged from
// MSXDescription (badig_zxspectrum.py's Description never overrides
// it) - same value as MSX/ZX81.
func (description) AssignmentOperator() string { return "=" }

// FunctionCallInstruction mirrors c_func_call, inherited unchanged
// from MSXDescription - same as MSX/ZX81.
func (description) FunctionCallInstruction() string { return "gosub" }

// FunctionReturnInstruction mirrors c_func_ret, inherited unchanged
// from MSXDescription - same as MSX/ZX81.
func (description) FunctionReturnInstruction() string { return "return" }

// FunctionCallBoundaryWords mirrors c_func_stop_kw
// (badig_zxspectrum.py:85: "self.c_reserved_kw") - unlike MSX/ZX81,
// which build a SMALLER list excluding operators, Spectrum reuses its
// ENTIRE ReservedKeywords() list unchanged.
func (description) FunctionCallBoundaryWords() []string {
	return description{}.ReservedKeywords()
}

// VariableAliasLength mirrors c_var_valid_chars (badig_zxspectrum.py:
// 83): present in the original but NOT used for any alias allocation
// here (confirmed reading process_variable's absence - Spectrum has no
// allocator at all, ver parser.go's own doc comment on
// ValidateIdentifier). Kept at 1 only to satisfy this interface; no
// Pass in this package reads it.
func (description) VariableAliasLength() int { return 1 }

// ReservedKeywords mirrors c_reserved_kw (badig_zxspectrum.py:72:
// "instructions + jumps + operators + list(FUNCTIONS) + ['REM']").
// FUNCTIONS is the tokenizer's own list (functions, tokenizer.go) -
// reused directly rather than re-transcribed, since it's the exact
// same Python tuple imported by badig_zxspectrum.py itself
// (`from .tokenizer_zxspectrum import FUNCTIONS`).
func (description) ReservedKeywords() []string {
	words := make([]string, 0, len(instructionWords)+len(jumpWords)+len(operatorWords)+len(functions)+1)
	words = append(words, instructionWords...)
	words = append(words, jumpWords...)
	words = append(words, operatorWords...)
	words = append(words, functions...)
	words = append(words, "REM")
	return words
}

// CompoundSymbols mirrors c_symb_comp (badig_zxspectrum.py:80: "symbols
// + [':', '_']") - unlike MSX, nothing is removed from symbols here
// (no "PRINT #1 needs a space after PRINT" concern at this level - the
// Spectrum tokenizer, Etapa 1, handles that split on its own, at the
// BYTE-encoding stage, not here).
func (description) CompoundSymbols() []string {
	words := make([]string, 0, len(symbolWords)+2)
	words = append(words, symbolWords...)
	words = append(words, ":", "_")
	return words
}

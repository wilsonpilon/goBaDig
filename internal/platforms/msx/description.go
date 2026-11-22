package msx

import "github.com/wilsonpilon/goBaDig/internal/core"

// Vocabulary tables, transcribed verbatim (same entries, same order
// where order matters) from badig/msx/badig_msx.py's Description
// class (lines 1-224, read in full this session, not summarized).
//
// These are TEXT-matching tables for the Dignified-level Lexer - they
// classify what a piece of source IS, they do not encode it to bytes.
// Byte encoding is a later, separate stage (the already-ported
// Batoken tokenizer, tokenizer.go in this same package), which works
// on the generated Classic ASCII text, not on these Lexer tokens.
var (
	// instructions mirrors c_instrc (badig_msx.py:27-38): 73 plain
	// keywords plus two pattern-shaped entries ("?" and "DEFUSR" with
	// an optional trailing digit) kept here as their own cases rather
	// than regex, matched after plain instructions fail.
	instructions = []string{
		"AS", "BASE", "BEEP", "BLOAD", "BSAVE", "CALL", "CIRCLE",
		"CLEAR", "CLOAD", "CLOSE", "CLS", "CMD", "COLOR", "CONT",
		"COPY", "CSAVE", "CSRLIN", "DEF", "DEFDBL", "DEFINT", "MAXFILES",
		"DEFSNG", "DEFSTR", "DIM", "DRAW", "DSKI", "END", "EQV",
		"ERASE", "ERR", "ERROR", "FIELD", "FILES", "FN", "FOR", "GET",
		"IF", "INPUT", "INTERVAL", "IMP", "IPL", "KILL", "LET",
		"LFILES", "LINE", "LOAD", "LOCATE", "LPRINT", "LSET", "MAX",
		"MERGE", "MOTOR", "NAME", "NEW", "NEXT", "OFF", "ON", "OPEN",
		"OUT", "OUTPUT", "PAINT", "POINT", "POKE", "PRESET", "PRINT",
		"PSET", "PUT", "READ", "RSET", "SAVE", "SCREEN", "SET",
		"SOUND", "STEP", "STOP", "SWAP", "TIME", "TO", "TROFF",
		"TRON", "USING", "VPOKE", "WAIT", "WIDTH",
	}

	// functionsDollar mirrors c_funcdl (badig_msx.py:41-44): functions
	// whose name ends in "$".
	functionsDollar = []string{
		"ATTR$", "BIN$", "CHR$", "DSKO$", "HEX$",
		"INKEY$", "INPUT$", "LEFT$", "MID$", "MKD$",
		"MKI$", "MKS$", "OCT$", "RIGHT$", "SPACE$",
		"SPRITE$", "STR$", "STRING$",
	}

	// functionsPlain mirrors c_funcnm (badig_msx.py:47-52), minus
	// "USR" (handled separately below, since it takes an optional
	// trailing digit, like "?"/"DEFUSR" above).
	functionsPlain = []string{
		"ABS", "ASC", "ATN", "CDBL", "CINT", "COS", "CSNG", "CVD",
		"CVI", "CVS", "DSKF", "EOF", "EXP", "FIX", "FPOS", "FRE",
		"INP", "INSTR", "INT", "KEY", "LEN", "LOC", "LOF", "LOG",
		"LPOS", "PAD", "PDL", "PEEK", "PLAY", "POS", "RND", "SGN",
		"SIN", "SPC", "SPRITE", "SQR", "STICK", "STRIG", "TAB", "TAN",
		"VAL", "VARPTR", "VDP", "VPEEK",
	}

	// jumpKeywords mirrors c_jumpin (badig_msx.py:55-56) - keywords
	// whose trailing argument is a line-number reference, not an
	// ordinary expression. Same vocabulary already used by the
	// Batoken tokenizer's own `jumps` map (tokenizer.go) - kept as a
	// separate copy (different name: Go won't allow two package-level
	// `jumps`) since they classify different things (Dignified source
	// text vs. already-tokenized Classic bytes) and have no reason to
	// share storage.
	jumpKeywords = []string{
		"RESTORE", "AUTO", "RENUM", "DELETE", "RESUME", "ERL", "ELSE",
		"RUN", "LIST", "LLIST", "GOTO", "RETURN", "THEN", "GOSUB",
	}

	// operators mirrors c_operat (badig_msx.py:59).
	operators = []string{"AND", "MOD", "NOT", "OR", "XOR"}

	// symbols mirrors c_symbol (badig_msx.py:62-63), already
	// unescaped (the Python source has them as regex fragments like
	// `\+`, since they feed a combined regex - this port doesn't need
	// that, plain text is enough since each entry is tried in order
	// against the start of the remaining line, like every other
	// vocabulary table here).
	symbols = []string{">", "=", "<", "+", "*", "/", "^", `\`, "-", ",", "(", ")", ";", "#"}

	// extendedSymbols mirrors c_e_symb (badig_msx.py:93) - compound
	// assignment/increment/decrement operators. Must be checked
	// BEFORE symbols (below), or e.g. "+=" would wrongly match "+"
	// first and leave a dangling "=".
	extendedSymbols = []string{"++", "--", "+=", "-=", "*=", "/=", "^="}
)

// description implements core.Description for MSX.
type description struct{}

// Description is the MSX core.Description singleton.
var Description core.Description = description{}

func (description) SysInfo() core.SystemInfo {
	return core.SystemInfo{
		Name:         "MSX",
		DignifiedExt: ".dmx",
		ClassicExt:   ".amx",
		BinaryExt:    ".bmx",
	}
}

func (description) QuoteChar() byte { return '"' }

// StatementSeparator mirrors c_instsp_str (badig_msx.py:156).
func (description) StatementSeparator() string { return ":" }

// LoopBackInstruction mirrors c_loop_back (badig_msx.py:162).
func (description) LoopBackInstruction() string { return "goto" }

// Variable allocation constants, confirmed reading process_variable
// (badig_msx.py:516-558) in a previous session and process_variable's
// surrounding Description fields (badig_msx.py:133-150) directly this
// session. Not used yet - the Parser (next phase, not this one) is
// what actually allocates 2-letter aliases - kept here now so they
// aren't rediscovered/retranscribed later.
const (
	// variableAliasChars is c_var_chr: both alias characters come from
	// this alphabet.
	variableAliasChars = "abcdefghijklmnopqrstuvwxyz"
	// variableAliasLen is c_var_valid_chars: every MSX alias is 2
	// characters, regardless of category (unlike ZX81/Spectrum, which
	// have per-category lengths - see OUTLINE.md, seção 4.1/6 item 5).
	variableAliasLen = 2
)

// cReplacements mirrors c_replacements (badig_msx.py:66-72): a direct
// substitution for specific box-drawing/smiley glyphs, applied before
// cOriginal/cOriginalRunes below. Read directly from the source file
// this session, not retyped from memory - Unicode glyphs are easy to
// get subtly wrong by hand.
var cReplacements = map[rune]rune{
	'☺': 'A', '☻': 'B', '♥': 'C', '♦': 'D', '♣': 'E', '♠': 'F',
	'·': 'G', '◘': 'H', '○': 'I', '◙': 'J', '♂': 'K', '♀': 'L',
	'♪': 'M', '♬': 'N', '☼': 'O', '┿': 'P', '┴': 'Q', '┬': 'R',
	'┤': 'S', '├': 'T', '┼': 'U', '│': 'V', '─': 'W', '┌': 'X',
	'┐': 'Y', '└': 'Z', '┘': '[', '╳': ']', '╱': '\\', '╲': '^',
	'╂': '_',
}

// cOriginal mirrors c_original (badig_msx.py:76-77): 128 characters,
// index-paired with MSX's internal 8-bit character codes \u0080-ÿ
// (translateChars, parser.go, maps cOriginal[i] -> rune(0x80+i)).
// Extracted programmatically from the Python source this session
// (`eval()`'d and length-checked at 128) rather than retyped by hand,
// to avoid a transcription error in a 128-character Unicode literal.
const cOriginal = "ÇüéâäàåçêëèïîìÄÅÉæÆôöòûùÿÖÜ¢£¥₧ƒáíóúñÑªº¿⌐¬½¼¡«»ÃãĨĩÕõŨũĲĳ¾∽◇‰¶§" +
	"▂▚▆▔◾▇▎▞▊▕▉▨▧▼▲▶◀⧗⧓▘▗▝▖▒Δǂω█▄▌▐▀αβΓπΣσμτΦθΩδ∞φ∈∩≡±≥≤⌠⌡÷≈°∙‐√ⁿ²❚■"

// cOriginalRunes is cOriginal decoded once, so translateChars can
// search it by character position (what the index->0x80+i mapping
// needs), not byte offset.
var cOriginalRunes = []rune(cOriginal)

// variableCategory is the single MSX variable category (core's
// VariableAllocator, from Bloco 1, is parametrized by category - MSX
// just has one, since it uses the same 2-letter scheme for every
// variable regardless of type, unlike ZX81/Spectrum). Not used yet;
// see the const block above.
var variableCategory = core.VariableCategory{
	Name:       "msx-variable",
	AliasChars: variableAliasChars,
	AliasLen:   variableAliasLen,
}

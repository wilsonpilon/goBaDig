package edit

// Keyword/function tables below are distilled from this project's own
// reference material for each dialect's existing Sublime Package/
// VSCode Extension syntax grammars (badig/<dialect>/{Sublime Package,
// VSCode Extension}) - NOT re-derived from the compiler's own lexer
// (internal/platforms/*), on purpose: those lexers are built for a
// correctness-first, whole-program compile pass and aren't meant to
// run on every keystroke against arbitrary, possibly mid-edit,
// possibly invalid text. These tables are deliberately simpler than
// the full grammars (no multi-word statement forms like "SET ADJUST"
// kept as one unit, no distinct "invalid input" detection, no
// separate operator-word vs. keyword distinction) - each dialect's
// own compound statement words (GO, TO, SUB, SET, DEF, etc.) are
// listed individually, so e.g. "GO TO" still highlights as two
// keyword-colored words even though it isn't tracked as one token.

var msxKeywords = wordSet(
	"AND", "AS", "AUTO", "BEEP", "BIN", "BLOAD", "BSAVE", "CIRCLE", "CLEAR",
	"CLOAD", "CLOSE", "CLS", "CMD", "COLOR", "CONT", "COPY", "CSAVE", "CSRLIN",
	"DATA", "DECLARE", "DEF", "DEFDBL", "DEFINE", "DEFINT", "DEFSNG", "DEFSTR",
	"DELETE", "DIM", "DRAW", "DSKO", "ELSE", "END", "ENDIF", "EQV", "ERASE",
	"ERROR", "EXIT", "FALSE", "FIELD", "FILES", "FOR", "FUNC", "GET", "GO",
	"GOSUB", "GOTO", "IF", "IMP", "INCLUDE", "INPUT", "IPL", "KANJI", "KEEP",
	"KEY", "KILL", "LET", "LINE", "LIST", "LLIST", "LOAD", "LOCATE", "LPRINT",
	"LSET", "MAXFILES", "MERGE", "MOD", "MOTOR", "NAME", "NEW", "NEXT", "NOT",
	"OFF", "ON", "OPEN", "OR", "OUT", "OUTPUT", "PAINT", "PLAY", "POINT",
	"POKE", "PRESET", "PRINT", "PSET", "PUT", "READ", "RENUM", "RESTORE",
	"RESUME", "RET", "RETURN", "RSET", "RUN", "SAVE", "SCREEN", "SET", "SOUND",
	"SPRITE", "STEP", "STOP", "SUB", "SWAP", "THEN", "TO", "TROFF", "TRON",
	"TRUE", "USING", "VPOKE", "WAIT", "WIDTH", "XOR",
)

var msxFunctions = wordSet(
	"ABS", "ASC", "ATN", "BASE", "CDBL", "CINT", "COS", "CSNG", "CVD", "CVI",
	"CVS", "DATE", "DSKF", "EOF", "ERL", "ERR", "EXP", "FIX", "FN", "FPOS",
	"FRE", "INP", "INSTR", "INTERVAL", "INT", "LEN", "LOC", "LOF", "LOG",
	"LPOS", "PAD", "PDL", "PEEK", "POS", "RND", "SGN", "SIN", "SPC", "SQR",
	"STICK", "STRIG", "TAB", "TAN", "TIME", "VAL", "VARPTR", "VDP", "VPEEK",
	"ATTR$", "BIN$", "CHR$", "HEX$", "INKEY$", "INPUT$", "LEFT$", "MID$",
	"MKD$", "MKI$", "MKS$", "OCT$", "RIGHT$", "SPACE$", "STR$", "STRING$",
)

var zx81Keywords = wordSet(
	"AND", "CLEAR", "CLS", "CONT", "CONTINUE", "COPY", "DECLARE", "DEFINE",
	"DIM", "ENDIF", "EXIT", "FALSE", "FAST", "FOR", "FUNC", "GO", "GOSUB",
	"GOTO", "IF", "INCLUDE", "INPUT", "KEEP", "LET", "LIST", "LLIST", "LOAD",
	"LPRINT", "NEW", "NEXT", "NOT", "OR", "PAUSE", "PLOT", "POKE", "PRINT",
	"RAND", "RANDOMIZE", "REMX", "RET", "RETURN", "RUN", "SAVE", "SCROLL",
	"SLOW", "STEP", "STOP", "THEN", "TO", "TRUE", "UNPLOT",
)

var zx81Functions = wordSet(
	"ABS", "ACS", "ASN", "AT", "ATN", "CHR$", "CODE", "COS", "EXP", "INKEY$",
	"INT", "LEN", "LN", "PEEK", "PI", "RND", "SGN", "SIN", "SQR", "STR$",
	"TAB", "TAN", "USR", "VAL",
)

var zxSpectrumKeywords = wordSet(
	"AND", "AT", "BEEP", "BIN", "BORDER", "BRIGHT", "CIRCLE", "CLEAR", "CLOSE",
	"CLS", "CONTINUE", "COPY", "DATA", "DECLARE", "DEF", "DEFINE", "DIM",
	"DRAW", "ENDIF", "EXIT", "FALSE", "FLASH", "FOR", "FUNC", "GO", "GOSUB",
	"GOTO", "IF", "INCLUDE", "INK", "INPUT", "INVERSE", "KEEP", "LET", "LINE",
	"LIST", "LLIST", "LOAD", "LPRINT", "MERGE", "NEW", "NEXT", "NOT", "OPEN",
	"OR", "OUT", "OVER", "PAPER", "PAUSE", "PLAY", "PLOT", "POKE", "PRINT",
	"RANDOMIZE", "READ", "REMX", "RESTORE", "RET", "RETURN", "RUN", "SAVE",
	"SPECTRUM", "STEP", "STOP", "THEN", "TO", "TRUE", "VERIFY",
)

var zxSpectrumFunctions = wordSet(
	"ABS", "ACS", "ASN", "ATN", "ATTR", "CHR$", "CODE", "COS", "EXP", "FN",
	"IN", "INKEY$", "INT", "LEN", "LN", "PEEK", "PI", "POINT", "RND",
	"SCREEN$", "SGN", "SIN", "SQR", "STR$", "TAN", "USR", "VAL", "VAL$",
)

func wordSet(words ...string) map[string]bool {
	set := make(map[string]bool, len(words))
	for _, w := range words {
		set[w] = true
	}
	return set
}

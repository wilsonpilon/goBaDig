package msx

import (
	"regexp"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// Token kinds produced by the MSX Lexer. These classify Dignified
// source text (badig_msx.py's Description/classic_groups) - a
// different, earlier stage than the already-ported Batoken tokenizer
// (tokenizer.go), which encodes already-generated Classic ASCII text
// into bytes. A KindNumber token here just carries the matched text
// (e.g. "3.14"); nothing is byte-encoded until the tokenizer runs,
// much later in the pipeline.
const (
	KindOperator    core.TokenKind = "MSX_OPERATOR"
	KindSymbol      core.TokenKind = "MSX_SYMBOL"
	KindFunction    core.TokenKind = "MSX_FUNCTION"
	KindJump        core.TokenKind = "MSX_JUMP"
	KindInstruction core.TokenKind = "MSX_INSTRUCTION"
	KindNumber      core.TokenKind = "MSX_NUMBER"
	KindComment     core.TokenKind = "MSX_COMMENT"
	KindData        core.TokenKind = "MSX_DATA"
	KindSeparator   core.TokenKind = "MSX_SEPARATOR"
)

// Lexer implements core.Lexer for MSX.
//
// Deliberately out of scope this round (none exercised by the test
// below; all are real MSX features, not bugs, just not needed yet to
// validate the core.Lexer "v1" contract for real - see OUTLINE.md,
// Bloco 2):
//   - The "”" multi-line comment block (c_o_brem) - spans several
//     source lines, which needs its own state across Lex calls.
//   - DATA's comma-separated literal-item semantics (get_data_line in
//     the original engine) - "DATA" itself is recognized as its own
//     token kind, but its arguments are lexed by the normal rules
//     below instead of as raw literal text. Close enough to validate
//     the contract; the real echoing behavior is Parser-stage work.
//   - Unicode-to-ASCII glyph translation (c_replacements/c_original/
//     c_translat) - that is Parser Pass 5 work in the original engine,
//     not the Lexer's.
type Lexer struct{}

func (Lexer) Init(*core.CompileContext) {}

func (Lexer) Lex(_ *core.CompileContext, remaining string) (int, []core.Token, bool, error) {
	if remaining == "" {
		return 0, nil, false, nil
	}

	// Extended (2-char) symbols MUST be checked before their
	// single-char prefixes, or e.g. "+=" would wrongly match "+" and
	// leave a dangling "=". (badig_msx.py's classic_groups lists
	// plain symbols before extended ones, but that is a single
	// whole-nugget-match regex, not prefix matching on raw text like
	// this port does - the two are not directly comparable; this
	// order is the one that is actually correct for prefix matching.)
	for _, s := range extendedSymbols {
		if strings.HasPrefix(remaining, s) {
			return len(s), one(KindSymbol, s), true, nil
		}
	}
	for _, s := range symbols {
		// "#" immediately followed by a letter is a shared Dignified
		// toggle-rem tag ("#a", "#none", "#all" - badig_dignified.py's
		// d_togrem), not MSX's own "#" symbol (file channel refs like
		// "#1" in PRINT/OPEN - confirmed used that way in
		// badig/examples/msx/CGK173.dmx). Refusing here lets core.Lex's
		// own shared fallback claim the whole "#word" instead; "#1"
		// etc. still falls through to the plain match right below.
		if s == "#" && len(remaining) > 1 && isLetter(remaining[1]) {
			continue
		}
		if strings.HasPrefix(remaining, s) {
			return len(s), one(KindSymbol, s), true, nil
		}
	}

	switch remaining[0] {
	case '\'':
		return len(remaining), one(KindComment, remaining), true, nil
	case ':':
		return 1, one(KindSeparator, ":"), true, nil
	case '?':
		return 1, one(KindInstruction, "?"), true, nil
	case '&':
		if n, tok, ok := scanBaseLiteral(remaining); ok {
			return n, []core.Token{tok}, true, nil
		}
	}

	if n, tok, ok := scanDecimalNumber(remaining); ok {
		return n, []core.Token{tok}, true, nil
	}

	if isLetter(remaining[0]) {
		if n, tok, ok := scanUSRLike(remaining); ok {
			return n, []core.Token{tok}, true, nil
		}

		word := scanKeywordCandidate(remaining)
		upper := strings.ToUpper(word)
		switch {
		case upper == "REM":
			return len(remaining), one(KindComment, remaining), true, nil
		case upper == "DATA":
			return len(word), one(KindData, word), true, nil
		case contains(operators, upper):
			return len(word), one(KindOperator, word), true, nil
		case contains(functionsDollar, upper), contains(functionsPlain, upper):
			return len(word), one(KindFunction, word), true, nil
		case contains(jumpKeywords, upper):
			return len(word), one(KindJump, word), true, nil
		case contains(instructions, upper):
			return len(word), one(KindInstruction, word), true, nil
		}
		// Not a keyword: let the driver's generic fallback take over.
		// It only recognizes a lowercase-starting run as a Dignified
		// identifier - an uppercase word that isn't a keyword falls
		// through all the way to a single raw character, same
		// tolerant behavior as the original engine's "character not
		// recognized" diagnostic path.
		return 0, nil, false, nil
	}

	return 0, nil, false, nil
}

func one(kind core.TokenKind, value string) []core.Token {
	return []core.Token{{Kind: kind, Value: value}}
}

func isLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// scanKeywordCandidate extracts a run of letters, plus one trailing
// "$" if present (so function names like "ATTR$" are captured whole).
// It deliberately does NOT consume trailing digits - that is handled
// separately by scanUSRLike, so an ordinary keyword immediately
// followed by a digit (unusual, but not a case to silently mismatch)
// isn't accidentally absorbed into one word.
func scanKeywordCandidate(s string) string {
	i := 0
	for i < len(s) && isLetter(s[i]) {
		i++
	}
	if i < len(s) && s[i] == '$' {
		i++
	}
	return s[:i]
}

// scanUSRLike handles the two instructions/functions whose name may
// carry a single trailing device-number digit (badig_msx.py:38,52:
// `DEFUSR(\d)?`, `USR(\d)?`).
func scanUSRLike(remaining string) (int, core.Token, bool) {
	for base, kind := range map[string]core.TokenKind{"DEFUSR": KindInstruction, "USR": KindFunction} {
		if len(remaining) < len(base) || !strings.EqualFold(remaining[:len(base)], base) {
			continue
		}
		n := len(base)
		if n < len(remaining) && isDigit(remaining[n]) {
			n++
		}
		return n, core.Token{Kind: kind, Value: remaining[:n]}, true
	}
	return 0, core.Token{}, false
}

var (
	hexLiteralPattern    = regexp.MustCompile(`(?i)^&H[0-9a-f]+`)
	octLiteralPattern    = regexp.MustCompile(`(?i)^&O[0-7]+`)
	binLiteralPattern    = regexp.MustCompile(`(?i)^&B[01]+`)
	decimalNumberPattern = regexp.MustCompile(`(?i)^(\d*\.?)?\d+([de][+-]\d+)?[%#!]?`)
)

// scanBaseLiteral recognizes &H/&O/&B literals. Mirrors c_numhex/
// c_numoct/c_numbin (badig_msx.py:105-109) - at this Lexer stage it
// only needs to find the TEXT EXTENT; byte encoding (and the real
// &B-doesn't-actually-compute-anything discrepancy) is the Batoken
// tokenizer's job, already handled in tokenizer.go.
func scanBaseLiteral(remaining string) (int, core.Token, bool) {
	for _, re := range []*regexp.Regexp{hexLiteralPattern, octLiteralPattern, binLiteralPattern} {
		if m := re.FindString(remaining); m != "" {
			return len(m), core.Token{Kind: KindNumber, Value: m}, true
		}
	}
	return 0, core.Token{}, false
}

// scanDecimalNumber mirrors c_numdec (badig_msx.py:103): an optional
// integer part with an optional decimal point, a mandatory digit run,
// an optional D/E exponent, and an optional %/#/! type suffix.
func scanDecimalNumber(remaining string) (int, core.Token, bool) {
	m := decimalNumberPattern.FindString(remaining)
	if m == "" {
		return 0, core.Token{}, false
	}
	return len(m), core.Token{Kind: KindNumber, Value: m}, true
}

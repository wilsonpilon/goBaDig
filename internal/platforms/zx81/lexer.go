package zx81

import (
	"regexp"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// Token kinds produced by the ZX81 Lexer. These classify Dignified
// source text (badig_zx81.py's Description/classic_groups) - a
// different, earlier stage than the already-ported tokenizer
// (tokenizer.go), which encodes already-generated Classic ASCII text
// into bytes.
const (
	KindOperator    core.TokenKind = "ZX81_OPERATOR"
	KindSymbol      core.TokenKind = "ZX81_SYMBOL"
	KindFunction    core.TokenKind = "ZX81_FUNCTION"
	KindJump        core.TokenKind = "ZX81_JUMP"
	KindInstruction core.TokenKind = "ZX81_INSTRUCTION"
	KindNumber      core.TokenKind = "ZX81_NUMBER"
	KindComment     core.TokenKind = "ZX81_COMMENT"
	KindSeparator   core.TokenKind = "ZX81_SEPARATOR"
	// KindHexByte is a Dignified-exclusive `\{XX}` raw-byte escape
	// (badig_zx81.py's own Lexer.lexing override, "C_HEXBYT") - lets
	// source text inject one arbitrary byte directly (into a REM, a
	// variable name, anywhere), decoded later by the tokenizer
	// (tokenizer.go's own `\{XX}` handling, already ported in Stage 1)
	// - this Lexer only needs to recognize and pass the escape through
	// verbatim, not decode it.
	KindHexByte core.TokenKind = "ZX81_HEXBYTE"
)

// Lexer implements core.Lexer for ZX81.
//
// Deliberately out of scope this round (real ZX81 features, not
// bugs - see OUTLINE.md, Bloco 3):
//   - "PRINT AT"'s own `[?](row,col)` builtin-define expansion and the
//     "LET" implicit-insertion rule in generate() - both Parser-stage
//     work (Stage 3 of the roadmap), not the Lexer's.
//   - remx (reserve raw bytes inside a REM line) - also Parser-stage
//     (a Dignified-level directive, Stage 3), though the tokenizer's
//     own remx handling (Stage 1) already exists independently.
//   - Numbered-listing detection/renumbering (dignify_numbered) -
//     explicitly deferred, documented in the plan's roadmap as a
//     later, optional Etapa 6.
type Lexer struct{}

func (Lexer) Init(*core.CompileContext) {}

func (Lexer) Lex(_ *core.CompileContext, remaining string, _ bool) (int, []core.Token, bool, error) {
	if remaining == "" {
		return 0, nil, false, nil
	}

	// Extended (2-char) Dignified-only compound operators MUST be
	// checked before any single-char symbol prefix they share, same
	// reasoning as MSX's own extendedSymbols (ver msx/lexer.go).
	for _, s := range extendedSymbols {
		if strings.HasPrefix(remaining, s) {
			return len(s), one(KindSymbol, s), true, nil
		}
	}

	switch remaining[0] {
	case '"':
		n, tok := scanZX81String(remaining)
		return n, []core.Token{tok}, true, nil
	case '\\':
		if strings.HasPrefix(remaining, `\{`) {
			n, tok, err := scanHexByte(remaining)
			if err != nil {
				return 0, nil, false, err
			}
			return n, []core.Token{tok}, true, nil
		}
	case ':':
		return 1, one(KindSeparator, ":"), true, nil
	case '?':
		return 1, one(KindInstruction, "?"), true, nil
	case '#':
		// "#" immediately followed by a letter or another "#" is
		// shared Dignified vocabulary (a KEEP tag, a toggle-block
		// marker, "##"/"###") - refusing here lets core.Lex's own
		// fallback claim it, same guard MSX's own lexer uses for the
		// same reason (ver msx/lexer.go).
		if len(remaining) > 1 && (isAlpha(remaining[1]) || remaining[1] == '#') {
			return 0, nil, false, nil
		}
	case '+', '-':
		// A leading sign merges into a following PLAIN DIGIT RUN only
		// - confirmed empirically against the real badig.py (ver
		// numberBytesWithSign's own doc comment for the full set of
		// cases this was checked against): "y-5"/"PRINT -5" -> ONE
		// number token "-5", but "-5.5" -> "-5" then ".5" as TWO
		// tokens (the sign does NOT extend through a decimal point or
		// an exponent the way an UNSIGNED number does), and "-.5" ->
		// "-" (plain symbol) then ".5" (the sign doesn't merge at all
		// when a digit isn't immediately next). This is quirky enough
		// that it's deliberately NOT generalized from numberPattern
		// (below) - it is its own, narrower pattern.
		if m := signedIntegerPattern.FindString(remaining); m != "" {
			return len(m), one(KindNumber, m), true, nil
		}
	}

	for _, s := range symbols {
		if strings.HasPrefix(remaining, s) {
			return len(s), one(KindSymbol, s), true, nil
		}
	}

	if m := numberPattern.FindString(remaining); m != "" {
		return len(m), one(KindNumber, m), true, nil
	}

	if isAlpha(remaining[0]) {
		word := scanKeywordCandidate(remaining)
		upper := strings.ToUpper(word)
		// Same guard as MSX's own lexer: a keyword match is only
		// complete if what immediately follows isn't itself a valid
		// identifier-continuation character (ver msx/lexer.go's own
		// doc comment on "copy_VRAM"/"color1" for the real bugs this
		// guard fixes).
		if len(word) >= len(remaining) || !isIdentifierContinuation(remaining[len(word)]) {
			switch {
			case upper == "REM":
				return len(remaining), one(KindComment, remaining), true, nil
			case contains(operators, upper):
				return len(word), one(KindOperator, word), true, nil
			case contains(functionsDollar, upper), contains(functionsPlain, upper):
				return len(word), one(KindFunction, word), true, nil
			case contains(jumpKeywords, upper):
				return len(word), one(KindJump, word), true, nil
			case contains(instructions, upper):
				return len(word), one(KindInstruction, word), true, nil
			}
		}
		// Not a keyword: let the driver's generic fallback take over
		// (a Dignified identifier, same tolerant behavior as MSX).
		return 0, nil, false, nil
	}

	return 0, nil, false, nil
}

func one(kind core.TokenKind, value string) []core.Token {
	return []core.Token{{Kind: kind, Value: value}}
}

// isIdentifierContinuation reports whether c could extend an
// identifier that already started before it - ver msx/lexer.go's own
// doc comment for the real bugs this guards against. isAlpha/isDigit
// are tokenizer.go's own byte-classification helpers, reused here
// rather than redeclared.
func isIdentifierContinuation(c byte) bool {
	return isAlpha(c) || isDigit(c) || c == '_'
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
// "$" if present (so function names like "CHR$" are captured whole).
// Does not consume trailing digits - ZX81 has no USR/DEFUSR-style
// device-number suffix to worry about (unlike MSX's own scanUSRLike),
// but the caller's own isIdentifierContinuation check still rejects a
// match followed immediately by a digit, so e.g. "to1" (an ordinary
// identifier starting with the reserved word "TO") still correctly
// falls through to the identifier pattern as one token.
func scanKeywordCandidate(s string) string {
	i := 0
	for i < len(s) && isAlpha(s[i]) {
		i++
	}
	if i < len(s) && s[i] == '$' {
		i++
	}
	return s[:i]
}

var (
	// numberPattern mirrors c_numdec (badig_zx81.py:186) WITHOUT its
	// own leading-sign group - a leading sign is handled separately
	// by signedIntegerPattern (ver Lex's own doc comment on why). No
	// %/#/! type suffix (unlike MSX) - ZX81 doesn't have one; no "D"
	// exponent letter either (only "E").
	numberPattern = regexp.MustCompile(`(?i)^(?:[0-9]*\.?[0-9]+)(?:E[+-]?[0-9]+)?`)
	// signedIntegerPattern is the narrower pattern a leading "+"/"-"
	// actually merges with - confirmed empirically to be JUST a plain
	// digit run, never a decimal point or an exponent (ver Lex's own
	// doc comment).
	signedIntegerPattern = regexp.MustCompile(`^[+-][0-9]+`)
	// hexByteEscape mirrors HEX_ESCAPE_MATCH (tokenizer_zx81.py) - the
	// SAME pattern the tokenizer uses to later decode this escape;
	// here it only needs to recognize/validate the Lexer-level token,
	// same two sides of one mechanism (ver tokenizer.go's own
	// hexEscapeRe).
	hexByteEscape = regexp.MustCompile(`^\\\{([0-9a-fA-F]{2})\}`)
)

// scanHexByte recognizes a `\{XX}` escape, mirroring
// badig_zx81.py's Lexer.lexing override: it scans forward looking for
// a closing "}" (stopping early, as a fatal error, if it runs off the
// end of remaining without finding one - there is no embedded newline
// to worry about, since remaining never crosses a line boundary), then
// validates the WHOLE captured span against the exact `\{XX}` shape
// (exactly 2 hex digits) before accepting it.
func scanHexByte(remaining string) (int, core.Token, error) {
	end := strings.IndexByte(remaining[2:], '}')
	if end < 0 {
		return 0, core.Token{}, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  "Invalid hexadecimal escape; expected \\{XX}.",
		}
	}
	span := remaining[:2+end+1]
	if !hexByteEscape.MatchString(span) {
		return 0, core.Token{}, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  "Invalid hexadecimal escape; expected \\{XX}.",
		}
	}
	return len(span), core.Token{Kind: KindHexByte, Value: span}, nil
}

// scanZX81String claims a Dignified string literal ENTIRELY (unlike
// MSX, which relies on core.Lex's own generic quoted-string fallback,
// since MSX needs no escaping at all) - ZX81 needs its own scanner
// because it supports BOTH `\"` (backslash escape) AND `""` (doubled
// quote, classic BASIC's own convention) as ways to embed a literal
// quote character inside a Dignified string. Mirrors
// badig_zx81.py's Lexer.quoted_string override exactly: a lone
// closing quote ends the string; a DOUBLED quote is consumed as a
// literal escaped quote and scanning continues; an unterminated
// string (no closing quote before the text runs out) is accepted
// tolerantly, consuming everything remaining - same "character not
// recognized"-tolerant spirit as the rest of this Lexer. The returned
// token's Value is the RAW matched text, delimiters and escapes
// included verbatim - nothing is decoded at this stage (same as
// core.Lex's own generic string fallback); decoding happens later, at
// the tokenizer stage (tokenizer.go), which already handles `\"`/`""`
// on the Classic ASCII side.
func scanZX81String(remaining string) (int, core.Token) {
	i := 1
	for i < len(remaining) {
		if remaining[i] == '\\' && i+1 < len(remaining) && remaining[i+1] == '"' {
			i += 2
			continue
		}
		if remaining[i] == '"' {
			if i+1 < len(remaining) && remaining[i+1] == '"' {
				i += 2
				continue
			}
			i++
			break
		}
		i++
	}
	return i, core.Token{Kind: core.KindString, Value: remaining[:i]}
}

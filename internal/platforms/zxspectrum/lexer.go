package zxspectrum

import (
	"regexp"
	"slices"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// jumpWordsSingle/instructionWordsSingle are the subsets of
// jumpWords/instructionWords that are a SINGLE word with no internal
// whitespace - matched by scanning a plain keyword candidate. The
// multi-word forms ("GO TO"/"GO SUB"/"DEF FN"/"OPEN #"/"CLOSE #") are
// matched separately, via compoundRe (tokenizer.go, reused here
// unchanged - same shape at both the Dignified-Lexer and Classic-
// tokenizer levels).
var jumpWordsSingle = []string{"GOTO", "GOSUB", "THEN"}

// Lexer implements core.Lexer for ZX Spectrum - ported from
// badig_zxspectrum.py's Lexer(MSXLexer), read in full this session.
//
// Deliberately NOT implemented this round (ver OUTLINE.md, Bloco 4,
// and parser.go's own doc comment): REMX's own argument parsing
// (Z_REMX/Pass1 in the original) - "REMX" is still recognized as a
// plain instruction word (it's in instructionWords), so a `.dzs`
// source using it fails with a clear "not implemented" diagnostic from
// Pass1, rather than being silently misclassified as a variable name.
type Lexer struct{}

func (Lexer) Init(*core.CompileContext) {}

func (Lexer) Lex(_ *core.CompileContext, remaining string, _ bool) (int, []core.Token, bool, error) {
	if remaining == "" {
		return 0, nil, false, nil
	}

	// Extended (2-char) Dignified-only compound operators MUST be
	// checked before any single-char symbol prefix they share, same
	// reasoning as MSX/ZX81's own extendedSymbols.
	for _, s := range extendedSymbols {
		if strings.HasPrefix(remaining, s) {
			return len(s), one(KindExtendedSymbol, s), true, nil
		}
	}

	switch remaining[0] {
	case '"':
		n, tok, err := scanSpectrumString(remaining)
		if err != nil {
			return 0, nil, false, err
		}
		return n, []core.Token{tok}, true, nil
	case '\\':
		if strings.HasPrefix(remaining, `\{`) {
			n, tok, err := scanHexByte(remaining)
			if err != nil {
				return 0, nil, false, err
			}
			// badig_zxspectrum.py's own Lexer.lexing override: once
			// "\{EA}" (the REM token code) is seen, everything else on
			// the line is raw REM text - confirmed empirically (ver
			// tokenizer_test.go's TestRawBytesAreLiteralInCodeStringsAndRem,
			// a vector taken straight from test_zxspectrum.py), "FN"/
			// quotes/etc. inside it stay completely literal. Returning
			// BOTH tokens from this one Lex call (the raw-byte token
			// itself, then a comment token covering the rest of the
			// line) mirrors the original's own two-step
			// `self.lex.lexer.append(tk)` + `get_lit_line` exactly,
			// without needing core.Lexer's interface to grow anything -
			// it already supports returning several tokens per call.
			if strings.EqualFold(tok.String(), `\{EA}`) {
				rest := remaining[n:]
				if rest == "" {
					return n, []core.Token{tok}, true, nil
				}
				return len(remaining), []core.Token{tok, {Kind: KindComment, Value: rest}}, true, nil
			}
			return n, []core.Token{tok}, true, nil
		}
	case ':':
		return 1, one(KindSeparator, ":"), true, nil
	case '#':
		// "#" immediately followed by a letter or another "#" is
		// shared Dignified vocabulary (a KEEP tag, a toggle-block
		// marker, "##"/"###") - refusing here lets core.Lex's own
		// fallback claim it, same guard MSX/ZX81 use for the same
		// reason.
		if len(remaining) > 1 && (isAlpha(remaining[1]) || remaining[1] == '#') {
			return 0, nil, false, nil
		}
	}

	for _, s := range symbolWords {
		if strings.HasPrefix(remaining, s) {
			return len(s), one(KindSymbol, s), true, nil
		}
	}

	// numberPattern mirrors c_number (badig_zxspectrum.py:98, built
	// from the SAME `number` regex as c_number's own definition at
	// line 87) - no leading-sign group at all (confirmed empirically:
	// running badig.py on "LET score2=-1" renders the sign as its own
	// plain "-" symbol, immediately adjacent to the number, never
	// merged - ver tokenizer_test.go's
	// TestNumbersUseIntegerAndFloatingPointFormats, which already
	// exercises this exact case against the real tokenizer output).
	if m := numberLexPattern.FindString(remaining); m != "" {
		return len(m), one(KindNumber, m), true, nil
	}

	if isAlpha(remaining[0]) {
		// Multi-word compounds ("GO TO"/"GO SUB"/"DEF FN"/"OPEN #"/
		// "CLOSE #") are checked before any single-word candidate -
		// same priority as tokenizer.go's own tokenizeLine, and the
		// SAME regex (compoundRe), since both levels recognize exactly
		// this one set of multi-word shapes.
		if m := compoundRe.FindString(remaining); m != "" {
			upper := normalizeKeywordSpacing(m)
			switch upper {
			case "GO TO", "GO SUB":
				return len(m), one(KindJump, m), true, nil
			default: // "DEF FN", "OPEN #", "CLOSE #"
				return len(m), one(KindInstruction, m), true, nil
			}
		}

		word := scanKeywordCandidate(remaining)
		upper := strings.ToUpper(word)
		// Same guard as MSX/ZX81's own lexer: a keyword match is only
		// complete if what immediately follows isn't itself a valid
		// identifier-continuation character.
		if len(word) >= len(remaining) || !isIdentifierContinuation(remaining[len(word)]) {
			switch {
			case upper == "REM":
				// Split into TWO tokens - the "REM" keyword itself
				// (eligible for the generic reserved-keyword uppercase
				// forcing AND spacing in Pass5/core.Generate, exactly
				// like any other instruction) and its content, kept
				// byte-for-byte as the user wrote it (never forced to
				// uppercase - mirrors c_lit_toks' own exclusion).
				// Confirmed empirically: `rem lowercase text` renders
				// as "REM  lowercase text" - TWO spaces, one from
				// core.Generate's own automatic reserved-word spacing,
				// one already PART of the content (its own leading
				// space, preserved verbatim) - NOT a single merged
				// blob the way MSX/ZX81's simpler REM handling works.
				return len(remaining), []core.Token{
					{Kind: KindInstruction, Value: word},
					{Kind: KindComment, Value: remaining[len(word):]},
				}, true, nil
			case contains(operatorWords, upper):
				return len(word), one(KindOperator, word), true, nil
			case contains(functions, upper):
				return len(word), one(KindFunction, word), true, nil
			case contains(jumpWordsSingle, upper):
				return len(word), one(KindJump, word), true, nil
			case contains(instructionWords, upper):
				return len(word), one(KindInstruction, word), true, nil
			}
		}
		// Not a keyword: let the driver's generic fallback take over
		// (a Dignified identifier, same tolerant behavior as MSX/ZX81).
		return 0, nil, false, nil
	}

	return 0, nil, false, nil
}

// normalizeKeywordSpacing collapses internal whitespace to a single
// space and uppercases - mirrors the SAME normalization tokenizeLine
// applies to a compound match (ver tokenizer.go), reused here only to
// CLASSIFY the match (the token's own Value keeps the original,
// unnormalized text - Pass5 is what actually rewrites "GO  TO"/"GOTO"
// to a canonical "GO TO" in the rendered Classic text, same as the
// original engine's own pass_5).
func normalizeKeywordSpacing(s string) string {
	return strings.Join(strings.Fields(strings.ToUpper(s)), " ")
}

func one(kind core.TokenKind, value string) []core.Token {
	return []core.Token{{Kind: kind, Value: value}}
}

func contains(list []string, s string) bool {
	return slices.Contains(list, s)
}

func isIdentifierContinuation(c byte) bool {
	return isASCIIAlnum(c) || c == '_'
}

// scanKeywordCandidate extracts a run of letters, plus one trailing
// "$" if present (so function names like "INKEY$"/"SCREEN$" are
// captured whole) - same shape as MSX/ZX81's own.
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

func isAlpha(ch byte) bool { return isASCIIAlpha(ch) }

// numberLexPattern mirrors c_number (badig_zxspectrum.py:87,98) - the
// SAME shape as the tokenizer's own numberRe (tokenizer.go), kept as a
// separate compiled regex at this layer since the two are conceptually
// independent (Lexer vocabulary vs. tokenizer byte-encoding), even
// though they happen to describe the same grammar.
var numberLexPattern = regexp.MustCompile(`^(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?`)

// hexByteEscapeLex mirrors RAW_ESCAPE (tokenizer_zxspectrum.py:61) -
// the SAME pattern the tokenizer uses to later decode this escape;
// here it only needs to recognize/validate the Lexer-level token.
var hexByteEscapeLex = regexp.MustCompile(`^\\\{([0-9a-fA-F]{2})\}`)

// scanHexByte recognizes a `\{XX}` escape - same approach as ZX81's
// own scanHexByte (zx81/lexer.go): scan forward for a closing "}",
// then validate the whole captured span against the exact shape.
func scanHexByte(remaining string) (int, core.Token, error) {
	end := strings.IndexByte(remaining[2:], '}')
	if end < 0 {
		return 0, core.Token{}, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  `Invalid Spectrum escape: expected \{XX}.`,
		}
	}
	span := remaining[:2+end+1]
	if !hexByteEscapeLex.MatchString(span) {
		return 0, core.Token{}, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  `Invalid Spectrum escape: expected \{XX}.`,
		}
	}
	return len(span), core.Token{Kind: KindHexByte, Value: span}, nil
}

// scanSpectrumString mirrors badig_zxspectrum.py's Lexer.quoted_string
// override (badig_zxspectrum.py:161-171): a Spectrum string literal
// escapes an embedded quote ONLY by doubling it ("" - no backslash
// form at all, unlike MSX/ZX81), and an unterminated string (no
// closing quote before the line ends) is a FATAL error - unlike ZX81,
// which tolerantly consumes to end-of-line instead.
func scanSpectrumString(remaining string) (int, core.Token, error) {
	i := 1
	for i < len(remaining) {
		if remaining[i] == '"' {
			if i+1 < len(remaining) && remaining[i+1] == '"' {
				i += 2
				continue
			}
			i++
			return i, core.Token{Kind: core.KindString, Value: remaining[:i]}, nil
		}
		i++
	}
	return 0, core.Token{}, core.Diagnostic{
		Severity: core.SeverityError,
		Message:  "Unterminated Spectrum string",
	}
}

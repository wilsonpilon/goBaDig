// Package zxspectrum implements the ZX Spectrum-dialect pieces of the
// goBaDig conversion engine. See OUTLINE.md, "Bloco 4 — Módulo ZX
// Spectrum".
//
// This first slice is the tokenizer only (Classic ASCII numbered text
// -> tokenized program bytes), mirroring how the MSX/ZX81 ports both
// landed their own standalone tokenizer before any Description/Lexer/
// Parser existed. Ported line by line from the real reference
// implementation (badig/zxspectrum/tokenizer_zxspectrum.py, 516 lines,
// read in full this session), confirmed against its own test vectors
// (badig/zxspectrum/test_zxspectrum.py's TokenizerTests/TapeTests).
//
// REMX is deliberately NOT ported in this slice (same scope decision
// already made for ZX81 - see zx81/tokenizer.go's own package comment
// history in OUTLINE.md, Bloco 3): badig_zxspectrum.py's own REMX
// lives entirely at the Dignified Parser level (Pass1 + a postprocess-
// style reinsertion that bypasses normal line numbering), unlike
// ZX81's own REMX, which the tokenizer itself recognizes as a special
// line prefix. There is no tokenizer-level REMX surface to port here
// at all - revisit together with the Parser (Etapa 3), if a real
// program ever needs it.
package zxspectrum

import (
	"fmt"
	"maps"
	"math/big"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// moduleVersion mirrors VERSION in badig/zxspectrum/__init__.py - used
// only to word an error message identically to the original ("is not
// supported in Spectrum vX.Y.Z"), not a real dependency on anything.
const moduleVersion = "1.10.0"

// keywords mirrors KEYWORDS (tokenizer_zxspectrum.py:10-22) - order is
// significant: TokenCodes assigns byte values by enumerating this list
// starting at 0xA5, exactly like the original's
// `enumerate(KEYWORDS, 0xA5)`.
var keywords = []string{
	"RND", "INKEY$", "PI", "FN", "POINT", "SCREEN$", "ATTR", "AT", "TAB",
	"VAL$", "CODE", "VAL", "LEN", "SIN", "COS", "TAN", "ASN", "ACS", "ATN",
	"LN", "EXP", "INT", "SQR", "SGN", "ABS", "PEEK", "IN", "USR", "STR$",
	"CHR$", "NOT", "BIN", "OR", "AND", "<=", ">=", "<>", "LINE", "THEN",
	"TO", "STEP", "DEF FN", "CAT", "FORMAT", "MOVE", "ERASE", "OPEN #",
	"CLOSE #", "MERGE", "VERIFY", "BEEP", "CIRCLE", "INK", "PAPER", "FLASH",
	"BRIGHT", "INVERSE", "OVER", "OUT", "LPRINT", "LLIST", "STOP", "READ",
	"DATA", "RESTORE", "NEW", "BORDER", "CONTINUE", "DIM", "REM", "FOR",
	"GO TO", "GO SUB", "INPUT", "LOAD", "LIST", "LET", "PAUSE", "NEXT",
	"POKE", "PRINT", "PLOT", "RUN", "SAVE", "RANDOMIZE", "IF", "CLS", "DRAW",
	"CLEAR", "RETURN", "COPY",
}

// model128Tokens mirrors MODEL_128_TOKENS (tokenizer_zxspectrum.py:25):
// on 48K these same byte codes are the last two UDG graphics (%T and
// %U); the 128K editor ROM repurposes them as SPECTRUM and PLAY.
var model128Tokens = map[string]byte{"SPECTRUM": 0xA3, "PLAY": 0xA4}

// tokenCodes/codeKeywords mirror TOKEN_CODES/CODE_KEYWORDS.
var tokenCodes = map[string]byte{}
var codeKeywords = map[byte]string{}

// functions mirrors FUNCTIONS (tokenizer_zxspectrum.py:31-35).
var functions = []string{
	"RND", "INKEY$", "PI", "POINT", "SCREEN$", "ATTR", "VAL$", "CODE",
	"VAL", "LEN", "SIN", "COS", "TAN", "ASN", "ACS", "ATN", "LN", "EXP",
	"INT", "SQR", "SGN", "ABS", "PEEK", "IN", "USR", "STR$", "CHR$", "FN",
}

// supported/statements mirror SUPPORTED/STATEMENTS (tokenizer_zxspectrum.py:
// 36-54): the deliberate subset of the real 48K ROM vocabulary this
// module accepts; statements is the smaller set of keywords that may
// legally START a BASIC statement.
var supported = map[string]bool{}
var statements = map[string]bool{}

// model128Supported/model128Statements mirror MODEL_128_SUPPORTED/
// MODEL_128_STATEMENTS (tokenizer_zxspectrum.py:29-30) - both equal to
// the same frozenset in the original.
var model128Supported = map[string]bool{"SPECTRUM": true, "PLAY": true}
var model128Statements = model128Supported

// commandsNeedingArgumentCheck mirrors the inline tuple tokenize_line
// checks (tokenizer_zxspectrum.py:395-397) before calling
// validateCommandArguments.
var commandsNeedingArgumentCheck = map[string]bool{
	"RUN": true, "LIST": true, "LLIST": true, "NEW": true, "CONTINUE": true,
	"OUT": true, "OPEN #": true, "CLOSE #": true, "COPY": true, "LOAD": true,
	"SAVE": true, "VERIFY": true, "MERGE": true, "PLAY": true, "SPECTRUM": true,
}

func init() {
	for i, word := range keywords {
		tokenCodes[word] = byte(0xA5 + i)
	}
	maps.Copy(tokenCodes, model128Tokens)
	for word, code := range tokenCodes {
		codeKeywords[code] = word
	}

	supportedWords := []string{
		"PRINT", "LET", "IF", "GO TO", "REM", "CLS", "STOP",
		"THEN", "AND", "OR", "NOT", "<=", ">=", "<>",
		"BORDER", "INK", "PAPER", "PAUSE", "BEEP", "INPUT",
		"GO SUB", "RETURN", "FOR", "NEXT", "TO", "STEP",
		"DIM", "DATA", "READ", "RESTORE", "RANDOMIZE",
		"AT", "TAB", "FLASH", "BRIGHT", "INVERSE", "OVER", "PLOT", "DRAW",
		"CIRCLE", "POKE", "CLEAR", "RUN", "LIST", "NEW", "CONTINUE", "OUT", "BIN",
		"DEF FN", "LOAD", "SAVE", "VERIFY", "MERGE", "LINE", "OPEN #", "CLOSE #",
		"LPRINT", "LLIST", "COPY",
	}
	supportedWords = append(supportedWords, functions...)
	for _, w := range supportedWords {
		supported[w] = true
	}

	statementWords := []string{
		"PRINT", "LET", "IF", "GO TO", "REM", "CLS", "STOP",
		"BORDER", "INK", "PAPER", "PAUSE", "BEEP", "INPUT",
		"GO SUB", "RETURN", "FOR", "NEXT",
		"DIM", "DATA", "READ", "RESTORE", "RANDOMIZE",
		"FLASH", "BRIGHT", "INVERSE", "OVER", "PLOT", "DRAW",
		"CIRCLE", "POKE", "CLEAR", "RUN", "LIST", "NEW",
		"CONTINUE", "OUT", "DEF FN", "LOAD", "SAVE", "VERIFY",
		"MERGE", "OPEN #", "CLOSE #", "LPRINT", "LLIST", "COPY",
	}
	for _, w := range statementWords {
		statements[w] = true
	}
}

var (
	numberRe        = regexp.MustCompile(`^(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?`)
	numberFullRe    = regexp.MustCompile(`^(?:(\d+)(?:\.(\d*))?|\.(\d+))(?:[eE]([+-]?\d+))?$`)
	wordRe          = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*[$%!]?`)
	binaryNumberRe  = regexp.MustCompile(`^[ \t]*([01](?:[ \t]*[01])*)`)
	lineRe          = regexp.MustCompile(`^\s*([0-9]+)\s+(.+)$`)
	compoundRe      = regexp.MustCompile(`^(?i:GO\s+(?:TO|SUB)\b|DEF\s+FN\b|(?:OPEN|CLOSE)\s*#)`)
	graphicEscapeRe = regexp.MustCompile(`^\\([0-9A-Fa-f]{2})`)
	rawEscapeRe     = regexp.MustCompile(`^\\\{([0-9A-Fa-f]{2})\}`)
	rawEscapeFindRe = regexp.MustCompile(`\\\{[0-9A-Fa-f]{2}\}`) // non-anchored, mirrors RAW_ESCAPE.search
	openCloseRe     = regexp.MustCompile(`^(OPEN|CLOSE)\s*#$`)
	fnDefRe         = regexp.MustCompile(`^(\s+[A-Za-z]\$?\s*\()([^()]*)(\)\s*=)`)
	fnParamRe       = regexp.MustCompile(`^\s*[A-Za-z]\$?\s*$`)
	fnCallOpenRe    = regexp.MustCompile(`^\s+[A-Za-z]\$?\s*\(`)
	fnCallRawRe     = regexp.MustCompile(`^\s+\\\{`)
	identifierRe    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*\$?$`)
	cmdArgScanRe    = regexp.MustCompile(`"(?:[^"]|"")*"|[(),:]`)
)

var twoPow32 = new(big.Int).Lsh(big.NewInt(1), 32)

// TokenizeError mirrors SpectrumTokenizeError in tokenizer_zxspectrum.py
// - a plain message, no structured line field: the original embeds
// line numbers directly into each message's text at the point it is
// raised/re-raised, rather than carrying them as data.
type TokenizeError struct{ Message string }

func (e *TokenizeError) Error() string { return e.Message }

func tokenizeErrorf(format string, args ...any) error {
	return &TokenizeError{Message: fmt.Sprintf(format, args...)}
}

// EncodeNumber encodes a nonnegative numeric literal as the Sinclair
// ROM's five-byte number record, mirroring encode_number exactly -
// including using exact rational arithmetic (math/big, not float64)
// the same way the original uses decimal.Decimal+fractions.Fraction,
// since this format (unlike ZX81's own hidden float) is defined in
// terms of an EXACT decimal-to-binary conversion, not IEEE754/any
// floating-point rounding.
func EncodeNumber(text string) ([]byte, error) {
	m := numberFullRe.FindStringSubmatch(text)
	if m == nil {
		return nil, tokenizeErrorf("Invalid number: %s", text)
	}
	intPart, fracPart := m[1], m[2]
	if intPart == "" && m[3] != "" {
		fracPart = m[3]
	}
	digitsStr := intPart + fracPart
	d := new(big.Int)
	if digitsStr != "" {
		if _, ok := d.SetString(digitsStr, 10); !ok {
			return nil, tokenizeErrorf("Invalid number: %s", text)
		}
	}
	if d.Sign() == 0 {
		return make([]byte, 5), nil
	}

	eExp := big.NewInt(0)
	if m[4] != "" {
		expText := strings.TrimPrefix(m[4], "+")
		if _, ok := eExp.SetString(expText, 10); !ok {
			return nil, tokenizeErrorf("Invalid number: %s", text)
		}
	}
	e := new(big.Int).Sub(eExp, big.NewInt(int64(len(fracPart))))
	numDigits := len(d.String())
	adjusted := new(big.Int).Add(e, big.NewInt(int64(numDigits-1)))
	if adjusted.Cmp(big.NewInt(-39)) < 0 || adjusted.Cmp(big.NewInt(38)) > 0 {
		return nil, tokenizeErrorf("Number outside Sinclair range: %s", text)
	}

	var numBig, denomBig *big.Int
	if e.Sign() >= 0 {
		pow := new(big.Int).Exp(big.NewInt(10), e, nil)
		numBig = new(big.Int).Mul(d, pow)
		denomBig = big.NewInt(1)
	} else {
		negE := new(big.Int).Neg(e)
		numBig = d
		denomBig = new(big.Int).Exp(big.NewInt(10), negE, nil)
	}
	value := new(big.Rat).SetFrac(numBig, denomBig)

	if value.IsInt() && value.Num().Cmp(big.NewInt(65535)) <= 0 {
		v := value.Num().Uint64()
		return []byte{0, 0, byte(v), byte(v >> 8), 0}, nil
	}

	numerator := new(big.Int).Set(value.Num())
	denominator := new(big.Int).Set(value.Denom())
	exponent := numerator.BitLen() - denominator.BitLen()

	var below bool
	if exponent >= 0 {
		below = numerator.Cmp(new(big.Int).Lsh(denominator, uint(exponent))) < 0
	} else {
		below = new(big.Int).Lsh(numerator, uint(-exponent)).Cmp(denominator) < 0
	}
	if below {
		exponent--
	}
	shift := 31 - exponent
	if shift >= 0 {
		numerator = new(big.Int).Lsh(numerator, uint(shift))
	} else {
		denominator = new(big.Int).Lsh(denominator, uint(-shift))
	}
	mantissa, remainder := new(big.Int).QuoRem(numerator, denominator, new(big.Int))
	if new(big.Int).Lsh(remainder, 1).Cmp(denominator) >= 0 {
		mantissa.Add(mantissa, big.NewInt(1))
	}
	if mantissa.Cmp(twoPow32) == 0 {
		mantissa.Rsh(mantissa, 1)
		exponent++
	}
	if exponent+129 < 1 || exponent+129 > 255 {
		return nil, tokenizeErrorf("Number outside Sinclair range: %s", text)
	}
	mantissa.And(mantissa, big.NewInt(0x7FFFFFFF))
	out := make([]byte, 4)
	mb := mantissa.Bytes()
	copy(out[4-len(mb):], mb)
	return append([]byte{byte(exponent + 129)}, out...), nil
}

// EncodeText mirrors encode_text (tokenizer_zxspectrum.py:112-158):
// escapes/UDGs/graphics aside, this operates on RUNES, not bytes - the
// £/© special cases (and the general character-range rejection) only
// make sense decoded as Unicode codepoints, unlike the ZX81 port's own
// deliberately byte-wise (latin1-style) scanning elsewhere in this
// project. raw=true allows the "\{XX}" raw-byte escape (used for code/
// string/REM text in practice); raw=false (not used by this package's
// own tokenizeLine, kept for API fidelity with the original, which
// Lexer/Parser code - not yet ported - may call with raw=false)
// restricts "\" to the graphic-escape form only.
func EncodeText(text string, raw bool, model string) ([]byte, error) {
	var result []byte
	index := 0
	for index < len(text) {
		ch := text[index]
		if strings.HasPrefix(text[index:], `\\`) || strings.HasPrefix(text[index:], `%%`) {
			result = append(result, ch)
			index += 2
			continue
		}
		if ch == '\\' {
			var code int
			consumed := 0
			if raw {
				if m := rawEscapeRe.FindStringSubmatchIndex(text[index:]); m != nil {
					v, _ := strconv.ParseUint(text[index+m[2]:index+m[3]], 16, 8)
					code = int(v)
					consumed = m[1]
				}
			}
			if consumed == 0 {
				m := graphicEscapeRe.FindStringSubmatchIndex(text[index:])
				if m == nil {
					return nil, tokenizeErrorf(`Invalid Spectrum escape: use \80..\A4, \\ or \{XX}`)
				}
				v, _ := strconv.ParseUint(text[index+m[2]:index+m[3]], 16, 8)
				code = int(v)
				if code < 128 || code > 164 {
					return nil, tokenizeErrorf(`Graphic escape must be within \80..\A4`)
				}
				if model == "128" && (code == int(model128Tokens["SPECTRUM"]) || code == int(model128Tokens["PLAY"])) {
					return nil, tokenizeErrorf(`\A3/\A4 are the SPECTRUM/PLAY tokens with --sp_model 128; use \{XX} for the raw byte`)
				}
				consumed = m[1]
			}
			result = append(result, byte(code))
			index += consumed
			continue
		}
		if ch == '%' && index+1 < len(text) && isASCIIAlpha(text[index+1]) {
			letter := toUpperASCII(text[index+1])
			if letter < 'A' || letter > 'U' {
				return nil, tokenizeErrorf("Spectrum UDG escapes use %%A..%%U")
			}
			if model == "128" && (letter == 'T' || letter == 'U') {
				return nil, tokenizeErrorf("UDG %%T/%%U are not available with --sp_model 128; those codes are the SPECTRUM/PLAY tokens")
			}
			result = append(result, byte(144+int(letter-'A')))
			index += 2
			continue
		}
		r, w := decodeRuneAt(text, index)
		code := int(r)
		switch r {
		case '£': // £
			code = 0x60
		case '©': // ©
			code = 0x7F
		}
		if code < 32 || code > 127 || r == '`' || r == 0x7f {
			return nil, tokenizeErrorf("Unsupported Spectrum character: U+%04X", r)
		}
		result = append(result, byte(code))
		index += w
	}
	return result, nil
}

// quotedEnd mirrors quoted_end (tokenizer_zxspectrum.py:161-171): a
// Spectrum string literal escapes an embedded quote only by doubling
// it ("" - no backslash form, unlike ZX81/MSX).
func quotedEnd(text string, start int) (int, error) {
	end := start + 1
	for end < len(text) {
		if text[end] == '"' {
			end++
			if end < len(text) && text[end] == '"' {
				end += 1
				continue
			}
			return end, nil
		}
		end++
	}
	return 0, tokenizeErrorf("Unterminated string")
}

// sourceToken mirrors SourceToken (tokenizer_zxspectrum.py:65).
type sourceToken struct {
	kind  string
	value string
	start int
	end   int
}

// sourceTokens mirrors source_tokens (tokenizer_zxspectrum.py:174-199):
// a shallow scan that keeps strings/REM/raw-escape payloads opaque,
// used by fn_definition/validate_fn_call/validateCommandArguments to
// look for an early statement separator or raw escape without fully
// tokenizing. Eagerly collects into a slice (unlike the Python
// generator) since every caller's input here is one statement's worth
// of text, never a whole program.
func sourceTokens(text string) ([]sourceToken, error) {
	var tokens []sourceToken
	index := 0
	for index < len(text) {
		start := index
		var kind string
		var end int
		if text[index] == '"' {
			e, err := quotedEnd(text, index)
			if err != nil {
				return nil, err
			}
			end = e
			kind = "string"
		} else {
			matched := false
			if m := rawEscapeRe.FindStringIndex(text[index:]); m != nil {
				end = index + m[1]
				kind = "raw"
				matched = true
			}
			if !matched {
				if m := compoundRe.FindStringIndex(text[index:]); m != nil {
					end = index + m[1]
					kind = "word"
					matched = true
				} else if m := wordRe.FindStringIndex(text[index:]); m != nil {
					end = index + m[1]
					kind = "word"
					matched = true
				}
			}
			if !matched {
				if m := numberRe.FindStringIndex(text[index:]); m != nil {
					end = index + m[1]
					kind = "number"
					matched = true
				}
			}
			if !matched {
				end = index + 1
				kind = "symbol"
			}
		}
		value := text[start:end]
		if (kind == "word" && strings.EqualFold(value, "REM")) ||
			(kind == "raw" && rawByteIsRem(value)) {
			tokens = append(tokens, sourceToken{kind: "rem", value: text[start:], start: start, end: len(text)})
			return tokens, nil
		}
		tokens = append(tokens, sourceToken{kind: kind, value: value, start: start, end: end})
		index = end
	}
	return tokens, nil
}

// rawByteIsRem reports whether a "\{XX}" raw-escape token's byte is
// the REM token code - mirrors source_tokens' own inline check
// (`int(value[2:4], 16) == TOKEN_CODES['REM']`).
func rawByteIsRem(value string) bool {
	if len(value) < 4 {
		return false
	}
	v, err := strconv.ParseUint(value[2:4], 16, 8)
	return err == nil && byte(v) == tokenCodes["REM"]
}

// FnDefinition encodes a "DEF FN" signature, including the ROM's
// six-byte hidden parameter slots, mirroring fn_definition
// (tokenizer_zxspectrum.py:202-219). text is everything after the
// "DEF FN" keyword itself; the returned int is how many of text's
// bytes the signature consumed (up to and including the "=").
func FnDefinition(text, model string) ([]byte, int, error) {
	m := fnDefRe.FindStringSubmatchIndex(text)
	if m == nil {
		return nil, 0, tokenizeErrorf("DEF FN requires letter[$](parameters)=expression")
	}
	prefix := text[m[2]:m[3]]
	parameters := text[m[4]:m[5]]
	suffix := text[m[6]:m[7]]
	matchEnd := m[1]

	var names []string
	if strings.TrimSpace(parameters) != "" {
		names = strings.Split(parameters, ",")
	}
	for _, name := range names {
		if !fnParamRe.MatchString(name) {
			return nil, 0, tokenizeErrorf("DEF FN parameters require single letters with optional $")
		}
	}
	rest := text[matchEnd:]
	trimmedRest := strings.TrimLeft(rest, " \t\n\r\f\v")
	if strings.TrimSpace(rest) == "" || strings.HasPrefix(trimmedRest, ":") {
		return nil, 0, tokenizeErrorf("DEF FN requires an expression after =")
	}

	encodedPrefix, err := EncodeText(prefix, true, model)
	if err != nil {
		return nil, 0, err
	}
	encoded := append([]byte{}, encodedPrefix...)

	toks, err := sourceTokens(parameters)
	if err != nil {
		return nil, 0, err
	}
	for _, tok := range toks {
		tb, err := EncodeText(tok.value, true, model)
		if err != nil {
			return nil, 0, err
		}
		encoded = append(encoded, tb...)
		if tok.kind == "word" {
			encoded = append(encoded, 0x0e)
			encoded = append(encoded, make([]byte, 5)...)
		}
	}
	suffixBytes, err := EncodeText(suffix, true, model)
	if err != nil {
		return nil, 0, err
	}
	encoded = append(encoded, suffixBytes...)
	return encoded, matchEnd, nil
}

// ValidateFnCall mirrors validate_fn_call (tokenizer_zxspectrum.py:
// 222-243): checks an "FN" call's syntax only - letter[$](arguments)
// with balanced parentheses - it never emits any byte of its own (the
// call tokenizes as plain text otherwise); the hidden 6-byte slots
// only ever apply to a DEF FN's own declaration.
func ValidateFnCall(text string) error {
	toks, err := sourceTokens(text)
	if err != nil {
		return err
	}
	for _, tok := range toks {
		if tok.kind == "symbol" && tok.value == ":" {
			break
		}
		if tok.kind == "raw" {
			return nil
		}
	}
	m := fnCallOpenRe.FindStringIndex(text)
	if m == nil {
		if fnCallRawRe.MatchString(text) {
			return nil
		}
		return tokenizeErrorf("FN requires letter[$](arguments)")
	}
	rest := text[m[1]:]
	toks2, err := sourceTokens(rest)
	if err != nil {
		return err
	}
	depth := 1
	for _, tok := range toks2 {
		if tok.kind != "symbol" {
			continue
		}
		switch tok.value {
		case "(":
			depth++
		case ")":
			depth--
			if depth == 0 {
				return nil
			}
		}
	}
	return tokenizeErrorf("FN requires closing argument parentheses")
}

// ValidateIdentifier mirrors validate_identifier (tokenizer_zxspectrum.py:
// 287-299) - the core of the Spectrum module's biggest structural
// difference from MSX/ZX81: there is no alias allocator here at all,
// only a per-category NAME-SHAPE validation (native Sinclair BASIC
// variable names are valid at any length).
func ValidateIdentifier(name string, array, forControl bool) error {
	if !identifierRe.MatchString(name) {
		return tokenizeErrorf("Invalid Sinclair name: %s; use DECLARE name:letter for snake_case", name)
	}
	var valid bool
	var category string
	if strings.HasSuffix(name, "$") {
		valid = len(name) == 2 && !forControl
		category = "String variables"
	} else {
		if array || forControl {
			valid = len(name) == 1
		} else {
			valid = true
		}
		category = "Arrays and FOR/NEXT variables"
	}
	if !valid {
		return tokenizeErrorf("%s require a single letter: %s; use an explicit DECLARE alias", category, name)
	}
	return nil
}

// validateCommandArguments mirrors _validate_command_arguments
// (tokenizer_zxspectrum.py:302-346): structural validation (argument
// count/emptiness, balanced parentheses) for the native commands whose
// argument shape the tokenizer itself enforces, since the Dignified
// Parser never looks at their contents.
func validateCommandArguments(command, text string) error {
	toks, err := sourceTokens(text)
	if err != nil {
		return err
	}
	for _, tok := range toks {
		if tok.kind == "symbol" && tok.value == ":" {
			break
		}
		if tok.kind == "raw" {
			return nil
		}
	}

	var arguments []string
	start := 0
	depth := 0
	end := len(text)
	for _, loc := range cmdArgScanRe.FindAllStringIndex(text, -1) {
		symbol := text[loc[0]:loc[1]]
		if symbol == ":" {
			end = loc[0]
			break
		}
		switch symbol {
		case "(":
			depth++
		case ")":
			depth--
			if depth < 0 {
				return tokenizeErrorf("%s: unbalanced parentheses", command)
			}
		case ",":
			if depth == 0 {
				arguments = append(arguments, strings.TrimSpace(text[start:loc[0]]))
				start = loc[1]
			}
		}
	}
	if depth != 0 {
		return tokenizeErrorf("%s: unbalanced parentheses", command)
	}
	arguments = append(arguments, strings.TrimSpace(text[start:end]))
	if len(arguments) == 1 && arguments[0] == "" {
		arguments = nil
	}

	switch command {
	case "NEW", "CONTINUE", "COPY", "SPECTRUM":
		if len(arguments) > 0 {
			return tokenizeErrorf("%s takes no arguments", command)
		}
	}
	maximum := 1
	if (command == "LIST" || command == "LLIST") && len(arguments) > 0 && strings.HasPrefix(arguments[0], "#") {
		maximum = 2
	}
	if (command == "RUN" || command == "LIST" || command == "LLIST") && len(arguments) > maximum {
		return tokenizeErrorf("%s takes at most one expression", command)
	}
	if command == "OUT" && (len(arguments) != 2 || !allNonEmpty(arguments)) {
		return tokenizeErrorf("OUT requires two expressions separated by a comma")
	}
	if command == "OPEN #" && (len(arguments) != 2 || !allNonEmpty(arguments)) {
		return tokenizeErrorf("OPEN # requires stream,channel")
	}
	if command == "CLOSE #" && (len(arguments) != 1 || arguments[0] == "") {
		return tokenizeErrorf("CLOSE # requires one stream")
	}
	switch command {
	case "LOAD", "SAVE", "VERIFY", "MERGE":
		if len(arguments) == 0 {
			return tokenizeErrorf("%s requires a filename expression", command)
		}
	}
	if command == "PLAY" && !(len(arguments) >= 1 && len(arguments) <= 3 && allNonEmpty(arguments)) {
		return tokenizeErrorf("PLAY requires 1 to 3 expressions separated by commas")
	}
	return nil
}

func allNonEmpty(ss []string) bool {
	return !slices.Contains(ss, "")
}

// tokenizeLine mirrors tokenize_line (tokenizer_zxspectrum.py:349-482):
// encodes one Classic BASIC statement's text (everything after its
// line number) into the bytes the Spectrum ROM stores for it.
func tokenizeLine(text, model string) ([]byte, error) {
	var result []byte
	index := 0
	atStatement := true
	previousKeyword := ""
	statement := ""

	for index < len(text) {
		ch := text[index]
		if ch == ' ' || ch == '\t' {
			result = append(result, 32)
			index++
			continue
		}

		if m := rawEscapeRe.FindStringSubmatchIndex(text[index:]); m != nil {
			v, _ := strconv.ParseUint(text[index+m[2]:index+m[3]], 16, 8)
			code := byte(v)
			result = append(result, code)
			index += m[1]
			if kw, ok := codeKeywords[code]; ok && statements[kw] {
				statement = kw
			}
			if code == tokenCodes["REM"] {
				tb, err := EncodeText(text[index:], true, model)
				if err != nil {
					return nil, err
				}
				return append(append(result, tb...), 0x0d), nil
			}
			atStatement = false
			previousKeyword = ""
			continue
		}
		if ch == '\\' {
			return nil, tokenizeErrorf(`Raw code bytes require \{XX}, exactly two hex digits`)
		}

		matchEnd := -1
		if m := compoundRe.FindStringIndex(text[index:]); m != nil {
			matchEnd = index + m[1]
		} else if m := wordRe.FindStringIndex(text[index:]); m != nil {
			matchEnd = index + m[1]
		}
		if matchEnd >= 0 {
			raw := text[index:matchEnd]
			fields := strings.Fields(strings.ToUpper(raw))
			keyword := strings.Join(fields, " ")
			switch keyword {
			case "GOTO":
				keyword = "GO TO"
			case "GOSUB":
				keyword = "GO SUB"
			}
			keyword = openCloseRe.ReplaceAllString(keyword, "$1 #")

			if code, ok := tokenCodes[keyword]; ok {
				model128Only := model128Supported[keyword]
				if model128Only && model != "128" {
					return nil, tokenizeErrorf("%s requires --sp_model 128", keyword)
				}
				if !model128Only && !supported[keyword] {
					return nil, tokenizeErrorf("%s is not supported in Spectrum v%s", keyword, moduleVersion)
				}
				isStatement := statements[keyword] || (model == "128" && model128Statements[keyword])
				if atStatement && !isStatement {
					return nil, tokenizeErrorf("Expected a statement, got %s", keyword)
				}
				if atStatement {
					statement = keyword
				}
				if commandsNeedingArgumentCheck[keyword] {
					if !atStatement {
						return nil, tokenizeErrorf("%s must start a statement", keyword)
					}
					if err := validateCommandArguments(keyword, text[matchEnd:]); err != nil {
						return nil, err
					}
				}
				result = append(result, code)

				switch keyword {
				case "DEF FN":
					tail := text[matchEnd:]
					beforeEq, _, _ := strings.Cut(tail, "=")
					if !rawEscapeFindRe.MatchString(beforeEq) {
						encoded, consumed, err := FnDefinition(tail, model)
						if err != nil {
							return nil, err
						}
						result = append(result, encoded...)
						index = matchEnd + consumed
					} else {
						index = matchEnd
					}
					atStatement = false
					previousKeyword = ""
					continue
				case "FN":
					if err := ValidateFnCall(text[matchEnd:]); err != nil {
						return nil, err
					}
				case "BIN":
					m2 := binaryNumberRe.FindStringSubmatchIndex(text[matchEnd:])
					if m2 == nil {
						return nil, tokenizeErrorf("BIN requires binary digits 0 or 1")
					}
					binEnd := matchEnd + m2[1]
					digitsRaw := text[matchEnd+m2[2] : matchEnd+m2[3]]
					suffix := text[binEnd:]
					following := strings.TrimLeft(suffix, " \t")
					bad := false
					if suffix != "" && (isASCIIAlnum(suffix[0]) || suffix[0] == '.' || suffix[0] == '_' || suffix[0] == '$') {
						bad = true
					}
					if !bad && following != "" && (isASCIIDigit(following[0]) || following[0] == '.') {
						bad = true
					}
					if bad {
						return nil, tokenizeErrorf("Invalid BIN literal: use only binary digits")
					}
					digits := strings.TrimLeft(strings.NewReplacer(" ", "", "\t", "").Replace(digitsRaw), "0")
					if len(digits) > 16 {
						return nil, tokenizeErrorf("BIN value exceeds 65535")
					}
					binText := strings.ReplaceAll(text[matchEnd:binEnd], "\t", " ")
					tb, err := EncodeText(binText, true, model)
					if err != nil {
						return nil, err
					}
					result = append(result, tb...)
					result = append(result, 0x0e)
					var val int64
					if digits != "" {
						val, _ = strconv.ParseInt(digits, 2, 64)
					}
					nb, err := EncodeNumber(strconv.FormatInt(val, 10))
					if err != nil {
						return nil, err
					}
					result = append(result, nb...)
					index = binEnd
					atStatement = false
					previousKeyword = ""
					continue
				case "REM":
					tb, err := EncodeText(text[matchEnd:], true, model)
					if err != nil {
						return nil, err
					}
					return append(append(result, tb...), 0x0d), nil
				}
				atStatement = keyword == "THEN"
				previousKeyword = keyword
			} else {
				if atStatement {
					return nil, tokenizeErrorf("Expected a supported statement (use explicit LET): %s", raw)
				}
				after := strings.TrimLeft(text[matchEnd:], " \t\n\r\f\v")
				forControl := previousKeyword == "FOR" || previousKeyword == "NEXT"
				if err := ValidateIdentifier(raw, strings.HasPrefix(after, "("), forControl); err != nil {
					return nil, err
				}
				result = append(result, []byte(raw)...)
				previousKeyword = ""
			}
			index = matchEnd
			continue
		}

		if atStatement {
			return nil, tokenizeErrorf("Expected a supported statement after line number, : or THEN")
		}
		if ch == '"' {
			end, err := quotedEnd(text, index)
			if err != nil {
				return nil, err
			}
			tb, err := EncodeText(text[index:end], true, model)
			if err != nil {
				return nil, err
			}
			result = append(result, tb...)
			index = end
			continue
		}
		if m := numberRe.FindStringIndex(text[index:]); m != nil {
			absEnd := index + m[1]
			if absEnd < len(text) {
				next := text[absEnd]
				if next == '.' || next == 'e' || next == 'E' {
					return nil, tokenizeErrorf("Invalid number near: %s", text[index:])
				}
			}
			raw := text[index:absEnd]
			result = append(result, []byte(raw)...)
			result = append(result, 0x0e)
			nb, err := EncodeNumber(raw)
			if err != nil {
				return nil, err
			}
			result = append(result, nb...)
			index = absEnd
			continue
		}
		if index+2 <= len(text) {
			pair := text[index : index+2]
			if pair == "<=" || pair == ">=" || pair == "<>" {
				result = append(result, tokenCodes[pair])
				index += 2
				continue
			}
		}
		if ch == '#' {
			if statement != "PRINT" && statement != "INPUT" && statement != "LIST" && statement != "LPRINT" && statement != "LLIST" {
				return nil, tokenizeErrorf("# stream selector requires PRINT, INPUT or LIST")
			}
		}
		if !strings.ContainsRune(`+-*/^=<>(),;:'#`, rune(ch)) {
			return nil, tokenizeErrorf("Unsupported symbol: %c", ch)
		}
		tb, err := EncodeText(string(ch), true, model)
		if err != nil {
			return nil, err
		}
		result = append(result, tb...)
		if ch == ':' {
			atStatement = true
		}
		index++
	}
	if atStatement {
		return nil, tokenizeErrorf("Missing statement after : or THEN")
	}
	return append(result, 0x0d), nil
}

// Tokenize mirrors tokenize (tokenizer_zxspectrum.py:485-516): converts
// a complete already-numbered Classic ASCII ZX Spectrum BASIC listing
// into tokenized program bytes - unlike ZX81's own Tokenize, this is
// JUST the program's bytes (line-number/length-prefixed records), not
// a full memory image: the Spectrum has no equivalent "system
// variables + display file" blob to also produce here. model must be
// "48" or "128".
func Tokenize(source, model string) ([]byte, error) {
	if model != "48" && model != "128" {
		return nil, tokenizeErrorf("Unknown Spectrum model: %s; use 48 or 128", model)
	}
	var program []byte
	previous := 0
	for i, text := range splitLines(source) {
		sourceLine := i + 1
		if strings.TrimSpace(text) == "" {
			continue
		}
		m := lineRe.FindStringSubmatch(text)
		if m == nil {
			return nil, tokenizeErrorf("Source line %d: expected a numbered BASIC line", sourceLine)
		}
		digits := strings.TrimLeft(m[1], "0")
		if digits == "" {
			digits = "0"
		}
		if len(digits) > 4 {
			return nil, tokenizeErrorf("Source line %d: line number exceeds 9999", sourceLine)
		}
		number, _ := strconv.Atoi(digits)
		if !(previous < number && number <= 9999) {
			return nil, tokenizeErrorf("Source line %d: line numbers must increase within 1..9999", sourceLine)
		}
		body, err := tokenizeLine(m[2], model)
		if err != nil {
			return nil, tokenizeErrorf("BASIC line %d: %s", number, err.Error())
		}
		if len(body) > 65535 {
			return nil, tokenizeErrorf("BASIC line %d: tokenized line too long", number)
		}
		program = append(program, byte(number>>8), byte(number))
		program = append(program, byte(len(body)), byte(len(body)>>8))
		program = append(program, body...)
		previous = number
	}
	if len(program) == 0 {
		return nil, tokenizeErrorf("No BASIC program to tokenize")
	}
	return program, nil
}

// splitLines mirrors Python's str.splitlines(): splits on \n, \r\n or
// lone \r, without keeping terminators and without a trailing empty
// entry for a final line terminator.
func splitLines(s string) []string {
	var lines []string
	start := 0
	i := 0
	for i < len(s) {
		c := s[i]
		if c == '\n' || c == '\r' {
			lines = append(lines, s[start:i])
			if c == '\r' && i+1 < len(s) && s[i+1] == '\n' {
				i++
			}
			i++
			start = i
			continue
		}
		i++
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func isASCIIAlpha(ch byte) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }
func isASCIIDigit(ch byte) bool { return ch >= '0' && ch <= '9' }
func isASCIIAlnum(ch byte) bool { return isASCIIAlpha(ch) || isASCIIDigit(ch) }

func toUpperASCII(ch byte) byte {
	if ch >= 'a' && ch <= 'z' {
		return ch - 32
	}
	return ch
}

// decodeRuneAt decodes one UTF-8 rune at text[index:], mirroring
// EncodeText's need to inspect a single Unicode character (£/© in
// particular) rather than a raw byte - see EncodeText's own doc
// comment for why this function, alone in this package, is rune-wise
// rather than byte-wise.
func decodeRuneAt(text string, index int) (rune, int) {
	return utf8.DecodeRuneInString(text[index:])
}

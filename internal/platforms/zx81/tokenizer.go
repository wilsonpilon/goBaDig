// Package zx81 implements the ZX81-dialect pieces of the goBaDig
// conversion engine. See OUTLINE.md, "Bloco 3 — Módulo ZX81".
//
// This first slice is the tokenizer only (Classic ASCII numbered text
// -> a complete ".p" memory image): no Description/Lexer/Parser yet,
// mirroring how the MSX port's own Batoken tokenizer landed before
// its Dignified dialect pieces did. Ported line by line from the real
// reference implementation (badig/zx81/tokenizer_zx81.py), confirmed
// against its own test vectors (badig/zx81/test_tokenizer_zx81.py) -
// see docs/golden-vectors/zx81-tokenizer.md for the committed copy of
// those vectors.
package zx81

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// tokenCodes is the ZX81 ROM's keyword table (lowercase match text ->
// token byte), transcribed verbatim from TOKEN_CODES in
// tokenizer_zx81.py. "go to"/"go sub" exist alongside "goto"/"gosub"
// because the ROM accepts either spelling - matched specially (see
// matchToken) since the whitespace between the two words is variable
// width, unlike every other entry here.
var tokenCodes = map[string]byte{
	"copy": 255, "return": 254, "clear": 253, "unplot": 252, "cls": 251,
	"if": 250, "randomize": 249, "rand": 249, "save": 248, "run": 247,
	"plot": 246, "print": 245, "poke": 244, "next": 243, "pause": 242,
	"let": 241, "list": 240, "load": 239, "input": 238, "go sub": 237,
	"gosub": 237, "go to": 236, "goto": 236, "for": 235, "rem": 234,
	"dim": 233, "continue": 232, "cont": 232, "scroll": 231, "new": 230,
	"fast": 229, "slow": 228, "stop": 227, "llist": 226, "lprint": 225,
	"step": 224, "to": 223, "then": 222, "<>": 221, ">=": 220, "<=": 219,
	"and": 218, "or": 217, "**": 216, "not": 215, "chr$": 214, "str$": 213,
	"usr": 212, "peek": 211, "abs": 210, "sgn": 209, "sqr": 208, "int": 207,
	"exp": 206, "ln": 205, "atn": 204, "acs": 203, "asn": 202, "tan": 201,
	"cos": 200, "sin": 199, "len": 198, "val": 197, "code": 196, "tab": 194,
	"at": 193, "pi": 66, "inkey$": 65, "rnd": 64,
}

// charCodes is CHAR_CODES from tokenizer_zx81.py: punctuation -> ZX81
// character code, used by characterCode as a fallback once a byte
// isn't a letter or digit.
var charCodes = map[byte]byte{
	' ': 0, '"': 11, '#': 12, '$': 13, ':': 14, '?': 15, '(': 16, ')': 17,
	'>': 18, '<': 19, '=': 20, '+': 21, '-': 22, '*': 23, '/': 24, ';': 25,
	',': 26, '.': 27,
}

// graphicsCodes is GRAPHICS_CODES from tokenizer_zx81.py: the 2-character
// `\XY` graphics escapes tokenizeLine recognizes.
var graphicsCodes = map[string]byte{
	"  ": 0, "' ": 1, " '": 2, "''": 3, ". ": 4, ": ": 5, ".'": 6, ":'": 7,
	"##": 8, ",,": 9, "~~": 10, "::": 128, ".:": 129, ":.": 130, "..": 131,
	"':": 132, " :": 133, "'.": 134, " .": 135, "@@": 136, ";;": 137, "!!": 138,
}

// tokenMatches holds tokenCodes' keys sorted longest-first, replicating
// TOKEN_MATCHES = tuple(sorted(TOKEN_CODES, key=len, reverse=True)) -
// matchToken scans it in this order so a longer keyword always wins
// over a shorter one that happens to be its prefix (there is no
// ambiguity between equal-length entries: two different token strings
// of the same length can never both match at the same text position).
var tokenMatches []string

func init() {
	tokenMatches = make([]string, 0, len(tokenCodes))
	for k := range tokenCodes {
		tokenMatches = append(tokenMatches, k)
	}
	sort.Slice(tokenMatches, func(i, j int) bool {
		return len(tokenMatches[i]) > len(tokenMatches[j])
	})
}

var (
	hexEscapeRe = regexp.MustCompile(`^\\\{([0-9a-fA-F]{2})\}`)
	numberRe    = regexp.MustCompile(`^(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?`)
	lineRe      = regexp.MustCompile(`^\s*([0-9]+)(?:\s+|$)(.*)$`)
	remxArgsRe  = regexp.MustCompile(`^\s*([0-9]+)\s*,\s*([0-9]+)\s*(?:,\s*([0-9]+|"[^"]"|'[^']'))?$`)
	goToRe      = regexp.MustCompile(`^go\s+to`)
	goSubRe     = regexp.MustCompile(`^go\s+sub`)
)

const (
	// programStart is PROGRAM_START - the fixed ZX81 address a
	// tokenized program's first byte lives at (0x407D), right after
	// the 116-byte system variables area.
	programStart = 0x407D
	// remxMax is REMX_MAX - the largest byte count a single remx
	// directive may request.
	remxMax = 49152
	// programMax is PROGRAM_MAX - the program area's own ceiling once
	// the fixed 793-byte display file and the 2-byte end-of-variables
	// marker (plus 1 byte of margin the original also reserves) are
	// accounted for, so the whole image still fits a 64 KB address
	// space starting at programStart.
	programMax = 0x10000 - programStart - 793 - 3
)

// TokenizeError reports a problem with one specific input line,
// mirroring ZX81TokenizeError in tokenizer_zx81.py.
type TokenizeError struct {
	Line    int
	Message string
}

func (e *TokenizeError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}

func tokenizeErrorf(line int, format string, args ...any) error {
	return &TokenizeError{Line: line, Message: fmt.Sprintf(format, args...)}
}

// Tokenize converts already-numbered Classic ASCII ZX81 BASIC lines
// (one program statement per entry, e.g. "10 PRINT 1", or a "remx
// <line>,<count>[,<byte>]" directive) into a complete ZX81 ".p" memory
// image - mirrors tokenize_lines in tokenizer_zx81.py, which (unlike
// the MSX tokenizer) already returns the WHOLE memory image (system
// variables + program + display file + variables area), not just the
// program's own bytes.
func Tokenize(lines []string) ([]byte, error) {
	var program []byte
	var lineNumbers []int

	for i, raw := range lines {
		sourceLine := i + 1
		text := strings.TrimRight(raw, "\r\n")
		trimmed := strings.TrimSpace(text)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		var lineNumber int
		var encodedLine []byte

		if rest, ok := matchRemxPrefix(text); ok {
			n, count, code, err := parseRemx(rest, sourceLine)
			if err != nil {
				return nil, err
			}
			lineNumber = n
			encodedLine = make([]byte, 0, count+2)
			encodedLine = append(encodedLine, tokenCodes["rem"])
			for range count {
				encodedLine = append(encodedLine, code)
			}
			encodedLine = append(encodedLine, 0x76)
		} else {
			m := lineRe.FindStringSubmatch(text)
			if m == nil {
				return nil, tokenizeErrorf(sourceLine, "input line %d: expected a numbered BASIC line", sourceLine)
			}
			n, err := strconv.Atoi(m[1])
			if err != nil {
				return nil, tokenizeErrorf(sourceLine, "invalid line number")
			}
			lineNumber = n
			encodedLine, err = tokenizeLine(m[2], lineNumber)
			if err != nil {
				return nil, err
			}
		}

		if lineNumber > 9999 {
			return nil, tokenizeErrorf(sourceLine, "input line %d: line number must be 0..9999", sourceLine)
		}
		if len(lineNumbers) > 0 && lineNumber <= lineNumbers[len(lineNumbers)-1] {
			return nil, tokenizeErrorf(sourceLine, "input line %d: line numbers must increase", sourceLine)
		}
		if len(program)+4+len(encodedLine) > programMax {
			return nil, tokenizeErrorf(sourceLine,
				"program is too large for a ZX-81 .p file: the program area cannot exceed %d bytes", programMax)
		}

		program = append(program, byte(lineNumber>>8), byte(lineNumber))
		program = append(program, byte(len(encodedLine)), byte(len(encodedLine)>>8))
		program = append(program, encodedLine...)
		lineNumbers = append(lineNumbers, lineNumber)
	}

	firstLine := 0
	if len(lineNumbers) > 0 {
		firstLine = lineNumbers[0]
	}
	return programImage(program, firstLine), nil
}

// matchRemxPrefix reports whether line (after leading whitespace) is
// a "remx ..." directive, returning the text after "remx" (trailing
// whitespace trimmed). Go's RE2 engine has no negative lookahead, so
// this hand-replicates REMX_IGNORE's `remx(?![A-Za-z0-9_$])` boundary
// check (rejecting "remxfoo" while accepting "remx 0,4,118") instead
// of translating the regex directly.
func matchRemxPrefix(line string) (rest string, ok bool) {
	trimmed := strings.TrimLeft(line, " \t\n\r\f\v")
	if len(trimmed) < 4 || !strings.EqualFold(trimmed[:4], "remx") {
		return "", false
	}
	if len(trimmed) > 4 && isRemxBoundaryChar(trimmed[4]) {
		return "", false
	}
	return strings.TrimRight(trimmed[4:], " \t\n\r\f\v"), true
}

func isRemxBoundaryChar(ch byte) bool {
	return isAlpha(ch) || isDigit(ch) || ch == '_' || ch == '$'
}

// parseRemx parses a remx directive's arguments ("<line>,<count>[,<byte>]"),
// mirroring parse_remx in tokenizer_zx81.py.
func parseRemx(arguments string, sourceLine int) (lineNumber, count int, code byte, err error) {
	m := remxArgsRe.FindStringSubmatch(arguments)
	if m == nil {
		return 0, 0, 0, tokenizeErrorf(sourceLine, "input line %d: remx expects <line>,<count>[,<byte>]", sourceLine)
	}
	lineNumber, _ = strconv.Atoi(m[1])
	count, _ = strconv.Atoi(m[2])
	fill := m[3]

	if lineNumber > 9999 {
		return 0, 0, 0, tokenizeErrorf(sourceLine, "input line %d: remx line number must be 0..9999", sourceLine)
	}
	if count < 1 || count > remxMax {
		return 0, 0, 0, tokenizeErrorf(sourceLine, "input line %d: remx count must be 1..%d", sourceLine, remxMax)
	}

	switch {
	case fill == "":
		code, err = characterCode('0', sourceLine)
	case fill[0] == '"' || fill[0] == '\'':
		code, err = characterCode(fill[1], sourceLine)
	default:
		v, convErr := strconv.Atoi(fill)
		if convErr != nil || v > 255 {
			return 0, 0, 0, tokenizeErrorf(sourceLine, "input line %d: remx byte must be 0..255", sourceLine)
		}
		code = byte(v)
	}
	if err != nil {
		return 0, 0, 0, err
	}
	return lineNumber, count, code, nil
}

// characterCode maps one Classic BASIC source character to its ZX81
// character code, mirroring _character_code in tokenizer_zx81.py.
func characterCode(ch byte, lineNumber int) (byte, error) {
	if isAlpha(ch) {
		return lowerByte(ch) - 59, nil
	}
	if isDigit(ch) {
		return ch - 20, nil
	}
	if code, ok := charCodes[ch]; ok {
		return code, nil
	}
	switch ch {
	case '!', '%', '&', '\'', '@':
		return 0, tokenizeErrorf(lineNumber, "character %q cannot be represented", string(ch))
	}
	if ch < 128 {
		return ch, nil
	}
	return 0, tokenizeErrorf(lineNumber, "non-ASCII character %q cannot be represented", string(ch))
}

// numberBytes encodes a numeric literal's FULL token form: the visible
// ASCII digits (each through characterCode) followed by the 5-byte
// "hidden" binary float the ZX81 ROM hides right after every number in
// a tokenized listing, invisible to LIST but used by the interpreter
// so it never has to re-parse the text. Mirrors _number_bytes in
// tokenizer_zx81.py exactly, including its Sinclair-style exponent
// bias (128, not IEEE754's 127) and the mantissa-overflow correction.
func numberBytes(text string, lineNumber int) ([]byte, error) {
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil, tokenizeErrorf(lineNumber, "invalid number %q", text)
	}
	magnitude := math.Abs(value)

	var exponent int
	var mantissa uint32
	if magnitude == 0 {
		exponent, mantissa = 0, 0
	} else {
		fraction, binaryExponent := math.Frexp(magnitude)
		exponent = binaryExponent + 128
		m := math.Round((fraction*2 - 1) * float64(0x80000000))
		mantissa = uint32(m)
		if mantissa == 0x80000000 {
			mantissa = 0
			exponent++
		}
		if exponent < 1 || exponent > 255 {
			return nil, tokenizeErrorf(lineNumber, "number exponent out of range: %s", text)
		}
	}

	encoded := make([]byte, 0, len(text)+5)
	for i := 0; i < len(text); i++ {
		code, err := characterCode(text[i], lineNumber)
		if err != nil {
			return nil, err
		}
		encoded = append(encoded, code)
	}
	encoded = append(encoded, 0x7E, byte(exponent))
	encoded = append(encoded, byte(mantissa>>24), byte(mantissa>>16), byte(mantissa>>8), byte(mantissa))
	return encoded, nil
}

// matchedToken is one matchToken result.
type matchedToken struct {
	end  int
	code byte
	text string
}

// matchToken finds the longest tokenCodes entry starting at text[index:],
// mirroring _match_token in tokenizer_zx81.py: a boundary guard on both
// sides keeps a keyword from matching inside a longer identifier (e.g.
// "RUN" inside "RUNNER").
func matchToken(text string, index int) (matchedToken, bool) {
	lower := toLowerASCII(text)
	for _, token := range tokenMatches {
		var end int
		switch token {
		case "go to":
			loc := goToRe.FindStringIndex(lower[index:])
			if loc == nil {
				continue
			}
			end = index + loc[1]
		case "go sub":
			loc := goSubRe.FindStringIndex(lower[index:])
			if loc == nil {
				continue
			}
			end = index + loc[1]
		default:
			if !strings.HasPrefix(lower[index:], token) {
				continue
			}
			end = index + len(token)
		}

		if isAlpha(token[0]) && index > 0 && isAlpha(text[index-1]) {
			continue
		}
		if isAlpha(token[len(token)-1]) && end < len(text) && isAlpha(text[end]) {
			continue
		}
		return matchedToken{end: end, code: tokenCodes[token], text: token}, true
	}
	return matchedToken{}, false
}

// tokenizeLine encodes one Classic BASIC statement's text (everything
// after its line number) into the bytes the ZX81 stores for it,
// mirroring _tokenize_line in tokenizer_zx81.py character by character.
func tokenizeLine(text string, lineNumber int) ([]byte, error) {
	output := make([]byte, 0, len(text)+1)
	index := 0
	inQuotes := false
	inRem := false

	for index < len(text) {
		ch := text[index]

		if strings.HasPrefix(text[index:], `\{`) {
			m := hexEscapeRe.FindStringSubmatchIndex(text[index:])
			if m == nil {
				return nil, tokenizeErrorf(lineNumber, "invalid hexadecimal escape; expected \\{XX}")
			}
			v, _ := strconv.ParseUint(text[index+m[2]:index+m[3]], 16, 8)
			output = append(output, byte(v))
			index += m[1]
			continue
		}

		if inQuotes && !inRem && ch == '"' && index+1 < len(text) && text[index+1] == '"' {
			output = append(output, 192)
			index += 2
			continue
		}

		if ch == '"' {
			inQuotes = !inQuotes
			output = append(output, charCodes['"'])
			index++
			continue
		}

		if !inQuotes && !inRem && isSpace(ch) {
			index++
			continue
		}

		if !inRem && ch == '%' && index+1 < len(text) {
			code, err := characterCode(text[index+1], lineNumber)
			if err != nil {
				return nil, err
			}
			output = append(output, code|0x80)
			index += 2
			continue
		}

		if !inRem && ch == '\\' && index+1 < len(text) {
			escaped := text[index+1]
			if escaped == '"' {
				output = append(output, 192)
				index += 2
				continue
			}
			graphic := ""
			if index+3 <= len(text) {
				graphic = text[index+1 : index+3]
			} else {
				graphic = text[index+1:]
			}
			if code, ok := graphicsCodes[graphic]; ok {
				output = append(output, code)
				index += 3
				continue
			}
			code, err := characterCode(escaped, lineNumber)
			if err != nil {
				return nil, err
			}
			output = append(output, code)
			index += 2
			continue
		}

		if !inQuotes && !inRem {
			if tok, ok := matchToken(text, index); ok {
				output = append(output, tok.code)
				index = tok.end
				if tok.text == "rem" {
					inRem = true
					if index < len(text) && text[index] == ' ' {
						index++
					}
				}
				continue
			}

			var preceding byte
			if index > 0 {
				preceding = text[index-1]
			}
			if !isAlpha(preceding) && isNumberStart(text, index, ch) {
				skip := 0
				if ch == '+' || ch == '-' {
					skip = 1
				}
				loc := numberRe.FindStringIndex(text[index+skip:])
				if loc != nil {
					end := index + skip + loc[1]
					var numberText string
					if ch == '+' || ch == '-' {
						numberText = string(ch) + text[index+skip:end]
					} else {
						numberText = text[index:end]
					}

					if end+5 <= len(text) && strings.EqualFold(text[end:end+5], `\{7e}`) {
						for i := 0; i < len(numberText); i++ {
							code, err := characterCode(numberText[i], lineNumber)
							if err != nil {
								return nil, err
							}
							output = append(output, code)
						}
					} else {
						nb, err := numberBytes(numberText, lineNumber)
						if err != nil {
							return nil, err
						}
						output = append(output, nb...)
					}
					index = end
					continue
				}
			}
		}

		code, err := characterCode(ch, lineNumber)
		if err != nil {
			return nil, err
		}
		output = append(output, code)
		index++
	}

	output = append(output, 0x76)
	if len(output) > 0xFFFF {
		return nil, tokenizeErrorf(lineNumber, "line is too long")
	}
	return output, nil
}

// isNumberStart reports whether a numeric literal starts at text[index]
// (ch == text[index]), mirroring the digit/sign/dot guard inline in
// _tokenize_line.
func isNumberStart(text string, index int, ch byte) bool {
	if isDigit(ch) {
		return true
	}
	if ch == '+' || ch == '-' || ch == '.' {
		if index+1 < len(text) {
			next := text[index+1]
			if isDigit(next) {
				return true
			}
			if next == '.' && index+2 < len(text) && isDigit(text[index+2]) {
				return true
			}
		}
	}
	return false
}

// programImage assembles the complete ".p" memory image from an
// already-tokenized program area, mirroring _program_file in
// tokenizer_zx81.py: 116 bytes of system variables (offsets confirmed
// against the ZX81's own public system-variables table, not just
// re-typed as unnamed magic numbers), the program itself, a blank
// 793-byte display file, and a 2-byte empty-variables marker.
func programImage(program []byte, firstLine int) []byte {
	dataFile := programStart + len(program)
	varsAddress := dataFile + 793

	systemVars := make([]byte, 116)
	putWord := func(offset, value int) {
		systemVars[offset] = byte(value)
		systemVars[offset+1] = byte(value >> 8)
	}
	putWord(1, firstLine)  // first line number of the program (PPC-ish).
	putWord(3, dataFile)   // D_FILE: end of program / start of the rest.
	putWord(5, dataFile+1) // duplicate pointer the ROM also expects here.
	putWord(7, varsAddress)
	putWord(11, varsAddress+1)
	putWord(13, varsAddress+2)
	putWord(17, varsAddress+3) // VARS/E_LINE region start, duplicated.
	putWord(19, varsAddress+3)
	putWord(22, 0x405D)
	systemVars[25] = 2
	systemVars[28] = 0xFF
	systemVars[29] = 0xFF
	systemVars[30] = 0x0F
	systemVars[31] = 55
	putWord(32, dataFile)
	putWord(39, 0x0C6B)
	putWord(43, 0xFD7D)
	systemVars[47] = 0xBC
	systemVars[48] = 0x21
	systemVars[49] = 0x18
	systemVars[50] = 0x40
	systemVars[83] = 0x76

	displayFile := make([]byte, 0, 793)
	displayFile = append(displayFile, 0x76)
	displayLine := make([]byte, 33) // 32 blank cells + line terminator.
	displayLine[32] = 0x76
	for range 24 {
		displayFile = append(displayFile, displayLine...)
	}

	result := make([]byte, 0, 116+len(program)+793+2)
	result = append(result, systemVars...)
	result = append(result, program...)
	result = append(result, displayFile...)
	result = append(result, 0x80, 0x76)
	return result
}

func isAlpha(ch byte) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }
func isDigit(ch byte) bool { return ch >= '0' && ch <= '9' }
func isSpace(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	}
	return false
}
func lowerByte(ch byte) byte {
	if ch >= 'A' && ch <= 'Z' {
		return ch + 32
	}
	return ch
}

// toLowerASCII lowercases only plain ASCII letters, leaving every other
// byte (including Latin-1 bytes >= 0x80) untouched - strings.ToLower is
// avoided here because it decodes runes, which would misinterpret a
// raw byte >= 0x80 as the start of a multi-byte UTF-8 sequence. ZX81
// source text is scanned byte-by-byte throughout this file precisely
// to match the original Python engine's own byte-oriented (`latin1`)
// handling.
func toLowerASCII(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

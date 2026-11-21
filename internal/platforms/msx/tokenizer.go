// Package msx implements the MSX-dialect pieces of the goBaDig
// conversion engine. See OUTLINE.md, "Bloco 2 — Módulo MSX".
package msx

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// base is the fixed MSX BASIC program load address. The very first
// byte of a tokenized .bmx file is a separate file-level marker
// (0xFF) that sits one byte BEFORE this address (at base-1) - it is
// not part of any line's own next-line-pointer arithmetic. Confirmed
// reading the original Python tokenizer (msxbatoken.py, BASE = 0x8001,
// and IO.save_list's own comment placing the 0xFF at BASE-1).
const base = 0x8001

// msxToken is one entry of the token table: the exact uppercase text
// it matches at the start of the remaining source, and the byte(s) it
// emits. Match lengths vary (1-7 characters); emitted byte counts vary
// too (1 byte for most keywords, 2 bytes for the "FF xx" extended
// function space plus ELSE/AS, 3 bytes for LOC( and the "'" comment
// shortcut).
type msxToken struct {
	Match string
	Bytes []byte
}

func hb(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("msx: invalid token literal " + s)
	}
	return b
}

// tokens is a deliberately ORDERED slice, not a map: the original
// Python tokenizer scans this table as a plain list
// (`for command, token in TOKENS: if text.startswith(command): ...`),
// and entry order is load-bearing - e.g. "ERROR" must come before
// "ERR", or the text "ERROR" would wrongly match "ERR" first and leave
// "OR" dangling. Transcribed verbatim, in the same order, from
// badig/msx/msxbatoken/msxbatoken.py (TOKENS, confirmed line by line
// against the source in this session).
var tokens = []msxToken{
	{">", hb("ee")}, {"PAINT", hb("bf")}, {"=", hb("ef")}, {"ERROR", hb("a6")}, {"ERR", hb("e2")}, {"<", hb("f0")}, {"+", hb("f1")},
	{"FIELD", hb("b1")}, {"PLAY", hb("c1")}, {"-", hb("f2")}, {"FILES", hb("b7")}, {"POINT", hb("ed")}, {"*", hb("f3")}, {"POKE", hb("98")},
	{"/", hb("f4")}, {"FN", hb("de")}, {"^", hb("f5")}, {"FOR", hb("82")}, {"PRESET", hb("c3")}, {"\\", hb("fc")}, {"PRINT", hb("91")}, {"?", hb("91")},
	{"PSET", hb("c2")}, {"AND", hb("f6")}, {"GET", hb("b2")}, {"PUT", hb("b3")}, {"GOSUB", hb("8d")}, {"READ", hb("87")}, {"GOTO", hb("89")},
	{"ATTR$", hb("e9")}, {"RENUM", hb("aa")}, {"AUTO", hb("a9")}, {"IF", hb("8b")}, {"RESTORE", hb("8c")}, {"BASE", hb("c9")}, {"IMP", hb("fa")},
	{"RESUME", hb("a7")}, {"BEEP", hb("c0")}, {"INKEY$", hb("ec")}, {"RETURN", hb("8e")}, {"BLOAD", hb("cf")}, {"INPUT", hb("85")},
	{"BSAVE", hb("d0")}, {"INSTR", hb("e5")}, {"RSET", hb("b9")}, {"CALL", hb("ca")}, {"_", hb("5f")}, {"RUN", hb("8a")}, {"IPL", hb("d5")}, {"SAVE", hb("ba")},
	{"KEY", hb("cc")}, {"SCREEN", hb("c5")}, {"KILL", hb("d4")}, {"SET", hb("d2")}, {"CIRCLE", hb("bc")}, {"CLEAR", hb("92")}, {"CLOAD", hb("9b")},
	{"LET", hb("88")}, {"SOUND", hb("c4")}, {"CLOSE", hb("b4")}, {"LFILES", hb("bb")}, {"CLS", hb("9f")}, {"LINE", hb("af")}, {"SPC(", hb("df")},
	{"CMD", hb("d7")}, {"LIST", hb("93")}, {"SPRITE", hb("c7")}, {"COLOR", hb("bd")}, {"LLIST", hb("9e")}, {"CONT", hb("99")}, {"LOAD", hb("b5")},
	{"STEP", hb("dc")}, {"COPY", hb("d6")}, {"LOCATE", hb("d8")}, {"STOP", hb("90")}, {"CSAVE", hb("9a")}, {"CSRLIN", hb("e8")},
	{"STRING$", hb("e3")}, {"LPRINT", hb("9d")}, {"SWAP", hb("a4")}, {"LSET", hb("b8")}, {"TAB(", hb("db")}, {"MAX", hb("cd")}, {"DATA", hb("84")},
	{"MERGE", hb("b6")}, {"THEN", hb("da")}, {"TIME", hb("cb")}, {"TO", hb("d9")}, {"DEFDBL", hb("ae")}, {"DEFINT", hb("ac")}, {"DEFSTR", hb("ab")},
	{"TROFF", hb("a3")}, {"DEFSNG", hb("ad")}, {"TRON", hb("a2")}, {"DEF", hb("97")}, {"MOD", hb("fb")}, {"USING", hb("e4")},
	{"DELETE", hb("a8")}, {"MOTOR", hb("ce")}, {"USR", hb("dd")}, {"DIM", hb("86")}, {"NAME", hb("d3")}, {"DRAW", hb("be")}, {"NEW", hb("94")},
	{"VARPTR", hb("e7")}, {"NEXT", hb("83")}, {"VDP", hb("c8")}, {"DSKI$", hb("ea")}, {"NOT", hb("e0")}, {"DSKO$", hb("d1")}, {"VPOKE", hb("c6")},
	{"OFF", hb("eb")}, {"WAIT", hb("96")}, {"END", hb("81")}, {"ON", hb("95")}, {"WIDTH", hb("a0")}, {"OPEN", hb("b0")}, {"XOR", hb("f8")},
	{"EQV", hb("f9")}, {"OR", hb("f7")}, {"ERASE", hb("a5")}, {"OUT", hb("9c")}, {"ERL", hb("e1")}, {"REM", hb("8f")},

	{"PDL", hb("ffa4")}, {"EXP", hb("ff8b")}, {"PEEK", hb("ff97")}, {"FIX", hb("ffa1")}, {"POS", hb("ff91")}, {"FPOS", hb("ffa7")},
	{"ABS", hb("ff86")}, {"FRE", hb("ff8f")}, {"ASC", hb("ff95")}, {"ATN", hb("ff8e")}, {"HEX$", hb("ff9b")}, {"BIN$", hb("ff9d")},
	{"INP", hb("ff90")}, {"RIGHT$", hb("ff82")}, {"RND", hb("ff88")}, {"INT", hb("ff85")}, {"CDBL", hb("ffa0")}, {"CHR$", hb("ff96")},
	{"CINT", hb("ff9e")}, {"LEFT$", hb("ff81")}, {"SGN", hb("ff84")}, {"LEN", hb("ff92")}, {"SIN", hb("ff89")}, {"SPACE$", hb("ff99")},
	{"SQR", hb("ff87")}, {"LOC(", hb("ffac28")}, {"STICK", hb("ffa2")}, {"COS", hb("ff8c")}, {"LOF", hb("ffad")}, {"STR$", hb("ff93")},
	{"CSNG", hb("ff9f")}, {"LOG", hb("ff8a")}, {"STRIG", hb("ffa3")}, {"LPOS", hb("ff9c")}, {"CVD", hb("ffaa")}, {"CVI", hb("ffa8")},
	{"CVS", hb("ffa9")}, {"TAN", hb("ff8d")}, {"MID$", hb("ff83")}, {"MKD$", hb("ffb0")}, {"MKI$", hb("ffae")}, {"MKS$", hb("ffaf")},
	{"VAL", hb("ff94")}, {"DSKF", hb("ffa6")}, {"VPEEK", hb("ff98")}, {"OCT$", hb("ff9a")}, {"EOF", hb("ffab")}, {"PAD", hb("ffa5")},

	{"'", hb("3a8fe6")}, {"ELSE", hb("3aa1")}, {"AS", hb("4153")},
}

// jumps names every keyword whose trailing numeric argument is a
// branch target (encoded specially - see scanJumpTargets), not an
// ordinary numeric literal. Order doesn't matter here, only
// membership, so a map is fine (unlike tokens above).
var jumps = map[string]bool{
	"RESTORE": true, "AUTO": true, "RENUM": true, "DELETE": true, "RESUME": true, "ERL": true,
	"ELSE": true, "RUN": true, "LIST": true, "LLIST": true, "GOTO": true, "RETURN": true, "THEN": true, "GOSUB": true,
}

// Tokenize converts already-numbered classic MSX Basic source lines
// (no trailing newline characters - one physical line per slice
// element) into the tokenized binary .bmx format.
//
// Deliberate gaps vs. the original msxbatoken.py, none exercised by
// the golden vectors in docs/golden-vectors/msx-tokenizer.md nor by
// real MSX hardware behavior worth reproducing:
//
//   - Scientific-notation literals (1E+10, 1D-5) are not supported yet
//   - they return an error instead of silently mis-encoding.
//   - Numeric literals may not contain embedded whitespace (the
//     original's regexes allow stray spaces between a number's digits
//     and its suffix/exponent, but the original's own consumption-
//     length arithmetic never accounts for those gaps either -
//     porting that faithfully would port a latent bug, not real MSX
//     behavior).
//   - Literals longer than fit in an int64 (~18-19 digits) return an
//     error instead of using arbitrary-precision arithmetic (the
//     original allows up to 10**63-1, a generous sentinel far beyond
//     anything a real Dignified MSX program would contain).
//   - The original's "AS" handling crashes (Python TypeError) on a
//     2-digit trailing number, via a bare ord() call on a
//     multi-character string; this port handles 1 or 2 digits
//     correctly instead of reproducing that crash.
func Tokenize(lines []string) ([]byte, error) {
	var out bytes.Buffer
	out.WriteByte(0xff)

	address := base
	lastLineNumber := -1
	for i, raw := range lines {
		if raw == "" {
			continue
		}
		lineNumber, rest, err := splitLineNumber(raw)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		if lineNumber <= lastLineNumber {
			return nil, fmt.Errorf("line %d: line number out of order: %d", i+1, lineNumber)
		}
		if lineNumber > 65529 {
			return nil, fmt.Errorf("line %d: line number too high: %d", i+1, lineNumber)
		}
		lastLineNumber = lineNumber

		tokenBytes, err := scanLine(rest)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}

		nextAddress := address + 5 + len(tokenBytes)
		out.Write(le16(nextAddress))
		out.Write(le16(lineNumber))
		out.Write(tokenBytes)
		out.WriteByte(0x00)
		address = nextAddress
	}

	out.Write([]byte{0x00, 0x00})
	return out.Bytes(), nil
}

var lineNumberPattern = regexp.MustCompile(`^\s*\d+\s?`)

func splitLineNumber(line string) (int, string, error) {
	m := lineNumberPattern.FindString(line)
	if m == "" {
		return 0, "", fmt.Errorf("line not starting with a number")
	}
	n, err := strconv.Atoi(strings.TrimSpace(m))
	if err != nil {
		return 0, "", fmt.Errorf("invalid line number: %w", err)
	}
	return n, line[len(m):], nil
}

// scanLine tokenizes one line's content (the part after its line
// number has already been split off), mirroring the per-line body of
// msxbatoken.py's Tokenize.tok().
func scanLine(line string) ([]byte, error) {
	var out []byte
	for len(line) > 0 {
		tok, matchLen := findToken(line)
		if tok == nil {
			b, n, err := scanOther(line)
			if err != nil {
				return nil, err
			}
			out = append(out, b...)
			line = line[n:]
			continue
		}

		out = append(out, tok.Bytes...)
		line = line[matchLen:]

		if tok.Match == "AS" {
			b, n := scanASDigits(line)
			out = append(out, b...)
			line = line[n:]
		}

		if jumps[tok.Match] {
			b, n, err := scanJumpTargets(line)
			if err != nil {
				return nil, err
			}
			out = append(out, b...)
			line = line[n:]
		}

		switch tok.Match {
		case "DATA", "REM", "'", "CALL", "_":
			b, n := scanLiteralRun(line, tok.Match)
			out = append(out, b...)
			line = line[n:]
		}
	}
	return out, nil
}

func findToken(line string) (*msxToken, int) {
	for i := range tokens {
		if hasUpperPrefix(line, tokens[i].Match) {
			return &tokens[i], len(tokens[i].Match)
		}
	}
	return nil, 0
}

// scanOther handles everything the token table doesn't match at the
// current position: numbers, &H/&O/&B literals, quoted strings,
// identifiers, and (as a last resort) a single literal character.
func scanOther(line string) ([]byte, int, error) {
	c := line[0]
	switch {
	case c >= '0' && c <= '9' || c == '.':
		return scanNumber(line)
	case c == '&':
		return scanBase(line)
	case c == '"':
		b, n := scanQuotedString(line)
		return b, n, nil
	case toUpperByte(c) >= 'A' && toUpperByte(c) <= 'Z':
		b, n := scanIdentifier(line)
		return b, n, nil
	default:
		return []byte{toUpperByte(c)}, 1, nil
	}
}

func toUpperByte(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 'A'
	}
	return c
}

func hasUpperPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return strings.ToUpper(s[:len(prefix)]) == prefix
}

func le16(v int) []byte {
	return []byte{byte(v & 0xff), byte((v >> 8) & 0xff)}
}

// scanASDigits handles the 1-2 digit numeric suffix on the "AS"
// keyword (e.g. a DEFUSR-adjacent "AS1" form). See the Tokenize
// doc comment: the original crashes on 2 digits, this port doesn't.
var asDigitsPattern = regexp.MustCompile(`^(\s*)(\d{1,2})`)

func scanASDigits(line string) ([]byte, int) {
	m := asDigitsPattern.FindStringSubmatch(line)
	if m == nil {
		return nil, 0
	}
	spaces, digits := m[1], m[2]
	out := bytes.Repeat([]byte{0x20}, len(spaces))
	out = append(out, digits...)
	return out, len(spaces) + len(digits)
}

// scanJumpTargets encodes the line-number argument(s) following a
// branch keyword (THEN/GOTO/GOSUB/...): each target is the literal
// decimal line number re-encoded as a raw little-endian 16-bit value,
// prefixed by a 0x0e marker byte - there is no address resolution or
// symbol table involved at all. Multiple comma-separated targets (ON
// x GOTO 10,20,30) are supported: each run of commas between targets
// is re-emitted as that many literal 0x2c bytes.
var jumpTargetPattern = regexp.MustCompile(`^(\s*)(\d+|,+)`)

func scanJumpTargets(line string) ([]byte, int, error) {
	var out []byte
	consumed := 0
	for {
		m := jumpTargetPattern.FindStringSubmatch(line)
		if m == nil {
			break
		}
		spaces, nugget := m[1], m[2]
		out = append(out, bytes.Repeat([]byte{0x20}, len(spaces))...)
		if nugget[0] >= '0' && nugget[0] <= '9' {
			n, err := strconv.Atoi(nugget)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid branch target: %w", err)
			}
			if n > 65529 {
				return nil, 0, fmt.Errorf("line number jump too high: %d", n)
			}
			out = append(out, 0x0e)
			out = append(out, le16(n)...)
		} else {
			out = append(out, bytes.Repeat([]byte{0x2c}, len(nugget))...)
		}
		advance := len(spaces) + len(nugget)
		line = line[advance:]
		consumed += advance
	}
	return out, consumed, nil
}

// scanLiteralRun echoes characters verbatim until the end of the
// statement, for keywords whose argument is raw text rather than
// tokenized expressions (DATA, REM, the "'" comment shortcut, CALL,
// and the Dignified call-instruction alias "_"). REM/"'" swallow the
// rest of the physical line; DATA/CALL/_ stop early at ":" (and
// CALL/_ also stop at "(") - and CALL/_ uppercase what they echo.
func scanLiteralRun(line, command string) ([]byte, int) {
	var out []byte
	consumed := 0
	for len(line) > 0 {
		c := line[0]
		if command == "CALL" || command == "_" {
			c = toUpperByte(c)
		}
		out = append(out, c)
		line = line[1:]
		consumed++

		if len(line) == 0 {
			break
		}
		stop := (command == "DATA" && line[0] == ':') ||
			(command == "_" && (line[0] == ':' || line[0] == '(')) ||
			(command == "CALL" && (line[0] == ':' || line[0] == '('))
		if stop {
			break
		}
	}
	return out, consumed
}

// scanBase handles &H (hex), &O (octal) and &B (binary) literals.
// &H/&O compute the real integer value and emit it as a 2-byte
// little-endian word behind a type-prefix token. &B does NOT compute
// anything: it stores the literal '0'/'1' characters as raw ASCII,
// behind the literal ASCII bytes for '&' and 'B' themselves (not a
// dedicated token) - this is a real discrepancy from true binary
// encoding, documented in the original's own docstring, and ported
// faithfully rather than "fixed".
var hexBasePattern = regexp.MustCompile(`^[0-9a-fA-F]*`)
var octBasePattern = regexp.MustCompile(`^[0-7]*`)
var binBasePattern = regexp.MustCompile(`^[01]*`)

func scanBase(line string) ([]byte, int, error) {
	prefix := ""
	if len(line) >= 2 {
		prefix = strings.ToUpper(line[:2])
	}
	switch prefix {
	case "&H":
		digits := hexBasePattern.FindString(line[2:])
		b, err := encodeBaseLiteral(digits, 16, "0c")
		if err != nil {
			return nil, 0, err
		}
		return b, 2 + len(digits), nil
	case "&O":
		digits := octBasePattern.FindString(line[2:])
		b, err := encodeBaseLiteral(digits, 8, "0b")
		if err != nil {
			return nil, 0, err
		}
		return b, 2 + len(digits), nil
	case "&B":
		digits := binBasePattern.FindString(line[2:])
		out := []byte{0x26, 0x42}
		out = append(out, digits...)
		return out, 2 + len(digits), nil
	default:
		return []byte{'&'}, 1, nil
	}
}

func encodeBaseLiteral(digits string, base int, prefixHex string) ([]byte, error) {
	prefix := hb(prefixHex)
	if digits == "" {
		return append(append([]byte{}, prefix...), 0x00, 0x00), nil
	}
	n, err := strconv.ParseInt(digits, base, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid base-%d literal: %q", base, digits)
	}
	if n > 65535 {
		return nil, fmt.Errorf("number overflow: %s", digits)
	}
	return append(append([]byte{}, prefix...), le16(int(n))...), nil
}

// scanQuotedString copies a string literal byte-for-byte, including
// its opening and (if present) closing quote - MSX string literals
// are stored literally, with no escaping.
func scanQuotedString(line string) ([]byte, int) {
	var out []byte
	quotes := 0
	i := 0
	for i < len(line) {
		c := line[i]
		if c == '"' {
			quotes++
		}
		out = append(out, c)
		i++
		if quotes > 1 {
			break
		}
	}
	return out, i
}

// scanIdentifier consumes a run of letters/digits as a variable name,
// re-checking the token table at EVERY character: if the remaining
// text starting here matches a keyword, the identifier scan stops
// immediately without consuming it. This reproduces the classic MSX
// Basic gotcha where a variable name containing an embedded keyword
// substring breaks apart - e.g. "SCORE" containing "OR" tokenizes as
// "SC" + OR(token) + "E", not as one identifier.
func scanIdentifier(line string) ([]byte, int) {
	var out []byte
	i := 0
	for i < len(line) {
		remaining := line[i:]
		if tok, _ := findToken(remaining); tok != nil {
			break
		}
		c := toUpperByte(remaining[0])
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')) {
			break
		}
		out = append(out, c)
		i++
	}
	return out, i
}

// numberFirstPattern/numberDecimalPattern mirror msxbatoken.py's own
// number regexes, minus the `\s*` gaps the original allows between a
// number's pieces (see the Tokenize doc comment: not supported here,
// not real MSX behavior, not covered by any golden vector). Trailing
// groups are optional (`.?` instead of `.`) because this port's lines
// have no trailing padding characters to fall back on at end-of-line,
// unlike the original (which always has a trailing "\r\n").
var numberFirstPattern = regexp.MustCompile(`^(\d*)(.?)(.?)`)
var numberDecimalPattern = regexp.MustCompile(`^(\d*)(.?)(\d*)(.?)(.?)`)

func scanNumber(line string) ([]byte, int, error) {
	m := numberFirstPattern.FindStringSubmatch(line)
	integerDigits := m[1]
	rawIntegerDigits := integerDigits
	fractionalDigits := ""
	signal := m[2]
	notifConfirm := m[3]
	numberDigits := integerDigits

	if signal == "." {
		m2 := numberDecimalPattern.FindStringSubmatch(line)
		group1 := m2[1]
		fracRun := m2[3]
		integerForNumber := group1
		if integerForNumber == "" {
			integerForNumber = "0"
		}
		numberDigits = integerForNumber + fracRun
		integerDigits = integerForNumber
		fractionalDigits = "." + fracRun
		signal = m2[4]
		notifConfirm = m2[5]
		rawIntegerDigits = group1
	}

	isInt := false
	switch {
	case signal == "%":
		numberDigits = integerDigits
		isInt = true
	case signal != "%" && signal != "!" && signal != "#" && !isExponentStart(signal, notifConfirm):
		signal = ""
		if fractionalDigits == "" {
			isInt = true
		}
	}

	if isExponentStart(signal, notifConfirm) {
		return nil, 0, fmt.Errorf("scientific notation literals are not supported yet: %q", line)
	}

	if numberDigits == "" {
		numberDigits = "0"
	}
	value, err := strconv.ParseInt(numberDigits, 10, 64)
	if err != nil {
		return nil, 0, fmt.Errorf("number literal too large (not supported): %q", numberDigits)
	}
	if signal == "%" && value >= 32768 {
		return nil, 0, fmt.Errorf("integer overflow: %d", value)
	}

	encoded, newIntegerDigits, err := encodeNumericValue(value, signal, isInt, integerDigits, fractionalDigits, rawIntegerDigits, numberDigits)
	if err != nil {
		return nil, 0, err
	}
	consumed := len(newIntegerDigits) + len(fractionalDigits) + len(signal)
	return encoded, consumed, nil
}

func isExponentStart(signal, notif string) bool {
	s := strings.ToUpper(signal)
	return (s == "E" || s == "D") && (notif == "-" || notif == "+")
}

// encodeNumericValue picks one of five encodings for a numeric
// literal, in the same priority order as msxbatoken.py's elif chain:
// single precision, then double precision, then the three plain
// integer ranges (0-9 get a dedicated single byte, 10-255 a one-byte
// prefix, 256-32767 a two-byte prefix) - note that a plain integer
// literal with no suffix bigger than 32767 is NOT an error: it falls
// through to the single/double precision float encoding instead,
// matching real MSX Basic auto-promotion.
func encodeNumericValue(value int64, signal string, isInt bool, integerDigits, fractionalDigits, rawIntegerDigits, numberDigits string) ([]byte, string, error) {
	switch {
	case (value >= 32768 && value <= 999999 && signal != "#") ||
		signal == "!" ||
		(!isInt && value <= 999999 && signal != "#"):
		hexDigits, newInteger := encodeDecimalFloat("1d", 6, integerDigits, fractionalDigits, rawIntegerDigits, numberDigits)
		b, err := decodeHexPadded(hexDigits, 10, false)
		return b, newInteger, err

	case value >= 1000000 || signal == "#" || !isInt:
		hexDigits, newInteger := encodeDecimalFloat("1f", 14, integerDigits, fractionalDigits, rawIntegerDigits, numberDigits)
		b, err := decodeHexPadded(hexDigits, 18, true)
		return b, newInteger, err

	case value >= 0 && value <= 9:
		return []byte{byte(value + 0x11)}, integerDigits, nil

	case value >= 10 && value <= 255:
		return []byte{0x0f, byte(value)}, integerDigits, nil

	case value >= 256 && value <= 32767:
		return append([]byte{0x1c}, le16(int(value))...), integerDigits, nil

	default:
		return nil, "", fmt.Errorf("number too high: %s", strings.TrimLeft(numberDigits, "0"))
	}
}

// encodeDecimalFloat mirrors msxbatoken.py's parse_sgn_dbl (lines
// 372-393 in the original). Despite OUTLINE.md's earlier assumption,
// this does NOT produce a binary float (neither IEEE-754 nor
// Microsoft Binary Format): MSX Basic's tokenized numeric literals
// store their decimal digits almost as-is - each decimal digit is
// already a valid hex nibble - together with one "exponent" byte
// (0x40 + count of significant digits before the decimal point, with
// special cases for values that are exactly zero or purely
// fractional). header is the type-prefix hex string ("1d" single /
// "1f" double); precision is how many significant decimal digits to
// keep (6 single, 14 double), rounded half-up at that position.
// Parameter names mirror the Python function's own, for traceability.
// Returns the encoded hex-nibble string (not yet padded to its final
// byte length - the caller does that, same as the original) and the
// (possibly reassigned) integerDigits the caller needs for its own
// character-consumption bookkeeping.
func encodeDecimalFloat(header string, precision int, integerDigits, fractionalDigits, rawIntegerDigits, numberDigits string) (string, string) {
	stripped := strings.TrimLeft(integerDigits, "0")
	var exponentByte int
	if stripped == "" {
		fracDigitsOnly := strings.TrimPrefix(fractionalDigits, ".")
		if fracDigitsOnly == "" || allDigitsZero(fracDigitsOnly) {
			exponentByte = 0x00
		} else {
			integerDigits = rawIntegerDigits
			if fracDigitsOnly[0] == '0' {
				trimmedTrailing := strings.TrimRight(fracDigitsOnly, "0")
				leadingZeros := len(trimmedTrailing) - len(strings.TrimLeft(trimmedTrailing, "0"))
				exponentByte = 64 - leadingZeros
			} else {
				exponentByte = 0x40
			}
		}
	} else {
		exponentByte = len(stripped) + 64
	}

	hexDigits := header + fmt.Sprintf("%02x", exponentByte)
	cropped := numberDigits
	if n, err := strconv.ParseInt(numberDigits, 10, 64); err == nil {
		cropped = strconv.FormatInt(n, 10)
	}
	roundDigit := byte(0)
	if precision < len(cropped) && cropped[precision] >= '0' && cropped[precision] <= '9' {
		roundDigit = cropped[precision] - '0'
	}
	limit := min(precision, len(cropped))
	if roundDigit < 5 {
		hexDigits += cropped[:limit]
	} else {
		n, _ := strconv.ParseInt(cropped[:limit], 10, 64)
		hexDigits += strconv.FormatInt(n+1, 10)
	}
	return hexDigits, integerDigits
}

func allDigitsZero(digits string) bool {
	for _, d := range digits {
		if d != '0' {
			return false
		}
	}
	return true
}

func decodeHexPadded(hexDigits string, totalLen int, truncate bool) ([]byte, error) {
	for len(hexDigits) < totalLen {
		hexDigits += "0"
	}
	if truncate && len(hexDigits) > totalLen {
		hexDigits = hexDigits[:totalLen]
	}
	return hex.DecodeString(hexDigits)
}

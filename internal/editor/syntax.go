package edit

import (
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/man"
)

// syntaxDialect is the BASIC Dignified dialect a document's syntax
// highlighting rules are drawn from, based on the open file's
// extension. The embedded WibEdit editor has no notion of "dialect"
// otherwise - this is the one addition that gives it one, purely for
// coloring (no compilation, no validation: an unrecognized or
// malformed line just doesn't light up, same as any plain-text
// editor's highlighter degrading gracefully on bad input).
type syntaxDialect int

const (
	dialectNone syntaxDialect = iota
	dialectMSX
	dialectZX81
	dialectZXSpectrum
	dialectASM
)

// dialectForPath picks a dialect from a file's extension - both the
// Dignified source extension and the classic (numbered-listing)
// output extension for each platform, since a user can open either
// in WibEdit (F4 on any file, not just sources). ".asm" is asMSX's own
// source extension (Fase 4/5 of OUTLINE.md's plan) - not a Dignified
// dialect (internal/dialect.go deliberately excludes it), but WibEdit
// highlights it anyway since it is a file type a project can contain.
func dialectForPath(path string) syntaxDialect {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".dmx", ".amx":
		return dialectMSX
	case ".dzx", ".azx":
		return dialectZX81
	case ".dzs", ".azs":
		return dialectZXSpectrum
	case ".asm":
		return dialectASM
	}
	return dialectNone
}

// syntaxKind is the highlight category a span of a line's text falls
// into. Deliberately coarse (six categories) rather than mirroring
// every scope the original Sublime/VSCode grammars define (operators,
// "entity.name.tag" for Dignified constructs, invalid-input
// detection, etc.) - those grammars are built for a real IDE's
// language server; this is a lightweight, redraw-every-keystroke
// highlighter for a terminal editor, where keyword/function/comment/
// string/number/label is most of the visual value for the least
// implementation risk. Autocomplete and context help (the user's own
// stated next steps) are a better place to reuse the finer-grained
// distinctions later, if ever.
type syntaxKind int

const (
	synNone syntaxKind = iota
	synKeyword
	synFunction
	synComment
	synString
	synNumber
	synLabel
	synConstant
)

// syntaxSpan is one highlighted run within a single line's own text,
// as BYTE offsets relative to that line's start (not the whole
// document) - the caller adds the line's own start offset to get
// document-absolute positions, same unit documentCell.start/end
// already use.
type syntaxSpan struct {
	start, end int
	kind       syntaxKind
}

// tokenizeLine scans one line of text and returns the highlighted
// spans in it, in order. Returns nil for dialectNone (no highlighting
// at all - the common case for a plain-text file) and for a blank
// line. Never errors: worst case for malformed/mid-edit input is that
// a span just doesn't get colored, exactly like any other editor's
// highlighter would degrade.
//
// Deliberately stateless across lines - a known, scoped-out-for-v1
// consequence (matches the user's own "syntax highlighting first,
// autocomplete/context help later" framing for this whole feature):
// this project's own multi-line block comments (a doubled apostrophe
// opening and closing one, or "###...###") aren't tracked as
// open/closed, so a line inside one of those blocks is still scanned
// normally - an English word inside a comment block that happens to
// also be a BASIC keyword (ON, TO, FOR,
// AND...) gets keyword-colored even though the whole block is really
// a comment. Properly tracking this needs a scan of every line from
// the top of the document (or a cached per-document "which lines are
// inside an unclosed block" index invalidated on edits), not just the
// handful of currently-visible lines this renders - a reasonable
// follow-up, not a v1 requirement.
func tokenizeLine(dialect syntaxDialect, text string) []syntaxSpan {
	if dialect == dialectNone || text == "" {
		return nil
	}
	var spans []syntaxSpan
	n := len(text)
	for i := 0; i < n; {
		c := text[i]
		switch {
		case c == ' ' || c == '\t':
			i++
		case c == '\'':
			spans = append(spans, syntaxSpan{i, n, synComment})
			i = n
		case c == '#' && i+1 < n && text[i+1] == '#':
			spans = append(spans, syntaxSpan{i, n, synComment})
			i = n
		case dialect == dialectASM && c == ';':
			spans = append(spans, syntaxSpan{i, n, synComment})
			i = n
		case dialect == dialectASM && c == '/' && i+1 < n && text[i+1] == '/':
			spans = append(spans, syntaxSpan{i, n, synComment})
			i = n
		case dialect == dialectASM && c == '-' && i+1 < n && text[i+1] == '-':
			spans = append(spans, syntaxSpan{i, n, synComment})
			i = n
		// Same-line only: /* ... */ left unclosed on this line is a
		// known, deliberate gap (see the func doc comment above on
		// stateless-across-lines scanning), same tradeoff as every
		// other multi-line construct this highlighter doesn't track.
		case dialect == dialectASM && c == '/' && i+1 < n && text[i+1] == '*':
			start := i
			if end := strings.Index(text[i:], "*/"); end >= 0 {
				i += end + 2
			} else {
				i = n
			}
			spans = append(spans, syntaxSpan{start, i, synComment})
		// asMSX uses "{ }" as a Pascal-style comment delimiter (asmsx.md
		// section 2.6) - checked ahead of the unconditional "{" case
		// below, which is a different dialect's label-anchor syntax.
		case dialect == dialectASM && c == '{':
			start := i
			i++
			for i < n && text[i] != '}' {
				i++
			}
			if i < n {
				i++
			}
			spans = append(spans, syntaxSpan{start, i, synComment})
		case c == '"':
			start := i
			i++
			for i < n && text[i] != '"' {
				i++
			}
			if i < n {
				i++
			}
			spans = append(spans, syntaxSpan{start, i, synString})
		case c == '{':
			start := i
			i++
			for i < n && text[i] != '}' {
				i++
			}
			if i < n {
				i++
			}
			spans = append(spans, syntaxSpan{start, i, synLabel})
		case dialect == dialectMSX && c == '&' && i+1 < n && isHexOctalBinaryPrefix(text[i+1]):
			start := i
			i += 2
			for i < n && isHexDigit(text[i]) {
				i++
			}
			spans = append(spans, syntaxSpan{start, i, synNumber})
		// Pascal-style "$8a" hex (asmsx.md section 2.3.1); a bare "$"
		// (the current-address pseudo-variable) is left uncolored.
		case dialect == dialectASM && c == '$' && i+1 < n && isHexDigit(text[i+1]):
			start := i
			i++
			for i < n && isHexDigit(text[i]) {
				i++
			}
			spans = append(spans, syntaxSpan{start, i, synNumber})
		// A directive used with its optional leading "." (asmsx.md:
		// "you can use them without first point character"), or a
		// "."-prefixed local label (asmsx.md section 2.2, the other
		// convention besides the "@@" prefix autocomplete doesn't need
		// to special-case - it already falls through to the plain
		// word branch below and gets colored there once its own ":"
		// is seen).
		case dialect == dialectASM && c == '.' && i+1 < n && isWordStart(text[i+1]):
			start := i
			i++
			wordStart := i
			i = scanWord(text, i)
			bare := strings.ToUpper(text[wordStart:i])
			switch {
			case asmsxDirectives[bare]:
				spans = append(spans, syntaxSpan{start, i, synKeyword})
			case i < n && text[i] == ':':
				spans = append(spans, syntaxSpan{start, i, synLabel})
			}
		case dialect == dialectASM && isDigit(c):
			start := i
			i = scanAsmNumber(text, i)
			spans = append(spans, syntaxSpan{start, i, synNumber})
		case isDigit(c):
			start := i
			i = scanNumber(text, i)
			spans = append(spans, syntaxSpan{start, i, synNumber})
		case isWordStart(c):
			start := i
			i = scanWord(text, i)
			if dialect != dialectASM && i < n && text[i] == '$' {
				i++
			}
			word := text[start:i]
			if strings.EqualFold(word, "REM") {
				spans = append(spans, syntaxSpan{start, n, synComment})
				i = n
				continue
			}
			// A trailing ":" immediately after the word, with no
			// space, is asmsx's own label-definition syntax (asmsx.md
			// section 2.2) - checked ahead of lookupWord so a label
			// that happens to share a mnemonic/directive/BIOS name
			// still colors as a label, not as whatever it collides
			// with (asmsx itself resolves this the same way: the
			// trailing colon is what makes it a label).
			if dialect == dialectASM && i < n && text[i] == ':' {
				spans = append(spans, syntaxSpan{start, i, synLabel})
				continue
			}
			if kind := lookupWord(dialect, word); kind != synNone {
				spans = append(spans, syntaxSpan{start, i, kind})
			}
		default:
			i++
		}
	}
	return spans
}

func isDigit(c byte) bool    { return c >= '0' && c <= '9' }
func isHexDigit(c byte) bool { return isDigit(c) || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f') }
func isHexOctalBinaryPrefix(c byte) bool {
	return c == 'H' || c == 'h' || c == 'O' || c == 'o' || c == 'B' || c == 'b'
}

func isWordStart(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isWordChar(c byte) bool {
	return isWordStart(c) || isDigit(c) || c == '_'
}

// scanNumber scans a decimal literal: digits, an optional decimal
// point and more digits, an optional exponent (E/D, +/-, digits), and
// an optional type-suffix character (%!#). Good enough for
// highlighting purposes across all three dialects; doesn't need to
// reject malformed numbers; a real lexer validates that at compile
// time.
func scanNumber(text string, i int) int {
	n := len(text)
	for i < n && isDigit(text[i]) {
		i++
	}
	if i < n && text[i] == '.' {
		i++
		for i < n && isDigit(text[i]) {
			i++
		}
	}
	if i < n && (text[i] == 'E' || text[i] == 'e' || text[i] == 'D' || text[i] == 'd') {
		j := i + 1
		if j < n && (text[j] == '+' || text[j] == '-') {
			j++
		}
		if j < n && isDigit(text[j]) {
			i = j
			for i < n && isDigit(text[i]) {
				i++
			}
		}
	}
	if i < n && (text[i] == '%' || text[i] == '!' || text[i] == '#') {
		i++
	}
	return i
}

func scanWord(text string, i int) int {
	n := len(text)
	for i < n && isWordChar(text[i]) {
		i++
	}
	return i
}

// scanAsmNumber scans one of asmsx.md's four numeric-literal
// conventions (section 2.3.1), given that text[i] is already known to
// be a decimal digit (asmsx requires a numeric literal using the
// trailing-letter conventions to start with a digit, same rule this
// relies on to avoid colliding with a hex-looking label like "FFh"):
//
//   - "0x8a" (C-style hex): a literal "0x"/"0X" prefix, then hex digits.
//   - a decimal float: digits, ".", more digits (no exponent letter in
//     asmsx's own grammar, unlike the BASIC dialects' scanNumber).
//   - digits followed by a single trailing radix letter - hex "h"/"H",
//     octal "o"/"O" or binary "b"/"B". Hex digits (a-f/A-F) are
//     consumed greedily before checking for this suffix, which is what
//     lets "0ffh" scan as one token instead of stopping at "0ff" and
//     leaving a stray "h".
//   - a plain decimal (or C-style leading-zero octal, "077" -
//     highlighted the same as decimal; telling them apart isn't needed
//     for coloring purposes) when none of the above suffixes follow.
func scanAsmNumber(text string, i int) int {
	n := len(text)
	if text[i] == '0' && i+1 < n && (text[i+1] == 'x' || text[i+1] == 'X') {
		i += 2
		for i < n && isHexDigit(text[i]) {
			i++
		}
		return i
	}
	for i < n && isHexDigit(text[i]) {
		i++
	}
	if i < n && text[i] == '.' {
		i++
		for i < n && isDigit(text[i]) {
			i++
		}
		return i
	}
	if i < n && (text[i] == 'h' || text[i] == 'H' || text[i] == 'o' || text[i] == 'O' ||
		text[i] == 'b' || text[i] == 'B') {
		i++
	}
	return i
}

// asmBiosNames are asMSX's predefined ".BIOS"/".BIOSVARS" symbol names
// (internal/man), cached once (built from a map copy on every
// keystroke would be wasteful) for lookupWord's dialectASM case below.
var asmBiosNames = man.AsmBiosSymbolNames()

func lookupWord(dialect syntaxDialect, word string) syntaxKind {
	upper := strings.ToUpper(word)
	switch dialect {
	case dialectMSX:
		if msxFunctions[upper] || isUSRFunction(upper) {
			return synFunction
		}
		if msxKeywords[upper] {
			return synKeyword
		}
	case dialectZX81:
		if zx81Functions[upper] {
			return synFunction
		}
		if zx81Keywords[upper] {
			return synKeyword
		}
	case dialectZXSpectrum:
		if zxSpectrumFunctions[upper] {
			return synFunction
		}
		if zxSpectrumKeywords[upper] {
			return synKeyword
		}
	case dialectASM:
		if asmBiosNames[upper] {
			return synConstant
		}
		if asmsxNumericFunctions[upper] {
			return synFunction
		}
		if z80Mnemonics[upper] || z80Registers[upper] || asmsxDirectives[upper] {
			return synKeyword
		}
	}
	return synNone
}

// isUSRFunction matches MSX's USR0..USR9 (and bare USR) function
// names - the one dialect with digit-suffixed function names.
func isUSRFunction(upper string) bool {
	if !strings.HasPrefix(upper, "USR") {
		return false
	}
	rest := upper[len("USR"):]
	if rest == "" {
		return true
	}
	return len(rest) == 1 && isDigit(rest[0])
}

// spanKindAt returns the highlight kind covering offset (a byte
// offset relative to the same line tokenizeLine was given), or
// synNone if offset falls outside every span. spans are produced in
// left-to-right order by tokenizeLine, so a linear scan is simplest;
// a line only ever has a handful of spans, so this isn't worth a
// binary search.
func spanKindAt(spans []syntaxSpan, offset int) syntaxKind {
	for _, s := range spans {
		if offset >= s.start && offset < s.end {
			return s.kind
		}
	}
	return synNone
}

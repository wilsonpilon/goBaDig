// Package core implements the platform-agnostic pieces of the BASIC
// Dignified conversion engine: the token/position model, the
// Lexer/Parser hook contract, diagnostics, and the pipeline that
// drives them. See OUTLINE.md, "Bloco 1 — Núcleo do motor", for the
// architecture this translates from the original Python engine.
package core

import (
	"fmt"
	"regexp"
)

// Line is one line of source code, carrying its original line number
// and the file it came from (for error reporting and INCLUDE
// tracking). Index 0 of a Source slice is always the synthetic
// PROGRAM entry a CompileContext starts from, mirroring the original
// engine's IO.load_file (support/helper.py).
type Line struct {
	Number int
	Text   string
	File   string
}

// TokenKind identifies what a Token represents. It is a string, not an
// int enum, because the set of kinds is open: internal/core defines
// only the structural kinds every platform shares (Program, Newline,
// EOF, Word); each platform package defines its own kinds in its own
// namespace without colliding with core's or another platform's -
// mirrors the original engine's Token.tok, a free-form string matched
// by comparison, never a closed enum.
type TokenKind string

const (
	// KindProgram marks the synthetic first token of a token stream -
	// a stable "line zero" anchor, never a real piece of source.
	KindProgram TokenKind = "PROGRAM"
	// KindNewline marks the end of a source line.
	KindNewline TokenKind = "NEWLINE"
	// KindEOF marks the end of the token stream.
	KindEOF TokenKind = "EOF"
	// KindIdentifier is a Dignified identifier (`^[a-z][a-z_0-9]*[%!#$]?`)
	// - the same pattern for every platform (badig_dignified.py), so
	// core.Lex recognizes it itself as a fallback, after the platform's
	// Lexer.Lex got first refusal (see Lexer, in contract.go).
	KindIdentifier TokenKind = "IDENTIFIER"
	// KindDignifiedSymbol is one of the 6 Dignified structural symbols
	// (`~ [ ] { } @` - d_symbol in badig_dignified.py): shared by every
	// platform, so core.Lex recognizes them itself, same reasoning as
	// KindIdentifier. Only `~` (keep-long-name) is acted on by any
	// Parser yet (MSX Pass 3, Bloco 2) - the other five (define/label
	// delimiters) are recognized but unused so far.
	KindDignifiedSymbol TokenKind = "DIGNIFIED_SYMBOL"
	// KindString is a quoted string literal, delimited by the
	// platform Description's QuoteChar, Value including both
	// delimiters verbatim (no escaping, same as the original engine).
	KindString TokenKind = "STRING"
	// KindUnknown is a single raw character core.Lex could not
	// classify by any means (not the platform's, not an identifier,
	// not a quoted string) - mirrors the original engine's "character
	// not recognized" diagnostic path, but tolerant: scanning
	// continues instead of aborting.
	KindUnknown TokenKind = "UNKNOWN"
)

// Position pins a Token to an exact place in the source, for error
// reporting. Source is a reference to the whole listing the Token
// came from (shared, not copied), so a Diagnostic can render the
// offending line's text.
type Position struct {
	Source []Line
	Line   int // index into Source
	Column int // 1-based
	Offset int // leading-whitespace compensation for the "^" error pointer
}

// Text returns the source text of the line this Position points at,
// or "" if Source doesn't cover it.
func (p Position) Text() string {
	if p.Line < 0 || p.Line >= len(p.Source) {
		return ""
	}
	return p.Source[p.Line].Text
}

// File returns the file this Position's line came from.
func (p Position) File() string {
	if p.Line < 0 || p.Line >= len(p.Source) {
		return ""
	}
	return p.Source[p.Line].File
}

// Token is the unit the lexer produces and the parser passes rewrite.
// Value is `any`, not `string`: lexing always produces text, but a
// later pass (resolving a Dignified label to a line number) replaces
// it with an int - confirmed reading the original engine's Pass 4,
// which reuses Token.val the same way.
type Token struct {
	Kind  TokenKind
	Value any
	Pos   Position
}

// String renders Value as text regardless of its underlying type.
func (t Token) String() string {
	return fmt.Sprint(t.Value)
}

// ParseIdentifier splits an identifier's value into its name and type
// suffix (e.g. MSX "a$" -> "a", "$") using idRegex, which must have
// exactly two capture groups: name and suffix. It is a free function,
// not a Token method, so a Token never needs to know its own grammar -
// the original engine's Token constructor builds a throwaway
// Description() just to fetch this regex on every token; passing it
// in avoids that wart in the port.
func ParseIdentifier(value string, idRegex *regexp.Regexp) (name, suffix string, ok bool) {
	m := idRegex.FindStringSubmatch(value)
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}

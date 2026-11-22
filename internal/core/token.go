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
	// KindDignifiedOperator is TRUE/FALSE (d_operat in
	// badig_dignified.py): shared by every platform, case-insensitive,
	// matched as a whole word (not a prefix) - same reasoning as
	// KindIdentifier/KindDignifiedSymbol. MSX Pass 5 (Bloco 2) is the
	// first - so far only - Parser pass that acts on it.
	KindDignifiedOperator TokenKind = "DIGNIFIED_OPERATOR"
	// KindDignifiedSeparator is `_` (d_instsp in badig_dignified.py):
	// shared by every platform, same reasoning as KindDignifiedSymbol.
	// At the end of a line (immediately followed by KindNewline) it
	// merges that line into the next one (core Pass 3, Bloco 1/2 -
	// "labels e numeração"); anywhere else it is MSX's own "_ as a
	// call separator" concern (Pass 1, still not implemented - see
	// OUTLINE.md).
	KindDignifiedSeparator TokenKind = "DIGNIFIED_SEPARATOR"
	// KindLabelDeclaration is a resolved `{name}` (or loop-label, not
	// implemented yet) appearing at the start of a line (D_LBLLIN in
	// badig.py) - collapsed from 3 tokens (`{`, name, `}`) into one by
	// core Pass 2. Removed entirely by core Pass 4 once its line
	// number is recorded (a label never gets a line of its own - see
	// OUTLINE.md, "labels e numeração").
	KindLabelDeclaration TokenKind = "LABEL_DECLARATION"
	// KindLabelJump is a resolved `{name}` (or `{@}`, the current-line
	// self-reference) appearing anywhere else (D_LBLJMP in badig.py) -
	// same 3-token collapse as KindLabelDeclaration. Core Pass 4
	// replaces its Value with the resolved line number (int); a
	// reference to a name that's never declared is a fatal error.
	KindLabelJump TokenKind = "LABEL_JUMP"
	// KindLineNumber is a synthetic token (C_LINENB in badig.py) core
	// Pass 4 inserts right after every KindNewline that starts a new
	// classic line, carrying that line's number (int) as Value.
	KindLineNumber TokenKind = "LINE_NUMBER"
	// KindDignifiedInstruction is one of the 8 shared Dignified
	// instruction words (d_instrc in badig_dignified.py: DEFINE,
	// DECLARE, INCLUDE, KEEP, ENDIF, FUNC, RET, EXIT), case-insensitive,
	// matched as a whole word - same reasoning as KindDignifiedOperator.
	// Only EXIT (loop-label exit, core Pass 2) is acted on so far; the
	// other 7 are recognized but unused until DEFINE/FUNC/etc. exist.
	KindDignifiedInstruction TokenKind = "DIGNIFIED_INSTRUCTION"
	// KindLoopReturn is a loop label's closing `}` (D_LBLRET in
	// badig.py), resolved exactly like KindLabelJump (against the
	// loop-opener's own line) but also recording, as a side effect,
	// the line right after it - what a KindLoopExit inside the same
	// loop resolves to.
	KindLoopReturn TokenKind = "LOOP_RETURN"
	// KindLoopExit is an `exit` inside a loop label (D_LBLEXT in
	// badig.py) - resolves to the line right after the innermost open
	// loop's closing `}` (recorded by that KindLoopReturn), not to a
	// declared label's own line.
	KindLoopExit TokenKind = "LOOP_EXIT"
	// KindSynthetic is literal text core itself inserts (not scanned
	// from source) - e.g. the statement separator and "loop back"
	// instruction a loop label's closing `}` emits (core Pass 2,
	// sourced from Description.StatementSeparator/LoopBackInstruction,
	// since those vary by platform). Deliberately generic: nothing
	// downstream needs to tell these apart by Kind yet (today's
	// minimal core.Generate doesn't distinguish any Kind for spacing -
	// that's real generate() work, not done yet, ver OUTLINE.md).
	KindSynthetic TokenKind = "SYNTHETIC"
	// KindDefinePlaceholder marks, inside a DEFINE's stored content
	// (D_DEFVAR in badig.py), the position of its one substitutable
	// argument - e.g. the `[0]` in `DEFINE [var0][poke MLs+0,[0]]`.
	// Purely internal to core Pass 1's DEFINE expansion
	// (pass1Defines, pipeline.go): never reaches core.Parse's output,
	// since every occurrence is always substituted away when a call
	// (`[var0]` or `[var0](&h20)`) expands.
	KindDefinePlaceholder TokenKind = "DEFINE_PLACEHOLDER"
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

package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// identifierPattern is the shared Dignified identifier pattern - the
// SAME for every platform (badig_dignified.py's d_idnttp), which is
// why core.Lex recognizes it itself instead of asking the platform.
// Case-insensitive: d_idnttp's `[a-z]` looks lowercase-only, but it's
// compiled inside the original engine's single combined regex with
// `re.I` (badig.py:127-132), so it actually matches any case - real
// Dignified source routinely uses camelCase/mixed-case names
// (confirmed in badig/examples/msx/CGK173.dmx: "gameName$",
// "fill_VRAM", "pauseKEY", the label "portuguese_ROM", etc.) - a
// strictly-lowercase pattern here would cut those at the first
// uppercase letter.
var identifierPattern = regexp.MustCompile(`(?i)^[a-z][a-z_0-9]*[%!#$]?`)

// dignifiedSymbols are badig_dignified.py's d_symbol: 6 structural
// symbols shared by every dialect (not platform vocabulary), so
// core.Lex recognizes them itself, same reasoning as identifierPattern.
// `~` (keep a long variable name verbatim) is the only one any Parser
// acts on so far (MSX Pass 3, Bloco 2); the rest (define/label
// delimiters) are recognized but unused until DEFINE/labels exist.
var dignifiedSymbols = map[byte]bool{
	'~': true, '[': true, ']': true, '{': true, '}': true, '@': true,
}

// dignifiedOperatorWord extracts a leading run of ASCII letters (not a
// platform's job - this is checked before any platform gets a chance,
// same priority as dignifiedSymbols/identifierPattern below).
var dignifiedOperatorWord = regexp.MustCompile(`^[A-Za-z]+`)

// dignifiedOperators are badig_dignified.py's d_operat (TRUE/FALSE) -
// shared by every dialect, matched as a whole word (not a prefix, so
// "truest" never matches "TRUE"), same reasoning as dignifiedSymbols.
var dignifiedOperators = map[string]bool{"TRUE": true, "FALSE": true}

// dignifiedInstructions are badig_dignified.py's d_instrc: 8 shared
// Dignified instruction words, matched the same way as
// dignifiedOperators. Only EXIT/DEFINE/FUNC/RET have a Parser acting
// on them so far; DECLARE/INCLUDE/KEEP/ENDIF are recognized but
// unused still.
var dignifiedInstructions = map[string]bool{
	"DEFINE": true, "DECLARE": true, "INCLUDE": true, "KEEP": true,
	"ENDIF": true, "FUNC": true, "RET": true, "EXIT": true,
}

// functionNamePattern is badig_dignified.py's d_funcnm (`^\.{d_idnocaps}$`,
// also compiled under re.I - ver identifierPattern's own doc comment
// for why that matters): a `.` followed by a letter-starting word -
// shared Dignified vocabulary (a FUNC's own declared name, or a call
// to it), not platform-specific. Confirmed used in
// badig/examples/msx/CGK173.dmx (".move_cursor_files", ".show_content",
// etc.). Doesn't collide with a leading-dot decimal number (e.g.
// ".5") - the platform's own Lexer.Lex already claims that first,
// since numbers are platform vocabulary and get first refusal.
var functionNamePattern = regexp.MustCompile(`(?i)^\.[a-z][a-z_0-9]*`)

// toggleRemPattern is badig_dignified.py's d_togrem (`^#{d_idnocaps}$`,
// also under re.I): `#` followed by a letter-starting word - shared
// Dignified vocabulary (a KEEP declaration's tag, or a toggle-block/
// line marker like "#a"). The platform's own Lexer.Lex gets first
// refusal as always; MSX's (internal/platforms/msx/lexer.go) was
// taught to refuse "#"+letter specifically, since "#" alone is ALSO
// real MSX vocabulary (file channel refs like "#1") - this fallback
// only ever gets a chance once that refusal happens.
var toggleRemPattern = regexp.MustCompile(`(?i)^#[a-z][a-z_0-9]*`)

// dignifiedLineCommentPrefix is badig_dignified.py's d_linrem ("##",
// D_LINREM) - a Dignified-exclusive comment that works ANYWHERE on a
// physical line (confirmed used both standalone and as a trailing
// comment after real code in badig/examples/msx/CGK173.dmx),
// consuming the rest of that line with zero tokens - badig.py's own
// lex() just advances its cursor to end-of-line and never appends
// anything for it (`self.pos.col = len(self.pos.text) - 1; continue`).
// "##BB:tag=value" (a remtag, support/badig_settings.py's
// read_remtags_from_code) is just this same "##" to the lexer/parser
// - its own extraction happens in a completely separate pass over the
// raw file text, part of the 4-layer config system already deferred
// since Bloco 1/Fase 4; this port only needs to not choke on the
// syntax, not apply it.
const dignifiedLineCommentPrefix = "##"

// dignifiedBlockCommentMarker is badig_dignified.py's d_o_rems
// ("###", D_O_EREM) - the Dignified-exclusive counterpart to MSX's
// own classic block comment (two apostrophes, Description.CommentChar()
// doubled) - ver findBlockCloseLine's own doc comment for the shared
// mechanics and get_lit_block (badig.py:331-380).
const dignifiedBlockCommentMarker = "###"

// scanCommentBlock mirrors get_lit_block's own line-by-line scan
// (badig.py:331-380) for a Dignified/classic comment BLOCK: openRest
// is whatever comes right after the marker on the OPENING line itself
// (already known to start at the beginning of that line); it scans
// forward - starting with openRest, then ctx.Source's own raw line
// text from lineIndex+1 onward - for a line ENDING in marker (not
// needing to be alone: real content can share that same line before
// it, confirmed used for real in badig/examples/msx/CGK173.dmx, e.g.
// a block closing with "...VRAM=&h1940###"). Returns every content
// chunk in between and the index of the line the close marker was
// found on, or ok=false if the end of the file is reached first.
//
// The OPENING line's own remainder (openRest) and the CLOSING line's
// own remainder-before-the-marker are each skipped when empty (marker
// alone, nothing else on that line) - confirmed empirically by
// running badig.py itself against a minimal "”"-block repro: a
// marker alone on its opening/closing line produces NO extra blank
// comment line, but a genuinely blank INTERIOR line still does (it's
// real content, just empty, same as any other content line - only
// the two boundary chunks get this special treatment).
func scanCommentBlock(ctx *CompileContext, lineIndex int, openRest, marker string) (content []string, closeLine int, ok bool) {
	chunk := openRest
	li := lineIndex
	first := true
	for {
		if strings.HasSuffix(chunk, marker) {
			inner := chunk[:len(chunk)-len(marker)]
			if strings.TrimSpace(inner) != "" {
				content = append(content, inner)
			}
			return content, li, true
		}
		if !first || strings.TrimSpace(chunk) != "" {
			content = append(content, chunk)
		}
		first = false
		li++
		if li >= len(ctx.Source) {
			return nil, 0, false
		}
		chunk = ctx.Source[li].Text
	}
}

// Lex scans ctx.Source into a flat token stream: a synthetic PROGRAM
// token first, then one token per recognized item plus a NEWLINE per
// source line, ending in EOF.
//
// At each non-whitespace position, platform.Lexer.Lex gets first
// refusal on the rest of the current line (see Lexer, in contract.go,
// for why it gets the whole remaining line rather than a pre-cut
// word - validating this against MSX, a real platform, showed a
// pre-cut word doesn't work). If it doesn't recognize anything there,
// core tries its own generic fallbacks, in order: a Dignified
// identifier (shared by every platform), a quoted string (delimited
// by platform.Description.QuoteChar), and finally a single raw,
// unclassified character - mirroring the original engine's priority
// (dialect keywords before the shared identifier pattern) and its
// tolerant "character not recognized" handling (scanning continues
// rather than aborting).
//
// LoadLines reads path from disk into Source-ready Lines: the real
// (non-test) implementation for CompileContext.LoadFile, used by
// INCLUDE (pass3Includes) and reusable later for loading the main
// file too, once there is real CLI wiring. Mirrors the Line shape
// every test helper in pipeline_test.go already uses: a dummy
// Line{Number: 0} placeholder first (core.Lex starts scanning real
// content at index 1), then one Line per real line, 1-indexed.
// tabSize expands every TAB the same way support/helper.py's
// IO.load_file does (`line.expandtabs(tab_lenght)`, badig.py:88) -
// ver expandTabs' own doc comment. Deliberately out of scope, same
// as the original engine's own IO.load_file options this doesn't
// need yet: file encoding normalization (CompileSettings doesn't
// model that).
//
// expandTabs mirrors Python's str.expandtabs(tabsize) exactly as
// support/helper.py's IO.load_file uses it
// (`line.expandtabs(tab_lenght)`, badig.py:88): replaces each TAB
// with just enough spaces to reach the next tabSize-column boundary,
// tracking column position as it scans so a TAB coming after other
// TABs/text still lands on the right stop. tabSize<=0 is a
// degenerate case Python itself doesn't sanely support either
// (division-by-zero territory for its own algorithm) - left
// unexpanded here rather than guessing a replacement behavior.
func expandTabs(s string, tabSize int) string {
	if tabSize <= 0 || !strings.ContainsRune(s, '\t') {
		return s
	}
	var b strings.Builder
	col := 0
	for _, r := range s {
		if r == '\t' {
			n := tabSize - (col % tabSize)
			b.WriteString(strings.Repeat(" ", n))
			col += n
			continue
		}
		b.WriteRune(r)
		col++
	}
	return b.String()
}

func LoadLines(path string, tabSize int) ([]Line, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rawLines := strings.Split(string(data), "\n")
	lines := make([]Line, 0, len(rawLines)+1)
	lines = append(lines, Line{Number: 0, Text: "PROGRAM", File: path})
	for i, text := range rawLines {
		text = expandTabs(strings.TrimRight(text, "\r"), tabSize)
		lines = append(lines, Line{Number: i + 1, Text: text, File: path})
	}
	return lines, nil
}

func Lex(ctx *CompileContext, platform Platform) ([]Token, error) {
	platform.Lexer.Init(ctx)

	tokens := []Token{{
		Kind:  KindProgram,
		Value: "PROGRAM",
		Pos:   Position{Source: ctx.Source, Line: 0, Column: 1},
	}}
	quote := platform.Description.QuoteChar()

	for lineIndex := 1; lineIndex < len(ctx.Source); lineIndex++ {
		line := ctx.Source[lineIndex]
		runes := []rune(line.Text)
		col := 1
		// blockConsumed is set when this line turned out to be the
		// OPENING of a "###"/"''" comment block (ver below) - the
		// whole multi-line span, including every NEWLINE inside it
		// and the one that would normally end THIS line, is swallowed
		// whole, so the usual trailing-NEWLINE append after the inner
		// loop must be skipped too.
		blockConsumed := false

		for pos := 0; pos < len(runes); {
			if unicode.IsSpace(runes[pos]) {
				pos++
				col++
				continue
			}
			startCol := col
			// pin stamps the real Position onto a token the platform
			// returned - it has no access to ctx.Source/line/col, only
			// the text. When Lex returns several tokens for one call
			// (e.g. a DATA statement's items), they all get pinned to
			// the match's START column, not their own individual
			// offset within it - a minor, accepted precision loss on
			// error-pointer placement for that case, not a
			// functional one.
			pin := func(tok Token) Token {
				tok.Pos = Position{Source: ctx.Source, Line: lineIndex, Column: startCol}
				return tok
			}

			remaining := string(runes[pos:])
			// atLineStart mirrors pass1Dignified's own
			// atOutputLineStart(), one level down: nothing has been
			// appended yet for the CURRENT physical line. Needed so
			// the platform can tell its own single-character
			// conventions (MSX's "'" REM, "#" file channel) apart
			// from Dignified-shared syntax that only means something
			// right at the start of a line (ver contract.go's own doc
			// comment on Lexer.Lex).
			atLineStart := len(tokens) == 0 || tokens[len(tokens)-1].Kind == KindNewline || tokens[len(tokens)-1].Kind == KindProgram
			consumed, newTokens, ok, err := platform.Lexer.Lex(ctx, remaining, atLineStart)
			if err != nil {
				return nil, err
			}
			if ok {
				if consumed <= 0 {
					return nil, fmt.Errorf("line %d: Lexer.Lex claimed a match but consumed nothing at %q", line.Number, remaining)
				}
				for _, t := range newTokens {
					tokens = append(tokens, pin(t))
				}
				pos += consumed
				col += consumed
				continue
			}

			// Dignified/classic comment BLOCKS ("###"..."###" or
			// Description.CommentChar() doubled, e.g. MSX's
			// "''"..."''") - only triggered right at the start of a
			// line (same as badig.py's own `self.last_tok.tok ==
			// 'NEWLINE'` check), but real content CAN share that same
			// opening/closing line with the marker - confirmed used
			// for real in badig/examples/msx/CGK173.dmx (a "###"
			// block whose opener is "###\tInsert ML routines" and
			// whose closer ends "...VRAM=&h1940###", neither alone on
			// its own line). scanCommentBlock's own doc comment has
			// the full mechanics.
			if atLineStart {
				hashes := 0
				for hashes < len(remaining) && remaining[hashes] == '#' {
					hashes++
				}
				classicBlockMarker := platform.Description.CommentChar() + platform.Description.CommentChar()
				isDignifiedBlock := hashes == len(dignifiedBlockCommentMarker)
				isClassicBlock := !isDignifiedBlock && strings.HasPrefix(remaining, classicBlockMarker)
				if isDignifiedBlock || isClassicBlock {
					marker := dignifiedBlockCommentMarker
					if isClassicBlock {
						marker = classicBlockMarker
					}
					content, closeLine, closed := scanCommentBlock(ctx, lineIndex, remaining[len(marker):], marker)
					if !closed {
						return nil, Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("block not closed from: %s", marker), Pos: &Position{Source: ctx.Source, Line: lineIndex, Column: startCol}}
					}
					// isDignifiedBlock ("###") discards the whole
					// span, content included - confirmed reading
					// get_lit_block's own tail
					// (`if tk.tok == 'D_O_EREM': rem_block = []`).
					// isClassicBlock ("''") instead turns each
					// content chunk into its own real classic comment
					// line - mirrors badig.py's own Pass 3 expansion
					// (badig.py:1536-1548) done directly here instead,
					// since the content is never re-tokenized as code
					// either way and no other pass needs to see the
					// block intact (ver pipeline.go's module doc for
					// why this simplification is safe).
					if isClassicBlock {
						com := platform.Description.CommentChar()
						for _, text := range content {
							tokens = append(tokens,
								Token{Kind: KindSynthetic, Value: com, Pos: Position{Source: ctx.Source, Line: lineIndex, Column: startCol}},
								Token{Kind: KindSynthetic, Value: text, Pos: Position{Source: ctx.Source, Line: lineIndex, Column: startCol}},
								Token{Kind: KindNewline, Value: "\n", Pos: Position{Source: ctx.Source, Line: lineIndex, Column: startCol}},
							)
						}
					}
					lineIndex = closeLine
					blockConsumed = true
					break
				}
			}

			// Dignified-exclusive line comment ("##", works ANYWHERE
			// on a physical line, not just at its start - confirmed
			// used both ways in badig/examples/msx/CGK173.dmx; a bare
			// 2-hash run that ISN'T the "###" block marker above also
			// lands here, e.g. "##BB:tag=value" remtags - ver
			// dignifiedLineCommentPrefix's own doc comment).
			if strings.HasPrefix(remaining, dignifiedLineCommentPrefix) {
				break
			}

			// Shared Dignified operators (TRUE/FALSE) are checked before
			// the identifier pattern - same priority as every other
			// dialect/shared keyword winning over identifier
			// interpretation (ver Lex's doc comment) - otherwise a
			// lowercase "true"/"false" would already have matched
			// identifierPattern below and never reach here.
			if word := dignifiedOperatorWord.FindString(remaining); word != "" && dignifiedOperators[strings.ToUpper(word)] {
				tokens = append(tokens, pin(Token{Kind: KindDignifiedOperator, Value: word}))
				pos += len(word)
				col += len(word)
				continue
			}

			// Same priority as the TRUE/FALSE fallback just above.
			if word := dignifiedOperatorWord.FindString(remaining); word != "" && dignifiedInstructions[strings.ToUpper(word)] {
				tokens = append(tokens, pin(Token{Kind: KindDignifiedInstruction, Value: word}))
				pos += len(word)
				col += len(word)
				continue
			}

			if name := functionNamePattern.FindString(remaining); name != "" {
				tokens = append(tokens, pin(Token{Kind: KindFunctionName, Value: name}))
				n := len([]rune(name))
				pos += n
				col += n
				continue
			}

			if tag := toggleRemPattern.FindString(remaining); tag != "" {
				tokens = append(tokens, pin(Token{Kind: KindToggleRem, Value: tag}))
				n := len([]rune(tag))
				pos += n
				col += n
				continue
			}

			if id := identifierPattern.FindString(remaining); id != "" {
				n := len([]rune(id))
				tokens = append(tokens, pin(Token{Kind: KindIdentifier, Value: id}))
				pos += n
				col += n
				continue
			}

			if byte(runes[pos]) == quote && runes[pos] <= 127 {
				n := scanGenericString(runes[pos:], quote)
				tokens = append(tokens, pin(Token{Kind: KindString, Value: string(runes[pos : pos+n])}))
				pos += n
				col += n
				continue
			}

			if runes[pos] <= 127 && dignifiedSymbols[byte(runes[pos])] {
				tokens = append(tokens, pin(Token{Kind: KindDignifiedSymbol, Value: string(runes[pos])}))
				pos++
				col++
				continue
			}

			if runes[pos] == '_' {
				tokens = append(tokens, pin(Token{Kind: KindDignifiedSeparator, Value: "_"}))
				pos++
				col++
				continue
			}

			tokens = append(tokens, pin(Token{Kind: KindUnknown, Value: string(runes[pos])}))
			pos++
			col++
		}

		if !blockConsumed {
			tokens = append(tokens, Token{
				Kind:  KindNewline,
				Value: "\n",
				Pos:   Position{Source: ctx.Source, Line: lineIndex, Column: col},
			})
		}
	}

	tokens = append(tokens, Token{
		Kind:  KindEOF,
		Value: "",
		Pos:   Position{Source: ctx.Source, Line: len(ctx.Source) - 1, Column: 1},
	})
	return tokens, nil
}

// scanGenericString returns how many runes make up a quoted string
// literal starting at runes[0] (the opening quote): consumes to the
// matching closing quote, or to the end of the line if unterminated -
// same tolerant behavior as the original engine's own quoted-line
// scanning.
func scanGenericString(runes []rune, quote byte) int {
	i := 1
	for i < len(runes) {
		if runes[i] <= 127 && byte(runes[i]) == quote {
			return i + 1
		}
		i++
	}
	return i
}

// passFunc is the shape shared by every Parser pass method - letting
// runPass drive any of them identically. It takes the pass's whole
// input slice plus the index being classified this call (not just
// that one token in isolation) - see Parser, in contract.go, for why:
// MSX's real Pass 3 (Bloco 2) needed to look at the previous token.
// On Replace, it emits every token in the returned slice (0, 1, or
// several) - MSX's real Pass 5 (Bloco 2) needed to expand ONE compound
// assignment token (`x+=5`) into several (`=`, `x`, `+`), which a
// single-Token Replace couldn't express.
type passFunc func(ctx *CompileContext, tokens []Token, index int) (HookAction, []Token, error)

// runPass walks tokens once, calling pass for each and applying its
// HookAction - the shared traversal behind every pass_N in the
// original engine's par() (badig.py), once there is no core-level
// Dignified processing (DEFINE/INCLUDE/etc.) of its own to special
// case first. That core-level processing is Fase 3 follow-up work
// (ver OUTLINE.md, Bloco 1, Fase 3); until it exists, every pass's
// traversal is identical, which is why this one helper drives all of
// them. pass always sees the ORIGINAL input tokens for lookback/
// lookahead, even though out, the slice actually being built, may
// already have skipped or replaced earlier entries.
func runPass(ctx *CompileContext, tokens []Token, pass passFunc) ([]Token, error) {
	out := make([]Token, 0, len(tokens))
	for i, tok := range tokens {
		action, result, err := pass(ctx, tokens, i)
		if err != nil {
			return nil, err
		}
		switch action {
		case Skip:
			continue
		case Replace:
			out = append(out, result...)
		default: // Continue
			out = append(out, tok)
		}
	}
	return out, nil
}

// readParenArgs parses a comma-separated, paren-delimited argument
// list starting at *i (which must point at `(`) - mirrors
// get_func_args (badig.py:1149-1193), used both by a FUNC's own
// signature (badig.py's get_func_def) and by a call site's arguments
// (replace_func_calls) - the original shares one function for both,
// so this does too, as a package-level helper rather than a closure
// private to one pass. Nested parens are tracked (depth) so a
// sub-expression's own `(...)` doesn't prematurely end the outer
// list; `(` / `,` / `)` are matched by VALUE, not Kind, same pragmatic
// simplification already used for DEFINE's call-argument parens
// (ver pass1Dignified) - every classic dialect uses these same
// characters regardless of what Kind its own Lexer gives them. An
// empty `()` returns (nil, nil), matching get_func_args' own
// `return []` for that case.
func readParenArgs(tokens []Token, i *int) ([][]Token, error) {
	if *i >= len(tokens) || tokens[*i].Value != "(" {
		pos := Position{}
		if *i < len(tokens) {
			pos = tokens[*i].Pos
		}
		return nil, Diagnostic{Severity: SeverityError, Message: `missing "("`, Pos: &pos}
	}
	open := tokens[*i]
	*i++
	if *i < len(tokens) && tokens[*i].Value == ")" {
		*i++
		return nil, nil
	}

	var args [][]Token
	var cur []Token
	depth := 1
	for {
		if *i >= len(tokens) || tokens[*i].Kind == KindNewline {
			return nil, Diagnostic{Severity: SeverityError, Message: `unbalanced "("`, Pos: &open.Pos}
		}
		tok := tokens[*i]
		switch {
		case tok.Value == "(":
			depth++
			cur = append(cur, tok)
			*i++
		case tok.Value == "," && depth == 1:
			args = append(args, cur)
			cur = nil
			*i++
		case tok.Value == ")":
			depth--
			*i++
			if depth == 0 {
				args = append(args, cur)
				return args, nil
			}
			cur = append(cur, tok)
		default:
			cur = append(cur, tok)
			*i++
		}
	}
}

// isAlphanumeric reports whether every character in s is an ASCII
// letter or digit - used to validate a DECLARE's explicit short-alias
// shape. Mirrors the OBVIOUSLY intended shape of c_varsna
// (badig_msx.py:138, `^[a-zA-z][a-zA-Z0-9]?$`) rather than its actual
// text: `A-z` is a well-known ASCII range typo (it spans A-Z plus
// `[\]^_` before reaching a-z), clearly not meant to accept those
// punctuation characters - corrected here rather than reproduced,
// same spirit as the tokenizer's 2-digit AS/ord() TypeError bug being
// fixed, not replicated (ver OUTLINE.md).
func isAlphanumeric(s string) bool {
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

// tokenTextEqual mirrors compare_func_args (badig.py:1139-1147): joins
// each slice's token values with a space and compares the text -
// used to skip a redundant `param = param` self-assignment when a
// function call's given argument (or captured return variable) is
// already exactly what the parameter/return name itself is.
func tokenTextEqual(a, b []Token) bool {
	text := func(toks []Token) string {
		parts := make([]string, len(toks))
		for i, t := range toks {
			parts[i] = t.String()
		}
		return strings.Join(parts, " ")
	}
	return text(a) == text(b)
}

// functionEntry is one FUNC's registered signature (badig.py's
// func_def namedtuple: args/rets/equl) - populated only once its
// matching RET is processed (badig.py's get_func_ret), never by FUNC
// alone. args/equl pair up positionally (one entry per declared
// parameter: args[n] is just the parameter's own name token, equl[n]
// is its default-value expression, empty if the parameter has none).
// rets is one entry per RET variable (each the function's own
// internal variable that holds that return value).
type functionEntry struct {
	args [][]Token
	equl [][]Token
	rets [][]Token
}

// defineEntry is one DEFINE's stored content (badig.py's def_grp
// namedtuple: repl/varv) - content may contain a KindDefinePlaceholder
// marker (at most one DISTINCT value, though the marker can repeat;
// badig_msx.py's get_defined_content, called with recur=2, allows
// exactly one level of `[...]` nesting inside a define's content, no
// deeper). defaultVar is the placeholder's own inner content (e.g.
// the `0` in `[0]`), used when a call gives no explicit argument;
// nil if the content has no placeholder at all.
type defineEntry struct {
	content    []Token
	defaultVar []Token
}

// pass1Dignified replaces the generic runPass(platform.Parser.Pass1)
// call in Parse with every Pass-1-level Dignified instruction that's
// implemented so far: DEFINE declarations/calls (badig.py:596-745:
// get_defines/get_defined_content/replace_define) and FUNC/RET
// declarations (badig.py:947-1031: get_func_def/get_func_ret) -
// before ever falling through to platform.Parser.Pass1 for every
// other token. All of it is shared Dignified vocabulary
// (badig_dignified.py), not MSX-specific. Named "Dignified", not
// "Defines", because it now covers more than DEFINE (ver OUTLINE.md).
//
// Function CALLS are a separate concern (badig.py's own
// replace_func_calls runs from pass_2, not pass_1 - ver
// pass2Functions below); this only registers a FUNC's signature once
// its matching RET is seen, into the functions map it returns, so
// pass2Functions can resolve calls against it (a call may legally
// appear, in the source, before its function's own declaration -
// pass_1 runs to completion before pass_2 ever starts, so by the time
// any call is resolved every function in the whole file is already
// registered).
//
// Unlike every other pass (even the bespoke pass2Labels/
// pass4Numbering, which still delegate their per-token work to
// runPass), this can't be expressed as "classify the token at index
// i, maybe peeking at a few fixed neighbors" at all: a DEFINE's
// content and a call's argument, and a FUNC's own parameter list, are
// arbitrary-length, recursively bracket-balanced constructs (a call's
// argument can itself be ANOTHER call, nested arbitrarily - confirmed
// in badig/examples/msx/CGK173.dmx, e.g. nested [name](args) chains).
// So this walks the input with its own explicit cursor instead,
// mirroring the original's own recursive-descent helpers
// (next_tok/prev_tok et al, badig.py:566-594) directly - and because
// it builds its output slice by hand, it has TRUE output-lookback
// (unlike the runPass-driven passes' input-based approximation): "is
// DEFINE/FUNC/RET at the start of a line" is answered by checking
// out's own last entry, exactly like the original's last_tok().
func pass1Dignified(ctx *CompileContext, tokens []Token, platform Platform) ([]Token, map[string]functionEntry, error) {
	defines := map[string]defineEntry{}
	for name, bd := range platform.Parser.BuiltinDefines() {
		defines[name] = defineEntry{content: bd.Content, defaultVar: bd.DefaultVar}
	}
	functions := map[string]functionEntry{}
	var out []Token
	i := 0

	// keeps holds every tag a KEEP declaration has registered so far
	// (badig.py's self.des.d_keeps, uppercased "#TAG" values) - purely
	// Pass-1-scoped: toggle_lines (readToggle below) is the only other
	// thing that ever reads it, and it only ever runs from within this
	// same pass, so unlike functions this never needs to escape the
	// function.
	var keeps []string

	// inFunc/pendingFuncName/pendingArgs/pendingEqul carry a FUNC
	// declaration's signature across to its matching RET (badig.py's
	// self.in_func/self.func_def_args/self.func_def_equl) - RET is
	// what actually registers into functions, not FUNC itself.
	var inFunc *Token
	var pendingFuncName string
	var pendingArgs, pendingEqul [][]Token

	isOpenBracket := func(tok Token) bool { return tok.Kind == KindDignifiedSymbol && tok.Value == "[" }
	isCloseBracket := func(tok Token) bool { return tok.Kind == KindDignifiedSymbol && tok.Value == "]" }
	atOutputLineStart := func() bool {
		return len(out) == 0 || out[len(out)-1].Kind == KindNewline || out[len(out)-1].Kind == KindProgram
	}

	// readFuncDef reads a FUNC declaration's name and parenthesized
	// signature at i (mirrors get_func_def, badig.py:947-994) -
	// "FUNC" itself has already been consumed. Each parameter is
	// `name[=default]`; no `=` means no default at all. Appends the
	// KindFunctionDefinition marker to out itself (the original
	// falls through to its own "append self.tk" after get_func_def
	// returns - there's no equivalent "fall through" here, so this
	// does it directly).
	readFuncDef := func() error {
		if inFunc != nil {
			return Diagnostic{Severity: SeverityError, Message: "already inside a function", Pos: &inFunc.Pos}
		}
		if i >= len(tokens) || tokens[i].Kind != KindFunctionName {
			pos := Position{}
			if i < len(tokens) {
				pos = tokens[i].Pos
			}
			return Diagnostic{Severity: SeverityError, Message: "invalid function name", Pos: &pos}
		}
		funcTok := tokens[i]
		i++
		name := strings.ToLower(funcTok.String())
		if _, dup := functions[name]; dup {
			return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("function name duplicated: %s", funcTok.String()), Pos: &funcTok.Pos}
		}

		rawArgs, err := readParenArgs(tokens, &i)
		if err != nil {
			return err
		}
		var defArgs, defEqul [][]Token
		for _, group := range rawArgs {
			if len(group) == 0 {
				return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("function definition missing argument: %s", funcTok.String()), Pos: &funcTok.Pos}
			}
			if group[0].Kind != KindIdentifier {
				return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("function definition only takes variables: %s", group[0].String()), Pos: &group[0].Pos}
			}
			var equl []Token
			if len(group) > 1 {
				if group[1].Value != "=" {
					return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("function definition only takes one variable: %s", group[1].String()), Pos: &group[1].Pos}
				}
				equl = group[2:]
			}
			defArgs = append(defArgs, []Token{group[0]})
			defEqul = append(defEqul, equl)
		}

		inFunc = &funcTok
		pendingFuncName = name
		pendingArgs, pendingEqul = defArgs, defEqul
		out = append(out, Token{Kind: KindFunctionDefinition, Value: name, Pos: funcTok.Pos})
		return nil
	}

	// readFuncRet reads RET's own optional comma-separated return
	// variable list at i (mirrors get_func_ret, badig.py:996-1031) -
	// "RET" itself (retTok) has already been consumed. Registers the
	// function's full signature into functions (only now, not at
	// FUNC time) and appends a synthetic "return" token (never a
	// KindFunctionDefinition/etc. - RET itself needs no resolution,
	// only FUNC's own name does).
	readFuncRet := func(retTok Token) error {
		if inFunc == nil {
			return Diagnostic{Severity: SeverityError, Message: "ret without function", Pos: &retTok.Pos}
		}
		sep := platform.Description.StatementSeparator()
		com := platform.Description.CommentChar()
		isBoundary := func(tok Token) bool {
			return tok.Kind == KindNewline || tok.Value == sep || strings.HasPrefix(tok.String(), com)
		}

		var retList [][]Token
		if i < len(tokens) && !isBoundary(tokens[i]) {
			var cur []Token
			for {
				if i >= len(tokens) {
					return Diagnostic{Severity: SeverityError, Message: "function return missing variable", Pos: &retTok.Pos}
				}
				tok := tokens[i]
				if tok.Value == "," {
					retList = append(retList, cur)
					cur = nil
					i++
					continue
				}
				if isBoundary(tok) {
					retList = append(retList, cur)
					break // leave the boundary token for the caller
				}
				cur = append(cur, tok)
				i++
			}
		}
		for _, r := range retList {
			if len(r) == 0 {
				return Diagnostic{Severity: SeverityError, Message: "function return missing variable", Pos: &retTok.Pos}
			}
		}

		functions[pendingFuncName] = functionEntry{args: pendingArgs, equl: pendingEqul, rets: retList}
		inFunc = nil
		out = append(out, Token{Kind: KindSynthetic, Value: platform.Description.FunctionReturnInstruction(), Pos: retTok.Pos})
		return nil
	}

	// readDeclares reads one whole `DECLARE name[:alias], ...` line
	// (mirrors get_declares, badig.py:747-846) - writes straight into
	// ctx.Symbols, the same 3 tables MSX's own Pass 3/4 (get_hard_
	// variable/process_variable) already consult before allocating an
	// alias, so nothing downstream needs to change to respect a
	// DECLARE. Consumes the terminating NEWLINE, same as DEFINE.
	//
	// The original never actually validates the `:` separator's own
	// value (badig.py:800-804: it only checks the NEXT token is
	// neither `,` nor NEWLINE, then blindly consumes one token
	// assuming it's `:`) - replicated exactly, not tightened.
	//
	// Deliberately not ported: 5 non-fatal (Info.log level 2)
	// prefix-collision warning loops, same as get_hard_variable's own
	// skipped warnings (ver OUTLINE.md).
	readDeclares := func() error {
		aliasLen := platform.Description.VariableAliasLength()
		reserved := platform.Description.ReservedKeywords()
		isReserved := func(name string) bool {
			upper := strings.ToUpper(name)
			for _, w := range reserved {
				if upper == w {
					return true
				}
			}
			return false
		}

		for {
			if i >= len(tokens) || tokens[i].Kind != KindIdentifier {
				pos := Position{}
				if i < len(tokens) {
					pos = tokens[i].Pos
				}
				return Diagnostic{Severity: SeverityError, Message: "invalid declared variable", Pos: &pos}
			}
			longTok := tokens[i]
			longName := strings.ToLower(longTok.String())
			i++

			if len(longName) == 1 {
				return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("declared variable too short: %s", longName), Pos: &longTok.Pos}
			}
			if isReserved(longName) {
				return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("variable is a reserved keyword: %s", longName), Pos: &longTok.Pos}
			}
			key := longName + "@" + longTok.Pos.File()

			if i < len(tokens) && (tokens[i].Value == "," || tokens[i].Kind == KindNewline) {
				// Bare reservation, no explicit alias.
				if len(longName) == aliasLen {
					ctx.Symbols.HardShortVars[longName] = true
				} else {
					ctx.Symbols.HardLongVars[key] = true
				}
				if tokens[i].Kind == KindNewline {
					i++
					return nil
				}
				i++ // consume ","
				continue
			}
			if i >= len(tokens) {
				return Diagnostic{Severity: SeverityError, Message: `expecting ":"`, Pos: &longTok.Pos}
			}
			i++ // consume whatever is assumed to be ":" - unchecked, see doc comment above

			if i >= len(tokens) {
				return Diagnostic{Severity: SeverityError, Message: "invalid declared short variable", Pos: &longTok.Pos}
			}
			shortTok := tokens[i]
			shortName := strings.ToLower(shortTok.String())
			i++

			if len(shortName) < 1 || len(shortName) > aliasLen || !isAlphanumeric(shortName) {
				return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("invalid declared short variable: %s", shortName), Pos: &shortTok.Pos}
			}
			if _, declared := ctx.Symbols.Declares[key]; declared {
				return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("long variable already declared: %s", longName), Pos: &shortTok.Pos}
			}
			for _, alias := range ctx.Symbols.Declares {
				if alias == shortName {
					return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("short variable already declared: %s", shortName), Pos: &shortTok.Pos}
				}
			}
			if ctx.Symbols.HardLongVars[key] {
				return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("long variable already reserved: %s", longName), Pos: &longTok.Pos}
			}
			ctx.Symbols.Declares[key] = shortName

			if i >= len(tokens) {
				return Diagnostic{Severity: SeverityError, Message: `expecting ","`, Pos: &shortTok.Pos}
			}
			if tokens[i].Value == "," {
				i++
				continue
			}
			if tokens[i].Kind == KindNewline {
				i++
				return nil
			}
			return Diagnostic{Severity: SeverityError, Message: `expecting ","`, Pos: &tokens[i].Pos}
		}
	}

	// readKeeps reads one whole `KEEP #tag1 #tag2 ...` line at i
	// (mirrors get_keeps, badig.py:1196-1211) - "KEEP" itself has
	// already been consumed. Consumes the terminating NEWLINE too,
	// same as DEFINE/DECLARE (the original's own dispatch has
	// `continue` for KEEP, same as those - unlike FUNC/RET).
	//
	// Correction, documented (same spirit as isAlphanumeric's own
	// fix above): the original compares `keep.val in self.des.d_keeps`
	// (whatever case the tag was WRITTEN in) against a list that only
	// ever stores `.uval` (always uppercase) - so its "duplicate tag"
	// check almost never actually fires. This compares uppercase on
	// both sides, the clearly intended check.
	readKeeps := func() error {
		for {
			if i >= len(tokens) {
				return Diagnostic{Severity: SeverityError, Message: "unexpected end of file in keep", Pos: &Position{}}
			}
			if tokens[i].Kind == KindNewline {
				i++
				return nil
			}
			if tokens[i].Kind != KindToggleRem {
				return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("invalid keep tag: %s", tokens[i].String()), Pos: &tokens[i].Pos}
			}
			tagUpper := strings.ToUpper(tokens[i].String())
			for _, existing := range keeps {
				if existing == tagUpper {
					return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("keep tag duplicated: %s", tokens[i].String()), Pos: &tokens[i].Pos}
				}
			}
			keeps = append(keeps, tagUpper)
			i++
		}
	}

	// readToggle reads one `#tag` toggle marker at i (i already past
	// tagTok itself) - mirrors toggle_lines (badig.py:1213-1240).
	// tagTok is kept (nothing removed) if its tag, or "#ALL", is in
	// keeps AND "#NONE" is not - with NO keep declared at all, keeps
	// is empty, so this is always false and every tagged block/line
	// is stripped by default (confirmed: these are opt-in extras).
	//
	// Otherwise removes either just the rest of this physical line
	// (single-line form: tagTok is NOT alone on its own line) or an
	// entire block up to a MATCHING closing "#tag" alone on its own
	// line (block form, same shape as badig/examples/msx/CGK173.dmx's
	// "#a"..."#a"). Traced the original's own next_tok()/continue
	// interaction carefully: a single-line toggle's own NEWLINE is
	// consumed too (the whole line vanishes); a block's closing tag is
	// consumed, but the NEWLINE right AFTER it is deliberately left
	// for the caller - that asymmetry is real, not a port mistake.
	readToggle := func(tagTok Token) error {
		tagUpper := strings.ToUpper(tagTok.String())
		keep := false
		for _, k := range keeps {
			if k == tagUpper || k == "#ALL" {
				keep = true
			}
			if k == "#NONE" {
				keep = false
				break // #NONE always overrides, checked last on purpose
			}
		}
		if keep {
			return nil
		}

		blockMode := i < len(tokens) && tokens[i].Kind == KindNewline
		for {
			if i >= len(tokens) || tokens[i].Kind == KindEOF {
				return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("toggle rem not closed from %s", tagTok.String()), Pos: &tagTok.Pos}
			}
			if !blockMode && tokens[i].Kind == KindNewline {
				i++ // the whole single line, including its own newline, vanishes
				return nil
			}
			if blockMode && tokens[i].Kind == KindToggleRem && strings.EqualFold(tokens[i].String(), tagTok.String()) &&
				i > 0 && tokens[i-1].Kind == KindNewline &&
				i+1 < len(tokens) && tokens[i+1].Kind == KindNewline {
				i++ // consume the closing tag only - its trailing newline is left for the caller
				return nil
			}
			i++
		}
	}

	// readBracketName reads `[name]` at i (mirrors get_in_bracket,
	// badig.py:821-838, but accepting any single token as the name,
	// not just an identifier - the original validates alnum shape
	// only loosely, and a define name never collides with platform
	// keywords in practice). Advances i past the closing `]`.
	readBracketName := func() (Token, error) {
		open := tokens[i]
		i++
		if i >= len(tokens) || isCloseBracket(tokens[i]) {
			if i < len(tokens) {
				i++
			}
			return Token{}, Diagnostic{Severity: SeverityError, Message: "define blank", Pos: &open.Pos}
		}
		name := tokens[i]
		i++
		if i >= len(tokens) || !isCloseBracket(tokens[i]) {
			return Token{}, Diagnostic{Severity: SeverityError, Message: `closing "]" not found`, Pos: &name.Pos}
		}
		i++
		return name, nil
	}

	// readNestedBracketOnce reads a `[...]` placeholder's own content
	// at i (i is at the inner `[`) - no further `[` nesting allowed
	// inside it (recur=1 in the original: the placeholder is already
	// one level deep). Advances i past the closing `]`.
	var readNestedBracketOnce func() ([]Token, error)
	readNestedBracketOnce = func() ([]Token, error) {
		open := tokens[i]
		i++
		var inner []Token
		for {
			if i >= len(tokens) || tokens[i].Kind == KindNewline {
				return nil, Diagnostic{Severity: SeverityError, Message: `unbalanced "["`, Pos: &open.Pos}
			}
			if isOpenBracket(tokens[i]) {
				return nil, Diagnostic{Severity: SeverityError, Message: `too many opened "["`, Pos: &tokens[i].Pos}
			}
			if isCloseBracket(tokens[i]) {
				i++
				return inner, nil
			}
			inner = append(inner, tokens[i])
			i++
		}
	}

	// readDefineContent reads a DEFINE's content at i (i is at the
	// content's own opening `[`) - mirrors get_defined_content
	// (badig.py:633-680) called with recur=2. Every nested `[...]`
	// found becomes one KindDefinePlaceholder marker in content; only
	// the FIRST one's own inner tokens are kept as the default value
	// (badig.py itself only ever reads def_rep[1][0][0] - the first -
	// even though it recurses generally), but every marker still gets
	// substituted identically at call time. Advances i past the
	// closing `]`.
	readDefineContent := func() ([]Token, []Token, error) {
		open := tokens[i]
		i++
		var content []Token
		var defaultVar []Token
		haveDefault := false
		for {
			if i >= len(tokens) || tokens[i].Kind == KindNewline {
				return nil, nil, Diagnostic{Severity: SeverityError, Message: `unbalanced "["`, Pos: &open.Pos}
			}
			if isOpenBracket(tokens[i]) {
				pos := tokens[i].Pos
				inner, err := readNestedBracketOnce()
				if err != nil {
					return nil, nil, err
				}
				if !haveDefault {
					haveDefault = true
					defaultVar = inner
				}
				content = append(content, Token{Kind: KindDefinePlaceholder, Pos: pos})
				continue
			}
			if isCloseBracket(tokens[i]) {
				i++
				return content, defaultVar, nil
			}
			content = append(content, tokens[i])
			i++
		}
	}

	// readDefines reads one whole `DEFINE [name][content], ...` line
	// (mirrors get_defines, badig.py:597-631), registering each name
	// (lowercased, matching the original's dict keyed by Token.lval)
	// into defines. Consumes the terminating NEWLINE too - the
	// original's own next_tok()/continue interaction (confirmed
	// tracing it) means a DEFINE line leaves no trace at all, not
	// even a blank line.
	readDefines := func() error {
		for {
			if i >= len(tokens) || !isOpenBracket(tokens[i]) {
				pos := Position{}
				if i < len(tokens) {
					pos = tokens[i].Pos
				}
				return Diagnostic{Severity: SeverityError, Message: `expecting "["`, Pos: &pos}
			}
			nameTok, err := readBracketName()
			if err != nil {
				return err
			}
			content, defaultVar, err := readDefineContent()
			if err != nil {
				return err
			}
			name := strings.ToLower(nameTok.String())
			if _, dup := defines[name]; dup {
				return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("define name duplicated: %s", nameTok.String()), Pos: &nameTok.Pos}
			}
			defines[name] = defineEntry{content: content, defaultVar: defaultVar}

			if i >= len(tokens) {
				return Diagnostic{Severity: SeverityError, Message: `expecting ","`, Pos: &nameTok.Pos}
			}
			if tokens[i].Value == "," {
				i++
				continue
			}
			if tokens[i].Kind == KindNewline {
				i++
				return nil
			}
			return Diagnostic{Severity: SeverityError, Message: `expecting ","`, Pos: &tokens[i].Pos}
		}
	}

	// expandDefineCall reads one `[name]` or `[name](args)` call at i
	// (mirrors replace_define, badig.py:682-745) and returns its fully
	// expanded tokens. Recurses for a nested call appearing inside
	// another call's argument. Content tokens copied verbatim get the
	// CALL's position (so a diagnostic about them points at the call
	// site, matching the original); the substituted argument/default
	// tokens keep their own original position.
	var expandDefineCall func() ([]Token, error)
	expandDefineCall = func() ([]Token, error) {
		nameTok, err := readBracketName()
		if err != nil {
			return nil, err
		}
		name := strings.ToLower(nameTok.String())
		def, ok := defines[name]
		if !ok {
			return nil, Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("%s define not defined", nameTok.String()), Pos: &nameTok.Pos}
		}

		var explicitArg []Token
		haveExplicitArg := false
		if i < len(tokens) && tokens[i].Value == "(" {
			haveExplicitArg = true
			open := tokens[i]
			i++
			depth := 1
			for {
				if i >= len(tokens) || tokens[i].Kind == KindNewline {
					return nil, Diagnostic{Severity: SeverityError, Message: `missing ")"`, Pos: &open.Pos}
				}
				if isOpenBracket(tokens[i]) {
					nested, err := expandDefineCall()
					if err != nil {
						return nil, err
					}
					explicitArg = append(explicitArg, nested...)
					continue
				}
				if tokens[i].Value == "(" {
					depth++
					explicitArg = append(explicitArg, tokens[i])
					i++
					continue
				}
				if tokens[i].Value == ")" {
					depth--
					i++
					if depth == 0 {
						break
					}
					explicitArg = append(explicitArg, tokens[i-1])
					continue
				}
				explicitArg = append(explicitArg, tokens[i])
				i++
			}
		}

		insertVar := def.defaultVar
		if haveExplicitArg {
			insertVar = explicitArg
		}

		var expanded []Token
		for _, tok := range def.content {
			if tok.Kind == KindDefinePlaceholder {
				expanded = append(expanded, insertVar...)
				continue
			}
			tok.Pos = nameTok.Pos
			expanded = append(expanded, tok)
		}
		return expanded, nil
	}

	for i < len(tokens) {
		tok := tokens[i]

		// Checked before KindDignifiedInstruction, same order as the
		// original's own dispatch (badig.py's pass_1: the D_TGLREM
		// check comes before D_INSTRC) - not that it matters for
		// correctness here, since KEEP's own "#tag" arguments are
		// consumed directly by readKeeps, never reaching this check.
		// No error if a bare "#tag" isn't at the start of a line - it
		// just isn't a toggle marker at all, falls through as an
		// ordinary token (same as the original: no dispatch branch
		// matches it, so it reaches self.clc.pass_1() unchanged).
		if tok.Kind == KindToggleRem && atOutputLineStart() {
			i++
			if err := readToggle(tok); err != nil {
				return nil, nil, err
			}
			continue
		}

		if tok.Kind == KindDignifiedInstruction {
			upper := strings.ToUpper(tok.String())
			if (upper == "DEFINE" || upper == "FUNC" || upper == "RET" || upper == "DECLARE" || upper == "KEEP") && !atOutputLineStart() {
				return nil, nil, Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("%s must be at the start of a line", tok.String()), Pos: &tok.Pos}
			}
			switch upper {
			case "DEFINE":
				i++
				if err := readDefines(); err != nil {
					return nil, nil, err
				}
				continue
			case "FUNC":
				i++
				if err := readFuncDef(); err != nil {
					return nil, nil, err
				}
				continue
			case "RET":
				i++
				if err := readFuncRet(tok); err != nil {
					return nil, nil, err
				}
				continue
			case "DECLARE":
				i++
				if err := readDeclares(); err != nil {
					return nil, nil, err
				}
				continue
			case "KEEP":
				i++
				if err := readKeeps(); err != nil {
					return nil, nil, err
				}
				continue
			}
		}

		if isOpenBracket(tok) {
			expanded, err := expandDefineCall()
			if err != nil {
				return nil, nil, err
			}
			out = append(out, expanded...)
			continue
		}

		action, result, err := platform.Parser.Pass1(ctx, tokens, i)
		if err != nil {
			return nil, nil, err
		}
		switch action {
		case Skip:
		case Replace:
			out = append(out, result...)
		default: // Continue
			out = append(out, tok)
		}
		i++
	}

	if inFunc != nil {
		return nil, nil, Diagnostic{Severity: SeverityError, Message: "func without ret", Pos: &inFunc.Pos}
	}

	return out, functions, nil
}

// pass2Functions expands every Dignified function call (badig.py's
// replace_func_calls, 1033-1137) against the functions map
// pass1Dignified registered. In the original this runs as the first
// branch of pass_2's own dispatch (`if self.tk.tok == "D_FUNCNM":`);
// here it's its own pass, running before pass2Labels, because it
// needs the same kind of cursor-based, build-output-by-hand approach
// pass1Dignified does - not for forward-scanning a variable-length
// construct this time, but for a variable-length BACKWARD scan: a
// call that captures return values (`a, b = .name()`) is written
// with the receiving variables and `=` BEFORE the call, already
// emitted to out by the time the call itself is reached, and has to
// be POPPED back off (badig.py's own tok_list_out.pop() loop) to
// recover them - not expressible as "decide this call's own
// HookAction" any more than pass1Dignified's forward constructs were.
//
// Deliberate simplification, documented: this never calls
// platform.Parser.Pass2 itself (pass2Labels, right after it in
// Parse, is left as the one place that still does, unchanged) -
// splitting what was one interleaved original pass into two
// sequential Go ones means a call's expanded tokens get one extra,
// harmless pass through Pass2's hook that the original never gave
// them (today a no-op for MSX either way).
func pass2Functions(ctx *CompileContext, tokens []Token, functions map[string]functionEntry, platform Platform) ([]Token, error) {
	var out []Token
	i := 0

	sep := platform.Description.StatementSeparator()
	eq := platform.Description.AssignmentOperator()
	boundaryWords := platform.Description.FunctionCallBoundaryWords()
	isBoundaryWord := func(tok Token) bool {
		upper := strings.ToUpper(tok.String())
		for _, w := range boundaryWords {
			if upper == w {
				return true
			}
		}
		return false
	}

	for i < len(tokens) {
		tok := tokens[i]

		if tok.Kind != KindFunctionName {
			out = append(out, tok)
			i++
			continue
		}

		funcTok := tok
		name := strings.ToLower(funcTok.String())
		def, ok := functions[name]
		if !ok {
			return nil, Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("function not defined: %s", funcTok.String()), Pos: &funcTok.Pos}
		}
		i++

		// Does this call capture return values ("a, b = .name()")?
		// Scan BACKWARD through out, popping tokens, collecting the
		// comma-separated receiving-variable groups - mirrors
		// replace_func_calls' own tok_list_out.pop() loop exactly.
		var callVars [][]Token
		if len(out) > 0 && out[len(out)-1].Value == eq {
			out = out[:len(out)-1] // pop the "="
			var group []Token
			for {
				if len(out) == 0 {
					return nil, Diagnostic{Severity: SeverityError, Message: "function call missing receiving variable", Pos: &funcTok.Pos}
				}
				last := out[len(out)-1]
				if last.Value == "," {
					out = out[:len(out)-1]
					callVars = append(callVars, reverseTokens(group))
					group = nil
					continue
				}
				if last.Kind == KindNewline || last.Value == sep || isBoundaryWord(last) {
					callVars = append(callVars, reverseTokens(group))
					break // leave the boundary token where it is
				}
				out = out[:len(out)-1]
				group = append(group, last)
			}
			reverseGroups(callVars)
		}
		for _, v := range callVars {
			if len(v) == 0 {
				continue
			}
			if v[0].Kind != KindIdentifier {
				return nil, Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("function definition only takes variables: %s", v[0].String()), Pos: &v[0].Pos}
			}
			if len(v) > 1 {
				return nil, Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("function definition only takes one variable: %s", v[1].String()), Pos: &v[1].Pos}
			}
		}

		callArgs, err := readParenArgs(tokens, &i)
		if err != nil {
			return nil, err
		}
		if len(callArgs) > len(def.args) {
			return nil, Diagnostic{Severity: SeverityError, Message: "function arguments mismatch", Pos: &funcTok.Pos}
		}
		if len(callVars) > len(def.rets) {
			return nil, Diagnostic{Severity: SeverityError, Message: "function variable mismatch", Pos: &funcTok.Pos}
		}
		if len(callArgs) == 0 {
			callArgs = [][]Token{{}}
		}

		var expanded []Token
		for idx := 0; idx < len(def.args) && idx < len(callArgs) && idx < len(def.equl); idx++ {
			d, c, e := def.args[idx], callArgs[idx], def.equl[idx]
			if tokenTextEqual(d, c) {
				continue
			}
			if len(c) > 0 {
				expanded = append(expanded, d[0])
				expanded = append(expanded, Token{Kind: KindSynthetic, Value: eq, Pos: funcTok.Pos})
				expanded = append(expanded, c...)
				expanded = append(expanded, Token{Kind: KindSynthetic, Value: sep, Pos: funcTok.Pos})
			} else if len(e) > 0 {
				expanded = append(expanded, d[0])
				expanded = append(expanded, Token{Kind: KindSynthetic, Value: eq, Pos: funcTok.Pos})
				expanded = append(expanded, e...)
				expanded = append(expanded, Token{Kind: KindSynthetic, Value: sep, Pos: funcTok.Pos})
			}
		}

		expanded = append(expanded,
			Token{Kind: KindSynthetic, Value: platform.Description.FunctionCallInstruction(), Pos: funcTok.Pos},
			Token{Kind: KindFunctionCall, Value: name, Pos: funcTok.Pos},
		)

		for idx := 0; idx < len(callVars) && idx < len(def.rets); idx++ {
			c, d := callVars[idx], def.rets[idx]
			if tokenTextEqual(d, c) || len(c) == 0 {
				continue
			}
			expanded = append(expanded, Token{Kind: KindSynthetic, Value: sep, Pos: funcTok.Pos})
			expanded = append(expanded, c[0])
			expanded = append(expanded, Token{Kind: KindSynthetic, Value: eq, Pos: funcTok.Pos})
			expanded = append(expanded, d...)
		}

		out = append(out, expanded...)
	}

	return out, nil
}

// reverseTokens returns toks reversed (a fresh slice) - used to undo
// the backward-to-forward order a backward scan naturally collects
// tokens in.
func reverseTokens(toks []Token) []Token {
	out := make([]Token, len(toks))
	for i, t := range toks {
		out[len(toks)-1-i] = t
	}
	return out
}

// reverseGroups reverses groups in place - mirrors
// replace_func_calls' `cal_vars = cal_vars[::-1]`: groups are found
// innermost-first by the backward scan, so the overall list needs
// reversing too, not just each group's own tokens.
func reverseGroups(groups [][]Token) {
	for i, j := 0, len(groups)-1; i < j; i, j = i+1, j-1 {
		groups[i], groups[j] = groups[j], groups[i]
	}
}

// pass2Labels wraps platform.Parser.Pass2 with the core-level
// Dignified label handling pass_2 does before ever calling
// self.clc.pass_2() (badig.py:1447-1504, loop-label helpers
// get_label_lines/get_jump_labels/get_loop_label_return/
// get_labels_exit, lines 849-921): turns every `{name}`/`{@}` -
// always 3 adjacent tokens, `{`/name-or-`@`/`}`, since whitespace
// never produces a token - into one KindLabelDeclaration (at the
// start of a line) or KindLabelJump (anywhere else) token, mirroring
// D_LBLLIN/D_LBLJMP; and handles loop labels (`name{ ... }` + `exit`):
// a loop-opener produces the SAME KindLabelDeclaration a plain
// `{name}` does (badig.py's get_label_lines ends in the same
// D_LBLLIN token either way) - only the detection differs, since a
// bare identifier immediately followed by `{` is never wrapped in
// its own braces. A loop's closing `}` (KindLoopReturn) and `exit`
// inside it (KindLoopExit) both emit a synthetic jump-back
// instruction, resolved in core Pass 4 (pass4Numbering below).
//
// Bespoke, not a passFunc, for the same reason pass4Numbering below
// isn't: it needs a local stack of open loop labels across the whole
// pass, plus a post-pass check that the stack is empty at the end
// ("loop label not closed") - badig.py's own pass_2 does the same
// check after its own traversal loop.
func pass2Labels(ctx *CompileContext, tokens []Token, platform Platform) ([]Token, error) {
	isOpenBrace := func(tok Token) bool { return tok.Kind == KindDignifiedSymbol && tok.Value == "{" }
	isCloseBrace := func(tok Token) bool { return tok.Kind == KindDignifiedSymbol && tok.Value == "}" }
	// isPlainLabelAt reports whether the `{` at bracePos starts a
	// plain `{name}`/`{@}` pattern (closed within 2 tokens) - the
	// same condition CASE A below checks, factored out so the
	// loop-opener identifier check (which looks at the `{` one slot
	// AHEAD of itself) can avoid skipping an identifier that a plain
	// label pattern will claim instead (ver badig.py's own dispatch
	// priority: a jump-label match always wins over a loop-opener
	// one, confirmed reading pass_2's D_SYMBOL branch order).
	isPlainLabelAt := func(bracePos int) bool {
		return bracePos+2 < len(tokens) && isCloseBrace(tokens[bracePos+2])
	}
	atLineStart := func(i int) bool {
		return i == 0 || tokens[i-1].Kind == KindNewline || tokens[i-1].Kind == KindProgram
	}

	var loopLabels []Token // stack of open loop labels: Value=name, Pos=opener's position

	pass := func(ctx *CompileContext, tokens []Token, index int) (HookAction, []Token, error) {
		tok := tokens[index]

		// Plain `{name}`/`{@}` (declaration or jump reference).
		if isOpenBrace(tok) && isPlainLabelAt(index) {
			name, ok := tokens[index+1].Value.(string)
			if !ok {
				return Continue, nil, nil
			}
			kind := KindLabelJump
			if atLineStart(index) {
				kind = KindLabelDeclaration
			}
			return Replace, []Token{{Kind: kind, Value: name, Pos: tok.Pos}}, nil
		}
		// The name/`@` and closing `}` of the pattern just handled,
		// reached on their own turn - already consumed.
		if index >= 1 && isOpenBrace(tokens[index-1]) && index+1 < len(tokens) && isCloseBrace(tokens[index+1]) {
			return Skip, nil, nil
		}
		if index >= 2 && isOpenBrace(tokens[index-2]) && isCloseBrace(tok) {
			return Skip, nil, nil
		}

		// Loop-opener name: an identifier at the start of a line,
		// immediately followed by `{` - unless that `{` is itself
		// about to be claimed as a plain label pattern above.
		if tok.Kind == KindIdentifier && atLineStart(index) &&
			index+1 < len(tokens) && isOpenBrace(tokens[index+1]) && !isPlainLabelAt(index+1) {
			return Skip, nil, nil
		}
		// Loop-opener `{`: the identifier just skipped above.
		if isOpenBrace(tok) && index >= 1 && tokens[index-1].Kind == KindIdentifier &&
			atLineStart(index-1) && !isPlainLabelAt(index) {
			name, _ := tokens[index-1].Value.(string)
			loopLabels = append(loopLabels, Token{Value: name, Pos: tokens[index-1].Pos})
			return Replace, []Token{{Kind: KindLabelDeclaration, Value: name, Pos: tokens[index-1].Pos}}, nil
		}

		// Loop-closer: any `}` not already consumed above.
		if isCloseBrace(tok) {
			if len(loopLabels) == 0 {
				return Continue, nil, Diagnostic{
					Severity: SeverityError,
					Message:  "loop label close without open",
					Pos:      &tok.Pos,
				}
			}
			label := loopLabels[len(loopLabels)-1]
			loopLabels = loopLabels[:len(loopLabels)-1]

			var out []Token
			if index == 0 || tokens[index-1].Kind != KindNewline {
				out = append(out, Token{Kind: KindSynthetic, Value: platform.Description.StatementSeparator(), Pos: tok.Pos})
			}
			out = append(out,
				Token{Kind: KindSynthetic, Value: platform.Description.LoopBackInstruction(), Pos: tok.Pos},
				Token{Kind: KindLoopReturn, Value: label.Value, Pos: tok.Pos},
			)
			return Replace, out, nil
		}

		// `exit`: jump to right after the innermost open loop's close.
		if tok.Kind == KindDignifiedInstruction && strings.EqualFold(tok.String(), "EXIT") {
			if len(loopLabels) == 0 {
				return Continue, nil, Diagnostic{
					Severity: SeverityError,
					Message:  "exit outside of any loop label",
					Pos:      &tok.Pos,
				}
			}
			label := loopLabels[len(loopLabels)-1]
			return Replace, []Token{
				{Kind: KindSynthetic, Value: platform.Description.LoopBackInstruction(), Pos: tok.Pos},
				{Kind: KindLoopExit, Value: label.Value, Pos: tok.Pos},
			}, nil
		}

		return platform.Parser.Pass2(ctx, tokens, index)
	}

	out, err := runPass(ctx, tokens, pass)
	if err != nil {
		return nil, err
	}
	if len(loopLabels) > 0 {
		last := loopLabels[len(loopLabels)-1]
		return nil, Diagnostic{
			Severity: SeverityError,
			Message:  fmt.Sprintf("loop label not closed from: %s", last.String()),
			Pos:      &last.Pos,
		}
	}
	return out, nil
}

// pass3Includes replaces the generic runPass(platform.Parser.Pass3)
// call in Parse with the core-level work pass_3 does
// (badig.py:1506-1613) before ever falling through to self.clc.pass_3():
// the line cleanup this absorbs from the former pass3LineHygiene
// (consecutive blank lines collapse to none, a `_` immediately
// followed by NEWLINE merges that line into the next one, the
// NEWLINE right after a label declaration vanishes too), plus
// INCLUDE, which badig.py only ever acts on here - Pass 1 lets the
// INCLUDE/filename tokens through completely unchanged
// (`d_instr == 'INCLUDE': pass`, badig.py:1399-1400), and Pass 2
// doesn't recognize them either.
//
// Doesn't fit passFunc/runPass as a single call, same reason as
// DEFINE/FUNC/DECLARE/KEEP before it: INCLUDE consumes a
// variable-length run of tokens (its own instruction + filename, and
// however many tokens the included file contributes), not just the
// one tokens[index] runPass hands a call.
//
// No "must be at the start of a line" requirement - confirmed neither
// badig.py's Pass 1 nor Pass 3 dispatch for INCLUDE checks that
// (unlike DEFINE/FUNC/RET/DECLARE/KEEP); in practice it only ever
// works at the start of a line anyway, since the next token must
// literally be a quoted filename.
func pass3Includes(ctx *CompileContext, tokens []Token, platform Platform) ([]Token, error) {
	quote := platform.Description.QuoteChar()
	sep := platform.Description.StatementSeparator()
	var out []Token
	i := 0

	// collapsesNewline reports whether a NEWLINE right after a token
	// shaped like (k, v) must vanish - every one of these is a
	// SEPARATE elif branch in badig.py's own pass_3 (badig.py:1556-1588),
	// all sharing the same "continue" shape: KindLabelDeclaration (a
	// label's own line never gets a blank line after it),
	// KindDignifiedSeparator (a `_` at end of line - already consumed
	// without ever reaching out, hence the dual lookback below),
	// KindFunctionDefinition ("Remove newline after function
	// definitions" - confirmed missing before this round: a FUNC's
	// body was wasting a blank line), and the classic statement
	// separator by VALUE, not Kind ("Remove newline after instruction
	// separator", badig.py:1568-1572) - MSX's own ":" has a
	// platform-specific Kind (KindSeparator) this core-level function
	// shouldn't reference, same value-only principle used everywhere
	// else in this pipeline. Unlike KindDignifiedSeparator, the
	// separator itself is NOT removed here - only its NEWLINE is -
	// confirmed empirically running badig.py against a minimal
	// "for f=0 to 5:\n\tprint f:\nnext:\n\tprint \"done\"\n" repro:
	// it becomes ONE line, "for f=0 to 5:print f:next:print "done"",
	// with every original ":" still there.
	// KindProgram is included too: a file whose first physical line is
	// itself blank (or whose content vanishes entirely - KEEP/comment
	// blocks right at the top, etc.) must not waste the file's first
	// line number on nothing. Confirmed empirically: a real
	// "\nprint 1\n" (blank first line) renders as just "10 print 1"
	// in badig.py, not "10" + "20 print 1" - found cross-checking the
	// FULL badig/examples/msx/CGK173.dmx (891 lines, starts with a
	// blank line before its own header comment block) against
	// badig.py's real output: every jump target was off by exactly
	// 10, consistently, which is exactly what this one extra wasted
	// leading line number would cause throughout the whole file.
	collapsesNewline := func(k TokenKind, v string) bool {
		switch k {
		case KindNewline, KindProgram, KindLabelDeclaration, KindDignifiedSeparator, KindFunctionDefinition:
			return true
		}
		return v == sep
	}

	for i < len(tokens) {
		tok := tokens[i]

		if tok.Kind == KindDignifiedInstruction && strings.ToUpper(tok.String()) == "INCLUDE" {
			included, next, err := resolveInclude(ctx, tokens, i, tok, quote, platform)
			if err != nil {
				return nil, err
			}
			out = append(out, included...)
			i = next
			continue
		}

		if tok.Kind == KindNewline && i > 0 {
			// Two lookbacks, needed for different reasons: tokens[i-1]
			// is the ORIGINAL input sequence, which a single skipped
			// token (e.g. a `_` separator right before this NEWLINE)
			// still shows correctly even though out never got it -
			// same rule runPass's own doc comment states for passFunc,
			// which this hand-rolled loop has to honor explicitly now
			// that it isn't runPass-driven anymore. out[-1] is needed
			// on top of that for INCLUDE specifically: when an
			// included file's own content vanishes entirely (e.g. it
			// was just a DECLARE line), tokens[i-1] is the filename
			// string, not a NEWLINE - but out's own last entry is
			// still the PREVIOUS real newline, and this one must
			// collapse against it all the same.
			prevTok := tokens[i-1]
			if len(out) > 0 {
				last := out[len(out)-1]
				if collapsesNewline(last.Kind, last.String()) {
					i++
					continue
				}
			}
			if collapsesNewline(prevTok.Kind, prevTok.String()) {
				i++
				continue
			}
			out = append(out, tok)
			i++
			continue
		}

		if tok.Kind == KindDignifiedSeparator {
			if i+1 < len(tokens) && tokens[i+1].Kind == KindNewline {
				i++
				continue
			}
			// Not at the end of a line: MSX's own Pass 1 "_ as a call
			// separator" territory, not implemented yet (ver
			// OUTLINE.md) - left unchanged.
			out = append(out, tok)
			i++
			continue
		}

		// The classic statement separator (":" for MSX) - by VALUE,
		// same reasoning as collapsesNewline above. Two adjacent
		// separators collapse into one ("Remove excess :",
		// badig.py:1556-1559); a separator right after a NEWLINE
		// (i.e. a line that STARTS with it) instead merges into the
		// PREVIOUS line by popping that boundary NEWLINE ("Remove
		// newline before instruction separator", badig.py:1574-1578) -
		// confirmed empirically: "print 1\n:print 2\nprint 3" becomes
		// "10 print 1:print 2" / "20 print 3".
		if tok.String() == sep {
			if n := len(out); n > 0 && out[n-1].Kind == KindNewline {
				out = out[:n-1]
			} else if n := len(out); n > 0 && out[n-1].String() == sep {
				i++
				continue
			}
			out = append(out, tok)
			i++
			continue
		}

		// Join adjacent string literals (badig.py:1590-1605) - same
		// line, or separated by just a NEWLINE (that NEWLINE vanishes
		// too). Confirmed real and necessary, not cosmetic: a
		// DEFINE's own expansion can produce two adjacent literals
		// with no operator between them (e.g. `define [x]["hello"]`
		// called as `[x]"world"`), which wouldn't even be valid
		// Classic BASIC without merging them into one string first -
		// confirmed running badig.py for real, both forms.
		if tok.Kind == KindString {
			if n := len(out); n > 0 && out[n-1].Kind == KindString {
				out[n-1] = mergeAdjacentStrings(out[n-1], tok, quote)
				i++
				continue
			}
			if n := len(out); n > 1 && out[n-1].Kind == KindNewline && out[n-2].Kind == KindString {
				merged := mergeAdjacentStrings(out[n-2], tok, quote)
				out = out[:n-2]
				out = append(out, merged)
				i++
				continue
			}
		}

		action, result, err := platform.Parser.Pass3(ctx, tokens, i)
		if err != nil {
			return nil, err
		}
		switch action {
		case Skip:
			i++
		case Replace:
			out = append(out, result...)
			i++
		default: // Continue
			out = append(out, tok)
			i++
		}
	}
	return out, nil
}

// mergeAdjacentStrings mirrors "Join adjacent strings" (badig.py:1590-1605):
// `self.last_tok().val.rstrip(qs) + self.tk.val.lstrip(qs)` - strips
// the touching quote off each side and concatenates. The merged
// token's position is b's own, same as the original mutating
// self.tk (the second/current token), not a's.
func mergeAdjacentStrings(a, b Token, quote byte) Token {
	av := strings.TrimRight(a.String(), string(quote))
	bv := strings.TrimLeft(b.String(), string(quote))
	return Token{Kind: KindString, Value: av + bv, Pos: b.Pos}
}

// resolveInclude reads the quoted filename right after an INCLUDE
// token at tokens[i], loads+lexes+parses that file as a nested
// compile, and returns the slice of its own output tokens to splice
// in, plus the index in tokens to resume from. Mirrors include_file
// (badig.py:1243-1286).
func resolveInclude(ctx *CompileContext, tokens []Token, i int, instr Token, quote byte, platform Platform) ([]Token, int, error) {
	i++
	if i >= len(tokens) || tokens[i].Kind != KindString {
		return nil, 0, Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("include error: %s", instr.String()), Pos: &instr.Pos}
	}
	nameTok := tokens[i]
	i++

	includePath := strings.TrimSpace(strings.Trim(nameTok.String(), string(quote)))
	if includePath == "" {
		return nil, 0, Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("include error: %s", nameTok.String()), Pos: &nameTok.Pos}
	}

	// Relative paths resolve against the MAIN file's own directory,
	// not the currently-including file's - confirmed reading
	// support/badig_settings.py:505 (`self.file_path =
	// os.path.dirname(self.file_load)`, set once, shared by reference
	// across every nested Parser instance, never updated per include).
	// IncludeStack[0] is always the main file.
	if !filepath.IsAbs(includePath) && len(ctx.IncludeStack) > 0 {
		includePath = filepath.Join(filepath.Dir(ctx.IncludeStack[0]), includePath)
	}

	if ctx.LoadFile == nil {
		return nil, 0, Diagnostic{Severity: SeverityError, Message: "include error: no file loader configured", Pos: &instr.Pos}
	}
	lines, err := ctx.LoadFile(includePath)
	if err != nil {
		return nil, 0, Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("include error: %v", err), Pos: &instr.Pos}
	}

	nestedCtx := &CompileContext{
		Source:       lines,
		Settings:     ctx.Settings,
		Symbols:      ctx.Symbols,
		Diagnostics:  ctx.Diagnostics,
		IncludeStack: append(append([]string{}, ctx.IncludeStack...), includePath),
		LoadFile:     ctx.LoadFile,
	}

	nestedTokens, err := Lex(nestedCtx, platform)
	if err != nil {
		return nil, 0, err
	}
	nestedTokens, err = Parse(nestedCtx, nestedTokens, platform)
	if err != nil {
		return nil, 0, err
	}

	// Mirrors parsed[2:-2] in include_file: badig.py's own token
	// stream has a synthetic leading NEWLINE right after PROGRAM
	// (lex(), badig.py:418-419) that core.Lex never inserts, so only
	// PROGRAM itself is dropped from the front here; EOF and the
	// included file's own trailing NEWLINE are dropped from the back
	// either way, so the including file's own NEWLINE (which
	// terminated its INCLUDE line) becomes the separator instead,
	// avoiding a doubled blank line.
	contentEnd := len(nestedTokens) - 2
	if contentEnd < 1 {
		contentEnd = 1
	}
	if contentEnd > len(nestedTokens) {
		contentEnd = len(nestedTokens)
	}
	return nestedTokens[1:contentEnd], i, nil
}

// remHeaderLine1/2 are support/badig_settings.py's own header1/header2
// (lines 68-69) - fixed, shared across every dialect (not
// platform-specific), used by pass4Numbering's own KindProgram case
// when CompileSettings.RemHeader is set.
const (
	remHeaderLine1 = "Converted with Basic Dignified"
	remHeaderLine2 = "https://github.com/farique1/basic-dignified"
)

// pass4Numbering replaces the generic runPass(platform.Parser.Pass4)
// call in Parse with the core-level label-resolution + line-numbering
// pass_4 does (badig.py:1615-1757) before ever falling through to
// self.clc.pass_4() (MSX's variable renaming, which keeps running
// exactly as before for every token this function doesn't itself
// recognize).
//
// It doesn't fit passFunc/runPass's one-call-per-token shape as a
// single pass, because the original resolves jump placeholders in a
// POST-pass that mutates a token emitted by an EARLIER call - not
// expressible as "decide this call's own HookAction". Instead of
// tracking an output index the way badig.py does (its own loop
// already builds the final list token by token, so it always knows
// that index), this runs the forward scan via the ordinary runPass
// first (registering each label declaration's line number, resolving
// `{@}` inline, but leaving every other KindLabelJump token
// untouched), then walks the RETURNED slice once more - Go slices
// share the underlying array, so overwriting tokens[i].Value there is
// exactly the original's "mutate tok_list_out[l.index].val" in effect,
// just found by a second linear scan instead of a tracked index.
func pass4Numbering(ctx *CompileContext, tokens []Token, platform Platform) ([]Token, error) {
	labelLines := map[string]int{}
	loopReturnLines := map[string]int{} // D_LBLRET's name+file -> line right after it, for a matching KindLoopExit
	funcDefLines := map[string]int{}    // D_FUNCDF's name+file -> its line - a SEPARATE namespace from labelLines, mirrors func_defs being its own dict in badig.py
	var lineNumber int

	forward := func(ctx *CompileContext, tokens []Token, index int) (HookAction, []Token, error) {
		tok := tokens[index]

		switch tok.Kind {
		case KindProgram:
			// The first real content line has no NEWLINE before it to
			// trigger numbering the way every later line does - number
			// it here instead, right after the sentinel (which never
			// renders in Generate, so this becomes that line's first
			// rendered token).
			lineNumber = ctx.Settings.LineStart
			out := []Token{tok, {Kind: KindLineNumber, Value: lineNumber, Pos: tok.Pos}}

			// Optional 2-line REM header (insert_rem_header,
			// badig.py:1289-1303, called twice right at the start of
			// pass_4, badig.py:1643-1645, BEFORE its own numbering
			// loop runs). Built by hand here instead of relying on
			// the generic KindNewline-triggered numbering below
			// (ver its own case) - a token this function Replaces
			// never gets individually re-visited by this same
			// forward pass (runPass just appends Replace's result
			// verbatim), so each header line's own LineNumber has to
			// be advanced and inserted explicitly, right here, same
			// as the original's own two back-to-back
			// insert_rem_header calls.
			if ctx.Settings.RemHeader {
				com := platform.Description.CommentChar()
				for _, text := range []string{remHeaderLine1, remHeaderLine2} {
					lineNumber += ctx.Settings.LineStep
					out = append(out,
						Token{Kind: KindSynthetic, Value: com, Pos: tok.Pos},
						Token{Kind: KindSynthetic, Value: text, Pos: tok.Pos},
						Token{Kind: KindNewline, Value: "\n", Pos: tok.Pos},
						Token{Kind: KindLineNumber, Value: lineNumber, Pos: tok.Pos},
					)
				}
			}
			return Replace, out, nil

		case KindLabelDeclaration:
			name, _ := tok.Value.(string)
			labelLines[name+tok.Pos.File()] = lineNumber
			return Skip, nil, nil

		case KindLabelJump:
			if name, _ := tok.Value.(string); name == "@" {
				tok.Value = lineNumber
				return Replace, []Token{tok}, nil
			}
			return Continue, nil, nil

		case KindLoopReturn:
			// Resolves against labelLines exactly like KindLabelJump
			// (ver a passada de resolução abaixo) - this additionally
			// records where a matching KindLoopExit (the `exit`
			// command, always encountered BEFORE this closing `}` in
			// token order, since it sits inside the loop body) should
			// land: the line right after this one.
			name, _ := tok.Value.(string)
			loopReturnLines[name+tok.Pos.File()] = lineNumber + ctx.Settings.LineStep
			return Continue, nil, nil

		case KindFunctionDefinition:
			// Resolves/removed exactly like KindLabelDeclaration, but
			// into its own namespace (func_defs is a separate dict
			// from label_lines in badig.py too).
			name, _ := tok.Value.(string)
			funcDefLines[name+tok.Pos.File()] = lineNumber
			return Skip, nil, nil

		case KindFunctionCall:
			// Resolved in the second pass below, same reasoning as
			// KindLabelJump (no `{@}`-style inline case exists for
			// function calls).
			return Continue, nil, nil

		case KindNewline:
			if index+1 < len(tokens) && tokens[index+1].Kind != KindEOF {
				lineNumber += ctx.Settings.LineStep
				return Replace, []Token{tok, {Kind: KindLineNumber, Value: lineNumber, Pos: tok.Pos}}, nil
			}
			return Continue, nil, nil

		default:
			return platform.Parser.Pass4(ctx, tokens, index)
		}
	}

	tokens, err := runPass(ctx, tokens, forward)
	if err != nil {
		return nil, err
	}

	for i := range tokens {
		switch tokens[i].Kind {
		case KindLabelJump, KindLoopReturn:
			// A `{@}` self-reference is already resolved to an int
			// line number inline, during the forward scan above - its
			// Value is no longer a name to look up.
			name, ok := tokens[i].Value.(string)
			if !ok {
				continue
			}
			line, ok := labelLines[name+tokens[i].Pos.File()]
			if !ok {
				return nil, Diagnostic{
					Severity: SeverityError,
					Message:  fmt.Sprintf("label does not exist: %s", name),
					Pos:      &tokens[i].Pos,
				}
			}
			tokens[i].Value = line

		case KindLoopExit:
			// loopReturnLines is only fully populated once the whole
			// forward scan above has finished - an `exit` always
			// appears, in token order, before the loop's closing `}`
			// that records its target, so this can't resolve inline
			// during the forward scan the way `{@}` does.
			name, _ := tokens[i].Value.(string)
			line := loopReturnLines[name+tokens[i].Pos.File()]
			if line > lineNumber-ctx.Settings.LineStep {
				return nil, Diagnostic{
					Severity: SeverityError,
					Message:  "loop exit past end of program",
					Pos:      &tokens[i].Pos,
				}
			}
			tokens[i].Value = line

		case KindFunctionCall:
			name, _ := tokens[i].Value.(string)
			line, ok := funcDefLines[name+tokens[i].Pos.File()]
			if !ok {
				return nil, Diagnostic{
					Severity: SeverityError,
					Message:  fmt.Sprintf("function not defined: %s", name),
					Pos:      &tokens[i].Pos,
				}
			}
			tokens[i].Value = line
		}
	}

	return tokens, nil
}

// Parse runs the parser's five passes over tokens (the output of
// Lex), gating passes 4 and 5 to the main file only - mirrors par()'s
// `pass_1(); pass_2(); pass_3(); if is_main_file(): pass_4(); pass_5()`
// in the original engine. Pass 1/2/3/4 wrap the platform's own hook
// with core-level Dignified work (DEFINE, FUNC/RET, function calls,
// labels, line hygiene, numbering - ver pass1Dignified/pass2Functions/
// pass2Labels/pass3LineHygiene/pass4Numbering) exactly the way the
// original's own pass_1/pass_2/pass_3/pass_4 do before ever calling
// self.clc.passN() (pass2Functions is the one exception that never
// calls it at all - ver its own doc comment for why). INCLUDE
// expansion happens inside pass3Includes, same as the original engine
// (ver its own doc comment).
func Parse(ctx *CompileContext, tokens []Token, platform Platform) ([]Token, error) {
	tokens, functions, err := pass1Dignified(ctx, tokens, platform)
	if err != nil {
		return nil, err
	}
	tokens, err = pass2Functions(ctx, tokens, functions, platform)
	if err != nil {
		return nil, err
	}
	tokens, err = pass2Labels(ctx, tokens, platform)
	if err != nil {
		return nil, err
	}
	tokens, err = pass3Includes(ctx, tokens, platform)
	if err != nil {
		return nil, err
	}
	if !ctx.IsMainFile() {
		return tokens, nil
	}
	tokens, err = pass4Numbering(ctx, tokens, platform)
	if err != nil {
		return nil, err
	}
	tokens, err = runPass(ctx, tokens, platform.Parser.Pass5)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// generalSpaces is the padding Generate's spacing rules insert -
// badig.py's self.stg.general_spaces, which defaults to a single
// space and squeezes to "" under the `-ss` CLI flag
// (support/badig_settings.py:524). No CLI flag wiring exists for that
// yet (ver OUTLINE.md), so this is hardcoded to the original's own
// default rather than added as unused CompileSettings plumbing.
const generalSpaces = " "

// maxLineLength is the classic listing's per-line character limit -
// a literal in the original engine's generate() (badig.py:1865),
// never read from its own Settings, so this mirrors it as a literal
// too rather than inventing new configurability.
const maxLineLength = 256

// Generate runs the parser's Generate hook over tokens (the output of
// Parse) and assembles the final classic listing, line by line -
// mirrors generate() (badig.py:1796-1878).
//
// Spacing is purely value-based, not Kind-based (same as the
// original's own `self.token.uval in self.des.c_reserved_kw`): a
// reserved keyword (platform.Description.ReservedKeywords(), already
// built for DECLARE - ver contract.go) gets a padding space on each
// side, unless the line already ends in one or the neighboring
// token's own value is a "compound symbol"
// (platform.Description.CompoundSymbols(), c_symb_comp in
// badig_msx.py - new for this). A KindLineNumber token always gets a
// trailing space, unconditionally (so "10" + "print" doesn't collide
// into "10print").
//
// Doesn't literally append the line's own trailing NEWLINE to the
// per-line token group the way badig.py's self.line does (its own
// negative-index lookbacks at the line's two ends never actually
// fire in practice, confirmed tracing the short-circuit `and`s - the
// line number is never reserved, and NEWLINE is never reserved
// either - so a bounds-checked `n > 0` / `n+1 < len` here reaches the
// exact same result without needing NEWLINE in the slice at all).
func Generate(ctx *CompileContext, tokens []Token, platform Platform) (GenerateResult, error) {
	tokens, err := runPass(ctx, tokens, platform.Parser.Generate)
	if err != nil {
		return GenerateResult{}, err
	}

	reserved := map[string]bool{}
	for _, w := range platform.Description.ReservedKeywords() {
		reserved[strings.ToUpper(w)] = true
	}
	compound := map[string]bool{}
	for _, s := range platform.Description.CompoundSymbols() {
		compound[s] = true
	}

	var code []string
	var line []Token

	flush := func() error {
		if len(line) == 0 {
			return nil
		}
		var b strings.Builder
		for n, tok := range line {
			text := tok.String()
			isReserved := reserved[strings.ToUpper(text)]

			if n > 0 && isReserved {
				rendered := b.String()
				lastIsSpace := len(rendered) > 0 && rendered[len(rendered)-1] == ' '
				if !lastIsSpace && !compound[line[n-1].String()] {
					b.WriteString(generalSpaces)
				}
			}

			b.WriteString(text)

			if tok.Kind == KindLineNumber || (isReserved && n+1 < len(line) && !compound[line[n+1].String()]) {
				b.WriteString(generalSpaces)
			}
		}

		rendered := strings.TrimRight(b.String(), " ")
		if len(rendered) > maxLineLength {
			pos := line[0].Pos
			return Diagnostic{Severity: SeverityError, Message: fmt.Sprintf("line too long: %d characters", len(rendered)), Pos: &pos}
		}
		code = append(code, rendered)
		line = nil
		return nil
	}

	for _, tok := range tokens {
		switch tok.Kind {
		case KindProgram:
			// sentinela: nunca faz parte do texto gerado
		case KindNewline, KindEOF:
			if err := flush(); err != nil {
				return GenerateResult{}, err
			}
		default:
			line = append(line, tok)
		}
	}
	if err := flush(); err != nil {
		return GenerateResult{}, err
	}

	return GenerateResult{Code: code}, nil
}

// Run drives one compilation from source through to the final classic
// listing: Lex -> Parse -> Generate - the Go equivalent of the
// Dignified path in the original engine's Main.dignified() (badig.py).
// The `-asc` bypass path (straight to Platform.Tools.Run, skipping
// Lex/Parse entirely) is a decision for whoever orchestrates a real
// CLI run, not something this function branches on (ver OUTLINE.md,
// Bloco 1, Fase 3).
func Run(ctx *CompileContext, platform Platform) (GenerateResult, error) {
	tokens, err := Lex(ctx, platform)
	if err != nil {
		return GenerateResult{}, err
	}
	tokens, err = Parse(ctx, tokens, platform)
	if err != nil {
		return GenerateResult{}, err
	}
	return Generate(ctx, tokens, platform)
}

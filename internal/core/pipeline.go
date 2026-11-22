package core

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// identifierPattern is the shared Dignified identifier pattern - the
// SAME for every platform (badig_dignified.py's d_idnttp), which is
// why core.Lex recognizes it itself instead of asking the platform.
var identifierPattern = regexp.MustCompile(`^[a-z][a-z_0-9]*[%!#$]?`)

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
// dignifiedOperators. Only EXIT has a Parser acting on it so far
// (core Pass 2, loop labels); the rest are recognized but unused
// until DEFINE/DECLARE/INCLUDE/KEEP/FUNC/RET exist.
var dignifiedInstructions = map[string]bool{
	"DEFINE": true, "DECLARE": true, "INCLUDE": true, "KEEP": true,
	"ENDIF": true, "FUNC": true, "RET": true, "EXIT": true,
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
			consumed, newTokens, ok, err := platform.Lexer.Lex(ctx, remaining)
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

		tokens = append(tokens, Token{
			Kind:  KindNewline,
			Value: "\n",
			Pos:   Position{Source: ctx.Source, Line: lineIndex, Column: col},
		})
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

// pass3LineHygiene wraps platform.Parser.Pass3 with the core-level
// line cleanup pass_3 does (badig.py:1536-1613) before ever falling
// through to self.clc.pass_3(): consecutive blank lines collapse to
// none, a `_` immediately followed by NEWLINE merges that line into
// the next one (both tokens vanish), and the NEWLINE right after a
// label declaration vanishes too (a label never gets a blank line of
// its own before the code that follows it). All three only ever look
// at tokens[index-1], same lookback already used by MSX's own Pass 3.
func pass3LineHygiene(platform Platform) passFunc {
	return func(ctx *CompileContext, tokens []Token, index int) (HookAction, []Token, error) {
		tok := tokens[index]

		if tok.Kind == KindNewline && index > 0 {
			switch tokens[index-1].Kind {
			case KindNewline, KindLabelDeclaration, KindDignifiedSeparator:
				return Skip, nil, nil
			}
			return Continue, nil, nil
		}

		if tok.Kind == KindDignifiedSeparator {
			if index+1 < len(tokens) && tokens[index+1].Kind == KindNewline {
				return Skip, nil, nil
			}
			// Not at the end of a line: MSX's own Pass 1 "_ as a call
			// separator" territory, not implemented yet (ver
			// OUTLINE.md) - left unchanged.
			return Continue, nil, nil
		}

		return platform.Parser.Pass3(ctx, tokens, index)
	}
}

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
			return Replace, []Token{tok, {Kind: KindLineNumber, Value: lineNumber, Pos: tok.Pos}}, nil

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
		}
	}

	return tokens, nil
}

// Parse runs the parser's five passes over tokens (the output of
// Lex), gating passes 4 and 5 to the main file only - mirrors par()'s
// `pass_1(); pass_2(); pass_3(); if is_main_file(): pass_4(); pass_5()`
// in the original engine. Pass 2/3/4 wrap the platform's own hook with
// core-level Dignified work (labels, line hygiene, numbering - ver
// pass2Labels/pass3LineHygiene/pass4Numbering) exactly the way the
// original's own pass_2/pass_3/pass_4 do before ever calling
// self.clc.passN(). INCLUDE expansion (which happens inside pass 3 in
// the original engine) is not implemented yet (ver OUTLINE.md, Bloco
// 1, Fase 3).
func Parse(ctx *CompileContext, tokens []Token, platform Platform) ([]Token, error) {
	tokens, err := runPass(ctx, tokens, platform.Parser.Pass1)
	if err != nil {
		return nil, err
	}
	tokens, err = pass2Labels(ctx, tokens, platform)
	if err != nil {
		return nil, err
	}
	tokens, err = runPass(ctx, tokens, pass3LineHygiene(platform))
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

// Generate runs the parser's Generate hook over tokens (the output of
// Parse) and assembles the final classic listing.
//
// Unlike the original engine's generate() (badig.py), which rejoins
// tokens with dialect-specific spacing rules sourced from Description
// and enforces a configurable per-line character limit (OUTLINE.md,
// seção 3.5), this assembly is deliberately minimal for now: it
// starts a new output line on each KindNewline token and joins every
// other token's text with a single space. Real spacing/line-limit
// rules are Bloco 2 work, once there is a real Description to drive
// them (ver OUTLINE.md, Bloco 1, Fase 3).
func Generate(ctx *CompileContext, tokens []Token, platform Platform) (GenerateResult, error) {
	tokens, err := runPass(ctx, tokens, platform.Parser.Generate)
	if err != nil {
		return GenerateResult{}, err
	}

	var code []string
	var currentLine []string
	flush := func() {
		if len(currentLine) > 0 {
			code = append(code, strings.Join(currentLine, " "))
			currentLine = nil
		}
	}
	for _, tok := range tokens {
		switch tok.Kind {
		case KindProgram:
			// sentinela: nunca faz parte do texto gerado
		case KindNewline, KindEOF:
			flush()
		default:
			currentLine = append(currentLine, tok.String())
		}
	}
	flush()

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

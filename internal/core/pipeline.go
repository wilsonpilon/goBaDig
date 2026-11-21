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
type passFunc func(ctx *CompileContext, tokens []Token, index int) (HookAction, Token, error)

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
			out = append(out, result)
		default: // Continue
			out = append(out, tok)
		}
	}
	return out, nil
}

// Parse runs the parser's five passes over tokens (the output of
// Lex), gating passes 4 and 5 to the main file only - mirrors par()'s
// `pass_1(); pass_2(); pass_3(); if is_main_file(): pass_4(); pass_5()`
// in the original engine. INCLUDE expansion (which happens inside
// pass 3 in the original engine) is not implemented yet (ver
// OUTLINE.md, Bloco 1, Fase 3).
func Parse(ctx *CompileContext, tokens []Token, platform Platform) ([]Token, error) {
	var err error
	for _, pass := range []passFunc{platform.Parser.Pass1, platform.Parser.Pass2, platform.Parser.Pass3} {
		tokens, err = runPass(ctx, tokens, pass)
		if err != nil {
			return nil, err
		}
	}
	if !ctx.IsMainFile() {
		return tokens, nil
	}
	for _, pass := range []passFunc{platform.Parser.Pass4, platform.Parser.Pass5} {
		tokens, err = runPass(ctx, tokens, pass)
		if err != nil {
			return nil, err
		}
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

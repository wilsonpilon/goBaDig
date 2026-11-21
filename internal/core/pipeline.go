package core

import (
	"strings"
	"unicode"
)

// Lex scans ctx.Source into a flat token stream: a synthetic PROGRAM
// token first, then one token per recognized word plus a NEWLINE per
// source line, ending in EOF.
//
// Unlike the original engine - whose "lexer" is really a single giant
// regex alternation combining every platform's own vocabulary
// (confirmed reading badig.py/badig_dignified.py: core, Dignified and
// Classic groups are all concatenated into one pattern, group order
// mattering) - this scanner only ever recognizes whitespace and line
// boundaries itself. Every other contiguous run of non-whitespace
// characters is staged as a KindWord token and handed to
// platform.Lexer.Lex for classification. A hand-written, table/trie
// based scanner that actually knows a dialect's full vocabulary is
// Bloco 2 work, once there is a real vocabulary to build one against
// (ver OUTLINE.md, Bloco 1, Fase 3).
func Lex(ctx *CompileContext, platform Platform) ([]Token, error) {
	platform.Lexer.Init(ctx)

	tokens := []Token{{
		Kind:  KindProgram,
		Value: "PROGRAM",
		Pos:   Position{Source: ctx.Source, Line: 0, Column: 1},
	}}

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

			start, startCol := pos, col
			for pos < len(runes) && !unicode.IsSpace(runes[pos]) {
				pos++
				col++
			}
			word := string(runes[start:pos])
			candidate := Token{
				Kind:  KindWord,
				Value: word,
				Pos:   Position{Source: ctx.Source, Line: lineIndex, Column: startCol},
			}

			action, tok, err := platform.Lexer.Lex(ctx, candidate)
			if err != nil {
				return nil, err
			}
			switch action {
			case Skip:
				// discard, keep scanning
			case Replace:
				tokens = append(tokens, tok)
			default: // Continue
				tokens = append(tokens, candidate)
			}
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

// passFunc is the shape shared by every Parser pass method - letting
// runPass drive any of them identically.
type passFunc func(ctx *CompileContext, tok Token) (HookAction, Token, error)

// runPass walks tokens once, calling pass for each and applying its
// HookAction - the shared traversal behind every pass_N in the
// original engine's par() (badig.py), once there is no core-level
// Dignified processing (DEFINE/INCLUDE/etc.) of its own to special
// case first. That core-level processing is Fase 3 follow-up work
// (ver OUTLINE.md, Bloco 1, Fase 3); until it exists, every pass's
// traversal is identical, which is why this one helper drives all of
// them.
func runPass(ctx *CompileContext, tokens []Token, pass passFunc) ([]Token, error) {
	out := make([]Token, 0, len(tokens))
	for _, tok := range tokens {
		action, result, err := pass(ctx, tok)
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

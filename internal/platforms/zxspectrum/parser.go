package zxspectrum

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// Parser implements core.Parser for ZX Spectrum - ported from
// badig_zxspectrum.py's Parser(MSXParser), read in full this session.
//
// closeParenAt carries state ACROSS Pass5 calls for a single compile:
// a pending compound-assignment expansion (`x+=5*2` -> `x=x+(5*2)`)
// opens its own "(" at the operator's own index but can only close it
// once the right-hand expression's own terminator (NEWLINE/EOF/the
// statement separator) is reached, several tokens later - unlike the
// original engine (whose own `tok_list_out`/`last_tok()` give every
// hook a direct view of its OWN already-emitted output), this
// project's runPass (internal/core/pipeline.go) always hands a per-
// pass hook the ORIGINAL input slice, never an evolving accumulator -
// so "insert something several tokens later" has no other channel
// than a little Parser-instance state, set by the operator's own call
// and consumed by whichever LATER call reaches that same index. -1
// means "none pending". Use NewParser, not a bare struct literal.
type Parser struct {
	closeParenAt int
}

// NewParser returns a ready-to-use ZX Spectrum Parser.
func NewParser() *Parser {
	return &Parser{closeParenAt: -1}
}

// Pass1 mirrors badig_zxspectrum.py's own pass_1, MINUS REMX (Z_REMX):
// REMX's own argument parsing/line-reservation is deliberately NOT
// implemented this round (ver OUTLINE.md, Bloco 4 - same "postprocess
// bypasses normal numbering" architecture concern already deferred for
// ZX81's own REMX, Bloco 3) - "REMX" is still a recognized instruction
// word (lexer.go), so this gives a clear, specific diagnostic instead
// of letting it fall through and fail confusingly at the tokenizer
// stage (Etapa 1) as an unsupported statement.
func (*Parser) Pass1(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	tok := tokens[index]
	if tok.Kind == KindInstruction && strings.EqualFold(tok.String(), "REMX") {
		return core.Continue, nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  "REMX is not implemented yet in this port of the ZX Spectrum module",
			Pos:      &tok.Pos,
		}
	}
	return core.Continue, nil, nil
}

// Pass2 has no Spectrum-specific work (badig_zxspectrum.py's own
// pass_2 always returns False).
func (*Parser) Pass2(_ *core.CompileContext, _ []core.Token, _ int) (core.HookAction, []core.Token, error) {
	return core.Continue, nil, nil
}

// Pass3 has no Spectrum-specific work (badig_zxspectrum.py's own
// pass_3 always returns False) - unlike MSX/ZX81, there is no
// alias-allocation bookkeeping to register here at all (ver Pass4's
// own doc comment on why this module has no allocator).
func (*Parser) Pass3(_ *core.CompileContext, _ []core.Token, _ int) (core.HookAction, []core.Token, error) {
	return core.Continue, nil, nil
}

// Pass4 mirrors pass_4 (badig_zxspectrum.py:200-219) - the module's
// biggest structural difference from MSX/ZX81: there is NO alias
// allocator at all (Sinclair BASIC accepts native variable names of
// any length), so this only ever VALIDATES a name's shape by category
// (ValidateIdentifier, tokenizer.go) and rejects a collision with a
// ROM keyword - never shortens/renames anything on its own. `~`
// (Dignified's "keep this long name verbatim" marker) has nothing to
// "keep" here either - confirmed reading the original, it's simply
// removed from the stream with no other effect (unlike ZX81, where
// it's always a fatal error, or MSX, where it actually changes
// allocation behavior).
func (*Parser) Pass4(ctx *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	tok := tokens[index]
	if tok.Kind == core.KindDignifiedSymbol && tok.Value == "~" {
		return core.Skip, nil, nil
	}
	if tok.Kind != core.KindIdentifier {
		return core.Continue, nil, nil
	}

	full, _ := tok.Value.(string)
	base, suffix := full, ""
	if strings.HasSuffix(full, "$") {
		base, suffix = full[:len(full)-1], "$"
	}
	key := base + "@" + tok.Pos.File()

	name := full
	changed := false
	if alias, ok := ctx.Symbols.Declares[key]; ok {
		name = alias + suffix
		changed = true
	}

	nextIsParen := index+1 < len(tokens) && tokens[index+1].Value == "("
	forControl := index > 0 &&
		(strings.EqualFold(tokens[index-1].String(), "FOR") || strings.EqualFold(tokens[index-1].String(), "NEXT"))
	if err := ValidateIdentifier(name, nextIsParen, forControl); err != nil {
		return core.Continue, nil, core.Diagnostic{Severity: core.SeverityError, Message: err.Error(), Pos: &tok.Pos}
	}
	if _, reserved := tokenCodes[strings.ToUpper(name)]; reserved {
		return core.Continue, nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  fmt.Sprintf("%s is reserved by Sinclair BASIC", name),
			Pos:      &tok.Pos,
		}
	}
	if changed {
		tok.Value = name
		return core.Replace, []core.Token{tok}, nil
	}
	return core.Continue, nil, nil
}

// atStatementStart reports whether prev is a position a new BASIC
// statement may legally begin right after - mirrors the shared
// condition badig_zxspectrum.py's own pass_5 (implicit LET) and
// Generate (below, which is where THIS port actually inserts LET -
// ver its own doc comment for why) both rely on: immediately after a
// line number, immediately after the statement separator (":" - a
// Spectrum-specific addition vs. ZX81, which only checks line-start/
// THEN), or immediately after "THEN".
func atStatementStart(prev core.Token) bool {
	if prev.Kind == core.KindLineNumber {
		return true
	}
	// Checked by VALUE, not Kind=KindSeparator - core's own shared
	// function-call expansion (pass2Functions) splices in its own
	// ":" between a call's param-assignments/GOSUB/return-assignment
	// as Kind=core.KindSynthetic, never this package's own
	// KindSeparator (confirmed by a real example, functions.DZS,
	// whose synthetic "...;GO SUB 130:answer=result" needs its OWN
	// implicit LET exactly like any other statement boundary).
	if prev.String() == ":" {
		return true
	}
	return strings.EqualFold(prev.String(), "then")
}

// Pass5 mirrors pass_5 (badig_zxspectrum.py:221-270), EXCEPT for where
// the implicit "LET" is actually inserted - see Generate's own doc
// comment for why that one piece moved to a later stage in this port.
func (p *Parser) Pass5(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	tok := tokens[index]

	// Finish a pending compound-assignment expression (ver Parser's
	// own doc comment on closeParenAt) - checked unconditionally,
	// first, since the token that closes it could be anything
	// (NEWLINE, ":", THEN...).
	if index == p.closeParenAt {
		p.closeParenAt = -1
		return core.Replace, []core.Token{{Kind: KindSymbol, Value: ")", Pos: tok.Pos}, tok}, nil
	}

	if tok.Kind == KindExtendedSymbol {
		return p.expandCompoundAssignment(tokens, index)
	}

	if tok.Kind == core.KindDignifiedOperator {
		switch strings.ToUpper(tok.String()) {
		case "TRUE":
			// Confirmed: 1/0, NOT MSX's -1/0 bit-pattern semantics.
			return core.Replace, []core.Token{{Kind: KindNumber, Value: "1", Pos: tok.Pos}}, nil
		case "FALSE":
			return core.Replace, []core.Token{{Kind: KindNumber, Value: "0", Pos: tok.Pos}}, nil
		}
	}
	if tok.Kind == KindInstruction && tok.String() == "?" {
		return core.Replace, []core.Token{{Kind: KindInstruction, Value: "PRINT", Pos: tok.Pos}}, nil
	}

	// GOTO/GOSUB normalization and the generic reserved-keyword
	// uppercase forcing are both checked by VALUE, not by Kind -
	// needed because core's own shared machinery (pass2Functions'
	// `gosub`/`return` for a Dignified function call, badig.py
	// FunctionCallInstruction/FunctionReturnInstruction; pass2Labels'
	// `goto` for a loop label's closing `}`, LoopBackInstruction) emits
	// these as Kind=core.KindSynthetic, never KindJump/KindInstruction -
	// confirmed by a real example (functions.DZS) rendering "gosub"/
	// "return" lowercase, unforced, until this was changed from a
	// Kind-based check to a value-based one.
	switch normalizeKeywordSpacing(tok.String()) {
	case "GOTO", "GO TO":
		return core.Replace, []core.Token{{Kind: tok.Kind, Value: "GO TO", Pos: tok.Pos}}, nil
	case "GOSUB", "GO SUB":
		return core.Replace, []core.Token{{Kind: tok.Kind, Value: "GO SUB", Pos: tok.Pos}}, nil
	}
	if upper := strings.ToUpper(tok.String()); reservedUpper[upper] && upper != tok.String() {
		tok.Value = upper
		return core.Replace, []core.Token{tok}, nil
	}
	return core.Continue, nil, nil
}

// reservedUpper is Description.ReservedKeywords(), uppercased, used by
// Pass5's generic forcing above - a value-based set rather than a
// Kind-based switch precisely so it also catches core's own synthetic
// tokens (ver Pass5's own doc comment), mirroring the original's own
// `tk.uval in self.des.c_reserved_kw` check, which is ALSO by value,
// never by any notion of "token kind".
var reservedUpper = func() map[string]bool {
	m := map[string]bool{}
	for _, w := range (description{}).ReservedKeywords() {
		m[strings.ToUpper(w)] = true
	}
	return m
}()

// expandCompoundAssignment mirrors the C_E_SYMB branch of pass_5
// (badig_zxspectrum.py:227-258): `x+=5*2` -> `x=x+(5*2)`,
// `x++`/`x--` -> `x=x+(1)`/`x=x-(1)`. Scans the right-hand side (every
// token up to the next NEWLINE/EOF/statement-separator) directly off
// the READ-ONLY input slice for validation - it never needs to
// "consume" those tokens away from the main sweep (ver Parser's own
// doc comment): they keep flowing through Pass5 individually and
// unchanged except for whatever their OWN per-token rule already does
// (TRUE/FALSE, uppercase-forcing) - which is exactly what the original
// achieves by manually reapplying those same two rules while
// collecting rhs (badig_zxspectrum.py:248-255), just reached here by
// every token's own independent pass through the SAME hook instead.
func (p *Parser) expandCompoundAssignment(tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	tok := tokens[index]
	op, _ := tok.Value.(string)

	if index < 2 {
		return core.Continue, nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  fmt.Sprintf("%s has no operand to its left", op),
			Pos:      &tok.Pos,
		}
	}
	operand := tokens[index-1]
	if operand.Kind != core.KindIdentifier || !atStatementStart(tokens[index-2]) {
		return core.Continue, nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  "Compound assignment requires a scalar LET target",
			Pos:      &tok.Pos,
		}
	}
	isString := strings.HasSuffix(operand.String(), "$")
	if isString && op != "+=" {
		return core.Continue, nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  "Strings support only += compound assignment",
			Pos:      &tok.Pos,
		}
	}

	end := index + 1
	for end < len(tokens) {
		switch tokens[end].Kind {
		case core.KindNewline, core.KindEOF, KindSeparator:
			goto scanned
		}
		end++
	}
scanned:
	rhs := tokens[index+1 : end]
	for _, t := range rhs {
		if t.Kind == KindExtendedSymbol {
			return core.Continue, nil, core.Diagnostic{
				Severity: core.SeverityError,
				Message:  "Nested compound assignments are not supported",
				Pos:      &tok.Pos,
			}
		}
	}

	if op == "++" || op == "--" {
		if len(rhs) > 0 || isString {
			return core.Continue, nil, core.Diagnostic{
				Severity: core.SeverityError,
				Message:  "Increment/decrement requires a numeric scalar and no right operand",
				Pos:      &tok.Pos,
			}
		}
		return core.Replace, []core.Token{
			{Kind: KindSymbol, Value: "=", Pos: tok.Pos},
			operand,
			{Kind: KindSymbol, Value: string(op[0]), Pos: tok.Pos},
			{Kind: KindSymbol, Value: "(", Pos: tok.Pos},
			{Kind: KindNumber, Value: "1", Pos: tok.Pos},
			{Kind: KindSymbol, Value: ")", Pos: tok.Pos},
		}, nil
	}

	if len(rhs) == 0 {
		return core.Continue, nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  "Compound assignment requires a right operand",
			Pos:      &tok.Pos,
		}
	}

	p.closeParenAt = end
	return core.Replace, []core.Token{
		{Kind: KindSymbol, Value: "=", Pos: tok.Pos},
		operand,
		{Kind: KindSymbol, Value: string(op[0]), Pos: tok.Pos},
		{Kind: KindSymbol, Value: "(", Pos: tok.Pos},
	}, nil
}

// Generate mirrors generate() (badig_zxspectrum.py:272-285) - EXCEPT
// the spacing rule itself, which core.Generate's own shared pass
// (pipeline.go) already implements identically (confirmed: Spectrum's
// own rule is "line number always followed by a space, reserved word
// adjacent to another token gets a space unless that neighbor is a
// compound symbol" - exactly core's generic rule, no platform-specific
// override needed for it at all).
//
// The implicit "LET" insertion is moved HERE from pass_5 (where the
// original puts it) rather than ported in place, same spirit as ZX81's
// own Generate-based LET (zx81/parser.go) - checked and found safe:
// Pass5's own compound-assignment expansion (above) validates its
// operand is at a statement-start position by checking tokens[index-2]
// directly, never by inspecting an actual "LET" token's presence - so
// deferring the INSERTION itself to Generate changes nothing it
// depends on. Doing it here instead avoids the one thing that WOULD
// have been awkward in Pass5: a hook driven by runPass only ever sees
// the pass's ORIGINAL input tokens (ver Parser's own doc comment), so
// an identifier's "was LET just inserted before me" can only be
// answered by asking IF it would be - which is exactly the same
// question Generate asks, just once, where it's needed.
func (*Parser) Generate(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	tok := tokens[index]
	if tok.Kind == core.KindLineNumber {
		if n, err := strconv.Atoi(tok.String()); err == nil && (n < 1 || n > 9999) {
			return core.Continue, nil, core.Diagnostic{
				Severity: core.SeverityError,
				Message:  "Spectrum line number exceeds 9999; reduce -ls or -lp",
				Pos:      &tok.Pos,
			}
		}
		return core.Continue, nil, nil
	}
	if tok.Kind != core.KindIdentifier || index == 0 || !atStatementStart(tokens[index-1]) {
		return core.Continue, nil, nil
	}
	return core.Replace, []core.Token{
		{Kind: KindInstruction, Value: "LET", Pos: tok.Pos},
		tok,
	}, nil
}

// BuiltinDefines mirrors badig_zxspectrum.py's Parser.initialization:
// "[?](row,col)" is pre-registered as shorthand for "print at row,col;"
// in every compile, unconditionally. Unlike ZX81's own "[?](row,col)"
// (zx81/parser.go), this uses the SIMPLE MSX-style mechanism (a single
// Content placeholder, no Expand) - confirmed empirically (running
// badig.py): calling it with only ONE argument, e.g. "[?](5)", splices
// that one argument in LITERALLY, producing "PRINT AT 5;" (no comma,
// no second coordinate defaulted independently) - a real behavior of
// the original to replicate faithfully, not a gap to fix.
func (*Parser) BuiltinDefines() map[string]core.BuiltinDefine {
	return map[string]core.BuiltinDefine{
		"?": {
			Content: []core.Token{
				{Kind: KindInstruction, Value: "print"},
				{Kind: KindInstruction, Value: "at"},
				{Kind: core.KindDefinePlaceholder},
				{Kind: KindSymbol, Value: ";"},
			},
			DefaultVar: []core.Token{
				{Kind: KindNumber, Value: "0"},
				{Kind: KindSymbol, Value: ","},
				{Kind: KindNumber, Value: "0"},
			},
		},
	}
}

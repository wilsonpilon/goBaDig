package msx

import (
	"fmt"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// Parser implements core.Parser for MSX. Pass 3/4 (variable
// allocation, confirmed line by line against badig_msx.py's
// get_hard_variable/process_variable) and Pass 5 (the final
// Classic-specific conversions, badig_msx.py:617-682) do real work.
// Pass 1 (the Dignified call-instruction separator) and a real
// generate() (reserved-keyword spacing, the 256-char line limit) are a
// later round - see OUTLINE.md, Bloco 2.
//
// Deliberately not ported from get_hard_variable: the non-fatal
// (Info.log level 2) prefix-collision WARNING loops. The two FATAL
// checks (using `~` on an already-short name; using `~` on an
// already-declared name) are ported, since skipping those silently
// would let genuinely invalid programs "succeed" incorrectly.
//
// Holds its own core.VariableAllocator, NOT a package-level global -
// allocator state (which short aliases are already handed out) is
// per-compile, so two compilations in the same process (e.g. two
// test cases) never see each other's allocations. Use NewParser,
// not a bare Parser{}, to get one correctly initialized.
//
// ConvertPrint/StripThenGoto/Translate mirror Settings fields the
// original engine exposes via `-cp`/`-tg` CLI flags and a translate
// flag (badig_msx.py:248-313) - none of that CLI/Settings wiring
// exists yet (OUTLINE.md's pending "Settings/Expose MSX" item), so
// these default to the same off/no-op values the original has before
// any flag is given, and can only be set programmatically for now.
type Parser struct {
	allocator core.VariableAllocator

	// ConvertPrint: "" = no conversion (default); "?" = convert PRINT
	// text to "?"; any other non-empty value = convert "?" to "print".
	ConvertPrint string
	// StripThenGoto: "" = no stripping (default); "T"/"t" = strip a
	// THEN immediately followed by GOTO; "G"/"g" = strip a GOTO
	// immediately preceded by THEN or ELSE.
	StripThenGoto string
	// Translate: Unicode-to-ASCII character translation for literal
	// tokens (strings, comments). Off by default.
	Translate bool
}

// NewParser returns an MSX Parser ready for one compile.
func NewParser() *Parser {
	return &Parser{allocator: core.NewDescendingAllocator()}
}

func (*Parser) Pass1(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	return core.Continue, nil, nil
}

func (*Parser) Pass2(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	return core.Continue, nil, nil
}

// Pass3 mirrors get_hard_variable (badig_msx.py:463-514): registers
// short names used directly in the source (so Pass 4's allocator
// never hands that same short name to a different long variable),
// and short-circuits the Dignified "keep this long name verbatim"
// marker, `~`, immediately before an identifier.
func (*Parser) Pass3(ctx *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	tok := tokens[index]

	if tok.Kind == core.KindDignifiedSymbol && tok.Value == "~" {
		// Consumed here; the following identifier (if any) looks back
		// at this same input slice to see it was preceded by "~" -
		// mirrors the original's `self.par.tok_list_out.pop()` without
		// needing to mutate anything already emitted (ver OUTLINE.md,
		// Bloco 2, "achado que exige mudar a interface core.Parser").
		return core.Skip, nil, nil
	}
	if tok.Kind != core.KindIdentifier {
		return core.Continue, nil, nil
	}

	name, _ := tok.Value.(string)
	key := name + "@" + tok.Pos.File()

	if index > 0 && tokens[index-1].Kind == core.KindDignifiedSymbol && tokens[index-1].Value == "~" {
		if len(name) <= variableAliasLen {
			return core.Continue, nil, core.Diagnostic{
				Severity: core.SeverityError,
				Message:  fmt.Sprintf("can't use ~ on a short named variable: %s", name),
				Pos:      &tok.Pos,
			}
		}
		if _, declared := ctx.Symbols.Declares[key]; declared {
			return core.Continue, nil, core.Diagnostic{
				Severity: core.SeverityError,
				Message:  fmt.Sprintf("long variable already declared: %s", name),
				Pos:      &tok.Pos,
			}
		}
		ctx.Symbols.HardLongVars[key] = true
		return core.Continue, nil, nil
	}

	if _, declared := ctx.Symbols.Declares[key]; declared {
		return core.Continue, nil, nil
	}
	if len(name) <= variableAliasLen {
		ctx.Symbols.HardShortVars[name] = true
	}
	return core.Continue, nil, nil
}

// Pass4 mirrors process_variable (badig_msx.py:516-558): renames
// every identifier not already resolved by Pass 3 into its 2-letter
// alias, allocating a new one (descending from "zz") the first time a
// given long name is seen.
func (p *Parser) Pass4(ctx *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	tok := tokens[index]
	if tok.Kind != core.KindIdentifier {
		return core.Continue, nil, nil
	}

	name, _ := tok.Value.(string)
	suffix := ""
	if n := len(name); n > 0 {
		if c := name[n-1]; c == '%' || c == '!' || c == '#' || c == '$' {
			suffix = string(c)
			name = name[:n-1]
		}
	}
	key := name + "@" + tok.Pos.File()

	if len(name) <= variableAliasLen {
		return core.Continue, nil, nil
	}
	if ctx.Symbols.HardLongVars[key] {
		return core.Continue, nil, nil
	}
	if alias, ok := ctx.Symbols.Declares[key]; ok {
		tok.Value = alias + suffix
		return core.Replace, []core.Token{tok}, nil
	}

	excluded := func(alias string) bool {
		if ctx.Symbols.HardShortVars[alias] {
			return true
		}
		for _, declaredAlias := range ctx.Symbols.Declares {
			if declaredAlias == alias {
				return true
			}
		}
		for longKey := range ctx.Symbols.HardLongVars {
			if len(longKey) >= variableAliasLen && longKey[:variableAliasLen] == alias {
				return true
			}
		}
		return false
	}

	alias, err := p.allocator.Allocate(variableCategory, key, excluded)
	if err != nil {
		return core.Continue, nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  fmt.Sprintf("too many variables used (max=%d): %s", aliasSpaceSize(), name),
			Pos:      &tok.Pos,
		}
	}
	ctx.Symbols.Declares[key] = alias
	tok.Value = alias + suffix
	return core.Replace, []core.Token{tok}, nil
}

// Pass5 mirrors pass_5 (badig_msx.py:617-682): the final
// Classic-specific conversions, applied once per compile (gated to
// the main file only, same as Pass 4 - see core.Parse).
func (p *Parser) Pass5(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	tok := tokens[index]

	// Convert ? to print or vice versa.
	if tok.Kind == KindInstruction && p.ConvertPrint != "" {
		if val, _ := tok.Value.(string); val == "?" || strings.EqualFold(val, "PRINT") {
			if p.ConvertPrint == "?" {
				tok.Value = "?"
			} else {
				tok.Value = "print"
			}
			return core.Replace, []core.Token{tok}, nil
		}
	}

	// Strip THEN before GOTO / GOTO after THEN or ELSE.
	if tok.Kind == KindJump {
		upper := strings.ToUpper(tok.String())
		if strings.ToUpper(p.StripThenGoto) == "T" && upper == "THEN" &&
			index+1 < len(tokens) && strings.EqualFold(tokens[index+1].String(), "GOTO") {
			return core.Skip, nil, nil
		}
		if strings.ToUpper(p.StripThenGoto) == "G" && upper == "GOTO" && index > 0 &&
			(strings.EqualFold(tokens[index-1].String(), "THEN") || strings.EqualFold(tokens[index-1].String(), "ELSE")) {
			return core.Skip, nil, nil
		}
	}

	// Expand compound assignment/increment/decrement
	// (x+=5 -> = x + 5; x++ -> = x + 1), reusing the already-emitted
	// operand (tokens[index-1], which reflects whatever Pass 4 already
	// renamed it to) - see "Verificação feita nesta sessão" in the
	// approved plan for why this produces `x = x + 5`, not `= x + 5`.
	if tok.Kind == KindSymbol {
		if op, _ := tok.Value.(string); contains(extendedSymbols, op) {
			if index == 0 {
				return core.Continue, nil, core.Diagnostic{
					Severity: core.SeverityError,
					Message:  fmt.Sprintf("%s has no operand to its left", op),
					Pos:      &tok.Pos,
				}
			}
			operand := tokens[index-1]
			out := []core.Token{
				{Kind: KindSymbol, Value: "=", Pos: tok.Pos},
				operand,
				{Kind: KindSymbol, Value: string(op[0]), Pos: tok.Pos},
			}
			if len(op) == 2 && op[0] == op[1] {
				out = append(out, core.Token{Kind: KindNumber, Value: "1", Pos: tok.Pos})
			}
			return core.Replace, out, nil
		}
	}

	// Replace TRUE/FALSE.
	if tok.Kind == core.KindDignifiedOperator {
		switch strings.ToUpper(tok.String()) {
		case "FALSE":
			return core.Replace, []core.Token{{Kind: KindNumber, Value: "0", Pos: tok.Pos}}, nil
		case "TRUE":
			return core.Replace, []core.Token{
				{Kind: KindSymbol, Value: "-", Pos: tok.Pos},
				{Kind: KindNumber, Value: "1", Pos: tok.Pos},
			}, nil
		}
	}

	// Translate Unicode characters to their ASCII equivalent - only
	// for literal tokens (strings, comments); DATA items aren't parsed
	// as their own literals yet (see lexer.go), so KindData isn't
	// included here.
	if p.Translate && (tok.Kind == core.KindString || tok.Kind == KindComment) {
		if s, ok := tok.Value.(string); ok {
			tok.Value = translateChars(s)
			return core.Replace, []core.Token{tok}, nil
		}
	}

	return core.Continue, nil, nil
}

func (*Parser) Generate(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	return core.Continue, nil, nil
}

// aliasSpaceSize is how many distinct 2-letter MSX aliases exist
// (c_var_max, badig_msx.py:149 - 26*26=676).
func aliasSpaceSize() int {
	return len(variableAliasChars) * len(variableAliasChars)
}

// translateChars mirrors trans_char (badig_msx.py:560-567): first a
// dict-based replacement of specific box-drawing/smiley characters
// (cReplacements), then a second pass mapping every remaining
// character found in cOriginal to its paired code point (0x80 + its
// index in cOriginal) - MSX's internal 8-bit character codes, stored
// as Unicode code points \u0080-ÿ so they round-trip through Go
// strings the same way the original stores them in Python strings.
func translateChars(s string) string {
	var b strings.Builder
	for _, r := range s {
		if replacement, ok := cReplacements[r]; ok {
			r = replacement
		}
		for i, o := range cOriginalRunes {
			if o == r {
				r = rune(0x80 + i)
				break
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

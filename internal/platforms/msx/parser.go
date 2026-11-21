package msx

import (
	"fmt"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// Parser implements core.Parser for MSX. Only Pass 3 and Pass 4 do
// real work in this round - the variable-allocation half of the
// Parser, confirmed line by line against badig_msx.py's
// get_hard_variable (Pass 3) and process_variable (Pass 4). Pass 1
// (the Dignified call-instruction separator), Pass 5 (`?`/PRINT
// conversion, THEN/GOTO stripping, compound-assignment expansion,
// TRUE/FALSE, Unicode translation) and a real generate() (spacing
// rules, the 256-char line limit) are a later round - see
// OUTLINE.md, Bloco 2.
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
type Parser struct {
	allocator core.VariableAllocator
}

// NewParser returns an MSX Parser ready for one compile.
func NewParser() *Parser {
	return &Parser{allocator: core.NewDescendingAllocator()}
}

func (*Parser) Pass1(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, core.Token, error) {
	return core.Continue, tokens[index], nil
}

func (*Parser) Pass2(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, core.Token, error) {
	return core.Continue, tokens[index], nil
}

// Pass3 mirrors get_hard_variable (badig_msx.py:463-514): registers
// short names used directly in the source (so Pass 4's allocator
// never hands that same short name to a different long variable),
// and short-circuits the Dignified "keep this long name verbatim"
// marker, `~`, immediately before an identifier.
func (*Parser) Pass3(ctx *core.CompileContext, tokens []core.Token, index int) (core.HookAction, core.Token, error) {
	tok := tokens[index]

	if tok.Kind == core.KindDignifiedSymbol && tok.Value == "~" {
		// Consumed here; the following identifier (if any) looks back
		// at this same input slice to see it was preceded by "~" -
		// mirrors the original's `self.par.tok_list_out.pop()` without
		// needing to mutate anything already emitted (ver OUTLINE.md,
		// Bloco 2, "achado que exige mudar a interface core.Parser").
		return core.Skip, core.Token{}, nil
	}
	if tok.Kind != core.KindIdentifier {
		return core.Continue, tok, nil
	}

	name, _ := tok.Value.(string)
	key := name + "@" + tok.Pos.File()

	if index > 0 && tokens[index-1].Kind == core.KindDignifiedSymbol && tokens[index-1].Value == "~" {
		if len(name) <= variableAliasLen {
			return core.Continue, tok, core.Diagnostic{
				Severity: core.SeverityError,
				Message:  fmt.Sprintf("can't use ~ on a short named variable: %s", name),
				Pos:      &tok.Pos,
			}
		}
		if _, declared := ctx.Symbols.Declares[key]; declared {
			return core.Continue, tok, core.Diagnostic{
				Severity: core.SeverityError,
				Message:  fmt.Sprintf("long variable already declared: %s", name),
				Pos:      &tok.Pos,
			}
		}
		ctx.Symbols.HardLongVars[key] = true
		return core.Continue, tok, nil
	}

	if _, declared := ctx.Symbols.Declares[key]; declared {
		return core.Continue, tok, nil
	}
	if len(name) <= variableAliasLen {
		ctx.Symbols.HardShortVars[name] = true
	}
	return core.Continue, tok, nil
}

// Pass4 mirrors process_variable (badig_msx.py:516-558): renames
// every identifier not already resolved by Pass 3 into its 2-letter
// alias, allocating a new one (descending from "zz") the first time a
// given long name is seen.
func (p *Parser) Pass4(ctx *core.CompileContext, tokens []core.Token, index int) (core.HookAction, core.Token, error) {
	tok := tokens[index]
	if tok.Kind != core.KindIdentifier {
		return core.Continue, tok, nil
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
		return core.Continue, tok, nil
	}
	if ctx.Symbols.HardLongVars[key] {
		return core.Continue, tok, nil
	}
	if alias, ok := ctx.Symbols.Declares[key]; ok {
		tok.Value = alias + suffix
		return core.Replace, tok, nil
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
		return core.Continue, tok, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  fmt.Sprintf("too many variables used (max=%d): %s", aliasSpaceSize(), name),
			Pos:      &tok.Pos,
		}
	}
	ctx.Symbols.Declares[key] = alias
	tok.Value = alias + suffix
	return core.Replace, tok, nil
}

func (*Parser) Pass5(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, core.Token, error) {
	return core.Continue, tokens[index], nil
}

func (*Parser) Generate(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, core.Token, error) {
	return core.Continue, tokens[index], nil
}

// aliasSpaceSize is how many distinct 2-letter MSX aliases exist
// (c_var_max, badig_msx.py:149 - 26*26=676).
func aliasSpaceSize() int {
	return len(variableAliasChars) * len(variableAliasChars)
}

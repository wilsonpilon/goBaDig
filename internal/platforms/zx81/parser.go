package zx81

import (
	"fmt"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// Parser implements core.Parser for ZX81 - ported from
// badig_zx81.py's Parser(MSXParser), read in full this session.
//
// allocator state (which short aliases are already handed out) is
// held per-instance, same reasoning as MSX's own Parser (ver
// msx/parser.go): different compiles (and future test cases) never
// see each other's allocations. Use NewParser, not a bare struct
// literal.
type Parser struct {
	allocator core.VariableAllocator
}

// NewParser returns a ready-to-use ZX81 Parser.
func NewParser() *Parser {
	return &Parser{allocator: core.NewDescendingAllocator()}
}

// Pass1 has no ZX81-specific work (badig_zx81.py's own Parser class
// never overrides pass_1 - every Dignified-level mechanism it needs,
// DEFINE/FUNC/RET/DECLARE/KEEP/labels, is already shared, platform-
// agnostic core machinery). Deliberately NOT mirroring MSX's own
// Pass1 override (its KindCallSeparator/KindCallIdentifier handling
// disambiguates an MSX-specific edge case around the Dignified "~"
// marker and the "_" separator that hasn't been confirmed to even
// apply to ZX81 yet - left for a future round if a real ZX81 program
// ever needs it, same "validate against something real first"
// discipline used throughout this project).
func (*Parser) Pass1(_ *core.CompileContext, _ []core.Token, _ int) (core.HookAction, []core.Token, error) {
	return core.Continue, nil, nil
}

// Pass2 has no per-token ZX81-specific work. PRINT AT - the one thing
// badig_zx81.py's own Parser.pass_2 override exists for - is handled
// entirely during BuiltinDefines' Expand (ver this file's own
// BuiltinDefines), which runs during Pass 1's already-bespoke,
// cursor-driven define-expansion instead: that code already computes
// a call's resolved argument tokens as a byproduct of substituting
// them in, so PRINT AT's own comma-splitting/defaulting can happen
// right there, without needing Pass2 to also become bespoke (ver
// OUTLINE.md, Bloco 3, for the full reasoning this decision is based
// on - the same kind of minimal, precedented interface growth already
// used for BuiltinDefines/Lexer's atLineStart/HookAction's []Token).
func (*Parser) Pass2(_ *core.CompileContext, _ []core.Token, _ int) (core.HookAction, []core.Token, error) {
	return core.Continue, nil, nil
}

// Pass3 mirrors get_hard_variable (inherited unchanged by
// badig_zx81.py's Parser from MSXParser - badig_msx.py:463-514):
// registers short names used directly in the source (so Pass 4's
// allocator never hands that same short name to a different long
// variable), and short-circuits the Dignified "keep this long name
// verbatim" marker, "~", immediately before an identifier. Unlike
// MSX, "~" can never actually be HONORED for ZX81 (ver Pass4's own
// doc comment) - but the MARKING itself (which Pass4 then reacts to)
// is identical, shared Dignified machinery, so this pass is ported
// unchanged from msx/parser.go's own Pass3.
func (*Parser) Pass3(ctx *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	tok := tokens[index]

	if tok.Kind == core.KindDignifiedSymbol && tok.Value == "~" {
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

// Pass4 mirrors process_variable (badig_zx81.py:330-356): renames
// every identifier not already 1 letter into its single-letter
// alias, allocating a new one (descending from "z") the first time a
// given long name is seen.
//
// Deliberately DIFFERENT from MSX's own Pass4 in one key way: MSX
// treats an identifier marked HardLongVars (Pass3's own "~" handling)
// as "skip, keep the long name as-is" - ZX81 classic BASIC has no way
// to represent a variable name longer than 1 letter (+ optional "$")
// AT ALL, so "~" can never actually be honored here; badig_zx81.py's
// own process_variable instead raises a fatal error the moment it
// sees a HardLongVars entry for this name (checked by NAME+FILE, same
// key format as MSX's own `key` - fires for EITHER the very
// occurrence marked "~" or a different, unmarked occurrence of the
// same name, confirmed reading process_variable's own loop, which
// compares by name+file unconditionally before ever checking whether
// THIS token itself carries the marker).
func (p *Parser) Pass4(ctx *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	tok := tokens[index]
	if tok.Kind != core.KindIdentifier {
		return core.Continue, nil, nil
	}

	name, _ := tok.Value.(string)
	suffix := ""
	if n := len(name); n > 0 && name[n-1] == '$' {
		suffix = "$"
		name = name[:n-1]
	}
	key := name + "@" + tok.Pos.File()

	if len(name) <= variableAliasLen {
		return core.Continue, nil, nil
	}
	if ctx.Symbols.HardLongVars[key] {
		return core.Continue, nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  "ZX-81 variables must be one letter, optionally followed by $.",
			Pos:      &tok.Pos,
		}
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
			Message:  "Too many variables used (max=26).",
			Pos:      &tok.Pos,
		}
	}
	ctx.Symbols.Declares[key] = alias
	tok.Value = alias + suffix
	return core.Replace, []core.Token{tok}, nil
}

// Pass5 has no ZX81-specific work yet. MSX's own Pass5 (TRUE/FALSE,
// compound-assignment expansion, Unicode translation) is shared
// Dignified behavior that badig_zx81.py's Parser never overrides
// either (it inherits pass_5 from MSXParser unchanged) - but since
// this port doesn't share code between platform packages, those
// pieces would need to be re-ported here verbatim to get the same
// behavior for ZX81. Deliberately deferred: no real ZX81 .dzx
// exercising x+=5/TRUE/FALSE exists yet to validate against (same
// "confirm against something real before porting" discipline as
// Pass1's own deferral above) - revisit once Etapa 4's end-to-end
// validation program needs one of these.
func (*Parser) Pass5(_ *core.CompileContext, _ []core.Token, _ int) (core.HookAction, []core.Token, error) {
	return core.Continue, nil, nil
}

// Generate mirrors the one ZX81-specific rule in generate()
// (badig_zx81.py:380-383): ZX-81 statements must always start with a
// keyword - classic BASIC has no bare assignment syntax - so a Dignified
// `a = 1` (no explicit LET) needs "let "/"LET " inserted automatically
// right before the identifier, whenever that identifier opens a new
// statement: immediately after the line number, or immediately after
// "THEN". core.Generate's own shared spacing pass (pipeline.go) runs
// AFTER this hook, over whatever token sequence it returns - so
// returning the synthetic "let" instruction token followed by the
// original identifier is enough; the shared pass adds the space
// between them on its own, the same way any other reserved-keyword-
// before-identifier adjacency already does.
func (*Parser) Generate(_ *core.CompileContext, tokens []core.Token, index int) (core.HookAction, []core.Token, error) {
	tok := tokens[index]
	if tok.Kind != core.KindIdentifier || index == 0 {
		return core.Continue, nil, nil
	}
	previous := tokens[index-1]
	if previous.Kind != core.KindLineNumber && !strings.EqualFold(previous.String(), "then") {
		return core.Continue, nil, nil
	}
	return core.Replace, []core.Token{
		{Kind: KindInstruction, Value: "let", Pos: tok.Pos},
		tok,
	}, nil
}

// BuiltinDefines mirrors badig_zx81.py's Parser.initialization +
// replace_print_at: ZX81 pre-registers "[?](row,col)" as shorthand
// for "print at row,col;" in every compile, unconditionally (never
// declared by the user with their own "DEFINE [?][...]" line) -
// confirmed used that way in the project's own research into real
// ZX81 Dignified source (e.g. `[?](10,5) "ZX-81"` -> `print at
// 10,5;"zx-81"`).
//
// Uses Expand (ver contract.go's own doc comment on BuiltinDefine)
// rather than a Content placeholder, because a missing coordinate
// needs to default to "0" INDEPENDENTLY per argument (row, then col),
// not as one fixed whole-call default the way MSX's own simpler
// "[?](x,y)" macro gets away with (its DefaultVar "0,0" only ever
// applies when NO arguments are given at all, never to just one of
// two). printAtExpand receives exactly the call's resolved argument
// tokens (its own explicit "(...)" contents, or "0,0" if called bare)
// and does the comma-splitting/defaulting itself.
func (*Parser) BuiltinDefines() map[string]core.BuiltinDefine {
	return map[string]core.BuiltinDefine{
		"?": {
			DefaultVar: []core.Token{
				{Kind: KindNumber, Value: "0"},
				{Kind: KindSymbol, Value: ","},
				{Kind: KindNumber, Value: "0"},
			},
			Expand: printAtExpand,
		},
	}
}

// printAtExpand splits args on a top-level comma (not inside parens)
// into (row, col), defaulting either side to "0" when empty, and
// renders "print at <row>,<col>;" - mirrors replace_print_at
// (badig_zx81.py:291-323) exactly, except it operates directly on the
// already-resolved argument tokens instead of re-scanning a marker-
// delimited span (ver BuiltinDefines' own doc comment for why that
// simplification is safe). A THIRD top-level comma (if any) is not
// treated as another separator - it and everything after it just
// become part of the col side's own token list, same as the original
// (its own `if len(coordinates) == 1` guard only ever fires once).
func printAtExpand(_ *core.CompileContext, args []core.Token) ([]core.Token, error) {
	var coords [2][]core.Token
	side := 0
	depth := 0
	for _, tok := range args {
		switch v := tok.String(); {
		case v == "(":
			depth++
		case v == ")" && depth > 0:
			depth--
		case depth == 0 && v == "," && side == 0:
			side = 1
			continue
		}
		coords[side] = append(coords[side], tok)
	}
	if len(coords[0]) == 0 {
		coords[0] = []core.Token{{Kind: KindNumber, Value: "0"}}
	}
	if len(coords[1]) == 0 {
		coords[1] = []core.Token{{Kind: KindNumber, Value: "0"}}
	}

	out := []core.Token{
		{Kind: KindInstruction, Value: "print"},
		{Kind: KindFunction, Value: "at"},
	}
	out = append(out, coords[0]...)
	out = append(out, core.Token{Kind: KindSymbol, Value: ","})
	out = append(out, coords[1]...)
	out = append(out, core.Token{Kind: KindSymbol, Value: ";"})
	return out, nil
}

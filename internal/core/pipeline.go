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

			if name := functionNamePattern.FindString(remaining); name != "" {
				tokens = append(tokens, pin(Token{Kind: KindFunctionName, Value: name}))
				n := len([]rune(name))
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
	functions := map[string]functionEntry{}
	var out []Token
	i := 0

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

		if tok.Kind == KindDignifiedInstruction {
			upper := strings.ToUpper(tok.String())
			if (upper == "DEFINE" || upper == "FUNC" || upper == "RET" || upper == "DECLARE") && !atOutputLineStart() {
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
// expansion (which happens inside pass 3 in the original engine) is
// not implemented yet (ver OUTLINE.md, Bloco 1, Fase 3).
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

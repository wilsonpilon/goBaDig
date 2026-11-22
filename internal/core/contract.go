package core

// Description exposes a dialect's vocabulary and basic identity.
//
// In the original engine this is, in practice, a single giant regex
// alternation (classic_groups | dignified_groups | main_groups),
// where EVERY GROUP A Description CONTRIBUTES HAS SIGNIFICANT ORDER
// (OUTLINE.md, seção 3.5: "ordem das alternativas importa"). The hand
// written scanner this interface drives (see Lex, in pipeline.go)
// consults ORDERED vocabulary rules, not unordered sets - validated
// for real against the MSX platform (internal/platforms/msx) in this
// round, after an earlier round only validated it against a fake one.
type Description interface {
	SysInfo() SystemInfo
	// QuoteChar is the character that opens and closes a string
	// literal (`"` for MSX; a dialect could in principle use another
	// one) - core.Lex's generic string-literal fallback uses this
	// instead of each platform reimplementing string scanning itself.
	QuoteChar() byte
}

// SystemInfo is a dialect's name and on-disk file extensions -
// mirrors sysinfo() in the original engine.
type SystemInfo struct {
	Name         string
	DignifiedExt string // ex.: ".dmx"
	ClassicExt   string // ex.: ".amx"
	BinaryExt    string // ex.: ".bmx"
}

// Lexer classifies whatever sits at the current scanning position.
//
// v0 of this interface (Bloco 1) handed the platform a single
// whitespace-delimited "word" candidate, mirroring the original
// engine's `self.clc.lexing()` fallback hook. Validating it against a
// real platform (MSX, this round) showed that doesn't work: numeric
// literals with a dialect-specific exponent suffix (MSX's `c_numdec`:
// optional `.`, optional `(D|E)(+|-)digits`, optional `%!#`) and
// compound symbols with no surrounding whitespace (`x+=5`) both need
// multi-character lookahead a generic word-splitter can't provide -
// exactly the same reason the Batoken tokenizer (internal/platforms/
// msx/tokenizer.go) hands its scanners the whole remaining line
// instead of a pre-cut word. Lex now does the same.
type Lexer interface {
	// Init runs once before scanning starts.
	Init(ctx *CompileContext)
	// Lex tries to classify whatever is at the start of remaining. If
	// it recognizes something, it returns how many characters it
	// consumed (consumed > 0) and ok=true, plus zero or more tokens -
	// zero lets a platform consume and discard text that produces no
	// token of its own (e.g. a separator its own handling already
	// absorbed), and more than one lets it handle something like an
	// MSX `DATA` statement's comma-separated items in a single call,
	// instead of forcing one token per call the way v0 of this
	// interface (Bloco 1) did. ok=false means "nothing dialect-specific
	// starts here" - the driver then tries its own generic fallbacks
	// (a Dignified identifier, a quoted string) before giving up and
	// emitting one raw character, mirroring the original engine's
	// priority: dialect keywords are checked before the shared
	// Dignified identifier pattern (confirmed by the combined-regex
	// ordering "classic_groups | dignified_groups | main_groups |
	// d_idnttp" - identifiers, d_idnttp, come last). A non-nil error is
	// fatal and aborts the whole pipeline immediately (see Diagnostic).
	Lex(ctx *CompileContext, remaining string) (consumed int, tokens []Token, ok bool, err error)
}

// Parser has one method per pass, called by the core pipeline driver
// for every token in that pass - the fallback "else" branch of each
// pass_N in the original engine (e.g. `self.clc.pass_1()`).
//
// Resumo de cada pass (confirmado lendo badig.py, docstrings + corpo):
//
//	Pass1    preparação: toggle de rem, instruções Dignified
//	         (DEFINE/DECLARE/KEEP/FUNC/RET/EXIT/INCLUDE/ENDIF), expande defines
//	Pass2    registra fluxo: resolve chamadas de função e labels
//	Pass3    adiciona/remove: expande includes, blocos de rem, remove
//	         newlines excedentes, junta linhas/strings
//	Pass4    (só arquivo principal) aplica fluxo: troca labels por
//	         linhas, numera, insere header
//	Pass5    (só arquivo principal) alterações classic-specific, capitalização
//	Generate produz a listagem classic final
//
// Nenhuma dessas responsabilidades de nível "core" (DEFINE/INCLUDE/
// numeração/etc.) está implementada ainda - ver OUTLINE.md, Bloco 1,
// Fase 3: o driver de hoje só chama Pass1..5/Generate por token, sem
// nenhum processamento Dignified próprio ainda.
//
// Each method gets the pass's whole INPUT token slice plus the index
// of the one it must classify this call - not just that one token in
// isolation, as v0 of this interface (Bloco 1) did. MSX's real Pass 3
// (Bloco 2) needed to look at the PREVIOUS token to detect the
// Dignified "keep long name" symbol (`~`) immediately before an
// identifier (badig_msx.py's get_hard_variable checks
// `self.par.last_tok()`) - a single-token hook can't do that. Still
// just classifies the one token at tokens[index] and returns a
// HookAction for it; it just gets to look at tokens[index-1]/[index+1]/
// etc. too, for context, without needing to retroactively change a
// decision already made about a neighbor.
//
// On Replace, a method returns []Token (not Token, v2 of this
// interface, Bloco 2) - MSX's real Pass 5 needed to expand ONE input
// token into several (`x+=5` -> `=`, `x`, `+`, badig_msx.py's pass_5),
// which a single-Token Replace can't express. 0 tokens behaves like
// Skip, 1 like the old single-Token Replace; several is the new case.
type Parser interface {
	Pass1(ctx *CompileContext, tokens []Token, index int) (HookAction, []Token, error)
	Pass2(ctx *CompileContext, tokens []Token, index int) (HookAction, []Token, error)
	Pass3(ctx *CompileContext, tokens []Token, index int) (HookAction, []Token, error)
	Pass4(ctx *CompileContext, tokens []Token, index int) (HookAction, []Token, error)
	Pass5(ctx *CompileContext, tokens []Token, index int) (HookAction, []Token, error)
	Generate(ctx *CompileContext, tokens []Token, index int) (HookAction, []Token, error)
}

// Settings holds dialect-specific config, applied after the shared
// CompileSettings is already resolved - mirrors `c_stg.properties(self)`
// in the original engine's Settings.init(): the platform reads
// already-resolved shared values and computes its own derived config.
type Settings interface {
	Apply(shared *CompileSettings) error
}

// GenerateResult is the 3-tuple generate() returns in the original
// engine: the final classic listing, the variable-substitution map
// (relatório -var) and the Dignified<->Classic line correspondence
// (relatório -lnr).
type GenerateResult struct {
	Code        []string
	VariableMap map[string]string
	LineReport  []LineCorrespondence
}

// LineCorrespondence maps one Dignified source line to the Classic
// line(s) it generated.
type LineCorrespondence struct {
	DignifiedLine int
	ClassicLine   int
}

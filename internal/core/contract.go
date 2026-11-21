package core

// Description exposes a dialect's vocabulary and basic identity.
//
// In the original engine this is, in practice, a single giant regex
// alternation (classic_groups | dignified_groups | main_groups),
// where EVERY GROUP A Description CONTRIBUTES HAS SIGNIFICANT ORDER
// (OUTLINE.md, seção 3.5: "ordem das alternativas importa"). The hand
// written scanner this interface will eventually drive must consult
// ORDERED vocabulary rules, not unordered sets - the exact shape of
// that access is deliberately left out of this "v0" interface; it
// gets validated against the fake platform in pipeline_test.go before
// any real dialect depends on it (ver OUTLINE.md, Bloco 1, Fase 2/3).
type Description interface {
	SysInfo() SystemInfo
}

// SystemInfo is a dialect's name and on-disk file extensions -
// mirrors sysinfo() in the original engine.
type SystemInfo struct {
	Name         string
	DignifiedExt string // ex.: ".dmx"
	ClassicExt   string // ex.: ".amx"
	BinaryExt    string // ex.: ".bmx"
}

// Lexer is called by the core scanner as a fallback for every
// contiguous run of non-whitespace characters it doesn't itself
// classify - exactly as the original engine's `self.clc.lexing()`
// hook (badig.py).
type Lexer interface {
	// Init runs once before scanning starts.
	Init(ctx *CompileContext)
	// Lex classifies a candidate token. A non-nil error is fatal and
	// aborts the whole pipeline immediately (see Diagnostic).
	Lex(ctx *CompileContext, tok Token) (HookAction, Token, error)
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
type Parser interface {
	Pass1(ctx *CompileContext, tok Token) (HookAction, Token, error)
	Pass2(ctx *CompileContext, tok Token) (HookAction, Token, error)
	Pass3(ctx *CompileContext, tok Token) (HookAction, Token, error)
	Pass4(ctx *CompileContext, tok Token) (HookAction, Token, error)
	Pass5(ctx *CompileContext, tok Token) (HookAction, Token, error)
	Generate(ctx *CompileContext, tok Token) (HookAction, Token, error)
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

package core

// Platform bundles one dialect's pluggable pieces - the Go equivalent
// of the Classic/Tools/Dignified trio the original engine assembles
// via dynamic import + global reassignment at process start
// (support/badig_settings.py). Go has no equivalent of that trick, so
// a Platform value is chosen and passed explicitly instead (decisão
// já tomada em OUTLINE.md, seção 6, item 1).
//
// Expose and Tools are deliberately not fields here yet: those
// interfaces aren't defined until Fase 2 is finished for real (ver
// OUTLINE.md, Bloco 1, Fase 2) - nothing else about Platform needs to
// change when they're added.
//
// The shared Dignified vocabulary (DEFINE/DECLARE/INCLUDE/KEEP/ENDIF/
// FUNC/RET/EXIT, TRUE/FALSE, símbolos `[ ] { } @ ~` etc.) is NOT a
// Platform field either: in the original engine it is the same object
// for every dialect (only Classic.Description varies), so in Go it
// belongs to internal/core itself, as fixed data, once Fase 3's
// Dignified-level processing is implemented.
type Platform struct {
	Description Description
	Lexer       Lexer
	Parser      Parser
	Settings    Settings
}

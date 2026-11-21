package core

// CompileSettings is the fully-resolved result of the layered config
// decided in OUTLINE.md, seção 6, item 7: SQLite global < SQLite
// projeto < flags de CLI < remtags no código-fonte. Resolving those
// layers for real is deferred (see OUTLINE.md, Bloco 1, Fase 4) - this
// type exists now only so CompileContext and Platform.Settings have
// something concrete to work with; fields grow as each layer is
// actually implemented.
type CompileSettings struct {
	LineStart int // primeira linha numerada (Python default: 10)
	LineStep  int // passo de numeração (Python default: 10)
}

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

	// RemHeader mirrors self.stg.rem_header (support/badig_settings.py:46,
	// default True in Python; a plain -rh CLI flag toggles it off,
	// support/badig_settings.py:432) - whether pass4Numbering inserts
	// the 2-line "Converted with Basic Dignified" REM header
	// (insert_rem_header, badig.py:1289-1303/1643-1645).
	//
	// Deliberately NOT defaulting to true here the way Python does:
	// Go's own zero-value for a bool is false, and no CLI wiring
	// exists yet to pass an explicit value either way - flipping the
	// zero-value to "header on by default" would silently start
	// generating a header for every existing CompileSettings{...}
	// literal in this codebase's own tests, none of which opts in
	// today. Whenever real CLI wiring exists, IT decides the true
	// default (matching Python's), not this struct's zero-value.
	RemHeader bool

	// TabLength mirrors self.stg.tab_lenght (support/badig_settings.py:58,
	// default 4) - how many columns a TAB advances to when
	// core.LoadLines expands it (support/helper.py's IO.load_file:
	// `line.expandtabs(tab_lenght)`, badig.py:88). Zero (Go's own
	// zero-value) means "don't expand" here, same reasoning as
	// RemHeader above - no CLI wiring exists yet to supply Python's
	// own default, so callers that care pass it explicitly.
	TabLength int
}

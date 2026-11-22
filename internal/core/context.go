package core

// SymbolTables holds the variable/declare bookkeeping that must
// survive across INCLUDE boundaries by explicit reference, instead of
// the original engine's convention of passing a tuple of three dicts
// (c_hard_short_vars, c_hard_long_vars, d_declares) back out of each
// recursive par() call (OUTLINE.md, seção 3.4/3.5). Field shapes
// confirmed reading badig_msx.py's get_hard_variable/process_variable
// (the first real Parser passes to use them - Bloco 2) line by line;
// an earlier Bloco 1 version of this type had two vague, never-used
// fields instead.
type SymbolTables struct {
	// Declares maps "longName@file" to the short alias already
	// allocated for it (d_declares) - both the allocator's own memory
	// of what it already handed out, and the record a repeat
	// reference to the same long name reuses.
	Declares map[string]string
	// HardShortVars is the set of short names (≤ a category's alias
	// length) a long name must never be shortened to, because some
	// variable already uses that exact short name directly in the
	// source (c_hard_short_vars).
	HardShortVars map[string]bool
	// HardLongVars is the set of "longName@file" kept verbatim
	// (never shortened) because the source marked them with the
	// Dignified "keep long name" symbol, `~` (c_hard_long_vars).
	HardLongVars map[string]bool
}

// NewSymbolTables returns an empty, ready-to-use SymbolTables.
func NewSymbolTables() *SymbolTables {
	return &SymbolTables{
		Declares:      make(map[string]string),
		HardShortVars: make(map[string]bool),
		HardLongVars:  make(map[string]bool),
	}
}

// CompileContext carries everything a Lexer/Parser hook needs, passed
// explicitly on every call instead of hanging off dynamic instance
// attributes the way the original engine does (decisão já tomada em
// OUTLINE.md, seção 6, item 4).
type CompileContext struct {
	Source      []Line
	Settings    *CompileSettings
	Symbols     *SymbolTables
	Diagnostics DiagnosticSink

	// IncludeStack names the files open right now, outermost first -
	// the current file is always the last entry.
	IncludeStack []string

	// LoadFile reads a source file from disk into Source-ready Lines
	// for INCLUDE (badig.py's IO.load_file, support/helper.py) -
	// injected rather than hardcoded to os.ReadFile, same principle as
	// Platform's own Description/Lexer/Parser (OUTLINE.md, seção 6,
	// item 1): keeps core.Parse's INCLUDE handling testable without
	// real files on disk. LoadLines (pipeline.go) is the real,
	// ready-to-use implementation. Left nil when the compile never
	// uses INCLUDE; attempting to use INCLUDE with it nil is a fatal
	// Diagnostic, not a panic.
	LoadFile func(path string) ([]Line, error)
}

// IsMainFile reports whether this context is compiling the top-level
// source file rather than a file pulled in via INCLUDE. Parser passes
// 4 and 5 only run for the main file - mirrors par()'s
// `pass_1(); pass_2(); pass_3(); if is_main_file(): pass_4(); pass_5()`
// in the original engine.
func (c *CompileContext) IsMainFile() bool {
	return len(c.IncludeStack) <= 1
}

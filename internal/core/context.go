package core

// SymbolTables holds the variable/declare bookkeeping that must
// survive across INCLUDE boundaries by explicit reference, instead of
// the original engine's convention of passing a tuple of three dicts
// (c_hard_short_vars, c_hard_long_vars, d_declares) back out of each
// recursive par() call (OUTLINE.md, seção 3.4/3.5).
type SymbolTables struct {
	ShortVars map[string]string
	LongVars  map[string]string
	Declares  map[string]string
}

// NewSymbolTables returns an empty, ready-to-use SymbolTables.
func NewSymbolTables() *SymbolTables {
	return &SymbolTables{
		ShortVars: make(map[string]string),
		LongVars:  make(map[string]string),
		Declares:  make(map[string]string),
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
}

// IsMainFile reports whether this context is compiling the top-level
// source file rather than a file pulled in via INCLUDE. Parser passes
// 4 and 5 only run for the main file - mirrors par()'s
// `pass_1(); pass_2(); pass_3(); if is_main_file(): pass_4(); pass_5()`
// in the original engine.
func (c *CompileContext) IsMainFile() bool {
	return len(c.IncludeStack) <= 1
}

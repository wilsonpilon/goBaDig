package msx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// CompileToBinary runs the complete MSX conversion chain in one call:
// core.Run's Dignified→Classic text pipeline (Lex/Parse/Generate),
// then Tokenize encodes that already-numbered Classic listing into
// the tokenized binary .bmx format - mirrors the original engine's
// own two-stage flow (badig.py's dignified(), which saves the
// generated Classic text, then msx/tokenizer_interface.py's Tokenizer
// reads it back and runs MSXBatoken.Main().execute() on it), done
// here directly in memory instead of round-tripping through disk,
// since nothing about the original's split is a real behavioral
// requirement - just how its CLI-oriented, file-based stages happen
// to be wired.
func CompileToBinary(ctx *core.CompileContext, platform core.Platform) ([]byte, error) {
	result, err := core.Run(ctx, platform)
	if err != nil {
		return nil, err
	}
	return Tokenize(result.Code)
}

// DefaultCompileSettings mirrors badig.py's own real defaults
// (support/badig_settings.py: line_start=10, line_step=10,
// rem_header=True, tab_lenght=4) - the first real caller to need
// them, now that CompileFile is a genuine end-user entry point.
// core.CompileSettings's own zero-value deliberately differs
// (RemHeader false) only because no caller existed yet to decide a
// real default one way or the other; this is that decision.
func DefaultCompileSettings() *core.CompileSettings {
	return &core.CompileSettings{
		LineStart: 10,
		LineStep:  10,
		RemHeader: true,
		TabLength: 4,
	}
}

// CompileFile reads a Dignified MSX source file (dmxPath), runs it
// through the full conversion chain (core.Run, then Tokenize) and
// writes both outputs next to dmxPath, under its own base name:
// amxPath (the numbered Classic text listing) and bmxPath (the
// tokenized binary) - mirrors badig.py's own dignified() end to end
// for the first time (ver OUTLINE.md: every piece this chains
// together - core.Run, Tokenize, core.LoadLines - already existed
// and was validated on its own; this is the first real "do it to an
// actual file on disk" entry point that ties them together).
//
// settings may be nil, in which case DefaultCompileSettings is used -
// every caller so far (the CLI, below) passes nil, since there is no
// real flag/option wiring yet to resolve a different value from.
//
// The Classic listing is written with CRLF line endings: badig.py's
// own generate() always appends a bare "\n" per line (badig.py:1811,
// self.des.newline_str_n - "\r" on non-Windows, but this project is
// developed and validated on Windows throughout), and IO.save_file
// opens its output in Python's text mode ('w'), which on Windows
// translates every "\n" it writes into "\r\n" - confirmed against
// real output from this project's own reference badig.py runs on
// this machine.
func CompileFile(dmxPath string, settings *core.CompileSettings) (amxPath, bmxPath string, err error) {
	if settings == nil {
		settings = DefaultCompileSettings()
	}

	absPath, err := filepath.Abs(dmxPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve path %s: %w", dmxPath, err)
	}
	lines, err := core.LoadLines(absPath, settings.TabLength)
	if err != nil {
		return "", "", fmt.Errorf("load %s: %w", dmxPath, err)
	}

	ctx := &core.CompileContext{
		Source:       lines,
		Settings:     settings,
		Symbols:      core.NewSymbolTables(),
		IncludeStack: []string{absPath},
		LoadFile: func(p string) ([]core.Line, error) {
			return core.LoadLines(p, settings.TabLength)
		},
	}
	platform := core.Platform{Description: Description, Lexer: Lexer{}, Parser: NewParser()}

	result, err := core.Run(ctx, platform)
	if err != nil {
		return "", "", fmt.Errorf("compile %s: %w", dmxPath, err)
	}
	tokenized, err := Tokenize(result.Code)
	if err != nil {
		return "", "", fmt.Errorf("tokenize %s: %w", dmxPath, err)
	}

	base := strings.TrimSuffix(dmxPath, filepath.Ext(dmxPath))
	amxPath = base + ".amx"
	bmxPath = base + ".bmx"

	amxContent := strings.Join(result.Code, "\r\n") + "\r\n"
	if err := os.WriteFile(amxPath, []byte(amxContent), 0o644); err != nil {
		return "", "", fmt.Errorf("write %s: %w", amxPath, err)
	}
	if err := os.WriteFile(bmxPath, tokenized, 0o644); err != nil {
		return "", "", fmt.Errorf("write %s: %w", bmxPath, err)
	}
	return amxPath, bmxPath, nil
}

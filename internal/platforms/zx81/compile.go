package zx81

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// DefaultCompileSettings mirrors badig.py's own real defaults
// (support/badig_settings.py: line_start=10, line_step=10,
// rem_header=True, tab_lenght=4) - same values as MSX's own
// DefaultCompileSettings (internal/platforms/msx/compile.go), since
// these are shared, dialect-agnostic settings, not ZX81-specific
// ones.
func DefaultCompileSettings() *core.CompileSettings {
	return &core.CompileSettings{
		LineStart: 10,
		LineStep:  10,
		RemHeader: true,
		TabLength: 4,
	}
}

// CompileFile reads a Dignified ZX81 source file (dzxPath), runs it
// through the full conversion chain (core.Run, then Tokenize) and
// writes both outputs next to dzxPath, under its own base name:
// azxPath (the numbered Classic text listing) and pPath (the
// tokenized ".p" memory image) - mirrors msx.CompileFile exactly
// (internal/platforms/msx/compile.go), ZX81's own extensions
// (Description.SysInfo) and tokenizer swapped in.
//
// settings may be nil, in which case DefaultCompileSettings is used.
//
// The Classic listing is written with CRLF line endings - same
// reasoning as msx.CompileFile's own doc comment (badig.py's
// generate() always appends a bare "\n", and IO.save_file's Python
// text-mode write translates that to "\r\n" on Windows, confirmed
// against real badig.py output on this machine).
func CompileFile(dzxPath string, settings *core.CompileSettings) (azxPath, pPath string, err error) {
	if settings == nil {
		settings = DefaultCompileSettings()
	}

	absPath, err := filepath.Abs(dzxPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve path %s: %w", dzxPath, err)
	}
	lines, err := core.LoadLines(absPath, settings.TabLength)
	if err != nil {
		return "", "", fmt.Errorf("load %s: %w", dzxPath, err)
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
		return "", "", fmt.Errorf("compile %s: %w", dzxPath, err)
	}
	tokenized, err := Tokenize(result.Code)
	if err != nil {
		return "", "", fmt.Errorf("tokenize %s: %w", dzxPath, err)
	}

	base := strings.TrimSuffix(dzxPath, filepath.Ext(dzxPath))
	azxPath = base + ".azx"
	pPath = base + ".p"

	azxContent := strings.Join(result.Code, "\r\n") + "\r\n"
	if err := os.WriteFile(azxPath, []byte(azxContent), 0o644); err != nil {
		return "", "", fmt.Errorf("write %s: %w", azxPath, err)
	}
	if err := os.WriteFile(pPath, tokenized, 0o644); err != nil {
		return "", "", fmt.Errorf("write %s: %w", pPath, err)
	}
	return azxPath, pPath, nil
}

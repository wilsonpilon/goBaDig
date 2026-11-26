package zxspectrum

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// DefaultCompileSettings mirrors badig.py's own real defaults
// (support/badig_settings.py: line_start=10, line_step=10,
// rem_header=True, tab_lenght=4) - same values as MSX/ZX81's own
// DefaultCompileSettings, since these are shared, dialect-agnostic
// settings, not Spectrum-specific ones.
func DefaultCompileSettings() *core.CompileSettings {
	return &core.CompileSettings{
		LineStart: 10,
		LineStep:  10,
		RemHeader: true,
		TabLength: 4,
	}
}

// CompileFile reads a Dignified ZX Spectrum source file (dzsPath),
// runs it through the full conversion chain (core.Run, then Tokenize,
// then MakeTap) and writes both outputs next to dzsPath, under its
// own base name: azsPath (the numbered Classic text listing) and
// tapPath (the tokenized .tap tape image) - mirrors msx.CompileFile/
// zx81.CompileFile, Spectrum's own extensions (Description.SysInfo)
// and tokenizer/tape packaging swapped in.
//
// Always targets the 48K model - the 128K model (SPECTRUM/PLAY
// tokens, a different EightyOne profile) is deliberately out of
// scope for this round's emulator integration (ver OUTLINE.md, Bloco
// 4 - the tokenizer itself, Etapa 1, already supports model="128" for
// whoever calls Tokenize directly; there just isn't a --sp_model-style
// setting wired to this entry point yet).
//
// settings may be nil, in which case DefaultCompileSettings is used.
//
// The Classic listing is written with CRLF line endings - same
// reasoning as msx.CompileFile's own doc comment (badig.py's
// generate() always appends a bare "\n", and IO.save_file's Python
// text-mode write translates that to "\r\n" on Windows, confirmed
// against real badig.py output on this machine).
func CompileFile(dzsPath string, settings *core.CompileSettings) (azsPath, tapPath string, err error) {
	if settings == nil {
		settings = DefaultCompileSettings()
	}

	absPath, err := filepath.Abs(dzsPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve path %s: %w", dzsPath, err)
	}
	lines, err := core.LoadLines(absPath, settings.TabLength)
	if err != nil {
		return "", "", fmt.Errorf("load %s: %w", dzsPath, err)
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
		return "", "", fmt.Errorf("compile %s: %w", dzsPath, err)
	}
	program, err := Tokenize(strings.Join(result.Code, "\n"), "48")
	if err != nil {
		return "", "", fmt.Errorf("tokenize %s: %w", dzsPath, err)
	}

	base := strings.TrimSuffix(dzsPath, filepath.Ext(dzsPath))
	azsPath = base + ".azs"
	tapPath = base + ".tap"

	tap, err := MakeTap(program, tapeName(base), nil)
	if err != nil {
		return "", "", fmt.Errorf("package %s: %w", dzsPath, err)
	}

	azsContent := strings.Join(result.Code, "\r\n") + "\r\n"
	if err := os.WriteFile(azsPath, []byte(azsContent), 0o644); err != nil {
		return "", "", fmt.Errorf("write %s: %w", azsPath, err)
	}
	if err := os.WriteFile(tapPath, tap, 0o644); err != nil {
		return "", "", fmt.Errorf("write %s: %w", tapPath, err)
	}
	return azsPath, tapPath, nil
}

// tapeName derives a .tap header name from a source file's own base
// name (no directory, no extension) - mirrors the original's own
// TK_NAME default ("nome do arquivo truncado em 10 chars", per the
// research this round started from): MakeTap requires 1..10 printable
// ASCII characters, so anything longer is simply truncated.
func tapeName(base string) string {
	name := filepath.Base(base)
	if len(name) > 10 {
		name = name[:10]
	}
	return name
}

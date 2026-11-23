package msx

import "github.com/wilsonpilon/goBaDig/internal/core"

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

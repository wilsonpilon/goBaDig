package msx

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/core"
)

// expandIncbload is the Expand callback for the "[INCBLOAD]("file.asm")"
// Dignified macro (Fase 5 of the autocomplete+man/help+ASMSX plan, ver
// OUTLINE.md): the Assembly->BASIC bridge the user asked for after
// Fase 4 shipped ASMSX compilation itself. Written as [INCBLOAD](...)
// rather than auto-inserted text, so nothing ever edits the user's own
// .dmx - every compile just re-reads whatever "file.asm"/"file.bin"/
// "file.sym" look like right now and expands to the equivalent of
// `BLOAD"FILE.BIN":DEFUSR=&Hxxxx`, letting the program call the
// routine with `x=USR(0)`/`print USR(0)` wherever it needs to. No ",R"
// on BLOAD: that flag auto-runs the binary once on load, which fits a
// standalone machine-code program taking over the whole machine, not
// a subroutine meant to be called (repeatedly) via USR() from BASIC.
//
// Requires the companion .asm to already be compiled (its .bin must
// exist) - this macro only READS already-produced artifacts, it never
// shells out to asmsx itself, keeping BASIC compilation and ASM
// compilation as two separate, already-existing steps (compile
// foo.asm, then compile foo.dmx) instead of coupling them.
func expandIncbload(ctx *core.CompileContext, args []core.Token) ([]core.Token, error) {
	if len(args) != 1 || args[0].Kind != core.KindString {
		pos := core.Position{}
		if len(args) > 0 {
			pos = args[0].Pos
		}
		return nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  `INCBLOAD espera um único nome de arquivo .asm entre aspas, ex.: [INCBLOAD]("rotina.asm")`,
			Pos:      &pos,
		}
	}
	nameTok := args[0]

	asmRelPath := strings.TrimSpace(strings.Trim(nameTok.String(), `"`))
	if asmRelPath == "" || !strings.EqualFold(filepath.Ext(asmRelPath), ".asm") {
		return nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  fmt.Sprintf("INCBLOAD espera um arquivo .asm: %s", nameTok.String()),
			Pos:      &nameTok.Pos,
		}
	}

	// Relative paths resolve against the MAIN file's own directory,
	// same rule INCLUDE itself already follows (ver resolveInclude).
	asmPath := asmRelPath
	if !filepath.IsAbs(asmPath) && len(ctx.IncludeStack) > 0 {
		asmPath = filepath.Join(filepath.Dir(ctx.IncludeStack[0]), asmPath)
	}

	binPath := strings.TrimSuffix(asmPath, filepath.Ext(asmPath)) + ".bin"
	if _, err := os.Stat(binPath); err != nil {
		return nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message: fmt.Sprintf(
				"INCBLOAD: %s não foi encontrado - compile %s com o asmsx antes de compilar este programa",
				filepath.Base(binPath), filepath.Base(asmPath),
			),
			Pos: &nameTok.Pos,
		}
	}

	address, err := resolveEntryAddress(asmPath)
	if err != nil {
		return nil, core.Diagnostic{
			Severity: core.SeverityError,
			Message:  fmt.Sprintf("INCBLOAD: %v", err),
			Pos:      &nameTok.Pos,
		}
	}

	binName := strings.ToUpper(filepath.Base(binPath))
	return []core.Token{
		{Kind: KindInstruction, Value: "bload", Pos: nameTok.Pos},
		{Kind: core.KindString, Value: `"` + binName + `"`, Pos: nameTok.Pos},
		{Kind: KindSeparator, Value: ":", Pos: nameTok.Pos},
		{Kind: KindInstruction, Value: "defusr", Pos: nameTok.Pos},
		{Kind: KindSymbol, Value: "=", Pos: nameTok.Pos},
		{Kind: KindNumber, Value: fmt.Sprintf("&H%X", address), Pos: nameTok.Pos},
	}, nil
}

// resolveEntryAddress finds the entry-point address a compiled .asm's
// companion BLOAD+DEFUSR line should use: the ".START" directive's
// argument if present, otherwise ".ORG"'s (mirrors asMSX's own
// run_address defaulting to start_address when .START is absent - ver
// resource/asMSX/src/dura.y:4201-4202). The argument is either a bare
// numeric literal (parseAsmNumber) or a bare label name, resolved by
// reading the .sym file asmsx itself produces alongside the .bin
// (resource/asMSX/src/dura.y:3974 write_sym, only emitted when the
// program defines at least one symbol).
//
// Known v1 limitation, not solved here: a label whose name happens to
// look like a valid "<hex digits>h" literal (e.g. a label named "ABh")
// is read as the number 0xAB instead of being looked up as a symbol -
// the same ambiguity asMSX's own grammar resolves with full symbol-
// table context, which this textual scan does not have. Avoid such
// label names on a .START/.ORG line used with INCBLOAD.
func resolveEntryAddress(asmPath string) (int, error) {
	raw, err := os.ReadFile(asmPath)
	if err != nil {
		return 0, fmt.Errorf("não foi possível ler %s: %w", filepath.Base(asmPath), err)
	}
	text := string(raw)

	arg, ok := findLastDirectiveArg(text, ".START")
	if !ok {
		arg, ok = findLastDirectiveArg(text, ".ORG")
	}
	if !ok {
		return 0, fmt.Errorf("%s não tem diretiva .START nem .ORG - não há como saber o endereço de entrada", filepath.Base(asmPath))
	}

	if value, ok := parseAsmNumber(arg); ok {
		return value, nil
	}

	symPath := strings.TrimSuffix(asmPath, filepath.Ext(asmPath)) + ".sym"
	symbols, err := parseSymFile(symPath)
	if err != nil {
		return 0, fmt.Errorf(
			"endereço de entrada é o símbolo %q, mas %s não foi encontrado ou não pôde ser lido (%w) - recompile o .asm com o asmsx",
			arg, filepath.Base(symPath), err,
		)
	}
	if value, ok := symbols[strings.ToUpper(arg)]; ok {
		return value, nil
	}
	return 0, fmt.Errorf("símbolo %q (do .START/.ORG de %s) não foi encontrado em %s", arg, filepath.Base(asmPath), filepath.Base(symPath))
}

// findLastDirectiveArg scans asmText for every unconditional line
// whose first word is directive (case-insensitive) and returns the
// LAST match's own second word - good enough for the common case of a
// single, unconditional ".START"/".ORG" line (ver resource/asMSX/doc/
// asmsx.md's own usage example, which places ".START" once, after any
// IFDEF/ENDIF block). Comments (";" to end of line) are stripped
// first so a commented-out directive is never matched.
func findLastDirectiveArg(asmText, directive string) (string, bool) {
	var last string
	found := false
	for line := range strings.SplitSeq(asmText, "\n") {
		if idx := strings.IndexByte(line, ';'); idx >= 0 {
			line = line[:idx]
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || !strings.EqualFold(fields[0], directive) {
			continue
		}
		last = fields[1]
		found = true
	}
	return last, found
}

// parseAsmNumber parses one of asMSX's 3 documented hexadecimal
// literal conventions (resource/asMSX/doc/asmsx.md: "0x" prefix, "$"
// prefix, or trailing "h"/"H") or a plain decimal integer. ok is false
// for anything else (a bare label name, most commonly).
func parseAsmNumber(tok string) (int, bool) {
	tok = strings.TrimSpace(tok)
	switch {
	case tok == "":
		return 0, false
	case strings.HasPrefix(tok, "0x"), strings.HasPrefix(tok, "0X"):
		v, err := strconv.ParseInt(tok[2:], 16, 64)
		return int(v), err == nil
	case strings.HasPrefix(tok, "$"):
		v, err := strconv.ParseInt(tok[1:], 16, 64)
		return int(v), err == nil
	case len(tok) > 1 && (tok[len(tok)-1] == 'h' || tok[len(tok)-1] == 'H'):
		v, err := strconv.ParseInt(tok[:len(tok)-1], 16, 64)
		return int(v), err == nil
	default:
		v, err := strconv.ParseInt(tok, 10, 64)
		return int(v), err == nil
	}
}

// parseSymFile parses a ".sym" file asmsx itself generates
// (resource/asMSX/src/dura.y:3974 write_sym): comment lines start with
// ";", data lines are "XXXXh NAME" (or, for a MEGAROM build this
// project never produces via INCBLOAD, "XXh:XXXXh NAME" - the page
// prefix before ":" is simply discarded). Returns symbol names
// uppercased, since MSX BASIC/asMSX identifiers are conventionally
// case-insensitive.
func parseSymFile(symPath string) (map[string]int, error) {
	raw, err := os.ReadFile(symPath)
	if err != nil {
		return nil, err
	}
	symbols := map[string]int{}
	for line := range strings.SplitSeq(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		addrField := fields[0]
		if idx := strings.IndexByte(addrField, ':'); idx >= 0 {
			addrField = addrField[idx+1:]
		}
		addrField = strings.TrimSuffix(strings.TrimSuffix(addrField, "H"), "h")
		value, err := strconv.ParseInt(addrField, 16, 64)
		if err != nil {
			continue
		}
		symbols[strings.ToUpper(fields[1])] = int(value)
	}
	return symbols, nil
}

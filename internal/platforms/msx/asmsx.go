package msx

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/options"
)

// ResolveAsmsxPath finds the asMSX Z80 cross-assembler to use for
// compiling ".asm" files inside an MSX BASIC project (Fase 4 of the
// autocomplete+man/help+ASMSX plan, ver OUTLINE.md): first the user's
// own ASMSX_PATH setting (internal/options' CategoryMSX), if set and
// pointing at a file that exists; otherwise the system PATH
// ("asmsx"). store may be nil (a caller with no settings database
// open yet just falls through to PATH) - EMULATOR_PATH's own existing
// entries follow the same "configured first, nothing else today"
// shape, except none of them are actually read anywhere yet (this is
// the first one that is).
//
// Deliberately out of scope: a "bundled next to goBaDig.exe" tier -
// no build pipeline today compiles asmsx.exe from the vendored C
// source in resource/asMSX/src (needs gcc+flex+bison, see
// resource/asMSX/Makefile); a future packaging step, not a blocker
// for this one.
func ResolveAsmsxPath(store *config.Store) (string, error) {
	if store != nil {
		if configured, err := store.GetOption(options.CategoryMSX, "ASMSX_PATH"); err == nil {
			configured = strings.TrimSpace(configured)
			if configured != "" {
				if _, statErr := os.Stat(configured); statErr == nil {
					return configured, nil
				}
			}
		}
	}
	path, err := exec.LookPath("asmsx")
	if err != nil {
		return "", fmt.Errorf("asmsx not found in PATH: %w", err)
	}
	return path, nil
}

// CompileAsm runs asmsxExe against asmPath (a ".asm" MSX Assembly
// source file) - the same subprocess-with-captured-output shape
// internal/editor/printer_unix.go already uses for the one other
// synchronous external-tool call in the project (every emulator
// launch, by contrast, is fire-and-forget - ver run.go). Runs inside
// asmPath's own directory, passing just the base filename, so
// asMSX's own relative #include paths resolve correctly; any failure
// wraps the captured output into the returned error.
//
// On success, binPath is the ".bin" asMSX produces by default. Other
// output extensions (.rom/.cas/.wav/.com) depend on a directive
// inside the .asm itself (rom/.megarom/basic/cas/wav) - detecting
// those is a known v1 gap (ver OUTLINE.md), not handled here.
func CompileAsm(asmPath, asmsxExe string) (binPath, output string, err error) {
	command := exec.Command(asmsxExe, filepath.Base(asmPath))
	command.Dir = filepath.Dir(asmPath)
	raw, runErr := command.CombinedOutput()
	output = string(bytes.TrimSpace(raw))
	if runErr != nil {
		return "", output, fmt.Errorf("asmsx: %w: %s", runErr, output)
	}

	binPath = strings.TrimSuffix(asmPath, filepath.Ext(asmPath)) + ".bin"
	if _, statErr := os.Stat(binPath); statErr != nil {
		return "", output, fmt.Errorf(
			"asmsx: compilação terminou sem erro, mas %s não foi gerado (verifique as diretivas de saída no .asm - rom/.megarom/basic/cas/wav)",
			filepath.Base(binPath),
		)
	}
	return binPath, output, nil
}

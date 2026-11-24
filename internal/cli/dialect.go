package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/platforms/msx"
	"github.com/wilsonpilon/goBaDig/internal/platforms/zx81"
)

// compileByExtension dispatches goBaDig's "compile a Dignified source
// file" entry point to the right dialect by the input file's own
// extension - .dmx for MSX, .dzx for ZX81 - picking automatically
// instead of requiring an explicit flag (ZX Spectrum has no engine of
// its own yet, so there is nothing else to dispatch to). Mirrors how
// the real badig.py picks a dialect via its own `-id` flag.
func compileByExtension(path string) (classicPath, binaryPath string, err error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".dmx":
		return msx.CompileFile(path, nil)
	case ".dzx":
		return zx81.CompileFile(path, nil)
	default:
		return "", "", fmt.Errorf("unrecognized dialect for extension %q (expected .dmx or .dzx)", filepath.Ext(path))
	}
}

// runInEmulatorByExtension dispatches goBaDig's "build a ready-to-run
// test artifact and open it in a real emulator" entry point the same
// way. MSX needs an extra disk-packaging step (msx.BuildTestDisk) that
// ZX81 doesn't - CompileFile's own ".p" output is already everything
// EightyOne needs - so each branch calls whatever its own dialect
// actually requires before launching its own emulator.
func runInEmulatorByExtension(path string) (artifactPath string, err error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".dmx":
		diskPath, err := msx.BuildTestDisk(path)
		if err != nil {
			return "", err
		}
		if err := msx.RunInEmulator(diskPath); err != nil {
			return "", err
		}
		return diskPath, nil
	case ".dzx":
		_, pPath, err := zx81.CompileFile(path, nil)
		if err != nil {
			return "", err
		}
		if err := zx81.RunInEmulator(pPath); err != nil {
			return "", err
		}
		return pPath, nil
	default:
		return "", fmt.Errorf("unrecognized dialect for extension %q (expected .dmx or .dzx)", filepath.Ext(path))
	}
}

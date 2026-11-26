package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/platforms/msx"
	"github.com/wilsonpilon/goBaDig/internal/platforms/zx81"
	"github.com/wilsonpilon/goBaDig/internal/platforms/zxspectrum"
)

// dialectExtensions lists the recognized input extensions, shared by
// both dispatchers' error messages so they never drift apart.
const dialectExtensions = ".dmx, .dzx or .dzs"

// compileByExtension dispatches goBaDig's "compile a Dignified source
// file" entry point to the right dialect by the input file's own
// extension - .dmx for MSX, .dzx for ZX81, .dzs for ZX Spectrum -
// picking automatically instead of requiring an explicit flag.
// Mirrors how the real badig.py picks a dialect via its own `-id`
// flag.
func compileByExtension(path string) (classicPath, binaryPath string, err error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".dmx":
		return msx.CompileFile(path, nil)
	case ".dzx":
		return zx81.CompileFile(path, nil)
	case ".dzs":
		return zxspectrum.CompileFile(path, nil)
	default:
		return "", "", fmt.Errorf("unrecognized dialect for extension %q (expected %s)", filepath.Ext(path), dialectExtensions)
	}
}

// runInEmulatorByExtension dispatches goBaDig's "build a ready-to-run
// test artifact and open it in a real emulator" entry point the same
// way. MSX needs an extra disk-packaging step (msx.BuildTestDisk) that
// ZX81/Spectrum don't - their own CompileFile output (".p"/".tap") is
// already everything their own emulator needs - so each branch calls
// whatever its own dialect actually requires before launching its own
// emulator.
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
	case ".dzs":
		_, tapPath, err := zxspectrum.CompileFile(path, nil)
		if err != nil {
			return "", err
		}
		if err := zxspectrum.RunInEmulator(tapPath); err != nil {
			return "", err
		}
		return tapPath, nil
	default:
		return "", fmt.Errorf("unrecognized dialect for extension %q (expected %s)", filepath.Ext(path), dialectExtensions)
	}
}

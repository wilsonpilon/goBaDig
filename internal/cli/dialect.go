package cli

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/platforms/msx"
	"github.com/wilsonpilon/goBaDig/internal/platforms/zx81"
	"github.com/wilsonpilon/goBaDig/internal/platforms/zxspectrum"
)

// dialectExtensions lists the recognized input extensions, shared by
// both dispatchers' error messages so they never drift apart.
const dialectExtensions = ".dmx, .dzx or .dzs"

// emulatorBackends lists, per extension, the valid `--emulator`/
// console `execute <file> <emulator>` names - the first entry in each
// list is that extension's own default (used when the caller passes
// ""), matching whatever RunInEmulator already did before any of this
// existed. Added 2026-11-26 at the user's own request: EightyOne/
// openMSX were the only backend each dialect could ever use before -
// ZEsarUX (ZX81/Spectrum) and fMSX/blueMSX+ (MSX) are real
// alternatives already installed on this machine, confirmed by
// actually launching each one. ZEsarUX was deliberately left out of
// MSX's own list - it only emulates MSX1, and the user judged that
// not worth adding alongside openMSX/fMSX/blueMSX+, which all cover
// MSX2/MSX2+ too.
var emulatorBackends = map[string][]string{
	".dmx": {"openmsx", "fmsx", "bluemsx+"},
	".dzx": {"eightyone", "zesarux"},
	".dzs": {"eightyone", "zesarux"},
}

// resolveBackend validates backend against ext's own allow-list
// (emulatorBackends), returning the extension's own default when
// backend is "". A backend name valid for a DIFFERENT dialect (e.g.
// "zesarux" against ".dmx") is still rejected - each dialect's own
// list is exactly the backends it actually supports, not a shared
// vocabulary.
func resolveBackend(ext, backend string) (string, error) {
	names, ok := emulatorBackends[ext]
	if !ok {
		return "", fmt.Errorf("unrecognized dialect for extension %q (expected %s)", ext, dialectExtensions)
	}
	if backend == "" {
		return names[0], nil
	}
	if slices.Contains(names, backend) {
		return backend, nil
	}
	return "", fmt.Errorf("unknown --emulator %q for extension %q (expected one of: %s)", backend, ext, strings.Join(names, ", "))
}

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
// way, honoring an explicit emulator backend when one is given
// (emulatorBackends) - "" picks each dialect's own default (the only
// backend each one had before 2026-11-26: openMSX/EightyOne/
// EightyOne). MSX needs an extra disk-packaging step
// (msx.BuildTestDisk) that ZX81/Spectrum don't - their own CompileFile
// output (".p"/".tap") is already everything their own emulator
// needs - so each branch calls whatever its own dialect actually
// requires before launching whichever backend was resolved.
func runInEmulatorByExtension(path, backend string) (artifactPath string, err error) {
	ext := strings.ToLower(filepath.Ext(path))
	resolved, err := resolveBackend(ext, backend)
	if err != nil {
		return "", err
	}

	switch ext {
	case ".dmx":
		diskPath, err := msx.BuildTestDisk(path)
		if err != nil {
			return "", err
		}
		switch resolved {
		case "fmsx":
			err = msx.RunInFMSX(diskPath)
		case "bluemsx+":
			err = msx.RunInBlueMSX(diskPath)
		default:
			err = msx.RunInEmulator(diskPath)
		}
		if err != nil {
			return "", err
		}
		return diskPath, nil
	case ".dzx":
		_, pPath, err := zx81.CompileFile(path, nil)
		if err != nil {
			return "", err
		}
		if resolved == "zesarux" {
			err = zx81.RunInZEsarUX(pPath)
		} else {
			err = zx81.RunInEmulator(pPath)
		}
		if err != nil {
			return "", err
		}
		return pPath, nil
	default: // ".dzs"
		_, tapPath, err := zxspectrum.CompileFile(path, nil)
		if err != nil {
			return "", err
		}
		if resolved == "zesarux" {
			err = zxspectrum.RunInZEsarUX(tapPath)
		} else {
			err = zxspectrum.RunInEmulator(tapPath)
		}
		if err != nil {
			return "", err
		}
		return tapPath, nil
	}
}

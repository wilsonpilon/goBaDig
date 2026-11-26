package zx81

import (
	"fmt"
	"os/exec"
)

// RunInEmulator launches EightyOne with pPath loaded directly -
// mirrors the real engine's own emulator_interface.py
// (`subprocess.Popen([emulator_path, file_save])`): EightyOne takes
// the ".p" file as its single positional argument, no flag at all
// (unlike openMSX's own "-diska <path>" - ver msx/run.go's
// RunInEmulator). Unlike MSX, there is no disk-wrapping step at all
// here: CompileFile's own ".p" output is already everything EightyOne
// needs to boot straight into the compiled program - ZX81 has no
// disk filesystem/AUTOEXEC convention to route through.
//
// EightyOne is located via PATH (same os/exec PATHEXT-aware lookup as
// MSX's own RunInEmulator) - no machine/settings customization yet
// (same deferred work as MSX's own doc comment notes). RunInEmulator
// does not wait for the emulator to exit.
func RunInEmulator(pPath string) error {
	exePath, err := exec.LookPath("eightyone")
	if err != nil {
		return fmt.Errorf("eightyone not found in PATH: %w", err)
	}
	cmd := exec.Command(exePath, pPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start eightyone: %w", err)
	}
	return nil
}

// RunInZEsarUX launches ZEsarUX with pPath loaded as its ZX81 tape -
// an alternative to EightyOne requested by the user (2026-11-26).
// Confirmed reading zesarux.exe's own `--help` text and by actually
// launching it this way on this machine: `--machine ZX81` selects the
// ZX81 model, `--tape <file>` accepts `.p` directly (listed under
// ZX81's own supported tape formats, alongside `.81`/`.z81` - a `.p`
// is ALSO listed as a valid `--snap` target, since it is really a
// full memory image rather than a true cassette recording, but
// `--tape` is what was confirmed to work end to end here). No
// machine/settings customization beyond that yet - same deferred
// scope as RunInEmulator's own doc comment. Does not wait for the
// emulator to exit, same as RunInEmulator.
func RunInZEsarUX(pPath string) error {
	exePath, err := exec.LookPath("zesarux")
	if err != nil {
		return fmt.Errorf("zesarux not found in PATH: %w", err)
	}
	cmd := exec.Command(exePath, "--machine", "ZX81", "--tape", pPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start zesarux: %w", err)
	}
	return nil
}

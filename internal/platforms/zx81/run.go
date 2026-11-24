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

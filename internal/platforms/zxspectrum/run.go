package zxspectrum

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// eightyOne48INI mirrors eightyone48.ini (badig/zxspectrum, read in
// full this session) - the default EightyOne profile for the 48K
// model. TZXMAN is already set to autostart - the real engine's own
// run() only ever rewrites these three keys to '1' when the tape file
// is .tap/.tzx (tape.suffix.lower() in ('.tap', '.tzx')) and this
// package's own CompileFile (Etapa 5) only ever produces .tap, so
// there is nothing to conditionally adjust here: the template's own
// defaults already match every case this port can produce.
const eightyOne48INI = `[MAIN]
ShowSplash=0

[HARDWARE]
MachineName=48K Spectrum
ROMSP48=spectrum48.rom

[TZXMAN]
FlashLoad=1
AutoStart=1
AutoLoad=1
`

// RunInEmulator launches EightyOne with a private, temporary copy of
// the 48K profile (eightyOne48INI) plus tapPath - mirrors the real
// engine's own emulator_interface.py, except for one deliberate
// divergence: the original WAITS for EightyOne to close
// (`process.wait()`) and fails loudly on a nonzero exit code; this
// does not, matching MSX/ZX81's own RunInEmulator instead (confirmed
// with the user, 2026-11-26, as the preferred behavior - consistency
// of goBaDig's own `execute` command across all three dialects over
// fidelity to Python on this one point).
//
// EightyOne takes its own INI profile AND the tape file as TWO
// positional arguments (`[ini, tape]`) - unlike ZX81/MSX, which each
// take only one. Always the 48K profile - the 128K model is
// deliberately out of scope for emulator integration this round (ver
// CompileFile's own doc comment).
//
// The temporary INI is left on disk, not cleaned up: since this
// function does not wait for EightyOne to exit, deleting it right
// away would race against EightyOne still reading it at startup - the
// same tradeoff the "don't wait" decision above already accepts
// elsewhere (a tiny, harmless leftover file in the OS temp directory).
func RunInEmulator(tapPath string) error {
	exePath, err := exec.LookPath("eightyone")
	if err != nil {
		return fmt.Errorf("eightyone not found in PATH: %w", err)
	}

	tmp, err := os.CreateTemp("", "badig-spectrum-*.ini")
	if err != nil {
		return fmt.Errorf("create temporary EightyOne profile: %w", err)
	}
	if _, err := tmp.WriteString(eightyOne48INI); err != nil {
		tmp.Close()
		return fmt.Errorf("write temporary EightyOne profile: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write temporary EightyOne profile: %w", err)
	}

	absTap, err := filepath.Abs(tapPath)
	if err != nil {
		return fmt.Errorf("resolve path %s: %w", tapPath, err)
	}
	cmd := exec.Command(exePath, tmp.Name(), absTap)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start eightyone: %w", err)
	}
	return nil
}

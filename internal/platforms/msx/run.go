package msx

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/platforms/msx/dsk"
)

// BuildTestDisk compiles dmxPath (ver CompileFile) and packages the
// tokenized result as a ready-to-boot MSX disk image, next to the
// source file under its own base name (<base>.dsk) - the "gerar um
// disco" half of the user's own request: compile, then make
// something a real MSX (or an emulator) can actually insert and run.
//
// The compiled program is stored on the disk as AUTOEXEC.BAS - the
// MSX disk ROM's own convention for a program that runs by itself as
// soon as the disk boots, no MSX-DOS and no typing required. Real
// MSX-DOS system files (MSXDOS.SYS/COMMAND.COM) are deliberately NOT
// bundled onto this disk: those are copyrighted third-party system
// software (confirmed real copies exist in the user's own
// E:\msxdiskutil, used there for the user's own local testing) - not
// something this project redistributes. dsk.Create's own default
// boot sector (ported from the same reference) still tries to
// chain-load MSXDOS.SYS first; without it present, real MSX
// firmware shows a brief "Boot error" prompt before falling through
// to BASIC and its own AUTOEXEC.BAS scan - confirmed by actually
// booting a disk this function built in openMSX (ver OUTLINE.md) -
// one keypress past that is the only practical cost, not a
// correctness bug.
func BuildTestDisk(dmxPath string) (diskPath string, err error) {
	_, bmxPath, err := CompileFile(dmxPath, nil)
	if err != nil {
		return "", err
	}

	diskPath = strings.TrimSuffix(dmxPath, filepath.Ext(dmxPath)) + ".dsk"
	img, err := dsk.Create(diskPath, nil)
	if err != nil {
		return "", fmt.Errorf("create disk %s: %w", diskPath, err)
	}
	if err := img.AddFile(bmxPath, "AUTOEXEC.BAS"); err != nil {
		img.Close()
		return "", fmt.Errorf("add %s to disk %s: %w", bmxPath, diskPath, err)
	}
	if err := img.Close(); err != nil {
		return "", fmt.Errorf("close disk %s: %w", diskPath, err)
	}
	return diskPath, nil
}

// RunInEmulator launches openMSX with diskPath inserted as drive A -
// mirrors the real engine's own emulator_interface.py (badig_msx.py:
// `cmd.extend(['-diska', self.stg.file_path])`), minus its
// stdio-command monitoring protocol (a materially bigger feature,
// already deferred in OUTLINE.md's own backlog - "Monitoramento de
// execução em tempo real"): this just opens the emulator window for
// the user to watch and interact with directly, which is exactly
// what was asked for ("para o usuario ja poder testar").
//
// openMSX is located via PATH (os/exec's own PATHEXT-aware lookup on
// Windows finds "openmsx.exe") - no machine/extension/settings
// customization yet (the options catalog already has
// EMULATOR_PATH/EMULATOR_MACHINE/etc. fields - ver internal/options -
// but nothing reads them into real behavior yet anywhere in the
// project; wiring that in is future work, not this round's).
// RunInEmulator does not wait for the emulator to exit.
func RunInEmulator(diskPath string) error {
	exePath, err := exec.LookPath("openmsx")
	if err != nil {
		return fmt.Errorf("openmsx not found in PATH: %w", err)
	}
	cmd := exec.Command(exePath, "-diska", diskPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start openmsx: %w", err)
	}
	return nil
}

// RunInFMSX launches fMSX with diskPath inserted as drive A - same
// flag shape as openMSX's own "-diska" (confirmed reading fMSX's own
// documentation, fMSX.html's "Command Line Options": "-diska
// <filename> - Set disk image used for drive A:"), and confirmed by
// actually launching fMSX this way on this machine. An alternative to
// openMSX requested by the user (2026-11-26), alongside blueMSX+
// below - ZEsarUX was considered too (it has an MSX1 mode) but
// deliberately left out: the user judged it not worth adding for MSX
// specifically, since it only covers MSX1, unlike openMSX/fMSX/
// blueMSX+ which all support MSX2/MSX2+.
func RunInFMSX(diskPath string) error {
	exePath, err := exec.LookPath("fmsx")
	if err != nil {
		return fmt.Errorf("fmsx not found in PATH: %w", err)
	}
	cmd := exec.Command(exePath, "-diska", diskPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start fmsx: %w", err)
	}
	return nil
}

// RunInBlueMSX launches blueMSX+ with diskPath as its one argument -
// blueMSX+'s own "smart" single-file-open form (confirmed reading its
// own embedded usage text: "blueMSX+ <file> - Open one media file...
// the form used for a double clicked file", and by actually launching
// it this way on this machine) - no "-diska"-style flag needed or
// accepted for this form. Specifically the "+" fork the user asked
// for (blueMSX+.exe), not the older plain blueMSX.exe also commonly
// installed alongside it.
func RunInBlueMSX(diskPath string) error {
	exePath, err := exec.LookPath("bluemsx+")
	if err != nil {
		return fmt.Errorf("bluemsx+ not found in PATH: %w", err)
	}
	cmd := exec.Command(exePath, diskPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start bluemsx+: %w", err)
	}
	return nil
}

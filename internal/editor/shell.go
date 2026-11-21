package edit

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func (e *Editor) runPowerShell() error {
	if e.virtual != nil {
		return errors.New("PowerShell is unavailable in a FILE SPACE session")
	}
	candidates := []string{"pwsh"}
	if runtime.GOOS == "windows" {
		candidates = append(candidates, "powershell.exe")
	}
	var shell string
	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate)
		if err == nil {
			shell = path
			break
		}
	}
	if shell == "" {
		return fmt.Errorf("PowerShell was not found (tried %s)", strings.Join(candidates, ", "))
	}

	var commandErr error
	if !e.app.Suspend(func() {
		command := exec.Command(shell)
		command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
		commandErr = command.Run()
	}) {
		return errors.New("unable to suspend the editor terminal for PowerShell")
	}
	if commandErr != nil {
		return fmt.Errorf("PowerShell failed: %w", commandErr)
	}
	return nil
}

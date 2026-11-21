//go:build !windows

package edit

import (
	"bytes"
	"fmt"
	"os/exec"
)

func sendToDefaultPrinter(stream []byte) error {
	for _, name := range []string{"lp", "lpr"} {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		command := exec.Command(path)
		command.Stdin = bytes.NewReader(stream)
		output, err := command.CombinedOutput()
		if err != nil {
			if len(output) > 0 {
				return fmt.Errorf("%s failed: %w: %s", name, err, bytes.TrimSpace(output))
			}
			return fmt.Errorf("%s failed: %w", name, err)
		}
		return nil
	}
	return fmt.Errorf("no default printer command found (lp or lpr)")
}

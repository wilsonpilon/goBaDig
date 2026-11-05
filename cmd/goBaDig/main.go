// Command goBaDig is the entrypoint shared by the CLI, TUI and GUI
// surfaces. See OUTLINE.md for the full architecture plan — this is
// still a stub: --help/--man and a minimal TUI/GUI with File > Exit
// and Help > About, nothing else implemented yet.
package main

import (
	"os"

	"github.com/wilsonpilon/goBaDig/internal/cli"
	"github.com/wilsonpilon/goBaDig/internal/gui"
	"github.com/wilsonpilon/goBaDig/internal/tui"
)

func main() {
	args := os.Args[1:]

	if handled, code := cli.Run(args); handled {
		os.Exit(code)
	}

	switch cli.Parse(args) {
	case cli.ModeTUI:
		if err := tui.Run(); err != nil {
			os.Exit(1)
		}
	case cli.ModeGUI:
		gui.Run()
	}
}

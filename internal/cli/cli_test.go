package cli

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want Mode
	}{
		{"no args", []string{}, ModeBanner},
		{"help long", []string{"--help"}, ModeHelp},
		{"help short", []string{"-h"}, ModeHelp},
		{"man", []string{"--man"}, ModeMan},
		{"tui", []string{"--tui"}, ModeTUI},
		{"gui", []string{"--gui"}, ModeGUI},
		{"cli", []string{"--cli"}, ModeConsole},
		{"unknown flag falls back to help", []string{"--nope"}, ModeHelp},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Parse(c.args); got != c.want {
				t.Fatalf("Parse(%v) = %v, want %v", c.args, got, c.want)
			}
		})
	}
}

func TestRunHandlesNonInteractiveModes(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"--man"}, {}} {
		handled, code := Run(args)
		if !handled {
			t.Fatalf("Run(%v): expected handled=true", args)
		}
		if code != 0 {
			t.Fatalf("Run(%v): expected exit code 0, got %d", args, code)
		}
	}
}

func TestRunLeavesInteractiveModesToCaller(t *testing.T) {
	for _, args := range [][]string{{"--tui"}, {"--gui"}, {"--cli"}} {
		handled, _ := Run(args)
		if handled {
			t.Fatalf("Run(%v): expected handled=false, caller must launch the mode", args)
		}
	}
}

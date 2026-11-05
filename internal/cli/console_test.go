package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunConsole(t *testing.T) {
	in := strings.NewReader("about\nversion\nnope\nexit\n")
	var out bytes.Buffer

	RunConsole(in, &out)

	got := out.String()
	for _, want := range []string{
		"Modo console",
		"goBaDig> ",
		"Basic Dignified Suite, em Go.", // from "about"
		"bloco NEW, feature outline",    // from "version"
		`comando desconhecido: "nope"`,  // unknown command
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("console output missing %q; got:\n%s", want, got)
		}
	}
}

func TestRunConsoleStopsOnEOF(t *testing.T) {
	// Empty input hits EOF on the first read; RunConsole must return
	// instead of blocking forever waiting for an "exit"/"quit" line.
	in := strings.NewReader("")
	var out bytes.Buffer

	RunConsole(in, &out)

	if !strings.Contains(out.String(), "goBaDig> ") {
		t.Fatalf("expected the prompt to be printed at least once; got:\n%s", out.String())
	}
}

package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/about"
	"github.com/wilsonpilon/goBaDig/internal/version"
)

// RunConsole implements the interactive console mode (--cli): a
// traditional prompt that waits for the user to type commands, one per
// line, until "exit"/"quit" or end-of-input (Ctrl-D/Ctrl-Z).
//
// This is a stub command set — real conversion commands arrive
// together with the core engine (see OUTLINE.md, Bloco 5).
func RunConsole(in io.Reader, out io.Writer) {
	fmt.Fprint(out, Banner())
	fmt.Fprintln(out, `Modo console. Digite "help" para ver os comandos, "exit" para sair.`)

	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, "goBaDig> ")
		if !scanner.Scan() {
			fmt.Fprintln(out)
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(out, "erro de leitura: %v\n", err)
			}
			return
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch line {
		case "exit", "quit":
			return
		case "help":
			fmt.Fprint(out, HelpText())
		case "man":
			fmt.Fprintln(out, about.Man())
		case "about":
			fmt.Fprintln(out, about.Short())
		case "version":
			fmt.Fprintf(out, "goBaDig v%s (bloco %s, feature %s)\n",
				version.Version, version.Block, version.Feature)
		default:
			fmt.Fprintf(out, "comando desconhecido: %q (digite \"help\")\n", line)
		}
	}
}

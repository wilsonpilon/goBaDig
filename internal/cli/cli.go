// Package cli implements the command-line surface: flag parsing, the
// --help/--man text, and dispatch into the interactive modes (TUI/GUI).
//
// Design note: the CLI, TUI and GUI are meant to expose, as far as
// possible, the same set of options (see OUTLINE.md, section "Paridade
// de opções entre interfaces"). This stub only has --help/--man/--tui/
// --gui; real conversion flags arrive together with the core engine.
package cli

import (
	"fmt"
	"os"

	"github.com/wilsonpilon/goBaDig/internal/about"
	"github.com/wilsonpilon/goBaDig/internal/version"
)

// Mode is what the CLI decided to do after looking at the arguments.
type Mode int

const (
	ModeBanner Mode = iota
	ModeHelp
	ModeMan
	ModeTUI
	ModeGUI
	ModeConsole
)

// Parse reads args (normally os.Args[1:]) and decides which mode to
// run. It never exits the process itself, so it stays easy to test.
func Parse(args []string) Mode {
	if len(args) == 0 {
		return ModeBanner
	}
	for _, a := range args {
		switch a {
		case "-h", "-help", "--help":
			return ModeHelp
		case "-man", "--man":
			return ModeMan
		case "-tui", "--tui":
			return ModeTUI
		case "-gui", "--gui":
			return ModeGUI
		case "-cli", "--cli":
			return ModeConsole
		}
	}
	// Unknown flag: stub behaviour is to show help rather than fail
	// silently, since there is no real option set to validate against yet.
	return ModeHelp
}

// HelpText is the --help output.
func HelpText() string {
	return fmt.Sprintf(`goBaDig v%s - Basic Dignified Suite preprocessor (MSX, ZX81, ZX Spectrum)

USO:
    goBaDig [opcao]

OPCOES (stub - o conjunto completo ainda esta em planejamento, ver OUTLINE.md):
    --help         mostra esta ajuda e sai
    --man          mostra a pagina de referencia (estilo "man"); nesta
                   versao de stub traz so o "about" do projeto
    --tui          abre a interface de terminal (TUI)
    --gui          abre a interface grafica (GUI)
    --cli          abre o modo console: um prompt interativo que fica
                   aguardando comandos, um por linha, ate "exit"/"quit"

Sem opcoes, mostra um banner curto (nao entra no modo console).

As tres interfaces (CLI, TUI, GUI) expoem, na medida do possivel, o
mesmo conjunto de opcoes - ver OUTLINE.md, secao "Paridade de opcoes
entre interfaces".
`, version.Version)
}

// Banner is the short message printed when goBaDig runs with no flags.
func Banner() string {
	return fmt.Sprintf("goBaDig v%s (bloco %s, feature %s) - use --help para as opcoes.\n",
		version.Version, version.Block, version.Feature)
}

// Run handles the non-interactive modes (help/man/banner), printing to
// stdout and reporting the exit code. It returns handled=false when the
// caller (main) still needs to launch an interactive mode (TUI/GUI).
func Run(args []string) (handled bool, exitCode int) {
	switch Parse(args) {
	case ModeHelp:
		fmt.Fprint(os.Stdout, HelpText())
		return true, 0
	case ModeMan:
		fmt.Fprintln(os.Stdout, about.Man())
		return true, 0
	case ModeBanner:
		fmt.Fprint(os.Stdout, Banner())
		return true, 0
	default:
		return false, 0
	}
}

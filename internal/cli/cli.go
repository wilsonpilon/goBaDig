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
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/about"
	"github.com/wilsonpilon/goBaDig/internal/platforms/msx"
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
    goBaDig <arquivo.dmx>

OPCOES (stub - o conjunto completo ainda esta em planejamento, ver OUTLINE.md):
    --help         mostra esta ajuda e sai
    --man          mostra a pagina de referencia (estilo "man"); nesta
                   versao de stub traz so o "about" do projeto
    --tui          abre a interface de terminal (TUI)
    --gui          abre a interface grafica (GUI)
    --cli          abre o modo console: um prompt interativo que fica
                   aguardando comandos, como "theme dracula",
                   "set INITIAL_LINE 10", "settings", "compile
                   programa.dmx" e "help"

Passar um arquivo MSX Dignified (<arquivo.dmx>) direto, sem nenhuma
opcao, converte ele: gera <arquivo.amx> (listagem Classic numerada) e
<arquivo.bmx> (binario tokenizado), no mesmo diretorio do arquivo de
entrada. Ainda so MSX (ZX81/ZX Spectrum nao tem motor de conversao
proprio implementado).

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

// Run handles the non-interactive modes (help/man/banner/compile),
// printing to stdout and reporting the exit code. It returns
// handled=false when the caller (main) still needs to launch an
// interactive mode (TUI/GUI).
func Run(args []string) (handled bool, exitCode int) {
	switch Parse(args) {
	case ModeHelp:
		// A bare positional file ("goBaDig programa.dmx") falls into
		// this same bucket - Parse has no flag to match it against,
		// mirroring the original engine's own CLI convention
		// (`badig.py <input> [output]`, no verb needed) instead of
		// requiring a "compile" subcommand. An explicit mode flag
		// (--tui, etc.) always takes precedence: Parse already
		// returns that other Mode first, so this is never reached
		// when one is present.
		if path, ok := compileFileArg(args); ok {
			return true, runCompile(path)
		}
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

// compileFileArg reports the single positional (non-flag) argument in
// args, if there is exactly one - the candidate source file for
// runCompile. Zero or more than one such argument means there is no
// unambiguous single file to compile, so the caller falls back to
// HelpText instead of guessing.
func compileFileArg(args []string) (string, bool) {
	var found string
	count := 0
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		found = a
		count++
	}
	if count != 1 {
		return "", false
	}
	return found, true
}

// runCompile implements `goBaDig <arquivo.dmx>`: the first real
// "do something to a user's file" command in the project (every
// piece it calls - msx.CompileFile, which wraps core.Run/Tokenize -
// already existed and was validated on its own; this is just the
// CLI entry point to it). MSX is hard-coded because it is the only
// dialect with a real Description/Lexer/Parser today - selecting a
// dialect by file extension or a `-id` flag (ver OUTLINE.md, "9
// perguntas") is deferred until ZX81/ZX Spectrum exist to choose
// between.
func runCompile(path string) int {
	amxPath, bmxPath, err := msx.CompileFile(path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "goBaDig: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stdout, "Convertido: %s\n  -> %s\n  -> %s\n", path, amxPath, bmxPath)
	return 0
}

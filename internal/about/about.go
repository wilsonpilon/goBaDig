// Package about holds the project description text shared by every
// interface (CLI --man, TUI Help > About, GUI Help > About), so the
// three surfaces never drift apart. See OUTLINE.md, section "Paridade
// de opções entre interfaces".
package about

import (
	"fmt"

	"github.com/wilsonpilon/goBaDig/internal/version"
)

// Short is a one-paragraph description, used by the TUI/GUI About
// dialogs and as the header of the CLI "man" output.
func Short() string {
	return fmt.Sprintf(
		"goBaDig v%s (bloco %s, feature %s)\n\n"+
			"Basic Dignified Suite, em Go.\n"+
			"Pre-processador de BASIC para MSX, ZX81 e ZX Spectrum,\n"+
			"com CLI, TUI e GUI, reescrito a partir da Basic Dignified\n"+
			"Suite original (Python, por Farique).\n\n"+
			"Licenca: GNU GPL v3. Sem garantias.",
		version.Version, version.Block, version.Feature,
	)
}

// Man is the CLI "man"-style reference. This stub only prints the
// project's "about" section (NAME/DESCRIPTION) — the full reference
// (BaDig commands/options and the embedded MSX/ZX81/ZX Spectrum BASIC
// manuals) is planned but not implemented yet. See OUTLINE.md, bloco
// "Sistema de ajuda e manuais".
func Man() string {
	return "GOBADIG(1)" + "\n\n" +
		"NAME\n" +
		"    goBaDig - Basic Dignified Suite preprocessor (MSX, ZX81, ZX Spectrum)\n\n" +
		"DESCRIPTION\n" +
		"    " + indentedShort() + "\n\n" +
		"STATUS\n" +
		"    Este e um programa STUB. Ainda nao ha conversor de linguagem,\n" +
		"    tokenizer ou integracao com emulador implementados. Ver\n" +
		"    OUTLINE.md no repositorio para o plano completo.\n\n" +
		"    O \"man\" completo (comandos, opcoes do BaDig e os manuais de\n" +
		"    BASIC do MSX, ZX81 e ZX Spectrum para consulta de sintaxe)\n" +
		"    ainda nao esta implementado nesta versao.\n\n" +
		"SEE ALSO\n" +
		"    --help        lista de opcoes de linha de comando\n" +
		"    --tui         abre a interface de terminal (TUI)\n" +
		"    --gui         abre a interface grafica (GUI)\n"
}

func indentedShort() string {
	return "goBaDig e um pre-processador de BASIC para MSX, ZX81 e ZX Spectrum,\n" +
		"    com CLI, TUI e GUI, reescrito em Go a partir da Basic Dignified\n" +
		"    Suite original (Python, por Farique)."
}

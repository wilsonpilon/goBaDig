// Package about holds the project description text shared by every
// interface (CLI --man, TUI Help > About, GUI Help > About), so the
// three surfaces never drift apart. See OUTLINE.md, section "Paridade
// de opções entre interfaces".
package about

import (
	"fmt"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/man"
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

// Man is the CLI "man"-style reference index, delegating to
// internal/man (the shared Page model behind "--man"/"man <tópico>"
// and WibEdit's own Manual viewer) rather than keeping its own text.
// See OUTLINE.md, bloco "Sistema de ajuda e manuais".
func Man() string {
	return "GOBADIG(1)\n\nNAME\n" +
		"    goBaDig - Basic Dignified Suite preprocessor (MSX, ZX81, ZX Spectrum)\n\n" +
		"DESCRIPTION\n    " + indentedShort() + "\n\n" +
		"TÓPICOS\n" + indent(man.Index().PlainLines()) +
		"\nUse \"man <tópico>\" (console) ou \"--man <tópico>\" (linha de comando)\n" +
		"para abrir um deles.\n"
}

func indent(lines []string) string {
	var b strings.Builder
	for _, line := range lines {
		b.WriteString("    ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func indentedShort() string {
	return "goBaDig e um pre-processador de BASIC para MSX, ZX81 e ZX Spectrum,\n" +
		"    com CLI, TUI e GUI, reescrito em Go a partir da Basic Dignified\n" +
		"    Suite original (Python, por Farique)."
}

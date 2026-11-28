package man

import (
	"fmt"
	"sort"
	"strings"
)

// AsmBiosKind distinguishes a BIOS call (a ROM routine, reached with
// CALL/RST) from a system variable (a RAM location read/written
// directly) - both are "predefined symbols" in asMSX terms (the
// ".BIOS"/".BIOSVARS" directives, see docs/man/asmsx.md section 2.5),
// but the right verb in a human-readable page differs.
type AsmBiosKind int

const (
	AsmBiosRoutine AsmBiosKind = iota
	AsmBiosVariable
)

func (k AsmBiosKind) String() string {
	if k == AsmBiosVariable {
		return "variável de sistema"
	}
	return "rotina da BIOS"
}

// AsmBiosSymbol is one predefined name asMSX resolves on its own, with
// no .EQU needed, when a ".BIOS"/".BIOSVARS" source uses it - extracted
// once from resource/asMSX/src/dura.y's msx_bios()/msx_bios_vars()
// (asmbios_data.go, generated, see that file's own header), not
// hand-written: that C source is asMSX's own authoritative list, the
// same one used to decide whether "CHPUT" assembles to an address or
// an undefined-symbol error.
type AsmBiosSymbol struct {
	Name    string
	Address int
	Kind    AsmBiosKind
}

// Page renders s as a short reference page - just enough to answer
// "what does this name resolve to", pointing to "man asmsx" (section
// 2.5, the ".BIOS"/".BIOSVARS" directives) for the full story. Unlike
// MsxDictEntry.Page (dict.go), there is no prose to show per symbol -
// asMSX's own manual documents the directives, not each individual
// name - so this is deliberately a one-liner, not a multi-section page.
func (s AsmBiosSymbol) Page() Page {
	page := Page{Title: s.Name}
	page.Lines = append(page.Lines, Line{Text: s.Name, Style: StyleHeading1})
	page.Lines = append(page.Lines, Line{
		Text:  fmt.Sprintf("Endereço: %04Xh (%s, predefinido pelo asMSX)", s.Address, s.Kind),
		Style: StyleNormal,
	})
	directive := ".BIOS"
	if s.Kind == AsmBiosVariable {
		directive = ".BIOSVARS"
	}
	page.Lines = append(page.Lines, Line{
		Text:  fmt.Sprintf("Disponível quando o .asm usa a diretiva %s - veja \"man asmsx\".", directive),
		Style: StyleNormal,
	})
	return page
}

// asmBiosByName indexes asmBiosRoutines/asmBiosVariables (asmbios_data.go)
// by uppercase name, built once - asMSX's own names are case-sensitive
// in source (always upper case by convention, per docs/man/asmsx.md),
// but a lookup here (man/Shift+F1/highlighting) is deliberately
// case-insensitive, matching every other lookup in this package.
var asmBiosByName = buildAsmBiosByName()

func buildAsmBiosByName() map[string]AsmBiosSymbol {
	index := make(map[string]AsmBiosSymbol, len(asmBiosRoutines)+len(asmBiosVariables))
	for _, s := range asmBiosRoutines {
		index[s.Name] = s
	}
	for _, s := range asmBiosVariables {
		index[s.Name] = s
	}
	return index
}

// FindAsmBiosSymbol looks up one of asMSX's predefined BIOS/system-area
// names (case-insensitive) - used by "man <nome>" (registry.go's
// Lookup, tried last, after the goBaDig command topics and the MSX
// BASIC dictionary, since a handful of these names - BEEP, CLS - also
// happen to be MSX BASIC commands and that reading should win), by
// WibEdit's Shift+F1 contextual help on a .asm file (view_man.go), and
// by the editor's own syntax highlighter to recognize these names as a
// distinct color (internal/editor's asmBiosNames, kept in sync via
// AsmBiosSymbolNames below rather than its own copy of this data).
func FindAsmBiosSymbol(name string) (AsmBiosSymbol, bool) {
	s, ok := asmBiosByName[strings.ToUpper(strings.TrimSpace(name))]
	return s, ok
}

// AsmBiosSymbolNames returns every predefined name's upper-case form,
// for the editor's syntax highlighter (internal/editor) to recognize
// without duplicating this package's data.
func AsmBiosSymbolNames() map[string]bool {
	names := make(map[string]bool, len(asmBiosByName))
	for name := range asmBiosByName {
		names[name] = true
	}
	return names
}

// asmBiosTopic is the reserved pseudo-topic ("man asmbios") that lists
// every predefined name, grouped by kind - mirrors msxDictTopic
// ("man msx") in registry.go.
const asmBiosTopic = "asmbios"

// AsmBiosIndex lists every predefined BIOS routine, then every system
// variable, each sorted by name and linked via TOC so Enter on a line
// opens that symbol's own Page (view_man.go's activate already follows
// any TOC entry through man.Lookup - no separate wiring needed there).
func AsmBiosIndex() Page {
	page := Page{Title: "asMSX - endereços predefinidos (.BIOS / .BIOSVARS)"}
	page.Lines = append(page.Lines, Line{Text: page.Title, Style: StyleHeading1})

	group := func(heading string, symbols []AsmBiosSymbol) {
		page.Lines = append(page.Lines, Line{Text: heading, Style: StyleHeading2})
		names := make([]string, 0, len(symbols))
		for _, s := range symbols {
			names = append(names, s.Name)
		}
		sort.Strings(names)
		for _, name := range names {
			page.TOC = append(page.TOC, TOCEntry{Title: name, Line: len(page.Lines)})
			s := asmBiosByName[name]
			page.Lines = append(page.Lines, Line{
				Text:  fmt.Sprintf("- %-7s %04Xh", s.Name, s.Address),
				Style: StyleBullet,
			})
		}
	}
	group(fmt.Sprintf("Rotinas da BIOS (%d, diretiva .BIOS)", len(asmBiosRoutines)), asmBiosRoutines)
	group(fmt.Sprintf("Variáveis de sistema (%d, diretiva .BIOSVARS)", len(asmBiosVariables)), asmBiosVariables)
	return page
}

package man

import "testing"

// A smoke count, not an exact pin to 124/144: the point is catching a
// truncated/empty generated table, not re-asserting dura.y's own
// numbers here (see asmbios_data.go's header for how to regenerate).
func TestAsmBiosDataIsPlausible(t *testing.T) {
	if len(asmBiosRoutines) < 100 {
		t.Fatalf("asmBiosRoutines has %d entries, want at least 100", len(asmBiosRoutines))
	}
	if len(asmBiosVariables) < 100 {
		t.Fatalf("asmBiosVariables has %d entries, want at least 100", len(asmBiosVariables))
	}
}

func TestFindAsmBiosSymbolIsCaseInsensitive(t *testing.T) {
	for _, name := range []string{"CHPUT", "chput", "ChPut"} {
		s, ok := FindAsmBiosSymbol(name)
		if !ok {
			t.Fatalf("FindAsmBiosSymbol(%q): not found", name)
		}
		if s.Address != 0x00a2 || s.Kind != AsmBiosRoutine {
			t.Fatalf("FindAsmBiosSymbol(%q) = %+v, want address 0x00a2, kind AsmBiosRoutine", name, s)
		}
	}
}

func TestFindAsmBiosSymbolFindsASystemVariable(t *testing.T) {
	s, ok := FindAsmBiosSymbol("HIMEM")
	if !ok || s.Address != 0xfc4a || s.Kind != AsmBiosVariable {
		t.Fatalf("FindAsmBiosSymbol(HIMEM) = %+v, %v, want address 0xfc4a, kind AsmBiosVariable", s, ok)
	}
}

func TestFindAsmBiosSymbolNotFound(t *testing.T) {
	if _, ok := FindAsmBiosSymbol("NOTASYMBOL"); ok {
		t.Fatalf("FindAsmBiosSymbol(NOTASYMBOL): found, want not found")
	}
}

func TestAsmBiosSymbolNamesIncludesBothGroups(t *testing.T) {
	names := AsmBiosSymbolNames()
	for _, want := range []string{"CHPUT", "HIMEM"} {
		if !names[want] {
			t.Fatalf("AsmBiosSymbolNames() missing %q", want)
		}
	}
}

func TestLookupPrefersMsxBasicCommandOverABiosRoutineOfTheSameName(t *testing.T) {
	// BEEP and CLS are both real MSX BASIC commands and real asMSX
	// BIOS routine names - "man beep" in this BASIC-first project
	// should resolve to the BASIC command (the MSX dictionary entry),
	// not the BIOS address, so FindAsmBiosSymbol is only ever tried as
	// Lookup's last resort (registry.go).
	page, err := Lookup("beep")
	if err != nil {
		t.Fatalf("Lookup(beep): %v", err)
	}
	if page.Title != "BEEP" {
		t.Fatalf("Lookup(beep).Title = %q, want the MSX dictionary's own title", page.Title)
	}
}

func TestLookupFindsABiosRoutineWithNoBasicCommandOfTheSameName(t *testing.T) {
	page, err := Lookup("CALSLT")
	if err != nil {
		t.Fatalf("Lookup(CALSLT): %v", err)
	}
	if page.Title != "CALSLT" {
		t.Fatalf("Lookup(CALSLT).Title = %q, want CALSLT", page.Title)
	}
}

func TestLookupAsmBiosPseudoTopic(t *testing.T) {
	page, err := Lookup("asmbios")
	if err != nil {
		t.Fatalf("Lookup(asmbios): %v", err)
	}
	if len(page.TOC) != len(asmBiosRoutines)+len(asmBiosVariables) {
		t.Fatalf("AsmBiosIndex TOC has %d entries, want %d", len(page.TOC), len(asmBiosRoutines)+len(asmBiosVariables))
	}
}

func TestIndexListsTheAsmBiosPseudoTopic(t *testing.T) {
	page := Index()
	for _, entry := range page.TOC {
		if entry.Title == asmBiosTopic {
			return
		}
	}
	t.Fatalf("Index() TOC is missing the %q entry", asmBiosTopic)
}

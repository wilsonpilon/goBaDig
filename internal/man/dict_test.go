package man

import (
	"strings"
	"testing"
)

func TestMsxDictHasAPlausibleEntryCount(t *testing.T) {
	// Sanity check against a generated data file silently ending up
	// empty or truncated: 140 MSX1 verbetes (the book, A-W, no J/Q) +
	// 41 MSX2+ verbetes (ACVS manual) = 181.
	if len(msxDict) != 181 {
		t.Fatalf("len(msxDict) = %d, want 181", len(msxDict))
	}
	msx1, msx2plus := 0, 0
	for _, e := range msxDict {
		switch e.System {
		case "MSX1":
			msx1++
		case "MSX2+":
			msx2plus++
		default:
			t.Errorf("entry %q has unexpected System %q", e.Title, e.System)
		}
		if e.Title == "" {
			t.Error("entry with empty Title")
		}
	}
	if msx1 != 140 {
		t.Errorf("MSX1 entries = %d, want 140", msx1)
	}
	if msx2plus != 41 {
		t.Errorf("MSX2+ entries = %d, want 41", msx2plus)
	}
}

func TestFindMsxDictEntryExactMatchIsCaseInsensitive(t *testing.T) {
	for _, keyword := range []string{"ABS", "abs", "Abs"} {
		entry, ok := FindMsxDictEntry(keyword)
		if !ok {
			t.Fatalf("FindMsxDictEntry(%q): not found", keyword)
		}
		if entry.Title != "ABS" {
			t.Fatalf("FindMsxDictEntry(%q).Title = %q, want ABS", keyword, entry.Title)
		}
	}
}

func TestFindMsxDictEntryResolvesACompoundTitle(t *testing.T) {
	entry, ok := FindMsxDictEntry("RETURN")
	if !ok {
		t.Fatal(`FindMsxDictEntry("RETURN"): not found`)
	}
	if entry.Title != "GOSUB-RETURN" {
		t.Fatalf("FindMsxDictEntry(RETURN).Title = %q, want GOSUB-RETURN", entry.Title)
	}
}

func TestFindMsxDictEntryUnknownKeyword(t *testing.T) {
	if _, ok := FindMsxDictEntry("NAOEXISTE"); ok {
		t.Fatal("FindMsxDictEntry(NAOEXISTE) should not be found")
	}
	if _, ok := FindMsxDictEntry(""); ok {
		t.Fatal("FindMsxDictEntry(\"\") should not be found")
	}
}

func TestMsxDictEntryPageHasTheBookSectionOrder(t *testing.T) {
	entry, ok := FindMsxDictEntry("ABS")
	if !ok {
		t.Fatal("ABS not found")
	}
	page := entry.Page()
	var headings []string
	for _, line := range page.Lines {
		if line.Style == StyleHeading1 {
			headings = append(headings, line.Text)
		}
	}
	want := []string{"NOME", "FORMATO", "EXEMPLO", "FUNÇÃO", "PROGRAMA EXEMPLO", "REFERÊNCIA"}
	if len(headings) != len(want) {
		t.Fatalf("headings = %v, want %v", headings, want)
	}
	for i := range want {
		if headings[i] != want[i] {
			t.Fatalf("headings[%d] = %q, want %q", i, headings[i], want[i])
		}
	}
}

func TestMsxDictIndexGroupsByFirstLetter(t *testing.T) {
	index := MsxDictIndex()
	if len(index.TOC) != len(msxDict) {
		t.Fatalf("MsxDictIndex TOC has %d entries, want %d", len(index.TOC), len(msxDict))
	}
	var headings []string
	for _, line := range index.Lines {
		if line.Style == StyleHeading2 {
			headings = append(headings, line.Text)
		}
	}
	if len(headings) == 0 {
		t.Fatal("MsxDictIndex has no letter headings")
	}
	if headings[0] != "A" {
		t.Fatalf("first letter heading = %q, want A", headings[0])
	}
}

func TestLookupPrefersTheMsxDictionaryOverMarkdownTopics(t *testing.T) {
	page, err := Lookup("abs")
	if err != nil {
		t.Fatalf("Lookup(abs) error: %v", err)
	}
	if !strings.Contains(page.Title, "ABS") {
		t.Fatalf("Lookup(abs).Title = %q, want it to mention ABS", page.Title)
	}
}

func TestLookupMsxPseudoTopicOpensTheDictionaryIndex(t *testing.T) {
	page, err := Lookup("msx")
	if err != nil {
		t.Fatalf("Lookup(msx) error: %v", err)
	}
	if len(page.TOC) != len(msxDict) {
		t.Fatalf("Lookup(msx) TOC has %d entries, want %d", len(page.TOC), len(msxDict))
	}
}

func TestIndexLinksToTheMsxDictionary(t *testing.T) {
	index := Index()
	found := false
	for _, entry := range index.TOC {
		if entry.Title == msxDictTopic {
			found = true
		}
	}
	if !found {
		t.Fatal(`Index() is missing a TOC link to the "msx" pseudo-topic`)
	}
}

package edit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestAutocompletePopupOpensWithMatchingCandidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.dmx")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	if err := editor.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	drawEditor(editor, screen)

	for _, r := range "PRI" {
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
	drawEditor(editor, screen)

	if editor.ac == nil {
		t.Fatal("typing a known MSX keyword prefix did not open the autocomplete popup")
	}
	found := false
	for _, c := range editor.ac.candidates {
		if c == "PRINT" {
			found = true
		}
	}
	if !found {
		t.Fatalf("candidates = %v, want PRINT among them", editor.ac.candidates)
	}
}

func TestAutocompleteOffersAWordFromAMsxDictCompoundTitle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.dmx")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	if err := editor.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	drawEditor(editor, screen)

	// "RETURN" only exists in the MSX dictionary as half of the
	// compound title "GOSUB-RETURN" - this only works once
	// candidatesFor pulls from man.AutocompleteWords(), not just the
	// plain msxKeywords/msxFunctions tables.
	for _, r := range "RET" {
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
	drawEditor(editor, screen)

	if editor.ac == nil {
		t.Fatal(`typing "RET" did not open the autocomplete popup`)
	}
	found := false
	for _, c := range editor.ac.candidates {
		if c == "RETURN" {
			found = true
		}
	}
	if !found {
		t.Fatalf("candidates = %v, want RETURN among them", editor.ac.candidates)
	}
}

func TestAutocompleteTabConfirmsAndReplacesPrefix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.dmx")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	if err := editor.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	drawEditor(editor, screen)

	for _, r := range "PRI" {
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
	drawEditor(editor, screen)
	if editor.ac == nil {
		t.Fatal("popup did not open before Tab")
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	drawEditor(editor, screen)

	if editor.ac != nil {
		t.Fatal("Tab did not close the popup")
	}
	if got := editor.area.GetText(); got != "PRINT" {
		t.Fatalf("document text = %q, want %q", got, "PRINT")
	}
}

func TestAutocompleteEscapeClosesWithoutInserting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.dmx")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	if err := editor.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	drawEditor(editor, screen)

	for _, r := range "PRI" {
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
	drawEditor(editor, screen)
	if editor.ac == nil {
		t.Fatal("popup did not open before Escape")
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	drawEditor(editor, screen)

	if editor.ac != nil {
		t.Fatal("Escape did not close the popup")
	}
	if got := editor.area.GetText(); got != "PRI" {
		t.Fatalf("Escape changed the document text: got %q, want %q", got, "PRI")
	}
}

func TestAutocompleteClosesOnSpace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.dmx")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	if err := editor.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	drawEditor(editor, screen)

	for _, r := range "PRI" {
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
	drawEditor(editor, screen)
	if editor.ac == nil {
		t.Fatal("popup did not open before space")
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	drawEditor(editor, screen)

	if editor.ac != nil {
		t.Fatal("typing a space (leaving the word) did not close the popup")
	}
	if got := editor.area.GetText(); got != "PRI " {
		t.Fatalf("document text = %q, want %q", got, "PRI ")
	}
}

func TestAutocompleteArrowsNavigateCandidates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.dmx")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	if err := editor.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	drawEditor(editor, screen)

	// "CL" matches several MSX keywords (CLEAR, CLOAD, CLOSE, CLS).
	for _, r := range "CL" {
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
	drawEditor(editor, screen)
	if editor.ac == nil || len(editor.ac.candidates) < 2 {
		t.Fatalf("expected multiple candidates for \"CL\", got %v", editor.ac)
	}
	if editor.ac.selected != 0 {
		t.Fatalf("selected = %d, want 0 right after opening", editor.ac.selected)
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	if editor.ac.selected != 1 {
		t.Fatalf("Down did not advance selection: selected = %d, want 1", editor.ac.selected)
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	if editor.ac.selected != 0 {
		t.Fatalf("Up did not step back: selected = %d, want 0", editor.ac.selected)
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	if editor.ac.selected != len(editor.ac.candidates)-1 {
		t.Fatalf("Up did not wrap around: selected = %d, want %d", editor.ac.selected, len(editor.ac.candidates)-1)
	}
}

func TestAutocompleteWorksForZX81AndZXSpectrumToo(t *testing.T) {
	cases := []struct {
		ext  string
		want string
	}{
		{"dzx", "PRINT"},
		{"dzs", "PRINT"},
	}
	for _, c := range cases {
		dir := t.TempDir()
		path := filepath.Join(dir, "game."+c.ext)
		if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		editor, app, screen := newTestEditor(t, 80, 25, nil)
		if err := editor.OpenFile(path); err != nil {
			t.Fatal(err)
		}
		drawEditor(editor, screen)

		for _, r := range "PRI" {
			sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
		}
		drawEditor(editor, screen)
		if editor.ac == nil {
			t.Fatalf("%s: popup did not open for a known keyword prefix", c.ext)
		}
		found := false
		for _, cand := range editor.ac.candidates {
			if cand == c.want {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s: candidates = %v, want %q among them", c.ext, editor.ac.candidates, c.want)
		}
	}
}

func TestAutocompleteNeverOpensWithoutARecognizedDialect(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	// A brand-new "Untitled" document has no path, so dialectForPath
	// gives dialectNone - nothing should ever pop up, even typing a
	// string that happens to be a real MSX keyword.
	for _, r := range "PRINT" {
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
	drawEditor(editor, screen)
	if editor.ac != nil {
		t.Fatal("autocomplete opened for a document without a recognized dialect")
	}
}

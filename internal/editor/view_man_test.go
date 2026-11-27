package edit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestManDisplayOpensTheIndexWithATOC(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	command, ok := commandByID("manual.open")
	if !ok || !editor.available(command) {
		t.Fatal("manual.open command is unavailable")
	}
	editor.execute(command)
	if editor.manView == nil {
		t.Fatal("Manual did not open")
	}
	if len(editor.manView.page().TOC) == 0 {
		t.Fatal("Manual index has no topics")
	}
	foundCompile := false
	for _, entry := range editor.manView.page().TOC {
		if entry.Title == "compile" {
			foundCompile = true
		}
	}
	if !foundCompile {
		t.Fatal(`Manual index is missing the "compile" topic`)
	}
}

func TestManViewEnterOpensALinkedTopicAndEscGoesBack(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	command, _ := commandByID("manual.open")
	editor.execute(command)

	view := editor.manView
	// Move the selection to the first TOC entry and activate it.
	view.selected = view.page().TOC[0].Line
	wantTopic := view.page().TOC[0].Title

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if editor.manView == nil {
		t.Fatal("Manual closed instead of navigating into the topic")
	}
	if len(editor.manView.history) != 2 {
		t.Fatalf("history length = %d, want 2", len(editor.manView.history))
	}
	if got := editor.manView.page().Title; got != wantTopic {
		t.Fatalf("opened topic = %q, want %q", got, wantTopic)
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if editor.manView == nil {
		t.Fatal("Esc closed the Manual instead of going back one level")
	}
	if len(editor.manView.history) != 1 {
		t.Fatalf("history length after Esc = %d, want 1", len(editor.manView.history))
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if editor.manView != nil {
		t.Fatal("Esc at the root should close the Manual view")
	}
	if editor.app.GetFocus() != editor.area {
		t.Fatal("closing the Manual did not restore focus to the document")
	}
}

func TestManViewUpDownNavigateSelection(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	command, _ := commandByID("manual.open")
	editor.execute(command)
	view := editor.manView
	start := view.selected

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	if view.selected != start+1 {
		t.Fatalf("selected after Down = %d, want %d", view.selected, start+1)
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	if view.selected != start {
		t.Fatalf("selected after Up = %d, want %d", view.selected, start)
	}
}

func TestShiftF1OpensTheDictionaryEntryForTheWordUnderTheCursor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.dmx")
	if err := os.WriteFile(path, []byte("PRINT 1"), 0o644); err != nil {
		t.Fatal(err)
	}
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	if err := editor.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	editor.area.document.cursor = documentPosition{line: 0, character: 3}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModShift))
	if editor.manView == nil {
		t.Fatal("Shift+F1 did not open the Manual viewer")
	}
	if got := editor.manView.page().Title; got != "PRINT" {
		t.Fatalf("opened page title = %q, want PRINT", got)
	}
}

func TestShiftF1NormalizesTheQuestionMarkShorthandToPrint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.dmx")
	if err := os.WriteFile(path, []byte("? 1"), 0o644); err != nil {
		t.Fatal(err)
	}
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	if err := editor.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	editor.area.document.cursor = documentPosition{line: 0, character: 0}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF13, 0, tcell.ModNone))
	if editor.manView == nil {
		t.Fatal("Shift+F1 (KeyF13) did not open the Manual viewer")
	}
	if got := editor.manView.page().Title; got != "PRINT" {
		t.Fatalf("opened page title = %q, want PRINT", got)
	}
}

func TestShiftF1ReportsAnUnknownWordInsteadOfOpeningAnything(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.dmx")
	if err := os.WriteFile(path, []byte("NAOEXISTE"), 0o644); err != nil {
		t.Fatal(err)
	}
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	if err := editor.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	editor.area.document.cursor = documentPosition{line: 0, character: 3}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModShift))
	if editor.manView != nil {
		t.Fatal("Shift+F1 opened the Manual viewer for an unknown word")
	}
	if !strings.Contains(editor.notice, "NAOEXISTE") {
		t.Fatalf("notice = %q, want it to mention the unknown word", editor.notice)
	}
}

func TestShiftF1FallsBackToTheGeneralIndexOutsideMsx(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.dzx")
	if err := os.WriteFile(path, []byte("PRINT 1"), 0o644); err != nil {
		t.Fatal(err)
	}
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	if err := editor.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	editor.area.document.cursor = documentPosition{line: 0, character: 3}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModShift))
	if editor.manView == nil {
		t.Fatal("Shift+F1 outside MSX should fall back to the general index")
	}
	if len(editor.manView.page().TOC) == 0 {
		t.Fatal("fallback page has no TOC - doesn't look like the general index")
	}
}

func TestPlainF1StillOpensTheHelpMenuUnaffectedByShiftF1(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone))
	if !editor.menuOpen || editor.currentMenu != 'H' {
		t.Fatal("plain F1 should still open the Help menu")
	}
}

func TestManViewIsModalAndDoesNotReachTheDocument(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("original", true)
	command, _ := commandByID("manual.open")
	editor.execute(command)

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	if editor.area.GetText() != "original" {
		t.Fatal("a key typed while the Manual is open reached the document")
	}
	if editor.manView == nil {
		t.Fatal("an unrelated key should not close the Manual view")
	}
}

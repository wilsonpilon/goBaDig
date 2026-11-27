package edit

import (
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

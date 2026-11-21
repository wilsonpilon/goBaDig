package edit

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestEditorScreenAndMenu(t *testing.T) {
	editor, app, screen := newTestEditor(t, 100, 25, nil)
	if row := editorScreenRow(screen, 0); !strings.Contains(row, "Text-layout") || !strings.Contains(row, "F1=Help") {
		t.Fatalf("menu bar is incomplete: %q", row)
	}
	if row := editorScreenRow(screen, 24); !strings.Contains(row, "Line=1") || !strings.Contains(row, "Col=1") ||
		!strings.Contains(row, "Untitled") || !strings.Contains(row, "Insert") || !strings.Contains(row, "WW=Off") {
		t.Fatalf("status bar is incomplete: %q", row)
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModAlt))
	drawEditor(editor, screen)
	name, primitive := editor.GetFrontPage()
	dropdown, ok := primitive.(*editorDropdown)
	if name != "dropdown" || !ok || dropdown.GetItemCount() != 10 {
		t.Fatal("Alt+T did not open the Text-layout menu")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))
	drawEditor(editor, screen)
	if row := editorScreenRow(screen, 24); !strings.Contains(row, "WW=On") {
		t.Errorf("Word wrap selection did not update the status line: %q", row)
	}
}

func TestEditorTracksCursorAndConfirmsDiscard(t *testing.T) {
	closed := false
	editor, app, screen := newTestEditor(t, 80, 25, func() { closed = true })
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	drawEditor(editor, screen)
	if row := editorScreenRow(screen, 24); !strings.Contains(row, "Col=2") || !strings.Contains(row, "Untitled*") {
		t.Fatalf("editing did not update cursor and modified state: %q", row)
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if closed {
		t.Fatal("editor closed without confirming unsaved text")
	}
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Unsaved document" {
		t.Fatal("unsaved exit did not open the Save/Don't Save/Cancel dialog")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	if editor.confirmExit {
		t.Fatal("Cancel did not dismiss the unsaved prompt")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	if !closed {
		t.Fatal("Don't Save did not discard and close the editor")
	}
}

func newTestEditor(t *testing.T, width, height int, onClose func()) (*Editor, *tview.Application, tcell.SimulationScreen) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)
	screen.SetSize(width, height)
	app := tview.NewApplication().SetScreen(screen)
	editor := NewEditor(app, onClose)
	editor.printer = func([]byte) error { return nil }
	app.SetInputCapture(editor.Capture)
	app.SetRoot(editor, true)
	drawEditor(editor, screen)
	return editor, app, screen
}

func sendEditorKey(editor *Editor, app *tview.Application, event *tcell.EventKey) {
	if event = editor.Capture(event); event != nil {
		if focus := app.GetFocus(); focus != nil {
			focus.InputHandler()(event, func(p tview.Primitive) { app.SetFocus(p) })
		}
	}
}

func drawEditor(editor *Editor, screen tcell.Screen) {
	width, height := screen.Size()
	screen.Clear()
	editor.SetRect(0, 0, width, height)
	editor.Draw(screen)
	screen.Show()
}

func editorScreenRow(screen tcell.Screen, y int) string {
	width, _ := screen.Size()
	var row strings.Builder
	for x := 0; x < width; x++ {
		character, _, _, _ := screen.GetContent(x, y)
		row.WriteRune(character)
	}
	return row.String()
}

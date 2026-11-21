package edit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestSecondWindowKeepsIndependentDocumentAndCursorState(t *testing.T) {
	dir := t.TempDir()
	firstPath := filepath.Join(dir, "first.prg")
	secondPath := filepath.Join(dir, "second.prg")
	firstContent := strings.Repeat("first\n", 50)
	if err := os.WriteFile(firstPath, []byte(firstContent), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondPath, []byte("second buffer"), 0600); err != nil {
		t.Fatal(err)
	}

	editor, app, screen := newTestEditor(t, 80, 25, nil)
	first, firstText, err := readEditorFile(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	editor.setFile(first, firstText, true)
	editor.area.document.setCursor(documentPosition{line: 10, character: 5}, false)
	editor.area.SetOffset(5, 0)

	editor.lastDir = dir
	openSecondFile(t, editor, app, secondPath)
	if len(editor.windows) != 2 || editor.activeWindow != 1 {
		t.Fatal("opening the second file must create and activate the second window")
	}
	if editor.area.GetText() != "second buffer" || editor.path != secondPath || editor.changed {
		t.Fatal("second window did not load its own file state")
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, '!', tcell.ModNone))
	if !editor.changed || editor.area.GetText() != "!second buffer" {
		t.Fatalf("editing second window: text=%q changed=%v windowChanged=%v focusIsActive=%v",
			editor.area.GetText(), editor.changed, editor.windows[1].changed, app.GetFocus() == editor.area)
	}
	drawEditor(editor, screen)
	if row := editorScreenRow(screen, 24); !strings.Contains(row, "second.prg* [2/2]") {
		t.Fatalf("status bar does not identify the active modified window: %q", row)
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF2, 0, tcell.ModNone))
	if editor.activeWindow != 0 || editor.area.GetText() != firstContent ||
		editor.area.document.cursor != (documentPosition{line: 10, character: 5}) {
		t.Fatal("F2 did not restore the first window's independent buffer and cursor")
	}
	if top, _ := editor.area.GetOffset(); top != 5 {
		t.Fatalf("switching windows changed the first window's scroll position: %d", top)
	}
	drawEditor(editor, screen)
	if row := editorScreenRow(screen, 24); !strings.Contains(row, "first.prg [1/2]") {
		t.Fatalf("status bar does not identify the first window: %q", row)
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF2, 0, tcell.ModNone))
	if editor.activeWindow != 1 || editor.area.GetText() != "!second buffer" || !editor.changed {
		t.Fatal("switching back lost the second window's modified state")
	}
	if editor.windows[0].area.GetText() != firstContent {
		t.Fatal("editing one window changed the other window's buffer")
	}
}

func TestSecondWindowCloseHonorsUnsavedChanges(t *testing.T) {
	dir := t.TempDir()
	secondPath := filepath.Join(dir, "second.prg")
	if err := os.WriteFile(secondPath, []byte("second"), 0600); err != nil {
		t.Fatal(err)
	}

	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("first", false)
	editor.changed = false
	editor.lastDir = dir
	openSecondFile(t, editor, app, secondPath)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, '!', tcell.ModNone))

	closeCommand, _ := commandByID("file.close-window")
	editor.execute(closeCommand)
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Unsaved document" {
		t.Fatal("closing a modified second window must ask what to do")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	if len(editor.windows) != 2 || editor.area.GetText() != "!second" || !editor.changed {
		t.Fatalf("canceling close: windows=%d text=%q changed=%v",
			len(editor.windows), editor.area.GetText(), editor.changed)
	}

	editor.execute(closeCommand)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	if len(editor.windows) != 1 || editor.activeWindow != 0 ||
		editor.area.GetText() != "first" || editor.changed {
		t.Fatal("discarding second-window changes did not return to the intact first window")
	}
}

func TestCloseWindowClosesTheActiveWindow(t *testing.T) {
	dir := t.TempDir()
	secondPath := filepath.Join(dir, "second.prg")
	if err := os.WriteFile(secondPath, []byte("second"), 0600); err != nil {
		t.Fatal(err)
	}

	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("first", false)
	editor.changed = false
	editor.lastDir = dir
	openSecondFile(t, editor, app, secondPath)
	secondWindow := editor.area
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF2, 0, tcell.ModNone))

	closeCommand, _ := commandByID("file.close-window")
	editor.execute(closeCommand)
	if len(editor.windows) != 1 || editor.activeWindow != 0 ||
		editor.area != secondWindow || editor.area.GetText() != "second" {
		t.Fatal("Close Window did not close the active first window and keep the other window")
	}
}

func TestExitChecksUnsavedChangesInInactiveWindow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "second.prg")
	if err := os.WriteFile(path, []byte("second"), 0600); err != nil {
		t.Fatal(err)
	}
	closed := false
	editor, app, _ := newTestEditor(t, 80, 25, func() { closed = true })
	editor.area.SetText("first", false)
	editor.changed = false
	editor.lastDir = dir
	openSecondFile(t, editor, app, path)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, '!', tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF2, 0, tcell.ModNone))

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if closed || editor.activeWindow != 1 || len(editor.dialogs) != 1 ||
		editor.dialogs[0].title != "Unsaved document" {
		t.Fatal("exit did not prompt for changes in the inactive second window")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if closed || len(editor.windows) != 2 || editor.windows[0].area.GetText() != "first" ||
		editor.windows[1].area.GetText() != "!second" || !editor.windows[1].changed {
		t.Fatal("canceling exit lost a document or its unsaved state")
	}
}

func TestSamePathWindowsRemainIndependentAndDetectExternalSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "shared.prg")
	if err := os.WriteFile(path, []byte("base"), 0600); err != nil {
		t.Fatal(err)
	}
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	absolute, text, err := readEditorFile(path)
	if err != nil {
		t.Fatal(err)
	}
	editor.setFile(absolute, text, true)
	editor.lastDir = dir
	openSecondFile(t, editor, app, path)
	if editor.area.GetText() != "base" || editor.windows[0].area.GetText() != "base" {
		t.Fatal("opening the same path did not create two independent copies")
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, '!', tcell.ModNone))
	save, _ := commandByID("file.save")
	editor.execute(save)
	if got, err := os.ReadFile(path); err != nil || string(got) != "!base" {
		t.Fatalf("saving the second buffer wrote %q, err=%v", got, err)
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF2, 0, tcell.ModNone))
	if editor.area.GetText() != "base" || editor.changed {
		t.Fatal("saving one window changed the other window's buffer or dirty state")
	}
	editor.execute(save)
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "File changed on disk" {
		t.Fatal("saving a stale same-path buffer must detect the external write")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if got, err := os.ReadFile(path); err != nil || string(got) != "!base" {
		t.Fatalf("canceling stale save changed the file to %q, err=%v", got, err)
	}
}

func TestSecondWindowDividerCanMoveAndPersistsPaneSizes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "second.prg")
	if err := os.WriteFile(path, []byte("second"), 0600); err != nil {
		t.Fatal(err)
	}
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	editor.lastDir = dir
	openSecondFile(t, editor, app, path)

	up, _ := commandByID("window.divider-up")
	down, _ := commandByID("window.divider-down")
	editor.execute(down)
	if editor.dividerPosition != 55 {
		t.Fatalf("moving the divider down set position %d, want 55", editor.dividerPosition)
	}
	editor.execute(up)
	if editor.dividerPosition != 50 {
		t.Fatalf("moving the divider up set position %d, want 50", editor.dividerPosition)
	}
	editor.execute(down)
	editor.execute(down)
	editor.execute(down)
	drawEditor(editor, screen)
	_, _, _, firstRows := editor.windows[0].area.GetRect()
	_, _, _, secondRows := editor.windows[1].area.GetRect()
	if firstRows <= secondRows {
		t.Fatalf("divider position did not resize panes: first=%d rows, second=%d rows", firstRows, secondRows)
	}
}

func openSecondFile(t *testing.T, editor *Editor, app *tview.Application, path string) {
	t.Helper()
	command, ok := commandByID("file.window")
	if !ok {
		t.Fatal("second-window command is missing")
	}
	editor.execute(command)
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Open file" {
		t.Fatal("opening the second window must show the file browser")
	}
	dialog := editor.dialogs[0]
	filename := dialog.fields[0]
	dialog.widgets[filename].value = path
	dialog.focus = 0
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if len(editor.dialogs) != 0 {
		t.Fatal("opening an existing file in the second window left a dialog open")
	}
}

package edit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestDialogWidgetsAndModality(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("text preserved", true)
	_, start, end := editor.area.GetSelection()
	for _, id := range []string{"file.merge", "file.window", "search.replace",
		"search.line", "layout.document", "layout.setup",
		"print.setup", "options.display", "options.mouse", "options.editor", "options.save"} {
		command, ok := commandByID(id)
		if !ok {
			t.Fatalf("missing command %s", id)
		}
		editor.execute(command)
		if len(editor.dialogs) != 1 {
			t.Fatalf("%s: expected one preview dialog", id)
		}
		d := editor.dialogs[0]
		drawEditor(editor, screen)
		x, y, width, height := d.GetRect()
		border, _, _, _ := screen.GetContent(x, y)
		if border != '\u2554' || width > 80 || height > 25 {
			t.Fatalf("%s: invalid double-border geometry", id)
		}
		before := d.focus
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone))
		if d.focus != len(d.fields)-1 {
			t.Fatalf("%s: Shift-Tab must wrap to last field", id)
		}
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
		if d.focus != before {
			t.Fatalf("%s: Tab did not restore focus", id)
		}
		editor.MouseHandler()(tview.MouseLeftClick, tcell.NewEventMouse(0, 1, tcell.Button1, tcell.ModNone), func(p tview.Primitive) { app.SetFocus(p) })
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))
		_, afterStart, afterEnd := editor.area.GetSelection()
		if editor.area.GetText() != "text preserved" || start != afterStart || end != afterEnd {
			t.Fatalf("%s: modal input leaked into document", id)
		}
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
		if len(editor.dialogs) != 0 || app.GetFocus() != editor.area {
			t.Fatalf("%s: cancel must restore editor focus", id)
		}
	}
	command, _ := commandByID("search.replace")
	editor.execute(command)
	d := editor.dialogs[0]
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
	if d.value("find") != "n" {
		t.Fatal("button initials must not steal typed field text")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	if d.value("replace") != "x" {
		t.Fatal("second search field did not accept text")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Search" {
		t.Fatal("Find and Replace must report when the search text is not present")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	command, _ = commandByID("search.find")
	editor.execute(command)
	d = editor.dialogs[0]
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	if !d.widgets[d.fields[d.focus]].checked {
		t.Fatal("Space must toggle a focused checkbox")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	command, _ = commandByID("options.editor")
	editor.execute(command)
	d = editor.dialogs[0]
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	if d.widgets[d.fields[0]].selected != 1 {
		t.Fatal("Down must change the focused radio group")
	}
}

func TestDialogButtonBracketsFollowFocus(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	dialog := newDialog("Button focus", 40, 8)
	dialog.buttons([]string{"ok", "cancel"}, []string{"OK", "Cancel"}, "cancel")
	dialog.focus = 1
	selected := ""
	dialog.result = func(id string) { selected = id }
	editor.openDialog(dialog)

	drawEditor(editor, screen)
	x, y, _, height := dialog.GetRect()
	buttonRow := y + height - 2
	cancel, _, _, _ := screen.GetContent(x+10, buttonRow)
	ok, _, _, _ := screen.GetContent(x+2, buttonRow)
	if cancel != '[' || ok != ' ' {
		t.Fatal("focused Cancel button must carry brackets instead of the default button")
	}

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	drawEditor(editor, screen)
	ok, _, _, _ = screen.GetContent(x+2, buttonRow)
	cancel, _, _, _ = screen.GetContent(x+10, buttonRow)
	if ok != '[' || cancel != ' ' {
		t.Fatal("button brackets did not follow Tab focus")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if selected != "ok" || dialog.defaultID != "cancel" {
		t.Fatal("focused button activation and default Enter action must remain independent")
	}
}

func TestDialogValidationNestedReturnAndSmallScreen(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	command, _ := commandByID("search.line")
	editor.execute(command)
	d := editor.dialogs[0]
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if len(editor.dialogs) != 2 || editor.dialogs[1].title != "Invalid value" {
		t.Fatal("invalid numbers must show an error, not silently default or close")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if len(editor.dialogs) != 1 || app.GetFocus() != d || d.value("number") != "x" {
		t.Fatal("dismissing the error must preserve the parent input and focus")
	}
	screen.SetSize(20, 8)
	drawEditor(editor, screen)
	x, y, width, height := d.GetRect()
	if x < 0 || y < 0 || x+width > 20 || y+height > 8 {
		t.Fatal("dialog exceeded small screen boundaries")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	command, _ = commandByID("file.open")
	editor.execute(command)
	d = editor.dialogs[0]
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone))
	w := d.widgets[d.fields[d.focus]]
	if len(w.choices) > 0 && w.selected != len(w.choices)-1 {
		t.Fatal("list End must select the final entry")
	}
}

func TestUnsavedChangesOfferSaveDiscardOrCancel(t *testing.T) {
	closed := false
	editor, app, _ := newTestEditor(t, 80, 25, func() { closed = true })
	editor.lastDir = t.TempDir()
	editor.area.SetText("keep me", true)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if len(editor.dialogs) != 1 {
		t.Fatal("unsaved exit needs a modal confirmation")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	if closed || len(editor.dialogs) != 2 || editor.dialogs[1].title != "Save file as" {
		t.Fatal("Save must open the real Save As dialog")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if editor.confirmExit || closed || editor.area.GetText() != "keep me" || app.GetFocus() != editor.area {
		t.Fatal("Cancel must preserve text and restore focus")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	if !closed || editor.changed || editor.area.GetText() != "" || len(editor.dialogs) != 0 {
		t.Fatal("Don't Save must discard before returning to the Shell")
	}
}

func TestSaveOnExitWritesFileBeforeClosing(t *testing.T) {
	dir := t.TempDir()
	closed := false
	editor, app, _ := newTestEditor(t, 80, 25, func() { closed = true })
	editor.lastDir = dir
	editor.area.SetText("saved content", true)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	saveDialog := editor.dialogs[len(editor.dialogs)-1]
	saveDialog.widgets[saveDialog.fields[0]].value = "saved.prg"
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	data, err := os.ReadFile(filepath.Join(dir, "saved.prg"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "saved content" || !closed || len(editor.dialogs) != 0 {
		t.Fatalf("Save on exit wrote %q, closed=%v dialogs=%d", data, closed, len(editor.dialogs))
	}
}

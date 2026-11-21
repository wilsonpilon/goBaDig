package edit

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestR13LayoutSetupAppliesOptionsWithoutChangingDocument(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("existing text", false)
	editor.changed = false
	editor.syncActiveWindow()
	document := editor.area.document
	revision := document.revision
	command, _ := commandByID("layout.setup")
	editor.execute(command)
	dialog := editor.dialogs[len(editor.dialogs)-1]
	for index := range dialog.widgets {
		switch dialog.widgets[index].id {
		case "line-length":
			dialog.widgets[index].value = "72"
		case "first-margin":
			dialog.widgets[index].value = "3"
		case "left-margin":
			dialog.widgets[index].value = "4"
		case "begin":
			dialog.widgets[index].selected = 0
		}
	}
	dialog.activate("ok")
	if len(editor.dialogs) != 0 {
		t.Fatal("successful setup did not close its dialog")
	}
	if document.lineLength != 72 || document.firstMargin != 3 || document.leftMargin != 4 ||
		!document.formatAtCursor || document.text != "existing text" || document.revision != revision ||
		editor.changed {
		t.Fatalf("setup state: line=%d first=%d left=%d cursor=%v text=%q revision=%d/%d dirty=%v",
			document.lineLength, document.firstMargin, document.leftMargin, document.formatAtCursor,
			document.text, document.revision, revision, editor.changed)
	}
}

func TestR13LayoutSetupCancelAndAutoToggle(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	command, _ := commandByID("layout.setup")
	editor.execute(command)
	dialog := editor.dialogs[len(editor.dialogs)-1]
	dialog.widgets[dialog.fields[0]].value = "90"
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if editor.area.document.lineLength != 65 {
		t.Fatal("cancel applied a layout setting")
	}
	auto, _ := commandByID("layout.auto")
	editor.execute(auto)
	if !editor.auto || !editor.area.document.autoIndent {
		t.Fatal("Auto did not enable indentation behavior")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if editor.area.GetText() != " \r\n " {
		t.Fatalf("Auto Enter failed to copy indentation: %q", editor.area.GetText())
	}
}

package edit

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestR15ExpandTabsTopAndCursorScopes(t *testing.T) {
	d := newDocument("a\tb\r\n\t界\tz")
	d.tabWidth = 4
	if !d.expandTabs(false) {
		t.Fatal("Top did not expand tabs")
	}
	if want := "a   b\r\n    界  z"; d.text != want {
		t.Fatalf("expanded text = %q, want %q", d.text, want)
	}

	d = newDocument("a\tb\tc")
	d.tabWidth = 4
	d.setCursor(documentPosition{0, 3}, false)
	if !d.expandTabs(true) {
		t.Fatal("Cursor did not expand tabs to the right of the cursor")
	}
	if want := "a\tb   c"; d.text != want {
		t.Fatalf("Cursor expansion = %q, want %q", d.text, want)
	}
	if d.column(d.cursor) != 5 {
		t.Fatalf("Cursor moved from its visual column: %d", d.column(d.cursor))
	}
}

func TestR15PutInTabsConvertsIndentOnly(t *testing.T) {
	d := newDocument("        alpha  beta\r\n  \t gamma")
	d.tabWidth = 4
	if !d.putInTabs(false) {
		t.Fatal("Top did not convert leading indentation")
	}
	if want := "\t\talpha  beta\r\n\t gamma"; d.text != want {
		t.Fatalf("Put in Tabs = %q, want %q", d.text, want)
	}

	d = newDocument("outside  text\r\n    current  text\r\n        next")
	d.tabWidth = 4
	d.setCursor(documentPosition{0, 8}, false)
	if !d.putInTabs(true) {
		t.Fatal("Cursor did not process following lines")
	}
	if want := "outside  text\r\n\tcurrent  text\r\n\t\tnext"; d.text != want {
		t.Fatalf("Cursor Put in Tabs = %q, want %q", d.text, want)
	}
}

func TestR15StartDialogCancelAndCursorAction(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("a\tb\r\n\tc", false)
	editor.area.document.tabWidth = 4
	editor.changed = false
	editor.syncActiveWindow()
	original := editor.area.GetText()
	command, _ := commandByID("layout.expand-tabs")
	editor.execute(command)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if editor.area.GetText() != original || editor.changed {
		t.Fatal("canceling tab expansion modified the document")
	}

	editor.area.document.setCursor(documentPosition{0, 2}, false)
	editor.execute(command)
	dialog := editor.dialogs[len(editor.dialogs)-1]
	dialog.activate("cursor")
	if editor.area.GetText() != "a\tb\r\n    c" || !editor.changed {
		t.Fatalf("Cursor tab expansion failed: %q, dirty=%v", editor.area.GetText(), editor.changed)
	}
}

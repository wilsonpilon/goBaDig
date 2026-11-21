package edit

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestR16InsertPageBreaksTopAndCursor(t *testing.T) {
	d := newDocument("one\r\ntwo\r\nthree\r\nfour\r\nfive\r\nsix")
	if d.pageLength != 62 {
		t.Fatalf("default page length = %d, want 62", d.pageLength)
	}
	d.pageLength = 2
	if !d.insertPageBreaks(false) {
		t.Fatal("Top did not insert page breaks")
	}
	if want := "one\r\ntwo\r\n\fthree\r\nfour\r\n\ffive\r\nsix"; d.text != want {
		t.Fatalf("Top result = %q, want %q", d.text, want)
	}
	if d.insertPageBreaks(false) {
		t.Fatal("repeating insertion added duplicate page breaks")
	}

	d = newDocument("before\r\none\r\ntwo\r\nthree\r\nfour\r\nfive")
	d.pageLength = 2
	d.setCursor(documentPosition{1, 0}, false)
	if !d.insertPageBreaks(true) {
		t.Fatal("Cursor did not insert page breaks")
	}
	if want := "before\r\none\r\ntwo\r\n\fthree\r\nfour\r\n\ffive"; d.text != want {
		t.Fatalf("Cursor result = %q, want %q", d.text, want)
	}
}

func TestR16InsertPageBreaksAvoidsTrailingEmptyPage(t *testing.T) {
	d := newDocument("one\ntwo\nthree\n")
	d.pageLength = 3
	if d.insertPageBreaks(false) {
		t.Fatalf("inserted a break after an exact-length document: %q", d.text)
	}

	d = newDocument("")
	if d.insertPageBreaks(false) {
		t.Fatal("inserted a page break into an empty document")
	}
}

func TestR16RemovePageBreaksTopAndCursor(t *testing.T) {
	d := newDocument("before\f\r\nremove\f here\r\nlast\f")
	d.setCursor(documentPosition{1, 0}, false)
	if !d.removePageBreaks(true) {
		t.Fatal("Cursor did not remove page breaks")
	}
	if want := "before\f\r\nremove here\r\nlast"; d.text != want {
		t.Fatalf("Cursor result = %q, want %q", d.text, want)
	}
	if !d.removePageBreaks(false) {
		t.Fatal("Top did not remove the remaining page break")
	}
	if want := "before\r\nremove here\r\nlast"; d.text != want {
		t.Fatalf("Top result = %q, want %q", d.text, want)
	}
	if d.removePageBreaks(false) {
		t.Fatal("removal reported a change when there were no page breaks")
	}
}

func TestR16PageBreakCommandsAndCancel(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("one\ntwo\nthree", false)
	editor.area.document.pageLength = 2
	editor.changed = false
	original := editor.area.GetText()

	insert, ok := commandByID("layout.insert-pages")
	if !ok || !editor.available(insert) {
		t.Fatal("Insert page breaks command is unavailable")
	}
	editor.execute(insert)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if editor.area.GetText() != original || editor.changed {
		t.Fatal("canceling page-break insertion modified the document")
	}

	editor.execute(insert)
	editor.dialogs[len(editor.dialogs)-1].activate("cursor")
	if want := "one\ntwo\n\fthree"; editor.area.GetText() != want || !editor.changed {
		t.Fatalf("Cursor insertion = %q, dirty=%v; want %q", editor.area.GetText(), editor.changed, want)
	}

	remove, ok := commandByID("layout.remove-pages")
	if !ok || !editor.available(remove) {
		t.Fatal("Remove page breaks command is unavailable")
	}
	editor.execute(remove)
	editor.dialogs[len(editor.dialogs)-1].activate("top")
	if editor.area.GetText() != original || !editor.changed {
		t.Fatalf("Top removal = %q, dirty=%v; want %q", editor.area.GetText(), editor.changed, original)
	}
}

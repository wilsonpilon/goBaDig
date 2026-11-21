package edit

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestR14ParagraphFormatUsesMarginsAndPreservesBlankLines(t *testing.T) {
	d := newDocument("one   two\r\nthree four\r\n\r\nnext")
	d.lineLength, d.firstMargin, d.leftMargin = 11, 1, 2
	if !d.formatParagraph() {
		t.Fatal("paragraph formatter did not apply configured margins")
	}
	if want := " one two\r\n  three four\r\n\r\nnext"; d.text != want {
		t.Fatalf("formatted text = %q, want %q", d.text, want)
	}
	if d.cursor != (documentPosition{0, 0}) {
		t.Fatalf("paragraph formatting cursor = %+v, want paragraph start", d.cursor)
	}
}

func TestR14DocumentFormatHonorsTopAndCursorScope(t *testing.T) {
	d := newDocument("aa bb cc\r\n\r\nunchanged words\r\none two three\r\n\r\nnext section")
	d.lineLength = 5
	if !d.formatDocument(false) {
		t.Fatal("Top document format made no changes")
	}
	if want := "aa bb\r\ncc\r\n\r\nunchanged\r\nwords\r\none\r\ntwo\r\nthree\r\n\r\nnext\r\nsection"; d.text != want {
		t.Fatalf("Top format = %q, want %q", d.text, want)
	}

	d = newDocument("unchanged words\r\none two three\r\n\r\nnext section")
	d.lineLength = 7
	d.setCursor(documentPosition{1, 4}, false)
	if !d.formatDocument(true) {
		t.Fatal("Cursor document format made no changes")
	}
	if want := "unchanged words\r\none two\r\nthree\r\n\r\nnext\r\nsection"; d.text != want {
		t.Fatalf("Cursor format = %q, want %q", d.text, want)
	}
	if d.cursor != (documentPosition{1, 3}) {
		t.Fatalf("Cursor format position = %+v, want start of formatted suffix", d.cursor)
	}
}

func TestR14CenterLineRespectsMarginsAndCursor(t *testing.T) {
	d := newDocument("  hello   \r\nnext")
	d.lineLength, d.leftMargin = 20, 2
	d.setCursor(documentPosition{0, 3}, false)
	if !d.centerLine() {
		t.Fatal("Center line made no change")
	}
	if want := "         hello\r\nnext"; d.text != want {
		t.Fatalf("centered text = %q, want %q", d.text, want)
	}
	if d.cursor != (documentPosition{0, 10}) {
		t.Fatalf("center cursor = %+v, want position after the original character", d.cursor)
	}
}

func TestR14DocumentFormatDialogCancelAndCursorAction(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("unchanged words\r\none two three", false)
	editor.changed = false
	editor.syncActiveWindow()
	editor.area.Select(2, 9)
	selection, selectionStart, selectionEnd := editor.area.GetSelection()
	cursor := editor.area.document.cursor
	documentFormat, _ := commandByID("layout.document")
	editor.execute(documentFormat)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	afterSelection, afterStart, afterEnd := editor.area.GetSelection()
	if editor.area.GetText() != "unchanged words\r\none two three" || editor.changed ||
		editor.area.document.cursor != cursor || afterSelection != selection ||
		afterStart != selectionStart || afterEnd != selectionEnd {
		t.Fatal("canceling Document format changed text, cursor or selection")
	}

	editor.area.document.lineLength = 7
	editor.area.document.setCursor(documentPosition{1, 4}, false)
	editor.execute(documentFormat)
	dialog := editor.dialogs[len(editor.dialogs)-1]
	dialog.activate("cursor")
	if editor.area.GetText() != "unchanged words\r\none two\r\nthree" || !editor.changed ||
		editor.notice != "Document formatted." {
		t.Fatalf("Cursor action did not format the document: %q, dirty=%v, notice=%q",
			editor.area.GetText(), editor.changed, editor.notice)
	}
}

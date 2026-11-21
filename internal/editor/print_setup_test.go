package edit

import (
	"bytes"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestR23PrintSetupAppliesAllOptionsWithoutChangingDocument(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("unchanged", false)
	editor.changed = false
	document := editor.area.document
	revision := document.revision

	command, ok := commandByID("print.setup")
	if !ok || !editor.available(command) {
		t.Fatal("Print Setup command is unavailable")
	}
	editor.execute(command)
	dialog := editor.dialogs[len(editor.dialogs)-1]
	setDialogValue(dialog, "top-margin", "3")
	setDialogValue(dialog, "left-margin", "4")
	setDialogValue(dialog, "page-length", "44")
	for index := range dialog.widgets {
		if dialog.widgets[index].id == "formfeed" {
			dialog.widgets[index].checked = false
		}
		if dialog.widgets[index].id == "after" {
			dialog.widgets[index].selected = 2
		}
	}
	dialog.activate("ok")

	if document.printTopMargin != 3 || document.printLeftMargin != 4 ||
		document.pageLength != 44 || document.formfeedBetween || document.formfeedAfter != 2 {
		t.Fatalf("print settings were not applied: %+v", document)
	}
	if document.text != "unchanged" || document.revision != revision || editor.changed {
		t.Fatal("Print Setup modified or dirtied the document")
	}
}

func TestR23PrintSetupCancelAndInvalidInputPreserveSettings(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	document := editor.area.document
	command, _ := commandByID("print.setup")
	editor.execute(command)
	setDialogValue(editor.dialogs[0], "page-length", "40")
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if document.pageLength != 62 {
		t.Fatal("cancel applied Print Setup values")
	}

	editor.execute(command)
	setDialogValue(editor.dialogs[0], "page-length", "0")
	editor.dialogs[0].activate("ok")
	if document.pageLength != 62 || len(editor.dialogs) != 2 ||
		editor.dialogs[1].title != "Invalid value" {
		t.Fatal("invalid page length was applied or not reported")
	}
}

func TestR23PrintStreamMarginsTabsAndPageFlow(t *testing.T) {
	document := newDocument("A\tB\nsecond\nthird")
	document.pageLength = 2
	document.printTopMargin = 1
	document.printLeftMargin = 2
	document.tabWidth = 4
	document.formfeedBetween = true
	document.formfeedAfter = 1

	stream, err := document.printStream()
	if err != nil {
		t.Fatal(err)
	}
	if want := "\n  A B\n  second\n\f\n  third\f"; string(stream) != want {
		t.Fatalf("print stream = %q, want %q", stream, want)
	}
}

func TestR23PrintStreamExactPageEmptyDocumentAndExplicitBreak(t *testing.T) {
	exact := newDocument("one\r\ntwo\r\n")
	exact.pageLength = 2
	exact.printTopMargin = 0
	exact.formfeedAfter = 0
	stream, err := exact.printStream()
	if err != nil {
		t.Fatal(err)
	}
	if want := "one\r\ntwo\r\n"; string(stream) != want {
		t.Fatalf("exact-page stream = %q, want %q", stream, want)
	}

	empty := newDocument("")
	stream, err = empty.printStream()
	if err != nil || string(stream) != "\r\n\r\n\f" {
		t.Fatalf("empty-document stream = %q, err=%v", stream, err)
	}

	withoutAutomaticFormfeed := newDocument("one\ntwo\nthree")
	withoutAutomaticFormfeed.pageLength = 2
	withoutAutomaticFormfeed.printTopMargin = 1
	withoutAutomaticFormfeed.formfeedBetween = false
	withoutAutomaticFormfeed.formfeedAfter = 0
	stream, err = withoutAutomaticFormfeed.printStream()
	if err != nil {
		t.Fatal(err)
	}
	if want := "\none\ntwo\n\nthree"; string(stream) != want {
		t.Fatalf("stream without automatic formfeed = %q, want %q", stream, want)
	}

	explicit := newDocument("one\n\ftwo")
	explicit.pageLength = 1
	explicit.printTopMargin = 1
	explicit.formfeedBetween = false
	explicit.formfeedAfter = 0
	stream, err = explicit.printStream()
	if err != nil {
		t.Fatal(err)
	}
	if want := "\none\n\f\ntwo"; string(stream) != want {
		t.Fatalf("explicit-break stream = %q, want %q", stream, want)
	}
}

func TestR23PrintStreamRejectsInvalidSettings(t *testing.T) {
	document := newDocument("text")
	document.pageLength = 0
	if _, err := document.printStream(); err == nil {
		t.Fatal("invalid page length was accepted")
	}
}

func TestR23PrintStreamAfterPrintFormfeedCounts(t *testing.T) {
	for after := 0; after <= 2; after++ {
		document := newDocument("x")
		document.printTopMargin = 0
		document.formfeedAfter = after
		stream, err := document.printStream()
		if err != nil {
			t.Fatal(err)
		}
		want := "x"
		for range after {
			want += "\f"
		}
		if string(stream) != want {
			t.Fatalf("after-print option %d produced %q, want %q", after, stream, want)
		}
	}
}

func TestR23PrintStreamPreservesInvalidUTF8Bytes(t *testing.T) {
	document := newDocument(string([]byte{0xff, 'x'}))
	document.printTopMargin = 0
	document.formfeedAfter = 0
	stream, err := document.printStream()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stream, []byte{0xff, 'x'}) {
		t.Fatalf("print stream changed source bytes: %v", stream)
	}
}

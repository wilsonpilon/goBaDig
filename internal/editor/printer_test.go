package edit

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestR24PrintDocumentSendsPagedStreamWithoutChangingDocument(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("first\nsecond", false)
	editor.changed = false
	document := editor.area.document
	document.pageLength = 1
	document.printTopMargin = 0
	document.formfeedAfter = 0
	revision := document.revision
	want, err := document.printStream()
	if err != nil {
		t.Fatal(err)
	}

	var sent []byte
	editor.printer = func(stream []byte) error {
		sent = append([]byte(nil), stream...)
		return nil
	}
	command, ok := commandByID("print.document")
	if !ok || !editor.available(command) {
		t.Fatal("Print command is unavailable")
	}
	editor.execute(command)

	if !bytes.Equal(sent, want) {
		t.Fatalf("printer received %q, want %q", sent, want)
	}
	if document.text != "first\nsecond" || document.revision != revision || editor.changed {
		t.Fatal("printing changed or dirtied the document")
	}
	if editor.notice != "Sent 13 bytes to the default printer." {
		t.Fatalf("print notice = %q", editor.notice)
	}
}

func TestR24PrintDocumentReportsPrinterFailureWithoutChangingDocument(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("unchanged", false)
	editor.changed = true
	document := editor.area.document
	revision := document.revision
	editor.printer = func([]byte) error { return errors.New("printer offline") }

	command, _ := commandByID("print.document")
	editor.execute(command)

	if document.text != "unchanged" || document.revision != revision || !editor.changed {
		t.Fatal("printer failure changed or cleared document state")
	}
	reported := false
	if len(editor.dialogs) == 1 {
		for _, widget := range editor.dialogs[0].widgets {
			reported = reported || strings.Contains(widget.label, "printer offline")
		}
	}
	if len(editor.dialogs) != 1 || editor.dialogs[0].context != "print.document" || !reported {
		t.Fatal("printer failure was not reported in the Print context")
	}
}

func TestR24dPrintDocumentIncludesUnloadedFileSuffix(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("prefix", false)
	editor.path = filepath.Join(t.TempDir(), "partial.txt")
	editor.diskText = "prefix-and-suffix"
	editor.diskKnown = true
	editor.changed = false
	document := editor.area.document
	document.setSourceBase(0)
	want, err := document.printStreamForText(editor.diskText)
	if err != nil {
		t.Fatal(err)
	}

	var sent []byte
	editor.printer = func(stream []byte) error {
		sent = append([]byte(nil), stream...)
		return nil
	}
	editor.printDocument()

	if !bytes.Equal(sent, want) {
		t.Fatalf("printer received %q, want full-file stream %q", sent, want)
	}
}

func TestR24bPrintBlockSendsOnlyMarkedTextWithPrintSetup(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("before\nselected\ttext\nafter", false)
	editor.changed = true
	document := editor.area.document
	document.pageLength = 1
	document.printTopMargin = 1
	document.printLeftMargin = 2
	document.tabWidth = 4
	document.formfeedAfter = 0
	start, end := len("before\n"), len("before\nselected\ttext")
	setMarkedBytes(document, start, end)
	revision, cursor := document.revision, document.cursor
	markers := document.blockMarkers
	want := []byte("\n  selected  text")

	var sent []byte
	editor.printer = func(stream []byte) error {
		sent = append([]byte(nil), stream...)
		return nil
	}
	command, ok := commandByID("print.block")
	if !ok || !editor.available(command) {
		t.Fatal("Print Block command is unavailable")
	}
	editor.execute(command)

	if !bytes.Equal(sent, want) {
		t.Fatalf("printer received %q, want %q", sent, want)
	}
	if document.text != "before\nselected\ttext\nafter" || document.revision != revision ||
		document.cursor != cursor || document.blockMarkers != markers ||
		!editor.changed {
		t.Fatal("Print Block changed document, cursor, markers or modified state")
	}
}

func TestR24bPrintBlockRequiresTwoMarkers(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	called := false
	editor.printer = func([]byte) error {
		called = true
		return nil
	}

	command, _ := commandByID("print.block")
	editor.execute(command)

	if called || editor.notice != "Set two block markers before using Print Block." {
		t.Fatal("Print Block without markers was not rejected before printing")
	}
}

func TestR24cEjectPageSendsOnlyFormfeedWithoutChangingDocument(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("unchanged", false)
	editor.changed = true
	document := editor.area.document
	revision, cursor := document.revision, document.cursor

	var sent []byte
	editor.printer = func(stream []byte) error {
		sent = append([]byte(nil), stream...)
		return nil
	}
	command, ok := commandByID("print.eject")
	if !ok || !editor.available(command) {
		t.Fatal("Eject page command is unavailable")
	}
	editor.execute(command)

	if !bytes.Equal(sent, []byte{'\f'}) {
		t.Fatalf("Eject page sent %q, want one formfeed", sent)
	}
	if document.text != "unchanged" || document.revision != revision ||
		document.cursor != cursor || !editor.changed {
		t.Fatal("Eject page changed the document or its modified state")
	}
	if editor.notice != "Sent an eject-page formfeed to the default printer." {
		t.Fatalf("Eject page notice = %q", editor.notice)
	}
}

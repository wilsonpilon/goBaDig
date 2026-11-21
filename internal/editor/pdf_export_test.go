package edit

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestR24dPDFEmbedsUnicodeFontAndProducesPagedOutput(t *testing.T) {
	document := newDocument("café\nsecond")
	document.pageLength = 1
	document.printTopMargin = 0
	document.formfeedBetween = true
	document.formfeedAfter = 0

	data, err := document.pdfForText(document.text)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) || !bytes.Contains(data, []byte("/FontFile2")) {
		t.Fatal("PDF header or embedded TrueType font is missing")
	}
}

func TestR24dPDFRejectsInvalidUTF8UnsupportedGlyphsAndControls(t *testing.T) {
	for _, test := range []struct {
		name, text, wantError string
	}{
		{"invalid UTF-8", string([]byte{0xff}), "valid UTF-8"},
		{"unsupported glyph", "\u4e00", "U+4E00"},
		{"control character", "\x00", "U+0000"},
	} {
		t.Run(test.name, func(t *testing.T) {
			document := newDocument(test.text)
			document.printTopMargin = 0
			document.formfeedAfter = 0
			if _, err := document.pdfForText(document.text); err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("PDF error = %v, want text %q", err, test.wantError)
			}
		})
	}
}

func TestR24dPDFRejectsLinesThatWouldBeClipped(t *testing.T) {
	document := newDocument(strings.Repeat("W", 200))
	document.printTopMargin = 0
	document.formfeedAfter = 0

	if _, err := document.pdfForText(document.text); err == nil ||
		!strings.Contains(err.Error(), "A4 printable width") {
		t.Fatalf("wide-line PDF error = %v", err)
	}
}

func TestR24dPDFExportDialogAddsExtensionAndPreservesDocument(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	dir := t.TempDir()
	editor.lastDir = dir
	editor.area.SetText("café", false)
	editor.changed = true
	revision := editor.area.document.revision

	command, ok := commandByID("print.pdf")
	if !ok || !editor.available(command) {
		t.Fatal("Export PDF command is unavailable")
	}
	editor.execute(command)
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Export PDF" {
		t.Fatal("Export PDF did not open its destination dialog")
	}
	dialog := editor.dialogs[0]
	dialog.widgets[dialog.fields[0]].value = "report"
	dialog.result("ok")

	path := filepath.Join(dir, "report.pdf")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("exported file is not a PDF")
	}
	if editor.area.GetText() != "café" || editor.area.document.revision != revision ||
		!editor.changed || editor.path != "" || len(editor.dialogs) != 0 {
		t.Fatal("PDF export changed the document, file identity, or modified state")
	}
}

func TestR24dPDFExportRejectsNonPDFExtension(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.lastDir = t.TempDir()
	command, ok := commandByID("print.pdf")
	if !ok {
		t.Fatal("Export PDF command is unavailable")
	}
	editor.execute(command)
	dialog := editor.dialogs[0]
	dialog.widgets[dialog.fields[0]].value = "report.txt"
	dialog.result("ok")

	if len(editor.dialogs) != 2 || editor.dialogs[1].context != "print.pdf" {
		t.Fatal("non-PDF extension was not reported")
	}
	if _, err := os.Stat(filepath.Join(editor.lastDir, "report.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("invalid extension created an output file")
	}
}

func TestR24dPDFOverwriteRequiresConfirmation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.pdf")
	original := []byte("keep this file")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.lastDir = dir
	editor.area.SetText("new PDF", false)
	command, _ := commandByID("print.pdf")
	editor.execute(command)
	dialog := editor.dialogs[0]
	dialog.widgets[dialog.fields[0]].value = "existing.pdf"
	dialog.result("ok")

	if len(editor.dialogs) != 2 || editor.dialogs[1].title != "Overwrite file?" {
		t.Fatal("existing PDF was not protected by an overwrite prompt")
	}
	editor.dialogs[1].result("cancel")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, original) || len(editor.dialogs) != 1 {
		t.Fatal("canceling PDF overwrite changed the existing file")
	}
}

func TestR24dPDFExportRejectsCurrentDocumentPath(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	path := filepath.Join(t.TempDir(), "document.pdf")
	editor.path = path
	editor.area.SetText("document", false)

	editor.exportPDFTo(path, false)

	if len(editor.dialogs) != 1 || editor.dialogs[0].context != "print.pdf" {
		t.Fatal("export to the active document path was not rejected")
	}
	reported := false
	for _, widget := range editor.dialogs[0].widgets {
		reported = reported || strings.Contains(widget.label, "cannot be the current document path")
	}
	if !reported {
		t.Fatal("current-path error was not shown")
	}
}

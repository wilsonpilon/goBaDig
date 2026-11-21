package edit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestR20aInformationDisplayReportsFullDocumentAndCursorData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "source.prg")
	text := "alpha\ncafé\n"
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.setFile(path, text, true)
	editor.area.document.cursor = documentPosition{line: 1, character: 3}

	lines, err := editor.informationLines()
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(lines, "\n")
	for _, want := range []string{
		"File: source.prg",
		"Path: " + path,
		"Directory: " + dir,
		"Size: 12 bytes",
		"Lines: 3  Words: 2  Characters: 11",
		"Cursor: line 2, column 4",
		"On-disk modified:",
		"System time:",
		"Free space available:",
		"Character at cursor: 'é' (U+00E9, 233)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("information display missing %q in:\n%s", want, got)
		}
	}
}

func TestR20aInformationDisplayIncludesUnloadedFileSuffix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "partial.prg")
	fullText := "prefix\nloaded\nunloaded suffix"
	if err := os.WriteFile(path, []byte(fullText), 0600); err != nil {
		t.Fatal(err)
	}
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.path, editor.filename = path, filepath.Base(path)
	editor.diskText, editor.diskKnown, editor.fileBase = fullText, true, len("prefix\n")
	editor.area.SetText("loaded\n", false)
	editor.area.document.setSourceBase(editor.fileBase)
	editor.area.document.cursor = documentPosition{line: 0, character: 6}

	lines, err := editor.informationLines()
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(lines, "\n")
	for _, want := range []string{
		"Size: 29 bytes",
		"Lines: 3",
		"In buffer: 7 bytes  Cursor: line 2, column 7",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("information display missing %q in:\n%s", want, got)
		}
	}
}

func TestR20aInformationCommandShowsDialogWithoutChangingDocument(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("document", false)
	editor.changed = true
	revision := editor.area.document.revision
	command, ok := commandByID("view.information")
	if !ok || !editor.available(command) {
		t.Fatal("Information display command is unavailable")
	}

	editor.execute(command)

	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Information display" {
		t.Fatal("Information display did not open its report")
	}
	if editor.area.GetText() != "document" || editor.area.document.revision != revision ||
		!editor.changed {
		t.Fatal("Information display changed the document")
	}
}

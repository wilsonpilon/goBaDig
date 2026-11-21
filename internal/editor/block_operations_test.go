package edit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func setMarkedBytes(d *document, start, end int) {
	d.setCursor(d.positionAtByte(start), false)
	d.setBlockMarkerAtCursor()
	d.setCursor(d.positionAtByte(end), false)
	d.setBlockMarkerAtCursor()
}

func TestDocumentBlockIndentUnindentAndSourceMapping(t *testing.T) {
	d := newDocument("  first\n\tsecond\n\nlast")
	d.setSourceBase(100)
	setMarkedBytes(d, 0, len(d.text))

	if !d.indentBlock(2) {
		t.Fatal("Indent did not modify the marked lines")
	}
	if want := "    first\n  \tsecond\n  \n  last"; d.text != want {
		t.Fatalf("indented text = %q, want %q", d.text, want)
	}
	if got := d.sourcePositionAt(len(d.text)); got != 121 {
		t.Fatalf("source end after indent = %d, want 121", got)
	}
	if !d.unindentBlock(2) {
		t.Fatal("Unindent did not modify the marked lines")
	}
	if want := "  first\n      second\n\nlast"; d.text != want {
		t.Fatalf("unindented text = %q, want %q", d.text, want)
	}
	if got := d.sourcePositionAt(len(d.text)); got != 121 {
		t.Fatalf("source end after unindent = %d, want 121", got)
	}
}

func TestUnindentPartialBlockAndTabUsesSpaces(t *testing.T) {
	d := newDocument("before\n\tselected\nafter")
	setMarkedBytes(d, len("before\n\t"), len("before\n\tselected"))
	if !d.unindentBlock(1) {
		t.Fatal("Unindent did not modify the marked line")
	}
	if want := "before\n       selected\nafter"; d.text != want {
		t.Fatalf("partial-block unindent = %q, want %q", d.text, want)
	}
}

func TestBlockCopyMoveAndDelete(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("abcdef", false)
	d := editor.area.document
	setMarkedBytes(d, 1, 3)
	d.setCursor(d.positionAtByte(5), false)
	editor.openMenu('B')
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	if got, want := d.text, "abcdebcf"; got != want {
		t.Fatalf("Copy text = %q, want %q", got, want)
	}

	editor.area.SetText("abcdef", false)
	d = editor.area.document
	setMarkedBytes(d, 1, 3)
	d.setCursor(d.positionAtByte(5), false)
	move, _ := commandByID("block.move")
	editor.execute(move)
	if got, want := d.text, "adebcf"; got != want {
		t.Fatalf("Move text = %q, want %q", got, want)
	}
	if start, end, ok := d.markedBlock(); !ok || d.text[start:end] != "bc" {
		t.Fatal("Move did not leave markers around the moved text")
	}

	editor.area.SetText("abcdef", false)
	d = editor.area.document
	setMarkedBytes(d, 1, 3)
	d.setCursor(d.positionAtByte(3), false)
	copyCommand, _ := commandByID("block.copy")
	editor.execute(copyCommand)
	if got, want := d.text, "abcbcdef"; got != want {
		t.Fatalf("Copy at block end = %q, want %q", got, want)
	}
	if start, end, ok := d.markedBlock(); !ok || d.text[start:end] != "bc" {
		t.Fatal("Copy at the block edge expanded the source markers")
	}

	editor.area.SetText("abcdef", false)
	d = editor.area.document
	setMarkedBytes(d, 1, 3)
	d.setCursor(d.positionAtByte(2), false)
	editor.execute(copyCommand)
	if d.text != "abcdef" || editor.notice != "Copy destination is inside the marked block." {
		t.Fatal("Copy into the marked block must be a no-op")
	}
	d.setCursor(d.positionAtByte(1), false)
	editor.execute(move)
	if d.text != "abcdef" || editor.notice != "Move destination is inside or at the edge of the marked block." {
		t.Fatal("Move to the block edge must be a no-op")
	}

	editor.area.SetText("abcdef", false)
	setMarkedBytes(editor.area.document, 1, 3)
	deleteCommand, _ := commandByID("block.delete")
	editor.execute(deleteCommand)
	if len(editor.dialogs) != 1 || editor.area.GetText() != "abcdef" {
		t.Fatal("Delete must confirm before changing the document")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if editor.area.GetText() != "abcdef" {
		t.Fatal("Canceling Delete changed the document")
	}
	setMarkedBytes(editor.area.document, 1, 3)
	editor.execute(deleteCommand)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	if got, want := editor.area.GetText(), "adef"; got != want {
		t.Fatalf("confirmed Delete text = %q, want %q", got, want)
	}

	editor.area.SetText("abcdef", false)
	setMarkedBytes(editor.area.document, 1, 3)
	deleteDirectCommand, _ := commandByID("block.delete-direct")
	editor.execute(deleteDirectCommand)
	if got, want := editor.area.GetText(), "adef"; got != want {
		t.Fatalf("direct Delete text = %q, want %q", got, want)
	}
}

func TestMoveRejectsPartialFileAndWriteToCursorAfterMove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partial.prg")
	text := "abcdef" + strings.Repeat("x", fileChunkBytes)
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	absolute, loaded, err := readEditorFile(path)
	if err != nil {
		t.Fatal(err)
	}
	editor.setFile(absolute, loaded, true)
	d := editor.area.document
	setMarkedBytes(d, 1, 3)
	d.setCursor(d.positionAtByte(5), false)
	editor.moveBlock()
	if editor.area.GetText() != text[:fileChunkEnd(text, 0)] || d.sourceReordered {
		t.Fatal("Move changed a partially loaded file")
	}

	editor.setFile(absolute, "abcdef", true)
	d = editor.area.document
	setMarkedBytes(d, 1, 3)
	d.setCursor(d.positionAtByte(5), false)
	editor.moveBlock()
	if d.text != "adebcf" || !d.sourceReordered {
		t.Fatal("Move did not reorder the complete document")
	}
	editor.writeToCursor()
	if editor.notice != "Save the moved block before using Write to Cursor." {
		t.Fatal("Write to Cursor must reject a moved, unsaved buffer")
	}
}

func TestBlockOutputWritesOnlySelectionAndConfirmsOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "block.txt")
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.lastDir = dir
	editor.setFile("", "prefix\r\nselected\x00suffix", false)
	setMarkedBytes(editor.area.document, len("prefix\r\n"), len("prefix\r\nselected\x00"))
	command, _ := commandByID("block.output")
	editor.execute(command)
	if editor.dialogs[0].title != "Output block to file" {
		t.Fatal("Output to file did not open its destination selector")
	}
	editor.dialogs[0].widgets[editor.dialogs[0].fields[0]].value = filepath.Base(path)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if got, want := string(readFileBytes(t, path)), "selected\x00"; got != want {
		t.Fatalf("output bytes = %q, want %q", got, want)
	}
	if editor.area.GetText() != "prefix\r\nselected\x00suffix" || editor.changed {
		t.Fatal("Output changed the active document or its dirty state")
	}

	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	editor.execute(command)
	editor.dialogs[0].widgets[editor.dialogs[0].fields[0]].value = filepath.Base(path)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if len(editor.dialogs) != 2 || editor.dialogs[1].title != "Overwrite file?" {
		t.Fatal("Output must confirm before overwriting an existing file")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if len(editor.dialogs) != 1 || string(readFileBytes(t, path)) != "old" {
		t.Fatal("Canceling Output overwrite modified the existing file")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if len(editor.dialogs) != 2 || editor.dialogs[1].title != "Overwrite file?" {
		t.Fatal("Output selector could not retry after canceling overwrite")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	if got := string(readFileBytes(t, path)); got != "selected\x00" {
		t.Fatalf("confirmed output = %q", got)
	}
	if got := string(readFileBytes(t, filepath.Join(dir, "block.{TX"))); got != "old" {
		t.Fatalf("overwrite backup = %q, want %q", got, "old")
	}
	if editor.area.GetText() != "prefix\r\nselected\x00suffix" {
		t.Fatal("confirmed Output changed the active document")
	}
}

func TestBlockOutputRejectsCurrentDocumentPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.txt")
	if err := os.WriteFile(path, []byte("whole document"), 0600); err != nil {
		t.Fatal(err)
	}
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	absolute, text, err := readEditorFile(path)
	if err != nil {
		t.Fatal(err)
	}
	editor.setFile(absolute, text, true)
	setMarkedBytes(editor.area.document, 0, 5)
	command, _ := commandByID("block.output")
	editor.execute(command)
	editor.dialogs[0].widgets[editor.dialogs[0].fields[0]].value = path
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if len(editor.dialogs) != 2 || editor.dialogs[1].title != "File error" ||
		string(readFileBytes(t, path)) != text {
		t.Fatal("Output must not replace the active document file with the selected block")
	}
}

func TestBlockIndentCommandAppliesDialogAmount(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("first\nsecond", false)
	setMarkedBytes(editor.area.document, 0, len(editor.area.GetText()))
	command, _ := commandByID("block.indent")
	editor.execute(command)
	if editor.dialogs[0].title != "Indent..." {
		t.Fatal("Indent did not open the amount dialog")
	}
	editor.dialogs[0].widgets[editor.dialogs[0].fields[0]].value = "2"
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if got, want := editor.area.GetText(), "  first\n  second"; got != want {
		t.Fatalf("Indent result = %q, want %q", got, want)
	}
}

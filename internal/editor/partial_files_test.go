package edit

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestFileChunkEndPreservesCRLFAndUTF8Boundaries(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "CRLF",
			text: strings.Repeat("x", fileChunkBytes-1) + "\r\nend",
			want: fileChunkBytes - 1,
		},
		{
			name: "UTF-8",
			text: strings.Repeat("x", fileChunkBytes-1) + "€end",
			want: fileChunkBytes - 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := fileChunkEnd(test.text, 0); got != test.want {
				t.Fatalf("fileChunkEnd() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestPartialFileWriteLoadMoreAndReturnToTop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.prg")
	original := strings.Repeat("abcdefgh", fileChunkBytes*3/8+20)
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	absolute, text, err := readEditorFile(path)
	if err != nil {
		t.Fatal(err)
	}
	editor.setFile(absolute, text, true)
	if len(editor.area.GetText()) != fileChunkBytes || editor.fileBase != 0 {
		t.Fatalf("initial buffer length/base = %d/%d", len(editor.area.GetText()), editor.fileBase)
	}

	editor.area.notify(editor.area.document.replaceBytes(0, 1, "Z"))
	editor.area.document.setCursor(documentPosition{character: 8}, false)
	editor.writeToCursor()
	want := "Z" + original[1:]
	if got := string(readFileBytes(t, path)); got != want {
		t.Fatal("Write to Cursor did not preserve the edited prefix and untouched suffix")
	}
	if editor.area.GetText() != original[8:fileChunkBytes] || editor.fileBase != 8 || editor.changed {
		t.Fatalf("post-write buffer/base/dirty = %d bytes/%d/%v",
			len(editor.area.GetText()), editor.fileBase, editor.changed)
	}
	assertFileBytes(t, filepath.Join(dir, "large.{PR"), []byte(original))

	editor.area.notify(editor.area.document.replaceBytes(0, 1, "Y"))
	editor.loadMore()
	if got, wantLength := len(editor.area.GetText()), fileChunkBytes-8+fileChunkBytes; got != wantLength {
		t.Fatalf("after Load More buffer length = %d, want %d", got, wantLength)
	}
	if editor.area.GetText()[0] != 'Y' || !editor.changed {
		t.Fatal("Load More discarded an unsaved edit in the existing buffer")
	}
	editor.loadMore()
	editor.loadMore()
	loadedLength := len(editor.area.GetText())
	if loadedLength != len(editor.diskText)-editor.fileBase {
		t.Fatalf("final loaded buffer length = %d, want %d", loadedLength, len(editor.diskText)-editor.fileBase)
	}
	editor.loadMore()
	if len(editor.area.GetText()) != loadedLength || editor.notice != "The entire file is already loaded." {
		t.Fatal("Load More at EOF must not duplicate data")
	}
	editor.saveDocument(nil)
	want = "Z" + original[1:8] + "Y" + original[9:]
	if got := string(readFileBytes(t, path)); got != want {
		t.Fatal("Save after partial loading did not preserve the full file")
	}

	editor.area.notify(editor.area.document.replaceBytes(0, 1, "Q"))
	editor.topOfFile()
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Unsaved document" {
		t.Fatal("Top of File must resolve unsaved changes before reloading")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	if editor.fileBase != 0 || len(editor.area.GetText()) != fileChunkBytes ||
		editor.area.GetText()[0] != 'Z' {
		t.Fatal("Top of File did not reload the saved file from its beginning")
	}
}

func TestPartialFileOperationsRejectExternalChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.prg")
	original := strings.Repeat("x", fileChunkBytes+10)
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	absolute, text, err := readEditorFile(path)
	if err != nil {
		t.Fatal(err)
	}
	editor.setFile(absolute, text, true)
	before := editor.area.GetText()
	if err := os.WriteFile(path, []byte("external"), 0600); err != nil {
		t.Fatal(err)
	}
	editor.loadMore()
	if len(editor.dialogs) != 1 || editor.area.GetText() != before {
		t.Fatal("Load More must report external changes without altering the buffer")
	}
}

func TestWriteToCursorRejectsExternalChangesWithoutDataLoss(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.prg")
	original := strings.Repeat("x", fileChunkBytes+10)
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	absolute, text, err := readEditorFile(path)
	if err != nil {
		t.Fatal(err)
	}
	editor.setFile(absolute, text, true)
	editor.area.document.setCursor(documentPosition{character: 1}, false)
	if err := os.WriteFile(path, []byte("external"), 0600); err != nil {
		t.Fatal(err)
	}

	editor.writeToCursor()
	if len(editor.dialogs) != 1 || editor.area.GetText() != original[:fileChunkBytes] ||
		string(readFileBytes(t, path)) != "external" || editor.fileBase != 0 {
		t.Fatal("Write to Cursor must reject external changes without altering either copy")
	}
}

func TestMergeInsertsAtCursorAndReportsMissingFile(t *testing.T) {
	dir := t.TempDir()
	mergePath := filepath.Join(dir, "merge.prg")
	if err := os.WriteFile(mergePath, []byte{0xff, '\r', '\n'}, 0600); err != nil {
		t.Fatal(err)
	}
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("hello", false)
	editor.area.document.selectBytes(1, 4)
	command, _ := commandByID("file.merge")
	editor.execute(command)
	if editor.dialogs[0].title != "Merge file" {
		t.Fatal("Merge did not open its file selector")
	}
	editor.dialogs[0].widgets[editor.dialogs[0].fields[0]].value = mergePath
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if got, want := editor.area.GetText(), string([]byte{'h', 'e', 'l', 'l', 0xff, '\r', '\n', 'o'}); got != want {
		t.Fatalf("merged text = %v, want %v", []byte(got), []byte(want))
	}
	if !editor.changed || len(editor.dialogs) != 0 {
		t.Fatal("Merge did not leave the inserted bytes dirty or close its selector")
	}

	editor.execute(command)
	missing := filepath.Join(dir, "missing.prg")
	editor.dialogs[0].widgets[editor.dialogs[0].fields[0]].value = missing
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if len(editor.dialogs) != 2 || editor.dialogs[1].title == "Create file?" {
		t.Fatal("a missing Merge source must report a read error, not offer to create it")
	}
	if _, err := os.Stat(missing); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Merge created a missing source file: %v", err)
	}
}

func TestPowerShellCommandUsesRunnerAndSurfacesFailure(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	called := false
	editor.shellRunner = func() error {
		called = true
		return nil
	}
	command, _ := commandByID("file.shell")
	editor.execute(command)
	if !called || editor.notice != "Returned from PowerShell." {
		t.Fatal("PowerShell command did not run its configured runner")
	}

	editor.shellRunner = func() error { return errors.New("launch failed") }
	editor.execute(command)
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "File error" {
		t.Fatal("PowerShell errors must be shown to the user")
	}
}

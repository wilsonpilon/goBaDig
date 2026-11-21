package edit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestFileBrowserNavigatesAndOpensRawFile(t *testing.T) {
	root := t.TempDir()
	subdirectory := filepath.Join(root, "subdirectory")
	if err := os.Mkdir(subdirectory, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(subdirectory, "program.PRG")
	want := []byte{0xef, 0xbb, 0xbf, 'x', 0x00, 0xff, '\r', '\n', 0x1a}
	if err := os.WriteFile(path, want, 0600); err != nil {
		t.Fatal(err)
	}

	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.lastDir = root
	command, _ := commandByID("file.open")
	editor.execute(command)
	dialog := editor.dialogs[0]
	directories := &dialog.widgets[dialog.fields[2]]
	dialog.focus = 2
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	if directories.choices[directories.selected] != "[subdirectory]" {
		t.Fatalf("directory selection = %q", directories.choices[directories.selected])
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	files := &dialog.widgets[dialog.fields[1]]
	if files.label != subdirectory || len(files.choices) != 1 || files.choices[0] != "program.PRG" {
		t.Fatalf("directory navigation produced label=%q files=%v", files.label, files.choices)
	}
	dialog.focus = 1
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if editor.path != path || editor.area.GetText() != string(want) || editor.changed {
		t.Fatalf("opened file path=%q bytes=%v changed=%v", editor.path, []byte(editor.area.GetText()), editor.changed)
	}
}

func TestOpenInitialFileAndCreatePromptDoNotCreateUntilSave(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "existing.prg")
	raw := []byte{0xef, 0xbb, 0xbf, 'a', 0x00, 0xff, '\n'}
	if err := os.WriteFile(existing, raw, 0600); err != nil {
		t.Fatal(err)
	}
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	if err := editor.openInitialFile(existing); err != nil {
		t.Fatal(err)
	}
	if editor.path != existing || editor.area.GetText() != string(raw) || editor.changed {
		t.Fatal("CLI startup did not load the file byte-for-byte")
	}

	missing := filepath.Join(dir, "new.prg")
	if err := editor.openInitialFile(missing); err != nil {
		t.Fatal(err)
	}
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Create file?" {
		t.Fatal("missing CLI path must ask before creating a document")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if _, err := os.Stat(missing); !errors.Is(err, os.ErrNotExist) || editor.path != existing {
		t.Fatal("canceling Create changed the active file or created the missing path")
	}

	if err := editor.openInitialFile(missing); err != nil {
		t.Fatal(err)
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone))
	if editor.path != missing || editor.diskKnown || editor.area.GetText() != "" || editor.changed {
		t.Fatal("Create must select a new in-memory document without writing it")
	}
	if _, err := os.Stat(missing); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Create wrote the file before Save: %v", err)
	}
}

func TestSaveRequiresConfirmationAfterExternalChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "document.PRG")
	if err := os.WriteFile(path, []byte("opened"), 0600); err != nil {
		t.Fatal(err)
	}
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	absolute, text, err := readEditorFile(path)
	if err != nil {
		t.Fatal(err)
	}
	editor.setFile(absolute, text, true)
	editor.area.SetText("editor contents", true)
	external := []byte("external contents")
	if err := os.WriteFile(path, external, 0600); err != nil {
		t.Fatal(err)
	}

	command, _ := commandByID("file.save")
	editor.execute(command)
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "File changed on disk" {
		t.Fatal("external modification must require explicit overwrite confirmation")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if string(readFileBytes(t, path)) != string(external) || !editor.changed {
		t.Fatal("canceling overwrite modified the external file or cleared dirty state")
	}

	editor.execute(command)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	if got := readFileBytes(t, path); string(got) != "editor contents" {
		t.Fatalf("confirmed save wrote %q", got)
	}
	if got := readFileBytes(t, filepath.Join(dir, "document.{PR")); string(got) != string(external) {
		t.Fatalf("backup contains %q, want external contents %q", got, external)
	}
	if editor.changed || len(editor.dialogs) != 0 {
		t.Fatal("successful confirmed save did not clear dirty state and close the prompt")
	}
}

func TestNewAndOpenDoNotDiscardUntilConfirmed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "next.prg")
	if err := os.WriteFile(path, []byte("next document"), 0600); err != nil {
		t.Fatal(err)
	}
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.lastDir = dir
	editor.area.SetText("unsaved document", true)

	newCommand, _ := commandByID("file.new")
	editor.execute(newCommand)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	if editor.area.GetText() != "unsaved document" || !editor.changed {
		t.Fatal("canceling New discarded the current document")
	}

	openCommand, _ := commandByID("file.open")
	editor.execute(openCommand)
	openDialog := editor.dialogs[0]
	openDialog.widgets[openDialog.fields[0]].value = filepath.Base(path)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	if editor.area.GetText() != "unsaved document" || !editor.changed || editor.path != "" {
		t.Fatal("canceling Open discarded or replaced the current document")
	}

	editor.execute(openCommand)
	openDialog = editor.dialogs[0]
	openDialog.widgets[openDialog.fields[0]].value = filepath.Base(path)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	if editor.path != path || editor.area.GetText() != "next document" || editor.changed {
		t.Fatal("confirmed Open did not replace the old document with the selected file")
	}
}

func readFileBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

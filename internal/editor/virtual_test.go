package edit

import (
	"errors"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestVirtualDocumentEditsSavesAndConfirmsExit(t *testing.T) {
	closed := false
	editor, app, _ := newTestEditor(t, 80, 25, func() { closed = true })
	stored := string([]byte{0xef, 0xbb, 0xbf, '?', ' ', '\'', 'a', '\'', '\r', '\n', 0xff, 0x1a})
	saves := 0
	source := VirtualDocument{
		Path: `filespace:\code\example.prg`, Text: stored, Exists: true,
		Read: func() (string, error) { return stored, nil },
		Save: func(text string) error { stored = text; saves++; return nil },
	}
	if err := editor.OpenVirtualDocument(source); err != nil {
		t.Fatal(err)
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, '*', tcell.ModNone))
	if !editor.changed {
		t.Fatal("virtual editing did not mark the document changed")
	}
	editor.requestClose()
	if closed || len(editor.dialogs) != 1 || editor.dialogs[0].title != "Unsaved document" {
		t.Fatal("virtual exit must confirm unsaved changes")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	if !closed || saves != 1 || stored != "*"+source.Text {
		t.Fatalf("Save on exit: closed=%v saves=%d content=%q", closed, saves, stored)
	}
}

func TestVirtualDocumentNewSaveAndStaleFailure(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	stored := ""
	fail := false
	source := VirtualDocument{
		Path: `filespace:\new.cmd`,
		Read: func() (string, error) { return stored, nil },
		Save: func(text string) error {
			if fail {
				return errors.New("stale virtual document")
			}
			stored = text
			return nil
		},
	}
	if err := editor.OpenVirtualDocument(source); err != nil {
		t.Fatal(err)
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	editor.saveDocument(nil)
	if stored != "x" || !editor.diskKnown || editor.changed {
		t.Fatal("new virtual document was not saved through its callback")
	}
	fail = true
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))
	editor.saveDocument(nil)
	if stored != "x" || !editor.changed || editor.diskText != "x" {
		t.Fatal("failed virtual save changed the saved baseline or cleared edits")
	}
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "File error" {
		t.Fatal("stale virtual saves must fail without offering overwrite")
	}
}

func TestVirtualDocumentPartialFileUsesCallbacks(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	stored := strings.Repeat("x", fileChunkBytes) + "tail"
	reads := 0
	if err := editor.OpenVirtualDocument(VirtualDocument{
		Path: `filespace:\large.prg`, Text: stored, Exists: true,
		Read: func() (string, error) { reads++; return stored, nil },
		Save: func(text string) error { stored = text; return nil },
	}); err != nil {
		t.Fatal(err)
	}
	editor.area.document.cursor = documentPosition{line: 0, character: 1}
	editor.writeToCursor()
	if reads != 1 || editor.fileBase != 1 || len(stored) != fileChunkBytes+4 {
		t.Fatalf("virtual write-to-cursor used wrong storage: reads=%d base=%d", reads, editor.fileBase)
	}
	editor.loadMore()
	if reads != 2 || !strings.HasSuffix(editor.area.GetText(), "tail") {
		t.Fatal("virtual load-more did not read the remaining stored bytes")
	}
	editor.reloadTopOfFile()
	if reads != 3 || editor.fileBase != 0 {
		t.Fatal("virtual top-of-file did not use the reader")
	}
}

func TestVirtualDocumentBlocksHostCommandsAndDestinations(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	saves := 0
	if err := editor.OpenVirtualDocument(VirtualDocument{
		Path: `filespace:\only.prg`,
		Read: func() (string, error) { return "", nil },
		Save: func(string) error { saves++; return nil },
	}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{
		"file.new", "file.open", "file.save-as", "file.merge", "file.window",
		"file.shell", "block.output", "print.pdf", "print.document", "options.save",
	} {
		command, ok := commandByID(id)
		if !ok {
			t.Fatalf("unknown command %q", id)
		}
		if editor.available(command) {
			t.Errorf("%s is available in a virtual session", id)
		}
		editor.execute(command)
		if len(editor.dialogs) != 0 || editor.path != `filespace:\only.prg` {
			t.Fatalf("%s escaped the virtual session", id)
		}
	}
	if err := editor.writeFile(`C:\host.prg`, "host", nil, true); err == nil || saves != 0 {
		t.Fatal("virtual Save As reached storage")
	}
	if err := editor.openInitialFile(`C:\host.prg`); err == nil {
		t.Fatal("virtual session permitted a host file to be opened")
	}
	if err := editor.runPowerShell(); err == nil {
		t.Fatal("virtual session permitted a host shell")
	}
}

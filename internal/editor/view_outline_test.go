package edit

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestR20bCompactOutlineLinesRemoveBlankLinesAndIndentation(t *testing.T) {
	got := compactOutlineLines(" \n  first  \r\n\tsecond\t \n\n  third\x01\xff\n")
	want := []string{"first", "second", "third\uFFFD\uFFFD"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("compactOutlineLines() = %#v, want %#v", got, want)
	}
}

func TestR20bOutlineIsTemporaryAndAnyKeyRestoresDocumentView(t *testing.T) {
	editor, _, screen := newTestEditor(t, 80, 25, nil)
	text := "  first line\n\n\tsecond line  \n"
	editor.area.SetText(text, false)
	editor.area.document.cursor = documentPosition{line: 2, character: 3}
	editor.changed = true
	revision := editor.area.document.revision
	cursor := editor.area.document.cursor
	command, ok := commandByID("view.outline")
	if !ok || !editor.available(command) {
		t.Fatal("Outline display command is unavailable")
	}

	for _, event := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone),
	} {
		editor.execute(command)
		if editor.outlineView == nil || len(editor.dialogs) != 0 {
			t.Fatal("Outline did not open as a temporary document view")
		}
		if got, want := editor.outlineView.lines, []string{"first line", "second line"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("outline lines = %#v, want %#v", got, want)
		}
		drawEditor(editor, screen)
		x, y, _, _ := editor.area.GetInnerRect()
		first, _, _, _ := screen.GetContent(x, y)
		second, _, _, _ := screen.GetContent(x, y+1)
		if first != 'f' || second != 's' {
			t.Fatalf("outline left-aligned rows start with %q and %q, want 'f' and 's'", first, second)
		}

		if got := editor.Capture(event); got != nil {
			t.Fatal("Outline key was not consumed")
		}
		if editor.outlineView != nil || editor.app.GetFocus() != editor.area {
			t.Fatal("key did not restore the normal document view")
		}
		if editor.area.GetText() != text || editor.area.document.revision != revision ||
			editor.area.document.cursor != cursor || !editor.changed {
			t.Fatal("Outline changed the document, cursor, or modified state")
		}
	}
}

func TestR20bOutlineShowsOnlyLoadedPartOfPartialFile(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.path = filepath.Join(t.TempDir(), "partial.prg")
	editor.diskText, editor.diskKnown = "loaded\nunloaded", true
	editor.area.SetText("loaded\n", false)
	editor.area.document.setSourceBase(0)

	command, _ := commandByID("view.outline")
	editor.execute(command)
	if got, want := editor.outlineView.lines, []string{"loaded"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("partial outline = %#v, want %#v", got, want)
	}
	if editor.notice != "Outline view (loaded portion only) - press any key to return." {
		t.Fatalf("partial outline notice = %q", editor.notice)
	}
}

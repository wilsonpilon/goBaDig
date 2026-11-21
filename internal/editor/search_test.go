package edit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestFindTextDirectionCaseAndUnicode(t *testing.T) {
	text := "Alpha alpha\nCAFÉ café"
	tests := []struct {
		name      string
		query     string
		cursor    int
		backwards bool
		matchCase bool
		wantStart int
		wantEnd   int
		wantFound bool
	}{
		{name: "forward case-insensitive", query: "alpha", cursor: 0, wantStart: 0, wantEnd: 5, wantFound: true},
		{name: "forward case-sensitive", query: "alpha", cursor: 0, matchCase: true, wantStart: 6, wantEnd: 11, wantFound: true},
		{name: "backward case-insensitive", query: "alpha", cursor: 11, backwards: true, wantStart: 6, wantEnd: 11, wantFound: true},
		{name: "backward stops before cursor", query: "Alpha", cursor: 4, backwards: true, matchCase: true},
		{name: "Unicode case folding", query: "café", cursor: 12, wantStart: 12, wantEnd: 17, wantFound: true},
		{name: "no wrap forward", query: "Alpha", cursor: len(text), matchCase: true},
		{name: "empty query", cursor: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start, end, found := findText(text, test.query, test.cursor, test.backwards, test.matchCase)
			if found != test.wantFound || found && (start != test.wantStart || end != test.wantEnd) {
				t.Fatalf("findText = (%d, %d, %v); want (%d, %d, %v)", start, end, found, test.wantStart, test.wantEnd, test.wantFound)
			}
		})
	}
}

func TestFindDialogFindsAndPreservesBlockMarkers(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("one two one", false)
	d := editor.area.document
	d.setBlockMarkerAtCursor()
	d.setCursor(d.positionAtByte(8), false)
	d.setBlockMarkerAtCursor()
	d.setCursor(d.positionAtByte(len(d.text)), false)

	command, _ := commandByID("search.find")
	editor.execute(command)
	dialog := editor.dialogs[0]
	for i := range dialog.widgets {
		switch dialog.widgets[i].id {
		case "find":
			dialog.widgets[i].value = "one"
		case "case", "backwards":
			dialog.widgets[i].checked = true
		}
	}
	sendEditorKey(editor, app, keyEnter())

	if len(editor.dialogs) != 0 || d.byteOffset(d.cursor) != 8 {
		t.Fatalf("Find did not move to the match: cursor=%d dialogs=%d", d.byteOffset(d.cursor), len(editor.dialogs))
	}
	if !d.searchMatch || d.searchMatchStart != 8 || d.searchMatchEnd != 11 {
		t.Fatalf("Find highlight = %d..%d, active=%v", d.searchMatchStart, d.searchMatchEnd, d.searchMatch)
	}
	if d.blockMarkerCount != 2 || d.text != "one two one" || !editor.lastSearch.valid ||
		!editor.lastSearch.matchCase || !editor.lastSearch.backwards {
		t.Fatal("Find changed document or block markers, or failed to retain search parameters")
	}
}

func TestFindCursorBoundaryControlsSubsequentTyping(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		query      string
		cursor     int
		backwards  bool
		wantCursor int
		wantText   string
	}{
		{"forward beta", "alpha beta alpha\r\n", "beta", 0, false, 10, "alpha betax alpha\r\n"},
		{"forward alpha", "alpha beta alpha\r\n", "alpha", 0, false, 5, "alphax beta alpha\r\n"},
		{"backward alpha", "alpha beta alpha\r\n", "alpha", 11, true, 0, "xalpha beta alpha\r\n"},
		{"Unicode folded match", "CAFÉ café", "café", 0, false, 5, "CAFÉx café"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			editor, app, _ := newTestEditor(t, 80, 25, nil)
			editor.area.SetText(test.text, false)
			editor.changed = false
			d := editor.area.document
			d.setCursor(d.positionAtByte(test.cursor), false)
			command, _ := commandByID("search.find")
			editor.execute(command)
			dialog := editor.dialogs[0]
			setDialogValue(dialog, "find", test.query)
			for i := range dialog.widgets {
				if dialog.widgets[i].id == "backwards" {
					dialog.widgets[i].checked = test.backwards
				}
			}
			sendEditorKey(editor, app, keyEnter())
			if got := d.byteOffset(d.cursor); got != test.wantCursor {
				t.Fatalf("Find cursor=%d, want %d", got, test.wantCursor)
			}
			if !d.searchMatch || editor.area.GetText() != test.text || editor.changed {
				t.Fatal("Find lost its highlight or modified the document")
			}
			sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
			if got := editor.area.GetText(); got != test.wantText {
				t.Fatalf("typing after Find=%q, want %q", got, test.wantText)
			}
		})
	}
}

func TestSearchDirectionChangesMatchDOSBoxReference(t *testing.T) {
	const text = "alpha beta alpha second beta line\r\n"
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText(text, false)
	editor.changed = false
	command, _ := commandByID("search.find")
	editor.execute(command)
	setDialogValue(editor.dialogs[0], "find", "alpha")
	sendEditorKey(editor, app, keyEnter())
	d := editor.area.document
	if got := d.byteOffset(d.cursor); got != 5 {
		t.Fatalf("Find cursor=%d, want 5", got)
	}
	for _, step := range []struct {
		id         string
		wantCursor int
	}{
		{"search.forward", 16},
		{"search.reverse", 11},
		{"search.reverse", 0},
		{"search.forward", 5},
	} {
		command, _ := commandByID(step.id)
		editor.execute(command)
		if got := d.byteOffset(d.cursor); got != step.wantCursor {
			t.Fatalf("%s cursor=%d, want %d", step.id, got, step.wantCursor)
		}
		if d.text != text || !d.searchMatch || editor.changed {
			t.Fatalf("%s changed the text or lost the match", step.id)
		}
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	path := filepath.Join(t.TempDir(), "wibchk1.txt")
	command, _ = commandByID("file.save-as")
	editor.execute(command)
	setDialogValue(editor.dialogs[0], "filename", path)
	sendEditorKey(editor, app, keyEnter())
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "alphax beta alpha second beta line\r\n"; string(got) != want {
		t.Fatalf("saved bytes=%q, want NE reference %q", got, want)
	}
}

func TestFindDialogReportsEmptyAndMissingText(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	command, _ := commandByID("search.find")
	editor.execute(command)
	sendEditorKey(editor, app, keyEnter())
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Search error" {
		t.Fatal("empty Find text was not reported")
	}
	sendEditorKey(editor, app, keyEscape())

	editor.area.SetText("present", false)
	editor.execute(command)
	for i := range editor.dialogs[0].widgets {
		if editor.dialogs[0].widgets[i].id == "find" {
			editor.dialogs[0].widgets[i].value = "missing"
		}
	}
	sendEditorKey(editor, app, keyEnter())
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Search" {
		t.Fatal("not-found result was not reported")
	}
}

func TestFindReplaceCanSkipAndReplaceAllWithoutLooping(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("cat cat", false)
	command, _ := commandByID("search.replace")
	editor.execute(command)
	setDialogValue(editor.dialogs[0], "find", "cat")
	setDialogValue(editor.dialogs[0], "replace", "dog")
	sendEditorKey(editor, app, keyEnter())

	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Replace text?" {
		t.Fatal("Find and Replace did not request confirmation")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Replace text?" ||
		editor.area.document.byteOffset(editor.area.document.cursor) != 4 {
		t.Fatal("Skip did not advance to the next occurrence")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone))
	if editor.area.GetText() != "cat dog" || editor.replaceSession != nil {
		t.Fatalf("single replacement result = %q, session=%v", editor.area.GetText(), editor.replaceSession)
	}

	editor.area.SetText("a a", false)
	editor.execute(command)
	setDialogValue(editor.dialogs[0], "find", "a")
	setDialogValue(editor.dialogs[0], "replace", "a!")
	sendEditorKey(editor, app, keyEnter())
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	if editor.area.GetText() != "a! a!" || editor.replaceSession != nil {
		t.Fatalf("Replace All result = %q, session=%v", editor.area.GetText(), editor.replaceSession)
	}
}

func TestContinueForwardStartsAfterMatchAndDoesNotWrap(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("a A a", false)
	editor.lastSearch = searchState{text: "a", matchCase: true, valid: true, start: 0, end: 1}
	editor.area.document.setCursor(editor.area.document.positionAtByte(1), false)
	editor.area.document.setSearchMatch(0, 1)

	command, _ := commandByID("search.forward")
	editor.execute(command)
	if editor.area.document.byteOffset(editor.area.document.cursor) != 5 ||
		editor.lastSearch.start != 4 || !editor.area.document.searchMatch {
		t.Fatal("Continue Forward did not find the next case-sensitive match")
	}
	editor.execute(command)
	if !strings.Contains(editor.notice, "No more matches forward") ||
		editor.area.document.byteOffset(editor.area.document.cursor) != 5 {
		t.Fatal("Continue Forward wrapped or moved the cursor when no match remained")
	}
}

func TestContinueReverseStartsBeforeMatchAndDoesNotWrap(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("a A a", false)
	editor.lastSearch = searchState{text: "a", matchCase: true, valid: true, start: 4, end: 5}
	editor.area.document.setCursor(editor.area.document.positionAtByte(4), false)
	editor.area.document.setSearchMatch(4, 5)

	command, _ := commandByID("search.reverse")
	editor.execute(command)
	if editor.area.document.byteOffset(editor.area.document.cursor) != 0 ||
		editor.lastSearch.start != 0 || !editor.area.document.searchMatch {
		t.Fatal("Continue Reverse did not find the previous case-sensitive match")
	}
	editor.execute(command)
	if !strings.Contains(editor.notice, "No more matches backward") ||
		editor.area.document.byteOffset(editor.area.document.cursor) != 0 {
		t.Fatal("Continue Reverse wrapped or moved the cursor when no match remained")
	}
}

func TestSearchDialogRestoresLastParameters(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.lastSearch = searchState{
		text:        "needle",
		replacement: "thread",
		matchCase:   true,
		backwards:   true,
		valid:       true,
	}
	command, _ := commandByID("search.replace")
	editor.execute(command)
	d := editor.dialogs[0]
	if d.value("find") != "needle" || d.value("replace") != "thread" ||
		!d.checked("case") || !d.checked("backwards") {
		t.Fatal("Find and Replace dialog did not restore the last search parameters")
	}
}

func TestGoToLineMovesAndRejectsOutOfRange(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("first\n\nthird", false)
	command, _ := commandByID("search.line")
	editor.execute(command)
	setDialogValue(editor.dialogs[0], "number", "3")
	sendEditorKey(editor, app, keyEnter())
	if editor.area.document.cursor.line != 2 || editor.area.document.cursor.character != 0 {
		t.Fatal("Go To Line did not move to the requested physical line")
	}

	editor.execute(command)
	setDialogValue(editor.dialogs[0], "number", "4")
	sendEditorKey(editor, app, keyEnter())
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Line number" ||
		editor.area.document.cursor.line != 2 {
		t.Fatal("out-of-range line number was not reported without moving the cursor")
	}
}

func TestMatchingBracketHandlesNestingAndMissingPairs(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("({[x]})", false)
	if !editor.matchingBracket() || editor.area.document.byteOffset(editor.area.document.cursor) != 6 {
		t.Fatal("outer opening bracket did not find its matching close")
	}
	editor.area.document.setCursor(editor.area.document.positionAtByte(4), false)
	if !editor.matchingBracket() || editor.area.document.byteOffset(editor.area.document.cursor) != 2 {
		t.Fatal("nested closing bracket did not find its matching open")
	}

	editor.area.SetText("(unclosed", false)
	if editor.matchingBracket() || len(editor.dialogs) != 1 ||
		editor.area.document.byteOffset(editor.area.document.cursor) != 0 {
		t.Fatal("unmatched opening bracket was not reported without moving the cursor")
	}
}

func TestWindowDifferencesMovesBothWindowsFromTheirCursors(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("abcX", false)
	editor.openSecondWindow("second", "abcY", false)
	editor.activateWindow(0, false)

	command, _ := commandByID("search.differences")
	editor.execute(command)
	if editor.area.document.byteOffset(editor.area.document.cursor) != 3 ||
		editor.windows[1].area.document.byteOffset(editor.windows[1].area.document.cursor) != 3 {
		t.Fatal("Window Differences did not move both cursors to the first differing byte")
	}

	editor.windows[1].area.SetText("abcX", false)
	editor.area.document.setCursor(documentPosition{}, false)
	editor.windows[1].area.document.setCursor(documentPosition{}, false)
	editor.execute(command)
	if !strings.Contains(editor.notice, "No differences found") {
		t.Fatal("equal window tails were not reported")
	}
}

func TestTextInBlockUsesMarkersAndPreservesThem(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("needle rest needle", false)
	d := editor.area.document
	d.setBlockMarkerAtCursor()
	d.setCursor(d.positionAtByte(6), false)
	d.setBlockMarkerAtCursor()
	d.setCursor(d.positionAtByte(7), false)
	editor.lastSearch = searchState{matchCase: true, valid: true}

	command, _ := commandByID("search.block")
	editor.execute(command)
	if d.byteOffset(d.cursor) != 18 || d.blockMarkerCount != 2 ||
		!editor.lastSearch.matchCase || editor.lastSearch.text != "needle" {
		t.Fatal("Text in Block failed to find the marked text or changed its markers/options")
	}
}

func TestSearchBlockMarkerCyclesAsItsOwnSearchCommand(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("a b", false)
	d := editor.area.document

	command, _ := commandByID("search.marker")
	editor.execute(command)
	if !strings.Contains(editor.notice, "No block markers are set [search.marker]") {
		t.Fatal("Search > Block Marker did not report the empty-marker case")
	}

	d.setBlockMarkerAtCursor()
	d.setCursor(d.positionAtByte(2), false)
	editor.execute(command)
	if d.byteOffset(d.cursor) != 0 {
		t.Fatal("Search > Block Marker did not return to the sole marker")
	}
	editor.execute(command)
	if d.byteOffset(d.cursor) != 0 {
		t.Fatal("Search > Block Marker did not remain on the sole marker")
	}

	d.setCursor(d.positionAtByte(2), false)
	d.setBlockMarkerAtCursor()
	d.setCursor(documentPosition{}, false)

	editor.execute(command)
	if d.byteOffset(d.cursor) != 0 || !strings.Contains(editor.notice, "[search.marker]") {
		t.Fatal("Search > Block Marker did not move to the first marker with distinct command feedback")
	}
	editor.execute(command)
	if d.byteOffset(d.cursor) != 2 {
		t.Fatal("Search > Block Marker did not cycle to the second marker")
	}
}

func setDialogValue(dialog *editorDialog, id, value string) {
	for i := range dialog.widgets {
		if dialog.widgets[i].id == id {
			dialog.widgets[i].value = value
			return
		}
	}
}

func keyEnter() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
}

func keyEscape() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
}

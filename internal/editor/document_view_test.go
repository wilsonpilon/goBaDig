package edit

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestDocumentViewNavigationKeysAndPhysicalStatus(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	text := "a\tb\r\n" + strings.Repeat("x", 120) + "\nshort"
	editor.area.SetText(text, false)
	editor.changed = false
	revision := editor.area.document.revision
	for _, step := range []struct {
		key      tcell.Key
		mod      tcell.ModMask
		position documentPosition
		column   int
	}{
		{tcell.KeyRight, tcell.ModNone, documentPosition{0, 1}, 1},
		{tcell.KeyRight, tcell.ModNone, documentPosition{0, 2}, 8},
		{tcell.KeyDown, tcell.ModNone, documentPosition{1, 8}, 8},
		{tcell.KeyDown, tcell.ModNone, documentPosition{2, 5}, 5},
		{tcell.KeyUp, tcell.ModNone, documentPosition{1, 8}, 8},
		{tcell.KeyHome, tcell.ModCtrl, documentPosition{}, 0},
		{tcell.KeyRight, tcell.ModCtrl, documentPosition{0, 2}, 8},
		{tcell.KeyLeft, tcell.ModCtrl, documentPosition{}, 0},
		{tcell.KeyEnd, tcell.ModCtrl, documentPosition{2, 5}, 5},
		{tcell.KeyHome, tcell.ModNone, documentPosition{2, 0}, 0},
		{tcell.KeyUp, tcell.ModNone, documentPosition{1, 0}, 0},
		{tcell.KeyEnd, tcell.ModNone, documentPosition{1, 120}, 120},
	} {
		sendEditorKey(editor, app, tcell.NewEventKey(step.key, 0, step.mod))
		d := editor.area.document
		if d.cursor != step.position || d.column(d.cursor) != step.column {
			t.Fatalf("%s: cursor=%+v column=%d", keyName(tcell.NewEventKey(step.key, 0, step.mod)), d.cursor, d.column(d.cursor))
		}
		drawEditor(editor, screen)
		if d.text != text || d.revision != revision || editor.changed {
			t.Fatal("navigation changed text, revision or dirty state")
		}
	}
	if row := editorScreenRow(screen, 24); !strings.Contains(row, "Line=2") || !strings.Contains(row, "Col=121") {
		t.Fatalf("status lost physical line/visual column: %q", row)
	}
	x, y, visible := screen.GetCursor()
	if !visible || x != 78 || y != 2 || editor.area.document.left != 42 {
		t.Fatalf("horizontal reveal or cursor rendering incorrect: cursor=%d,%d visible=%v left=%d", x, y, visible, editor.area.document.left)
	}
	editor.area.SetText(strings.Repeat("x", 100000), true)
	drawEditor(editor, screen)
	if row := editorScreenRow(screen, 24); !strings.Contains(row, "Col=100001") {
		t.Fatalf("large columns must not be truncated in status: %q", row)
	}
}

func TestDocumentViewScrollAndResizeKeepBytes(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	text := strings.Repeat("abcdefghij\n", 99) + "last"
	editor.area.SetText(text, false)
	editor.changed = false
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone))
	if d := editor.area.document; d.top != 22 || d.cursor != (documentPosition{22, 0}) {
		t.Fatal("PgDn must route through document navigation")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModCtrl))
	if editor.area.document.cursor.line != 44 {
		t.Fatal("Ctrl-PgDn must move to the bottom of the current viewport")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModCtrl))
	if editor.area.document.cursor.line != 22 {
		t.Fatal("Ctrl-PgUp must move to the top of the current viewport")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModAlt))
	if d := editor.area.document; d.top != 23 || d.cursor.line != 23 {
		t.Fatal("Alt-Z must scroll upwards, clamping the cursor if needed")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyCtrlZ, 0, tcell.ModCtrl))
	if d := editor.area.document; d.top != 22 || d.cursor.line != 23 || d.text != text {
		t.Fatal("Ctrl-Z must scroll down, never undo")
	}
	editor.area.SetOffset(60, 0)
	drawEditor(editor, screen)
	if _, _, visible := screen.GetCursor(); visible {
		t.Fatal("manual scrolling may temporarily hide the cursor")
	}
	editor.showHelp("cursor.page-down")
	drawEditor(editor, screen)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	drawEditor(editor, screen)
	if d := editor.area.document; d.top != 60 || d.cursor.line != 23 || app.GetFocus() != editor.area {
		t.Fatal("help focus restoration must not snap a manually scrolled viewport")
	}
	editor.area.SetOffset(22, 0)
	before := editor.area.document.cursor
	screen.SetSize(25, 8)
	drawEditor(editor, screen)
	if d := editor.area.document; d.width != 24 || d.height != 6 || d.cursor != before || d.text != text || editor.changed {
		t.Fatal("resize must update viewport, not document/cursor/dirty state")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone))
	if editor.area.document.cursor.line != before.line+5 {
		t.Fatal("paging must use the resized viewport")
	}
}

func TestDocumentViewRendersTabsControlsAndUnicode(t *testing.T) {
	editor, _, screen := newTestEditor(t, 80, 25, nil)
	text := "a\t界e\u0301\x00\x1b\x7f\u0085"
	editor.area.SetText(text, false)
	drawEditor(editor, screen)
	for _, expected := range []struct {
		x  int
		ch rune
	}{{0, 'a'}, {7, ' '}, {8, '界'}, {10, 'e'}, {11, '\u2400'}, {12, '\u241b'}, {13, '\u2421'}, {14, '\ufffd'}} {
		ch, combining, _, _ := screen.GetContent(expected.x, 1)
		if ch != expected.ch {
			t.Fatalf("column %d: rendered %q, want %q", expected.x, ch, expected.ch)
		}
		if expected.ch == 'e' && (len(combining) != 1 || combining[0] != '\u0301') {
			t.Fatal("combining marks must remain attached to the displayed grapheme")
		}
	}
	if editor.area.GetText() != text {
		t.Fatal("displaying control pictures must not rewrite control bytes")
	}
}

func TestDocumentViewWrapMouseAndActiveSelectionCursor(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("first\nsecond", false)
	editor.area.Select(10, 1)
	drawEditor(editor, screen)
	if row := editorScreenRow(screen, 24); !strings.Contains(row, "Line=1") || !strings.Contains(row, "Col=2") {
		t.Fatal("status must follow the active end of a backwards selection")
	}
	editor.area.SetText("a\t界e\u0301", false)
	drawEditor(editor, screen)
	editor.MouseHandler()(tview.MouseLeftClick, tcell.NewEventMouse(10, 1, tcell.Button1, tcell.ModNone), func(p tview.Primitive) { app.SetFocus(p) })
	if d := editor.area.document; d.cursor.character != 3 || d.byteOffset(d.cursor) != 5 || d.column(d.cursor) != 10 {
		t.Fatal("mouse must map terminal columns to grapheme and byte boundaries")
	}
	text := strings.Repeat("word ", 30)
	editor.area.SetText(text, true)
	editor.changed = false
	editor.openMenu('T')
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))
	drawEditor(editor, screen)
	if d := editor.area.document; !d.wrap || d.text != text || d.cursor.line != 0 || editor.changed {
		t.Fatal("display wrap must not change physical lines or bytes")
	}
	if row := editorScreenRow(screen, 24); !strings.Contains(row, "Line=1") || !strings.Contains(row, "Col=151") {
		t.Fatalf("wrapped status must still show physical coordinates: %q", row)
	}
	x, y, visible := screen.GetCursor()
	if !visible || x < 0 || x >= 79 || y < 1 || y >= 24 {
		t.Fatal("wrapped cursor must remain inside the text viewport")
	}
}

func TestDocumentViewBlockCommandsAndHighlight(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("first\r\nsecond", false)
	d := editor.area.document
	for _, id := range []string{
		"block.set", "block.remove", "block.find", "block.mark-line", "block.mark-end", "search.marker",
	} {
		command, ok := commandByID(id)
		if !ok || !editor.available(command) {
			t.Fatalf("%s must be available in R09", id)
		}
	}

	d.setCursor(documentPosition{1, 2}, false)
	editor.openMenu('B')
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	if selected, _, _ := editor.area.GetSelection(); selected != "second" {
		t.Fatalf("Mark line selected %q, want %q", selected, "second")
	}
	drawEditor(editor, screen)
	startMarker, _, startStyle, _ := screen.GetContent(0, 2)
	endMarker, _, endStyle, _ := screen.GetContent(1, 2)
	_, _, markedStyle, _ := screen.GetContent(2, 2)
	if startMarker != '\u25aa' || endMarker != '\u25aa' ||
		startStyle != markerStyle ||
		endStyle != startStyle || markedStyle != blockStyle {
		t.Fatal("block markers and marked text must be visibly distinct")
	}

	editor.openMenu('B')
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	if selected, _, _ := editor.area.GetSelection(); selected != "cond" {
		t.Fatalf("Mark to end selected %q, want %q", selected, "cond")
	}

	remove, _ := commandByID("block.remove")
	editor.execute(remove)
	d.setCursor(documentPosition{0, 1}, false)
	set, _ := commandByID("block.set")
	editor.execute(set)
	d.setCursor(documentPosition{1, 2}, false)
	editor.execute(set)
	if selected, _, _ := editor.area.GetSelection(); selected != "irst\r\nse" {
		t.Fatalf("Set marker selected %q, want %q", selected, "irst\r\nse")
	}
	find, _ := commandByID("search.marker")
	editor.execute(find)
	if d.byteOffset(d.cursor) != 1 {
		t.Fatal("Search > Block Marker did not find the first boundary")
	}
	find, _ = commandByID("block.find")
	editor.execute(find)
	if d.byteOffset(d.cursor) != 9 {
		t.Fatal("Find marker did not advance to the second boundary")
	}
	editor.execute(remove)
	if editor.area.HasSelection() {
		t.Fatal("Remove markers left a marked selection")
	}
}

func TestTypingAtCursorDoesNotReplaceMarkedBlock(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	text := "zero\r\nblock\r\nend"
	editor.area.SetText(text, false)
	d := editor.area.document
	setMarkedBytes(d, len("zero\r\n"), len("zero\r\nblock"))
	d.setCursor(d.positionAtByte(len(text)-1), false)

	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	if want := "zero\r\nblock\r\nen\r\nd"; d.text != want {
		t.Fatalf("Enter at cursor changed the wrong text: got %q, want %q", d.text, want)
	}
	start, end, ok := d.markedBlock()
	if !ok || d.text[start:end] != "block" {
		t.Fatal("typing at the cursor removed or changed the marked block")
	}
}

func TestDocumentViewBasicEditingPreserved(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	for _, ch := range "ab" {
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, ch, tcell.ModNone))
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, '界', tcell.ModNone))
	if editor.area.GetText() != "ab\r\n\t界" {
		t.Fatal("model migration lost prototype typing, Enter or Tab")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone))
	if editor.area.GetText() != "ab" {
		t.Fatal("backspace must remove graphemes and join lines")
	}
	editor.area.Select(0, 2)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'X', tcell.ModNone))
	if editor.area.GetText() != "X" {
		t.Fatal("typing must still replace a basic selection")
	}
	editor.area.SetText("e\u0301界", false)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone))
	if editor.area.GetText() != "界" {
		t.Fatal("Delete split a grapheme")
	}
	editor.area.SetText("\u0301", false)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	if d := editor.area.document; d.text != "e\u0301" || d.byteOffset(d.cursor) != len(d.text) {
		t.Fatal("insertion must place the cursor after a newly joined grapheme")
	}
	drawEditor(editor, screen)
}

func TestDocumentMouseDragIsCapturedOnlyFromDocument(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	text := "abcdef"
	editor.area.SetText(text, false)
	drawEditor(editor, screen)
	handler := editor.area.MouseHandler()
	focus := func(p tview.Primitive) { app.SetFocus(p) }
	consumed, _ := handler(tview.MouseMove, tcell.NewEventMouse(4, 0, tcell.Button1, tcell.ModNone), focus)
	if consumed || editor.area.document.cursor != (documentPosition{}) {
		t.Fatal("foreign mouse drags must not move the document cursor")
	}

	_, capture := handler(tview.MouseLeftDown, tcell.NewEventMouse(5, 1, tcell.Button1, tcell.ModNone), focus)
	if capture != editor.area {
		t.Fatal("document drag must retain capture")
	}
	handler(tview.MouseMove, tcell.NewEventMouse(1, 1, tcell.Button1, tcell.ModNone), focus)
	handler(tview.MouseLeftUp, tcell.NewEventMouse(1, 1, tcell.ButtonNone, tcell.ModNone), focus)
	handler(tview.MouseLeftClick, tcell.NewEventMouse(1, 1, tcell.ButtonNone, tcell.ModNone), focus)
	selected, _, _ := editor.area.GetSelection()
	if selected != "bcde" || editor.area.document.cursor.character != 1 || editor.area.dragging {
		t.Fatal("backwards mouse selection lost its active endpoint or capture")
	}
	if editor.area.document.blockMarkerCount != 2 {
		t.Fatal("mouse drag did not create persistent block markers")
	}
	editor.area.navigate("cursor.right")
	if selected, _, _ = editor.area.GetSelection(); selected != "bcde" {
		t.Fatal("mouse-marked block did not remain selected during navigation")
	}
	editor.area.navigate("cursor.file-end")
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if editor.area.GetText() != text+"\r\n" {
		t.Fatal("typing after a mouse-marked block must insert at the cursor")
	}
	if selected, _, _ = editor.area.GetSelection(); selected != "bcde" {
		t.Fatal("typing at the cursor removed the mouse-marked block")
	}
}

func TestDocumentNavigationThroughApplicationEvents(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	text := "a\tb\r\n" + strings.Repeat("xxxxxxxxxx\n", 30)
	editor.area.SetText(text, false)
	editor.changed = false
	for _, event := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModCtrl),
		tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModCtrl),
		tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModAlt),
	} {
		app.QueueEvent(event)
	}
	rendered := make(chan struct{}, 1)
	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		if d := editor.area.document; d.top == 1 && len(editor.dialogs) == 0 &&
			d.cursor == (documentPosition{1, 8}) && d.text == text && !editor.changed &&
			strings.Contains(editorScreenRow(screen, 24), "Col=9") {
			select {
			case rendered <- struct{}{}:
			default:
			}
		}
	})
	result := make(chan error, 1)
	go func() { result <- app.Run() }()
	t.Cleanup(app.Stop)
	select {
	case <-rendered:
	case err := <-result:
		t.Fatalf("application stopped before rendering navigation: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("application did not render navigation/help/scroll sequence")
	}
	app.Stop()
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("application did not shut down")
	}
}

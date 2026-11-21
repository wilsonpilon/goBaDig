package edit

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestAllMenusAndMnemonics(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	for _, menu := range neMenus {
		for index, command := range menu.items {
			editor.area.SetText("unchanged", false)
			editor.changed = true
			editor.syncActiveWindow()
			_, start, end := editor.area.GetSelection()
			editor.openMenu(menu.key)
			drawEditor(editor, screen)
			dropdown := editor.dropdown
			if dropdown.GetItemCount() != len(menu.items) {
				t.Fatalf("%s: missing menu entries", menu.label)
			}
			if command.mnemonic != 0 {
				sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, unicodeLower(command.mnemonic), tcell.ModNone))
			} else {
				dropdown.selected = index
				sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
			}
			if editor.menuOpen {
				t.Fatalf("%s: command did not dismiss the menu", command.id)
			}
			if !editor.available(command) && !strings.Contains(editor.notice, command.id) {
				t.Fatalf("%s: unavailable command missing explicit notice: %q", command.id, editor.notice)
			}
			_, afterStart, afterEnd := editor.area.GetSelection()
			if command.id == "layout.center" {
				if editor.area.GetText() != strings.Repeat(" ", 28)+"unchanged" {
					t.Fatal("Center line did not center text within the default line length")
				}
			} else if editor.area.GetText() != "unchanged" || start != afterStart || end != afterEnd {
				t.Fatalf("%s modified document or cursor during menu navigation", command.id)
			}
			if editor.confirmExit {
				sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
			}
			for len(editor.dialogs) > 0 {
				sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
			}
			if editor.outlineView != nil {
				sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
			}
		}
	}
}

func unicodeLower(ch rune) rune {
	if ch >= 'A' && ch <= 'Z' {
		return ch + ('a' - 'A')
	}
	return ch
}

func TestMenuCellsMouseAndSmallScreens(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	if row := editorScreenRow(screen, 0); !strings.Contains(row, "Options") || row[70:77] != "F1=Help" {
		t.Fatalf("80-column menu geometry incorrect: %q", row)
	}
	up, _, _, _ := screen.GetContent(79, 1)
	down, _, _, _ := screen.GetContent(79, 23)
	if up != '\u2191' || down != '\u2193' {
		t.Fatal("vertical scrollbar must span the document viewport")
	}
	editor.openMenu('F')
	drawEditor(editor, screen)
	x, y, w, h := editor.dropdown.GetRect()
	corner, _, _, _ := screen.GetContent(x, y)
	shadow, _, shadowStyle, _ := screen.GetContent(x+w, y+h)
	if corner != '\u250c' || shadow != ' ' || shadowStyle != panelShadowStyle {
		t.Fatal("menu border or shadow missing")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone))
	if editor.currentMenu != 'B' {
		t.Fatal("Right must switch to the next menu")
	}
	drawEditor(editor, screen)
	handler := editor.MouseHandler()
	consumed, _ := handler(tview.MouseLeftClick, tcell.NewEventMouse(editor.menuX['P'], 0, tcell.Button1, tcell.ModNone), func(p tview.Primitive) { app.SetFocus(p) })
	if !consumed || editor.currentMenu != 'P' {
		t.Fatal("mouse must switch menus without editing text")
	}
	drawEditor(editor, screen)
	var printed []byte
	editor.printer = func(stream []byte) error {
		printed = append([]byte(nil), stream...)
		return nil
	}
	x, y, _, _ = editor.dropdown.GetRect()
	consumed, _ = editor.MouseHandler()(tview.MouseLeftClick, tcell.NewEventMouse(x+2, y+3, tcell.Button1, tcell.ModNone), func(p tview.Primitive) { app.SetFocus(p) })
	if !consumed || string(printed) != "\f" || editor.notice != "Sent an eject-page formfeed to the default printer." {
		t.Fatalf("mouse selected the wrong print command or payload: %q, %q", editor.notice, printed)
	}
	screen.SetSize(20, 8)
	editor.openMenu('F')
	drawEditor(editor, screen)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone))
	drawEditor(editor, screen)
	x, y, w, h = editor.dropdown.GetRect()
	if x < 0 || y < 0 || x+w > 20 || y+h > 8 || editor.dropdown.top == 0 {
		t.Fatal("menu must fit and scroll in a small terminal")
	}
}

func TestInsertOverAndStatusIndicators(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("abc\nnext", false)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyInsert, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'X', tcell.ModNone))
	if editor.area.GetText() != "Xbc\nnext" {
		t.Fatal("Over must replace a character, not just change its indicator")
	}

	editor.area.Select(3, 3)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'Y', tcell.ModNone))
	if editor.area.GetText() != "XbcY\nnext" {
		t.Fatal("overstrike must not delete a newline")
	}
	editor.openMenu('T')
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	drawEditor(editor, screen)
	if row := editorScreenRow(screen, 24); !strings.Contains(row, "Over") || !strings.Contains(row, "Auto") {
		t.Fatalf("mode indicators missing: %q", row)
	}
}

func TestScrollbarClicksMoveViewportNotCursor(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	editor.area.SetText(strings.Repeat("long "+strings.Repeat("x", 90)+"\n", 50), false)
	drawEditor(editor, screen)
	_, start, end := editor.area.GetSelection()
	handler := editor.MouseHandler()
	focus := func(p tview.Primitive) { app.SetFocus(p) }
	handler(tview.MouseLeftClick, tcell.NewEventMouse(79, 23, tcell.Button1, tcell.ModNone), focus)
	drawEditor(editor, screen)
	row, col := editor.area.GetOffset()
	if row != 1 || col != 0 {
		t.Fatalf("vertical scroll did not persist after drawing: %d,%d", row, col)
	}
	handler(tview.MouseLeftClick, tcell.NewEventMouse(79, 24, tcell.Button1, tcell.ModNone), focus)
	drawEditor(editor, screen)
	row, col = editor.area.GetOffset()
	_, afterStart, afterEnd := editor.area.GetSelection()
	if row != 1 || col != 1 || start != afterStart || end != afterEnd {
		t.Fatalf("horizontal scroll moved cursor or failed: viewport=%d,%d cursor=%d,%d", row, col, afterStart, afterEnd)
	}
}

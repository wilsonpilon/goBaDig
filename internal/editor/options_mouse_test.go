package edit

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestR26MouseAppliesRightButtonAndScrollThrottle(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	editor.area.SetText(strings.Repeat("line\n", 60), false)
	command, ok := commandByID("options.mouse")
	if !ok || !editor.available(command) {
		t.Fatal("Mouse setup command is unavailable")
	}

	editor.execute(command)
	dialog := editor.dialogs[0]
	dialogWidgetByID(t, dialog, "button").selected = 1
	dialogWidgetByID(t, dialog, "delay").selected = 3
	dialogWidgetByID(t, dialog, "speed").checked = true
	dialog.activate("ok")

	if !editor.mouseButtonRight || editor.mouseScrollDelay != 26*time.Millisecond ||
		editor.area.mouseButton != tcell.Button3 || editor.area.scrollDelay != 26*time.Millisecond {
		t.Fatal("mouse button or scroll delay was not applied to the active editor")
	}
	if !strings.Contains(editor.notice, "Not applied: mouse cursor speed") {
		t.Fatalf("unsupported mouse option was not reported: %q", editor.notice)
	}

	editor.area.SetText("first\nsecond", false)
	drawEditor(editor, screen)
	x, y, _, _ := editor.area.GetInnerRect()
	focus := func(p tview.Primitive) { app.SetFocus(p) }
	handler := editor.area.MouseHandler()
	handler(tview.MouseRightClick, tcell.NewEventMouse(x+3, y+1, tcell.Button3, tcell.ModNone), focus)
	if editor.area.document.cursor.line != 1 {
		t.Fatal("configured right button did not position the cursor")
	}
	cursor := editor.area.document.cursor
	handler(tview.MouseLeftClick, tcell.NewEventMouse(x+1, y, tcell.Button1, tcell.ModNone), focus)
	if editor.area.document.cursor != cursor {
		t.Fatal("inactive left button changed the cursor")
	}

	editor.area.SetText(strings.Repeat("line\n", 60), false)
	drawEditor(editor, screen)
	handler = editor.area.MouseHandler()
	editor.area.lastScroll = time.Now()
	handler(tview.MouseScrollDown, tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone), focus)
	if row, _ := editor.area.GetOffset(); row != 0 {
		t.Fatalf("scroll event inside the configured delay was applied at row %d", row)
	}
	editor.area.lastScroll = time.Now().Add(-time.Second)
	handler(tview.MouseScrollDown, tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone), focus)
	if row, _ := editor.area.GetOffset(); row != 3 {
		t.Fatalf("scroll event after the configured delay moved to row %d, want 3", row)
	}
}

func TestR26MouseReportsSyncClockAndSpeedAsNotApplied(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	command, _ := commandByID("options.mouse")
	editor.execute(command)
	dialog := editor.dialogs[0]
	dialogWidgetByID(t, dialog, "delay").selected = 4
	dialogWidgetByID(t, dialog, "speed").checked = true
	dialog.activate("ok")

	if editor.mouseScrollDelay != 0 {
		t.Fatal("unsupported sync-to-clock setting changed the scroll delay")
	}
	if !strings.Contains(editor.notice, "Not applied: sync to clock, mouse cursor speed") {
		t.Fatalf("host-controlled mouse options were not reported: %q", editor.notice)
	}
}

func TestR26MouseCancelDoesNotApplySettings(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	command, _ := commandByID("options.mouse")
	editor.execute(command)
	dialog := editor.dialogs[0]
	dialogWidgetByID(t, dialog, "button").selected = 1
	dialogWidgetByID(t, dialog, "delay").selected = 3
	dialog.activate("cancel")

	if editor.mouseButtonRight || editor.mouseScrollDelay != 0 ||
		editor.area.mouseButton != tcell.Button1 {
		t.Fatal("cancel applied mouse settings")
	}
}

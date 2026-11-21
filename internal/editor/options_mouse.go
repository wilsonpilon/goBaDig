package edit

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (e *Editor) activeMouseClick(action tview.MouseAction) bool {
	_, _, click := mouseButtonActions(e.mouseButtonMask())
	return action == click
}

func (e *Editor) mouseButtonMask() tcell.ButtonMask {
	if e.mouseButtonRight {
		return tcell.Button3
	}
	return tcell.Button1
}

func mouseButtonActions(button tcell.ButtonMask) (down, up, click tview.MouseAction) {
	if button == tcell.Button3 {
		return tview.MouseRightDown, tview.MouseRightUp, tview.MouseRightClick
	}
	return tview.MouseLeftDown, tview.MouseLeftUp, tview.MouseLeftClick
}

func (v *documentView) setMouseOptions(rightButton bool, scrollDelay time.Duration) {
	v.mouseButton = tcell.Button1
	if rightButton {
		v.mouseButton = tcell.Button3
	}
	v.scrollDelay = scrollDelay
	v.lastScroll = time.Time{}
}

func (e *Editor) mouseSetup() {
	d := newDialog("Mouse setup", 66, 17)
	d.context = "options.mouse"
	d.label(2, 2, "Mouse button:")
	button := 0
	if e.mouseButtonRight {
		button = 1
	}
	d.add(dialogWidget{kind: widgetRadio, id: "button", choices: []string{"Left", "Right"}, selected: button, x: 2, y: 3, width: 22, height: 2})
	d.label(28, 2, "Scroll delay:")
	d.add(dialogWidget{
		kind: widgetRadio, id: "delay",
		choices:  []string{"No delay", "12 ms", "18 ms", "26 ms", "Sync to clock (not applied)"},
		selected: mouseScrollDelayIndex(e.mouseScrollDelay), x: 28, y: 3, width: 36, height: 5,
	})
	d.add(dialogWidget{kind: widgetCheck, id: "speed", label: "Adjust mouse cursor speed (not supported; not applied)", x: 2, y: 10, width: 61})
	d.label(2, 12, "Scroll delay is enforced by the editor; pointer speed is host-controlled.")
	d.buttons([]string{"ok", "help", "cancel"}, []string{"OK", "F1=Help", "Cancel"}, "ok")
	d.result = func(id string) {
		if id == "help" {
			e.showHelp(d.context)
			return
		}
		if id == "cancel" {
			e.closeDialog()
			return
		}
		e.closeDialog()

		before := e.currentConfiguration()
		e.mouseButtonRight = d.widgets[d.fields[0]].selected == 1
		selectedDelay := d.widgets[d.fields[1]].selected
		var unavailable []string
		switch selectedDelay {
		case 0:
			e.mouseScrollDelay = 0
		case 1:
			e.mouseScrollDelay = 12 * time.Millisecond
		case 2:
			e.mouseScrollDelay = 18 * time.Millisecond
		case 3:
			e.mouseScrollDelay = 26 * time.Millisecond
		default:
			unavailable = append(unavailable, "sync to clock")
		}
		if d.checked("speed") {
			unavailable = append(unavailable, "mouse cursor speed")
		}
		for _, window := range e.windows {
			window.area.setMouseOptions(e.mouseButtonRight, e.mouseScrollDelay)
		}
		e.configChanged = e.configChanged || before != e.currentConfiguration()

		e.notice = "Mouse button and scroll delay applied."
		if len(unavailable) > 0 {
			e.notice += " Not applied: " + strings.Join(unavailable, ", ") + " (controlled by the host)."
		}
		e.updateStatus()
	}
	e.openDialog(d)
}

func mouseScrollDelayIndex(delay time.Duration) int {
	switch delay {
	case 12 * time.Millisecond:
		return 1
	case 18 * time.Millisecond:
		return 2
	case 26 * time.Millisecond:
		return 3
	default:
		return 0
	}
}

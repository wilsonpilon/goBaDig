package edit

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rivo/uniseg"
)

type widgetKind int

const (
	widgetLabel widgetKind = iota
	widgetInput
	widgetList
	widgetCheck
	widgetRadio
	widgetButton
)

type dialogWidget struct {
	kind                  widgetKind
	id, label, value      string
	x, y, width, height   int
	cursor, selected, top int
	choices               []string
	checked, fresh        bool
	numeric               bool
	minimum, maximum      int
	mnemonic              rune
	onSelect, onActivate  func(int)
}

type editorDialog struct {
	*tview.Box
	editor        *Editor
	title, page   string
	context       string
	width, height int
	widgets       []dialogWidget
	fields        []int
	focus, scroll int
	defaultID     string
	previousFocus tview.Primitive
	result        func(string)
	helpPages     [][]string
	helpPage      int
}

func newDialog(title string, width, height int) *editorDialog {
	return &editorDialog{Box: tview.NewBox(), title: title, width: width, height: height, defaultID: "ok"}
}

func (d *editorDialog) add(w dialogWidget) {
	if w.kind != widgetLabel {
		d.fields = append(d.fields, len(d.widgets))
	}
	if w.kind == widgetInput {
		w.cursor = len([]rune(w.value))
		w.fresh = true
	}
	if w.height == 0 {
		w.height = 1
	}
	if w.mnemonic == 0 && w.kind != widgetLabel && w.label != "" {
		w.mnemonic = unicode.ToUpper([]rune(w.label)[0])
	}
	d.widgets = append(d.widgets, w)
}

func (d *editorDialog) label(x, y int, text string) {
	d.add(dialogWidget{kind: widgetLabel, x: x, y: y, label: text, width: len(text)})
}

func (d *editorDialog) input(id, label, value string, y int, minimum, maximum int) {
	d.label(2, y, label)
	d.add(dialogWidget{kind: widgetInput, id: id, label: label, value: value, x: 3 + len(label), y: y,
		width: max(4, d.width-len(label)-7), numeric: maximum > 0, minimum: minimum, maximum: maximum})
}

func (d *editorDialog) buttons(ids, labels []string, defaultID string) {
	d.defaultID = defaultID
	x := 2
	for i, id := range ids {
		label := labels[i]
		mnemonic := unicode.ToUpper([]rune(label)[0])
		if id == "help" {
			mnemonic = 'H'
		}
		d.add(dialogWidget{kind: widgetButton, id: id, label: label, mnemonic: mnemonic, x: x, y: -1, width: len(label) + 4})
		x += len(label) + 6
	}
}

func (d *editorDialog) value(id string) string {
	for _, w := range d.widgets {
		if w.id == id {
			return w.value
		}
	}
	return ""
}

func (d *editorDialog) checked(id string) bool {
	for _, w := range d.widgets {
		if w.id == id {
			return w.checked
		}
	}
	return false
}

func (e *Editor) openDialog(d *editorDialog) {
	e.literal = false
	e.closeAutocomplete()
	d.editor = e
	d.previousFocus = e.app.GetFocus()
	d.page = fmt.Sprintf("dialog-%d", len(e.dialogs))
	e.dialogs = append(e.dialogs, d)
	e.AddPage(d.page, d, true, true)
	e.app.SetFocus(d)
}

func (e *Editor) closeDialog() {
	if len(e.dialogs) == 0 {
		return
	}
	d := e.dialogs[len(e.dialogs)-1]
	e.dialogs = e.dialogs[:len(e.dialogs)-1]
	e.RemovePage(d.page)
	if len(e.dialogs) > 0 {
		e.app.SetFocus(e.dialogs[len(e.dialogs)-1])
	} else if d.previousFocus != nil {
		e.app.SetFocus(d.previousFocus)
	} else {
		e.app.SetFocus(e.area)
	}
	e.updateStatus()
}

func (e *Editor) showMessage(title, text string) {
	lines := wrapDialogText(text, 60)
	d := newDialog(title, 64, len(lines)+5)
	for i, line := range lines {
		d.label(2, i+1, line)
	}
	d.buttons([]string{"ok"}, []string{"OK"}, "ok")
	if len(e.dialogs) > 0 {
		d.context = e.dialogs[len(e.dialogs)-1].context
	}
	d.result = func(id string) {
		if id == "help" {
			topic := d.context
			if topic == "" {
				topic = "help.general"
			}
			e.showHelp(topic)
			return
		}
		e.closeDialog()
	}
	e.openDialog(d)
}

func wrapDialogText(text string, width int) []string {
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		line := ""
		for _, word := range strings.Fields(paragraph) {
			if len(line)+len(word)+1 > width && line != "" {
				lines = append(lines, line)
				line = ""
			}
			if line != "" {
				line += " "
			}
			line += word
		}
		lines = append(lines, line)
	}
	return lines
}

func (d *editorDialog) activate(id string) {
	if id != "cancel" && id != "help" {
		for _, w := range d.widgets {
			if w.kind == widgetInput && w.numeric {
				n, err := strconv.Atoi(w.value)
				if err != nil || n < w.minimum || n > w.maximum {
					d.editor.showMessage("Invalid value", fmt.Sprintf("%s must be an integer from %d to %d.", w.id, w.minimum, w.maximum))
					return
				}
			}
		}
	}
	if d.result != nil {
		d.result(id)
	}
}

func (d *editorDialog) selectList(w *dialogWidget, index int) {
	if len(w.choices) == 0 {
		return
	}
	w.selected = max(0, min(len(w.choices)-1, index))
	if w.onSelect != nil {
		w.onSelect(w.selected)
	}
}

func (d *editorDialog) key(event *tcell.EventKey) {
	if event.Key() == tcell.KeyEscape {
		d.activate("cancel")
		return
	}
	if event.Key() == tcell.KeyF1 {
		d.activate("help")
		return
	}
	if len(d.helpPages) > 0 {
		_, _, _, height := d.GetRect()
		limit := max(0, 18-max(1, height-3))
		switch event.Key() {
		case tcell.KeyUp:
			d.scroll = max(0, d.scroll-1)
			return
		case tcell.KeyDown:
			d.scroll = min(limit, d.scroll+1)
			return
		case tcell.KeyPgUp:
			d.scroll = max(0, d.scroll-max(1, height-4))
			return
		case tcell.KeyPgDn:
			d.scroll = min(limit, d.scroll+max(1, height-4))
			return
		}
	}
	if len(d.fields) == 0 {
		return
	}
	switch event.Key() {
	case tcell.KeyTab:
		d.focus = (d.focus + 1) % len(d.fields)
		return
	case tcell.KeyBacktab:
		d.focus = (d.focus + len(d.fields) - 1) % len(d.fields)
		return
	}
	w := &d.widgets[d.fields[d.focus]]
	if event.Key() == tcell.KeyEnter {
		if w.kind == widgetButton {
			d.activate(w.id)
		} else if w.kind == widgetList && w.onActivate != nil && len(w.choices) > 0 {
			w.onActivate(w.selected)
		} else {
			d.activate(d.defaultID)
		}
		return
	}
	if event.Key() == tcell.KeyRune && event.Modifiers()&tcell.ModAlt != 0 {
		for focus, index := range d.fields {
			target := &d.widgets[index]
			if target.mnemonic != 0 && target.mnemonic == unicode.ToUpper(event.Rune()) {
				d.focus = focus
				switch target.kind {
				case widgetButton:
					d.activate(target.id)
				case widgetCheck:
					target.checked = !target.checked
				}
				return
			}
		}
	}
	if event.Key() == tcell.KeyRune && w.kind != widgetInput {
		for _, button := range d.widgets {
			if button.kind == widgetButton && button.mnemonic == unicode.ToUpper(event.Rune()) {
				d.activate(button.id)
				return
			}
		}
	}
	switch w.kind {
	case widgetInput:
		text := []rune(w.value)
		switch event.Key() {
		case tcell.KeyLeft:
			w.cursor = max(0, w.cursor-1)
			w.fresh = false
		case tcell.KeyRight:
			w.cursor = min(len(text), w.cursor+1)
			w.fresh = false
		case tcell.KeyHome:
			w.cursor = 0
			w.fresh = false
		case tcell.KeyEnd:
			w.cursor = len(text)
			w.fresh = false
		case tcell.KeyCtrlU:
			text = nil
			w.cursor = 0
			w.fresh = false
		case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
			if w.fresh {
				text = nil
				w.cursor = 0
			} else if event.Key() == tcell.KeyDelete && w.cursor < len(text) {
				text = append(text[:w.cursor], text[w.cursor+1:]...)
			} else if event.Key() != tcell.KeyDelete && w.cursor > 0 {
				text = append(text[:w.cursor-1], text[w.cursor:]...)
				w.cursor--
			}
			w.fresh = false
		case tcell.KeyRune:
			if event.Modifiers()&^(tcell.ModShift) == 0 && !unicode.IsControl(event.Rune()) {
				if w.fresh {
					text = nil
					w.cursor = 0
				}
				text = append(text, 0)
				copy(text[w.cursor+1:], text[w.cursor:])
				text[w.cursor] = event.Rune()
				w.cursor++
				w.fresh = false
			}
		}
		w.value = string(text)
	case widgetCheck:
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			w.checked = !w.checked
		}
	case widgetRadio, widgetList:
		count := len(w.choices)
		if count == 0 {
			return
		}
		switch event.Key() {
		case tcell.KeyLeft, tcell.KeyUp:
			w.selected = (w.selected + count - 1) % count
		case tcell.KeyRight, tcell.KeyDown:
			w.selected = (w.selected + 1) % count
		case tcell.KeyHome:
			w.selected = 0
		case tcell.KeyEnd:
			w.selected = count - 1
		case tcell.KeyPgUp:
			w.selected = max(0, w.selected-max(1, w.height-2))
		case tcell.KeyPgDn:
			w.selected = min(count-1, w.selected+max(1, w.height-2))
		}
		if w.kind == widgetList && w.onSelect != nil {
			w.onSelect(w.selected)
		}
	case widgetButton:
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			d.activate(w.id)
		}
	}
}

func (d *editorDialog) Draw(screen tcell.Screen) {
	sw, sh := screen.Size()
	width, height := min(d.width, sw), min(d.height, max(0, sh-2))
	x, y := max(0, (sw-width)/2), max(0, (sh-height)/2)
	d.SetRect(x, y, width, height)
	drawPanel(screen, x, y, width, height, d.title, true)
	screen.HideCursor()
	if width < 4 || height < 4 {
		return
	}
	visible := height - 3
	if len(d.fields) > 0 {
		w := d.widgets[d.fields[d.focus]]
		target := w.y
		if w.kind == widgetRadio {
			target += w.selected
		}
		if target >= 0 {
			if target < d.scroll+1 {
				d.scroll = max(0, target-1)
			}
			if target >= d.scroll+visible {
				d.scroll = max(0, target-visible+1)
			}
		}
	}
	for i := range d.widgets {
		w := &d.widgets[i]
		focused := len(d.fields) > 0 && d.fields[d.focus] == i
		wy := w.y - d.scroll
		if w.kind == widgetButton {
			wy = height - 2
		}
		style := dialogStyle
		wx := max(1, min(w.x, width-2))
		available := max(0, width-wx-1)
		if w.kind != widgetButton {
			available = min(available, w.width)
		}
		if w.kind == widgetButton && focused {
			wx = min(wx, max(1, width-w.width-1))
			available = width - wx - 1
		}
		put := func(dy int, text string, style tcell.Style) {
			if dy > 0 && dy < height-1 {
				drawLabel(screen, x+wx, y+dy, available, text, 0, style)
			}
		}
		switch w.kind {
		case widgetLabel:
			if wy < height-2 {
				put(wy, w.label, style)
			}
		case widgetInput:
			if wy <= 0 || wy >= height-2 {
				continue
			}
			fieldWidth := min(w.width, available)
			if fieldWidth < 3 {
				continue
			}
			text := []rune(w.value)
			start := 0
			for start < w.cursor && uniseg.StringWidth(string(text[start:w.cursor])) > fieldWidth-3 {
				start++
			}
			value := string(text[start:])
			put(wy, "[ "+strings.Repeat(" ", fieldWidth-3)+"]", dialogStyle)
			if focused && w.fresh {
				style = selectedStyle
			}
			drawLabel(screen, x+wx+1, y+wy, fieldWidth-2, value, 0, style)
			if focused {
				screen.ShowCursor(x+wx+1+uniseg.StringWidth(string(text[start:w.cursor])), y+wy)
			}
		case widgetCheck:
			marker := " "
			if w.checked {
				marker = "x"
			}
			if wy < height-2 {
				put(wy, "["+marker+"] "+w.label, style)
				if focused && wy > 0 {
					screen.ShowCursor(x+wx+1, y+wy)
				}
			}
		case widgetRadio:
			for index, choice := range w.choices {
				marker := " "
				if index == w.selected {
					marker = "*"
				}
				if wy+index < height-2 {
					put(wy+index, "("+marker+") "+choice, style)
					if focused && index == w.selected && wy+index > 0 {
						screen.ShowCursor(x+wx+1, y+wy+index)
					}
				}
			}
		case widgetList:
			rows := max(1, w.height-2)
			if w.selected < w.top {
				w.top = w.selected
			}
			if w.selected >= w.top+rows {
				w.top = w.selected - rows + 1
			}
			put(wy, w.label, dialogStyle)
			for row := 0; row < rows && row+w.top < len(w.choices); row++ {
				if wy+row+1 >= height-2 {
					break
				}
				rowStyle := dialogStyle
				if focused && w.selected == row+w.top {
					rowStyle = selectedStyle
				}
				put(wy+row+1, w.choices[row+w.top], rowStyle)
			}
		case widgetButton:
			label := "  " + w.label + "  "
			if focused {
				label = "[ " + w.label + " ]"
			}
			put(wy, label, selectedStyle)
		}
	}
}

func (d *editorDialog) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, _ func(tview.Primitive)) { d.key(event) })
}

func (d *editorDialog) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return d.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, _ func(tview.Primitive)) (bool, tview.Primitive) {
		if d.editor.activeMouseClick(action) && d.InRect(event.Position()) {
			mx, my := event.Position()
			x, y, width, height := d.GetRect()
			for focus, index := range d.fields {
				w := &d.widgets[index]
				wy, wx := w.y-d.scroll, max(1, min(w.x, width-2))
				if w.kind == widgetButton {
					wy = height - 2
					if d.focus == focus {
						wx = min(wx, max(1, width-w.width-1))
					}
				}
				if mx < x+wx || mx >= x+min(width-1, wx+w.width) || my < y+wy || my >= y+wy+w.height {
					continue
				}
				d.focus = focus
				switch w.kind {
				case widgetCheck:
					w.checked = !w.checked
				case widgetRadio:
					w.selected = min(len(w.choices)-1, my-y-wy)
				case widgetList:
					d.selectList(w, w.top+my-y-wy-1)
				case widgetButton:
					d.activate(w.id)
				}
				break
			}
		}
		return true, nil
	})
}

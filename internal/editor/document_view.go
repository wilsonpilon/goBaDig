package edit

import (
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type documentView struct {
	*tview.Box
	document       *document
	changed, moved func()
	activated      func()
	scrollbar      bool
	mouseButton    tcell.ButtonMask
	scrollDelay    time.Duration
	lastScroll     time.Time
	over           bool
	dragged        bool
	dragging       bool
}

func newDocumentView() *documentView {
	v := &documentView{Box: tview.NewBox(), document: newDocument(""), scrollbar: true, mouseButton: tcell.Button1}
	v.SetBackgroundColor(tcell.ColorBlack)
	return v
}

func (v *documentView) notify(changed bool) {
	if changed {
		v.document.searchMatch = false
	}
	if changed && v.changed != nil {
		v.changed()
	} else if v.moved != nil {
		v.moved()
	}
}

func (v *documentView) SetText(text string, cursorAtEnd bool) {
	changed := text != v.document.text
	sourceStart := v.document.sourcePositionAt(0)
	sourceEnd := v.document.sourcePositionAt(len(v.document.text))
	v.dragged, v.dragging = false, false
	v.document.load(text)
	v.document.setSourceRuns([]documentSourceRun{{
		textStart: 0, textEnd: len(text), sourceStart: sourceStart, sourceEnd: sourceEnd,
	}})
	if cursorAtEnd {
		last := len(v.document.lines) - 1
		v.document.setCursor(documentPosition{last, len(v.document.lines[last].cells)}, false)
	}
	if changed {
		v.document.revision++
	}
	v.notify(changed)
}

func (v *documentView) GetText() string { return v.document.text }

func (v *documentView) GetSelection() (string, int, int) {
	start, end := v.document.selection()
	return v.document.text[start:end], start, end
}

func (v *documentView) HasSelection() bool {
	start, end := v.document.selection()
	return start != end
}

func (v *documentView) Select(start, end int) {
	v.document.selectBytes(start, end)
	v.notify(false)
}

func (v *documentView) GetOffset() (int, int) {
	return v.document.top, v.document.left
}

func (v *documentView) SetOffset(row, column int) {
	d := v.document
	d.top = max(0, min(row, len(d.rows)-1))
	longest := 0
	for _, line := range d.lines {
		longest = max(longest, line.width)
	}
	d.left = max(0, min(column, max(0, longest-d.width+1)))
	if d.wrap {
		d.left = 0
	}
	v.notify(false)
}

func (v *documentView) SetWrap(wrap bool) {
	d := v.document
	d.resize(d.width, d.height, wrap)
	v.notify(false)
}

func (v *documentView) navigate(id string) bool {
	if !v.document.navigate(id) {
		return false
	}
	v.notify(false)
	return true
}

func (v *documentView) edit(id string) (bool, string) {
	changed, notice, handled := v.document.edit(id, v.over)
	if handled {
		v.notify(changed)
	}
	return handled, notice
}

func (v *documentView) Draw(screen tcell.Screen) {
	v.Box.DrawForSubclass(screen, v)
	x, y, width, height := v.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}
	d := v.document
	gutter := d.markerGutterWidth()
	scrollbarWidth := 0
	if v.scrollbar {
		scrollbarWidth = 1
	}
	contentWidth := max(1, width-gutter-scrollbarWidth)
	d.resize(contentWidth, height, d.wrap)
	style := contentStyle
	start, end := d.selection()
	for index := 0; index < d.blockMarkerCount; index++ {
		row := d.rowAt(d.positionAtByte(d.blockMarkers[index]))
		if row >= d.top && row < d.top+height {
			screen.SetContent(x+index, y+row-d.top, '\u25aa', nil, markerStyle)
		}
	}
	for dy := 0; dy < height && dy+d.top < len(d.rows); dy++ {
		row := d.rows[dy+d.top]
		line := d.lines[row.line]
		left := d.left
		if d.wrap {
			left = row.from
		}
		first := sort.Search(len(line.cells), func(i int) bool {
			return line.cells[i].column+line.cells[i].width > left
		})
		for _, cell := range line.cells[first:] {
			cx := cell.column - left
			if cell.column >= row.to || cx >= contentWidth {
				break
			}
			if cx+cell.width <= 0 {
				continue
			}
			cellStyle := style
			if cell.start >= start && cell.start < end {
				cellStyle = blockStyle
			}
			if d.searchMatch && cell.start < d.searchMatchEnd && cell.end > d.searchMatchStart {
				cellStyle = selectedStyle
			}
			if cell.tab {
				for dx := max(0, cx); dx < min(contentWidth, cx+cell.width); dx++ {
					screen.SetContent(x+gutter+dx, y+dy, ' ', nil, cellStyle)
				}
			} else if cx >= 0 && cx+cell.width <= contentWidth {
				drawLabel(screen, x+gutter+cx, y+dy, cell.width, cell.text, 0, cellStyle)
			}
		}
	}
	if v.scrollbar {
		for dy := 0; dy < height; dy++ {
			ch := '\u2591'
			scrollStyle := panelStyle
			switch dy {
			case 0:
				ch = '\u2191'
			case height - 1:
				ch = '\u2193'
			default:
				track := max(1, height-2)
				thumb := min(track-1, d.top*(track-1)/max(1, len(d.rows)-1))
				if dy == thumb+1 {
					ch = '\u2588'
					scrollStyle = scrollStyle.Foreground(scrollbarThumbColor)
				}
			}
			screen.SetContent(x+width-1, y+dy, ch, nil, scrollStyle)
		}
	}
	if v.HasFocus() {
		row := d.rowAt(d.cursor)
		column := d.column(d.cursor) - d.left
		if d.wrap {
			column = d.column(d.cursor) - d.rows[row].from
		}
		if row >= d.top && row < d.top+height && column >= 0 && column < contentWidth {
			screen.ShowCursor(x+gutter+column, y+row-d.top)
		} else {
			screen.HideCursor()
		}
	}
}

func (v *documentView) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return v.WrapInputHandler(func(event *tcell.EventKey, _ func(tview.Primitive)) {
		d := v.document
		if event.Key() == tcell.KeyRune {
			if event.Modifiers()&^tcell.ModShift != 0 {
				return
			}
			v.notify(d.insertText(string(event.Rune()), v.over))
			return
		}
		if command, ok := commandForKey(event); ok && strings.HasPrefix(command.id, "edit.") {
			v.edit(command.id)
		}
	})
}

func (v *documentView) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return v.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, focus func(tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()
		x, y, width, height := v.GetInnerRect()
		d := v.document
		buttonDown, buttonUp, buttonClick := mouseButtonActions(v.mouseButton)
		activeMove := action == tview.MouseMove && event.Buttons()&v.mouseButton != 0
		if !v.InRect(mx, my) && !(v.dragging && (action == tview.MouseMove || action == buttonUp)) {
			return false, nil
		}
		if action == buttonUp {
			wasDragging := v.dragging
			v.dragging = false
			if wasDragging && d.cursor != d.anchor {
				d.setBlockRange(d.byteOffset(d.anchor), d.byteOffset(d.cursor))
				v.notify(false)
			}
			return wasDragging, nil
		}
		if v.scrollbar && mx == x+width-1 && v.InRect(mx, my) {
			if action == buttonClick {
				row := d.top
				switch my {
				case y:
					row--
				case y + height - 1:
					row++
				default:
					row = (my - y - 1) * (len(d.rows) - 1) / max(1, height-3)
				}
				v.SetOffset(row, d.left)
			}
			return true, nil
		}
		switch action {
		case tview.MouseScrollUp, tview.MouseScrollDown:
			if v.scrollDelay > 0 && !v.lastScroll.IsZero() && time.Since(v.lastScroll) < v.scrollDelay {
				return true, nil
			}
			v.lastScroll = time.Now()
			delta := 3
			if action == tview.MouseScrollUp {
				delta = -delta
			}
			v.SetOffset(d.top+delta, d.left)
		default:
			if action != buttonDown && action != buttonClick && !activeMove {
				return false, nil
			}
			if (action == buttonDown || action == buttonClick) && v.activated != nil {
				v.activated()
			}
			if action == buttonClick && v.dragged {
				v.dragged = false
				return true, nil
			}
			focus(v)
			row := max(0, min(my-y, height-1)) + d.top
			row = min(row, len(d.rows)-1)
			column := max(0, mx-x-d.markerGutterWidth())
			if !d.wrap {
				column += d.left
			}
			p := d.positionOnRow(row, column)
			if activeMove {
				if p != d.cursor {
					d.deleted = nil
					d.searchMatch = false
				}
				d.cursor = p
				d.goal = d.column(p)
				v.dragged = d.cursor != d.anchor
			} else {
				if action == buttonDown || action == buttonClick {
					d.clearBlockMarkers()
				}
				d.searchMatch = false
				d.setCursor(p, false)
				v.dragged = false
			}
			v.notify(false)
			if action == buttonDown {
				v.dragging = true
				return true, v
			}
		}
		return true, nil
	})
}

func (v *documentView) PasteHandler() func(string, func(tview.Primitive)) {
	return v.WrapPasteHandler(func(text string, _ func(tview.Primitive)) {
		v.notify(v.document.insertText(text, false))
	})
}

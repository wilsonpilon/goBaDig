package edit

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/goBaDig/internal/man"
)

// Styles for man.LineStyle, independent of the editor's own display-
// profile palette (markerStyle etc., options_display.go) so a man page
// always reads the same regardless of the user's chosen color profile.
var (
	manHeadingStyle    = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack).Bold(true)
	manSubheadingStyle = tcell.StyleDefault.Foreground(tcell.ColorAqua).Background(tcell.ColorBlack).Bold(true)
	manCodeStyle       = tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)
	manQuoteStyle      = tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
)

// editorManView is a full-screen, scrollable viewer for man.Page
// content - the "Manual" menu item today, and (Fase 3) Shift+F1's
// contextual lookup for MSX keywords. Modeled on editorOutlineView
// (view_outline.go) but, unlike that "press any key to return"
// display, this one has real navigation: Up/Down/PgUp/PgDn/Home/End
// scroll, Enter on a table-of-contents line opens the topic it links
// to (pushing it onto history), Esc goes back one level or closes the
// view entirely once history is exhausted.
type editorManView struct {
	*tview.Box
	editor         *Editor
	history        []man.Page
	top, selected  int
	previousNotice string
}

// openManView opens page in the Manual viewer, or, if it is already
// open, navigates into page from wherever it currently is (used by
// Enter on a TOC line - see activate, below).
func (e *Editor) openManView(page man.Page) {
	if e.manView != nil {
		e.manView.push(page)
		return
	}
	e.closeAutocomplete()
	view := &editorManView{
		Box:            tview.NewBox(),
		editor:         e,
		history:        []man.Page{page},
		previousNotice: e.notice,
	}
	e.manView = view
	e.notice = "Manual - Enter abre o topico, Esc volta/fecha."
	e.AddPage("man", view, true, true)
	e.app.SetFocus(view)
	e.updateStatus()
}

// manDisplay opens the Manual viewer at its index - the "Manual" menu
// (commands.go's "manual.open"). Deliberately not under the "help."
// id prefix: Editor.execute intercepts every "help."-prefixed command
// unconditionally into showHelp (WibEdit's own F1 "how do I use this
// editor" reference) before the command switch below is ever reached -
// this is a different, goBaDig-specific reference (internal/man).
func (e *Editor) manDisplay() {
	e.openManView(man.Index())
}

func (e *Editor) closeManView() {
	view := e.manView
	if view == nil {
		return
	}
	e.RemovePage("man")
	e.manView = nil
	e.notice = view.previousNotice
	e.app.SetFocus(e.area)
	e.updateStatus()
}

func (v *editorManView) push(page man.Page) {
	v.history = append(v.history, page)
	v.top, v.selected = 0, 0
}

func (v *editorManView) page() man.Page {
	return v.history[len(v.history)-1]
}

func (v *editorManView) visibleHeight() int {
	_, _, _, height := v.editor.area.GetInnerRect()
	return height
}

// key handles every key while the Manual viewer is open - wired from
// Editor.Capture the same way the dialog stack is (editor.go), ahead
// of literal/dialog/autocomplete handling, since the viewer is modal:
// no other editor action is reachable until it closes.
func (v *editorManView) key(event *tcell.EventKey) {
	height := v.visibleHeight()
	switch event.Key() {
	case tcell.KeyUp:
		v.move(-1, height)
	case tcell.KeyDown:
		v.move(1, height)
	case tcell.KeyPgUp:
		v.move(-max(1, height-1), height)
	case tcell.KeyPgDn:
		v.move(max(1, height-1), height)
	case tcell.KeyHome:
		v.selected, v.top = 0, 0
	case tcell.KeyEnd:
		v.selected = max(0, len(v.page().Lines)-1)
		v.ensureVisible(height)
	case tcell.KeyEnter:
		v.activate()
	case tcell.KeyEscape:
		if len(v.history) > 1 {
			v.history = v.history[:len(v.history)-1]
			v.top, v.selected = 0, 0
		} else {
			v.editor.closeManView()
		}
	}
	v.editor.updateStatus()
}

func (v *editorManView) move(delta, height int) {
	lines := len(v.page().Lines)
	if lines == 0 {
		return
	}
	v.selected = clampInt(v.selected+delta, 0, lines-1)
	v.ensureVisible(height)
}

func (v *editorManView) ensureVisible(height int) {
	if height <= 0 {
		return
	}
	if v.selected < v.top {
		v.top = v.selected
	}
	if v.selected >= v.top+height {
		v.top = v.selected - height + 1
	}
}

// activate follows the selected line to a linked topic, if the
// current page's TOC has an entry at that line (on the Index page,
// every line is one) - a plain line with no TOC entry does nothing.
func (v *editorManView) activate() {
	page := v.page()
	for _, entry := range page.TOC {
		if entry.Line == v.selected {
			if linked, err := man.Lookup(entry.Title); err == nil {
				v.push(linked)
			}
			return
		}
	}
}

func clampInt(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func (v *editorManView) Draw(screen tcell.Screen) {
	x, y, width, height := v.editor.area.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}
	page := v.page()
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			screen.SetContent(x+col, y+row, ' ', nil, contentStyle)
		}
		index := v.top + row
		if index >= len(page.Lines) {
			continue
		}
		line := page.Lines[index]
		style := manStyleFor(line.Style)
		if index == v.selected {
			style = selectedStyle
		}
		drawLabel(screen, x, y+row, width, line.Text, 0, style)
	}
}

func manStyleFor(style man.LineStyle) tcell.Style {
	switch style {
	case man.StyleHeading1:
		return manHeadingStyle
	case man.StyleHeading2, man.StyleHeading3:
		return manSubheadingStyle
	case man.StyleCode:
		return manCodeStyle
	case man.StyleQuote:
		return manQuoteStyle
	default:
		return contentStyle
	}
}

func (v *editorManView) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return v.WrapInputHandler(func(event *tcell.EventKey, _ func(tview.Primitive)) { v.key(event) })
}

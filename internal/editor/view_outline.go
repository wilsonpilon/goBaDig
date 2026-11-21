package edit

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type editorOutlineView struct {
	*tview.Box
	editor         *Editor
	lines          []string
	previousNotice string
}

func (e *Editor) outlineDisplay() {
	text := e.area.GetText()
	view := &editorOutlineView{
		Box:            tview.NewBox(),
		editor:         e,
		lines:          compactOutlineLines(text),
		previousNotice: e.notice,
	}
	e.outlineView = view
	e.notice = "Outline view - press any key to return."
	if e.diskKnown && e.path != "" &&
		e.area.document.sourcePositionAt(len(e.area.document.text)) < len(e.diskText) {
		e.notice = "Outline view (loaded portion only) - press any key to return."
	}
	e.AddPage("outline", view, true, true)
	e.app.SetFocus(view)
	e.updateStatus()
}

func (e *Editor) closeOutlineView() {
	view := e.outlineView
	if view == nil {
		return
	}
	e.RemovePage("outline")
	e.outlineView = nil
	e.notice = view.previousNotice
	e.app.SetFocus(e.area)
	e.updateStatus()
}

func compactOutlineLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSuffix(line, "\r")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.ToValidUTF8(line, "\uFFFD")
		line = strings.Map(func(character rune) rune {
			if character == '\t' {
				return ' '
			}
			if unicode.IsControl(character) {
				return '\uFFFD'
			}
			return character
		}, line)
		lines = append(lines, line)
	}
	return lines
}

func (v *editorOutlineView) Draw(screen tcell.Screen) {
	x, y, width, height := v.editor.area.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}
	style := contentStyle
	for row := 0; row < height; row++ {
		for column := 0; column < width; column++ {
			screen.SetContent(x+column, y+row, ' ', nil, style)
		}
		if row < len(v.lines) {
			drawLabel(screen, x, y+row, width, v.lines[row], 0, style)
		}
	}
}

func (v *editorOutlineView) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return v.WrapInputHandler(func(_ *tcell.EventKey, _ func(tview.Primitive)) {
		v.editor.closeOutlineView()
	})
}

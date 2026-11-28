package edit

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func setDisplayPalette(profile int) {
	if profile < 0 || profile > 4 {
		profile = 0
	}
	black := tcell.NewRGBColor(0, 0, 0)
	white := tcell.NewRGBColor(255, 255, 255)
	blue := tcell.NewRGBColor(0, 0, 170)
	cyan := tcell.NewRGBColor(0, 170, 170)
	gray := tcell.NewRGBColor(170, 170, 170)
	darkGray := tcell.NewRGBColor(85, 85, 85)
	brown := tcell.NewRGBColor(170, 85, 0)
	red := tcell.NewRGBColor(170, 0, 0)
	yellow := tcell.NewRGBColor(255, 255, 85)
	panelStyle = tcell.StyleDefault.Foreground(black).Background(gray)
	menuStyle = panelStyle
	dialogStyle = panelStyle
	dialogTitleStyle = dialogStyle
	selectedStyle = tcell.StyleDefault.Foreground(gray).Background(black)
	pendingStyle = panelStyle
	contentStyle = tcell.StyleDefault.Foreground(gray).Background(black)
	blockStyle = panelStyle
	panelShadowStyle = tcell.StyleDefault.Foreground(darkGray).Background(darkGray)
	mnemonicForegroundColor = white
	scrollbarThumbColor = black
	switch profile {
	case 0:
		panelStyle = tcell.StyleDefault.Foreground(black).Background(cyan)
		menuStyle = tcell.StyleDefault.Foreground(white).Background(cyan)
		dialogStyle = menuStyle
		dialogTitleStyle = dialogStyle.Foreground(yellow)
		selectedStyle = tcell.StyleDefault.Foreground(white).Background(black)
		pendingStyle = panelStyle.Foreground(darkGray)
		contentStyle = tcell.StyleDefault.Foreground(white).Background(blue)
		blockStyle = tcell.StyleDefault.Foreground(blue).Background(gray)
		panelShadowStyle = tcell.StyleDefault.Foreground(black).Background(black)
		mnemonicForegroundColor = yellow
		scrollbarThumbColor = red
	case 1:
		contentStyle = tcell.StyleDefault.Foreground(white).Background(blue)
		blockStyle = tcell.StyleDefault.Foreground(blue).Background(gray)
		panelShadowStyle = tcell.StyleDefault.Foreground(black).Background(black)
		mnemonicForegroundColor = red
	case 2:
		// These are the NE attributes observed on a VGA color adapter.
		pendingStyle = panelStyle.Foreground(brown)
		blockStyle = panelStyle.Background(brown)
		panelShadowStyle = tcell.StyleDefault.Foreground(blue).Background(blue)
	case 4:
		mnemonicForegroundColor = tcell.ColorDefault
	}
	markerStyle = blockStyle
	_, contentBackgroundColor, _ = contentStyle.Decompose()
	statusForegroundColor, statusBackgroundColor, _ = panelStyle.Decompose()
	syntaxKeywordStyle = syntaxKeywordStyle.Background(contentBackgroundColor)
	syntaxFunctionStyle = syntaxFunctionStyle.Background(contentBackgroundColor)
	syntaxCommentStyle = syntaxCommentStyle.Background(contentBackgroundColor)
	syntaxStringStyle = syntaxStringStyle.Background(contentBackgroundColor)
	syntaxNumberStyle = syntaxNumberStyle.Background(contentBackgroundColor)
	syntaxLabelStyle = syntaxLabelStyle.Background(contentBackgroundColor)
	syntaxConstantStyle = syntaxConstantStyle.Background(contentBackgroundColor)
}

// syntaxStyleFor returns the display style for a highlight category,
// or ok=false for synNone (callers fall back to the plain content
// style).
func syntaxStyleFor(kind syntaxKind) (tcell.Style, bool) {
	switch kind {
	case synKeyword:
		return syntaxKeywordStyle, true
	case synFunction:
		return syntaxFunctionStyle, true
	case synComment:
		return syntaxCommentStyle, true
	case synString:
		return syntaxStringStyle, true
	case synNumber:
		return syntaxNumberStyle, true
	case synLabel:
		return syntaxLabelStyle, true
	case synConstant:
		return syntaxConstantStyle, true
	default:
		return tcell.StyleDefault, false
	}
}

func (e *Editor) applyDisplayPalette(profile int) {
	e.colorProfile = profile
	setDisplayPalette(profile)
	e.SetBackgroundColor(contentBackgroundColor)
	if e.root != nil {
		e.root.SetBackgroundColor(contentBackgroundColor)
	}
	if e.windowLayout != nil {
		e.windowLayout.SetBackgroundColor(contentBackgroundColor)
	}
	for _, window := range e.windows {
		window.area.SetBackgroundColor(contentBackgroundColor)
	}
	if e.statusBar != nil {
		e.updateDisplayStatusColors()
	}
}

func (e *Editor) displaySetup() {
	d := newDialog("Display setup", 76, 21)
	d.context = "options.display"
	d.label(2, 2, "Color profile:")
	d.add(dialogWidget{
		kind: widgetRadio, id: "color",
		choices:  []string{"Color 1", "Color 2", "VGA Monochrome", "Monochrome", "LCD"},
		selected: e.colorProfile, x: 2, y: 3, width: 27, height: 5,
	})
	d.label(32, 2, "Display lines:")
	d.add(dialogWidget{
		kind: widgetRadio, id: "lines",
		choices:  []string{"25", "33/35", "40/43", "50", "Don't change"},
		selected: 4, x: 32, y: 3, width: 39, height: 5,
	})
	d.add(dialogWidget{kind: widgetCheck, id: "scrollbar", label: "Display scrollbar", checked: e.displayScrollbar, x: 2, y: 9, width: 30})
	d.add(dialogWidget{kind: widgetCheck, id: "blink", label: "Blinking bracket (not supported; not applied)", x: 2, y: 10, width: 68})
	d.add(dialogWidget{kind: widgetCheck, id: "zoom", label: "Zoom window effects (not supported; not applied)", x: 2, y: 11, width: 68})
	d.add(dialogWidget{kind: widgetCheck, id: "cursor", label: "Large cursor (not supported; not applied)", x: 2, y: 12, width: 68})
	d.label(2, 14, "Tab spacing:")
	d.add(dialogWidget{kind: widgetInput, id: "tab-spacing", value: fmt.Sprint(e.area.document.tabWidth), x: 16, y: 14, width: 6, numeric: true, minimum: 1, maximum: 32})
	d.label(2, 16, "Terminal controls size; display line counts are not applied.")
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
		tabWidth, err := parseDisplayTabSpacing(d.value("tab-spacing"))
		if err != nil {
			e.showMessage("Invalid value", err.Error())
			return
		}
		e.closeDialog()

		before := e.currentConfiguration()
		e.applyDisplayPalette(d.widgets[d.fields[0]].selected)
		e.displayScrollbar = d.checked("scrollbar")
		for _, window := range e.windows {
			window.area.scrollbar = e.displayScrollbar
			window.area.document.tabWidth = tabWidth
			window.area.document.indexLines()
			window.area.notify(false)
		}
		e.configChanged = e.configChanged || before != e.currentConfiguration()

		var unavailable []string
		if d.widgets[d.fields[1]].selected != 4 {
			unavailable = append(unavailable, "terminal height")
		}
		if d.checked("blink") {
			unavailable = append(unavailable, "blinking bracket")
		}
		if d.checked("zoom") {
			unavailable = append(unavailable, "zoom effects")
		}
		if d.checked("cursor") {
			unavailable = append(unavailable, "large cursor")
		}
		e.notice = "Display colors, scrollbar and tab spacing applied."
		if len(unavailable) > 0 {
			e.notice += " Not applied: " + strings.Join(unavailable, ", ") + " (unsupported by this terminal UI)."
		} else {
			e.notice += " Terminal size and effects remain unchanged."
		}
		e.updateStatus()
	}
	e.openDialog(d)
}

func parseDisplayTabSpacing(value string) (int, error) {
	width, err := strconv.Atoi(value)
	if err != nil || width < 1 || width > 32 {
		return 0, fmt.Errorf("Tab spacing must be an integer from 1 to 32.")
	}
	return width, nil
}

func (e *Editor) updateDisplayStatusColors() {
	e.statusBar.SetBackgroundColor(statusBackgroundColor)
	for _, cell := range []*tview.TextView{
		e.lineStatus, e.columnStatus, e.nameStatus, e.modeStatus, e.wrapStatus, e.autoStatus,
	} {
		cell.SetTextColor(statusForegroundColor)
		cell.SetBackgroundColor(statusBackgroundColor)
	}
}

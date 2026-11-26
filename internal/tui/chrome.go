package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/goBaDig/internal/chrome"
)

// desktop is the application's background: a dithered Turbo-Vision-style
// texture with a centered banner. It replaces a plain TextView so the
// banner can be drawn on top of the texture instead of covering it with a
// solid rectangle.
//
// Kept here, unlike the rest of the old chrome.go (now
// internal/chrome) - this is genuinely tui-specific (the main shell's
// own background), not something internal/filemanager or any other
// future full-screen component needs to share.
type desktop struct {
	*tview.Box
	lines     []string
	textFg    tcell.Color
	patternFg tcell.Color
}

func newDesktop() *desktop {
	return &desktop{Box: tview.NewBox()}
}

// SetColors sets the desktop's background color, the banner text's
// foreground color, and the (dimmer) foreground color used for the
// dithered texture surrounding it.
func (d *desktop) SetColors(bg, textFg, patternFg tcell.Color) *desktop {
	d.SetBackgroundColor(bg)
	d.textFg = textFg
	d.patternFg = patternFg
	return d
}

// SetLines sets the centered lines of the banner shown over the desktop
// texture.
func (d *desktop) SetLines(lines ...string) *desktop {
	d.lines = lines
	return d
}

// Draw draws this primitive onto the screen.
func (d *desktop) Draw(screen tcell.Screen) {
	d.DrawForSubclass(screen, d)
	x, y, width, height := d.GetInnerRect()
	shade := tcell.StyleDefault.Foreground(d.patternFg).Background(d.GetBackgroundColor())
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			screen.SetContent(col, row, chrome.DesktopShade, nil, shade)
		}
	}
	top := y + height/2 - len(d.lines)/2
	for i, line := range d.lines {
		tview.Print(screen, line, x, top+i, width, tview.AlignCenter, d.textFg)
	}
}

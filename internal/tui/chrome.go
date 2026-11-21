package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Unicode glyphs used throughout the TUI to give menus, windows and the
// desktop the look of a 1990s DOS "window manager" toolkit (Borland's
// Turbo Vision, used by Turbo Pascal/C++, and Norton Commander): a system
// box in each window's top-left corner, small icons next to menu entries,
// and a dithered desktop pattern. All of them are single-width glyphs from
// the WGL4/CP437-compatible symbol set so they render correctly in classic
// monospace console fonts.
const (
	iconFile    = "▭" // ▭ WHITE RECTANGLE - a blank sheet of paper.
	iconSetup   = "⚙" // ⚙ GEAR.
	iconHelp    = "ℹ" // ℹ INFORMATION SOURCE.
	iconExit    = "✕" // ✕ MULTIPLICATION X.
	iconInstall = "◐" // ◐ CIRCLE WITH LEFT HALF BLACK (a theme/contrast swatch).
	iconAbout   = "★" // ★ BLACK STAR.
	iconWindow  = "■" // ■ BLACK SQUARE, used as the window system/close box.
	iconSwatch  = "■" // ■ BLACK SQUARE, used as a per-theme color swatch.

	desktopShade = '▒' // ▒ MEDIUM SHADE - the classic dithered desktop texture.
	bannerRule   = '═' // ═ BOX DRAWINGS DOUBLE HORIZONTAL - a decorative banner rule.
)

// shadowStyle is the style used to darken the cells a floating window casts
// its drop shadow onto. The rune already on screen is kept (so whatever was
// drawn underneath - desktop texture or another window - still shows its
// shape), only its colors are replaced with black, exactly like Turbo
// Vision's monochrome window shadows.
var shadowStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlack)

// drawShadow paints the classic L-shaped drop shadow cast by a window
// occupying the rectangle (x, y, width, height): two columns along its
// right edge (offset one row down) and one row along its bottom edge
// (offset two columns right). This is the defining visual trait of
// Turbo-Vision-style floating windows.
func drawShadow(screen tcell.Screen, x, y, width, height int) {
	darken := func(col, row int) {
		mainc, combc, style, _ := screen.GetContent(col, row)
		_, _, attrs := style.Decompose()
		screen.SetContent(col, row, mainc, combc, shadowStyle.Attributes(attrs))
	}
	for row := y + 1; row <= y+height; row++ {
		darken(x+width, row)
		darken(x+width+1, row)
	}
	for col := x + 2; col < x+width+2; col++ {
		darken(col, y+height)
	}
}

// attachShadow makes a bordered box cast a drop shadow (see drawShadow)
// without otherwise changing how it draws its border, background or title.
func attachShadow(box *tview.Box) {
	box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		drawShadow(screen, x, y, width, height)
		return x + 1, y + 1, width - 2, height - 2
	})
}

// windowChrome turns a bordered box into a Turbo-Vision-style floating
// window: a "[■]" system/close box glyph and an icon+title label stamped
// into the top border line, plus the drop shadow from attachShadow. The
// box must already have a border; its own SetTitle is cleared since the
// title is now painted here instead.
//
// closeBoxGlyph is "[■]": the opening bracket is immediately followed by a
// non-ASCII, non-tag-syntax rune, so tview's "[style]" tag parser treats it
// as plain text (see [tview.Print]) rather than an (invalid) style tag -
// this is safe without needing [tview.Escape].
const closeBoxGlyph = "[" + iconWindow + "]"

func windowChrome(box *tview.Box, icon, title string, titleColor tcell.Color) {
	box.SetTitle("")
	box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		drawShadow(screen, x, y, width, height)
		if width >= 5 {
			tview.Print(screen, closeBoxGlyph, x+1, y, 3, tview.AlignLeft, tcell.ColorYellow)
		}
		label := strings.TrimSpace(icon + " " + title)
		if label != "" && width >= 9 {
			tview.Print(screen, " "+label+" ", x+4, y, width-5, tview.AlignLeft, titleColor)
		}
		return x + 1, y + 1, width - 2, height - 2
	})
}

// desktop is the application's background: a dithered Turbo-Vision-style
// texture with a centered banner. It replaces a plain TextView so the
// banner can be drawn on top of the texture instead of covering it with a
// solid rectangle.
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
			screen.SetContent(col, row, desktopShade, nil, shade)
		}
	}
	top := y + height/2 - len(d.lines)/2
	for i, line := range d.lines {
		tview.Print(screen, line, x, top+i, width, tview.AlignCenter, d.textFg)
	}
}

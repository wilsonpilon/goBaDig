// Package chrome holds the Turbo-Vision-style visual helpers shared
// by every full-screen TUI component (the main shell in internal/tui,
// and the file manager in internal/filemanager): a drop shadow, a
// "[■]" system-box + icon + title stamped into a window's border, and
// the Unicode glyph set used throughout for menus/icons.
//
// Extracted from internal/tui/chrome.go (2026-11-26) for the same
// reason internal/dialect was extracted from internal/cli that same
// day: internal/filemanager needs this chrome too, but internal/tui
// ALSO needs to import internal/filemanager (to embed it as a page,
// same pattern already used for internal/editor) - internal/
// filemanager importing internal/tui directly would be a cycle.
// internal/tui/chrome.go keeps only what's genuinely tui-specific (the
// desktop background primitive) and re-uses this package for
// everything else.
package chrome

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
	IconFile    = "▭" // ▭ WHITE RECTANGLE - a blank sheet of paper.
	IconEdit    = "✎" // ✎ LOWER RIGHT PENCIL - the Editor menu.
	IconFolder  = "▦" // ▦ SQUARE WITH ORTHOGONAL CROSSHATCH FILL - the file manager's own menu entry/directories.
	IconSetup   = "⚙" // ⚙ GEAR.
	IconHelp    = "ℹ" // ℹ INFORMATION SOURCE.
	IconExit    = "✕" // ✕ MULTIPLICATION X.
	IconInstall = "◐" // ◐ CIRCLE WITH LEFT HALF BLACK (a theme/contrast swatch).
	IconAbout   = "★" // ★ BLACK STAR.
	IconWindow  = "■" // ■ BLACK SQUARE, used as the window system/close box.
	IconSwatch  = "■" // ■ BLACK SQUARE, used as a per-theme color swatch.

	// The file manager's own file-category markers (internal/
	// filemanager): a user asked for files of interest to stand out
	// with both a stronger color AND a distinct icon, to make a
	// panel listing scannable at a glance - which file is YOUR
	// source, which is the classic listing, which is ready to run.
	IconSource  = "◆" // ◆ BLACK DIAMOND - BADIG Dignified source (.dmx/.dzx/.dzs), the "real" code.
	IconListing = "≡" // ≡ IDENTICAL TO - a classic ASCII listing (.amx/.azx/.azs).
	IconBinary  = "▣" // ▣ WHITE SQUARE CONTAINING BLACK SMALL SQUARE - a packaged/compiled artifact (.bmx/.p/.tap/.tzx/.dsk).

	DesktopShade = '▒' // ▒ MEDIUM SHADE - the classic dithered desktop texture.
	BannerRule   = '═' // ═ BOX DRAWINGS DOUBLE HORIZONTAL - a decorative banner rule.
)

// ShadowStyle is the style used to darken the cells a floating window casts
// its drop shadow onto. The rune already on screen is kept (so whatever was
// drawn underneath - desktop texture or another window - still shows its
// shape), only its colors are replaced with black, exactly like Turbo
// Vision's monochrome window shadows.
var ShadowStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlack)

// DrawShadow paints the classic L-shaped drop shadow cast by a window
// occupying the rectangle (x, y, width, height): two columns along its
// right edge (offset one row down) and one row along its bottom edge
// (offset two columns right). This is the defining visual trait of
// Turbo-Vision-style floating windows.
func DrawShadow(screen tcell.Screen, x, y, width, height int) {
	darken := func(col, row int) {
		mainc, combc, style, _ := screen.GetContent(col, row)
		_, _, attrs := style.Decompose()
		screen.SetContent(col, row, mainc, combc, ShadowStyle.Attributes(attrs))
	}
	for row := y + 1; row <= y+height; row++ {
		darken(x+width, row)
		darken(x+width+1, row)
	}
	for col := x + 2; col < x+width+2; col++ {
		darken(col, y+height)
	}
}

// AttachShadow makes a bordered box cast a drop shadow (see DrawShadow)
// without otherwise changing how it draws its border, background or title.
func AttachShadow(box *tview.Box) {
	box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		DrawShadow(screen, x, y, width, height)
		return x + 1, y + 1, width - 2, height - 2
	})
}

// WindowChrome turns a bordered box into a Turbo-Vision-style floating
// window: a "[■]" system/close box glyph and an icon+title label stamped
// into the top border line, plus the drop shadow from AttachShadow. The
// box must already have a border; its own SetTitle is cleared since the
// title is now painted here instead.
//
// CloseBoxGlyph is "[■]": the opening bracket is immediately followed by a
// non-ASCII, non-tag-syntax rune, so tview's "[style]" tag parser treats it
// as plain text (see [tview.Print]) rather than an (invalid) style tag -
// this is safe without needing [tview.Escape]. Exported so tests driving a
// tcell.SimulationScreen can assert a window's chrome actually rendered.
const CloseBoxGlyph = "[" + IconWindow + "]"

func WindowChrome(box *tview.Box, icon, title string, titleColor tcell.Color) {
	box.SetTitle("")
	box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		DrawShadow(screen, x, y, width, height)
		if width >= 5 {
			tview.Print(screen, CloseBoxGlyph, x+1, y, 3, tview.AlignLeft, tcell.ColorYellow)
		}
		label := strings.TrimSpace(icon + " " + title)
		if label != "" && width >= 9 {
			tview.Print(screen, " "+label+" ", x+4, y, width-5, tview.AlignLeft, titleColor)
		}
		return x + 1, y + 1, width - 2, height - 2
	})
}

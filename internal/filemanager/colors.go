package filemanager

import "github.com/gdamore/tcell/v2"

// Colors is the palette the file manager renders with - handed in by
// internal/tui from the TUI's own active theme (internal/theme),
// instead of a fixed scheme of its own. Previously this component had
// a hardcoded yellow-on-black/navy look (closer to a 1979 BASIC
// listing than the rest of goBaDig's retro-but-themed chrome) -
// changed 2026-11-26 after the user pointed out both the visual
// mismatch and a real readability bug: unselectable rows fell back to
// tview.Table's implicit "invert" selection style, which combined
// badly with a cell that had no explicit BackgroundColor set,
// producing black-on-black text for some directory entries (still
// navigable - you could tell you were "in" one - just invisible).
// SetSelectedStyle below (see newPanel) fixes that directly; using
// the active theme's own palette additionally keeps the look
// consistent with Setup > Install's dialogs and the rest of the TUI.
type Colors struct {
	PanelBackground tcell.Color // panel fill (theme Surface)
	FileText        tcell.Color // normal (non-directory, non-recognized) entries (theme Foreground)
	DirectoryText   tcell.Color // directory entries, an accent (theme Primary)
	SourceText      tcell.Color // BADIG Dignified source (.dmx/.dzx/.dzs) (theme Primary)
	ClassicText     tcell.Color // classic ASCII listing (.amx/.azx/.azs) (theme Hover)
	BinaryText      tcell.Color // compiled/packaged artifact (.bmx/.p/.tap/.tzx/.dsk) (theme Error)
	InactiveBorder  tcell.Color // the non-focused panel's border (theme Foreground)
	ActiveBorder    tcell.Color // the focused panel's border (theme Primary)
	SelectionBg     tcell.Color // highlighted row background (theme Selection)
	SelectionFg     tcell.Color // highlighted row text - fixed white, same choice
	// already proven for the TUI's own menu dropdowns (tui.go's
	// list.SetSelectedTextColor(colors.desktopFg)): a saturated
	// Selection color from any theme reads fine under plain white.
	DialogBackground tcell.Color // dialog/list window fill (theme Surface)
	DialogText       tcell.Color // dialog/list body text (theme Foreground)
	DialogBorder     tcell.Color // dialog/list border + buttons (theme Primary)
	StatusBg         tcell.Color // status bar background - fixed silver,
	StatusFg         tcell.Color // status bar text - fixed black,
	// matching the TUI's own top menu bar/bottom status bar exactly
	// (tui.go's applyPalette: colors.menuBg/menuFg are themselves
	// fixed Silver/Black, never theme-driven).
}

// DefaultColors is used only where no theme is available yet (tests
// constructing a FileManager directly) - the same fixed Silver/Black
// bars plus a neutral navy/white/yellow panel scheme, close to this
// component's original look, so a test doesn't need a *theme.Theme.
func DefaultColors() Colors {
	return Colors{
		PanelBackground:  tcell.ColorBlack,
		FileText:         tcell.ColorWhite,
		DirectoryText:    tcell.ColorYellow,
		SourceText:       tcell.ColorYellow,
		ClassicText:      tcell.ColorAqua,
		BinaryText:       tcell.ColorGreen,
		InactiveBorder:   tcell.ColorWhite,
		ActiveBorder:     tcell.ColorYellow,
		SelectionBg:      tcell.ColorNavy,
		SelectionFg:      tcell.ColorWhite,
		DialogBackground: tcell.ColorNavy,
		DialogText:       tcell.ColorWhite,
		DialogBorder:     tcell.ColorSilver,
		StatusBg:         tcell.ColorSilver,
		StatusFg:         tcell.ColorBlack,
	}
}

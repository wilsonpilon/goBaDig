package tui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// TestFileManagerEntriesAreNotBlackOnBlack is a regression test
// reported directly by the user: the file manager's own fixed yellow-
// on-black scheme clashed with tview.Table's default "invert"
// selection style (used whenever no SetSelectedStyle is configured),
// which combines a cell's text color with its (never explicitly set)
// per-cell BackgroundColor - producing black-on-black text for a
// selected directory entry: present, navigable ("eu entro neles e dá
// pra saber que diretório é"), but invisible ("ficam preto com
// preto"). Fixed two ways: internal/filemanager.panel now calls
// table.SetSelectedStyle explicitly (panel.go's applyColors), and the
// whole component renders with the active theme's own palette
// (internal/filemanager.Colors) instead of a fixed scheme, matching
// the rest of the TUI's dialogs.
//
// Same whole-screen-scan technique as
// TestSetupMenuUnselectedItemsAreNotBlackOnBlack: not tied to precise
// coordinates, and catches this class of bug anywhere on screen. The
// file manager's own two panels always have at least one directory
// entry selected by default (the ".." pseudo-entry, row 0), so no
// fixture directory is needed to reproduce the exact scenario
// reported.
func TestFileManagerEntriesAreNotBlackOnBlack(t *testing.T) {
	screen, app, errCh := startSimulatedTUI(t)

	press := func(key tcell.Key, r rune, mod tcell.ModMask) {
		screen.InjectKey(key, r, mod)
		time.Sleep(30 * time.Millisecond)
	}
	// Navigate File -> Editor -> Setup -> Help -> Files, open the
	// menu, select "Open" (same sequence as
	// TestFilesMenuOpensAndClosesTheFileManager).
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)

	contents := dumpScreenText(screen)
	cells, width, _ := screen.GetContents()
	for i, cell := range cells {
		if len(cell.Runes) == 0 || cell.Runes[0] == ' ' {
			continue
		}
		fg, bg, _ := cell.Style.Decompose()
		if fg == bg {
			row, col := i/width, i%width
			t.Fatalf("cell (row=%d, col=%d) rune=%q has matching foreground and background (%v) - invisible text.\nScreen:\n%s",
				row, col, string(cell.Runes), fg, contents)
		}
	}

	stopSimulatedTUI(t, app, errCh)
}

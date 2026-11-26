package tui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/theme"
)

// TestFileManagerEntriesAreNotBlackOnBlack is an end-to-end regression
// test for a readability bug reported directly by the user, in two
// rounds: first "preto com preto" (black-on-black) on a selected
// directory row, then - after a fix that only addressed the SELECTED
// row's own style - "cinza com fundo cinza" (gray-on-gray) on
// ordinary rows.
//
// Two real causes:
//
//  1. No table.SetSelectedStyle was configured on the panels, so
//     tview.Table fell back to inverting the selected cell's text
//     color against its (never set) per-cell BackgroundColor - no
//     guarantee of contrast. Fixed with an explicit SetSelectedStyle
//     (internal/filemanager/panel.go's applyColors).
//  2. tview.NewTableCell seeds cell.Style to a concrete, non-default
//     value (tview's own global white-on-black), and
//     TableCell.SetTextColor only mutates that style's FOREGROUND
//     component in place when cell.Style != tcell.StyleDefault -
//     which is always, straight from the constructor. The background
//     stays exactly as the constructor left it (tview's own
//     hardcoded black) unless SetBackgroundColor is ALSO called. The
//     actually-shipped first fix called only SetTextColor per cell,
//     so text correctly picked up the active theme's foreground, but
//     sat on a plain black box the theme never chose - "gray on
//     gray" on a theme whose foreground happens to be muted. Fixed by
//     building the whole style (fg AND bg) with SetStyle instead
//     (internal/filemanager/panel.go's render - see
//     internal/filemanager/panel_test.go for a unit test reading the
//     TableCell's own Style field directly, which is what actually
//     found and confirmed this root cause: a screen-level dump alone
//     wasn't specific enough to catch it reliably).
//
// This end-to-end version drives the real TUI (same path a user
// takes: Files > Open) and checks an actual rendered screen cell's
// colors against the REAL, persisted active theme (config.
// OpenDefault - not a fixture), both foreground AND background, for
// an UNSELECTED row - "tui.go" (this package's own source file,
// always present in the file manager's starting directory, never
// sorts to the selected row 0).
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

	// run() reads the REAL, persisted active TUI theme via
	// config.OpenDefault() (the actual user's settings.db under
	// os.UserConfigDir(), not a fixture) - "tui-vscode-dark" is only
	// the fallback EnsureActiveTheme seeds a brand new database with,
	// never assume it's still active. Open the same store the same
	// way and ask it directly, instead of hardcoding a theme ID.
	store, err := config.OpenDefault()
	if err != nil {
		t.Fatalf("open default settings store: %v", err)
	}
	defer store.Close()
	if err := store.InstallThemes(theme.CatalogFor(theme.ModeTUI)); err != nil {
		t.Fatalf("install TUI theme catalog: %v", err)
	}
	activeTheme, err := store.EnsureActiveTheme(theme.ModeTUI, "tui-vscode-dark")
	if err != nil {
		t.Fatalf("load active TUI theme: %v", err)
	}
	wantFg := tcell.GetColor(activeTheme.Palette.Foreground)
	wantBg := tcell.GetColor(activeTheme.Palette.Surface)

	cells, width, height := screen.GetContents()
	runes := []rune("tui.go")
	foundEntry := false
	for row := 0; row < height; row++ {
		for col := 0; col <= width-len(runes); col++ {
			matches := true
			for k, r := range runes {
				if len(cells[row*width+col+k].Runes) == 0 || cells[row*width+col+k].Runes[0] != r {
					matches = false
					break
				}
			}
			if !matches {
				continue
			}
			foundEntry = true
			fg, bg, _ := cells[row*width+col].Style.Decompose()
			if fg != wantFg {
				t.Fatalf(`"tui.go" entry (row=%d, col=%d) foreground = %v, want the theme's Foreground color %v`,
					row, col, fg, wantFg)
			}
			if bg != wantBg {
				t.Fatalf(`"tui.go" entry (row=%d, col=%d) background = %v, want the theme's Surface color %v (fell back to tview's own hardcoded default - the exact mechanism behind the reported gray-on-gray bug)`,
					row, col, bg, wantBg)
			}
		}
	}
	if !foundEntry {
		t.Fatalf(`"tui.go" file entry not found anywhere on screen.\nScreen:\n%s`, dumpScreenText(screen))
	}

	stopSimulatedTUI(t, app, errCh)
}

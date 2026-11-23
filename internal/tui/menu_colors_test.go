package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestSetupMenuUnselectedItemsAreNotBlackOnBlack is a regression test
// reported directly by the user: every unselected item in an open
// menu dropdown (Setup, in this case - the only one with more than
// one item at the time this was filed) rendered as black text on a
// black background, invisible until navigated onto. Root cause:
// tview.List's own SetMainTextColor only ever touches the
// FOREGROUND half of its internal mainTextStyle - the BACKGROUND
// half stays at tview.NewList()'s own seeded default
// (Styles.PrimitiveBackgroundColor, black), unrelated to this menu's
// own SetBackgroundColor call, and List.Draw paints each unselected
// row's text with that whole (wrong) style, overwriting the Box's
// own correctly-colored background fill exactly at those character
// cells. Fixed with SetMainTextStyle (sets both halves at once).
//
// Scans the WHOLE simulated screen for any non-space glyph whose
// cell has identical foreground and background - true of any
// invisible-text bug anywhere on screen, not just this one menu, and
// not tied to precise coordinates that would make the test fragile.
func TestSetupMenuUnselectedItemsAreNotBlackOnBlack(t *testing.T) {
	screen, app, errCh := startSimulatedTUI(t)

	press := func(key tcell.Key, r rune, mod tcell.ModMask) {
		screen.InjectKey(key, r, mod)
		time.Sleep(30 * time.Millisecond)
	}
	press(tcell.KeyRight, 0, tcell.ModNone) // active menu: Editor
	press(tcell.KeyRight, 0, tcell.ModNone) // active menu: Setup
	press(tcell.KeyEnter, 0, tcell.ModNone) // open Setup dropdown (5 items; only "Install" is selected)

	contents := dumpScreenText(screen)
	left, right, top, bottom := dropdownInteriorBounds(t, contents)

	cells, width, _ := screen.GetContents()
	for row := top; row <= bottom; row++ {
		for col := left; col <= right; col++ {
			cell := cells[row*width+col]
			if len(cell.Runes) == 0 || cell.Runes[0] == ' ' {
				continue
			}
			fg, bg, _ := cell.Style.Decompose()
			if fg == bg {
				t.Fatalf("cell (row=%d, col=%d) rune=%q has matching foreground and background (%v) - invisible text.\nScreen:\n%s",
					row, col, string(cell.Runes), fg, contents)
			}
		}
	}

	stopSimulatedTUI(t, app, errCh)
}

// dropdownInteriorBounds locates the Setup dropdown's own box (drawn
// with "║" side borders and a "╚...╝" bottom border, the same chrome
// windowChrome/tview's SetBorder always draw) and returns the column/
// row range strictly INSIDE it - excluding this box lets a scan avoid
// the desktop's own "▒" shading pattern just outside, which legitimately
// paints its glyph with matching foreground/background (the pattern
// comes from the glyph's own shape, not a color contrast) and would
// otherwise be a false positive for an "invisible text" check.
func dropdownInteriorBounds(t *testing.T, contents string) (left, right, top, bottom int) {
	t.Helper()
	rows := strings.Split(contents, "\n")
	top, bottom = -1, -1
	for i, line := range rows {
		if strings.Contains(line, "⚙ Setup") {
			top = i + 1 // first item row is right below the title bar
		}
		if strings.Contains(line, "╚") {
			bottom = i - 1 // last item row is right above the bottom border
			break
		}
	}
	if top < 0 || bottom < 0 || top > bottom {
		t.Fatalf("could not locate the Setup dropdown's own box on screen:\n%s", contents)
	}
	// Column indices must be RUNE counts (one per screen cell), not
	// byte offsets - dumpScreenText writes one rune per column, and
	// box-drawing characters like "║" are multi-byte in UTF-8, so a
	// byte-based strings.Index here would misalign every later column
	// lookup into cells[row*width+col].
	runes := []rune(rows[top])
	left, right = -1, -1
	for i, r := range runes {
		if r != '║' {
			continue
		}
		if left < 0 {
			left = i
		}
		right = i
	}
	if left < 0 || right <= left {
		t.Fatalf("could not locate the Setup dropdown's left/right borders on row %d:\n%s", top, contents)
	}
	return left + 1, right - 1, top, bottom
}

// TestSetupMenuSelectedItemHighlightSpansTheFullMenuWidth is the
// second half of the same user report: the selected item's
// highlighted background stopped right after its own text instead of
// filling the whole menu width, leaving a ragged, partially-colored
// bar (e.g. a 10-character-wide menu with a 6-character label left
// the remaining 4 columns showing the plain menu background instead
// of the highlight). tview.List's own SetHighlightFullLine(true) is
// exactly tview's knob for this - checked here by comparing the
// background a few columns past the selected "Install" label's own
// text against the background AT the label itself; they must match.
func TestSetupMenuSelectedItemHighlightSpansTheFullMenuWidth(t *testing.T) {
	screen, app, errCh := startSimulatedTUI(t)

	press := func(key tcell.Key, r rune, mod tcell.ModMask) {
		screen.InjectKey(key, r, mod)
		time.Sleep(30 * time.Millisecond)
	}
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone) // "Install" is the default-selected (currentItem 0) item

	contents := dumpScreenText(screen)
	rows := strings.Split(contents, "\n")
	row, labelStart, labelEnd := -1, -1, -1
	label := []rune("Install")
	for r, line := range rows {
		// Column indices must be RUNE counts (one per screen cell),
		// not byte offsets - dumpScreenText writes one rune per
		// column, and the row has multi-byte box-drawing characters
		// (▒, ║, ◐) before "Install", so a byte-based strings.Index
		// would misalign the column lookup into
		// cells[row*width+col] below.
		runes := []rune(line)
		for i := 0; i+len(label) <= len(runes); i++ {
			if string(runes[i:i+len(label)]) == string(label) {
				row, labelStart = r, i
				labelEnd = i + len(label)
				break
			}
		}
		if row >= 0 {
			break
		}
	}
	if row < 0 {
		t.Fatalf("could not find the selected \"Install\" label on screen:\n%s", contents)
	}

	cells, width, _ := screen.GetContents()
	labelBg := cellBackground(t, cells, width, row, labelStart)
	// A few columns past the label's own text, still comfortably
	// inside the dropdown's own box (computed at 23 columns wide for
	// Setup's longest item, "Basic Dignified") and short of its
	// border.
	paddingBg := cellBackground(t, cells, width, row, labelEnd+3)
	if paddingBg != labelBg {
		t.Fatalf("background past the selected label's own text = %v, want it to match the label's own background %v (highlight should span the full menu width, not just the text)\nScreen:\n%s",
			paddingBg, labelBg, contents)
	}

	stopSimulatedTUI(t, app, errCh)
}

func cellBackground(t *testing.T, cells []tcell.SimCell, width, row, col int) tcell.Color {
	t.Helper()
	idx := row*width + col
	if idx < 0 || idx >= len(cells) {
		t.Fatalf("cell (row=%d, col=%d) is out of bounds (width=%d, total=%d)", row, col, width, len(cells))
	}
	_, bg, _ := cells[idx].Style.Decompose()
	return bg
}

// startSimulatedTUI boots run() against a fresh simulation screen,
// same setup every TUI screen-driving test in this package repeats -
// factored out here since this file adds two more such tests.
func startSimulatedTUI(t *testing.T) (tcell.SimulationScreen, *tview.Application, chan error) {
	t.Helper()
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("init simulation screen: %v", err)
	}
	screen.SetSize(80, 24)

	appCh := make(chan *tview.Application, 1)
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(screen, func(a *tview.Application) { appCh <- a })
	}()

	select {
	case app := <-appCh:
		return screen, app, errCh
	case <-time.After(2 * time.Second):
		t.Fatal("run() never became ready")
		return nil, nil, nil
	}
}

func stopSimulatedTUI(t *testing.T, app *tview.Application, errCh chan error) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		app.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("app.Stop() did not return")
	}
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run() returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run() did not return after Stop()")
	}
}

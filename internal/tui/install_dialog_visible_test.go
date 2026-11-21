package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestSetupInstallDialogIsVisible is a regression test for a bug where
// Setup > Install's dialog never drew at all: .SetBackgroundColor(...),
// chained directly onto tview.NewGrid(), is inherited from *tview.Box
// (Grid has no override) and returns *tview.Box, so the ":=" assignment
// silently downgraded the dialog window's static type from *tview.Grid
// to *tview.Box. Box.Draw() runs and returns without ever positioning
// or drawing its children, so the dialog's form and chrome never
// appeared on screen - no panic, no log, no hang. TestSetupInstallFreeze
// above could not have caught this: the application kept responding to
// input throughout, since nothing was actually deadlocked. This test
// instead reads back the simulated screen's real contents and checks
// that the dialog's window chrome (its close box and title, painted by
// windowChrome in chrome.go) is actually there.
func TestSetupInstallDialogIsVisible(t *testing.T) {
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

	var app *tview.Application
	select {
	case app = <-appCh:
	case <-time.After(2 * time.Second):
		t.Fatal("run() never became ready")
	}

	press := func(key tcell.Key, r rune, mod tcell.ModMask) {
		screen.InjectKey(key, r, mod)
		time.Sleep(30 * time.Millisecond)
	}

	// Navigate File -> Editor -> Setup, open the menu, select "Install" ->
	// showThemePicker().
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)

	contents := dumpScreenText(screen)
	if !strings.Contains(contents, "Install TUI Theme") {
		t.Fatalf("Install dialog title not found on screen; dialog is invisible.\nScreen contents:\n%s", contents)
	}
	if !strings.Contains(contents, closeBoxGlyph) {
		t.Fatalf("Install dialog close box %q not found on screen; dialog is invisible.\nScreen contents:\n%s", closeBoxGlyph, contents)
	}

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

// dumpScreenText renders a simulation screen's current contents as plain
// text, one line per row, so tests can make substring assertions about
// what would actually be visible to a user.
func dumpScreenText(screen tcell.SimulationScreen) string {
	cells, width, height := screen.GetContents()
	var b strings.Builder
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			cell := cells[row*width+col]
			if len(cell.Runes) == 0 {
				b.WriteRune(' ')
				continue
			}
			b.WriteRune(cell.Runes[0])
		}
		b.WriteByte('\n')
	}
	return b.String()
}

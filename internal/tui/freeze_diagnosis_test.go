package tui

import (
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestSetupInstallFreeze drives the TUI with a simulated terminal to
// reproduce the "freezes on Setup > Install" bug reported in OUTLINE.md /
// resume.txt, without needing a real terminal. If the application stops
// responding to further key events after opening the theme picker, this
// test times out and dumps all goroutine stacks to help pinpoint the
// deadlock.
func TestSetupInstallFreeze(t *testing.T) {
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

	// Navigate File -> Editor -> Setup, open the menu, select "Install".
	press(tcell.KeyRight, 0, tcell.ModNone) // active menu: Editor
	press(tcell.KeyRight, 0, tcell.ModNone) // active menu: Setup
	press(tcell.KeyEnter, 0, tcell.ModNone) // open Setup dropdown
	press(tcell.KeyEnter, 0, tcell.ModNone) // select Install -> showThemePicker()

	// From here on, the bug report says the UI stops responding. Prove it
	// one way or the other by sending one more keystroke (Escape, which
	// should cancel the theme picker dialog) on its own goroutine and
	// racing it against a timeout.
	escapeDone := make(chan struct{})
	go func() {
		press(tcell.KeyEscape, 0, tcell.ModNone)
		close(escapeDone)
	}()

	select {
	case <-escapeDone:
		t.Log("Escape after Setup > Install was processed without external hang")
	case <-time.After(2 * time.Second):
		dumpGoroutines(t)
		t.Fatal("UI stopped responding to input after Setup > Install (reproduced the freeze)")
	}

	// Whether or not Escape got through, make sure the application can
	// still be stopped cleanly; if Stop() itself never returns control to
	// Run(), that's the deadlock.
	stopDone := make(chan struct{})
	go func() {
		app.Stop()
		close(stopDone)
	}()

	select {
	case <-stopDone:
	case <-time.After(2 * time.Second):
		dumpGoroutines(t)
		t.Fatal("app.Stop() did not return - Application is wedged")
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run() returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		dumpGoroutines(t)
		t.Fatal("run() did not return after Stop() - Run() event loop is deadlocked")
	}
}

func dumpGoroutines(t *testing.T) {
	var b strings.Builder
	_ = pprof.Lookup("goroutine").WriteTo(&b, 2)
	t.Log("goroutine dump:\n" + b.String())
}

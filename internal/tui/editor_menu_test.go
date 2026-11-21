package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestEditorMenuOpensAndClosesWibEdit is a regression test for the
// integration of WibEdit (internal/editor) as the "Editor" page of the
// goBaDig TUI's shared *tview.Pages root. It exercises the same class of
// bug this project has already hit twice with this Pages/input-capture
// wiring: a page that silently never draws, or an app.SetInputCapture
// routing mistake that leaves the wrong widget receiving keys. It opens
// Editor > Open, asserts the editor's own chrome actually renders (not
// just that nothing hangs), then exits the editor (Escape, with no
// unsaved changes, falls straight through to its onClose callback) and
// asserts the desktop is visible again.
func TestEditorMenuOpensAndClosesWibEdit(t *testing.T) {
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

	// Navigate File -> Editor, open the menu, select "Open".
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)

	contents := dumpScreenText(screen)
	if !strings.Contains(contents, "Block") || !strings.Contains(contents, "Untitled") {
		t.Fatalf("WibEdit chrome not found on screen after Editor > Open.\nScreen contents:\n%s", contents)
	}

	// A fresh, unchanged document closes immediately on Escape (no "save
	// changes?" prompt), returning control to goBaDig via the onClose
	// callback passed to edit.NewEditor.
	press(tcell.KeyEscape, 0, tcell.ModNone)

	contents = dumpScreenText(screen)
	if strings.Contains(contents, "Untitled") {
		t.Fatalf("WibEdit chrome still on screen after Escape; editor did not close.\nScreen contents:\n%s", contents)
	}
	if !strings.Contains(contents, "goBaDig") {
		t.Fatalf("goBaDig desktop not visible again after closing the editor.\nScreen contents:\n%s", contents)
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

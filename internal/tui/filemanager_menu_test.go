package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestFilesMenuOpensAndClosesTheFileManager is the same class of
// regression test as TestEditorMenuOpensAndClosesWibEdit, for the new
// "Files" menu wiring internal/filemanager's two-panel component into
// the shared *tview.Pages root: opens Files > Open, asserts the file
// manager's own chrome (its status bar of shortcuts) actually
// renders, then exits it (Escape) and asserts the desktop is visible
// again.
func TestFilesMenuOpensAndClosesTheFileManager(t *testing.T) {
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

	// Navigate File -> Editor -> Setup -> Help -> Files, open the
	// menu, select "Open".
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)

	contents := dumpScreenText(screen)
	if !strings.Contains(contents, "Tab Painel") {
		t.Fatalf("file manager chrome not found on screen after Files > Open.\nScreen contents:\n%s", contents)
	}

	press(tcell.KeyEscape, 0, tcell.ModNone)

	contents = dumpScreenText(screen)
	if strings.Contains(contents, "Tab Painel") {
		t.Fatalf("file manager chrome still on screen after Escape; it did not close.\nScreen contents:\n%s", contents)
	}
	if !strings.Contains(contents, "goBaDig") {
		t.Fatalf("goBaDig desktop not visible again after closing the file manager.\nScreen contents:\n%s", contents)
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

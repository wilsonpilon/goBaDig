package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/goBaDig/internal/chrome"
)

// TestSetupOptionsPanelsRenderTheRightCategory confirms each new Setup
// menu item (Etapa C of the config panel, see OUTLINE.md Bloco 1/Fase
// 4) opens showOptionsPanel for its own category - checked by reading
// back the simulated screen's real contents (same technique as
// TestSetupInstallDialogIsVisible), not just that nothing deadlocked.
// internal/config and internal/cli already test SetOption/GetOption/
// EnsureOptionDefaults and the SET/SETTINGS/DIALECT command dispatch
// thoroughly on their own - this only needs to confirm the TUI wires
// the right category into the form it draws.
func TestSetupOptionsPanelsRenderTheRightCategory(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("init simulation screen: %v", err)
	}
	screen.SetSize(80, 30)

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

	// waitForText polls the screen until want appears or timeout
	// elapses, instead of trusting a single fixed sleep after the key
	// that is supposed to trigger it - under heavy system load (this
	// package's full test run alongside every other package's, each
	// driving its own tview.Application goroutine) a fixed 30ms isn't
	// always enough for this test's own chained Down/Down/Enter
	// sequence to have been fully processed yet.
	waitForText := func(want string) string {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		var contents string
		for time.Now().Before(deadline) {
			contents = dumpScreenText(screen)
			if strings.Contains(contents, want) {
				return contents
			}
			time.Sleep(20 * time.Millisecond)
		}
		return contents
	}

	// Navigate File -> Editor -> Setup, open the dropdown, move down
	// to "MSX" (items: 0=Install, 1=Basic Dignified, 2=MSX, ...).
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyRight, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone) // open Setup dropdown
	press(tcell.KeyDown, 0, tcell.ModNone)
	press(tcell.KeyDown, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone) // select "MSX"

	contents := waitForText("CONVERT_PRINT")
	for _, want := range []string{"MSX", "CONVERT_PRINT", "EMULATOR_MACHINE", chrome.CloseBoxGlyph} {
		if !strings.Contains(contents, want) {
			t.Fatalf("MSX options panel missing %q; screen contents:\n%s", want, contents)
		}
	}

	press(tcell.KeyEscape, 0, tcell.ModNone) // close the MSX panel

	// Reopen, move down to "ZX81" (index 3) this time.
	press(tcell.KeyEnter, 0, tcell.ModNone)
	press(tcell.KeyDown, 0, tcell.ModNone)
	press(tcell.KeyDown, 0, tcell.ModNone)
	press(tcell.KeyDown, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone) // select "ZX81"

	contents = waitForText("EMULATOR_PATH")
	for _, want := range []string{"ZX81", "TOKENIZE", "EMULATOR_PATH"} {
		if !strings.Contains(contents, want) {
			t.Fatalf("ZX81 options panel missing %q; screen contents:\n%s", want, contents)
		}
	}
	if strings.Contains(contents, "EMULATOR_MACHINE") {
		t.Fatalf("ZX81 options panel must not show MSX-only fields; screen contents:\n%s", contents)
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

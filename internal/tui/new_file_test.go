package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestF3CreatesAFileAndOpensItInTheEditor drives the whole "File
// Manager as control center" loop the user asked for end-to-end
// through real simulated keystrokes (not a direct Go call into
// internal/filemanager, which every other filemanager test uses and
// which bypasses tview.Form entirely): open File > File Manager,
// press F3, type a filename, confirm it through the dialog, and
// check the file actually exists and WibEdit opened with it.
//
// Confirms a real quirk of tview.Form worth recording: pressing Enter
// while an InputField is focused does not immediately submit the
// form. The field's own InputHandler treats Enter as "I'm done with
// this field" (calling its finished callback, which Form uses to
// advance focus), and a SECOND Enter is needed to actually move focus
// onto the first button, with a THIRD to click it - three Enters
// total from the input field to a submitted dialog, not one or two,
// confirmed here by trial (not Tab/Down, which do nothing useful from
// a focused InputField in this tview version).
func TestF3CreatesAFileAndOpensItInTheEditor(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

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
	typeText := func(s string) {
		for _, r := range s {
			press(tcell.KeyRune, r, tcell.ModNone)
		}
	}

	// File is the active menu at startup; open it and select "File
	// Manager" (item 0).
	press(tcell.KeyEnter, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)

	press(tcell.KeyF3, 0, tcell.ModNone)
	typeText("new.dmx")
	press(tcell.KeyEnter, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)

	// Poll rather than trust the fixed per-keystroke sleep above for
	// this final check: under heavy system load (this package's full
	// test run drives several of its own tview.Application goroutines
	// at once) three chained Enters don't always land within a fixed
	// window - same reasoning as options_panel_test.go's own
	// waitForText helper.
	deadline := time.Now().Add(2 * time.Second)
	var statErr error
	for time.Now().Before(deadline) {
		if _, statErr = os.Stat(filepath.Join(dir, "new.dmx")); statErr == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if statErr != nil {
		t.Fatalf("expected new.dmx to have been created: %v", statErr)
	}

	contents := dumpScreenText(screen)
	if !strings.Contains(contents, "Block") {
		t.Fatalf("expected WibEdit to have opened with the new file.\nScreen:\n%s", contents)
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

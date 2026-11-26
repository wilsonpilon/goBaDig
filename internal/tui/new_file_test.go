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

// submitFormDialog presses Enter until the dialog named by
// closedWhenGone (a substring of its title/chrome) is no longer on
// screen, up to a handful of tries. Needed because tview.Form's Enter
// handling isn't "submit on the first press": pressing Enter while an
// InputField is focused only finishes that FIELD (advancing focus to
// the next item, per InputField.InputHandler's "finish" callback) -
// it can take two or more Enters total (field -> button focus ->
// button click) before the dialog actually closes, and the exact
// count isn't a reliable constant to hardcode.
func submitFormDialog(t *testing.T, screen tcell.SimulationScreen, press func(tcell.Key, rune, tcell.ModMask), closedWhenGone string) {
	t.Helper()
	isGone := func() bool { return !strings.Contains(dumpScreenText(screen), closedWhenGone) }
	for i := 0; i < 5; i++ {
		if isGone() {
			return
		}
		press(tcell.KeyEnter, 0, tcell.ModNone)
		deadline := time.Now().Add(300 * time.Millisecond)
		for time.Now().Before(deadline) {
			if isGone() {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	if !isGone() {
		t.Fatalf("dialog %q still on screen after 5 Enters.\nScreen:\n%s", closedWhenGone, dumpScreenText(screen))
	}
}

// TestF3CreatesAFileAndOpensItInTheEditor drives the whole "File
// Manager as control center" loop the user asked for end-to-end
// through real simulated keystrokes (not a direct Go call into
// internal/filemanager, which every other filemanager test uses and
// which bypasses tview.Form entirely): open File > File Manager,
// press F3, type a filename, confirm it through the dialog, and
// check the file actually exists and WibEdit opened with it.
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
	submitFormDialog(t, screen, press, "Novo Arquivo")

	// Poll rather than trust the fixed per-keystroke sleep above for
	// this final check: under heavy system load (this package's full
	// test run drives several of its own tview.Application goroutines
	// at once) a chain of key presses doesn't always land within a
	// fixed window - same reasoning as options_panel_test.go's own
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

// TestClosingTheEditorReturnsToTheFileManagerItWasOpenedFrom is a
// regression test reported directly by the user: "ao ir para o
// editor, voltando a barra de status fica bagunçada, não mostrando
// mais as teclas de atalho" - closing the editor after opening a file
// from the file manager (F3/F4) landed on the DESKTOP instead of back
// on the file manager, because closeEditor always switched to the
// "app" page unconditionally. What looked like "the status bar got
// messed up" was really the desktop's own, different status bar
// ("F10 Menu...") showing instead of the file manager's shortcut
// list ("Tab Painel..."), because the wrong page was visible. Fixed
// with editorReturnPage, set to "filemanager" by the file manager's
// own onOpenInEditor callback and back to "app" by the Editor menu's
// own "Open" action.
//
// Uses F4 (edit an already-existing file) rather than F3 (create a
// new one): F3 goes through a tview.Form dialog first, whose exact
// Enter-count-to-submit isn't a reliable constant (confirmed during
// development - a fixed or polled Enter sequence occasionally
// overshoots by one press once the dialog closes, landing inside the
// now-focused editor and typing a newline instead). F4 exercises the
// exact same editorReturnPage code path with no dialog in the way.
func TestClosingTheEditorReturnsToTheFileManagerItWasOpenedFrom(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	if err := os.WriteFile(filepath.Join(dir, "roundtrip.dmx"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

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
	waitFor := func(want string) string {
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

	// File > File Manager; row 0 is always "..", selected by default -
	// Down once selects roundtrip.dmx (the only other entry). F4 opens it.
	press(tcell.KeyEnter, 0, tcell.ModNone)
	press(tcell.KeyEnter, 0, tcell.ModNone)
	press(tcell.KeyDown, 0, tcell.ModNone)
	press(tcell.KeyF4, 0, tcell.ModNone)

	if contents := waitFor("Block"); !strings.Contains(contents, "Block") {
		t.Fatalf("expected WibEdit to have opened.\nScreen:\n%s", contents)
	}

	// A fresh, unchanged document closes immediately on Escape (no
	// "save changes?" prompt) - same as TestEditorMenuOpensAndClosesWibEdit.
	press(tcell.KeyEscape, 0, tcell.ModNone)

	contents := waitFor("Tab Painel")
	if !strings.Contains(contents, "Tab Painel") {
		t.Fatalf("expected the file manager's own status bar after closing the editor, got the desktop's instead.\nScreen:\n%s", contents)
	}
	if strings.Contains(contents, "F10") {
		t.Fatalf("desktop status bar text found - landed on the desktop instead of the file manager.\nScreen:\n%s", contents)
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

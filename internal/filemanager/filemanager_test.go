package filemanager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/options"
)

func newTestStore(t *testing.T) *config.Store {
	t.Helper()
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func keyEvent(key tcell.Key) *tcell.EventKey {
	return tcell.NewEventKey(key, 0, tcell.ModNone)
}

func TestNewFileManagerStartsBothPanelsAtTheWorkingDirectory(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, nil)
	wantDir, _ := filepath.Abs(wd)
	if fm.panels[0].dir != wantDir || fm.panels[1].dir != wantDir {
		t.Fatalf("panels = (%q, %q), want both %q", fm.panels[0].dir, fm.panels[1].dir, wantDir)
	}
	if fm.active != 0 {
		t.Fatalf("active = %d, want 0", fm.active)
	}
}

func TestTabSwitchesTheActivePanel(t *testing.T) {
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, nil)
	if got := fm.Capture(keyEvent(tcell.KeyTab)); got != nil {
		t.Fatalf("Capture(Tab) = %v, want nil (consumed)", got)
	}
	if fm.active != 1 {
		t.Fatalf("active = %d, want 1 after Tab", fm.active)
	}
	fm.Capture(keyEvent(tcell.KeyTab))
	if fm.active != 0 {
		t.Fatalf("active = %d, want 0 after a second Tab", fm.active)
	}
}

func TestEnterDescendsIntoADirectoryAndBackspaceGoesBackUp(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, nil)
	if err := fm.current().load(parent); err != nil {
		t.Fatal(err)
	}
	fm.current().table.Select(indexOf(t, fm.current(), "child"), 0)

	fm.Capture(keyEvent(tcell.KeyEnter))
	wantChild, _ := filepath.Abs(child)
	if fm.current().dir != wantChild {
		t.Fatalf("after Enter on child, dir = %q, want %q", fm.current().dir, wantChild)
	}

	fm.Capture(keyEvent(tcell.KeyBackspace2))
	wantParent, _ := filepath.Abs(parent)
	if fm.current().dir != wantParent {
		t.Fatalf("after Backspace, dir = %q, want %q", fm.current().dir, wantParent)
	}
}

func indexOf(t *testing.T, p *panel, name string) int {
	t.Helper()
	for i, e := range p.entries {
		if e.name == name {
			return i
		}
	}
	t.Fatalf("entry %q not found in panel listing %v", name, p.entries)
	return -1
}

func TestEscClosesTheFileManager(t *testing.T) {
	closed := false
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, func() { closed = true })
	fm.Capture(keyEvent(tcell.KeyEsc))
	if !closed {
		t.Fatal("Esc did not call onClose")
	}
}

func TestEditSelectedCallsOnOpenInEditorWithTheSelectedPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "hello.dmx")
	if err := os.WriteFile(file, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var opened string
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), func(path string) { opened = path }, nil)
	if err := fm.current().load(dir); err != nil {
		t.Fatal(err)
	}
	fm.current().table.Select(indexOf(t, fm.current(), "hello.dmx"), 0)
	fm.Capture(keyEvent(tcell.KeyF4))
	want, _ := filepath.Abs(file)
	if opened != want {
		t.Fatalf("onOpenInEditor called with %q, want %q", opened, want)
	}
}

func TestF2CompilesTheSelectedFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "hello.dmx")
	if err := os.WriteFile(file, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, nil)
	if err := fm.current().load(dir); err != nil {
		t.Fatal(err)
	}
	fm.current().table.Select(indexOf(t, fm.current(), "hello.dmx"), 0)
	fm.Capture(keyEvent(tcell.KeyF2))

	if _, err := os.Stat(filepath.Join(dir, "hello.bmx")); err != nil {
		t.Fatalf("expected the compiled .bmx artifact to exist: %v", err)
	}
	if status := fm.status.GetText(true); status == statusHelp {
		t.Fatalf("status was not updated after compiling, still %q", status)
	}
}

// TestF2DispatchesAsmFilesToAsmsx confirms ".asm" is routed into the
// ASMSX branch (not dialect.CompileByExtension, which would reject it
// as an unrecognized dialect) - PATH is pointed at an empty directory
// so asmsx can never actually be found; CompileAsm's own real
// subprocess/output logic is covered separately, deeply, in
// internal/platforms/msx/asmsx_test.go, so this only needs to confirm
// the dispatch itself happened.
func TestF2DispatchesAsmFilesToAsmsx(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	dir := t.TempDir()
	file := filepath.Join(dir, "game.asm")
	if err := os.WriteFile(file, []byte("; nothing"), 0o644); err != nil {
		t.Fatal(err)
	}
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, nil)
	if err := fm.current().load(dir); err != nil {
		t.Fatal(err)
	}
	fm.current().table.Select(indexOf(t, fm.current(), "game.asm"), 0)
	fm.Capture(keyEvent(tcell.KeyF2))

	status := fm.status.GetText(true)
	if !strings.Contains(status, "asmsx") {
		t.Fatalf("status = %q, want it to mention asmsx (confirms .asm was dispatched to ASMSX, not CompileByExtension)", status)
	}
}

func TestF9ReportsAnErrorWhenNoEmulatorIsOnPathButStillBuildsTheArtifact(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	dir := t.TempDir()
	file := filepath.Join(dir, "hello.dmx")
	if err := os.WriteFile(file, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, nil)
	if err := fm.current().load(dir); err != nil {
		t.Fatal(err)
	}
	fm.current().table.Select(indexOf(t, fm.current(), "hello.dmx"), 0)
	fm.Capture(keyEvent(tcell.KeyF9))

	if _, err := os.Stat(filepath.Join(dir, "hello.dsk")); err != nil {
		t.Fatalf("expected the MSX test disk to still have been built: %v", err)
	}
	status := fm.status.GetText(true)
	if status == statusHelp {
		t.Fatal("status was not updated after F9 with no emulator on PATH")
	}
}

func TestF7CreatesADirectoryInTheActivePanel(t *testing.T) {
	dir := t.TempDir()
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, nil)
	if err := fm.current().load(dir); err != nil {
		t.Fatal(err)
	}
	fm.mkdir("novapasta")
	if info, err := os.Stat(filepath.Join(dir, "novapasta")); err != nil || !info.IsDir() {
		t.Fatalf("expected novapasta to have been created as a directory: %v", err)
	}
}

func TestCopySelectedCopiesAFileToTheOtherPanel(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	file := filepath.Join(dirA, "hello.dmx")
	if err := os.WriteFile(file, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, nil)
	if err := fm.panels[0].load(dirA); err != nil {
		t.Fatal(err)
	}
	if err := fm.panels[1].load(dirB); err != nil {
		t.Fatal(err)
	}
	fm.panels[0].table.Select(indexOf(t, fm.panels[0], "hello.dmx"), 0)
	fm.copySelected()

	if _, err := os.Stat(filepath.Join(dirB, "hello.dmx")); err != nil {
		t.Fatalf("expected hello.dmx to have been copied into dirB: %v", err)
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatalf("expected the original file in dirA to remain after copy: %v", err)
	}
}

func TestMoveSelectedMovesAFileAfterConfirmation(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	file := filepath.Join(dirA, "hello.dmx")
	if err := os.WriteFile(file, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, nil)
	if err := fm.panels[0].load(dirA); err != nil {
		t.Fatal(err)
	}
	if err := fm.panels[1].load(dirB); err != nil {
		t.Fatal(err)
	}
	fm.panels[0].table.Select(indexOf(t, fm.panels[0], "hello.dmx"), 0)
	fm.moveSelected()
	if !fm.dialogOpen {
		t.Fatal("moveSelected did not open a confirmation dialog")
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatal("expected the source file to still exist before confirmation")
	}
}

func TestCreateProjectThenListAppearsInProjectsList(t *testing.T) {
	store := newTestStore(t)
	parent := t.TempDir()
	fm := NewFileManager(tview.NewApplication(), store, DefaultColors(), nil, nil)
	fm.createProject("MeuJogo", parent, options.CategoryMSX)

	projects, err := store.ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].Name != "MeuJogo" {
		t.Fatalf("ListProjects() = %+v, want one project named MeuJogo", projects)
	}
	wantDir, _ := filepath.Abs(filepath.Join(parent, "MeuJogo"))
	if fm.current().dir != wantDir {
		t.Fatalf("active panel dir = %q, want %q (jumped to the new project's own subfolder)", fm.current().dir, wantDir)
	}
}

func TestCreateProjectCreatesItsOwnSubfolderInsideTheParent(t *testing.T) {
	store := newTestStore(t)
	parent := t.TempDir()
	fm := NewFileManager(tview.NewApplication(), store, DefaultColors(), nil, nil)
	fm.createProject("MeuJogo", parent, options.CategoryMSX)

	projectDir := filepath.Join(parent, "MeuJogo")
	if info, err := os.Stat(projectDir); err != nil || !info.IsDir() {
		t.Fatalf("expected %s to have been created as the project's own subfolder: %v", projectDir, err)
	}
	// The parent itself must stay clean - the project's files live in
	// its own subfolder, not scattered into whatever directory the
	// user was browsing (the user's own request).
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "MeuJogo" {
		t.Fatalf("parent directory = %v, want only the MeuJogo subfolder", entries)
	}
}

// TestCreateProjectStillShowsTheDirectoryWhenTheDatabaseStepFails is
// a regression test reported directly by the user: "ao criar um
// projeto novo, às vezes dá erro e não mostra os arquivos no file
// manager, mas entrando e saindo do programa o diretório está lá".
//
// Root cause: config.Store.CreateProject runs os.MkdirAll BEFORE its
// database insert - a failure at the database step (most commonly a
// duplicate project name, since "name" is that table's primary key
// and persists across restarts in the user's real settings.db) still
// leaves a real, newly-created directory on disk. The file manager's
// own createProject used to return as soon as store.CreateProject
// failed, without ever reloading the panel - so that directory
// stayed invisible in the CURRENT session until the whole program was
// restarted (which lists the real filesystem fresh, revealing it was
// there all along). Reproduced here with a duplicate name, the
// simplest real way to trigger the database step's own failure.
func TestCreateProjectStillShowsTheDirectoryWhenTheDatabaseStepFails(t *testing.T) {
	store := newTestStore(t)
	parent := t.TempDir()
	fm := NewFileManager(tview.NewApplication(), store, DefaultColors(), nil, nil)

	fm.createProject("MeuJogo", parent, options.CategoryMSX)
	if status := fm.status.GetText(true); strings.Contains(status, "erro") {
		t.Fatalf("first createProject: unexpected error status %q", status)
	}

	// Move the panel away to an UNRELATED directory - one that
	// genuinely doesn't contain "MeuJogo" - before recreating with the
	// same name. If createProject's own reload doesn't fire on the
	// error path below, the panel stays on this unrelated listing
	// instead of reflecting "parent" (where MeuJogo actually is), so
	// this test can tell the difference; reloading "parent" directly
	// here first would make both the buggy and fixed code look the
	// same, since the panel would already be showing MeuJogo before
	// the second createProject call even ran.
	elsewhere := t.TempDir()
	if err := fm.current().load(elsewhere); err != nil {
		t.Fatal(err)
	}

	// store.CreateProject's database insert fails (duplicate primary
	// key), but MeuJogo's own directory already exists from the first
	// call above.
	fm.createProject("MeuJogo", parent, options.CategoryMSX)

	status := fm.status.GetText(true)
	if !strings.Contains(status, "erro") {
		t.Fatalf("second createProject with a duplicate name: want an error status, got %q", status)
	}
	found := false
	for _, e := range fm.current().entries {
		if e.name == "MeuJogo" {
			found = true
		}
	}
	if !found {
		t.Fatalf("MeuJogo not found in the reloaded panel after the duplicate-name error: %+v", fm.current().entries)
	}
}

func TestCreateProjectWritesAMainStarterFileWithTheRightExtension(t *testing.T) {
	store := newTestStore(t)
	parent := t.TempDir()
	var opened string
	fm := NewFileManager(tview.NewApplication(), store, DefaultColors(), func(path string) { opened = path }, nil)
	fm.createProject("Jogo81", parent, options.CategoryZX81)

	want := filepath.Join(parent, "Jogo81", "main.dzx")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected a main.dzx starter file: %v", err)
	}
	wantAbs, _ := filepath.Abs(want)
	if opened != wantAbs {
		t.Fatalf("onOpenInEditor called with %q, want the starter file %q", opened, wantAbs)
	}
}

func TestCreateFileCreatesAnEmptyFileAndOpensItInTheEditor(t *testing.T) {
	dir := t.TempDir()
	var opened string
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), func(path string) { opened = path }, nil)
	if err := fm.current().load(dir); err != nil {
		t.Fatal(err)
	}
	fm.createFile("main.dmx")

	want := filepath.Join(dir, "main.dmx")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected main.dmx to have been created: %v", err)
	}
	wantAbs, _ := filepath.Abs(want)
	if opened != wantAbs {
		t.Fatalf("onOpenInEditor called with %q, want %q", opened, wantAbs)
	}
}

func TestCreateFileRejectsAnEmptyName(t *testing.T) {
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, nil)
	if err := fm.current().load(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	fm.createFile("")
	if status := fm.status.GetText(true); status == statusHelp {
		t.Fatal("status was not updated after createFile(\"\")")
	}
}

func TestCreateFileRefusesToOverwriteAnExistingFile(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "main.dmx")
	if err := os.WriteFile(existing, []byte("10 PRINT 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var opened string
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), func(path string) { opened = path }, nil)
	if err := fm.current().load(dir); err != nil {
		t.Fatal(err)
	}
	fm.createFile("main.dmx")

	contents, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "10 PRINT 1\n" {
		t.Fatalf("existing file was overwritten: %q", contents)
	}
	if opened != "" {
		t.Fatalf("onOpenInEditor should not have been called, got %q", opened)
	}
}

// TestSelectingADifferentRowResetsTheStatusBar is a regression test
// reported directly by the user: "a linha de status ainda não está
// sendo atualizada a cada opção usada, algumas alteram a barra de
// status e a mesma não volta" - some actions (F2 compile, F5 copy,
// F7 mkdir, etc.) set a one-off status message that never reverted,
// because the only place that ever reset it back to statusHelp was
// Enter/Backspace (navigating a directory). Arrow-key navigation
// within a panel never goes through FileManager.Capture at all -
// Up/Down/PageUp/PageDown aren't in its switch, so tview.Table's own
// InputHandler processes them directly - meaning a stale action
// message could sit on screen indefinitely once the user just moved
// the selection with the keyboard instead of pressing Enter/
// Backspace. Fixed with a SetSelectionChangedFunc on each panel
// (registered in NewFileManager), confirmed here by directly calling
// tview.Table's own Select (the same call arrow-key navigation
// ultimately makes) rather than simulating a real keypress.
func TestSelectingADifferentRowResetsTheStatusBar(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fm := NewFileManager(tview.NewApplication(), newTestStore(t), DefaultColors(), nil, nil)
	if err := fm.current().load(dir); err != nil {
		t.Fatal(err)
	}

	// Leave a stale action message on the status bar, exactly like F2
	// (compile) or F7 (mkdir) would.
	fm.setStatus(" compilado: a.amx + a.bmx")
	if status := fm.status.GetText(true); status != " compilado: a.amx + a.bmx" {
		t.Fatalf("setup: status = %q, want the stale compile message", status)
	}

	// Simulate arrow-key navigation: Table.Select is exactly what
	// Up/Down ultimately call (table.go's own InputHandler), and
	// Select always fires SetSelectionChangedFunc regardless of
	// whether it's a real keypress or a direct call.
	fm.current().table.Select(1, 0)

	if status := fm.status.GetText(true); status != statusHelp {
		t.Fatalf("status after moving the selection = %q, want it reset to the shortcut list %q", status, statusHelp)
	}
}

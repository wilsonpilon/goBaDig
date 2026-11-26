package filemanager

import (
	"os"
	"path/filepath"
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

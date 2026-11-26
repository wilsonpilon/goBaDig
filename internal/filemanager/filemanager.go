// Package filemanager implements a two-panel, Norton-Commander-style
// file manager for the TUI: browse the local filesystem in two side-
// by-side panels, open a file in the embedded WibEdit editor, compile
// or execute a BASIC source file through internal/dialect, do basic
// file operations (copy/move/mkdir/delete), and create or reopen a
// goBaDig "project" (internal/config's projects table).
//
// Modeled on internal/editor's own shape: FileManager embeds
// *tview.Pages (its "main" page plus any dialog overlay), is built
// with NewFileManager(app, store, onOpenInEditor, onClose), and
// exposes a Capture method for internal/tui's input-routing switch -
// the same three-point wiring already used for the editor.
package filemanager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/goBaDig/internal/config"
)

// FileManager is a full-screen TUI component: two directory panels,
// a status bar showing the available shortcuts and the last action's
// result, and its own dialog overlays (new project, project list,
// confirmations) layered on top via its embedded *tview.Pages.
type FileManager struct {
	*tview.Pages

	app            *tview.Application
	store          *config.Store
	onOpenInEditor func(path string)
	onClose        func()

	panels     [2]*panel
	active     int
	status     *tview.TextView
	dialogOpen bool
}

const statusHelp = " Tab Painel  Enter/Backspace Navega  F2 Compila  F4 Edita  F5 Copia  F6 Move  F7 Dir  F8 Apaga  F9 Executa  Ctrl+N Projeto  Ctrl+P Projetos  Esc Fecha"

// NewFileManager builds a ready-to-show file manager starting both
// panels at the current working directory. onOpenInEditor is called
// with the path of a file the user chose to edit (F4) - the caller
// (internal/tui) is expected to hand it to the already-embedded
// WibEdit editor and switch pages to it, exactly like the Editor menu
// already does. onClose is called when the user asks to leave the
// file manager (Esc/F10).
func NewFileManager(app *tview.Application, store *config.Store, onOpenInEditor func(path string), onClose func()) *FileManager {
	if onOpenInEditor == nil {
		onOpenInEditor = func(string) {}
	}
	if onClose == nil {
		onClose = func() {}
	}
	fm := &FileManager{
		Pages:          tview.NewPages(),
		app:            app,
		store:          store,
		onOpenInEditor: onOpenInEditor,
		onClose:        onClose,
	}
	fm.panels[0] = newPanel()
	fm.panels[1] = newPanel()
	start, err := os.Getwd()
	if err != nil {
		start = "."
	}
	_ = fm.panels[0].load(start)
	_ = fm.panels[1].load(start)

	fm.status = tview.NewTextView().SetDynamicColors(true)
	fm.status.SetBackgroundColor(tcell.ColorBlue)
	fm.status.SetTextColor(tcell.ColorWhite)
	fm.setStatus(statusHelp)

	body := tview.NewFlex().
		AddItem(fm.panels[0].table, 0, 1, true).
		AddItem(fm.panels[1].table, 0, 1, false)
	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(body, 0, 1, true).
		AddItem(fm.status, 1, 0, false)
	fm.AddPage("main", root, true, true)
	fm.setActive(0)
	return fm
}

func (fm *FileManager) setStatus(text string) {
	fm.status.SetText(text)
}

func (fm *FileManager) setActive(i int) {
	fm.active = i
	for idx, p := range fm.panels {
		if idx == i {
			p.table.SetBorderColor(tcell.ColorYellow)
		} else {
			p.table.SetBorderColor(tcell.ColorWhite)
		}
	}
	fm.app.SetFocus(fm.panels[i].table)
}

func (fm *FileManager) current() *panel { return fm.panels[fm.active] }
func (fm *FileManager) other() *panel   { return fm.panels[1-fm.active] }

// Capture is this component's key-routing hook, called from
// internal/tui's own app.SetInputCapture switch whenever the file
// manager page is the front page - same contract as editor.Editor's
// own Capture method.
func (fm *FileManager) Capture(event *tcell.EventKey) *tcell.EventKey {
	if fm.dialogOpen {
		return event
	}
	switch event.Key() {
	case tcell.KeyTab:
		fm.setActive(1 - fm.active)
		return nil
	case tcell.KeyEsc, tcell.KeyF10:
		fm.onClose()
		return nil
	case tcell.KeyEnter:
		fm.enterSelected()
		return nil
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		fm.goUp()
		return nil
	case tcell.KeyF2:
		fm.compileSelected()
		return nil
	case tcell.KeyF4:
		fm.editSelected()
		return nil
	case tcell.KeyF5:
		fm.copySelected()
		return nil
	case tcell.KeyF6:
		fm.moveSelected()
		return nil
	case tcell.KeyF7:
		fm.promptMkdir()
		return nil
	case tcell.KeyF8:
		fm.promptDelete()
		return nil
	case tcell.KeyF9:
		fm.executeSelected()
		return nil
	case tcell.KeyCtrlN:
		fm.promptNewProject()
		return nil
	case tcell.KeyCtrlP:
		fm.promptProjectList()
		return nil
	}
	return event
}

func (fm *FileManager) enterSelected() {
	p := fm.current()
	e, ok := p.selected()
	if !ok || !e.isDir {
		return
	}
	var next string
	if e.isParent() {
		next = filepath.Dir(p.dir)
	} else {
		next = filepath.Join(p.dir, e.name)
	}
	if err := p.load(next); err != nil {
		fm.setStatus(fmt.Sprintf(" erro: %v", err))
		return
	}
	fm.setStatus(statusHelp)
}

func (fm *FileManager) goUp() {
	p := fm.current()
	parent := filepath.Dir(p.dir)
	if parent == p.dir {
		return
	}
	if err := p.load(parent); err != nil {
		fm.setStatus(fmt.Sprintf(" erro: %v", err))
		return
	}
	fm.setStatus(statusHelp)
}

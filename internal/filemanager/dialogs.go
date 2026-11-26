package filemanager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/goBaDig/internal/chrome"
	"github.com/wilsonpilon/goBaDig/internal/options"
)

// projectExtensions maps each known project dialect to the source
// file extension internal/dialect dispatches on, for the optional
// starter file a new project can be created with.
var projectExtensions = map[string]string{
	options.CategoryMSX:        ".dmx",
	options.CategoryZX81:       ".dzx",
	options.CategoryZXSpectrum: ".dzs",
}

var projectDialectLabels = []struct {
	id    string
	label string
}{
	{options.CategoryMSX, "MSX"},
	{options.CategoryZX81, "ZX81"},
	{options.CategoryZXSpectrum, "ZX Spectrum"},
}

func (fm *FileManager) newDialogWindow(form *tview.Form, icon, title string, rows int) *tview.Grid {
	colors := fm.colors
	form.SetBorder(true)
	form.SetBorderColor(colors.DialogBorder)
	chrome.WindowChrome(form.Box, icon, title, colors.DialogText)
	form.SetBackgroundColor(colors.DialogBackground)
	form.SetFieldBackgroundColor(tcell.ColorWhite)
	form.SetFieldTextColor(tcell.ColorBlack)
	form.SetLabelColor(colors.DialogText)
	form.SetButtonBackgroundColor(colors.DialogBorder)
	form.SetButtonTextColor(tcell.ColorBlack)
	form.SetButtonsAlign(tview.AlignRight)
	// SetBackgroundColor on Grid must stay on its own statement - see
	// the IMPORTANT comment on internal/tui's showThemePicker: chained
	// onto NewGrid()/AddItem(), it silently downgrades the chain's
	// result to a plain *tview.Box, so nothing ever draws the grid's
	// children.
	window := tview.NewGrid().SetRows(0, rows, 0).SetColumns(0, 60, 0).
		AddItem(form, 1, 1, 1, 1, 0, 0, true)
	// Fixed blue backdrop around the centered form, matching the
	// margin color internal/tui's own showThemePicker/
	// showOptionsPanel dialogs use (colors.desktopBg there) - that
	// margin is the TUI's desktop color, which is itself fixed, not
	// theme-driven (see paletteFor in tui.go).
	window.SetBackgroundColor(tcell.ColorBlue)
	return window
}

func (fm *FileManager) openDialog(name string, primitive tview.Primitive, focus tview.Primitive) {
	fm.dialogOpen = true
	fm.AddPage(name, primitive, true, true)
	fm.app.SetFocus(focus)
}

func (fm *FileManager) closeDialog(name string) {
	fm.RemovePage(name)
	fm.dialogOpen = false
	fm.app.SetFocus(fm.current().table)
}

// confirm shows a Yes/No modal and calls onYes only if the user
// confirms - used for the two actions that change or remove existing
// files (move, delete).
func (fm *FileManager) confirm(text string, onYes func()) {
	colors := fm.colors
	modal := tview.NewModal().SetText(text).AddButtons([]string{"Sim", "Nao"})
	modal.SetBackgroundColor(colors.DialogBackground)
	modal.SetTextColor(colors.DialogText)
	modal.SetButtonBackgroundColor(colors.DialogBorder)
	modal.SetButtonTextColor(tcell.ColorBlack)
	chrome.AttachShadow(modal.Box)
	modal.SetDoneFunc(func(_ int, label string) {
		fm.closeDialog("confirm")
		if label == "Sim" {
			onYes()
		}
	})
	fm.openDialog("confirm", modal, modal)
}

func (fm *FileManager) promptMkdir() {
	form := tview.NewForm()
	form.AddInputField("Nome do diretorio", "", 40, nil, nil)
	form.AddButton("Criar", func() {
		name := form.GetFormItemByLabel("Nome do diretorio").(*tview.InputField).GetText()
		fm.closeDialog("mkdir")
		fm.mkdir(name)
	})
	form.AddButton("Cancelar", func() { fm.closeDialog("mkdir") })
	form.SetCancelFunc(func() { fm.closeDialog("mkdir") })
	window := fm.newDialogWindow(form, chrome.IconFolder, "Novo Diretorio", 7)
	fm.openDialog("mkdir", window, form)
}

// promptNewProject implements Ctrl+N: name + parent folder (defaults
// to the active panel's current directory) + dialect. Per-project
// settings beyond name/path/dialect are deliberately not part of this
// dialog (confirmed with the user).
//
// "Pasta" is the PARENT location, not the project's own folder:
// createProject below always creates a new subfolder named after the
// project inside it, so a project's files stay organized in their own
// place instead of being dumped straight into whatever directory the
// user happened to be browsing (requested by the user: "crie uma
// pasta dentro do diretório corrente para organizar os arquivos").
//
// The starter file is always created (no longer an optional
// checkbox - the file manager is meant to be the project's full
// control center, so a project without anything to immediately edit
// defeats the point) as "main<ext>" - a fixed, predictable entry
// point name, not "<project name><ext>", so every project's starter
// file is found the same way regardless of what the project itself
// is called. It's opened in the embedded WibEdit editor right away,
// completing the create -> edit -> save -> compile -> execute loop
// entirely from the file manager.
func (fm *FileManager) promptNewProject() {
	labels := make([]string, len(projectDialectLabels))
	for i, d := range projectDialectLabels {
		labels[i] = d.label
	}
	selectedDialect := 0

	form := tview.NewForm()
	form.AddInputField("Nome", "", 40, nil, nil)
	form.AddInputField("Pasta", fm.current().dir, 40, nil, nil)
	form.AddDropDown("Dialeto", labels, selectedDialect, func(_ string, index int) {
		selectedDialect = index
	})
	form.AddButton("Criar", func() {
		name := form.GetFormItemByLabel("Nome").(*tview.InputField).GetText()
		parent := form.GetFormItemByLabel("Pasta").(*tview.InputField).GetText()
		dialectID := projectDialectLabels[selectedDialect].id
		fm.closeDialog("new-project")
		fm.createProject(name, parent, dialectID)
	})
	form.AddButton("Cancelar", func() { fm.closeDialog("new-project") })
	form.SetCancelFunc(func() { fm.closeDialog("new-project") })
	window := fm.newDialogWindow(form, chrome.IconFile, "Novo Projeto", 9)
	fm.openDialog("new-project", window, form)
}

// createProject creates a new subfolder named name inside parent -
// not parent itself - and makes THAT the project's own directory.
func (fm *FileManager) createProject(name, parent, dialectID string) {
	if name == "" {
		fm.setStatus(" o projeto precisa de um nome")
		return
	}
	path := filepath.Join(parent, name)
	err := fm.store.CreateProject(name, path, dialectID)
	// Reload the active panel either way, even on error below: store.
	// CreateProject's own os.MkdirAll runs BEFORE its database insert,
	// so a failure at the database step (most commonly a duplicate
	// project name - "name" is that table's primary key, and it
	// persists across restarts in the user's real settings.db, so
	// reusing a name from a past session fails here too) still leaves
	// a real, newly-created directory on disk that this panel hadn't
	// shown yet. Previously this returned before ever reloading,
	// leaving that directory invisible until the whole program was
	// restarted (reported bug: "às vezes dá erro e não mostra os
	// arquivos... mas entrando e saindo do programa o diretório está
	// lá").
	_ = fm.current().load(parent)
	if err != nil {
		fm.setStatus(fmt.Sprintf(" erro ao criar projeto: %v", err))
		return
	}
	starter := filepath.Join(path, "main"+projectExtensions[dialectID])
	if _, err := os.Stat(starter); os.IsNotExist(err) {
		_ = os.WriteFile(starter, []byte{}, 0o644)
	}
	_ = fm.current().load(path)
	fm.setStatus(fmt.Sprintf(" projeto %q criado em %s", name, path))
	fm.onOpenInEditor(starter)
}

// promptNewFile implements F3: a blank file in the active panel's
// current directory, named by the user (so the extension picks the
// dialect - .dmx/.dzx/.dzs for a source file the file manager can
// later compile/execute, or anything else for a plain text file).
// Opened in the embedded editor immediately, same reasoning as
// promptNewProject's starter file.
func (fm *FileManager) promptNewFile() {
	form := tview.NewForm()
	form.AddInputField("Nome do arquivo", "", 40, nil, nil)
	form.AddButton("Criar", func() {
		name := form.GetFormItemByLabel("Nome do arquivo").(*tview.InputField).GetText()
		fm.closeDialog("new-file")
		fm.createFile(name)
	})
	form.AddButton("Cancelar", func() { fm.closeDialog("new-file") })
	form.SetCancelFunc(func() { fm.closeDialog("new-file") })
	window := fm.newDialogWindow(form, chrome.IconSource, "Novo Arquivo", 7)
	fm.openDialog("new-file", window, form)
}

func (fm *FileManager) createFile(name string) {
	if name == "" {
		fm.setStatus(" o arquivo precisa de um nome")
		return
	}
	path := filepath.Join(fm.current().dir, name)
	if _, err := os.Stat(path); err == nil {
		fm.setStatus(fmt.Sprintf(" %q ja existe", name))
		return
	}
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		fm.setStatus(fmt.Sprintf(" erro ao criar arquivo: %v", err))
		return
	}
	_ = fm.current().reload()
	fm.setStatus(fmt.Sprintf(" arquivo %q criado", name))
	fm.onOpenInEditor(path)
}

// promptProjectList implements Ctrl+P: pick a saved project and jump
// the active panel straight to its folder.
func (fm *FileManager) promptProjectList() {
	projects, err := fm.store.ListProjects()
	if err != nil {
		fm.setStatus(fmt.Sprintf(" erro ao listar projetos: %v", err))
		return
	}
	if len(projects) == 0 {
		fm.setStatus(" nenhum projeto salvo ainda (Ctrl+N pra criar um)")
		return
	}

	colors := fm.colors
	list := tview.NewList().ShowSecondaryText(true)
	list.SetBorder(true)
	list.SetBorderColor(colors.DialogBorder)
	chrome.WindowChrome(list.Box, chrome.IconFolder, "Projetos", colors.DialogText)
	list.SetBackgroundColor(colors.DialogBackground)
	list.SetMainTextColor(colors.DialogText)
	list.SetSecondaryTextColor(colors.DialogBorder)
	list.SetSelectedBackgroundColor(colors.SelectionBg)
	list.SetSelectedTextColor(colors.SelectionFg)
	for _, project := range projects {
		p := project
		// tview.Escape: p.Name/p.Path are user-controlled (typed into
		// the New Project dialog) - see entryLabel in panel.go for
		// why this isn't optional.
		list.AddItem(tview.Escape(p.Name), fmt.Sprintf("%s - %s", p.Dialect, tview.Escape(p.Path)), 0, func() {
			fm.closeDialog("project-list")
			if err := fm.current().load(p.Path); err != nil {
				fm.setStatus(fmt.Sprintf(" erro: %v", err))
				return
			}
			fm.setStatus(fmt.Sprintf(" projeto %q aberto", p.Name))
		})
	}
	list.AddItem("Cancelar", "", 0, func() { fm.closeDialog("project-list") })
	rows := min(len(projects)+6, 20)
	window := tview.NewGrid().SetRows(0, rows, 0).SetColumns(0, 60, 0).
		AddItem(list, 1, 1, 1, 1, 0, 0, true)
	window.SetBackgroundColor(tcell.ColorBlue)
	fm.openDialog("project-list", window, list)
}

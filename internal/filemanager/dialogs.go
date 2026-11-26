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

// Fixed color scheme for this component's own dialogs - the file
// manager doesn't integrate with the TUI's active theme, same
// precedent as the embedded WibEdit editor (its own chrome is also
// independent of internal/theme).
const (
	dialogFg     = tcell.ColorWhite
	dialogBg     = tcell.ColorNavy
	dialogBorder = tcell.ColorSilver
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

func newDialogWindow(form *tview.Form, icon, title string, rows int) *tview.Grid {
	form.SetBorder(true)
	form.SetBorderColor(dialogBorder)
	chrome.WindowChrome(form.Box, icon, title, dialogFg)
	form.SetBackgroundColor(dialogBg)
	form.SetFieldBackgroundColor(tcell.ColorWhite)
	form.SetFieldTextColor(tcell.ColorBlack)
	form.SetLabelColor(dialogFg)
	form.SetButtonBackgroundColor(dialogBorder)
	form.SetButtonTextColor(tcell.ColorBlack)
	form.SetButtonsAlign(tview.AlignRight)
	// SetBackgroundColor on Grid must stay on its own statement - see
	// the IMPORTANT comment on internal/tui's showThemePicker: chained
	// onto NewGrid()/AddItem(), it silently downgrades the chain's
	// result to a plain *tview.Box, so nothing ever draws the grid's
	// children.
	window := tview.NewGrid().SetRows(0, rows, 0).SetColumns(0, 60, 0).
		AddItem(form, 1, 1, 1, 1, 0, 0, true)
	window.SetBackgroundColor(tcell.ColorBlack)
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
	modal := tview.NewModal().SetText(text).AddButtons([]string{"Sim", "Nao"})
	modal.SetBackgroundColor(dialogBg)
	modal.SetTextColor(dialogFg)
	modal.SetButtonBackgroundColor(dialogBorder)
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
	window := newDialogWindow(form, chrome.IconFolder, "Novo Diretorio", 7)
	fm.openDialog("mkdir", window, form)
}

// promptNewProject implements Ctrl+N: name + folder (defaulting to
// the active panel's current directory, created if it doesn't exist)
// + dialect, with an optional starter source file - exactly the
// scope confirmed with the user (per-project settings beyond
// name/path/dialect are deliberately not part of this dialog).
func (fm *FileManager) promptNewProject() {
	labels := make([]string, len(projectDialectLabels))
	for i, d := range projectDialectLabels {
		labels[i] = d.label
	}
	selectedDialect := 0
	createStarter := true

	form := tview.NewForm()
	form.AddInputField("Nome", "", 40, nil, nil)
	form.AddInputField("Pasta", fm.current().dir, 40, nil, nil)
	form.AddDropDown("Dialeto", labels, selectedDialect, func(_ string, index int) {
		selectedDialect = index
	})
	form.AddCheckbox("Criar arquivo inicial", createStarter, func(checked bool) {
		createStarter = checked
	})
	form.AddButton("Criar", func() {
		name := form.GetFormItemByLabel("Nome").(*tview.InputField).GetText()
		path := form.GetFormItemByLabel("Pasta").(*tview.InputField).GetText()
		dialectID := projectDialectLabels[selectedDialect].id
		fm.closeDialog("new-project")
		fm.createProject(name, path, dialectID, createStarter)
	})
	form.AddButton("Cancelar", func() { fm.closeDialog("new-project") })
	form.SetCancelFunc(func() { fm.closeDialog("new-project") })
	window := newDialogWindow(form, chrome.IconFile, "Novo Projeto", 11)
	fm.openDialog("new-project", window, form)
}

func (fm *FileManager) createProject(name, path, dialectID string, createStarter bool) {
	if name == "" {
		fm.setStatus(" o projeto precisa de um nome")
		return
	}
	if err := fm.store.CreateProject(name, path, dialectID); err != nil {
		fm.setStatus(fmt.Sprintf(" erro ao criar projeto: %v", err))
		return
	}
	if createStarter {
		ext := projectExtensions[dialectID]
		starter := filepath.Join(path, name+ext)
		if _, err := os.Stat(starter); os.IsNotExist(err) {
			_ = os.WriteFile(starter, []byte{}, 0o644)
		}
	}
	_ = fm.current().load(path)
	fm.setStatus(fmt.Sprintf(" projeto %q criado em %s", name, path))
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

	list := tview.NewList().ShowSecondaryText(true)
	list.SetBorder(true)
	list.SetBorderColor(dialogBorder)
	chrome.WindowChrome(list.Box, chrome.IconFolder, "Projetos", dialogFg)
	list.SetBackgroundColor(dialogBg)
	list.SetMainTextColor(dialogFg)
	list.SetSecondaryTextColor(tcell.ColorSilver)
	list.SetSelectedBackgroundColor(dialogBorder)
	list.SetSelectedTextColor(tcell.ColorBlack)
	for _, project := range projects {
		p := project
		list.AddItem(p.Name, fmt.Sprintf("%s - %s", p.Dialect, p.Path), 0, func() {
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
	window.SetBackgroundColor(tcell.ColorBlack)
	fm.openDialog("project-list", window, list)
}

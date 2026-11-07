// Package gui implements the graphical interface, built with Fyne.
package gui

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/goBaDig/internal/about"
	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/theme"
	"github.com/wilsonpilon/goBaDig/internal/version"
)

const defaultGUIThemeID = "vscode-dark"

// Run opens the GUI and blocks until the window is closed.
func Run() (runErr error) {
	store, err := config.OpenDefault()
	if err != nil {
		return err
	}
	defer func() {
		if err := store.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("close settings database: %w", err))
		}
	}()
	if err := store.InstallThemes(theme.Catalog()); err != nil {
		return err
	}
	activeTheme, err := store.EnsureActiveTheme(theme.ModeGUI, defaultGUIThemeID)
	if err != nil {
		return err
	}
	fyneTheme, err := newPaletteTheme(activeTheme)
	if err != nil {
		return err
	}

	a := app.New()
	a.Settings().SetTheme(fyneTheme)
	w := a.NewWindow("goBaDig v" + version.Version)
	activeThemeLabel := widget.NewLabel("Tema ativo: " + activeTheme.Name)

	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Exit", func() { a.Quit() }),
	)
	setupMenu := fyne.NewMenu("Setup",
		fyne.NewMenuItem("Install", func() {
			showThemeInstaller(a, w, store, activeThemeLabel)
		}),
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			dialog.ShowInformation("About goBaDig", about.Short(), w)
		}),
	)
	w.SetMainMenu(fyne.NewMainMenu(fileMenu, setupMenu, helpMenu))

	w.SetContent(container.NewVBox(
		widget.NewLabel("goBaDig"),
		activeThemeLabel,
		widget.NewLabel("Use Setup > Install para escolher outro tema."),
	))
	w.Resize(fyne.NewSize(480, 240))
	w.ShowAndRun()
	return nil
}

func showThemeInstaller(a fyne.App, w fyne.Window, store *config.Store, activeThemeLabel *widget.Label) {
	themes, err := store.ListThemes(theme.ModeGUI)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	activeTheme, err := store.ActiveTheme(theme.ModeGUI)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	labels := make([]string, len(themes))
	labelToTheme := make(map[string]theme.Theme, len(themes))
	selected := ""
	for i, item := range themes {
		labels[i] = item.Name
		labelToTheme[item.Name] = item
		if item.ID == activeTheme.ID {
			selected = item.Name
		}
	}
	description := widget.NewLabel(activeTheme.Description)
	selector := widget.NewRadioGroup(labels, func(selected string) {
		if item, ok := labelToTheme[selected]; ok {
			description.SetText(item.Description)
		}
	})
	selector.SetSelected(selected)
	content := container.NewVBox(
		widget.NewLabel("Escolha o tema que sera usado na interface grafica:"),
		selector,
		description,
	)
	confirm := dialog.NewCustomConfirm("Install GUI Theme", "Install", "Cancel", content, func(ok bool) {
		if !ok {
			return
		}
		item, ok := labelToTheme[selector.Selected]
		if !ok {
			dialog.ShowError(fmt.Errorf("selecione um tema da lista"), w)
			return
		}
		fyneTheme, err := newPaletteTheme(item)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if err := store.SetActiveTheme(theme.ModeGUI, item.ID); err != nil {
			dialog.ShowError(err, w)
			return
		}
		a.Settings().SetTheme(fyneTheme)
		activeThemeLabel.SetText("Tema ativo: " + item.Name)
		dialog.ShowInformation("Tema instalado", fmt.Sprintf("Tema ativo: %s", item.Name), w)
	}, w)
	confirm.Resize(fyne.NewSize(440, 360))
	confirm.Show()
}

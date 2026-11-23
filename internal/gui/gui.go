// Package gui implements the graphical interface, built with Fyne.
package gui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/goBaDig/internal/about"
	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/options"
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
		fyne.NewMenuItem("Basic Dignified", showOptionsPanel(w, store, options.CategoryCore, "Basic Dignified")),
		fyne.NewMenuItem("MSX", showOptionsPanel(w, store, options.CategoryMSX, "MSX")),
		fyne.NewMenuItem("ZX81", showOptionsPanel(w, store, options.CategoryZX81, "ZX81")),
		fyne.NewMenuItem("ZX Spectrum", showOptionsPanel(w, store, options.CategoryZXSpectrum, "ZX Spectrum")),
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

// showOptionsPanel builds one category's configuration panel (Basic
// Dignified / MSX / ZX81 / ZX Spectrum) - same catalog
// (internal/options) and the same four categories as the TUI's own
// showOptionsPanel (internal/tui/tui.go), using Fyne's dialog.ShowForm
// (a form dialog with its own Apply/Cancel buttons and validation
// wiring) instead of the TUI's hand-built tview.Form - a closer match
// to that widget than reusing dialog.NewCustomConfirm (which already
// supplies its own Confirm/Dismiss, same as a widget.Form would -
// using both together would double up on buttons).
//
// One options.Field becomes one *widget.FormItem: TypeBool ->
// widget.Check, TypeEnum -> widget.Select (its own options list is
// already a closed set, nothing to validate), TypeInt/TypeString ->
// widget.Entry. Unlike the TUI (which only validates at Apply time,
// since tview has no live-validation wiring this round would need to
// build itself), Entry.Validator here hooks options.Validate directly
// - Fyne's dialog.ShowForm already disables its own Apply button
// while any field reports invalid, giving int fields real-time Min/
// Max feedback for free.
func showOptionsPanel(w fyne.Window, store *config.Store, category, title string) func() {
	return func() {
		if err := store.EnsureOptionDefaults(category, options.Catalog()); err != nil {
			dialog.ShowError(err, w)
			return
		}
		values, err := store.ListOptions(category)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		fields := options.ForCategory(category)
		items, read := buildOptionsFormItems(category, values)

		dialog.ShowForm(title, "Apply", "Cancel", items, func(ok bool) {
			if !ok {
				return
			}
			for i, field := range fields {
				if err := store.SetOption(category, field.Key, read[i]()); err != nil {
					dialog.ShowError(err, w)
					return
				}
			}
		}, w)
	}
}

// buildOptionsFormItems turns one category's catalog fields into Fyne
// form items plus one getter per item that reads its widget's current
// value back as a string, ready for Store.SetOption - split out from
// showOptionsPanel so this wiring (which options.Type maps to which
// widget type, how its initial value gets set from values/Default) is
// unit-testable without a running window or dialog.
func buildOptionsFormItems(category string, values map[string]string) ([]*widget.FormItem, []func() string) {
	fields := options.ForCategory(category)
	items := make([]*widget.FormItem, 0, len(fields))
	read := make([]func() string, 0, len(fields))

	for _, field := range fields {
		current, ok := values[field.Key]
		if !ok {
			current = field.Default
		}
		switch field.Type {
		case options.TypeBool:
			check := widget.NewCheck("", nil)
			check.SetChecked(current == "true")
			items = append(items, widget.NewFormItem(field.Key, check))
			read = append(read, func() string { return strconv.FormatBool(check.Checked) })
		case options.TypeEnum:
			selected := current
			for _, choice := range field.Enum {
				if strings.EqualFold(choice, current) {
					selected = choice
					break
				}
			}
			sel := widget.NewSelect(field.Enum, nil)
			sel.SetSelected(selected)
			items = append(items, widget.NewFormItem(field.Key, sel))
			read = append(read, func() string { return sel.Selected })
		default: // options.TypeInt, options.TypeString
			entry := widget.NewEntry()
			entry.SetText(current)
			entry.Validator = func(raw string) error {
				return options.Validate(category, field.Key, raw)
			}
			items = append(items, widget.NewFormItem(field.Key, entry))
			read = append(read, func() string { return entry.Text })
		}
	}
	return items, read
}

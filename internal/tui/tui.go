// Package tui implements the terminal interface with a classic suspended
// drop-down menu bar in the visual tradition of Turbo Pascal and Norton.
package tui

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/goBaDig/internal/about"
	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/theme"
	"github.com/wilsonpilon/goBaDig/internal/version"
)

type menuItem struct {
	label  string
	action func()
}

type menuDef struct {
	label string
	items []menuItem
}

type palette struct {
	menuBg, menuFg, desktopBg, desktopFg, dialogBg, dialogFg, highlight, border tcell.Color
}

func paletteFor(item theme.Theme) palette {
	return palette{
		menuBg:    tcell.ColorSilver,
		menuFg:    tcell.ColorBlack,
		desktopBg: tcell.ColorBlue,
		desktopFg: tcell.ColorWhite,
		dialogBg:  tcell.GetColor(item.Palette.Surface),
		dialogFg:  tcell.GetColor(item.Palette.Foreground),
		highlight: tcell.GetColor(item.Palette.Selection),
		border:    tcell.GetColor(item.Palette.Primary),
	}
}

// Run opens the TUI and blocks until the user exits (File > Exit).
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
	if err := store.InstallThemes(theme.CatalogFor(theme.ModeTUI)); err != nil {
		return err
	}
	activeTheme, err := store.EnsureActiveTheme(theme.ModeTUI, "tui-vscode-dark")
	if err != nil {
		return err
	}
	themes, err := store.ListThemes(theme.ModeTUI)
	if err != nil {
		return err
	}

	app := tview.NewApplication()
	pages := tview.NewPages()
	colors := paletteFor(activeTheme)
	menus := []menuDef{
		{label: "File", items: []menuItem{{"Exit", func() { app.Stop() }}}},
		{label: "Setup", items: []menuItem{{"Install", nil}}},
		{label: "Help", items: []menuItem{{"About", nil}}},
	}
	activeMenu := 0
	menuOpen := false
	dialogOpen := false
	var dialogFocus tview.Primitive
	offsets := make([]int, len(menus))
	col := 1
	for i, m := range menus {
		offsets[i] = col
		col += len(m.label) + 3
	}

	titleBar := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	bar := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	desktop := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	statusBar := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	applyPalette := func() {
		titleBar.SetBackgroundColor(tcell.ColorNavy)
		titleBar.SetTextColor(tcell.ColorWhite)
		bar.SetBackgroundColor(colors.menuBg)
		bar.SetTextColor(colors.menuFg)
		desktop.SetBackgroundColor(colors.desktopBg)
		desktop.SetTextColor(colors.desktopFg)
		statusBar.SetBackgroundColor(tcell.ColorSilver)
		statusBar.SetTextColor(tcell.ColorBlack)
	}
	renderBar := func() {
		var b strings.Builder
		for i, m := range menus {
			if i == activeMenu {
				fmt.Fprintf(&b, "[white:blue] %s [-:-:-] ", m.label)
			} else {
				fmt.Fprintf(&b, " %s  ", m.label)
			}
		}
		bar.SetText(b.String())
	}
	setDesktop := func() {
		desktop.SetText(fmt.Sprintf("\n\n\n\n\n[::b]goBaDig[::b]\n\nBasic Dignified Suite\n\nTema TUI: %s", activeTheme.Name))
	}
	titleBar.SetText(fmt.Sprintf(" goBaDig v%s  -  BASIC Development Environment", version.Version))
	statusBar.SetText(" F10 Menu     Arrows Move     Enter Select     Esc Close")
	applyPalette()
	renderBar()
	setDesktop()
	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(titleBar, 1, 0, false).
		AddItem(bar, 1, 0, false).
		AddItem(desktop, 0, 1, false).
		AddItem(statusBar, 1, 0, false)
	pages.AddPage("app", root, true, true)

	closeMenu := func() {
		pages.RemovePage("dropdown")
		menuOpen = false
		app.SetFocus(pages)
	}
	showAbout := func() {
		modal := tview.NewModal().SetText(about.Short()).AddButtons([]string{"OK"}).
			SetDoneFunc(func(int, string) {
				pages.RemovePage("about")
				dialogOpen = false
				dialogFocus = nil
				app.SetFocus(pages)
			})
		modal.SetBackgroundColor(colors.dialogBg)
		modal.SetTextColor(colors.dialogFg)
		modal.SetButtonBackgroundColor(colors.highlight)
		modal.SetButtonTextColor(colors.desktopFg)
		dialogOpen = true
		dialogFocus = modal
		pages.AddPage("about", modal, true, true)
		app.SetFocus(modal)
	}
	var showError func(error)
	showThemePicker := func() {
		options := make([]string, len(themes))
		selected := 0
		for i, item := range themes {
			options[i] = item.Name
			if item.ID == activeTheme.ID {
				selected = i
			}
		}
		form := tview.NewForm()
		form.SetBorder(true).SetTitle(" Install TUI Theme ").SetBorderColor(colors.border)
		form.SetBackgroundColor(colors.dialogBg)
		form.SetFieldBackgroundColor(tcell.ColorWhite)
		form.SetFieldTextColor(tcell.ColorBlack)
		form.SetLabelColor(colors.dialogFg)
		form.SetButtonBackgroundColor(colors.highlight)
		form.SetButtonTextColor(colors.desktopFg)
		var description *tview.TextView
		form.AddDropDown("Theme", options, selected, func(_ string, index int) {
			if index >= 0 && index < len(themes) {
				selected = index
				if description != nil {
					description.SetText(themes[index].Description)
				}
			}
		})
		form.AddTextView("Description", themes[selected].Description, 38, 3, true, false)
		description = form.GetFormItemByLabel("Description").(*tview.TextView)
		description.SetTextColor(colors.dialogFg)
		description.SetBackgroundColor(colors.dialogBg)
		closeDialog := func() {
			pages.RemovePage("theme-picker")
			dialogOpen = false
			dialogFocus = nil
			app.SetFocus(pages)
		}
		form.AddButton("Apply", func() {
			item := themes[selected]
			if err := store.SetActiveTheme(theme.ModeTUI, item.ID); err != nil {
				showError(err)
				return
			}
			activeTheme = item
			colors = paletteFor(item)
			applyPalette()
			renderBar()
			setDesktop()
			closeDialog()
		})
		form.AddButton("Cancel", closeDialog)
		form.SetCancelFunc(closeDialog)
		form.SetButtonsAlign(tview.AlignRight)
		window := tview.NewGrid().SetRows(0, 13, 0).SetColumns(0, 64, 0).
			AddItem(form, 1, 1, 1, 1, 0, 0, true).
			SetBackgroundColor(colors.desktopBg)
		dialogOpen = true
		dialogFocus = form
		pages.AddPage("theme-picker", window, true, true)
		app.SetFocus(form)
	}
	showError = func(err error) {
		parentFocus := dialogFocus
		modal := tview.NewModal().SetText(fmt.Sprintf("Nao foi possivel ativar o tema:\n%v", err)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(int, string) {
				pages.RemovePage("theme-error")
				dialogFocus = parentFocus
				app.SetFocus(parentFocus)
			})
		modal.SetBackgroundColor(colors.dialogBg)
		modal.SetTextColor(colors.dialogFg)
		modal.SetButtonBackgroundColor(colors.highlight)
		modal.SetButtonTextColor(colors.desktopFg)
		dialogFocus = modal
		pages.AddPage("theme-error", modal, true, true)
		app.SetFocus(modal)
	}
	menus[1].items[0].action = showThemePicker
	menus[2].items[0].action = showAbout

	var openMenu func(i int)
	openMenu = func(i int) {
		activeMenu = i
		renderBar()
		m := menus[i]
		width := 14
		for _, it := range m.items {
			if w := len(it.label) + 6; w > width {
				width = w
			}
		}
		list := tview.NewList().ShowSecondaryText(false)
		list.SetBorder(true).SetBorderColor(colors.menuFg)
		list.SetBackgroundColor(colors.menuBg)
		list.SetMainTextColor(colors.menuFg)
		list.SetSelectedBackgroundColor(colors.highlight)
		list.SetSelectedTextColor(colors.desktopFg)
		for _, it := range m.items {
			action := it.action
			list.AddItem(it.label, "", 0, func() {
				closeMenu()
				if action != nil {
					action()
				}
			})
		}
		list.SetDoneFunc(closeMenu)
		positioned := tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, offsets[i], 0, false).
			AddItem(list, width, 0, true).
			AddItem(nil, 0, 1, false)
		dropdown := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 1, 0, false).
			AddItem(positioned, len(m.items)+2, 0, true).
			AddItem(nil, 0, 1, false)
		pages.RemovePage("dropdown")
		pages.AddPage("dropdown", dropdown, true, true)
		app.SetFocus(list)
		menuOpen = true
	}
	moveActive := func(delta int) {
		activeMenu = (activeMenu + delta + len(menus)) % len(menus)
		renderBar()
	}
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if dialogOpen {
			return event
		}
		if event.Modifiers()&tcell.ModAlt != 0 {
			r := unicode.ToUpper(event.Rune())
			for i, m := range menus {
				if rune(m.label[0]) == r {
					openMenu(i)
					return nil
				}
			}
		}
		if !menuOpen {
			switch event.Key() {
			case tcell.KeyLeft:
				moveActive(-1)
				return nil
			case tcell.KeyRight:
				moveActive(1)
				return nil
			case tcell.KeyDown, tcell.KeyEnter, tcell.KeyF10:
				openMenu(activeMenu)
				return nil
			}
			return event
		}
		switch event.Key() {
		case tcell.KeyLeft:
			moveActive(-1)
			openMenu(activeMenu)
			return nil
		case tcell.KeyRight:
			moveActive(1)
			openMenu(activeMenu)
			return nil
		}
		return event
	})

	return app.SetRoot(pages, true).SetFocus(pages).Run()
}

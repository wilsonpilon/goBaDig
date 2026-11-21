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
	"github.com/wilsonpilon/goBaDig/internal/editor"
	"github.com/wilsonpilon/goBaDig/internal/theme"
	"github.com/wilsonpilon/goBaDig/internal/version"
)

type menuItem struct {
	icon   string
	label  string
	action func()
}

type menuDef struct {
	icon  string
	label string
	items []menuItem
}

// visibleLabel is how a menu bar entry or dropdown item renders: its icon
// followed by its text label.
func visibleLabel(icon, label string) string {
	return icon + " " + label
}

type palette struct {
	menuBg, menuFg                        tcell.Color
	desktopBg, desktopFg, desktopPattern  tcell.Color
	dialogBg, dialogFg, highlight, border tcell.Color
}

func paletteFor(item theme.Theme) palette {
	return palette{
		menuBg:         tcell.ColorSilver,
		menuFg:         tcell.ColorBlack,
		desktopBg:      tcell.ColorBlue,
		desktopFg:      tcell.ColorWhite,
		desktopPattern: tcell.ColorTeal,
		dialogBg:       tcell.GetColor(item.Palette.Surface),
		dialogFg:       tcell.GetColor(item.Palette.Foreground),
		highlight:      tcell.GetColor(item.Palette.Selection),
		border:         tcell.GetColor(item.Palette.Primary),
	}
}

// Run opens the TUI and blocks until the user exits (File > Exit).
func Run() (runErr error) {
	return run(nil, nil)
}

// run is the real implementation behind Run(). It accepts an optional
// tcell.Screen (used by tests to drive the UI with a simulation screen
// instead of a real terminal) and an optional callback invoked with the
// *tview.Application once it has been constructed, before Run() blocks.
func run(screen tcell.Screen, ready func(*tview.Application)) (runErr error) {
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
	if screen != nil {
		app.SetScreen(screen)
	}
	pages := tview.NewPages()
	colors := paletteFor(activeTheme)
	menus := []menuDef{
		{icon: iconFile, label: "File", items: []menuItem{{iconExit, "Exit", func() { app.Stop() }}}},
		{icon: iconEdit, label: "Editor", items: []menuItem{{iconEdit, "Open", nil}}},
		{icon: iconSetup, label: "Setup", items: []menuItem{{iconInstall, "Install", nil}}},
		{icon: iconHelp, label: "Help", items: []menuItem{{iconAbout, "About", nil}}},
	}
	activeMenu := 0
	menuOpen := false
	dialogOpen := false
	var dialogFocus tview.Primitive
	offsets := make([]int, len(menus))
	col := 1
	for i, m := range menus {
		offsets[i] = col
		col += len([]rune(visibleLabel(m.icon, m.label))) + 3
	}

	titleBar := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	bar := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	desk := newDesktop()
	statusBar := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	applyPalette := func() {
		titleBar.SetBackgroundColor(tcell.ColorNavy)
		titleBar.SetTextColor(tcell.ColorWhite)
		bar.SetBackgroundColor(colors.menuBg)
		bar.SetTextColor(colors.menuFg)
		desk.SetColors(colors.desktopBg, colors.desktopFg, colors.desktopPattern)
		statusBar.SetBackgroundColor(tcell.ColorSilver)
		statusBar.SetTextColor(tcell.ColorBlack)
	}
	renderBar := func() {
		var b strings.Builder
		for i, m := range menus {
			label := visibleLabel(m.icon, m.label)
			if i == activeMenu {
				fmt.Fprintf(&b, "[white:blue] %s [-:-:-] ", label)
			} else {
				fmt.Fprintf(&b, " %s  ", label)
			}
		}
		bar.SetText(b.String())
	}
	setDesktop := func() {
		themeLine := "Tema TUI: " + activeTheme.Name
		rule := strings.Repeat(string(bannerRule), len([]rune(themeLine)))
		desk.SetLines("goBaDig", rule, "Basic Dignified Suite", "", themeLine)
	}
	titleBar.SetText(fmt.Sprintf(" ▣ goBaDig v%s  -  BASIC Development Environment", version.Version))
	statusBar.SetText(" F10 ☰ Menu     ←↑↓→ Move     Enter Select     Esc Close")
	applyPalette()
	renderBar()
	setDesktop()
	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(titleBar, 1, 0, false).
		AddItem(bar, 1, 0, false).
		AddItem(desk, 0, 1, false).
		AddItem(statusBar, 1, 0, false)
	pages.AddPage("app", root, true, true)

	closeEditor := func() {
		pages.SwitchToPage("app")
		app.SetFocus(pages)
	}
	ed := edit.NewEditor(app, closeEditor)
	pages.AddPage("editor", ed, true, false)
	openEditor := func() {
		pages.SwitchToPage("editor")
		app.SetFocus(ed.Input())
	}

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
		attachShadow(modal.Box)
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
			options[i] = fmt.Sprintf("[%s]%s[-] %s", item.Palette.Primary, iconSwatch, item.Name)
			if item.ID == activeTheme.ID {
				selected = i
			}
		}
		form := tview.NewForm()
		form.SetBorder(true).SetBorderColor(colors.border)
		windowChrome(form.Box, iconInstall, "Install TUI Theme", colors.dialogFg)
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
		// IMPORTANT: SetBackgroundColor is inherited from *tview.Box (Grid
		// has no override), so it returns *tview.Box, not *tview.Grid. If it
		// were chained onto NewGrid()/AddItem() like the other setters, the
		// chain's result would silently become a plain *Box - identical
		// Primitive interface, but Box.Draw() short-circuits before Grid's
		// own layout/child-drawing logic ever runs. That is the bug that
		// made this dialog (and whatever it contained) invisible: nothing
		// ever positioned or drew `form`, which is what the TUI's
		// "congela" ("freezes") report on opening Setup > Install actually
		// was. Keep this call on its own statement, discarding its *Box
		// return value, so `window` below stays a *tview.Grid.
		window := tview.NewGrid().SetRows(0, 13, 0).SetColumns(0, 64, 0).
			AddItem(form, 1, 1, 1, 1, 0, 0, true)
		window.SetBackgroundColor(colors.desktopBg)
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
		attachShadow(modal.Box)
		dialogFocus = modal
		pages.AddPage("theme-error", modal, true, true)
		app.SetFocus(modal)
	}
	menus[1].items[0].action = openEditor
	menus[2].items[0].action = showThemePicker
	menus[3].items[0].action = showAbout

	var openMenu func(i int)
	openMenu = func(i int) {
		activeMenu = i
		renderBar()
		m := menus[i]
		width := 14
		for _, it := range m.items {
			if w := len([]rune(visibleLabel(it.icon, it.label))) + 6; w > width {
				width = w
			}
		}
		if w := len([]rune(visibleLabel(m.icon, m.label))) + 10; w > width {
			width = w
		}
		list := tview.NewList().ShowSecondaryText(false)
		list.SetBorder(true).SetBorderColor(colors.menuFg)
		list.SetBackgroundColor(colors.menuBg)
		list.SetMainTextColor(colors.menuFg)
		list.SetSelectedBackgroundColor(colors.highlight)
		list.SetSelectedTextColor(colors.desktopFg)
		windowChrome(list.Box, m.icon, m.label, colors.menuFg)
		for _, it := range m.items {
			action := it.action
			list.AddItem(visibleLabel(it.icon, it.label), "", 0, func() {
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
			AddItem(nil, 2, 0, false). // Skip the title bar and the menu bar itself.
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
		if name, _ := pages.GetFrontPage(); name == "editor" {
			return ed.Capture(event)
		}
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

	app.SetRoot(pages, true).SetFocus(pages)
	if ready != nil {
		ready(app)
	}
	return app.Run()
}

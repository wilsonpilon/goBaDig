// Package tui implements the terminal interface with a classic suspended
// drop-down menu bar in the visual tradition of Turbo Pascal and Norton.
package tui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/goBaDig/internal/about"
	"github.com/wilsonpilon/goBaDig/internal/chrome"
	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/editor"
	"github.com/wilsonpilon/goBaDig/internal/filemanager"
	"github.com/wilsonpilon/goBaDig/internal/options"
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

// filemanagerColorsFor maps the active theme onto internal/filemanager's
// own Colors type, so the file manager renders with the same palette
// as the rest of the TUI's dialogs (see paletteFor above) instead of
// a fixed scheme of its own. The status bar stays fixed Silver/Black,
// matching the TUI's own top menu bar/bottom status bar, which are
// themselves never theme-driven either (menuBg/menuFg above).
//
// SourceText/ClassicText/BinaryText repurpose Palette.Hover and
// Palette.Error - neither is used anywhere else in this TUI, and
// both are deliberately vivid/high-contrast in every curated theme
// (Hover is an interactive-highlight accent, Error is the
// attention-grabbing destructive color) - exactly the "stands out"
// quality the user asked for when a panel listing should make BADIG
// source, classic listings, and ready-to-run artifacts scannable at
// a glance. SourceText shares Primary with DirectoryText on purpose:
// both are "the important things", and the "[name]" bracket
// notation already disambiguates a directory from a file.
func filemanagerColorsFor(item theme.Theme) filemanager.Colors {
	return filemanager.Colors{
		PanelBackground:  tcell.GetColor(item.Palette.Surface),
		FileText:         tcell.GetColor(item.Palette.Foreground),
		DirectoryText:    tcell.GetColor(item.Palette.Primary),
		SourceText:       tcell.GetColor(item.Palette.Primary),
		ClassicText:      tcell.GetColor(item.Palette.Hover),
		BinaryText:       tcell.GetColor(item.Palette.Error),
		InactiveBorder:   tcell.GetColor(item.Palette.Foreground),
		ActiveBorder:     tcell.GetColor(item.Palette.Primary),
		SelectionBg:      tcell.GetColor(item.Palette.Selection),
		SelectionFg:      tcell.ColorWhite,
		DialogBackground: tcell.GetColor(item.Palette.Surface),
		DialogText:       tcell.GetColor(item.Palette.Foreground),
		DialogBorder:     tcell.GetColor(item.Palette.Primary),
		StatusBg:         tcell.ColorSilver,
		StatusFg:         tcell.ColorBlack,
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
		{icon: chrome.IconFile, label: "File", items: []menuItem{
			{chrome.IconFolder, "File Manager", nil},
			{chrome.IconExit, "Exit", func() { app.Stop() }},
		}},
		{icon: chrome.IconEdit, label: "Editor", items: []menuItem{{chrome.IconEdit, "Open", nil}}},
		{icon: chrome.IconSetup, label: "Setup", items: []menuItem{
			{chrome.IconInstall, "Install", nil},
			{chrome.IconSetup, "Basic Dignified", nil},
			{chrome.IconSetup, "MSX", nil},
			{chrome.IconSetup, "ZX81", nil},
			{chrome.IconSetup, "ZX Spectrum", nil},
		}},
		{icon: chrome.IconHelp, label: "Help", items: []menuItem{{chrome.IconAbout, "About", nil}}},
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
		rule := strings.Repeat(string(chrome.BannerRule), len([]rune(themeLine)))
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

	// editorReturnPage is which page closeEditor switches back to -
	// "app" (the desktop) when the editor was opened from the Editor
	// menu, "filemanager" when it was opened from there instead (F3/F4
	// - see the file manager wiring below). Previously closeEditor
	// always switched to "app" unconditionally, so closing a file
	// opened FROM the file manager dropped the user onto the desktop
	// instead of back onto the file manager - reported by the user as
	// the file manager's own status bar (its shortcut list) "getting
	// messed up" after a round trip through the editor: what they were
	// actually seeing was the DESKTOP's own, different status bar,
	// having been sent to the wrong page.
	editorReturnPage := "app"
	// fm is declared here (assigned below, after the file manager is
	// built) rather than with ":=" at its construction site, so
	// closeEditor above can reference it - a closure in the same
	// function scope can't forward-reference a ":="-declared name.
	var fm *filemanager.FileManager
	closeEditor := func() {
		pages.SwitchToPage(editorReturnPage)
		if editorReturnPage == "filemanager" {
			app.SetFocus(fm)
		} else {
			app.SetFocus(pages)
		}
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
		chrome.AttachShadow(modal.Box)
		dialogOpen = true
		dialogFocus = modal
		pages.AddPage("about", modal, true, true)
		app.SetFocus(modal)
	}
	var showError func(context string, err error)
	closeFileManager := func() {
		pages.SwitchToPage("app")
		app.SetFocus(pages)
	}
	fm = filemanager.NewFileManager(app, store, filemanagerColorsFor(activeTheme), func(path string) {
		if err := ed.OpenFile(path); err != nil {
			showError("Nao foi possivel abrir o arquivo no editor", err)
			return
		}
		editorReturnPage = "filemanager"
		openEditor()
	}, closeFileManager)
	pages.AddPage("filemanager", fm, true, false)
	openFileManager := func() {
		pages.SwitchToPage("filemanager")
		app.SetFocus(fm)
	}
	showThemePicker := func() {
		options := make([]string, len(themes))
		selected := 0
		for i, item := range themes {
			options[i] = fmt.Sprintf("[%s]%s[-] %s", item.Palette.Primary, chrome.IconSwatch, item.Name)
			if item.ID == activeTheme.ID {
				selected = i
			}
		}
		form := tview.NewForm()
		form.SetBorder(true).SetBorderColor(colors.border)
		chrome.WindowChrome(form.Box, chrome.IconInstall, "Install TUI Theme", colors.dialogFg)
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
				showError("Nao foi possivel ativar o tema", err)
				return
			}
			activeTheme = item
			colors = paletteFor(item)
			applyPalette()
			renderBar()
			setDesktop()
			fm.SetColors(filemanagerColorsFor(item))
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
	showError = func(context string, err error) {
		parentFocus := dialogFocus
		modal := tview.NewModal().SetText(fmt.Sprintf("%s:\n%v", context, err)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(int, string) {
				pages.RemovePage("dialog-error")
				dialogFocus = parentFocus
				app.SetFocus(parentFocus)
			})
		modal.SetBackgroundColor(colors.dialogBg)
		modal.SetTextColor(colors.dialogFg)
		modal.SetButtonBackgroundColor(colors.highlight)
		modal.SetButtonTextColor(colors.desktopFg)
		chrome.AttachShadow(modal.Box)
		dialogFocus = modal
		pages.AddPage("dialog-error", modal, true, true)
		app.SetFocus(modal)
	}
	// showOptionsPanel builds one category's configuration form (Basic
	// Dignified / MSX / ZX81 / ZX Spectrum) - same shape as
	// showThemePicker above (tview.Form in a centered tview.Grid,
	// Apply/Cancel), but one widget per options.Field instead of a
	// single theme dropdown: options.TypeBool -> Checkbox,
	// options.TypeEnum -> DropDown, options.TypeInt -> InputField
	// restricted to tview.InputFieldInteger (shape-only; range is
	// still enforced by options.Validate, inside Store.SetOption, at
	// Apply time), options.TypeString -> plain InputField.
	//
	// Deliberate simplification: Apply calls Store.SetOption once per
	// field and stops at the first rejected one, reporting which -
	// fields applied before that point keep their new value (not
	// rolled back). In practice this can only happen on an
	// options.TypeInt field outside its Min/Max (every other widget
	// only ever offers an already-legal shape/choice), and the error
	// names the field, so re-opening Apply after fixing it is enough -
	// not worth a transactional multi-field rollback for that.
	showOptionsPanel := func(category, title string) func() {
		return func() {
			if err := store.EnsureOptionDefaults(category, options.Catalog()); err != nil {
				showError(fmt.Sprintf("Nao foi possivel carregar %s", title), err)
				return
			}
			fields := options.ForCategory(category)
			values, err := store.ListOptions(category)
			if err != nil {
				showError(fmt.Sprintf("Nao foi possivel carregar %s", title), err)
				return
			}

			form := tview.NewForm()
			form.SetBorder(true).SetBorderColor(colors.border)
			chrome.WindowChrome(form.Box, chrome.IconSetup, title, colors.dialogFg)
			form.SetBackgroundColor(colors.dialogBg)
			form.SetFieldBackgroundColor(tcell.ColorWhite)
			form.SetFieldTextColor(tcell.ColorBlack)
			form.SetLabelColor(colors.dialogFg)
			form.SetButtonBackgroundColor(colors.highlight)
			form.SetButtonTextColor(colors.desktopFg)

			for _, field := range fields {
				current, ok := values[field.Key]
				if !ok {
					current = field.Default
				}
				switch field.Type {
				case options.TypeBool:
					form.AddCheckbox(field.Key, current == "true", nil)
				case options.TypeEnum:
					selected := 0
					for i, choice := range field.Enum {
						if strings.EqualFold(choice, current) {
							selected = i
						}
					}
					form.AddDropDown(field.Key, field.Enum, selected, nil)
				case options.TypeInt:
					form.AddInputField(field.Key, current, 20, tview.InputFieldInteger, nil)
				default: // options.TypeString
					form.AddInputField(field.Key, current, 30, nil, nil)
				}
			}

			closeDialog := func() {
				pages.RemovePage("options-panel")
				dialogOpen = false
				dialogFocus = nil
				app.SetFocus(pages)
			}
			form.AddButton("Apply", func() {
				for _, field := range fields {
					var raw string
					switch widget := form.GetFormItemByLabel(field.Key).(type) {
					case *tview.Checkbox:
						raw = strconv.FormatBool(widget.IsChecked())
					case *tview.DropDown:
						_, raw = widget.GetCurrentOption()
					case *tview.InputField:
						raw = widget.GetText()
					}
					if err := store.SetOption(category, field.Key, raw); err != nil {
						showError(fmt.Sprintf("Opcao invalida: %s", field.Key), err)
						return
					}
				}
				closeDialog()
			})
			form.AddButton("Cancel", closeDialog)
			form.SetCancelFunc(closeDialog)
			form.SetButtonsAlign(tview.AlignRight)

			// Tall enough to fit every field without scrolling for the
			// smaller panels (ZX81 has 4), capped for the larger ones
			// (MSX/Basic Dignified have 14) - tview.Form scrolls its own
			// focused item into view automatically past that cap (ver
			// form.go's Draw, "Determine vertical offset").
			rows := max(min(len(fields)+5, 22), 10)
			// See the IMPORTANT comment on showThemePicker's own
			// identical window construction above - SetBackgroundColor
			// must stay on its own statement here too.
			window := tview.NewGrid().SetRows(0, rows, 0).SetColumns(0, 64, 0).
				AddItem(form, 1, 1, 1, 1, 0, 0, true)
			window.SetBackgroundColor(colors.desktopBg)
			dialogOpen = true
			dialogFocus = form
			pages.AddPage("options-panel", window, true, true)
			app.SetFocus(form)
		}
	}
	menus[0].items[0].action = openFileManager
	menus[1].items[0].action = func() { editorReturnPage = "app"; openEditor() }
	menus[2].items[0].action = showThemePicker
	menus[2].items[1].action = showOptionsPanel(options.CategoryCore, "Basic Dignified")
	menus[2].items[2].action = showOptionsPanel(options.CategoryMSX, "MSX")
	menus[2].items[3].action = showOptionsPanel(options.CategoryZX81, "ZX81")
	menus[2].items[4].action = showOptionsPanel(options.CategoryZXSpectrum, "ZX Spectrum")
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
		// SetMainTextColor alone only ever touches the FOREGROUND half
		// of List's own mainTextStyle - tview.NewList() seeds that
		// style's background to tview's own global
		// Styles.PrimitiveBackgroundColor (black, unrelated to this
		// menu's own palette), and List.Draw paints each unselected
		// row's text with that exact style (foreground+background
		// together), overwriting the Box's own SetBackgroundColor fill
		// at those character cells. The net effect: unselected text
		// rendered black-on-black (menuFg happens to be black too),
		// with the UNtouched padding past the text's own width still
		// showing the Box's real menuBg - a visible seam. Fixing this
		// needs SetMainTextStyle, which sets both halves at once.
		list.SetMainTextStyle(tcell.StyleDefault.Foreground(colors.menuFg).Background(colors.menuBg))
		list.SetSelectedBackgroundColor(colors.highlight)
		list.SetSelectedTextColor(colors.desktopFg)
		// The selected row's own highlight otherwise only spans its
		// own text's width too (same underlying List.Draw behavior),
		// leaving a short highlighted patch instead of a full menu-bar
		// width band - SetHighlightFullLine(true) is tview's own
		// knob for exactly this ("Norton/PC Tools"-style solid bars).
		list.SetHighlightFullLine(true)
		chrome.WindowChrome(list.Box, m.icon, m.label, colors.menuFg)
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
		if name, _ := pages.GetFrontPage(); name == "filemanager" {
			return fm.Capture(event)
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

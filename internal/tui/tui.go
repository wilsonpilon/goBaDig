// Package tui implements the terminal interface: a classic, early-1990s
// MS-DOS style "suspended" drop-down menu bar (File, Help), in the
// visual/keyboard tradition of Turbo Pascal 7, Norton Commander and the
// Norton Editor — a fixed top bar, inverse highlight on the active
// entry, and a floating box with the submenu drawn over the body
// content (not replacing the screen). Built with tview/tcell.
//
// Real configuration/pre-processing screens arrive together with the
// core engine and the SQLite config layer (see OUTLINE.md, Bloco 6).
package tui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/goBaDig/internal/about"
)

type menuItem struct {
	label    string
	shortcut rune
	action   func()
}

type menuDef struct {
	label string
	items []menuItem
}

// Classic DOS-era palette: light bar, blue "desktop", blue inverse
// highlight for the active bar entry and the selected dropdown row —
// the look shared by Turbo Pascal 7, Norton Commander and Norton Editor.
var (
	colBarBg     = tcell.ColorWhite
	colBarFg     = tcell.ColorBlack
	colBodyBg    = tcell.ColorBlue
	colBodyFg    = tcell.ColorWhite
	colHighlight = tcell.ColorBlue
)

// Run opens the TUI and blocks until the user exits (File > Exit).
func Run() error {
	app := tview.NewApplication()
	pages := tview.NewPages()

	menus := []menuDef{
		{label: "File", items: []menuItem{
			{"Exit", 'x', func() { app.Stop() }},
		}},
		{label: "Help", items: []menuItem{
			{"About", 'a', nil}, // wired to showAbout below
		}},
	}

	activeMenu := 0
	menuOpen := false

	// Column where each top-level label starts, used to position its
	// dropdown directly beneath it, like a real suspended menu.
	offsets := make([]int, len(menus))
	col := 1
	for i, m := range menus {
		offsets[i] = col
		col += len(m.label) + 3 // " Label " + 1 column of gap
	}

	bar := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)
	bar.SetBackgroundColor(colBarBg)
	bar.SetTextColor(colBarFg)

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
	renderBar()

	body := tview.NewTextView().
		SetText("\n  goBaDig - stub\n\n" +
			"  Setas ou Alt+letra para navegar o menu, Enter abre,\n" +
			"  Esc fecha o menu suspenso.").
		SetTextColor(colBodyFg)
	body.SetBackgroundColor(colBodyBg)

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(bar, 1, 0, false).
		AddItem(body, 0, 1, false)
	pages.AddPage("app", root, true, true)

	showAbout := func() {
		modal := tview.NewModal().
			SetText(about.Short()).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(int, string) { pages.RemovePage("about") })
		modal.SetBackgroundColor(colBarBg)
		modal.SetTextColor(colBarFg)
		modal.SetButtonBackgroundColor(colHighlight)
		modal.SetButtonTextColor(tcell.ColorWhite)
		pages.AddPage("about", modal, true, true)
		app.SetFocus(modal)
	}
	menus[1].items[0].action = showAbout

	closeMenu := func() {
		pages.RemovePage("dropdown")
		menuOpen = false
		app.SetFocus(pages)
	}

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
		height := len(m.items) + 2

		list := tview.NewList().ShowSecondaryText(false)
		list.SetBorder(true).SetBorderColor(colBarFg)
		list.SetBackgroundColor(colBarBg)
		list.SetMainTextColor(colBarFg)
		list.SetSelectedBackgroundColor(colHighlight)
		list.SetSelectedTextColor(tcell.ColorWhite)
		for _, it := range m.items {
			action := it.action
			list.AddItem(it.label, "", it.shortcut, func() {
				closeMenu()
				if action != nil {
					action()
				}
			})
		}
		list.SetDoneFunc(closeMenu)

		// Suspended box: a transparent page (nil spacers draw nothing,
		// letting the bar/body from the "app" page show through) with
		// the dropdown positioned right under its bar entry.
		positioned := tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, offsets[i], 0, false).
			AddItem(list, width, 0, true).
			AddItem(nil, 0, 1, false)
		dropdown := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 1, 0, false).
			AddItem(positioned, height, 0, true).
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

		// A dropdown is open: Left/Right switches straight to the
		// neighbouring menu's dropdown, matching Turbo Pascal/Norton
		// Commander behaviour. Up/Down/Enter/Esc stay with the list.
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

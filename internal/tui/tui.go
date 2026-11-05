// Package tui implements the terminal interface stub: a two-level
// menu (File > Exit, Help > About), built with tview/tcell. Real
// configuration/pre-processing screens arrive together with the core
// engine and the SQLite config layer (see OUTLINE.md, Bloco 6).
package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/goBaDig/internal/about"
)

// Run opens the stub TUI and blocks until the user exits (File > Exit
// or Ctrl-C).
func Run() error {
	app := tview.NewApplication()
	pages := tview.NewPages()

	mainMenu := tview.NewList().
		AddItem("File", "", 'f', nil).
		AddItem("Help", "", 'h', nil)
	mainMenu.SetBorder(true).SetTitle(" goBaDig ")

	fileMenu := tview.NewList().
		AddItem("Exit", "", 'x', func() { app.Stop() })
	fileMenu.SetBorder(true).SetTitle(" File ")
	fileMenu.SetDoneFunc(func() { pages.SwitchToPage("main") })

	aboutModal := tview.NewModal().
		SetText(about.Short()).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(int, string) { pages.SwitchToPage("main") })

	helpMenu := tview.NewList().
		AddItem("About", "", 'a', func() { pages.SwitchToPage("about") })
	helpMenu.SetBorder(true).SetTitle(" Help ")
	helpMenu.SetDoneFunc(func() { pages.SwitchToPage("main") })

	mainMenu.SetSelectedFunc(func(_ int, mainText, _ string, _ rune) {
		switch mainText {
		case "File":
			pages.SwitchToPage("file")
		case "Help":
			pages.SwitchToPage("help")
		}
	})

	pages.AddPage("main", mainMenu, true, true)
	pages.AddPage("file", fileMenu, true, false)
	pages.AddPage("help", helpMenu, true, false)
	pages.AddPage("about", aboutModal, true, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if name, _ := pages.GetFrontPage(); name != "main" {
				pages.SwitchToPage("main")
				return nil
			}
		}
		return event
	})

	return app.SetRoot(pages, true).SetFocus(pages).Run()
}

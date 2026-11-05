// Package gui implements the graphical interface stub: a window with a
// File > Exit and Help > About menu, built with Fyne. Real
// configuration/pre-processing screens arrive together with the core
// engine and the SQLite config layer (see OUTLINE.md, Bloco 7).
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/goBaDig/internal/about"
	"github.com/wilsonpilon/goBaDig/internal/version"
)

// Run opens the stub GUI and blocks until the window is closed.
func Run() {
	a := app.New()
	w := a.NewWindow("goBaDig v" + version.Version)

	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Exit", func() { a.Quit() }),
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			dialog.ShowInformation("About goBaDig", about.Short(), w)
		}),
	)
	w.SetMainMenu(fyne.NewMainMenu(fileMenu, helpMenu))

	w.SetContent(widget.NewLabel(
		"goBaDig - stub GUI.\nUse o menu File > Exit ou Help > About.",
	))
	w.Resize(fyne.NewSize(480, 240))
	w.ShowAndRun()
}

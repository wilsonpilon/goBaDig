package edit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rivo/uniseg"
)

type Editor struct {
	*tview.Pages
	app              *tview.Application
	area             *documentView
	root             *tview.Flex
	windowLayout     *tview.Flex
	windows          []*editorWindow
	activeWindow     int
	dividerPosition  int
	lineStatus       *tview.TextView
	columnStatus     *tview.TextView
	nameStatus       *tview.TextView
	modeStatus       *tview.TextView
	wrapStatus       *tview.TextView
	autoStatus       *tview.TextView
	statusBar        *tview.Flex
	onClose          func()
	filename         string
	path             string
	diskText         string
	diskKnown        bool
	lastDir          string
	fileBase         int
	shellRunner      func() error
	printer          func([]byte) error
	backupCharacter  rune
	backupsEnabled   bool
	menuX            map[rune]int
	menuOrder        []rune
	currentMenu      rune
	dropdown         *editorDropdown
	menuOpen         bool
	confirmExit      bool
	changed          bool
	wrap             bool
	over             bool
	colorProfile     int
	displayScrollbar bool
	mouseButtonRight bool
	mouseScrollDelay time.Duration
	configPath       string
	configChanged    bool
	startMode        int
	literal          bool
	auto             bool
	notice           string
	dialogs          []*editorDialog
	outlineView      *editorOutlineView
	lastSearch       searchState
	replaceSession   *replaceSession
	virtual          *VirtualDocument
}

type editorWindow struct {
	area                     *documentView
	filename, path, diskText string
	lastDir                  string
	diskKnown, changed, wrap bool
	fileBase                 int
	over, literal, auto      bool
	startMode                int
}

func NewEditor(app *tview.Application, onClose func()) *Editor {
	if onClose == nil {
		onClose = func() {}
	}
	e := &Editor{
		Pages:            tview.NewPages(),
		app:              app,
		area:             newDocumentView(),
		onClose:          onClose,
		filename:         "Untitled",
		backupCharacter:  '{',
		backupsEnabled:   true,
		displayScrollbar: true,
		dividerPosition:  50,
		menuX:            make(map[rune]int),
		menuOrder:        []rune{'F', 'B', 'T', 'S', 'P', 'V', 'O', 'H'},
		printer:          sendToDefaultPrinter,
	}
	setDisplayPalette(0)
	e.shellRunner = e.runPowerShell
	e.build()
	return e
}

func Run() error {
	return runEditor("")
}

func RunFile(path string) error {
	if path == "" {
		return errors.New("a file name is required")
	}
	return runEditor(path)
}

func runEditor(path string) error {
	app := tview.NewApplication().EnableMouse(true)
	editor := NewEditor(app, app.Stop)
	if err := editor.loadStartupConfiguration(); err != nil {
		return err
	}
	app.SetInputCapture(editor.Capture)
	app.SetRoot(editor, true)
	if path != "" {
		if err := editor.openInitialFile(path); err != nil {
			return err
		}
	}
	return app.Run()
}

func (e *Editor) openInitialFile(name string) error {
	if e.virtual != nil {
		return errors.New("host files are unavailable in a FILE SPACE session")
	}
	path, text, err := readEditorFile(name)
	if errors.Is(err, os.ErrNotExist) {
		info, parentErr := os.Stat(filepath.Dir(path))
		if parentErr != nil {
			return parentErr
		}
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", filepath.Dir(path))
		}
		e.offerNewFile(path, func() { e.setFile(path, "", false) })
		return nil
	}
	if err != nil {
		return err
	}
	e.setFile(path, text, true)
	return nil
}

func (e *Editor) Input() tview.Primitive { return e.area }

func (e *Editor) build() {
	menuBar := &editorMenuBar{Box: tview.NewBox(), editor: e}
	e.windows = []*editorWindow{{area: e.area, filename: e.filename, startMode: e.startMode}}
	e.bindWindow(e.windows[0], 0)
	e.area.SetBackgroundColor(contentBackgroundColor)
	e.SetBackgroundColor(contentBackgroundColor)

	e.lineStatus = statusCell("")
	e.columnStatus = statusCell("")
	e.nameStatus = statusCell("")
	e.modeStatus = statusCell("")
	e.wrapStatus = statusCell("")
	e.autoStatus = statusCell("")
	statusBar := tview.NewFlex()
	e.statusBar = statusBar
	statusBar.SetBackgroundColor(statusBackgroundColor)
	statusBar.AddItem(&editorScrollArrow{Box: tview.NewBox(), editor: e, step: -1}, 3, 0, false)
	statusBar.AddItem(e.lineStatus, 10, 0, false)
	statusBar.AddItem(e.columnStatus, 10, 0, false)
	statusBar.AddItem(e.nameStatus, 0, 1, false)
	statusBar.AddItem(e.modeStatus, 8, 0, false)
	statusBar.AddItem(e.wrapStatus, 8, 0, false)
	statusBar.AddItem(e.autoStatus, 6, 0, false)
	statusBar.AddItem(&editorScrollArrow{Box: tview.NewBox(), editor: e, step: 1}, 1, 0, false)

	e.windowLayout = tview.NewFlex().SetDirection(tview.FlexRow)
	e.rebuildWindowLayout()
	e.windowLayout.SetBackgroundColor(contentBackgroundColor)
	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.SetBackgroundColor(contentBackgroundColor)
	root.AddItem(menuBar, 1, 0, false)
	root.AddItem(e.windowLayout, 0, 1, true)
	root.AddItem(statusBar, 1, 0, false)
	e.root = root
	e.AddPage("editor", root, true, true)
	e.updateStatus()
	e.app.SetFocus(e.area)
}

func (e *Editor) bindWindow(window *editorWindow, index int) {
	window.area.activated = func() { e.activateWindow(index, false) }
	window.area.changed = func() {
		window.changed = true
		if e.activeWindow == index {
			e.changed = true
			e.notice = ""
			e.updateStatus()
		}
	}
	window.area.moved = func() {
		if e.activeWindow == index {
			e.notice = ""
			e.updateStatus()
		}
	}
}

func (e *Editor) rebuildWindowLayout() {
	e.windowLayout.Clear()
	if len(e.windows) == 1 {
		e.windowLayout.AddItem(e.windows[0].area, 0, 1, true)
		return
	}
	e.windowLayout.AddItem(e.windows[0].area, 0, e.dividerPosition, false)
	e.windowLayout.AddItem(&editorWindowDivider{Box: tview.NewBox()}, 1, 0, false)
	e.windowLayout.AddItem(e.windows[1].area, 0, 100-e.dividerPosition, false)
}

func (e *Editor) syncActiveWindow() {
	if e.activeWindow < 0 || e.activeWindow >= len(e.windows) {
		return
	}
	window := e.windows[e.activeWindow]
	window.area = e.area
	window.filename, window.path, window.diskText = e.filename, e.path, e.diskText
	window.lastDir, window.diskKnown, window.changed = e.lastDir, e.diskKnown, e.changed
	window.fileBase = e.fileBase
	window.wrap, window.over, window.literal, window.auto = e.wrap, e.over, e.literal, e.auto
	window.startMode = e.startMode
}

func (e *Editor) activateWindow(index int, focus bool) {
	if index < 0 || index >= len(e.windows) || index == e.activeWindow {
		if focus && e.area != nil {
			e.app.SetFocus(e.area)
		}
		return
	}
	e.syncActiveWindow()
	e.activeWindow = index
	window := e.windows[index]
	e.area = window.area
	e.filename, e.path, e.diskText = window.filename, window.path, window.diskText
	e.lastDir, e.diskKnown, e.changed = window.lastDir, window.diskKnown, window.changed
	e.fileBase = window.fileBase
	e.wrap, e.over, e.literal, e.auto = window.wrap, window.over, window.literal, window.auto
	e.startMode = window.startMode
	e.area.document.wordWrap, e.area.document.autoIndent = e.wrap, e.auto
	e.area.SetWrap(e.wrap)
	e.updateStatus()
	if focus {
		e.app.SetFocus(e.area)
	}
}

func (e *Editor) openSecondWindow(path, text string, known bool) {
	e.syncActiveWindow()
	if len(e.windows) == 1 {
		window := &editorWindow{
			area:      newDocumentView(),
			filename:  "Untitled",
			lastDir:   e.lastDir,
			startMode: e.startMode,
			wrap:      e.wrap,
			auto:      e.auto,
		}
		window.area.scrollbar = e.displayScrollbar
		window.area.setMouseOptions(e.mouseButtonRight, e.mouseScrollDelay)
		window.area.document.tabWidth = e.area.document.tabWidth
		window.area.document.lineLength = e.area.document.lineLength
		window.area.document.firstMargin = e.area.document.firstMargin
		window.area.document.leftMargin = e.area.document.leftMargin
		window.area.document.pageLength = e.area.document.pageLength
		window.area.document.printTopMargin = e.area.document.printTopMargin
		window.area.document.printLeftMargin = e.area.document.printLeftMargin
		window.area.document.formfeedBetween = e.area.document.formfeedBetween
		window.area.document.formfeedAfter = e.area.document.formfeedAfter
		window.area.document.formatAtCursor = e.area.document.formatAtCursor
		window.area.document.wordWrap = e.wrap
		window.area.document.autoIndent = e.auto
		e.bindWindow(window, 1)
		e.windows = append(e.windows, window)
		e.rebuildWindowLayout()
	}
	e.activateWindow(1, true)
	e.setFile(path, text, known)
}

func (e *Editor) closeActiveWindowNow() {
	if len(e.windows) != 2 {
		return
	}
	e.syncActiveWindow()
	remaining := e.windows[1-e.activeWindow]
	e.windows = []*editorWindow{remaining}
	e.activeWindow = 0
	window := e.windows[0]
	e.bindWindow(window, 0)
	e.area = window.area
	e.filename, e.path, e.diskText = window.filename, window.path, window.diskText
	e.lastDir, e.diskKnown, e.changed = window.lastDir, window.diskKnown, window.changed
	e.fileBase = window.fileBase
	e.wrap, e.over, e.literal, e.auto = window.wrap, window.over, window.literal, window.auto
	e.startMode = window.startMode
	e.area.document.wordWrap, e.area.document.autoIndent = e.wrap, e.auto
	e.area.SetWrap(e.wrap)
	e.rebuildWindowLayout()
	e.app.SetFocus(e.area)
	e.updateStatus()
}

func (e *Editor) closeWindow() {
	if len(e.windows) != 2 {
		e.notice = "No other window is open [file.close-window]."
		e.updateStatus()
		return
	}
	e.confirmUnsaved("file.close-window", e.closeActiveWindowNow)
}

func (e *Editor) moveDivider(delta int) {
	if len(e.windows) != 2 {
		e.notice = "Open the second window before moving the divider."
		e.updateStatus()
		return
	}
	next := max(20, min(80, e.dividerPosition+delta))
	if next == e.dividerPosition {
		return
	}
	e.dividerPosition = next
	e.rebuildWindowLayout()
	e.app.SetFocus(e.area)
	e.notice = fmt.Sprintf("Window divider: %d%%", e.dividerPosition)
	e.updateStatus()
}

func (e *Editor) menu(key rune) (menuSpec, bool) {
	for _, menu := range neMenus {
		if menu.key == key {
			return menu, true
		}
	}
	if key == 'H' {
		return neHelpMenu, true
	}
	return menuSpec{}, false
}

func (e *Editor) available(command commandSpec) bool {
	if e.virtual != nil && !virtualCommandAllowed(command.id) {
		return false
	}
	switch command.id {
	case "file.window":
		return true
	case "file.close-window", "window.divider-up", "window.divider-down":
		return len(e.windows) == 2
	case "layout.paragraph", "layout.center", "layout.expand-tabs", "layout.put-tabs",
		"layout.insert-pages", "layout.remove-pages":
		return true
	case "layout.document", "layout.setup":
		return true
	case "options.display":
		return true
	case "options.mouse":
		return true
	case "options.save":
		return true
	}
	switch command.id {
	case "file.new", "file.open", "file.save", "file.save-as":
		return true
	case "file.merge", "file.load-more", "file.write-cursor", "file.top", "file.shell":
		return true
	}
	if strings.HasPrefix(command.id, "edit.") || command.id == "options.editor" {
		return true
	}
	if strings.HasPrefix(command.id, "cursor.") {
		return true
	}
	if strings.HasPrefix(command.id, "help.") {
		return true
	}
	switch command.id {
	case "file.exit", "layout.wrap", "layout.auto", "edit.insert", "menu.open":
		return true
	case "print.setup":
		return true
	case "print.document":
		return true
	case "print.block":
		return true
	case "print.eject":
		return true
	case "print.pdf":
		return true
	case "view.information":
		return true
	case "view.outline":
		return true
	case "view.about":
		return true
	case "block.copy", "block.move", "block.delete", "block.delete-direct",
		"block.set", "block.remove", "block.find", "block.indent", "block.unindent",
		"block.output", "block.mark-line", "block.mark-end", "search.marker",
		"search.find", "search.replace", "search.forward", "search.reverse",
		"search.line", "search.bracket", "search.differences", "search.block":
		return true
	default:
		return false
	}
}

func (e *Editor) openMenu(key rune) {
	e.literal = false
	menu, ok := e.menu(key)
	if !ok {
		return
	}
	if e.menuOpen {
		e.RemovePage("dropdown")
	}
	e.currentMenu = key
	e.menuOpen = true
	e.dropdown = &editorDropdown{Box: tview.NewBox(), editor: e, items: menu.items}
	e.AddPage("dropdown", e.dropdown, true, true)
	e.app.SetFocus(e.dropdown)
	e.menuNotice(menu.items[0])
}

func (e *Editor) menuNotice(command commandSpec) {
	e.notice = command.description
	if !e.available(command) {
		e.notice = command.label + " - planned " + command.phase + " [" + command.id + "]"
	}
	e.updateStatus()
}

func (e *Editor) closeMenu() {
	if !e.menuOpen {
		return
	}
	e.RemovePage("dropdown")
	e.menuOpen = false
	e.dropdown = nil
	e.notice = ""
	e.updateStatus()
	e.app.SetFocus(e.area)
}

func (e *Editor) execute(command commandSpec) {
	if e.virtual != nil && !virtualCommandAllowed(command.id) {
		e.notice = "Unavailable in a FILE SPACE session. Save and return to WIBBase to choose another file."
		e.updateStatus()
		return
	}
	if strings.HasPrefix(command.id, "edit.") {
		if handled, notice := e.area.edit(command.id); handled {
			e.notice = notice
			e.updateStatus()
			return
		}
	}
	if strings.HasPrefix(command.id, "cursor.") && e.area.navigate(command.id) {
		return
	}
	if strings.HasPrefix(command.id, "help.") {
		if command.id == "help.index" {
			e.openMenu('H')
		} else {
			e.showHelp(command.id)
		}
		return
	}
	switch command.id {
	case "file.new":
		e.confirmUnsaved(command.id, func() { e.setFile("", "", false) })
	case "file.open":
		e.fileDialog(command.id, nil)
	case "file.save":
		e.saveDocument(nil)
	case "file.save-as":
		e.fileDialog(command.id, nil)
	case "file.merge":
		e.fileDialog(command.id, nil)
	case "file.window":
		if len(e.windows) == 2 {
			e.activateWindow(1-e.activeWindow, true)
		} else {
			e.fileDialog(command.id, nil)
		}
	case "file.close-window":
		e.closeWindow()
	case "file.load-more":
		e.loadMore()
	case "file.write-cursor":
		e.writeToCursor()
	case "file.top":
		e.topOfFile()
	case "file.shell":
		if err := e.shellRunner(); err != nil {
			e.fileError(err, "file.shell")
		} else {
			e.notice = "Returned from PowerShell."
		}
	case "file.exit":
		e.requestClose()
	case "search.find":
		e.notice = "Find text in the current document."
		e.previewDialog(command)
	case "search.replace":
		e.notice = "Find and replace text in the current document."
		e.previewDialog(command)
	case "search.forward":
		e.continueSearch(false, command.id)
	case "search.reverse":
		e.continueSearch(true, command.id)
	case "search.line":
		e.previewDialog(command)
	case "search.bracket":
		e.matchingBracket()
	case "search.differences":
		e.windowDifferences()
	case "search.block":
		e.searchTextInBlock()
	case "print.setup":
		e.printSetup()
	case "print.document":
		e.printDocument()
	case "print.block":
		e.printBlock()
	case "print.eject":
		e.ejectPage()
	case "print.pdf":
		e.fileDialog(command.id, nil)
	case "view.information":
		e.informationDisplay()
	case "view.outline":
		e.outlineDisplay()
	case "view.about":
		e.aboutDisplay()
	case "options.display":
		e.displaySetup()
	case "options.mouse":
		e.mouseSetup()
	case "options.save":
		e.saveConfigurationDialog()
	case "window.divider-up":
		e.moveDivider(-5)
	case "window.divider-down":
		e.moveDivider(5)
	case "layout.wrap":
		e.wrap = !e.wrap
		e.area.document.wordWrap = e.wrap
		e.area.SetWrap(e.wrap)
		e.notice = fmt.Sprintf("Word wrap %s.", map[bool]string{true: "on", false: "off"}[e.wrap])
	case "layout.auto":
		e.auto = !e.auto
		e.area.document.autoIndent = e.auto
		e.notice = fmt.Sprintf("Automatic indentation %s.", map[bool]string{true: "on", false: "off"}[e.auto])
	case "layout.setup":
		e.layoutSetup()
	case "layout.paragraph":
		e.formatParagraph()
	case "layout.document":
		e.documentFormatDialog()
	case "layout.center":
		e.centerCurrentLine()
	case "layout.expand-tabs":
		e.layoutStartDialog(command, func(fromCursor bool) bool {
			return e.area.document.expandTabs(fromCursor)
		})
	case "layout.put-tabs":
		e.layoutStartDialog(command, func(fromCursor bool) bool {
			return e.area.document.putInTabs(fromCursor)
		})
	case "layout.insert-pages":
		e.layoutStartDialog(command, func(fromCursor bool) bool {
			return e.area.document.insertPageBreaks(fromCursor)
		})
	case "layout.remove-pages":
		e.layoutStartDialog(command, func(fromCursor bool) bool {
			return e.area.document.removePageBreaks(fromCursor)
		})
	case "edit.insert":
		e.over = !e.over && e.startMode != 2
		e.area.over = e.over
	case "edit.literal":
		e.literal = true
		e.notice = "Literal input: press the next character or control key."
	case "options.editor":
		e.editorSetup()
	case "menu.open":
		e.openMenu('F')
	case "block.set":
		marker, complete := e.area.document.setBlockMarkerAtCursor()
		if complete {
			e.notice = "Block selected between markers [block.set]."
		} else {
			e.notice = fmt.Sprintf("Block marker %d set; set another marker to select a block [block.set].", marker)
		}
	case "block.remove":
		hadMarkers := e.area.document.blockMarkerCount > 0
		e.area.document.clearBlockMarkers()
		if hadMarkers {
			e.notice = "Block markers removed [block.remove]."
		} else {
			e.notice = "No block markers were set [block.remove]."
		}
	case "block.find", "search.marker":
		if marker, _, ok := e.area.document.findBlockMarker(); ok {
			if command.id == "search.marker" {
				e.notice = fmt.Sprintf("Search moved to block marker %d [search.marker].", marker+1)
			} else {
				e.notice = fmt.Sprintf("Moved to block marker %d [block.find].", marker+1)
			}
		} else {
			if command.id == "search.marker" {
				e.notice = "No block markers are set [search.marker]."
			} else {
				e.notice = "No block markers are set [block.find]."
			}
		}
	case "block.mark-line":
		e.area.document.markCurrentLine()
		e.notice = "Current line marked [block.mark-line]."
	case "block.mark-end":
		e.area.document.markToEndOfLine()
		e.notice = "Text to end of line marked [block.mark-end]."
	case "block.copy":
		e.copyBlock()
	case "block.move":
		e.moveBlock()
	case "block.delete":
		e.deleteBlock(false)
	case "block.delete-direct":
		e.deleteBlock(true)
	case "block.indent", "block.unindent":
		e.blockIndent(command.id)
	case "block.output":
		e.outputBlock()
	default:
		e.dispatch(command)
		if command.dialog != "" {
			e.previewDialog(command)
		}
	}
	e.updateStatus()
}

func (e *Editor) Capture(event *tcell.EventKey) *tcell.EventKey {
	if e.outlineView != nil {
		e.closeOutlineView()
		return nil
	}
	if e.literal {
		e.literal = false
		if text, ok := literalText(event); ok {
			e.area.notify(e.area.document.insertText(text, false))
		} else {
			e.showMessage("Literal input", "That key has no literal character value. No text was changed.")
		}
		return nil
	}
	if len(e.dialogs) > 0 {
		e.dialogs[len(e.dialogs)-1].key(event)
		return nil
	}
	if event.Key() == tcell.KeyEscape {
		if e.menuOpen {
			e.closeMenu()
		} else {
			e.requestClose()
		}
		return nil
	}
	if event.Key() == tcell.KeyF1 {
		if e.menuOpen {
			e.showHelp(e.dropdown.items[e.dropdown.selected].id)
		} else {
			e.openMenu('H')
		}
		return nil
	}
	if event.Key() == tcell.KeyF10 {
		if e.menuOpen {
			e.closeMenu()
		} else {
			e.openMenu('F')
		}
		return nil
	}
	if event.Key() == tcell.KeyRune && event.Modifiers()&tcell.ModAlt != 0 {
		key := unicode.ToUpper(event.Rune())
		if _, ok := e.menu(key); ok && key != 'H' {
			e.openMenu(key)
			return nil
		}
	}
	if e.menuOpen {
		if event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyRight {
			step := 1
			if event.Key() == tcell.KeyLeft {
				step = len(e.menuOrder) - 1
			}
			for i, key := range e.menuOrder {
				if key == e.currentMenu {
					e.openMenu(e.menuOrder[(i+step)%len(e.menuOrder)])
					break
				}
			}
			return nil
		}
		return event
	}
	if command, ok := commandForKey(event); ok {
		e.execute(command)
		return nil
	}
	if event.Key() != tcell.KeyRune || event.Modifiers()&^tcell.ModShift != 0 {
		e.notice = "Key not available in " + compatibilityProfile + ": " + keyName(event)
		e.updateStatus()
		return nil
	}
	return event
}

func (e *Editor) requestClose() {
	if e.configChanged && e.virtual == nil {
		e.confirmConfigurationExit()
		return
	}
	e.confirmCloseDocuments(false)
}

func (e *Editor) confirmCloseDocuments(discardConfiguration bool) {
	e.confirmExit = true
	order := []int{e.activeWindow}
	if len(e.windows) == 2 {
		order = append(order, 1-e.activeWindow)
	}
	e.confirmCloseWindows(order, 0, discardConfiguration)
}

func (e *Editor) confirmCloseWindows(order []int, position int, discardConfiguration bool) {
	if position >= len(order) {
		e.confirmExit = false
		if discardConfiguration {
			e.configChanged = false
		}
		e.setFile("", "", false)
		e.onClose()
		return
	}
	index := order[position]
	e.activateWindow(index, true)
	if !e.changed {
		e.confirmCloseWindows(order, position+1, discardConfiguration)
		return
	}
	e.confirmUnsaved("file.exit", func() {
		e.confirmCloseWindows(order, position+1, discardConfiguration)
	})
}

func (e *Editor) updateStatus() {
	e.syncActiveWindow()
	d := e.area.document
	e.lineStatus.SetText(fmt.Sprintf(" Line=%d", d.cursor.line+1))
	e.columnStatus.SetText(fmt.Sprintf(" Col=%d", d.column(d.cursor)+1))
	e.statusBar.ResizeItem(e.lineStatus, max(10, len(e.lineStatus.GetText(false))), 0)
	e.statusBar.ResizeItem(e.columnStatus, max(10, len(e.columnStatus.GetText(false))), 0)
	name := e.filename
	if e.changed {
		name += "*"
	}
	if len(e.windows) == 2 {
		name += fmt.Sprintf(" [%d/2]", e.activeWindow+1)
	}
	if e.notice != "" {
		name = e.notice
	}
	e.nameStatus.SetText(" " + name)
	mode := "Insert"
	if e.over {
		mode = "Over"
	}
	e.modeStatus.SetText(" " + mode)
	wrap := "Off"
	if e.wrap {
		wrap = "On"
	}
	e.wrapStatus.SetText(" WW=" + wrap)
	auto := ""
	if e.auto {
		auto = " Auto"
	}
	e.autoStatus.SetText(auto)
}

type editorWindowDivider struct {
	*tview.Box
}

func (d *editorWindowDivider) Draw(screen tcell.Screen) {
	x, y, width, _ := d.GetRect()
	for column := 0; column < width; column++ {
		character := '\u2500'
		if column == width/2 {
			character = '\u253c'
		}
		screen.SetContent(x+column, y, character, nil, panelStyle)
	}
}

func statusCell(text string) *tview.TextView {
	cell := tview.NewTextView().SetText(text).SetWrap(false)
	cell.SetTextColor(statusForegroundColor)
	cell.SetBackgroundColor(statusBackgroundColor)
	return cell
}

var panelStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorSilver)
var menuStyle = panelStyle
var dialogStyle = panelStyle
var dialogTitleStyle = panelStyle
var blockStyle = panelStyle
var selectedStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
var pendingStyle = tcell.StyleDefault.Foreground(tcell.ColorMaroon).Background(tcell.ColorSilver)
var contentStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
var contentBackgroundColor = tcell.ColorBlack
var markerStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
var panelShadowStyle = tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
var statusForegroundColor = tcell.ColorBlack
var statusBackgroundColor = tcell.ColorSilver
var mnemonicForegroundColor = tcell.ColorDefault
var scrollbarThumbColor = tcell.ColorBlack

type editorMenuBar struct {
	*tview.Box
	editor *Editor
}

func (b *editorMenuBar) Draw(screen tcell.Screen) {
	x, y, width, _ := b.GetRect()
	for col := 0; col < width; col++ {
		screen.SetContent(x+col, y, ' ', nil, panelStyle)
	}
	offset := 2
	for _, key := range b.editor.menuOrder {
		menu, _ := b.editor.menu(key)
		if key == 'H' {
			offset = max(offset, width-10)
		}
		b.editor.menuX[key] = x + offset
		style := panelStyle
		if b.editor.menuOpen && b.editor.currentMenu == key {
			style = selectedStyle
		}
		for i, ch := range menu.label {
			if offset+i >= width {
				break
			}
			screen.SetContent(x+offset+i, y, ch, nil, style)
		}
		offset += len(menu.label) + 2
	}
}

func (b *editorMenuBar) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return b.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, _ func(tview.Primitive)) (bool, tview.Primitive) {
		if !b.editor.activeMouseClick(action) || !b.InRect(event.Position()) {
			return false, nil
		}
		x, _ := event.Position()
		for _, key := range b.editor.menuOrder {
			menu, _ := b.editor.menu(key)
			left := b.editor.menuX[key]
			if x >= left && x < left+len(menu.label) {
				b.editor.openMenu(key)
				return true, nil
			}
		}
		return true, nil
	})
}

type editorDropdown struct {
	*tview.Box
	editor   *Editor
	items    []commandSpec
	selected int
	top      int
	rows     []int
}

func (d *editorDropdown) GetItemCount() int { return len(d.items) }

func (d *editorDropdown) Draw(screen tcell.Screen) {
	sw, sh := screen.Size()
	width, totalRows := 0, 0
	d.rows = d.rows[:0]
	for i, item := range d.items {
		if item.separator {
			d.rows = append(d.rows, -1)
			totalRows++
		}
		d.rows = append(d.rows, i)
		totalRows++
		shortcut := ""
		if len(item.keys) > 0 {
			shortcut = item.keys[0]
		}
		width = max(width, len(item.label)+len(shortcut)+5)
	}
	width = min(width, sw)
	height := min(totalRows+2, max(0, sh-2))
	x := max(0, min(d.editor.menuX[d.editor.currentMenu]-1, sw-width))
	d.SetRect(x, 1, width, height)
	drawPanel(screen, x, 1, width, height, "", false)
	visible := max(0, height-2)
	selectedRow := 0
	for row, item := range d.rows {
		if item == d.selected {
			selectedRow = row
		}
	}
	if selectedRow < d.top {
		d.top = selectedRow
	}
	if selectedRow >= d.top+visible {
		d.top = max(0, selectedRow-visible+1)
	}
	for row := 0; row < visible && row+d.top < len(d.rows); row++ {
		index := d.rows[row+d.top]
		if index < 0 {
			for col := 1; col < width-1; col++ {
				screen.SetContent(x+col, row+2, '\u2500', nil, panelStyle)
			}
			continue
		}
		item := d.items[index]
		style := menuStyle
		mnemonic := item.mnemonic
		available := d.editor.available(item)
		if !available {
			style = pendingStyle
			mnemonic = 0
		}
		if index == d.selected {
			if available {
				style = selectedStyle
			} else {
				_, background, _ := selectedStyle.Decompose()
				style = style.Background(background)
			}
		}
		for col := 1; col < width-1; col++ {
			screen.SetContent(x+col, row+2, ' ', nil, style)
		}
		drawLabel(screen, x+2, row+2, width-4, item.label, mnemonic, style)
		if len(item.keys) > 0 {
			key := item.keys[0]
			drawLabel(screen, x+width-len(key)-2, row+2, len(key), key, 0, style)
		}
	}
}

func (d *editorDropdown) choose(index int) {
	command := d.items[index]
	d.editor.closeMenu()
	d.editor.execute(command)
}

func (d *editorDropdown) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, _ func(tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			d.selected = (d.selected + len(d.items) - 1) % len(d.items)
		case tcell.KeyDown:
			d.selected = (d.selected + 1) % len(d.items)
		case tcell.KeyHome:
			d.selected = 0
		case tcell.KeyEnd:
			d.selected = len(d.items) - 1
		case tcell.KeyEnter:
			d.choose(d.selected)
			return
		case tcell.KeyRune:
			for i, item := range d.items {
				if item.mnemonic != 0 && item.mnemonic == unicode.ToUpper(event.Rune()) {
					d.choose(i)
					return
				}
			}
			if d.editor.currentMenu == 'V' && unicode.ToUpper(event.Rune()) == 'A' {
				command, _ := commandByID("view.about")
				d.editor.closeMenu()
				d.editor.execute(command)
				return
			}
		case tcell.KeyDelete:
			if d.editor.currentMenu == 'B' {
				command, _ := commandByID("block.delete-direct")
				d.editor.closeMenu()
				d.editor.execute(command)
				return
			}
		}
		if d.editor.currentMenu == 'B' && event.Key() == tcell.KeyRune {
			id := map[rune]string{'L': "block.mark-line", 'E': "block.mark-end"}[unicode.ToUpper(event.Rune())]
			if command, ok := commandByID(id); ok {
				d.editor.closeMenu()
				d.editor.execute(command)
				return
			}
		}
		d.editor.menuNotice(d.items[d.selected])
	})
}

func (d *editorDropdown) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return d.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, _ func(tview.Primitive)) (bool, tview.Primitive) {
		x, y := event.Position()
		if !d.InRect(x, y) {
			// Consume document clicks while a menu is open, but let the menu bar
			// handle clicks that switch menus.
			if y > 0 && d.editor.activeMouseClick(action) {
				d.editor.closeMenu()
				return true, nil
			}
			return false, nil
		}
		_, top, _, _ := d.GetRect()
		row := y - top - 1 + d.top
		if row >= 0 && row < len(d.rows) && d.rows[row] >= 0 {
			d.selected = d.rows[row]
			d.editor.menuNotice(d.items[d.selected])
			if d.editor.activeMouseClick(action) {
				d.choose(d.selected)
			}
		}
		return true, nil
	})
}

func drawLabel(screen tcell.Screen, x, y, width int, text string, mnemonic rune, style tcell.Style) {
	highlighted := false
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		runes := graphemes.Runes()
		cellWidth := graphemes.Width()
		if cellWidth > width || width <= 0 {
			break
		}
		cellStyle := style
		if !highlighted && mnemonic != 0 && unicode.ToUpper(runes[0]) == mnemonic {
			if mnemonicForegroundColor == tcell.ColorDefault {
				cellStyle = style.Reverse(true)
			} else {
				cellStyle = style.Foreground(mnemonicForegroundColor)
			}
			highlighted = true
		}
		screen.SetContent(x, y, runes[0], runes[1:], cellStyle)
		x += cellWidth
		width -= cellWidth
	}
}

func drawPanel(screen tcell.Screen, x, y, width, height int, title string, double bool) {
	sw, sh := screen.Size()
	shadow := panelShadowStyle
	for row := y + 1; row <= y+height && row < sh; row++ {
		for col := x + width; col < x+width+2 && col < sw; col++ {
			screen.SetContent(col, row, ' ', nil, shadow)
		}
	}
	for col := x + 2; col < x+width+2 && col < sw; col++ {
		if y+height < sh {
			screen.SetContent(col, y+height, ' ', nil, shadow)
		}
	}
	style, titleStyle := panelStyle, panelStyle
	horizontal, vertical := '\u2500', '\u2502'
	corners := []rune{'\u250c', '\u2510', '\u2514', '\u2518'}
	if double {
		style, titleStyle = dialogStyle, dialogTitleStyle
		horizontal, vertical = '\u2550', '\u2551'
		corners = []rune{'\u2554', '\u2557', '\u255a', '\u255d'}
	}
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			ch := ' '
			switch {
			case row == 0 && col == 0:
				ch = corners[0]
			case row == 0 && col == width-1:
				ch = corners[1]
			case row == height-1 && col == 0:
				ch = corners[2]
			case row == height-1 && col == width-1:
				ch = corners[3]
			case row == 0 || row == height-1:
				ch = horizontal
			case col == 0 || col == width-1:
				ch = vertical
			}
			screen.SetContent(x+col, y+row, ch, nil, style)
		}
	}
	if title != "" {
		drawLabel(screen, x+max(1, (width-len(title)-2)/2), y, max(0, width-2), " "+title+" ", 0, titleStyle)
	}
}

type editorScrollArrow struct {
	*tview.Box
	editor *Editor
	step   int
}

func (a *editorScrollArrow) Draw(screen tcell.Screen) {
	x, y, width, _ := a.GetRect()
	for col := 0; col < width; col++ {
		screen.SetContent(x+col, y, ' ', nil, panelStyle)
	}
	ch := '\u2192'
	if a.step < 0 {
		ch = '\u2190'
	}
	if width > 0 {
		screen.SetContent(x+width/2, y, ch, nil, panelStyle)
	}
}

func (a *editorScrollArrow) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return a.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, _ func(tview.Primitive)) (bool, tview.Primitive) {
		if !a.InRect(event.Position()) {
			return false, nil
		}
		if a.editor.activeMouseClick(action) && !a.editor.wrap {
			row, column := a.editor.area.GetOffset()
			a.editor.area.SetOffset(row, column+a.step)
		}
		return true, nil
	})
}

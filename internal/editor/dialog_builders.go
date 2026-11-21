package edit

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func (e *Editor) previewDialog(command commandSpec) {
	d := newDialog(command.label, 62, 12)
	d.context = command.id
	switch command.dialog {
	case "file":
		d.title, d.height = "Open file", 19
		switch command.id {
		case "file.save-as", "block.output", "print.pdf":
			d.title = "Save file as"
		case "file.merge":
			d.title = "Merge file"
		case "file.window":
			d.title = "2nd window"
		}
		dir, err := os.Getwd()
		if err != nil {
			e.showMessage("File dialog error", err.Error())
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			e.showMessage("File dialog error", err.Error())
			return
		}
		files, dirs := []string{}, []string{".."}
		for _, entry := range entries {
			if entry.IsDir() {
				dirs = append(dirs, "["+entry.Name()+"]")
			} else {
				files = append(files, entry.Name())
			}
		}
		mask := "*"
		if ext := filepath.Ext(e.filename); ext != "" {
			mask += ext
		}
		d.input("filename", "File name:", mask, 2, 0, 0)
		d.add(dialogWidget{kind: widgetList, id: "files", label: dir, choices: files, x: 2, y: 4, width: 36, height: 10})
		d.add(dialogWidget{kind: widgetList, id: "dirs", label: "Dirs:", choices: dirs, x: 40, y: 4, width: 18, height: 10})
	case "search", "replace":
		d.title = "Search for text"
		findValue, replaceValue := "", ""
		matchCase, backwards := false, false
		if e.lastSearch.valid {
			findValue = e.lastSearch.text
			replaceValue = e.lastSearch.replacement
			matchCase, backwards = e.lastSearch.matchCase, e.lastSearch.backwards
		}
		d.input("find", "Find:", findValue, 2, 0, 0)
		row := 4
		if command.dialog == "replace" {
			d.title, d.height = "Search & Replace", 13
			d.input("replace", "Replace:", replaceValue, 3, 0, 0)
			row++
		}
		d.add(dialogWidget{kind: widgetCheck, id: "case", label: "Upper/lower case must match", checked: matchCase, x: 6, y: row, width: 38})
		d.add(dialogWidget{kind: widgetCheck, id: "backwards", label: "Search backwards", checked: backwards, x: 6, y: row + 1, width: 38})
	case "line", "indent":
		d.height = 8
		label := "Line number:"
		minimum := 1
		if command.dialog == "indent" {
			label, minimum = "Amount:", 0
		} else {
			d.title = "Find line"
		}
		d.input("number", label, "1", 2, minimum, 100000000)
	case "start":
		d.title, d.height = "Starting position", 8
		d.label(2, 2, command.label+" starting at...")
		d.buttons([]string{"top", "cursor", "cancel"}, []string{"Top", "Cursor", "Cancel"}, "top")
	case "layout":
		d.title, d.width, d.height = "Text-layout setup", 44, 14
		d.input("line-length", "Line Length:", "65", 2, 1, 10000)
		d.input("first-margin", "1st line margin:", "0", 3, 0, 10000)
		d.input("left-margin", "Left Margin:", "0", 4, 0, 10000)
		d.label(2, 6, "Begin formatting at:")
		d.add(dialogWidget{kind: widgetRadio, id: "begin", choices: []string{"Cursor", "Top of paragraph"}, selected: 1, x: 4, y: 7, width: 32, height: 2})
	case "display":
		d.title, d.width, d.height = "Display setup", 66, 21
		d.label(2, 2, "Color:")
		d.add(dialogWidget{kind: widgetRadio, id: "color", choices: []string{"Color 1", "Color 2", "VGA Monochrome", "Monochrome", "LCD"}, selected: 4, x: 2, y: 3, width: 25, height: 5})
		d.label(34, 2, "Display lines:")
		d.add(dialogWidget{kind: widgetRadio, id: "lines", choices: []string{"25", "33/35", "40/43", "50", "Don't change"}, x: 34, y: 3, width: 28, height: 5})
		for i, label := range []string{"Display scrollbar", "Blinking bracket", "Zoom window effects", "Large cursor"} {
			d.add(dialogWidget{kind: widgetCheck, id: label, label: label, checked: i < 3, x: 2, y: 10 + i, width: 31})
		}
		d.label(34, 11, "Tab spacing:")
		d.add(dialogWidget{kind: widgetInput, id: "tab-spacing", value: "8", x: 48, y: 11, width: 6, numeric: true, minimum: 1, maximum: 32})
	case "mouse":
		d.title, d.width, d.height = "Mouse setup", 64, 17
		d.label(2, 2, "Mouse button:")
		d.add(dialogWidget{kind: widgetRadio, id: "button", choices: []string{"Left", "Right"}, x: 2, y: 3, width: 22, height: 2})
		d.label(28, 2, "Scroll delay:")
		d.add(dialogWidget{kind: widgetRadio, id: "delay", choices: []string{"No delay", "12 ms", "18 ms", "26 ms", "Sync to clock"}, x: 28, y: 3, width: 30, height: 5})
		d.add(dialogWidget{kind: widgetCheck, id: "speed", label: "Adj mouse cursor speed", checked: true, x: 2, y: 10, width: 40})
	case "config":
		d.title, d.height = "Save configuration", 10
		d.label(2, 2, "Update:")
		d.add(dialogWidget{kind: widgetRadio, id: "scope", choices: []string{"current config file", "config for this directory"}, x: 2, y: 3, width: 56, height: 2})
	default:
		e.showMessage("Not available", command.description+" Planned "+command.phase+" ["+command.id+"]")
		return
	}
	if command.id != "search.find" && command.id != "search.replace" && command.id != "search.line" {
		d.label(2, 1, "Preview only - execution planned "+command.phase)
	}
	if command.dialog != "start" {
		d.buttons([]string{"ok", "help", "cancel"}, []string{"OK", "F1=Help", "Cancel"}, "ok")
	}
	d.result = func(id string) {
		if id == "help" {
			e.showHelp(command.id)
			return
		}
		e.closeDialog()
		if id == "cancel" {
			return
		}
		if command.id == "search.line" {
			line, err := strconv.Atoi(d.value("number"))
			if err != nil {
				e.showMessage("Invalid value", "Line number must be a positive integer.")
				return
			}
			e.goToLine(line)
			return
		}
		if command.id == "search.find" || command.id == "search.replace" {
			query := d.value("find")
			if query == "" {
				e.showMessage("Search error", "Find text cannot be empty.")
				return
			}
			search := searchState{
				text:        query,
				replacement: e.lastSearch.replacement,
				matchCase:   d.checked("case"),
				backwards:   d.checked("backwards"),
				valid:       true,
				start:       -1,
				end:         -1,
			}
			if command.id == "search.replace" {
				search.replacement = d.value("replace")
			}
			e.lastSearch = search
			if command.id == "search.replace" {
				e.beginReplaceSearch(search, d.value("replace"))
			} else if !e.findSearchText(query, search.matchCase, search.backwards) {
				e.showMessage("Search", "Text not found.")
			}
			return
		}
		if id != "cancel" {
			e.showMessage("Not available", fmt.Sprintf("%s is planned %s [%s]. Preview values were not applied.", command.label, command.phase, command.id))
		}
	}
	e.openDialog(d)
}

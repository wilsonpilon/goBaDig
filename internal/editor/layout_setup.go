package edit

import (
	"fmt"
	"strconv"
)

func (e *Editor) layoutSetup() {
	document := e.area.document
	d := newDialog("Text-layout setup", 44, 14)
	d.context = "layout.setup"
	d.input("line-length", "Line Length:", strconv.Itoa(document.lineLength), 2, 1, 10000)
	d.input("first-margin", "1st line margin:", strconv.Itoa(document.firstMargin), 3, 0, 10000)
	d.input("left-margin", "Left Margin:", strconv.Itoa(document.leftMargin), 4, 0, 10000)
	d.label(2, 6, "Begin formatting at:")
	start := 1
	if document.formatAtCursor {
		start = 0
	}
	d.add(dialogWidget{
		kind: widgetRadio, id: "begin", choices: []string{"Cursor", "Top of paragraph"},
		selected: start, x: 4, y: 7, width: 32, height: 2,
	})
	d.buttons([]string{"ok", "help", "cancel"}, []string{"OK", "F1=Help", "Cancel"}, "ok")
	d.result = func(id string) {
		if id == "help" {
			e.showHelp(d.context)
			return
		}
		if id == "ok" {
			values := make([]int, 3)
			for index, field := range []string{"line-length", "first-margin", "left-margin"} {
				value, err := strconv.Atoi(d.value(field))
				if err != nil {
					e.showMessage("Invalid value", fmt.Sprintf("%s: %v", field, err))
					return
				}
				values[index] = value
			}
			document.lineLength, document.firstMargin, document.leftMargin = values[0], values[1], values[2]
			document.formatAtCursor = d.widgets[d.fields[3]].selected == 0
		}
		e.closeDialog()
		if id == "ok" {
			e.notice = "Text-layout setup applied; existing text was not reformatted."
			e.updateStatus()
		}
	}
	e.openDialog(d)
}

package edit

import (
	"fmt"
	"strconv"
)

func (e *Editor) printSetup() {
	document := e.area.document
	d := newDialog("Print setup", 48, 18)
	d.context = "print.setup"
	d.input("top-margin", "Top margin:", strconv.Itoa(document.printTopMargin), 2, 0, 10000)
	d.input("left-margin", "Left margin:", strconv.Itoa(document.printLeftMargin), 3, 0, 10000)
	d.input("page-length", "Page length:", strconv.Itoa(document.pageLength), 4, 1, 10000)
	d.add(dialogWidget{
		kind: widgetCheck, id: "formfeed", label: "Formfeed between pages",
		checked: document.formfeedBetween, x: 2, y: 6, width: 40,
	})
	d.label(2, 8, "Formfeed after printing:")
	d.add(dialogWidget{
		kind: widgetRadio, id: "after", choices: []string{"No", "One page", "Two pages"},
		selected: document.formfeedAfter, x: 4, y: 9, width: 32, height: 3,
	})
	d.buttons([]string{"ok", "help", "cancel"}, []string{"OK", "F1=Help", "Cancel"}, "ok")
	d.result = func(action string) {
		if action == "help" {
			e.showHelp(d.context)
			return
		}
		e.closeDialog()
		if action != "ok" {
			return
		}

		values := make([]int, 3)
		for index, field := range []struct {
			id, label string
		}{
			{"top-margin", "Top margin"},
			{"left-margin", "Left margin"},
			{"page-length", "Page length"},
		} {
			value, err := strconv.Atoi(d.value(field.id))
			if err != nil {
				e.showMessage("Invalid value", fmt.Sprintf("%s: %v", field.label, err))
				return
			}
			minimum := 0
			if field.id == "page-length" {
				minimum = 1
			}
			if value < minimum || value > 10000 {
				e.showMessage("Invalid value", fmt.Sprintf("%s must be between %d and 10000.", field.label, minimum))
				return
			}
			values[index] = value
		}

		after := -1
		for index := range d.widgets {
			if d.widgets[index].id == "after" {
				after = d.widgets[index].selected
				break
			}
		}
		if after < 0 || after > 2 {
			e.showMessage("Invalid value", "Formfeed after printing must be No, One page, or Two pages.")
			return
		}

		document.printTopMargin = values[0]
		document.printLeftMargin = values[1]
		document.pageLength = values[2]
		document.formfeedBetween = d.checked("formfeed")
		document.formfeedAfter = after
		e.notice = "Print setup applied [print.setup]."
		e.updateStatus()
	}
	e.openDialog(d)
}

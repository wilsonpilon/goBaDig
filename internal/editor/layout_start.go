package edit

func (e *Editor) layoutStartDialog(command commandSpec, action func(bool) bool) {
	d := newDialog("Starting position", 46, 9)
	d.context = command.id
	d.label(2, 2, command.label+" starting at...")
	d.buttons(
		[]string{"top", "cursor", "help", "cancel"},
		[]string{"Top", "Cursor", "F1=Help", "Cancel"},
		"top",
	)
	d.result = func(id string) {
		if id == "help" {
			e.showHelp(d.context)
			return
		}
		if id != "top" && id != "cursor" {
			e.closeDialog()
			return
		}
		changed := action(id == "cursor")
		e.closeDialog()
		e.area.notify(changed)
		if changed {
			e.notice = command.label + " completed."
		} else {
			e.notice = "No changes were needed."
		}
		e.updateStatus()
	}
	e.openDialog(d)
}

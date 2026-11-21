package edit

func (e *Editor) documentFormatDialog() {
	d := newDialog("Document format", 46, 9)
	d.context = "layout.document"
	d.label(2, 2, "Begin formatting at...")
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
		changed := e.area.document.formatDocument(id == "cursor")
		e.closeDialog()
		e.area.notify(changed)
		if changed {
			e.notice = "Document formatted."
		} else {
			e.notice = "Document has no text to format."
		}
		e.updateStatus()
	}
	e.openDialog(d)
}

func (e *Editor) formatParagraph() {
	changed := e.area.document.formatParagraph()
	e.area.notify(changed)
	if changed {
		e.notice = "Paragraph formatted."
	} else {
		e.notice = "No paragraph formatting was needed."
	}
	e.updateStatus()
}

func (e *Editor) centerCurrentLine() {
	changed := e.area.document.centerLine()
	e.area.notify(changed)
	if changed {
		e.notice = "Line centered."
	} else {
		e.notice = "Line is empty or already centered."
	}
	e.updateStatus()
}

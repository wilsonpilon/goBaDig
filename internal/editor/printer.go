package edit

import "fmt"

func (e *Editor) printDocument() {
	text, err := e.currentFileText()
	if err != nil {
		e.fileError(err, "print.document")
		return
	}
	stream, err := e.area.document.printStreamForText(text)
	if err != nil {
		e.fileError(err, "print.document")
		return
	}
	e.sendPrintStream(stream, "print.document", fmt.Sprintf("Sent %d bytes to the default printer.", len(stream)))
}

func (e *Editor) printBlock() {
	document := e.area.document
	start, end, ok := document.markedBlock()
	if !ok {
		e.notice = "Set two block markers before using Print Block."
		e.updateStatus()
		return
	}
	stream, err := document.printStreamForText(document.text[start:end])
	if err != nil {
		e.fileError(err, "print.block")
		return
	}
	e.sendPrintStream(stream, "print.block", fmt.Sprintf("Sent %d bytes to the default printer.", len(stream)))
}

func (e *Editor) ejectPage() {
	e.sendPrintStream([]byte{'\f'}, "print.eject", "Sent an eject-page formfeed to the default printer.")
}

func (e *Editor) sendPrintStream(stream []byte, context, successNotice string) {
	if e.printer == nil {
		e.fileError(fmt.Errorf("default printer is unavailable"), context)
		return
	}
	if err := e.printer(stream); err != nil {
		e.fileError(fmt.Errorf("sending to the default printer: %w", err), context)
		return
	}
	e.notice = successNotice
	e.updateStatus()
}

package edit

import (
	"errors"
	"github.com/wilsonpilon/goBaDig/internal/printpdf"
)

func (d *document) pdfForText(text string) ([]byte, error) {
	stream, err := d.printStreamForText(text)
	if err != nil {
		return nil, err
	}
	if len(stream) < d.formfeedAfter {
		return nil, errors.New("invalid trailing print eject count")
	}
	stream = stream[:len(stream)-d.formfeedAfter]
	return printpdf.Render(stream, false)
}

func (e *Editor) exportPDFTo(path string, overwrite bool) {
	if sameEditorPath(path, e.path) {
		e.fileError(errors.New("the PDF output path cannot be the current document path"), "print.pdf")
		return
	}
	text, err := e.currentFileText()
	if err != nil {
		e.fileError(err, "print.pdf")
		return
	}
	data, err := e.area.document.pdfForText(text)
	if err != nil {
		e.fileError(err, "print.pdf")
		return
	}
	err = saveEditorFile(path, string(data), e.backupCharacter, nil, overwrite)
	if errors.Is(err, errFileExists) {
		e.confirmOverwriteAction(path, "print.pdf", false, func() {
			e.exportPDFTo(path, true)
		})
		return
	}
	if err != nil {
		e.fileError(err, "print.pdf")
		return
	}
	e.closeDialog()
	e.notice = "Exported PDF to " + path
	e.updateStatus()
}

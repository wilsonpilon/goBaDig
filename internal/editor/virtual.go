package edit

import (
	"errors"
	"strings"

	"github.com/rivo/tview"
)

// VirtualDocument provides a fixed-path document whose storage belongs to its caller.
// Save must detect stale content; it is never retried with overwrite enabled.
type VirtualDocument struct {
	Path   string
	Text   string
	Exists bool
	Read   func() (string, error)
	Save   func(text string) error
}

// RunVirtualDocument runs WIBEdit without reading or staging host files.
func RunVirtualDocument(source VirtualDocument) error {
	app := tview.NewApplication().EnableMouse(true)
	editor := NewEditor(app, app.Stop)
	if err := editor.OpenVirtualDocument(source); err != nil {
		return err
	}
	app.SetInputCapture(editor.Capture)
	app.SetRoot(editor, true)
	return app.Run()
}

// OpenVirtualDocument binds an editor to a single virtual document.
func (e *Editor) OpenVirtualDocument(source VirtualDocument) error {
	if source.Path == "" || source.Read == nil || source.Save == nil {
		return errors.New("a virtual document requires a path, reader and save callback")
	}
	if len(e.windows) != 1 || e.changed {
		return errors.New("open a virtual document in a fresh editor")
	}
	e.virtual = &source
	e.setFile(source.Path, source.Text, source.Exists)
	return nil
}

func virtualCommandAllowed(id string) bool {
	switch id {
	case "file.new", "file.open", "file.save-as", "file.merge", "file.window",
		"file.close-window", "file.shell", "block.output", "options.save",
		"view.information":
		return false
	}
	return !strings.HasPrefix(id, "print.")
}

func (e *Editor) writeFile(path, text string, expected *string, overwrite bool) error {
	if e.virtual == nil {
		return saveEditorFile(path, text, e.backupCharacterForSave(), expected, overwrite)
	}
	if path != e.virtual.Path {
		return errors.New("Save As and host paths are unavailable in a FILE SPACE session")
	}
	return e.virtual.Save(text)
}

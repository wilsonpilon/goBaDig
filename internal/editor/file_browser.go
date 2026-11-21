package edit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func fileMatches(mask, name string) (bool, error) {
	if mask == "*.*" {
		mask = "*"
	}
	if runtime.GOOS == "windows" {
		mask, name = strings.ToLower(mask), strings.ToLower(name)
	}
	return filepath.Match(mask, name)
}

func (e *Editor) fileDialog(context string, done func()) {
	if e.virtual != nil {
		e.fileError(errors.New("host file dialogs are unavailable in a FILE SPACE session"), context)
		return
	}
	blockStart, blockEnd := 0, 0
	if context == "block.output" {
		var ok bool
		blockStart, blockEnd, ok = e.area.document.markedBlock()
		if !ok {
			e.notice = "Set two block markers before using Output to file."
			e.updateStatus()
			return
		}
	}
	dir := e.lastDir
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			e.fileError(err, context)
			return
		}
	}
	title := "Open file"
	if context == "file.merge" {
		title = "Merge file"
	} else if context == "block.output" {
		title = "Output block to file"
	} else if context == "print.pdf" {
		title = "Export PDF"
	}
	d := newDialog(title, 64, 19)
	d.context = context
	value, mask := "*", "*"
	if ext := filepath.Ext(e.path); ext != "" {
		mask += ext
		value = mask
	}
	if context == "file.save-as" {
		d.title, value = "Save file as", ""
		if e.path != "" {
			value = filepath.Base(e.path)
		}
	} else if context == "block.output" {
		value = ""
	} else if context == "print.pdf" {
		mask = "*.pdf"
		base := strings.TrimSuffix(e.filename, filepath.Ext(e.filename))
		if base == "" {
			base = "Untitled"
		}
		value = base + ".pdf"
	}
	d.input("filename", "File name:", value, 2, 0, 0)
	d.add(dialogWidget{kind: widgetList, id: "files", x: 2, y: 4, width: 36, height: 10})
	d.add(dialogWidget{kind: widgetList, id: "dirs", label: "Dirs / drives:", x: 40, y: 4, width: 20, height: 10})
	d.buttons([]string{"ok", "help", "cancel"}, []string{"OK", "F1=Help", "Cancel"}, "ok")
	filename, files, directories := &d.widgets[1], &d.widgets[2], &d.widgets[3]
	var directoryPaths []string
	setName := func(name string) {
		filename.value, filename.cursor, filename.fresh = name, len([]rune(name)), true
	}
	refresh := func(nextDir, nextMask string) error {
		if _, err := fileMatches(nextMask, ""); err != nil {
			return fmt.Errorf("invalid file mask %q: %w", nextMask, err)
		}
		entries, err := os.ReadDir(nextDir)
		if err != nil {
			return err
		}
		drives, err := editorDrives()
		if err != nil {
			return fmt.Errorf("listing drives: %w", err)
		}
		var names []string
		labels, paths := []string{"[..]"}, []string{filepath.Dir(nextDir)}
		for _, entry := range entries {
			path := filepath.Join(nextDir, entry.Name())
			isDir := entry.IsDir()
			if entry.Type()&os.ModeSymlink != 0 {
				info, err := os.Stat(path)
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
				isDir = err == nil && info.IsDir()
			}
			if isDir {
				labels, paths = append(labels, "["+entry.Name()+"]"), append(paths, path)
			} else if match, err := fileMatches(nextMask, entry.Name()); err != nil {
				return err
			} else if match {
				names = append(names, entry.Name())
			}
		}
		for _, drive := range drives {
			labels, paths = append(labels, "["+drive+"]"), append(paths, drive)
		}
		dir, mask, e.lastDir = nextDir, nextMask, nextDir
		files.label, files.choices, files.selected, files.top = nextDir, names, 0, 0
		directories.choices, directories.selected, directories.top = labels, 0, 0
		directoryPaths = paths
		return nil
	}
	if err := refresh(dir, mask); err != nil {
		e.fileError(err, context)
		return
	}
	submit := func() {
		name := filename.value
		if strings.TrimSpace(name) == "" {
			e.fileError(errors.New("a file name is required"), context)
			return
		}
		target := name
		if !filepath.IsAbs(target) {
			target = filepath.Join(dir, target)
		}
		target = filepath.Clean(target)
		if strings.ContainsAny(filepath.Base(target), "*?[") {
			if err := refresh(filepath.Dir(target), filepath.Base(target)); err != nil {
				e.fileError(err, context)
			}
			return
		}
		if info, err := os.Stat(target); err == nil && info.IsDir() {
			if err := refresh(target, mask); err != nil {
				e.fileError(err, context)
			} else {
				setName(mask)
			}
			return
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			e.fileError(err, context)
			return
		}
		if context == "print.pdf" {
			extension := filepath.Ext(target)
			if extension == "" {
				target += ".pdf"
			} else if !strings.EqualFold(extension, ".pdf") {
				e.fileError(errors.New("PDF output files must use the .pdf extension"), context)
				return
			}
			e.exportPDFTo(target, false)
			return
		}
		if context == "block.output" {
			if sameEditorPath(target, e.path) {
				e.fileError(errors.New("the block output path cannot be the current document path"), context)
				return
			}
			blockText := e.area.document.text[blockStart:blockEnd]
			e.saveBlockTo(target, blockText, false, func() {
				e.closeDialog()
				if done != nil {
					done()
				}
			})
			return
		}
		if context == "file.save-as" {
			e.saveTo(target, false, func() {
				e.closeDialog()
				if done != nil {
					done()
				}
			})
			return
		}
		path, text, err := readEditorFile(target)
		if errors.Is(err, os.ErrNotExist) && context != "file.merge" {
			e.offerNewFile(path, func() {
				e.closeDialog()
				if context == "file.window" {
					e.openSecondWindow(path, "", false)
					return
				}
				e.confirmUnsaved("file.open", func() { e.setFile(path, "", false) })
			})
			return
		}
		if err != nil {
			e.fileError(err, context)
			return
		}
		e.closeDialog()
		if context == "file.merge" {
			e.mergeText(text)
			return
		}
		if context == "file.window" {
			e.openSecondWindow(path, text, true)
			return
		}
		e.confirmUnsaved("file.open", func() { e.setFile(path, text, true) })
	}
	files.onSelect = func(index int) {
		if len(files.choices) > 0 {
			setName(files.choices[index])
		}
	}
	files.onActivate = func(index int) {
		if len(files.choices) > 0 {
			setName(files.choices[index])
			submit()
		}
	}
	directories.onActivate = func(index int) {
		if err := refresh(directoryPaths[index], mask); err != nil {
			e.fileError(err, context)
		} else {
			setName(mask)
		}
	}
	d.result = func(id string) {
		switch id {
		case "help":
			e.showHelp(context)
		case "ok":
			submit()
		default:
			e.closeDialog()
		}
	}
	e.openDialog(d)
}

func sameEditorPath(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	left, right = filepath.Clean(left), filepath.Clean(right)
	if runtime.GOOS == "windows" {
		if strings.EqualFold(left, right) {
			return true
		}
	}
	if left == right {
		return true
	}
	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)
	return leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo)
}

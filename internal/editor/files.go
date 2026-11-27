package edit

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const fileChunkBytes = 1 << 20

var (
	errFileExists  = errors.New("the destination already exists")
	errFileChanged = errors.New("the destination changed on disk")
)

func readEditorFile(name string) (path, text string, err error) {
	if name == "" {
		return "", "", errors.New("a file name is required")
	}
	path, err = filepath.Abs(name)
	if err != nil {
		return path, "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return path, "", err
	}
	if !info.Mode().IsRegular() {
		return path, "", fmt.Errorf("%s is not a regular file", path)
	}
	data, err := os.ReadFile(path)
	return path, string(data), err
}

func fileChunkEnd(text string, start int) int {
	end := min(len(text), start+fileChunkBytes)
	if end < len(text) && end > start && text[end-1] == '\r' && text[end] == '\n' {
		end--
	}
	for end > start && end < len(text) && !utf8.RuneStart(text[end]) {
		end--
	}
	return end
}

func backupName(path string, character rune) string {
	ext := filepath.Ext(path)
	letters := []rune(strings.TrimPrefix(ext, "."))
	letters = letters[:min(2, len(letters))]
	return strings.TrimSuffix(path, ext) + "." + string(character) + strings.ToUpper(string(letters))
}

func atomicWriteFile(path string, text string, mode fs.FileMode, beforeCommit func() error) (err error) {
	file, err := os.CreateTemp(filepath.Dir(path), ".wibedit-*")
	if err != nil {
		return err
	}
	temp := file.Name()
	closed, committed := false, false
	defer func() {
		if !closed {
			err = errors.Join(err, file.Close())
		}
		if !committed {
			if cleanupErr := os.Remove(temp); cleanupErr != nil && !errors.Is(cleanupErr, os.ErrNotExist) {
				err = errors.Join(err, fmt.Errorf("removing temporary file: %w", cleanupErr))
			}
		}
	}()
	if err = file.Chmod(mode.Perm()); err != nil {
		return err
	}
	var count int
	if count, err = file.WriteString(text); err != nil {
		return err
	}
	if count != len(text) {
		return io.ErrShortWrite
	}
	if err = file.Sync(); err != nil {
		return err
	}
	err = file.Close()
	closed = true
	if err != nil {
		return err
	}
	if beforeCommit != nil {
		if err = beforeCommit(); err != nil {
			return err
		}
	}
	if err = os.Rename(temp, path); err != nil {
		return err
	}
	committed = true
	return nil
}

func saveEditorFile(path, text string, character rune, expected *string, overwrite bool) error {
	info, err := os.Lstat(path)
	exists := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	mode := fs.FileMode(0600)
	old := ""
	if exists {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file; symbolic links must be saved through their real target", path)
		}
		if info.Mode().Perm()&0222 == 0 {
			return fmt.Errorf("%s is read-only", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		old, mode = string(data), info.Mode()
		if expected == nil && !overwrite {
			return errFileExists
		}
	}
	if expected != nil && (!exists || old != *expected) {
		return fmt.Errorf("%w: use Save As to choose another destination or explicitly overwrite", errFileChanged)
	}
	if exists && old == text {
		return nil
	}
	if exists && character != 0 {
		if err := atomicWriteFile(backupName(path, character), old, mode, nil); err != nil {
			return fmt.Errorf("creating backup: %w", err)
		}
	}
	check := func() error {
		data, err := os.ReadFile(path)
		if !exists {
			if err == nil {
				return errFileExists
			}
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if err != nil {
			return err
		}
		if string(data) != old {
			return fmt.Errorf("%w while saving; its contents were not replaced", errFileChanged)
		}
		return nil
	}
	return atomicWriteFile(path, text, mode, check)
}

func (e *Editor) setFile(path, text string, known bool) {
	e.closeAutocomplete()
	e.fileBase = 0
	loadedText := text
	if known && len(text) > fileChunkBytes {
		loadedText = text[:fileChunkEnd(text, 0)]
	}
	e.area.SetText(loadedText, false)
	e.area.document.setSourceBase(e.fileBase)
	e.path, e.diskText, e.diskKnown = path, text, known
	e.filename = "Untitled"
	if path != "" {
		e.filename, e.lastDir = filepath.Base(path), filepath.Dir(path)
	}
	e.area.dialect = dialectForPath(path)
	e.over = e.startMode == 1
	e.area.over = e.over
	e.changed, e.literal = false, false
	e.notice = ""
	if len(loadedText) < len(text) {
		e.notice = "Loaded first 1 MiB. Use File > Load More."
	}
	e.updateStatus()
	e.app.SetFocus(e.area)
}

func (e *Editor) fileError(err error, context string) {
	e.showMessage("File error", err.Error())
	e.dialogs[len(e.dialogs)-1].context = context
}

func (e *Editor) saveTo(path string, overwrite bool, done func()) {
	var expected *string
	if path == e.path && e.diskKnown && !overwrite {
		expected = &e.diskText
	}
	text, err := e.currentFileText()
	if err != nil {
		e.fileError(err, "file.save")
		return
	}
	err = e.writeFile(path, text, expected, overwrite)
	if err != nil && e.virtual != nil {
		e.fileError(err, "file.save")
		return
	}
	if errors.Is(err, errFileExists) {
		e.confirmOverwrite(path, done, false)
		return
	}
	if errors.Is(err, errFileChanged) {
		e.confirmOverwrite(path, done, true)
		return
	}
	if err != nil {
		e.fileError(err, "file.save")
		return
	}
	e.path, e.filename, e.lastDir = path, filepath.Base(path), filepath.Dir(path)
	e.area.dialect = dialectForPath(path)
	e.diskText, e.diskKnown, e.changed = text, true, false
	e.area.document.setSourceBase(e.fileBase)
	e.notice = "Saved " + path
	e.updateStatus()
	if done != nil {
		done()
	}
}

func (e *Editor) saveDocument(done func()) {
	if e.path == "" {
		e.fileDialog("file.save-as", done)
	} else {
		e.saveTo(e.path, false, done)
	}
}

func (e *Editor) currentFileText() (string, error) {
	if !e.diskKnown || e.path == "" {
		return e.area.GetText(), nil
	}
	d := e.area.document
	end := d.sourcePositionAt(len(d.text))
	if e.fileBase < 0 || end < e.fileBase || end > len(e.diskText) {
		return "", errors.New("invalid partial-file boundaries; the document was not saved")
	}
	return e.diskText[:e.fileBase] + d.text + e.diskText[end:], nil
}

func (e *Editor) verifyDiskFile() error {
	if e.virtual != nil {
		text, err := e.virtual.Read()
		if err != nil {
			return err
		}
		if text != e.diskText {
			return errFileChanged
		}
		return nil
	}
	path, text, err := readEditorFile(e.path)
	if err != nil {
		return err
	}
	if path != e.path || text != e.diskText {
		return fmt.Errorf("%w; reopen the file before continuing", errFileChanged)
	}
	return nil
}

func (e *Editor) loadMore() {
	if e.path == "" || !e.diskKnown {
		e.notice = "Open a file before using Load More."
		e.updateStatus()
		return
	}
	if err := e.verifyDiskFile(); err != nil {
		e.fileError(err, "file.load-more")
		return
	}
	d := e.area.document
	offset := d.sourcePositionAt(len(d.text))
	if offset < e.fileBase || offset > len(e.diskText) {
		e.fileError(errors.New("invalid partial-file boundary; no data was loaded"), "file.load-more")
		return
	}
	if offset == len(e.diskText) {
		e.notice = "The entire file is already loaded."
		e.updateStatus()
		return
	}
	end := fileChunkEnd(e.diskText, offset)
	if end <= offset {
		e.fileError(errors.New("unable to advance the file chunk boundary"), "file.load-more")
		return
	}
	d.appendSourceText(e.diskText[offset:end], offset)
	e.notice = fmt.Sprintf("Loaded %d more bytes; %d bytes remain.", end-offset, len(e.diskText)-end)
	e.updateStatus()
}

func (e *Editor) writeToCursor() {
	if e.path == "" || !e.diskKnown {
		e.notice = "Open or save a file before using Write to Cursor."
		e.updateStatus()
		return
	}
	d := e.area.document
	if d.sourceReordered {
		e.notice = "Save the moved block before using Write to Cursor."
		e.updateStatus()
		return
	}
	cursor := d.byteOffset(d.cursor)
	if cursor == 0 {
		e.notice = "Cursor is at the top; nothing was written."
		e.updateStatus()
		return
	}
	sourceCursor := d.sourcePositionAt(cursor)
	sourceEnd := d.sourcePositionAt(len(d.text))
	if e.fileBase < 0 || sourceCursor < e.fileBase || sourceEnd < sourceCursor || sourceEnd > len(e.diskText) {
		e.fileError(errors.New("invalid partial-file boundaries; the document was not written"), "file.write-cursor")
		return
	}
	if err := e.verifyDiskFile(); err != nil {
		e.fileError(err, "file.write-cursor")
		return
	}
	prefix := d.text[:cursor]
	nextDisk := e.diskText[:e.fileBase] + prefix + e.diskText[sourceCursor:]
	if err := e.writeFile(e.path, nextDisk, &e.diskText, false); err != nil {
		e.fileError(err, "file.write-cursor")
		return
	}
	newBase := e.fileBase + len(prefix)
	remainder := d.text[cursor:]
	sourceRuns := d.sourceRunsAfter(cursor, sourceCursor, newBase)
	expectedEnd := newBase + sourceEnd - sourceCursor
	e.diskText, e.fileBase = nextDisk, newBase
	e.area.SetText(remainder, false)
	e.area.document.setSourceRuns(sourceRuns)
	e.changed = remainder != nextDisk[newBase:expectedEnd]
	e.notice = fmt.Sprintf("Wrote %d bytes through the cursor.", len(prefix))
	e.updateStatus()
}

func (e *Editor) topOfFile() {
	if e.path == "" || !e.diskKnown {
		e.notice = "Open a file before using Top of File."
		e.updateStatus()
		return
	}
	e.confirmUnsaved("file.top", e.reloadTopOfFile)
}

func (e *Editor) reloadTopOfFile() {
	if err := e.verifyDiskFile(); err != nil {
		e.fileError(err, "file.top")
		return
	}
	e.setFile(e.path, e.diskText, true)
	e.notice = "Returned to the top of the file."
	e.updateStatus()
}

func (e *Editor) mergeText(text string) {
	d := e.area.document
	offset := d.byteOffset(d.cursor)
	changed := d.replaceBytes(offset, offset, text)
	e.area.notify(changed)
	e.notice = fmt.Sprintf("Merged %d bytes at the cursor.", len(text))
	e.updateStatus()
}

func (e *Editor) saveBlockTo(path, text string, overwrite bool, done func()) {
	if e.virtual != nil {
		e.fileError(errors.New("block output is unavailable in a FILE SPACE session"), "block.output")
		return
	}
	err := saveEditorFile(path, text, e.backupCharacterForSave(), nil, overwrite)
	if errors.Is(err, errFileExists) {
		e.confirmOverwriteAction(path, "block.output", false, func() {
			e.saveBlockTo(path, text, true, done)
		})
		return
	}
	if err != nil {
		e.fileError(err, "block.output")
		return
	}
	e.notice = fmt.Sprintf("Wrote %d bytes to %s.", len(text), path)
	e.updateStatus()
	if done != nil {
		done()
	}
}

func (e *Editor) confirmOverwrite(path string, done func(), changed bool) {
	context := "file.save-as"
	if changed {
		context = "file.save"
	}
	e.confirmOverwriteAction(path, context, changed, func() {
		e.saveTo(path, true, done)
	})
}

func (e *Editor) confirmOverwriteAction(path, context string, changed bool, overwrite func()) {
	title, message := "Overwrite file?", filepath.Base(path)+" already exists."
	if changed {
		title, message = "File changed on disk", filepath.Base(path)+" changed or was removed since it was opened."
	}
	d := newDialog(title, 64, 8)
	d.context = context
	d.label(2, 2, message)
	d.label(2, 3, "Replace or recreate it? Previous contents, if any, will be backed up.")
	d.buttons([]string{"overwrite", "cancel"}, []string{"Overwrite", "Cancel"}, "cancel")
	d.focus = 1
	d.result = func(id string) {
		if id == "help" {
			e.showHelp(d.context)
			return
		}
		e.closeDialog()
		if id == "overwrite" {
			overwrite()
		}
	}
	e.openDialog(d)
}

func (e *Editor) offerNewFile(path string, done func()) {
	d := newDialog("Create file?", 64, 8)
	d.context = "file.open"
	d.label(2, 2, filepath.Base(path)+" was not found.")
	d.label(2, 3, "Create a new document? Disk creation happens on Save.")
	d.buttons([]string{"create", "cancel"}, []string{"Create", "Cancel"}, "cancel")
	d.widgets[d.fields[0]].mnemonic = 'R'
	d.focus = 1
	d.result = func(id string) {
		if id == "help" {
			e.showHelp(d.context)
			return
		}
		e.closeDialog()
		if id == "create" {
			done()
		}
	}
	e.openDialog(d)
}

func (e *Editor) confirmUnsaved(context string, done func()) {
	if !e.changed {
		done()
		return
	}
	d := newDialog("Unsaved document", 62, 8)
	d.context = context
	d.label(2, 2, "This document has changed. Save before continuing?")
	d.buttons([]string{"save", "discard", "cancel"}, []string{"Save", "Don't Save", "Cancel"}, "cancel")
	d.focus = 2
	d.result = func(id string) {
		proceed := func() {
			e.closeDialog()
			e.confirmExit = false
			done()
		}
		switch id {
		case "save":
			e.saveDocument(proceed)
		case "help":
			e.showHelp(context)
		case "discard":
			proceed()
		default:
			e.closeDialog()
			e.confirmExit = false
		}
	}
	e.openDialog(d)
}

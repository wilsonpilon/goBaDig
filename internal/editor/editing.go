package edit

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

type tabKeyMode int

const (
	tabInsert tabKeyMode = iota
	tabSpaces
	tabMove
)

type documentDeletion struct {
	start         int
	text          string
	before, after documentPosition
	revision      uint64
}

func (d *document) insertText(text string, over bool) bool {
	start, end := d.cursorSelection()
	if over && start == end {
		p := d.cursor
		graphemes := uniseg.NewGraphemes(text)
		for graphemes.Next() && p.character < len(d.lines[p.line].cells) {
			p.character++
		}
		end = d.byteOffset(p)
	}
	changed := d.replace(start, end, text)
	if changed && d.wordWrap {
		for d.autoWrapCurrentLine() {
			changed = true
		}
	}
	return changed
}

func (d *document) autoWrapCurrentLine() bool {
	line := d.lines[d.cursor.line]
	lineLimit := d.lineLength + d.leftMargin
	if d.cursor.line == 0 || len(d.lines[d.cursor.line-1].cells) == 0 {
		lineLimit = d.lineLength + d.firstMargin
	}
	lineLimit = max(1, lineLimit)
	if line.width <= lineLimit || d.column(d.cursor) <= lineLimit {
		return false
	}
	overflow := 0
	for overflow < len(line.cells) && line.cells[overflow].column+line.cells[overflow].width <= lineLimit {
		overflow++
	}
	for index := overflow - 1; index >= 0; index-- {
		cell := line.cells[index]
		raw := d.text[cell.start:cell.end]
		if !cell.tab && strings.TrimSpace(raw) != "" {
			continue
		}
		end := index + 1
		for end < len(line.cells) {
			next := d.text[line.cells[end].start:line.cells[end].end]
			if next != "\t" && strings.TrimSpace(next) != "" {
				break
			}
			end++
		}
		if end == len(line.cells) {
			return false
		}
		startOffset, endOffset := cell.start, line.cells[end-1].end
		cursorOffset := d.byteOffset(d.cursor)
		indent := strings.Repeat(" ", d.leftMargin)
		replacement := d.newline + indent
		d.replaceBytes(startOffset, endOffset, replacement)
		cursorOffset += len(replacement) - (endOffset - startOffset)
		d.setCursor(d.positionAtByte(cursorOffset), false)
		return true
	}
	return false
}

func (d *document) deleteText(start, end int) bool {
	start, end = min(start, end), max(start, end)
	if start == end {
		return false
	}
	deletion := &documentDeletion{start: start, text: d.text[start:end], before: d.cursor}
	changed := d.replace(start, end, "")
	deletion.after, deletion.revision = d.cursor, d.revision
	d.deleted = deletion
	return changed
}

func (d *document) edit(id string, over bool) (changed bool, notice string, handled bool) {
	p := d.cursor
	line := d.lines[p.line]
	start, end := d.cursorSelection()
	handled = true
	switch id {
	case "edit.undelete":
		deletion := d.deleted
		if deletion == nil || deletion.after != d.cursor || deletion.revision != d.revision {
			d.deleted = nil
			return false, "No deletion to restore; cursor movement or editing cancels undelete.", true
		}
		changed = d.replaceBytes(deletion.start, deletion.start, deletion.text)
		d.setCursor(deletion.before, false)
		return changed, "", true
	case "edit.backspace", "edit.word-left":
		if start == end {
			if p.character > 0 {
				p.character--
				if id == "edit.word-left" {
					for p.character > 0 && !wordCell(line.cells[p.character]) {
						p.character--
					}
					if wordCell(line.cells[p.character]) {
						for p.character > 0 && wordCell(line.cells[p.character-1]) {
							p.character--
						}
					}
				}
			} else if p.line > 0 {
				p.line--
				p.character = len(d.lines[p.line].cells)
			}
			start = d.byteOffset(p)
		}
	case "edit.delete", "edit.word-right":
		if start == end {
			if p.character < len(line.cells) {
				if id == "edit.word-right" {
					for p.character < len(line.cells) && !wordCell(line.cells[p.character]) {
						p.character++
					}
					for p.character < len(line.cells) && wordCell(line.cells[p.character]) {
						p.character++
					}
				} else {
					p.character++
				}
			} else if p.line+1 < len(d.lines) {
				p.line++
				p.character = 0
			}
			end = d.byteOffset(p)
		}
	case "edit.line-left":
		start, end = line.start, d.byteOffset(p)
	case "edit.line-right":
		start, end = d.byteOffset(p), line.end
	case "edit.kill-line":
		start, end = line.start, line.end
		if p.line+1 < len(d.lines) {
			end = d.lines[p.line+1].start
		} else if p.line > 0 {
			start = d.lines[p.line-1].end
		}
	case "edit.case-left", "edit.case-right":
		start, end = line.start, d.byteOffset(p)
		if id == "edit.case-right" {
			start, end = end, line.end
		}
		var text strings.Builder
		for _, cell := range line.cells {
			if cell.start < start || cell.start >= end {
				continue
			}
			raw := d.text[cell.start:cell.end]
			if utf8.ValidString(raw) {
				raw = strings.Map(func(r rune) rune {
					if unicode.IsLower(r) {
						return unicode.ToUpper(r)
					}
					if unicode.IsUpper(r) {
						return unicode.ToLower(r)
					}
					return r
				}, raw)
			}
			text.WriteString(raw)
		}
		offset := d.byteOffset(p)
		if id == "edit.case-left" {
			offset += text.Len() - (end - start)
		}
		changed = d.replace(start, end, text.String())
		d.setCursor(d.positionAtByte(offset), false)
		return changed, "", true
	case "edit.enter", "edit.indent-enter":
		indent := ""
		if id == "edit.indent-enter" || d.autoIndent {
			endIndent := line.start
			for index, cell := range line.cells {
				if index >= p.character || !cell.tab && cell.text != " " {
					break
				}
				endIndent = cell.end
			}
			indent = d.text[line.start:endIndent]
		}
		return d.insertText(d.newline+indent, false), "", true
	case "edit.tab":
		next := d.column(p) + d.tabWidth - d.column(p)%d.tabWidth
		switch d.tabMode {
		case tabSpaces:
			return d.insertText(strings.Repeat(" ", next-d.column(p)), over), "", true
		case tabMove:
			d.deleted = nil
			if next > line.width {
				return d.replace(line.end, line.end, strings.Repeat(" ", next-line.width)), "", true
			}
			d.setCursor(d.positionAtColumn(p.line, next), false)
			return false, "", true
		default:
			return d.insertText("\t", over), "", true
		}
	case "edit.unindent":
		if p.character == 0 {
			return false, "", true
		}
		for _, cell := range line.cells[:p.character] {
			if !cell.tab && cell.text != " " {
				return false, "Unindent applies only inside leading whitespace.", true
			}
		}
		target := (d.column(p) - 1) / d.tabWidth * d.tabWidth
		start, end = d.byteOffset(d.positionAtColumn(p.line, target)), d.byteOffset(p)
	default:
		return false, "", false
	}
	changed = d.deleteText(start, end)
	if changed && id == "edit.kill-line" {
		d.cursor.character = 0
		d.anchor = d.cursor
		d.goal = 0
		d.deleted.after = d.cursor
		d.reveal()
	}
	return changed, "", true
}

func literalText(event *tcell.EventKey) (string, bool) {
	if event.Key() >= tcell.KeyCtrlSpace && event.Key() <= tcell.KeyCtrlUnderscore {
		return string([]byte{byte(event.Key() - tcell.KeyCtrlSpace)}), true
	}
	if event.Key() >= tcell.KeyNUL && event.Key() <= tcell.KeyUS {
		return string([]byte{byte(event.Key())}), true
	}
	if event.Key() == tcell.KeyBackspace2 || event.Key() == tcell.KeyDelete {
		return "\x7f", true
	}
	if event.Key() != tcell.KeyRune {
		return "", false
	}
	r := event.Rune()
	if event.Modifiers()&tcell.ModCtrl != 0 {
		r = unicode.ToUpper(r)
		switch {
		case r == ' ' || r == '2':
			r = 0
		case r == '6':
			r = 30
		case r == '-':
			r = 31
		case r >= '@' && r <= '_':
			r &= 31
		default:
			return "", false
		}
	} else if event.Modifiers()&tcell.ModAlt != 0 {
		return "", false
	}
	return string(r), true
}

func (e *Editor) editorSetup() {
	d := newDialog("Editor setup", 76, 19)
	d.context = "options.editor"
	d.label(2, 2, "Insert / overstrike mode:")
	d.add(dialogWidget{kind: widgetRadio, id: "insert-mode", choices: []string{"Start in insert", "Start in overstrike", "Never overstrike"}, selected: e.startMode, x: 2, y: 3, width: 32, height: 3})
	d.label(2, 8, "Tab key:")
	d.add(dialogWidget{kind: widgetRadio, id: "tab-mode", choices: []string{"Inserts a tab", "Inserts spaces", "Moves to tab stop"}, selected: int(e.area.document.tabMode), x: 2, y: 9, width: 32, height: 3})
	d.label(38, 2, "Legacy backup character:")
	d.add(dialogWidget{kind: widgetInput, id: "backup-character", value: string(e.backupCharacter), x: 63, y: 2, width: 4})
	d.add(dialogWidget{kind: widgetCheck, id: "backups", label: "Create backups", checked: e.backupsEnabled, x: 38, y: 4, width: 27})
	d.label(38, 6, "Only the latest .{XX backup is retained.")
	d.label(38, 8, "Backup deletion policy is not configurable.")
	d.label(38, 10, "Alt-alone uses F10; terminals control isolated Alt.")
	d.buttons([]string{"ok", "help", "cancel"}, []string{"OK", "F1=Help", "Cancel"}, "ok")
	d.result = func(id string) {
		if id == "help" {
			e.showHelp(d.context)
			return
		}
		if id != "ok" {
			e.closeDialog()
			return
		}
		backupCharacter, err := parseBackupCharacter(d.value("backup-character"))
		if err != nil {
			e.showMessage("Invalid backup character", err.Error())
			return
		}
		e.closeDialog()
		before := e.currentConfiguration()
		e.startMode = d.widgets[d.fields[0]].selected
		e.backupCharacter = backupCharacter
		e.backupsEnabled = d.checked("backups")
		e.over = e.startMode == 1
		for _, window := range e.windows {
			window.startMode = e.startMode
			window.over = e.over
			window.area.over = e.over
			window.area.document.tabMode = tabKeyMode(d.widgets[d.fields[1]].selected)
		}
		e.configChanged = e.configChanged || before != e.currentConfiguration()
		e.notice = "Editor insertion, Tab and backup settings applied."
		if !e.backupsEnabled {
			e.notice += " New backups are disabled; existing backup files are kept."
		}
		e.notice += " Backup deletion policy and isolated Alt are not configurable."
		e.updateStatus()
	}
	e.openDialog(d)
}

func parseBackupCharacter(value string) (rune, error) {
	runes := []rune(value)
	if len(runes) != 1 || runes[0] < '!' || runes[0] > '~' || strings.ContainsRune("\\/:*?\"<>|.", runes[0]) {
		return 0, fmt.Errorf("Backup character must be one printable ASCII character allowed in a file name.")
	}
	return runes[0], nil
}

func (e *Editor) backupCharacterForSave() rune {
	if !e.backupsEnabled {
		return 0
	}
	return e.backupCharacter
}

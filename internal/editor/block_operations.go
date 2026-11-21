package edit

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	maxBlockIndentSpaces = 10000
	maxBlockIndentGrowth = 16 << 20
)

func (d *document) markedBlock() (int, int, bool) {
	if d.blockMarkerCount != len(d.blockMarkers) {
		return 0, 0, false
	}
	return min(d.blockMarkers[0], d.blockMarkers[1]),
		max(d.blockMarkers[0], d.blockMarkers[1]), true
}

func (d *document) blockLineRange(start, end int) (int, int, bool) {
	if start >= end {
		return 0, 0, false
	}
	first := d.positionAtByte(start).line
	last := d.positionAtByte(end - 1).line
	return first, last, true
}

func (d *document) indentBlock(amount int) bool {
	start, end, ok := d.markedBlock()
	if !ok || amount <= 0 {
		return false
	}
	first, last, ok := d.blockLineRange(start, end)
	if !ok {
		return false
	}
	spaces := strings.Repeat(" ", amount)
	replacements := make([]documentReplacement, 0, last-first+1)
	for line := first; line <= last; line++ {
		offset := d.lines[line].start
		replacements = append(replacements, documentReplacement{
			start: offset, end: offset, text: spaces,
		})
	}
	return d.replaceMany(replacements)
}

func (d *document) unindentBlock(amount int) bool {
	start, end, ok := d.markedBlock()
	if !ok || amount <= 0 {
		return false
	}
	first, last, ok := d.blockLineRange(start, end)
	if !ok {
		return false
	}
	replacements := make([]documentReplacement, 0, last-first+1)
	for lineIndex := first; lineIndex <= last; lineIndex++ {
		line := d.lines[lineIndex]
		indentEnd, indentWidth := line.start, 0
		for _, cell := range line.cells {
			if !cell.tab && cell.text != " " {
				break
			}
			indentEnd = cell.end
			indentWidth += cell.width
		}
		if indentWidth == 0 {
			continue
		}

		targetWidth := max(0, indentWidth-amount)
		column := 0
		var candidate strings.Builder
		for _, cell := range line.cells {
			if !cell.tab && cell.text != " " {
				break
			}
			cellEnd := column + cell.width
			switch {
			case cellEnd <= amount:
			case column < amount:
				candidate.WriteString(strings.Repeat(" ", cellEnd-amount))
			default:
				candidate.WriteString(d.text[cell.start:cell.end])
			}
			column = cellEnd
		}
		candidateWidth := 0
		for _, character := range candidate.String() {
			if character == '\t' {
				candidateWidth += d.tabWidth - candidateWidth%d.tabWidth
			} else {
				candidateWidth++
			}
		}
		replacement := candidate.String()
		if candidateWidth != targetWidth {
			replacement = strings.Repeat(" ", targetWidth)
		}
		if replacement != d.text[line.start:indentEnd] {
			replacements = append(replacements, documentReplacement{
				start: line.start, end: indentEnd, text: replacement,
			})
		}
	}
	return d.replaceMany(replacements)
}

func (e *Editor) blockRangeOrNotice(command string) (int, int, bool) {
	start, end, ok := e.area.document.markedBlock()
	if !ok {
		e.notice = "Set two block markers before using " + command + "."
		e.updateStatus()
		return 0, 0, false
	}
	return start, end, true
}

func (e *Editor) copyBlock() {
	start, end, ok := e.blockRangeOrNotice("Copy")
	if !ok {
		return
	}
	if start == end {
		e.notice = "The marked block is empty."
		e.updateStatus()
		return
	}
	d := e.area.document
	destination := d.byteOffset(d.cursor)
	if destination > start && destination < end {
		e.notice = "Copy destination is inside the marked block."
		e.updateStatus()
		return
	}
	text := d.text[start:end]
	changed := d.replaceBytes(destination, destination, text)
	if destination == end {
		d.setBlockRange(start, end)
	}
	e.area.notify(changed)
	e.notice = fmt.Sprintf("Copied %d bytes at the cursor.", len(text))
	e.updateStatus()
}

func (e *Editor) moveBlock() {
	start, end, ok := e.blockRangeOrNotice("Move")
	if !ok {
		return
	}
	if start == end {
		e.notice = "The marked block is empty."
		e.updateStatus()
		return
	}
	d := e.area.document
	destination := d.byteOffset(d.cursor)
	if destination >= start && destination <= end {
		e.notice = "Move destination is inside or at the edge of the marked block."
		e.updateStatus()
		return
	}
	if e.path != "" && e.diskKnown && d.sourcePositionAt(len(d.text)) != len(e.diskText) {
		e.notice = "Load the entire file before moving a block."
		e.updateStatus()
		return
	}

	text := d.text
	block := text[start:end]
	newText := ""
	newDestination := destination
	if destination < start {
		newText = text[:destination] + block + text[destination:start] + text[end:]
	} else {
		newDestination -= len(block)
		newText = text[:start] + text[end:destination] + block + text[destination:]
	}
	if newText == text {
		e.notice = "Moving the block would not change the document."
		e.updateStatus()
		return
	}

	d.replaceBytes(start, end, "")
	d.replaceBytes(newDestination, newDestination, block)
	d.setSourceRuns([]documentSourceRun{{
		textStart: 0, textEnd: len(d.text),
		sourceStart: e.fileBase, sourceEnd: len(e.diskText),
	}})
	d.sourceReordered = true
	d.setBlockRange(newDestination, newDestination+len(block))
	d.setCursor(d.positionAtByte(newDestination+len(block)), false)
	e.area.notify(true)
	e.notice = fmt.Sprintf("Moved %d bytes to the cursor.", len(block))
	e.updateStatus()
}

func (e *Editor) deleteBlock(confirmed bool) {
	start, end, ok := e.blockRangeOrNotice("Delete")
	if !ok {
		return
	}
	if start == end {
		e.notice = "The marked block is empty."
		e.updateStatus()
		return
	}
	if !confirmed {
		d := newDialog("Delete marked block?", 62, 8)
		d.context = "block.delete"
		d.label(2, 2, "Remove the marked text from this document?")
		d.buttons([]string{"delete", "cancel"}, []string{"Delete", "Cancel"}, "cancel")
		d.focus = 1
		d.result = func(id string) {
			if id == "help" {
				e.showHelp(d.context)
				return
			}
			e.closeDialog()
			if id == "delete" {
				e.deleteMarkedRange(start, end)
			}
		}
		e.openDialog(d)
		return
	}
	e.deleteMarkedRange(start, end)
}

func (e *Editor) deleteMarkedRange(start, end int) {
	d := e.area.document
	changed := d.replaceBytes(start, end, "")
	d.clearBlockMarkers()
	e.area.notify(changed)
	e.notice = "Marked block deleted."
	e.updateStatus()
}

func (e *Editor) blockIndent(command string) {
	start, end, ok := e.blockRangeOrNotice(command)
	if !ok {
		return
	}
	if start == end {
		e.notice = "The marked block is empty."
		e.updateStatus()
		return
	}
	commandSpec, _ := commandByID(command)
	d := newDialog(commandSpec.label, 62, 8)
	d.context = command
	d.input("amount", "Spaces/columns:", "1", 2, 0, maxBlockIndentSpaces)
	d.buttons([]string{"ok", "help", "cancel"}, []string{"OK", "F1=Help", "Cancel"}, "ok")
	d.result = func(id string) {
		if id == "help" {
			e.showHelp(d.context)
			return
		}
		e.closeDialog()
		if id != "ok" {
			return
		}
		amount, err := strconv.Atoi(d.value("amount"))
		if err != nil {
			e.fileError(fmt.Errorf("invalid indentation amount: %w", err), command)
			return
		}
		if amount == 0 {
			e.notice = "Indentation amount is zero; the document was not changed."
			e.updateStatus()
			return
		}
		first, last, ok := e.area.document.blockLineRange(start, end)
		if !ok {
			e.notice = "The marked block is empty."
			e.updateStatus()
			return
		}
		lines := last - first + 1
		if command == "block.indent" && amount > maxBlockIndentGrowth/lines {
			e.fileError(fmt.Errorf("indentation would add more than %d bytes", maxBlockIndentGrowth), command)
			return
		}
		changed := false
		if command == "block.indent" {
			changed = e.area.document.indentBlock(amount)
		} else {
			changed = e.area.document.unindentBlock(amount)
		}
		e.area.notify(changed)
		if changed {
			e.notice = fmt.Sprintf("Applied %d-space %s to the marked lines.", amount, strings.TrimPrefix(command, "block."))
		} else {
			e.notice = "The marked lines do not need this indentation change."
		}
		e.updateStatus()
	}
	e.openDialog(d)
}

func (e *Editor) outputBlock() {
	if _, _, ok := e.blockRangeOrNotice("Output to file"); !ok {
		return
	}
	e.fileDialog("block.output", nil)
}

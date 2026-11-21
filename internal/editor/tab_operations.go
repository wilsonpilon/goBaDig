package edit

import "strings"

func mappedOffset(offset int, replacements []documentReplacement) int {
	shift := 0
	for _, replacement := range replacements {
		delta := len(replacement.text) - (replacement.end - replacement.start)
		if replacement.start == replacement.end {
			if offset >= replacement.start {
				shift += len(replacement.text)
			}
			continue
		}
		if offset <= replacement.start {
			break
		}
		if offset < replacement.end {
			return replacement.start + shift + len(replacement.text)
		}
		shift += delta
	}
	return offset + shift
}

func (d *document) applyLayoutReplacements(replacements []documentReplacement) bool {
	if len(replacements) == 0 {
		return false
	}
	cursorOffset, anchorOffset := d.byteOffset(d.cursor), d.byteOffset(d.anchor)
	cursorTarget := mappedOffset(cursorOffset, replacements)
	anchorTarget := mappedOffset(anchorOffset, replacements)
	if !d.replaceMany(replacements) {
		return false
	}
	d.cursor, d.anchor = d.positionAtByte(cursorTarget), d.positionAtByte(anchorTarget)
	d.goal = d.column(d.cursor)
	d.reveal()
	return true
}

func (d *document) expandTabs(fromCursor bool) bool {
	tabWidth := max(1, d.tabWidth)
	firstLine := 0
	cursorOffset := 0
	if fromCursor {
		firstLine = d.cursor.line
		cursorOffset = d.byteOffset(d.cursor)
	}
	var replacements []documentReplacement
	for lineIndex := firstLine; lineIndex < len(d.lines); lineIndex++ {
		line := d.lines[lineIndex]
		for _, cell := range line.cells {
			if !cell.tab || fromCursor && lineIndex == firstLine && cell.start < cursorOffset {
				continue
			}
			spaces := tabWidth - cell.column%tabWidth
			replacements = append(replacements, documentReplacement{
				start: cell.start, end: cell.end, text: strings.Repeat(" ", spaces),
			})
		}
	}
	return d.applyLayoutReplacements(replacements)
}

func (d *document) putInTabs(fromCursor bool) bool {
	tabWidth := max(1, d.tabWidth)
	firstLine := 0
	if fromCursor {
		firstLine = d.cursor.line
	}
	var replacements []documentReplacement
	for lineIndex := firstLine; lineIndex < len(d.lines); lineIndex++ {
		line := d.lines[lineIndex]
		leadingCells, columns := 0, 0
		for leadingCells < len(line.cells) {
			cell := line.cells[leadingCells]
			raw := d.text[cell.start:cell.end]
			switch raw {
			case " ":
				columns++
			case "\t":
				columns += tabWidth - columns%tabWidth
			default:
				goto leadingWhitespaceEnd
			}
			leadingCells++
		}
	leadingWhitespaceEnd:
		if leadingCells == 0 || fromCursor && lineIndex == firstLine && d.cursor.character > leadingCells {
			continue
		}
		start, end := line.start, line.cells[leadingCells-1].end
		converted := strings.Repeat("\t", columns/tabWidth) + strings.Repeat(" ", columns%tabWidth)
		if d.text[start:end] != converted {
			replacements = append(replacements, documentReplacement{start: start, end: end, text: converted})
		}
	}
	return d.applyLayoutReplacements(replacements)
}

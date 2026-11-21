package edit

import "strings"

func (d *document) insertPageBreaks(fromCursor bool) bool {
	firstLine := 0
	if fromCursor {
		firstLine = d.cursor.line
	}
	lineCount := len(d.lines)
	if strings.HasSuffix(d.text, "\n") {
		lineCount--
	}
	pageLength := max(1, d.pageLength)
	var replacements []documentReplacement
	for boundary := firstLine + pageLength; boundary < lineCount; {
		line := d.lines[boundary]
		if !strings.HasPrefix(d.text[line.start:line.end], "\f") {
			replacements = append(replacements, documentReplacement{
				start: line.start, end: line.start, text: "\f",
			})
		}
		if pageLength > lineCount-boundary {
			break
		}
		boundary += pageLength
	}
	return d.applyLayoutReplacements(replacements)
}

func (d *document) removePageBreaks(fromCursor bool) bool {
	firstLine := 0
	if fromCursor {
		firstLine = d.cursor.line
	}
	var replacements []documentReplacement
	for _, line := range d.lines[firstLine:] {
		for _, cell := range line.cells {
			if d.text[cell.start:cell.end] == "\f" {
				replacements = append(replacements, documentReplacement{
					start: cell.start, end: cell.end,
				})
			}
		}
	}
	return d.applyLayoutReplacements(replacements)
}

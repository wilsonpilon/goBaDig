package edit

import (
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
)

func (d *document) paragraphBounds(line int) (int, int, bool) {
	if line < 0 || line >= len(d.lines) || strings.TrimSpace(d.text[d.lines[line].start:d.lines[line].end]) == "" {
		return 0, 0, false
	}
	start, end := line, line+1
	for start > 0 && strings.TrimSpace(d.text[d.lines[start-1].start:d.lines[start-1].end]) != "" {
		start--
	}
	for end < len(d.lines) && strings.TrimSpace(d.text[d.lines[end].start:d.lines[end].end]) != "" {
		end++
	}
	return start, end, true
}

func (d *document) formatStartAtCursor() (int, int) {
	position := d.cursor
	line := d.lines[position.line]
	for position.character > 0 {
		cell := line.cells[position.character-1]
		raw := d.text[cell.start:cell.end]
		if raw != "\t" && strings.TrimSpace(raw) != "" {
			break
		}
		position.character--
	}
	return d.byteOffset(position), d.column(position)
}

func visualWidth(text string, column, tabWidth int) int {
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		if graphemes.Str() == "\t" {
			column += tabWidth - column%tabWidth
		} else {
			column += max(1, graphemes.Width())
		}
	}
	return column
}

func (d *document) reflowWords(text string, initialColumn, firstIndent, firstLimit int, hasPrefix bool) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var output strings.Builder
	column := initialColumn
	if firstIndent > 0 {
		output.WriteString(strings.Repeat(" ", firstIndent))
	}
	hasWord := hasPrefix
	firstLine := true
	for _, word := range words {
		wordWidth := visualWidth(word, 0, d.tabWidth)
		space := 0
		if hasWord {
			space = 1
		}
		limit := d.lineLength + d.leftMargin
		if firstLine {
			limit = firstLimit
		}
		if hasWord && column+space+wordWidth > limit {
			output.WriteString(d.newline)
			if d.leftMargin > 0 {
				output.WriteString(strings.Repeat(" ", d.leftMargin))
			}
			column = d.leftMargin
			hasWord = false
			space = 0
			firstLine = false
		}
		if space > 0 {
			output.WriteByte(' ')
			column++
		}
		output.WriteString(word)
		column += wordWidth
		hasWord = true
	}
	return output.String()
}

func (d *document) formatParagraph() bool {
	startLine, endLine, ok := d.paragraphBounds(d.cursor.line)
	if !ok {
		return false
	}
	start := d.lines[startLine].start
	firstIndent := d.firstMargin
	initialColumn, firstLimit := firstIndent, d.lineLength+firstIndent
	hasPrefix := false
	if d.formatAtCursor {
		start, initialColumn = d.formatStartAtCursor()
		firstIndent = 0
		firstLimit = d.lineLength + d.leftMargin
		hasPrefix = initialColumn > 0
	}
	end := d.lines[endLine-1].end
	formatted := d.reflowWords(d.text[start:end], initialColumn, firstIndent, firstLimit, hasPrefix)
	if formatted == d.text[start:end] || !d.replaceBytes(start, end, formatted) {
		return false
	}
	d.setCursor(d.positionAtByte(start), false)
	return true
}

func (d *document) formatDocument(fromCursor bool) bool {
	firstStart := -1
	var replacements []documentReplacement
	for line := 0; line < len(d.lines); {
		startLine, endLine, ok := d.paragraphBounds(line)
		if !ok {
			line++
			continue
		}
		start, end := d.lines[startLine].start, d.lines[endLine-1].end
		if fromCursor && d.cursor.line >= endLine {
			line = endLine
			continue
		}
		partial := fromCursor && d.cursor.line >= startLine && d.cursor.line < endLine
		firstIndent := d.firstMargin
		initialColumn, firstLimit := firstIndent, d.lineLength+firstIndent
		hasPrefix := false
		if partial {
			start, initialColumn = d.formatStartAtCursor()
			firstIndent = 0
			firstLimit = d.lineLength + d.leftMargin
			hasPrefix = initialColumn > 0
		}
		if start < end {
			raw := d.text[start:end]
			formatted := d.reflowWords(raw, initialColumn, firstIndent, firstLimit, hasPrefix)
			if formatted != raw {
				replacements = append(replacements, documentReplacement{start: start, end: end, text: formatted})
				if firstStart < 0 {
					firstStart = start
				}
			}
		}
		line = endLine
	}
	if len(replacements) == 0 || !d.replaceMany(replacements) {
		return false
	}
	d.setCursor(d.positionAtByte(firstStart), false)
	return true
}

func (d *document) centerLine() bool {
	line := d.lines[d.cursor.line]
	raw := d.text[line.start:line.end]
	trimmed := strings.TrimLeftFunc(raw, unicode.IsSpace)
	leadingBytes := len(raw) - len(trimmed)
	trimmed = strings.TrimRightFunc(trimmed, unicode.IsSpace)
	if trimmed == "" {
		return false
	}
	margin := d.leftMargin
	if lineIndex, _, ok := d.paragraphBounds(d.cursor.line); ok && lineIndex == d.cursor.line {
		margin += d.firstMargin
	}
	padding := margin + max(0, (d.lineLength-(visualWidth(trimmed, margin, d.tabWidth)-margin))/2)
	for range d.tabWidth {
		textWidth := visualWidth(trimmed, padding, d.tabWidth) - padding
		next := margin + max(0, (d.lineLength-textWidth)/2)
		if next == padding {
			break
		}
		padding = next
	}
	centered := strings.Repeat(" ", padding) + trimmed
	if centered == raw {
		return false
	}
	cursorOffset := d.byteOffset(d.cursor)
	relative := max(0, min(len(trimmed), cursorOffset-line.start-leadingBytes))
	if !d.replaceBytes(line.start, line.end, centered) {
		return false
	}
	d.setCursor(d.positionAtByte(line.start+padding+relative), false)
	return true
}

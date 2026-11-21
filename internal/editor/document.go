package edit

// Navigation semantics adapted from the user's fwMSX editor (BSD-3-Clause).
// See docs/FWMSX-EDITOR-NOTICE.txt. This model has no terminal dependency.

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

type documentPosition struct {
	line, character int
}

type documentCell struct {
	start, end, column, width int
	text                      string
	tab, control, breakAfter  bool
}

type documentLine struct {
	start, end, width int
	cells             []documentCell
}

type documentRow struct {
	line, from, to int
}

type documentSourceRun struct {
	textStart, textEnd     int
	sourceStart, sourceEnd int
	linear                 bool
}

type document struct {
	text             string
	lines            []documentLine
	rows             []documentRow
	sourceRuns       []documentSourceRun
	sourceReordered  bool
	cursor, anchor   documentPosition
	goal, top, left  int
	width, height    int
	tabWidth         int
	pageLength       int
	printTopMargin   int
	printLeftMargin  int
	formfeedBetween  bool
	formfeedAfter    int
	lineLength       int
	firstMargin      int
	leftMargin       int
	formatAtCursor   bool
	wrap             bool
	wordWrap         bool
	autoIndent       bool
	revision         uint64
	newline          string
	tabMode          tabKeyMode
	deleted          *documentDeletion
	blockMarkers     [2]int
	blockMarkerCount int
	nextBlockMarker  int
	searchMatchStart int
	searchMatchEnd   int
	searchMatch      bool
}

func newDocument(text string) *document {
	d := &document{
		width: 79, height: 23, tabWidth: 8, pageLength: 62, printTopMargin: 2,
		formfeedBetween: true, formfeedAfter: 1, lineLength: 65,
	}
	d.load(text)
	return d
}

func (d *document) load(text string) {
	d.text = text
	d.newline = "\r\n"
	if index := strings.IndexByte(text, '\n'); index >= 0 && (index == 0 || text[index-1] != '\r') {
		d.newline = "\n"
	}
	d.deleted = nil
	d.cursor, d.anchor = documentPosition{}, documentPosition{}
	d.blockMarkers, d.blockMarkerCount, d.nextBlockMarker = [2]int{}, 0, 0
	d.searchMatchStart, d.searchMatchEnd, d.searchMatch = 0, 0, false
	d.goal, d.top, d.left = 0, 0, 0
	d.setSourceBase(0)
	d.indexLines()
}

func (d *document) indexLines() {
	d.lines = nil
	for start := 0; ; {
		end := strings.IndexByte(d.text[start:], '\n')
		next := len(d.text) + 1
		if end < 0 {
			end = len(d.text)
		} else {
			end += start
			next = end + 1
			if end > start && d.text[end-1] == '\r' {
				end--
			}
		}
		line := documentLine{start: start, end: end}
		graphemes := uniseg.NewGraphemes(d.text[start:end])
		for graphemes.Next() {
			from, to := graphemes.Positions()
			text := graphemes.Str()
			tab := text == "\t"
			width := max(1, graphemes.Width())
			r, _ := utf8.DecodeRuneInString(text)
			control := !tab && unicode.IsControl(r)
			if tab {
				width = d.tabWidth - line.width%d.tabWidth
			} else {
				switch {
				case r < 32:
					text = string('\u2400' + r)
				case r == 127:
					text = "\u2421"
				case control:
					text = "\ufffd"
				case graphemes.Width() == 0:
					text = "\u25cc" + text
				}
			}
			line.cells = append(line.cells, documentCell{
				start: start + from, end: start + to, column: line.width, width: width,
				text: text, tab: tab, control: control, breakAfter: graphemes.LineBreak() != uniseg.LineDontBreak,
			})
			line.width += width
		}
		d.lines = append(d.lines, line)
		if next > len(d.text) {
			break
		}
		start = next
	}
	d.layout()
}

func (d *document) layout() {
	d.rows = nil
	for index, line := range d.lines {
		if !d.wrap || line.width == 0 {
			d.rows = append(d.rows, documentRow{index, 0, line.width})
			continue
		}
		from := 0
		for from < line.width {
			to := min(from+d.width, line.width)
			if to < line.width {
				lastBreak := from
				first := sort.Search(len(line.cells), func(i int) bool {
					return line.cells[i].column+line.cells[i].width > from
				})
				for _, cell := range line.cells[first:] {
					end := cell.column + cell.width
					if cell.column >= to {
						break
					}
					if end > to && !cell.tab && cell.column > from {
						to = cell.column
						break
					}
					if cell.breakAfter && end <= to {
						lastBreak = end
					}
				}
				if lastBreak > from {
					to = lastBreak
				}
			}
			d.rows = append(d.rows, documentRow{index, from, to})
			from = to
		}
		last := d.rows[len(d.rows)-1]
		if last.to-last.from >= d.width {
			d.rows = append(d.rows, documentRow{index, line.width, line.width})
		}
	}
	d.top = max(0, min(d.top, len(d.rows)-1))
	if d.wrap {
		d.left = 0
	}
}

func (d *document) clamp(p documentPosition) documentPosition {
	p.line = max(0, min(p.line, len(d.lines)-1))
	p.character = max(0, min(p.character, len(d.lines[p.line].cells)))
	return p
}

func (d *document) byteOffset(p documentPosition) int {
	p = d.clamp(p)
	line := d.lines[p.line]
	if p.character == len(line.cells) {
		return line.end
	}
	return line.cells[p.character].start
}

func (d *document) positionAtByte(offset int) documentPosition {
	offset = max(0, min(offset, len(d.text)))
	line := max(0, sort.Search(len(d.lines), func(i int) bool { return d.lines[i].start > offset })-1)
	character := sort.Search(len(d.lines[line].cells), func(i int) bool { return d.lines[line].cells[i].end > offset })
	return documentPosition{line, character}
}

func (d *document) column(p documentPosition) int {
	p = d.clamp(p)
	line := d.lines[p.line]
	if p.character == len(line.cells) {
		return line.width
	}
	return line.cells[p.character].column
}

func (d *document) positionAtColumn(line, column int) documentPosition {
	line = max(0, min(line, len(d.lines)-1))
	cells := d.lines[line].cells
	character := sort.Search(len(cells), func(i int) bool { return cells[i].column+cells[i].width > column })
	return documentPosition{line, character}
}

func (d *document) rowAt(p documentPosition) int {
	p = d.clamp(p)
	column := d.column(p)
	row := sort.Search(len(d.rows), func(i int) bool {
		return d.rows[i].line > p.line || (d.rows[i].line == p.line && d.rows[i].from > column)
	}) - 1
	return max(0, row)
}

func (d *document) reveal() {
	row, column := d.rowAt(d.cursor), d.column(d.cursor)
	cursorWidth := 1
	if d.cursor.character < len(d.lines[d.cursor.line].cells) {
		cell := d.lines[d.cursor.line].cells[d.cursor.character]
		if !cell.tab {
			cursorWidth = min(cell.width, d.width)
		}
	}
	if row < d.top {
		d.top = row
	} else if row >= d.top+d.height {
		d.top = row - d.height + 1
	}
	if d.wrap {
		d.left = 0
	} else if column < d.left {
		d.left = column
	} else if column+cursorWidth > d.left+d.width {
		d.left = column + cursorWidth - d.width
	}
}

func (d *document) resize(width, height int, wrap bool) {
	width, height = max(1, width), max(1, height)
	if d.width == width && d.height == height && d.wrap == wrap {
		return
	}
	d.width, d.height, d.wrap = width, height, wrap
	d.layout()
	d.reveal()
}

func (d *document) setCursor(p documentPosition, keepGoal bool) {
	p = d.clamp(p)
	if p != d.cursor {
		d.deleted = nil
		d.searchMatch = false
	}
	d.cursor = p
	d.anchor = d.cursor
	if !keepGoal {
		d.goal = d.column(d.cursor)
	}
	d.reveal()
}

func (d *document) setSearchMatch(start, end int) {
	d.searchMatchStart = max(0, min(start, len(d.text)))
	d.searchMatchEnd = max(d.searchMatchStart, min(end, len(d.text)))
	d.searchMatch = d.searchMatchStart < d.searchMatchEnd
}

func (d *document) positionOnRow(row, column int) documentPosition {
	row = max(0, min(row, len(d.rows)-1))
	target := d.rows[row]
	if d.wrap {
		column = min(target.to, target.from+column)
	}
	return d.positionAtColumn(target.line, column)
}

func wordCell(cell documentCell) bool {
	if cell.control || cell.tab {
		return false
	}
	r, _ := utf8.DecodeRuneInString(cell.text)
	return r == '_' || r >= 128 || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z'
}

func (d *document) navigate(id string) bool {
	d.deleted = nil
	p := d.cursor
	line := d.lines[p.line]
	keepGoal := false
	screenGoal := d.goal
	if d.wrap {
		screenGoal = max(0, screenGoal-d.rows[d.rowAt(p)].from)
	}
	switch id {
	case "cursor.left":
		if p.character > 0 {
			p.character--
		} else if p.line > 0 {
			p.line--
			p.character = len(d.lines[p.line].cells)
		}
	case "cursor.right":
		if p.character < len(line.cells) {
			p.character++
		} else if p.line+1 < len(d.lines) {
			p.line++
			p.character = 0
		}
	case "cursor.word-left":
		if p.character == 0 && p.line > 0 {
			p.line--
			p.character = len(d.lines[p.line].cells)
		} else {
			for p.character > 0 && !wordCell(line.cells[p.character-1]) {
				p.character--
			}
			for p.character > 0 && wordCell(line.cells[p.character-1]) {
				p.character--
			}
		}
	case "cursor.word-right":
		if p.character == len(line.cells) && p.line+1 < len(d.lines) {
			p.line++
			p.character = 0
		} else {
			for p.character < len(line.cells) && wordCell(line.cells[p.character]) {
				p.character++
			}
			for p.character < len(line.cells) && !wordCell(line.cells[p.character]) {
				p.character++
			}
		}
	case "cursor.up", "cursor.down":
		delta := 1
		if id == "cursor.up" {
			delta = -1
		}
		p = d.positionAtColumn(p.line+delta, d.goal)
		keepGoal = true
	case "cursor.home":
		p.character = 0
	case "cursor.end":
		p.character = len(line.cells)
	case "cursor.file-top":
		p = documentPosition{}
	case "cursor.file-end":
		p.line = len(d.lines) - 1
		p.character = len(d.lines[p.line].cells)
	case "cursor.page-up", "cursor.page-down":
		delta := max(1, d.height-1)
		if id == "cursor.page-up" {
			delta = -delta
		}
		d.top = max(0, min(d.top+delta, len(d.rows)-1))
		p = d.positionOnRow(d.rowAt(p)+delta, screenGoal)
		keepGoal = true
	case "cursor.screen-top", "cursor.screen-bottom":
		row := d.top
		if id == "cursor.screen-bottom" {
			row += d.height - 1
		}
		p = d.positionOnRow(row, screenGoal)
		keepGoal = true
	case "cursor.scroll-up", "cursor.scroll-down":
		delta := 1
		if id == "cursor.scroll-down" {
			delta = -1
		}
		d.top = max(0, min(d.top+delta, len(d.rows)-1))
		row := d.rowAt(p)
		target := max(d.top, min(row, d.top+d.height-1))
		if target != row {
			d.cursor = d.positionOnRow(target, screenGoal)
		}
		d.anchor = d.cursor
		return true
	default:
		return false
	}
	d.setCursor(p, keepGoal)
	return true
}

func (d *document) selection() (int, int) {
	if d.blockMarkerCount == len(d.blockMarkers) {
		return min(d.blockMarkers[0], d.blockMarkers[1]), max(d.blockMarkers[0], d.blockMarkers[1])
	}
	return d.cursorSelection()
}

func (d *document) markerGutterWidth() int {
	return d.blockMarkerCount
}

func (d *document) cursorSelection() (int, int) {
	a, b := d.byteOffset(d.anchor), d.byteOffset(d.cursor)
	return min(a, b), max(a, b)
}

func (d *document) selectBytes(start, end int) {
	d.clearBlockMarkers()
	d.deleted = nil
	d.anchor, d.cursor = d.positionAtByte(start), d.positionAtByte(end)
	d.goal = d.column(d.cursor)
	d.reveal()
}

func (d *document) setSourceBase(base int) {
	d.sourceRuns = []documentSourceRun{{
		textStart: 0, textEnd: len(d.text),
		sourceStart: base, sourceEnd: base + len(d.text), linear: true,
	}}
	d.sourceReordered = false
}

func (d *document) setSourceRuns(runs []documentSourceRun) {
	d.sourceRuns = append([]documentSourceRun(nil), runs...)
	if len(d.sourceRuns) == 0 {
		d.setSourceBase(0)
	}
}

func (d *document) sourcePositionAt(offset int) int {
	offset = max(0, min(offset, len(d.text)))
	d.splitSourceRunAt(offset)
	position, found := 0, false
	for _, run := range d.sourceRuns {
		switch {
		case run.textStart == offset && run.textEnd == offset:
			position, found = run.sourceEnd, true
		case run.textStart == offset && run.textEnd > offset:
			return run.sourceStart
		case run.textEnd == offset && run.textStart < offset:
			position, found = run.sourceEnd, true
		}
	}
	if found {
		return position
	}
	return 0
}

func (d *document) splitSourceRunAt(offset int) {
	for index, run := range d.sourceRuns {
		if offset <= run.textStart || offset >= run.textEnd {
			continue
		}
		left, right := run, run
		left.textEnd, right.textStart = offset, offset
		if run.linear {
			sourceOffset := run.sourceStart + offset - run.textStart
			left.sourceEnd, right.sourceStart = sourceOffset, sourceOffset
		} else {
			right.sourceStart, right.sourceEnd = run.sourceEnd, run.sourceEnd
			right.linear = false
		}
		d.sourceRuns = append(d.sourceRuns[:index], append([]documentSourceRun{left, right}, d.sourceRuns[index+1:]...)...)
		return
	}
}

func (d *document) replaceSourceRange(start, end, inserted, sourceStart, sourceEnd int) {
	d.splitSourceRunAt(start)
	d.splitSourceRunAt(end)
	prefix, suffix := make([]documentSourceRun, 0, len(d.sourceRuns)), make([]documentSourceRun, 0, len(d.sourceRuns))
	for _, run := range d.sourceRuns {
		if run.textEnd <= start {
			prefix = append(prefix, run)
		} else if run.textStart >= end {
			suffix = append(suffix, run)
		}
	}
	delta := inserted - (end - start)
	for index := range suffix {
		suffix[index].textStart += delta
		suffix[index].textEnd += delta
	}
	if inserted > 0 {
		prefix = append(prefix, documentSourceRun{
			textStart: start, textEnd: start + inserted,
			sourceStart: sourceStart, sourceEnd: sourceEnd,
		})
	} else if sourceEnd > sourceStart {
		prefix = append(prefix, documentSourceRun{
			textStart: start, textEnd: start,
			sourceStart: sourceStart, sourceEnd: sourceEnd,
		})
	}
	d.sourceRuns = append(prefix, suffix...)
}

func (d *document) appendSourceText(text string, sourceStart int) {
	if text == "" {
		return
	}
	start := len(d.text)
	d.text += text
	d.sourceRuns = append(d.sourceRuns, documentSourceRun{
		textStart: start, textEnd: len(d.text),
		sourceStart: sourceStart, sourceEnd: sourceStart + len(text), linear: true,
	})
	d.indexLines()
	d.reveal()
}

type documentReplacement struct {
	start, end             int
	text                   string
	sourceStart, sourceEnd int
}

func (d *document) replaceMany(replacements []documentReplacement) bool {
	if len(replacements) == 0 {
		return false
	}
	lastEnd := 0
	for index := range replacements {
		replacement := &replacements[index]
		if replacement.start < lastEnd || replacement.start < 0 ||
			replacement.end < replacement.start || replacement.end > len(d.text) {
			return false
		}
		replacement.sourceStart = d.sourcePositionAt(replacement.start)
		replacement.sourceEnd = d.sourcePositionAt(replacement.end)
		lastEnd = replacement.end
	}
	for _, replacement := range replacements {
		d.splitSourceRunAt(replacement.start)
		d.splitSourceRunAt(replacement.end)
	}

	oldText := d.text
	oldRuns := append([]documentSourceRun(nil), d.sourceRuns...)
	var output strings.Builder
	netChange := 0
	for _, replacement := range replacements {
		netChange += len(replacement.text) - (replacement.end - replacement.start)
	}
	output.Grow(max(0, len(oldText)+netChange))
	var runs []documentSourceRun
	appendUnchanged := func(start, end, outputStart int, includeEnd bool) {
		for _, run := range oldRuns {
			if run.textStart == run.textEnd {
				if run.textStart < start || run.textStart > end ||
					!includeEnd && run.textStart == end {
					continue
				}
				run.textStart = outputStart + run.textStart - start
				run.textEnd = run.textStart
				runs = append(runs, run)
				continue
			}
			if run.textStart < start || run.textEnd > end {
				continue
			}
			run.textStart = outputStart + run.textStart - start
			run.textEnd = outputStart + run.textEnd - start
			runs = append(runs, run)
		}
	}
	transformOffset := func(offset int) int {
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

	oldOffset, outputOffset := 0, 0
	for _, replacement := range replacements {
		output.WriteString(oldText[oldOffset:replacement.start])
		appendUnchanged(oldOffset, replacement.start, outputOffset, false)
		outputOffset += replacement.start - oldOffset
		if replacement.text != "" {
			runs = append(runs, documentSourceRun{
				textStart: outputOffset, textEnd: outputOffset + len(replacement.text),
				sourceStart: replacement.sourceStart, sourceEnd: replacement.sourceEnd,
			})
			output.WriteString(replacement.text)
			outputOffset += len(replacement.text)
		} else if replacement.sourceEnd > replacement.sourceStart {
			runs = append(runs, documentSourceRun{
				textStart: outputOffset, textEnd: outputOffset,
				sourceStart: replacement.sourceStart, sourceEnd: replacement.sourceEnd,
			})
		}
		oldOffset = replacement.end
	}
	output.WriteString(oldText[oldOffset:])
	appendUnchanged(oldOffset, len(oldText), outputOffset, true)

	cursorOffset, anchorOffset := transformOffset(d.byteOffset(d.cursor)), transformOffset(d.byteOffset(d.anchor))
	for index := 0; index < d.blockMarkerCount; index++ {
		d.blockMarkers[index] = transformOffset(d.blockMarkers[index])
	}
	d.text = output.String()
	d.sourceRuns = runs
	if len(d.sourceRuns) == 0 {
		d.setSourceBase(0)
	}
	d.deleted = nil
	d.indexLines()
	d.cursor, d.anchor = d.positionAtByte(cursorOffset), d.positionAtByte(anchorOffset)
	d.goal = d.column(d.cursor)
	d.normalizeBlockMarkers()
	d.reveal()
	d.revision++
	return true
}

func (d *document) sourceRunsAfter(offset, sourcePosition, newBase int) []documentSourceRun {
	d.splitSourceRunAt(offset)
	var runs []documentSourceRun
	for _, run := range d.sourceRuns {
		if run.textEnd <= offset {
			continue
		}
		run.textStart -= offset
		run.textEnd -= offset
		run.sourceStart = newBase + max(0, run.sourceStart-sourcePosition)
		run.sourceEnd = newBase + max(0, run.sourceEnd-sourcePosition)
		runs = append(runs, run)
	}
	if len(runs) == 0 {
		runs = []documentSourceRun{{linear: true, sourceStart: newBase, sourceEnd: newBase}}
	}
	return runs
}

func (d *document) setBlockMarkerAtCursor() (int, bool) {
	d.deleted = nil
	offset := d.byteOffset(d.cursor)
	if d.blockMarkerCount == len(d.blockMarkers) {
		d.blockMarkers[0], d.blockMarkers[1] = offset, 0
		d.blockMarkerCount = 1
		d.nextBlockMarker = 0
	} else {
		d.blockMarkers[d.blockMarkerCount] = offset
		d.blockMarkerCount++
	}
	d.anchor = d.cursor
	return d.blockMarkerCount, d.blockMarkerCount == len(d.blockMarkers)
}

func (d *document) setBlockRange(start, end int) {
	d.deleted = nil
	start = d.byteOffset(d.positionAtByte(start))
	end = d.byteOffset(d.positionAtByte(end))
	d.blockMarkers = [2]int{start, end}
	d.blockMarkerCount = len(d.blockMarkers)
	d.nextBlockMarker = 0
	d.anchor = d.cursor
}

func (d *document) markCurrentLine() {
	line := d.lines[d.cursor.line]
	end := line.end
	if d.cursor.line+1 < len(d.lines) {
		end = d.lines[d.cursor.line+1].start
	}
	d.setBlockRange(line.start, end)
}

func (d *document) markToEndOfLine() {
	d.setBlockRange(d.byteOffset(d.cursor), d.lines[d.cursor.line].end)
}

func (d *document) findBlockMarker() (int, int, bool) {
	if d.blockMarkerCount == 0 {
		return 0, 0, false
	}
	index := d.nextBlockMarker % d.blockMarkerCount
	offset := d.blockMarkers[index]
	d.nextBlockMarker = (index + 1) % d.blockMarkerCount
	d.setCursor(d.positionAtByte(offset), false)
	return index, offset, true
}

func (d *document) clearBlockMarkers() {
	d.blockMarkers, d.blockMarkerCount, d.nextBlockMarker = [2]int{}, 0, 0
	d.anchor = d.cursor
}

func (d *document) adjustBlockMarkers(start, end, inserted int) {
	if d.blockMarkerCount == 0 {
		return
	}
	delta := inserted - (end - start)
	newLength := len(d.text) - (end - start) + inserted
	for index := 0; index < d.blockMarkerCount; index++ {
		offset := d.blockMarkers[index]
		switch {
		case start == end && offset >= start:
			offset += inserted
		case start < end && offset > start && offset < end:
			offset = start + inserted
		case start < end && offset >= end:
			offset += delta
		}
		d.blockMarkers[index] = max(0, min(offset, newLength))
	}
}

func (d *document) normalizeBlockMarkers() {
	for index := 0; index < d.blockMarkerCount; index++ {
		d.blockMarkers[index] = d.byteOffset(d.positionAtByte(d.blockMarkers[index]))
	}
}

func (d *document) replace(start, end int, text string) bool {
	start = d.byteOffset(d.positionAtByte(start))
	end = d.byteOffset(d.positionAtByte(end))
	start, end = min(start, end), max(start, end)
	return d.replaceBytes(start, end, text)
}

func (d *document) replaceBytes(start, end int, text string) bool {
	d.deleted = nil
	sourceStart, sourceEnd := d.sourcePositionAt(start), d.sourcePositionAt(end)
	d.replaceSourceRange(start, end, len(text), sourceStart, sourceEnd)
	d.adjustBlockMarkers(start, end, len(text))
	updated := d.text[:start] + text + d.text[end:]
	changed := updated != d.text
	d.text = updated
	d.indexLines()
	d.normalizeBlockMarkers()
	offset := start + len(text)
	cursor := d.positionAtByte(offset)
	if d.byteOffset(cursor) < offset && cursor.character < len(d.lines[cursor.line].cells) {
		cursor.character++
	}
	d.setCursor(cursor, false)
	if changed {
		d.revision++
	}
	return changed
}

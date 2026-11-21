package edit

import (
	"fmt"
	"strings"

	"github.com/rivo/uniseg"
)

type printLine struct {
	text, ending string
	pageBreak    bool
}

func (d *document) printStream() ([]byte, error) {
	if d.pageLength < 1 || d.printTopMargin < 0 || d.printLeftMargin < 0 ||
		d.formfeedAfter < 0 || d.formfeedAfter > 2 || d.tabWidth < 1 {
		return nil, fmt.Errorf("invalid print setup")
	}
	if d.newline != "\n" && d.newline != "\r\n" {
		return nil, fmt.Errorf("unsupported document line ending")
	}

	lineCount := len(d.lines)
	if strings.HasSuffix(d.text, "\n") {
		lineCount--
	}
	lines := make([]printLine, 0, lineCount)
	for index := 0; index < lineCount; index++ {
		line := d.lines[index]
		text := d.text[line.start:line.end]
		ending := ""
		if index < len(d.lines)-1 {
			end := line.end
			if end < len(d.text) && d.text[end] == '\r' {
				end++
			}
			if end < len(d.text) && d.text[end] == '\n' {
				end++
				ending = d.text[line.end:end]
			}
		}
		pageBreak := strings.HasPrefix(text, "\f")
		if pageBreak {
			text = text[1:]
		}
		lines = append(lines, printLine{text: text, ending: ending, pageBreak: pageBreak})
	}

	var output strings.Builder
	pageLines := 0
	pageStarted := false
	writeTopMargin := func() {
		for range d.printTopMargin {
			output.WriteString(d.newline)
		}
	}
	for _, line := range lines {
		if line.pageBreak {
			output.WriteByte('\f')
			pageLines = 0
			pageStarted = true
			writeTopMargin()
		} else if !pageStarted {
			pageStarted = true
			writeTopMargin()
		} else if pageLines == d.pageLength {
			if d.formfeedBetween {
				output.WriteByte('\f')
			}
			pageLines = 0
			writeTopMargin()
		}

		output.WriteString(strings.Repeat(" ", d.printLeftMargin))
		writePrintText(&output, line.text, d.printLeftMargin, d.tabWidth)
		output.WriteString(line.ending)
		pageLines++
	}
	for range d.formfeedAfter {
		output.WriteByte('\f')
	}
	return []byte(output.String()), nil
}

func (d *document) printStreamForText(text string) ([]byte, error) {
	selected := newDocument(text)
	selected.newline = d.newline
	selected.pageLength = d.pageLength
	selected.printTopMargin = d.printTopMargin
	selected.printLeftMargin = d.printLeftMargin
	selected.formfeedBetween = d.formfeedBetween
	selected.formfeedAfter = d.formfeedAfter
	selected.tabWidth = d.tabWidth
	return selected.printStream()
}

func writePrintText(output *strings.Builder, text string, column, tabWidth int) {
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		grapheme := graphemes.Str()
		if grapheme == "\t" {
			spaces := tabWidth - column%tabWidth
			output.WriteString(strings.Repeat(" ", spaces))
			column += spaces
			continue
		}
		output.WriteString(grapheme)
		column += graphemes.Width()
	}
}

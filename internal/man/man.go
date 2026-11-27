// Package man is the shared "manual page" model behind goBaDig's --man/
// "man <tópico>" (CLI/console) and the WibEdit "Manual" viewer: a Page is
// the same structure whether it came from Markdown (markdown.go) or,
// later (Fase 3), from the MSX BASIC dictionary - callers never need to
// know which. See OUTLINE.md ("Sistema de ajuda e manuais") for the
// phased plan this is part of.
package man

import "strings"

// LineStyle marks how a Page's Line should be rendered: a heading level
// (used for both emphasis and the TOC), a code block, a blockquote, a
// bullet, or plain paragraph text.
type LineStyle int

const (
	StyleNormal LineStyle = iota
	StyleHeading1
	StyleHeading2
	StyleHeading3
	StyleCode
	StyleQuote
	StyleBullet
)

// Line is one row of a rendered Page.
type Line struct {
	Text  string
	Style LineStyle
}

// TOCEntry links a heading's title to the Line index it starts at,
// letting a viewer jump straight to it (or, for the Index page, letting
// every entry act as a link to another topic's Page).
type TOCEntry struct {
	Title string
	Line  int
}

// Page is a single topic ready to display: the CLI prints Lines as
// plain text, WibEdit's viewer colors them by Style and lets Enter on a
// TOC line jump to it.
type Page struct {
	Title string
	Lines []Line
	TOC   []TOCEntry
}

// PlainLines renders p for a plain-text surface (CLI stdout): headings
// are upper-cased for visibility, everything else is printed as-is.
func (p Page) PlainLines() []string {
	lines := make([]string, 0, len(p.Lines))
	for _, line := range p.Lines {
		text := line.Text
		switch line.Style {
		case StyleHeading1, StyleHeading2, StyleHeading3:
			text = strings.ToUpper(text)
		}
		lines = append(lines, text)
	}
	return lines
}

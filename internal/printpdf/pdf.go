package printpdf

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-pdf/fpdf"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
)

func Render(stream []byte, fitWidth bool) ([]byte, error) {
	if !utf8.Valid(stream) {
		return nil, errors.New("PDF export requires valid UTF-8 text")
	}
	fontData, err := opentype.Parse(gomono.TTF)
	if err != nil {
		return nil, fmt.Errorf("loading embedded Go Mono font: %w", err)
	}
	face, err := opentype.NewFace(fontData, nil)
	if err != nil {
		return nil, fmt.Errorf("creating Go Mono font face: %w", err)
	}
	defer face.Close()
	text := strings.ReplaceAll(string(stream), "\r\n", "\n")
	for _, character := range text {
		if character == '\n' || character == '\f' {
			continue
		}
		if unicode.IsControl(character) {
			return nil, fmt.Errorf("PDF export does not support control character U+%04X", character)
		}
		if _, _, ok := face.GlyphBounds(character); !ok {
			return nil, fmt.Errorf("embedded Go Mono font has no glyph for U+%04X", character)
		}
	}

	const margin = 10.0
	fontSize, lineHeight := 9.0, 4.0
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(margin, margin, margin)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddUTF8FontFromBytes("GoMono", "", gomono.TTF)
	pdf.SetFont("GoMono", "", fontSize)
	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("configuring PDF font: %w", err)
	}
	pageWidth, pageHeight := pdf.GetPageSize()
	maxWidth := pageWidth - 2*margin
	if fitWidth {
		widest := 0.0
		for _, line := range strings.FieldsFunc(text, func(r rune) bool { return r == '\n' || r == '\f' }) {
			widest = max(widest, pdf.GetStringWidth(line))
		}
		if widest > maxWidth {
			pdf.SetFontSize(fontSize * maxWidth / widest)
		}
	}

	lineNumber := 0
	for _, section := range strings.Split(text, "\f") {
		pdf.AddPage()
		y := pdf.GetY() + lineHeight
		for _, line := range strings.Split(strings.TrimSuffix(section, "\n"), "\n") {
			lineNumber++
			if !fitWidth && pdf.GetStringWidth(line) > maxWidth {
				return nil, fmt.Errorf("PDF line %d exceeds the A4 printable width", lineNumber)
			}
			if y > pageHeight-margin {
				pdf.AddPage()
				y = pdf.GetY() + lineHeight
			}
			if line != "" {
				pdf.Text(margin, y, line)
			}
			y += lineHeight
		}
	}
	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, fmt.Errorf("generating PDF: %w", err)
	}
	return output.Bytes(), nil
}

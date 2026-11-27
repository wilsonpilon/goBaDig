package man

import "strings"

// wrapWidth is the column count ParseMarkdown wraps plain paragraphs
// to - narrow enough to stay readable in WibEdit's document area and a
// standard 80-column terminal alike.
const wrapWidth = 70

// ParseMarkdown turns a Markdown source into a Page: "#".."######"
// headings become TOC entries plus heading-styled lines, fenced code
// blocks (```) become verbatim StyleCode lines, "> " lines become
// StyleQuote, "-"/"*"/"N." list items become StyleBullet, and everything
// else is a paragraph, word-wrapped to wrapWidth. Page.Title is the
// first heading found, if any. v1 does not render tables or inline
// emphasis (*/**/_ pass through literally) - docs/man content avoids
// relying on those.
func ParseMarkdown(src string) Page {
	var page Page
	inCode := false
	var paragraph []string

	flush := func() {
		if len(paragraph) == 0 {
			return
		}
		text := strings.Join(paragraph, " ")
		for _, wrapped := range wrapText(text, wrapWidth) {
			page.Lines = append(page.Lines, Line{Text: wrapped, Style: StyleNormal})
		}
		paragraph = nil
	}

	for raw := range strings.SplitSeq(src, "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			flush()
			inCode = !inCode
			continue
		}
		if inCode {
			page.Lines = append(page.Lines, Line{Text: line, Style: StyleCode})
			continue
		}
		if trimmed == "" {
			flush()
			continue
		}
		if level, title, ok := parseHeading(trimmed); ok {
			flush()
			style := StyleHeading3
			switch level {
			case 1:
				style = StyleHeading1
			case 2:
				style = StyleHeading2
			}
			page.TOC = append(page.TOC, TOCEntry{Title: title, Line: len(page.Lines)})
			page.Lines = append(page.Lines, Line{Text: title, Style: style})
			continue
		}
		if strings.HasPrefix(trimmed, ">") {
			flush()
			text := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			page.Lines = append(page.Lines, Line{Text: text, Style: StyleQuote})
			continue
		}
		if text, ok := parseBullet(trimmed); ok {
			flush()
			page.Lines = append(page.Lines, Line{Text: "- " + text, Style: StyleBullet})
			continue
		}
		paragraph = append(paragraph, trimmed)
	}
	flush()

	if len(page.TOC) > 0 {
		page.Title = page.TOC[0].Title
	}
	return page
}

// parseHeading reports the level (1-6) and title of an ATX heading
// ("# Title".."###### Title"); trimmed must already have leading/
// trailing whitespace removed.
func parseHeading(trimmed string) (level int, title string, ok bool) {
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level >= len(trimmed) || trimmed[level] != ' ' {
		return 0, "", false
	}
	return level, strings.TrimSpace(trimmed[level:]), true
}

// parseBullet reports the text of a "-"/"*" or "N." list item;
// trimmed must already have leading/trailing whitespace removed.
func parseBullet(trimmed string) (string, bool) {
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return trimmed[2:], true
	}
	i := 0
	for i < len(trimmed) && trimmed[i] >= '0' && trimmed[i] <= '9' {
		i++
	}
	if i > 0 && i+1 < len(trimmed) && trimmed[i] == '.' && trimmed[i+1] == ' ' {
		return trimmed[i+2:], true
	}
	return "", false
}

func wrapText(text string, width int) []string {
	var lines []string
	line := ""
	for word := range strings.FieldsSeq(text) {
		if line != "" && len(line)+1+len(word) > width {
			lines = append(lines, line)
			line = ""
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	if line != "" {
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		lines = append(lines, "")
	}
	return lines
}

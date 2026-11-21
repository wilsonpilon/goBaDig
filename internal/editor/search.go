package edit

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type searchState struct {
	text        string
	replacement string
	matchCase   bool
	backwards   bool
	valid       bool
	start       int
	end         int
}

func findText(text, query string, cursor int, backwards, matchCase bool) (int, int, bool) {
	if query == "" {
		return 0, 0, false
	}
	cursor = max(0, min(cursor, len(text)))
	if matchCase {
		if backwards {
			if cursor == 0 {
				return 0, 0, false
			}
			index := strings.LastIndex(text[:cursor], query)
			return index, index + len(query), index >= 0
		}
		index := strings.Index(text[cursor:], query)
		if index < 0 {
			return 0, 0, false
		}
		start := cursor + index
		return start, start + len(query), true
	}

	queryRunes := []rune(query)
	if backwards {
		found, foundEnd := -1, 0
		for offset := 0; offset <= cursor; {
			end, ok := foldedMatchEnd(text, queryRunes, offset)
			if ok && end <= cursor {
				found, foundEnd = offset, end
			}
			if offset == cursor || offset == len(text) {
				break
			}
			_, size := utf8.DecodeRuneInString(text[offset:])
			offset += size
		}
		return found, foundEnd, found >= 0
	}

	for offset := cursor; offset < len(text); {
		if end, ok := foldedMatchEnd(text, queryRunes, offset); ok {
			return offset, end, true
		}
		_, size := utf8.DecodeRuneInString(text[offset:])
		offset += size
	}
	return 0, 0, false
}

func foldedMatchEnd(text string, query []rune, start int) (int, bool) {
	offset := start
	for _, expected := range query {
		if offset >= len(text) {
			return 0, false
		}
		actual, size := utf8.DecodeRuneInString(text[offset:])
		if !strings.EqualFold(string(actual), string(expected)) {
			return 0, false
		}
		offset += size
	}
	return offset, true
}

func (e *Editor) findSearchText(query string, matchCase, backwards bool) bool {
	e.area.document.searchMatch = false
	start, end, found := findText(
		e.area.document.text,
		query,
		e.area.document.byteOffset(e.area.document.cursor),
		backwards,
		matchCase,
	)
	if !found {
		e.area.notify(false)
		e.notice = "Text not found [search.find]."
		e.updateStatus()
		return false
	}

	e.lastSearch.start = start
	e.lastSearch.end = end
	cursor := end
	if backwards {
		cursor = start
	}
	e.area.document.setCursor(e.area.document.positionAtByte(cursor), false)
	e.area.document.setSearchMatch(start, end)
	e.area.notify(false)
	e.notice = "Found text [search.find]."
	e.updateStatus()
	return true
}

func (e *Editor) continueSearch(backwards bool, commandID string) bool {
	if !e.lastSearch.valid {
		e.notice = "Run Find before continuing a search [" + commandID + "]."
		e.updateStatus()
		return false
	}

	cursor := e.area.document.byteOffset(e.area.document.cursor)
	start, end, found := findText(
		e.area.document.text,
		e.lastSearch.text,
		cursor,
		backwards,
		e.lastSearch.matchCase,
	)
	if !found {
		direction := "forward"
		if backwards {
			direction = "backward"
		}
		e.notice = "No more matches " + direction + " [" + commandID + "]."
		e.updateStatus()
		return false
	}

	e.lastSearch.start, e.lastSearch.end = start, end
	cursor = end
	if backwards {
		cursor = start
	}
	e.area.document.setCursor(e.area.document.positionAtByte(cursor), false)
	e.area.document.setSearchMatch(start, end)
	e.area.notify(false)
	e.notice = "Found next match [" + commandID + "]."
	e.updateStatus()
	return true
}

func (e *Editor) goToLine(line int) bool {
	if line < 1 || line > len(e.area.document.lines) {
		e.showMessage("Line number", fmt.Sprintf("Line %d is outside this document.", line))
		return false
	}
	e.area.document.setCursor(documentPosition{line: line - 1}, false)
	e.area.notify(false)
	e.notice = fmt.Sprintf("Moved to line %d [search.line].", line)
	e.updateStatus()
	return true
}

func (e *Editor) searchTextInBlock() bool {
	d := e.area.document
	start, end, marked := d.markedBlock()
	if !marked {
		e.showMessage("Text in Block", "Set two block markers before searching.")
		return false
	}
	query := d.text[start:end]
	if query == "" {
		e.showMessage("Text in Block", "The marked block is empty.")
		return false
	}

	matchCase, backwards, replacement := false, false, ""
	if e.lastSearch.valid {
		matchCase, backwards = e.lastSearch.matchCase, e.lastSearch.backwards
		replacement = e.lastSearch.replacement
	}
	e.lastSearch = searchState{
		text:        query,
		replacement: replacement,
		matchCase:   matchCase,
		backwards:   backwards,
		valid:       true,
		start:       -1,
		end:         -1,
	}
	if !e.findSearchText(query, matchCase, backwards) {
		e.showMessage("Text in Block", "The marked text was not found from the cursor.")
		return false
	}
	e.notice = "Found marked text [search.block]."
	e.updateStatus()
	return true
}

package edit

func matchingBracket(text string, cursor int) (int, bool, bool) {
	if cursor < 0 || cursor >= len(text) {
		return 0, false, false
	}

	opening, closing := byte(0), byte(0)
	step := 1
	switch text[cursor] {
	case '(', ')':
		opening, closing = '(', ')'
	case '[', ']':
		opening, closing = '[', ']'
	case '{', '}':
		opening, closing = '{', '}'
	default:
		return 0, false, false
	}
	if text[cursor] == closing {
		step = -1
	}

	depth := 0
	for index := cursor; index >= 0 && index < len(text); index += step {
		switch text[index] {
		case opening:
			if step > 0 {
				depth++
			} else {
				depth--
			}
		case closing:
			if step > 0 {
				depth--
			} else {
				depth++
			}
		}
		if depth == 0 {
			return index, true, true
		}
	}
	return 0, true, false
}

func (e *Editor) matchingBracket() bool {
	d := e.area.document
	offset := d.byteOffset(d.cursor)
	match, isBracket, found := matchingBracket(d.text, offset)
	if !isBracket {
		e.showMessage("Matching bracket", "Place the cursor on a bracket.")
		return false
	}
	if !found {
		e.showMessage("Matching bracket", "No matching bracket was found.")
		return false
	}

	d.setCursor(d.positionAtByte(match), false)
	d.setSearchMatch(match, match+1)
	e.area.notify(false)
	e.notice = "Matching bracket found [search.bracket]."
	e.updateStatus()
	return true
}

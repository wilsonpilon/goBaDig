package edit

func firstWindowDifference(first, second string, firstOffset, secondOffset int) (int, int, bool) {
	firstOffset = max(0, min(firstOffset, len(first)))
	secondOffset = max(0, min(secondOffset, len(second)))
	firstRemaining := first[firstOffset:]
	secondRemaining := second[secondOffset:]
	limit := min(len(firstRemaining), len(secondRemaining))
	for index := 0; index < limit; index++ {
		if firstRemaining[index] != secondRemaining[index] {
			return firstOffset + index, secondOffset + index, true
		}
	}
	if len(firstRemaining) != len(secondRemaining) {
		return firstOffset + limit, secondOffset + limit, true
	}
	return 0, 0, false
}

func (e *Editor) windowDifferences() bool {
	if len(e.windows) != 2 {
		e.showMessage("Window Differences", "Open a second window before comparing.")
		return false
	}
	e.syncActiveWindow()
	otherIndex := 1 - e.activeWindow
	active := e.windows[e.activeWindow].area.document
	other := e.windows[otherIndex].area.document
	firstOffset, secondOffset, found := firstWindowDifference(
		active.text,
		other.text,
		active.byteOffset(active.cursor),
		other.byteOffset(other.cursor),
	)
	if !found {
		e.notice = "No differences found [search.differences]."
		e.updateStatus()
		return false
	}

	active.setCursor(active.positionAtByte(firstOffset), false)
	active.setSearchMatch(firstOffset, min(len(active.text), firstOffset+1))
	other.setCursor(other.positionAtByte(secondOffset), false)
	other.setSearchMatch(secondOffset, min(len(other.text), secondOffset+1))
	e.windows[otherIndex].area.notify(false)
	e.area.notify(false)
	e.notice = "Moved both windows to their next difference [search.differences]."
	e.updateStatus()
	return true
}

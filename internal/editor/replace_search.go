package edit

import (
	"fmt"
)

type replaceSession struct {
	search      string
	replacement string
	matchCase   bool
	backwards   bool
	start       int
	end         int
	replaced    int
}

func (e *Editor) beginReplaceSearch(search searchState, replacement string) {
	e.area.document.searchMatch = false
	e.replaceSession = &replaceSession{
		search:      search.text,
		replacement: replacement,
		matchCase:   search.matchCase,
		backwards:   search.backwards,
	}
	if !e.locateReplaceMatch(e.area.document.byteOffset(e.area.document.cursor)) {
		e.replaceSession = nil
		e.showMessage("Search", "Text not found.")
		return
	}
	e.showReplacePrompt()
}

func (e *Editor) locateReplaceMatch(cursor int) bool {
	session := e.replaceSession
	if session == nil {
		return false
	}
	start, end, found := findText(
		e.area.document.text,
		session.search,
		cursor,
		session.backwards,
		session.matchCase,
	)
	if !found {
		return false
	}
	session.start, session.end = start, end
	e.lastSearch.start, e.lastSearch.end = start, end
	e.area.document.setCursor(e.area.document.positionAtByte(start), false)
	e.area.document.setSearchMatch(start, end)
	e.area.notify(false)
	return true
}

func (e *Editor) showReplacePrompt() {
	session := e.replaceSession
	if session == nil {
		return
	}
	d := newDialog("Replace text?", 58, 9)
	d.context = "search.replace"
	d.label(2, 2, fmt.Sprintf("Replace this occurrence with %q?", session.replacement))
	d.buttons(
		[]string{"replace", "skip", "all", "cancel"},
		[]string{"Replace", "Skip", "All", "Cancel"},
		"replace",
	)
	d.result = func(action string) {
		e.closeDialog()
		switch action {
		case "replace":
			e.replaceCurrentMatch()
			e.advanceReplace(e.nextReplaceCursor(true))
		case "skip":
			e.advanceReplace(e.nextReplaceCursor(false))
		case "all":
			e.replaceAllRemaining()
		case "cancel":
			e.finishReplace(true)
		}
	}
	e.openDialog(d)
}

func (e *Editor) nextReplaceCursor(replaced bool) int {
	session := e.replaceSession
	if session.backwards {
		return session.start
	}
	if replaced {
		return session.start + len(session.replacement)
	}
	return session.end
}

func (e *Editor) replaceCurrentMatch() {
	session := e.replaceSession
	d := e.area.document
	changed := d.replaceBytes(session.start, session.end, session.replacement)
	e.area.notify(changed)
	session.replaced++
}

func (e *Editor) advanceReplace(cursor int) {
	if e.locateReplaceMatch(cursor) {
		e.showReplacePrompt()
		return
	}
	e.finishReplace(false)
}

func (e *Editor) replaceAllRemaining() {
	for {
		e.replaceCurrentMatch()
		if !e.locateReplaceMatch(e.nextReplaceCursor(true)) {
			e.finishReplace(false)
			return
		}
	}
}

func (e *Editor) finishReplace(cancelled bool) {
	session := e.replaceSession
	if session == nil {
		return
	}
	count := session.replaced
	e.replaceSession = nil
	if cancelled {
		e.notice = fmt.Sprintf("Replace cancelled after %d occurrence(s).", count)
	} else {
		e.notice = fmt.Sprintf("Replaced %d occurrence(s).", count)
	}
	e.updateStatus()
}

package edit

// Prefix-based autocomplete for WibEdit, reusing the per-dialect
// keyword/function tables syntax.go already maintains for highlighting
// (internal/editor/syntax_keywords.go) - no new data source needed. See
// OUTLINE.md/resume.txt ("autocomplete + ajuda sensivel ao contexto")
// for the feature this is the first half of.

import (
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// maxAutocompleteCandidates caps how many matches the popup ever shows
// at once, same spirit as the user's own MSXIDE (MAX_AC_MATCHES=30
// there) - keeps the popup from growing to cover the whole screen on a
// one-letter prefix.
const maxAutocompleteCandidates = 30

// autocompleteState is nil whenever the popup is closed; Editor.ac
// holds the only live instance, mirroring how e.dropdown/e.dialogs work.
type autocompleteState struct {
	anchor     documentPosition // start of the word being completed
	candidates []string
	selected   int
}

// refreshAutocomplete re-evaluates the popup against the current
// cursor position. Called from bindWindow's changed/moved callbacks
// (editor.go) - i.e. after every edit and every cursor movement, not
// from Capture directly, so it naturally reacts to typing, Backspace,
// and the arrow keys without Capture needing a case for each.
func (e *Editor) refreshAutocomplete() {
	dialect := e.area.dialect
	if dialect == dialectNone {
		e.closeAutocomplete()
		return
	}
	prefix, anchor := wordPrefixBeforeCursor(e.area.document)
	candidates := candidatesFor(dialect, prefix)
	if len(candidates) == 0 {
		e.closeAutocomplete()
		return
	}
	// anchor/candidates are recomputed from scratch on every keystroke
	// (there is no separate "typing" vs. "navigating" state) - selected
	// always resets to 0 here. It only ever moves away from 0 via the
	// Up/Down branches in Capture, which don't go through this
	// function at all.
	e.ac = &autocompleteState{anchor: anchor, candidates: candidates}
}

func (e *Editor) closeAutocomplete() {
	e.ac = nil
}

// confirmAutocomplete replaces the prefix being completed with the
// selected candidate, the same way the literal-input branch in
// Capture replaces text: compute the byte range, close the popup
// first (so the refresh that notify() triggers below sees e.ac == nil
// and starts clean), then replace and notify.
func (e *Editor) confirmAutocomplete() {
	if e.ac == nil {
		return
	}
	if len(e.ac.candidates) == 0 {
		e.closeAutocomplete()
		return
	}
	choice := e.ac.candidates[e.ac.selected]
	d := e.area.document
	start := d.byteOffset(d.clamp(e.ac.anchor))
	end := d.byteOffset(d.cursor)
	e.closeAutocomplete()
	e.area.notify(d.replace(start, end, choice))
}

// wordPrefixBeforeCursor scans backwards from the cursor on its own
// line using the same isWordChar rule syntax.go's tokenizer already
// applies, returning the partial word immediately before the cursor
// (empty if the cursor isn't right after a word character - e.g. at
// the start of a line, after a space, or after punctuation).
func wordPrefixBeforeCursor(d *document) (string, documentPosition) {
	cursor := d.cursor
	line := d.lines[cursor.line]
	end := cursor.character
	start := end
	for start > 0 {
		cell := line.cells[start-1]
		if len(cell.text) != 1 || !isWordChar(cell.text[0]) {
			break
		}
		start--
	}
	if start == end {
		return "", cursor
	}
	var prefix strings.Builder
	for _, cell := range line.cells[start:end] {
		prefix.WriteString(cell.text)
	}
	return prefix.String(), documentPosition{line: cursor.line, character: start}
}

// candidatesFor combines a dialect's keyword and function tables
// (syntax_keywords.go) into a sorted, capped, case-insensitive
// prefix-match candidate list. Excludes a candidate identical to the
// already-fully-typed prefix (nothing to complete to), same as the
// user's own MSXIDE.
func candidatesFor(dialect syntaxDialect, prefix string) []string {
	if prefix == "" {
		return nil
	}
	upperPrefix := strings.ToUpper(prefix)
	var tables []map[string]bool
	switch dialect {
	case dialectMSX:
		tables = []map[string]bool{msxKeywords, msxFunctions}
	case dialectZX81:
		tables = []map[string]bool{zx81Keywords, zx81Functions}
	case dialectZXSpectrum:
		tables = []map[string]bool{zxSpectrumKeywords, zxSpectrumFunctions}
	default:
		return nil
	}
	var matches []string
	for _, table := range tables {
		for word := range table {
			if word != upperPrefix && strings.HasPrefix(word, upperPrefix) {
				matches = append(matches, word)
			}
		}
	}
	sort.Strings(matches)
	if len(matches) > maxAutocompleteCandidates {
		matches = matches[:maxAutocompleteCandidates]
	}
	return matches
}

// drawAutocompletePopup draws Editor.ac's candidate list anchored at
// the word being completed, recomputing its screen position fresh
// every call (same approach documentView.Draw already uses to
// position the terminal cursor) so it tracks scrolling correctly.
//
// Deliberately NOT a tview.Primitive added as a Pages entry: an
// earlier version was (mirroring editorDropdown), but
// tview.Pages.AddPage steals focus to the newly added page whenever
// the Pages container already has focus (rivo/tview@v0.42.0/
// pages.go:88-90, Pages.Focus picks the topmost visible page) - since
// *tview.Box has no InputHandler of its own, that silently swallowed
// every keystroke after the popup first opened. Editor.Draw (below)
// calls this directly after Pages.Draw instead, so the popup is pure
// overlay: it never owns the application's focus, which stays on
// documentView the whole time, and Editor.Capture intercepts only the
// handful of keys that matter to the popup itself (Tab/Enter/Up/Down/
// Esc) before they would otherwise reach the document or the menu/
// command dispatch.
func drawAutocompletePopup(screen tcell.Screen, e *Editor) {
	ac := e.ac
	if ac == nil || len(ac.candidates) == 0 {
		return
	}
	v := e.area
	d := v.document
	vx, vy, vwidth, vheight := v.GetInnerRect()
	if vwidth <= 0 || vheight <= 0 {
		return
	}
	gutter := d.markerGutterWidth()
	scrollbarWidth := 0
	if v.scrollbar {
		scrollbarWidth = 1
	}
	contentWidth := max(1, vwidth-gutter-scrollbarWidth)

	anchor := d.clamp(ac.anchor)
	row := d.rowAt(anchor)
	column := d.column(anchor) - d.left
	if d.wrap {
		column = d.column(anchor) - d.rows[row].from
	}
	if row < d.top || row >= d.top+vheight || column < 0 || column >= contentWidth {
		return
	}
	anchorX := vx + gutter + column
	anchorY := vy + row - d.top

	width := 0
	for _, candidate := range ac.candidates {
		width = max(width, len(candidate)+2)
	}
	width = max(1, min(width, vwidth))
	height := min(len(ac.candidates)+2, vheight)

	sw, sh := screen.Size()
	x := max(0, min(anchorX, sw-width))
	y := anchorY + 1
	if y+height > sh || y+height > vy+vheight {
		if anchorY-height >= vy {
			y = anchorY - height
		}
	}

	drawPanel(screen, x, y, width, height, "", false)
	visible := max(0, height-2)
	top := 0
	if ac.selected >= visible {
		top = ac.selected - visible + 1
	}
	for row := 0; row < visible && row+top < len(ac.candidates); row++ {
		candidate := ac.candidates[row+top]
		style := menuStyle
		if row+top == ac.selected {
			style = selectedStyle
		}
		for col := 1; col < width-1; col++ {
			screen.SetContent(x+col, y+row+1, ' ', nil, style)
		}
		drawLabel(screen, x+1, y+row+1, width-2, candidate, 0, style)
	}
}

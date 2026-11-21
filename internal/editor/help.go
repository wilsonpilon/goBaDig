package edit

import (
	"fmt"
	"strings"
)

func (e *Editor) commandHelp(command commandSpec) []string {
	lines := []string{command.label + " [" + command.id + "]"}
	for _, key := range command.keys {
		lines = append(lines, "Key: "+key)
	}
	if command.description != "" {
		lines = append(lines, wrapDialogText(command.description, 60)...)
	}
	state := "Planned " + command.phase + ". No document operation yet."
	if strings.HasPrefix(command.id, "edit.") {
		state = "Available R06: explicit NE-profile editing."
	}
	switch command.id {
	case "file.exit":
		state = "Available R07: Save, Don't Save and Cancel preserve the selected action."
	case "file.new":
		state = "Available R07: starts an empty document after resolving unsaved changes."
	case "file.open":
		state = "Available R07: opens a file without changing its bytes or line endings."
	case "file.save":
		state = "Available R07: saves with a backup and checks for external changes."
	case "file.save-as":
		state = "Available R07: chooses a destination and confirms overwriting existing files."
	case "file.merge":
		state = "Available R12: inserts the selected file at the cursor without replacing a selection."
	case "file.load-more":
		state = "Available R21: appends the next 1 MiB from the open file."
	case "file.write-cursor":
		state = "Available R21: writes through the cursor and keeps the remaining text editable."
	case "file.top":
		state = "Available R21: reloads the saved file from its beginning."
	case "file.shell":
		state = "Available R22: suspends the editor and starts PowerShell."
	case "search.find":
		state = "Available R17: finds text from the current cursor."
		lines = append(lines, "Search can match case and proceed backwards.",
			"Forward leaves the cursor after the match; backward before it.",
			"Text not found and an empty query are reported; search does not wrap.")
	case "search.forward":
		state = "Available R19a: continues the last search forwards."
		lines = append(lines, "Uses the last search text and case option.",
			"Starts at the cursor and leaves it after the match; no wrap.",
			"Reversing direction can find the same occurrence again.")
	case "search.reverse":
		state = "Available R19b: continues the last search backwards."
		lines = append(lines, "Uses the last search text and case option.",
			"Starts at the cursor and leaves it before the match; no wrap.",
			"Reversing direction can find the same occurrence again.")
	case "search.line":
		state = "Available R19c: moves to the start of a requested line."
		lines = append(lines, "Line numbers are 1-based; a number past the",
			"document end is reported without moving the cursor.")
	case "search.bracket":
		state = "Available R19d: moves to the matching bracket."
		lines = append(lines, "Place the cursor on (, ), [, ], { or }.",
			"Nested occurrences of the same bracket pair are counted.")
	case "search.differences":
		state = "Available R19e: locates the next difference between two windows."
		lines = append(lines, "Comparison begins at each window's cursor.",
			"Both windows move to their respective differing positions.")
	case "search.block":
		state = "Available R19f: searches for the text between two block markers."
		lines = append(lines, "Uses the last search's case and direction options.",
			"The marked text and block markers remain unchanged.")
	case "search.replace":
		state = "Available R18: finds and replaces matching text."
		lines = append(lines, "Each occurrence can be replaced, skipped, or",
			"replaced along with the remaining matches.")
	case "file.window":
		state = "Available R08: opens a second independent document window; F2 switches between open windows."
		lines = append(lines, "Each window keeps its own cursor, scroll position and modified state.",
			"Opening the same path twice creates separate buffers; external-change checks protect saves.")
	case "file.close-window":
		if len(e.windows) == 2 {
			state = "Available R08: closes the active window after resolving its unsaved changes."
		} else {
			state = "Available when the second window is open."
		}
	case "block.set":
		state = "Available R09: sets visible start/end markers; two markers select a block."
		lines = append(lines, "A third Set starts a new block at the current cursor.")
	case "block.remove":
		state = "Available R09: clears block markers and the marked selection."
	case "block.find":
		state = "Available R09: moves between the set block boundaries."
		lines = append(lines, "With one marker, Find returns to that boundary.")
	case "search.marker":
		state = "Available R19g: Search > Block Marker moves between set boundaries."
		lines = append(lines, "It retains R09 marker cycling; validation is part of",
			"the final Norton Editor 2 Search-menu comparison.")
	case "block.mark-line":
		state = "Available R09: marks the current physical line, including its line ending when present."
	case "block.mark-end":
		state = "Available R09: marks from the cursor through the current line end."
	case "block.copy":
		state = "Available R10: copies the marked bytes at the cursor."
		lines = append(lines, "Copy is a no-op when the cursor is inside the block.")
	case "block.move":
		state = "Available R10: moves the marked bytes to the cursor."
		lines = append(lines, "Destinations inside or at the edges of the block are ignored.",
			"Move requires the entire source file to be loaded.")
	case "block.delete":
		state = "Available R10: asks before deleting the marked bytes."
	case "block.delete-direct":
		state = "Available R10: Delete in the Block menu removes the block without confirmation."
	case "block.indent":
		state = "Available R11: adds the requested number of spaces to each marked line."
	case "block.unindent":
		state = "Available R11: removes the requested number of leading columns."
		lines = append(lines, "A partially removed tab is replaced with spaces.")
	case "block.output":
		state = "Available R12: writes only the marked bytes to a separate file."
		lines = append(lines, "Existing files require confirmation and receive a backup.",
			"The active document path cannot be used as the output path.")
	case "layout.wrap":
		state = "Available R13: wraps at word boundaries while editing."
		lines = append(lines, "Toggling Word wrap does not reformat existing text.")
	case "layout.auto":
		state = "Available R13: Enter copies the current line indentation."
		lines = append(lines, "Ctrl-Enter continues to indent even when Auto is off.")
	case "layout.setup":
		state = "Available R13: configures line length, margins and format start."
		lines = append(lines, "Applying setup changes options only; existing text is unchanged.")
	case "layout.paragraph":
		state = "Available R14: reformats the current paragraph."
		lines = append(lines, "The setup option chooses Cursor or Top of paragraph.",
			"Blank lines separate paragraphs; words are separated by one space.")
	case "layout.document":
		state = "Available R14: reformats paragraphs from Top or Cursor onward."
		lines = append(lines, "Earlier paragraphs remain unchanged when starting at Cursor.")
	case "layout.center":
		state = "Available R14: centers the current physical line within configured margins."
		lines = append(lines, "Leading and trailing whitespace is normalized.")
	case "layout.expand-tabs":
		state = "Available R15: expands tabs to spaces using configured tab stops."
		lines = append(lines, "Top processes the full document; Cursor starts on the current line.")
	case "layout.put-tabs":
		state = "Available R15: converts leading indentation to tabs and residual spaces."
		lines = append(lines, "Spaces inside text are preserved; Top and Cursor select line scope.")
	case "layout.insert-pages":
		state = "Available R16: inserts form feeds after each configured page length."
		lines = append(lines, "The current default is 62 lines; Top and Cursor select the starting line.")
	case "layout.remove-pages":
		state = "Available R16: removes form-feed page breaks."
		lines = append(lines, "Top processes the full document; Cursor starts on the current line.")
	case "print.setup":
		state = "Available R23: sets page margins, page length and formfeed options."
		lines = append(lines, "Settings drive the deterministic paged print stream.",
			"Print sends that stream to the operating system's default printer.",
			"Printer errors are reported; the document remains unchanged.")
	case "print.document":
		state = "Available R24a: sends the complete paged document to the default system printer."
		lines = append(lines, "Print Setup controls margins, page length and formfeeds.",
			"Printer errors are reported; the document remains unchanged.")
	case "print.block":
		state = "Available R24b: sends only the marked bytes to the default system printer."
		lines = append(lines, "Print Setup controls margins, page length and formfeeds.",
			"Two block markers are required; printer errors are reported.")
	case "print.eject":
		state = "Available R24c: sends one formfeed to the default system printer."
		lines = append(lines, "Eject page is separate from PDF export.",
			"Printer errors are reported; the document remains unchanged.")
	case "print.pdf":
		state = "Available R24d: exports the paged document to PDF with an embedded Unicode font."
		lines = append(lines, "Go Mono covers many scripts; unsupported characters and",
			"lines too wide for A4 are reported instead of omitted.")
	case "view.information":
		state = "Available R20a: shows full-file counts, file details, time, free space and the character at the cursor."
		lines = append(lines, "Counts include text outside a partially loaded buffer.",
			"File-system values come from the current operating system.")
	case "view.outline":
		state = "Available R20b: temporarily shows a compact view of loaded text."
		lines = append(lines, "Blank lines and leading indentation are hidden in this view.",
			"The document, cursor and modified state are unchanged.",
			"Press any key to restore the normal view; partial files show loaded text only.")
	case "view.about":
		state = "Available R20c: identifies WibEdit, its version and authorship."
		lines = append(lines, "The About display does not claim complete legacy compatibility.")
	case "options.display":
		state = "Available R25: applies color profile, scrollbar and tab spacing."
		lines = append(lines, "Display height and effects unsupported by the terminal UI are not applied.",
			"The dialog identifies unsupported choices explicitly.")
	case "options.mouse":
		state = "Available R26: selects the active mouse button and editor scroll delay."
		lines = append(lines, "Pointer speed and sync-to-clock timing are host-controlled and not applied.")
	case "edit.insert":
		state = "Available R06: Insert/Over; Never Over prevents overwrite."
	case "edit.literal":
		state = "Available R06: Ctrl-P inserts the next character/control literally."
		lines = append(lines, "Literal Esc is data, not exit. NUL is preserved.",
			"Literal Enter stores CR; the Unicode view displays it as a control.")
	case "edit.undelete":
		state = "Available R06: restores the last deletion and its original cursor."
		lines = append(lines, "One-shot restoration; movement or further editing cancels it.")
	case "edit.word-right":
		lines = append(lines, "Deletes through the word, leaving its following separator.")
	case "edit.unindent":
		lines = append(lines, "Inside leading whitespace, deletes back to the previous tab stop.")
	case "options.editor":
		state = "Available R27: insertion/Tab modes, backup character and backup creation."
		lines = append(lines, "Only the latest backup is retained; deletion policy is unavailable.",
			"Alt-alone uses F10 because isolated Alt is terminal-controlled.")
	case "options.save":
		state = "Available R28: saves editor Options to the current or directory configuration."
		lines = append(lines, "Directory configuration overrides the global configuration.",
			"Changed settings prompt to Save, Don't Save or Cancel on exit.")
	case "menu.open":
		state = "Available: F10 is the terminal alternative to Alt alone."
	case "edit.backspace", "edit.delete", "edit.enter", "edit.tab":
		state = "Available R06: grapheme editing, CRLF defaults and preserved LF files."
	}
	if command.id == "window.divider-up" || command.id == "window.divider-down" {
		if len(e.windows) == 2 {
			state = "Available R08: moves the split between the two document windows."
		} else {
			state = "Available when the second window is open."
		}
	}
	if strings.HasPrefix(command.id, "cursor.") {
		state = "Available R05: model navigation never changes document bytes."
		switch command.id {
		case "cursor.up", "cursor.down":
			lines = append(lines, "Retains the desired visual column across short lines/tabs.")
		case "cursor.page-up", "cursor.page-down":
			lines = append(lines, "Moves one viewport minus one row (22 rows at 80x25).")
		case "cursor.scroll-up", "cursor.scroll-down":
			lines = append(lines, "Alt-Z scrolls text up; Ctrl-Z scrolls text down, not undo.",
				"The cursor moves only when it would leave the viewport.")
		}
	}
	if command.dialog != "" && !e.available(command) {
		lines = append(lines, "Dialog preview available; OK never applies future operations.")
	}
	lines = append(lines, wrapDialogText(state, 60)...)
	lines = append(lines, "")
	return lines
}

func paginateHelp(lines []string, limit int) [][]string {
	var pages [][]string
	for len(lines) > 0 {
		count := min(limit, len(lines))
		pages = append(pages, append([]string(nil), lines[:count]...))
		lines = lines[count:]
	}
	return pages
}

func keySummaryLine(description string, ids ...string) string {
	var keys []string
	for _, id := range ids {
		if command, ok := commandByID(id); ok {
			keys = append(keys, command.keys...)
		}
	}
	return strings.Join(keys, "/") + ": " + description
}

func keySummaryPages() [][]string {
	return [][]string{
		{
			"Movement",
			keySummaryLine("move by character or physical line.", "cursor.left", "cursor.right", "cursor.up", "cursor.down"),
			keySummaryLine("move to the previous or next word.", "cursor.word-left", "cursor.word-right"),
			keySummaryLine("move to line start or end.", "cursor.home", "cursor.end"),
			keySummaryLine("move to file start or end.", "cursor.file-top", "cursor.file-end"),
			keySummaryLine("move by one page.", "cursor.page-up", "cursor.page-down"),
			keySummaryLine("move to viewport top/bottom.", "cursor.screen-top", "cursor.screen-bottom"),
			"Up/Down keeps visual column; tabs stop every 8 columns.",
			"",
			"Scrolling",
			keySummaryLine("scroll the display up.", "cursor.scroll-up"),
			keySummaryLine("scroll the display down (not Undo).", "cursor.scroll-down"),
			"The cursor moves only if it would leave the view.",
			"",
		},
		{
			"Editing",
			keySummaryLine("delete to the left or right.", "edit.backspace", "edit.delete"),
			keySummaryLine("delete a word to the left or right.", "edit.word-left", "edit.word-right"),
			keySummaryLine("delete to the start or end of the line.", "edit.line-left", "edit.line-right"),
			keySummaryLine("delete the current line.", "edit.kill-line"),
			keySummaryLine("restore the last deletion if the cursor stays put.", "edit.undelete"),
			keySummaryLine("flip case before/after cursor on this line.", "edit.case-left", "edit.case-right"),
			keySummaryLine("insert the next key literally; Esc becomes data.", "edit.literal"),
			keySummaryLine("new line; Ctrl-Enter copies indentation.", "edit.enter", "edit.indent-enter"),
			keySummaryLine("unindent one level (Shift-Tab).", "edit.unindent"),
			keySummaryLine("insert a tab, spaces, or move to the next stop.", "edit.tab"),
			keySummaryLine("toggle Insert/Over; Never Over blocks overstrike.", "edit.insert"),
			"New files use CRLF; opened LF files retain LF.",
			"",
		},
		{
			"Help and files",
			"F1 opens help; F10 opens the File menu.",
			"Alt-F/B/T/S/P/V/O opens the corresponding menu.",
			"Esc closes a menu/dialog or requests editor exit.",
			"File > New/Open/Save/Save As perform real operations.",
			"wibedit file.prg opens that file; missing paths prompt.",
			"Saving an existing file creates a legacy .{XX backup.",
			"Unsaved New/Open/exit offers Save/Don't Save/Cancel.",
			"",
			"Preview-only keys (their actions are not implemented):",
			keySummaryLine("open previews; no search or replace runs.", "search.find", "search.replace"),
			keySummaryLine("planned search-repeat/bracket commands.", "search.forward", "search.reverse", "search.bracket"),
			keySummaryLine("move the divider between open windows.", "window.divider-up", "window.divider-down"),
			"Use Next for detailed command notes and limitations.",
		},
	}
}

func (e *Editor) helpContent(topic string) (string, [][]string, bool) {
	switch topic {
	case "help.general":
		return "General help", [][]string{
			{
				"WibEdit - source code and plain text in a terminal.",
				"Profile: " + compatibilityProfile + " [ui.screen]",
				"",
				"F1 opens the help index. F10 opens File [menu.open].",
				"Alt-F/B/T/S/P/V/O opens the corresponding menu.",
				"Up/Down selects; Left/Right changes menus.",
				"Enter or an item letter activates the selection.",
				"Mouse clicks select menus and items [ui.menu].",
				"",
				"Esc closes menus/help/dialogs, or requests editor exit.",
				"New, Open and exit protect modified text [file.new/file.open/file.exit].",
				"Press N or Next for dialog and compatibility notes.",
			},
			{
				"Dialogs [dialog.focus]: Tab/Shift-Tab changes fields.",
				"Arrows select list entries and radio choices.",
				"Space toggles a checkbox or activates a button.",
				"Enter uses the default button; Esc cancels.",
				"Button initials work outside text fields.",
				"Alt+accelerator works without stealing typed text.",
				"F1 shows context help and returns to the same field.",
				"",
				"File dialogs perform real operations; other previews do not apply values.",
				"Pending commands report their requirement ID and phase.",
				"WW wraps at word boundaries while editing; Auto indents Enter.",
				"R06: Ctrl-P literal input; Ctrl-U conditional undelete.",
				"Native NE behavior certification remains pending.",
			},
		}, true
	case "help.keys":
		var lines []string
		for _, page := range keySummaryPages() {
			lines = append(lines, page...)
		}
		for _, command := range commands() {
			if len(command.keys) > 0 {
				lines = append(lines, e.commandHelp(command)...)
			}
		}
		lines = append(lines, "Terminal note: Ctrl-2, Ctrl-Enter and modified navigation",
			"depend on terminal delivery. F10 substitutes for Alt alone.",
			"The historical Ctrl-left indentation glyph is interpreted as",
			"Ctrl-Backspace; original runtime confirmation is pending.",
			"Ctrl-Z is scroll, not undo. Ctrl-U is conditional undelete.",
			"Text navigation uses the independent Go document model.",
			"Tabs use stops every 8 columns; status shows visual columns.",
			"Word movement follows fwMSX letter/digit/underscore rules.",
			"Byte offsets and grapheme positions remain distinct.")
		return "Help on keys", paginateHelp(lines, 14), true
	case "help.about":
		return "About WibEdit", [][]string{append(e.aboutSummary(),
			"Independent source and plain-text editor, implemented in Go.",
			"Built into goBaDig, the Basic Dignified Suite.",
			"All rights reserved. See the goBaDig LICENSE.")}, true
	}
	menuTopics := map[string]rune{
		"help.file": 'F', "help.block": 'B', "help.layout": 'T',
		"help.search": 'S', "help.print": 'P', "help.view": 'V', "help.options": 'O',
	}
	if key, ok := menuTopics[topic]; ok {
		menu, _ := e.menu(key)
		var lines []string
		for _, command := range menu.items {
			if command.mnemonic != 0 {
				lines = append(lines, fmt.Sprintf("Menu: Alt-%c %c", key, command.mnemonic))
			}
			lines = append(lines, e.commandHelp(command)...)
		}
		if key == 'B' {
			for _, id := range []string{"block.delete-direct", "block.mark-line", "block.mark-end"} {
				command, _ := commandByID(id)
				lines = append(lines, e.commandHelp(command)...)
			}
			lines = append(lines, "Hidden menu keys: L marks the line; E marks to line end (R09).",
				"Delete removes without confirmation.")
		}
		if key == 'V' {
			lines = append(lines, "Alt-V A opens About WibEdit [help.about].")
		}
		return menu.label + " menu help", paginateHelp(lines, 14), true
	}
	if command, ok := commandByID(topic); ok && !strings.HasPrefix(topic, "help.") {
		return command.label + " help", paginateHelp(e.commandHelp(command), 14), true
	}
	return "", nil, false
}

func (e *Editor) showHelp(topic string) {
	title, pages, ok := e.helpContent(topic)
	if !ok || len(pages) == 0 {
		e.showMessage("Help not available", "No help topic for "+topic)
		return
	}
	d := newDialog(title, 66, 22)
	d.context = topic
	d.helpPages = pages
	d.result = func(id string) {
		switch id {
		case "next":
			if d.helpPage+1 < len(d.helpPages) {
				d.helpPage++
				d.setHelpPage()
			}
		case "previous":
			if d.helpPage > 0 {
				d.helpPage--
				d.setHelpPage()
			}
		case "help":
			d.helpPage = 0
			d.setHelpPage()
		default:
			e.closeDialog()
		}
	}
	d.setHelpPage()
	e.openDialog(d)
}

func (d *editorDialog) setHelpPage() {
	d.widgets, d.fields = nil, nil
	d.focus, d.scroll = 0, 0
	for row, line := range d.helpPages[d.helpPage] {
		d.label(2, row+1, line)
	}
	footer := fmt.Sprintf("Page %d/%d", d.helpPage+1, len(d.helpPages))
	if d.helpPage+1 < len(d.helpPages) {
		footer += " - More"
	}
	d.label(2, 17, footer)
	var ids, labels []string
	defaultID := "cancel"
	if d.helpPage+1 < len(d.helpPages) {
		ids, labels = append(ids, "next"), append(labels, "Next")
		defaultID = "next"
	}
	if d.helpPage > 0 {
		ids, labels = append(ids, "previous"), append(labels, "Previous")
	}
	ids, labels = append(ids, "cancel"), append(labels, "Cancel")
	d.buttons(ids, labels, defaultID)
	for focus, index := range d.fields {
		if d.widgets[index].id == defaultID {
			d.focus = focus
		}
	}
}

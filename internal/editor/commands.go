package edit

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

const compatibilityProfile = "NE2-terminal"

type commandSpec struct {
	id, label, phase, description, dialog string
	mnemonic                              rune
	keys                                  []string
	separator                             bool
}

type menuSpec struct {
	key   rune
	label string
	items []commandSpec
}

// Labels and bindings are independently implemented from the local NE references
// and the user's fwMSX specification, not from the reorganized fwMSX menus.
var neMenus = []menuSpec{
	{'F', "File", []commandSpec{
		{"file.new", "New", "F2.3", "Start an empty document.", "", 'N', nil, false},
		{"file.open", "Open...", "F2.3", "Load a text or program file.", "file", 'O', nil, false},
		{"file.save", "Save", "F2.3", "Write the current document safely.", "", 'S', nil, false},
		{"file.save-as", "Save As...", "F2.3", "Write the document with another name.", "file", 'A', nil, false},
		{"file.merge", "Merge...", "F2.3", "Insert another file at the cursor.", "file", 'M', nil, false},
		{"file.window", "2nd window...", "F2.4", "Open or activate the second window.", "file", '2', []string{"F2"}, true},
		{"file.close-window", "Close window", "F2.4", "Close the active window.", "", 'C', nil, false},
		{"file.load-more", "Load more", "F6.1", "Read the next part of a large file.", "", 'L', nil, true},
		{"file.write-cursor", "Write to cursor", "F6.1", "Write the prefix and reclaim buffer space.", "", 'W', nil, false},
		{"file.top", "Top of file", "F6.1", "Return to the start of a partial file.", "", 'T', nil, false},
		{"file.shell", "PowerShell", "F6.2", "Suspend editing and open PowerShell.", "", 'P', nil, true},
		{"file.exit", "eXit", "F1.3", "Return to goBaDig; confirm unsaved text.", "", 'X', nil, false},
	}},
	{'B', "Block", []commandSpec{
		{"block.copy", "Copy", "F3.2", "Copy marked text to the cursor.", "", 'C', nil, false},
		{"block.move", "Move", "F3.2", "Move marked text to the cursor.", "", 'M', nil, false},
		{"block.delete", "Delete", "F3.2", "Confirm removal of marked text.", "", 'D', nil, false},
		{"block.set", "Set marker", "F3.1", "Set a boundary of the text block.", "", 'S', nil, true},
		{"block.remove", "Remove markers", "F3.1", "Clear block boundaries.", "", 'R', nil, false},
		{"block.find", "Find marker", "F3.1", "Move to a block boundary.", "", 'F', nil, false},
		{"block.indent", "Indent...", "F3.3", "Indent lines in the marked block.", "indent", 'I', nil, true},
		{"block.unindent", "Unindent...", "F3.3", "Remove indentation in the marked block.", "indent", 'U', nil, false},
		{"block.output", "Output to file...", "F3.4", "Write only the marked block.", "file", 'O', nil, true},
	}},
	{'T', "Text-layout", []commandSpec{
		{"layout.paragraph", "Paragraph format", "F4.2", "Reflow the current paragraph using configured margins.", "", 'P', nil, false},
		{"layout.document", "Document format...", "F4.2", "Reflow from the top or cursor.", "start", 'D', nil, false},
		{"layout.center", "Center line", "F4.2", "Center the current line within configured margins.", "", 'C', nil, false},
		{"layout.setup", "Setup...", "F4.1", "Set line length and paragraph margins.", "layout", 'S', nil, true},
		{"layout.wrap", "Word wrap", "F4.1", "Toggle automatic word wrap while editing.", "", 'W', nil, false},
		{"layout.insert-pages", "Insert page breaks...", "F4.4", "Insert form feeds at page-length intervals.", "start", 'I', nil, true},
		{"layout.remove-pages", "Remove page breaks...", "F4.4", "Remove form-feed page breaks.", "start", 'B', nil, false},
		{"layout.expand-tabs", "Expand tabs...", "F4.3", "Replace tab characters with spaces.", "start", 'E', nil, true},
		{"layout.put-tabs", "Put in Tabs...", "F4.3", "Replace leading spaces with tabs.", "start", 'T', nil, false},
		{"layout.auto", "Auto indent", "F4.1", "Toggle indentation copying when Enter creates a line.", "", 'A', nil, true},
	}},
	{'S', "Search", []commandSpec{
		{"search.find", "Find string...", "F5.1", "Locate text with case and direction options.", "search", 'F', []string{"Ctrl-F"}, false},
		{"search.replace", "Find and Replace...", "F5.2", "Find text and confirm replacement.", "replace", 'R', []string{"Ctrl-R"}, false},
		{"search.forward", "Continue fwd", "F5.3/R19a", "Repeat the last search forwards.", "", 'C', []string{"Alt-C"}, false},
		{"search.reverse", "Continue rev", "F5.3/R19b", "Repeat the last search backwards.", "", 0, []string{"Ctrl-C"}, false},
		{"search.line", "Line number...", "F5.3/R19c", "Move to a specified document line.", "line", 'L', nil, true},
		{"search.bracket", "Matching bracket", "F5.3/R19d", "Find the partner of a bracket.", "", 0, []string{"Ctrl-]"}, false},
		{"search.differences", "Window Differences", "F5.3/R19e", "Locate differences between two windows.", "", 'D', nil, false},
		{"search.block", "Text in Block", "F5.3/R19f", "Use marked text for the next search.", "", 'B', nil, true},
		{"search.marker", "Block Marker", "F5.3/R19g", "Move to a block boundary from Search.", "", 'M', nil, false},
	}},
	{'P', "Print", []commandSpec{
		{"print.document", "Print", "F6.4", "Print the whole document.", "", 'P', nil, false},
		{"print.block", "Print Block", "F6.4", "Print only the marked block.", "", 'B', nil, false},
		{"print.eject", "Eject page", "F6.4", "Advance the printer to the next page; retained alongside PDF.", "", 'E', nil, false},
		{"print.setup", "Setup...", "F6.3", "Set margins, page length and formfeed policy.", "print", 'S', nil, true},
		{"print.pdf", "Export PDF... (WIB)", "F6.4", "Generate a PDF without replacing Eject page.", "file", 'X', nil, true},
	}},
	{'V', "View", []commandSpec{
		{"view.information", "Information display", "F5.4/R20a", "Inspect document and buffer information.", "information", 'I', nil, false},
		{"view.outline", "Outline display", "F5.4/R20b", "Inspect and navigate a compact document overview.", "", 'O', nil, false},
	}},
	{'O', "Options", []commandSpec{
		{"options.display", "Display...", "F7.1/R25", "Configure terminal display options.", "display", 'D', nil, false},
		{"options.mouse", "Mouse...", "F7.2/R26", "Configure mouse interaction.", "mouse", 'M', nil, false},
		{"options.editor", "Editor...", "F7.3/R27", "Configure insertion, tabs and backups.", "editor", 'E', nil, false},
		{"options.save", "Save configuration...", "F7.4/R28", "Save global or directory configuration.", "config", 'S', nil, true},
	}},
	// Manual is goBaDig's own addition, not part of the NE2 reference
	// this editor otherwise faithfully reproduces (see the comment
	// above neMenus) - it opens the Markdown/BASIC-dictionary/ASMSX
	// reference content (internal/man), unrelated to WibEdit's own
	// F1 "how do I use this editor" help (neHelpMenu, below).
	{'M', "Manual", []commandSpec{
		{"manual.open", "Abrir manual...", "goBaDig", "Abre o indice do manual de referencia do goBaDig (comandos, BASIC, ASMSX).", "", 'A', nil, false},
	}},
}

var neKeyCommands = []commandSpec{
	{id: "cursor.left", label: "Cursor left", phase: "F2.1", keys: []string{"Left"}},
	{id: "cursor.right", label: "Cursor right", phase: "F2.1", keys: []string{"Right"}},
	{id: "cursor.up", label: "Cursor up", phase: "F2.1", keys: []string{"Up"}},
	{id: "cursor.down", label: "Cursor down", phase: "F2.1", keys: []string{"Down"}},
	{id: "cursor.word-left", label: "Previous word", phase: "F2.1", keys: []string{"Ctrl-Left"}},
	{id: "cursor.word-right", label: "Next word", phase: "F2.1", keys: []string{"Ctrl-Right"}},
	{id: "cursor.home", label: "Start of line", phase: "F2.1", keys: []string{"Home"}},
	{id: "cursor.end", label: "End of line", phase: "F2.1", keys: []string{"End"}},
	{id: "cursor.page-up", label: "Page up", phase: "F2.1", keys: []string{"PgUp"}},
	{id: "cursor.page-down", label: "Page down", phase: "F2.1", keys: []string{"PgDn"}},
	{id: "cursor.screen-top", label: "Top of screen", phase: "F2.1", keys: []string{"Ctrl-PgUp"}},
	{id: "cursor.screen-bottom", label: "Bottom of screen", phase: "F2.1", keys: []string{"Ctrl-PgDn"}},
	{id: "cursor.file-top", label: "Start of file", phase: "F2.1", keys: []string{"Ctrl-Home"}},
	{id: "cursor.file-end", label: "End of file", phase: "F2.1", keys: []string{"Ctrl-End"}},
	{id: "cursor.scroll-up", label: "Scroll up", phase: "F2.1", keys: []string{"Alt-Z"}},
	{id: "cursor.scroll-down", label: "Scroll down", phase: "F2.1", keys: []string{"Ctrl-Z"}},
	{id: "edit.backspace", label: "Delete left", phase: "F2.2", keys: []string{"Backspace"}},
	{id: "edit.delete", label: "Delete right", phase: "F2.2", keys: []string{"Delete"}},
	{id: "edit.word-left", label: "Delete word left", phase: "F2.2", keys: []string{"Ctrl-W"}},
	{id: "edit.word-right", label: "Delete word right", phase: "F2.2", keys: []string{"Alt-W"}},
	{id: "edit.line-left", label: "Delete to start of line", phase: "F2.2", keys: []string{"Ctrl-L"}},
	{id: "edit.line-right", label: "Delete to end of line", phase: "F2.2", keys: []string{"Alt-L"}},
	{id: "edit.kill-line", label: "Delete line", phase: "F2.2", keys: []string{"Alt-K"}},
	{id: "edit.undelete", label: "Restore last deletion without cursor movement", phase: "F2.2", keys: []string{"Ctrl-U"}},
	{id: "edit.case-left", label: "Flip case towards line start", phase: "F2.2", keys: []string{"Ctrl-A"}},
	{id: "edit.case-right", label: "Flip case towards line end", phase: "F2.2", keys: []string{"Alt-A"}},
	{id: "edit.literal", label: "Insert next key literally", phase: "F2.2", keys: []string{"Ctrl-P"}},
	{id: "edit.indent-enter", label: "New line with indentation", phase: "F2.2", keys: []string{"Ctrl-Enter"}},
	{id: "edit.unindent", label: "Remove indentation level", phase: "F2.2", keys: []string{"Ctrl-Backspace", "Backtab"}},
	{id: "view.about", label: "About WibEdit", phase: "F5.4/R20c"},
	{id: "window.divider-up", label: "Move divider up", phase: "F2.4", keys: []string{"Ctrl-2"}},
	{id: "window.divider-down", label: "Move divider down", phase: "F2.4", keys: []string{"Alt-2"}},
	{id: "edit.insert", label: "Toggle Insert/Over", phase: "F2.2", keys: []string{"Insert"}},
	{id: "edit.enter", label: "New line", phase: "F2.2", keys: []string{"Enter"}},
	{id: "edit.tab", label: "Insert tab", phase: "F2.2", keys: []string{"Tab"}},
	{id: "help.index", label: "Help index", phase: "F1.4", keys: []string{"F1"}},
	{id: "menu.open", label: "Open File menu (terminal alternative to Alt alone)", phase: "F1.2", keys: []string{"F10"}},
	{id: "block.delete-direct", label: "Delete block without confirmation", phase: "F3.2"},
	{id: "block.mark-line", label: "Mark current line", phase: "F3.1"},
	{id: "block.mark-end", label: "Mark to end of line", phase: "F3.1"},
}

var neHelpMenu = menuSpec{'H', "F1=Help", []commandSpec{
	{id: "help.general", label: "General", phase: "F1.4", mnemonic: 'G'},
	{id: "help.keys", label: "Keys", phase: "F1.4", mnemonic: 'K'},
	{id: "help.file", label: "File", phase: "F1.4", mnemonic: 'F'},
	{id: "help.block", label: "Block", phase: "F1.4", mnemonic: 'B'},
	{id: "help.layout", label: "Text-layout", phase: "F1.4", mnemonic: 'T'},
	{id: "help.search", label: "Search", phase: "F1.4", mnemonic: 'S'},
	{id: "help.print", label: "Print", phase: "F1.4", mnemonic: 'P'},
	{id: "help.view", label: "View", phase: "F1.4", mnemonic: 'V'},
	{id: "help.options", label: "Options", phase: "F1.4", mnemonic: 'O'},
	{id: "help.about", label: "About WibEdit", phase: "F1.4", mnemonic: 'A', separator: true},
}}

func commands() []commandSpec {
	var result []commandSpec
	for _, menu := range neMenus {
		result = append(result, menu.items...)
	}
	result = append(result, neKeyCommands...)
	return append(result, neHelpMenu.items...)
}

func commandByID(id string) (commandSpec, bool) {
	for _, command := range commands() {
		if command.id == id {
			return command, true
		}
	}
	return commandSpec{}, false
}

func keyName(event *tcell.EventKey) string {
	base := tcell.KeyNames[event.Key()]
	switch event.Key() {
	case tcell.KeyRune:
		base = string(unicode.ToUpper(event.Rune()))
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		base = "Backspace"
	case tcell.KeyEnter:
		base = "Enter"
	case tcell.KeyTab:
		base = "Tab"
	}
	modifiers := event.Modifiers()
	if strings.HasPrefix(base, "Ctrl-") {
		base = strings.TrimPrefix(base, "Ctrl-")
		modifiers |= tcell.ModCtrl
	}
	prefix := ""
	if modifiers&tcell.ModCtrl != 0 {
		prefix += "Ctrl-"
	}
	if modifiers&tcell.ModAlt != 0 {
		prefix += "Alt-"
	}
	if modifiers&tcell.ModShift != 0 {
		prefix += "Shift-"
	}
	return prefix + base
}

func commandForKey(event *tcell.EventKey) (commandSpec, bool) {
	name := keyName(event)
	for _, command := range commands() {
		for _, binding := range command.keys {
			if binding == name {
				return command, true
			}
		}
	}
	return commandSpec{}, false
}

func (e *Editor) dispatch(command commandSpec) {
	e.notice = command.label + " - planned " + command.phase + " [" + command.id + "]"
	e.updateStatus()
}

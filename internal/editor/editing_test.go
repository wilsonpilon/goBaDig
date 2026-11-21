package edit

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestNEEditingDeletionAndRestore(t *testing.T) {
	for _, test := range []struct {
		id, text, want string
		at, after      documentPosition
	}{
		{"edit.word-right", "alpha beta gamma", " beta gamma", documentPosition{}, documentPosition{}},
		{"edit.word-right", "alpha, beta", "alpha", documentPosition{0, 5}, documentPosition{0, 5}},
		{"edit.word-left", "alpha beta gamma", "alpha beta ", documentPosition{0, 16}, documentPosition{0, 11}},
		{"edit.line-left", "alpha beta gamma", "beta gamma", documentPosition{0, 6}, documentPosition{}},
		{"edit.line-right", "alpha beta gamma", "alpha ", documentPosition{0, 6}, documentPosition{0, 6}},
		{"edit.delete", "left\r\nright", "leftright", documentPosition{0, 4}, documentPosition{0, 4}},
		{"edit.backspace", "left\r\nright", "leftright", documentPosition{1, 0}, documentPosition{0, 4}},
		{"edit.kill-line", "a\r\nb\r\nc", "a\r\nc", documentPosition{1, 1}, documentPosition{1, 0}},
		{"edit.kill-line", "a\r\nb", "a", documentPosition{1, 1}, documentPosition{}},
		{"edit.unindent", "    second", "second", documentPosition{0, 4}, documentPosition{}},
		{"edit.unindent", "        second", "    second", documentPosition{0, 4}, documentPosition{}},
		{"edit.unindent", "\t\tsecond", "\tsecond", documentPosition{0, 2}, documentPosition{0, 1}},
		{"edit.delete", "e\r\n\u0301", "e\u0301", documentPosition{0, 1}, documentPosition{0, 1}},
	} {
		t.Run(test.id+"/"+test.text, func(t *testing.T) {
			d := newDocument(test.text)
			d.setCursor(test.at, false)
			changed, notice, handled := d.edit(test.id, false)
			if !handled || !changed || notice != "" || d.text != test.want || d.cursor != test.after {
				t.Fatalf("delete: text=%q cursor=%+v changed=%v notice=%q", d.text, d.cursor, changed, notice)
			}
			changed, notice, _ = d.edit("edit.undelete", false)
			if !changed || notice != "" || d.text != test.text || d.cursor != test.at {
				t.Fatalf("restore: text=%q cursor=%+v notice=%q", d.text, d.cursor, notice)
			}
			if changed, notice, _ = d.edit("edit.undelete", false); changed || notice == "" {
				t.Fatal("restore must be one-shot with an explicit unavailable notice")
			}
		})
	}
	d := newDocument("abc")
	d.edit("edit.delete", false)
	d.navigate("cursor.right")
	d.navigate("cursor.left")
	if changed, notice, _ := d.edit("edit.undelete", false); changed || notice == "" || d.text != "bc" {
		t.Fatal("moving away and back must invalidate restoration")
	}
}

func TestNEEditingTabsEnterAndCasePreserveUnrelatedBytes(t *testing.T) {
	d := newDocument("a\tb\r\n    second")
	d.setCursor(documentPosition{1, 6}, false)
	d.edit("edit.indent-enter", true)
	if d.text != "a\tb\r\n    se\r\n    cond" || d.cursor != (documentPosition{2, 4}) {
		t.Fatalf("indented CRLF Enter: %q %+v", d.text, d.cursor)
	}
	d = newDocument("    text")
	d.edit("edit.unindent", false)
	if d.text != "    text" {
		t.Fatal("unindent at column one must not remove indentation after the cursor")
	}
	d = newDocument("abcdef\r\nnext")
	d.setCursor(documentPosition{0, 1}, false)
	d.tabMode = tabSpaces
	d.edit("edit.tab", true)
	if d.text != "a       \r\nnext" || d.column(d.cursor) != 8 {
		t.Fatal("space Tab in Over must replace cells without removing line separators")
	}
	d = newDocument("abcdefghijk")
	d.tabMode = tabMove
	d.edit("edit.tab", false)
	if d.text != "abcdefghijk" || d.cursor.character != 8 || d.revision != 0 {
		t.Fatal("move Tab inside a line must not change its bytes")
	}
	d = newDocument("abc")
	d.tabMode = tabMove
	d.setCursor(documentPosition{0, 3}, false)
	d.edit("edit.tab", false)
	if d.text != "abc     " || d.column(d.cursor) != 8 || d.revision != 1 {
		t.Fatal("original NE pads with spaces when move Tab goes past EOL")
	}
	d = newDocument("é\xffAb\r\n\x00last")
	d.setCursor(documentPosition{0, 2}, false)
	d.edit("edit.case-left", false)
	if d.text != "É\xffAb\r\n\x00last" || d.cursor.character != 2 {
		t.Fatal("case conversion corrupted an invalid UTF-8 byte or cursor")
	}
	d.edit("edit.case-right", false)
	if d.text != "É\xffaB\r\n\x00last" || d.cursor.character != 2 {
		t.Fatal("case conversion must only affect the requested line segment")
	}
}

func TestR13AutoIndentAndWordWrapEditPhysicalText(t *testing.T) {
	d := newDocument("  first")
	d.autoIndent = true
	d.setCursor(documentPosition{0, len(d.lines[0].cells)}, false)
	if changed, _, handled := d.edit("edit.enter", false); !handled || !changed || d.text != "  first\r\n  " {
		t.Fatalf("Auto Enter must copy leading indentation: %q", d.text)
	}
	if d.cursor != (documentPosition{1, 2}) {
		t.Fatalf("Auto Enter cursor = %+v, want the end of copied indentation", d.cursor)
	}

	d = newDocument("")
	d.lineLength = 8
	d.leftMargin = 2
	d.wordWrap = true
	for _, char := range "one two three" {
		d.insertText(string(char), false)
	}
	if d.text != "one two\r\n  three" {
		t.Fatalf("word wrap = %q, want a physical break with left margin", d.text)
	}
	if d.cursor != (documentPosition{1, 7}) {
		t.Fatalf("word wrap cursor = %+v, want to stay after typed text", d.cursor)
	}
}

func TestR13WordWrapDoesNotSplitLongWordsOrReflowWhenEnabled(t *testing.T) {
	d := newDocument("")
	d.lineLength = 4
	d.wordWrap = true
	for _, char := range "longword" {
		d.insertText(string(char), false)
	}
	if d.text != "longword" || d.cursor.line != 0 {
		t.Fatalf("word wrap split a word without a boundary: %q", d.text)
	}

	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("one two three", false)
	revision := editor.area.document.revision
	command, _ := commandByID("layout.wrap")
	editor.execute(command)
	if editor.area.GetText() != "one two three" || editor.area.document.revision != revision ||
		!editor.area.document.wordWrap {
		t.Fatal("enabling Word wrap must not reformat existing text")
	}
}

func TestNEEditingKeysLiteralAndSetup(t *testing.T) {
	closed := false
	editor, app, _ := newTestEditor(t, 80, 25, func() { closed = true })
	editor.area.SetText("alpha beta gamma", false)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModAlt))
	if editor.area.GetText() != " beta gamma" || !editor.changed {
		t.Fatal("Alt-W must execute real NE word deletion")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyCtrlU, 0, tcell.ModNone))
	if editor.area.GetText() != "alpha beta gamma" {
		t.Fatal("Ctrl-U must restore deletion, not inherit framework behavior")
	}
	editor.area.SetText("", false)
	for _, event := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyCtrlSpace, 0, tcell.ModCtrl),
		tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone),
	} {
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyCtrlP, 0, tcell.ModNone))
		sendEditorKey(editor, app, event)
	}
	if editor.area.GetText() != "\x1b\x00\r" || closed || editor.literal {
		t.Fatal("literal input must preserve ESC/NUL/CR and bypass global commands")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyCtrlP, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone))
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Literal input" {
		t.Fatal("unsupported literal keys must report a visible error")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	command, _ := commandByID("options.editor")
	editor.execute(command)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if editor.startMode != 0 {
		t.Fatal("cancel must not apply editor settings")
	}
	editor.execute(command)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if editor.startMode != 2 || editor.area.document.tabMode != tabMove {
		t.Fatal("editor setup must apply Never Over and move-Tab modes")
	}
	editor.area.SetText("abc", false)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyInsert, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'X', tcell.ModNone))
	if editor.over || editor.area.GetText() != "Xabc" {
		t.Fatal("Never Over must prevent destructive overwrite")
	}
	editor.literal = true
	editor.showHelp("edit.literal")
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	if editor.literal || editor.area.GetText() != "Xabc" {
		t.Fatal("opening a modal must cancel pending literal input")
	}
}

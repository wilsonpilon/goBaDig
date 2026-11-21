package edit

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestHelpTopicsAndPagination(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone))
	if !editor.menuOpen || editor.currentMenu != 'H' || editor.dropdown.GetItemCount() != 10 {
		t.Fatal("F1 must open the complete help index")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone))
	if len(editor.dialogs) != 1 || editor.dialogs[0].title != "General help" {
		t.Fatal("General help is missing")
	}
	d := editor.dialogs[0]
	drawEditor(editor, screen)
	var initial strings.Builder
	for row := 0; row < 25; row++ {
		initial.WriteString(editorScreenRow(screen, row))
	}
	if !strings.Contains(initial.String(), "Page 1/2 - More") {
		t.Fatal("help must show page number and More")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
	if d.helpPage != 1 {
		t.Fatal("N must select Next")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone))
	if d.helpPage != 0 {
		t.Fatal("P must select Previous")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	for _, entry := range neHelpMenu.items {
		editor.showHelp(entry.id)
		d = editor.dialogs[len(editor.dialogs)-1]
		for page := 0; page < len(d.helpPages); page++ {
			drawEditor(editor, screen)
			if d.helpPage != page {
				t.Fatalf("%s: page traversal skipped content", entry.id)
			}
			if page+1 < len(d.helpPages) {
				sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
			}
		}
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	}
	for _, command := range commands() {
		if strings.HasPrefix(command.id, "help.") {
			continue
		}
		_, pages, ok := editor.helpContent(command.id)
		if !ok {
			t.Fatalf("no contextual help for %s", command.id)
		}
		var text []string
		for _, page := range pages {
			text = append(text, page...)
		}
		if !strings.Contains(strings.Join(text, "\n"), command.id) {
			t.Fatalf("%s: help lost requirement traceability", command.id)
		}
	}
}

func TestKeysHelpStartsWithConciseSummary(t *testing.T) {
	editor, _, screen := newTestEditor(t, 80, 25, nil)
	title, pages, ok := editor.helpContent("help.keys")
	if !ok || title != "Help on keys" || len(pages) < 3 {
		t.Fatal("key help must include the three-page quick summary")
	}

	var summary strings.Builder
	for page := 0; page < 3; page++ {
		if len(pages[page]) != 14 {
			t.Fatalf("summary page %d has %d lines, want 14", page+1, len(pages[page]))
		}
		for _, line := range pages[page] {
			if len(line) > 60 {
				t.Errorf("summary page %d line is too wide (%d): %s", page+1, len(line), line)
			}
		}
		summary.WriteString(strings.Join(pages[page], "\n"))
		summary.WriteByte('\n')
	}
	for _, text := range []string{
		"Movement",
		"Editing",
		"Help and files",
		"Ctrl-Z: scroll the display down (not Undo).",
		"Ctrl-F/Ctrl-R: open previews; no search or replace runs.",
		"move the divider between open windows.",
	} {
		if !strings.Contains(summary.String(), text) {
			t.Errorf("key summary is missing %q", text)
		}
	}

	editor.showHelp("help.keys")
	drawEditor(editor, screen)
	var rendered strings.Builder
	for row := 0; row < 25; row++ {
		rendered.WriteString(editorScreenRow(screen, row))
	}
	if !strings.Contains(rendered.String(), "Movement") ||
		!strings.Contains(rendered.String(), "Ctrl-Z: scroll the display down") {
		t.Fatal("the first summary page must render its movement and scrolling keys")
	}
}

func TestContextHelpPreservesDialogAndMenu(t *testing.T) {
	editor, app, screen := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("document", true)
	_, start, end := editor.area.GetSelection()
	command, _ := commandByID("search.find")
	editor.execute(command)
	parent := editor.dialogs[0]
	for _, char := range "needle" {
		sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, char, tcell.ModNone))
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	focus := parent.focus
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone))
	if len(editor.dialogs) != 2 || editor.dialogs[1].context != "search.find" {
		t.Fatal("F1 must open command help above the current dialog")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if len(editor.dialogs) != 1 || app.GetFocus() != parent || parent.focus != focus ||
		parent.value("find") != "needle" || !parent.widgets[parent.fields[focus]].checked {
		t.Fatal("return from help lost input, choice or focus")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	editor.openMenu('F')
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	menu := editor.dropdown
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone))
	if len(editor.dialogs) != 1 || editor.dialogs[0].context != "file.open" {
		t.Fatal("menu F1 must describe the selected command")
	}
	drawEditor(editor, screen)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if !editor.menuOpen || editor.dropdown != menu || menu.selected != 1 || app.GetFocus() != menu {
		t.Fatal("return from help lost the dropdown selection")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	_, afterStart, afterEnd := editor.area.GetSelection()
	if editor.area.GetText() != "document" || start != afterStart || end != afterEnd {
		t.Fatal("help changed document or cursor")
	}
	editor.showHelp("file.exit")
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
	if len(editor.dialogs) != 1 || editor.dialogs[0].context != "file.exit" {
		t.Fatal("discard-dialog Y/N aliases must not leak into help")
	}
	screen.SetSize(30, 8)
	drawEditor(editor, screen)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone))
	if editor.dialogs[0].scroll == 0 {
		t.Fatal("small-screen help must allow scrolling otherwise hidden content")
	}
}

func TestErrorContextHelpKeepsUnsavedConfirmation(t *testing.T) {
	closed := false
	editor, app, _ := newTestEditor(t, 80, 25, func() { closed = true })
	editor.path = t.TempDir()
	editor.area.SetText("unsaved", true)
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone))
	if len(editor.dialogs) != 3 || editor.dialogs[2].context != "file.save" {
		t.Fatal("error F1 must preserve both notification and confirmation")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))
	if closed || len(editor.dialogs) != 3 || editor.area.GetText() != "unsaved" {
		t.Fatal("a key in nested help changed the unsaved document")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if closed || len(editor.dialogs) != 1 {
		t.Fatal("dismissing the error must preserve the unsaved confirmation")
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if closed || editor.area.GetText() != "unsaved" || app.GetFocus() != editor.area {
		t.Fatal("nested return changed unsaved text or final focus")
	}
}

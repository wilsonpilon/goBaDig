package edit

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestNEProfileDispatchDoesNotInheritFrameworkShortcuts(t *testing.T) {
	editor, app, _ := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("first\nsecond", true)
	_, start, end := editor.area.GetSelection()
	for _, test := range []struct {
		key tcell.Key
		id  string
	}{
		{tcell.KeyCtrlF, "search.find"},
		{tcell.KeyCtrlR, "search.replace"},
		{tcell.KeyCtrlC, "search.reverse"},
		{tcell.KeyCtrlRightSq, "search.bracket"},
	} {
		sendEditorKey(editor, app, tcell.NewEventKey(test.key, 0, tcell.ModNone))
		if test.id == "search.find" || test.id == "search.replace" {
			if len(editor.dialogs) != 1 {
				t.Fatalf("%v: %s did not open its dialog", test.key, test.id)
			}
		} else if test.id == "search.bracket" {
			if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Matching bracket" {
				t.Fatalf("%v: missing-bracket result was not reported", test.key)
			}
		} else if test.id == "search.differences" {
			if len(editor.dialogs) != 1 || editor.dialogs[0].title != "Window Differences" {
				t.Fatalf("%v: missing second-window result was not reported", test.key)
			}
		} else if !strings.Contains(editor.notice, test.id) {
			t.Errorf("%v: wrong dispatch: %q", test.key, editor.notice)
		}
		_, afterStart, afterEnd := editor.area.GetSelection()
		if editor.area.GetText() != "first\nsecond" || afterStart != start || afterEnd != end {
			t.Fatalf("%v inherited a conflicting TextArea action", test.key)
		}
		for len(editor.dialogs) > 0 {
			sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
		}
	}
	sendEditorKey(editor, app, tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModAlt))
	if !strings.Contains(editor.notice, "search.forward") {
		t.Fatal("Alt-C must dispatch continue forward, not a menu or copy")
	}
}

func TestCommandInventoryIsUnambiguous(t *testing.T) {
	ids, keys := map[string]bool{}, map[string]string{}
	for _, command := range commands() {
		if ids[command.id] || command.id == "" || command.phase == "" || command.label == "" {
			t.Fatalf("incomplete or duplicate requirement: %+v", command)
		}
		ids[command.id] = true
		for _, key := range command.keys {
			if previous, exists := keys[key]; exists {
				t.Fatalf("%s is bound to both %s and %s", key, previous, command.id)
			}
			keys[key] = command.id
		}
	}
	for _, event := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyCtrlF, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyCtrlF, 0, tcell.ModCtrl),
		tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModCtrl),
	} {
		command, ok := commandForKey(event)
		if !ok || command.id != "search.find" {
			t.Fatalf("control event not normalized: %q", event.Name())
		}
	}
}

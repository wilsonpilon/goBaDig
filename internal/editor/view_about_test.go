package edit

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wilsonpilon/goBaDig/internal/version"
)

func TestR20cViewAboutShowsProjectIdentityAndAuthorship(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.openMenu('V')
	editor.dropdown.InputHandler()(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone), func(_ tview.Primitive) {})

	if editor.menuOpen || len(editor.dialogs) != 1 || editor.dialogs[0].title != "About WibEdit" {
		t.Fatal("View > About did not open its About dialog")
	}
	content := editor.dialogs[0].widgets
	for _, want := range []string{
		"WibEdit " + version.Version + " (goBaDig)",
		"Created by " + editorCopyrightHolder + ".",
		"All rights reserved.",
		"certification is incomplete",
	} {
		found := false
		for _, widget := range content {
			found = found || strings.Contains(widget.label, want)
		}
		if !found {
			t.Errorf("About dialog missing %q", want)
		}
	}

	editor.dialogs[0].key(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if len(editor.dialogs) != 0 || editor.app.GetFocus() != editor.area {
		t.Fatal("closing About did not return to the editor")
	}
}

func TestR20cHelpAboutUsesCurrentVersionAndAttribution(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	_, pages, ok := editor.helpContent("help.about")
	if !ok {
		t.Fatal("About help is unavailable")
	}
	content := strings.Join(pages[0], "\n")
	for _, want := range []string{
		"WibEdit " + version.Version + " (goBaDig)",
		"Created by " + editorCopyrightHolder,
		"certification is incomplete",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("About help missing %q in:\n%s", want, content)
		}
	}
}

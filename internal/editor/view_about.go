package edit

import (
	"fmt"

	"github.com/wilsonpilon/goBaDig/internal/version"
)

const editorCopyrightHolder = "Wilson Pilon / WIB Projetos"

func (e *Editor) aboutDisplay() {
	d := newDialog("About WibEdit", 66, 14)
	d.context = "view.about"
	lines := []string{
		"WibEdit " + version.Version + " (goBaDig)",
		"Independent source and plain-text editor, implemented in Go.",
		"Built into goBaDig, the Basic Dignified Suite.",
		"Created by " + editorCopyrightHolder + ".",
		"All rights reserved. See the goBaDig LICENSE.",
		"Norton Editor 2 is a compatibility target; certification is incomplete.",
	}
	for row, line := range lines {
		d.label(2, row+1, line)
	}
	d.buttons([]string{"ok", "help"}, []string{"OK", "F1=Help"}, "ok")
	d.result = func(id string) {
		if id == "help" {
			e.showHelp("view.about")
			return
		}
		e.closeDialog()
	}
	e.openDialog(d)
}

func (e *Editor) aboutSummary() []string {
	return []string{
		"WibEdit " + version.Version + " (goBaDig)",
		"Created by " + editorCopyrightHolder + ".",
		fmt.Sprintf("Compatibility target: %s; certification is incomplete.", compatibilityProfile),
	}
}

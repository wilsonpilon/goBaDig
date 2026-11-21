package edit

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestR25DisplayAppliesSupportedSettingsAndReportsUnsupportedChoices(t *testing.T) {
	editor, _, screen := newTestEditor(t, 80, 25, nil)
	editor.area.SetText("a\tb", false)
	command, ok := commandByID("options.display")
	if !ok || !editor.available(command) {
		t.Fatal("Display setup command is unavailable")
	}

	editor.execute(command)
	dialog := editor.dialogs[0]
	dialogWidgetByID(t, dialog, "color").selected = 1
	dialogWidgetByID(t, dialog, "lines").selected = 2
	dialogWidgetByID(t, dialog, "scrollbar").checked = false
	dialogWidgetByID(t, dialog, "tab-spacing").value = "4"
	dialogWidgetByID(t, dialog, "blink").checked = true
	dialog.activate("ok")

	if editor.colorProfile != 1 || editor.displayScrollbar || editor.area.scrollbar {
		t.Fatal("supported color and scrollbar options were not applied")
	}
	if editor.area.document.tabWidth != 4 || editor.area.document.lines[0].cells[1].width != 3 {
		t.Fatal("tab spacing was not applied to the document view")
	}
	if !strings.Contains(editor.notice, "Not applied: terminal height, blinking bracket") {
		t.Fatalf("unsupported display settings were not identified: %q", editor.notice)
	}
	if panelStyle != tcell.StyleDefault.Foreground(tcell.NewRGBColor(0, 0, 0)).Background(tcell.NewRGBColor(170, 170, 170)) {
		t.Fatal("selected Color 2 profile was not applied to the panel")
	}

	drawEditor(editor, screen)
	x, y, width, _ := editor.area.GetInnerRect()
	lastColumn, _, _, _ := screen.GetContent(x+width-1, y)
	if lastColumn != ' ' {
		t.Fatal("hidden scrollbar was still drawn")
	}
}

func TestR25DisplayPalettesMatchReferencesAndApplyToDialogs(t *testing.T) {
	black := tcell.NewRGBColor(0, 0, 0)
	white := tcell.NewRGBColor(255, 255, 255)
	blue := tcell.NewRGBColor(0, 0, 170)
	cyan := tcell.NewRGBColor(0, 170, 170)
	gray := tcell.NewRGBColor(170, 170, 170)
	darkGray := tcell.NewRGBColor(85, 85, 85)
	brown := tcell.NewRGBColor(170, 85, 0)
	red := tcell.NewRGBColor(170, 0, 0)
	yellow := tcell.NewRGBColor(255, 255, 85)
	profiles := []struct {
		name               string
		index              int
		panelBackground    tcell.Color
		menuForeground     tcell.Color
		contentForeground  tcell.Color
		contentBackground  tcell.Color
		selectedForeground tcell.Color
		mnemonicForeground tcell.Color
		pendingForeground  tcell.Color
		dialogForeground   tcell.Color
		titleForeground    tcell.Color
		shadowBackground   tcell.Color
		blockForeground    tcell.Color
		blockBackground    tcell.Color
		thumbColor         tcell.Color
	}{
		{"Color 1", 0, cyan, white, white, blue, white, yellow, darkGray, white, yellow, black, blue, gray, red},
		{"Color 2", 1, gray, black, white, blue, gray, red, black, black, black, black, blue, gray, black},
		{"VGA Monochrome", 2, gray, black, gray, black, gray, white, brown, black, black, blue, black, brown, black},
		{"Monochrome", 3, gray, black, gray, black, gray, white, black, black, black, darkGray, black, gray, black},
		{"LCD", 4, gray, black, gray, black, gray, tcell.ColorDefault, black, black, black, darkGray, black, gray, black},
	}

	for _, profile := range profiles {
		t.Run(profile.name, func(t *testing.T) {
			editor, _, screen := newTestEditor(t, 80, 25, nil)
			editor.area.SetText("sample text", false)
			editor.changed = false
			document := editor.area.document
			revision, cursor := document.revision, document.cursor
			if profile.index != 0 {
				command, _ := commandByID("options.display")
				editor.execute(command)
				dialogWidgetByID(t, editor.dialogs[0], "color").selected = profile.index
				editor.dialogs[0].activate("ok")
			}
			drawEditor(editor, screen)

			_, _, menuStyle, _ := screen.GetContent(1, 0)
			assertStyleColors(t, menuStyle, black, profile.panelBackground, "menu bar")
			x, y, areaWidth, _ := editor.area.GetInnerRect()
			_, _, contentStyleAtScreen, _ := screen.GetContent(x, y)
			assertStyleColors(t, contentStyleAtScreen, profile.contentForeground, profile.contentBackground, "document background")
			_, _, statusStyle, _ := screen.GetContent(3, 24)
			assertStyleColors(t, statusStyle, black, profile.panelBackground, "status bar")
			_, _, thumbStyle, _ := screen.GetContent(x+areaWidth-1, y+1)
			assertStyleColors(t, thumbStyle, profile.thumbColor, profile.panelBackground, "scrollbar thumb")
			document.setBlockRange(0, 6)
			drawEditor(editor, screen)
			_, _, markedStyle, _ := screen.GetContent(x+2, y)
			assertStyleColors(t, markedStyle, profile.blockForeground, profile.blockBackground, "marked text")
			_, _, markerStyleAtScreen, _ := screen.GetContent(x, y)
			assertStyleColors(t, markerStyleAtScreen, profile.blockForeground, profile.blockBackground, "block marker")
			document.clearBlockMarkers()

			editor.openMenu('F')
			drawEditor(editor, screen)
			menuX, menuY, menuWidth, _ := editor.dropdown.GetRect()
			_, _, selectedRow, _ := screen.GetContent(menuX+3, menuY+1)
			assertStyleColors(t, selectedRow, profile.selectedForeground, black, "selected menu item")
			_, _, menuText, _ := screen.GetContent(menuX+3, menuY+2)
			assertStyleColors(t, menuText, profile.menuForeground, profile.panelBackground, "menu item")
			for _, row := range []int{1, 2} {
				_, _, mnemonicStyle, _ := screen.GetContent(menuX+2, menuY+row)
				foreground, background := profile.menuForeground, profile.panelBackground
				if row == 1 {
					foreground, background = profile.selectedForeground, black
				}
				expected := tcell.StyleDefault.Foreground(profile.mnemonicForeground).Background(background)
				if profile.mnemonicForeground == tcell.ColorDefault {
					expected = tcell.StyleDefault.Foreground(foreground).Background(background).Reverse(true)
				}
				if mnemonicStyle != expected {
					t.Fatalf("menu mnemonic row %d = %v, want %v", row, mnemonicStyle, expected)
				}
			}
			for row, index := range editor.dropdown.rows {
				if index < 0 || editor.available(editor.dropdown.items[index]) {
					continue
				}
				_, _, disabledStyle, _ := screen.GetContent(menuX+2, menuY+row+1)
				assertStyleColors(t, disabledStyle, profile.pendingForeground, profile.panelBackground, "disabled menu item")
				editor.dropdown.selected = index
				drawEditor(editor, screen)
				_, _, disabledStyle, _ = screen.GetContent(menuX+2, menuY+row+1)
				assertStyleColors(t, disabledStyle, profile.pendingForeground, black, "selected disabled menu item")
				break
			}
			shadowRune, _, shadowStyle, _ := screen.GetContent(menuX+menuWidth, menuY+1)
			assertStyleColors(t, shadowStyle, profile.shadowBackground, profile.shadowBackground, "menu shadow")
			if shadowRune != ' ' {
				t.Fatal("panel shadows must be solid, not stippled")
			}
			editor.closeMenu()

			dialog := newDialog("Palette", 50, 12)
			dialog.label(2, 1, "Label")
			dialog.add(dialogWidget{kind: widgetInput, id: "input", value: "ab", x: 2, y: 2, width: 12})
			dialog.add(dialogWidget{kind: widgetCheck, label: "Check", x: 2, y: 4, width: 15})
			dialog.add(dialogWidget{kind: widgetRadio, choices: []string{"One", "Two"}, x: 2, y: 5, width: 15, height: 2})
			dialog.add(dialogWidget{kind: widgetList, choices: []string{"One", "Two"}, x: 25, y: 2, width: 15, height: 4})
			dialog.buttons([]string{"ok", "cancel"}, []string{"OK", "Cancel"}, "ok")
			editor.openDialog(dialog)
			drawEditor(editor, screen)
			dialogX, dialogY, dialogWidth, dialogHeight := dialog.GetRect()
			_, _, borderStyle, _ := screen.GetContent(dialogX, dialogY)
			assertStyleColors(t, borderStyle, profile.dialogForeground, profile.panelBackground, "dialog border")
			_, _, bodyStyle, _ := screen.GetContent(dialogX+1, dialogY+1)
			assertStyleColors(t, bodyStyle, profile.dialogForeground, profile.panelBackground, "dialog background")
			_, _, titleStyle, _ := screen.GetContent(dialogX+(dialogWidth-len(dialog.title)-2)/2+1, dialogY)
			assertStyleColors(t, titleStyle, profile.titleForeground, profile.panelBackground, "dialog title")
			_, _, inputStyle, _ := screen.GetContent(dialogX+3, dialogY+2)
			assertStyleColors(t, inputStyle, profile.selectedForeground, black, "fresh input selection")
			_, _, inputPadding, _ := screen.GetContent(dialogX+6, dialogY+2)
			assertStyleColors(t, inputPadding, profile.dialogForeground, profile.panelBackground, "input padding")
			dialog.key(tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone))
			drawEditor(editor, screen)
			_, _, inputStyle, _ = screen.GetContent(dialogX+3, dialogY+2)
			assertStyleColors(t, inputStyle, profile.dialogForeground, profile.panelBackground, "editing input")
			for focus, row := range map[int]int{1: 4, 2: 5} {
				dialog.focus = focus
				drawEditor(editor, screen)
				_, _, fieldStyle, _ := screen.GetContent(dialogX+2, dialogY+row)
				assertStyleColors(t, fieldStyle, profile.dialogForeground, profile.panelBackground, "focused check/radio")
				if _, _, visible := screen.GetCursor(); !visible {
					t.Fatal("check/radio focus must retain a visible cursor")
				}
			}
			dialog.focus = 3
			drawEditor(editor, screen)
			_, _, listStyle, _ := screen.GetContent(dialogX+25, dialogY+3)
			assertStyleColors(t, listStyle, profile.selectedForeground, black, "focused list row")
			for _, focus := range []int{3, 4} {
				dialog.focus = focus
				drawEditor(editor, screen)
				_, _, buttonStyle, _ := screen.GetContent(dialogX+4, dialogY+dialogHeight-2)
				assertStyleColors(t, buttonStyle, profile.selectedForeground, black, "normal/focused button")
			}
			editor.closeDialog()
			if document.text != "sample text" || document.revision != revision || document.cursor != cursor || editor.changed {
				t.Fatal("display palette or drawing changed the document")
			}
		})
	}
}

func assertStyleColors(t *testing.T, style tcell.Style, wantForeground, wantBackground tcell.Color, location string) {
	t.Helper()
	foreground, background, _ := style.Decompose()
	if foreground != wantForeground || background != wantBackground {
		t.Fatalf("%s colors = (%v, %v), want (%v, %v)", location, foreground, background, wantForeground, wantBackground)
	}
}

func TestR25DisplayCancelDoesNotApplySettings(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	command, _ := commandByID("options.display")
	editor.execute(command)
	dialog := editor.dialogs[0]
	dialogWidgetByID(t, dialog, "color").selected = 3
	dialogWidgetByID(t, dialog, "tab-spacing").value = "4"
	dialog.activate("cancel")

	if editor.colorProfile != 0 || !editor.displayScrollbar || editor.area.document.tabWidth != 8 {
		t.Fatal("cancel applied display settings")
	}
}

func dialogWidgetByID(t *testing.T, dialog *editorDialog, id string) *dialogWidget {
	t.Helper()
	for index := range dialog.widgets {
		if dialog.widgets[index].id == id {
			return &dialog.widgets[index]
		}
	}
	t.Fatalf("dialog is missing widget %q", id)
	return nil
}

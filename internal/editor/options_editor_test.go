package edit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestR27EditorSetupAppliesBackupAndEditingSettingsToOpenWindows(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	first := editor.area
	editor.openSecondWindow("", "", false)
	second := editor.area
	command, ok := commandByID("options.editor")
	if !ok || !editor.available(command) {
		t.Fatal("Editor setup command is unavailable")
	}

	editor.execute(command)
	dialog := editor.dialogs[0]
	dialogWidgetByID(t, dialog, "insert-mode").selected = 1
	dialogWidgetByID(t, dialog, "tab-mode").selected = int(tabSpaces)
	dialogWidgetByID(t, dialog, "backup-character").value = "!"
	dialogWidgetByID(t, dialog, "backups").checked = false
	dialog.activate("ok")

	for _, view := range []*documentView{first, second} {
		if view.document.tabMode != tabSpaces || !view.over {
			t.Fatal("editor mode and Tab settings were not applied to each window")
		}
	}
	if editor.startMode != 1 || editor.backupCharacter != '!' || editor.backupsEnabled {
		t.Fatal("Editor setup did not apply insertion or backup settings")
	}

	path := filepath.Join(t.TempDir(), "backup.prg")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := saveEditorFile(path, "new", editor.backupCharacterForSave(), nil, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "backup.!PR")); !os.IsNotExist(err) {
		t.Fatalf("disabled backups left a new backup file: %v", err)
	}
	editor.backupsEnabled = true
	if err := saveEditorFile(path, "newer", editor.backupCharacterForSave(), nil, true); err != nil {
		t.Fatal(err)
	}
	assertFileBytes(t, filepath.Join(filepath.Dir(path), "backup.!PR"), []byte("new"))
}

func TestR27EditorSetupRejectsUnsafeBackupCharacter(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	command, _ := commandByID("options.editor")
	editor.execute(command)
	dialog := editor.dialogs[0]
	dialogWidgetByID(t, dialog, "backup-character").value = `\`
	dialog.activate("ok")

	if editor.backupCharacter != '{' || !editor.backupsEnabled ||
		len(editor.dialogs) != 2 || editor.dialogs[1].title != "Invalid backup character" {
		t.Fatal("unsafe backup character was applied or not reported")
	}
}

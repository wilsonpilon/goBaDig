package edit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestR28DirectoryConfigurationOverridesGlobal(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "global.json")
	directoryPath := filepath.Join(dir, editorDirectoryConfigName)
	global := defaultEditorConfig()
	local := global
	local.ColorProfile = 4
	local.TabWidth = 4
	if err := writeEditorConfig(globalPath, global); err != nil {
		t.Fatal(err)
	}
	if err := writeEditorConfig(directoryPath, local); err != nil {
		t.Fatal(err)
	}

	got, path, err := loadEditorConfiguration(globalPath, directoryPath)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || *got != local || path != directoryPath {
		t.Fatalf("loaded configuration = %#v at %q, want directory configuration at %q", got, path, directoryPath)
	}

	if err := os.Remove(directoryPath); err != nil {
		t.Fatal(err)
	}
	got, path, err = loadEditorConfiguration(globalPath, directoryPath)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || *got != global || path != globalPath {
		t.Fatalf("fallback configuration = %#v at %q, want global configuration at %q", got, path, globalPath)
	}
}

func TestR28ConfigurationRejectsMalformedAndUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wibedit.json")
	data, err := json.Marshal(defaultEditorConfig())
	if err != nil {
		t.Fatal(err)
	}
	data = append(data[:len(data)-1], []byte(`,"unexpected":true}`)...)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readEditorConfig(path); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown-field configuration error = %v, want explicit error", err)
	}

	if err := os.WriteFile(path, []byte(`{"version":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readEditorConfig(path); err == nil || !strings.Contains(err.Error(), "missing required field") {
		t.Fatalf("incomplete configuration error = %v, want explicit missing-field error", err)
	}

	configuration := defaultEditorConfig()
	configuration.TabWidth = 0
	if err := writeEditorConfig(path, configuration); err == nil {
		t.Fatal("invalid configuration was written")
	}
}

func TestR28SaveConfigurationDialogWritesSelectedScope(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	globalPath := filepath.Join(t.TempDir(), "global", "wibedit.json")
	editor.configPath = globalPath
	editor.configChanged = true
	command, _ := commandByID("options.save")
	editor.execute(command)
	editor.dialogs[0].activate("ok")

	got, err := readEditorConfig(globalPath)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || *got != editor.currentConfiguration() || editor.configChanged {
		t.Fatal("current configuration was not saved and marked clean")
	}

	directory := t.TempDir()
	t.Chdir(directory)
	editor.configChanged = true
	editor.execute(command)
	editor.dialogs[0].widgets[editor.dialogs[0].fields[0]].selected = 1
	editor.dialogs[0].activate("ok")
	directoryPath := filepath.Join(directory, editorDirectoryConfigName)
	got, err = readEditorConfig(directoryPath)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || *got != editor.currentConfiguration() || editor.configPath != directoryPath || editor.configChanged {
		t.Fatal("directory configuration was not saved and selected as current")
	}
}

func TestR28ExitConfigurationPromptSaveDiscardAndCancel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	closed := false
	editor, _, _ := newTestEditor(t, 80, 25, func() { closed = true })
	editor.configPath = path
	editor.configChanged = true

	editor.requestClose()
	editor.dialogs[0].activate("cancel")
	if closed || !editor.configChanged {
		t.Fatal("canceling the configuration prompt changed exit state")
	}

	editor.requestClose()
	editor.dialogs[0].activate("save")
	if !closed || editor.configChanged {
		t.Fatal("saving configuration did not continue the exit")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("exit configuration was not saved: %v", err)
	}

	discarded, _, _ := newTestEditor(t, 80, 25, nil)
	discarded.configChanged = true
	discarded.requestClose()
	discarded.dialogs[0].activate("discard")
	if discarded.configChanged {
		t.Fatal("discarding configuration did not clear its dirty state on exit")
	}
}

func TestR28ConfigurationWriteFailureKeepsDirtyState(t *testing.T) {
	editor, _, _ := newTestEditor(t, 80, 25, nil)
	editor.configChanged = true
	path := t.TempDir()
	if err := editor.saveConfiguration(path); err == nil {
		t.Fatal("saving over a directory unexpectedly succeeded")
	}
	if !editor.configChanged {
		t.Fatal("failed configuration save cleared its dirty state")
	}
}

func defaultEditorConfig() editorConfig {
	return editorConfig{
		Version: editorConfigVersion, ColorProfile: 0, DisplayScrollbar: true, TabWidth: 8,
		MouseButtonRight: false, MouseScrollDelayMS: 0, StartMode: 0, TabMode: int(tabInsert),
		BackupCharacter: "{", BackupsEnabled: true,
	}
}

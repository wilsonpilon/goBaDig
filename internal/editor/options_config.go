package edit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const editorConfigVersion = 1
const editorDirectoryConfigName = ".wibedit.json"

type editorConfig struct {
	Version            int    `json:"version"`
	ColorProfile       int    `json:"colorProfile"`
	DisplayScrollbar   bool   `json:"displayScrollbar"`
	TabWidth           int    `json:"tabWidth"`
	MouseButtonRight   bool   `json:"mouseButtonRight"`
	MouseScrollDelayMS int    `json:"mouseScrollDelayMs"`
	StartMode          int    `json:"startMode"`
	TabMode            int    `json:"tabMode"`
	BackupCharacter    string `json:"backupCharacter"`
	BackupsEnabled     bool   `json:"backupsEnabled"`
}

func (e *Editor) currentConfiguration() editorConfig {
	return editorConfig{
		Version:            editorConfigVersion,
		ColorProfile:       e.colorProfile,
		DisplayScrollbar:   e.displayScrollbar,
		TabWidth:           e.area.document.tabWidth,
		MouseButtonRight:   e.mouseButtonRight,
		MouseScrollDelayMS: int(e.mouseScrollDelay / time.Millisecond),
		StartMode:          e.startMode,
		TabMode:            int(e.area.document.tabMode),
		BackupCharacter:    string(e.backupCharacter),
		BackupsEnabled:     e.backupsEnabled,
	}
}

func (e *Editor) applyConfiguration(configuration editorConfig) {
	e.applyDisplayPalette(configuration.ColorProfile)
	e.displayScrollbar = configuration.DisplayScrollbar
	e.mouseButtonRight = configuration.MouseButtonRight
	e.mouseScrollDelay = time.Duration(configuration.MouseScrollDelayMS) * time.Millisecond
	e.startMode = configuration.StartMode
	e.backupCharacter, _ = parseBackupCharacter(configuration.BackupCharacter)
	e.backupsEnabled = configuration.BackupsEnabled
	for _, window := range e.windows {
		window.area.scrollbar = e.displayScrollbar
		window.area.setMouseOptions(e.mouseButtonRight, e.mouseScrollDelay)
		window.startMode = e.startMode
		window.over = e.startMode == 1
		window.area.over = window.over
		window.area.document.tabWidth = configuration.TabWidth
		window.area.document.tabMode = tabKeyMode(configuration.TabMode)
		window.area.document.indexLines()
	}
	if len(e.windows) > 0 {
		e.over = e.windows[e.activeWindow].over
	}
	e.updateStatus()
}

func (e *Editor) loadStartupConfiguration() error {
	globalPath, err := globalEditorConfigPath()
	if err != nil {
		return err
	}
	directory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("finding current directory for editor configuration: %w", err)
	}
	directoryPath := filepath.Join(directory, editorDirectoryConfigName)
	configuration, path, err := loadEditorConfiguration(globalPath, directoryPath)
	if err != nil {
		return err
	}
	e.configPath = path
	if configuration != nil {
		e.applyConfiguration(*configuration)
	}
	e.configChanged = false
	return nil
}

func loadEditorConfiguration(globalPath, directoryPath string) (*editorConfig, string, error) {
	global, err := readEditorConfig(globalPath)
	if err != nil {
		return nil, "", err
	}
	directory, err := readEditorConfig(directoryPath)
	if err != nil {
		return nil, "", err
	}
	if directory != nil {
		return directory, directoryPath, nil
	}
	if global != nil {
		return global, globalPath, nil
	}
	return nil, globalPath, nil
}

func globalEditorConfigPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("finding user configuration directory: %w", err)
	}
	return filepath.Join(directory, "goBaDig", "wibedit.json"), nil
}

func readEditorConfig(path string) (*editorConfig, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading editor configuration %s: %w", path, err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("decoding editor configuration %s: %w", path, err)
	}
	for _, key := range []string{
		"version", "colorProfile", "displayScrollbar", "tabWidth",
		"mouseButtonRight", "mouseScrollDelayMs", "startMode", "tabMode",
		"backupCharacter", "backupsEnabled",
	} {
		if value, ok := fields[key]; !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, fmt.Errorf("decoding editor configuration %s: missing required field %q", path, key)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var configuration editorConfig
	if err := decoder.Decode(&configuration); err != nil {
		return nil, fmt.Errorf("decoding editor configuration %s: %w", path, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = errors.New("configuration contains more than one JSON value")
		}
		return nil, fmt.Errorf("decoding editor configuration %s: %w", path, err)
	}
	if err := validateEditorConfig(configuration); err != nil {
		return nil, fmt.Errorf("invalid editor configuration %s: %w", path, err)
	}
	return &configuration, nil
}

func validateEditorConfig(configuration editorConfig) error {
	if configuration.Version != editorConfigVersion {
		return fmt.Errorf("unsupported version %d", configuration.Version)
	}
	if configuration.ColorProfile < 0 || configuration.ColorProfile > 4 {
		return errors.New("color profile must be from 0 to 4")
	}
	if configuration.TabWidth < 1 || configuration.TabWidth > 32 {
		return errors.New("tab width must be from 1 to 32")
	}
	switch configuration.MouseScrollDelayMS {
	case 0, 12, 18, 26:
	default:
		return errors.New("mouse scroll delay must be 0, 12, 18 or 26 milliseconds")
	}
	if configuration.StartMode < 0 || configuration.StartMode > 2 {
		return errors.New("start mode must be from 0 to 2")
	}
	if configuration.TabMode < int(tabInsert) || configuration.TabMode > int(tabMove) {
		return errors.New("Tab mode is invalid")
	}
	if _, err := parseBackupCharacter(configuration.BackupCharacter); err != nil {
		return err
	}
	return nil
}

func (e *Editor) saveConfiguration(path string) error {
	if path == "" {
		return errors.New("editor configuration path is empty")
	}
	if err := validateEditorConfig(e.currentConfiguration()); err != nil {
		return fmt.Errorf("cannot save editor configuration: %w", err)
	}
	if err := writeEditorConfig(path, e.currentConfiguration()); err != nil {
		return err
	}
	e.configPath = path
	e.configChanged = false
	return nil
}

func writeEditorConfig(path string, configuration editorConfig) error {
	if err := validateEditorConfig(configuration); err != nil {
		return fmt.Errorf("cannot write editor configuration: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating editor configuration directory: %w", err)
	}
	data, err := json.MarshalIndent(configuration, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding editor configuration: %w", err)
	}
	if err := atomicWriteFile(path, string(data)+"\n", 0600, nil); err != nil {
		return fmt.Errorf("saving editor configuration %s: %w", path, err)
	}
	return nil
}

func (e *Editor) saveConfigurationDialog() {
	d := newDialog("Save configuration", 62, 10)
	d.context = "options.save"
	d.label(2, 2, "Update:")
	d.add(dialogWidget{
		kind: widgetRadio, id: "scope",
		choices: []string{"current config file", "config for this directory"},
		x:       2, y: 3, width: 56, height: 2,
	})
	d.buttons([]string{"ok", "help", "cancel"}, []string{"OK", "F1=Help", "Cancel"}, "ok")
	d.result = func(id string) {
		if id == "help" {
			e.showHelp(d.context)
			return
		}
		if id != "ok" {
			e.closeDialog()
			return
		}
		selectedScope := d.widgets[d.fields[0]].selected
		path := e.configPath
		if selectedScope == 1 {
			directory, err := os.Getwd()
			if err != nil {
				e.showMessage("Configuration error", fmt.Sprintf("Finding current directory: %v", err))
				return
			}
			path = filepath.Join(directory, editorDirectoryConfigName)
		} else if path == "" {
			var err error
			path, err = globalEditorConfigPath()
			if err != nil {
				e.showMessage("Configuration error", err.Error())
				return
			}
		}
		e.closeDialog()
		if err := e.saveConfiguration(path); err != nil {
			e.showMessage("Configuration error", err.Error())
		} else {
			e.notice = "Saved editor configuration to " + path
			e.updateStatus()
		}
	}
	e.openDialog(d)
}

func (e *Editor) confirmConfigurationExit() {
	d := newDialog("Configuration changed", 64, 8)
	d.label(2, 2, "Save the changed editor configuration before exiting?")
	d.buttons([]string{"save", "discard", "cancel"}, []string{"Save", "Don't Save", "Cancel"}, "save")
	d.result = func(id string) {
		e.closeDialog()
		switch id {
		case "save":
			if e.configPath == "" {
				path, err := globalEditorConfigPath()
				if err != nil {
					e.showMessage("Configuration error", err.Error())
					return
				}
				e.configPath = path
			}
			if err := e.saveConfiguration(e.configPath); err != nil {
				e.showMessage("Configuration error", err.Error())
				return
			}
			e.confirmCloseDocuments(false)
		case "discard":
			e.confirmCloseDocuments(true)
		}
	}
	e.openDialog(d)
}

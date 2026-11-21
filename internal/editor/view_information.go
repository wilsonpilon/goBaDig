package edit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

func (e *Editor) informationDisplay() {
	lines, err := e.informationLines()
	if err != nil {
		e.fileError(err, "view.information")
		return
	}
	e.showMessage("Information display", strings.Join(lines, "\n"))
}

func (e *Editor) informationLines() ([]string, error) {
	text, err := e.currentFileText()
	if err != nil {
		return nil, err
	}
	directory := filepath.Dir(e.path)
	if e.path == "" {
		directory, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getting the current directory: %w", err)
		}
	}
	freeBytes, err := editorAvailableSpace(directory)
	if err != nil {
		return nil, fmt.Errorf("getting available space for %q: %w", directory, err)
	}

	d := e.area.document
	cursorOffset := d.byteOffset(d.cursor)
	cursorPrefix := d.text[:cursorOffset]
	if e.diskKnown && e.path != "" {
		if e.fileBase < 0 || e.fileBase > len(e.diskText) {
			return nil, errors.New("invalid partial-file boundary; information is unavailable")
		}
		cursorPrefix = e.diskText[:e.fileBase] + cursorPrefix
	}
	cursorLine := strings.Count(cursorPrefix, "\n") + 1
	cursorColumn := d.column(d.cursor) + 1

	modified := "Not saved"
	if e.path != "" {
		info, statErr := os.Stat(e.path)
		switch {
		case statErr == nil:
			modified = info.ModTime().Format(time.RFC3339)
		case errors.Is(statErr, os.ErrNotExist):
			modified = "Not on disk"
		default:
			return nil, fmt.Errorf("reading file information for %q: %w", e.path, statErr)
		}
	}

	character := "None at cursor"
	line := d.lines[d.cursor.line]
	if d.cursor.character < len(line.cells) {
		cell := line.cells[d.cursor.character]
		raw := d.text[cell.start:cell.end]
		r, size := utf8.DecodeRuneInString(raw)
		if r == utf8.RuneError && size == 1 && !utf8.ValidString(raw) {
			character = fmt.Sprintf("Invalid UTF-8 byte 0x%02X", raw[0])
		} else {
			character = fmt.Sprintf("%q (U+%04X, %d)", r, r, r)
		}
	}

	name := e.filename
	if name == "" {
		name = "Untitled"
	}
	return []string{
		"File: " + name,
		"Path: " + displayPath(e.path),
		"Directory: " + directory,
		fmt.Sprintf("Size: %d bytes (%s)", len(text), formatInformationSize(uint64(len(text)))),
		fmt.Sprintf("Lines: %d  Words: %d  Characters: %d",
			strings.Count(text, "\n")+1, len(strings.Fields(text)), utf8.RuneCountInString(text)),
		fmt.Sprintf("In buffer: %d bytes  Cursor: line %d, column %d",
			len(d.text), cursorLine, cursorColumn),
		"On-disk modified: " + modified,
		"System time: " + time.Now().Format(time.RFC3339),
		"Free space available: " + formatInformationSize(freeBytes),
		"Character at cursor: " + character,
	}, nil
}

func displayPath(path string) string {
	if path == "" {
		return "Not saved"
	}
	return path
}

func formatInformationSize(size uint64) string {
	const base = uint64(1024)
	units := [...]string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	value := float64(size)
	unit := 0
	for value >= float64(base) && unit < len(units)-1 {
		value /= float64(base)
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d %s", size, units[unit])
	}
	return fmt.Sprintf("%.1f %s", value, units[unit])
}

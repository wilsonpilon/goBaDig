package filemanager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestRenderUsesTheConfiguredColorsNotTviewDefaults is a regression
// test for a readability bug reported directly by the user, in two
// rounds: first "preto com preto" on a selected directory row (fixed
// with an explicit table.SetSelectedStyle - see TestFileManagerEntriesAreNotBlackOnBlack
// in internal/tui), then "cinza com fundo cinza" on ordinary rows
// even after that fix.
//
// Root cause of the second round: tview.NewTableCell seeds cell.Style
// to a concrete, non-default value (tview's own global white-on-
// black), and TableCell.SetTextColor only mutates that style's
// FOREGROUND component in place when cell.Style != tcell.StyleDefault
// - which is always, straight from the constructor. The background
// component is left exactly as tview's own default (black) unless
// SetBackgroundColor is ALSO called. The actually-shipped v1.41.1
// "fix" called only SetTextColor per cell, so every row's text
// correctly picked up the active theme's foreground, but sat on a
// plain black box the theme never chose - "gray on gray" whenever
// that theme's foreground happens to be a muted color on a terminal
// whose rendering of pure black isn't far from it.
//
// This test reads the TableCell's own Style field directly after
// render() - not a rendered tcell.SimulationScreen - specifically
// because an earlier attempt at a screen-level regression test
// passed while this exact bug was still live (a screen dump can't
// easily distinguish "tview's hardcoded default happens to equal
// what I wanted" from "my color was actually applied"); reading the
// cell's own fields is unambiguous.
func TestRenderUsesTheConfiguredColorsNotTviewDefaults(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "afile.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "adir"), 0o755); err != nil {
		t.Fatal(err)
	}

	colors := DefaultColors()
	colors.PanelBackground = tcell.NewRGBColor(0x12, 0x34, 0x56)
	colors.FileText = tcell.NewRGBColor(0x65, 0x43, 0x21)
	colors.DirectoryText = tcell.NewRGBColor(0xAB, 0xCD, 0xEF)

	p := newPanel(colors)
	if err := p.load(dir); err != nil {
		t.Fatal(err)
	}

	var fileRow, dirRow int = -1, -1
	for i, e := range p.entries {
		switch e.name {
		case "afile.txt":
			fileRow = i
		case "adir":
			dirRow = i
		}
	}
	if fileRow < 0 || dirRow < 0 {
		t.Fatalf("fixtures not found in panel entries: %+v", p.entries)
	}

	checkCell := func(row int, wantFg tcell.Color, label string) {
		cell := p.table.GetCell(row, 0)
		fg, bg, _ := cell.Style.Decompose()
		if fg != wantFg {
			t.Errorf("%s: cell.Style foreground = %v, want %v (SetTextColor/SetStyle not taking effect)", label, fg, wantFg)
		}
		if bg != colors.PanelBackground {
			t.Errorf("%s: cell.Style background = %v, want %v (the actual bug: SetBackgroundColor/SetStyle not taking effect, falling back to tview's own hardcoded default)", label, bg, colors.PanelBackground)
		}
	}
	checkCell(fileRow, colors.FileText, "file entry")
	checkCell(dirRow, colors.DirectoryText, "directory entry")
}

// TestMultipleDirectoryEntriesAreActuallyVisibleOnScreen is a
// regression test for the REAL bug behind the user's report - the
// previous two "color" fixes (TestRenderUsesTheConfiguredColorsNotTviewDefaults
// above, and the SetSelectedStyle fix in internal/tui) both read
// cell.Style/cell.Color directly, which stayed correct the whole time
// and never caught this: the problem was never a color at all.
//
// tview's print functions (used by every Table cell, via
// printWithStyle -> step -> parseTag) treat any "[...]" substring as
// an attempted [color]/["region"] style tag whenever the character
// right after '[' is a letter, digit, '#', ':' or '-' - matched and
// CONSUMED (producing zero visible characters) whether or not the
// "color" inside resolves to anything real. This project's own
// Norton-Commander-style "[name]" directory wrapping (panel.go's
// entryLabel) collided with that directly: "[..]" happened to survive
// (the character after '[' is '.', not a valid tag-start character),
// but "[cmd]", "[docs]", "[internal]" - any real directory starting
// with a letter or digit - vanished completely. Confirmed by the user
// screenshotting exactly this (images/erro-filemanager.png, 2026-11-26):
// only ".." and ".claude" (an escaped-by-accident ".name" pattern)
// were visible; every other directory in the repo root showed only
// its date column, the name silently blank.
//
// Root-caused by vendoring a local copy of rivo/tview v0.42.0 with a
// temporary `replace` in go.mod and debug prints inside Table.Draw -
// confirmed printWithStyle's own (start, end) return values were
// (0, 0) - zero characters printed - for "[aaa]"-style text, while
// "[..]" printed all 4. Fixed with tview.Escape (entryLabel), which
// is the library's own documented tool for this exact situation.
//
// This test renders a REAL tcell.SimulationScreen (not just read back
// cell fields) with several directories whose names are ordinary
// letters - exactly the pattern that silently vanished - and asserts
// their full escaped label text is actually present somewhere on
// screen.
func TestMultipleDirectoryEntriesAreActuallyVisibleOnScreen(t *testing.T) {
	dir := t.TempDir()
	// Deliberately includes names that are valid tview tag starts
	// (letters right after where "[" would go) - "aaa"/"bbb"/etc. are
	// even valid 3-digit hex colors, the worst case for this bug.
	names := []string{"cmd", "docs", "internal", "aaa", "resource"}
	for _, name := range names {
		if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	p := newPanel(DefaultColors())
	if err := p.load(dir); err != nil {
		t.Fatal(err)
	}

	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(60, len(p.entries)+4)
	p.table.SetRect(0, 0, 60, len(p.entries)+4)
	p.table.Draw(screen)
	screen.Show()

	cells, width, height := screen.GetContents()
	var line string
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			c := cells[row*width+col]
			if len(c.Runes) > 0 {
				line += string(c.Runes[0])
			} else {
				line += " "
			}
		}
		line += "\n"
	}

	for _, name := range names {
		want := "[" + name + "]"
		if !strings.Contains(line, want) {
			t.Errorf("directory label %q not found anywhere on screen (silently swallowed as a style tag).\nScreen:\n%s", want, line)
		}
	}
}

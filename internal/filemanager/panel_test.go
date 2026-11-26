package filemanager

import (
	"os"
	"path/filepath"
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

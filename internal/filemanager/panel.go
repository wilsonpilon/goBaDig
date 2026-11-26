package filemanager

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// entry is one row shown in a panel: a directory, a file, or the
// ".." pseudo-entry used to go up to the parent directory.
type entry struct {
	name    string
	isDir   bool
	size    int64
	modTime time.Time
}

func (e entry) isParent() bool { return e.name == ".." }

// panel is one side of the two-panel file manager: the directory
// it's currently showing, and the tview.Table rendering its entries
// in three aligned columns (name, size, date) - the Norton-Commander
// shape this component is modeled on.
type panel struct {
	dir     string
	entries []entry
	table   *tview.Table
}

func newPanel() *panel {
	table := tview.NewTable().SetSelectable(true, false)
	table.SetBorder(true)
	table.SetBackgroundColor(tcell.ColorBlack)
	return &panel{table: table}
}

// load reads dir's contents into the panel and renders them. On
// error the panel keeps showing whatever it had before - a failed
// navigation (permission denied, removed directory) shouldn't blank
// the screen.
func (p *panel) load(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	osEntries, err := os.ReadDir(abs)
	if err != nil {
		return err
	}
	entries := make([]entry, 0, len(osEntries)+1)
	if parent := filepath.Dir(abs); parent != abs {
		entries = append(entries, entry{name: "..", isDir: true})
	}
	for _, de := range osEntries {
		info, err := de.Info()
		if err != nil {
			// Unreadable entry (permission denied, broken symlink, a
			// file that vanished between ReadDir and Info) - skip it
			// rather than failing the whole listing over one entry.
			continue
		}
		entries = append(entries, entry{
			name:    de.Name(),
			isDir:   de.IsDir(),
			size:    info.Size(),
			modTime: info.ModTime(),
		})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].isParent() {
			return true
		}
		if entries[j].isParent() {
			return false
		}
		if entries[i].isDir != entries[j].isDir {
			return entries[i].isDir
		}
		return strings.ToLower(entries[i].name) < strings.ToLower(entries[j].name)
	})
	p.dir = abs
	p.entries = entries
	p.render()
	return nil
}

func (p *panel) render() {
	p.table.Clear()
	p.table.SetTitle(" " + p.dir + " ")
	for row, e := range p.entries {
		name := e.name
		size := ""
		if e.isDir {
			name = "[" + name + "]"
		} else {
			size = fmt.Sprintf("%d", e.size)
		}
		date := ""
		if !e.modTime.IsZero() {
			date = e.modTime.Format("2006-01-02 15:04")
		}
		p.table.SetCell(row, 0, tview.NewTableCell(name).SetExpansion(1).SetTextColor(entryColor(e)))
		p.table.SetCell(row, 1, tview.NewTableCell(size).SetAlign(tview.AlignRight).SetTextColor(entryColor(e)))
		p.table.SetCell(row, 2, tview.NewTableCell(date).SetTextColor(entryColor(e)))
	}
	if len(p.entries) > 0 {
		p.table.Select(0, 0)
	}
}

func entryColor(e entry) tcell.Color {
	if e.isDir {
		return tcell.ColorYellow
	}
	return tcell.ColorWhite
}

// selected returns the entry currently highlighted in the panel.
func (p *panel) selected() (entry, bool) {
	row, _ := p.table.GetSelection()
	if row < 0 || row >= len(p.entries) {
		return entry{}, false
	}
	return p.entries[row], true
}

// selectedPath returns the absolute path of the highlighted entry,
// or false if nothing is selected or the selection is "..".
func (p *panel) selectedPath() (string, bool) {
	e, ok := p.selected()
	if !ok || e.isParent() {
		return "", false
	}
	return filepath.Join(p.dir, e.name), true
}

// reload re-reads the current directory, e.g. after a file operation
// changed its contents.
func (p *panel) reload() error {
	return p.load(p.dir)
}

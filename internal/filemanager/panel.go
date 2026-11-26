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

	"github.com/wilsonpilon/goBaDig/internal/chrome"
)

// fileCategory distinguishes the file types a goBaDig user cares
// about at a glance in a panel listing: their own Dignified source,
// the classic ASCII listing a compile produces, and the final
// compiled/packaged artifact ready to run in an emulator. Requested
// by the user once the file manager became "the control center" for
// the whole compile/edit/run cycle - plain file-manager coloring
// (just directory vs. file) no longer told that story.
type fileCategory int

const (
	categoryOther fileCategory = iota
	categorySource
	categoryClassic
	categoryBinary
)

var sourceExtensions = map[string]bool{".dmx": true, ".dzx": true, ".dzs": true}
var classicExtensions = map[string]bool{".amx": true, ".azx": true, ".azs": true}
var binaryExtensions = map[string]bool{".bmx": true, ".p": true, ".tap": true, ".tzx": true, ".dsk": true}

func categorize(name string) fileCategory {
	switch ext := strings.ToLower(filepath.Ext(name)); {
	case sourceExtensions[ext]:
		return categorySource
	case classicExtensions[ext]:
		return categoryClassic
	case binaryExtensions[ext]:
		return categoryBinary
	default:
		return categoryOther
	}
}

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
	colors  Colors
}

func newPanel(colors Colors) *panel {
	table := tview.NewTable().SetSelectable(true, false)
	table.SetBorder(true)
	p := &panel{table: table, colors: colors}
	p.applyColors(colors)
	return p
}

// applyColors restyles the panel in place - called once at
// construction and again whenever the active theme changes
// (internal/tui calls FileManager.SetColors after Setup > Install).
// Explicitly setting the table's own SetSelectedStyle is what fixes
// the black-on-black bug described on the Colors type: without it,
// tview.Table falls back to inverting a selected cell's foreground
// with its (never set here) per-cell BackgroundColor, which isn't
// guaranteed to produce a readable combination.
func (p *panel) applyColors(colors Colors) {
	p.colors = colors
	p.table.SetBackgroundColor(colors.PanelBackground)
	p.table.SetBorderColor(colors.InactiveBorder)
	p.table.SetSelectedStyle(tcell.StyleDefault.
		Background(colors.SelectionBg).
		Foreground(colors.SelectionFg))
	p.render()
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
	// tview.Escape is required here, not optional: a path/filename is
	// arbitrary user/filesystem-controlled text, and tview's print
	// functions (used by both Box titles and Table cells) treat ANY
	// "[...]" shaped substring as a style/region tag - see the long
	// comment on entryLabel below for the bug this caused in
	// practice (directory names silently vanishing).
	p.table.SetTitle(" " + tview.Escape(p.dir) + " ")
	for row, e := range p.entries {
		name := entryLabel(e)
		size := ""
		if !e.isDir {
			size = fmt.Sprintf("%d", e.size)
		}
		date := ""
		if !e.modTime.IsZero() {
			date = e.modTime.Format("2006-01-02 15:04")
		}
		// SetStyle (both fg and bg in one call) - NOT just
		// SetTextColor: tview.NewTableCell seeds cell.Style to a
		// concrete, non-default value (Styles.PrimaryTextColor on
		// Styles.PrimitiveBackgroundColor - tview's own global
		// white-on-black), and SetTextColor only mutates that
		// style's FOREGROUND component in place
		// (table.go: `c.Style = c.Style.Foreground(color)` when
		// c.Style != tcell.StyleDefault, which is always). The
		// background component is left exactly as the constructor
		// set it - tview's own hardcoded black - unless
		// SetBackgroundColor is ALSO called. The actually-shipped
		// v1.41.1 "fix" called only SetTextColor, so every row's
		// text correctly picked up the active theme's color, but
		// sat on a plain black box the theme never chose - "gray on
		// gray" if that theme's foreground happens to be a muted
		// color (confirmed directly: internal/filemanager/panel_test.go
		// reads cell.Style's own fg/bg fields after render, not a
		// rendered screen, to catch this precisely).
		style := tcell.StyleDefault.Foreground(p.entryColor(e)).Background(p.colors.PanelBackground)
		p.table.SetCell(row, 0, tview.NewTableCell(name).SetExpansion(1).SetStyle(style))
		p.table.SetCell(row, 1, tview.NewTableCell(size).SetAlign(tview.AlignRight).SetStyle(style))
		p.table.SetCell(row, 2, tview.NewTableCell(date).SetStyle(style))
	}
	if len(p.entries) > 0 {
		p.table.Select(0, 0)
	}
}

// entryLabel renders an entry's filename for display: directories get
// wrapped in "[name]", exactly like Norton Commander. tview.Escape is
// not optional here - it's the actual fix for a real bug reported
// by the user ("alguns diretórios... ficam preto com preto", later
// found to really be a name silently not printing at all, not a
// color problem): tview's own Print functions (used by every Table
// cell and Box title in this UI) treat ANY "[...]" substring whose
// first character is a letter, digit, '#', ':' or '-' as an attempted
// [color] or ["region"] style tag - matched and CONSUMED whether or
// not the "color" inside actually resolves to anything, producing no
// visible text at all. A name like "[..]" survives by accident (the
// character right after '[' is '.', which isn't a valid tag-start
// character), but "[cmd]", "[docs]", "[internal]" - or any real
// directory/file name starting with a letter or digit - vanish
// completely. tview.Escape wraps the "[...]" sequence in its escaped
// form so the print functions render it as literal text instead of
// trying to parse it as a tag.
func entryLabel(e entry) string {
	if e.isDir {
		return tview.Escape("[" + e.name + "]")
	}
	icon := ""
	switch categorize(e.name) {
	case categorySource:
		icon = chrome.IconSource + " "
	case categoryClassic:
		icon = chrome.IconListing + " "
	case categoryBinary:
		icon = chrome.IconBinary + " "
	}
	return tview.Escape(icon + e.name)
}

func (p *panel) entryColor(e entry) tcell.Color {
	if e.isDir {
		return p.colors.DirectoryText
	}
	switch categorize(e.name) {
	case categorySource:
		return p.colors.SourceText
	case categoryClassic:
		return p.colors.ClassicText
	case categoryBinary:
		return p.colors.BinaryText
	default:
		return p.colors.FileText
	}
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

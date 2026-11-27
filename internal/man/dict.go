package man

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// MsxDictEntry is one verbete of the MSX BASIC dictionary - converted
// once from the user's own MSXIDE project (E:\msxide\ajuda\
// MsxBasicDictData.pbi/MsxBasic2PlusDictData.pbi, see msxdict_data.go)
// rather than hand-written. Source: "Linguagem BASIC MSX" (Denise
// Santoro Cruz, Editora Aleph/Gradiente, 4a edicao 1986) for System
// "MSX1", and the "Manual MSX2+ FM" (ACVS Eletronica) for "MSX2+".
type MsxDictEntry struct {
	Title          string
	IsFunction     bool
	IsAdvanced     bool
	Origin         string
	Summary        string
	Format         string
	Example        string
	Function       string
	ProgramExample string
	PageNumber     int
	System         string
}

// Page renders e in the book's own fixed section order (NOME/FORMATO/
// EXEMPLO/FUNÇÃO/PROGRAMA EXEMPLO), producing the same Page type
// ParseMarkdown does - the CLI and WibEdit's viewer never need to know
// a page came from the dictionary instead of Markdown.
func (e MsxDictEntry) Page() Page {
	page := Page{Title: e.Title}
	heading := func(text string) {
		page.TOC = append(page.TOC, TOCEntry{Title: text, Line: len(page.Lines)})
		page.Lines = append(page.Lines, Line{Text: text, Style: StyleHeading1})
	}
	// Splits on the original source's own "\n" first and wraps each
	// line independently, instead of joining everything into one blob
	// and rewrapping - some verbetes (e.g. GOSUB-RETURN's flow-
	// diagram examples, AUTO's aligned comparison table) depend on
	// their own blank lines/line breaks to stay readable.
	paragraph := func(text string) {
		for line := range strings.SplitSeq(text, "\n") {
			if line == "" {
				page.Lines = append(page.Lines, Line{Text: "", Style: StyleNormal})
				continue
			}
			for _, wrapped := range wrapText(line, wrapWidth) {
				page.Lines = append(page.Lines, Line{Text: wrapped, Style: StyleNormal})
			}
		}
	}
	code := func(text string) {
		for line := range strings.SplitSeq(text, "\n") {
			page.Lines = append(page.Lines, Line{Text: line, Style: StyleCode})
		}
	}

	heading("NOME")
	kind := "instrução"
	if e.IsFunction {
		kind = "função"
	}
	summary := fmt.Sprintf("%s - %s (%s)", e.Title, kind, e.System)
	if e.IsAdvanced {
		summary += " [uso avançado]"
	}
	paragraph(summary)
	if e.Origin != "" {
		paragraph("Origem: " + e.Origin)
	}
	paragraph(e.Summary)

	if e.Format != "" {
		heading("FORMATO")
		code(e.Format)
	}
	if e.Example != "" {
		heading("EXEMPLO")
		code(e.Example)
	}
	if e.Function != "" {
		heading("FUNÇÃO")
		paragraph(e.Function)
	}
	if e.ProgramExample != "" {
		heading("PROGRAMA EXEMPLO")
		code(e.ProgramExample)
	}
	heading("REFERÊNCIA")
	paragraph(fmt.Sprintf("Página %d do manual original (%s).", e.PageNumber, e.System))

	return page
}

// FindMsxDictEntry finds a dictionary entry by keyword (case-
// insensitive exact match on Title first; falling back to a compound
// entry whose title's last "-"-separated part matches, e.g. "RETURN"
// resolving to the combined "GOSUB-RETURN" verbete, the same way the
// book itself documents paired keywords together).
func FindMsxDictEntry(keyword string) (MsxDictEntry, bool) {
	key := strings.ToUpper(strings.TrimSpace(keyword))
	if key == "" {
		return MsxDictEntry{}, false
	}
	for _, e := range msxDict {
		if strings.ToUpper(e.Title) == key {
			return e, true
		}
	}
	for _, e := range msxDict {
		for part := range strings.SplitSeq(e.Title, "-") {
			if strings.ToUpper(strings.TrimSpace(part)) == key {
				return e, true
			}
		}
	}
	return MsxDictEntry{}, false
}

// AutocompleteWords returns every single-word keyword the MSX BASIC
// dictionary knows about, normalized for prefix-matching: a handful
// of entries document several related keywords under one compound
// title (e.g. "GOSUB-RETURN", "CIRCLE, DRAW, LINE, PAINT, POINT,
// PRESET e PSET (MSX2+)") - each becomes its own candidate here, the
// same way the user's own MSXIDE splits/normalizes dictionary titles
// before offering them as autocomplete matches. Computed once; the
// dictionary itself never changes at runtime.
func AutocompleteWords() []string {
	return autocompleteWords
}

var autocompleteWords = buildAutocompleteWords()

var parenGroup = regexp.MustCompile(`\([^)]*\)`)

func buildAutocompleteWords() []string {
	seen := make(map[string]bool)
	for _, e := range msxDict {
		for _, word := range splitDictTitle(e.Title) {
			seen[word] = true
		}
	}
	words := make([]string, 0, len(seen))
	for word := range seen {
		words = append(words, word)
	}
	sort.Strings(words)
	return words
}

// splitDictTitle normalizes one dictionary Title into its component
// keyword(s): parenthesized groups (a "(MSX2+)"-style system marker,
// or a "(n)"-style parameter placeholder) are dropped entirely, "/"
// becomes a separator alongside space/comma/hyphen, and the bare
// Portuguese conjunction "e" ("and", from comma-separated lists like
// "CIRCLE, DRAW, ... PRESET e PSET") is dropped too.
func splitDictTitle(title string) []string {
	cleaned := parenGroup.ReplaceAllString(title, " ")
	cleaned = strings.ReplaceAll(cleaned, "/", " ")
	var words []string
	for _, field := range strings.FieldsFunc(cleaned, func(r rune) bool {
		return r == ' ' || r == ',' || r == '-'
	}) {
		field = strings.ToUpper(strings.TrimSpace(field))
		if field == "" || field == "E" {
			continue
		}
		words = append(words, field)
	}
	return words
}

// MsxDictIndex lists every dictionary entry grouped by its first
// letter, each line linked via TOC - the "man msx"/WibEdit-viewer
// index into the dictionary, separate from Index()'s own list of
// Markdown-sourced topics (compile/execute/...).
func MsxDictIndex() Page {
	sorted := make([]MsxDictEntry, len(msxDict))
	copy(sorted, msxDict)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToUpper(sorted[i].Title) < strings.ToUpper(sorted[j].Title)
	})

	page := Page{Title: "Dicionário MSX BASIC"}
	page.Lines = append(page.Lines, Line{Text: page.Title, Style: StyleHeading1})
	currentLetter := byte(0)
	for _, e := range sorted {
		title := strings.ToUpper(e.Title)
		if title[0] != currentLetter {
			currentLetter = title[0]
			page.Lines = append(page.Lines, Line{Text: string(currentLetter), Style: StyleHeading2})
		}
		page.TOC = append(page.TOC, TOCEntry{Title: e.Title, Line: len(page.Lines)})
		page.Lines = append(page.Lines, Line{Text: "- " + e.Title, Style: StyleBullet})
	}
	return page
}

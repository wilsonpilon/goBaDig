package man

import (
	"fmt"
	"sort"
	"strings"

	docsman "github.com/wilsonpilon/goBaDig/docs/man"
)

// topics holds every page this registry can serve, keyed by lowercase
// topic title, built once from the embedded Markdown sources at
// startup. Fase 3 adds the MSX BASIC dictionary here too (preferred
// over a plain Markdown page when both exist for the same keyword);
// Fase 4 adds an "asmsx" topic the same way Badig is added below.
var topics = buildTopics()

func buildTopics() map[string]Page {
	pages := make(map[string]Page)
	addMarkdownSource(pages, docsman.Badig)
	return pages
}

// addMarkdownSource splits src's top-level ("# ") sections into
// separate pages, keyed by each section's own heading text lowercased.
// docs/man/badig.md (and, from Fase 4, asmsx.md) is written as one
// Markdown file per "manual", with exactly one "# " heading per topic,
// rather than one file per topic - it reads naturally as a single
// document in an editor or on GitHub, while still serving as several
// independent man topics at runtime.
func addMarkdownSource(pages map[string]Page, src string) {
	for _, section := range splitTopLevelSections(src) {
		page := ParseMarkdown(section)
		key := strings.ToLower(strings.TrimSpace(page.Title))
		if key == "" {
			continue
		}
		pages[key] = page
	}
}

func splitTopLevelSections(src string) []string {
	var sections []string
	var current []string
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "# ") && len(current) > 0 {
			sections = append(sections, strings.Join(current, "\n"))
			current = nil
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		sections = append(sections, strings.Join(current, "\n"))
	}
	return sections
}

// Lookup finds a topic by name (case-insensitive) - the CLI's "man
// <tópico>"/"--man <tópico>", and WibEdit's Manual viewer following a
// TOC link (view_man.go), and Shift+F1's contextual lookup
// (internal/editor). The MSX BASIC dictionary (dict.go) is tried
// first - its keywords never collide with the goBaDig command topics
// below, but a verbete always wins if they ever do, since it is the
// more specific reference for an actual BASIC keyword.
func Lookup(topic string) (Page, error) {
	if entry, ok := FindMsxDictEntry(topic); ok {
		return entry.Page(), nil
	}
	key := strings.ToLower(strings.TrimSpace(topic))
	if key == msxDictTopic {
		return MsxDictIndex(), nil
	}
	if page, ok := topics[key]; ok {
		return page, nil
	}
	return Page{}, fmt.Errorf("tópico de man desconhecido: %q (use \"man\" sem argumento para ver o índice)", topic)
}

// msxDictTopic is the reserved pseudo-topic ("man msx") that opens the
// MSX BASIC dictionary's own index (dict.go's MsxDictIndex) - not a
// real dictionary keyword, so it is checked separately from
// FindMsxDictEntry above.
const msxDictTopic = "msx"

// Index lists every available topic as a Page of its own, each line
// linked via TOC so a viewer can jump straight into it - the "man"/
// "--man" output with no topic argument, and the page WibEdit's Manual
// viewer opens to first.
func Index() Page {
	names := make([]string, 0, len(topics))
	for name := range topics {
		names = append(names, name)
	}
	sort.Strings(names)

	page := Page{Title: "goBaDig - índice do manual"}
	page.Lines = append(page.Lines, Line{Text: page.Title, Style: StyleHeading1})
	page.Lines = append(page.Lines, Line{Text: "Comandos do goBaDig", Style: StyleHeading2})
	for _, name := range names {
		page.TOC = append(page.TOC, TOCEntry{Title: topics[name].Title, Line: len(page.Lines)})
		page.Lines = append(page.Lines, Line{Text: "- " + topics[name].Title, Style: StyleBullet})
	}
	page.Lines = append(page.Lines, Line{Text: "BASIC", Style: StyleHeading2})
	page.TOC = append(page.TOC, TOCEntry{Title: msxDictTopic, Line: len(page.Lines)})
	page.Lines = append(page.Lines, Line{
		Text:  fmt.Sprintf("- Dicionário MSX BASIC (%d verbetes)", len(msxDict)),
		Style: StyleBullet,
	})
	return page
}

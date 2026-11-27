package man

import (
	"strings"
	"testing"
)

func TestParseMarkdownHeadingsCodeQuoteBulletAndParagraph(t *testing.T) {
	src := strings.Join([]string{
		"# Título",
		"",
		"## Sub",
		"",
		"Um parágrafo comum que deveria permanecer numa única linha curta.",
		"",
		"```",
		"codigo literal",
		"```",
		"",
		"> uma citação",
		"",
		"- item um",
		"* item dois",
		"1. item três",
	}, "\n")

	page := ParseMarkdown(src)

	if page.Title != "Título" {
		t.Fatalf("Title = %q, want %q", page.Title, "Título")
	}
	if len(page.TOC) != 2 || page.TOC[0].Title != "Título" || page.TOC[1].Title != "Sub" {
		t.Fatalf("TOC = %+v, want [Título Sub]", page.TOC)
	}

	var gotCode, gotQuote, gotBullets int
	for _, line := range page.Lines {
		switch line.Style {
		case StyleCode:
			gotCode++
			if line.Text != "codigo literal" {
				t.Errorf("code line = %q, want %q", line.Text, "codigo literal")
			}
		case StyleQuote:
			gotQuote++
			if line.Text != "uma citação" {
				t.Errorf("quote line = %q, want %q", line.Text, "uma citação")
			}
		case StyleBullet:
			gotBullets++
		}
	}
	if gotCode != 1 {
		t.Errorf("code lines = %d, want 1", gotCode)
	}
	if gotQuote != 1 {
		t.Errorf("quote lines = %d, want 1", gotQuote)
	}
	if gotBullets != 3 {
		t.Errorf("bullet lines = %d, want 3", gotBullets)
	}
}

func TestParseMarkdownWrapsLongParagraphs(t *testing.T) {
	long := strings.Repeat("palavra ", 30)
	page := ParseMarkdown(long)
	if len(page.Lines) < 2 {
		t.Fatalf("expected a long paragraph to wrap into multiple lines, got %d", len(page.Lines))
	}
	for _, line := range page.Lines {
		if len(line.Text) > wrapWidth {
			t.Errorf("line %q exceeds wrapWidth %d", line.Text, wrapWidth)
		}
	}
}

func TestLookupAndIndex(t *testing.T) {
	page, err := Lookup("compile")
	if err != nil {
		t.Fatalf("Lookup(compile) error: %v", err)
	}
	if page.Title != "compile" {
		t.Errorf("Title = %q, want %q", page.Title, "compile")
	}

	if _, err := Lookup("COMPILE"); err != nil {
		t.Errorf("Lookup should be case-insensitive: %v", err)
	}

	if _, err := Lookup("nao-existe"); err == nil {
		t.Error("Lookup(nao-existe) should return an error")
	}

	index := Index()
	if len(index.TOC) == 0 {
		t.Fatal("Index() should list at least one topic")
	}
	found := false
	for _, entry := range index.TOC {
		if entry.Title == "compile" {
			found = true
		}
	}
	if !found {
		t.Error("Index() should list \"compile\" among its topics")
	}
}

func TestPlainLinesUppercasesHeadings(t *testing.T) {
	page := ParseMarkdown("# título\n\ntexto normal")
	lines := page.PlainLines()
	if lines[0] != "TÍTULO" {
		t.Errorf("heading line = %q, want %q", lines[0], "TÍTULO")
	}
	if lines[1] != "texto normal" {
		t.Errorf("paragraph line = %q, want %q", lines[1], "texto normal")
	}
}

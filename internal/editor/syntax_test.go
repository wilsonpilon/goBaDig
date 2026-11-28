package edit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestDialectForPath(t *testing.T) {
	cases := map[string]syntaxDialect{
		"game.dmx":    dialectMSX,
		"game.amx":    dialectMSX,
		"GAME.DMX":    dialectMSX,
		"game.dzx":    dialectZX81,
		"game.azx":    dialectZX81,
		"game.dzs":    dialectZXSpectrum,
		"game.azs":    dialectZXSpectrum,
		"rotina.asm":  dialectASM,
		"ROTINA.ASM":  dialectASM,
		"game.txt":    dialectNone,
		"game.bmx":    dialectNone,
		"":            dialectNone,
		"noextension": dialectNone,
	}
	for path, want := range cases {
		if got := dialectForPath(path); got != want {
			t.Errorf("dialectForPath(%q) = %v, want %v", path, got, want)
		}
	}
}

func kindsOf(t *testing.T, dialect syntaxDialect, text string) []syntaxKind {
	t.Helper()
	spans := tokenizeLine(dialect, text)
	kinds := make([]syntaxKind, len(spans))
	for i, s := range spans {
		kinds[i] = s.kind
	}
	return kinds
}

func textOf(spans []syntaxSpan, text string) []string {
	out := make([]string, len(spans))
	for i, s := range spans {
		out[i] = text[s.start:s.end]
	}
	return out
}

func TestTokenizeLineNoDialectProducesNoSpans(t *testing.T) {
	if spans := tokenizeLine(dialectNone, `10 PRINT "hi"`); spans != nil {
		t.Fatalf("tokenizeLine(dialectNone, ...) = %v, want nil", spans)
	}
}

func TestTokenizeLineMSXKeywordsFunctionsStringsNumbers(t *testing.T) {
	text := `print "hello" + str$(1.5e2) : rem a comment`
	spans := tokenizeLine(dialectMSX, text)
	words := textOf(spans, text)
	kinds := make([]syntaxKind, len(spans))
	for i, s := range spans {
		kinds[i] = s.kind
	}
	want := []struct {
		word string
		kind syntaxKind
	}{
		{"print", synKeyword},
		{`"hello"`, synString},
		{"str$", synFunction},
		{"1.5e2", synNumber},
		{"rem a comment", synComment},
	}
	if len(spans) != len(want) {
		t.Fatalf("tokenizeLine(%q) = %v (words %v), want %d spans", text, spans, words, len(want))
	}
	for i, w := range want {
		if words[i] != w.word || kinds[i] != w.kind {
			t.Errorf("span %d = (%q, %v), want (%q, %v)", i, words[i], kinds[i], w.word, w.kind)
		}
	}
}

func TestTokenizeLineMSXHexNumberAndLabel(t *testing.T) {
	text := `{loop} poke &H1F00, 1`
	spans := tokenizeLine(dialectMSX, text)
	words := textOf(spans, text)
	foundLabel, foundHex := false, false
	for i, s := range spans {
		switch {
		case s.kind == synLabel && words[i] == "{loop}":
			foundLabel = true
		case s.kind == synNumber && words[i] == "&H1F00":
			foundHex = true
		}
	}
	if !foundLabel {
		t.Errorf("tokenizeLine(%q): no label span for {loop}, got %v / %v", text, spans, words)
	}
	if !foundHex {
		t.Errorf("tokenizeLine(%q): no number span for &H1F00, got %v / %v", text, spans, words)
	}
}

func TestTokenizeLineMSXApostropheCommentToEndOfLine(t *testing.T) {
	text := `print 1 ' trailing comment with "quotes" and 123`
	spans := tokenizeLine(dialectMSX, text)
	last := spans[len(spans)-1]
	if last.kind != synComment {
		t.Fatalf("last span kind = %v, want synComment", last.kind)
	}
	if text[last.start:last.end] != `' trailing comment with "quotes" and 123` {
		t.Fatalf("comment span = %q, want the rest of the line verbatim", text[last.start:last.end])
	}
}

func TestTokenizeLineZX81KeywordsAndFunctions(t *testing.T) {
	kinds := kindsOf(t, dialectZX81, `let a=peek 16514`)
	want := []syntaxKind{synKeyword, synFunction, synNumber}
	if len(kinds) != len(want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Errorf("kind %d = %v, want %v", i, kinds[i], want[i])
		}
	}
}

func TestTokenizeLineZXSpectrumKeywordsAndFunctions(t *testing.T) {
	kinds := kindsOf(t, dialectZXSpectrum, `print at 10,10; chr$ 65`)
	// "10" and "10" are two separate numeric literals (either side of
	// the comma), not one span - each gets its own.
	want := []syntaxKind{synKeyword, synKeyword, synNumber, synNumber, synFunction, synNumber}
	if len(kinds) != len(want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Errorf("kind %d = %v, want %v", i, kinds[i], want[i])
		}
	}
}

func TestTokenizeLineHashCommentAllThreeDialects(t *testing.T) {
	for _, dialect := range []syntaxDialect{dialectMSX, dialectZX81, dialectZXSpectrum} {
		text := `print 1 ## exclusive comment`
		spans := tokenizeLine(dialect, text)
		last := spans[len(spans)-1]
		if last.kind != synComment {
			t.Errorf("dialect %v: last span kind = %v, want synComment", dialect, last.kind)
		}
	}
}

func TestTokenizeLineUSRFunctionVariants(t *testing.T) {
	for _, word := range []string{"USR", "USR0", "USR9"} {
		spans := tokenizeLine(dialectMSX, "x = "+word+"(0)")
		found := false
		for i, s := range spans {
			if s.kind == synFunction && textOf(spans, "x = "+word+"(0)")[i] == word {
				found = true
			}
		}
		if !found {
			t.Errorf("USR variant %q not recognized as a function: %v", word, spans)
		}
	}
}

func TestLookupWordIsCaseInsensitive(t *testing.T) {
	for _, word := range []string{"print", "PRINT", "Print", "PrInT"} {
		if kind := lookupWord(dialectMSX, word); kind != synKeyword {
			t.Errorf("lookupWord(dialectMSX, %q) = %v, want synKeyword", word, kind)
		}
	}
}

// TestOpeningADmxFileActuallyRendersSyntaxColors is the end-to-end
// check for the whole feature, not just the tokenizer in isolation:
// opens a real .dmx file through Editor.OpenFile (the exact path
// internal/filemanager's F4 uses), draws to a simulated screen, and
// confirms the keyword and the string literal actually render in
// their own distinct colors - not just that tokenizeLine classifies
// them correctly off-screen.
func TestOpeningADmxFileActuallyRendersSyntaxColors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.dmx")
	if err := os.WriteFile(path, []byte(`print "hi"`), 0o644); err != nil {
		t.Fatal(err)
	}

	editor, _, screen := newTestEditor(t, 80, 25, nil)
	if err := editor.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	drawEditor(editor, screen)

	width, _ := screen.Size()
	var keywordFg, stringFg tcell.Color
	foundKeyword, foundString := false, false
	for x := 0; x < width; x++ {
		ch, _, style, _ := screen.GetContent(x, 1)
		fg, _, _ := style.Decompose()
		switch ch {
		case 'p':
			if !foundKeyword {
				keywordFg = fg
			}
			foundKeyword = true
		case '"':
			if !foundString {
				stringFg = fg
			}
			foundString = true
		}
	}
	if !foundKeyword || !foundString {
		t.Fatalf("expected to find both 'p' (print) and '\"' (string) on row 1; row=%q", editorScreenRow(screen, 1))
	}
	wantKeywordFg, _, _ := syntaxKeywordStyle.Decompose()
	wantStringFg, _, _ := syntaxStringStyle.Decompose()
	if keywordFg != wantKeywordFg {
		t.Errorf("'print' rendered with fg %v, want the keyword color %v", keywordFg, wantKeywordFg)
	}
	if stringFg != wantStringFg {
		t.Errorf(`'"hi"' rendered with fg %v, want the string color %v`, stringFg, wantStringFg)
	}
	if keywordFg == stringFg {
		t.Error("keyword and string rendered with the SAME color - highlighting isn't actually distinguishing them")
	}
}

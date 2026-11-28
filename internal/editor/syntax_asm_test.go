package edit

import "testing"

// findSpan returns the kind of the first span whose text equals word,
// or synNone with ok=false if none matches - convenient for assertions
// that only care about one or two specific tokens within a realistic
// line, rather than the whole span list in order.
func findSpan(spans []syntaxSpan, text, word string) (syntaxKind, bool) {
	for _, s := range spans {
		if text[s.start:s.end] == word {
			return s.kind, true
		}
	}
	return synNone, false
}

func TestTokenizeLineAsmCommentConventions(t *testing.T) {
	cases := []struct {
		text string
		want string // the exact comment span text expected
	}{
		{"ld a,1 ; a comment", "; a comment"},
		{"ld a,1 // a comment", "// a comment"},
		{"ld a,1 -- a comment", "-- a comment"},
		{"ld a,1 /* a comment */ nop", "/* a comment */"},
		{"ld a,1 { a comment } nop", "{ a comment }"},
	}
	for _, c := range cases {
		spans := tokenizeLine(dialectASM, c.text)
		kind, ok := findSpan(spans, c.text, c.want)
		if !ok || kind != synComment {
			t.Errorf("tokenizeLine(%q): comment span %q not found as synComment (spans %v)", c.text, c.want, spans)
		}
	}
}

func TestTokenizeLineAsmNumericConventions(t *testing.T) {
	for _, number := range []string{"0x8a", "0xff", "$8a", "8ah", "0ffh", "10h", "77o", "1o", "11110000b", "1001b", "2048", "3.14"} {
		text := "db " + number
		spans := tokenizeLine(dialectASM, text)
		kind, ok := findSpan(spans, text, number)
		if !ok || kind != synNumber {
			t.Errorf("tokenizeLine(%q): %q not found as a single synNumber span (spans %v)", text, number, spans)
		}
	}
}

func TestTokenizeLineAsmLabelDefinitions(t *testing.T) {
	for _, text := range []string{"loop:", ".loop:", "@@loop:", "Function1:"} {
		spans := tokenizeLine(dialectASM, text)
		if len(spans) != 1 || spans[0].kind != synLabel {
			t.Errorf("tokenizeLine(%q) = %v, want exactly one synLabel span", text, spans)
		}
	}
}

func TestTokenizeLineAsmDirectivesWithAndWithoutLeadingDot(t *testing.T) {
	for _, text := range []string{".ORG 08000h", "ORG 08000h", ".org 08000h", "org 08000h"} {
		spans := tokenizeLine(dialectASM, text)
		kind, ok := findSpan(spans, text, text[:len(text)-len(" 08000h")])
		if !ok || kind != synKeyword {
			t.Errorf("tokenizeLine(%q): directive not found as synKeyword (spans %v)", text, spans)
		}
	}
}

func TestTokenizeLineAsmMnemonicsRegistersAndBiosConstants(t *testing.T) {
	text := "call CHPUT ; writes A to the screen"
	spans := tokenizeLine(dialectASM, text)
	if kind, ok := findSpan(spans, text, "call"); !ok || kind != synKeyword {
		t.Errorf("tokenizeLine(%q): %q not synKeyword (spans %v)", text, "call", spans)
	}
	if kind, ok := findSpan(spans, text, "CHPUT"); !ok || kind != synConstant {
		t.Errorf("tokenizeLine(%q): %q not synConstant (spans %v)", text, "CHPUT", spans)
	}
}

func TestTokenizeLineAsmRegisterIsKeyword(t *testing.T) {
	text := "ld a,HL"
	spans := tokenizeLine(dialectASM, text)
	for _, word := range []string{"ld", "a", "HL"} {
		if kind, ok := findSpan(spans, text, word); !ok || kind != synKeyword {
			t.Errorf("tokenizeLine(%q): %q = %v, want synKeyword", text, word, kind)
		}
	}
}

func TestTokenizeLineAsmLabelWinsOverAMatchingMnemonicName(t *testing.T) {
	// "DI" is a real Z80 mnemonic (disable interrupts); as a label
	// ("DI:") the trailing colon should still win, matching how
	// tokenizeLine checks for it ahead of lookupWord for dialectASM.
	text := "DI:"
	spans := tokenizeLine(dialectASM, text)
	if len(spans) != 1 || spans[0].kind != synLabel {
		t.Fatalf("tokenizeLine(%q) = %v, want exactly one synLabel span", text, spans)
	}
}

func TestTokenizeLineAsmRealisticExampleFromAsmsxManual(t *testing.T) {
	text := `points:        db 20,30,40,50  ; points table`
	spans := tokenizeLine(dialectASM, text)
	if kind, ok := findSpan(spans, text, "points"); !ok || kind != synLabel {
		t.Errorf("label not found: spans %v", spans)
	}
	if kind, ok := findSpan(spans, text, "db"); !ok || kind != synKeyword {
		t.Errorf("db directive not found: spans %v", spans)
	}
	if kind, ok := findSpan(spans, text, "; points table"); !ok || kind != synComment {
		t.Errorf("comment not found: spans %v", spans)
	}
}

package edit

import (
	"fmt"
	"strings"
	"testing"
)

func TestDocumentNavigationFWMSXCases(t *testing.T) {
	tests := []struct {
		name, text string
		at         documentPosition
		commands   []string
		want       documentPosition
		column     int
	}{
		{"left across line", "hello\nx", documentPosition{1, 0}, []string{"cursor.left"}, documentPosition{0, 5}, 5},
		{"right across line", "hello\nx", documentPosition{0, 5}, []string{"cursor.right"}, documentPosition{1, 0}, 0},
		{"previous word", "one two\nthree", documentPosition{1, 0}, []string{"cursor.word-left", "cursor.word-left"}, documentPosition{0, 4}, 4},
		{"next word then next line", "one two\nthree", documentPosition{}, []string{"cursor.word-right", "cursor.word-right", "cursor.word-right"}, documentPosition{1, 0}, 0},
		{"punctuation and tabs", "one,\t_two!", documentPosition{}, []string{"cursor.word-right"}, documentPosition{0, 5}, 8},
		{"line home", "hello\nx", documentPosition{0, 3}, []string{"cursor.home"}, documentPosition{}, 0},
		{"line end", "hello\nx", documentPosition{}, []string{"cursor.end"}, documentPosition{0, 5}, 5},
		{"file top", "hello\nx", documentPosition{1, 1}, []string{"cursor.file-top"}, documentPosition{}, 0},
		{"file end", "hello\nx", documentPosition{}, []string{"cursor.file-end"}, documentPosition{1, 1}, 1},
		{"tab goal down", "a\tb\nxxxxxxxxxxxx", documentPosition{0, 2}, []string{"cursor.down"}, documentPosition{1, 8}, 8},
		{"tab goal back up", "a\tb\nxxxxxxxxxxxx", documentPosition{0, 2}, []string{"cursor.down", "cursor.up"}, documentPosition{0, 2}, 8},
		{"goal across empty line", "abcdefghij\n\nx\nabcdefghij", documentPosition{0, 8}, []string{"cursor.down", "cursor.down", "cursor.down"}, documentPosition{3, 8}, 8},
		{"goal inside tab floors", "abcd\n\tend\nabcd", documentPosition{0, 3}, []string{"cursor.down"}, documentPosition{1, 0}, 0},
		{"goal after tab floor survives", "abcd\n\tend\nabcd", documentPosition{0, 3}, []string{"cursor.down", "cursor.down"}, documentPosition{2, 3}, 3},
		{"horizontal resets goal", "abcdefghij\nx\nabcdefghij", documentPosition{0, 8}, []string{"cursor.down", "cursor.home", "cursor.down"}, documentPosition{2, 0}, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := newDocument(test.text)
			d.setCursor(test.at, false)
			for _, command := range test.commands {
				if !d.navigate(command) {
					t.Fatalf("unhandled command %s", command)
				}
			}
			if d.cursor != test.want || d.column(d.cursor) != test.column {
				t.Fatalf("cursor=%+v column=%d; want %+v column=%d", d.cursor, d.column(d.cursor), test.want, test.column)
			}
			if d.text != test.text || d.revision != 0 {
				t.Fatal("navigation changed the document")
			}
		})
	}
}

func TestDocumentPagesScreenAndKeyboardScroll(t *testing.T) {
	var text strings.Builder
	for i := 0; i < 100; i++ {
		fmt.Fprintf(&text, "line %03d\n", i)
	}
	d := newDocument(strings.TrimSuffix(text.String(), "\n"))
	d.navigate("cursor.page-down")
	if d.cursor.line != 22 || d.top != 22 {
		t.Fatalf("80x25 page down must move 22 rows: cursor=%+v top=%d", d.cursor, d.top)
	}
	d.navigate("cursor.page-up")
	if d.cursor.line != 0 || d.top != 0 {
		t.Fatal("page up did not return to top")
	}
	d.navigate("cursor.scroll-up")
	if d.top != 1 || d.cursor.line != 1 {
		t.Fatal("Alt-Z must scroll text up and clamp a cursor leaving the viewport")
	}
	d.navigate("cursor.scroll-down")
	if d.top != 0 || d.cursor.line != 1 {
		t.Fatal("Ctrl-Z must scroll text down without undo or needless cursor movement")
	}
	d.navigate("cursor.screen-bottom")
	if d.cursor.line != 22 || d.top != 0 {
		t.Fatal("Ctrl-PgDn must move to the bottom of the current screen")
	}
	d.navigate("cursor.screen-top")
	if d.cursor.line != 0 || d.top != 0 {
		t.Fatal("Ctrl-PgUp must move to the top of the current screen")
	}
	d.setCursor(documentPosition{12, 4}, false)
	d.navigate("cursor.scroll-up")
	if d.cursor != (documentPosition{12, 4}) || d.top != 1 {
		t.Fatal("scroll must leave an already visible cursor in place")
	}
	d.navigate("cursor.file-end")
	for i := 0; i < 110; i++ {
		d.navigate("cursor.scroll-up")
	}
	if d.top != 99 || d.cursor.line != 99 {
		t.Fatal("keyboard scrolling exceeded the final line")
	}
	d.resize(12, 4, false)
	d.navigate("cursor.page-up")
	if d.cursor.line != 96 || d.top != 96 {
		t.Fatalf("page size must follow viewport height: cursor=%+v top=%d", d.cursor, d.top)
	}
}

func TestDocumentGraphemesBytesColumnsAndControls(t *testing.T) {
	text := "a\t界e\u0301\x00\x1b\x7f\r\nlast\n"
	d := newDocument(text)
	for _, test := range []struct {
		character, offset, column int
	}{
		{0, 0, 0}, {1, 1, 1}, {2, 2, 8}, {3, 5, 10},
		{4, 8, 11}, {5, 9, 12}, {6, 10, 13}, {7, 11, 14},
	} {
		p := documentPosition{0, test.character}
		if d.byteOffset(p) != test.offset || d.column(p) != test.column || d.positionAtByte(test.offset) != p {
			t.Fatalf("character %d conflates bytes, graphemes or display columns", test.character)
		}
	}
	d.setCursor(documentPosition{0, 3}, false)
	d.navigate("cursor.right")
	if d.cursor.character != 4 || d.byteOffset(d.cursor) != 8 {
		t.Fatal("Right must cross the whole combining grapheme")
	}
	d.navigate("cursor.left")
	if d.byteOffset(d.cursor) != 5 {
		t.Fatal("Left split a combining grapheme")
	}
	d.setCursor(documentPosition{0, 7}, false)
	d.navigate("cursor.right")
	if d.cursor != (documentPosition{1, 0}) || d.byteOffset(d.cursor) != 13 {
		t.Fatal("CRLF must be one line boundary while preserving both bytes")
	}
	d.navigate("cursor.file-end")
	if d.cursor != (documentPosition{2, 0}) || d.byteOffset(d.cursor) != len(text) || d.text != text {
		t.Fatal("trailing newline or controls were lost")
	}
}

func TestDocumentBlockMarkersTrackEditsAndFindBoundaries(t *testing.T) {
	d := newDocument("abcdef")
	d.setCursor(d.positionAtByte(1), false)
	if marker, complete := d.setBlockMarkerAtCursor(); marker != 1 || complete {
		t.Fatal("first Set marker must add only the first boundary")
	}
	d.setCursor(d.positionAtByte(4), false)
	if marker, complete := d.setBlockMarkerAtCursor(); marker != 2 || !complete {
		t.Fatal("second Set marker must complete the block")
	}
	if start, end := d.selection(); d.text[start:end] != "bcd" {
		t.Fatalf("marked text = %q, want %q", d.text[start:end], "bcd")
	}
	for _, expected := range []struct {
		index, offset int
	}{{0, 1}, {1, 4}} {
		index, offset, ok := d.findBlockMarker()
		if !ok || index != expected.index || offset != expected.offset || d.byteOffset(d.cursor) != offset {
			t.Fatalf("Find marker = %d, %d, %v; want %d, %d", index, offset, ok, expected.index, expected.offset)
		}
	}

	d.replaceBytes(0, 0, "X")
	if start, end := d.selection(); d.text[start:end] != "bcd" || start != 2 || end != 5 {
		t.Fatalf("markers did not shift with preceding insertion: %q [%d,%d)", d.text[start:end], start, end)
	}
	d.replaceBytes(3, 3, "Y")
	if start, end := d.selection(); d.text[start:end] != "bYcd" || start != 2 || end != 6 {
		t.Fatalf("insertion inside block was not included: %q [%d,%d)", d.text[start:end], start, end)
	}
	d.replaceBytes(2, 6, "Q")
	if start, end := d.selection(); d.text[start:end] != "Q" || start != 2 || end != 3 {
		t.Fatalf("replacement did not preserve the marked range: %q [%d,%d)", d.text[start:end], start, end)
	}

	d.setCursor(d.positionAtByte(4), false)
	if marker, complete := d.setBlockMarkerAtCursor(); marker != 1 || complete || d.blockMarkerCount != 1 {
		t.Fatal("a third marker must begin a new block instead of keeping a stale second boundary")
	}
	start, end := d.selection()
	if start != end {
		t.Fatalf("starting a new block retained a stale selection: [%d,%d)", start, end)
	}
	d.clearBlockMarkers()
	if d.blockMarkerCount != 0 {
		t.Fatal("Remove markers did not clear the block boundaries")
	}
}

func TestDocumentBlockMarkLineAndEndOfLine(t *testing.T) {
	d := newDocument("first\r\nsecond\n")
	d.setCursor(documentPosition{0, 2}, false)
	d.markCurrentLine()
	start, end := d.selection()
	if d.text[start:end] != "first\r\n" {
		t.Fatalf("current line = %q, want its text and CRLF", d.text[start:end])
	}

	d.setCursor(documentPosition{1, 2}, false)
	d.markToEndOfLine()
	start, end = d.selection()
	if d.text[start:end] != "cond" {
		t.Fatalf("mark to line end = %q, want %q", d.text[start:end], "cond")
	}

	d.setCursor(documentPosition{1, 6}, false)
	d.markToEndOfLine()
	start, end = d.selection()
	if start != end || d.blockMarkerCount != 2 {
		t.Fatal("marking from EOL must represent an empty block with two boundaries")
	}
}

func FuzzDocumentNavigationPreservesBytes(f *testing.F) {
	for _, text := range []string{"", "a\tb\r\n\nlast", "界e\u0301\x00\x1b\xff", strings.Repeat("x", 160)} {
		f.Add(text, uint8(79), uint8(23), false)
	}
	f.Fuzz(func(t *testing.T, text string, width, height uint8, wrap bool) {
		if len(text) > 4096 {
			t.Skip()
		}
		d := newDocument(text)
		d.resize(max(1, int(width)), max(1, int(height)), wrap)
		for _, command := range neKeyCommands {
			if command.phase != "F2.1" {
				continue
			}
			d.navigate(command.id)
			if d.text != text || d.revision != 0 || d.cursor != d.clamp(d.cursor) ||
				d.positionAtByte(d.byteOffset(d.cursor)) != d.cursor || d.top < 0 || d.top >= len(d.rows) || d.left < 0 {
				t.Fatalf("%s changed bytes or escaped document bounds", command.id)
			}
		}
	})
}

func TestDocumentNavigationPreservesAllBytesAndBounds(t *testing.T) {
	texts := []string{"", "\n\n", "a\r\n\r\nb\r", "\x00\x1b\x1a\x7f\xff\t\n", "a\t界e\u0301\nx\n", strings.Repeat("x", 100000)}
	var keys []string
	for _, command := range neKeyCommands {
		if command.phase == "F2.1" {
			keys = append(keys, command.id)
		}

	}
	for index, text := range texts {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			d := newDocument(text)
			for _, size := range [][2]int{{79, 23}, {7, 3}, {1, 1}} {
				for _, wrap := range []bool{false, true} {
					d.resize(size[0], size[1], wrap)
					for _, command := range keys {
						for repeat := 0; repeat < 3; repeat++ {
							d.navigate(command)
							if d.cursor != d.clamp(d.cursor) || d.top < 0 || d.top >= len(d.rows) || d.left < 0 {
								t.Fatalf("%s escaped bounds: cursor=%+v viewport=%d,%d", command, d.cursor, d.top, d.left)
							}
							offset := d.byteOffset(d.cursor)
							if d.positionAtByte(offset) != d.cursor || d.text != text || d.revision != 0 {
								t.Fatalf("%s changed bytes or split a character", command)
							}
						}
					}
				}
			}
		})
	}
}

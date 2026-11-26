package zxspectrum

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

func hexBytes(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(strings.ReplaceAll(s, " ", ""))
	if err != nil {
		t.Fatalf("hexBytes(%q): %v", s, err)
	}
	return b
}

// tokenizeDefault mirrors every test_zxspectrum.py call site that
// omits model= (defaults to "48").
func tokenizeDefault(t *testing.T, source string) []byte {
	t.Helper()
	program, err := Tokenize(source, "48")
	if err != nil {
		t.Fatalf("Tokenize(%q): %v", source, err)
	}
	return program
}

// body strips the 4-byte line-number+length record header a single-
// line program's Tokenize output starts with, mirroring this test
// file's own `[4:]` slicing convention (copied straight from
// test_zxspectrum.py, which slices the same way for the same reason).
func body(program []byte) []byte { return program[4:] }

// TestRawBytesAreLiteralInCodeStringsAndRem mirrors
// test_raw_bytes_are_literal_in_code_strings_and_rem.
func TestRawBytesAreLiteralInCodeStringsAndRem(t *testing.T) {
	var literal strings.Builder
	for code := range 256 {
		literal.WriteString(`\{`)
		literal.WriteString(strings.ToUpper(hexByte(byte(code))))
		literal.WriteString(`}`)
	}
	expected := make([]byte, 256)
	for i := range expected {
		expected[i] = byte(i)
	}

	got := body(tokenizeDefault(t, "10 "+literal.String()))
	want := append(append([]byte{}, expected...), 0x0d)
	if !bytes.Equal(got, want) {
		t.Fatalf("raw bytes in code: got % x", got)
	}

	got = body(tokenizeDefault(t, `10 PRINT "`+literal.String()+`"`))
	want = append(append([]byte{0xf5, ' ', '"'}, expected...), '"', 0x0d)
	if !bytes.Equal(got, want) {
		t.Fatalf("raw bytes in string: got % x", got)
	}

	got = body(tokenizeDefault(t, `10 \{EA} 123:FN "%%A"`))
	// Once \{EA} (the REM token code) fires, the rest of the line is raw
	// REM text (encodeText, raw=True) - "FN" stays literal ASCII, never
	// tokenized as the FN keyword, and quotes are not special either.
	want = hexBytes(t, "ea 20 31 32 33 3a 46 4e 20 22 25 41 22 0d")
	if !bytes.Equal(got, want) {
		t.Fatalf(`\{EA} as REM: got % x, want % x`, got, want)
	}

	a := tokenizeDefault(t, `10 DEF FN f(x\{0E}\{00}\{00}\{00}\{00}\{00})=x`)
	b := tokenizeDefault(t, "10 DEF FN f(x)=x")
	if !bytes.Equal(a, b) {
		t.Fatalf("explicit hidden slot bytes should equal the implicit ones: % x vs % x", a, b)
	}

	hashLine := body(tokenizeDefault(t, `10 \{F5}#2;"OK"`))
	printLine := body(tokenizeDefault(t, `10 PRINT#2;"OK"`))
	if !bytes.Equal(hashLine, printLine) {
		t.Fatalf(`\{F5}#2 should equal PRINT#2: % x vs % x`, hashLine, printLine)
	}

	a = tokenizeDefault(t, `10 PRINT FN f\{28}3\{29}`)
	b = tokenizeDefault(t, "10 PRINT FN f(3)")
	if !bytes.Equal(a, b) {
		t.Fatalf(`\{28}/\{29} should equal (/): % x vs % x`, a, b)
	}

	for _, text := range []string{`\{0}`, `\{000}`, `\{GG}`, `\{FF`, `\FF`} {
		if _, err := Tokenize("10 "+text, "48"); err == nil {
			t.Fatalf("expected error for %q", text)
		}
	}
}

func hexByte(b byte) string {
	const digits = "0123456789abcdef"
	return string([]byte{digits[b>>4], digits[b&0xf]})
}

// TestNativeFnParameterSlotsAndTapeStreamCommands mirrors
// test_native_fn_parameter_slots_and_tape_stream_commands.
func TestNativeFnParameterSlotsAndTapeStreamCommands(t *testing.T) {
	got := body(tokenizeDefault(t, "10 DEF FN f(x,a$)=x+LEN a$"))
	want := append(append([]byte{0xce, ' ', 'f', '('}, append([]byte{'x', 0x0e}, make([]byte, 5)...)...), ',')
	want = append(want, 'a', '$', 0x0e)
	want = append(want, make([]byte, 5)...)
	want = append(want, ')', '=', 'x', '+', 0xb1, ' ', 'a', '$', 0x0d)
	if !bytes.Equal(got, want) {
		t.Fatalf("DEF FN hidden slots: got % x, want % x", got, want)
	}

	ok := []string{
		"DEF FN z()=42", "PRINT FN z()", `PRINT LEN (FN s$("A"))`,
		`LOAD ""`, `SAVE "P" LINE 10`, `VERIFY "P" CODE 30000,16`,
		`MERGE "P"`, `LOAD "P" CODE`, `LOAD "P" SCREEN$`,
		`SAVE "P" DATA a()`, `LOAD "P" DATA a$()`,
		`OPEN#4,"S":PRINT #4;"OK":CLOSE#4`,
		`INPUT #0;LINE a$`, `LIST #2;10`, `LIST #2,10`,
		`LIST #FN s(1,2),10`, `LPRINT "P"`, "LLIST", "COPY",
	}
	for _, statement := range ok {
		if _, err := Tokenize("10 "+statement, "48"); err != nil {
			t.Errorf("Tokenize(%q): unexpected error: %v", statement, err)
		}
	}

	bad := []string{
		"DEF FN long(x)=x", "DEF FN f(long)=long", "DEF FN f(x)",
		"DEF FN f(x)=", "PRINT FN f 1", "PRINT FN f(1",
		"OPEN#4", "CLOSE#", "CLOSE#1,2", "LIST #2,10,20",
		"COPY 1", "SAVE", "LOAD",
		"LET x#=1",
	}
	for _, statement := range bad {
		if _, err := Tokenize("10 "+statement, "48"); err == nil {
			t.Errorf("Tokenize(%q): expected error", statement)
		}
	}
}

// TestGraphicAndUdgEscapesEmitRealCharacterBytes mirrors
// test_graphic_and_udg_escapes_emit_real_character_bytes.
func TestGraphicAndUdgEscapesEmitRealCharacterBytes(t *testing.T) {
	var graphics, graphicsAlt, udgs strings.Builder
	for code := 128; code < 144; code++ {
		graphics.WriteString(`\`)
		graphics.WriteString(strings.ToUpper(hexByte(byte(code))))
	}
	for letter := 'A'; letter <= 'U'; letter++ {
		udgs.WriteRune('%')
		udgs.WriteRune(letter)
	}
	for code := 128; code <= 164; code++ {
		graphicsAlt.WriteString(`\`)
		graphicsAlt.WriteString(strings.ToLower(hexByte(byte(code))))
	}
	expected := make([]byte, 0, 37)
	for code := 128; code <= 164; code++ {
		expected = append(expected, byte(code))
	}

	for _, literal := range []string{graphics.String() + udgs.String(), graphicsAlt.String()} {
		got := body(tokenizeDefault(t, `10 PRINT "`+literal+`"`))
		want := append(append([]byte{0xf5, ' ', '"'}, expected...), '"', 0x0d)
		if !bytes.Equal(got, want) {
			t.Fatalf("PRINT graphics/UDG literal %q: got % x, want % x", literal, got, want)
		}
		got = body(tokenizeDefault(t, "10 REM "+literal))
		want = append(append([]byte{0xea, ' '}, expected...), 0x0d)
		if !bytes.Equal(got, want) {
			t.Fatalf("REM graphics/UDG literal %q: got % x, want % x", literal, got, want)
		}
	}

	got := body(tokenizeDefault(t, `10 PRINT "%%ABC\\80""%a"`))
	want := append([]byte{0xf5, ' ', '"', '%', 'A', 'B', 'C', '\\', '8', '0', '"', '"'}, 0x90, '"', 0x0d)
	if !bytes.Equal(got, want) {
		t.Fatalf("literal %%/backslash + UDG: got % x, want % x", got, want)
	}

	got = body(tokenizeDefault(t, `10 REM \{00}\{A5}\{FF}`))
	want = []byte{0xea, ' ', 0x00, 0xa5, 0xff, 0x0d}
	if !bytes.Equal(got, want) {
		t.Fatalf(`REM \{XX} raw bytes: got % x, want % x`, got, want)
	}

	for _, literal := range []string{`\7F`, `\A5`, `\8`, `\GG`, "%V", "%Z"} {
		if _, err := Tokenize(`10 PRINT "`+literal+`"`, "48"); err == nil {
			t.Errorf("literal %q: expected error", literal)
		}
	}

	got = body(tokenizeDefault(t, `10 REM \{0D}`))
	want = []byte{0xea, ' ', 0x0d, 0x0d}
	if !bytes.Equal(got, want) {
		t.Fatalf(`REM \{0D}: got % x, want % x`, got, want)
	}

	for _, literal := range []string{`\{0}`, `\{GG}`} {
		if _, err := Tokenize("10 REM "+literal, "48"); err == nil {
			t.Errorf("REM %q: expected error", literal)
		}
	}
}

// TestBinaryLiteralsUseBinaryValuesInRomNumberRecords mirrors
// test_binary_literals_use_binary_values_in_rom_number_records.
func TestBinaryLiteralsUseBinaryValuesInRomNumberRecords(t *testing.T) {
	got := tokenizeDefault(t, "10 PRINT BIN 1010")
	want := hexBytes(t, "000a0f00f520c420313031300e00000a00000d")
	if !bytes.Equal(got, want) {
		t.Fatalf("BIN 1010: got % x, want % x", got, want)
	}

	cases := []struct {
		digits string
		number string
	}{
		{"0", "0000000000"}, {"1", "0000010000"},
		{"1111111111111111", "0000ffff00"},
		{"00000000000000001", "0000010000"},
		{strings.Repeat("0", 64), "0000000000"}, {"1 0\t1 0", "00000a0000"},
	}
	for _, c := range cases {
		expected := append([]byte{0xf5, ' ', 0xc4, ' '}, []byte(strings.ReplaceAll(c.digits, "\t", " "))...)
		expected = append(expected, 0x0e)
		expected = append(expected, hexBytes(t, c.number)...)
		got := body(tokenizeDefault(t, "10 PRINT BIN "+c.digits))
		expected = append(expected, 0x0d)
		if !bytes.Equal(got, expected) {
			t.Errorf("BIN %q: got % x, want % x", c.digits, got, expected)
		}
	}

	got = body(tokenizeDefault(t, `10 PRINT "BIN 102:":REM BIN 2`))
	want = []byte{0xf5, ' ', '"', 'B', 'I', 'N', ' ', '1', '0', '2', ':', '"', ':', 0xea, ' ', 'B', 'I', 'N', ' ', '2', 0x0d}
	if !bytes.Equal(got, want) {
		t.Fatalf("BIN inside string/REM not tokenized: got % x, want % x", got, want)
	}
}

// TestBinaryLiteralsRejectMissingInvalidAndOverflowingDigits mirrors
// test_binary_literals_reject_missing_invalid_and_overflowing_digits.
func TestBinaryLiteralsRejectMissingInvalidAndOverflowingDigits(t *testing.T) {
	for _, expr := range []string{
		"BIN", "BIN 2", "BIN 102", "BIN 1 2", "BIN 10.1",
		"BIN 10E2", "BIN 10x", "BIN -1", `BIN "10"`,
		"BIN (10)", "BIN 10000000000000000",
	} {
		if _, err := Tokenize("10 PRINT "+expr, "48"); err == nil {
			t.Errorf("%q: expected error", expr)
		}
	}
}

// TestNativeControlAndPortCommands mirrors
// test_native_control_and_port_commands.
func TestNativeControlAndPortCommands(t *testing.T) {
	cases := []struct {
		command  string
		expected []byte
	}{
		{"RUN", []byte{0xf7}}, {"LIST", []byte{0xf0}}, {"NEW", []byte{0xe6}},
		{"CONTINUE", []byte{0xe8}},
	}
	for _, c := range cases {
		got := body(tokenizeDefault(t, "10 "+c.command))
		want := append(append([]byte{}, c.expected...), 0x0d)
		if !bytes.Equal(got, want) {
			t.Errorf("%s: got % x, want % x", c.command, got, want)
		}
	}

	got := body(tokenizeDefault(t, "10 OUT 254,1"))
	expected := []byte{0xdf, ' ', '2', '5', '4', 0x0e, 0x00, 0x00, 0xfe, 0x00, 0x00, ',', '1', 0x0e, 0x00, 0x00, 0x01, 0x00, 0x00, 0x0d}
	if !bytes.Equal(got, expected) {
		t.Fatalf("OUT 254,1: got % x, want % x", got, expected)
	}

	for _, statement := range []string{
		"RUN start+10", "LIST ABS (-10)", "OUT ATTR (0,0),BIN 1",
		`LIST VAL "10"`, `OUT CODE ",",1`,
	} {
		if _, err := Tokenize("10 "+statement, "48"); err != nil {
			t.Errorf("Tokenize(%q): unexpected error: %v", statement, err)
		}
	}
}

// TestNativeCommandsCheckArgumentStructureAndStatementContext mirrors
// test_native_commands_check_argument_structure_and_statement_context.
func TestNativeCommandsCheckArgumentStructureAndStatementContext(t *testing.T) {
	for _, statement := range []string{
		"NEW 1", "CONTINUE 1", "RUN 1,2", "LIST 1,2",
		"OUT", "OUT 254", "OUT ,1", "OUT 254,", "OUT 1,2,3",
		"OUT (1,2", "OUT 1),2", "PRINT RUN", "LET x=LIST",
	} {
		if _, err := Tokenize("10 "+statement, "48"); err == nil {
			t.Errorf("%q: expected error", statement)
		}
	}
}

// TestModel128EnablesSpectrumAndPlayTokens mirrors
// test_model_128_enables_spectrum_and_play_tokens.
func TestModel128EnablesSpectrumAndPlayTokens(t *testing.T) {
	for _, statement := range []string{"SPECTRUM", `PLAY "a"`} {
		_, err := Tokenize("10 "+statement, "48")
		if err == nil || !strings.Contains(err.Error(), "--sp_model 128") {
			t.Errorf("%q on 48K: expected --sp_model 128 error, got %v", statement, err)
		}
	}
	got := body(mustTokenize(t, "10 SPECTRUM", "128"))
	if !bytes.Equal(got, []byte{0xa3, 0x0d}) {
		t.Fatalf("SPECTRUM on 128K: got % x", got)
	}
	got = body(mustTokenize(t, `10 PLAY "a","b","c"`, "128"))
	want := []byte{0xa4, ' ', '"', 'a', '"', ',', '"', 'b', '"', ',', '"', 'c', '"', 0x0d}
	if !bytes.Equal(got, want) {
		t.Fatalf("PLAY on 128K: got % x, want % x", got, want)
	}
	for _, statement := range []string{"SPECTRUM 1", "PLAY", `PLAY "a","b","c","d"`, `PLAY "a",`} {
		if _, err := Tokenize("10 "+statement, "128"); err == nil {
			t.Errorf("%q: expected error", statement)
		}
	}
	if _, err := Tokenize("10 STOP", "256"); err == nil {
		t.Error("model 256: expected error")
	}
}

func mustTokenize(t *testing.T, source, model string) []byte {
	t.Helper()
	program, err := Tokenize(source, model)
	if err != nil {
		t.Fatalf("Tokenize(%q, %q): %v", source, model, err)
	}
	return program
}

// TestModel128UdgTAndUBecomeReservedTokens mirrors
// test_model_128_udg_t_and_u_become_reserved_tokens.
func TestModel128UdgTAndUBecomeReservedTokens(t *testing.T) {
	got, err := EncodeText("%T%U", true, "48")
	if err != nil {
		t.Fatalf("EncodeText: %v", err)
	}
	if !bytes.Equal(got, []byte{0xa3, 0xa4}) {
		t.Fatalf("%%T%%U on 48K: got % x", got)
	}
	for _, literal := range []string{"%T", "%U", `\A3`, `\a4`} {
		if _, err := EncodeText(literal, true, "128"); err == nil {
			t.Errorf("%q on 128K: expected error", literal)
		}
	}
	for _, code := range []byte{0xA3, 0xA4} {
		raw := `\{` + strings.ToUpper(hexByte(code)) + `}`
		for _, model := range []string{"48", "128"} {
			got, err := EncodeText(raw, true, model)
			if err != nil {
				t.Fatalf("EncodeText(%q, %q): %v", raw, model, err)
			}
			if !bytes.Equal(got, []byte{code}) {
				t.Errorf("EncodeText(%q, %q) = % x, want %02x", raw, model, got, code)
			}
		}
	}
}

// TestKnownRomTokensAndLineLayout mirrors
// test_known_rom_tokens_and_line_layout.
func TestKnownRomTokensAndLineLayout(t *testing.T) {
	got := tokenizeDefault(t, "10 PRINT \"HI\"\n20 STOP")
	want := hexBytes(t, "000a0700f520224849220d00140200e20d")
	if !bytes.Equal(got, want) {
		t.Fatalf("got % x, want % x", got, want)
	}
}

// TestNumbersUseIntegerAndFloatingPointFormats mirrors
// test_numbers_use_integer_and_floating_point_formats.
func TestNumbersUseIntegerAndFloatingPointFormats(t *testing.T) {
	vectors := map[string]string{
		"0":     "0000000000",
		"1":     "0000010000",
		"255":   "0000ff0000",
		"256":   "0000000100",
		"65535": "0000ffff00",
		"65536": "9100000000",
		".5":    "8000000000",
		"0.1":   "7d4ccccccd",
		"1e2":   "0000640000",
	}
	for text, want := range vectors {
		got, err := EncodeNumber(text)
		if err != nil {
			t.Fatalf("EncodeNumber(%q): %v", text, err)
		}
		if hex.EncodeToString(got) != want {
			t.Errorf("EncodeNumber(%q) = %x, want %s", text, got, want)
		}
	}

	got := body(tokenizeDefault(t, "10 LET score2=-1"))
	want := []byte{0xf1, ' ', 's', 'c', 'o', 'r', 'e', '2', '=', '-', '1', 0x0e, 0x00, 0x00, 0x01, 0x00, 0x00, 0x0d}
	if !bytes.Equal(got, want) {
		t.Fatalf("LET score2=-1: got % x, want % x", got, want)
	}
}

// TestStringsAndRemAreNotTokenizedAsCode mirrors
// test_strings_and_rem_are_not_tokenized_as_code.
func TestStringsAndRemAreNotTokenizedAsCode(t *testing.T) {
	got := body(tokenizeDefault(t, `10 PRINT "GO TO 20 ""HI"""'"X": REM 30: PRINT`))
	want := []byte{0xf5, ' ', '"', 'G', 'O', ' ', 'T', 'O', ' ', '2', '0', ' ', '"', '"', 'H', 'I', '"', '"', '"', '\'', '"', 'X', '"', ':', ' ', 0xea, ' ', '3', '0', ':', ' ', 'P', 'R', 'I', 'N', 'T', 0x0d}
	if !bytes.Equal(got, want) {
		t.Fatalf("got % x, want % x", got, want)
	}
}

// TestJumpAliasAndRelationalTokens mirrors
// test_jump_alias_and_relational_tokens.
func TestJumpAliasAndRelationalTokens(t *testing.T) {
	a := tokenizeDefault(t, "10 IF count<=3 THEN GOTO 10")
	b := tokenizeDefault(t, "10 IF count<=3 THEN GO TO 10")
	if !bytes.Equal(a, b) {
		t.Fatalf("GOTO should equal GO TO: % x vs % x", a, b)
	}
	if !bytes.Contains(b, []byte{0xc7}) {
		t.Errorf("expected <= token 0xc7 in % x", b)
	}
	if !bytes.Contains(b, []byte{0xec}) {
		t.Errorf("expected GO TO token 0xec in % x", b)
	}
}

// TestNativeFunctionAndStringArrayBytes mirrors
// test_native_function_and_string_array_bytes.
func TestNativeFunctionAndStringArrayBytes(t *testing.T) {
	got := tokenizeDefault(t, "10 PRINT LEN a$")
	want := hexBytes(t, "000a0700f520b12061240d")
	if !bytes.Equal(got, want) {
		t.Fatalf("got % x, want % x", got, want)
	}

	dim := tokenizeDefault(t, "10 DIM a$(2,5)")
	needle := []byte{'a', '$', '(', '2', 0x0e, 0x00, 0x00, 0x02, 0x00, 0x00, ','}
	if !bytes.Contains(dim, needle) {
		t.Errorf("DIM a$(2,5): missing %x in % x", needle, dim)
	}

	idx := tokenizeDefault(t, "10 PRINT a$(1 TO 2)")
	needle2 := []byte{'a', '$', '(', '1', 0x0e, 0x00, 0x00, 0x01, 0x00, 0x00, ' ', 0xcc, ' '}
	if !bytes.Contains(idx, needle2) {
		t.Errorf("a$(1 TO 2): missing %x in % x", needle2, idx)
	}
}

// TestRejectsUnrepresentableOrUnsupportedSource mirrors
// test_rejects_unrepresentable_or_unsupported_source.
func TestRejectsUnrepresentableOrUnsupportedSource(t *testing.T) {
	sources := []string{
		"", `PRINT "HI"`, "0 STOP", "10 STOP\n10 STOP", "20 STOP\n10 STOP",
		"10000 STOP", `10 PRINT "oops`, "10 PRINT 1e+", "10 LET n=1e1000",
		"10 PRINT 1e-1000", "10 PRINT \"é\"", "10 CAT",
		`10 LET name$="X"`, "10 LET total_score=1", "10 LET array(1)=2",
		"10 a=1", "10 IF a=1 THEN 20", "10 PRINT 1:",
		"10 FOR counter=1 TO 2", "10 NEXT count", "10 DEF FN long(x)=x",
		"10 PRINT 1e999999999999999999999999",
	}
	for _, source := range sources {
		if _, err := Tokenize(source, "48"); err == nil {
			t.Errorf("source %q: expected error", source)
		}
	}
}

// --- TapeTests ---

// TestAutorunHeaderAndValidation mirrors test_autorun_header_and_validation.
func TestAutorunHeaderAndValidation(t *testing.T) {
	program := tokenizeDefault(t, "10 STOP")
	for _, line := range []int{0, 10, 9999} {
		l := line
		tap, err := MakeTap(program, "RUN", &l)
		if err != nil {
			t.Fatalf("MakeTap autorun=%d: %v", line, err)
		}
		got := int(tap[16]) | int(tap[17])<<8
		if got != line {
			t.Errorf("autorun=%d: header says %d", line, got)
		}
		var checksum byte
		for _, b := range tap[2:21] {
			checksum ^= b
		}
		if checksum != 0 {
			t.Errorf("autorun=%d: header checksum = %x, want 0", line, checksum)
		}
		tzx, err := MakeTzx(program, "RUN", &l)
		if err != nil {
			t.Fatalf("MakeTzx autorun=%d: %v", line, err)
		}
		if !bytes.Equal(tzx[13:34], tap[:21]) {
			t.Errorf("autorun=%d: tzx header block does not match tap", line)
		}
	}
	for _, line := range []int{-1, 10000} {
		l := line
		if _, err := MakeTap(program, "RUN", &l); err == nil {
			t.Errorf("autorun=%d: expected error", line)
		}
	}
}

// TestTzxStandardBlocksMatchTap mirrors test_tzx_standard_blocks_match_tap.
func TestTzxStandardBlocksMatchTap(t *testing.T) {
	program := tokenizeDefault(t, "10 STOP")
	tap, err := MakeTap(program, "HI", nil)
	if err != nil {
		t.Fatal(err)
	}
	tzx, err := MakeTzx(program, "HI", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tzx[:10], []byte("ZXTape!\x1a\x01\x14")) {
		t.Fatalf("tzx signature: % x", tzx[:10])
	}
	if !bytes.Equal(tzx[10:13], []byte{0x10, 0xe8, 0x03}) {
		t.Fatalf("tzx first block id/pause: % x", tzx[10:13])
	}
	if !bytes.Equal(tzx[13:34], tap[:21]) {
		t.Fatalf("tzx header payload mismatch")
	}
	if !bytes.Equal(tzx[34:37], []byte{0x10, 0xe8, 0x03}) {
		t.Fatalf("tzx second block id/pause: % x", tzx[34:37])
	}
	if !bytes.Equal(tzx[37:], tap[21:]) {
		t.Fatalf("tzx data payload mismatch")
	}
}

// TestKnownTapHeaderAndChecksum mirrors test_known_tap_header_and_checksum.
func TestKnownTapHeaderAndChecksum(t *testing.T) {
	program := tokenizeDefault(t, "10 STOP")
	tape, err := MakeTap(program, "HI", nil)
	if err != nil {
		t.Fatal(err)
	}
	want := hexBytes(t, "130000004849202020202020202006000080060081"+
		"0800ff000a0200e20d18")
	if !bytes.Equal(tape, want) {
		t.Fatalf("got % x, want % x", tape, want)
	}

	offset := 0
	var blocks [][]byte
	for offset < len(tape) {
		size := int(tape[offset]) | int(tape[offset+1])<<8
		block := tape[offset+2 : offset+2+size]
		var checksum byte
		for _, b := range block {
			checksum ^= b
		}
		if checksum != 0 {
			t.Fatalf("block checksum = %x, want 0", checksum)
		}
		blocks = append(blocks, block)
		offset += size + 2
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	header, data := blocks[0], blocks[1]
	if !bytes.Equal(header[:2], []byte{0x00, 0x00}) {
		t.Errorf("header flag: % x", header[:2])
	}
	if string(header[2:12]) != "HI        " {
		t.Errorf("header name: %q", header[2:12])
	}
	if data[0] != 255 {
		t.Errorf("data flag: %d", data[0])
	}
	if !bytes.Equal(data[1:len(data)-1], program) {
		t.Errorf("data payload does not match program")
	}
}

// TestRejectsInvalidNameAndOversizedProgram mirrors
// test_rejects_invalid_name_and_oversized_program.
func TestRejectsInvalidNameAndOversizedProgram(t *testing.T) {
	program := tokenizeDefault(t, "10 STOP")
	for _, name := range []string{"", "ELEVENCHARS", "é", "BAD\nNAME"} {
		if _, err := MakeTap(program, name, nil); err == nil {
			t.Errorf("name %q: expected error", name)
		}
	}
	if _, err := MakeTap(make([]byte, MaxProgramBytes+1), "BIG", nil); err == nil {
		t.Error("oversized program: expected error")
	}
}

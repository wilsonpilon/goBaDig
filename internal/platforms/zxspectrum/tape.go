package zxspectrum

// MaxProgramBytes mirrors MAX_PROGRAM_BYTES (tape_zxspectrum.py:7): the
// default 48K PROG=23755/RAMTOP=65367 initial program area ceiling -
// runtime variables/stacks need additional space beyond this.
const MaxProgramBytes = 65367 - 23755

// tapBlock mirrors tap_block (tape_zxspectrum.py:10-17): wraps data in
// a standard TAP block (2-byte little-endian length, the data, then a
// 1-byte XOR checksum of the data).
func tapBlock(data []byte) ([]byte, error) {
	var checksum byte
	for _, b := range data {
		checksum ^= b
	}
	block := append(append([]byte{}, data...), checksum)
	if len(block) > 65535 {
		return nil, tokenizeErrorf("TAP block exceeds 65535 bytes")
	}
	size := len(block)
	out := make([]byte, 0, 2+len(block))
	out = append(out, byte(size), byte(size>>8))
	return append(out, block...), nil
}

// MakeTap mirrors make_tap (tape_zxspectrum.py:20-34): wraps an
// already-tokenized program (Tokenize's own output) into a standard
// ROM-loadable .tap - a header block (name/size/autostart) followed by
// a data block (flag byte 0xFF + the program itself). autorun is nil
// for "no autostart line" (encoded as the ROM's own 32768 sentinel),
// matching the original's autorun=None.
func MakeTap(program []byte, name string, autorun *int) ([]byte, error) {
	if len(program) == 0 {
		return nil, tokenizeErrorf("Cannot save an empty BASIC program")
	}
	if len(program) > MaxProgramBytes {
		return nil, tokenizeErrorf(
			"BASIC exceeds the initial 48K program area (%d bytes); runtime variables and stacks need additional space",
			MaxProgramBytes)
	}
	if len(name) < 1 || len(name) > 10 {
		return nil, tokenizeErrorf("Tape name must contain 1..10 printable ASCII characters")
	}
	for i := 0; i < len(name); i++ {
		if name[i] < 32 || name[i] > 126 {
			return nil, tokenizeErrorf("Tape name must contain 1..10 printable ASCII characters")
		}
	}
	if autorun != nil && (*autorun < 0 || *autorun > 9999) {
		return nil, tokenizeErrorf("Autorun must be a line within 0..9999 or None")
	}

	size := len(program)
	header := make([]byte, 0, 2+10+2+2+2)
	header = append(header, 0x00, 0x00)
	nameBytes := []byte(name)
	for len(nameBytes) < 10 {
		nameBytes = append(nameBytes, ' ')
	}
	header = append(header, nameBytes...)
	header = append(header, byte(size), byte(size>>8))
	autostart := 32768
	if autorun != nil {
		autostart = *autorun
	}
	header = append(header, byte(autostart), byte(autostart>>8))
	header = append(header, byte(size), byte(size>>8))

	headerBlock, err := tapBlock(header)
	if err != nil {
		return nil, err
	}
	dataBlock, err := tapBlock(append([]byte{0xff}, program...))
	if err != nil {
		return nil, err
	}
	return append(headerBlock, dataBlock...), nil
}

// MakeTzx mirrors make_tzx (tape_zxspectrum.py:37-47): TZX 1.20 with a
// fixed signature, re-wrapping every block of an already-built .tap as
// a "Standard Speed Data Block" (id 0x10) with a fixed one-second
// (1000 ms, 0x03E8 little-endian) pause.
func MakeTzx(program []byte, name string, autorun *int) ([]byte, error) {
	tape, err := MakeTap(program, name, autorun)
	if err != nil {
		return nil, err
	}
	result := []byte("ZXTape!\x1a\x01\x14")
	offset := 0
	for offset < len(tape) {
		size := int(tape[offset]) | int(tape[offset+1])<<8
		result = append(result, 0x10, 0xe8, 0x03)
		result = append(result, tape[offset:offset+size+2]...)
		offset += size + 2
	}
	return result, nil
}
